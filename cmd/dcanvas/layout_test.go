// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: BlueOak-1.0.0

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

// update regenerates the golden file: `go test ./cmd/dcanvas/ -update`.
var update = flag.Bool("update", false, "update the golden output file")

const goldenPath = "testdata/golden.d.canvas"

// runLayoutToTemp copies the input testdata to a temp file with the given
// extension, runs the CLI's layout command against it, and returns the output
// bytes. Driving through run() exercises the tool end-to-end.
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
	if err := run([]string{"layout", "-o", outPath, inPath}); err != nil {
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

func TestLayoutCLI_PositionsAndPreserves(t *testing.T) {
	got := runLayoutToTemp(t, "in", ".d.canvas")

	var doc map[string]any
	if err := json.Unmarshal(got, &doc); err != nil {
		t.Fatalf("parse output: %v", err)
	}

	// At least one connected node must have moved off its (0,0) input position.
	nodes := doc["nodes"].([]any)
	moved := false
	for _, raw := range nodes {
		n := raw.(map[string]any)
		if n["x"].(float64) != 0 || n["y"].(float64) != 0 {
			moved = true
			break
		}
	}
	if !moved {
		t.Error("expected at least one node to be positioned off (0,0)")
	}

	// Untouched fields, including Layer 2 project fields, must be preserved.
	if _, ok := doc["x-project"]; !ok {
		t.Error("top-level Layer 2 field x-project was dropped")
	}
	first := nodes[0].(map[string]any)
	if first["x-journalText"] != "keep me" {
		t.Errorf("node Layer 2 field x-journalText not preserved: %v", first["x-journalText"])
	}
}

func TestLayoutCLI_AcceptsBothExtensions(t *testing.T) {
	for _, ext := range []string{".d.canvas", ".dcanvas"} {
		t.Run(ext, func(t *testing.T) {
			runLayoutToTemp(t, "in", ext) // fails the test if run() errors
		})
	}
}

func TestLayoutCLI_RejectsUnknownExtension(t *testing.T) {
	if err := run([]string{"layout", "canvas.txt"}); err == nil {
		t.Error("expected an error for an unsupported extension")
	}
}
