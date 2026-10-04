---
epic: "[[EPIC-0005 - Migrate to Go 1.27 and json v2]]"
parent:
status: To do
priority: 3
blocked by:
  - "[[0024 - Bump the library, the CLI and CI to Go 1.27]]"
kind: Task
mode: AFK
tier: high
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

The library reads and writes dCanvas through `encoding/json/v2` and `jsontext`, and the behaviour a user sees changes in the three deliberate ways described in [[EPIC-0005 - Migrate to Go 1.27 and json v2|EPIC-0005]]. Reading is strict: a duplicated key, invalid UTF-8 or a lone surrogate in a string, and anything after the document are errors, with the existing `can't decode dcanvas:` prefix. Writing is literal: `<`, `>`, `&` and U+2028/2029 are written as they are, including inside an unknown field, whose value comes back byte for byte; the document is indented with a tab and still ends with a newline. The hand-written marshalers stay as they are: only the package and the raw-value type change, so that a regression can be attributed to the change of semantics and not to a rewrite.

The strict-reading and literal-writing cases are written as tests first, fail on v1 and pass after the swap. The golden file of the CLI and every fixture containing `<`, `>` or `&` are refreshed once, and the diff is read by hand. A new ADR records the minimum Go version, the strict reading, the literal writing and the release as v1.1.1; it builds on [[ADR-0016 - The module version follows the format version and library-only changes go into the patch|ADR-0016]] and supersedes nothing.

# Acceptance criteria

- [ ] `Decode` rejects a document with a duplicated key, with invalid UTF-8 in a string, with a lone surrogate escape in a string, and with anything after the document, each covered by a test that failed before the change
- [ ] `Encode` writes `<`, `>`, `&` and U+2028 unescaped, in known fields and in an unknown field, and an unknown field's numbers and strings come back byte for byte, each covered by a test
- [ ] `Encode` output ends with a newline and is indented with a tab
- [ ] No source file imports `encoding/json` v1 any more, in the library, the CLI and the tests
- [ ] The existing round-trip, version and conformance tests pass with no change in meaning; the CLI golden file and the changed fixtures were refreshed and their diff reviewed by hand
- [ ] A new ADR in `docs/adr/` records the four decisions, its references to other ADRs are wikilinks with an alias, and the spec and the schema are not edited
- [ ] The probes of v2 behaviour for the cases that matter to the format showed no surprise on Go 1.27
