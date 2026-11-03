// SPDX-License-Identifier: Apache-2.0

package formal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tamnd/algorithm-agent/node"
)

func TestLoadAndCheck(t *testing.T) {
	path := filepath.Join(t.TempDir(), "decls.jsonl")
	lines := []string{
		`{"decl":"Algo.M.model","kind":"def","module":"Algo.M","axioms":[],"computable":true}`,
		`{"decl":"Algo.M.good","kind":"theorem","module":"Algo.M","axioms":["Classical.choice","Quot.sound","propext"]}`,
		`{"decl":"Algo.M.bad","kind":"theorem","module":"Algo.M","axioms":["propext","sorryAx"],"sorry":true}`,
		`{"decl":"Algo.M.fast","kind":"theorem","module":"Algo.M","axioms":["Lean.ofReduceBool"]}`,
	}
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ds, err := Load(path)
	if err != nil || len(ds) != 4 {
		t.Fatalf("Load = %d decls, %v", len(ds), err)
	}
	c := &node.Corpus{Nodes: []*node.Node{{Slug: "m", Lean: map[string]string{
		"module":  "Algo.M",
		"model":   "Algo.M.good",
		"correct": "Algo.M.missing",
	}}}}
	var got []string
	for _, f := range Check(c, ds) {
		got = append(got, f.String())
	}
	want := []string{
		"L03 Algo.M.bad: uses the axiom sorryAx (a sorry)",
		"L03 Algo.M.fast: uses the axiom Lean.ofReduceBool",
		"L01 m lean.correct: Algo.M.missing is not a declaration of the Algo package",
		"L01 m lean.model: Algo.M.good is a theorem, want a def",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("Check =\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}
