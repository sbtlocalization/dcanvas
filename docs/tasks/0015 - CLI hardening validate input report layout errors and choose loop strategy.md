---
epic: "[[EPIC-0002 - dCanvas CLI TUI and round-trip library robustness]]"
parent:
status: Backlog
priority: 3
blocked by: ["[[0011 - CLI walking skeleton lay out a file end-to-end]]", "[[0012 - Validate API for structural checks over a canvas]]", "[[0013 - Lossless round-trip preserve version and stop rejecting x-kind on write]]", "[[0014 - Layout returns an error instead of silently recovering]]"]
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

Harden the CLI now that the library pieces exist: turn the walking skeleton into a tool that fails loudly and honestly, and lets the user steer layout.

- **Validate on load.** Run `Validate` on the decoded canvas before operating; on failure, report the problem and exit non-zero without writing an output file.
- **Report layout failure.** Surface an error returned by `Layout` (non-zero exit, clear message) instead of writing a "done" that changed nothing.
- **Loop strategy flag.** A `--loop` option selecting `LoopCut` (default) or `LoopDFS`, wired to `WithLoopStrategy`.
- Writing back relies on the lossless round-trip (0013), so the output keeps the input's version and any data the tool did not touch.

# Acceptance criteria

- [ ] Running on an invalid file (dangling edge, duplicate id, missing required field) prints the problem, exits non-zero, and writes no output.
- [ ] A layout failure is reported with a non-zero exit; success writes the laid-out file.
- [ ] `--loop cut` (default) and `--loop dfs` change the layout behaviour accordingly.
- [ ] The output preserves the input's `x-dCanvasVersion` and untouched fields.
- [ ] Tests drive the CLI for the invalid-input, layout-failure, and both loop-strategy paths.
- [ ] `go test ./...` passes.
