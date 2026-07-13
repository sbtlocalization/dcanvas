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

func TestEncode_VersionContract(t *testing.T) {
	// Encode preserves a decoded version whose major matches the library's, and
	// otherwise stamps the library version.
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{"no version stamped with library version", "", Version},
		{"same-major minor is preserved", "3.1", "3.1"},
		{"different major replaced with library version", "9.9", Version},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := encodeToMap(t, &Canvas{Version: tc.version})
			if got := m["x-dCanvasVersion"]; got != tc.want {
				t.Errorf("x-dCanvasVersion = %v, want %q", got, tc.want)
			}
		})
	}
}

func TestEncode_NodesEdgesAlwaysArrays(t *testing.T) {
	m := encodeToMap(t, &Canvas{})
	if _, ok := m["nodes"].([]any); !ok {
		t.Errorf("nodes should be emitted as an array, got %T", m["nodes"])
	}
	if _, ok := m["edges"].([]any); !ok {
		t.Errorf("edges should be emitted as an array, got %T", m["edges"])
	}
}

func TestRoundTrip_PreservesMinorVersion(t *testing.T) {
	// A 3.1 document must round-trip as 3.1, not be downgraded to the library's
	// 3.0.
	c, err := decodeString(t, `{"x-dCanvasVersion":"3.1","nodes":[],"edges":[]}`)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	m := encodeToMap(t, c)
	if got := m["x-dCanvasVersion"]; got != "3.1" {
		t.Errorf("x-dCanvasVersion = %v, want \"3.1\"", got)
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

// objectKeysInOrder reads the keys of the first JSON object found in data, in
// document order (json.Decoder preserves it, unlike a map). Nested values are
// skipped so only the immediate object's keys are returned.
func objectKeysInOrder(t *testing.T, data []byte) []string {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(data))
	for { // advance to the first '{'
		tok, err := dec.Token()
		if err != nil {
			t.Fatalf("scanning for object start: %v", err)
		}
		if d, ok := tok.(json.Delim); ok && d == '{' {
			break
		}
	}
	var keys []string
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			t.Fatalf("reading key: %v", err)
		}
		keys = append(keys, keyTok.(string))
		if err := skipJSONValue(dec); err != nil {
			t.Fatalf("skipping value: %v", err)
		}
	}
	return keys
}

// skipJSONValue consumes exactly one value (scalar or a fully nested
// object/array) from the decoder.
func skipJSONValue(dec *json.Decoder) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	d, ok := tok.(json.Delim)
	if !ok || (d != '{' && d != '[') {
		return nil // scalar
	}
	for depth := 1; depth > 0; {
		t2, err := dec.Token()
		if err != nil {
			return err
		}
		if dd, ok := t2.(json.Delim); ok {
			if dd == '{' || dd == '[' {
				depth++
			} else {
				depth--
			}
		}
	}
	return nil
}

func TestEncode_KeyOrder(t *testing.T) {
	c := &Canvas{
		Nodes: []*Node{{
			ID: "n1", Type: "text", X: 1, Y: 2, Width: 400, Height: 300,
			Color: "3", Text: "hi", XID: "X", Kind: KindLine, Role: "state",
		}},
		Edges: []*Edge{{ID: "e1", FromNode: "n1", ToNode: "n1", Label: "go", Kind: KindNormal}},
	}
	var buf bytes.Buffer
	if err := Encode(c, &buf); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	data := buf.Bytes()

	top := objectKeysInOrder(t, data)
	wantTop := []string{"x-dCanvasVersion", "nodes", "edges"}
	if strings.Join(top, ",") != strings.Join(wantTop, ",") {
		t.Errorf("top-level key order = %v, want %v", top, wantTop)
	}

	// The node object is the first object inside the "nodes" array; find it by
	// scanning past the top-level keys to the first nested '{'.
	idx := bytes.IndexByte(data, '{')
	nodeStart := bytes.IndexByte(data[idx+1:], '{')
	nodeKeys := objectKeysInOrder(t, data[idx+1+nodeStart:])
	wantNode := []string{"id", "type", "x", "y", "width", "height", "color", "text", "x-id", "x-kind", "x-role"}
	if strings.Join(nodeKeys, ",") != strings.Join(wantNode, ",") {
		t.Errorf("node key order = %v, want %v", nodeKeys, wantNode)
	}
}

func TestRoundTrip_PreservesUnrecognisedKind(t *testing.T) {
	// Encode no longer enforces the closed x-kind set (that is Validate's job),
	// so a file whose x-kind the library does not recognise round-trips without
	// error and keeps its value on the wire.
	input := `{
		"x-dCanvasVersion": "3.0",
		"nodes": [{"id":"n","type":"text","x":0,"y":0,"width":400,"height":300,"text":"hi","x-kind":"shout"}],
		"edges": [{"id":"e","fromNode":"n","toNode":"n","x-kind":"weird"}]
	}`
	c, err := decodeString(t, input)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	var buf bytes.Buffer
	if err := Encode(c, &buf); err != nil {
		t.Fatalf("Encode should not reject an unrecognised x-kind: %v", err)
	}
	m := encodeToMap(t, c)
	if got := m["nodes"].([]any)[0].(map[string]any)["x-kind"]; got != "shout" {
		t.Errorf("node x-kind = %v, want \"shout\"", got)
	}
	if got := m["edges"].([]any)[0].(map[string]any)["x-kind"]; got != "weird" {
		t.Errorf("edge x-kind = %v, want \"weird\"", got)
	}
}

func TestNode_SetExtra(t *testing.T) {
	c := &Canvas{
		Nodes: []*Node{{ID: "n", Type: "text", Text: "hi"}},
	}
	if err := c.Nodes[0].SetExtra("x-journalText", "The Captive Nymph"); err != nil {
		t.Fatalf("SetExtra: %v", err)
	}

	m := encodeToMap(t, c)
	node := m["nodes"].([]any)[0].(map[string]any)
	if node["x-journalText"] != "The Captive Nymph" {
		t.Errorf("node[\"x-journalText\"] = %v, want \"The Captive Nymph\"", node["x-journalText"])
	}

	// Round-trips through Decode like any other unknown field.
	var buf bytes.Buffer
	if err := Encode(c, &buf); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	c2, err := Decode(&buf)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	m2 := encodeToMap(t, c2)
	node2 := m2["nodes"].([]any)[0].(map[string]any)
	if node2["x-journalText"] != "The Captive Nymph" {
		t.Errorf("after round-trip, node[\"x-journalText\"] = %v", node2["x-journalText"])
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
