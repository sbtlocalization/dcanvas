// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: BlueOak-1.0.0

package main

import "github.com/sbtlocalization/dcanvas"

// This file holds the view command's pure tree logic, deliberately kept apart
// from the Bubble Tea Model (model.go): build turns a canvas into a walkable
// forest, and visibleLines flattens that forest under a given expand state.
// Both are free of any UI concern so they can be reasoned about and tested on
// their own.

// item is a text node as it appears at one position in the walkable tree.
//
// The dialogue is a DAG (fan-in and loops), so a single node can be reached by
// several paths. It is rendered on its *first* occurrence (a canonical item
// with children); every later occurrence — a loop back-edge or a second forward
// parent (merge) — is a jump: a non-expandable reference back to that canonical
// item.
type item struct {
	node   *dcanvas.Node // the underlying text node (the jump target, for a jump)
	edge   *dcanvas.Edge // the edge that led here; nil for a root
	depth  int           // nesting depth from its root (roots are 0)
	parent *item         // nil for a root

	jump     bool    // true → a non-expandable reference to node's first occurrence
	target   *item   // a jump's canonical occurrence to move to (nil if unresolved)
	children []*item // populated only on a canonical (non-jump) occurrence
}

// expandable reports whether the item is a branch that can be opened or closed
// (a canonical occurrence with at least one child). Jumps and childless leaves
// are not expandable.
func (it *item) expandable() bool { return !it.jump && len(it.children) > 0 }

// build derives the walkable forest from a canvas, best-effort: it never runs
// Validate and silently ignores any edge whose endpoint node is missing or is
// not a text node. Only text nodes are visible; group/file/link nodes are
// opaque to the dialogue (ADR-0007) and never appear.
//
// Roots are text nodes with no incoming forward (normal) edge — loop edges do
// not count — taken in nodes-array order, so a forest is supported. Children
// come from a node's outgoing edges in edges-array order (authoring order,
// independent of geometry). If the whole graph is cyclic and no node qualifies
// as a root, the first text node is used so the forest is never empty.
func build(c *dcanvas.Canvas) []*item {
	text := make(map[string]*dcanvas.Node)
	for _, n := range c.Nodes {
		if n.Type == "text" {
			text[n.ID] = n
		}
	}

	// Outgoing edges per source in edges-array order, and a count of incoming
	// forward edges per node for root-finding. Only edges whose endpoints are
	// both visible text nodes are usable; the rest are silently dropped.
	outgoing := make(map[string][]*dcanvas.Edge)
	incomingForward := make(map[string]int)
	for _, e := range c.Edges {
		if text[e.FromNode] == nil || text[e.ToNode] == nil {
			continue
		}
		outgoing[e.FromNode] = append(outgoing[e.FromNode], e)
		if e.Kind != dcanvas.KindLoop {
			incomingForward[e.ToNode]++
		}
	}

	var rootIDs []string
	for _, n := range c.Nodes {
		if n.Type == "text" && incomingForward[n.ID] == 0 {
			rootIDs = append(rootIDs, n.ID)
		}
	}
	if len(rootIDs) == 0 { // fully cyclic: fall back to the first text node
		for _, n := range c.Nodes {
			if n.Type == "text" {
				rootIDs = append(rootIDs, n.ID)
				break
			}
		}
	}

	visited := make(map[string]bool)
	canonical := make(map[string]*item)
	var jumps []*item

	var walk func(id string, via *dcanvas.Edge, depth int, parent *item) *item
	walk = func(id string, via *dcanvas.Edge, depth int, parent *item) *item {
		it := &item{node: text[id], edge: via, depth: depth, parent: parent}
		isLoop := via != nil && via.Kind == dcanvas.KindLoop
		if isLoop || visited[id] {
			// A loop back-edge, or a node already expanded elsewhere (merge):
			// render it as a jump and resolve its target after the full walk,
			// once every canonical occurrence exists.
			it.jump = true
			jumps = append(jumps, it)
			return it
		}
		visited[id] = true
		canonical[id] = it
		for _, e := range outgoing[id] {
			it.children = append(it.children, walk(e.ToNode, e, depth+1, it))
		}
		return it
	}

	var forest []*item
	for _, id := range rootIDs {
		forest = append(forest, walk(id, nil, 0, nil))
	}
	for _, j := range jumps {
		j.target = canonical[j.node.ID]
	}
	return forest
}

// line is one visible row: an item plus its current expand state, so the view
// can draw the right indicator without re-deriving it.
type line struct {
	item       *item
	expandable bool
	expanded   bool
}

// visibleLines flattens the forest into the rows currently on screen, in
// pre-order. collapsed is the set of branch items the user has closed; an item
// absent from it is expanded, so the default (empty set) shows everything —
// the reading tool opens fully. A collapsed branch hides its whole subtree.
func visibleLines(forest []*item, collapsed map[*item]bool) []line {
	var out []line
	var walk func(it *item)
	walk = func(it *item) {
		exp := it.expandable()
		open := exp && !collapsed[it]
		out = append(out, line{item: it, expandable: exp, expanded: open})
		if open {
			for _, c := range it.children {
				walk(c)
			}
		}
	}
	for _, r := range forest {
		walk(r)
	}
	return out
}
