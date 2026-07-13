---
adr-status: Accepted
superseded-by:
---

# The JSON Schema is the cross-language contract, enforced by writer conformance

Fulfils the "Later" stage of [[ADR-0008 - Extract format and library to a separate repository|ADR-0008]]: a second consumer now exists, so the format moves to its own repository — **this** repository becomes the canonical master source of the dCanvas format (spec, JSON Schema, and the Go reference library).

The format is consumed across two language ecosystems with a one-directional data flow:

- a **writer** (Go, write-only) generates `.d.canvas` files via the Go reference library;
- a **reader** (TypeScript + Svelte, read-only) opens and displays them.

The reader cannot use the Go library at all. The only thing both sides share is the **format itself**. Therefore the cross-language contract is the **JSON Schema** (`docs/dcanvas-3.0.schema.json`), not the Go library — the Go library is merely the writer's reference implementation.

The risk in a two-implementation, two-language setup is silent drift: the Go writer emitting something the reader's schema validation rejects. Nothing currently couples the library code to the schema text, so they can diverge unnoticed until a reader user hits a broken file.

## Decision

1. **The JSON Schema is the contract.** Both implementations are judged against it. The Go library does not define the format; the schema does.
2. **Writer conformance is enforced in CI.** A test takes representative output of the Go writer and validates it against the very schema the reader consumes, so any divergence (e.g. a `text` node emitted without `text`) fails the build rather than the reader's user. The schema validator is a **test-only** dependency; the writer's runtime stays standard-library-only.
3. **This repository is the canonical master.** The schema's `$id` resolves to this repository, and the format ships as **versioned releases** (`v0.x` tags) so consumers pin a format version rather than a floating `HEAD`.

How the reader *obtains* the schema (submodule, vendored copy, package) is the reader repository's concern, out of scope here. This repository's only obligation is to be a clean, canonical, versioned source.

## Consequences

- The format vocabulary is **not** changed: it was already designed with the new writer in mind, so this work is conformance plumbing, not format evolution.
- Findings that only affect a hypothetical **Go round-tripping consumer** (isolating the `autog` dependency, an `Extra` read accessor, minor-version round-trip, `Decode`/`Encode` `x-kind` symmetry, the unknown-field preservation machinery) have **no consumer** in this topology. They remain a deferred backlog, pulled only when a consumer actually exercises them — consistent with [[ADR-0008 - Extract format and library to a separate repository|ADR-0008]]'s rule against validating the format in a vacuum.
- Unknown-field preservation is now exercised only by the **Obsidian round-trip** path, not by either of this project's two tools.
