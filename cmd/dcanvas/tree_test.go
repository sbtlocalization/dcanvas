// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: BlueOak-1.0.0

package main

import (
	"os"
	"testing"

	"github.com/sbtlocalization/dcanvas"
)

// TestBuildTree is the one sanity check on the pure tree logic (the TUI itself
// is deliberately untested — no teatest, no Update-transition tests). It builds
// the forest from the example dialogue fixture and asserts the two shapes that
// matter: the root is n-greet, and both loop edges surface as jump lines.
func TestBuildTree(t *testing.T) {
	f, err := os.Open("testdata/example.d.canvas")
	if err != nil {
		t.Fatalf("open example: %v", err)
	}
	defer f.Close()

	c, err := dcanvas.Decode(f)
	if err != nil {
		t.Fatalf("decode example: %v", err)
	}

	forest := build(c)
	if len(forest) != 1 {
		t.Fatalf("want a single-root forest, got %d roots", len(forest))
	}
	if got := forest[0].node.ID; got != "n-greet" {
		t.Errorf("root = %q, want %q", got, "n-greet")
	}

	// Both loop edges (e8, e9 back to n-greet) must render as jump lines.
	var loopJumps int
	var walk func(it *item)
	walk = func(it *item) {
		if it.jump && it.edge != nil && it.edge.Kind == dcanvas.KindLoop {
			loopJumps++
		}
		for _, c := range it.children {
			walk(c)
		}
	}
	for _, r := range forest {
		walk(r)
	}
	if loopJumps != 2 {
		t.Errorf("loop jump lines = %d, want 2", loopJumps)
	}
}
