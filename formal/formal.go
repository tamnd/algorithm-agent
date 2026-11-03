// SPDX-License-Identifier: Apache-2.0

// Package formal builds the Lean package of the content repository and reads what it proves (spec 07).
//
// Build runs lake build in the lean directory, then runs AlgoExtract (Extract.lean) with lake env lean. AlgoExtract writes one JSON line for each declaration in the namespace Algo to decls.jsonl. Check reads that file and reports each declaration that uses an axiom outside the policy of 07.5, and each name of a lean field that does not exist.
package formal

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/tamnd/algorithm-agent/node"
)

//go:embed Extract.lean
var extractSrc []byte

// Allowed are the axioms that a proof can use (spec 07.5). Each other axiom, such as sorryAx or Lean.ofReduceBool, fails the build.
var Allowed = []string{"propext", "Classical.choice", "Quot.sound"}

// Decl is one line of decls.jsonl.
type Decl struct {
	Decl       string   `json:"decl"`
	Kind       string   `json:"kind"`
	Module     string   `json:"module"`
	Type       string   `json:"type"`
	TypeHash   string   `json:"type_hash"`
	Axioms     []string `json:"axioms"`
	Deps       []string `json:"deps"`
	Sorry      bool     `json:"sorry"`
	Computable bool     `json:"computable"`
	Doc        string   `json:"doc,omitempty"`
	Pos        *Pos     `json:"pos,omitempty"`
}

// Pos is the line range of a declaration in its module.
type Pos struct {
	Line    int `json:"line"`
	EndLine int `json:"end_line"`
}

// Disallowed returns the axioms of d that the policy does not permit.
func (d Decl) Disallowed() []string {
	var out []string
	for _, a := range d.Axioms {
		if !slices.Contains(Allowed, a) {
			out = append(out, a)
		}
	}
	return out
}

// Decls maps each full name to its declaration.
type Decls map[string]Decl

// Load reads a decls.jsonl file.
func Load(path string) (Decls, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := Decls{}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<16), 1<<24)
	for i := 1; sc.Scan(); i++ {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		var d Decl
		if err := json.Unmarshal(line, &d); err != nil {
			return nil, fmt.Errorf("%s:%d: %w", path, i, err)
		}
		out[d.Decl] = d
	}
	return out, sc.Err()
}

// Options tell Build where the package is and how to run lake.
type Options struct {
	// Dir is the Lean package, the lean directory of the content repository.
	Dir string
	// Out is the directory for decls.jsonl and the copy of Extract.lean, such as build/lean.
	Out string
	// NoBuild skips lake build, when the package is built already.
	NoBuild bool
	// Wrap changes a command before it runs, for example to run it in a container. Nil runs it on the host.
	Wrap func(argv []string) []string
	// Log gets the output of lake.
	Log io.Writer
}

// Build runs lake build and AlgoExtract, and returns the path of decls.jsonl.
func Build(ctx context.Context, o Options) (string, error) {
	dir, err := filepath.Abs(o.Dir)
	if err != nil {
		return "", err
	}
	out, err := filepath.Abs(o.Out)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(filepath.Join(dir, "lakefile.toml")); err != nil {
		return "", fmt.Errorf("%s is not a Lake package: %w", dir, err)
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return "", err
	}
	if !o.NoBuild {
		if err := o.run(ctx, dir, nil, "lake", "build"); err != nil {
			return "", fmt.Errorf("lake build: %w", err)
		}
	}
	src := filepath.Join(out, "Extract.lean")
	if err := os.WriteFile(src, extractSrc, 0o644); err != nil {
		return "", err
	}
	decls := filepath.Join(out, "decls.jsonl")
	_ = os.Remove(decls)
	if err := o.run(ctx, dir, []string{"ALGO_DECLS=" + decls}, "lake", "env", "lean", src); err != nil {
		return "", fmt.Errorf("AlgoExtract: %w", err)
	}
	return decls, nil
}

func (o Options) run(ctx context.Context, dir string, env []string, argv ...string) error {
	if o.Wrap != nil {
		// The container gets the variables as arguments of env, because the engine does not pass the environment of the client.
		if len(env) > 0 {
			argv = append(append([]string{"env"}, env...), argv...)
			env = nil
		}
		argv = o.Wrap(argv)
	}
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	log := o.Log
	if log == nil {
		log = io.Discard
	}
	cmd.Stdout, cmd.Stderr = log, log
	return cmd.Run()
}

// Finding is one problem that Check finds.
type Finding struct {
	// Rule is the audit rule of 07.15, such as L01 or L03.
	Rule string
	// Where is the declaration, or the node and its field.
	Where string
	Msg   string
}

func (f Finding) String() string { return f.Rule + " " + f.Where + ": " + f.Msg }

// Kinds are the kinds that each key of a lean field must name.
var Kinds = map[string][]string{
	"model":   {"def"},
	"spec":    {"def", "inductive", "structure"},
	"correct": {"theorem"},
	"cost":    {"theorem"},
	"space":   {"theorem"},
}

