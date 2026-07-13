---
epic: "[[EPIC-0001 - dCanvas 3.0 Reference Library and Migration]]"
parent:
status: Postponed
priority:
blocked by:
  - "[[0004 - Port dialog conversion to dCanvas 3.0]]"
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

Wire the `dialog ex` command to write dCanvas 3.0 output via the ported conversion and the library's encoder, and prove the whole path end-to-end. From the user's perspective the command behaves as before; only the produced file format is now 3.0.

This is the acceptance test of the format against real Infinity data: a real DLG flows through parsing → conversion → layout → encode and yields a valid, round-trip-safe dCanvas 3.0 file.

# Acceptance criteria

- [ ] `dialog ex` produces a `.dcanvas`/`.d.canvas` file stamped `x-dCanvasVersion: "3.0"`.
- [ ] The produced file is accepted by the library's `Decode` as valid 3.0 and round-trips (decode → encode) without losing any field.
- [ ] Stripping all `x-` fields from the produced file yields a valid JSON Canvas 1.0 document (Obsidian-openable).
- [ ] An end-to-end test exports a fixture DLG and asserts the above on the result.
- [ ] The command's existing flags and user-facing behaviour are unchanged apart from output format.
