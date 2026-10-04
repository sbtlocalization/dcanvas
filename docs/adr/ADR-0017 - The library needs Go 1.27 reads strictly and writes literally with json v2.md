---
adr-status: Accepted
superseded-by:
---

# The library needs Go 1.27, reads strictly and writes literally with json v2

Go 1.27 makes `encoding/json/v2` and `encoding/json/jsontext` stable. With v1 the library let damage through silently, which the format's field preservation ([[ADR-0003 - Mandatory recursive preservation of unknown fields|ADR-0003]]) and literal text cannot afford: invalid UTF-8 and lone surrogates were replaced with U+FFFD, a duplicated key kept its last value, and anything after the first document was ignored. On writing, v1 escaped `<`, `>`, `&` and U+2028/2029, so a git diff of a dialogue showed `\u003c` instead of the words.

The library and the CLI therefore read and write through json v2, and four things follow from it:

- **The minimum Go version is 1.27**, for the library and, since it depends on the library, for the CLI. The stable v2 packages do not exist in older releases.
- **Reading is strict.** `Decode` takes the v2 defaults: a duplicated key, invalid UTF-8 or a lone surrogate in a string, and anything but whitespace after the document are errors, with the usual `can't decode dcanvas:` prefix. No compatibility option is offered. Writing is as strict: text that is not valid UTF-8 is an error rather than a replacement character.
- **Writing is literal.** `Encode` escapes nothing that JSON does not require: `<`, `>`, `&` and U+2028/2029 are written as they are, in known fields and in unknown ones. An unknown field's numbers come back verbatim; its strings come back equal in value, an escape such as `\u003c` being written as its character. The document is indented with a tab and ends with a newline, as before.
- **The release is v1.1.1.** The format does not change, so the module minor stays 1.1 and the change goes into the patch ([[ADR-0016 - The module version follows the format version and library-only changes go into the patch|ADR-0016]]), even though a raised minimum Go version and stricter reading are visible to a consumer. The release notes say so.

The hand-written marshalers are kept in this step; only the package and the raw-value type change, so that a regression can be attributed to the change of semantics and not to a rewrite of the preservation code.

## Considered Options

- **Stay on v1.** Rejected: the silent alterations above contradict the format's own guarantees, and v1's HTML escaping is noise in every diff of a dialogue.
- **v2 with compatibility options** (allowing duplicate names and invalid UTF-8, or keeping HTML escaping). Rejected: a file with a duplicated key or broken text is damaged, and accepting it would carry the damage into a game. No real-world file is expected to be rejected.
- **A minor or major release.** Rejected: the module minor follows the format minor ([[ADR-0016 - The module version follows the format version and library-only changes go into the patch|ADR-0016]]), and the API does not break.
- **Build tags to keep older Go working.** Rejected: two JSON paths would double the code whose only job is preservation.

## Consequences

- A consumer on Go older than 1.27 cannot update past v1.1.0.
- A file that v1 read with a duplicated key, broken UTF-8 or trailing data is now rejected; the error names the problem instead of a silently altered dialogue.
- Files written by the library change once: `<`, `>`, `&` and U+2028/2029 that used to be escaped are now written literally. Escapes inside the strings of an unknown field can also be spelled differently, a solidus written as `\/` becoming `/`, while the value stays equal.
