---
status: Done
priority: 1
blocked by:
kind: PRD
---
# Tasks

<!-- live the following block unchanged -->
```base
summaries: {}
filters:
  and:
    - epic == [this.file]
formulas:
  is_blocked: |-
    note["blocked by"].reduce(
      acc || (
        value.asFile().properties.status != "Done" && 
        value.asFile().properties.status != "Obsolete"
      ),
      false
    )
  progress_icon: |-
    if(status == "Done",
      icon("square-check-big"),
      if (status == "Obsolete",
	    icon("square-arrow-right"),
        if(formula.is_blocked,
          icon("construction"), 
          icon("square")
        )
      )
    )
  progress_sort: |-
    if(status == "Done" || status == "Obsolete", 
      20, 
      if(formula.is_blocked, 
        10, 
        if(status == "In progress", 
          5, 
          0)))
  progress_string: |-
    if(status == "Done" || status == "Obsolete", 
      "Done", 
      if(formula.is_blocked, 
        "On hold",
        "Ready"))
  priority: if(priority, icon("tally-" + priority), null)
  kind_icon: |-
    if(kind == "Bug",
      icon('bug'),
      ''
    )
  mode_icon: |-
    if(mode == "HITL",
      icon('speech'),
      ''
    )
properties:
  note.status:
    displayName: status
  file.name:
    displayName: task
  formula.progress_icon:
    displayName: " "
  formula.progress_string:
    displayName: Tasks that are
  formula.kind_icon:
    displayName: " "
  formula.mode_icon:
    displayName: " "
views:
  - type: table
    name: Таблиця
    groupBy:
      property: formula.progress_string
      direction: DESC
    order:
      - formula.progress_icon
      - formula.kind_icon
      - formula.mode_icon
      - file.name
      - formula.priority
      - status
      - blocked by
    sort:
      - property: formula.progress_sort
        direction: ASC
      - property: priority
        direction: DESC
      - property: file.basename
        direction: ASC
    summaries: {}
    columnSize:
      formula.progress_icon: -56
      formula.kind_icon: -1
      formula.mode_icon: -1
      file.name: 435
      note.status: 125
      note.blocked by: 522
```
<!-- end unchanged block -->

# Problem Statement

The library reads and writes dCanvas with `encoding/json` v1, and every Layer 2 and unknown-field guarantee of the format rests on hand-written marshalers: a `map[string]json.RawMessage` catch-all per object, helpers that pop known keys out of it, and an ordered object writer that merges the leftovers back. That is a lot of code whose only job is to teach v1 something v2 does natively. v1 also lets damage through silently, which the format's *Field preservation* and *Literal text* guarantees cannot afford: invalid UTF-8 and lone surrogates in a string are replaced with U+FFFD, a duplicated key silently keeps its last value, and anything after the first JSON value of a file is ignored. Go 1.27 makes `encoding/json/v2` and `encoding/json/jsontext` stable, so staying on v1 is now a choice rather than a necessity.

# Solution

The library and the CLI move to Go 1.27 and use `encoding/json/v2` and `jsontext` for all reading and writing. The behaviour a user sees changes in three deliberate ways, all of them toward the format's own promises. Reading is strict: a file with a duplicated key, invalid UTF-8 or a lone surrogate in a string, or anything after the document is rejected with an error instead of being silently altered. Writing is literal: text is no longer HTML-escaped, so `<`, `>`, `&` and U+2028/2029 appear as written, and an unknown field's value is carried through with its numbers verbatim and its strings equal in value and written literally, instead of being re-escaped. Writing still ends the document with a newline. The release is **v1.1.1**: the module version follows the format version and a library-only change goes into the patch ([[ADR-0016 - The module version follows the format version and library-only changes go into the patch|ADR-0016]]), even though its minimum Go version rises.

# User Stories

1. As a dialogue writer, I want a file with a duplicated key to be rejected with a clear error, so that I never save a dialogue in which one of two conflicting values was silently dropped.
2. As a localization engineer, I want invalid UTF-8 in a text field to be rejected on read, so that a string-table line is never silently rewritten with replacement characters.
3. As a localization engineer, I want a lone surrogate escape in a string to be rejected on read, so that corrupted text is caught at the door instead of travelling into a game.
4. As a dialogue writer, I want a file with anything after the closing brace to be rejected, so that two concatenated documents are not read as the first one only.
5. As a dialogue writer, I want `<`, `>` and `&` in a line to be written as typed, so that a git diff of a dialogue shows the words and not `\u003c` escapes.
6. As a project maintainer, I want an unknown `x-` field to come back from a round-trip with its numbers verbatim and its strings equal in value, so that a Layer 2 project's data is carried through exactly.
7. As a project maintainer, I want unknown fields to keep the order they had in the file, so that re-saving a file does not reshuffle my project's extension fields.
8. As a consumer of the library, I want `Edge.SetExtra` and `Node.SetExtra` to keep ignoring a key that a typed field already owns, so that attaching an extension never produces a document with a duplicated key.
9. As a consumer of the library, I want `Encode` to keep ending the document with a newline, so that written files stay well-behaved in editors and diffs.
10. As a consumer of the library, I want the module version to stay `v1.1.x`, so that the tag still reads as "format 1.1" ([[ADR-0016 - The module version follows the format version and library-only changes go into the patch|ADR-0016]]).
11. As a consumer of the library, I want the release notes to say that the minimum Go version is now 1.27 and that reading is stricter, so that I know why my build or my files started to fail after the update.
12. As a maintainer, I want one ADR to record the minimum Go version, the strict reading and the literal writing together, so that a future reader finds the reason for all three in one place.
13. As a maintainer, I want the hand-written preservation machinery replaced by the declarative unknown-member mechanism of v2, so that there is less code to keep correct.
14. As a CLI user, I want `dcanvas layout` to keep producing the same file apart from the deliberate byte changes, so that a layout run does not move nodes or lose a field.
15. As a contributor, I want CI to build with Go 1.27 taken from `go.mod`, so that the toolchain I test with is the toolchain I declare.

