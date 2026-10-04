# dCanvas — Context

The ubiquitous language and domain model behind the dCanvas format. Read this before the spec (`spec/dCanvas-1.1.md`); the spec is the *what*, this is the *why* and the *vocabulary*.

## What dCanvas is

dCanvas (**d**ialogue **canvas**) is a file format for **dialogue graphs** in games — who says what, and how the conversation branches. It is a **strict superset of JSON Canvas 1.0**, so any JSON Canvas editor (notably Obsidian Canvas) can open, view, and re-save a dCanvas file without data loss.

It is **dialogue-only by design**. It is not a generic diagram or flowchart format. The reference implementation also owns graph **layout** (auto-positioning of nodes).

## The three layers

Every field belongs to exactly one layer. This is the spine of the whole design.

| Layer | Owner | Examples | A generic dCanvas tool… |
|---|---|---|---|
| **0 — JSON Canvas** | the JSON Canvas standard | `id`, `x`, `y`, `text`, edge `label` | …fully understands |
| **1 — dialogue vocabulary** | the dCanvas spec | `d-kind`, `d-role`, `d-character`, `d-condition`, … | …fully understands |
| **2 — project extensions** | one game/engine | Infinity's `x-journalText`, `x-journalSound`, … | …does **not** understand, but **preserves** |

The format problem that triggered the current design: the old internal format (0.2) baked engine-specific fields (e.g. journal text) into the core, so reusing it for another project didn't fit. Layer 2 + mandatory preservation solves that.

## Extension namespaces

Two prefixes partition the extension fields ([[ADR-0014 - Namespace split d- for the format x- for projects|ADR-0014]]):

- **`d-`** — belongs to the spec. The defined `d-` fields are Layer 1; any *other* `d-` field is reserved for future format versions and must never carry a project meaning.
- **`x-`** — the conventional home of Layer 2 project extensions.

Preservation is prefix-blind: every unknown field survives a round-trip, whatever it is called.

## Glossary

- **Dialogue graph** — the directed graph of a single conversation. One dCanvas file = one graph.
- **Node** — a unit of spoken content. Always a JSON Canvas `text` node when it carries dialogue.
- **Line** (`d-kind: "line"`) — an utterance spoken *to* the player: NPC, narrator, object. Not necessarily an "NPC" — hence not called that.
- **Reply** (`d-kind: "reply"`) — the player's *own* utterance. **A reply is always a node.**
- **Choice** — what the player *sees and picks* from a menu. It rides on the **edge `label`**, not on a node. A choice is either a shortened paraphrase of the target reply, or coincides with it.
  - Consequence: "is the player choice a node or an edge?" has no single answer across projects. The *reply* is always a node; the *choice* is edge text. This is why text placement is split between `node.text` and `edge.label`.
- **`d-kind`** — a **closed** vocabulary, understood by every tool. Nodes: `line` / `reply` (rendering hints only). Edges: `normal` / `loop`. `loop` is largely a rendering hint, but layout also reads it: by default the library **excludes `loop` (back-)edges from positioning** so a cyclic dialogue still lays out as a clean top-down tree (configurable — see [[ADR-0009 - Layout may read x-kind and the loop strategy is configurable|ADR-0009]]).
- **`d-role`** — an **open** free-string label, defined per project (e.g. `state`, `transition`, `paraphrase`). Generic tools ignore it; a viewer may map `role → style`.
- **Canvas id vs domain id** — two distinct identifiers:
  - `id` — unique *within the file*; edges reference it. The "canvas" id.
  - `d-id` — the engine/domain identifier (e.g. `ABELA[5]`). Its value format is a project convention, not part of the format.
- **`d-character`** — speaker info (`name`, `textId`, `portrait`, `gender`), an **open object**; projects may add keys, preserved recursively. An absent or empty `portrait` means "no portrait". `textId` (since 1.1) is the string-table reference for `name`; like `d-textId`, its value format is a project convention.
- **`d-condition`** — engine condition/trigger gating a node or transition.
- **`d-action`** — engine action executed at a node.
- **`d-textId`** — string-table reference for the visible text (`node.text` or `edge.label`). Value format is a project convention.

## Consumers

The format is consumed across two languages with a one-directional flow (writer → file → reader).

- **Writer** — a Go tool that *generates* `.d.canvas` files via the Go reference library. Write-only.
  _Avoid_: exporter, producer.
- **Reader** — a TypeScript + Svelte tool that *opens and displays* `.d.canvas` files. Read-only; consumes the JSON Schema, never the Go library.
  _Avoid_: viewer, consumer (ambiguous — every tool is a "consumer" of the format).
- **CLI** — a Go command-line/TUI tool (in `cmd/`) that *operates* on `.d.canvas` files: reads a file, performs an operation (first: auto-layout), and writes it back. The only tool that round-trips (read → operate → write). See [[ADR-0011 - An in-repo reference CLI is a round-tripping consumer and the layout engine stays swappable|ADR-0011]].
- **Master spec** — this repository: the canonical source of the format (spec, JSON Schema, Go reference library). The cross-language contract is the **JSON Schema**, not the library (see [[ADR-0010 - The JSON Schema is the cross-language contract enforced by writer conformance|ADR-0010]]).

## Invariants

1. **Visible text lives in standard fields.** Character/player speech → `node.text`; player choice → `edge.label`. Never encode metadata inside `text`. This is what keeps Obsidian rendering correct.
2. **Preserve everything unknown, recursively.** Read → write must not drop any unrecognised field at any depth. (Same behaviour Obsidian already has for unknown JSON Canvas properties.)
3. **Closed vs open classification.** `d-kind` is closed (format owns the values); `d-role` is open (project owns the values).

## Status & roadmap

- **dCanvas 1.0 is the first published version of the format** ([[ADR-0013 - Public versioning restarts at 1.0 and internal specs are archived as 0.x|ADR-0013]]). The earlier internal drafts were never published and are archived in `docs/` as **0.1** and **0.2** (historically numbered 1.0 and 2.0). 1.0 is a clean break: a reader knows only `d-version` with major 1 and rejects anything else, including files stamped with the pre-1.0 internal version field — no migration code.
- **dCanvas 1.1 is the current version.** It only adds fields to 1.0 (the first is `d-character.textId`), so every 1.0 reader keeps working. Spec and schema stay one file per minor ([[ADR-0015 - The schema constrains d-version by major while schema files stay per minor|ADR-0015]]): `spec/dCanvas-1.0.*` keeps describing 1.0, `spec/dCanvas-1.1.*` describes 1.1.
- The format lives in its **own repository** — this one is the canonical master spec ([[ADR-0008 - Extract format and library to a separate repository|ADR-0008]]'s "Later" stage, fulfilled by [[ADR-0010 - The JSON Schema is the cross-language contract enforced by writer conformance|ADR-0010]]). The second consumer that triggered the move is the TypeScript reader.
- A Go writer and a TS+Svelte reader are the two live consumers; neither round-trips. Unknown-field preservation is exercised only via the Obsidian round-trip path.
