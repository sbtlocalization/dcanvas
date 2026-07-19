// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: BlueOak-1.0.0

package dcanvas

import (
	"strings"
	"testing"
)

// wellFormed returns a canvas that passes every Validate check, so a test can
// perturb a single aspect and confirm that aspect (and only it) is reported.
func wellFormed() *Canvas {
	return &Canvas{
		Nodes: []*Node{
			{ID: "n1", Type: "text", Width: 400, Height: 300, Text: "Greetings.", Kind: KindLine,
				Character: &Character{Name: "Abela"}},
			{ID: "n2", Type: "text", Width: 400, Height: 300, Text: "Hello.", Kind: KindReply},
			{ID: "g1", Type: "group"}, // non-text node needs no text
		},
		Edges: []*Edge{
			{ID: "e1", FromNode: "n1", ToNode: "n2", Label: "reply", Kind: KindNormal},
		},
	}
}

func TestValidate_HappyPath(t *testing.T) {
	if err := Validate(wellFormed()); err != nil {
		t.Fatalf("well-formed canvas should validate, got: %v", err)
	}
}

func TestValidate_Failures(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Canvas)
		wantSub []string // substrings the error must name so a user can locate it
	}{
		{
			name:    "duplicate node id",
			mutate:  func(c *Canvas) { c.Nodes[1].ID = "n1" },
			wantSub: []string{"duplicate", "n1"},
		},
		{
			name:    "edge references missing fromNode",
			mutate:  func(c *Canvas) { c.Edges[0].FromNode = "ghost" },
			wantSub: []string{"e1", "ghost"},
		},
		{
			name:    "edge references missing toNode",
			mutate:  func(c *Canvas) { c.Edges[0].ToNode = "ghost" },
			wantSub: []string{"e1", "ghost"},
		},
		{
			name:    "out-of-set node d-kind",
			mutate:  func(c *Canvas) { c.Nodes[0].Kind = "shout" },
			wantSub: []string{"n1", "shout"},
		},
		{
			name:    "out-of-set edge d-kind",
			mutate:  func(c *Canvas) { c.Edges[0].Kind = "weird" },
			wantSub: []string{"e1", "weird"},
		},
		{
			name:    "d-character without name",
			mutate:  func(c *Canvas) { c.Nodes[0].Character.Name = "" },
			wantSub: []string{"n1", "name"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := wellFormed()
			tc.mutate(c)
			err := Validate(c)
			if err == nil {
				t.Fatalf("expected an error, got nil")
			}
			msg := err.Error()
			for _, sub := range tc.wantSub {
				if !strings.Contains(msg, sub) {
					t.Errorf("error %q does not name %q", msg, sub)
				}
			}
		})
	}
}

// TestValidate_EmptyTextIsValid pins the rule that an empty text on a text node
// is valid: the format requires the text field to be present, not non-empty. A
// terminal "end of dialogue" node legitimately carries no spoken line.
func TestValidate_EmptyTextIsValid(t *testing.T) {
	c := wellFormed()
	c.Nodes = append(c.Nodes, &Node{
		ID: "end", Type: "text", Width: 400, Height: 300, Text: "",
		Character: &Character{Name: "End dialog"},
	})
	if err := Validate(c); err != nil {
		t.Fatalf("a text node with empty text should validate, got: %v", err)
	}
}

func TestValidate_ReportsMultipleProblems(t *testing.T) {
	c := wellFormed()
	c.Nodes[1].ID = "n1"      // duplicate id
	c.Edges[0].Kind = "weird" // bad edge kind
	err := Validate(c)
	if err == nil {
		t.Fatal("expected an error for a canvas with multiple problems")
	}
	msg := err.Error()
	if !strings.Contains(msg, "duplicate") || !strings.Contains(msg, "weird") {
		t.Errorf("expected both problems reported, got: %v", msg)
	}
}
