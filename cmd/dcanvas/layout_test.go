// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: BlueOak-1.0.0

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/sbtlocalization/dcanvas"
	"github.com/sbtlocalization/dcanvas/layout"
)

// update regenerates the golden file: `go test ./cmd/dcanvas/ -update`.
var update = flag.Bool("update", false, "update the golden output file")

const goldenPath = "testdata/golden.d.canvas"

// execute drives the CLI end-to-end: it builds a fresh root command, feeds it
// the given arguments, and returns the resulting error. Output is discarded so
// tests assert on behaviour (files, errors), not on printed text.
func execute(args ...string) error {
	root := newRootCmd()
	root.SetArgs(args)
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	return root.Execute()
}

// runLayoutToTemp copies the input testdata to a temp file with the given
// extension, runs the CLI's layout command against it, and returns the output
// bytes. Driving through execute() exercises the tool end-to-end.
func runLayoutToTemp(t *testing.T, inputName, ext string) []byte {
	t.Helper()
	src, err := os.ReadFile("testdata/input.d.canvas")
	if err != nil {
		t.Fatalf("read input fixture: %v", err)
	}
	dir := t.TempDir()
	inPath := filepath.Join(dir, inputName+ext)
	if err := os.WriteFile(inPath, src, 0o644); err != nil {
		t.Fatalf("write temp input: %v", err)
	}
	outPath := filepath.Join(dir, "out"+ext)
	if err := execute("layout", "-o", outPath, inPath); err != nil {
		t.Fatalf("run layout: %v", err)
	}
	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	return got
}

func TestLayoutCLI_Golden(t *testing.T) {
	got := runLayoutToTemp(t, "in", ".d.canvas")

	if *update {
		if err := os.WriteFile(goldenPath, got, 0o644); err != nil {
			t.Fatalf("update golden: %v", err)
		}
	}
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden (run with -update to create it): %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("CLI output does not match %s (run with -update to refresh)\n--- got ---\n%s", goldenPath, got)
	}
}

// anyNodeMoved reports whether any node in the encoded document sits off the
// (0,0) input origin — i.e. layout positioned it.
func anyNodeMoved(t *testing.T, doc map[string]any) bool {
	t.Helper()
	for _, raw := range doc["nodes"].([]any) {
		n := raw.(map[string]any)
		if n["x"].(float64) != 0 || n["y"].(float64) != 0 {
			return true
		}
	}
	return false
}

func TestLayoutCLI_PositionsAndPreserves(t *testing.T) {
	got := runLayoutToTemp(t, "in", ".d.canvas")

	var doc map[string]any
	if err := json.Unmarshal(got, &doc); err != nil {
		t.Fatalf("parse output: %v", err)
	}

	// At least one connected node must have moved off its (0,0) input position.
	if !anyNodeMoved(t, doc) {
		t.Error("expected at least one node to be positioned off (0,0)")
	}

	// Untouched fields, including Layer 2 project fields, must be preserved.
	if _, ok := doc["x-project"]; !ok {
		t.Error("top-level Layer 2 field x-project was dropped")
	}
	first := doc["nodes"].([]any)[0].(map[string]any)
	if first["x-journalText"] != "keep me" {
		t.Errorf("node Layer 2 field x-journalText not preserved: %v", first["x-journalText"])
	}
}

func TestLayoutCLI_AcceptsBothExtensions(t *testing.T) {
	for _, ext := range []string{".d.canvas", ".dcanvas"} {
		t.Run(ext, func(t *testing.T) {
			runLayoutToTemp(t, "in", ext) // fails the test if execute() errors
		})
	}
}

func TestLayoutCLI_RejectsUnknownExtension(t *testing.T) {
	if err := execute("layout", "canvas.txt"); err == nil {
		t.Error("expected an error for an unsupported extension")
	}
}

// copyFixture copies a testdata file into dir (keeping its name) and returns the
// destination path, so a test can run the CLI against a throwaway copy.
func copyFixture(t *testing.T, dir, name string) string {
	t.Helper()
	src, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	dst := filepath.Join(dir, name)
	if err := os.WriteFile(dst, src, 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return dst
}

func TestLayoutCLI_InvalidInputWritesNothing(t *testing.T) {
	// invalid.d.canvas decodes fine but has a dangling edge, which Validate
	// rejects; the CLI must fail and write no output.
	dir := t.TempDir()
	in := copyFixture(t, dir, "invalid.d.canvas")
	out := filepath.Join(dir, "out.d.canvas")

	err := execute("layout", "-o", out, in)
	if err == nil {
		t.Fatal("expected an error for an invalid canvas")
	}
	if _, statErr := os.Stat(out); !os.IsNotExist(statErr) {
		t.Errorf("no output should be written on validation failure; stat err = %v", statErr)
	}
}

func TestLayoutCLI_LayoutFailureWritesNothing(t *testing.T) {
	orig := layoutFn
	layoutFn = func(*dcanvas.Canvas, ...layout.Option) error { return errors.New("engine boom") }
	defer func() { layoutFn = orig }()

	dir := t.TempDir()
	in := copyFixture(t, dir, "input.d.canvas")
	out := filepath.Join(dir, "out.d.canvas")

	err := execute("layout", "-o", out, in)
	if err == nil {
		t.Fatal("expected the layout failure to be surfaced")
	}
	if _, statErr := os.Stat(out); !os.IsNotExist(statErr) {
		t.Errorf("no output should be written when layout fails; stat err = %v", statErr)
	}
}

func TestLayoutCLI_LoopStrategy(t *testing.T) {
	// loop-only.d.canvas has a single loop edge: cut drops it (no-op, nodes stay
	// at the origin); dfs lays it out (at least one node moves).
	anyMoved := func(loopVal string) bool {
		dir := t.TempDir()
		in := copyFixture(t, dir, "loop-only.d.canvas")
		out := filepath.Join(dir, "out.d.canvas")
		if err := execute("layout", "--loop", loopVal, "-o", out, in); err != nil {
			t.Fatalf("run --loop %s: %v", loopVal, err)
		}
		b, err := os.ReadFile(out)
		if err != nil {
			t.Fatalf("read output: %v", err)
		}
		var doc map[string]any
		if err := json.Unmarshal(b, &doc); err != nil {
			t.Fatalf("parse output: %v", err)
		}
		return anyNodeMoved(t, doc)
	}

	if anyMoved("cut") {
		t.Error("--loop cut should drop the loop edge and leave nodes at the origin")
	}
	if !anyMoved("dfs") {
		t.Error("--loop dfs should lay out the loop edge and move at least one node")
	}
}

func TestLayoutCLI_PreservesInputVersion(t *testing.T) {
	// A 3.1 input must round-trip through the CLI as 3.1, not be downgraded to
	// the library's 3.0 (relies on the lossless round-trip from task 0013).
	dir := t.TempDir()
	in := copyFixture(t, dir, "version-3.1.d.canvas")
	out := filepath.Join(dir, "out.d.canvas")
	if err := execute("layout", "-o", out, in); err != nil {
		t.Fatalf("run layout: %v", err)
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatalf("parse output: %v", err)
	}
	if doc["x-dCanvasVersion"] != "3.1" {
		t.Errorf("x-dCanvasVersion = %v, want \"3.1\"", doc["x-dCanvasVersion"])
	}
}

func TestLayoutCLI_RejectsUnknownLoopStrategy(t *testing.T) {
	if err := execute("layout", "--loop", "spiral", "canvas.d.canvas"); err == nil {
		t.Error("expected an error for an unknown --loop value")
	}
}
