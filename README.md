# algorithm-agent

[![CI](https://github.com/tamnd/algorithm-agent/actions/workflows/ci.yml/badge.svg)](https://github.com/tamnd/algorithm-agent/actions/workflows/ci.yml)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/tamnd/algorithm-agent/badge)](https://scorecard.dev/viewer/?uri=github.com/tamnd/algorithm-agent)
[![Go Reference](https://pkg.go.dev/badge/github.com/tamnd/algorithm-agent.svg)](https://pkg.go.dev/github.com/tamnd/algorithm-agent)

This repository holds `algo`, the toolchain of [algorithm.tamnd.com](https://algorithm.tamnd.com). The content is in [tamnd/algorithm](https://github.com/tamnd/algorithm).

`algo` reads the content tree, checks it against the specification, runs the tests of the implementations, bundles the visualizations and renders the site.

## Install

You need Go 1.27 or later.

```sh
go install github.com/tamnd/algorithm-agent/cmd/algo@latest
```

You can also download a binary from the [releases](https://github.com/tamnd/algorithm-agent/releases). Each binary has a build provenance attestation. To verify a binary, run `gh attestation verify algo_linux_amd64.tar.gz -R tamnd/algorithm-agent`.

## Use

Run these commands in a clone of tamnd/algorithm.

| Command | What it does |
| --- | --- |
| `algo new algorithm heap-sort -area sorting/comparison-sorts` | Makes a node directory with the sections of its kind. |
| `algo fmt` | Puts node.md, the proofs, the presets, the registers and the topics in canonical form. Add `-check` to only report. It never changes code. |
| `algo mint` | Gives each provisional id `NEW-<slug>` a permanent tag, and writes it into tags.reg and the provenance. Add `-dry-run` to see the tags first. |
| `algo mv <slug> <new-slug>` | Renames a node. It changes each edge and each link that names the old slug, and adds the old slug to `registers/aliases.reg`, so the old URL redirects. Give an area such as `sorting/comparison-sorts` to move the node to another area. The tag does not change. |
| `algo check` | Runs the audit rules. The S rules (S01 to S06, S09) check the contract and the cases, and the L rules (L01 to L05, L07, L09) check the Lean part of a node. The L rules that need the declarations read `build/lean/decls.jsonl`, or the file of `-decls`. A rule of the level `level` does not fail the run, but it keeps the node at a lower rung. Add `-checkers` to run each custom checker on its `checker_cases.jsonl` (S03), in the Python image of the sandbox. Add `-changed origin/main` to check only the nodes that a branch changes. |
| `algo status binary-search` | Shows the ladders of a node and what the next level needs. |
| `algo gen -lang rust` | Writes the build tree of each language to `build/lang/<lang>/`: a copy of each implementation, its build files, `algo-test.sh` and the adapters. `SOURCES` maps each file back to the node. |
| `algo test -unsafe-local binary-search` | Runs the native tests of the implementations, then the shared cases of each problem that the node solves. A wrong result names the case and the language. A case that is marked flaky counts as a failure, also when it passes (S09). With `-changed origin/main`, only the implementations that a change touches run. With `-cache dir`, a result that passed is used again while algo, the toolchain, the files and the cases stay the same. |
| `algo lean build` | Runs `lake build` in `lean/`, then AlgoExtract, which writes each declaration of the namespace `Algo` to `build/lean/decls.jsonl` with its type and its axioms. It fails if a declaration uses an axiom other than `propext`, `Classical.choice` and `Quot.sound`, such as `sorryAx`, and names the declaration. It also checks that each name of a `lean` field exists and has the right kind, and that each time claim of a node with `lean.cost` has a class lemma of the same class in its module (rule L05). With `-time`, it elaborates each module again with the Lean profiler and reports a module whose own time, without the imports, is more than 60 seconds (rule L10, a warning up to 300 seconds). `algo status` and the site read the file to compute the L ladder. The site shows the declarations in the section "Formal proof" of the node page and on the page `/<slug>/lean/`. |
| `algo viz check` | Builds each scene and runs each preset two times. |
| `algo site serve` | Renders the site and serves it at http://localhost:8150. It renders again when a file changes. |
| `algo site check` | Renders the site and checks each internal link. |
| `algo site deploy` | Renders the site, checks the limits of Cloudflare Pages (19500 files, 24 MB for one file), and uploads the tree with wrangler. It reads `CLOUDFLARE_API_TOKEN` and `CLOUDFLARE_ACCOUNT_ID` from the environment and does not print them. `-preview` deploys to the alias of a branch, `-no-build` uploads the `-out` directory as it is. |

`algo help` shows all commands. Commands that are planned but not built yet exit with code 3 and name their milestone.

The exit codes are: 0 for success, 1 when a check finds a problem, 2 for a usage error and 3 for a planned command.

## Safety of `algo test`

`algo test` compiles and runs code from the content tree. When the content repository has `manifests/images.yaml`, each build, test, adapter and custom checker runs in a new container of a runner image, with no network, a read-only root and limits on memory, CPU and processes. The images are in [images/](images/README.md). Give `-sandbox podman` to use Podman. Without the sandbox, `algo test` runs on your computer only with the flag `-unsafe-local`. Use it only on code that you have read. In CI, each language runs in its own job with no secrets.

## Layout

| Directory | Contents |
| --- | --- |
| `cmd/algo` | The command. |
| `node` | Reads the content tree. |
| `check` | The audit rules of spec part 13. |
| `status` | The ladders of a node. |
| `formal` | Builds the Lean package, runs AlgoExtract (`formal/Extract.lean`), and checks the axiom policy. |
| `contract` | Reads the contract and the shared cases of a problem, and decides if an output is right. |
| `lang` | Writes the build trees, and runs the native tests and the adapters for each language. Errors show the node paths, not the paths of the tree. The JSON codecs of the adapters are in `lang/codec`, under the MIT licence. |
| `viz` | Bundles the scenes with esbuild and checks them in a JavaScript runtime. |
| `markup` | Markdown, pseudocode and math. |
| `katex` | Renders TeX to HTML on the server. |
| `site` | The templates, assets and pages of the site. |
| `spec` | The specification. The site publishes it at [/spec/](https://algorithm.tamnd.com/spec/). |
| `testdata/corpus` | A small content tree for the tests. |

## Develop

```sh
make test     # go vet and go test
make lint     # golangci-lint
make site     # render testdata/corpus into build/site
```

Read [CONTRIBUTING.md](CONTRIBUTING.md) before you open a pull request. To report a security problem, read [SECURITY.md](SECURITY.md).

To release, push a tag `vX.Y.Z` and run `scripts/release.sh vX.Y.Z`. The script builds the archives from a clean tree of the tag and publishes the release. The release workflow then replaces the files with copies that have build provenance. A patch release comes every few merged pull requests. A minor release comes at the end of each milestone.

## Roadmap

The work is in six milestones, M0 to M5, from November 2026 to December 2028. [Spec 19](https://algorithm.tamnd.com/spec/19-milestones/) gives the goal and the exit criteria of each one. The [milestones](https://github.com/tamnd/algorithm-agent/milestones) list the open issues. Issues with the label `good first issue` are small and clear.

## Licence

The code is under the [Apache License 2.0](LICENSE). The scene package `viz/packages/core` is under the MIT licence, so scenes in the content repository can use it. The parts from other projects are listed in [THIRD_PARTY.md](THIRD_PARTY.md).
