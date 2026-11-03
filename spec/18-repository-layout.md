# 18. Repository layout

## 18.1 `tamnd/algorithm`, the content

```
algorithm/
  README.md                 what the repository is, how to read and add a node
  CONTRIBUTING.md           the steps for people; the DCO
  CURATION.md               how to decide inventory outcomes and merges
  REVIEWING.md              how to review a proof, a spec, an implementation or a scene
  LICENCE                   the reasons for the zones
  LICENSES/                 CC-BY-SA-4.0, MIT, Apache-2.0, CC0-1.0
  REUSE.toml                the licence of each path (17.1)
  CITATION.cff
  nodes/
    <top-area>/<sub-area>/<slug>/
      node.md               front matter and text (04.4, 06)
      proofs/*.md           human proofs (06.5)
      contract.yaml         problems and abstract types (09.1)
      cases/*.jsonl         shared cases
      cases/generated.yaml  generator specs (09.4)
      checker.py            custom checker (09.3)
      generators/*.py       input generators
      impl/<lang>/…         implementations and native tests (08)
      viz/scene.ts          the scene (10)
      viz/presets.yaml      named inputs
      bench.yaml            complexity check settings (09.7)
      figures/*.svg         static figures
  lean/
    lakefile.toml           the `Algo` package; requires Mathlib and CSLib at the pins
    lean-toolchain          leanprover/lean4:v4.35.0
    lake-manifest.json
    Algo.lean               imports every module
    Algo/
      Prelude/              shared definitions: cost ticks, sortedness, permutations, graphs
      <Area>/<Name>.lean    one module per node with a Lean part (07)
    AlgoExtract/            the extractor (07)
    test/                   tests of the package itself
  lean-aeneas/              the L4 workspace for Rust refinement, when its Lean version differs (Q-05)
  topics/
    areas.yaml              the area tree (05.2)
    core.yaml               the core (16.3)
    courses/*.yaml          courses (05.6)
    crosswalk/<scheme>.yaml acm-ccs-2012.yaml, msc2020.yaml, wikipedia.yaml, dads.yaml (16.4)
  registers/
    tags.reg                each tag, slug and commit (04.2)
    aliases.reg             old slugs and their targets
    retired.reg             retired tags
    languages.yaml          the language register (08.2)
    sources.yaml            the source register (11.1)
    canon/*.yaml            the tables of contents of the canon books (11.8)
    universe/items.jsonl    the universe register (16.1), sharded when it grows
    abbreviations.yaml      abbreviations and their expansions (05.7, 15.2)
    constraint-words.yaml   words that the palette reads as constraints (15.4)
    search-pins.yaml        pinned search results, at most 200 (15.7)
    search-eval.yaml        the search evaluation queries (15.8)
    site-tiers.yaml         which files go to the assets and which to R2 (14.9)
    patents.yaml            algorithms with patents that may still be in force (17.5)
    disambiguation.yaml     shared names (05.7)
    difficulty.yaml         the common difficulty scale of exercises (11.4)
    reviewers.reg           reviewers, their areas, skills, keys and affiliations (13.2.3)
    waivers.yaml            audit waivers, each with a person, a reason and an end date (13.3.3)
    reserved.yaml           reserved words that cannot be slugs (14.2)
  records/
    mentions/<source>/      mentions by source, sharded (11.3)
    inventory/              inventory rows, sharded
  prov/<tag>.jsonl          provenance of each node's files (04.7)
  reviews/                  signed review records, bound to hashes (13.2)
  schema/
    node.v1.json  contract.v1.json  case.v1.json  scene.v1.json  source.v1.json
    language.v1.json  universe.v1.json  review.v1.json  properties.v1.yaml  levels.v1.yaml  coverage-parts.v1.yaml
    edges.md  outcomes.v1.yaml  complexity.grammar
  manifests/
    releases/next.yaml      the pinned `algo` version and the pins of the next release
    releases/<YYYY.MM>.yaml each release (18.4)
    toolchains.yaml         the pinned compiler, runtime and tool versions of each language (08.2)
    images.yaml             the runner images and their digests (08.7)
    limits.yaml             the time and memory limits of tests, per language (09.8)
    models.yaml             the model ids and prompt versions that agents use (12.3)
    budgets.yaml            the monthly budget of each agent role (12.6)
  reports/
    audit/  coverage/  tests/  bench/  inventory/  agents/  sampling/  reviews/  incidents/
  build/                    generated, not committed (.gitignore)
    lang/<lang>/            generated build trees (08.5)
  .github/
    workflows/              18.3
    scripts/                Python checks that work without `algo`, as in math2
```

