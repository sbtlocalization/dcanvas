---
epic:
parent:
status: Done
priority: 3
blocked by:
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

The format goes public as **dCanvas 1.0**. The internal 3.0 was never published, so the public history starts clean: the spec and JSON Schema become `spec/dCanvas-1.0.md` and `spec/dCanvas-1.0.schema.json` (next to `JSON-Canvas-1.0.md`, matching its naming style), and the Layer 1 vocabulary moves from the `x-` prefix to a **reserved `d-` prefix**: `d-version` (replacing `x-dCanvasVersion`), `d-id`, `d-kind`, `d-role`, `d-textId`, `d-condition`, `d-action`, `d-sound`, `d-character`. The `d-` namespace belongs to the spec — unknown `d-*` fields are reserved for future format versions; `x-` remains the conventional home of Layer 2 project extensions (e.g. `x-journalText`). Mandatory recursive preservation applies to every unknown field regardless of prefix.

This is a **clean break**: the reader knows only `d-version` with major 1 and rejects anything else, including files stamped `x-dCanvasVersion` — no 3.0 shim, no migration code. The only 3.0-stamped files are this repo's examples and test fixtures; they are regenerated (the major-version tolerance fixture becomes a 1.1 file). The library, CLI, and their tests are renamed accordingly, with zero remaining references to 3.0 or the old field spellings.

The two never-published internal specs stay in `docs/` but are renumbered **0.1 and 0.2** — file names and titles only, each with a note that they were historically versioned 1.0/2.0 and never published; the version stamps described in their bodies (`x-dCanvasVersion: "1.0"/"2.0"`) are left intact because they describe real legacy artifacts.

Deliberate content changes to the spec and schema, settled in the design review (nothing else changes):

- Node `text` and edge `label` are **literal text**; a note states that Markdown rendering by JSON Canvas tools is a display concern — the format never escapes anything.
- The `"None.png"` sentinel disappears from `d-character.portrait` (a leaked project convention): absent or empty means "no portrait".
- An explicit sentence that the value format of `d-id` / `d-textId` is a project convention, not part of the format.
- The spec loses its historical notes (the 2.0 deprecation paragraph) and the layout implementation detail on `loop` edges — the spec describes the format only.

Documentation follows the three-mode split: code/spec/schema/fixtures carry no trace of the old era; `CONTEXT.md` (a living glossary) is rewritten in `d-*` terms with history referred to via the internal 0.x numbers; existing ADRs, epics, and tasks are **not touched**. Two new ADRs record the decisions: [[ADR-0013 - Public versioning restarts at 1.0 and internal specs are archived as 0.x|ADR-0013]] (public versioning restarts at 1.0; internal specs archived as 0.x) and [[ADR-0014 - Namespace split d- for the format x- for projects|ADR-0014]] (the `d-`/`x-` namespace split), the latter explicitly amending — not superseding — [[ADR-0002 - Two-axis classification x-kind vs x-role|ADR-0002]], [[ADR-0004 - Three-layer x- extension model|ADR-0004]], and [[ADR-0009 - Layout may read x-kind and the loop strategy is configurable|ADR-0009]], whose decisions stand with new field spellings. `AGENTS.md` pointers move to the new spec location. The mempalace memory is re-synced afterwards so it stops describing the old names.

# Acceptance criteria

- [x] `spec/dCanvas-1.0.md` and `spec/dCanvas-1.0.schema.json` exist; the schema's `$id`, `title`, and `d-version` const say 1.0; `docs/` no longer contains `dcanvas-3.0.*`.
- [x] The old internal specs live on as `docs/dcanvas-0.1.*` / `docs/dcanvas-0.2.*` with the "historically 1.0/2.0, never published" note; their body text about legacy version stamps is unchanged.
- [x] All Layer 1 fields in spec, schema, library, CLI, and fixtures use the `d-` prefix; the spec reserves unknown `d-*` fields for future versions and keeps `x-` for project extensions.
- [x] Decoding a file stamped `x-dCanvasVersion: "3.0"` fails with a clear error; decoding a `d-version: "1.1"` file succeeds with unknown fields preserved.
- [x] The spec states literal-text semantics, drops `"None.png"`, marks `d-id`/`d-textId` value formats as project conventions, and contains no historical or implementation references.
- [x] `CONTEXT.md` and `AGENTS.md` reflect the new names and paths; no existing ADR, epic, or task file is modified.
- [x] ADR-0013 and ADR-0014 exist, Accepted, with the amendment wikilinks per house rules.
- [x] `rg -i 'x-dCanvasVersion|"3\.0"|x-kind|x-role|x-textId|x-condition|x-action|x-sound|x-character|x-id' ` over code, spec, schema, and fixtures returns nothing (docs history excluded).
- [x] `go test -C . ./...` and `go test -C cmd/dcanvas ./...` pass.

# Out of scope

- Any content change to the format beyond the items above (no new fields, types, or constraints).
- Migration tooling or compatibility shims for 3.0-stamped files.
- Rewriting historical ADRs, epics, or tasks.
