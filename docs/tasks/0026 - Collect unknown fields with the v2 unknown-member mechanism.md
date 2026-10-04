---
epic: "[[EPIC-0005 - Migrate to Go 1.27 and json v2]]"
parent:
status: To do
priority: 3
blocked by:
  - "[[0025 - Read and write with json v2 strictly and literally]]"
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

The hand-written preservation machinery is replaced by the unknown-member mechanism of json v2, which in Go 1.27 is an embedded fallback: a `jsontext.Value` field tagged `embed`, not the `unknown` option of the experimental package, which does not exist and is ignored without an error. The unknown members of the canvas, of nodes, edges and characters are collected by it and written back after the known ones, in the order they had in the file. The alphabetical order of unknown fields is given up on purpose ([[EPIC-0005 - Migrate to Go 1.27 and json v2|EPIC-0005]]). What a tag cannot express stays in code: a `text` node always writes its `text` while other node types omit it when empty, and a key owned by a typed field is never duplicated by an extension, because json v2 turns a duplicated key on writing into an error where v1 silently let the typed field win. The documented contract of `Node.SetExtra` and `Edge.SetExtra` does not change.

This is a refactoring: the observable behaviour, apart from the order of unknown fields, is the one the previous task left, and the round-trip tests are the safety net.

# Acceptance criteria

- [ ] The catch-all maps, the helpers that pop known keys, and the ordered object writer are gone from the library
- [ ] Unknown fields at every level (canvas, node, edge, character) survive a round-trip, in the order of the source file, and a test fails if the mechanism is mis-tagged and an unknown field silently disappears
- [ ] The key-order test is rewritten from alphabetical to file order on purpose, and a test shows unknown fields keeping the order of the file
- [ ] `SetExtra` on a key owned by a typed field still leaves the typed value and produces no error and no duplicated key, for nodes and edges
- [ ] A `text` node still always writes `text`, and other node types still omit it when empty
- [ ] The remaining round-trip, version and conformance tests pass with no change in meaning, and the CLI golden test passes without a refresh
