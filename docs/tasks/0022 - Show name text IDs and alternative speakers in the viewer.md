---
epic: "[[EPIC-0004 - dCanvas 1.1 alternative characters and a character's text ID]]"
parent:
status: Done
priority: 3
blocked by:
  - "[[0021 - Add alternative characters to dialogue nodes]]"
kind: Task
mode: AFK
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

The reference CLI's viewer shows what 1.1 adds. The detail view of a node shows the speaker's name text ID next to the name, and lists every alternative speaker after `d-character` with the same details — name, text ID, portrait, gender. The tree view keeps showing `d-character` alone, as a 1.0 reader would.

# Acceptance criteria

- [x] The detail view shows `textId` of `d-character` when present
- [x] The detail view lists every alternative speaker, in document order, with its name and whichever of text ID, portrait and gender it has
- [x] The tree view still shows only `d-character`
- [x] `go test -C cmd/dcanvas ./...` passes
