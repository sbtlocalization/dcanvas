---
adr-status: Superseded
superseded-by: "[[ADR-0009 - Layout may read x-kind and the loop strategy is configurable]]"
---

# Graph layout (autog) belongs to the dCanvas library

Auto-positioning of dialogue nodes (currently `autog` in `dialog/format.go`) becomes part of the dCanvas reference library, exposed as a generic operation over a canvas.

Today the layout call is tangled inside `dialog.ToDCanvas`, mixed with engine-specific conversion (colors, characters, triggers). But layout only ever reads/writes Layer 0 geometry — node `width`/`height`/`x`/`y` and edge `fromNode`/`toNode`. It is therefore generic and reusable, and belongs with the format, not with one engine's conversion code. The engine-specific conversion (`newNode`/`newEdge`/`newCharacter`) stays in the consuming project.

The layout engine resolves cycles on its own, so `x-kind: "loop"` edges are **not** excluded from layout (see [[ADR-0002 - Two-axis classification x-kind vs x-role]]).
