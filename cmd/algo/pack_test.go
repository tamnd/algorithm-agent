package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTree(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, body := range files {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestPackSite(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"index.html":                       "home",
		"search.json":                      "{}",
		"llms.txt":                         "llms",
		"_headers":                         "/*\n  X-Content-Type-Options: nosniff\n\n/*.md\n  Content-Type: text/markdown; charset=utf-8\n",
		"assets/site.css":                  "css",
		"kind/algorithm/index.html":        "kind",
		"embed/merge-sort/index.html":      "embed",
		"merge-sort/index.html":            "main page",
		"merge-sort.md":                    "# merge sort",
		"merge-sort.json":                  `{"slug":"merge-sort"}`,
		"merge-sort/python.txt":            "def merge_sort(): pass",
		"merge-sort/python/index.html":     "python page",
		"merge-sort/proof/index.html":      "proof page",
		"merge-sort/visualize/index.html":  "visualize page",
		"merge-sort/other/index.html":      "not packed",
		"heap/index.html":                  "heap page",
		"heap.md":                          strings.Repeat("h", 1000),
		"heap/rust.txt":                    "fn heap() {}",
		"heap/rust/index.html":             "rust page",
		"big/index.html":                   "big page",
		"big.md":                           strings.Repeat("b", maxPacked+1),
	}
	writeTree(t, dir, files)
	n, err := packSite(dir, "test")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"embed/merge-sort/index.html", "merge-sort.md", "merge-sort.json", "merge-sort/python.txt", "merge-sort/python/index.html", "merge-sort/proof/index.html", "merge-sort/visualize/index.html", "heap.md", "heap/rust.txt", "heap/rust/index.html"}
	if n != len(want) {
		t.Errorf("packed %d files, want %d", n, len(want))
	}
	// Each packed file is gone from the tree, and its index entry gives back its bytes.
	got := map[string]string{}
	packs, _ := filepath.Glob(filepath.Join(dir, packDir, "*.json"))
	for _, ix := range packs {
		var index map[string][2]int64
		b, err := os.ReadFile(ix)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(b, &index); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(strings.TrimSuffix(ix, ".json") + ".pack")
		if err != nil {
			t.Fatal(err)
		}
		i := strings.TrimSuffix(filepath.Base(ix), ".json")
		for p, e := range index {
			if fmt.Sprint(packOf(p, len(packs))) != i {
				t.Errorf("%s is in pack %s, want %d", p, i, packOf(p, len(packs)))
			}
			got[p] = string(data[e[0] : e[0]+e[1]])
		}
	}
	for _, p := range want {
		if got[p] != files[p] {
			t.Errorf("pack gives %q for %s, want %q", got[p], p, files[p])
		}
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(p))); err == nil {
			t.Errorf("%s is still a static file", p)
		}
	}
	for _, p := range []string{"index.html", "search.json", "llms.txt", "assets/site.css", "kind/algorithm/index.html", "merge-sort/index.html", "merge-sort/other/index.html", "big.md", "heap/index.html"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(p))); err != nil {
			t.Errorf("%s is not a static file: %v", p, err)
		}
	}
	w, err := os.ReadFile(filepath.Join(dir, "_worker.js"))
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"const packs = 1;", `const build = "test";`, "X-Content-Type-Options", "nosniff"} {
		if !strings.Contains(strings.ToLower(string(w)), strings.ToLower(s)) {
			t.Errorf("_worker.js does not have %s", s)
		}
	}
	var r pagesRoutes
	b, err := os.ReadFile(filepath.Join(dir, "_routes.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &r); err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"/*.md", "/*/python/*", "/*/rust/*", "/*/proof/*", "/embed/*"} {
		if !strings.Contains(string(b), `"`+s+`"`) {
			t.Errorf("_routes.json does not include %s", s)
		}
	}
	// A second pass leaves the tree as it is.
	if n, err := packSite(dir, "test"); err != nil || n != 0 {
		t.Errorf("second pass = %d, %v", n, err)
	}
}

func TestPackRoutesLimit(t *testing.T) {
	var subs []string
	for i := 0; i < 60; i++ {
		subs = append(subs, fmt.Sprint("l", i))
	}
	r := packRoutes(subs)
	if len(r.Include)+len(r.Exclude) > maxRoutes {
		t.Errorf("%d routes, the limit is %d", len(r.Include)+len(r.Exclude), maxRoutes)
	}
}
