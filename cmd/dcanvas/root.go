// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: BlueOak-1.0.0

package main

import "github.com/spf13/cobra"

// newRootCmd builds the root dcanvas command and registers its subcommands. A
// fresh instance is returned each call so tests can execute the CLI in
// isolation.
func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "dcanvas",
		Short:         "Operate on dCanvas dialogue-graph files",
		Long:          "dcanvas is a CLI over the dcanvas library for reading, laying out, and writing .d.canvas files.",
		SilenceUsage:  true, // a runtime failure should not dump usage
		SilenceErrors: false,
	}
	root.AddCommand(newLayoutCmd())
	root.AddCommand(newViewCmd())
	return root
}
