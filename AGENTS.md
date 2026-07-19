# Agent guidelines — dCanvas

## Documentation boundaries (strict)

**The specification is the specification.** `spec/dCanvas-1.0.md` and `spec/dCanvas-1.0.schema.json` describe the **format only** — fields, types, values, structure, semantics. They MUST NOT mention the reference library, layout, `autog`, positioning, "configurable", or any other implementation or tooling detail. If a sentence describes what *code* does rather than what the *format is*, it does not belong in the spec or schema.

Implementation and design rationale live in **ADRs** (`docs/adr/`), not in the spec.

## ADRs

- Do **not** rewrite the decision of an Accepted ADR. To change a decision, write a **new** ADR and mark the old one `adr-status: Superseded` with `superseded-by: "[[<new ADR>]]"`. The new ADR gets `supersedes: "[[<old ADR>]]"`.
- If only a *consequence* of an ADR changes (not its core decision), the new ADR amends it explicitly in prose; do not supersede the whole ADR.

## Cross-references in Obsidian docs

Every ADR reference in `CONTEXT.md`, `docs/epics/`, `docs/tasks/`, and `docs/adr/` MUST be a wikilink with an alias:
`[[ADR-0009 - Layout may read x-kind and the loop strategy is configurable|ADR-0009]]`
Never a bare `ADR-0009` in prose. (The spec/schema are standalone and reference no ADRs at all — see above.)

## Format conventions

- Default dCanvas file extension is **`.d.canvas`** (the `.dcanvas` form is the compact alternative).
- License is **BlueOak-1.0.0**; keep SPDX headers consistent.
