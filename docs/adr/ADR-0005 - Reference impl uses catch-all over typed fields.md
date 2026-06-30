---
adr-status: Accepted
superseded-by:
---

# Reference implementation uses a catch-all map over typed fields, not struct embedding

We initially planned to model the layers with Go struct embedding (Layer 0 struct embedded by a Layer 1 struct embedded by a project struct), relying on `encoding/json`'s automatic field promotion to flatten everything to top level.

Mandatory recursive preservation ([[ADR-0003 - Mandatory recursive preservation of unknown fields]]) breaks that plan: capturing unknown fields requires a custom `UnmarshalJSON`, and once any embedded layer defines `UnmarshalJSON`, Go's automatic field promotion for marshaling stops working — the outer layer's own fields stop flattening cleanly. Embedding + catch-all do not compose.

So the reference implementation will instead decode into a `map[string]json.RawMessage`, populate the typed known fields from it, and keep the remaining (unknown) keys in a catch-all map; on encode it merges typed fields with the preserved map. The catch-all is recursive for known nested objects (currently only `x-character`).

## Consequences

This is purely a reference-implementation concern; the **format** is unaffected. A future reader who expects clean struct embedding should know it was tried and rejected for the reason above.
