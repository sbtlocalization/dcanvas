---
status: In progress
priority:
blocked by:
kind: PRD
---
# Tasks

<!-- live the following block unchanged -->
```base
summaries: {}
filters:
  and:
    - epic == [this.file]
formulas:
  is_blocked: |-
    note["blocked by"].reduce(
      acc || (
        value.asFile().properties.status != "Done" && 
        value.asFile().properties.status != "Obsolete"
      ),
      false
    )
  progress_icon: |-
    if(status == "Done",
      icon("square-check-big"),
      if (status == "Obsolete",
        icon("square-arrow-right"),
        if(formula.is_blocked,
          icon("construction"), 
          icon("square")
        )
      )
    )
  progress_sort: |-
    if(status == "Done" || status == "Obsolete", 
      20, 
      if(formula.is_blocked, 
        10, 
        if(status == "In progress", 
          5, 
          0)))
  progress_string: |-
    if(status == "Done" || status == "Obsolete", 
      "Done", 
      if(formula.is_blocked, 
        "On hold",
        "Ready"))
  priority: if(priority, icon("tally-" + priority), null)
  kind_icon: |-
    if(kind == "Bug",
      icon('bug'),
      ''
    )
properties:
  note.status:
    displayName: status
  file.name:
    displayName: task
  formula.progress_icon:
    displayName: " "
  formula.progress_string:
    displayName: Tasks that are
  formula.kind_icon:
    displayName: " "
views:
  - type: table
    name: Таблиця
    groupBy:
      property: formula.progress_string
      direction: DESC
    order:
      - formula.progress_icon
      - formula.kind_icon
      - file.name
      - formula.priority
      - status
      - blocked by
      - mode
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
      formula.kind_icon: -1
      file.name: 435
      note.status: 125
      note.blocked by: 627

```
<!-- end unchanged block -->

# Problem Statement

The dCanvas format started life as a backward-compatible superset of JSON Canvas for Infinity Engine dialogues. Trying to reuse it in a second project revealed the format doesn't fit: engine-specific fields (journal text/sound) are baked into the core, the node role is a fixed `state`/`transition` enum that uses one engine's vocabulary, and there is no contract that unknown fields survive a round-trip. The design has since been reworked into **dCanvas 3.0** (spec, schema, CONTEXT, and ADRs already written in `dcanvas/docs/`), but there is no code yet: the Go types still describe 2.0, layout is tangled inside Infinity-specific conversion, and `sbt-infinity` still emits 2.0.

We need a **reference implementation** of dCanvas 3.0 — a clean, dialogue-only, extensible library that other projects can depend on — and we need to prove it fits real use by migrating `sbt-infinity` onto it.

# Solution

Build the dCanvas 3.0 reference library as a self-contained package (staged inside `dcanvas/` so it can later be lifted into its own repository per [[ADR-0008 - Extract format and library to a separate repository|ADR-0008]]), then port `sbt-infinity`'s dialogue export onto it as the format's acceptance test.

The library provides three capabilities, mapped to the three layers of the format:

