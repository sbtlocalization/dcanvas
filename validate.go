// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: BlueOak-1.0.0

package dcanvas

import (
	"errors"
	"fmt"
)

// Validate reports the structural and semantic problems a dCanvas document may
// have, so a consumer (e.g. the CLI) can gate an operation on a valid input and
// tell the user why a file is unusable. It does not mutate the canvas.
//
// It checks that:
//
//   - node ids are unique within the file;
//   - every edge fromNode/toNode references an existing node id;
//   - x-kind values stay within their closed sets (nodes: line/reply; edges:
//     normal/loop);
//   - required fields are present: name on x-character.
//
// It deliberately does not check that a text node's text is non-empty: neither
// JSON Canvas 1.0 nor dCanvas requires text to be non-empty (only present), and
// an empty string is a legitimate value — e.g. a terminal "end of dialogue"
// node carries no spoken line. Presence of the text field itself is a per-node
// invariant the JSON Schema already expresses and Encode already guarantees, so
// it is not re-checked here.
//
// All problems found are collected and returned as a single joined error, so a
// caller sees every issue in one pass; a well-formed canvas returns nil. This
// is a Go-side convenience distinct from — and complementary to — the
// JSON-Schema conformance guarantee: it verifies invariants (id uniqueness,
// edge referential integrity) that the schema, being per-object, cannot.
func Validate(c *Canvas) error {
	var problems []error

	// Node ids must be unique; collect the set for the edge reference check.
	known := make(map[string]bool, len(c.Nodes))
	for _, n := range c.Nodes {
		if known[n.ID] {
			problems = append(problems, fmt.Errorf("dcanvas: duplicate node id %q", n.ID))
		}
		known[n.ID] = true

		if !validKind(n.Kind, KindLine, KindReply) {
			problems = append(problems, fmt.Errorf(
				"dcanvas: node %q has invalid x-kind %q (want %q or %q)", n.ID, n.Kind, KindLine, KindReply))
		}
		if n.Character != nil && n.Character.Name == "" {
			problems = append(problems, fmt.Errorf(
				"dcanvas: node %q has an x-character without a required name", n.ID))
		}
	}

	// Every edge endpoint must reference an existing node id.
	for _, e := range c.Edges {
		if !known[e.FromNode] {
			problems = append(problems, fmt.Errorf(
				"dcanvas: edge %q references missing fromNode %q", e.ID, e.FromNode))
		}
		if !known[e.ToNode] {
			problems = append(problems, fmt.Errorf(
				"dcanvas: edge %q references missing toNode %q", e.ID, e.ToNode))
		}
		if !validKind(e.Kind, KindNormal, KindLoop) {
			problems = append(problems, fmt.Errorf(
				"dcanvas: edge %q has invalid x-kind %q (want %q or %q)", e.ID, e.Kind, KindNormal, KindLoop))
		}
	}

	return errors.Join(problems...)
}
