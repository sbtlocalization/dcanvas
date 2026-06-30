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
)

// decodeString is a small helper around Decode for table-driven tests.
func decodeString(t *testing.T, s string) (*Canvas, error) {
	t.Helper()
	return Decode(strings.NewReader(s))
}

// encodeToMap encodes c and parses the bytes back into a generic map, so tests
// assert on the externally observable JSON shape rather than the internal
// catch-all representation.
func encodeToMap(t *testing.T, c *Canvas) map[string]any {
	t.Helper()
	var buf bytes.Buffer
	if err := Encode(c, &buf); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("re-parse encoded canvas: %v", err)
	}
	return m
}

func TestDecode_VersionGate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"v3.0 accepted", `{"x-dCanvasVersion":"3.0","nodes":[],"edges":[]}`, false},
		{"v3.1 accepted (higher minor, same major)", `{"x-dCanvasVersion":"3.1"}`, false},
		{"v2.0 rejected", `{"x-dCanvasVersion":"2.0","nodes":[],"edges":[]}`, true},
		{"v4.0 rejected", `{"x-dCanvasVersion":"4.0"}`, true},
		{"missing version rejected", `{"nodes":[],"edges":[]}`, true},
		{"malformed json rejected", `{not json`, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := decodeString(t, tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Decode(%q) err = %v, wantErr = %v", tc.input, err, tc.wantErr)
			}
		})
	}
}

func TestEncode_StampsVersion(t *testing.T) {
	// Even a canvas whose Version field says something else is stamped "3.0".
	c := &Canvas{Version: "9.9"}
	m := encodeToMap(t, c)
	if got := m["x-dCanvasVersion"]; got != "3.0" {
		t.Errorf("x-dCanvasVersion = %v, want \"3.0\"", got)
	}
	if _, ok := m["nodes"].([]any); !ok {
		t.Errorf("nodes should be emitted as an array, got %T", m["nodes"])
	}
	if _, ok := m["edges"].([]any); !ok {
		t.Errorf("edges should be emitted as an array, got %T", m["edges"])
	}
}

func TestEncode_StrippedIsValidJSONCanvas(t *testing.T) {
	// A document with no Layer 1 data must produce no x- fields on the node
	// beyond nothing — i.e. only standard JSON Canvas fields remain.
	c := &Canvas{
		Nodes: []*Node{{ID: "n1", Type: "text", X: 1, Y: 2, Width: 400, Height: 300, Text: "hi"}},
	}
	m := encodeToMap(t, c)
	node := m["nodes"].([]any)[0].(map[string]any)
	for k := range node {
		if strings.HasPrefix(k, "x-") {
			t.Errorf("node carried unexpected x- field %q for empty Layer 1", k)
		}
	}
	// Standard fields present.
	for _, k := range []string{"id", "type", "x", "y", "width", "height", "text"} {
		if _, ok := node[k]; !ok {
			t.Errorf("node missing standard field %q", k)
		}
	}
}

func TestRoundTrip_PreservesUnknownFields(t *testing.T) {
	input := `{
		"x-dCanvasVersion": "3.0",
		"x-unknownTopLevel": {"deep": [1, 2, 3]},
		"nodes": [
			{
				"id": "n1", "type": "text",
				"x": 0, "y": 0, "width": 400, "height": 300,
				"text": "Greetings.",
				"x-kind": "line",
				"x-projectField": "keep me",
				"unknownStandardLooking": 42
			}
		],
		"edges": [
			{
				"id": "e1", "fromNode": "n1", "toNode": "n1",
				"label": "loop back",
				"x-kind": "loop",
				"x-edgeProjectField": ["a", "b"]
			}
		]
	}`

	c, err := decodeString(t, input)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	m := encodeToMap(t, c)

	if _, ok := m["x-unknownTopLevel"]; !ok {
		t.Error("unknown top-level field was dropped")
	}
	node := m["nodes"].([]any)[0].(map[string]any)
	if node["x-projectField"] != "keep me" {
		t.Errorf("unknown node x- field dropped: %v", node["x-projectField"])
	}
	if node["unknownStandardLooking"].(float64) != 42 {
		t.Errorf("unknown standard-looking node field dropped: %v", node["unknownStandardLooking"])
	}
	edge := m["edges"].([]any)[0].(map[string]any)
	if _, ok := edge["x-edgeProjectField"]; !ok {
		t.Error("unknown edge x- field was dropped")
	}
}

