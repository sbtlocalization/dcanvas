---
adr-status: Accepted
superseded-by:
---

# The schema constrains d-version by major, while schema files stay per minor

Amends a consequence of [[ADR-0010 - The JSON Schema is the cross-language contract enforced by writer conformance|ADR-0010]]: the contract artifact now constrains the version field by **major**, not by the exact minor the schema file was written for.

`spec/dCanvas-1.0.schema.json` pinned `d-version` with `"const": "1.0"`. That reads correctly from one side of the contract and wrongly from the other, because the schema serves two consumers asking two different questions:

- The **writer** asks *"does what I just emitted match the version I target?"* — for that question `const` is exactly right.
- The **reader** asks *"is this file structurally something I understand?"* — for that question `const` produced a false rejection.

The spec's *Version detection* section and [[ADR-0013 - Public versioning restarts at 1.0 and internal specs are archived as 0.x|ADR-0013]] both say the **major** governs compatibility: a same-major document may add fields, and those fields survive untouched through preservation. The Go reference library implements exactly that — it decodes a `d-version: "1.1"` document without complaint (`io_test.go`, fixture `cmd/dcanvas/testdata/version-1.1.d.canvas`). The schema rejected that same document. So the artifact designated as *the* cross-language contract contradicted both the spec it encodes and the reference implementation it exists to keep honest, and every reader was left to invent its own programmatic loosening of `const` in code — the compatibility rule re-implemented once per language, which is precisely what a shared contract is meant to prevent.

## Decision

- `d-version` is constrained by **pattern, not constant**: `"pattern": "^1\\.\\d+$"`. The 1.0 schema accepts any 1.x document.
- **Schema files stay one per minor.** Minors differ structurally, since a minor may add fields, so `spec/dCanvas-1.0.schema.json` keeps describing the 1.0 structure precisely. Only the constraint on the *version field* is widened to the major. Files per minor, version constraint per major.
- The compatibility rule therefore lives **inside the contract artifact**, not only in the spec's prose. A reader validates and gets the spec's answer; it does not re-implement the rule.
- The writer's obligation to stamp its own version is asserted where it belongs — in the writer's unit test on the version constant (`TestEncode_VersionContract` in `io_test.go`) — rather than through schema validation.

## Consequences

- The "a minor only adds" promise becomes **checkable** instead of merely stated: a 1.1 document that only added fields validates against the 1.0 schema, and that validation is the mechanical proof of structural compatibility.
- A 1.1 document that added a value to a **closed** vocabulary — a new `d-kind`, say — still fails the 1.0 schema. That is the correct signal rather than a defect: it tells the reader "this minor knows a value I do not", and the reader decides what to do with it.
- The trade-off, stated plainly: writer conformance no longer catches a 1.0 writer that stamps `d-version: "1.3"`. One schema cannot both accept every compatible file and pin one writer's own output. That check is deliberately relocated to the writer's unit test on the version constant — cheaper, and in the right place.
- [[ADR-0010 - The JSON Schema is the cross-language contract enforced by writer conformance|ADR-0010]]'s core decision is unchanged and **not** superseded: the JSON Schema is still the cross-language contract, and writer conformance is still enforced in CI. Only the shape of one constraint inside it changes.
