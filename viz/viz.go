// Package viz bundles the scenes of the nodes (spec 10) and checks their traces.
//
// A scene is viz/scene.ts. It imports only "@algo/viz". Build bundles it with esbuild into a script that registers the scene in window.AlgoScenes. The player of the site runs it in the browser.
package viz

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

//go:embed packages/core/index.ts
var core string

// Core is the source of @algo/viz.
func Core() string { return core }

// Build bundles the scene file into a browser script. The script sets window.AlgoScenes[slug] to the scene, and window.AlgoViz to the runtime.
func Build(scenePath string, minify bool) ([]byte, error) {
	abs, err := filepath.Abs(scenePath)
	if err != nil {
		return nil, err
	}
	entry := `import s from ` + quote(abs) + `;
import * as viz from "@algo/viz";
const g = globalThis as any;
g.AlgoViz = viz;
(g.AlgoScenes ||= {})[s.slug] = s;
`
	res := api.Build(api.BuildOptions{
		Stdin: &api.StdinOptions{
			Contents:   entry,
			ResolveDir: filepath.Dir(abs),
			Sourcefile: "entry.ts",
			Loader:     api.LoaderTS,
		},
		Bundle:            true,
		Format:            api.FormatIIFE,
		Target:            api.ES2020,
		MinifyWhitespace:  minify,
		MinifySyntax:      minify,
		MinifyIdentifiers: minify,
		LegalComments:     api.LegalCommentsNone,
		Write:             false,
		Plugins:           []api.Plugin{onlyViz(abs)},
	})
	if len(res.Errors) > 0 {
		var msgs []string
		for _, m := range res.Errors {
			loc := ""
			if m.Location != nil {
				loc = fmt.Sprintf("%s:%d: ", filepath.Base(m.Location.File), m.Location.Line)
			}
			msgs = append(msgs, loc+m.Text)
		}
		return nil, errors.New(strings.Join(msgs, "; "))
	}
	if len(res.OutputFiles) == 0 {
		return nil, errors.New("esbuild wrote no output")
	}
	return res.OutputFiles[0].Contents, nil
}

func quote(s string) string {
	b, _ := json.Marshal(filepath.ToSlash(s))
	return string(b)
}

// onlyViz resolves "@algo/viz" to the embedded runtime, and rejects every other import of the scene (rule V01).
func onlyViz(scene string) api.Plugin {
	return api.Plugin{
		Name: "algo-viz",
		Setup: func(b api.PluginBuild) {
			b.OnResolve(api.OnResolveOptions{Filter: `.*`}, func(a api.OnResolveArgs) (api.OnResolveResult, error) {
				switch {
				case a.Path == "@algo/viz":
					return api.OnResolveResult{Path: "algo-viz", Namespace: "algo-viz"}, nil
				case a.Path == filepath.ToSlash(scene) || a.Path == scene:
					return api.OnResolveResult{Path: scene}, nil
				default:
					return api.OnResolveResult{}, fmt.Errorf("the scene imports %q; a scene imports only @algo/viz (rule V01)", a.Path)
				}
			})
			b.OnLoad(api.OnLoadOptions{Filter: `.*`, Namespace: "algo-viz"}, func(api.OnLoadArgs) (api.OnLoadResult, error) {
				return api.OnLoadResult{Contents: &core, Loader: api.LoaderTS}, nil
			})
		},
	}
}

// Preset is a named input of a scene.
type Preset struct {
	Name  string `json:"name"`
	Input any    `json:"input"`
}

// Report is the result of checking one scene.
type Report struct {
	Slug    string   `json:"slug"`
	Fails   []string `json:"fails"`
	Lines   []string `json:"lines"` // the line ids that the presets reach
	Steps   []int    `json:"steps"` // the number of steps of each preset
	Skipped string   `json:"skipped,omitempty"`
}

//go:embed check.js
var checkJS string

// Check runs the bundled scene in Node on each preset (rules V01, V02, V04, V05, V10, and the state vocabulary of V09). It also checks that each event names a view of its kind. ids are the line ids of the pseudocode of the page. Without Node, the check is skipped and says so.
func Check(slug string, bundle []byte, presets []Preset, ids []string) Report {
	r := Report{Slug: slug}
	nodeBin, err := exec.LookPath("node")
	if err != nil {
		r.Skipped = "node is not installed"
		return r
	}
	if len(presets) < 3 {
		r.Fails = append(r.Fails, fmt.Sprintf("the scene has %d of the 3 presets that it needs (V10)", len(presets)))
	}
	dir, err := os.MkdirTemp("", "algo-viz-")
	if err != nil {
		r.Fails = append(r.Fails, err.Error())
		return r
	}
	defer os.RemoveAll(dir)
	pj, _ := json.Marshal(presets)
	files := map[string][]byte{"scene.js": bundle, "check.js": []byte(checkJS), "presets.json": pj}
	for name, b := range files {
		if err := os.WriteFile(filepath.Join(dir, name), b, 0o600); err != nil {
			r.Fails = append(r.Fails, err.Error())
			return r
		}
	}
	cmd := exec.Command(nodeBin, "check.js", slug)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		msg := err.Error()
		var ee *exec.ExitError
		if errors.As(err, &ee) && len(ee.Stderr) > 0 {
			msg = strings.TrimSpace(string(ee.Stderr))
		}
		r.Fails = append(r.Fails, "the scene does not run: "+msg)
		return r
	}
	var res struct {
		Presets []struct {
			Name    string   `json:"name"`
			Error   string   `json:"error"`
			Same    bool     `json:"same"`
			Lines   []string `json:"lines"`
			Steps   int      `json:"steps"`
			Invalid []string `json:"invalid"`
		} `json:"presets"`
	}
	if err := json.Unmarshal(out, &res); err != nil {
		r.Fails = append(r.Fails, "the check output does not parse: "+err.Error())
		return r
	}
	reached := map[string]bool{}
	for _, p := range res.Presets {
		r.Steps = append(r.Steps, p.Steps)
		if p.Error != "" {
			r.Fails = append(r.Fails, fmt.Sprintf("preset %s: %s", p.Name, p.Error))
			continue
		}
		for _, m := range p.Invalid {
			r.Fails = append(r.Fails, fmt.Sprintf("preset %s: %s", p.Name, m))
		}
		if !p.Same {
			r.Fails = append(r.Fails, fmt.Sprintf("preset %s: two runs gave different traces (V02)", p.Name))
		}
		for _, l := range p.Lines {
			if !slices.Contains(ids, l) {
				r.Fails = append(r.Fails, fmt.Sprintf("preset %s: the line %s is not in the pseudocode (V04)", p.Name, l))
			}
			reached[l] = true
		}
	}
	for l := range reached {
		r.Lines = append(r.Lines, l)
	}
	slices.Sort(r.Lines)
	var miss []string
	for _, id := range ids {
		if !reached[id] {
			miss = append(miss, id)
		}
	}
	if len(miss) > 0 {
		r.Fails = append(r.Fails, "no preset reaches "+strings.Join(miss, ", ")+" (V05)")
	}
	return r
}
