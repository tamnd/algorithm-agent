# 04. The object model

## 4.1 The 12 kinds

| Kind | It is | Example | Has implementations | Has a Lean part | Has a scene |
|---|---|---|---|---|---|
| `area` | a node of the area tree (05.2) | Shortest paths | no | no | optional |
| `concept` | an idea used to design or analyse algorithms | Loop invariant, amortized analysis | no | optional | optional (explorable) |
| `definition` | a precise term | Stable sort, flow network, topological order | no | yes, a Lean `def` or `structure` | optional |
| `model` | a model of computation or of cost | Word RAM, comparison model, external memory | no | optional | optional |
| `problem` | a computational problem with input, output and their relation | Sorting, single-source shortest paths, LCS | no, it has the contract | yes, the spec | optional |
| `abstract-type` | an abstract data type: operations and their laws | Priority queue, dictionary, disjoint sets | no, it has the contract | yes, the laws | optional |
| `algorithm` | a method that solves a problem | Merge sort, Dijkstra's algorithm, KMP | yes | yes | yes |
| `data-structure` | a representation that implements an abstract type | Binary heap, red–black tree, disjoint-set forest | yes | yes | yes |
| `technique` | a paradigm or a reusable pattern | Divide and conquer, dynamic programming, two pointers | no, it links examples | optional | optional |
| `theorem` | a result about algorithms or problems | Master theorem, the Ω(n log n) sorting lower bound, max-flow min-cut | no | yes | optional |
| `exercise` | a practice problem on a public judge or in a public text | CSES 1671 "Shortest Routes I" | no, link and metadata | no | no |
| `person` | a person who found or named items | Edsger W. Dijkstra | no | no | no |

Notes:

- A **variant** is an algorithm or a data structure with a `variant_of` edge, for example `merge-sort-bottom-up` → `merge-sort`. A variant has its own page when it has its own pseudocode, analysis or implementation. A difference of one line is a section of the parent page.
- A **proof** is not a node. A proof belongs to the node that it proves. A node can have several proofs (06.5). Each proof has an id in the node, for example `merge-sort#proof-invariant`.
- A **lemma** that only one proof uses lives in that proof. A lemma that many pages use (the harmonic bound, the potential method) is a `theorem` or a `concept`.

## 4.2 Tags, slugs and URIs

- **Tag.** 6 characters of Crockford base32, the alphabet of math2, for example `4MRG7S`. A tag is permanent and never reused. `registers/tags.reg` lists each tag and its slug at minting. A new node has the provisional id `NEW-<slug>` until `algo mint` gives it a tag before the merge. `algo mint` never gives a tag that is in `registers/tags.reg` or `registers/retired.reg`.
- **Slug.** Lowercase ASCII words joined by `-`, unique over all nodes of all kinds, for example `dijkstra`, `binary-heap`, `single-source-shortest-paths`. The slug is the URL. It can change; the old slug goes to `registers/aliases.reg` and the site redirects it. A slug never names a site section (14.2 lists the reserved words).
- **URI.** `al://4MRG7S`, and `al://4MRG7S@2027.03` for the node in a release. Fragments name parts: `al://4MRG7S#proof-invariant`, `al://4MRG7S/impl/rust`.
- **Names.** `name` is the common English name. `aliases` lists the other names, each with a language when it is not English (`{name: "Алгоритм Дейкстры", lang: ru}`). Search indexes every alias (15.2).

## 4.3 The node directory

A node lives at `nodes/<top-area>/<sub-area>/<slug>/`. The two area levels are the primary area of the node (05.2). They keep each directory under 1,000 entries, which GitHub can list. Moving a node to another area changes its path but not its tag or slug, and `algo mv` does it.

An algorithm directory has this form. Other kinds use the parts that they need (4.1).

