# dCanvas Spec

<small>Version 2.0 — 2026-02-20</small>

## Overview

dCanvas 2.0 is a dialog-tree format for game dialog data. It is a **strict superset of [JSON Canvas 1.0](https://jsoncanvas.org/)**: all game-specific metadata is stored in `x-` prefixed extension fields. Stripping every `x-` field from a dCanvas 2.0 file produces a valid JSON Canvas 1.0 document.

### Design principles

1. **Clean separation** — the `text` field contains only human-readable dialog text. No YAML front-matter, no journal separators, no encoded metadata.
2. **Explicit metadata** — every piece of game-engine data has its own named field.
3. **Round-trip safe** — any JSON Canvas editor can open and save a dCanvas file without losing extension data, provided it preserves unknown properties.

## File extensions

dCanvas files use one of two equivalent extensions:

- `.dcanvas` — compact form
- `.d.canvas` — alternative form that preserves the `.canvas` suffix

Both are treated identically. Consumers should accept either extension. When both exist for the same dialog, `.d.canvas` takes priority.

Legacy v1.0 files use the plain `.canvas` extension.

## Version detection

A dCanvas 2.0 file **MUST** contain the following top-level property:

```json
"x-dCanvasVersion": "2.0"
```

Consumers should check this field before attempting to interpret any `x-` extension fields. If the field is absent, the file should be treated as plain JSON Canvas 1.0.

## Top-level structure

```json
{
  "x-dCanvasVersion": "2.0",
  "nodes": [],
  "edges": []
}
```

| Property | Required | Type | Description |
|---|---|---|---|
| `x-dCanvasVersion` | **yes** | string (const `"2.0"`) | Format version identifier |
| `nodes` | no | array of nodes | Canvas nodes in ascending z-index order |
| `edges` | no | array of edges | Connections between nodes |

## Nodes

### Standard JSON Canvas fields

All standard JSON Canvas 1.0 node fields apply unchanged. Only `text` type nodes are used for dialog data.

| Field | Required | Type | Description |
|---|---|---|---|
| `id` | yes | string | Unique node identifier |
| `type` | yes | string | Node type — `"text"` for dialog nodes |
| `x` | yes | integer | X position in pixels |
| `y` | yes | integer | Y position in pixels |
| `width` | yes | integer | Width in pixels |
| `height` | yes | integer | Height in pixels |
| `color` | no | `canvasColor` | Node color (hex string or preset `"1"`–`"6"`) |
| `text` | yes | string | **Clean dialog text only** — no YAML, no journal separator |

### Extension fields

Extension fields carry game-specific metadata. All are prefixed with `x-` per JSON Canvas convention.

| Field | Required | Type | Description |
|---|---|---|---|
| `x-nodeRole` | **yes** | `"state"` \| `"transition"` | Role in the dialog tree. **State** nodes represent NPC lines or dialog states. **Transition** nodes represent player choices or dialog flow control. |
| `x-nodeId` | no | string | Game-engine node identifier (e.g. `"ABELA[5]"`) |
| `x-character` | no | object | Speaker information — see below |
| `x-textId` | no | string | Dialog string reference in the game's string table (e.g. `"#2687"`) |
| `x-sound` | no | string | Sound resource name associated with this dialog line (may include path prefix and extension, e.g. `"sounds/TTB011.wav"`) |
| `x-journalTextId` | no | string | Journal string reference in the game's string table |
| `x-journalText` | no | string | Journal text content, stored separately from dialog text |
| `x-journalSound` | no | string | Sound resource name associated with the journal entry (may include path prefix and extension) |
| `x-trigger` | no | string | Game-engine trigger/condition script evaluated when entering this node |
| `x-action` | no | string | Game-engine action script executed when this node is reached |

#### `x-character` object

| Field | Required | Type | Description |
|---|---|---|---|
| `name` | yes | string | Character display name (e.g. `"Abela the Nymph"`, `"End dialog"`) |
| `portrait` | no | string | Portrait image filename relative to portraits directory; empty string or `"None.png"` means no portrait |
| `gender` | no | string | Creature gender resolved from `GENDER.IDS` (e.g. `"MALE"`, `"FEMALE"`) |

## Edges

### Standard JSON Canvas fields

All standard JSON Canvas 1.0 edge fields apply unchanged.

| Field | Required | Type | Description |
|---|---|---|---|
| `id` | yes | string | Unique edge identifier |
| `fromNode` | yes | string | Source node `id` |
| `fromSide` | no | string | Side of source node (`"top"`, `"right"`, `"bottom"`, `"left"`) |
| `fromEnd` | no | string | Endpoint shape at start — `"none"` (default) or `"arrow"` |
| `toNode` | yes | string | Destination node `id` |
| `toSide` | no | string | Side of destination node (`"top"`, `"right"`, `"bottom"`, `"left"`) |
| `toEnd` | no | string | Endpoint shape at end — `"arrow"` (default) or `"none"` |
| `color` | no | `canvasColor` | Edge color |
| `label` | no | string | **Free text label for display** — no longer overloaded for conditions |

### Extension fields

| Field | Required | Type | Description |
|---|---|---|---|
| `x-condition` | no | string | Game-engine condition expression that gates this transition. In v1.0, this data was packed into the `label` field; in v2.0 it has its own dedicated field. |

## Color

Color handling is identical to JSON Canvas 1.0. The `canvasColor` type accepts either a hex string (e.g. `"#FF0000"`) or a preset number string:

- `"1"` red
- `"2"` orange
- `"3"` yellow
- `"4"` green
- `"5"` cyan
- `"6"` purple

## Examples

### State node

An NPC dialog line with a trigger condition:

```json
{
  "id": "state-ABELA[5]",
  "type": "text",
  "x": 0,
  "y": 0,
  "width": 400,
  "height": 300,
  "color": "3",
  "text": "My poor Ragefast. Like many humans he could not understand the feelings my kind elicits...",
  "x-nodeRole": "state",
  "x-nodeId": "ABELA[5]",
  "x-character": {
    "name": "Abela the Nymph",
    "portrait": "None.png"
  },
  "x-textId": "#2687",
  "x-trigger": "Global(\"RagefastDead\",\"GLOBAL\",0)\r\nDead(\"Ragefast\")"
}
```

### Transition node with journal text

A flow-control node that ends dialog, gives an item, and writes a journal entry:

```json
{
  "id": "transition-ABELA[8]",
  "type": "text",
  "x": 920,
  "y": 2000,
  "width": 400,
  "height": 300,
  "color": "1",
  "text": "",
  "x-nodeRole": "transition",
  "x-nodeId": "ABELA[8]",
  "x-character": {
    "name": "End dialog",
    "portrait": ""
  },
  "x-journalTextId": "#26807",
  "x-journalText": "The Captive Nymph\nAbela is free from her confinement at last.",
  "x-action": "DialogueInterrupt(FALSE)\r\nGiveItem(\"CLCK07\",LastTalkedToBy)\r\n..."
}
```

### Conditional edge

An edge with a game-engine condition that must be satisfied for the transition to be available:

```json
{
  "id": "state-ABELA[5]-transition-ABELA[6]",
  "fromNode": "state-ABELA[5]",
  "fromSide": "bottom",
  "toNode": "transition-ABELA[6]",
  "toSide": "top",
  "toEnd": "arrow",
  "x-condition": "Global(\"HelpRamazith\",\"GLOBAL\",1)\r\n!Dead(\"Ramazith\")"
}
```

### Minimal complete document

```json
{
  "x-dCanvasVersion": "2.0",
  "nodes": [
    {
      "id": "state-START",
      "type": "text",
      "x": 0,
      "y": 0,
      "width": 400,
      "height": 200,
      "text": "Greetings, traveler.",
      "x-nodeRole": "state",
      "x-character": { "name": "Guard" }
    },
    {
      "id": "transition-REPLY1",
      "type": "text",
      "x": 0,
      "y": 300,
      "width": 400,
      "height": 100,
      "text": "Hello there.",
      "x-nodeRole": "transition"
    }
  ],
  "edges": [
    {
      "id": "state-START-transition-REPLY1",
      "fromNode": "state-START",
      "fromSide": "bottom",
      "toNode": "transition-REPLY1",
      "toSide": "top",
      "toEnd": "arrow"
    }
  ]
}
```

## Migration from v1.0

The v1.0 format (used by this project's `.canvas` files) stores game metadata inside the `text` field as YAML front-matter and uses the edge `label` field for condition expressions. To migrate:

### Node migration

1. **Parse YAML front-matter** from the `text` field (content between `---` delimiters).
2. **Extract** each front-matter field into its corresponding `x-` extension field:
   - `nodeId` becomes `x-nodeId`
   - `character` becomes `x-character`
   - `textId` becomes `x-textId`
   - `journalTextId` becomes `x-journalTextId`
   - `trigger` becomes `x-trigger`
   - `action` becomes `x-action`
3. **Split journal text**: if the text body contains the `>---- JOURNAL ----<` separator, move everything after it into `x-journalText`.
4. **Clean `text`**: set `text` to only the dialog text (after front-matter, before journal separator). Trim whitespace.
5. **Set `x-nodeRole`**: determine from the node's role in the dialog tree — typically derived from the node's color or connectivity pattern. State nodes (NPC lines) and transition nodes (player choices, flow control) should be distinguished.

### Edge migration

1. If `label` contains a game-engine condition expression, move it to `x-condition`.
2. Clear `label` or repurpose it for display text.

### Version stamp

Add `"x-dCanvasVersion": "2.0"` at the top level of the document.
