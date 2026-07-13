---
epic: "[[EPIC-0003 - JSON Schema as cross-language contract and writer conformance]]"
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

Give the JSON Schema a trustworthy canonical identity now that this repository is the master source of the format. The `$id` currently points at a stale, unrelated repository (`github.com/SBT/sbt-dialog-viewer`) with an inconsistent org name.

Point `$id` at the master repository — `github.com/sbtlocalization/dcanvas` — and reconcile the `SBT` vs `sbtlocalization` inconsistency so the canonical URL is consistent throughout the schema. Add a guard test asserting the schema's `$id` matches the canonical identity, so it cannot silently drift back.

Keep the schema free of implementation/tooling detail (it describes the format only, per the project's documentation boundaries).

# Acceptance criteria

- [x] The schema `$id` resolves to `github.com/sbtlocalization/dcanvas` (canonical form agreed in the frontmatter of the epic).
- [x] No remaining reference to `sbt-dialog-viewer` or the inconsistent `SBT` org spelling in the schema.
- [x] A test asserts the loaded schema's `$id` equals the canonical value.
- [x] The schema still validates the representative outputs from the conformance harness.
- [x] `go test ./...` passes.
