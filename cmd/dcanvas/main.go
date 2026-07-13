// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: BlueOak-1.0.0

// Command dcanvas is a small CLI over the dcanvas library. Its first (and so
// far only) subcommand, layout, reads a .d.canvas file, runs auto-layout with
// the default strategy, and writes the laid-out canvas back — the read →
// operate → write tracer bullet. Fields the tool does not touch (including
// Layer 2 project fields) survive via the library's preservation.
package main

import (
	"fmt"
	"os"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "dcanvas:", err)
		os.Exit(1)
	}
}

const usage = "usage: dcanvas layout [-o output] <input>"

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("%s", usage)
	}
	switch args[0] {
	case "layout":
		return runLayout(args[1:])
	default:
		return fmt.Errorf("unknown command %q\n%s", args[0], usage)
	}
}
