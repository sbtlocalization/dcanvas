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

A dialogue node names exactly one speaker. Games do not always know which one: in some games a conversation is owned by whichever actor the level script starts it with, and different scripts start the same conversation with different actors, so one line may be spoken by any of several characters. A writer forced to pick one asserts something the data does not support; a writer that picks none leaves the most common speaker of the game anonymous. Either way a translator inflecting the line for a gendered language does not see every voice it may be spoken in.

A speaker's name is also a translatable string in its own right, but `d-character` gives it no string-table reference the way `d-textId` does for a node's text. A reader cannot tell where the name came from, nor tell apart two characters whose names read the same.

# Solution

dCanvas 1.1 adds two optional fields. `d-character` gains `textId`, the string-table reference for its `name`. A dialogue node gains `d-alternativeCharacters`, a list of further possible speakers beside `d-character`: the line is spoken by exactly one of `d-character` and the alternatives, and `d-character` has no precedence over the rest beyond being the one a reader shows first.

Both fields are additive, so the format moves to 1.1 and every 1.0 reader keeps working: a reader that does not know them preserves them and shows `d-character` as before. The reference library models both as typed fields, the reference CLI's viewer shows them, and the module is released as `v1.1.0`.

# User Stories

1. As a writer, I want to list every possible speaker of a line, so that I do not have to assert one the data does not support.
2. As a writer, I want the alternatives to share the schema of `d-character`, so that each carries a name, portrait and gender in the same shape.
3. As a writer, I want to record the string-table reference of a speaker's name, so that the name can be traced to the string it was resolved from.
4. As a reader, I want `d-character` to stay where it was, so that a reader that knows nothing of alternatives still shows a real speaker.
5. As a reader, I want the spec to say that exactly one of the listed speakers speaks the line, so that I do not render the alternatives as a chorus.
6. As a reader, I want a 1.1 document to validate against the schema of its own minor, so that I can check it structurally.
7. As a reader, I want a 1.1 document that only uses the new fields to still validate against the 1.0 schema, so that the "a minor only adds" promise holds mechanically.
8. As a translator, I want to see every possible speaker of a line in the viewer, so that my translation works for each of them.
9. As a translator, I want to tell two characters with the same name apart by their name's string-table reference.
10. As a library user, I want `textId` and the alternatives as typed fields, so that I do not have to reach for preserved extras.
11. As a library user, I want `Validate` to reject an alternative without a name and alternatives without a `d-character`, so that a malformed document is caught before it is written.
12. As a library user, I want the library to stamp `1.1` and still read every 1.x document, so that upgrading costs nothing.
13. As a maintainer, I want the module tagged `v1.1.0`, so that the tag keeps reading as the format version it writes.

# Implementation Decisions

- **Spec and schema per minor.** `spec/dCanvas-1.1.md` and `spec/dCanvas-1.1.schema.json` are added beside the 1.0 files, which stay as they are: minors differ structurally, and the 1.0 files keep describing 1.0 precisely ([[ADR-0015 - The schema constrains d-version by major while schema files stay per minor|ADR-0015]]). The 1.1 schema constrains `d-version` by major, as the 1.0 one does. CONTEXT.md starts pointing at the 1.1 spec.
- **`d-character.textId`** — optional string, the string-table reference for `name`. Its value format is a project convention, exactly as for `d-textId`. It joins the known fields of the open `d-character` object; any other key is still preserved.
- **`d-alternativeCharacters`** — optional node field, a non-empty list of objects with the schema of `d-character`, `name` required in each. It is a node field beside `d-character`, not a key inside it: the `d-` prefix belongs to node fields, and a character that contained characters would invite alternatives of alternatives. It may appear only together with `d-character`, since the pair means "one of these", and alternatives without a first speaker leave a 1.0 reader with none.
- **Semantics, stated in the spec.** The line is spoken by exactly one of `d-character` and `d-alternativeCharacters`; `d-character` is the one a reader that shows a single speaker shows, and has no other precedence. The order of the alternatives is the writer's; the format attaches no meaning to it.
- **Library.** `Character` gains a typed `TextID`; `Node` gains a typed list of alternative characters, each a `Character` with the same preservation of unknown keys. `Version` becomes `"1.1"`. `Validate` requires a name on every alternative and rejects alternatives on a node without `d-character`.
- **Version tolerance fixture.** The fixture and tests that prove a higher same-major minor is accepted and preserved use 1.1 today; once the library is 1.1 they prove nothing, so they move to 1.2.
- **Viewer.** The reference CLI's detail view shows the name's text ID and every alternative speaker; the tree view keeps showing `d-character` alone.
- **Release.** The module is tagged `v1.1.0`: a new format minor is the one thing that moves the module minor ([[ADR-0016 - The module version follows the format version and library-only changes go into the patch|ADR-0016]]). Tagging and pushing wait for the maintainer.

# Testing Decisions

A good test drives the public surface — encode, decode, `Validate`, the schema — with documents and asserts what a reader or writer would observe, never internals.

- **Round trip** (prior art: the node and character round-trip tests). A document with `textId` and alternatives decodes into the typed fields and encodes back unchanged, unknown keys inside each alternative included.
- **Validate** (prior art: the `d-character` name check). An alternative without a name, and alternatives on a node without `d-character`, are each reported; a well-formed node is not.
- **Writer conformance** (prior art: the conformance harness of [[ADR-0010 - The JSON Schema is the cross-language contract enforced by writer conformance|ADR-0010]]). The library's output validates against the 1.1 schema, and a 1.1 document using both new fields also validates against the 1.0 schema.
- **Version** (prior art: the version contract and same-major tests). The library stamps `1.1`; a 1.2 document is accepted and preserved as 1.2; a 2.0 document is rejected.
- **Viewer** (prior art: the CLI's model tests). The detail view of a node with alternatives lists every speaker and the name's text ID.

# Out of Scope

- Any meaning for the order of the alternatives beyond "shown first".
- Simultaneous speakers — a chorus, a crowd. `d-alternativeCharacters` means "one of", and a line spoken by several at once is a different statement this minor does not make.
- Changes to the 1.0 spec and schema files.
- The TypeScript reader. It is a consumer of the format and adopts 1.1 on its own schedule.