```
nodes/sorting/comparison/merge-sort/
  node.md                 front matter and the text (06)
  proofs/
    invariant.md          a human proof (06.5)
    recurrence.md         a second proof, of the running time
  impl/
    python/merge_sort.py  the implementation (08)
    python/test_merge_sort.py
    cpp/merge_sort.hpp
    cpp/merge_sort_test.cpp
    rust/merge_sort.rs    tests in `mod tests`
    go/merge_sort.go
    go/merge_sort_test.go
    …                     one directory per language
  viz/
    scene.ts              the scene (10)
    presets.yaml          named inputs: random, sorted, reversed, worst case
  bench.yaml              optional: sizes and generators for the complexity check (09.7)
  figures/                optional: SVG figures that are not scenes
```

A problem directory has the contract instead of implementations:

```
nodes/sorting/sorting-problem/
  node.md
  contract.yaml           the signature, the input and output schemas, the checker (09.1)
  cases/
    basic.jsonl           hand-written cases with names
    edge.jsonl            empty, one element, duplicates, extremes
    generated.yaml        generator specs for random and adversarial cases
  checker.py              the checker, when the output is not unique (09.3)
  generators/             generators of inputs (09.4)
```

The Lean part lives in the Lean package, not in the node directory, because Lean module names cannot have hyphens: `lean/Algo/Sorting/MergeSort.lean`. The front matter names the module and the declarations, and audit rule L01 checks that each one exists (13.3).

## 4.4 The front matter

YAML front matter of `node.md`. The schema is `schema/node.v1.json`. Fields that do not apply to a kind are absent.

```yaml
---
tag: 4MRG7S
kind: algorithm
slug: merge-sort
name: Merge sort
aliases: [mergesort, "two-way merge sort"]
area: sorting/comparison              # primary area (05.2)
areas: [techniques/divide-and-conquer] # secondary areas
level: intro                          # intro | core | advanced | research (05.5)
summary: >-
  Sort a list by splitting it in half, sorting each half and merging the two sorted halves.
solves: [sorting-problem]             # problem slugs; the contract comes from here
techniques: [divide-and-conquer]
uses: [merge-two-sorted-lists]        # algorithms and data structures that it calls
variant_of: null
model: comparison                     # the cost model of the complexity claims
complexity:
  vars: {n: "the length of the input"}
  time:
    worst: "Θ(n log n)"
    average: "Θ(n log n)"
    best: "Θ(n log n)"
  space:
    extra: "Θ(n)"
  comparisons:
    worst: "n⌈lg n⌉ − 2^⌈lg n⌉ + 1"
properties:
  stable: true
  in_place: false
  online: false
  deterministic: true
  adaptive: false
  parallel: "Θ(n) work-efficient merge gives Θ(log³ n) span"   # free text, optional
history:
  year: 1945
  people: [john-von-neumann]
  first: {source: "goldstine-von-neumann-1948", note: "described in a 1945 manuscript"}
prereqs: [recursion, asymptotic-notation, loop-invariant]     # what to read first
same_as:
  wikidata: Q332015
  wikipedia: Merge_sort
  dads: mergesort
  cpalgorithms: null
  rosetta: Sorting_algorithms/Merge_sort
  math2: null
lean:
  module: Algo.Sorting.MergeSort
  model: Algo.Sorting.MergeSort.mergeSort            # the executable model
  spec: Algo.Sorting.SortingProblem.IsSort            # from the problem
  correct: Algo.Sorting.MergeSort.mergeSort_isSort
  cost: Algo.Sorting.MergeSort.mergeSort_time_le      # in TimeM
  cslib: Cslib.Algorithms.MergeSort                   # an upstream declaration, when one exists
sources:
  - {book: "isbn:9780262046305", locator: "CLRS 4e, Section 2.3"}
  - {book: "isbn:9780321573513", locator: "Sedgewick–Wayne 4e, Section 2.2"}
  - {url: "https://jeffe.cs.illinois.edu/teaching/algorithms/", locator: "Erickson, Chapter 1.4"}
exercises: [cses-1621, leetcode-912]
---
```

Rules:

