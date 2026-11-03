// SPDX-License-Identifier: Apache-2.0

package status

import (
	"testing"

	"github.com/tamnd/algorithm-agent/node"
)

func TestImplLadder(t *testing.T) {
	n := &node.Node{Slug: "merge-sort"}
	im := node.Impl{Lang: "c", Files: []node.File{{Name: "merge_sort.c"}, {Name: "test_merge_sort.c", Test: true}}}
	tests := []struct {
		name  string
		test  Test
		level int
	}{
		{"all pass", Test{Pass: true, Shared: []Shared{{Problem: "sorting-problem", Cases: 18, Passed: 18}}}, 1},
		{"expected skip", Test{Pass: true, Shared: []Shared{{Problem: "sorting-problem", Cases: 18, Passed: 13, Skipped: 5, SkipExpected: true}}}, 1},
		{"skip", Test{Pass: true, Shared: []Shared{{Problem: "sorting-problem", Cases: 18, Passed: 13, Skipped: 5, SkipReason: "no build"}}}, 0},
		{"failure", Test{Pass: true, Shared: []Shared{{Problem: "sorting-problem", Cases: 18, Passed: 17, Failures: []CaseFailure{{Case: "basic/example"}}}}}, 0},
		{"native fail", Test{Pass: false, Shared: []Shared{{Problem: "sorting-problem", Cases: 18, Passed: 18}}}, 0},
		{"no shared", Test{Pass: true}, 0},
	}
	for _, tc := range tests {
		env := Env{Results: map[string]map[string]Test{"merge-sort": {"c": tc.test}}}
		r := impl(n, im, env)
		if r.Level != tc.level {
			t.Errorf("%s: level %d, want %d (next %v)", tc.name, r.Level, tc.level, r.Next)
		}
	}
}
