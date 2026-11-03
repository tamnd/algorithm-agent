// Package lang runs the native tests of the implementations (spec 09.9).
//
// Each language has a recipe. A recipe copies the files of one implementation to a new temporary directory, writes the build files that the language needs, and runs the test command there. The recipe does not change the content tree.
//
// The tests run code from the content repository. Run them only in a sandbox, such as a CI runner or a container. The CLI asks for -unsafe-local outside CI.
package lang

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/tamnd/algorithm-agent/node"
	"github.com/tamnd/algorithm-agent/status"
)

// Timeout is the longest time that the tests of one implementation can take.
var Timeout = 3 * time.Minute

// JUnitVersion is the version of the JUnit console launcher for Java and Kotlin.
const JUnitVersion = "1.13.4"

// Recipe runs the tests of one language.
type Recipe struct {
	ID string
	// Tools are the commands that the recipe needs. If one is missing, the language is skipped.
	Tools []string
	// Version prints the version of the toolchain.
	Version []string
	// Run prepares dir, which holds a copy of the files, and returns the commands to run in it.
	Run func(r *Runner, dir string, n *node.Node, im node.Impl) ([][]string, error)
}

// Recipes are the tier 1 languages (spec 08.2).
var Recipes = map[string]Recipe{
	"python":     {ID: "python", Tools: []string{"python3"}, Version: []string{"python3", "--version"}, Run: python},
	"go":         {ID: "go", Tools: []string{"go"}, Version: []string{"go", "version"}, Run: golang},
	"rust":       {ID: "rust", Tools: []string{"cargo"}, Version: []string{"rustc", "--version"}, Run: rust},
	"c":          {ID: "c", Tools: []string{"cc"}, Version: []string{"cc", "--version"}, Run: clang},
	"cpp":        {ID: "cpp", Tools: []string{"c++"}, Version: []string{"c++", "--version"}, Run: cpp},
	"javascript": {ID: "javascript", Tools: []string{"node"}, Version: []string{"node", "--version"}, Run: javascript},
	"typescript": {ID: "typescript", Tools: []string{"node"}, Version: []string{"node", "--version"}, Run: typescript},
	"java":       {ID: "java", Tools: []string{"javac", "java"}, Version: []string{"javac", "-version"}, Run: java},
	"kotlin":     {ID: "kotlin", Tools: []string{"kotlinc", "java"}, Version: []string{"kotlinc", "-version"}, Run: kotlin},
	"csharp":     {ID: "csharp", Tools: []string{"dotnet"}, Version: []string{"dotnet", "--version"}, Run: csharp},
}

// Runner runs tests and keeps the results.
type Runner struct {
	// Cache holds the downloaded jars. The default is the user cache directory.
	Cache string
	// Log gets one line for each implementation.
	Log func(format string, args ...any)
	// Keep keeps the temporary directories, for debugging.
	Keep bool
}

// Skip is the error of a language whose tools are missing.
type Skip struct{ Lang, Tool string }

func (s Skip) Error() string { return fmt.Sprintf("%s: the tool %s is not installed", s.Lang, s.Tool) }

// Available tells if the tools of a recipe are installed.
func (rc Recipe) Available() error {
	for _, t := range rc.Tools {
		if _, err := exec.LookPath(t); err != nil {
			return Skip{rc.ID, t}
		}
	}
	// A tool can be a stub that only says how to install the real one, as /usr/bin/javac on macOS is.
	if len(rc.Version) > 0 {
		if err := exec.Command(rc.Version[0], rc.Version[1:]...).Run(); err != nil {
			return Skip{rc.ID, rc.Version[0]}
		}
	}
	return nil
}

// ToolVersion returns the first line of the version output of the toolchain.
func (rc Recipe) ToolVersion() string {
	if len(rc.Version) == 0 {
		return ""
	}
	out, err := exec.Command(rc.Version[0], rc.Version[1:]...).CombinedOutput()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(strings.SplitN(string(out), "\n", 2)[0])
}