// Check reports each declaration that uses an axiom outside the policy (rule L03), and each name of a lean field that is not in ds or has the wrong kind (rule L01). The findings are in a stable order.
func Check(c *node.Corpus, ds Decls) []Finding {
	var out []Finding
	names := make([]string, 0, len(ds))
	for n := range ds {
		names = append(names, n)
	}
	slices.Sort(names)
	for _, n := range names {
		d := ds[n]
		if bad := d.Disallowed(); len(bad) > 0 {
			msg := "uses the axiom " + strings.Join(bad, ", ")
			if d.Sorry {
				msg += " (a sorry)"
			}
			out = append(out, Finding{"L03", d.Decl, msg})
		}
	}
	if c == nil {
		return out
	}
	for _, nd := range c.Nodes {
		out = append(out, CheckNode(nd, ds)...)
		out = append(out, CheckClaims(nd, ds)...)
	}
	return out
}

// CheckNode checks the names of the lean field of one node (rule L01).
func CheckNode(n *node.Node, ds Decls) []Finding {
	var out []Finding
	mod := n.Lean["module"]
	for _, k := range slices.Sorted(maps.Keys(n.Lean)) {
		v := n.Lean[k]
		if k == "module" || k == "cslib" || v == "" {
			continue
		}
		where := n.Slug + " lean." + k
		d, ok := ds[v]
		if !ok {
			out = append(out, Finding{"L01", where, v + " is not a declaration of the Algo package"})
			continue
		}
		if want := Kinds[k]; want != nil && !slices.Contains(want, d.Kind) {
			out = append(out, Finding{"L01", where, fmt.Sprintf("%s is a %s, want a %s", v, d.Kind, strings.Join(want, " or "))})
		}
		if mod != "" && k != "spec" && d.Module != mod {
			out = append(out, Finding{"L01", where, fmt.Sprintf("%s is in the module %s, not in %s", v, d.Module, mod)})
		}
	}
	return out
}

// classOps are the relations of Mathlib's asymptotics, as the pretty printer shows them.
var classOps = []string{"=O[", "=Θ[", "=o["}

// IsClassLemma tells if a type states an asymptotic class, such as `f =O[Filter.atTop] g`.
func IsClassLemma(typ string) bool {
	for _, op := range classOps {
		if strings.Contains(typ, op) {
			return true
		}
	}
	return false
}

var (
	funRE   = regexp.MustCompile(`^fun \(?(\w+)(?: : [^)]*)?\)? => `)
	claimRE = regexp.MustCompile(`^[ΘOΩo]\((.*)\)$`)
	ascRE   = regexp.MustCompile(`\((\w+) : [^()]*\)`)
	dropRE  = regexp.MustCompile(`↑|Real\.|Nat\.|\\|[\s*·()]`)
)

// ClassOf returns the class of a class lemma in a normal form, such as "nlogn" for `f =O[Filter.atTop] fun n => ↑n * Real.log ↑n`. It returns "" when typ is not a class lemma.
func ClassOf(typ string) string {
	for _, op := range classOps {
		i := strings.Index(typ, op)
		if i < 0 {
			continue
		}
		rest := typ[i+len(op):]
		j := strings.Index(rest, "] ")
		if j < 0 {
			return ""
		}
		g := strings.TrimSpace(rest[j+2:])
		g = funRE.ReplaceAllString(g, "")
		return normClass(g)
	}
	return ""
}

// ClaimClass returns the class of a front matter bound in the same normal form, such as "nlogn" for "Θ(n log n)". It returns "" when the bound is not of the form Θ(...), O(...) or Ω(...).
func ClaimClass(bound string) string {
	m := claimRE.FindStringSubmatch(strings.TrimSpace(bound))
	if m == nil {
		return ""
	}
	return normClass(m[1])
}

func normClass(s string) string {
	s = strings.ReplaceAll(s, "lg", "log")
	s = ascRE.ReplaceAllString(s, "$1")
	return dropRE.ReplaceAllString(s, "")
}

// CheckClaims compares the time claims of a node that has lean.cost with the class lemmas of its module (rule L05). A claim with no class lemma of the same class is a finding. Space claims count only when lean.space is set.
func CheckClaims(n *node.Node, ds Decls) []Finding {
	mod := n.Lean["module"]
	if mod == "" || n.Complexity == nil {
		return nil
	}
	have := map[string]bool{}
	for _, d := range ds {
		if d.Module == mod && d.Kind == "theorem" && len(d.Disallowed()) == 0 {
			if c := ClassOf(d.Type); c != "" {
				have[c] = true
			}
		}
	}
	var out []Finding
	check := func(group string, m map[string]string) {
		for _, k := range slices.Sorted(maps.Keys(m)) {
			c := ClaimClass(m[k])
			if c == "" || have[c] {
				continue
			}
			out = append(out, Finding{"L05", n.Slug + " " + group + "." + k, fmt.Sprintf("%s has no class lemma in %s", m[k], mod)})
		}
	}
	if n.Lean["cost"] != "" {
		check("time", n.Complexity.Time)
	}
	if n.Lean["space"] != "" {
		check("space", n.Complexity.Space)
	}
	return out
}
