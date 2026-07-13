---
epic: "[[EPIC-0003 - JSON Schema as cross-language contract and writer conformance]]"
parent:
status: Done
priority: 4
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

A test that proves the Go library's encoded output satisfies the shared JSON Schema — the same `docs/dcanvas-3.0.schema.json` the TypeScript reader consumes. This is the end-to-end conformance mechanism: build a canvas through the public API, encode it, validate the produced bytes against the schema.

Cover the representative surface documented in the spec so the whole format is exercised:

- a `line` node with `x-character`;
- a `reply` node carrying a Layer 2 project field;
- an edge with a `label` and an `x-condition`;
- a minimal complete document.

The canvases are built programmatically through the public API (no committed `example.d.canvas` fixture). The schema validator is a **test-only** dependency (a draft-07 validator such as `santhosh-tekuri/jsonschema`); the library/writer runtime stays standard-library-only.

Assert on external behaviour only — the encoded bytes validate — never the internal catch-all representation, matching the philosophy of `encodeToMap` in `io_test.go`.

# Acceptance criteria

- [x] A draft-07 JSON-Schema validator is a test-only dependency; `go list` shows no new non-stdlib dependency in the library's non-test imports.
- [x] The test loads `docs/dcanvas-3.0.schema.json` and validates encoded output against it.
- [x] Representative canvases (line + `x-character`, reply + Layer 2 field, edge with label + condition, minimal document) are each built via the public API, encoded, and pass validation.
- [x] The test asserts on encoded bytes, not on internal fields.
- [x] `go test ./...` passes.
