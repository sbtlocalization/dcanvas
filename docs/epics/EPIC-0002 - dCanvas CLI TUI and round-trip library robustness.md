---
status: Backlog
priority: 4
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

Someone has a `.d.canvas` file and needs to run an operation on it from the command line — first and foremost **auto-layout**: a dialogue authored (or generated) without coordinates is handed to a program to be positioned into a readable top-down tree. There is no such tool today.

Worse, the reference library's **read → operate → write** path — which any such tool must use — has correctness gaps that would silently corrupt a file on round-trip:

- the format version stamp is rewritten to `3.0` on every write, so a `3.x` file is downgraded even when the tool changed nothing but coordinates;
- a file whose `x-kind` value the library does not recognise can be read but then **fails to write back**, so an unremarkable layout run breaks;
- if the layout engine fails, the failure is swallowed silently — the tool would report success while nothing moved;
- a file handed to the tool from an untrusted source (dangling edges, duplicate ids, missing required fields) has no validation entry point, so the tool cannot tell the user *why* a file is unusable.

The library is also welded to one layout engine (`autog`), which is not ideal and which we do not want to be locked into.

# Solution

An in-repo Go **CLI/TUI** (under `cmd/`) that reads a `.d.canvas` file, performs an operation (starting with auto-layout via the existing `Layout`), and writes it back — preserving verbatim everything it did not touch, including project (Layer 2) fields authored by the writer.

Building the CLI is the acceptance test that hardens the library's round-trip path: the version survives, `x-kind` round-trips, unknown/Layer 2 fields survive a real operate step, layout reports failure, and files are validated on load. The layout engine is quarantined into a `layout` subpackage behind an engine-neutral API so it can be replaced later without touching consumers (see [[ADR-0011 - An in-repo reference CLI is a round-tripping consumer and the layout engine stays swappable|ADR-0011]]).

# User Stories

1. As a dialogue author, I want to hand a coordinate-less `.d.canvas` file to the CLI and get back a laid-out file, so that I do not position nodes by hand.
2. As a dialogue author, I want a file I laid out to keep every field I never intended to change, so that my project-specific data is not lost by running a layout.
3. As a dialogue author, I want the version identifier of my file to be unchanged by a layout run, so that a `3.1` file stays a `3.1` file.
4. As a dialogue author, I want the CLI to tell me clearly when my file is not a valid dCanvas (dangling edge, duplicate id, missing required field), so that I can fix the source.
5. As a dialogue author, I want the CLI to report an error when layout could not be computed, so that I never get a "done" that did nothing.
6. As a dialogue author, I want to choose the loop-handling strategy (`LoopCut` default, `LoopDFS`) from the command line, so that cyclic dialogues lay out the way I expect.
7. As a dialogue author, I want the CLI to accept both `.d.canvas` and `.dcanvas` files, so that either naming works.
8. As a dialogue author, I want a TUI mode to preview the graph before writing, so that I can eyeball the layout.
9. As a CLI consumer of the library, I want to read a canvas, mutate only geometry, and write it back without the write step ever rejecting data the read step accepted, so that round-trip is lossless.
10. As a CLI consumer of the library, I want a `Validate` call that reports structural problems (id uniqueness, edge endpoint existence, closed-set `x-kind`, required fields), so that I can gate an operation on a valid input.
11. As a CLI consumer of the library, I want `Layout` to return an error rather than silently recovering, so that I can surface it.
12. As a maintainer, I want the layout engine isolated in its own subpackage so the core format package has no third-party dependencies, so that a future non-layout consumer (or a different engine) is not forced to depend on `autog`.
13. As a maintainer, I want the public `Layout` API to expose no `autog` types, so that replacing the engine does not break callers.
14. As a maintainer, I want the CLI's round-trip exercised by tests over real fixtures, so that a regression in preservation or versioning fails the build.

# Implementation Decisions

