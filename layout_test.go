// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: BlueOak-1.0.0

package dcanvas

import "testing"

// sized returns a node with an id and a bounding box; X/Y default to 0 so a
// post-layout move is observable.
func sized(id string, w, h int) *Node {
	return &Node{ID: id, Width: w, Height: h}
}

func edge(from, to string) *Edge {
	return &Edge{FromNode: from, ToNode: to}
}

func TestLayout_ConnectedGraphPositionsAndNoOverlap(t *testing.T) {
	c := &Canvas{
		Nodes: []*Node{
			sized("a", 400, 300),
			sized("b", 200, 150),
			sized("c", 400, 300),
			sized("d", 100, 100),
		},
		Edges: []*Edge{
			edge("a", "b"),
			edge("a", "c"),
			edge("b", "d"),
			edge("c", "d"),
		},
	}

	Layout(c)

	// At least one node must move off the default (0,0).
	moved := false
	for _, n := range c.Nodes {
		if n.X != 0 || n.Y != 0 {
			moved = true
			break
		}
	}
	if !moved {
		t.Error("expected layout to assign non-default coordinates")
	}

	if c.HasOverlappingNodes() {
		t.Error("expected laid-out nodes not to overlap")
	}
}

func TestLayout_CycleDoesNotPanic(t *testing.T) {
	// a -> b -> c -> a, with the back-edge marked as a loop.
	c := &Canvas{
		Nodes: []*Node{
			sized("a", 300, 200),
			sized("b", 300, 200),
			sized("c", 300, 200),
		},
		Edges: []*Edge{
			edge("a", "b"),
			edge("b", "c"),
			{FromNode: "c", ToNode: "a", Kind: KindLoop},
		},
	}

	// Must complete without panicking.
	Layout(c)

	if c.HasOverlappingNodes() {
		t.Error("expected laid-out nodes not to overlap for a cyclic graph")
	}

	// Depth-first cycle breaking should keep the entry node ("a", whose only
	// incoming edge is the back-edge) at the top rather than letting the loop
	// drag it below its successors.
	pos := map[string]int{}
	for _, n := range c.Nodes {
		pos[n.ID] = n.Y
	}
	if pos["a"] >= pos["b"] || pos["a"] >= pos["c"] {
		t.Errorf("entry node should be above its successors: a=%d b=%d c=%d", pos["a"], pos["b"], pos["c"])
	}
}

func TestLayout_LoopStrategy(t *testing.T) {
	// A canvas whose only edge is a loop: under LoopCut it is dropped (no layout
	// edges → no-op), under LoopDFS it lays the two nodes out.
	build := func() *Canvas {
		return &Canvas{
			Nodes: []*Node{sized("a", 400, 300), sized("b", 400, 300)},
			Edges: []*Edge{{FromNode: "a", ToNode: "b", Kind: KindLoop}},
		}
	}

	cut := build()
	Layout(cut) // default LoopCut
	for _, n := range cut.Nodes {
		if n.X != 0 || n.Y != 0 {
			t.Errorf("LoopCut should ignore the loop edge and leave %s at origin, got (%d,%d)", n.ID, n.X, n.Y)
		}
	}

	dfs := build()
	Layout(dfs, WithLoopStrategy(LoopDFS))
	moved := false
	for _, n := range dfs.Nodes {
		if n.X != 0 || n.Y != 0 {
			moved = true
		}
	}
	if !moved {
		t.Error("LoopDFS should lay out the loop edge and move at least one node")
	}
	if dfs.HasOverlappingNodes() {
		t.Error("LoopDFS layout should not overlap")
	}
}

func TestLayout_NoOp(t *testing.T) {
	tests := []struct {
		name   string
		canvas *Canvas
	}{
		{
			name:   "empty graph",
			canvas: &Canvas{},
		},
		{
			name: "nodes but no edges",
			canvas: &Canvas{Nodes: []*Node{
				sized("a", 400, 300),
				sized("b", 200, 150),
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture positions before layout.
			before := make([][2]int, len(tt.canvas.Nodes))
			for i, n := range tt.canvas.Nodes {
				before[i] = [2]int{n.X, n.Y}
			}

			Layout(tt.canvas)

			for i, n := range tt.canvas.Nodes {
				if n.X != before[i][0] || n.Y != before[i][1] {
					t.Errorf("node %d moved (%d,%d -> %d,%d); expected no-op",
						i, before[i][0], before[i][1], n.X, n.Y)
				}
			}
		})
	}
}