The path of a node is `nodes/<top-area>/<sub-area>/<slug>/`. A `person` node lives in `nodes/people/<initial>/<slug>/`, and an `exercise` node in `nodes/exercises/<judge>/<slug>/`. They are not in the area tree.

## 18.2 `tamnd/algorithm-agent`, the toolchain

```
algorithm-agent/
  README.md  CONTRIBUTING.md  SECURITY.md  LICENSE  Makefile  go.mod
  cmd/algo/                 main
  internal/ (or top-level packages, as in math2-agent)
    node/  schema/  tag/  graph/  taxonomy/  markup/  katex/
    lean/                   drives lake and AlgoExtract
    lang/                   the language register and the generators of build trees
      gen/<lang>/           one generator per language: build files and the JSON adapter (09.2)
    runner/                 sandboxed runs in images
    testkit/                contracts, cases, checkers, generators, oracle, differential, mutation
    mutate/<lang>/          mutation operators per language (09.6)
    bench/                  complexity checks
    viz/                    bundling, headless runs, frames, trace checks
    status/  audit/  coverage/  universe/  ingest/  copycheck/  prov/
    agent/                  roles and the loop
    queue/  llm/  sandbox/  release/  search/  site/  mcp/
  site/
    templates/*.html        the layouts of 14
    assets/                 site.css, site.js, fonts, icons
  viz/                      the TypeScript workspace of @algo/viz
    packages/core/          the trace protocol and the player
    packages/primitives/    array, grid, graph, tree, table, heap, hash, list, stack, plane, string, bits, numberline, chart
    packages/explorables/   the widgets of concept pages
    test/
  images/
    <lang>/Dockerfile       one runner image per language (08.7)
  prompts/
    <role>/<version>.md     versioned prompts (12.3)
  spec/                     a copy of this specification, as in math2-agent
  testdata/                 small trees for the tests of `algo`
```

## 18.3 The `algo` command tree

