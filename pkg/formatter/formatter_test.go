// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package formatter_test

import (
	"context"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/obsidian-lint/pkg/formatter"
	"github.com/bborbe/obsidian-lint/pkg/model"
)

var _ = Describe("Formatter", func() {
	var (
		ctx context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
	})

	Context("TextFormatter", func() {
		var f formatter.Formatter

		BeforeEach(func() {
			f = formatter.NewTextFormatter()
		})

		It("formats broken links in text format", func() {
			result := &model.ValidationResult{
				BrokenLinks: map[string][]model.BrokenLink{
					"/vault/file1.md": {
						{Link: "[[Dead1]]", Line: 5},
						{Link: "[[Dead2]]", Line: 10},
					},
					"/vault/file2.md": {
						{Link: "![[missing.png]]", Line: 3},
					},
				},
			}

			output, err := f.Format(ctx, result)
			Expect(err).NotTo(HaveOccurred())

			Expect(output).To(ContainSubstring("Broken links found in vault:"))
			Expect(output).To(ContainSubstring("/vault/file1.md:"))
			Expect(output).To(ContainSubstring("Line 5: [[Dead1]]"))
			Expect(output).To(ContainSubstring("Line 10: [[Dead2]]"))
			Expect(output).To(ContainSubstring("/vault/file2.md:"))
			Expect(output).To(ContainSubstring("Line 3: ![[missing.png]]"))
		})

		It("returns success message when no broken links", func() {
			result := &model.ValidationResult{
				BrokenLinks: map[string][]model.BrokenLink{},
			}

			output, err := f.Format(ctx, result)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal("No broken links found.\n"))
		})

		It("sorts files alphabetically", func() {
			result := &model.ValidationResult{
				BrokenLinks: map[string][]model.BrokenLink{
					"/vault/zzz.md": {{Link: "[[Link3]]", Line: 1}},
					"/vault/aaa.md": {{Link: "[[Link1]]", Line: 1}},
					"/vault/mmm.md": {{Link: "[[Link2]]", Line: 1}},
				},
			}

			output, err := f.Format(ctx, result)
			Expect(err).NotTo(HaveOccurred())

			aIdx := indexOf(output, "/vault/aaa.md")
			mIdx := indexOf(output, "/vault/mmm.md")
			zIdx := indexOf(output, "/vault/zzz.md")

			Expect(aIdx).To(BeNumerically("<", mIdx))
			Expect(mIdx).To(BeNumerically("<", zIdx))
		})
	})

	Context("JSONFormatter", func() {
		var f formatter.Formatter

		BeforeEach(func() {
			f = formatter.NewJSONFormatter()
		})

		It("formats broken links in JSON format", func() {
			result := &model.ValidationResult{
				BrokenLinks: map[string][]model.BrokenLink{
					"/vault/file1.md": {
						{Link: "[[Dead1]]", Line: 5},
						{Link: "[[Dead2]]", Line: 10},
					},
					"/vault/file2.md": {
						{Link: "![[missing.png]]", Line: 3},
					},
				},
			}

			output, err := f.Format(ctx, result)
			Expect(err).NotTo(HaveOccurred())

			// Parse JSON to verify structure
			var parsed map[string][]model.BrokenLink
			err = json.Unmarshal([]byte(output), &parsed)
			Expect(err).NotTo(HaveOccurred())

			Expect(parsed).To(HaveLen(2))
			Expect(parsed["/vault/file1.md"]).To(HaveLen(2))
			Expect(parsed["/vault/file1.md"][0].Link).To(Equal("[[Dead1]]"))
			Expect(parsed["/vault/file1.md"][0].Line).To(Equal(5))
			Expect(parsed["/vault/file1.md"][1].Link).To(Equal("[[Dead2]]"))
			Expect(parsed["/vault/file1.md"][1].Line).To(Equal(10))
			Expect(parsed["/vault/file2.md"]).To(HaveLen(1))
			Expect(parsed["/vault/file2.md"][0].Link).To(Equal("![[missing.png]]"))
			Expect(parsed["/vault/file2.md"][0].Line).To(Equal(3))
		})

		It("returns empty JSON object when no broken links", func() {
			result := &model.ValidationResult{
				BrokenLinks: map[string][]model.BrokenLink{},
			}

			output, err := f.Format(ctx, result)
			Expect(err).NotTo(HaveOccurred())

			var parsed map[string][]model.BrokenLink
			err = json.Unmarshal([]byte(output), &parsed)
			Expect(err).NotTo(HaveOccurred())
			Expect(parsed).To(BeEmpty())
		})

		It("outputs valid JSON", func() {
			result := &model.ValidationResult{
				BrokenLinks: map[string][]model.BrokenLink{
					"/vault/test.md": {{Link: "[[Test]]", Line: 1}},
				},
			}

			output, err := f.Format(ctx, result)
			Expect(err).NotTo(HaveOccurred())

			// Must be valid JSON
			var parsed interface{}
			err = json.Unmarshal([]byte(output), &parsed)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

// indexOf returns the index of substr in s, or -1 if not found
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
