// SPDX-License-Identifier: MIT
// Runs a bundled scene on its presets, twice each, and prints what algo viz check needs.
// The scene must be pure (rule V01): Math.random, timers, fetch and the DOM throw here.
"use strict";
const fs = require("fs");
const slug = process.argv[2];
const deny = (name) => () => { throw new Error(`the scene calls ${name} (V01)`); };
Math.random = deny("Math.random");
globalThis.setTimeout = deny("setTimeout");
globalThis.setInterval = deny("setInterval");
globalThis.fetch = deny("fetch");
globalThis.window = globalThis;
new Function(fs.readFileSync("scene.js", "utf8"))();
const scene = globalThis.AlgoScenes && globalThis.AlgoScenes[slug];
if (!scene) {
  process.stderr.write(`the bundle registers no scene with the slug ${slug}\n`);
  process.exit(1);
}
const presets = JSON.parse(fs.readFileSync("presets.json", "utf8"));
const STATES = new Set(["default", "active", "compare", "swap", "pivot", "frontier", "visited", "path", "rejected", "invariant"]);
// The place of the state in the arguments of each event that has one.
const STATE_AT = { "array.mark": 3, "graph.mark": 3, "graph.edge": 4, "tree.mark": 3, "tree.edge": 3, "table.mark": 3, "heap.mark": 3 };
// validate walks the events of a trace. Each event of a view names a view of its kind that exists, and each state is in the vocabulary (V09). It gives the first problems.
function validate(t) {
  const views = new Map(), bad = [];
  t.steps.forEach((s, i) => {
    for (const e of s.ev) {
      const [op, name] = e;
      const kind = op.split(".")[0];
      if (op === "destroy") { views.delete(name); continue; }
      if (op.endsWith(".create")) { views.set(name, kind); continue; }
      if (!["array", "graph", "tree", "table", "heap"].includes(kind)) { bad.push(`step ${i + 1}: the event ${op} is not in @algo/viz`); continue; }
      if (op === "array.match" && views.get(e[3]) !== "array") bad.push(`step ${i + 1}: array.match names ${JSON.stringify(e[3])}, which is not a string view that exists`);
      if (views.get(name) !== kind) bad.push(`step ${i + 1}: ${op} names ${JSON.stringify(name)}, which is not a ${kind} view that exists`);
      const at = STATE_AT[op];
      if (at !== undefined && e[at] !== null && !STATES.has(e[at])) bad.push(`step ${i + 1}: the state ${JSON.stringify(e[at])} is not in the vocabulary (V09)`);
    }
  });
  return bad.slice(0, 5);
}
const out = [];
for (const p of presets) {
  try {
    const a = JSON.stringify(globalThis.AlgoViz.trace(scene, p.input));
    const b = JSON.stringify(globalThis.AlgoViz.trace(scene, p.input));
    const t = JSON.parse(a);
    out.push({ name: p.name, same: a === b, steps: t.steps.length, lines: [...new Set(t.steps.map((s) => s.line))], invalid: validate(t) });
  } catch (e) {
    out.push({ name: p.name, error: String(e && e.message ? e.message : e) });
  }
}
process.stdout.write(JSON.stringify({ presets: out }));
