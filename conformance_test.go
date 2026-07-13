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
			name: "line node with x-character",
			canvas: &Canvas{
				Nodes: []*Node{
					{ID: "n1", Type: "text", X: 0, Y: 0, Width: 400, Height: 300,
						Text: "My poor Ragefast.", Kind: KindLine,
						Character: &Character{Name: "Abela the Nymph", Portrait: "None.png", Gender: "female"}},
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
			name: "edge with a label and an x-condition",
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