| Command | Does | Spec |
|---|---|---|
| `algo new <kind> <slug>` | make a node directory with a provisional id and the sections of its kind | 04, 06 |
| `algo fmt [paths]` | format front matter, complexity strings, pseudocode, and each implementation with its language's formatter | 04.4, 08.3 |
| `algo check [-changed]` | schema, audit of the changed nodes, Lean of the changed modules, tests of the changed implementations, scene checks | 03.5 F2 |
| `algo audit [-changed] [-rule ID]` | the audit rules | 13.3 |
| `algo mint` | give tags to provisional ids before the merge | 04.2 |
| `algo mv <slug> <new-slug \| area>` | rename a node or move it to another area; the old slug goes to `aliases.reg` | 04.2, 04.3 |
| `algo gen [lang]` | make the build trees under `build/lang/` | 08.5 |
| `algo test [slug] [-lang L] [-kind cases\|diff\|prop\|mutation]` | run tests in the runner images | 09 |
| `algo bench [slug] [-lang L]` | run the complexity checks | 09.7 |
| `algo lean build \| extract \| status` | build the `Algo` package, extract bindings and axioms, show the L level | 07 |
| `algo oracle <problem>` | build and run the oracle of a problem on given cases | 09.5 |
| `algo viz build \| frames \| check [slug]` | bundle scenes, render frames, check traces | 10 |
| `algo build` | the full build with the cache | 03.5 F3 |
| `algo status <slug>` | the ladders of a node and the reasons | 04.8 |
| `algo ingest <source>` | harvest a source into the raw store and the mentions | 11.3 |
| `algo universe merge \| report` | join mentions into inventory rows; report open rows | 11.3, 16.1 |
| `algo coverage` | the coverage levels and reports | 16 |
| `algo queue [list\|add\|lease]` | the work queue | 12.4 |
| `algo agent run <role> [-item ID] [-budget USD]` | run an agent role | 12 |
| `algo prov <slug>` | show or write provenance | 04.7 |
| `algo review <slug> <part> [-renew]` | make a review record for the current hashes, to sign and commit | 13.2 |
| `algo quarantine <slug> <part> -reason R` | withdraw a part from the site while it is checked; `-lift` restores it | 13.5 |
| `algo search build` | build the search index | 15 |
| `algo site build \| serve \| check \| deploy` | render, preview, check and publish the site | 14 |
| `algo mcp` | run the MCP server | 14.8 |
| `algo release cut \| verify \| tag` | the monthly release | 18.4 |
| `algo bump [lean\|lang L]` | move a pin to a new version in a pull request | 07, 08.2 |

## 18.4 Releases

Releases use calendar versions, `YYYY.MM`, one each month, as in math2. The first is 2027.03 (M1).

1. `algo release cut` builds `main`, writes `manifests/releases/<YYYY.MM>.yaml` (the commit, the `algo` version, the Lean, Mathlib and CSLib pins, the image digests, the model ids and the prompt versions, the ladder counts and the coverage), and opens the pull request `release/<YYYY.MM>`.
2. A second runner rebuilds it with `algo release verify`, and the check compares the graph, the ladders and the test results.
3. A person writes the summary and merges.
4. `algo release tag` runs the full audit, makes the tag and the GitHub release, and attaches the offline archive `algorithm-<YYYY.MM>-site.zip`, which opens from disk. Zenodo gets a DOI for each release from M3.

## 18.5 The workflows of `tamnd/algorithm`

| Workflow | Runs on | Does |
|---|---|---|
| `check.yml` | each pull request | `algo check -changed`: schema, audit, Lean of changed modules, tests of changed implementations in a matrix of images, scene checks; posts the ladder table |
| `tags.yml` | each pull request | fails while a provisional id exists; the maintainer runs `mint` |
| `build.yml` | each push to `main`, and nightly on the build server | `algo build` with the cache; the nightly run adds benchmarks and mutation tests of the whole core |
| `pages.yml` | after `build.yml` | `algo site build`, `site check`, `site deploy` to Cloudflare; the mirror to GitHub Pages |
| `import.yml` | each Monday | `algo ingest`, `universe merge`, `coverage`; a pull request with the changes |
| `queue.yml` | daily | `algo queue` from the new gaps; agents lease items on the build server |
| `bump.yml` | weekly | `algo bump lean` and `algo bump lang` for each language with a new stable release; a pull request with the report |
| `release.yml` | the first day of each month | `algo release cut` when the variable `ALGO_RELEASES` is `true` |
| `audit.yml` | daily | the full audit; a pull request when the report changes |
| `codeql.yml`, `scorecard.yml` | weekly | security checks, as in math2 |

## 18.6 The workflows of `tamnd/algorithm-agent`

| Workflow | Does |
|---|---|
| `ci.yml` | Go tests, lint, the tests of `@algo/viz`, a build of `testdata/` trees |
| `images.yml` | build and push the runner images when `images/` changes, and record the digests |
| `release.yml` | build `algo` for Linux, macOS and Windows on a tag; publish the images and `@algo/viz` |
| `spec.yml` | check the links and the section references of `spec/` |
