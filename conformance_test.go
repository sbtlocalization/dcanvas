// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: BlueOak-1.0.0

package dcanvas

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// compileSchema compiles the shared JSON Schema (draft-07) once for the
// conformance cases. The validator is a test-only dependency; the library
// runtime stays standard-library-only.
func compileSchema(t *testing.T) *jsonschema.Schema {
	t.Helper()
	c := jsonschema.NewCompiler()
	if err := c.AddResource(canonicalSchemaID, bytes.NewReader(loadSchemaBytes(t))); err != nil {
		t.Fatalf("add schema resource: %v", err)
	}
	s, err := c.Compile(canonicalSchemaID)
	if err != nil {
		t.Fatalf("compile schema: %v", err)
	}
	return s
}

// encodeCanvas encodes c through the public API and returns the produced bytes.
func encodeCanvas(t *testing.T, c *Canvas) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := Encode(c, &buf); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	return buf.Bytes()
}

// replyWithProjectField builds a reply node carrying a Layer 2 project field
// through the public SetExtra API, panicking only on a programmer error (the
// value is a plain string, which always marshals).
func replyWithProjectField(t *testing.T) *Node {
	t.Helper()
	n := &Node{ID: "n2", Type: "text", X: 0, Y: 400, Width: 400, Height: 300,
		Text: "I'll help you.", Kind: KindReply}
	if err := n.SetExtra("x-journalText", "The Captive Nymph"); err != nil {
		t.Fatalf("SetExtra: %v", err)
	}
	return n
}

// TestConformance_EncodedOutputValidatesAgainstSchema is the end-to-end
// cross-language guarantee: representative canvases built through the public API
// and encoded must satisfy the very schema the TypeScript reader consumes. It
// asserts on the encoded bytes, never on internal fields.
func TestConformance_EncodedOutputValidatesAgainstSchema(t *testing.T) {
	schema := compileSchema(t)

	cases := []struct {
		name   string
		canvas *Canvas
	}{
		{
			name: "line node with d-character",
			canvas: &Canvas{
				Nodes: []*Node{
					{ID: "n1", Type: "text", X: 0, Y: 0, Width: 400, Height: 300,
						Text: "My poor Ragefast.", Kind: KindLine,
						Character: &Character{Name: "Abela the Nymph", Portrait: "abela.png", Gender: "female"}},
				},
			},
		},
		{
			name: "reply node with a Layer 2 project field",
			canvas: &Canvas{
				Nodes: []*Node{replyWithProjectField(t)},
			},
		},
		{
			name: "edge with a label and a d-condition",
			canvas: &Canvas{
				Nodes: []*Node{
					{ID: "n1", Type: "text", X: 0, Y: 0, Width: 400, Height: 300, Text: "Choose.", Kind: KindLine},
					{ID: "n2", Type: "text", X: 0, Y: 400, Width: 400, Height: 300, Text: "Onward.", Kind: KindReply},
				},
				Edges: []*Edge{
					{ID: "e1", FromNode: "n1", ToNode: "n2", ToEnd: "arrow",
						Label: "Help her escape", Kind: KindNormal, Condition: "Global(\"x\",\"GLOBAL\",1)"},
				},
			},
		},
		{
			name:   "minimal complete document",
			canvas: &Canvas{},
		},
		{
			// The schema requires text on text-type nodes. A text node with empty
			// text must still emit a (possibly empty) text field so the output
			// conforms — see task 0007. Before that fix this case is red.
			name: "text node with empty text still conforms",
			canvas: &Canvas{
				Nodes: []*Node{
					{ID: "n1", Type: "text", X: 0, Y: 0, Width: 400, Height: 300},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := encodeCanvas(t, tc.canvas)

			var instance any
			if err := json.Unmarshal(data, &instance); err != nil {
				t.Fatalf("re-parse encoded canvas: %v", err)
			}
			if err := schema.Validate(instance); err != nil {
				t.Errorf("encoded output does not conform to schema:\n%s\noutput:\n%s", err, data)
			}
		})
	}
}

// TestConformance_FileAndLinkInheritRequiredFields verifies that dCanvas keeps
// JSON Canvas 1.0's required fields for the node types it does not redefine: a
// file node MUST carry file, a link node MUST carry url. dCanvas does not use
// these node types, so it inherits them wholesale rather than relaxing them.
// The cases run raw JSON against the schema (the typed API models neither
// field), asserting both directions: present passes, absent is rejected.
func TestConformance_FileAndLinkInheritRequiredFields(t *testing.T) {
	schema := compileSchema(t)

	const geom = `"x":0,"y":0,"width":400,"height":300`
	cases := []struct {
		name     string
		node     string
		wantPass bool
	}{
		{"file node with file", `{"id":"n","type":"file",` + geom + `,"file":"a.png"}`, true},
		{"file node without file", `{"id":"n","type":"file",` + geom + `}`, false},
		{"link node with url", `{"id":"n","type":"link",` + geom + `,"url":"https://x"}`, true},
		{"link node without url", `{"id":"n","type":"link",` + geom + `}`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			doc := `{"d-version":"1.0","nodes":[` + tc.node + `]}`
			var instance any
			if err := json.Unmarshal([]byte(doc), &instance); err != nil {
				t.Fatalf("parse instance: %v", err)
			}
			err := schema.Validate(instance)
			if tc.wantPass && err != nil {
				t.Errorf("expected the document to conform, got: %v", err)
			}
			if !tc.wantPass && err == nil {
				t.Errorf("expected the document to be rejected, but it validated:\n%s", doc)
			}
		})
	}
}
