# dCanvas Spec

<small>Version 1.0 — 2025-09-27</small>

## Overview

dCanvas 1.0 is the original dialog-tree format used by this project. It is a **de facto extension of [JSON Canvas 1.0](https://jsoncanvas.org/)** where game-specific metadata is encoded inside standard JSON Canvas fields rather than in dedicated extension properties.

This document retroactively formalises the conventions that were already in use, for the purpose of backward compatibility and as a reference for the v1.0 → v2.0 migration.

### Key characteristics

1. **YAML front-matter in `text`** — the node `text` field contains an optional YAML front-matter block (delimited by `---`), followed by dialog text, and optionally journal text after a separator.
2. **Conditions in `label`** — edge `label` is repurposed for game-engine condition expressions, not display text.
3. **Role inferred from `id` prefix** — node IDs prefixed with `state-` or `transition-` indicate the node's role in the dialog tree.

## File extension

dCanvas v1.0 files use the standard `.canvas` extension (plain JSON Canvas files with conventions applied).

## Version detection

A dCanvas 1.0 file **MAY** contain the top-level property:

```json
"x-dCanvasVersion": "1.0"
```

However, most v1.0 files do not have this field. Consumers should treat any file **without** `"x-dCanvasVersion": "2.0"` as v1.0.

## Top-level structure

```json
{
  "nodes": [],
  "edges": []
}
```

| Property           | Required | Type           | Description                             |
| ------------------ | -------- | -------------- | --------------------------------------- |
| `x-dCanvasVersion` | no       | string         | If present, `"1.0"`. Usually absent.    |
| `nodes`            | no       | array of nodes | Canvas nodes in ascending z-index order |
| `edges`            | no       | array of edges | Connections between nodes               |

## Nodes

### Standard JSON Canvas fields

| Field    | Required | Type          | Description                                                      |
| -------- | -------- | ------------- | ---------------------------------------------------------------- |
| `id`     | yes      | string        | Unique node identifier — encodes role via prefix (see below)     |
| `type`   | yes      | string        | Node type — `"text"` for dialog nodes                            |
| `x`      | yes      | integer       | X position in pixels                                             |
| `y`      | yes      | integer       | Y position in pixels                                             |
| `width`  | yes      | integer       | Width in pixels                                                  |
| `height` | yes      | integer       | Height in pixels                                                 |
| `color`  | no       | `canvasColor` | Node color (hex string or preset `"1"`–`"6"`)                    |
| `text`   | yes      | string        | Contains YAML front-matter + dialog text + optional journal text |

### Node ID conventions

The `id` field encodes the node's role via its prefix:

| Pattern                   | Role       | Description                                     |
| ------------------------- | ---------- | ----------------------------------------------- |
| `state-CHARACTER[N]`      | state      | NPC dialog line or dialog state                 |
| `transition-CHARACTER[N]` | transition | Player choice, flow control, or end-dialog node |

Where `CHARACTER` is an uppercase character identifier and `N` is a numeric index (e.g. `state-ABELA[5]`, `transition-ABELA[8]`).

### Text field format

The `text` field has a composite structure:

```
---
nodeId: ABELA[5]
character:
    name: Abela the Nymph
    portrait: None.png
textId: '#2687'
journalTextId: '#26807'
trigger: "Global(\"RagefastDead\",\"GLOBAL\",0)"
action: "GiveItem(\"CLCK07\",LastTalkedToBy)"
---
Dialog text goes here.

>---- JOURNAL ----<

Journal-only text goes here.
```

#### YAML front-matter

The front-matter block is delimited by `---` on its own line. All fields are optional:

| Field           | Type   | Description                                                        |
| --------------- | ------ | ------------------------------------------------------------------ |
| `nodeId`        | string | Game-engine node identifier (e.g. `"ABELA[5]"`)                    |
| `character`     | object | Speaker info with `name` (string) and optional `portrait` (string) |
| `textId`        | string | Dialog string reference (e.g. `"#2687"`)                           |
| `journalTextId` | string | Journal string reference                                           |
| `trigger`       | string | Game-engine trigger/condition script                               |
| `action`        | string | Game-engine action script                                          |

`character.portrait` is a filename relative to the portraits directory (e.g. `None.png`). Empty string or `"None.png"` means no portrait.

#### Journal text separator

For transition nodes, dialog text and journal text may be separated by:

```
\n\n>---- JOURNAL ----<\n\n
```

Everything before the separator is dialog text; everything after is journal text.

#### Nodes without front-matter

If the `text` field does not start with `---\n`, its entire content is treated as plain dialog text with no metadata.

## Edges

| Field      | Required | Type          | Description                                                                      |
| ---------- | -------- | ------------- | -------------------------------------------------------------------------------- |
| `id`       | yes      | string        | Unique edge identifier                                                           |
| `fromNode` | yes      | string        | Source node `id`                                                                 |
| `fromSide` | no       | string        | Side of source node (`"top"`, `"right"`, `"bottom"`, `"left"`)                   |
| `fromEnd`  | no       | string        | Endpoint shape at start — `"none"` (default) or `"arrow"`                        |
| `toNode`   | yes      | string        | Destination node `id`                                                            |
| `toSide`   | no       | string        | Side of destination node (`"top"`, `"right"`, `"bottom"`, `"left"`)              |
| `toEnd`    | no       | string        | Endpoint shape at end — `"arrow"` (default) or `"none"`                          |
| `color`    | no       | `canvasColor` | Edge color                                                                       |
| `label`    | no       | string        | **Game-engine condition expression** — repurposed from JSON Canvas display label |

> **Note**: In v1.0, the `label` field does not contain display text. It stores game-engine condition expressions that gate the transition. In v2.0, this data moves to the dedicated `x-condition` field.

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
  "text": "---\nnodeId: ABELA[5]\ncharacter:\n    name: Abela the Nymph\n    portrait: None.png\ntextId: '#2687'\ntrigger: \"Global(\\\"RagefastDead\\\",\\\"GLOBAL\\\",0)\\r\\nDead(\\\"Ragefast\\\")\"\n---\nMy poor Ragefast. Like many humans he could not understand the feelings my kind elicits..."
}
```

### Transition node with journal text

A flow-control node with an action script and journal entry:

```json
{
  "id": "transition-ABELA[8]",
  "type": "text",
  "x": 920,
  "y": 2000,
  "width": 400,
  "height": 300,
  "color": "1",
  "text": "---\nnodeId: ABELA[8]\ncharacter:\n    name: End dialog\n    portrait: \"\"\njournalTextId: '#26807'\naction: \"DialogueInterrupt(FALSE)\\r\\nGiveItem(\\\"CLCK07\\\",LastTalkedToBy)\\r\\n...\"\n---\n\n\n>---- JOURNAL ----<\n\nThe Captive Nymph\nAbela is free from her confinement at last."
}
```

### Edge with condition

An edge whose `label` contains a game-engine condition:

```json
{
  "id": "state-ABELA[5]-transition-ABELA[6]",
  "fromNode": "state-ABELA[5]",
  "fromSide": "bottom",
  "toNode": "transition-ABELA[6]",
  "toSide": "top",
  "toEnd": "arrow",
  "label": "Global(\"HelpRamazith\",\"GLOBAL\",1)\r\n!Dead(\"Ramazith\")"
}
```
