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

// canonicalSchemaID is the schema's canonical identity now that this repository
// is the master source of the format. A guard test pins it so the id cannot
// silently drift back to a stale, unrelated repository.
const canonicalSchemaID = "https://github.com/sbtlocalization/dcanvas/docs/dcanvas-3.0.schema.json"

// schemaPath is the location of the shared JSON Schema relative to the package
// directory (the repository root, where `go test` runs).
var schemaPath = filepath.Join("docs", "dcanvas-3.0.schema.json")

// loadSchemaBytes reads the shared JSON Schema, failing the test if it cannot.
func loadSchemaBytes(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("read schema %s: %v", schemaPath, err)
	}
	return b
}

// TestSchema_CanonicalID guards the schema's identity: the loaded schema's $id
// must equal the canonical master-repository URL.
func TestSchema_CanonicalID(t *testing.T) {
	var doc struct {
		ID string `json:"$id"`
	}
	if err := json.Unmarshal(loadSchemaBytes(t), &doc); err != nil {
		t.Fatalf("parse schema: %v", err)
	}
	if doc.ID != canonicalSchemaID {
		t.Errorf("schema $id = %q, want %q", doc.ID, canonicalSchemaID)
	}
}
