---
epic: "[[EPIC-0004 - dCanvas 1.1 alternative characters and a character text ID]]"
parent:
status: Done
priority: 3
blocked by:
kind: Task
mode: AFK
tier: high
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

The first slice of dCanvas 1.1, end to end: the format gains its first 1.1 field and the library writes 1.1.

`spec/dCanvas-1.1.md` and `spec/dCanvas-1.1.schema.json` are added beside the 1.0 files, which stay untouched — minors differ structurally, and the version field is constrained by major in the new schema as in the old ([[ADR-0015 - The schema constrains d-version by major while schema files stay per minor|ADR-0015]]). The 1.1 spec states what changed since 1.0. CONTEXT.md and the README point at the 1.1 spec.

`d-character` gains an optional `textId`: the string-table reference for `name`, its value format a project convention exactly as for `d-textId`. It joins the known fields of the open object; every other key is still preserved.

The library models it as a typed field of `Character` and its `Version` becomes `"1.1"`. The tests and fixture that prove a higher same-major minor is accepted and preserved move from 1.1 to 1.2, since the library now writes 1.1 itself.

# Acceptance criteria

- [x] `spec/dCanvas-1.1.md` and `spec/dCanvas-1.1.schema.json` exist, the 1.0 files are unchanged, and the 1.1 schema constrains `d-version` to major 1
- [x] The 1.1 spec and schema define `d-character.textId` as an optional string referencing `name`, with a project-defined value format, and the 1.1 spec lists it as the change since 1.0
- [x] The spec says nothing about the library, as AGENTS.md requires
- [x] `Character` has a typed text ID that decodes from and encodes to `textId`, and a character's unknown keys still round-trip
- [x] `Version` is `"1.1"`, and the writer's version contract test asserts it
- [x] The library's output validates against the 1.1 schema, and a 1.1 document using `textId` also validates against the 1.0 schema
- [x] The same-major tolerance tests and fixture use 1.2 and still pass; a 2.0 document is still rejected
- [x] CONTEXT.md and the README refer to the 1.1 spec
- [x] `go test -C . ./...` and `go test -C cmd/dcanvas ./...` pass
