---
epic: "[[EPIC-0005 - Migrate to Go 1.27 and json v2]]"
parent:
status: Done
priority: 3
blocked by:
  - "[[0025 - Read and write with json v2 strictly and literally]]"
  - "[[0026 - Collect unknown fields with the v2 unknown-member mechanism]]"
kind: Task
mode: HITL
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

The migration to Go 1.27 and json v2 is released: the module is tagged `v1.1.1`, because the major and minor of the tag equal the format version and a library-only change goes into the patch ([[ADR-0016 - The module version follows the format version and library-only changes go into the patch|ADR-0016]]), and `cmd/dcanvas` depends on it. The release notes say that the minimum Go version is now 1.27, that reading is stricter (duplicated keys, invalid UTF-8, trailing data), and that written files are no longer HTML-escaped. If [[0026 - Collect unknown fields with the v2 unknown-member mechanism|0026]] proves problematic it may be dropped from the release, which then waits only for [[0025 - Read and write with json v2 strictly and literally|0025]]. Tagging and pushing are the maintainer's call, hence HITL.

# Acceptance criteria

- [x] The module is tagged `v1.1.1` and `Version` is still `"1.1"`
- [x] `cmd/dcanvas` depends on `v1.1.1`
- [x] `go test -C . ./...` and `go test -C cmd/dcanvas ./...` pass at the tag on Go 1.27
- [x] The new minimum Go version, the stricter reading and the literal writing are stated for consumers in the accepted ADRs ([[ADR-0017 - The library needs Go 1.27 reads strictly and writes literally with json v2|ADR-0017]] and [[ADR-0018 - Unknown fields are kept by a json v2 embedded fallback over typed fields|ADR-0018]]); earlier releases had no release notes beyond their tags and no GitHub release was created, which is the maintainer's call
