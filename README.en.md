<!-- 
SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
SPDX-License-Identifier: BlueOak-1.0.0
SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com> 
-->

# dCanvas

[☑️ Українська](README.md) | **✅ English**

A file format for dialogue graphs from games.

## What it is

dCanvas (**d**ialogue **canvas**) is a superset of [JSON Canvas](https://jsoncanvas.org/) that carries the extra information dialogues in games need: who is speaking, what the player sees, and under which conditions a reply becomes available. Every dCanvas file is also a valid JSON Canvas file, so any JSON Canvas editor (notably Obsidian) can open, display, and re-save it without data loss.

The format is built out of three layers:

- **Layer 0 — JSON Canvas.** The standard fields (`id`, `x`, `y`, `text`, …): geometry and all visible text.
- **Layer 1 — the dialogue vocabulary.** Fields with the `d-` prefix (`d-kind`, `d-character`, `d-condition`, …), defined by the dCanvas specification.
- **Layer 2 — project extensions.** Fields with the `x-` prefix, owned by a particular game or engine. The format does not interpret them, but every tool **must preserve** all unknown fields on re-save.

## Example

A guard won't let the player into the city. The player has four replies: three lead onward (each gated by its own `d-condition`), and the fourth loops the conversation back to the start (`d-kind: "loop"` on the edge).

```json
{
  "d-version": "1.0",
  "nodes": [
    {
      "id": "gate", "type": "text", "d-kind": "line",
      "x": 0, "y": 0, "width": 280, "height": 120, "color": "3",
      "d-character": { "name": "Guard", "portrait": "guard.png" },
      "text": "Halt! The city gates are closed till dawn. No one gets through."
    },
    {
      "id": "letter", "type": "text", "d-kind": "reply",
      "x": 0, "y": 320, "width": 280, "height": 120, "color": "4",
      "d-condition": "has_item('letter')",
      "text": "I'm a courier. I carry a letter for the burgomaster."
    },
    {
      "id": "bribe", "type": "text", "d-kind": "reply",
      "x": 340, "y": 320, "width": 280, "height": 120, "color": "4",
      "d-condition": "gold >= 50",
      "text": "And what if this purse were to end up in your pocket?"
    },
    {
      "id": "threat", "type": "text", "d-kind": "reply",
      "x": 680, "y": 320, "width": 280, "height": 120, "color": "4",
      "d-condition": "rank('royal_guard')",
      "text": "Open up, in the name of the king!"
    },
    {
      "id": "wait", "type": "text", "d-kind": "reply",
      "x": 1020, "y": 320, "width": 280, "height": 120, "color": "4",
      "text": "Fine, I'll wait until morning."
    },
    {
      "id": "open", "type": "text", "d-kind": "line",
      "x": 0, "y": 640, "width": 280, "height": 120, "color": "1",
      "d-character": { "name": "Guard", "portrait": "guard.png" },
      "text": "Hm, the seal is genuine… Go on through, and stay out of trouble."
    },
    {
      "id": "angry", "type": "text", "d-kind": "line",
      "x": 680, "y": 640, "width": 280, "height": 120, "color": "1",
      "d-character": { "name": "Guard", "portrait": "guard.png" },
      "text": "Look how fearsome! Off with you, before I call the captain."
    }
  ],
  "edges": [
    { "id": "e1", "fromNode": "gate", "fromSide": "bottom", "toNode": "letter", "toSide": "top", "color": "#000000" },
    { "id": "e2", "fromNode": "gate", "fromSide": "bottom", "toNode": "bribe", "toSide": "top", "color": "#000000" },
    { "id": "e3", "fromNode": "gate", "fromSide": "bottom", "toNode": "threat", "toSide": "top", "color": "#000000" },
    { "id": "e4", "fromNode": "gate", "fromSide": "bottom", "toNode": "wait", "toSide": "top", "color": "#000000" },
    { "id": "e5", "fromNode": "letter", "fromSide": "bottom", "toNode": "open", "toSide": "top", "color": "#000000" },
    { "id": "e6", "fromNode": "bribe", "fromSide": "bottom", "toNode": "open", "toSide": "top", "color": "#000000" },
    { "id": "e7", "fromNode": "threat", "fromSide": "bottom", "toNode": "angry", "toSide": "top", "color": "#000000" },
    { "id": "e8", "fromNode": "wait", "fromSide": "top", "toNode": "gate", "toSide": "top", "d-kind": "loop", "color": "1" }
  ]
}
```

This is how this dialog could look like in a viewer application:
![[readme-example.png]]
## Specification

- [dCanvas 1.0 specification](spec/dCanvas-1.0.md)
- [JSON Schema](spec/dCanvas-1.0.schema.json) — the cross-language contract of the format

The default file extension is `.d.canvas`; the compact alternative is `.dcanvas`.

This repository also hosts the reference Go library `github.com/sbtlocalization/dcanvas` (read/write with unknown-field preservation, plus automatic graph layout) and the `dcanvas` CLI tool in `cmd/dcanvas`.

## License

[BlueOak-1.0.0](LICENSES/BlueOak-1.0.0.txt)
