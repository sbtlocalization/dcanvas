# dCanvas — Context

The ubiquitous language and domain model behind the dCanvas format. Read this before the spec (`dcanvas-3.0.md`); the spec is the *what*, this is the *why* and the *vocabulary*.

## What dCanvas is

dCanvas (**d**ialogue **canvas**) is a file format for **dialogue graphs** in games — who says what, and how the conversation branches. It is a **strict superset of JSON Canvas 1.0**, so any JSON Canvas editor (notably Obsidian Canvas) can open, view, and re-save a dCanvas file without data loss.

It is **dialogue-only by design**. It is not a generic diagram or flowchart format. The reference implementation also owns graph **layout** (auto-positioning of nodes).

## The three layers

Every field belongs to exactly one layer. This is the spine of the whole design.

| Layer | Owner | Examples | A generic dCanvas tool… |
|---|---|---|---|
| **0 — JSON Canvas** | the JSON Canvas standard | `id`, `x`, `y`, `text`, edge `label` | …fully understands |
| **1 — dialogue vocabulary** | the dCanvas spec | `x-kind`, `x-role`, `x-character`, `x-condition`, … | …fully understands |
| **2 — project extensions** | one game/engine | Infinity's `x-journalText`, `x-journalSound`, … | …does **not** understand, but **preserves** |

The format problem that triggered the 3.0 redesign: the old format baked engine-specific fields (e.g. journal text) into the core, so reusing it for another project didn't fit. Layer 2 + mandatory preservation solves that.

## Glossary

- **Dialogue graph** — the directed graph of a single conversation. One dCanvas file = one graph.
- **Node** — a unit of spoken content. Always a JSON Canvas `text` node when it carries dialogue.
- **Line** (`x-kind: "line"`) — an utterance spoken *to* the player: NPC, narrator, object. Not necessarily an "NPC" — hence not called that.
- **Reply** (`x-kind: "reply"`) — the player's *own* utterance. **A reply is always a node.**
- **Choice** — what the player *sees and picks* from a menu. It rides on the **edge `label`**, not on a node. A choice is either a shortened paraphrase of the target reply, or coincides with it.
  - Consequence: "is the player choice a node or an edge?" has no single answer across projects. The *reply* is always a node; the *choice* is edge text. This is why text placement is split between `node.text` and `edge.label`.
- **`x-kind`** — a **closed** vocabulary, understood by every tool. Nodes: `line` / `reply` (rendering hints only). Edges: `normal` / `loop`. `loop` is largely a rendering hint, but layout also reads it: by default the library **excludes `loop` (back-)edges from positioning** so a cyclic dialogue still lays out as a clean top-down tree (configurable — see [[ADR-0009 - Layout may read x-kind and the loop strategy is configurable|ADR-0009]]).
- **`x-role`** — an **open** free-string label, defined per project (e.g. `state`, `transition`, `paraphrase`). Generic tools ignore it; a viewer may map `role → style`.
- **Canvas id vs domain id** — two distinct identifiers:
  - `id` — unique *within the file*; edges reference it. The "canvas" id.
  - `x-id` — the engine/domain identifier (e.g. `ABELA[5]`).
- **`x-character`** — speaker info (`name`, `portrait`, `gender`), an **open object**; projects may add keys, preserved recursively.
- **`x-condition`** — engine condition/trigger gating a node or transition.
- **`x-action`** — engine action executed at a node.
- **`x-textId`** — string-table reference for the visible text (`node.text` or `edge.label`).

## Invariants

1. **Visible text lives in standard fields.** Character/player speech → `node.text`; player choice → `edge.label`. Never encode metadata inside `text`. This is what keeps Obsidian rendering correct.
2. **Preserve everything unknown, recursively.** Read → write must not drop any unrecognised field at any depth. (Same behaviour Obsidian already has for unknown JSON Canvas properties.)
3. **Closed vs open classification.** `x-kind` is closed (format owns the values); `x-role` is open (project owns the values).

## Status & roadmap

- **dCanvas 2.0 is deprecated.** 3.0 is a clean redesign; no migration code.
- The format and its Go **reference implementation** will move to a **separate repository** when a second consumer exists to validate it (see [[ADR-0008 - Extract format and library to a separate repository|ADR-0008]]).
- This `sbt-infinity` tool still emits 2.0 today; porting it to 3.0 (the natural acceptance test of the format) is planned but not yet done. Until then this `dcanvas/` package and its docs are the staging area.