1. **Format types & IO** — the JSON Canvas core (Layer 0) plus the dialogue vocabulary (Layer 1) as typed fields, with `Encode`/`Decode` that validate the major version and **preserve every unrecognised field recursively** (so a Layer 2 project's data is never lost).
2. **Layout** — a generic auto-positioning operation over a canvas, using `autog`, reading and writing only Layer 0 geometry, resolving cycles on its own.
3. **A clean extension seam** — projects attach their own `x-` fields (Layer 2) and their own `x-role` vocabulary without the library knowing about them.

`sbt-infinity` keeps its engine-specific conversion (NPC colours, characters, triggers, journal fields) but builds dCanvas 3.0 canvases, calls the library's layout, and writes `.dcanvas` files that are valid 3.0 and round-trip-safe.

# User Stories

1. As a developer of another dialogue project, I want to depend on the dCanvas library for reading and writing the format, so that I don't reimplement JSON Canvas parsing and layout.
2. As a developer of another dialogue project, I want to attach my own `x-` fields to nodes and edges, so that I can store engine-specific metadata the library doesn't know about.
3. As a developer of another dialogue project, I want my custom `x-role` values to pass through untouched, so that I'm not forced into Infinity's `state`/`transition` vocabulary.
4. As a consumer of the format, I want unrecognised fields to survive a read-then-write cycle, so that re-saving a file produced by a newer or different project never destroys data.
5. As a consumer of the format, I want unrecognised keys inside `x-character` to survive too, so that a project can extend the speaker object without losing data when a generic tool re-saves.
6. As a consumer of the format, I want decoding a non-3.0 (e.g. 2.0) file to fail loudly, so that I never get a silently half-interpreted document.
7. As a consumer of the format, I want the typed dialogue fields (`x-kind`, `x-role`, `x-character`, `x-condition`, `x-action`, `x-sound`, `x-textId`, `x-id`) available as first-class fields, so that I work with them statically rather than through a map.
8. As an author of a dialogue viewer, I want `x-kind` to carry a closed, known set of values (`line`/`reply`, `normal`/`loop`), so that I can style nodes and edges without understanding any project's vocabulary.
9. As an author of a dialogue viewer, I want to render the visible text from standard JSON Canvas fields (`node.text`, `edge.label`), so that my generic renderer works without dialogue-specific knowledge.
10. As an Obsidian user, I want to open a `.dcanvas` file as a plain JSON Canvas and re-save it, so that I can inspect and tweak a dialogue without losing its `x-` metadata.
11. As an `sbt-infinity` maintainer, I want `dialog ex` to emit dCanvas 3.0 files, so that the tool's output matches the current format.
12. As an `sbt-infinity` maintainer, I want the Infinity journal fields emitted as Layer 2 extensions, so that they no longer pollute the format core but still appear in the output.
13. As an `sbt-infinity` maintainer, I want the engine-specific conversion (colours, characters, triggers, sound prefixes) to stay in this project, so that the library remains domain-agnostic.
14. As an `sbt-infinity` maintainer, I want node auto-positioning to come from the library's layout, so that positioning logic is no longer tangled with conversion.
15. As an `sbt-infinity` maintainer, I want the player's choice text on the edge `label` and the spoken text on the node `text`, so that exported files follow the text-placement rule and render correctly in Obsidian.
16. As an `sbt-infinity` maintainer, I want each node to carry both a canvas `id` and an engine `x-id`, so that edges reference a stable file-local id while the engine identifier is preserved.
17. As a maintainer, I want the library structured so it can be lifted into its own repository with minimal change, so that the eventual extraction ([[ADR-0008 - Extract format and library to a separate repository|ADR-0008]]) is cheap.
18. As a maintainer, I want a dialogue graph with cycles to lay out without crashing, so that looping conversations are positioned rather than falling back to default coordinates.
19. As a maintainer, I want laid-out nodes to not overlap, so that the exported canvas is readable.
20. As a maintainer, I want the library to have no dependency on the Infinity domain packages, so that the dependency direction is one-way (project → library).

# Implementation Decisions

- **Package shape.** The dCanvas 3.0 reference library is a self-contained package with no dependency on `dialog`, `parser`, or other Infinity domain code (dependency direction is project → library only). It is staged inside `dcanvas/` and structured for later extraction to a standalone repository ([[ADR-0008 - Extract format and library to a separate repository|ADR-0008]]).
- **Format types.** Model Layer 0 (JSON Canvas core) and Layer 1 (dialogue vocabulary) as typed fields. Node Layer 1: `x-id`, `x-kind` (`line`/`reply`), `x-role` (free string), `x-textId`, `x-condition`, `x-action`, `x-sound`, `x-character`. Edge Layer 1: `x-id`, `x-kind` (`normal`/`loop`), `x-role` (free string), `x-condition`, `x-textId`. All `x-` fields optional.
- **Preservation mechanism.** Per [[ADR-0005 - Reference impl uses catch-all over typed fields|ADR-0005]], do **not** use struct embedding. Decode into a raw map, populate typed known fields, retain unknown keys in a catch-all; on encode, merge typed fields with the preserved map. Preservation is recursive for known nested objects (currently only `x-character`).
- **Version handling.** `Decode` validates the major version of `x-dCanvasVersion` and errors on mismatch (a 2.0 file is rejected). `Encode` always stamps `"3.0"`. No 2.0 migration code (2.0 is deprecated).
- **Layout.** Per ~~[[ADR-0006 - Layout belongs to the dCanvas library|ADR-0006]]~~ [[ADR-0009 - Layout may read x-kind and the loop strategy is configurable|ADR-0009]], expose a generic layout operation over a canvas (`width`/`height` → `x`/`y`) driven by edges. It uses `autog`, wraps the call in panic recovery (degrading to default positions on failure). ~~It reads/writes only Layer 0 geometry, and does **not** exclude `loop` edges — the layout engine resolves cycles itself.~~ By default it **excludes** `x-kind: loop` edges from positioning (reading Layer 1) so cyclic dialogues lay out as a clean tree; configurable via `WithLoopStrategy` (default `LoopCut`).
- **Constructor.** Provide a `Canvas` constructor that stamps the format version and initialises non-nil node/edge slices (the one real invariant worth a constructor). Nodes/edges/character are plain typed values, built via literals.
- **`sbt-infinity` port.** The `dialog` package keeps `ToDCanvas` and its `newNode`/`newEdge`/`newCharacter` helpers, but: builds 3.0 nodes/edges; maps NPC line → `x-kind: line`, player reply → `x-kind: reply`; sets the engine-specific term in `x-role`; puts journal fields as Layer 2 `x-` fields; sets canvas `id` and `x-id` distinctly; puts player-facing choice text on edge `label` and spoken text on `node.text`; and delegates positioning to the library's layout instead of calling `autog` inline.
- **CLI.** `dialog ex` behaviour is unchanged from the user's perspective except that output is dCanvas 3.0.

# Testing Decisions

- **What makes a good test here:** assert externally observable format behaviour — bytes in, bytes out, and the structure of the decoded canvas — never the internal catch-all representation. Prefer table-driven Go tests, matching the style already in `dcanvas_test.go`.
- **IO / round-trip seam (`Decode`/`Encode`, the highest and existing seam):**
  - Round-trip preserves an unknown top-level field, an unknown node `x-` field, an unknown edge `x-` field, and an unknown key inside `x-character` (recursive preservation, the core guarantee — [[ADR-0003 - Mandatory recursive preservation of unknown fields|ADR-0003]]).
  - Typed Layer 1 fields decode and re-encode faithfully.
  - Decoding a document whose major version ≠ 3 returns an error; decoding a `"3.0"` document succeeds.
  - A stripped document (all `x-` removed) is still valid JSON Canvas (encode produces no `x-` for an empty Layer 1).
- **Layout seam (new `Layout` function, highest point for positioning):**
  - A multi-node graph gets non-default, non-overlapping coordinates — reuse the existing `HasOverlappingNodes` assertion.
  - A graph containing a cycle lays out without panicking.
  - An empty graph is a no-op.
- **Conversion seam (`dialog.ToDCanvas`, existing):** a representative dialogue converts to a 3.0 canvas with correct `x-kind`/`x-role`, journal fields present as Layer 2, choice text on the edge `label`, and distinct `id`/`x-id`.
- **End-to-end seam (`dialog ex` CLI, existing):** exporting a fixture DLG produces a file that the library's `Decode` accepts as valid 3.0 and that round-trips without loss.
- **Prior art:** `dcanvas_test.go` (geometry/overlap tests) is the model for library-level tests; existing `cmd/dialog` tests (if any) for the CLI seam.

# Out of Scope

- Creating the standalone repository and actually moving the code out of `sbt-infinity` — deferred until a second consumer exists ([[ADR-0008 - Extract format and library to a separate repository|ADR-0008]]). This epic only structures the code for that move.
- Any dCanvas 2.0 → 3.0 migration tooling (2.0 is deprecated).
- First-class `group`-node support / sub-dialogue clustering ([[ADR-0007 - Dialogue semantics only on text nodes|ADR-0007]] keeps it opaque for now).
- Building or reworking the dialogue **viewer** (a separate effort; the format is designed to enable a flexible viewer, but the viewer itself is not part of this epic).
- Changing the dialogue domain model, DLG parsing, or any non-dialogue command.

# Further Notes

- The format design is already settled and documented: `dcanvas/docs/dcanvas-3.0.md`, `dcanvas-3.0.schema.json`, `CONTEXT.md`, and ADRs `0001`–`0008`. This epic is the implementation of that design.
- Natural delivery order: (1) format types + IO with preservation and version validation; (2) layout; (3) `sbt-infinity` port and CLI output switch. The port is the acceptance test that validates the format against real use.
- The schema `$id` still points at the old `sbt-dialog-viewer` host; fix it when the code is extracted to its own repository.
