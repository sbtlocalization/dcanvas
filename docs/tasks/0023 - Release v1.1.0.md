---
epic: "[[EPIC-0004 - dCanvas 1.1 alternative characters and a character text ID]]"
parent:
status: Done
priority: 3
blocked by:
  - "[[0022 - Show name text IDs and alternative speakers in the viewer]]"
kind: Task
mode: HITL
tier: low
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

dCanvas 1.1 is released: the module is tagged `v1.1.0`, since a new format minor is the one thing that moves the module minor ([[ADR-0016 - The module version follows the format version and library-only changes go into the patch|ADR-0016]]), and `cmd/dcanvas` depends on it. Tagging and pushing are the maintainer's call, hence HITL.

# Acceptance criteria

- [x] The module is tagged `v1.1.0` and `Version` is `"1.1"`
- [x] `cmd/dcanvas` depends on `v1.1.0`
- [x] `go test -C . ./...` and `go test -C cmd/dcanvas ./...` pass at the tag
