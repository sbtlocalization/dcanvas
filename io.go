// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: BlueOak-1.0.0

package dcanvas

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Encode writes a Canvas as indented JSON to w. The document is always stamped
// with the current format version and never loses preserved unknown fields.
func Encode(c *Canvas, w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "\t")
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("can't encode dcanvas: %w", err)
	}
	return nil
}

// Decode reads a Canvas from JSON in r. It validates the major version of
// x-dCanvasVersion and rejects any document that is not 3.x (a 2.0 file is
// rejected); there is no migration path. Unrecognised fields are preserved.
func Decode(r io.Reader) (*Canvas, error) {
	var c Canvas
	if err := json.NewDecoder(r).Decode(&c); err != nil {
		return nil, fmt.Errorf("can't decode dcanvas: %w", err)
	}
	if c.Version == "" {
		return nil, fmt.Errorf("can't decode dcanvas: missing x-dCanvasVersion")
	}
	major, _, _ := strings.Cut(c.Version, ".")
	want, _, _ := strings.Cut(Version, ".")
	if major != want {
		return nil, fmt.Errorf("can't decode dcanvas: unsupported major version %q (want %s.x)", c.Version, want)
	}
	return &c, nil
}
