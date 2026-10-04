// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: BlueOak-1.0.0

package dcanvas

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// schemaIDBase is the canonical home of the schema files now that this
// repository is the master source of the format. A guard test pins each
// schema's $id to it so the id cannot silently drift back to a stale,
// unrelated repository.
const schemaIDBase = "https://github.com/sbtlocalization/dcanvas/spec/"

// Schema files stay one per minor (ADR-0015). schema11 is the contract for what
// the library writes today; schema10 is kept to prove that a 1.1 document which
// only adds fields is still structurally a 1.0 document.
const (
	schema10 = "dCanvas-1.0.schema.json"
	schema11 = "dCanvas-1.1.schema.json"
)

// schemas lists every per-minor schema file, oldest first.
var schemas = []string{schema10, schema11}

// loadSchemaBytes reads a shared JSON Schema from spec/ relative to the package
// directory (the repository root, where `go test` runs), failing the test if it
// cannot.
func loadSchemaBytes(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join("spec", name)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read schema %s: %v", path, err)
	}
	return b
}

// TestSchema_CanonicalID guards each schema's identity: its $id must equal the
// canonical master-repository URL of its own file.
func TestSchema_CanonicalID(t *testing.T) {
	for _, name := range schemas {
		t.Run(name, func(t *testing.T) {
			var doc struct {
				ID string `json:"$id"`
			}
			if err := json.Unmarshal(loadSchemaBytes(t, name), &doc); err != nil {
				t.Fatalf("parse schema: %v", err)
			}
			if want := schemaIDBase + name; doc.ID != want {
				t.Errorf("schema $id = %q, want %q", doc.ID, want)
			}
		})
	}
}
