// SPDX-License-Identifier: Apache-2.0

package node

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsBuildFile(t *testing.T) {
	for name, want := range map[string]bool{"go.mod": true, "x/Cargo.toml": true, "Impl.csproj": true, "merge_sort.go": false, "package.json": true, "README.md": false} {
		if got := IsBuildFile(name); got != want {
			t.Errorf("IsBuildFile(%q) = %v", name, got)
		}
	}
}

func TestReadKeepsTagLikeNumber(t *testing.T) {
	dir := t.TempDir()
	front := "---\ntag: 860E65\nkind: algorithm\nslug: binary-search\n---\n"
	if err := os.WriteFile(filepath.Join(dir, "node.md"), []byte(front), 0o644); err != nil {
		t.Fatal(err)
	}
	n, err := Read(dir, dir)
	if err != nil {
		t.Fatal(err)
	}
	if n.Front.Tag != "860E65" {
		t.Errorf("tag is %q, want 860E65", n.Front.Tag)
	}
}
