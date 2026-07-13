// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: BlueOak-1.0.0

package dcanvas

import "testing"

// characterOf decodes a single-node canvas carrying the given x-character JSON
// and returns the node's re-encoded x-character map, so assertions are on the
// externally observable JSON rather than the internal catch-all.
func characterOf(t *testing.T, characterJSON string) map[string]any {
	t.Helper()
	input := `{
		"x-dCanvasVersion": "3.0",
		"nodes": [
			{
				"id": "n1", "type": "text",
				"x": 0, "y": 0, "width": 400, "height": 300,
				"text": "hi",
				"x-character": ` + characterJSON + `
			}
		],
		"edges": []
	}`
	c, err := decodeString(t, input)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	m := encodeToMap(t, c)
	node := m["nodes"].([]any)[0].(map[string]any)
	ch, ok := node["x-character"].(map[string]any)
	if !ok {
		t.Fatalf("x-character missing or not an object: %v", node["x-character"])
	}
	return ch
}

func TestCharacter_TypedFields(t *testing.T) {
	tests := []struct {
		name          string
		characterJSON string
		want          map[string]any
		absent        []string
	}{
		{
			name:          "all known fields",
			characterJSON: `{"name": "Abela", "portrait": "abela.png", "gender": "female"}`,
			want:          map[string]any{"name": "Abela", "portrait": "abela.png", "gender": "female"},
		},
		{
			name:          "name only, omitempty drops the rest",
			characterJSON: `{"name": "Narrator"}`,
			want:          map[string]any{"name": "Narrator"},
			absent:        []string{"portrait", "gender"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ch := characterOf(t, tc.characterJSON)
			for k, want := range tc.want {
				if ch[k] != want {
					t.Errorf("x-character[%q] = %v, want %v", k, ch[k], want)
				}
			}
			for _, k := range tc.absent {
				if _, ok := ch[k]; ok {
					t.Errorf("x-character[%q] should be absent (omitempty), got %v", k, ch[k])
				}
			}
		})
	}
}

func TestCharacter_PreservesUnknownScalar(t *testing.T) {
	ch := characterOf(t, `{"name": "Abela", "mood": "angry"}`)
	if ch["name"] != "Abela" {
		t.Errorf("typed field lost: name = %v", ch["name"])
	}
	if ch["mood"] != "angry" {
		t.Errorf("unknown nested scalar dropped: mood = %v", ch["mood"])
	}
}

func TestCharacter_PreservesUnknownNested(t *testing.T) {
	ch := characterOf(t, `{
		"name": "Abela",
		"x-stats": {"hp": 10, "tags": ["nymph", "fae"]},
		"x-aliases": ["Abby", "The Nymph"]
	}`)

	if ch["name"] != "Abela" {
		t.Errorf("typed field lost: name = %v", ch["name"])
	}

	stats, ok := ch["x-stats"].(map[string]any)
	if !ok {
		t.Fatalf("unknown nested object dropped: x-stats = %v", ch["x-stats"])
	}
	if stats["hp"].(float64) != 10 {
		t.Errorf("nested object value lost: x-stats.hp = %v", stats["hp"])
	}
	tags, ok := stats["tags"].([]any)
	if !ok || len(tags) != 2 || tags[0] != "nymph" || tags[1] != "fae" {
		t.Errorf("nested array within object lost: x-stats.tags = %v", stats["tags"])
	}

	aliases, ok := ch["x-aliases"].([]any)
	if !ok || len(aliases) != 2 || aliases[0] != "Abby" || aliases[1] != "The Nymph" {
		t.Errorf("unknown nested array dropped: x-aliases = %v", ch["x-aliases"])
	}
}
