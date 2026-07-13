---
epic: "[[EPIC-0002 - dCanvas CLI TUI and round-trip library robustness]]"
parent:
status: Backlog
priority: 2
blocked by: ["[[0012 - Validate API for structural checks over a canvas]]"]
kind: Bug
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

Make `Decode → Encode` faithful: writing back a file the library just read must never fail on, or silently alter, data that file legitimately contained.

Two fixes:

- **Version.** `Encode` currently always stamps `3.0`, downgrading a `3.x` file. Instead, preserve the decoded `x-dCanvasVersion` when its major matches the library's (so `3.1` stays `3.1`). A canvas with no decoded version (hand-built) or a different major is stamped with the library version — so the existing "hand-built canvas gets the library version, a `9.9` value becomes `3.0`" behaviour still holds for the different-major case.
- **`x-kind`.** `Encode` currently rejects an unrecognised `x-kind`, which breaks the round-trip of a file `Decode` happily accepted. `Encode` no longer enforces the closed set; that check now lives in `Validate` (task 0012). Consumers that want the guarantee call `Validate`.

# Acceptance criteria

- [ ] A `3.1` document round-trips (`Decode` → `Encode`) with `x-dCanvasVersion` still `3.1`.
- [ ] A hand-built canvas with no version is stamped with the library version; a different-major value is replaced with the library version.
- [ ] A document whose `x-kind` the library does not recognise round-trips without `Encode` returning an error.
- [ ] The closed-set `x-kind` check remains available via `Validate`.
- [ ] The previous always-stamp-`3.0` test is updated to the new contract; `go test ./...` passes.
