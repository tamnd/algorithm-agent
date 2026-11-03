// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/tamnd/algorithm-agent/contract"
	"github.com/tamnd/algorithm-agent/lang"
	"github.com/tamnd/algorithm-agent/node"
)

type diffFlags struct {
	classes, seeds string
	write          bool
	root           string
	box            *lang.Sandbox
}

// problemInputs are the generated inputs of one problem, made once for all the implementations, or the reason why the problem has none.
type problemInputs struct {
	in   *lang.Inputs
	skip string
	err  error
}

// runDiff runs the differential test of each implementation against the oracle of each problem that it solves, on the generated inputs (spec 09.5). It shrinks each failure, writes the regression cases to build/regress/<problem>.jsonl, and with -write-regress also to cases/regress.jsonl of the problem.
func runDiff(r *lang.Runner, nodes []*node.Node, want map[string]bool, problems func(*node.Node) []*sharedCases, d diffFlags) error {
	classes := strings.Split(d.classes, ",")
	for _, cl := range classes {
		if !contains(contract.ClassOrder, cl) {
			return usageError{fmt.Sprintf("the class %q is not one of %s", cl, strings.Join(contract.ClassOrder, ", "))}
		}
	}
	var seeds []int
	for _, s := range strings.Split(d.seeds, ",") {
		if s = strings.TrimSpace(s); s == "" {
			continue
		}
		n, err := strconv.Atoi(s)
		if err != nil {
			return usageError{"-seeds is a comma list of integers"}
		}
		seeds = append(seeds, n)
	}
	inputs := map[string]*problemInputs{}
	var closers []func()
	defer func() {
		for _, c := range closers {
			c()
		}
	}()
	get := func(k *contract.Contract) *problemInputs {
		if pi, ok := inputs[k.Problem]; ok {
			return pi
		}
		pi := &problemInputs{}
		inputs[k.Problem] = pi
		g, err := k.LoadGenerated()
		if err != nil {
			pi.err = err
			return pi
		}
		if g == nil {
			pi.skip = "the problem has no cases/generated.yaml"
			return pi
		}
		if d.box != nil {
			if err := d.box.Ready("python"); err != nil {
				pi.err = fmt.Errorf("the generators and the oracle run in the Python image: %w", err)
				return pi
			}
		}
		o, err := k.StartOracle(lang.CheckerWrap(d.box, k))
		if errors.Is(err, contract.ErrNoOracle) {
			pi.skip = err.Error()
			return pi
		}
		if err != nil {
			pi.err = err
			return pi
		}
		closers = append(closers, o.Close)
		gen, err := k.StartGenerator(lang.CheckerWrap(d.box, k))
		if err != nil {
			pi.err = err
			return pi
		}
		closers = append(closers, gen.Close)
		ss := seeds
		if len(ss) == 0 {
			ss = g.Seeds.PR
		}
		start := time.Now()
		pi.in, pi.err = lang.MakeInputs(k, g, gen, o, g.Plan(k, classes, ss))
		if pi.err == nil {
			fmt.Printf("%s: %d generated inputs with the oracle %s (%d ms)\n", k.Problem, len(pi.in.Cases), o.File, time.Since(start).Milliseconds())
		}
		return pi
	}
	failed := 0
	regress := map[string][]contract.Case{}
	dirs := map[string]string{}
	for _, n := range nodes {
		for _, im := range n.Impls {
			if len(want) > 0 && !want[im.Lang] {
				continue
			}
			rc, ok := lang.Recipes[im.Lang]
			if !ok {
				continue
			}
			if err := r.Available(rc); err != nil {
				var skip lang.Skip
				if errors.As(err, &skip) {
					fmt.Fprintf(os.Stderr, "skip %s/%s: %v\n", n.Slug, im.Lang, err)
					continue
				}
				return err
			}
			for _, p := range problems(n) {
				if p.err != nil {
					return fmt.Errorf("the contract of a problem that %s solves: %w", n.Slug, p.err)
				}
				pi := get(p.k)
				if pi.err != nil {
					return fmt.Errorf("%s: %w", p.k.Problem, pi.err)
				}
				if pi.skip != "" {
					fmt.Printf("skip %s/%s differential run of %s: %s\n", n.Slug, im.Lang, p.k.Problem, pi.skip)
					continue
				}
				ss, err := r.Open(n, im, p.k)
				if err != nil {
					fmt.Printf("FAIL %s/%s differential run of %s: %v\n", n.Slug, im.Lang, p.k.Problem, err)
					failed++
					continue
				}
				res, err := lang.Diff(ss, pi.in)
				ss.Close()
				if err != nil {
					return fmt.Errorf("%s/%s: %w", n.Slug, im.Lang, err)
				}
				state := "pass"
				if len(res.Failures) > 0 {
					state = "FAIL"
					failed++
				}
				line := fmt.Sprintf("%-4s %s/%s differential run of %s: %d of %d pass", state, n.Slug, im.Lang, p.k.Problem, res.Passed, res.Cases)
				if len(res.Failures) > 0 && res.Passed+len(res.Failures) < res.Cases {
					line = fmt.Sprintf("%-4s %s/%s differential run of %s: %d failures, the run stopped after %d of %d cases", state, n.Slug, im.Lang, p.k.Problem, len(res.Failures), res.Passed+len(res.Failures), res.Cases)
				}
				if res.Skipped > 0 {
					line += fmt.Sprintf(", %d skipped", res.Skipped)
				}
				fmt.Printf("%s (%d ms)\n", line, res.Took.Milliseconds())
				for _, f := range res.Failures {
					reg := lang.Regression(p.k, f, n.Slug+"/"+im.Lang, time.Now().UTC())
					fmt.Println(indent(fmt.Sprintf("case %s: %s\n  got:  %s\n  want: %s\nshrunk in %d runs to %s: %s\n  args: %s\n  got:  %s\n  want: %s", f.Case.Name, f.Reason, short(f.Got), short(f.Case.Expect), f.Calls, reg.Name, f.ShrunkReason, short(argsJSON(reg.Args)), short(f.ShrunkGot), short(reg.Expect))))
					regress[p.k.Problem] = append(regress[p.k.Problem], reg)
					dirs[p.k.Problem] = p.k.Dir
				}
			}
		}
	}
	for prob, cs := range regress {
		out := filepath.Join(d.root, "build", "regress", prob+".jsonl")
		if err := appendCases(out, cs); err != nil {
			return err
		}
		fmt.Printf("wrote %d regression cases to %s\n", len(cs), out)
		if d.write {
			path := filepath.Join(dirs[prob], "cases", "regress.jsonl")
			if err := appendCases(path, cs); err != nil {
				return err
			}
			fmt.Printf("added %d regression cases to %s\n", len(cs), path)
		}
	}
	if failed > 0 {
		return errFindings
	}
	return nil
}

func contains(xs []string, x string) bool {
	for _, y := range xs {
		if y == x {
			return true
		}
	}
	return false
}

func short(b []byte) string {
	if len(b) == 0 {
		return "(none)"
	}
	var buf bytes.Buffer
	if json.Compact(&buf, b) == nil {
		b = buf.Bytes()
	}
	return contract.Short(string(b))
}

func argsJSON(a map[string]json.RawMessage) []byte {
	b, _ := json.Marshal(a)
	return b
}

// appendCases adds the cases to a case file, and skips a case whose name is in the file already.
func appendCases(path string, cs []contract.Case) error {
	old, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var b bytes.Buffer
	b.Write(old)
	if len(old) > 0 && !bytes.HasSuffix(old, []byte("\n")) {
		b.WriteByte('\n')
	}
	for _, c := range cs {
		if bytes.Contains(old, []byte(`"name":"`+c.Name+`"`)) {
			continue
		}
		line, err := json.Marshal(c)
		if err != nil {
			return err
		}
		b.Write(line)
		b.WriteByte('\n')
	}
	return os.WriteFile(path, b.Bytes(), 0o644)
}
