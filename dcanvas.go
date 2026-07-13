// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: BlueOak-1.0.0

// Package dcanvas provides types and IO for the dCanvas 3.0 format, a strict
// superset of JSON Canvas 1.0 for dialogue graphs.
//
// Every field belongs to exactly one of three layers:
//
//   - Layer 0 — the JSON Canvas core (id, type, x, y, width, height, color,
//     text, edge label, …);
//   - Layer 1 — the dialogue vocabulary, modelled below as typed x- fields;
//   - Layer 2 — project-specific x- fields the spec does not know about.
//
// The package understands Layers 0 and 1 as typed members. Any other field —
// at the top level, on a node, or on an edge — is unknown to the package and
// is preserved verbatim on a read → write round-trip (see ADR-0003 and
// ADR-0005). The package depends on nothing outside the standard library; the
// dependency direction is one-way (a project depends on dcanvas, never the
// reverse).
package dcanvas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

// Version is the format version this package reads and writes. Encode always
// stamps it; Decode rejects any document whose major version differs.
const Version = "3.0"

// Node x-kind values (closed set).
const (
	KindLine  = "line"  // an utterance spoken to the player (NPC, narrator, object)
	KindReply = "reply" // the player's own utterance
)

// Edge x-kind values (closed set).
const (
	KindNormal = "normal" // a forward transition
	KindLoop   = "loop"   // a back-edge (a cycle in the dialogue), a rendering hint only
)

// Canvas is a dCanvas 3.0 document: a single dialogue graph.
type Canvas struct {
	// Version is the document's x-dCanvasVersion. Encode always writes "3.0"
	// regardless of this value; it is populated on Decode.
	Version string
	Nodes   []*Node
	Edges   []*Edge

	// extra holds unrecognised top-level fields, preserved verbatim.
	extra map[string]json.RawMessage
}

// Node is a canvas node. Dialogue semantics (Layer 1) attach to text nodes
// that carry x-kind; other text nodes are plain annotation cards, and file /
// link / group nodes are opaque to dialogue logic. All are preserved.
type Node struct {
	// Layer 0 — JSON Canvas core.
	ID     string // canvas-local id; edges reference it via fromNode/toNode
	Type   string // "text" for dialogue nodes
	X      int
	Y      int
	Width  int
	Height int
	Color  string // optional
	Text   string // spoken content (clean text only)

	// Layer 1 — dialogue vocabulary. All optional.
	XID       string     // engine/domain id (distinct from the canvas ID)
	Kind      string     // x-kind: KindLine | KindReply
	Role      string     // x-role: open, project-defined label
	TextID    string     // x-textId: string-table reference for Text
	Condition string     // x-condition: engine condition gating this node
	Action    string     // x-action: engine action executed at this node
	Sound     string     // x-sound: sound resource for this node's line
	Character *Character // x-character: speaker information

	// extra holds unrecognised node fields (including Layer 2 x- fields),
	// preserved verbatim.
	extra map[string]json.RawMessage
}

// Edge is a connection between two nodes.
type Edge struct {
	// Layer 0 — JSON Canvas core.
	ID       string
	FromNode string
	FromSide string // optional: "top" | "right" | "bottom" | "left"
	FromEnd  string // optional: "none" (default) | "arrow"
	ToNode   string
	ToSide   string // optional
	ToEnd    string // optional: "arrow" (default) | "none"
	Color    string // optional
	Label    string // optional: player-facing choice text

	// Layer 1 — dialogue vocabulary. All optional.
	XID       string // engine/domain id
	Kind      string // x-kind: KindNormal | KindLoop
	Role      string // x-role: open, project-defined label
	Condition string // x-condition: engine condition gating this transition
	TextID    string // x-textId: string-table reference for Label

	// extra holds unrecognised edge fields, preserved verbatim.
	extra map[string]json.RawMessage
}

// Character holds speaker information for a dialogue node. It is an open object
// in the spec: known fields are modelled as typed members, and any other key
// is preserved verbatim on a round-trip via the same catch-all pattern used by
// Node and Edge. This makes recursive preservation reusable for any future
// known nested object, not specific to x-character.
type Character struct {
	Name     string // required
	Portrait string // optional
	Gender   string // optional

	// extra holds unrecognised character fields, preserved verbatim.
	extra map[string]json.RawMessage
}

// --- preservation helpers ---------------------------------------------------

// objectWriter builds a JSON object that emits known fields in a fixed,
// human-friendly order (rather than the alphabetical order json.Marshal gives a
// map), then appends any preserved unknown fields. This keeps round-trip
// preservation while producing readable, JSON-Canvas-conventional output.
type objectWriter struct {
	fields []field
}

type field struct {
	key string
	val json.RawMessage
}

// always marshals a value and always emits it (used for required fields).
func (w *objectWriter) always(key string, val any) error {
	b, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("dcanvas: field %q: %w", key, err)
	}
	w.fields = append(w.fields, field{key, b})
	return nil
}

