// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: BlueOak-1.0.0

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sbtlocalization/dcanvas"
	"github.com/sbtlocalization/dcanvas/layout"
)

// runLayout reads the input canvas, lays it out with the default strategy
// (LoopCut), and writes it back. Without -o it overwrites the input in place.
func runLayout(args []string) error {
	fs := flag.NewFlagSet("layout", flag.ContinueOnError)
	out := fs.String("o", "", "output file (default: overwrite the input)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("layout: expected exactly one input file\n%s", usage)
	}

	in := fs.Arg(0)
	if err := checkExt(in); err != nil {
		return err
	}

	c, err := readCanvas(in)
	if err != nil {
		return err
	}

	if err := layout.Layout(c); err != nil {
		return fmt.Errorf("layout %s: %w", in, err)
	}

	dst := *out
	if dst == "" {
		dst = in
	}
	return writeCanvas(dst, c)
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
