---
adr-status: Accepted
superseded-by:
---

# Extract the dCanvas format and reference library to a separate repository (staged)

The dCanvas format and its Go reference library will live in their own repository, independent of `sbt-infinity`, so future projects can depend on the format without depending on this tool.

We deliberately **stage** this rather than doing it now:

1. **Now** — capture the design only: the 3.0 spec, JSON Schema, this CONTEXT, and these ADRs, kept in `dcanvas/docs/` as a staging area. `sbt-infinity` is **not** touched and keeps emitting 2.0.
2. **Later** — when a second, concrete consumer exists, create the repository, build the reference library, and let that consumer (or a port of `sbt-infinity`) be the format's acceptance test.

The reasoning: extracting code is hard to reverse, and a format validated only in a vacuum risks repeating the very mistake that triggered 3.0 (a format that didn't fit a real second use). The format design is independent of where the code lives, so we lose nothing by deferring the move. dCanvas 2.0 is deprecated; no migration code will be written.
