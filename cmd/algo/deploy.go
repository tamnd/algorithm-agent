// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// The limits of Cloudflare Pages, with a margin: 20000 files in a deployment and 25 MiB for one file.
const (
	maxSiteFiles = 19500
	maxFileBytes = 24 << 20
)

// wranglerVersion is the version of wrangler that algo site deploy runs. A pinned version gives the same upload each time.
const wranglerVersion = "4.156.2"

type deployOptions struct {
	Dir     string
	Project string
	Branch  string
	Commit  string
	Preview bool
}

// checkLimits reports each limit of the host that the site tree breaks.
func checkLimits(dir string) (files int, problems []string, err error) {
	err = filepath.WalkDir(dir, func(p string, e fs.DirEntry, err error) error {
		if err != nil || e.IsDir() {
			return err
		}
		files++
		info, err := e.Info()
		if err != nil {
			return err
		}
		if info.Size() > maxFileBytes {
			rel, _ := filepath.Rel(dir, p)
			problems = append(problems, fmt.Sprintf("%s is %.1f MB, and the limit is %d MB", filepath.ToSlash(rel), float64(info.Size())/(1<<20), maxFileBytes>>20))
		}
		return nil
	})
	if files > maxSiteFiles {
		problems = append(problems, fmt.Sprintf("the site has %d files, and the limit is %d", files, maxSiteFiles))
	}
	return files, problems, err
}

var aliasRE = regexp.MustCompile(`[^a-z0-9]+`)

// alias returns the subdomain that Cloudflare Pages gives a branch: lower case, with each run of other characters changed to a hyphen.
func alias(branch string) string {
	a := strings.Trim(aliasRE.ReplaceAllString(strings.ToLower(branch), "-"), "-")
	if len(a) > 28 {
		a = strings.TrimRight(a[:28], "-")
	}
	return a
}

func gitBranch(root string) string {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = root
	if b, err := cmd.Output(); err == nil && len(bytes.TrimSpace(b)) > 0 {
		return strings.TrimSpace(string(b))
	}
	return "preview"
}

var pagesURL = regexp.MustCompile(`https://[a-z0-9.-]+\.pages\.dev`)

// deploy checks the limits of the host and uploads the site tree to Cloudflare Pages with wrangler. The account and the token come from CLOUDFLARE_ACCOUNT_ID and CLOUDFLARE_API_TOKEN. algo does not print them, and it passes them to wrangler only through the environment.
func deploy(o deployOptions) error {
	var missing []string
	for _, k := range []string{"CLOUDFLARE_API_TOKEN", "CLOUDFLARE_ACCOUNT_ID"} {
		if os.Getenv(k) == "" {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("deploy needs %s in the environment", strings.Join(missing, " and "))
	}
	if !o.Preview && o.Branch != "main" {
		return fmt.Errorf("a production deploy uses the branch main, not %s; use -preview", o.Branch)
	}
	if o.Preview && o.Branch == "main" {
		return errors.New("a preview needs a branch other than main")
	}
	build := o.Commit
	if build == "" {
		build = fmt.Sprint(time.Now().Unix())
	}
	n, err := packSite(o.Dir, build)
	if err != nil {
		return fmt.Errorf("pack the site: %w", err)
	}
	if n > 0 {
		fmt.Fprintf(os.Stderr, "packed %d files\n", n)
	}
	files, problems, err := checkLimits(o.Dir)
	if err != nil {
		return err
	}
	for _, p := range problems {
		fmt.Fprintln(os.Stderr, p)
	}
	if len(problems) > 0 {
		return errFindings
	}
	if _, err := exec.LookPath("npx"); err != nil {
		return fmt.Errorf("deploy runs wrangler with npx, and npx is not on PATH: %w", err)
	}
	argv := []string{"npx", "--yes", "wrangler@" + wranglerVersion, "pages", "deploy", o.Dir, "--project-name", o.Project, "--branch", o.Branch, "--commit-dirty=true"}
	if o.Commit != "" {
		argv = append(argv, "--commit-hash", o.Commit)
	}
	fmt.Fprintf(os.Stderr, "uploading %d files to the project %s, branch %s\n", files, o.Project, o.Branch)
	var log bytes.Buffer
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdout, cmd.Stderr = &log, &log
	cmd.Env = append(os.Environ(), "WRANGLER_SEND_METRICS=false")
	if err := cmd.Run(); err != nil {
		_, _ = io.Copy(os.Stderr, &log)
		return fmt.Errorf("wrangler pages deploy: %w", err)
	}
	// wrangler prints the URL of the deployment, and for a branch also the URL of its alias. The alias stays the same for each deploy of the branch, so it is the better link for a preview.
	urls := pagesURL.FindAllString(log.String(), -1)
	url := ""
	for _, u := range urls {
		if url == "" || strings.HasPrefix(u, "https://"+alias(o.Branch)+".") {
			url = u
		}
	}
	if url == "" {
		return errors.New("wrangler did not print the URL of the deployment")
	}
	fmt.Printf("deployed to %s\n", url)
	if out := os.Getenv("GITHUB_OUTPUT"); out != "" {
		f, err := os.OpenFile(out, os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := fmt.Fprintf(f, "url=%s\n", url); err != nil {
			return err
		}
	}
	return nil
}