// str emits a string field only when it is non-empty (omitempty).
func (w *objectWriter) str(key, val string) {
	if val == "" {
		return
	}
	b, _ := json.Marshal(val) // marshalling a string never fails
	w.fields = append(w.fields, field{key, b})
}

// kind validates a closed x-kind value and emits it when non-empty.
func (w *objectWriter) kind(val, a, b, owner string) error {
	if val == "" {
		return nil
	}
	if val != a && val != b {
		return fmt.Errorf("dcanvas: invalid %s x-kind %q (want %q or %q)", owner, val, a, b)
	}
	w.str("x-kind", val)
	return nil
}

// bytes serialises the ordered known fields followed by the preserved unknown
// fields (sorted for determinism, since their original order is not retained).
func (w *objectWriter) bytes(extra map[string]json.RawMessage) []byte {
	var buf bytes.Buffer
	buf.WriteByte('{')
	first := true
	emit := func(k string, v json.RawMessage) {
		if !first {
			buf.WriteByte(',')
		}
		first = false
		kb, _ := json.Marshal(k)
		buf.Write(kb)
		buf.WriteByte(':')
		buf.Write(v)
	}
	for _, f := range w.fields {
		emit(f.key, f.val)
	}
	if len(extra) > 0 {
		keys := make([]string, 0, len(extra))
		for k := range extra {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			emit(k, extra[k])
		}
	}
	buf.WriteByte('}')
	return buf.Bytes()
}

// popString unmarshals a known string field out of raw and removes its key, so
// that whatever remains in raw is genuinely unknown.
func popString(raw map[string]json.RawMessage, key string, dst *string) error {
	if v, ok := raw[key]; ok {
		if err := json.Unmarshal(v, dst); err != nil {
			return fmt.Errorf("dcanvas: field %q: %w", key, err)
		}
		delete(raw, key)
	}
	return nil
}

// popInt is popString's integer counterpart.
func popInt(raw map[string]json.RawMessage, key string, dst *int) error {
	if v, ok := raw[key]; ok {
		if err := json.Unmarshal(v, dst); err != nil {
			return fmt.Errorf("dcanvas: field %q: %w", key, err)
		}
		delete(raw, key)
	}
	return nil
}

// --- Character IO -----------------------------------------------------------

// UnmarshalJSON decodes a character, keeping any unrecognised field in a
// catch-all so nested unknown keys survive a round-trip.
func (ch *Character) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for key, dst := range map[string]*string{
		"name":     &ch.Name,
		"portrait": &ch.Portrait,
		"gender":   &ch.Gender,
	} {
		if err := popString(raw, key, dst); err != nil {
			return err
		}
	}
	ch.extra = raw
	return nil
}

// MarshalJSON encodes a character, merging back any preserved unknown fields.
func (ch *Character) MarshalJSON() ([]byte, error) {
	var w objectWriter
	if err := w.always("name", ch.Name); err != nil {
		return nil, err
	}
	w.str("portrait", ch.Portrait)
	w.str("gender", ch.Gender)
	return w.bytes(ch.extra), nil
}

// --- Canvas IO --------------------------------------------------------------

// UnmarshalJSON decodes a canvas, keeping any unrecognised top-level field.
func (c *Canvas) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if err := popString(raw, "x-dCanvasVersion", &c.Version); err != nil {
		return err
	}
	if v, ok := raw["nodes"]; ok {
		if err := json.Unmarshal(v, &c.Nodes); err != nil {
			return fmt.Errorf("dcanvas: field %q: %w", "nodes", err)
		}
		delete(raw, "nodes")
	}
	if v, ok := raw["edges"]; ok {
		if err := json.Unmarshal(v, &c.Edges); err != nil {
			return fmt.Errorf("dcanvas: field %q: %w", "edges", err)
		}
		delete(raw, "edges")
	}
	c.extra = raw
	return nil
}

// MarshalJSON encodes a canvas, always stamping the current format version and
// merging back any preserved unknown top-level fields. nodes and edges are
// always emitted as arrays so the result is a clean JSON Canvas document.
func (c *Canvas) MarshalJSON() ([]byte, error) {
	var w objectWriter
	w.str("x-dCanvasVersion", Version)

	nodes := c.Nodes
	if nodes == nil {
		nodes = []*Node{}
	}
	if err := w.always("nodes", nodes); err != nil {
		return nil, err
	}
	edges := c.Edges
	if edges == nil {
		edges = []*Edge{}
	}
	if err := w.always("edges", edges); err != nil {
		return nil, err
	}
	return w.bytes(c.extra), nil
}

// SetExtra attaches a Layer 2 (project-specific) field to the node by
// JSON-marshaling val. It is preserved verbatim on encode and survives a
// round-trip through Decode, the same as any other unrecognised field; a
// typed field with the same key always wins.
func (n *Node) SetExtra(key string, val any) error {
	b, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("dcanvas: field %q: %w", key, err)
	}
	if n.extra == nil {
		n.extra = make(map[string]json.RawMessage)
	}
	n.extra[key] = b
	return nil
}

// --- Node IO ----------------------------------------------------------------

