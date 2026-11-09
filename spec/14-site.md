# 14. The site: algorithm.tamnd.com and its mirror

The site is where people use the corpus. It is built from the content tree and `build/` by `algo site build` (03.5 F4). It follows the design of math2 (https://math2.tamnd.com): one reading column, the Inter typeface, a hue for each area, no rounded corners, space and tints instead of lines, a light and a dark theme, and pages that work with no script. It adds what algorithms need: code in many languages, proofs at two levels, visualizations and comparison tables.

Evidence for the hosting numbers is in 01 and in spec 2149 10.6: Workers static assets take at most 20,000 files on Free and 100,000 on Paid, 25 MiB per file; GitHub Pages takes 1 GB per site; R2 costs $0.015/GB-month with no egress fee.

## 14.1 Design principles

1. **Reading first.** The text is one column of at most 42rem, at 17px with a line height of 1.75. Code, tables and the player may go wider (up to 72rem) but never push the text.
2. **The answer at the top.** Each node page starts with what a reader came for: what it is, what it solves, how fast it is, and whether it is proved and tested. The summary card (14.7.4) holds this in the first screen.
3. **One page, many depths.** The plain idea, the walk-through, the pseudocode, the human proof, the Lean proof and the code are all on the page, in that order. A reader stops where they want. Long parts fold, and nothing needs a click to be found by search.
4. **Evidence is visible and honest.** Each claim of proof, test or review is a badge that links to its evidence. A badge comes only from the computed ladders (04.8). What an agent wrote and nobody checked carries a label (14.4).
5. **Same idea, same look.** Each area has a hue (05.2). Its tiles, its cards, its breadcrumbs and the tint of its summary cards share it. Each kind has a word, not an icon, so the page says "Algorithm", "Problem", "Data structure".
6. **No script needed to read.** Script adds search, tabs that remember, sortable tables, the player and course progress. With no script, tabs become stacked sections, tables are in their default order, and the player becomes a filmstrip of static frames with captions (10.7).
7. **Fast on a cheap phone.** The budgets of 14.13 are checks in CI, not hopes.
8. **Stable links.** A URL that worked once keeps working. Slugs change only through `registers/aliases.reg`, and the old URL redirects.

## 14.2 URLs and reserved words

The node namespace is flat at the root of the site. Each node is at `/<slug>/`, whatever its kind and its area, so a URL does not change when a node moves in the tree (05.2).

| URL | Page |
|---|---|
| `/` | home (14.7.1) |
| `/<slug>/` | the node page of any kind (14.4, 14.7) |
| `/<slug>/<lang>/` | the implementation of an algorithm or a data structure in one language, for example `/dijkstra/rust/` (14.7.5). `<lang>` is a language id of `registers/languages.yaml` (08.2) |
| `/<slug>/proof/` | every human proof of the node at full length (14.7.6) |
| `/<slug>/lean/` | the Lean part: spec, model, theorems and the highlighted module (14.7.7) |
| `/<slug>/visualize/` | the full-window player (14.7.8) |
| `/<slug>/tests/` | the contract and the shared cases, for a problem or an abstract type (14.7.3) |
| `/<slug>.md`, `/<slug>.json` | the machine copies of the node (14.8) |
| `/<slug>/<lang>.txt` | the raw source file of the implementation, `text/plain` |
| `/embed/<slug>/` | the player alone, for an `<iframe>` in slides and course pages (14.7.8) |
| `/t/<tag>` | the node by tag; a redirect to `/<slug>/` |
| `/t/<tag>@<release>` | the node in a release; a redirect to the release site (14.18) |
| `/areas/` | the area tree (14.7.2) |
| `/problems/` | the problem index (05.3) |
| `/techniques/` | every technique (14.7.10) |
| `/learn/`, `/learn/<course>/` | courses and paths (14.6) |
| `/visualize/` | the gallery of scenes |
| `/cheatsheet/`, `/cheatsheet/<area>/` | the generated complexity tables (14.7.11) |
| `/compare/?a=<slug>&b=<slug>` | two or three nodes side by side, built in the browser from their JSON |
| `/kind/<kind>/` | every node of one kind |
| `/languages/`, `/languages/<lang>/` | each language, its toolchain pin, its coverage and its implementations |
| `/exercises/`, `/exercises/<judge>/` | the linked practice problems (14.7.12) |
| `/people/` | every person page; a person page itself is at `/<slug>/` |
| `/search/` | search (document 15) |
| `/random/` | a random node, with filters in the query |
| `/coverage/`, `/matrix/`, `/gaps/` | the dashboards (14.19) |
| `/latest/`, `/latest/feed.xml` | the newest and changed pages, with an Atom feed |
| `/badges/` | what each badge and level means |
| `/sources/`, `/universe/` | the source register (11) and the universe register (16) |
| `/reviewers/`, `/contribute/`, `/spec/`, `/about/` | the project |
| `/api/…`, `/llms.txt`, `/llms-full.txt`, `/sitemap.xml` | machine access (14.8) |

**Reserved words.** A slug cannot be one of these words, because each is a section of the site. `registers/reserved.yaml` holds the list, and audit rule X06 rejects a node slug that is on it (13.3).

```
about agents all api areas assets badges cheatsheet compare contribute courses coverage
embed exercises feed gaps index kind languages latest learn lean llms matrix mcp people
problems proof random releases reviewers robots search sitemap sources spec static
t techniques tests universe visualize 404
```

The second segment under a node is reserved for `proof`, `lean`, `visualize`, `tests` and the language ids. A language id is never a node slug at the second level, and the first level does not care, so a node may be called `rust` only if no language is called so; the rule X06 checks both lists.

Trailing slashes are canonical. `/merge-sort` redirects to `/merge-sort/`. Case is folded: `/Merge-Sort/` redirects to `/merge-sort/`. Common aliases redirect to their node: `/kmp/` goes to `/knuth-morris-pratt/` when the alias is unique (05.7). An ambiguous alias goes to its disambiguation page.

## 14.3 Information architecture and navigation

**The top bar** (sticky, blurred, as in math2):

```
[▲ algorithm]  Learn  Areas  Problems  Techniques  Visualize      [ Search algorithms…  ⌘K ]  ◐
```

| Item | Goes to | Answers |
|---|---|---|
| Learn | `/learn/` | "What should I learn next?" (05, question 4) |
| Areas | `/areas/` | "What is in this field?" |
| Problems | `/problems/` | "How do I solve this?" |
| Techniques | `/techniques/` | "Which idea solves problems like this?" |
| Visualize | `/visualize/` | "Show me." |
| Search box | the palette (15.3) | "What is X?" |
| ◐ | the theme | light, dark, or the system's |

On a narrow screen, the five links go into a menu button, and the search box becomes a search button that opens the palette.

**The breadcrumb** on each node page is the primary area: `Sorting / Comparison sorts / Algorithm`. The last word is the kind, linked to `/kind/<kind>/`.

**The footer** has four columns, as in math2:

| Learn | Reference | Build it with us | For machines |
|---|---|---|---|
| Courses, Areas, Problems, Techniques, Search | Cheat sheets, Every node, Languages, Exercises, People, Visualize | Contribute, Specification, Coverage, Test matrix, Gaps, Sources, Universe, Reviewers, Badges | For agents, llms.txt, JSON API, MCP, Content on GitHub, Toolchain on GitHub |

The fine print gives the licences of 17.1, the `algo` version, the commit and the build date, as in math2.

## 14.4 The node page

Every node page has the same frame. The parts of the body depend on the kind (14.7) and on the sections of document 06.

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│ top bar                                                                          │
├──────────────────────────────────────────────────────────────────────────────────┤
│ Sorting / Comparison sorts / Algorithm                                           │
│ Merge sort                                                    [Share] [Edit]     │
│ Sort a list by splitting it in half, sorting each half and merging the halves.   │
│ ┌──────────────────────────── summary card (14.7.4) ────────────────────────┐   │
│ └────────────────────────────────────────────────────────────────────────────┘   │
│ Read first  Recursion · Asymptotic notation · Loop invariant · Merging lists    │
│ ◂ Algorithms I · 14 of 250 · Insertion sort ←        → Quicksort                 │  step strip
├───────────────────────────────────────────────────────────┬──────────────────────┤
│ body: the sections of the kind (06)                       │ On this page         │
│                                                           │  Idea                │
│                                                           │  See it run          │
│                                                           │  Pseudocode          │
│                                                           │  Why it is correct   │
│                                                           │  Running time        │
│                                                           │  Code                │
│                                                           │  Variants            │
│                                                           │  In Lean             │
│                                                           │  Practice            │
│                                                           │  Further reading     │
├───────────────────────────────────────────────────────────┴──────────────────────┤
│ Reading path ▸ (fold)      Where this leads (12 pages)      Related              │
│ About this page ▸ (fold): tag, URI, kind, levels, history, links here, formats   │
└──────────────────────────────────────────────────────────────────────────────────┘
```

The parts:

- **Head.** The breadcrumb, the name (`h1`), the one-line `summary`, and the buttons "Share" (the Web Share API, else copy the link) and "Edit" (the file on GitHub).
- **Summary card.** For algorithms, data structures and problems (14.7.4). Other kinds have a short line of facts instead: a definition shows "Used by 48 pages", a theorem shows its proof levels.
- **Read first.** The direct `prereqs` of the node, as links. A link to a page that does not exist yet is shown in the "missing" colour and has a title that says so, as in math2.
- **Step strip.** When the node is in a course, one strip for each course: the course name, the position, the previous and the next node. When a node is in more than three courses, the strip shows the course the reader is following (14.6) and a menu for the rest.
- **On this page.** A table of contents of the `h2` sections. On screens 1,200px and wider it sits in a right column and is sticky, with the current section marked as the reader scrolls. On narrower screens it is a folded "Contents" line under the head.
- **Body.** The sections of the kind (06.2 to 06.7), rendered from `node.md` and the other parts of the node directory.
- **Agent labels.** A section, an implementation or a proof that an agent wrote and no person reviewed carries a small label "AI · not reviewed" at its heading, which links to the run record. When a page has more than three such parts, one notice under the head says so instead, as in math2: "An agent wrote some of this page, and no person checked it yet: the Kotlin, Swift and Lua code, and the proof of the running time." Quarantined text (12.7) is not shown, and a notice says which parts are hidden.
- **Reading path.** Folded. The topological order of all the prerequisites, foundations first, as a numbered list and as an SVG graph on request, as in math2. Each box shows the kind and the evidence of the node.
- **Where this leads.** The nodes that list this one in `prereqs`, `uses` or `variant_of`, as a cloud of links, with a count.
- **Related.** `see_also`, variants, and the other algorithms that solve the same problem, at most 8.
- **About this page.** Folded. A table of facts: the tag (click to copy), the URI, the aliases, the kind, the level, the five ladders with their words (H2 "full proof", L3 "correctness and cost proved", …), the reasons the build gives each level ("L3 because `mergeSort_time_le` proves the time claim; the space claim has no theorem yet"), and what the node needs for the next level. Then **History**: each change from `prov/<tag>.jsonl` (04.7), the oldest first, with the date, the file, the operation and the author; an agent's change names its run and links to the run record. Then **What links here**, grouped by edge type. Then the formats: Markdown, JSON, "Edit on GitHub", "Report a problem" (a new GitHub issue with the URL filled in).

A **retired** node keeps its page. A notice says why and links to the node that replaces it. A merged node redirects.

## 14.5 Comparison tables

Problem pages, abstract-type pages and area pages have comparison tables. The build makes them from the front matter of the nodes: `complexity`, `properties`, `model`, `history` and the ladders. Nobody writes a comparison table by hand, so a table cannot disagree with the pages it lists.

The classic table on `/sorting-problem/`:

```
Algorithms that solve Sorting                       [filter: stable ☐ in place ☐ comparison ☑]
 Algorithm        Worst        Average      Best       Extra space  Stable In place Adaptive  Evidence
 ───────────────────────────────────────────────────────────────────────────────────────────────────
 Merge sort       Θ(n log n)   Θ(n log n)   Θ(n log n) Θ(n)         yes    no       no        L3 H3 23
 Heapsort         Θ(n log n)   Θ(n log n)   Θ(n log n) Θ(1)         no     yes      no        L3 H3 23
 Quicksort        Θ(n²)        Θ(n log n)ᵉ  Θ(n log n) Θ(log n)ᵉ    no     yes      no        L2 H3 23
 Introsort        Θ(n log n)   Θ(n log n)   Θ(n log n) Θ(log n)     no     yes      no        L1 H2 10
 Timsort          Θ(n log n)   Θ(n log n)   Θ(n)       Θ(n)         yes    no       yes       L0 H2 10
 Insertion sort   Θ(n²)        Θ(n²)        Θ(n)       Θ(1)         yes    yes      yes       L3 H3 23
 Counting sort    Θ(n + k)     Θ(n + k)     Θ(n + k)   Θ(n + k)     yes    no       no        L2 H3 23
 ───────────────────────────────────────────────────────────────────────────────────────────────────
 ᵉ expected over the random pivots.   k = the size of the key range.   Evidence: Lean level, proof level, languages.
 Lower bound: Ω(n log n) comparisons in the comparison model (theorem: comparison-sort-lower-bound).
```

Rules:

- **Columns come from the kind.** The columns of a table are the complexity cases that any row has and the property keys of `schema/properties.v1.yaml` for the problem's area (04.4). A property that no row has is not a column.
- **Models are never mixed silently.** Rows in different models go in separate groups with the model as a heading ("Comparison model", "Word RAM with w-bit keys"). Variables are defined under the table.
- **Sortable and filterable.** With script, a click on a column header sorts. Complexity cells sort by the parsed expression (E04), so Θ(n) comes before Θ(n log n) before Θ(n²). The filters are the property keys. The state goes into the URL query, so a filtered table can be linked.
- **Footnotes, not asterisks without meaning.** Expected, amortized and with-high-probability bounds carry a superscript letter, and the note under the table says what it means.
- **The best known bound** for the problem, the lower bound theorem and its model, and open gaps ("Is there an O(m + n) algorithm for SSSP in the comparison-addition model? Open.") are shown under the table.
- **Data structures** get an operation table on the abstract-type page: one row per structure, one column per operation (14.7.9).
- **Area pages** show one compact table per problem, with the three best-known rows and a link to the full table.

## 14.6 Paths and courses

- **Courses** (05.6) are at `/learn/<course>/`. A course page has a head with the hue of the course, the count of nodes, a meter of how many have a page at W3 or better, a "Start with …" button, and the sections of the course as numbered lists, as in math2's course page. A node with no page yet is listed in grey with "not written yet".
- **Progress.** The reader marks a node as done with a button at the end of the page, or the site marks it after the reader reaches the end. Progress is stored in the browser only, in `localStorage` under `al-progress` (a map of course to a set of tags). No account and no server. The course page shows a tick for each done node and the meter shows the reader's progress next to the written share. An "Export" and an "Import" button move progress between browsers as a small JSON file.
- **Following a course.** When the reader opens a node from a course page, the site remembers the course (`al-course`), and the step strip of each node page shows that course first.
- **Step strip** keys: `[` the previous node, `]` the next node in the followed course.
- **Paths.** On any node page, "Reading path" (14.4) gives the path to that node. `/learn/path/?to=<slug>&from=<slug>` gives the path between two nodes, minus what the reader has marked as done.
- **Course for teachers.** Each course has a printable outline (A4, with the URLs as QR codes) and an `.ics`-free plain list for syllabus pages.

## 14.7 The page types

### 14.7.1 The home page

```
A free, open map of algorithms
Every algorithm, proved and tested.
11,482 pages. 1,214 algorithms proved in Lean. Code in 23 languages, each tested against a proved program.

[ Search for an algorithm, a problem or a data structure, for example "shortest path" ]

[Start learning]  [Browse areas]  [I need to solve…]

Start with an area                                                       All areas →
 ■ Sorting 212   ■ Graphs 1,140   ■ Strings 386   ■ Data structures 904   ■ Number theory 241 …

I need to…                                                               Problem index →
 sort a list · find a shortest path · match a pattern · find a minimum spanning tree ·
 schedule jobs · compress a file · test if a number is prime · solve a linear program …

See it run                                                                Gallery →
 [frame] Dijkstra   [frame] Quicksort   [frame] KMP   [frame] Union–find   [frame] Convex hull

Just proved                                                               Latest →
 Red–black tree insertion · L3 · 2 days ago     Edmonds–Karp · L2 · 5 days ago …

Courses: Algorithms I · Data structures · Competitive programming · Interview essentials …
```

The counts come from the build. When the graph is small (M0 and M1), the home page says so and points to Contribute, as in math2.

### 14.7.2 Area pages

`/areas/` shows the 30 top areas as tiles with their hue, their node count and a thin meter of C6 coverage (16.2). An area page (`/graphs/`, `/shortest-paths/`) has:

- the head with the hue, the description of the area and what it does not hold (from `topics/areas.yaml`);
- **Start here**: the `intro` nodes of the area as cards;
- **Map**: the sub-areas, each with its count and its coverage meter, and the nodes in teaching order (by the prerequisite DAG, then by level);
- **Problems in this area** with compact comparison tables (14.5);
- **Techniques used here**, the top 8 by count;
- **Every node** by kind, in a filterable list;
- the crosswalk line: "This area matches ACM CCS … and MSC 05C85" with links (16.4).

### 14.7.3 Problem pages

A problem page (`/single-source-shortest-paths/`) has:

1. **The problem**, in plain words, then exactly: input, output, the relation, with the variables named. The text comes from `node.md`. A box shows the signature of the contract (`contract.yaml`, 09.1) in a neutral pseudo-type form and links to `/…/tests/`.
2. **Examples**: the first three named cases of `cases/basic.jsonl` rendered as small pictures where the contract has a picture type (a graph, an array, a grid), else as text.
3. **Which algorithm should I use?**: a short decision list written in `node.md` ("Non-negative weights: Dijkstra. Negative weights, no negative cycle: Bellman–Ford, or Johnson for all pairs. A DAG: relax in topological order…").
4. **The comparison table** (14.5).
5. **Lower bounds and hardness**: `lower_bound` theorems, `reduces_to` edges in both directions, the complexity class.
6. **Variants and generalizations**: the `generalizes` edges.
7. **Practice**: the exercises of the algorithms that solve it, by difficulty.

`/<slug>/tests/` shows the contract in full, the checker (when the output is not unique), every case file with its cases as a table (name, input size, notes), and the generators with their size classes. It links to the raw files.

### 14.7.4 Algorithm pages

The most important page. The summary card first:

```
┌─ Merge sort ────────────────────────────────────────────────────── tinted with the area hue ─┐
│ Solves     Sorting                        Technique  Divide and conquer                        │
│ Time       Θ(n log n) worst · average · best                                                   │
│ Space      Θ(n) extra                     Model      Comparison model   n = length of input    │
│ Properties [stable] [not in place] [not adaptive] [deterministic]                              │
│ Evidence   ✓ Proved in Lean · correctness and cost    ✓ Proof reviewed                         │
│            ✓ Tested in 23 languages                   ✓ Visual                                 │
│ Origin     John von Neumann, 1945                                                              │
└────────────────────────────────────────────────────────────────────────────────────────────────┘
```

- Each line of the card links: "Sorting" to the problem, the time to the analysis section, "Comparison model" to the model node, each property chip to its definition node (`stable-sort`), each badge to its evidence.
- **Badges** (13.1). "Proved in Lean" needs L2; it says "correctness" at L2 and "correctness and cost" at L3, and "and the Rust code" at L4. "Proof reviewed" needs H3; at H2 the badge reads "Full proof" without the tick colour. "Tested in N languages" counts the languages at I2 or better; a hover (or a tap) lists them with their levels. "Visual" needs V2. A missing badge is not shown as a red cross; the "About this page" table lists what is missing.
- Several complexity cases with the same bound collapse ("Θ(n log n) worst · average · best"). Different bounds are listed in order worst, average, best, amortized, expected.

Then the body. The order follows the sections of document 06 for an algorithm:

| Section | What the reader sees |
|---|---|
| Idea | the plain-words explanation, at most a few paragraphs, often with one static figure |
| See it run | the embedded player (10.5) with the "random" preset, 16:9, height capped at 60% of the viewport. A button "Open full window" goes to `/…/visualize/`. The player loads only when it comes into view |
| Pseudocode | the pseudocode block (06.4), with line numbers. Each line has an id; when the player runs on the same page, the current line is highlighted here too |
| Walk-through | an example traced by hand, often frames from the scene with captions |
| Why it is correct | the main human proof (06.5), with lemma proofs folded (14.7.6) |
| Running time and space | the analysis, one subsection per claim of the front matter, each with a link to its Lean theorem when there is one |
| Code | the implementation tabs (below) |
| Variants | the `variant_of` children as cards, and small variants as subsections |
| Pitfalls | the common mistakes, often with a wrong version and a counterexample case |
| Applications | where it is used |
| History | who, when, where first published, with the sources |
| In Lean | the statements of the spec and the two theorems, highlighted, and a link to `/…/lean/` |
| Practice | exercises, by difficulty and judge |
| Further reading | the sources, numbered, with locators |

**The code tabs.**

```
Code
 Python  C++  Java  C  C#  JavaScript  TypeScript  Go  Rust  Kotlin   More ▾   Lean
 ───────────────────────────────────────────────────────────────────────────────────────
 merge_sort.rs · 41 lines · I3 tested, lint clean, complexity checked   [Copy] [Raw] [GitHub]
  1  /// Sorts `a` in place with merge sort. Stable. O(n log n) time, O(n) extra space.
  2  pub fn merge_sort<T: Ord + Clone>(a: &mut [T]) {
  …
 ▸ Tests (merge_sort.rs, mod tests) · 6 tests, plus 1,240 shared cases and the oracle
 Run it:  algo test merge-sort -lang rust        Open the Rust page →
```

- The tab row has the 10 tier-1 languages in a fixed order, "More" for tier 2 (and tier 3 when they exist), and Lean last, because the Lean model is runnable code too.
- The reader's language is remembered across pages in `localStorage` under `al-lang`. The URL fragment `#code-rust` selects a tab and wins over the stored choice. A page that has no implementation in the stored language shows the first tier-1 language and says "No Kotlin version yet. [Help write it]".
- The tier-1 implementations are in the HTML of the page, highlighted at build time, so they work with no script (as stacked sections with headings) and are found by search engines. Tier-2 implementations are fetched from their language page when their tab is opened, to keep the page small (14.13).
- The header line of each file names the file, the line count, the I-level in words, and an "AI · not reviewed" label when it applies.
- **Copy** copies the file. **Raw** opens `/…/<lang>.txt`. **GitHub** opens the file at the commit of the build.
- **Code sync.** Lines of an implementation that carry a `@line` marker (10.6) map to pseudocode lines. Hovering or focusing a pseudocode line highlights the matching code lines in the open tab, and the reverse.
- **Tests** fold under the code: the native test file, and a line with the count of shared cases and generators that the implementation passed in the last build.

### 14.7.5 Language pages

`/dijkstra/rust/` answers searches like "dijkstra in rust". It is a full page, not a copy of the node page:

- the title "Dijkstra's algorithm in Rust" and a two-line summary with the bound;
- the full source file, highlighted, with Copy, Raw and GitHub;
- the native test file;
- the notes of the implementation, when it has them (08): the idioms used, why a `BinaryHeap` with `Reverse`, what the standard library gives;
- the toolchain pin from `manifests/toolchains.yaml` and the commands to run the tests with and without `algo`;
- the last results: shared cases passed, differential runs, mutation score, the complexity check with its fitted exponent (09.7);
- the pseudocode, folded, with code sync;
- links: the node page, the same algorithm in the other languages, and the other algorithms of the same problem in Rust.

The canonical URL of a language page is itself. Its structured data is `SoftwareSourceCode` (14.15).

### 14.7.6 Proofs

The main proof is on the node page. `/<slug>/proof/` has every proof of the node at full length (06.5): correctness by invariant, correctness by induction, each complexity proof, and alternative proofs.

The proof layout:

```
Theorem 1 (Correctness). For each list a, merge-sort(a) is a sorted permutation of a.

Proof. By strong induction on n = |a|.
  1. Base: n ≤ 1. …                                                          [Lean: mergeSort_base]
  2. Step: split a into l and r with |l| = ⌊n/2⌋ …
     ▸ Why are both halves shorter than a?                                     (folded detail)
  3. By the induction hypothesis, merge-sort(l) and merge-sort(r) are sorted permutations …
  4. By Lemma 2 (merge), …                                                                    ∎
```

- Theorems, lemmas and invariants have boxes with a left bar in the area hue, numbered on the page, with anchors (`#thm-1`).
- A step can fold its justification ("Why?"), so a reader sees the outline first. "Unfold all" and "Fold all" buttons are at the top of the proof.
- A step that matches a Lean lemma shows a small link to it. The link goes to the line on the Lean page.
- A sketch (H1) is marked "Sketch" at its heading, with what it leaves out.
- The proof level and the reviewer are named under the proof: "Full proof. Reviewed by @name on 3 March 2027."

### 14.7.7 Lean pages

`/<slug>/lean/` shows the formal part:

- the **pins**: Lean, Mathlib and CSLib versions, from `lean/lake-manifest.json`;
- the **spec**, the **model** and the **theorems**, each as a highlighted Lean block with a plain-words line above it ("This says: the output is sorted and is a permutation of the input");
- the **fidelity review** of the spec (L1): who reviewed it and for which statement hash;
- the **axioms** that the theorems use, from `#print axioms` in the extractor (07.5), which must be the standard three;
- the **whole module**, highlighted with the info that the build exports with SubVerso (the highlighter of Verso): hovering an identifier shows its type and its docstring; clicking a tactic shows the proof state before and after it in a side panel; identifiers from Mathlib and CSLib link to their documentation;
- "Check it yourself": the commands `lake exe cache get`, `lake build Algo.Sorting.MergeSort`;
- links to the upstream CSLib or Mathlib declaration when the node binds one.

With no script, the hovers are gone, and the proof states are in folded sections after each tactic block.

### 14.7.8 The player page and embeds

`/<slug>/visualize/` is the player of document 10 at full window: the canvas, the controls, the input editor, the presets, the pseudocode and code panel with the language choice, the variables, and the narration. The page has the node head and the summary in a thin bar, and a link back to the node page.

`/embed/<slug>/` is the player with no site frame, for an `<iframe>`. The query sets the options: `?preset=worst&lang=python&speed=0.5&autoplay=0&controls=1`. The node page has a "Embed" button that gives the `<iframe>` code. Embeds send no data anywhere.

`/visualize/` is the gallery: a grid of the V2 scenes with a still frame each, filterable by area and by primitive (arrays, graphs, trees, grids, geometry).

### 14.7.9 Data structure and abstract type pages

A data structure page has the summary card with an **operation table** instead of one time line:

```
Binary heap · implements Priority queue
 Operation      Time        Notes
 insert         O(log n)    sift up
 find-min       O(1)
 delete-min     O(log n)    sift down
 decrease-key   O(log n)    needs a handle
 build          O(n)        bottom-up heapify
 Space          Θ(n)        an array, no pointers
```

The player of a data structure takes a list of operations as its input (10.4). The abstract-type page (`/priority-queue/`) has the operations and their laws, and the comparison table of 14.5 with one row per implementing structure and one column per operation (binary, binomial, pairing, Fibonacci, radix heaps).

### 14.7.10 Technique pages

A technique page (`/dynamic-programming/`) has the idea, the signals ("You need the best of many overlapping choices…"), the template as pseudocode with holes, the examples from simple to hard as cards with their level, the sub-techniques, the common mistakes, and the exercises by difficulty band, from the `exercises` edges of the examples (05.4). `/techniques/` lists them all, grouped by family.

### 14.7.11 Cheat sheets

`/cheatsheet/` is the complexity cheat sheet, generated from the front matter: one compact table per problem family (sorting, searching, priority queues, dictionaries, shortest paths, spanning trees, string matching, …), each row linked to its page. Cells are coloured by the time class (05.5) on a scale that also works in grey. A print style fits each table on A4. `/cheatsheet/<area>/` gives the tables of one area. A small explorable at the top plots the growth of the common classes for n up to a reader-chosen bound.

### 14.7.12 Exercises

`/exercises/` lists the exercise nodes with their judge, their difficulty band (normalized from the judge's rating, 11.4), their techniques and the algorithms that solve them. Filters: judge, band, technique, area. Each exercise has a small page with the title, the link to the judge, the techniques, and a note when one exists. The statement is never copied (17.2). `/exercises/<judge>/` lists one judge.

### 14.7.13 Other kinds

- **Concept, definition, model, theorem**: the head, a short fact line, the body of 06, the Lean statement when it has one, and "Used by" with a count.
- **Theorem**: the statement in a box, the proofs, the Lean binding, the problems and algorithms it is about.
- **Model**: what an operation costs, what it can and cannot do, and the list of claims on the site that use it.
- **Person**: name, years, a short line, the items they found or named (from `people` edges) in order of year, and links to Wikidata and the Mathematics Genealogy Project. No photo unless its licence allows (17.2).

## 14.8 Machine access

Each page has copies for programs, as in math2:

| Path | Content |
|---|---|
| `/<slug>.md` | the node as Markdown: the front matter, the body, the proofs, and the tier-1 code as fenced blocks |
| `/<slug>.json` | the node as JSON: the front matter, the computed ladders, the edges in both directions, the list of implementations with their paths and levels, the scene and its presets |
| `/<slug>/<lang>.txt` | the raw implementation |
| `/api/nodes.json` | every node: tag, slug, kind, name, area, summary, levels. Sharded by kind when it passes 5 MB |
| `/api/graph.json` | every edge |
| `/api/problems/<slug>/cases.jsonl` | the shared cases of a problem, so other projects can test their code |
| `/api/coverage.json`, `/api/matrix.json` | the dashboards as data |
| `/api/search/…` | the search shards (15.5) |
| `/llms.txt` | the site for language models: what it is, the areas, the problem index, how to fetch a node, the licences |
| `/llms-full.txt` | the summaries of every node, one line each |

Each HTML page links its copies with `<link rel="alternate">`.

**The MCP server.** `algo mcp` serves the corpus to agents over MCP (stdio, and streamable HTTP when run as a service). The same server runs at `https://algorithm.tamnd.com/mcp` as a Worker that reads the static JSON, without the tool `run_cases`. Tools:

| Tool | Does |
|---|---|
| `get_node(slug or tag)` | the JSON copy of a node |
| `search(query, filters)` | the search of document 15 |
| `find_algorithm(problem, constraints)` | the "I need to…" search of 15.4 |
| `get_implementation(slug, lang)` | the code, the tests and the levels |
| `get_cases(problem, file?)` | the shared cases of a problem |
| `run_cases(slug or code, lang, cases?)` | local only: runs code against the contract in the runner image of the language (08.7) and returns the checker's verdicts. Used by agents to test generated code |
| `get_lean(slug)` | the spec, the model and the theorems |
| `reading_path(to, from?)` | the path of 14.6 |

The `/agents/` page explains all of this, with examples, as math2's page "For agents" does.

## 14.9 Hosting and deploy

**The page and file budget** (estimates at the scale of 02.4):

| Files | Count (estimate) | Size (estimate) | Where |
|---|---|---|---|
| node pages | 10,000 | 300 MB | Worker assets |
| language pages, core: ~800 algorithms and structures × 23 languages | 18,000 | 450 MB | Worker assets for tier 1, R2 for tier 2 |
| language pages, long tail: ~3,500 × 10 tier-1 languages | 35,000 | 700 MB | R2 |
| proof, Lean, visualize, tests and embed pages | 8,000 | 250 MB | Worker assets for the core, R2 for the rest |
| exercise pages | 5,000 | 50 MB | R2 |
| area, course, list, cheat sheet, dashboard pages | 1,500 | 40 MB | Worker assets |
| `.md`, `.json`, `.txt` copies | 70,000 | 600 MB | R2 |
| scene bundles and static frames (SVG) | 15,000 | 200 MB | Worker assets for bundles, R2 for frames |
| social images (JPEG, 1200×630) | 10,000 | 800 MB | R2 |
| search shards, CSS, JS, fonts, icons | 3,000 | 60 MB | Worker assets |
| **total** | **~175,000** | **~3.5 GB** | |

This does not fit one static deployment on any plan. The site uses the outlets of olymmath:

| Outlet | Files | Serves |
|---|---|---|
| **Workers static assets** on `algorithm.tamnd.com` (Workers Paid, $5/month) | the hot set, at most 95,000 files: the shell, all node pages, the core sub-pages, all list pages, scene bundles and search shards | static requests, the fastest path |
| **R2 behind the same Worker** (`algorithm-site` bucket) | every other file, at its URL path | a request that misses the assets goes to the Worker. It reads `r2://algorithm-site/<path>`, sets the headers and caches at the edge. The Worker does no templating |
| **GitHub Pages** mirror (`tamnd.github.io/algorithm`) | the core: its node pages, its tier-1 language pages, its proof, Lean and visualize pages, the list pages, search, under 900 MB | the mirror for people who cannot reach Cloudflare. A banner says which pages it leaves out and links to the main site |

**The start.** Until the site has 19,500 files, it is one Cloudflare Pages project, `algorithm`, with the custom domain `algorithm.tamnd.com`. Pages takes at most 20,000 files and 25 MiB for each file. Each push to `main` deploys to production, and each pull request from the repository deploys a preview. The move to the Worker and R2 is an issue of M2, and the URLs do not change.

**Packs on Pages.** The site passed 19,500 files at about 3,500 written nodes. Before the move to R2, `algo site deploy` packs the files that readers ask for less often: the `.md`, `.json` and `.txt` copies, the language, proof, Lean and visualize pages, and the embeds. `packs/N.pack` holds the bytes of its files one after another, and `packs/N.json` maps each path to its offset and length. The pack of a path is the FNV-1a hash of the path modulo the count of packs, and a pack has about 512 KB. A `_worker.js` serves each packed path with a range read of its pack, applies the rules of `_headers`, and gives each other path to the static files. `_routes.json` runs the worker only for the packed paths. The URLs do not change. With every node written, the tree has about 11,000 files.

`registers/site-tiers.yaml` says which files go to the assets, and `algo site check` fails when the assets pass 95,000 files or a file passes 25 MiB.

**The Worker** (in `algorithm-agent/worker/`, a few hundred lines) does:

1. serve from the static assets when the path is there;
2. else the redirects: `/t/<tag>`, the slug aliases of `registers/aliases.reg`, case folding and missing slashes, from a redirect map that the build writes (about 15,000 entries, loaded once per isolate);
3. else read R2 and cache it at the edge;
4. else the 404 page, which offers a search for the words of the path.

Headers: hashed assets (`site.3fa94c.css`) have `Cache-Control: public, max-age=31536000, immutable`; HTML has `max-age=300, stale-while-revalidate=86400`; every response has a strict Content Security Policy (scripts from the site only, no inline script except the theme line with a hash), `X-Content-Type-Options: nosniff`, `Referrer-Policy: strict-origin-when-cross-origin`, and no third-party requests at all. The site has no analytics cookies. Counts of page views come from Cloudflare's own logs, with no tracker on the page.

**Build and deploy.**

- `algo site build` writes `build/site/` and a manifest of each file with its content hash. KaTeX output, highlighted code, SubVerso exports, scene frames and social images are cached by the hash of their input, so a normal build only renders what changed.
- `algo site deploy` compares the manifest with the last deploy, uploads changed R2 objects, and deploys the Worker assets when a hot file changed.
- The build server runs `scripts/deploy-cloudflare.sh` every 5 minutes, as for math2. It does nothing when the commits of `tamnd/algorithm` and `tamnd/algorithm-agent` did not change. A change on `main` is live in 10 to 20 minutes: the tests of the changed nodes and the Lean build run first (03.5 F3).
- A full build of the site must finish in under 30 minutes on the build server. A full test of all languages is separate and nightly (09.8).
- The workflow `pages` builds the mirror on GitHub Actions on each push to `main` and each night, so the mirror can be some hours behind.
- Each release (18.4) is a separate R2 prefix, served on `https://<release>.algorithm.tamnd.com` (for example `2027-03.algorithm.tamnd.com`) by the same Worker. Its paths are the same as the main site's. `/t/<tag>@<release>` redirects there (14.18).

## 14.10 Visual design

- **Typefaces.** Inter (variable, self-hosted, subset to Latin, Latin Extended, Greek and Cyrillic) for text, as in math2. Code uses the system monospace stack (`ui-monospace, SF Mono, JetBrains Mono, Menlo, Consolas`), so no font is downloaded for it, at 0.86em with `font-variant-ligatures: none` so that `!=` and `->` look as typed. Math uses STIX Two Math, loaded after the first paint, as in math2.
- **Colour.** The tokens of math2 (`--bg`, `--surface`, `--text`, `--link`, `--accent`, …) are kept. Each top area has a hue (`--hue`), and tints are computed from it with `--tint-l` and `--tint-s`, so the light and the dark theme both work. Evidence badges use the accent colour with a tick, not traffic-light colours, so that the site does not look like a test dashboard. The dashboards (14.19) use the validated sequential palette of the dataviz skill, with a text value in each cell.
- **Code highlighting** is done at build time with Chroma, with two themes that match the site's light and dark tokens. Comments are muted but stay above 4.5:1 contrast.
- **Shapes.** No rounded corners, no borders around boxes: blocks are separated by space and by tints of the surface colour, as in math2. Theorem boxes have a 3px left bar in the area hue.
- **Math.** KaTeX renders at build time to MathML (the `katex` package, in goja, as in math2). A formula that fails is logged by the build and shown as its source in the "missing" colour; rule W07 counts these.
- **Pseudocode** has its own block style: keywords in small caps weight, line numbers in the gutter, comments in the muted colour, and the font of the text (not monospace), as in CLRS-style pseudocode, so that it reads as mathematics rather than as code.

## 14.11 Interaction

**Keyboard shortcuts.** `?` shows them all.

| Key | Does |
|---|---|
| `/` or `⌘K` / `Ctrl K` | open the search palette (15.3) |
| `g h`, `g a`, `g p`, `g t`, `g l` | go to home, areas, problems, techniques, learn |
| `[`, `]` | previous and next node in the followed course |
| `l` | next language tab; `Shift L` the previous |
| `c` | copy the code of the open tab |
| `v` | open the player full window |
| `p` | unfold all the steps of the proofs on the page |
| `t` | change the theme |
| in the player: `Space`, `←`, `→`, `Home`, `End`, `+`, `−` | play or pause, step back, step forward, start, end, faster, slower (10.5) |

Shortcuts never fire inside an input, and single-letter shortcuts can be turned off in the settings, as WCAG 2.1.4 asks.

**Preferences** are in `localStorage` only, under keys that start with `al-`: the theme, the language, the followed course, the progress, the player speed, reduced motion, and whether shortcuts are on. A settings panel in the footer can reset them.

## 14.12 Mobile

- Under 720px: one column; the table of contents folds under the head; the summary card stacks its lines; the code tab row scrolls sideways with the selected tab in view and "More" at its end.
- Code blocks scroll sideways inside themselves; the page never does. Long lines are not wrapped by default, and a toggle wraps them.
- Comparison tables keep the first column fixed and scroll the rest.
- The player switches to a stacked layout: the canvas, then the controls as large touch targets (at least 44 × 44 px), then the narration; the code panel is a sheet that slides up.
- Touch: swipe left and right on the canvas steps the player.

## 14.13 Performance budgets

`algo site check` measures each budget on a fixed sample of 60 pages (the 20 largest node pages and 40 random ones) and fails the build over budget.

| Budget | Limit |
|---|---|
| HTML of a node page, gzip | ≤ 60 KB (tier-1 code included) |
| HTML of any page, uncompressed | ≤ 600 KB |
| CSS, gzip | ≤ 20 KB, one file |
| site script, gzip | ≤ 15 KB, deferred |
| the player runtime `@algo/viz`, gzip | ≤ 60 KB, loaded when the player comes into view |
| one scene bundle, gzip | ≤ 15 KB |
| fonts before first paint | Inter text subset only, ≤ 110 KB |
| Largest Contentful Paint | ≤ 1.8 s on a mid-range phone on "Fast 4G" (Lighthouse) |
| Cumulative Layout Shift | ≤ 0.05; the player reserves its box |
| Interaction to Next Paint | ≤ 200 ms |
| Lighthouse performance score | ≥ 95 on the sample |

## 14.14 Accessibility

The target is WCAG 2.2 level AA on every page.

- Landmarks, one `h1`, headings in order, a skip link, a visible focus ring.
- Text contrast of at least 4.5:1, and 3:1 for large text and for the parts of the player's pictures that carry meaning. Meaning is never in colour alone: states in the player also use shape, pattern or a label (10.8).
- Tabs follow the ARIA tab pattern, with arrow keys. Folds are `<details>`. Sortable tables use buttons in the headers with `aria-sort`.
- Math is MathML, so screen readers read it. Pseudocode is a list of lines with numbers, not an image.
- The player has a text track: each step's narration goes to an `aria-live="polite"` region, and the current state can be read as a table (10.8). Reduced motion turns animation into steps.
- Code blocks are scroll regions with `tabindex="0"` and a label.
- `algo site check` runs axe-core on the sample of 14.13 in a headless browser, and a finding of "serious" or "critical" fails the build (rule W08). A person tests the core page types with VoiceOver and NVDA at each milestone.

## 14.15 SEO and social cards

- **Titles.** Node pages: "Merge sort: algorithm, proof and code · algorithm". Language pages: "Merge sort in Rust · algorithm". Problem pages: "Sorting: the problem and every algorithm for it · algorithm".
- **Descriptions** come from `summary`, at most 160 characters.
- **Canonical** links on every page. Language pages are canonical to themselves; the `.md` and `.json` copies are not indexed (`X-Robots-Tag: noindex` on them).
- **Structured data** (JSON-LD): `TechArticle` for node pages, with `about`, `author` (people and, for agent text, the project as the publisher), `dateModified` and `isPartOf`; `SoftwareSourceCode` for language pages, with `programmingLanguage`, `codeRepository` and `license`; `BreadcrumbList` for the breadcrumb; `Course` for courses; `Person` for people; `Dataset` for `/api/`.
- **Sitemaps**: `/sitemap.xml` is an index of sitemaps by kind and by language, each under 50,000 URLs, with `lastmod` from the provenance log.
- **Social images**: one 1200 × 630 JPEG per node page, made by the build: the name, the kind, the area hue, the bound, and a frame of the scene when the node has one (10.7). Language pages reuse the image of their node with the language name added.
- `robots.txt` allows everything but `/embed/` and `/compare/`, and names the sitemap.

## 14.16 Rendering

- The renderer is the `site` package of `algo`, in Go, with `html/template`, as in math2-agent. It writes plain HTML. No client framework.
- Markdown is goldmark with the extensions of `markup` (03.4): math, pseudocode blocks, `[[slug]]` links, theorem and proof blocks, and folds.
- Script is in small islands, each in the one site script or loaded on demand: the theme, tabs that remember, copy buttons, the search palette, table sort and filter, course progress, the player, the compare page and the Lean hovers.
- Each template has a golden test: a fixed node directory renders to a fixed HTML file, as in math2-agent's tests.

## 14.17 Languages of the site

English first. Through M4 the site is in English only. Aliases in other languages are already in the front matter (04.2), so a search for "Алгоритм Дейкстры" or "迪杰斯特拉算法" finds the page (15.2).

After M4, translations can come in as a layer `i18n/<lang>/nodes/...` that holds translated `node.md` and proof files only. Code, tests, Lean and scenes are shared. Translated pages live at `/<lang>/<slug>/` with `hreflang` links, keep the English slug, and show "Translated from the English page of <date>" with a link to the diff when the English page changed since. Document 20 has the question of which languages first.

## 14.18 Releases and the offline archive

Each monthly release (18.4) builds the site of its commit:

- served at `https://<release>.algorithm.tamnd.com`, with the same paths as the main site;
- archived as `algorithm-<release>-site.zip`, attached to the GitHub release: the core pages, their tier-1 code, the scenes and search, made to open from disk with `index.html`. Search reads an embedded index (`window.algoSearch`) instead of fetching JSON, as math2's archive does. The player works offline because its bundles are local files.

## 14.19 Dashboards

| Page | Shows |
|---|---|
| `/coverage/` | the coverage levels C0–C6 of document 16, overall and per area, as a grid of areas × levels with a percentage in each cell; a click opens the gap list of that cell |
| `/matrix/` | the test matrix: one row per algorithm or data structure, one column per language, each cell the I-level as a shade and a digit; filters by area and level; the totals per language; links to the failing runs |
| `/gaps/` | the ranked gap list from the work queue (12.4): what is missing and its value, with "Help write it" links that open a ready-made issue or a `algo new` command |
| `/languages/<lang>/` | the toolchain pin, the share of nodes with I2 or better, the oldest unreviewed files |
| `/latest/` | new and changed pages, new levels ("Red–black tree insertion reached L3"), with an Atom feed |
| `/reviewers/` | the people who review, with their counts, as in math2 |
| `/universe/` | the universe register, searchable, each item with its outcome (16.1) |

The dashboards are built from the same `build/` data as the pages, and each has its JSON copy under `/api/`.
