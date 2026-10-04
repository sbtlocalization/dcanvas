---
adr-status: Accepted
supersedes: "[[ADR-0005 - Reference impl uses catch-all over typed fields|ADR-0005]]"
superseded-by:
---

# Unknown fields are kept by a json v2 embedded fallback over typed fields

[[ADR-0005 - Reference impl uses catch-all over typed fields|ADR-0005]] decoded every object into a map of raw values, popped the known keys out of it into typed fields, and merged the rest back with an ordered object writer. That was the only way to preserve unknown fields ([[ADR-0003 - Mandatory recursive preservation of unknown fields|ADR-0003]]) with json v1. With json v2 ([[ADR-0017 - The library needs Go 1.27 reads strictly and writes literally with json v2|ADR-0017]]) the library can let the JSON package do it.

The canvas, nodes, edges and characters therefore describe their known fields with `json` tags, and keep everything else in an embedded fallback: a `jsontext.Value` field tagged `embed`, which json v2 fills with the members no typed field owns and writes back after the known ones. The `MarshalJSON` and `UnmarshalJSON` methods are kept and do this work, so the public API does not change. The `unknown` tag option of the experimental package does not exist in Go 1.27: a field tagged with it is written as an ordinary member named after the field and the unknown members are lost, without an error.

Four things follow from it:

- **Unknown fields keep the order of the file.** They used to be written in alphabetical order; a raw object keeps them as they were read, where a map fallback would sort them.
- **What a tag cannot express stays in code.** A `text` node always writes its `text`, while other node types omit it when it is empty.
- **`SetExtra` ignores a key owned by a typed field, even when that field is empty.** json v2 rejects a duplicated key on writing, where v1 let the typed field win, so such a key is never stored. Before, a key whose typed field was empty and so not written did reach the output, and became the typed value on the next read; now the typed field wins by being absent. A key that is not valid UTF-8 is an error.
- **`"d-character": null` is dropped.** It used to be written back as an object with an empty name; now it is read as no character and not written.

## Considered Options

- **Keep the catch-all map and the ordered writer on json v2.** Rejected: it is a lot of code whose only job is to teach the JSON package what v2 does natively, and it cannot keep the order of the file.
- **A `map[string]jsontext.Value` fallback.** Rejected: it writes the unknown fields sorted, not in the order of the file.
- **An exported field for the unknown members.** Rejected: json v2 accepts tags only on exported fields, but an exported field changes the public types. Each type is read and written through a copy of itself without the JSON methods, wrapped together with the fallback.

## Consequences

- Re-saving a file no longer reorders its unknown fields; a file written by an earlier version has them reordered once, from alphabetical to the order of the file.
- Error messages for a malformed known field name it by its JSON pointer, as json v2 reports it.
