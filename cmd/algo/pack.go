package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Cloudflare Pages takes at most 20,000 files in one deployment. The site has a page, a Markdown copy and a JSON copy for each node, and a page and a raw file for each implementation, so a deploy packs the files that readers ask for less often.
//
// The pass moves each packed file into a pack: packs/N.pack holds the bytes of its files one after another, and packs/N.json maps the path of each file to its offset and length. The pack of a path is the FNV-1a hash of the path modulo the count of packs. The pass chooses the count so that a pack has about packBytes bytes.
//
// The pass then writes _worker.js and _routes.json. The routes send only the packed paths to the worker. The worker finds the pack of the path, reads the index, and reads the range of the pack that holds the file. It gives each other path to the static files.

// packDir is the directory of the packs in the site tree.
const packDir = "packs"

// packBytes is the size of a pack that the pass aims for. The worker reads a whole pack when the host does not answer a range request, so a pack stays small.
const packBytes = 512 << 10

// maxPacked is the size above which a file stays a static file.
const maxPacked = 256 << 10

// maxRoutes is the count of include and exclude rules that Pages takes in _routes.json.
const maxRoutes = 100

// packSub is the set of the second parts of a path, as in /SLUG/proof/, whose pages are packed. The pass adds the language of each raw file it finds.
var packSub = []string{"proof", "lean", "visualize"}

// packOf returns the pack of the path p, for n packs. The worker computes the same hash.
func packOf(p string, n int) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(p))
	return int(h.Sum32() % uint32(n)) // #nosec G115 -- n is a small positive count.
}

// packed reports the files of the site tree that the pass packs, and the second parts of the packed subpages.
func packed(dir string) (files []string, subs []string, err error) {
	sub := map[string]bool{}
	for _, s := range packSub {
		sub[s] = true
	}
	var all []string
	err = filepath.WalkDir(dir, func(p string, e fs.DirEntry, err error) error {
		if err != nil || e.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		rel = filepath.ToSlash(rel)
		all = append(all, rel)
		if parts := strings.Split(rel, "/"); len(parts) == 2 && strings.HasSuffix(parts[1], ".txt") && isPage(dir, parts[0]) {
			sub[strings.TrimSuffix(parts[1], ".txt")] = true
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	for _, rel := range all {
		if !packable(dir, rel, sub) {
			continue
		}
		info, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel)))
		if err != nil {
			return nil, nil, err
		}
		if info.Size() <= maxPacked {
			files = append(files, rel)
		}
	}
	for s := range sub {
		subs = append(subs, s)
	}
	sort.Strings(subs)
	return files, subs, nil
}

// isPage reports whether the directory name at the top of the tree holds the page of a node.
func isPage(dir, name string) bool {
	if name == packDir || name == "assets" || name == "embed" {
		return false
	}
	_, err := os.Stat(filepath.Join(dir, name, "index.html"))
	return err == nil
}

// packable reports whether the pass packs the file rel: the Markdown and JSON copy of a node, the raw file of an implementation, a subpage of a node in sub, and an embed page.
func packable(dir, rel string, sub map[string]bool) bool {
	parts := strings.Split(rel, "/")
	switch len(parts) {
	case 1:
		ext := filepath.Ext(rel)
		return (ext == ".md" || ext == ".json") && isPage(dir, strings.TrimSuffix(rel, ext))
	case 2:
		return strings.HasSuffix(parts[1], ".txt") && sub[strings.TrimSuffix(parts[1], ".txt")] && isPage(dir, parts[0])
	case 3:
		if parts[2] != "index.html" {
			return false
		}
		if parts[0] == "embed" {
			return true
		}
		return sub[parts[1]] && isPage(dir, parts[0])
	}
	return false
}

// headerRule is one block of _headers: a path pattern and its headers.
type headerRule struct {
	Pattern string            `json:"p"`
	Headers map[string]string `json:"h"`
}

// headerRules reads the _headers file of the site tree. Pages does not apply it to the responses of the worker, so the worker applies the rules itself. A later rule replaces a header of an earlier rule.
func headerRules(dir string) ([]headerRule, error) {
	data, err := os.ReadFile(filepath.Join(dir, "_headers"))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var rules []headerRule
	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		line := sc.Text()
		t := strings.TrimSpace(line)
		switch {
		case t == "" || strings.HasPrefix(t, "#"):
		case !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t"):
			pat := "^" + strings.ReplaceAll(regexp.QuoteMeta(t), `\*`, ".*") + "$"
			rules = append(rules, headerRule{Pattern: pat, Headers: map[string]string{}})
		case len(rules) > 0:
			if k, v, ok := strings.Cut(t, ":"); ok {
				rules[len(rules)-1].Headers[strings.ToLower(strings.TrimSpace(k))] = strings.TrimSpace(v)
			}
		}
	}
	return rules, sc.Err()
}

