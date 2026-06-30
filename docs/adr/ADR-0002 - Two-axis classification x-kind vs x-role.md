---
adr-status: Accepted
superseded-by:
---

# Classify nodes and edges on two axes: x-kind (closed) and x-role (open)

dCanvas 2.0 had a single `x-nodeRole` with a fixed enum (`state`/`transition`), which leaked one engine's vocabulary into the format and didn't fit other projects. 3.0 splits classification into two fields:

- **`x-kind`** — a **closed** vocabulary every tool understands, used purely as a rendering hint. Nodes: `line`/`reply`. Edges: `normal`/`loop`.
- **`x-role`** — an **open** free string, defined per project (e.g. `state`, `transition`, `paraphrase`).

We picked `line`/`reply` over `state`/`transition` (engine-internal) and `entry`/`reply` (`entry` too vague). `line` covers any speaker (NPC, narrator, object), not just NPCs.

We rejected renaming `x-kind` → `x-type` (collides with the standard JSON Canvas `type` field) and kept `x-role` over `x-category` (`kind`+`role` reads as a cleaner closed/open pair).

## Consequences

Nothing *mechanical* depends on `x-kind` — it is a rendering hint only. In particular, edge `loop` is **not** used to exclude back-edges from layout; the layout engine resolves cycles itself (see [[ADR-0006 - Layout belongs to the dCanvas library]]).
