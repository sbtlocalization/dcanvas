---
epic: "[[EPIC-0001 - dCanvas 3.0 Reference Library and Migration]]"
parent:
status: Done
priority:
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

The dCanvas 3.0 format as a self-contained Go package with typed fields and round-trip IO. This is the tracer bullet for the whole library: it establishes the package shape, the typed model, and the real IO mechanism in one cut.

- Typed model for the JSON Canvas core (Layer 0) plus the dialogue vocabulary (Layer 1) on `Canvas`, `Node`, `Edge`, and `Character`. Node Layer 1: `x-id`, `x-kind` (`line`/`reply`), `x-role` (free string), `x-textId`, `x-condition`, `x-action`, `x-sound`, `x-character`. Edge Layer 1: `x-id`, `x-kind` (`normal`/`loop`), `x-role` (free string), `x-condition`, `x-textId`. All `x-` fields optional.
- `Encode` always stamps `x-dCanvasVersion: "3.0"`.
- `Decode` validates the **major** version and returns an error on mismatch (a 2.0 file is rejected); no migration.
- **Field preservation** via the catch-all mechanism (ADR-0005, not struct embedding): decode into a raw map, populate typed known fields, retain unrecognised keys in a catch-all; on encode, merge typed fields with the preserved map. This applies at the top level, node, and edge. (Recursion into nested known objects is task 0002.)
- The package depends on nothing from the Infinity domain (one-way dependency: project → library).

This replaces the existing dCanvas 2.0 types and IO in the package.

# Acceptance criteria

- [x] `Canvas`, `Node`, `Edge`, `Character` expose Layer 0 and Layer 1 fields as typed members; `x-kind` is constrained to its closed value sets in encode/validation.
- [x] `Encode` produces a document with `x-dCanvasVersion: "3.0"`; stripping all `x-` fields yields a valid JSON Canvas 1.0 document.
- [x] `Decode` accepts a `"3.0"` document and rejects a document whose major version ≠ 3 with an error.
- [x] Round-trip preserves an unknown top-level field, an unknown node `x-` field, and an unknown edge `x-` field.
- [x] Typed Layer 1 fields decode and re-encode faithfully.
- [x] The package has no import of `dialog`, `parser`, or other Infinity domain packages.
- [x] Table-driven tests cover the above, in the style of the existing package tests.
