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

func TestClasses(t *testing.T) {
	for typ, want := range map[string]string{
		"(fun n => ↑n * ↑(Nat.clog 2 n)) =O[Filter.atTop] fun n => ↑n * Real.log ↑n": "nlogn",
		"f =Θ[Filter.atTop] fun (n : ℕ) => (n : ℝ) ^ 2":                              "n^2",
		"∀ n, f n ≤ n": "",
	} {
		if got := ClassOf(typ); got != want {
			t.Errorf("ClassOf(%q) = %q, want %q", typ, got, want)
		}
	}
	for claim, want := range map[string]string{"Θ(n log n)": "nlogn", "O(n lg n)": "nlogn", "Θ(n^2)": "n^2", "Θ(n²)": "n²", "amortized": ""} {
		if got := ClaimClass(claim); got != want {
			t.Errorf("ClaimClass(%q) = %q, want %q", claim, got, want)
		}
	}
}

func TestCheckClaims(t *testing.T) {
	ds := Decls{
		"Algo.M.cost":  {Decl: "Algo.M.cost", Kind: "theorem", Module: "Algo.M"},
		"Algo.M.class": {Decl: "Algo.M.class", Kind: "theorem", Module: "Algo.M", Type: "f =O[Filter.atTop] fun n => ↑n * Real.log ↑n"},
	}
	n := &node.Node{Slug: "m", Lean: map[string]string{"module": "Algo.M", "cost": "Algo.M.cost"}, Complexity: &node.Complexity{
		Time:  map[string]string{"worst": "Θ(n log n)", "best": "Θ(n)"},
		Space: map[string]string{"extra": "Θ(n)"},
	}}
	var got []string
	for _, f := range CheckClaims(n, ds) {
		got = append(got, f.String())
	}
	want := "L05 m time.best: Θ(n) has no class lemma in Algo.M"
	if strings.Join(got, "\n") != want {
		t.Errorf("CheckClaims = %q, want %q", got, want)
	}
	delete(n.Lean, "cost")
	if fs := CheckClaims(n, ds); len(fs) != 0 {
		t.Errorf("a node without lean.cost gets findings: %v", fs)
	}
}
