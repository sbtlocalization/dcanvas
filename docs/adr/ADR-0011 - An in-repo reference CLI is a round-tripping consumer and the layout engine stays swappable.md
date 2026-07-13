---
adr-status: Accepted
superseded-by:
---

# An in-repo reference CLI is a round-tripping consumer, and the layout engine stays swappable

A Go CLI/TUI lives in `cmd/` of this repository. It **reads** a `.d.canvas` file, **operates** on it (first operation: auto-layout — a file authored without coordinates is handed to the tool to be positioned), and **writes** it back.

This makes the CLI the first consumer that does a full **read → operate → write** round-trip, unlike the write-only writer and the read-only reader.

## Decision

- The reference CLI/TUI ships **in this repository** under `cmd/`, alongside the format spec and reference library. It doubles as the library's living round-trip acceptance test.
- The layout engine (`autog`) is **quarantined into a `layout` subpackage** so the core format package keeps zero third-party dependencies, and the public `Layout` API stays **engine-neutral** (no `autog` types in its signatures) so the engine can be replaced without breaking consumers.
- We deliberately do **not** build a pluggable `LayoutEngine` interface now. An engine-neutral API plus quarantine gives swap-freedom cheaply; a formal abstraction waits until a second engine actually exists (same "don't build in a vacuum" reasoning as [[ADR-0008 - Extract format and library to a separate repository|ADR-0008]]).

## Consequences

- **Amends [[ADR-0010 - The JSON Schema is the cross-language contract enforced by writer conformance|ADR-0010]].** ADR-0010's consequence said the round-trip and unknown-field preservation machinery had "no consumer in this topology" and that findings like minor-version round-trip and `Decode`/`Encode` `x-kind` symmetry were purely a deferred backlog. The CLI is that consumer: it exercises `Decode`, unknown-field preservation, and version round-trip for real. Those findings are now in scope (see the CLI epic), not deferred.
- The library gains motivated work: a `Validate` API (the CLI is handed files it did not produce), and `Layout` returning an error instead of silently recovering (the CLI must report a failed layout to the user).
- Preservation of unrecognised fields (see [[ADR-0003 - Mandatory recursive preservation of unknown fields|ADR-0003]], [[ADR-0005 - Reference impl uses catch-all over typed fields|ADR-0005]]) is now exercised by a first-party Go tool, not only the Obsidian round-trip path.
- Layout ownership and the loop-strategy design ([[ADR-0009 - Layout may read x-kind and the loop strategy is configurable|ADR-0009]]) are unchanged; only the engine's packaging (a subpackage) is decided here.
