---
epic: "[[EPIC-0004 - dCanvas 1.1 alternative characters and a character's text ID]]"
parent:
status: Done
priority: 3
blocked by:
  - "[[0020 - Open dCanvas 1.1 with a text ID for a character's name]]"
kind: Task
mode: AFK
tier: mid
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

A dialogue node gains `d-alternativeCharacters`: a non-empty list of objects with the schema of `d-character`, `name` required in each, unknown keys preserved in each. It sits on the node beside `d-character`, never inside it, and appears only together with `d-character`.

The 1.1 spec states its meaning: the line is spoken by exactly one of `d-character` and the alternatives; `d-character` is the speaker a reader that shows only one shows, and has no other precedence; the order of the alternatives carries no meaning in the format. A line spoken by several characters at once is not what this field says.

The library models it as a typed list of `Character` on `Node`. `Validate` reports an alternative without a name and alternatives on a node without `d-character`.

# Acceptance criteria

- [x] The 1.1 spec and schema define `d-alternativeCharacters` as an optional non-empty list of `d-character` objects on a node, require `name` in each, and allow it only together with `d-character`
- [x] The 1.1 spec states that exactly one of `d-character` and the alternatives speaks the line, that `d-character` has no precedence beyond being shown first, and that the order of the alternatives carries no meaning
- [x] `Node` has a typed list of alternative characters that decodes from and encodes to `d-alternativeCharacters`, with unknown keys of each alternative preserved
- [x] `Validate` reports an alternative without a name, and a node with alternatives but no `d-character`, and reports nothing for a well-formed node
- [x] The library's output with alternatives validates against the 1.1 schema, and also against the 1.0 schema
- [x] `go test -C . ./...` and `go test -C cmd/dcanvas ./...` pass
