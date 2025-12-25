// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package validator

import (
	"context"

	"github.com/bborbe/errors"

	"github.com/bborbe/obsidian-lint/pkg/index"
	"github.com/bborbe/obsidian-lint/pkg/model"
	"github.com/bborbe/obsidian-lint/pkg/parser"
	"github.com/bborbe/obsidian-lint/pkg/resolver"
	"github.com/bborbe/obsidian-lint/pkg/scanner"
)

//counterfeiter:generate -o ../../mocks/validator.go --fake-name Validator . Validator

// Validator orchestrates vault scanning and link validation
type Validator interface {
	Validate(ctx context.Context, vaultPath string) (*model.ValidationResult, error)
}

// New creates a new Validator
func New(
	scanner scanner.Scanner,
	parser parser.Parser,
	indexBuilder index.Builder,
	resolver resolver.Resolver,
) Validator {
	return &validator{
		scanner:      scanner,
		parser:       parser,
		indexBuilder: indexBuilder,
		resolver:     resolver,
	}
}

type validator struct {
	scanner      scanner.Scanner
	parser       parser.Parser
	indexBuilder index.Builder
	resolver     resolver.Resolver
}

// Validate scans vault and returns broken links
func (v *validator) Validate(
	ctx context.Context,
	vaultPath string,
) (*model.ValidationResult, error) {
	// Scan vault for markdown files
	files, err := v.scanner.Scan(ctx, vaultPath)
	if err != nil {
		return nil, errors.Wrap(ctx, err, "scan failed")
	}

	// Build vault index
	idx, err := v.indexBuilder.Build(ctx, vaultPath, files)
	if err != nil {
		return nil, errors.Wrap(ctx, err, "build index failed")
	}

	// Validate links in each file
	result := &model.ValidationResult{
		BrokenLinks: make(map[string][]model.BrokenLink),
	}

	for _, file := range files {
		links, err := v.parser.ParseFile(ctx, file)
		if err != nil {
			return nil, errors.Wrap(ctx, err, "parse file failed")
		}

		for _, link := range links {
			if !v.resolver.Resolve(ctx, link, idx) {
				result.BrokenLinks[file] = append(result.BrokenLinks[file], model.BrokenLink{
					Link: link.Raw,
					Line: link.Line,
				})
			}
		}
	}

	return result, nil
}
