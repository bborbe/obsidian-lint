// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package formatter

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/bborbe/errors"

	"github.com/bborbe/obsidian-lint/pkg/model"
)

//counterfeiter:generate -o ../../mocks/formatter.go --fake-name Formatter . Formatter

// Formatter formats validation results for output
type Formatter interface {
	Format(ctx context.Context, result *model.ValidationResult) (string, error)
}

// NewTextFormatter creates a text formatter
func NewTextFormatter() Formatter {
	return &textFormatter{}
}

type textFormatter struct{}

// Format outputs broken links in human-readable format
func (f *textFormatter) Format(
	ctx context.Context,
	result *model.ValidationResult,
) (string, error) {
	if len(result.BrokenLinks) == 0 {
		return "No broken links found.\n", nil
	}

	var sb strings.Builder
	sb.WriteString("Broken links found in vault:\n\n")

	// Sort files for consistent output
	files := make([]string, 0, len(result.BrokenLinks))
	for file := range result.BrokenLinks {
		files = append(files, file)
	}
	sort.Strings(files)

	for _, file := range files {
		links := result.BrokenLinks[file]

		sb.WriteString(file)
		sb.WriteString(":\n")

		for _, link := range links {
			sb.WriteString(fmt.Sprintf("  Line %d: %s\n", link.Line, link.Link))
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

// NewJSONFormatter creates a JSON formatter
func NewJSONFormatter() Formatter {
	return &jsonFormatter{}
}

type jsonFormatter struct{}

// Format outputs broken links in JSON format (grouped by file)
func (f *jsonFormatter) Format(
	ctx context.Context,
	result *model.ValidationResult,
) (string, error) {
	bytes, err := json.MarshalIndent(result.BrokenLinks, "", "  ")
	if err != nil {
		return "", errors.Wrap(ctx, err, "marshal json failed")
	}

	return string(bytes) + "\n", nil
}
