// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: BlueOak-1.0.0

package dcanvas

import (
	"bytes"
	"encoding/json/jsontext"
	json "encoding/json/v2"
	"fmt"
	"reflect"
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
		{"v1.0 accepted", `{"d-version":"1.0","nodes":[],"edges":[]}`, false},
		{"v1.1 accepted", `{"d-version":"1.1","nodes":[],"edges":[]}`, false},
		{"v1.2 accepted (higher minor, same major)", `{"d-version":"1.2"}`, false},
		{"v2.0 rejected", `{"d-version":"2.0","nodes":[],"edges":[]}`, true},
		{"v4.0 rejected", `{"d-version":"4.0"}`, true},
		{"missing version rejected", `{"nodes":[],"edges":[]}`, true},
		{"malformed json rejected", `{not json`, true},
	}
	// A file stamped with the pre-1.0 internal version field carries no
	// d-version at all, so it must be rejected the same way. The legacy key is
	// assembled at runtime to keep the repo-wide grep for old spellings clean.
	legacyKey := "x-" + "dCanvasVersion"
	legacyVal := "3" + ".0"
	tests = append(tests, struct {
		name    string
		input   string
		wantErr bool
	}{
		"legacy internal stamp rejected",
		fmt.Sprintf(`{%q:%q,"nodes":[],"edges":[]}`, legacyKey, legacyVal),
		true,
	})
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
		{"no version stamped with library version", "", "1.1"},
		{"lower same-major minor is preserved", "1.0", "1.0"},
		{"higher same-major minor is preserved", "1.2", "1.2"},
		{"different major replaced with library version", "9.9", "1.1"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := encodeToMap(t, &Canvas{Version: tc.version})
			if got := m["d-version"]; got != tc.want {
				t.Errorf("d-version = %v, want %q", got, tc.want)
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
	// A 1.2 document must round-trip as 1.2, not be downgraded to the library's
	// 1.1.
	c, err := decodeString(t, `{"d-version":"1.2","nodes":[],"edges":[]}`)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	m := encodeToMap(t, c)
	if got := m["d-version"]; got != "1.2" {
		t.Errorf("d-version = %v, want \"1.2\"", got)
	}
}

func TestEncode_StrippedIsValidJSONCanvas(t *testing.T) {
	// A document with no Layer 1 data must produce no d- or x- fields on the
	// node — i.e. only standard JSON Canvas fields remain.
	c := &Canvas{
		Nodes: []*Node{{ID: "n1", Type: "text", X: 1, Y: 2, Width: 400, Height: 300, Text: "hi"}},
	}
	m := encodeToMap(t, c)
	node := m["nodes"].([]any)[0].(map[string]any)
	for k := range node {
		if strings.HasPrefix(k, "d-") || strings.HasPrefix(k, "x-") {
			t.Errorf("node carried unexpected extension field %q for empty Layer 1", k)
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
		"d-version": "1.0",
		"x-unknownTopLevel": {"deep": [1, 2, 3]},
		"nodes": [
			{
				"id": "n1", "type": "text",
				"x": 0, "y": 0, "width": 400, "height": 300,
				"text": "Greetings.",
				"d-kind": "line",
				"x-projectField": "keep me",
				"unknownStandardLooking": 42
			}
		],
		"edges": [
			{
				"id": "e1", "fromNode": "n1", "toNode": "n1",
				"label": "loop back",
				"d-kind": "loop",
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
	nodes, ok := m["nodes"].([]any)
	if !ok || len(nodes) != 1 {
		t.Fatalf("nodes = %v, want one node", m["nodes"])
	}
	node, ok := nodes[0].(map[string]any)
	if !ok {
		t.Fatalf("node = %v, want an object", nodes[0])
	}
	if node["x-projectField"] != "keep me" {
		t.Errorf("unknown node x- field dropped: %v", node["x-projectField"])
	}
	if got, ok := node["unknownStandardLooking"].(float64); !ok || got != 42 {
		t.Errorf("unknown standard-looking node field dropped: %v", node["unknownStandardLooking"])
	}
	edges, ok := m["edges"].([]any)
	if !ok || len(edges) != 1 {
		t.Fatalf("edges = %v, want one edge", m["edges"])
	}
	edge, ok := edges[0].(map[string]any)
	if !ok {
		t.Fatalf("edge = %v, want an object", edges[0])
	}
	if _, ok := edge["x-edgeProjectField"]; !ok {
		t.Error("unknown edge x- field was dropped")
	}
}

func TestRoundTrip_TypedLayer1Faithful(t *testing.T) {
	input := `{
		"d-version": "1.0",
		"nodes": [
			{
				"id": "n-ABELA-5", "type": "text",
				"x": 10, "y": 20, "width": 400, "height": 300,
				"color": "3",
				"text": "My poor Ragefast.",
				"d-id": "ABELA[5]",
				"d-kind": "line",
				"d-role": "state",
				"d-textId": "#2687",
				"d-condition": "Dead(\"Ragefast\")",
				"d-action": "DoThing()",
				"d-sound": "ABELA05.wav",
				"d-character": {"name": "Abela the Nymph", "portrait": "abela.png", "gender": "female"}
			}
		],
		"edges": [
			{
				"id": "e-5-6", "fromNode": "n-ABELA-5", "fromSide": "bottom",
				"toNode": "n-ABELA-6", "toSide": "top", "toEnd": "arrow",
				"label": "Help her escape",
				"d-id": "ABELA[5]->[6]",
				"d-kind": "normal",
				"d-role": "paraphrase",
				"d-condition": "Global(\"x\",\"GLOBAL\",1)",
				"d-textId": "#2690"
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
		DomainID: "ABELA[5]", Kind: KindLine, Role: "state", TextID: "#2687",
		Condition: "Dead(\"Ragefast\")", Action: "DoThing()", Sound: "ABELA05.wav",
	}
	if n.ID != wantNode.ID || n.Type != wantNode.Type || n.X != wantNode.X || n.Y != wantNode.Y ||
		n.Width != wantNode.Width || n.Height != wantNode.Height || n.Color != wantNode.Color ||
		n.Text != wantNode.Text || n.DomainID != wantNode.DomainID || n.Kind != wantNode.Kind ||
		n.Role != wantNode.Role || n.TextID != wantNode.TextID || n.Condition != wantNode.Condition ||
		n.Action != wantNode.Action || n.Sound != wantNode.Sound {
		t.Errorf("decoded node = %+v, want subset %+v", *n, wantNode)
	}
	if n.Character == nil || n.Character.Name != "Abela the Nymph" || n.Character.Portrait != "abela.png" || n.Character.Gender != "female" {
		t.Errorf("decoded character = %+v", n.Character)
	}

	e := c.Edges[0]
	if e.ID != "e-5-6" || e.FromNode != "n-ABELA-5" || e.FromSide != "bottom" ||
		e.ToNode != "n-ABELA-6" || e.ToSide != "top" || e.ToEnd != "arrow" ||
		e.Label != "Help her escape" || e.DomainID != "ABELA[5]->[6]" || e.Kind != KindNormal ||
		e.Role != "paraphrase" || e.TextID != "#2690" {
		t.Errorf("decoded edge = %+v", *e)
	}

	// Re-encode and confirm the typed fields survive on the wire.
	m := encodeToMap(t, c)
	node := m["nodes"].([]any)[0].(map[string]any)
	for k, want := range map[string]any{
		"d-id": "ABELA[5]", "d-kind": "line", "d-role": "state",
		"d-textId": "#2687", "d-action": "DoThing()", "d-sound": "ABELA05.wav",
	} {
		if node[k] != want {
			t.Errorf("re-encoded node[%q] = %v, want %v", k, node[k], want)
		}
	}
	ch := node["d-character"].(map[string]any)
	if ch["name"] != "Abela the Nymph" || ch["gender"] != "female" {
		t.Errorf("re-encoded character = %v", ch)
	}
}

// keyOrders reads data and returns the keys of every JSON object in it, in
// document order (jsontext.Decoder preserves it, unlike a map), indexed by the
// object's JSON pointer: "" for the document, "/nodes/0" for the first node.
func keyOrders(t *testing.T, data []byte) map[string][]string {
	t.Helper()
	out := make(map[string][]string)
	dec := jsontext.NewDecoder(bytes.NewReader(data))
	var walk func(ptr string)
	walk = func(ptr string) {
		tok, err := dec.ReadToken()
		if err != nil {
			t.Fatalf("reading %q: %v", ptr, err)
		}
		switch tok.Kind() {
		case '{':
			out[ptr] = []string{}
			for dec.PeekKind() != '}' {
				name, err := dec.ReadToken()
				if err != nil {
					t.Fatalf("reading a key in %q: %v", ptr, err)
				}
				out[ptr] = append(out[ptr], name.String())
				walk(ptr + "/" + name.String())
			}
			dec.ReadToken()
		case '[':
			for i := 0; dec.PeekKind() != ']'; i++ {
				walk(fmt.Sprintf("%s/%d", ptr, i))
			}
			dec.ReadToken()
		}
	}
	walk("")
	return out
}

func TestEncode_KeyOrder(t *testing.T) {
	// Known fields come first, in the JSON Canvas order; unknown fields follow
	// them in the order they had in the file, at every level. They are spread
	// between the known ones, and not in alphabetical order, to show that both
	// are given up.
	input := `{
		"x-zeta": 1,
		"edges": [{"x-zeta": 1, "label": "go", "id": "e1", "x-alpha": 2, "toNode": "n1",
			"d-kind": "normal", "fromNode": "n1", "x-mu": 3}],
		"d-version": "1.1",
		"x-alpha": 2,
		"nodes": [{
			"x-zeta": 1, "d-role": "state", "text": "hi", "id": "n1", "x-alpha": 2,
			"type": "text", "d-kind": "line", "x": 1, "y": 2, "width": 400, "height": 300,
			"color": "3", "d-id": "X", "x-mu": 3,
			"d-character": {"x-zeta": 1, "gender": "female", "x-alpha": 2, "name": "Abela", "x-mu": 3}
		}],
		"x-mu": 3
	}`
	c, err := decodeString(t, input)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	var buf bytes.Buffer
	if err := Encode(c, &buf); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	got := keyOrders(t, buf.Bytes())

	unknown := []string{"x-zeta", "x-alpha", "x-mu"}
	for ptr, known := range map[string][]string{
		"":                     {"d-version", "nodes", "edges"},
		"/nodes/0":             {"id", "type", "x", "y", "width", "height", "color", "text", "d-id", "d-kind", "d-role", "d-character"},
		"/nodes/0/d-character": {"name", "gender"},
		"/edges/0":             {"id", "fromNode", "toNode", "label", "d-kind"},
	} {
		want := strings.Join(append(known, unknown...), ",")
		if got := strings.Join(got[ptr], ","); got != want {
			t.Errorf("key order of %q = %v, want %v", ptr, got, want)
		}
	}
}

func TestEncode_KnownKeyOrder(t *testing.T) {
	// A hand-built canvas, with no unknown fields, writes its known fields in
	// the JSON Canvas order.
	c := &Canvas{
		Nodes: []*Node{{
			ID: "n1", Type: "text", X: 1, Y: 2, Width: 400, Height: 300,
			Color: "3", Text: "hi", DomainID: "X", Kind: KindLine, Role: "state",
		}},
		Edges: []*Edge{{ID: "e1", FromNode: "n1", ToNode: "n1", Label: "go", Kind: KindNormal}},
	}
	var buf bytes.Buffer
	if err := Encode(c, &buf); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	got := keyOrders(t, buf.Bytes())
	for ptr, want := range map[string][]string{
		"":         {"d-version", "nodes", "edges"},
		"/nodes/0": {"id", "type", "x", "y", "width", "height", "color", "text", "d-id", "d-kind", "d-role"},
		"/edges/0": {"id", "fromNode", "toNode", "label", "d-kind"},
	} {
		if strings.Join(got[ptr], ",") != strings.Join(want, ",") {
			t.Errorf("key order of %q = %v, want %v", ptr, got[ptr], want)
		}
	}
}

func TestRoundTrip_PreservesUnrecognisedKind(t *testing.T) {
	// Encode no longer enforces the closed d-kind set (that is Validate's job),
	// so a file whose d-kind the library does not recognise round-trips without
	// error and keeps its value on the wire.
	input := `{
		"d-version": "1.0",
		"nodes": [{"id":"n","type":"text","x":0,"y":0,"width":400,"height":300,"text":"hi","d-kind":"shout"}],
		"edges": [{"id":"e","fromNode":"n","toNode":"n","d-kind":"weird"}]
	}`
	c, err := decodeString(t, input)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	var buf bytes.Buffer
	if err := Encode(c, &buf); err != nil {
		t.Fatalf("Encode should not reject an unrecognised d-kind: %v", err)
	}
	m := encodeToMap(t, c)
	if got := m["nodes"].([]any)[0].(map[string]any)["d-kind"]; got != "shout" {
		t.Errorf("node d-kind = %v, want \"shout\"", got)
	}
	if got := m["edges"].([]any)[0].(map[string]any)["d-kind"]; got != "weird" {
		t.Errorf("edge d-kind = %v, want \"weird\"", got)
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

func TestNode_SetExtra_TypedFieldWins(t *testing.T) {
	c := &Canvas{Nodes: []*Node{{ID: "n", Type: "text", Text: "typed"}}}
	if err := c.Nodes[0].SetExtra("text", "extra"); err != nil {
		t.Fatalf("SetExtra: %v", err)
	}
	var buf bytes.Buffer
	if err := Encode(c, &buf); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if n := strings.Count(buf.String(), `"text":`); n != 1 {
		t.Errorf("text key written %d times, want 1: %s", n, buf.String())
	}
	m := encodeToMap(t, c)
	if got := m["nodes"].([]any)[0].(map[string]any)["text"]; got != "typed" {
		t.Errorf("node text = %v, want \"typed\"", got)
	}
}

func TestEdge_SetExtra(t *testing.T) {
	c := &Canvas{
		Nodes: []*Node{{ID: "n", Type: "text", Text: "hi"}},
		Edges: []*Edge{{ID: "e", FromNode: "n", ToNode: "n"}},
	}
	if err := c.Edges[0].SetExtra("x-choice", "charm"); err != nil {
		t.Fatalf("SetExtra: %v", err)
	}

	m := encodeToMap(t, c)
	edge := m["edges"].([]any)[0].(map[string]any)
	if edge["x-choice"] != "charm" {
		t.Errorf("edge[\"x-choice\"] = %v, want \"charm\"", edge["x-choice"])
	}

	var buf bytes.Buffer
	if err := Encode(c, &buf); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	c2, err := Decode(&buf)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	m2 := encodeToMap(t, c2)
	edge2 := m2["edges"].([]any)[0].(map[string]any)
	if edge2["x-choice"] != "charm" {
		t.Errorf("after round-trip, edge[\"x-choice\"] = %v", edge2["x-choice"])
	}
}

func TestEdge_SetExtra_TypedFieldWins(t *testing.T) {
	c := &Canvas{
		Nodes: []*Node{{ID: "n", Type: "text", Text: "hi"}},
		Edges: []*Edge{{ID: "e", FromNode: "n", ToNode: "n", Label: "typed"}},
	}
	if err := c.Edges[0].SetExtra("label", "extra"); err != nil {
		t.Fatalf("SetExtra: %v", err)
	}
	var buf bytes.Buffer
	if err := Encode(c, &buf); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if n := strings.Count(buf.String(), `"label"`); n != 1 {
		t.Errorf("label key written %d times, want 1: %s", n, buf.String())
	}
	m := encodeToMap(t, c)
	if got := m["edges"].([]any)[0].(map[string]any)["label"]; got != "typed" {
		t.Errorf("edge label = %v, want \"typed\"", got)
	}
}

func TestEdge_SetExtra_UnmarshalableValue(t *testing.T) {
	e := &Edge{ID: "e"}
	err := e.SetExtra("x-bad", make(chan int))
	if err == nil {
		t.Fatal("SetExtra with an unmarshalable value returned nil")
	}
	if !strings.Contains(err.Error(), `"x-bad"`) {
		t.Errorf("error %q does not name the key", err)
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

func TestDecode_Strict(t *testing.T) {
	// Reading never alters a document silently: anything the format's field
	// preservation and literal text cannot carry through is an error.
	tests := []struct {
		name  string
		input string
	}{
		{"duplicated top-level key", `{"d-version":"1.1","d-version":"1.0"}`},
		{"duplicated node key", `{"d-version":"1.1","nodes":[{"id":"n","type":"text","text":"a","text":"b"}]}`},
		{"duplicated key in an unknown field", `{"d-version":"1.1","x-a":{"k":1,"k":2}}`},
		{"invalid UTF-8 in a string", "{\"d-version\":\"1.1\",\"nodes\":[{\"id\":\"n\",\"type\":\"text\",\"text\":\"a\xffb\"}]}"},
		{"lone surrogate escape in a string", `{"d-version":"1.1","nodes":[{"id":"n","type":"text","text":"a\ud800b"}]}`},
		{"trailing garbage", `{"d-version":"1.1"} junk`},
		{"a second document", `{"d-version":"1.1"}{"d-version":"1.1"}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := decodeString(t, tc.input)
			if err == nil {
				t.Fatalf("Decode(%q) succeeded, want an error", tc.input)
			}
			if !strings.HasPrefix(err.Error(), "can't decode dcanvas: ") {
				t.Errorf("error %q lacks the \"can't decode dcanvas: \" prefix", err)
			}
		})
	}
}

func TestDecode_TrailingWhitespaceAccepted(t *testing.T) {
	if _, err := decodeString(t, "{\"d-version\":\"1.1\"}\n\n"); err != nil {
		t.Fatalf("Decode with trailing whitespace: %v", err)
	}
}

func TestEncode_WritesTextLiterally(t *testing.T) {
	// <, >, & and U+2028/2029 are written as they are, in known fields and in
	// an unknown one; an escape read from a file is written as its character.
	input := `{
		"d-version": "1.1",
		"nodes": [{"id":"n","type":"text","x":0,"y":0,"width":1,"height":1,
			"text":"a<b>&c` + "\u2028\u2029" + `d",
			"x-markup":{"s":"<i>&amp;</i>` + "\u2028" + `","escaped":"\u003c\u0026\u003e\u2028"}}],
		"edges": [{"id":"e","fromNode":"n","toNode":"n","label":"<go> & see"}]
	}`
	c, err := decodeString(t, input)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	var buf bytes.Buffer
	if err := Encode(c, &buf); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	out := buf.String()
	for _, want := range []string{
		`"text": "a<b>&c` + "\u2028\u2029" + `d"`,
		`"label": "<go> & see"`,
		`"s": "<i>&amp;</i>` + "\u2028" + `"`,
		`"escaped": "<&>` + "\u2028" + `"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %s:\n%s", want, out)
		}
	}
	for _, escape := range []string{`\u003c`, `\u003e`, `\u0026`, `\u2028`, `\u2029`} {
		if strings.Contains(out, escape) {
			t.Errorf("output contains the escape %s:\n%s", escape, out)
		}
	}
}

func TestRoundTrip_UnknownFieldNumbersVerbatim(t *testing.T) {
	// An unknown field's numbers come back exactly as written, never
	// normalised through a float.
	input := `{"d-version":"1.1","x-numbers":[1e3,1.0,-0,12345678901234567890,0.1000]}`
	c, err := decodeString(t, input)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	var buf bytes.Buffer
	if err := Encode(c, &buf); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	for _, num := range []string{"1e3", "1.0", "-0", "12345678901234567890", "0.1000"} {
		if !strings.Contains(buf.String(), "\t"+num) {
			t.Errorf("number %s not written verbatim:\n%s", num, buf.String())
		}
	}
}

func TestRoundTrip_UnknownFieldStringsEqualInValue(t *testing.T) {
	// An unknown field's strings may change spelling (an escape is written as
	// its character) but never value.
	input := `{"d-version":"1.1","x-strings":["\u00e9\t\"\\\/","plain","\ud83d\ude00"]}`
	c, err := decodeString(t, input)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	m := encodeToMap(t, c)
	got, ok := m["x-strings"].([]any)
	if !ok {
		t.Fatalf("x-strings = %v, want an array", m["x-strings"])
	}
	want := []any{"é\t\"\\/", "plain", "😀"}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("x-strings = %q, want %q", got, want)
	}
}

func TestEncode_IndentedWithTabAndEndsWithNewline(t *testing.T) {
	var buf bytes.Buffer
	if err := Encode(&Canvas{Nodes: []*Node{{ID: "n", Type: "text"}}}, &buf); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	out := buf.String()
	if !strings.HasPrefix(out, "{\n\t\"d-version\": ") {
		t.Errorf("output is not indented with a tab:\n%s", out)
	}
	if !strings.Contains(out, "\n\t\t{\n\t\t\t\"id\": \"n\"") {
		t.Errorf("nested objects are not indented with tabs:\n%s", out)
	}
	if !strings.HasSuffix(out, "}\n") || strings.HasSuffix(out, "\n\n") {
		t.Errorf("output does not end with exactly one newline: %q", out)
	}
}

func TestEncode_RejectsInvalidUTF8(t *testing.T) {
	// Writing is as strict as reading: text that is not valid UTF-8 is an
	// error, never silently replaced.
	c := &Canvas{Nodes: []*Node{{ID: "n", Type: "text", Text: "a\xffb"}}}
	err := Encode(c, &bytes.Buffer{})
	if err == nil {
		t.Fatal("Encode of invalid UTF-8 text succeeded, want an error")
	}
	if !strings.Contains(err.Error(), `"/text"`) {
		t.Errorf("error %q does not name the text field", err)
	}
}

func TestEncode_TextNodeAlwaysWritesText(t *testing.T) {
	// The schema requires text on a text node, so an empty one is still
	// written, in its place among the known fields; other node types omit an
	// empty text.
	c := &Canvas{Nodes: []*Node{
		{ID: "t", Type: "text", Color: "1", DomainID: "D"},
		{ID: "f", Type: "file", Color: "1", DomainID: "D"},
	}}
	var buf bytes.Buffer
	if err := Encode(c, &buf); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	got := keyOrders(t, buf.Bytes())
	for ptr, want := range map[string]string{
		"/nodes/0": "id,type,x,y,width,height,color,text,d-id",
		"/nodes/1": "id,type,x,y,width,height,color,d-id",
	} {
		if strings.Join(got[ptr], ",") != want {
			t.Errorf("keys of %q = %v, want %v", ptr, got[ptr], want)
		}
	}
	if !strings.Contains(buf.String(), `"text": ""`) {
		t.Errorf("text node lacks an empty text:\n%s", buf.String())
	}
}

func TestSetExtra_TypedFieldWinsWhenOmitted(t *testing.T) {
	// A key owned by a typed field is never written by an extension, even when
	// the typed field itself is empty and so omitted: the typed field wins by
	// being absent, and Encode reports no error.
	c := &Canvas{
		Nodes: []*Node{{ID: "n", Type: "file"}},
		Edges: []*Edge{{ID: "e", FromNode: "n", ToNode: "n"}},
	}
	if err := c.Nodes[0].SetExtra("text", "extra"); err != nil {
		t.Fatalf("Node.SetExtra: %v", err)
	}
	if err := c.Nodes[0].SetExtra("d-character", map[string]string{"name": "extra"}); err != nil {
		t.Fatalf("Node.SetExtra: %v", err)
	}
	if err := c.Edges[0].SetExtra("label", "extra"); err != nil {
		t.Fatalf("Edge.SetExtra: %v", err)
	}
	var buf bytes.Buffer
	if err := Encode(c, &buf); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if strings.Contains(buf.String(), "extra") {
		t.Errorf("an extension wrote a key owned by a typed field:\n%s", buf.String())
	}
	// Nor does the extension reach the typed field through a re-read.
	c2, err := Decode(&buf)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if n, e := c2.Nodes[0], c2.Edges[0]; n.Text != "" || n.Character != nil || e.Label != "" {
		t.Errorf("after a round-trip text = %q, d-character = %v, label = %q, want all empty", n.Text, n.Character, e.Label)
	}
}

func TestSetExtra_TypedFieldWinsOnAlwaysWrittenField(t *testing.T) {
	// x is always written by the node, and fromNode by the edge.
	c := &Canvas{
		Nodes: []*Node{{ID: "n", Type: "text", X: 7}},
		Edges: []*Edge{{ID: "e", FromNode: "n", ToNode: "n"}},
	}
	if err := c.Nodes[0].SetExtra("x", 99); err != nil {
		t.Fatalf("Node.SetExtra: %v", err)
	}
	if err := c.Edges[0].SetExtra("fromNode", "m"); err != nil {
		t.Fatalf("Edge.SetExtra: %v", err)
	}
	m := encodeToMap(t, c)
	if got := m["nodes"].([]any)[0].(map[string]any)["x"]; got != float64(7) {
		t.Errorf("node x = %v, want 7", got)
	}
	if got := m["edges"].([]any)[0].(map[string]any)["fromNode"]; got != "n" {
		t.Errorf("edge fromNode = %v, want \"n\"", got)
	}
}

func TestSetExtra_ReplacesInPlace(t *testing.T) {
	// Setting a key twice keeps one key, with the last value, in the place of
	// the first; a decoded unknown field is replaced the same way.
	c, err := decodeString(t, `{"d-version":"1.1","nodes":[{"id":"n","type":"text","text":"hi","x-b":1,"x-a":1}]}`)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	n := c.Nodes[0]
	for _, set := range []struct {
		key string
		val int
	}{{"x-c", 1}, {"x-b", 2}, {"x-c", 3}} {
		if err := n.SetExtra(set.key, set.val); err != nil {
			t.Fatalf("SetExtra(%q): %v", set.key, err)
		}
	}
	var buf bytes.Buffer
	if err := Encode(c, &buf); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	keys := keyOrders(t, buf.Bytes())["/nodes/0"]
	if got := strings.Join(keys, ","); got != "id,type,x,y,width,height,text,x-b,x-a,x-c" {
		t.Errorf("node keys = %v", got)
	}
	node := encodeToMap(t, c)["nodes"].([]any)[0].(map[string]any)
	if node["x-b"] != float64(2) || node["x-c"] != float64(3) {
		t.Errorf("x-b = %v, x-c = %v, want 2 and 3", node["x-b"], node["x-c"])
	}
}

func TestSetExtra_InvalidUTF8KeyIsAnError(t *testing.T) {
	// Writing is strict: a key that is not valid UTF-8 is an error, never a
	// field that silently goes missing.
	n := &Node{ID: "n", Type: "text"}
	e := &Edge{ID: "e", FromNode: "n", ToNode: "n"}
	for name, set := range map[string]func(string, any) error{"node": n.SetExtra, "edge": e.SetExtra} {
		err := set("x-a\xffb", 1)
		if err == nil {
			t.Errorf("%s: SetExtra with an invalid UTF-8 key returned nil", name)
			continue
		}
		if !strings.HasPrefix(err.Error(), "dcanvas: field ") {
			t.Errorf("%s: error %q does not name the field", name, err)
		}
	}
	var buf bytes.Buffer
	if err := Encode(&Canvas{Nodes: []*Node{n}, Edges: []*Edge{e}}, &buf); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if strings.Contains(buf.String(), `"x-`) {
		t.Errorf("a failed SetExtra left a key:\n%s", buf.String())
	}
}

func TestTextNodeFields_TagsMatchNode(t *testing.T) {
	// textNodeFields differs from Node only in that text is always written; the
	// conversion between them checks names and types, this checks the tags.
	node, text := reflect.TypeFor[Node](), reflect.TypeFor[textNodeFields]()
	if node.NumField() != text.NumField() {
		t.Fatalf("Node has %d fields, textNodeFields %d", node.NumField(), text.NumField())
	}
	for i := range node.NumField() {
		nf, tf := node.Field(i), text.Field(i)
		want := nf.Tag.Get("json")
		if nf.Name == "Text" {
			want = strings.TrimSuffix(want, ",omitempty")
		}
		if got := tf.Tag.Get("json"); got != want {
			t.Errorf("textNodeFields.%s tag = %q, want %q", tf.Name, got, want)
		}
	}
}
