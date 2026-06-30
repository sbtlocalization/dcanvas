---
adr-status: Accepted
superseded-by:
---

# Attach dialogue semantics only to text nodes; preserve other JSON Canvas node types opaquely

Of the four JSON Canvas node types (`text`, `file`, `link`, `group`), only **`text`** nodes carry dialogue semantics (Layer 1 fields). `group`/`file`/`link` nodes are valid and **preserved**, but opaque to dialogue logic — no `x-kind`/`x-role`, not laid out as dialogue nodes. A full viewer should still render them per JSON Canvas.

We rejected making `group` a first-class dialogue construct now. It is the obvious future use (clustering a sub-dialogue, replacing today's per-source color coding), but currently **one file holds exactly one graph**, so grouping is unneeded. Treating non-text nodes as preserved-but-opaque is free (it falls out of [[ADR-0003 - Mandatory recursive preservation of unknown fields]]) and leaves the door open without adding nesting/bounds complexity to layout.
