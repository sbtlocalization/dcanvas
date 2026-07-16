// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: BlueOak-1.0.0

// Command dcanvas is a small CLI over the dcanvas library. Its layout
// subcommand reads a .d.canvas file, validates it, runs auto-layout with the
// chosen strategy, and writes the laid-out canvas back — the read → operate →
// write tracer bullet; fields the tool does not touch (including Layer 2
// project fields) survive via the library's preservation. Its view subcommand
// opens a read-only TUI that browses the dialogue as a navigable tree; it never
// runs layout and never writes anything.
package main

import "os"

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
