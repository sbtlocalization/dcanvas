---
status: Done
priority: 3
blocked by:
kind: PRD
---
# Tasks

<!-- live the following block unchanged -->
```base
summaries: {}
filters:
  and:
    - epic == [this.file]
formulas:
  is_blocked: |-
    note["blocked by"].reduce(
      acc || (
        value.asFile().properties.status != "Done" && 
        value.asFile().properties.status != "Obsolete"
      ),
      false
    )
  progress_icon: |-
    if(status == "Done",
      icon("square-check-big"),
      if (status == "Obsolete",
	    icon("square-arrow-right"),
        if(formula.is_blocked,
          icon("construction"), 
          icon("square")
        )
      )
    )
  progress_sort: |-
    if(status == "Done" || status == "Obsolete", 
      20, 
      if(formula.is_blocked, 
        10, 
        if(status == "In progress", 
          5, 
          0)))
  progress_string: |-
    if(status == "Done" || status == "Obsolete", 
      "Done", 
      if(formula.is_blocked, 
        "On hold",
        "Ready"))
  priority: if(priority, icon("tally-" + priority), null)
  kind_icon: |-
    if(kind == "Bug",
      icon('bug'),
      ''
    )
properties:
  note.status:
    displayName: status
  file.name:
    displayName: task
  formula.progress_icon:
    displayName: " "
  formula.progress_string:
    displayName: Tasks that are
  formula.kind_icon:
    displayName: " "
views:
  - type: table
    name: Таблиця
    groupBy:
      property: formula.progress_string
      direction: DESC
    order:
      - formula.progress_icon
      - formula.kind_icon
      - file.name
      - formula.priority
      - status
      - blocked by
      - mode
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
      formula.kind_icon: -1
      file.name: 435
      note.status: 125
      note.blocked by: 522
```
<!-- end unchanged block -->

# Problem Statement

The dCanvas format is consumed across two languages: a Go **writer** generates files, and a TypeScript + Svelte **reader** opens them and validates them against `docs/dcanvas-3.0.schema.json`. The **JSON Schema is the contract** between them — the reader cannot use the Go library at all.

Nothing today couples the Go library to that schema, so the two can drift silently. Concretely, the library can emit a document that violates its own schema: a `text` node with empty text is written *without* a `text` field, but the schema requires `text` on `text` nodes. A reader validating such a file would reject it — and the failure surfaces in a reader user's hands, not in the writer's build.

The schema's `$id` also points at a stale, unrelated repository (`github.com/SBT/sbt-dialog-viewer`) and uses an inconsistent org name, so the canonical contract has no trustworthy identity now that this repository is the master source.

# Solution

Make the JSON Schema the **enforced** cross-language contract:

1. A conformance test in CI takes representative output of the Go writer/library and validates it against the very schema the reader consumes, so any divergence fails the build instead of a reader user.
2. Fix the library so it cannot emit schema-invalid output (starting with the missing-`text` case the test surfaces).
3. Canonicalise the schema's `$id` to this repository, which is now the master source of the format ([[ADR-0010 - The JSON Schema is the cross-language contract enforced by writer conformance|ADR-0010]]).

# User Stories

1. As a reader developer, I want every file the writer produces to validate against the JSON Schema, so that my reader never receives a file it must reject.
2. As a reader developer, I want the schema's `$id` to resolve to the real master repository, so that I can trust and reference the contract.
3. As a maintainer, I want drift between the Go library and the schema caught in CI, so that a change to one that breaks the other fails the build.
4. As a maintainer, I want the library to be structurally unable to emit a document that violates the schema, so that conformance is guaranteed, not hoped for.
5. As a maintainer, I want the conformance check to validate representative outputs (line node, reply node with a Layer 2 field, edges with labels and conditions, `x-character`, and a minimal document), so that the whole documented surface is exercised.
6. As a maintainer, I want the schema validator to be a test-only dependency, so that the writer's runtime stays standard-library-only.
7. As a maintainer, I want the spec and schema to remain free of implementation/tooling detail, so that they describe only the format (per the project's documentation boundaries).

# Implementation Decisions

- **Conformance test (test-only dependency).** A Go test builds representative canvases through the public API, encodes them, and validates the produced bytes against `docs/dcanvas-3.0.schema.json` using a draft-07 JSON-Schema validator (e.g. `santhosh-tekuri/jsonschema`). The dependency is confined to test code; the library/writer runtime remains standard-library-only.
- **Fix schema-invalid output.** The library must not emit a `text` node without `text`. Resolution (task-level): emit `text` unconditionally for `text`-type nodes rather than treating it as `omitempty`, or the narrowest fix the conformance test surfaces. Any further divergences the test finds are fixed the same way.
- **Canonical `$id`.** Point the schema `$id` at this repository (`github.com/sbtlocalization/dcanvas`) and reconcile the `SBT` vs `sbtlocalization` inconsistency.
- **Contract, not library.** The schema is the cross-language contract; the Go library is the writer's reference implementation, judged against the schema — not the other way round ([[ADR-0010 - The JSON Schema is the cross-language contract enforced by writer conformance|ADR-0010]]).
- **Documentation boundaries hold.** The spec (`dcanvas-3.0.md`) and schema describe the format only — no mention of the library, layout, the CLI, or tooling.

# Testing Decisions

- **The conformance test is the deliverable.** It asserts external behaviour — the encoded bytes satisfy the shared schema — never internal representation. This continues the philosophy already in `io_test.go`, where `encodeToMap` asserts on the observable JSON shape; here the assertion is "validates against the schema."
- Cover representative outputs (the spec's examples are a good source: line node, reply node with a Layer 2 field, conditional edge, `x-character`, minimal document). Optionally validate a committed golden fixture (`example.d.canvas`).
- A deliberately schema-invalid canvas (e.g. `text` node with empty text, before the fix) is a useful red test proving the check bites.

# Out of Scope

- **How the reader obtains the schema** (git submodule, vendored copy, published package) — the reader repository's concern. This repository's obligation is only to be a clean, canonical master source.
- **The Go-side `Validate` API and the CLI's round-trip robustness** — owned by [[EPIC-0002 - dCanvas CLI TUI and round-trip library robustness|EPIC-0002]]. (`Validate` is a Go convenience; this epic's conformance test is the cross-language guarantee. They are complementary.)
- **Semver release tagging / `go get` versioning** — deferred until the writer stops depending on the live source via a `go.work`/`replace` and needs a pinned version. The format version (`3.0`) is a separate axis from the Go module version and does not change here.
- **README / godoc examples** — adoption polish; best written once the public API has stabilised under real consumer use.

# Further Notes

- The **format vocabulary is not changed**: it was designed with the new writer in mind, so this is conformance plumbing, not format evolution.
- Files written by the CLI in [[EPIC-0002 - dCanvas CLI TUI and round-trip library robustness|EPIC-0002]] must also conform; the conformance test should cover library output generally so it protects both the writer and the CLI. The two epics can proceed in parallel.
