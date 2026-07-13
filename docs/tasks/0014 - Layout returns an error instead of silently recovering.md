---
epic: "[[EPIC-0002 - dCanvas CLI TUI and round-trip library robustness]]"
parent:
status: To do
priority: 3
blocked by:
  - "[[0010 - Quarantine layout into a layout subpackage]]"
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

Change `Layout` to return an error instead of silently swallowing a failed layout with `recover`. A CLI whose primary job is layout must be able to report a failure to the user rather than claim success while nothing moved.

Semantics:

- a genuine layout failure (e.g. the engine panics on malformed input) is returned as an error, not recovered-and-ignored;
- a **no-op** — an empty graph, or a graph with no layout edges (e.g. every edge cut as a loop under `LoopCut`) — is a success (nil error) that leaves positions untouched, not a failure.

# Acceptance criteria

- [ ] `Layout` returns an error; callers can detect and surface a failed layout.
- [ ] An empty graph and a graph with no layout edges return nil and leave positions unchanged.
- [ ] A connected graph lays out and returns nil.
- [ ] A cyclic graph still completes without a panic escaping, per [[ADR-0009 - Layout may read x-kind and the loop strategy is configurable|ADR-0009]].
- [ ] Existing layout tests are updated to the new signature; `go test ./...` passes.
