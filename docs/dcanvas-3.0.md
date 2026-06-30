# dCanvas Spec

<small>Version 3.0 — 2026-06-25</small>

## Overview

dCanvas 3.0 is a **dialogue** graph format. It is a **strict superset of [JSON Canvas 1.0](https://jsoncanvas.org/)**: all dialogue metadata lives in `x-` prefixed extension fields. Stripping every `x-` field from a dCanvas 3.0 file produces a valid JSON Canvas 1.0 document.

dCanvas is **dialogue-only by design** — it describes who says what and how the conversation branches. It is not a general-purpose diagram format.

### Design principles

1. **JSON Canvas first** — the visible text always lives in the standard JSON Canvas fields: a node's `text` and an edge's `label`. No metadata is ever encoded inside `text`.
2. **Three layers of meaning** — every field belongs to exactly one layer:
   - **Layer 0 — JSON Canvas** (`id`, `type`, `x`, `y`, `width`, `height`, `color`, `text`, edge `label`, …): the universal base.
   - **Layer 1 — dialogue vocabulary** (the `x-` fields defined in this spec): the shared language of any dialogue project.
   - **Layer 2 — project extensions** (any other `x-` field): specific to one game/engine. The core spec does not know them.
3. **Open and extensible** — `x-role` is a free string, `x-character` is open, and projects add their own `x-` fields freely. A generic dCanvas tool need not understand Layer 2 to handle a file correctly.
4. **Round-trip safe (mandatory)** — see [Field preservation](#field-preservation).

## File extensions

dCanvas files use one of two equivalent extensions:

- `.dcanvas` — compact form
- `.d.canvas` — alternative form that preserves the `.canvas` suffix

Both are treated identically. Consumers should accept either. When both exist for the same dialogue, `.d.canvas` takes priority.

## Version detection

A dCanvas 3.0 file **MUST** contain the top-level property:

```json
"x-dCanvasVersion": "3.0"
```

The value is `major.minor`. The **major** version governs compatibility:

- **Same major** → compatible. A higher minor may add fields; they survive untouched via [field preservation](#field-preservation).
- **Different major** → incompatible. A consumer **MUST NOT** silently half-interpret the file; it should reject it with an error.

If the field is absent, the file is plain JSON Canvas 1.0 and carries no dialogue semantics.

> **dCanvas 2.0 is deprecated.** There is no migration path; 3.0 is a clean redesign. A 3.0 reader rejects a 2.0 file on the major-version check.

## Field preservation

> **A conforming implementation MUST preserve, on round-trip (read → write), every field it does not recognise — at any depth of nesting.**

This is what makes dCanvas extensible rather than merely having a fixed set of extensions. Concretely:

- Unknown top-level, node, and edge fields are preserved (both unknown `x-` fields and any unknown standard-looking field).
- Preservation is **recursive**: unknown keys inside known objects (e.g. inside `x-character`) are preserved too.
- A tool that understands only Layer 0+1 must not destroy a Layer 2 project's data when it re-saves a file.

## Top-level structure

```json
{
  "x-dCanvasVersion": "3.0",
  "nodes": [],
  "edges": []
}
```

| Property | Required | Type | Description |
|---|---|---|---|
| `x-dCanvasVersion` | **yes** | string (const `"3.0"`) | Format version identifier |
| `nodes` | no | array of nodes | Canvas nodes in ascending z-index order |
| `edges` | no | array of edges | Connections between nodes |

By current convention **one dCanvas file contains exactly one dialogue graph.** Grouping multiple graphs via `group` nodes is allowed by JSON Canvas but is not used today (see [Node types](#node-types)).

## Nodes

### Standard JSON Canvas fields (Layer 0)

All standard JSON Canvas 1.0 node fields apply unchanged. Dialogue semantics attach only to `text` nodes (see [Node types](#node-types)).

| Field | Required | Type | Description |
|---|---|---|---|
| `id` | yes | string | Unique node identifier **within the file**. Edges reference it via `fromNode`/`toNode`. This is the *canvas* id, not the engine id — see `x-id`. |
| `type` | yes | string | Node type — `"text"` for dialogue nodes |
| `x`, `y` | yes | integer | Position in pixels |
| `width`, `height` | yes | integer | Size in pixels |
| `color` | no | `canvasColor` | Node color (hex string or preset `"1"`–`"6"`) |
| `text` | yes (for `text` nodes) | string | **The spoken content of this node** — clean text only. For a `line` node, what the character says; for a `reply` node, what the player says. |

### Dialogue fields (Layer 1)

All optional and `x-` prefixed. A `text` node that carries `x-kind` is a dialogue node; a `text` node without it is a plain annotation card and is preserved as-is.

| Field | Type | Description |
|---|---|---|
| `x-id` | string | **Engine/domain** identifier of this node (e.g. `"ABELA[5]"`). Distinct from the canvas `id`. |
| `x-kind` | `"line"` \| `"reply"` | **Closed, rendering hint.** `line` = an utterance spoken *to* the player (NPC, narrator, object). `reply` = the player's own utterance. Both values are purely presentational. |
| `x-role` | string | **Open, project-defined** semantic label (e.g. `"state"`, `"transition"`, `"paraphrase"`). Generic tools need not understand it. |
| `x-textId` | string | String-table reference for `text` (e.g. `"#2687"`) |
| `x-condition` | string | Engine condition/trigger gating this node |
| `x-action` | string | Engine action executed when this node is reached |
| `x-sound` | string | Sound resource for this node's spoken line (may include path/extension) |
| `x-character` | object | Speaker information — see below |

#### `x-character` object

Open object — projects may add keys (preserved recursively).

| Field | Required | Type | Description |
|---|---|---|---|
| `name` | yes | string | Character display name |
| `portrait` | no | string | Portrait image filename; empty or `"None.png"` means none |
| `gender` | no | string | Speaker gender (project-defined vocabulary) |

### Project fields (Layer 2)

Anything else `x-` prefixed. The core spec does not define them; they are valid and **preserved**. Example — the Infinity Engine port adds `x-journalText`, `x-journalTextId`, `x-journalSound` to `reply` nodes. A generic dCanvas tool carries these through without understanding them.

## Edges

### Standard JSON Canvas fields (Layer 0)

| Field | Required | Type | Description |
|---|---|---|---|
| `id` | yes | string | Unique edge identifier within the file |
| `fromNode` | yes | string | Source node `id` |
| `fromSide` | no | string | `"top"` \| `"right"` \| `"bottom"` \| `"left"` |
| `fromEnd` | no | string | `"none"` (default) \| `"arrow"` |
| `toNode` | yes | string | Destination node `id` |
| `toSide` | no | string | Side of destination node |
| `toEnd` | no | string | `"arrow"` (default) \| `"none"` |
| `color` | no | `canvasColor` | Edge color |
| `label` | no | string | **The player-facing choice text** — what the player sees and picks. When the choice is a shortened paraphrase of a `reply` node, the short form lives here; when choice and reply coincide, it may duplicate the reply text or be omitted. |

### Dialogue fields (Layer 1)

| Field | Type | Description |
|---|---|---|
| `x-id` | string | Engine/domain identifier of this edge |
| `x-kind` | `"normal"` \| `"loop"` | **Closed, rendering hint.** `loop` marks a back-edge (a cycle in the dialogue); `normal` is a forward transition. May be drawn differently (e.g. `loop` dashed). |
| `x-role` | string | Open, project-defined label |
| `x-condition` | string | Engine condition expression gating this transition |
| `x-textId` | string | String-table reference for `label`, when present |

## Text placement rule

This is the heart of JSON-Canvas compatibility — the two visible texts ride on standard fields:

- **What the character says** → the **node** `text` (a `line` node).
- **What the player says** (full reply) → the **node** `text` (a `reply` node).
- **What the player sees and picks** (the choice/paraphrase) → the **edge** `label`.

Because both are standard JSON Canvas fields, the visible text needs no dialogue-specific interpretation to display.

## Node types

JSON Canvas defines four node types: `text`, `file`, `link`, `group`.

- Only **`text`** nodes carry dialogue semantics (Layer 1 fields).
- `group`, `file`, and `link` nodes are **valid and preserved** but carry no dialogue semantics (no `x-kind`/`x-role`).
- `group` is reserved for future clustering of sub-dialogues; it is unused today because one file holds one graph.

## Color

Identical to JSON Canvas 1.0. `canvasColor` is either a hex string (`"#FF0000"`) or a preset:

`"1"` red · `"2"` orange · `"3"` yellow · `"4"` green · `"5"` cyan · `"6"` purple

## Examples

### Line node (character speaks)

```json
{
  "id": "n-ABELA-5",
  "type": "text",
  "x": 0, "y": 0, "width": 400, "height": 300,
  "color": "3",
  "text": "My poor Ragefast. Like many humans he could not understand the feelings my kind elicits...",
  "x-id": "ABELA[5]",
  "x-kind": "line",
  "x-role": "state",
  "x-character": { "name": "Abela the Nymph", "portrait": "None.png" },
  "x-textId": "#2687",
  "x-condition": "Global(\"RagefastDead\",\"GLOBAL\",0)\r\nDead(\"Ragefast\")"
}
```

### Reply node with a project (Layer 2) extension

```json
{
  "id": "n-ABELA-8",
  "type": "text",
  "x": 920, "y": 2000, "width": 400, "height": 300,
  "color": "1",
  "text": "Be free, then.",
  "x-id": "ABELA[8]",
  "x-kind": "reply",
  "x-role": "transition",
  "x-action": "DialogueInterrupt(FALSE)\r\nGiveItem(\"CLCK07\",LastTalkedToBy)",
  "x-journalText": "The Captive Nymph\nAbela is free from her confinement at last.",
  "x-journalTextId": "#26807"
}
```

### Edge: player choice with a condition

```json
{
  "id": "e-ABELA-5-6",
  "fromNode": "n-ABELA-5",
  "fromSide": "bottom",
  "toNode": "n-ABELA-6",
  "toSide": "top",
  "toEnd": "arrow",
  "label": "Help her escape",
  "x-id": "ABELA[5]->[6]",
  "x-kind": "normal",
  "x-role": "paraphrase",
  "x-textId": "#2690",
  "x-condition": "Global(\"HelpRamazith\",\"GLOBAL\",1)\r\n!Dead(\"Ramazith\")"
}
```

### Minimal complete document

```json
{
  "x-dCanvasVersion": "3.0",
  "nodes": [
    {
      "id": "n-start", "type": "text",
      "x": 0, "y": 0, "width": 400, "height": 200,
      "text": "Greetings, traveler.",
      "x-kind": "line",
      "x-character": { "name": "Guard" }
    },
    {
      "id": "n-reply1", "type": "text",
      "x": 0, "y": 300, "width": 400, "height": 100,
      "text": "Hello there.",
      "x-kind": "reply"
    }
  ],
  "edges": [
    {
      "id": "e-start-reply1",
      "fromNode": "n-start", "fromSide": "bottom",
      "toNode": "n-reply1", "toSide": "top", "toEnd": "arrow",
      "label": "Greet back",
      "x-kind": "normal"
    }
  ]
}
```
