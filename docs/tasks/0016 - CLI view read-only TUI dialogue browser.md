---
epic: "[[EPIC-0002 - dCanvas CLI TUI and round-trip library robustness]]"
parent:
status: To do
priority: 1
blocked by:
  - "[[0011 - CLI walking skeleton lay out a file end-to-end]]"
kind: Task
mode: AFK
---

# Sub-tasks

<!-- live the following block unchanged -->
```base
summaries: {}
filters:
  and:
    - parent == [this.file]
formulas:
  is_blocked: |-
    note["blocked by"].reduce(
      value.asFile().properties.status != 'Done' || acc,
      false
    )
  progress_icon: |-
    if(status == "Done",
      icon('square-check-big'),
      if(formula.is_blocked,
        icon('construction'), 
        icon('square')
      )
    )
  progress_sort: |-
    if(status == "Done", 
      20, 
      if(formula.is_blocked, 
        10, 
        if(status == 'In progress', 
          5, 
          0)))
  progress_string: |-
    if(status == "Done", 
      "Done", 
      if(formula.is_blocked, 
        "On hold",
        "Ready"))
  priority: if(priority, icon('tally-' + priority), null)
properties:
  note.status:
    displayName: status
  file.name:
    displayName: task
  formula.progress_icon:
    displayName: " "
  formula.progress_string:
    displayName: Tasks that are
views:
  - type: table
    name: Таблиця
    groupBy:
      property: formula.progress_string
      direction: DESC
    order:
      - formula.progress_icon
      - file.name
      - formula.priority
      - status
      - blocked by
      - epic
    sort:
      - property: formula.progress_sort
        direction: ASC
      - property: priority
        direction: DESC
      - property: file.basename
        direction: ASC
    summaries: {}
    columnSize:
      formula.progress_icon: -56
      note.status: 125
      note.blocked by: 332

```
<!-- end unchanged block -->

# What to build

A new subcommand, `dcanvas view <input>`, that opens a **read-only** interactive TUI to browse a dCanvas dialogue as a navigable, collapsible tree of connected utterances.

`view` is for *reading* the dialogue — following who says what and how choices branch — not for previewing layout. It therefore **does not run layout and never writes anything**: layout stays the job of the non-interactive `dcanvas layout` command. This deliberately replaces the original "preview the laid-out graph, confirm the write" framing of this task: a tree derived from the dialogue's topology would be identical before and after layout (layout only moves coordinates), so it previews nothing about what layout did. The valuable feature is a legible, navigable *content* browser, and that is what this task builds.

The TUI framework is **Bubble Tea** (`github.com/charmbracelet/bubbletea`) with lipgloss for styling and `bubbles/viewport` for scrolling.

# Repository restructure (prerequisite — see [[ADR-0012 - The reference CLI is a separate Go module|ADR-0012]])

Adding a second CLI-only framework (Bubble Tea) to the shared `go.mod` is the trigger to split the repo into two Go modules. Do this first:

- **Library module** — repo root, `github.com/sbtlocalization/dcanvas`: the core `dcanvas` package (stdlib only) **and** the `layout` package (keeps its `autog` dependency). Unchanged in content; it just stops carrying CLI dependencies.
- **CLI module** — `cmd/dcanvas/`, its own `go.mod` with module path `github.com/sbtlocalization/dcanvas/cmd/dcanvas`, requiring the library module plus cobra and Bubble Tea. `cobra` moves out of the root `go.mod` into here. The library's test-only `jsonschema` dependency stays in the root module.
- **Binding** — a committed `go.work` at the repo root listing `.` and `./cmd/dcanvas`, so in-repo builds resolve the library locally without a `replace` directive.
- **CI** — a single job that vets and tests each module explicitly (a root `./...` does not descend into the nested module):
  - `go vet -C . ./...` and `go test -C . ./...`
  - `go vet -C cmd/dcanvas ./...` and `go test -C cmd/dcanvas ./...`

The `view` command imports only the `dcanvas` package (types + `Decode`); it does **not** import `layout`.

# The `view` command

- A cobra subcommand `view <input>` (`cobra.ExactArgs(1)`). Reuse the existing `checkExt` (`.d.canvas` / `.dcanvas`); take a file argument only (no stdin — stdin belongs to the TUI for key input).
- Read the file with `dcanvas.Decode`. Build the tree **best-effort**: do not run `Validate` (a viewer is not a gate), and silently ignore any edge whose endpoint node is missing or is not a `text` node.
- **Require an interactive terminal.** Check stdin+stdout with `term.IsTerminal`; if either is not a TTY, exit with a clear error (e.g. "`view` requires an interactive terminal") and a non-zero code. Do not start Bubble Tea in that case.

# Tree model

The dialogue is a DAG (fan-in and loops), rendered as a walkable tree:

- **Roots** — `text` nodes with no incoming *forward* (normal) edge. Support several (render a forest); a well-formed dialogue usually has one. Loop edges (`x-kind: loop`, `dcanvas.KindLoop`) do **not** count as an incoming edge for root-finding.
- **Node visibility** — show only `type == "text"` nodes. Hide `group` / `file` / `link` nodes (opaque to dialogue per [[ADR-0007 - Dialogue semantics only on text nodes|ADR-0007]]). An isolated `text` node (no edges) is its own single-node tree.
- **Children** — from normal (forward) edges. **Sibling order = order of edges in the canvas `edges` array** (deterministic, independent of X/Y, reflects authoring order).
- **Repeat occurrences (loops and DAG fan-in), one mechanism** — a node is expanded on its **first** occurrence (DFS pre-order from the roots, following the sibling order above). Every later occurrence — whether reached by a loop back-edge or by a second forward parent (merge) — renders as a non-expandable **jump line** `↩ … → <target>`. For a loop edge, the edge's `label` (e.g. `"(she calls you back)"`) is the jump line's caption, since it is the only content there.
- **Cycle fallback** — if no node qualifies as a root (every node has an incoming forward edge, i.e. a fully cyclic graph), take the first node in the `nodes` array as the root so the tree is never empty.

# UI and interaction

- **Responsive master–detail layout.** When the terminal width is ≥ ~90 columns, split horizontally: tree on the left (~60%), detail pane on the right (~40%). Below that, stack vertically: tree on top, detail pane at the bottom. Recompute on `tea.WindowSizeMsg`. (The 90-column threshold is a starting value, tune on real data.)
- **Initial state — fully expanded** (this is a reading tool; the whole flow should be visible, with collapse available for long branches).
- **Tree lines** carry: an expand/collapse indicator for branch nodes; a marker distinguishing a `line` node (NPC utterance — show the speaker's `Character.Name` when present, then the text) from a `reply` node (player utterance); and the node's **truncated** `text`. Jump lines render as described above. Exact glyphs/styling are the implementer's choice; the information above is the requirement.
- **Detail pane** shows, for the selected node: the **full** (untruncated) `text`; speaker (`Character.Name`, plus `Portrait`/`Gender` when present); `x-kind` and `x-role`; `x-condition`, `x-action`, `x-sound` when present; and the `label` of the edge that led to this node, when non-empty.
- **Keys (arrow-based; no vim keys):**
  - `↑` / `↓` — move the cursor across visible lines.
  - `→` — expand a collapsed branch; if already expanded, move into its first child.
  - `←` — collapse an expanded branch; if already collapsed or a leaf, move to the parent.
  - `Enter` — context-sensitive: on a branch node it toggles expand/collapse (alias of `→`/`←`); on a jump line it **jumps** to the target node's canonical (first) occurrence.
  - `Backspace` — pop the **jump history stack**: return to where the last jump started. A jump pushes the origin onto the stack; `Backspace` with an empty stack does nothing.
  - `Home` / `End` (with `Ctrl+Home` / `Ctrl+End` as aliases on terminals that require a modifier) — first / last line.
  - `Esc` (also `q`, `Ctrl+C`) — quit.
  - A jump must **auto-expand the target's ancestors** and scroll it into view, so the cursor never lands on a hidden line.

# Status bar (Turbo Vision style, Ukrainian labels)

A bottom status bar in the Turbo Vision idiom: a solid-background strip with each hotkey highlighted and its label beside it, styled with lipgloss. It is **context-sensitive** — it shows only the action applicable to the currently selected line, and `Enter` is never shown next to the arrow that does the same thing:

- Collapsed branch node: `↑↓ Рух   → Розгорнути   Home Початок   End Кінець   Esc Вихід`
- Expanded branch node: `↑↓ Рух   ← Згорнути   Home Початок   End Кінець   Esc Вихід`
- Leaf (no children): `↑↓ Рух   Home Початок   End Кінець   Esc Вихід`
- Jump line: `↑↓ Рух   ⏎ Стрибнути   Home Початок   End Кінець   Esc Вихід`
- Append `⌫ Назад` in any state **only when the jump history stack is non-empty**.

# Testing (deliberately light)

This command is intentionally small — a first outing with Bubble Tea — so keep tests minimal.

- Structure the pure logic apart from the Bubble Tea `Model`: a `build(*dcanvas.Canvas) → tree` function (roots, ordered children, jump lines, hidden non-text nodes) and a `visibleLines(tree, expandedState) → []line` flattening function. `Model` wraps state + viewport + size and delegates to these.
- Write **one** sanity unit test on `build` using `cmd/dcanvas/testdata/` fixtures or `example.d.canvas`: assert the root is `n-greet` and that both loop edges become jump lines.
- **No** `teatest`, **no** `Update`-transition tests, **no** new fixtures.

# Acceptance criteria

- [ ] The repo is two Go modules (library at root, CLI at `cmd/dcanvas/`) bound by a committed `go.work`; `cobra` is in the CLI module's `go.mod`, not the root's.
- [ ] `dcanvas view <file.d.canvas>` opens a TUI showing the dialogue as a collapsible tree of `text` nodes; edges appear as parent→child nesting.
- [ ] Loop edges and DAG merges render as jump lines; `Enter` on a jump line moves to the target and `Backspace` returns.
- [ ] Arrow navigation, expand/collapse, `Home`/`End`, and quit all work; the layout adapts to terminal width; the status bar is context-sensitive with the Ukrainian labels above.
- [ ] `view` never writes to disk and never runs layout.
- [ ] Running `view` without a TTY exits with a clear error, not a crash.
- [ ] `go test -C . ./...` and `go test -C cmd/dcanvas ./...` both pass in CI.

# Out of scope

- Any writing, layout, or editing of the file (read-only browser).
- `group` / `file` / `link` nodes (hidden).
- A pluggable render/theme system, a help overlay, `teatest` end-to-end tests.
- Semver release tagging of the two modules (deferred as before).