// Test runs the native tests of one implementation.
func (r *Runner) Test(n *node.Node, im node.Impl) status.Test {
	t := status.Test{Slug: n.Slug, Lang: im.Lang}
	rc, ok := Recipes[im.Lang]
	if !ok {
		t.Output = "no recipe for the language " + im.Lang
		return t
	}
	if !im.HasTests() {
		t.Output = "the implementation has no test file"
		return t
	}
	start := time.Now()
	dir, err := os.MkdirTemp("", "algo-test-"+im.Lang+"-")
	if err != nil {
		t.Output = err.Error()
		return t
	}
	if !r.Keep {
		defer os.RemoveAll(dir)
	}
	if err := stage(dir, im, true); err != nil {
		t.Output = err.Error()
		return t
	}
	cmds, err := rc.Run(r, dir, n, im)
	if err != nil {
		t.Output = err.Error()
		t.Ms = time.Since(start).Milliseconds()
		return t
	}
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()
	var out bytes.Buffer
	t.Pass = true
	for _, c := range cmds {
		cmd := exec.CommandContext(ctx, c[0], c[1:]...)
		cmd.Dir = dir
		cmd.Stdout, cmd.Stderr = &out, &out
		cmd.Env = append(os.Environ(), testEnv...)
		if err := cmd.Run(); err != nil {
			t.Pass = false
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				fmt.Fprintf(&out, "\nthe tests took more than %s\n", Timeout)
			}
			break
		}
	}
	t.Ms = time.Since(start).Milliseconds()
	t.Output = tail(out.String(), 4000)
	return t
}

func tail(s string, k int) string {
	s = strings.TrimSpace(s)
	if len(s) <= k {
		return s
	}
	return "...\n" + s[len(s)-k:]
}

func files(im node.Impl, test bool, ext ...string) []string {
	var out []string
	for _, f := range im.Files {
		if f.Test != test {
			continue
		}
		for _, e := range ext {
			if strings.HasSuffix(f.Name, e) {
				out = append(out, f.Name)
			}
		}
	}
	sort.Strings(out)
	return out
}

func all(im node.Impl, ext ...string) []string {
	return append(files(im, false, ext...), files(im, true, ext...)...)
}

func write(dir, name, content string) error {
	return os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)
}

func python(_ *Runner, dir string, _ *node.Node, _ node.Impl) ([][]string, error) {
	if err := exec.Command("python3", "-c", "import pytest").Run(); err != nil {
		return nil, Skip{"python", "pytest"}
	}
	return [][]string{{"python3", "-m", "pytest", "-q", "-p", "no:cacheprovider", "."}}, nil
}

func golang(_ *Runner, dir string, _ *node.Node, _ node.Impl) ([][]string, error) {
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err != nil {
		if err := write(dir, "go.mod", "module example.com/impl\n\ngo 1.24\n"); err != nil {
			return nil, err
		}
	}
	return [][]string{{"go", "vet", "./..."}, {"go", "test", "-count=1", "./..."}}, nil
}

var nonIdent = regexp.MustCompile(`[^a-z0-9_]+`)

func crate(slug string) string {
	return "algo_" + nonIdent.ReplaceAllString(strings.ReplaceAll(slug, "-", "_"), "")
}

func rust(_ *Runner, dir string, n *node.Node, im node.Impl) ([][]string, error) {
	main := files(im, false, ".rs")
	if len(main) == 0 {
		return nil, errors.New("rust: no source file")
	}
	// A test file is an integration test. It goes to tests/ and imports the crate.
	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0o755); err != nil {
		return nil, err
	}
	for _, t := range files(im, true, ".rs") {
		if err := os.Rename(filepath.Join(dir, t), filepath.Join(dir, "tests", filepath.Base(t))); err != nil {
			return nil, err
		}
	}
	toml := fmt.Sprintf("[package]\nname = %q\nversion = \"0.0.0\"\nedition = \"2024\"\npublish = false\n\n[lib]\npath = %q\n\n[lints.rust]\nwarnings = \"deny\"\n", crate(n.Slug), main[0])
	if err := write(dir, "Cargo.toml", toml); err != nil {
		return nil, err
	}
	return [][]string{{"cargo", "test", "--offline", "--quiet"}}, nil
}

var cFlags = []string{"-Wall", "-Wextra", "-Wpedantic", "-Werror", "-g", "-fsanitize=address,undefined", "-fno-omit-frame-pointer"}

func clang(_ *Runner, dir string, _ *node.Node, im node.Impl) ([][]string, error) {
	src := all(im, ".c")
	args := append([]string{"cc", "-std=c2x"}, cFlags...)
	args = append(append(args, src...), "-o", "test_main")
	return [][]string{args, {"./test_main"}}, nil
}

func cpp(_ *Runner, dir string, _ *node.Node, im node.Impl) ([][]string, error) {
	src := all(im, ".cpp", ".cc")
	args := append([]string{"c++", "-std=c++23"}, cFlags...)
	args = append(append(args, src...), "-o", "test_main")
	return [][]string{args, {"./test_main"}}, nil
}

func javascript(_ *Runner, dir string, _ *node.Node, im node.Impl) ([][]string, error) {
	if err := write(dir, "package.json", "{\"type\": \"module\", \"private\": true}\n"); err != nil {
		return nil, err
	}
	return [][]string{append([]string{"node", "--test"}, files(im, true, ".js", ".mjs")...)}, nil
}

