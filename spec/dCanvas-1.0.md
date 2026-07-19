# dCanvas Spec

<small>Version 1.0 — 2026-07-18</small>

## Overview

dCanvas 1.0 is a **dialogue** graph format. It is a **strict superset of [JSON Canvas 1.0](https://jsoncanvas.org/)**: all dialogue metadata lives in `d-` prefixed extension fields. Stripping every `d-` and `x-` prefixed field from a dCanvas 1.0 file produces a valid JSON Canvas 1.0 document.

dCanvas is **dialogue-only by design** — it describes who says what and how the conversation branches. It is not a general-purpose diagram format.

### Design principles

1. **JSON Canvas first** — the visible text always lives in the standard JSON Canvas fields: a node's `text` and an edge's `label`. No metadata is ever encoded inside `text`.
2. **Three layers of meaning** — every field belongs to exactly one layer:
   - **Layer 0 — JSON Canvas** (`id`, `type`, `x`, `y`, `width`, `height`, `color`, `text`, edge `label`, …): the universal base.
   - **Layer 1 — dialogue vocabulary** (the `d-` fields defined in this spec): the shared language of any dialogue project.
   - **Layer 2 — project extensions** (`x-` prefixed fields): specific to one game/engine. The core spec does not know them.
3. **Open and extensible** — `d-role` is a free string, `d-character` is open, and projects add their own `x-` fields freely. A generic dCanvas tool need not understand Layer 2 to handle a file correctly.
4. **Round-trip safe (mandatory)** — see [Field preservation](#field-preservation).

## Extension namespaces

Two prefixes partition the extension fields:

- **`d-`** — reserved for this spec. The `d-` fields defined below are the Layer 1 dialogue vocabulary; any other `d-` field is **reserved for future versions of the format** and MUST NOT be given a project-specific meaning.
- **`x-`** — the conventional home of Layer 2 project extensions (e.g. `x-journalText`). The spec never defines an `x-` field.

Mandatory [field preservation](#field-preservation) applies to every unknown field regardless of prefix.

## File extensions

dCanvas files use one of two equivalent extensions:

- `.dcanvas` — compact form
- `.d.canvas` — alternative form that preserves the `.canvas` suffix

Both are treated identically. Consumers should accept either. When both exist for the same dialogue, `.d.canvas` takes priority.

## Version detection

A dCanvas 1.0 file **MUST** contain the top-level property:

```json
"d-version": "1.0"
```

The value is `major.minor`. The **major** version governs compatibility:

- **Same major** → compatible. A higher minor may add fields; they survive untouched via [field preservation](#field-preservation).
- **Different major** → incompatible. A consumer **MUST NOT** silently half-interpret the file; it should reject it with an error.

If the field is absent, the file is plain JSON Canvas 1.0 and carries no dialogue semantics.

## Field preservation

> **A conforming implementation MUST preserve, on round-trip (read → write), every field it does not recognise — at any depth of nesting.**

This is what makes dCanvas extensible rather than merely having a fixed set of extensions. Concretely:

- Unknown top-level, node, and edge fields are preserved (`d-` prefixed, `x-` prefixed, and any unknown standard-looking field alike).
- Preservation is **recursive**: unknown keys inside known objects (e.g. inside `d-character`) are preserved too.
- A tool that understands only Layer 0+1 must not destroy a Layer 2 project's data when it re-saves a file.

## Literal text

A node's `text` and an edge's `label` are **literal text**: the format assigns them no markup semantics and never escapes anything. Some JSON Canvas tools render `text` as Markdown; that is a display concern of those tools, not a property of the format.

## Top-level structure

```json
{
  "d-version": "1.0",
  "nodes": [],
  "edges": []
}
```

| Property | Required | Type | Description |
|---|---|---|---|
| `d-version` | **yes** | string (const `"1.0"`) | Format version identifier |
| `nodes` | no | array of nodes | Canvas nodes in ascending z-index order |
| `edges` | no | array of edges | Connections between nodes |

By current convention **one dCanvas file contains exactly one dialogue graph.** Grouping multiple graphs via `group` nodes is allowed by JSON Canvas but is not used today (see [Node types](#node-types)).

## Nodes

### Standard JSON Canvas fields (Layer 0)

All standard JSON Canvas 1.0 node fields apply unchanged. Dialogue semantics attach only to `text` nodes (see [Node types](#node-types)).

| Field | Required | Type | Description |
|---|---|---|---|
| `id` | yes | string | Unique node identifier **within the file**. Edges reference it via `fromNode`/`toNode`. This is the *canvas* id, not the engine id — see `d-id`. |
| `type` | yes | string | Node type — `"text"` for dialogue nodes |
| `x`, `y` | yes | integer | Position in pixels |
| `width`, `height` | yes | integer | Size in pixels |
| `color` | no | `canvasColor` | Node color (hex string or preset `"1"`–`"6"`) |
| `text` | yes (for `text` nodes) | string | **The spoken content of this node** — literal text. For a `line` node, what the character says; for a `reply` node, what the player says. |

### Dialogue fields (Layer 1)

All optional and `d-` prefixed. A `text` node that carries `d-kind` is a dialogue node; a `text` node without it is a plain annotation card and is preserved as-is.

| Field | Type | Description |
|---|---|---|
| `d-id` | string | **Engine/domain** identifier of this node (e.g. `"ABELA[5]"`). Distinct from the canvas `id`. |
| `d-kind` | `"line"` \| `"reply"` | **Closed, rendering hint.** `line` = an utterance spoken *to* the player (NPC, narrator, object). `reply` = the player's own utterance. Both values are purely presentational. |
| `d-role` | string | **Open, project-defined** semantic label (e.g. `"state"`, `"transition"`, `"paraphrase"`). Generic tools need not understand it. |
| `d-textId` | string | String-table reference for `text` (e.g. `"#2687"`) |
| `d-condition` | string | Engine condition/trigger gating this node |
| `d-action` | string | Engine action executed when this node is reached |
| `d-sound` | string | Sound resource for this node's spoken line (may include path/extension) |
| `d-character` | object | Speaker information — see below |

The *value* format of `d-id` and `d-textId` (e.g. `"ABELA[5]"`, `"#2687"`) is a project convention, not part of the format: the spec only says the fields are strings and what they refer to.

#### `d-character` object

Open object — projects may add keys (preserved recursively).

| Field | Required | Type | Description |
|---|---|---|---|
| `name` | yes | string | Character display name |
| `portrait` | no | string | Portrait image filename; absent or empty means no portrait |
| `gender` | no | string | Speaker gender (project-defined vocabulary) |

### Project fields (Layer 2)

Anything `x-` prefixed. The core spec does not define them; they are valid and **preserved**. Example — the Infinity Engine port adds `x-journalText`, `x-journalTextId`, `x-journalSound` to `reply` nodes. A generic dCanvas tool carries these through without understanding them.

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
| `label` | no | string | **The player-facing choice text** — what the player sees and picks, as literal text. When the choice is a shortened paraphrase of a `reply` node, the short form lives here; when choice and reply coincide, it may duplicate the reply text or be omitted. |

### Dialogue fields (Layer 1)

| Field | Type | Description |
|---|---|---|
| `d-id` | string | Engine/domain identifier of this edge |
| `d-kind` | `"normal"` \| `"loop"` | **Closed, rendering hint.** `loop` marks a back-edge (a cycle in the dialogue); `normal` is a forward transition. May be drawn differently (e.g. `loop` dashed). |
| `d-role` | string | Open, project-defined label |
| `d-condition` | string | Engine condition expression gating this transition |
| `d-textId` | string | String-table reference for `label`, when present |

As on nodes, the value format of `d-id` and `d-textId` is a project convention.

## Text placement rule

This is the heart of JSON-Canvas compatibility — the two visible texts ride on standard fields:

- **What the character says** → the **node** `text` (a `line` node).
- **What the player says** (full reply) → the **node** `text` (a `reply` node).
- **What the player sees and picks** (the choice/paraphrase) → the **edge** `label`.

Because both are standard JSON Canvas fields, the visible text needs no dialogue-specific interpretation to display.

## Node types

JSON Canvas defines four node types: `text`, `file`, `link`, `group`.

- Only **`text`** nodes carry dialogue semantics (Layer 1 fields).
- `group`, `file`, and `link` nodes are **valid and preserved** but carry no dialogue semantics (no `d-kind`/`d-role`).
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
  "d-id": "ABELA[5]",
  "d-kind": "line",
  "d-role": "state",
  "d-character": { "name": "Abela the Nymph" },
  "d-textId": "#2687",
  "d-condition": "Global(\"RagefastDead\",\"GLOBAL\",0)\r\nDead(\"Ragefast\")"
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
  "d-id": "ABELA[8]",
  "d-kind": "reply",
  "d-role": "transition",
  "d-action": "DialogueInterrupt(FALSE)\r\nGiveItem(\"CLCK07\",LastTalkedToBy)",
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
  "d-id": "ABELA[5]->[6]",
  "d-kind": "normal",
  "d-role": "paraphrase",
  "d-textId": "#2690",
  "d-condition": "Global(\"HelpRamazith\",\"GLOBAL\",1)\r\n!Dead(\"Ramazith\")"
}
```

### Minimal complete document

```json
{
  "d-version": "1.0",
  "nodes": [
    {
      "id": "n-start", "type": "text",
      "x": 0, "y": 0, "width": 400, "height": 200,
      "text": "Greetings, traveler.",
      "d-kind": "line",
      "d-character": { "name": "Guard" }
    },
    {
      "id": "n-reply1", "type": "text",
      "x": 0, "y": 300, "width": 400, "height": 100,
      "text": "Hello there.",
      "d-kind": "reply"
    }
  ],
  "edges": [
    {
      "id": "e-start-reply1",
      "fromNode": "n-start", "fromSide": "bottom",
      "toNode": "n-reply1", "toSide": "top", "toEnd": "arrow",
      "label": "Greet back",
      "d-kind": "normal"
    }
  ]
}
```
