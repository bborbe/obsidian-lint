// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parser

import (
	"context"
	"os"
	"regexp"
	"strings"

	"github.com/bborbe/errors"
	"gopkg.in/yaml.v3"

	"github.com/bborbe/obsidian-lint/pkg/model"
)

//counterfeiter:generate -o ../../mocks/parser.go --fake-name Parser . Parser

// Parser extracts wiki links from markdown files
type Parser interface {
	ParseFile(ctx context.Context, filePath string) ([]*model.Link, error)
	ParseAliases(ctx context.Context, content string) ([]string, error)
}

// New creates a new Parser
func New() Parser {
	return &parser{
		linkRegex: regexp.MustCompile(`(!?\[\[([^\]]+)\]\])`),
	}
}

type parser struct {
	linkRegex *regexp.Regexp
}

// ParseFile extracts all wiki links from a markdown file
func (p *parser) ParseFile(ctx context.Context, filePath string) ([]*model.Link, error) {
	// #nosec G304 -- filePath comes from scanner.Scan(), not user input
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrap(ctx, err, "read file failed")
	}

	return p.parseContent(string(content)), nil
}

// parseContent extracts links from markdown content
func (p *parser) parseContent(content string) []*model.Link {
	var links []*model.Link
	lines := strings.Split(content, "\n")

	for lineNum, line := range lines {
		matches := p.linkRegex.FindAllStringSubmatch(line, -1)

		for _, match := range matches {
			if len(match) < 3 {
				continue
			}

			raw := match[1]   // Full match: ![[Note#Heading|alias]]
			inner := match[2] // Inner content: Note#Heading|alias
			isEmbed := strings.HasPrefix(raw, "!")

			link := p.parseLink(raw, inner, isEmbed, lineNum+1)
			links = append(links, link)
		}
	}

	return links
}

// parseLink parses a single link into components
func (p *parser) parseLink(raw, inner string, isEmbed bool, lineNum int) *model.Link {
	link := &model.Link{
		Raw:     raw,
		IsEmbed: isEmbed,
		Line:    lineNum,
	}

	// Split on | first to separate alias
	parts := strings.SplitN(inner, "|", 2)
	targetPart := parts[0]
	if len(parts) > 1 {
		link.Alias = parts[1]
	}

	// Split on # to separate heading
	targetParts := strings.SplitN(targetPart, "#", 2)
	link.Target = strings.TrimSpace(targetParts[0])
	if len(targetParts) > 1 {
		link.Heading = strings.TrimSpace(targetParts[1])
	}

	return link
}

// ParseAliases extracts aliases from YAML frontmatter
func (p *parser) ParseAliases(ctx context.Context, content string) ([]string, error) {
	frontmatter := extractFrontmatter(content)
	if frontmatter == "" {
		return nil, nil
	}

	var data struct {
		Aliases interface{} `yaml:"aliases"`
	}

	if err := yaml.Unmarshal([]byte(frontmatter), &data); err != nil {
		// Silently ignore YAML parsing errors - frontmatter might be malformed
		return nil, nil
	}

	if data.Aliases == nil {
		return nil, nil
	}

	// Handle both string and []interface{} (yaml array)
	switch v := data.Aliases.(type) {
	case string:
		return []string{v}, nil
	case []interface{}:
		var aliases []string
		for _, item := range v {
			if str, ok := item.(string); ok {
				aliases = append(aliases, str)
			}
		}
		return aliases, nil
	default:
		return nil, nil
	}
}

// extractFrontmatter extracts YAML frontmatter between --- markers
func extractFrontmatter(content string) string {
	if !strings.HasPrefix(content, "---\n") {
		return ""
	}

	// Find second --- delimiter
	parts := strings.SplitN(content[4:], "\n---\n", 2)
	if len(parts) < 2 {
		return ""
	}

	return parts[0]
}
