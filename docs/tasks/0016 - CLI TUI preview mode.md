---
epic: "[[EPIC-0002 - dCanvas CLI TUI and round-trip library robustness]]"
parent:
status: To do
priority: 1
blocked by:
  - "[[0011 - CLI walking skeleton lay out a file end-to-end]]"
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

A TUI mode that renders the laid-out graph so the user can eyeball the result before it is written to disk, then confirm the write or abort.

The TUI reuses the same read → layout pipeline as the CLI; it adds a visual preview and a confirm/cancel step. The TUI framework is the implementer's choice.

# Acceptance criteria

- [ ] Launching the TUI on a `.d.canvas` file shows a preview of the laid-out graph (nodes positioned, edges between them).
- [ ] The user can confirm to write the result or cancel to leave the file untouched.
- [ ] Cancelling leaves the input file unchanged on disk.
- [ ] `go test ./...` passes (logic under the TUI is covered where practical).