- `complexity` values are strings in a small grammar: `Θ(...)`, `O(...)`, `Ω(...)`, exact counts, and the words `amortized`, `expected`, `w.h.p.`. `algo fmt` normalizes them, and audit rule E04 parses them. Each variable is defined in `vars`. Each claim names its case: `worst`, `average` (with the input distribution in a note), `best`, `amortized` or `expected` (over the random choices of the algorithm).
- `model` is the slug of a `model` node. A claim in another model goes in `complexity_in`, a list of `{model, time, space}`.
- `properties` uses fixed keys for each kind, listed in `schema/properties.v1.yaml`. Sorting has `stable`, `in_place`, `adaptive`, `online`. Graph algorithms have `negative_weights`, `directed`, `weighted`. Data structures have `persistent`, `concurrent`, `external`. The comparison tables of the site (14.5) read these keys.
- Status fields do not exist. The build computes the ladders (4.8), and audit rule T05 rejects a status written by hand.

**Conditions on `solves`.** The short form `solves: [sorting-problem]` means "with no condition, under the bounds of `complexity`". An algorithm that is correct only under conditions, or whose bound depends on them, uses the long form:

```yaml
# in nodes/graphs/shortest-paths/dijkstra/node.md
solves:
  - problem: single-source-shortest-paths
    requires: {negative_weights: false}
    time: "O((n + m) log n)"          # with a binary heap; optional, defaults to complexity.time.worst
  - problem: single-source-shortest-paths
    requires: {negative_weights: false, integer_weights: true, max_weight: "C"}
    time: "O(m + nC)"                  # Dial's bucket queue
    note: "with a bucket queue instead of a heap"
```

A problem node declares the keys that its algorithms may use, with their types, in `constraints`:

```yaml
# in nodes/graphs/shortest-paths/single-source-shortest-paths/node.md
constraints:
  weighted: {type: bool, default: true}
  negative_weights: {type: bool, default: false}
  directed: {type: bool, default: true}
  acyclic: {type: bool, default: false}
  integer_weights: {type: bool, default: false}
  max_weight: {type: var, doc: "the largest absolute weight, C"}
  n_max: {type: size}
  m_max: {type: size}
```

- Each key of a `requires` is a key of the problem's `constraints`, or of a problem that it `generalizes` (rule G09 checks the coverage, and E04 parses the bounds).
- A family of problems shares keys through `schema/properties.v1.yaml`, so `negative_weights` means the same thing in every graph problem.
- The constraint search of 15.4 reads only these fields.

## 4.5 Abstract types and data structures

An `abstract-type` node lists its operations with signatures and laws:

```yaml
operations:
  - {name: insert, sig: "(Q, key) -> Q"}
  - {name: find_min, sig: "(Q) -> key?", law: "find_min(Q) = min(keys(Q))"}
  - {name: delete_min, sig: "(Q) -> (key?, Q)"}
  - {name: decrease_key, sig: "(Q, handle, key) -> Q", optional: true}
```

A `data-structure` node names the types it implements (`implements: [priority-queue]`) and gives the cost of each operation:

```yaml
implements: [priority-queue]
operations:
  insert:      {time: "O(log n)"}
  find_min:    {time: "O(1)"}
  delete_min:  {time: "O(log n)"}
  decrease_key: {time: "O(log n)"}
  build:       {time: "O(n)"}
```

The contract of a data structure is the contract of its abstract type: a test case is a sequence of operations with expected results (09.1).

## 4.6 Edges

Edges are typed fields of the front matter. The build derives the reverse edges. `schema/edges.md` defines each one.

