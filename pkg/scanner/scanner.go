// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scanner

import (
	"context"
	"io/fs"
	"path/filepath"

	"github.com/bborbe/errors"
)

//counterfeiter:generate -o ../../mocks/scanner.go --fake-name Scanner . Scanner

// Scanner walks a vault directory and returns all markdown files
type Scanner interface {
	Scan(ctx context.Context, vaultPath string) ([]string, error)
}

// New creates a new Scanner
func New() Scanner {
	return &scanner{}
}

type scanner struct{}

// Scan walks the vault directory and returns all .md files
func (s *scanner) Scan(ctx context.Context, vaultPath string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(vaultPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return errors.Wrap(ctx, err, "walk error")
		}

		// Skip context cancellation check
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only include .md files
		if filepath.Ext(path) == ".md" {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, errors.Wrap(ctx, err, "scan failed")
	}

	return files, nil
}
