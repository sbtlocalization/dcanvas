// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: BlueOak-1.0.0

package dcanvas

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// compileSchema compiles the named shared JSON Schema (draft-07) for the
// conformance cases. The validator is a test-only dependency; the library
// runtime stays standard-library-only.
func compileSchema(t *testing.T, name string) *jsonschema.Schema {
	t.Helper()
	id := schemaIDBase + name
	c := jsonschema.NewCompiler()
	if err := c.AddResource(id, bytes.NewReader(loadSchemaBytes(t, name))); err != nil {
		t.Fatalf("add schema resource: %v", err)
	}
	s, err := c.Compile(id)
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
	schema := compileSchema(t, schema11)

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
						Character: &Character{Name: "Abela the Nymph", TextID: "#1001", Portrait: "abela.png", Gender: "female"}},
				},
			},
		},
		{
			name: "line node with alternative characters",
			canvas: &Canvas{
				Nodes: []*Node{
					{ID: "n1", Type: "text", X: 0, Y: 0, Width: 400, Height: 300,
						Text: "Halt!", Kind: KindLine,
						Character:             &Character{Name: "Guard"},
						AlternativeCharacters: []*Character{{Name: "Captain", TextID: "#2"}}},
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
	schema := compileSchema(t, schema11)

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
			doc := `{"d-version":"1.1","nodes":[` + tc.node + `]}`
			assertConforms(t, validateJSON(t, schema, []byte(doc)), tc.wantPass, doc)
		})
	}
}

// assertConforms fails the test when a validation result err does not match
// the expectation wantPass for document doc.
func assertConforms(t *testing.T, err error, wantPass bool, doc string) {
	t.Helper()
	if wantPass && err != nil {
		t.Errorf("expected the document to conform, got: %v\n%s", err, doc)
	}
	if !wantPass && err == nil {
		t.Errorf("expected the document to be rejected, but it validated:\n%s", doc)
	}
}

// validateJSON parses doc and validates it against schema, returning the
// validation error, if any.
func validateJSON(t *testing.T, schema *jsonschema.Schema, doc []byte) error {
	t.Helper()
	var instance any
	if err := json.Unmarshal(doc, &instance); err != nil {
		t.Fatalf("parse instance: %v", err)
	}
	return schema.Validate(instance)
}

// TestConformance_MinorOnlyAdds is the mechanical proof of "a minor only adds"
// (ADR-0015): a 1.1 document that uses the 1.1 fields validates against the 1.1
// schema and, because those fields are additive, against the 1.0 schema too.
func TestConformance_MinorOnlyAdds(t *testing.T) {
	data := encodeCanvas(t, &Canvas{
		Nodes: []*Node{
			{ID: "n1", Type: "text", X: 0, Y: 0, Width: 400, Height: 300,
				Text: "Greetings.", Kind: KindLine,
				Character: &Character{Name: "Guard", TextID: "#1001"}},
		},
	})
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("re-parse encoded canvas: %v", err)
	}
	ch := m["nodes"].([]any)[0].(map[string]any)["d-character"].(map[string]any)
	if m["d-version"] != "1.1" || ch["textId"] != "#1001" {
		t.Fatalf("encoded output is not a 1.1 document using d-character.textId:\n%s", data)
	}
	for _, name := range schemas {
		t.Run(name, func(t *testing.T) {
			if err := validateJSON(t, compileSchema(t, name), data); err != nil {
				t.Errorf("1.1 document does not conform to %s:\n%s\noutput:\n%s", name, err, data)
			}
		})
	}
}

// TestConformance_CharacterTextIDIsAString checks the 1.1 schema types
// d-character.textId as a string, so a malformed reference is caught.
func TestConformance_CharacterTextIDIsAString(t *testing.T) {
	schema := compileSchema(t, schema11)
	const node = `{"id":"n","type":"text","x":0,"y":0,"width":400,"height":300,"text":"hi","d-character":`
	cases := []struct {
		name      string
		character string
		wantPass  bool
	}{
		{"string textId", `{"name":"Guard","textId":"#1001"}`, true},
		{"numeric textId", `{"name":"Guard","textId":1001}`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			doc := `{"d-version":"1.1","nodes":[` + node + tc.character + `}]}`
			assertConforms(t, validateJSON(t, schema, []byte(doc)), tc.wantPass, doc)
		})
	}
}

// TestConformance_SchemasAcceptMajorOne checks that each per-minor schema
// constrains d-version by major (ADR-0015): any 1.x passes, 2.0 fails.
func TestConformance_SchemasAcceptMajorOne(t *testing.T) {
	for _, name := range schemas {
		schema := compileSchema(t, name)
		for version, wantPass := range map[string]bool{"1.0": true, "1.1": true, "1.2": true, "2.0": false} {
			t.Run(name+"/"+version, func(t *testing.T) {
				doc := `{"d-version":"` + version + `"}`
				assertConforms(t, validateJSON(t, schema, []byte(doc)), wantPass, doc)
			})
		}
	}
}

// TestConformance_AlternativeCharacters checks the 1.1 schema: a non-empty list
// of characters with a required name, allowed only beside d-character. A 1.0
// schema, which does not know the field, accepts a valid document using it.
func TestConformance_AlternativeCharacters(t *testing.T) {
	const node = `{"id":"n","type":"text","x":0,"y":0,"width":400,"height":300,"text":"hi",`
	cases := []struct {
		name     string
		fields   string
		wantPass bool
	}{
		{"with d-character", `"d-character":{"name":"A"},"d-alternativeCharacters":[{"name":"B"}]`, true},
		{"without d-character", `"d-alternativeCharacters":[{"name":"B"}]`, false},
		{"empty list", `"d-character":{"name":"A"},"d-alternativeCharacters":[]`, false},
		{"alternative without name", `"d-character":{"name":"A"},"d-alternativeCharacters":[{"textId":"#1"}]`, false},
		{"not an array", `"d-character":{"name":"A"},"d-alternativeCharacters":{"name":"B"}`, false},
	}
	schema := compileSchema(t, schema11)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			doc := `{"d-version":"1.1","nodes":[` + node + tc.fields + `}]}`
			assertConforms(t, validateJSON(t, schema, []byte(doc)), tc.wantPass, doc)
		})
	}
}

// TestConformance_AlternativeCharactersOutputValidatesAgainstOneZero checks that
// the library's output with alternatives, additive in 1.1, also conforms to 1.0.
func TestConformance_AlternativeCharactersOutputValidatesAgainstOneZero(t *testing.T) {
	data := encodeCanvas(t, &Canvas{
		Nodes: []*Node{
			{ID: "n1", Type: "text", Width: 400, Height: 300, Text: "Halt!", Kind: KindLine,
				Character:             &Character{Name: "Guard"},
				AlternativeCharacters: []*Character{{Name: "Captain", TextID: "#2"}}},
		},
	})
	if !strings.Contains(string(data), "d-alternativeCharacters") {
		t.Fatalf("encoded output lacks d-alternativeCharacters:\n%s", data)
	}
	assertConforms(t, validateJSON(t, compileSchema(t, schema10), data), true, string(data))
}
