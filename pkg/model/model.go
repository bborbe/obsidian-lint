// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package model

// Link represents a wiki link found in a markdown file
type Link struct {
	Raw     string // "[[Note#Heading|alias]]"
	Target  string // "Note" (before # or |)
	Heading string // "Heading" (optional)
	Alias   string // "alias" (optional)
	IsEmbed bool   // true if "![[..."
	Line    int    // line number in file
}

// BrokenLink represents a broken link in output
type BrokenLink struct {
	Link string `json:"link"`
	Line int    `json:"line"`
}

// ValidationResult contains all broken links grouped by file
type ValidationResult struct {
	BrokenLinks map[string][]BrokenLink // file path -> broken links
}
