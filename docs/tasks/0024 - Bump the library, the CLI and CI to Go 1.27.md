---
epic: "[[EPIC-0005 - Migrate to Go 1.27 and json v2]]"
parent:
status: To do
priority: 3
blocked by:
kind: Task
mode: AFK
tier: low
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

The library, the CLI and CI build and test on Go 1.27, with no change in behaviour: the code still reads and writes through `encoding/json` v1, and nothing a user can observe differs. Both Go modules and the workspace declare Go 1.27, and CI keeps taking its Go version from `go.mod`. This slice only separates "the new toolchain breaks something" from "json v2 changes behaviour", so that the next task has a clean baseline. It is also the moment to re-run, on a real Go 1.27, the probes of json v2 behaviour that were made on Go 1.25 with the experiment flag during the design of [[EPIC-0005 - Migrate to Go 1.27 and json v2|EPIC-0005]]: `null` into a string, a number with a fraction into an integer, duplicate keys, invalid UTF-8, a lone surrogate, trailing data, HTML escaping, and how a raw value is written back.

# Acceptance criteria

- [ ] Both `go.mod` files and `go.work` declare Go 1.27
- [ ] `go vet` and `go test` pass for the library and for the CLI on a Go 1.27 toolchain
- [ ] CI passes on Go 1.27 without a hard-coded version outside `go.mod`
- [ ] No source file imports `encoding/json/v2` or `jsontext` yet, and no test expectation was edited
- [ ] The probes of v2 behaviour were re-run on Go 1.27, and any difference from the Go 1.25 results is written into the report for the next task