| Edge | From → to | Meaning |
|---|---|---|
| `solves` | algorithm → problem | the algorithm meets the contract of the problem |
| `implements` | data-structure → abstract-type | the structure meets the laws of the type |
| `uses` | algorithm, data-structure → algorithm, data-structure | it calls or contains the target |
| `techniques` | algorithm, data-structure → technique | it is an instance of the technique |
| `variant_of` | algorithm → algorithm | it changes the target in a named way |
| `generalizes` | problem, algorithm → problem, algorithm | the target is a special case |
| `reduces_to` | problem → problem | a reduction, with its cost in a note |
| `lower_bound` | theorem → problem | the theorem bounds every algorithm for the problem in a model |
| `proves` | theorem → algorithm, problem | the theorem is about the target |
| `prereqs` | any → any | read the target first; it forms the reading DAG (05.6) |
| `model` | algorithm, theorem → model | the model of the cost claims |
| `exercises` | algorithm, technique, data-structure → exercise | practice for the target |
| `people` | any → person | found, named or improved it |
| `uses_math` | any → math2 node | a mathematical prerequisite in math2 |
| `see_also` | any → any | a related page, at most 8 |

Audit rules G01–G08 check the graph: no cycle in `prereqs`, `variant_of` or `generalizes`; each `solves` target is a problem; each `implements` target is an abstract type; each node of the core has at least one `prereqs` path from a foundation node (13.3).

## 4.7 Provenance

Each file in a node directory has a record in `prov/<tag>.jsonl`, one line per change: the date, the file, the operation, the author (a person's GitHub login or an agent run id), the reviewer when there is one, and the source ids that the change used. `algo prov` writes it from the commit and the run record. The page shows it under "History" (14.4), as in math2.

## 4.8 The computed ladders

The build computes five ladders for each node that can have them. No person and no agent writes them (rule T05).

**H. Human proof** (for algorithms, data structures and theorems)

| Level | Meaning |
|---|---|
| H0 | no proof |
| H1 | a sketch: the main idea, marked as a sketch |
| H2 | a full proof of correctness and of each complexity claim of the front matter, to the standard of 06.5 |
| H3 | H2, and a person who is not the author reviewed this exact proof hash |

**L. Lean** (07.8)

| Level | Meaning |
|---|---|
| L0 | no Lean part |
| L1 | the spec and the executable model exist, and the spec is reviewed against the problem text (fidelity) |
| L2 | the correctness theorem is proved: no `sorry`, only the standard axioms |
| L3 | L2, and each time and space claim of the front matter is proved in the cost model |
| L4 | L3, and an implementation is proved to refine the model (the Rust one through Aeneas, 07.6) |

**I. Implementation** (for each language, 08.8)

| Level | Meaning |
|---|---|
| I0 | no implementation |
| I1 | it builds with no warnings and passes the shared cases of the contract |
| I2 | I1, and it passes the random and the adversarial differential tests against the oracle, and its tests kill at least 85% of the mutants (09.6) |
| I3 | I2, and lint and format are clean, the complexity check passes (09.7), and the code follows the conventions of 08.3 |
| I4 | I3, and a person reviewed this exact file hash |

**V. Visualization** (10.9)

| Level | Meaning |
|---|---|
| V0 | none |
| V1 | static figures or frames only |
| V2 | an interactive scene with step back, custom input and code sync, that passes the trace checks |
| V3 | V2, and a person reviewed it for clarity and accessibility |

**W. Writing** (06.8): W0 stub to W5 complete, as in math2's writing levels.

A node is **complete** when it has H3, L3, I3 in each tier-1 language, I2 in each tier-2 language, V2 and W4. The badges of the site (13.1) and the coverage levels (16.2) are computed from these ladders.

A review counts only for the hash that it reviewed. When the proof, the Lean statement or the file changes, the review stops counting and the level drops until a new review, as in math2.

## 4.9 Prohibitions

- No status, level or badge in front matter (T05).
- No copied text or code from a source whose licence forbids it (11.5, rule C01).
- No implementation that imports a package outside the standard library of its language (08.4, rule I07).
- No `sorry`, `admit`, `native_decide` or new axiom in a module that a node cites as proved (07.5, rule L03).
- No complexity claim without a model and defined variables (E04).
- No model-written statement of a problem contract that no person has reviewed, for the core (09.1, rule S06).