// UnmarshalJSON decodes a node, keeping any unrecognised field (including
// Layer 2 x- fields) in a catch-all.
func (n *Node) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for key, dst := range map[string]*string{
		"id":          &n.ID,
		"type":        &n.Type,
		"color":       &n.Color,
		"text":        &n.Text,
		"x-id":        &n.XID,
		"x-kind":      &n.Kind,
		"x-role":      &n.Role,
		"x-textId":    &n.TextID,
		"x-condition": &n.Condition,
		"x-action":    &n.Action,
		"x-sound":     &n.Sound,
	} {
		if err := popString(raw, key, dst); err != nil {
			return err
		}
	}
	for key, dst := range map[string]*int{
		"x": &n.X, "y": &n.Y, "width": &n.Width, "height": &n.Height,
	} {
		if err := popInt(raw, key, dst); err != nil {
			return err
		}
	}
	if v, ok := raw["x-character"]; ok {
		var ch Character
		if err := json.Unmarshal(v, &ch); err != nil {
			return fmt.Errorf("dcanvas: field %q: %w", "x-character", err)
		}
		n.Character = &ch
		delete(raw, "x-character")
	}
	n.extra = raw
	return nil
}

// MarshalJSON encodes a node, validating x-kind and merging back any preserved
// unknown fields.
func (n *Node) MarshalJSON() ([]byte, error) {
	var w objectWriter

	// Layer 0 — required geometry/identity always emitted, in canvas order.
	if err := w.always("id", n.ID); err != nil {
		return nil, err
	}
	if err := w.always("type", n.Type); err != nil {
		return nil, err
	}
	_ = w.always("x", n.X)
	_ = w.always("y", n.Y)
	_ = w.always("width", n.Width)
	_ = w.always("height", n.Height)
	w.str("color", n.Color)
	w.str("text", n.Text)

	// Layer 1 — optional dialogue vocabulary.
	w.str("x-id", n.XID)
	if err := w.kind(n.Kind, KindLine, KindReply, "node"); err != nil {
		return nil, err
	}
	w.str("x-role", n.Role)
	w.str("x-textId", n.TextID)
	w.str("x-condition", n.Condition)
	w.str("x-action", n.Action)
	w.str("x-sound", n.Sound)
	if n.Character != nil {
		if err := w.always("x-character", n.Character); err != nil {
			return nil, err
		}
	}
	return w.bytes(n.extra), nil
}

// --- Edge IO ----------------------------------------------------------------

// UnmarshalJSON decodes an edge, keeping any unrecognised field in a catch-all.
func (e *Edge) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for key, dst := range map[string]*string{
		"id":          &e.ID,
		"fromNode":    &e.FromNode,
		"fromSide":    &e.FromSide,
		"fromEnd":     &e.FromEnd,
		"toNode":      &e.ToNode,
		"toSide":      &e.ToSide,
		"toEnd":       &e.ToEnd,
		"color":       &e.Color,
		"label":       &e.Label,
		"x-id":        &e.XID,
		"x-kind":      &e.Kind,
		"x-role":      &e.Role,
		"x-condition": &e.Condition,
		"x-textId":    &e.TextID,
	} {
		if err := popString(raw, key, dst); err != nil {
			return err
		}
	}
	e.extra = raw
	return nil
}

// MarshalJSON encodes an edge, validating x-kind and merging back any preserved
// unknown fields.
func (e *Edge) MarshalJSON() ([]byte, error) {
	var w objectWriter

	// Layer 0 — required identity/endpoints first, then optional geometry.
	if err := w.always("id", e.ID); err != nil {
		return nil, err
	}
	if err := w.always("fromNode", e.FromNode); err != nil {
		return nil, err
	}
	w.str("fromSide", e.FromSide)
	w.str("fromEnd", e.FromEnd)
	if err := w.always("toNode", e.ToNode); err != nil {
		return nil, err
	}
	w.str("toSide", e.ToSide)
	w.str("toEnd", e.ToEnd)
	w.str("color", e.Color)
	w.str("label", e.Label)

	// Layer 1 — optional dialogue vocabulary.
	w.str("x-id", e.XID)
	if err := w.kind(e.Kind, KindNormal, KindLoop, "edge"); err != nil {
		return nil, err
	}
	w.str("x-role", e.Role)
	w.str("x-condition", e.Condition)
	w.str("x-textId", e.TextID)
	return w.bytes(e.extra), nil
}

// HasOverlappingNodes reports whether any two nodes in the canvas have
// overlapping bounding boxes. Nodes that merely touch at an edge are
// not considered overlapping (strict less-than comparison).
func (c *Canvas) HasOverlappingNodes() bool {
	nodes := c.Nodes
	for i := 0; i < len(nodes); i++ {
		a := nodes[i]
		for j := i + 1; j < len(nodes); j++ {
			b := nodes[j]
			if a.X < b.X+b.Width &&
				b.X < a.X+a.Width &&
				a.Y < b.Y+b.Height &&
				b.Y < a.Y+a.Height {
				return true
			}
		}
	}
	return false
}
