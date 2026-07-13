---
epic: "[[EPIC-0002 - dCanvas CLI TUI and round-trip library robustness]]"
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

Move auto-layout into its own `layout` subpackage so the core `dcanvas` package (types, IO, preservation) carries **no third-party dependency**, and `autog` is quarantined behind an engine-neutral boundary.

The public layout API keeps the same shape it has today — `Layout(canvas, opts...)`, `LoopStrategy`/`LoopCut`/`LoopDFS`, `WithLoopStrategy` — and exposes **no `autog` types** in its signatures, so the engine can be replaced later without breaking callers. No pluggable engine interface is introduced (deferred until a second engine exists), per [[ADR-0011 - An in-repo reference CLI is a round-tripping consumer and the layout engine stays swappable|ADR-0011]] and [[ADR-0009 - Layout may read x-kind and the loop strategy is configurable|ADR-0009]].

This is a pure restructuring: behaviour is unchanged; only the package layout and dependency graph change.

# Acceptance criteria

- [x] Layout lives in a `layout` subpackage; its tests move with it and pass.
- [x] The root `dcanvas` package imports nothing outside the standard library (verifiable via `go list` on its non-test imports).
- [x] `autog` appears only as a dependency of the `layout` subpackage.
- [x] The public layout API exposes no `autog`/`graph` types in its signatures.
- [x] `go vet ./...` and `go test ./...` pass.
