---
epic: "[[EPIC-0003 - JSON Schema as cross-language contract and writer conformance]]"
parent:
status: To do
priority: 3
blocked by:
  - "[[0006 - Conformance harness validating writer output against the JSON Schema]]"
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

A GitHub Actions workflow that runs the Go test suite on every push and pull request, so the cross-language contract is enforced automatically: the conformance harness (task 0006) becomes part of what gates CI, and drift between the Go library and the JSON Schema fails the build rather than surfacing in a reader user's hands.

The workflow checks out the repository, sets up the Go toolchain matching `go.mod`, and runs `go vet ./...` and `go test ./...`.

# Acceptance criteria

- [ ] A GitHub Actions workflow triggers on push and pull request.
- [ ] It sets up the Go version declared in `go.mod`.
- [ ] It runs `go vet ./...` and `go test ./...`; a failing conformance test fails the workflow.
- [ ] The workflow is green on the current `main`.
