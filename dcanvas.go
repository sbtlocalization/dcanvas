// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: BlueOak-1.0.0

// Package dcanvas provides types and IO for the dCanvas 1.1 format, a strict
// superset of JSON Canvas 1.0 for dialogue graphs.
//
// Every field belongs to exactly one of three layers:
//
//   - Layer 0 — the JSON Canvas core (id, type, x, y, width, height, color,
//     text, edge label, …);
//   - Layer 1 — the dialogue vocabulary, modelled below as typed d- fields;
//   - Layer 2 — project-specific x- fields the spec does not know about.
//
// The package understands Layers 0 and 1 as typed members. Any other field —
// at the top level, on a node, or on an edge — is unknown to the package and
// is preserved on a read → write round-trip (see ADR-0003 and ADR-0018), in
// the order of the file. The package depends on nothing outside the standard
// library; the dependency direction is one-way (a project depends on dcanvas,
// never the reverse).
package dcanvas

import (
	"bytes"
	"encoding/json/jsontext"
	json "encoding/json/v2"
	"fmt"
)

// Version is the format version this package reads and writes. Encode always
// stamps it; Decode rejects any document whose major version differs.
const Version = "1.1"

// NodeKind is a node's d-kind; its values form a closed set.
type NodeKind string

const (
	KindLine  NodeKind = "line"  // an utterance spoken to the player (NPC, narrator, object)
	KindReply NodeKind = "reply" // the player's own utterance
)

// EdgeKind is an edge's d-kind; its values form a closed set.
type EdgeKind string

const (
	KindNormal EdgeKind = "normal" // a forward transition
	KindLoop   EdgeKind = "loop"   // a back-edge (a cycle in the dialogue), a rendering hint only
)

// Canvas is a dCanvas 1.x document: a single dialogue graph.
type Canvas struct {
	// Version is the document's d-version, populated on Decode. Encode keeps a
	// same-major value and otherwise stamps the library Version.
	Version string  `json:"d-version"`
	Nodes   []*Node `json:"nodes"`
	Edges   []*Edge `json:"edges"`

	// extra holds unrecognised top-level fields, preserved verbatim as a JSON
	// object in the order of the file.
	extra jsontext.Value
}

// Node is a canvas node. Dialogue semantics (Layer 1) attach to text nodes
// that carry d-kind; other text nodes are plain annotation cards, and file /
// link / group nodes are opaque to dialogue logic. All are preserved.
type Node struct {
	// Layer 0 — JSON Canvas core.
	ID     string `json:"id"`              // canvas-local id; edges reference it via fromNode/toNode
	Type   string `json:"type"`            // "text" for dialogue nodes
	X      int    `json:"x"`               //
	Y      int    `json:"y"`               //
	Width  int    `json:"width"`           //
	Height int    `json:"height"`          //
	Color  string `json:"color,omitempty"` // optional
	Text   string `json:"text,omitempty"`  // spoken content (literal text); always written on a text node

	// Layer 1 — dialogue vocabulary. All optional.
	DomainID              string       `json:"d-id,omitempty"`                    // engine/domain id (distinct from the canvas ID)
	Kind                  NodeKind     `json:"d-kind,omitempty"`                  // d-kind: KindLine | KindReply
	Role                  string       `json:"d-role,omitempty"`                  // d-role: open, project-defined label
	TextID                string       `json:"d-textId,omitempty"`                // d-textId: string-table reference for Text
	Condition             string       `json:"d-condition,omitempty"`             // d-condition: engine condition gating this node
	Action                string       `json:"d-action,omitempty"`                // d-action: engine action executed at this node
	Sound                 string       `json:"d-sound,omitempty"`                 // d-sound: sound resource for this node's line
	Character             *Character   `json:"d-character,omitzero"`              // d-character: speaker information
	AlternativeCharacters []*Character `json:"d-alternativeCharacters,omitempty"` // d-alternativeCharacters: other possible speakers; requires Character

	// extra holds unrecognised node fields (including Layer 2 x- fields),
	// preserved verbatim as a JSON object in the order of the file.
	extra jsontext.Value
}

// Edge is a connection between two nodes.
type Edge struct {
	// Layer 0 — JSON Canvas core.
	ID       string `json:"id"`                 //
	FromNode string `json:"fromNode"`           //
	FromSide string `json:"fromSide,omitempty"` // optional: "top" | "right" | "bottom" | "left"
	FromEnd  string `json:"fromEnd,omitempty"`  // optional: "none" (default) | "arrow"
	ToNode   string `json:"toNode"`             //
	ToSide   string `json:"toSide,omitempty"`   // optional
	ToEnd    string `json:"toEnd,omitempty"`    // optional: "arrow" (default) | "none"
	Color    string `json:"color,omitempty"`    // optional
	Label    string `json:"label,omitempty"`    // optional: player-facing choice text

	// Layer 1 — dialogue vocabulary. All optional.
	DomainID  string   `json:"d-id,omitempty"`        // engine/domain id
	Kind      EdgeKind `json:"d-kind,omitempty"`      // d-kind: KindNormal | KindLoop
	Role      string   `json:"d-role,omitempty"`      // d-role: open, project-defined label
	Condition string   `json:"d-condition,omitempty"` // d-condition: engine condition gating this transition
	TextID    string   `json:"d-textId,omitempty"`    // d-textId: string-table reference for Label

	// extra holds unrecognised edge fields, preserved verbatim as a JSON object
	// in the order of the file.
	extra jsontext.Value
}

