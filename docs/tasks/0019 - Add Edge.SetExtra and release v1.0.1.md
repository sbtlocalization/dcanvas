---
epic:
parent:
status: Done
priority: 3
blocked by:
  - "[[0018 - Type the d-kind values of nodes and edges]]"
kind: Task
mode: AFK
model: sonnet
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

A project can attach its own Layer 2 `x-` field to a node through `Node.SetExtra`, but not to an edge, although the spec allows `x-` fields on edges just as on nodes and the library already preserves unknown edge fields on a round trip. `Edge.SetExtra` closes that gap with the same contract as its node counterpart: the value is JSON-marshaled, written verbatim on encode, survives decoding like any other unrecognised field, and a typed field with the same key always wins.

The first consumer is sbt-mele, which marks charm and intimidate choices of a conversation with an `x-` field on the edge that carries the choice.

Once this and [[0018 - Type the d-kind values of nodes and edges|0018]] are in, the module is tagged `v1.0.1`: both are library-only changes, which go into the patch under [[ADR-0016 - The module version follows the format version and library-only changes go into the patch|ADR-0016]]. `Version` and the spec stay at 1.0.

# Acceptance criteria

- [x] `Edge.SetExtra` attaches a field that appears in the encoded edge and is still there after decoding and encoding again
- [x] A typed edge field wins over an extra set under the same key, as it does for nodes
- [x] A value that cannot be marshaled returns an error naming the key, as it does for nodes
- [x] `go test -C . ./...` and `go test -C cmd/dcanvas ./...` pass
- [x] The module is tagged `v1.0.1` and `Version` is still `"1.0"`
