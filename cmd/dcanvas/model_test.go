// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: BlueOak-1.0.0

package main

import (
	"strings"
	"testing"

	"github.com/sbtlocalization/dcanvas"
)

// TestRenderCharacterFields verifies that renderCharacterFields outputs
// character attributes with the correct indentation.
func TestRenderCharacterFields(t *testing.T) {
	tests := []struct {
		name    string
		ch      *dcanvas.Character
		indent  string
		wantSub []string
		notWant []string
	}{
		{
			name:    "all fields present",
			ch:      &dcanvas.Character{Name: "Alice", TextID: "#alice", Portrait: "alice.png", Gender: "female"},
			indent:  "",
			wantSub: []string{"Текст ID", "#alice", "Портрет", "alice.png", "Стать", "female"},
			notWant: nil,
		},
		{
			name:    "only TextID",
			ch:      &dcanvas.Character{Name: "Bob", TextID: "#bob"},
			indent:  "",
			wantSub: []string{"Текст ID", "#bob"},
			notWant: []string{"Портрет", "Стать"},
		},
		{
			name:    "no optional fields",
			ch:      &dcanvas.Character{Name: "Charlie"},
			indent:  "",
			wantSub: nil,
			notWant: []string{"Текст ID", "Портрет", "Стать"},
		},
		{
			name:    "with indent",
			ch:      &dcanvas.Character{TextID: "#dave", Portrait: "dave.png"},
			indent:  "    ",
			wantSub: []string{"    Текст ID", "#dave", "    Портрет", "dave.png"},
			notWant: []string{"Стать"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b strings.Builder
			renderCharacterFields(&b, tt.ch, tt.indent)
			got := b.String()

			for _, want := range tt.wantSub {
				if !strings.Contains(got, want) {
					t.Errorf("got %q, want substring %q", got, want)
				}
			}

			for _, notWant := range tt.notWant {
				if strings.Contains(got, notWant) {
					t.Errorf("got %q, should not contain %q", got, notWant)
				}
			}
		})
	}
}

// TestRenderDetailWithCharacter verifies that renderDetail shows the main
// character's name and TextID when present.
func TestRenderDetailWithCharacter(t *testing.T) {
	m := model{
		lines: []line{
			{
				item: &item{
					node: &dcanvas.Node{
						ID:   "n1",
						Text: "Hello",
						Kind: dcanvas.KindLine,
						Character: &dcanvas.Character{
							Name:    "Alice",
							TextID:  "#alice",
							Portrait: "alice.png",
						},
					},
				},
			},
		},
		cursor: 0,
	}

	got := m.renderDetail(80, 20)

	// Verify main character is shown
	if !strings.Contains(got, "Мовець") {
		t.Error("detail missing 'Мовець' (Speaker) label")
	}
	if !strings.Contains(got, "Alice") {
		t.Error("detail missing character name 'Alice'")
	}

	// Verify TextID is shown
	if !strings.Contains(got, "Текст ID") {
		t.Error("detail missing 'Текст ID' label")
	}
	if !strings.Contains(got, "#alice") {
		t.Error("detail missing character TextID '#alice'")
	}

	// Verify other fields are shown
	if !strings.Contains(got, "Портрет") {
		t.Error("detail missing 'Портрет' (Portrait) label")
	}
	if !strings.Contains(got, "alice.png") {
		t.Error("detail missing portrait 'alice.png'")
	}
}