// Character holds speaker information for a dialogue node. It is an open object
// in the spec: known fields are modelled as typed members, and any other key
// is preserved verbatim on a round-trip, the same way as on Node and Edge. This
// makes recursive preservation reusable for any future known nested object,
// not specific to d-character.
type Character struct {
	Name     string `json:"name"`               // required
	TextID   string `json:"textId,omitempty"`   // optional: string-table reference for Name
	Portrait string `json:"portrait,omitempty"` // optional
	Gender   string `json:"gender,omitempty"`   // optional

	// extra holds unrecognised character fields, preserved verbatim as a JSON
	// object in the order of the file.
	extra jsontext.Value
}

// validKind reports whether a d-kind value is empty (the field is optional) or
// a member of its closed two-value set. It is the single source of truth for
// closed-set membership, used by Validate; Encode does not enforce the set (it
// preserves whatever a decoded file contained), so the check lives only there.
func validKind[K ~string](val, a, b K) bool {
	return val == "" || val == a || val == b
}

// --- preservation -----------------------------------------------------------

// The typed fields of each type are read and written through a copy of the
// type without its JSON methods (canvasFields, nodeFields, …), so that json v2
// handles them by their tags and does not call the methods again.
type (
	canvasFields    Canvas
	nodeFields      Node
	edgeFields      Edge
	characterFields Character
)

// textNodeFields is nodeFields for a text node: the schema requires text on a
// text node, so it is written even when empty. Its fields must match Node's,
// tags aside; the conversion from Node stops compiling if they drift apart.
type textNodeFields struct {
	ID                    string       `json:"id"`
	Type                  string       `json:"type"`
	X                     int          `json:"x"`
	Y                     int          `json:"y"`
	Width                 int          `json:"width"`
	Height                int          `json:"height"`
	Color                 string       `json:"color,omitempty"`
	Text                  string       `json:"text"`
	DomainID              string       `json:"d-id,omitempty"`
	Kind                  NodeKind     `json:"d-kind,omitempty"`
	Role                  string       `json:"d-role,omitempty"`
	TextID                string       `json:"d-textId,omitempty"`
	Condition             string       `json:"d-condition,omitempty"`
	Action                string       `json:"d-action,omitempty"`
	Sound                 string       `json:"d-sound,omitempty"`
	Character             *Character   `json:"d-character,omitzero"`
	AlternativeCharacters []*Character `json:"d-alternativeCharacters,omitempty"`
	extra                 jsontext.Value
}

// object is the JSON shape of a type with preserved unknown fields: the typed
// fields of F, followed by the members no typed field owns. Extra is the
// embedded fallback of json v2, which collects those members on reading and
// writes them back, both in the order of the file.
type object[F any] struct {
	Fields *F             `json:",embed"`
	Extra  jsontext.Value `json:",embed"`
}

func marshalObject[F any](fields *F, extra jsontext.Value) ([]byte, error) {
	return json.Marshal(object[F]{fields, extra})
}

// unmarshalObject decodes data into fields and returns the unknown members.
func unmarshalObject[F any](data []byte, fields *F) (jsontext.Value, error) {
	obj := object[F]{Fields: fields}
	err := json.Unmarshal(data, &obj)
	return obj.Extra, err
}

// owns reports whether a typed field of F owns key, by asking json v2 itself:
// a member it does not collect as unknown belongs to a typed field. A key that
// cannot be written, such as one that is not valid UTF-8, is an error.
func owns[F any](key string) (bool, error) {
	probe, err := json.Marshal(map[string]any{key: nil})
	if err != nil {
		return false, err
	}
	extra, err := unmarshalObject(probe, new(F))
	if err != nil {
		return false, err
	}
	return len(extra) == 0, nil
}

// SetExtra attaches a Layer 2 (project-specific) field to the node by
// JSON-marshaling val. It is preserved verbatim on encode and survives a
// round-trip through Decode, the same as any other unrecognised field. A key
// owned by a typed field is ignored, even when that typed field is empty and
// so not written: the typed field always wins.
func (n *Node) SetExtra(key string, val any) error {
	return setExtra(&n.extra, owns[nodeFields], key, val)
}

// SetExtra attaches a Layer 2 (project-specific) field to the edge; the
// contract is the same as Node.SetExtra, so a key owned by a typed field is
// ignored even when that typed field is empty.
func (e *Edge) SetExtra(key string, val any) error {
	return setExtra(&e.extra, owns[edgeFields], key, val)
}