- **New `cmd/` tool.** A Go CLI (and TUI mode) in this repository. It composes the existing public library: `Decode` → operation → `Encode`. First operation is auto-layout. Command surface (subcommands vs flags) and TUI framework are a task-level decision.
- **Layout moves to a `layout` subpackage.** `autog` is quarantined there; the root `dcanvas` package (types + IO + preservation) becomes dependency-free. The public `Layout` signature stays engine-neutral (`Layout(c, opts...)`, no `autog` types). No pluggable engine interface yet (deferred until a second engine exists). Per [[ADR-0011 - An in-repo reference CLI is a round-tripping consumer and the layout engine stays swappable|ADR-0011]] and [[ADR-0009 - Layout may read x-kind and the loop strategy is configurable|ADR-0009]].
- **Version round-trip fix.** `Encode` preserves the decoded `x-dCanvasVersion` when its major matches the library's (so `3.1` stays `3.1`); a canvas with no decoded version (hand-built) or a different major is stamped with the library version. Replaces the current "always stamp `3.0`" behaviour and its test.
- **Unify the `x-kind` contract so round-trip is lossless.** `Decode` and `Encode` must agree: a value one accepts, the other must not reject. Recommended resolution — closed-set validation lives in `Validate` (see below), `Decode` stays lenient, and `Encode` stops rejecting unknown `x-kind` (enforcement on write breaks round-trip and is the wrong place). The CLI runs `Validate` on load. Alternative — `Decode` rejects unknown closed-set values up front. This is a task-level decision; the invariant is that `Decode`-then-`Encode` never fails on data the file legitimately contained.
- **`Validate(*Canvas) error`.** Structural/semantic checks: unique node ids, every edge endpoint references an existing node, closed-set `x-kind` values, required fields (`text` on `text` nodes, character `name`). Distinct from the JSON-Schema conformance test in [[EPIC-0003 - JSON Schema as cross-language contract and writer conformance|EPIC-0003]] (that validates encoded bytes against the shared schema; this is a Go-side convenience for consumers). The two are complementary.
- **`Layout` returns an error** instead of `defer recover()` swallowing it. A no-op (empty graph, no layout edges) is a nil-error success, not a failure.
- **Preservation is relied upon, not rebuilt.** The existing catch-all preservation carries Layer 0/1/2 fields through the CLI's operate step; the round-trip through the CLI is its acceptance test.

# Testing Decisions

- **Highest seam — the CLI's observable behaviour.** Golden-file tests: an input `.d.canvas` in, an expected output file out. Assert coordinates changed, unknown/Layer 2 fields survived byte-for-byte in value, and the version stamp is unchanged. A committed representative fixture (the existing `example.d.canvas` can serve) is optional but useful; it can also seed the TS reader's tests.
- **Library round-trip seam.** `Decode → Layout → Encode` preserves unknown top-level/node/edge/character fields, preserves version, moves at least one connected node, and never rejects on write what it accepted on read. Prior art: the round-trip and preservation tests in `io_test.go` and `character_test.go`, and the no-overlap/no-op/cycle tests in `layout_test.go`.
- **`Validate` seam.** Table-driven malformed inputs (dangling edge, duplicate id, unknown `x-kind`, `text` node without `text`, character without `name`) each produce an error; a well-formed canvas produces none.
- Tests assert **external behaviour** (bytes/positions/errors), never the internal catch-all representation — as the existing tests already do via `encodeToMap`.

# Out of Scope

- **A reader accessor for Layer 2 fields** (`Extra` getter) and `SetExtra` symmetry on edges/canvas/character — only warranted once a CLI operation needs to *inspect or edit* project fields. Auto-layout does not; it only preserves them. Deferred until an operation pulls the need.
- **A pluggable `LayoutEngine` interface** — deferred until a second engine exists; engine-neutral API + quarantine suffice now.
- **Operations beyond auto-layout** — layout is the tracer bullet; further operations are separate work.
- **Semver release tagging / `go get` versioning** — during co-development the CLI and library live in one module; tagging is deferred (see [[EPIC-0003 - JSON Schema as cross-language contract and writer conformance|EPIC-0003]] notes).
- **The JSON-Schema conformance test and the schema `$id`** — owned by [[EPIC-0003 - JSON Schema as cross-language contract and writer conformance|EPIC-0003]].

# Further Notes

- Files the CLI writes must also conform to the shared JSON Schema; the conformance test in [[EPIC-0003 - JSON Schema as cross-language contract and writer conformance|EPIC-0003]] should cover CLI/library output, and can reuse `Validate` for structural checks. The two epics can proceed in parallel with no hard blocking dependency.
- This epic is the concrete realisation of the round-trip robustness that [[ADR-0011 - An in-repo reference CLI is a round-tripping consumer and the layout engine stays swappable|ADR-0011]] moved back into scope from [[ADR-0010 - The JSON Schema is the cross-language contract enforced by writer conformance|ADR-0010]]'s deferred backlog.
