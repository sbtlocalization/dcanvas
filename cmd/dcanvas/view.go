// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: BlueOak-1.0.0

package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// newViewCmd builds the `dcanvas view` command: a read-only TUI browser for a
// dialogue. Unlike layout it imports only the dcanvas package — it never runs
// layout and never writes anything.
func newViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "view <input>",
		Short: "Browse a dialogue as an interactive tree (read-only)",
		Long: `Open a read-only TUI that shows a .d.canvas (or .dcanvas) dialogue as a
navigable, collapsible tree of connected utterances.

view is for reading the dialogue — following who says what and how choices
branch — not for previewing geometry. It never runs layout and never writes to
disk. It requires an interactive terminal.`,
		Args: cobra.ExactArgs(1),
		RunE: runView,
	}
}

// runView reads the canvas best-effort (no Validate — a viewer is not a gate),
// requires a TTY, and hands the built model to Bubble Tea.
func runView(cmd *cobra.Command, args []string) error {
	in := args[0]
	if err := checkExt(in); err != nil {
		return err
	}
	c, err := readCanvas(in)
	if err != nil {
		return err
	}

	// A viewer needs stdin for keys and stdout for the screen; without a TTY on
	// both there is nothing to drive, so fail clearly rather than start Bubble
	// Tea against a pipe.
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return fmt.Errorf("view requires an interactive terminal")
	}

	_, err = tea.NewProgram(newModel(c), tea.WithAltScreen()).Run()
	return err
}
