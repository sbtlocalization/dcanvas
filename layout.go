// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: BlueOak-1.0.0

package dcanvas

import (
	"github.com/nulab/autog"
	"github.com/nulab/autog/graph"
)

// LoopStrategy controls how loop (cyclic) edges participate in layout.
type LoopStrategy int

const (
	// LoopCut excludes edges marked x-kind:loop from the layout graph, so the
	// remaining (acyclic) graph lays out as a clean top-down dialogue tree with
	// the entry node on top. This is the default and matches how the original
	// sbt-infinity export positioned nodes. Loop edges still appear in the
	// encoded canvas; they are only ignored while computing positions.
	LoopCut LoopStrategy = iota

	// LoopDFS keeps every edge in the layout graph and breaks cycles with a
	// depth-first edge reversal. Loop edges then influence positions (and are
	// routed back up through the layers), which can push their endpoints aside.
	LoopDFS
)

// LayoutOption configures Layout.
type LayoutOption func(*layoutConfig)

type layoutConfig struct {
	loops LoopStrategy
}

// WithLoopStrategy selects how loop (cyclic) edges are handled during layout.
// The default is LoopCut.
func WithLoopStrategy(s LoopStrategy) LayoutOption {
	return func(cfg *layoutConfig) { cfg.loops = s }
}

// Layout assigns positions to the canvas's nodes with a generic layered
// auto-layout. It reads node Width/Height and edge FromNode/ToNode, and writes
// back only each node's X and Y.
//
// By default (LoopCut) it reads the Layer 1 field x-kind to drop loop edges
// from the layout graph, producing a clean top-down tree; pass
// WithLoopStrategy(LoopDFS) to instead lay out every edge and break cycles
// depth-first. Either way, cycles never crash layout.
//
// Layout is a best-effort, optional operation: a canvas can be encoded with
// hand-set positions without ever calling it. An empty graph, or one with no
// layout edges (e.g. every edge cut as a loop), is a no-op that leaves
// positions untouched; if the underlying engine panics, the panic is recovered
// and nodes keep their existing positions. Nodes not referenced by any layout
// edge are left where they are (autog only places connected nodes).
func Layout(c *Canvas, opts ...LayoutOption) {
	cfg := layoutConfig{loops: LoopCut}
	for _, o := range opts {
		o(&cfg)
	}

	if len(c.Nodes) == 0 {
		return
	}

	// Index nodes by id so results can be written back, and build per-node
	// sizes for the layout engine.
	nodesByID := make(map[string]*Node, len(c.Nodes))
	sizes := make(map[string]graph.Size, len(c.Nodes))
	for _, n := range c.Nodes {
		nodesByID[n.ID] = n
		sizes[n.ID] = graph.Size{W: float64(n.Width), H: float64(n.Height)}
	}

	// Build the layout edge set, dropping x-kind:loop edges under LoopCut.
	layoutEdges := make([][]string, 0, len(c.Edges))
	for _, e := range c.Edges {
		if cfg.loops == LoopCut && e.Kind == KindLoop {
			continue
		}
		layoutEdges = append(layoutEdges, []string{e.FromNode, e.ToNode})
	}
	if len(layoutEdges) == 0 {
		return // nothing to lay out; leave positions as-is
	}

	// autog can panic on malformed input; never let that escape Layout.
	defer func() { _ = recover() }()

	layout := autog.Layout(
		graph.EdgeSlice(layoutEdges),
		autog.WithNodeSize(sizes),
		autog.WithLayerSpacing(200),
		// Depth-first cycle breaking reverses true back-edges (found by a DFS
		// from source nodes), preserving the natural top-down hierarchy. It is
		// a safety net for any residual cycle after LoopCut, and the cycle
		// resolver for LoopDFS; the default greedy breaker tangles dialogues.
		autog.WithCycleBreaking(autog.CycleBreakingDepthFirst),
	)
	for _, n := range layout.Nodes {
		if cn, ok := nodesByID[n.ID]; ok {
			cn.X = int(n.X)
			cn.Y = int(n.Y)
		}
	}
}
