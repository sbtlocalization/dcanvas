---
adr-status: Accepted
superseded-by:
---

# Public versioning restarts at 1.0, and the internal specs are archived as 0.x

The format goes public. Its internal version history (1.0 → 2.0 → 3.0) was never published: the only artifacts stamped with those versions were this repository's own examples and test fixtures, plus legacy files of the originating project. Publishing the format as "3.0" would carry two phantom majors into the public record and invite the question "where are 1.0 and 2.0?" forever.

## Decision

- The never-published internal 3.0 is published as **dCanvas 1.0**. The public history starts clean at 1.0.
- The spec and JSON Schema move out of `docs/` to **`spec/dCanvas-1.0.md`** and **`spec/dCanvas-1.0.schema.json`**, next to `spec/JSON-Canvas-1.0.md` and matching its naming style. The schema's `$id` follows: `https://github.com/sbtlocalization/dcanvas/spec/dCanvas-1.0.schema.json`.
- The two never-published internal specs stay in `docs/` but are renumbered **0.1** and **0.2** — file names and titles only, each with a note that they were historically versioned 1.0/2.0 and never published. The version stamps described in their bodies (`x-dCanvasVersion: "1.0"/"2.0"`) describe real legacy artifacts and are left intact.
- This is a **clean break**: the reader knows only `d-version` (see [[ADR-0014 - Namespace split d- for the format x- for projects|ADR-0014]]) with major 1 and rejects anything else, including files stamped `x-dCanvasVersion` — no 3.0 shim, no migration code. The repository's examples and test fixtures are regenerated; the major-version tolerance fixture becomes a 1.1 file.
- Documentation follows a three-mode split: code, spec, schema, and fixtures carry **no trace** of the old era; `CONTEXT.md` (a living glossary) is rewritten in current terms with history referred to via the internal 0.x numbers; existing ADRs, epics, and tasks are **not touched** — they are the historical record.

## Consequences

- Consumers and search engines only ever see one dCanvas lineage: 1.0 onward. The 0.x archive in `docs/` remains for archaeology.
- The library `Version` constant and every fixture stamp become `1.0`; a file stamped with the pre-1.0 internal version field fails decoding with a missing-`d-version` error, which is the designed behaviour, not an oversight.
- Older ADRs, epics, and tasks keep speaking of "3.0" and `x-` fields; readers should map that era to the published 1.0 via this ADR and [[ADR-0014 - Namespace split d- for the format x- for projects|ADR-0014]].