// packSite packs the files of the site tree at dir and writes the worker and its routes. build names the deployment, so that the cache of the worker does not keep the files of an earlier one. It returns the count of packed files.
func packSite(dir, build string) (int, error) {
	if _, err := os.Stat(filepath.Join(dir, packDir)); err == nil {
		return 0, nil // The tree is packed already, as when deploy runs with -no-build on a tree it deployed before.
	}
	files, subs, err := packed(dir)
	if err != nil {
		return 0, err
	}
	if len(files) == 0 {
		return 0, nil
	}
	var total int64
	sizes := map[string]int64{}
	for _, f := range files {
		info, err := os.Stat(filepath.Join(dir, filepath.FromSlash(f)))
		if err != nil {
			return 0, err
		}
		sizes[f] = info.Size()
		total += info.Size()
	}
	n := int(total/packBytes) + 1
	groups := make([][]string, n)
	for _, f := range files {
		i := packOf(f, n)
		groups[i] = append(groups[i], f)
	}
	if err := os.MkdirAll(filepath.Join(dir, packDir), 0o755); err != nil {
		return 0, err
	}
	for i, g := range groups {
		var data bytes.Buffer
		index := map[string][2]int64{}
		for _, f := range g {
			b, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(f)))
			if err != nil {
				return 0, err
			}
			index[f] = [2]int64{int64(data.Len()), int64(len(b))}
			data.Write(b)
		}
		ix, err := json.Marshal(index)
		if err != nil {
			return 0, err
		}
		base := filepath.Join(dir, packDir, fmt.Sprint(i))
		if err := os.WriteFile(base+".pack", data.Bytes(), 0o644); err != nil {
			return 0, err
		}
		if err := os.WriteFile(base+".json", ix, 0o644); err != nil {
			return 0, err
		}
	}
	for _, f := range files {
		if err := os.Remove(filepath.Join(dir, filepath.FromSlash(f))); err != nil {
			return 0, err
		}
	}
	removeEmptyDirs(dir)

	rules, err := headerRules(dir)
	if err != nil {
		return 0, err
	}
	rj, err := json.Marshal(rules)
	if err != nil {
		return 0, err
	}
	bj, _ := json.Marshal(build)
	w := strings.NewReplacer("PACKS_COUNT", fmt.Sprint(n), "BUILD_NAME", string(bj), "HEADER_RULES", string(rj)).Replace(workerJS)
	if err := os.WriteFile(filepath.Join(dir, "_worker.js"), []byte(w), 0o644); err != nil {
		return 0, err
	}
	routes, err := json.MarshalIndent(packRoutes(subs), "", "  ")
	if err != nil {
		return 0, err
	}
	if err := os.WriteFile(filepath.Join(dir, "_routes.json"), append(routes, '\n'), 0o644); err != nil {
		return 0, err
	}
	return len(files), nil
}

// pagesRoutes is the content of _routes.json.
type pagesRoutes struct {
	Version int      `json:"version"`
	Include []string `json:"include"`
	Exclude []string `json:"exclude"`
}

// packRoutes returns the routes of the worker: the packed paths, and the same paths without the final slash, so that the worker can add it.
func packRoutes(subs []string) pagesRoutes {
	r := pagesRoutes{Version: 1, Include: []string{"/*.md", "/*.json", "/*.txt", "/embed/*"}, Exclude: []string{"/assets/*", "/" + packDir + "/*"}}
	for _, s := range subs {
		if len(r.Include)+len(r.Exclude)+2 > maxRoutes {
			// Too many subpages for the routes: the worker then runs for each path of two parts and more.
			r.Include = []string{"/*"}
			r.Exclude = []string{"/assets/*", "/" + packDir + "/*"}
			return r
		}
		r.Include = append(r.Include, "/*/"+s+"/*", "/*/"+s)
	}
	return r
}

// removeEmptyDirs removes each empty directory under dir.
func removeEmptyDirs(dir string) {
	var dirs []string
	_ = filepath.WalkDir(dir, func(p string, e fs.DirEntry, err error) error {
		if err == nil && e.IsDir() && p != dir {
			dirs = append(dirs, p)
		}
		return nil
	})
	for i := len(dirs) - 1; i >= 0; i-- {
		_ = os.Remove(dirs[i]) // Remove fails on a directory that is not empty, which is what the pass wants.
	}
}

