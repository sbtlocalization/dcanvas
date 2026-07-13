---
epic: "[[EPIC-0002 - dCanvas CLI TUI and round-trip library robustness]]"
parent:
status: Backlog
priority: 2
blocked by: ["[[0010 - Quarantine layout into a layout subpackage]]"]
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

The walking skeleton of the CLI: a `cmd/` tool that reads a `.d.canvas` file, runs auto-layout with the default strategy (`LoopCut`), and writes the laid-out canvas back — end-to-end. This is the tracer bullet that proves the read → operate → write path.

It accepts either file extension (`.d.canvas` / `.dcanvas`). Unknown/Layer 2 fields survive the operation via the library's existing preservation. The command surface (a `layout` subcommand vs a flag) is the implementer's call; keep it minimal.

Robustness hardening (input validation, layout-failure reporting, loop-strategy choice) is layered on in later tasks; this slice uses the library as it stands and is demoable on a well-formed `3.0` file.

# Acceptance criteria

- [ ] Running the CLI on a coordinate-less `.d.canvas` file produces a file whose nodes have been positioned (at least one connected node moved off its input position).
- [ ] Both `.d.canvas` and `.dcanvas` inputs are accepted.
- [ ] Fields the tool did not touch (including Layer 2 project fields) are preserved in the output.
- [ ] A golden-file test drives the CLI end-to-end (input file → expected output file).
- [ ] `go test ./...` passes.