// TestRenderDetailWithAlternativeCharacters verifies that renderDetail lists
// alternative speakers with their attributes.
func TestRenderDetailWithAlternativeCharacters(t *testing.T) {
	m := model{
		lines: []line{
			{
				item: &item{
					node: &dcanvas.Node{
						ID:   "n1",
						Text: "Can you help?",
						Kind: dcanvas.KindLine,
						Character: &dcanvas.Character{
							Name:   "Guard",
							TextID: "#guard",
						},
						AlternativeCharacters: []*dcanvas.Character{
							{Name: "Captain", TextID: "#captain", Portrait: "captain.png"},
							{Name: "Soldier", Gender: "male"},
						},
					},
				},
			},
		},
		cursor: 0,
	}

	got := m.renderDetail(80, 30)

	// Verify main character
	if !strings.Contains(got, "Guard") {
		t.Error("detail missing main character 'Guard'")
	}
	if !strings.Contains(got, "#guard") {
		t.Error("detail missing main character TextID '#guard'")
	}

	// Verify alternative speakers section header
	if !strings.Contains(got, "Альтернативні мовці") {
		t.Error("detail missing 'Альтернативні мовці' (Alternative speakers) section")
	}

	// Verify first alternative
	if !strings.Contains(got, "Captain") {
		t.Error("detail missing alternative speaker 'Captain'")
	}
	if !strings.Contains(got, "#captain") {
		t.Error("detail missing alternative speaker TextID '#captain'")
	}
	if !strings.Contains(got, "captain.png") {
		t.Error("detail missing alternative speaker portrait 'captain.png'")
	}

	// Verify second alternative
	if !strings.Contains(got, "Soldier") {
		t.Error("detail missing alternative speaker 'Soldier'")
	}
	if !strings.Contains(got, "male") {
		t.Error("detail missing alternative speaker gender 'male'")
	}
}

// TestRenderLineDoesNotShowAlternativeCharacters verifies that the tree view
// (renderLine) only shows the main character name, not alternatives.
func TestRenderLineDoesNotShowAlternativeCharacters(t *testing.T) {
	m := model{}
	l := line{
		item: &item{
			depth: 0,
			node: &dcanvas.Node{
				ID:   "n1",
				Text: "What's up?",
				Kind: dcanvas.KindLine,
				Character: &dcanvas.Character{
					Name: "Guard",
				},
				AlternativeCharacters: []*dcanvas.Character{
					{Name: "Captain"},
					{Name: "Soldier"},
				},
			},
		},
		expandable: false,
		expanded:   false,
	}

	got := m.renderLine(l, 80, false)

	// Verify main character is shown
	if !strings.Contains(got, "Guard") {
		t.Error("tree line missing main character 'Guard'")
	}

	// Verify alternatives are NOT shown in the tree
	if strings.Contains(got, "Captain") {
		t.Error("tree line should not show alternative 'Captain'")
	}
	if strings.Contains(got, "Soldier") {
		t.Error("tree line should not show alternative 'Soldier'")
	}
}

// TestBuildDoesNotIncludeAlternativeCharactersInTree verifies that build()
// constructs the forest without considering AlternativeCharacters as separate
// nodes — they only appear in the detail view.
func TestBuildDoesNotIncludeAlternativeCharactersInTree(t *testing.T) {
	c := &dcanvas.Canvas{
		Version: "1.1",
		Nodes: []*dcanvas.Node{
			{
				ID:   "n1",
				Type: "text",
				Text: "Hello",
				Kind: dcanvas.KindLine,
				Character: &dcanvas.Character{
					Name: "Guard",
				},
				AlternativeCharacters: []*dcanvas.Character{
					{Name: "Captain"},
				},
			},
			{
				ID:   "n2",
				Type: "text",
				Text: "Goodbye",
				Kind: dcanvas.KindReply,
			},
		},
		Edges: []*dcanvas.Edge{
			{FromNode: "n1", ToNode: "n2", Kind: dcanvas.KindNormal},
		},
	}

	forest := build(c)

	// Should have exactly one root (n1)
	if len(forest) != 1 {
		t.Fatalf("want 1 root, got %d", len(forest))
	}

	// Root should be n1
	if forest[0].node.ID != "n1" {
		t.Errorf("root ID = %q, want n1", forest[0].node.ID)
	}

	// n1 should have exactly one child (n2)
	if len(forest[0].children) != 1 {
		t.Fatalf("n1 should have 1 child, got %d", len(forest[0].children))
	}

	// Child should be n2
	if forest[0].children[0].node.ID != "n2" {
		t.Errorf("child ID = %q, want n2", forest[0].children[0].node.ID)
	}

	// AlternativeCharacters should NOT create additional tree nodes
	if len(forest[0].children[0].children) != 0 {
		t.Errorf("n2 should have no children, got %d", len(forest[0].children[0].children))
	}
}
