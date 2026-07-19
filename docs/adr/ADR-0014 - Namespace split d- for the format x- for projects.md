---
adr-status: Accepted
superseded-by:
---

# Namespace split: d- is the format's vocabulary, x- is the projects' home

Until now the Layer 1 dialogue vocabulary and Layer 2 project extensions shared the `x-` prefix, distinguished only by which fields the spec happened to define. That works, but it gives the format no room to grow safely: any new `x-` field a future format version defines could collide with some project's existing extension, and a reader cannot tell "vocabulary I'm too old to know" from "someone's project data".

## Decision

- The Layer 1 vocabulary moves to a **reserved `d-` prefix**: `d-version` (replacing `x-dCanvasVersion`), `d-id`, `d-kind`, `d-role`, `d-textId`, `d-condition`, `d-action`, `d-sound`, `d-character`.
- The **`d-` namespace belongs to the spec**: unknown `d-*` fields are reserved for future format versions and must never be given a project-specific meaning.
- **`x-` remains the conventional home of Layer 2 project extensions** (e.g. `x-journalText`). The spec never defines an `x-` field.
- Mandatory recursive preservation ([[ADR-0003 - Mandatory recursive preservation of unknown fields|ADR-0003]]) is **prefix-blind** and unchanged: every unknown field is preserved regardless of spelling.

This ADR explicitly **amends — not supersedes** — the following ADRs, whose decisions stand with the new field spellings:

- [[ADR-0002 - Two-axis classification x-kind vs x-role|ADR-0002]] — the closed/open two-axis classification is unchanged; the fields are now spelled `d-kind` and `d-role`.
- [[ADR-0004 - Three-layer x- extension model|ADR-0004]] — the three-layer model is unchanged; Layer 1 is now the `d-` vocabulary while Layer 2 keeps `x-`, so the layer boundary is now visible in the field name itself.
- [[ADR-0009 - Layout may read x-kind and the loop strategy is configurable|ADR-0009]] — layout's default loop handling now reads `d-kind: "loop"`; the decision and both strategies are unchanged.

## Consequences

- A generic tool can now classify any field by prefix alone: `d-` known-or-future format vocabulary, `x-` project data, anything else Layer 0 / unknown-standard-looking. Collisions between future format versions and project extensions are impossible by construction.
- Together with [[ADR-0013 - Public versioning restarts at 1.0 and internal specs are archived as 0.x|ADR-0013]] this is a clean break: there is no reader for the `x-` spelled Layer 1, and the repository's code, spec, schema, and fixtures carry no references to the old spellings.
- Prose in older ADRs, epics, and tasks that names `x-kind`, `x-role`, `x-character`, etc. describes the pre-1.0 internal era; map those names to their `d-` spellings when reading.
