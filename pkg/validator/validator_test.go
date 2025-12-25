// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package validator_test

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/obsidian-lint/pkg/index"
	"github.com/bborbe/obsidian-lint/pkg/parser"
	"github.com/bborbe/obsidian-lint/pkg/resolver"
	"github.com/bborbe/obsidian-lint/pkg/scanner"
	"github.com/bborbe/obsidian-lint/pkg/validator"
)

var _ = Describe("Validator", func() {
	var (
		ctx     context.Context
		v       validator.Validator
		tempDir string
		err     error
	)

	BeforeEach(func() {
		ctx = context.Background()

		// Create real implementations
		s := scanner.New()
		p := parser.New()
		b := index.New(p)
		r := resolver.New()
		v = validator.New(s, p, b, r)

		tempDir, err = os.MkdirTemp("", "validator-test")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if tempDir != "" {
			_ = os.RemoveAll(tempDir)
		}
	})

	Context("Validate", func() {
		It("detects broken link", func() {
			// Create vault with broken link
			note := filepath.Join(tempDir, "Note.md")
			content := "This links to [[DoesNotExist]]"
			Expect(os.WriteFile(note, []byte(content), 0600)).To(Succeed())

			result, err := v.Validate(ctx, tempDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.BrokenLinks).To(HaveLen(1))
			Expect(result.BrokenLinks[note]).To(HaveLen(1))
			Expect(result.BrokenLinks[note][0].Link).To(Equal("[[DoesNotExist]]"))
			Expect(result.BrokenLinks[note][0].Line).To(Equal(1))
		})

		It("allows valid link to existing note", func() {
			// Create vault with valid link
			note1 := filepath.Join(tempDir, "Note1.md")
			note2 := filepath.Join(tempDir, "Note2.md")
			Expect(os.WriteFile(note1, []byte("This links to [[Note2]]"), 0600)).To(Succeed())
			Expect(os.WriteFile(note2, []byte("Content"), 0600)).To(Succeed())

			result, err := v.Validate(ctx, tempDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.BrokenLinks).To(BeEmpty())
		})

		It("allows link via alias", func() {
			note1 := filepath.Join(tempDir, "Note1.md")
			note2 := filepath.Join(tempDir, "Note2.md")

			Expect(os.WriteFile(note1, []byte("Link: [[MyAlias]]"), 0600)).To(Succeed())
			Expect(os.WriteFile(note2, []byte(`---
aliases: [MyAlias]
---
Content`), 0600)).To(Succeed())

			result, err := v.Validate(ctx, tempDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.BrokenLinks).To(BeEmpty())
		})

		It("detects broken embed", func() {
			note := filepath.Join(tempDir, "Note.md")
			content := "Embed: ![[missing.png]]"
			Expect(os.WriteFile(note, []byte(content), 0600)).To(Succeed())

			result, err := v.Validate(ctx, tempDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.BrokenLinks).To(HaveLen(1))
			Expect(result.BrokenLinks[note][0].Link).To(Equal("![[missing.png]]"))
		})

		It("allows valid embed to existing file", func() {
			note := filepath.Join(tempDir, "Note.md")
			image := filepath.Join(tempDir, "image.png")

			Expect(os.WriteFile(note, []byte("Embed: ![[image.png]]"), 0600)).To(Succeed())
			Expect(os.WriteFile(image, []byte("fake image"), 0600)).To(Succeed())

			result, err := v.Validate(ctx, tempDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.BrokenLinks).To(BeEmpty())
		})

		It("detects multiple broken links in same file", func() {
			note := filepath.Join(tempDir, "Note.md")
			content := "Link1: [[Dead1]]\nLink2: [[Dead2]]\nLink3: [[Dead3]]"
			Expect(os.WriteFile(note, []byte(content), 0600)).To(Succeed())

			result, err := v.Validate(ctx, tempDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.BrokenLinks).To(HaveLen(1))
			Expect(result.BrokenLinks[note]).To(HaveLen(3))
		})

		It("detects broken links across multiple files", func() {
			note1 := filepath.Join(tempDir, "Note1.md")
			note2 := filepath.Join(tempDir, "Note2.md")

			Expect(os.WriteFile(note1, []byte("[[Dead1]]"), 0600)).To(Succeed())
			Expect(os.WriteFile(note2, []byte("[[Dead2]]"), 0600)).To(Succeed())

			result, err := v.Validate(ctx, tempDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.BrokenLinks).To(HaveLen(2))
			Expect(result.BrokenLinks[note1]).To(HaveLen(1))
			Expect(result.BrokenLinks[note2]).To(HaveLen(1))
		})

		It("ignores heading when validating", func() {
			note1 := filepath.Join(tempDir, "Note1.md")
			note2 := filepath.Join(tempDir, "Note2.md")

			// Link has heading but heading doesn't exist in Note2
			Expect(os.WriteFile(note1, []byte("[[Note2#NonExistentHeading]]"), 0600)).To(Succeed())
			Expect(os.WriteFile(note2, []byte("No headings here"), 0600)).To(Succeed())

			result, err := v.Validate(ctx, tempDir)
			Expect(err).NotTo(HaveOccurred())
			// Should be valid because we only check if Note2 exists
			Expect(result.BrokenLinks).To(BeEmpty())
		})

		It("returns empty result for vault with no broken links", func() {
			note1 := filepath.Join(tempDir, "Note1.md")
			note2 := filepath.Join(tempDir, "Note2.md")

			Expect(os.WriteFile(note1, []byte("Valid: [[Note2]]"), 0600)).To(Succeed())
			Expect(os.WriteFile(note2, []byte("Valid: [[Note1]]"), 0600)).To(Succeed())

			result, err := v.Validate(ctx, tempDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.BrokenLinks).To(BeEmpty())
		})

		It("handles empty vault", func() {
			result, err := v.Validate(ctx, tempDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.BrokenLinks).To(BeEmpty())
		})
	})
})