func TestRoundTrip_TypedLayer1Faithful(t *testing.T) {
	input := `{
		"x-dCanvasVersion": "3.0",
		"nodes": [
			{
				"id": "n-ABELA-5", "type": "text",
				"x": 10, "y": 20, "width": 400, "height": 300,
				"color": "3",
				"text": "My poor Ragefast.",
				"x-id": "ABELA[5]",
				"x-kind": "line",
				"x-role": "state",
				"x-textId": "#2687",
				"x-condition": "Dead(\"Ragefast\")",
				"x-action": "DoThing()",
				"x-sound": "ABELA05.wav",
				"x-character": {"name": "Abela the Nymph", "portrait": "None.png", "gender": "female"}
			}
		],
		"edges": [
			{
				"id": "e-5-6", "fromNode": "n-ABELA-5", "fromSide": "bottom",
				"toNode": "n-ABELA-6", "toSide": "top", "toEnd": "arrow",
				"label": "Help her escape",
				"x-id": "ABELA[5]->[6]",
				"x-kind": "normal",
				"x-role": "paraphrase",
				"x-condition": "Global(\"x\",\"GLOBAL\",1)",
				"x-textId": "#2690"
			}
		]
	}`

	c, err := decodeString(t, input)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	n := c.Nodes[0]
	wantNode := Node{
		ID: "n-ABELA-5", Type: "text", X: 10, Y: 20, Width: 400, Height: 300,
		Color: "3", Text: "My poor Ragefast.",
		XID: "ABELA[5]", Kind: KindLine, Role: "state", TextID: "#2687",
		Condition: "Dead(\"Ragefast\")", Action: "DoThing()", Sound: "ABELA05.wav",
	}
	if n.ID != wantNode.ID || n.Type != wantNode.Type || n.X != wantNode.X || n.Y != wantNode.Y ||
		n.Width != wantNode.Width || n.Height != wantNode.Height || n.Color != wantNode.Color ||
		n.Text != wantNode.Text || n.XID != wantNode.XID || n.Kind != wantNode.Kind ||
		n.Role != wantNode.Role || n.TextID != wantNode.TextID || n.Condition != wantNode.Condition ||
		n.Action != wantNode.Action || n.Sound != wantNode.Sound {
		t.Errorf("decoded node = %+v, want subset %+v", *n, wantNode)
	}
	if n.Character == nil || n.Character.Name != "Abela the Nymph" || n.Character.Portrait != "None.png" || n.Character.Gender != "female" {
		t.Errorf("decoded character = %+v", n.Character)
	}

	e := c.Edges[0]
	if e.ID != "e-5-6" || e.FromNode != "n-ABELA-5" || e.FromSide != "bottom" ||
		e.ToNode != "n-ABELA-6" || e.ToSide != "top" || e.ToEnd != "arrow" ||
		e.Label != "Help her escape" || e.XID != "ABELA[5]->[6]" || e.Kind != KindNormal ||
		e.Role != "paraphrase" || e.TextID != "#2690" {
		t.Errorf("decoded edge = %+v", *e)
	}

	// Re-encode and confirm the typed fields survive on the wire.
	m := encodeToMap(t, c)
	node := m["nodes"].([]any)[0].(map[string]any)
	for k, want := range map[string]any{
		"x-id": "ABELA[5]", "x-kind": "line", "x-role": "state",
		"x-textId": "#2687", "x-action": "DoThing()", "x-sound": "ABELA05.wav",
	} {
		if node[k] != want {
			t.Errorf("re-encoded node[%q] = %v, want %v", k, node[k], want)
		}
	}
	ch := node["x-character"].(map[string]any)
	if ch["name"] != "Abela the Nymph" || ch["gender"] != "female" {
		t.Errorf("re-encoded character = %v", ch)
	}
}

func TestEncode_RejectsInvalidKind(t *testing.T) {
	tests := []struct {
		name   string
		canvas *Canvas
	}{
		{"bad node kind", &Canvas{Nodes: []*Node{{ID: "n", Type: "text", Kind: "shout"}}}},
		{"bad edge kind", &Canvas{Edges: []*Edge{{ID: "e", FromNode: "a", ToNode: "b", Kind: "weird"}}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := Encode(tc.canvas, &bytes.Buffer{})
			if err == nil {
				t.Fatal("expected an error for invalid x-kind, got nil")
			}
		})
	}
}

func TestEncode_AcceptsValidKinds(t *testing.T) {
	c := &Canvas{
		Nodes: []*Node{{ID: "n", Type: "text", Kind: KindReply, Text: "ok"}},
		Edges: []*Edge{{ID: "e", FromNode: "n", ToNode: "n", Kind: KindLoop}},
	}
	if err := Encode(c, &bytes.Buffer{}); err != nil {
		t.Fatalf("Encode with valid kinds: %v", err)
	}
}
