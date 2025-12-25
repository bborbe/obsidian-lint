// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parser_test

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/obsidian-lint/pkg/parser"
)

var _ = Describe("Parser", func() {
	var (
		ctx     context.Context
		p       parser.Parser
		tempDir string
		err     error
	)

	BeforeEach(func() {
		ctx = context.Background()
		p = parser.New()

		tempDir, err = os.MkdirTemp("", "parser-test")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if tempDir != "" {
			_ = os.RemoveAll(tempDir)
		}
	})

	Context("ParseFile", func() {
		It("extracts basic wiki link", func() {
			content := "This is a link to [[Note]]."
			file := filepath.Join(tempDir, "test.md")
			Expect(os.WriteFile(file, []byte(content), 0600)).To(Succeed())

			links, err := p.ParseFile(ctx, file)
			Expect(err).NotTo(HaveOccurred())
			Expect(links).To(HaveLen(1))
			Expect(links[0].Raw).To(Equal("[[Note]]"))
			Expect(links[0].Target).To(Equal("Note"))
			Expect(links[0].Heading).To(BeEmpty())
			Expect(links[0].Alias).To(BeEmpty())
			Expect(links[0].IsEmbed).To(BeFalse())
			Expect(links[0].Line).To(Equal(1))
		})

		It("extracts link with alias", func() {
			content := "Link: [[Note|Display Text]]"
			file := filepath.Join(tempDir, "test.md")
			Expect(os.WriteFile(file, []byte(content), 0600)).To(Succeed())

			links, err := p.ParseFile(ctx, file)
			Expect(err).NotTo(HaveOccurred())
			Expect(links).To(HaveLen(1))
			Expect(links[0].Raw).To(Equal("[[Note|Display Text]]"))
			Expect(links[0].Target).To(Equal("Note"))
			Expect(links[0].Alias).To(Equal("Display Text"))
			Expect(links[0].IsEmbed).To(BeFalse())
		})

		It("extracts link with heading", func() {
			content := "Link: [[Note#Heading]]"
			file := filepath.Join(tempDir, "test.md")
			Expect(os.WriteFile(file, []byte(content), 0600)).To(Succeed())

			links, err := p.ParseFile(ctx, file)
			Expect(err).NotTo(HaveOccurred())
			Expect(links).To(HaveLen(1))
			Expect(links[0].Raw).To(Equal("[[Note#Heading]]"))
			Expect(links[0].Target).To(Equal("Note"))
			Expect(links[0].Heading).To(Equal("Heading"))
			Expect(links[0].Alias).To(BeEmpty())
		})

		It("extracts link with heading and alias", func() {
			content := "Link: [[Note#Heading|Display]]"
			file := filepath.Join(tempDir, "test.md")
			Expect(os.WriteFile(file, []byte(content), 0600)).To(Succeed())

			links, err := p.ParseFile(ctx, file)
			Expect(err).NotTo(HaveOccurred())
			Expect(links).To(HaveLen(1))
			Expect(links[0].Raw).To(Equal("[[Note#Heading|Display]]"))
			Expect(links[0].Target).To(Equal("Note"))
			Expect(links[0].Heading).To(Equal("Heading"))
			Expect(links[0].Alias).To(Equal("Display"))
		})

		It("extracts embed link", func() {
			content := "Embed: ![[Note]]"
			file := filepath.Join(tempDir, "test.md")
			Expect(os.WriteFile(file, []byte(content), 0600)).To(Succeed())

			links, err := p.ParseFile(ctx, file)
			Expect(err).NotTo(HaveOccurred())
			Expect(links).To(HaveLen(1))
			Expect(links[0].Raw).To(Equal("![[Note]]"))
			Expect(links[0].Target).To(Equal("Note"))
			Expect(links[0].IsEmbed).To(BeTrue())
		})

		It("extracts embed with file extension", func() {
			content := "Image: ![[image.png]]"
			file := filepath.Join(tempDir, "test.md")
			Expect(os.WriteFile(file, []byte(content), 0600)).To(Succeed())

			links, err := p.ParseFile(ctx, file)
			Expect(err).NotTo(HaveOccurred())
			Expect(links).To(HaveLen(1))
			Expect(links[0].Raw).To(Equal("![[image.png]]"))
			Expect(links[0].Target).To(Equal("image.png"))
			Expect(links[0].IsEmbed).To(BeTrue())
		})

		It("extracts multiple links from same line", func() {
			content := "Links: [[Note1]] and [[Note2]]"
			file := filepath.Join(tempDir, "test.md")
			Expect(os.WriteFile(file, []byte(content), 0600)).To(Succeed())

			links, err := p.ParseFile(ctx, file)
			Expect(err).NotTo(HaveOccurred())
			Expect(links).To(HaveLen(2))
			Expect(links[0].Target).To(Equal("Note1"))
			Expect(links[1].Target).To(Equal("Note2"))
			Expect(links[0].Line).To(Equal(1))
			Expect(links[1].Line).To(Equal(1))
		})

		It("tracks line numbers correctly", func() {
			content := "Line 1: [[Note1]]\nLine 2\nLine 3: [[Note2]]"
			file := filepath.Join(tempDir, "test.md")
			Expect(os.WriteFile(file, []byte(content), 0600)).To(Succeed())

			links, err := p.ParseFile(ctx, file)
			Expect(err).NotTo(HaveOccurred())
			Expect(links).To(HaveLen(2))
			Expect(links[0].Line).To(Equal(1))
			Expect(links[1].Line).To(Equal(3))
		})

		It("handles folder paths in links", func() {
			content := "Link: [[folder/Note]]"
			file := filepath.Join(tempDir, "test.md")
			Expect(os.WriteFile(file, []byte(content), 0600)).To(Succeed())

			links, err := p.ParseFile(ctx, file)
			Expect(err).NotTo(HaveOccurred())
			Expect(links).To(HaveLen(1))
			Expect(links[0].Target).To(Equal("folder/Note"))
		})

		It("returns empty slice when no links exist", func() {
			content := "No links here, just text."
			file := filepath.Join(tempDir, "test.md")
			Expect(os.WriteFile(file, []byte(content), 0600)).To(Succeed())

			links, err := p.ParseFile(ctx, file)
			Expect(err).NotTo(HaveOccurred())
			Expect(links).To(BeEmpty())
		})
	})

	Context("ParseAliases", func() {
		It("extracts single alias from frontmatter", func() {
			content := `---
aliases: MyAlias
---
Content here`

			aliases, err := p.ParseAliases(ctx, content)
			Expect(err).NotTo(HaveOccurred())
			Expect(aliases).To(Equal([]string{"MyAlias"}))
		})

		It("extracts array of aliases from frontmatter", func() {
			content := `---
aliases: [AI, Artificial Intelligence, ML]
---
Content here`

			aliases, err := p.ParseAliases(ctx, content)
			Expect(err).NotTo(HaveOccurred())
			Expect(aliases).To(ConsistOf("AI", "Artificial Intelligence", "ML"))
		})

		It("extracts multiline aliases from frontmatter", func() {
			content := `---
aliases:
  - First Alias
  - Second Alias
  - Third Alias
---
Content here`

			aliases, err := p.ParseAliases(ctx, content)
			Expect(err).NotTo(HaveOccurred())
			Expect(aliases).To(ConsistOf("First Alias", "Second Alias", "Third Alias"))
		})

		It("returns nil when no aliases field exists", func() {
			content := `---
title: My Note
---
Content here`

			aliases, err := p.ParseAliases(ctx, content)
			Expect(err).NotTo(HaveOccurred())
			Expect(aliases).To(BeNil())
		})

		It("returns nil when no frontmatter exists", func() {
			content := "Just content, no frontmatter"

			aliases, err := p.ParseAliases(ctx, content)
			Expect(err).NotTo(HaveOccurred())
			Expect(aliases).To(BeNil())
		})

		It("returns nil when frontmatter is incomplete", func() {
			content := `---
aliases: [AI, ML]
Content without closing ---`

			aliases, err := p.ParseAliases(ctx, content)
			Expect(err).NotTo(HaveOccurred())
			Expect(aliases).To(BeNil())
		})
	})
})
