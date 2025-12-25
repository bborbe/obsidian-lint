// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package index

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/bborbe/errors"

	"github.com/bborbe/obsidian-lint/pkg/parser"
)

//counterfeiter:generate -o ../../mocks/index_builder.go --fake-name IndexBuilder . Builder

// Builder builds a vault index from files
type Builder interface {
	Build(ctx context.Context, vaultPath string, files []string) (*VaultIndex, error)
}

// New creates a new Builder
func New(parser parser.Parser) Builder {
	return &indexBuilder{
		parser: parser,
	}
}

type indexBuilder struct {
	parser parser.Parser
}

// Build creates a VaultIndex from markdown files and all files in vault
func (b *indexBuilder) Build(
	ctx context.Context,
	vaultPath string,
	files []string,
) (*VaultIndex, error) {
	index := &VaultIndex{
		files:   make(map[string]string),
		aliases: make(map[string]string),
	}

	// Index all files in vault (for embeds to images, PDFs, etc.)
	err := filepath.WalkDir(vaultPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// Index all files by normalized basename
		baseName := filepath.Base(path)
		normalized := normalizeTarget(baseName)
		index.files[normalized] = path

		return nil
	})

	if err != nil {
		return nil, errors.Wrap(ctx, err, "walk vault failed")
	}

	// Parse and index aliases from markdown files
	for _, file := range files {
		// #nosec G304 -- file paths come from scanner.Scan(), not user input
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, errors.Wrap(ctx, err, "read file failed")
		}

		aliases, err := b.parser.ParseAliases(ctx, string(content))
		if err != nil {
			return nil, errors.Wrap(ctx, err, "parse aliases failed")
		}

		for _, alias := range aliases {
			normalizedAlias := normalizeTarget(alias)
			index.aliases[normalizedAlias] = file
		}
	}

	return index, nil
}

// VaultIndex contains normalized file and alias mappings
type VaultIndex struct {
	files   map[string]string // normalized filename -> absolute path
	aliases map[string]string // normalized alias -> absolute path
}

// Resolve checks if a target exists in the index (case-insensitive)
func (v *VaultIndex) Resolve(target string) bool {
	normalized := normalizeTarget(target)

	// Check files first
	if _, exists := v.files[normalized]; exists {
		return true
	}

	// Check aliases
	if _, exists := v.aliases[normalized]; exists {
		return true
	}

	return false
}

// normalizeTarget converts a target to normalized form for case-insensitive matching
func normalizeTarget(target string) string {
	// Remove .md extension if present
	target = strings.TrimSuffix(target, ".md")

	// Convert to lowercase for case-insensitive matching
	return strings.ToLower(target)
}
