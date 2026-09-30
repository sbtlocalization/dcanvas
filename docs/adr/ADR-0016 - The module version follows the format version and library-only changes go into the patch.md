---
adr-status: Accepted
superseded-by:
---

# The module version follows the format version, and library-only changes go into the patch

The Go module and the format share one repository and one tag, and the tag `v1.0.0` reads naturally as "dCanvas 1.0". The moment the library grows an API the format does not need — `Edge.SetExtra`, for one — strict semver asks for `v1.1.0`, and a reader seeing that tag next to `d-version: "1.0"` would take it for a format 1.1 that does not exist.

The major and minor of the module tag therefore always equal the format version the library writes (its `Version` constant), and a release that changes only the library, additions to its API included, bumps the patch. A new format minor is the only thing that moves the module minor.

## Considered Options

- **Strict semver with a note in the README** that the module version is not the format version. Rejected: the tag is read far more often than the README, and the confusion it causes is exactly the one the note would have to undo.
