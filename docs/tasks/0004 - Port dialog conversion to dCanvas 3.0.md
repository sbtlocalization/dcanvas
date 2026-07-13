---
epic: "[[EPIC-0001 - dCanvas 3.0 Reference Library and Migration]]"
parent:
status: Postponed
priority:
blocked by:
  - "[[0001 - dCanvas 3.0 types and IO with preservation]]"
  - "[[0003 - Generic layout over a canvas]]"
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

Port the Infinity-specific dialogue conversion (`ToDCanvas` and its node/edge/character helpers) to produce dCanvas 3.0 canvases using the new library, while keeping all engine-specific logic in this project ([[ADR-0006 - Layout belongs to the dCanvas library|ADR-0006]]: the library stays domain-agnostic).

- NPC line → node with `x-kind: "line"`; player reply → node with `x-kind: "reply"`.
- The engine-specific term (the old `state`/`transition`) goes into the open `x-role` string, not the closed `x-kind`.
- Journal data (`x-journalText`, `x-journalTextId`, `x-journalSound`) is emitted as **Layer 2** project extensions on the relevant nodes — no longer part of the format core.
- Each node carries a stable file-local canvas `id` (referenced by edges) and the engine identifier in `x-id`.
- The player-facing choice text goes on the edge `label` (with `x-textId` when applicable); spoken text goes on the node `text`.
- Positioning is delegated to the library's layout operation; the inline `autog` call is removed from the conversion.

# Acceptance criteria

- [ ] A representative dialogue converts to a canvas whose nodes carry correct `x-kind` (`line`/`reply`) and the engine term in `x-role`.
- [ ] Journal fields appear as `x-` Layer 2 extensions on the produced nodes and round-trip through the library unchanged.
- [ ] Nodes have distinct canvas `id` and `x-id`; edges reference nodes by canvas `id`.
- [ ] Player choice text is on the edge `label`; spoken text is on the node `text`.
- [ ] Node positions come from the library layout; no `autog` call remains in the conversion code.
- [ ] A test verifies the conversion output for a representative dialogue.
