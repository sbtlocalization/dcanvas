// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: BlueOak-1.0.0

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sbtlocalization/dcanvas"
	"github.com/sbtlocalization/dcanvas/layout"
)

// layoutFn is the layout entry point, indirected through a variable so a test
// can substitute a failing layout and exercise the CLI's failure-reporting path
// (the engine does not fail deterministically on any craftable input).
var layoutFn = layout.Layout

// newLayoutCmd builds the `dcanvas layout` command.
func newLayoutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "layout <input>",
		Short: "Auto-layout a canvas and write it back",
		Long: `Read a .d.canvas (or .dcanvas) file, validate it, assign node positions
with the chosen loop strategy, and write the result back.

Without -o the input file is overwritten in place. The input's version and any
fields the tool does not touch (including Layer 2 project fields) are preserved.`,
		Example: `  Lay out a file in place:
    dcanvas layout dialogue.d.canvas

  Write to a separate file, keeping loop edges in the graph:
    dcanvas layout --loop dfs -o out.d.canvas dialogue.d.canvas`,
		Args: cobra.ExactArgs(1),
		RunE: runLayout,
	}

	cmd.Flags().StringP("output", "o", "", "output file `path` (default: overwrite the input)")
	cmd.Flags().String("loop", "cut", "loop-edge strategy: cut or dfs")
	cmd.MarkFlagFilename("output", "d.canvas", "dcanvas")

	return cmd
}

// runLayout reads the input canvas, validates it, lays it out with the chosen
// strategy, and writes it back. It fails loudly and writes no output when the
// input is invalid or layout fails.
func runLayout(cmd *cobra.Command, args []string) error {
	output, _ := cmd.Flags().GetString("output")
	loop, _ := cmd.Flags().GetString("loop")

	strategy, err := loopStrategy(loop)
	if err != nil {
		return err
	}

	in := args[0]
	if err := checkExt(in); err != nil {
		return err
	}

	c, err := readCanvas(in)
	if err != nil {
		return err
	}

	// Validate before operating: refuse an unusable file and write nothing.
	if err := dcanvas.Validate(c); err != nil {
		return fmt.Errorf("invalid canvas %s:\n%w", in, err)
	}

	if err := layoutFn(c, layout.WithLoopStrategy(strategy)); err != nil {
		return fmt.Errorf("layout %s: %w", in, err)
	}

	dst := output
	if dst == "" {
		dst = in
	}
	return writeCanvas(dst, c)
}

// loopStrategy maps the --loop flag value to a layout strategy.
func loopStrategy(v string) (layout.LoopStrategy, error) {
	switch v {
	case "cut":
		return layout.LoopCut, nil
	case "dfs":
		return layout.LoopDFS, nil
	default:
		return layout.LoopCut, fmt.Errorf("unknown --loop %q (want \"cut\" or \"dfs\")", v)
	}
}

// acceptedExts are the format's file extensions: the default .d.canvas and the
// compact .dcanvas alternative.
var acceptedExts = []string{".d.canvas", ".dcanvas"}

// checkExt accepts only the format's file extensions.
func checkExt(path string) error {
	for _, ext := range acceptedExts {
		if strings.HasSuffix(path, ext) {
			return nil
		}
	}
	return fmt.Errorf("unsupported extension for %q (want %s)", filepath.Base(path), strings.Join(acceptedExts, " or "))
}

func readCanvas(path string) (*dcanvas.Canvas, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	c, err := dcanvas.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return c, nil
}

func writeCanvas(path string, c *dcanvas.Canvas) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := dcanvas.Encode(c, f); err != nil {
		f.Close()
		return fmt.Errorf("write %s: %w", path, err)
	}
	return f.Close()
}
