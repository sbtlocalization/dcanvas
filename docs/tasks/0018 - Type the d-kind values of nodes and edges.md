---
epic:
parent:
status: Backlog
priority: 3
blocked by:
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

`Node.Kind` and `Edge.Kind` are plain strings, so nothing stops a node from being given an edge kind or an arbitrary string, although both vocabularies are closed sets. Each gets its own named type, `NodeKind` and `EdgeKind`, and the `Kind*` constants are typed accordingly; a node kind can no longer be assigned to an edge, or the other way round, without a conversion.

The change was started in the working tree and left unfinished: the types and fields are declared, but encoding, decoding, validation and the reference CLI still treat the kind as a plain string, and the module does not build. The wire format is untouched — `d-kind` stays a JSON string with the same values.

Consumers that only assign and compare the `Kind*` constants keep compiling unchanged; one that stores a kind in a plain `string` needs a conversion. The release goes out as a patch under [[ADR-0016 - The module version follows the format version and library-only changes go into the patch|ADR-0016]], together with [[0019 - Add Edge.SetExtra and release v1.0.1|0019]].

# Acceptance criteria

- [ ] `NodeKind` and `EdgeKind` exist, `Node.Kind` and `Edge.Kind` have those types, and `KindLine`, `KindReply`, `KindNormal`, `KindLoop` are typed constants
- [ ] Assigning an `EdgeKind` to `Node.Kind`, or a `NodeKind` to `Edge.Kind`, does not compile
- [ ] Encoding and decoding produce and accept exactly the same `d-kind` JSON as before, and validation still rejects an unknown kind
- [ ] The reference CLI builds and shows `d-kind` as before
- [ ] `go test -C . ./...` and `go test -C cmd/dcanvas ./...` pass
