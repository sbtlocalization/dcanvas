---
epic: "[[EPIC-0001 - dCanvas 3.0 Reference Library and Migration]]"
parent:
status: Done
priority:
blocked by: ["[[0001 - dCanvas 3.0 types and IO with preservation]]"]
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

Extend field preservation so it is **recursive** into known nested objects, satisfying the spec's "preserve unrecognised fields at any depth" guarantee ([[ADR-0003 - Mandatory recursive preservation of unknown fields|ADR-0003]]). The only known nested object today is `x-character`, but the mechanism should generalise to any future nested known object.

A project may add its own keys to `x-character` (e.g. a mood). When a tool that only understands `name`/`portrait`/`gender` reads and re-saves the file, those extra keys must survive untouched.

# Acceptance criteria

- [x] An unknown key inside `x-character` survives a decode → encode round-trip.
- [x] Known `x-character` fields (`name`, `portrait`, `gender`) still decode/encode as typed members.
- [x] The recursion approach is reusable for additional known nested objects, not hard-wired to `x-character` alone.
- [x] A test covers the nested-preservation case.
