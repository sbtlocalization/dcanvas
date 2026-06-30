---
adr-status: Accepted
superseded-by:
---

# Keep dCanvas a dialogue-only strict superset of JSON Canvas 1.0

dCanvas describes dialogue graphs only — who says what and how a conversation branches — not generic diagrams. It remains a **strict superset of JSON Canvas 1.0**: all dialogue metadata lives in `x-` fields, so stripping every `x-` yields a valid JSON Canvas document that any editor (notably Obsidian) can open and re-save unchanged.

We chose this because the primary win is free, mature tooling: editing, viewing, and round-tripping come from the JSON Canvas ecosystem rather than a bespoke editor. The cost is that the visible text must conform to JSON Canvas conventions (text in `node.text` / `edge.label`), which constrains how metadata is laid out — an acceptable trade we make deliberately (see [[ADR-0004 - Three-layer x- extension model]]).