# Implementation Decisions

- **Go version.** Both Go modules, the workspace and CI move to Go 1.27 together. The library cannot be built with an older Go once it imports the stable v2 packages, and the CLI depends on the library, so the CLI follows. CI already reads its Go version from `go.mod`.
- **Versioning.** The release is `v1.1.1`. The major and minor of the tag equal the format version the library writes, and a library-only change bumps the patch ([[ADR-0016 - The module version follows the format version and library-only changes go into the patch|ADR-0016]]). Raising the minimum Go version is accepted as a patch-level change by the maintainer; the release notes say so.
- **Two steps, two commits.** Step A swaps the packages and keeps the hand-written marshalers; step B replaces them with the unknown-member mechanism of v2. Keeping them apart means that a regression can be attributed either to a change of semantics or to the rewrite.
- **Step A: reading.** Reading takes the strict defaults of v2: duplicate member names, invalid UTF-8 and lone surrogates are errors, and so is anything following the document. No compatibility option is enabled. Error messages keep the existing `can't decode dcanvas:` prefix.
- **Step A: writing.** Writing takes the v2 defaults: no HTML escaping, no escaping of U+2028/2029, indented with a tab, and an explicit newline after the document. The raw-value type of v2 replaces the v1 one wherever unknown fields are stored.
- **Step B: preservation.** Unknown members of the canvas, nodes, edges and characters are collected by the embedded fallback of v2 (a `jsontext.Value` field tagged `embed`; the `unknown` tag option of the experimental package does not exist in Go 1.27, and a field tagged with it is written as an ordinary member named after the field while the unknown members are lost, without an error) and written back after the known ones, in the order they had in the file. The previous alphabetical order of unknown fields is given up on purpose. Behaviour that a tag cannot express stays in code: the rule that a `text` node always writes its `text` while other node types omit it when empty, and the guarantee that a key owned by a typed field is never duplicated by an extension.
- **Release and CLI dependency.** The library is tagged first and the CLI then depends on that tag; tagging and pushing are the maintainer's call, as for earlier releases.
- **Record.** Two new ADRs. The first records the minimum Go version, the strict reading, the literal writing and the release as v1.1.1; it builds on [[ADR-0016 - The module version follows the format version and library-only changes go into the patch|ADR-0016]] and supersedes nothing, and it is written in the first task. The second, written with step B, records the preservation through typed fields and the embedded fallback of v2 and supersedes [[ADR-0005 - Reference impl uses catch-all over typed fields|ADR-0005]]. The ADR cross-reference rule of the repository applies: every reference is a wikilink with an alias.
- **Spec and schema untouched.** The format does not change, so neither `spec/` nor the schema is edited; all of this lives in the ADR and the code.

# Testing Decisions

- **Seams, highest first.** (1) The public `Decode` and `Encode` of the library, through the existing `io_test.go` round-trip, key-order and version tests. (2) The schema conformance tests in `conformance_test.go`, which validate encoded output against the schema files. (3) The CLI golden test `TestLayoutCLI_Golden`, which compares the bytes of a laid-out file. No new seam is introduced.
- **A good test here** feeds a document in and checks what comes out of `Decode`, `Encode` or the CLI: the error, the preserved fields, the bytes. It never inspects the internal catch-all maps or the writer.
- **New behaviour is test-first.** Before step A, tests are added for each strict-reading case (duplicate key, invalid UTF-8, lone surrogate, trailing data), for literal writing (`<`, `>`, `&`, U+2028 appear unescaped, also inside an unknown field) and for the trailing newline. They fail on v1 and pass after the swap.
- **Preservation is the safety net for step B.** The existing round-trip tests (`PreservesUnknownFields`, `TypedLayer1Faithful`, `PreservesUnrecognisedKind`, `SetExtra` and `TypedFieldWins`) must pass unchanged in meaning; the key-order test is the one that is rewritten on purpose, from alphabetical to file order, and tests are added for the numbers of an unknown field coming back verbatim, for its strings coming back equal in value, and for an unknown field never disappearing silently.
- **Golden file.** `golden.d.canvas` and any fixture that contains `<`, `>` or `&` are refreshed once, deliberately, in step A, and the diff is reviewed by hand rather than accepted blindly.
- **Toolchain.** The migration is verified on a real Go 1.27 toolchain, not only on 1.25 with the experiment flag; the probes made during the design of this epic ran on 1.25 and may differ in detail from the final release.

# Out of Scope

- Any change to the format, the spec or the schema.
- A mode that reads files with duplicated keys or invalid UTF-8, or that writes HTML-escaped output, for compatibility with v1.
- Changing the JSON handling of the schema validation library, of the CLI's other dependencies or of the test helpers, beyond what the build requires.
- A new major version of the module or a new format minor.
- Keeping the library buildable with Go older than 1.27, for example behind build tags.

# Further Notes

- The v2 packages have different defaults from v1 in places the probes did not touch, notably how a `null` or a number with a fraction is decoded into typed fields; the probes showed no difference for the cases that matter to this format, but the first task should re-run them on Go 1.27.
- The first task is the right place to find out whether any real-world file, for example an export of an Infinity Engine dialogue, is rejected by the strict reading; the maintainer expects none, and chose strictness over a compatibility option on that basis.
- The planned tasks are: the migration with the ADR (step A), the preservation rewrite (step B, blocked by the first), and the v1.1.1 release (HITL, blocked by both). If step B proves problematic it can be dropped from the release, which then waits only for step A.
