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

A generic auto-layout operation, owned by the library (~~[[ADR-0006 - Layout belongs to the dCanvas library|ADR-0006]]~~ [[ADR-0009 - Layout may read x-kind and the loop strategy is configurable|ADR-0009]]), that assigns positions to a canvas's nodes. It uses `autog`, reads node `width`/`height` → `x`/`y`, driven by edge `fromNode`/`toNode`. ~~It reads and writes only Layer 0 geometry, and knows nothing about dialogue semantics or any project's `x-` fields.~~ Per [[ADR-0009 - Layout may read x-kind and the loop strategy is configurable|ADR-0009]] it also reads the Layer 1 field `x-kind: loop` to identify back-edges; it ignores all other `x-` fields.

- ~~Cycles are **not** excluded; the layout engine resolves them itself. `x-kind: "loop"` edges participate in layout like any other.~~ Per [[ADR-0009 - Layout may read x-kind and the loop strategy is configurable|ADR-0009]], `x-kind: loop` edges are **excluded** from positioning by default (`LoopCut`) so cyclic dialogues lay out as a clean tree; `LoopDFS` keeps them. Cycles never crash layout either way.
- The call is wrapped in panic recovery: if layout fails, nodes keep their existing/default positions and the operation does not crash.
- This lifts the positioning logic currently tangled inside the Infinity-specific conversion into a standalone, reusable operation.

# Acceptance criteria

- [x] A multi-node connected graph receives non-default coordinates after layout.
- [x] Laid-out nodes do not overlap (verifiable via the existing overlap check).
- [x] A graph containing a cycle lays out without panicking.
- [x] An empty graph (or a graph with no edges) is handled gracefully as a no-op / default placement.
- [x] The layout operation has no dependency on the Infinity domain.
