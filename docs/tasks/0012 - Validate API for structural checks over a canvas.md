---
epic: "[[EPIC-0002 - dCanvas CLI TUI and round-trip library robustness]]"
parent:
status: Done
priority: 3
blocked by:
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

A `Validate(*Canvas) error` entry point that reports the structural and semantic problems a file handed to a tool might have, so a consumer (the CLI) can gate an operation on a valid input and tell the user *why* a file is unusable.

Checks:

- node `id`s are unique within the file;
- every edge `fromNode`/`toNode` references an existing node `id`;
- `x-kind` values are within their closed sets (nodes: `line`/`reply`; edges: `normal`/`loop`);
- required fields are present: `text` on `text`-type nodes, `name` on `x-character`.

`Validate` is a Go-side convenience distinct from the JSON-Schema conformance test in [[EPIC-0003 - JSON Schema as cross-language contract and writer conformance|EPIC-0003]]; the two are complementary. It reports problems as errors; it does not mutate the canvas.

# Acceptance criteria

- [x] `Validate` returns an error identifying: a duplicate node id; an edge referencing a missing node; an out-of-set `x-kind` (node and edge); a `text` node without `text`; an `x-character` without `name`.
- [x] A well-formed canvas returns no error.
- [x] Errors name the offending element enough for a user to locate it.
- [x] Table-driven tests cover each failure and the happy path.
- [x] `go test ./...` passes.