// workerJS is the _worker.js of a deployment. PACKS_COUNT is the count of packs, BUILD_NAME the name of the deployment and HEADER_RULES the rules of _headers.
const workerJS = `// The worker of the algorithm site. algo site deploy writes this file.
// A request for a packed file gets the file from its pack with a range read. Each other request gets the static file.
const packs = PACKS_COUNT;
const build = BUILD_NAME;
const rules = HEADER_RULES.map((r) => [new RegExp(r.p), r.h]);
const types = { html: "text/html; charset=utf-8", md: "text/markdown; charset=utf-8", json: "application/json; charset=utf-8", txt: "text/plain; charset=utf-8" };
const indexes = new Map();
const bodies = new Map();

// packOf is the FNV-1a hash of the path modulo the count of packs, as in the build.
function packOf(p) {
  let h = 0x811c9dc5;
  for (const c of new TextEncoder().encode(p)) {
    h ^= c;
    h = Math.imul(h, 0x01000193) >>> 0;
  }
  return h % packs;
}

function remember(map, key, value, size) {
  if (map.size >= size) map.delete(map.keys().next().value);
  map.set(key, value);
}

// locate returns the pack, offset and length of the path p, or undefined.
async function locate(env, url, p) {
  const i = packOf(p);
  let index = indexes.get(i);
  if (!index) {
    const r = await env.ASSETS.fetch(new URL("/` + packDir + `/" + i + ".json", url));
    if (!r.ok) return undefined;
    index = await r.json();
    remember(indexes, i, index, 512);
  }
  const e = index[p];
  return e && { pack: i, offset: e[0], length: e[1] };
}

// read returns the bytes of a located file. It asks for the range of the file. A host that does not answer ranges sends the whole pack, and the worker keeps it for the next file.
async function read(env, url, e) {
  let whole = bodies.get(e.pack);
  if (!whole) {
    const req = new Request(new URL("/` + packDir + `/" + e.pack + ".pack", url), { headers: { range: "bytes=" + e.offset + "-" + (e.offset + e.length - 1) } });
    const r = await env.ASSETS.fetch(req);
    if (!r.ok) return undefined;
    const b = await r.arrayBuffer();
    if (r.status === 206) return b.byteLength === e.length ? b : undefined;
    whole = b;
    remember(bodies, e.pack, whole, 24);
  }
  if (e.offset + e.length > whole.byteLength) return undefined;
  return whole.slice(e.offset, e.offset + e.length);
}

export default {
  async fetch(request, env, ctx) {
    if (request.method !== "GET" && request.method !== "HEAD") return env.ASSETS.fetch(request);
    const url = new URL(request.url);
    let p;
    try {
      p = decodeURIComponent(url.pathname).replace(/^\/+/, "");
    } catch (_) {
      return env.ASSETS.fetch(request);
    }
    if (p === "" || p.startsWith("` + packDir + `/")) return env.ASSETS.fetch(request);
    if (p.endsWith("/")) p += "index.html";
    else if (!p.slice(p.lastIndexOf("/") + 1).includes(".")) {
      if (await locate(env, url, p + "/index.html")) return Response.redirect(url.origin + url.pathname + "/" + url.search, 308);
      return env.ASSETS.fetch(request);
    }
    // The cache key names the deployment, so a new deployment does not serve the files of an earlier one.
    const key = new Request(new URL("/__pack/" + build + "/" + p, url));
    if (request.method === "GET") {
      const hit = await caches.default.match(key);
      if (hit) return hit;
    }
    const e = await locate(env, url, p);
    const body = e && (await read(env, url, e));
    if (body === undefined) return env.ASSETS.fetch(request);
    const headers = new Headers({ "content-type": types[p.slice(p.lastIndexOf(".") + 1)] || "application/octet-stream", "cache-control": "public, max-age=0, must-revalidate" });
    const path = "/" + p.replace(/(^|\/)index\.html$/, "$1");
    for (const [re, h] of rules) {
      if (re.test(path)) for (const k in h) headers.set(k, h[k]);
    }
    const response = new Response(request.method === "HEAD" ? null : body, { headers });
    if (request.method === "GET") ctx.waitUntil(caches.default.put(key, response.clone()));
    return response;
  },
};
`
