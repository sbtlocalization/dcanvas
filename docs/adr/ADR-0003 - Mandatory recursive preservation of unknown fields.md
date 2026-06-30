---
adr-status: Accepted
superseded-by:
---

# Require recursive preservation of unrecognised fields on round-trip

A conforming implementation **MUST** preserve, on read → write, every field it does not recognise, at **any depth of nesting** (including unknown keys inside `x-character`). This applies to all unknown fields, not just `x-` ones.

This is what makes dCanvas genuinely *extensible* rather than merely having a fixed extension set — the original pain point that triggered the 3.0 redesign was a format that didn't fit a second project. It also matches how JSON Canvas editors (Obsidian) already treat unknown properties: they leave them untouched. We preserve *all* unknown fields (not only `x-`) so that future JSON Canvas standard additions also survive a dCanvas round-trip.

## Consequences

This is the reason the reference implementation cannot rely on plain Go struct embedding — see [[ADR-0005 - Reference impl uses catch-all over typed fields]].
