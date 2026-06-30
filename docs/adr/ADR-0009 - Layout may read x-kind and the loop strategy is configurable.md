---
adr-status: Accepted
supersedes: "[[ADR-0006 - Layout belongs to the dCanvas library]]"
superseded-by:
---

# Layout may read x-kind, and the loop strategy is configurable

Supersedes [[ADR-0006 - Layout belongs to the dCanvas library]] and amends a consequence of [[ADR-0002 - Two-axis classification x-kind vs x-role]].

[[ADR-0006 - Layout belongs to the dCanvas library|ADR-0006]] placed graph layout in the dCanvas library (still true) and decided two things that turned out wrong in practice:

1. **"Layout reads only Layer 0 geometry."**
2. **"Loop edges are not excluded; the layout engine resolves cycles itself."**

The reference implementation (`autog`) does not *remove* a cycle — it **reverses** a back-edge and routes it back up through every layer, inserting virtual nodes. So feeding `x-kind:loop` back-edges to the engine pushes their endpoints sideways and tangles the dialogue tree (e.g. a player reply whose only extra edge loops to the start gets shoved far to the right). The original `sbt-infinity` export never had this problem because it **excluded loop edges from layout entirely** and only ever laid out the acyclic subgraph — it never relied on the engine resolving cycles.

dCanvas is a *dialogue* format, not plain JSON Canvas, and the library owns both the format and its layout. There is therefore no reason to forbid layout from using Layer 1: knowing which edges are loops is exactly the information needed to lay a dialogue out well.

## Decision

`Layout` takes options and handles loop edges by a selectable strategy:

- **`LoopCut` (default)** — drop edges whose `x-kind` is `loop` from the graph fed to the engine. The remaining acyclic graph lays out as a clean top-down tree with the entry node on top. This reads Layer 1 (`x-kind`) deliberately. Loop edges remain in the encoded canvas; they are ignored only while computing positions.
- **`LoopDFS`** — keep every edge and break cycles with a depth-first edge reversal (`autog`'s DFS cycle breaker, which reverses true back-edges rather than the greedy default that tangles dialogues). Loop edges then influence positions.

Both strategies use depth-first cycle breaking underneath (a safety net for any residual cycle under `LoopCut`). Node sizing stays **per-node**, read from each node's `Width`/`Height`; the consuming project controls sizes by setting them on nodes (the original used a uniform 400×300 by doing exactly that).

Layout remains owned by the library and stays an optional operation (a canvas may be encoded with hand-set positions without calling it) — the parts of [[ADR-0006 - Layout belongs to the dCanvas library|ADR-0006]] that are unaffected.

## Consequences

- [[ADR-0006 - Layout belongs to the dCanvas library|ADR-0006]]'s "layout reads only Layer 0" no longer holds: `LoopCut` reads `x-kind`.
- This amends [[ADR-0002 - Two-axis classification x-kind vs x-role|ADR-0002]]'s stated consequence that *nothing mechanical* depends on `x-kind`. Layout's default `LoopCut` now depends on `x-kind:loop`. `x-kind` is still a **closed** vocabulary, but it is no longer *purely* a rendering hint — one of its values steers layout. [[ADR-0002 - Two-axis classification x-kind vs x-role|ADR-0002]]'s core decision (the two-axis `x-kind`/`x-role` split) is unchanged and not superseded.
- The spec (`dcanvas-3.0.md`) and `CONTEXT.md` still carry the old "layout does not exclude loops" wording; those are documentation of behaviour, not the format itself, and should be reconciled with this ADR separately.