func typescript(_ *Runner, dir string, _ *node.Node, im node.Impl) ([][]string, error) {
	if err := write(dir, "package.json", "{\"type\": \"module\", \"private\": true}\n"); err != nil {
		return nil, err
	}
	// Node 24 and later run TypeScript by type stripping. The code uses only erasable syntax.
	return [][]string{append([]string{"node", "--test"}, files(im, true, ".ts")...)}, nil
}

// junit returns the path of the JUnit console launcher, and downloads it once.
func (r *Runner) junit() (string, error) {
	cache := r.Cache
	if cache == "" {
		c, err := os.UserCacheDir()
		if err != nil {
			return "", err
		}
		cache = filepath.Join(c, "algo")
	}
	name := "junit-platform-console-standalone-" + JUnitVersion + ".jar"
	path := filepath.Join(cache, name)
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}
	if err := os.MkdirAll(cache, 0o755); err != nil {
		return "", err
	}
	url := "https://repo1.maven.org/maven2/org/junit/platform/junit-platform-console-standalone/" + JUnitVersion + "/" + name
	resp, err := http.Get(url) //nolint:noctx // the URL is a constant
	if err != nil {
		return "", fmt.Errorf("download JUnit: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download JUnit: %s", resp.Status)
	}
	tmp := path + ".part"
	f, err := os.Create(tmp)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		_ = f.Close()
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return path, os.Rename(tmp, path)
}

func java(r *Runner, dir string, _ *node.Node, im node.Impl) ([][]string, error) {
	jar, err := r.junit()
	if err != nil {
		return nil, err
	}
	args := append([]string{"javac", "-Xlint:all", "-Werror", "-d", "out", "-cp", jar}, all(im, ".java")...)
	run := []string{"java", "-jar", jar, "execute", "--class-path", "out", "--scan-class-path", "--disable-banner", "--details=summary", "--fail-if-no-tests"}
	return [][]string{args, run}, nil
}

// kotlinStdlib returns the path of kotlin-stdlib.jar of the kotlinc on the PATH.
func kotlinStdlib() (string, error) {
	kc, err := exec.LookPath("kotlinc")
	if err != nil {
		return "", err
	}
	if real, err := filepath.EvalSymlinks(kc); err == nil {
		kc = real
	}
	stdlib := filepath.Join(filepath.Dir(filepath.Dir(kc)), "lib", "kotlin-stdlib.jar")
	if _, err := os.Stat(stdlib); err != nil {
		return "", fmt.Errorf("kotlin: no kotlin-stdlib.jar next to kotlinc: %w", err)
	}
	return stdlib, nil
}

func kotlin(r *Runner, dir string, _ *node.Node, im node.Impl) ([][]string, error) {
	jar, err := r.junit()
	if err != nil {
		return nil, err
	}
	stdlib, err := kotlinStdlib()
	if err != nil {
		return nil, err
	}
	args := append([]string{"kotlinc", "-Werror", "-cp", jar, "-d", "out"}, all(im, ".kt")...)
	cp := "out" + string(os.PathListSeparator) + stdlib
	run := []string{"java", "-jar", jar, "execute", "--class-path", cp, "--scan-class-path", "--disable-banner", "--details=summary", "--fail-if-no-tests"}
	return [][]string{args, run}, nil
}

const csproj = `<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <TargetFramework>net%s</TargetFramework>
    <Nullable>enable</Nullable>
    <ImplicitUsings>enable</ImplicitUsings>
    <TreatWarningsAsErrors>true</TreatWarningsAsErrors>
    <IsPackable>false</IsPackable>
  </PropertyGroup>
  <ItemGroup>
    <PackageReference Include="Microsoft.NET.Test.Sdk" Version="17.14.1" />
    <PackageReference Include="xunit" Version="2.9.3" />
    <PackageReference Include="xunit.runner.visualstudio" Version="3.1.4" />
  </ItemGroup>
</Project>
`

// dotnetMajor returns the major version of the .NET SDK, such as 10.
func dotnetMajor() (string, error) {
	out, err := exec.Command("dotnet", "--version").Output()
	if err != nil {
		return "", err
	}
	major, _, _ := strings.Cut(strings.TrimSpace(string(out)), ".")
	return major, nil
}

func csharp(_ *Runner, dir string, _ *node.Node, _ node.Impl) ([][]string, error) {
	major, err := dotnetMajor()
	if err != nil {
		return nil, err
	}
	if err := write(dir, "Impl.Tests.csproj", fmt.Sprintf(csproj, major+".0")); err != nil {
		return nil, err
	}
	return [][]string{{"dotnet", "test", "--nologo", "-v", "q"}}, nil
}
