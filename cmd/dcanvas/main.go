// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: BlueOak-1.0.0

// Command dcanvas is a small CLI over the dcanvas library. Its first (and so
// far only) subcommand, layout, reads a .d.canvas file, validates it, runs
// auto-layout with the chosen strategy, and writes the laid-out canvas back —
// the read → operate → write tracer bullet. Fields the tool does not touch
// (including Layer 2 project fields) survive via the library's preservation.
package main

import "os"

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
