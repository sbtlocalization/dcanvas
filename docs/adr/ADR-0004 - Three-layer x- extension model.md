---
adr-status: Accepted
superseded-by:
---

# Organise fields into three layers over the x- namespace

Every field belongs to exactly one layer:

- **Layer 0 — JSON Canvas** (`id`, `x`, `y`, `text`, edge `label`, …): the universal base.
- **Layer 1 — dialogue vocabulary** (the `x-` fields defined in the spec: `x-kind`, `x-role`, `x-character`, `x-condition`, `x-action`, `x-sound`, `x-textId`, `x-id`): the shared language of any dialogue project.
- **Layer 2 — project extensions** (any other `x-` field, e.g. Infinity's `x-journalText`): specific to one engine; the core spec does not know them.

The 2.0 format baked Layer 2 fields (journal text/sound) into the core, which is exactly why it didn't fit a second project. Separating the layers — with Layer 1 as a reusable dialogue vocabulary and Layer 2 left entirely to projects — is the structural fix. Combined with mandatory preservation ([[ADR-0003 - Mandatory recursive preservation of unknown fields]]), a tool that understands only Layers 0+1 can handle any project's file correctly.