// setExtra sets key to val among the unknown members in extra, in the place
// of an existing key or else after the others. A key a typed field owns is
// left out, since json v2 rejects a duplicated key on writing where v1 let
// the typed field win.
func setExtra(extra *jsontext.Value, owns func(string) (bool, error), key string, val any) error {
	if err := setMember(extra, owns, key, val); err != nil {
		return fmt.Errorf("dcanvas: field %q: %w", key, err)
	}
	return nil
}

func setMember(extra *jsontext.Value, owns func(string) (bool, error), key string, val any) error {
	// Deterministic keeps the keys of a map value sorted, as v1 did, so the
	// same value always writes the same bytes.
	b, err := json.Marshal(val, json.Deterministic(true))
	if err != nil {
		return err
	}
	if owned, err := owns(key); err != nil || owned {
		return err
	}
	var buf bytes.Buffer
	enc := jsontext.NewEncoder(&buf)
	write := func(name string, val jsontext.Value) error {
		if err := enc.WriteToken(jsontext.String(name)); err != nil {
			return err
		}
		return enc.WriteValue(val)
	}
	if err := enc.WriteToken(jsontext.BeginObject); err != nil {
		return err
	}
	found := false
	if len(*extra) > 0 {
		dec := jsontext.NewDecoder(bytes.NewReader(*extra))
		if _, err := dec.ReadToken(); err != nil { // the opening '{'
			return err
		}
		for dec.PeekKind() == '"' {
			tok, err := dec.ReadToken()
			if err != nil {
				return err
			}
			name := tok.String() // read before ReadValue voids the token
			v, err := dec.ReadValue()
			if err != nil {
				return err
			}
			if name == key {
				v, found = b, true
			}
			if err := write(name, v); err != nil {
				return err
			}
		}
	}
	if !found {
		if err := write(key, b); err != nil {
			return err
		}
	}
	if err := enc.WriteToken(jsontext.EndObject); err != nil {
		return err
	}
	*extra = bytes.TrimSpace(buf.Bytes())
	return nil
}

// --- Character IO -----------------------------------------------------------

// UnmarshalJSON decodes a character, keeping any unrecognised field so nested
// unknown keys survive a round-trip.
func (ch *Character) UnmarshalJSON(data []byte) (err error) {
	ch.extra, err = unmarshalObject(data, (*characterFields)(ch))
	return err
}

// MarshalJSON encodes a character, writing back any preserved unknown fields.
func (ch *Character) MarshalJSON() ([]byte, error) {
	return marshalObject((*characterFields)(ch), ch.extra)
}

// --- Canvas IO --------------------------------------------------------------

// UnmarshalJSON decodes a canvas, keeping any unrecognised top-level field.
func (c *Canvas) UnmarshalJSON(data []byte) (err error) {
	c.extra, err = unmarshalObject(data, (*canvasFields)(c))
	return err
}

// MarshalJSON encodes a canvas, stamping the format version and writing back
// any preserved unknown top-level fields. nodes and edges are always emitted as
// arrays so the result is a clean JSON Canvas document.
//
// The decoded d-version is preserved when its major matches the
// library's, so a 1.2 file round-trips as 1.2; a canvas with no version
// (hand-built) or a different major is stamped with the library Version.
func (c *Canvas) MarshalJSON() ([]byte, error) {
	stamped := *c
	if c.Version == "" || majorVersion(c.Version) != majorVersion(Version) {
		stamped.Version = Version
	}
	return marshalObject((*canvasFields)(&stamped), c.extra)
}

// --- Node IO ----------------------------------------------------------------

// UnmarshalJSON decodes a node, keeping any unrecognised field (including
// Layer 2 x- fields).
func (n *Node) UnmarshalJSON(data []byte) (err error) {
	n.extra, err = unmarshalObject(data, (*nodeFields)(n))
	return err
}

// MarshalJSON encodes a node, writing back any preserved unknown fields. d-kind
// is written verbatim; the closed-set check is Validate's job, not the
// writer's (a decoded file's value must round-trip even if unrecognised).
func (n *Node) MarshalJSON() ([]byte, error) {
	if n.Type == "text" {
		return marshalObject((*textNodeFields)(n), n.extra)
	}
	return marshalObject((*nodeFields)(n), n.extra)
}

// --- Edge IO ----------------------------------------------------------------

// UnmarshalJSON decodes an edge, keeping any unrecognised field.
func (e *Edge) UnmarshalJSON(data []byte) (err error) {
	e.extra, err = unmarshalObject(data, (*edgeFields)(e))
	return err
}

// MarshalJSON encodes an edge, writing back any preserved unknown fields; like
// Node.MarshalJSON, it writes d-kind verbatim.
func (e *Edge) MarshalJSON() ([]byte, error) {
	return marshalObject((*edgeFields)(e), e.extra)
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
