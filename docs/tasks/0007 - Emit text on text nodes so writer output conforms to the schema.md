---
epic: "[[EPIC-0003 - JSON Schema as cross-language contract and writer conformance]]"
parent:
status: Done
priority: 3
blocked by:
  - "[[0006 - Conformance harness validating writer output against the JSON Schema]]"
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

Close the one known divergence between the library and the schema: the schema requires `text` on `text`-type nodes, but `Encode` treats `text` as `omitempty` and drops it when empty — producing a document that fails schema validation.

Add a failing (red) conformance case first: a `text` node with empty text is encoded and fails validation against the schema. Then fix encoding so a `text`-type node always emits its `text` field (even when empty), turning the case green. Non-`text` node types (`group`, `file`, `link`) are unaffected — the schema only requires `text` when `type` is `text`.

If the conformance harness surfaces any further divergences, fix them the same way (make the library structurally unable to emit schema-invalid output).

# Acceptance criteria

- [x] A red test demonstrates that, before the fix, a `text` node with empty text encodes to a document that fails schema validation.
- [x] After the fix, a `text`-type node always emits a `text` field; the same test passes.
- [x] `group` / `file` / `link` nodes are not forced to emit `text`.
- [x] The existing "stripped is valid JSON Canvas" and key-order tests still pass.
- [x] `go test ./...` passes.
