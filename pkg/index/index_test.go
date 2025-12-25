// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package index_test

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/obsidian-lint/pkg/index"
	"github.com/bborbe/obsidian-lint/pkg/parser"
)

var _ = Describe("IndexBuilder", func() {
	var (
		ctx     context.Context
		builder index.Builder
		p       parser.Parser
		tempDir string
		err     error
	)

	BeforeEach(func() {
		ctx = context.Background()
		p = parser.New()
		builder = index.New(p)

		tempDir, err = os.MkdirTemp("", "index-test")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if tempDir != "" {
			_ = os.RemoveAll(tempDir)
		}
	})

	Context("Build", func() {
		It("indexes files by normalized filename", func() {
			file1 := filepath.Join(tempDir, "Note1.md")
			file2 := filepath.Join(tempDir, "Note2.md")

			Expect(os.WriteFile(file1, []byte("content"), 0600)).To(Succeed())
			Expect(os.WriteFile(file2, []byte("content"), 0600)).To(Succeed())

			idx, err := builder.Build(ctx, tempDir, []string{file1, file2})
			Expect(err).NotTo(HaveOccurred())

			// Case-insensitive resolution
			Expect(idx.Resolve("Note1")).To(BeTrue())
			Expect(idx.Resolve("note1")).To(BeTrue())
			Expect(idx.Resolve("NOTE1")).To(BeTrue())
			Expect(idx.Resolve("Note2")).To(BeTrue())
			Expect(idx.Resolve("note2")).To(BeTrue())
		})

		It("indexes files without .md extension", func() {
			file := filepath.Join(tempDir, "MyNote.md")
			Expect(os.WriteFile(file, []byte("content"), 0600)).To(Succeed())

			idx, err := builder.Build(ctx, tempDir, []string{file})
			Expect(err).NotTo(HaveOccurred())

			// Should resolve both with and without extension
			Expect(idx.Resolve("MyNote")).To(BeTrue())
			Expect(idx.Resolve("MyNote.md")).To(BeTrue())
			Expect(idx.Resolve("mynote")).To(BeTrue())
		})

		It("indexes aliases from YAML frontmatter", func() {
			content := `---
aliases: [AI, Artificial Intelligence]
---
Note content`

			file := filepath.Join(tempDir, "Note.md")
			Expect(os.WriteFile(file, []byte(content), 0600)).To(Succeed())

			idx, err := builder.Build(ctx, tempDir, []string{file})
			Expect(err).NotTo(HaveOccurred())

			// Should resolve by filename
			Expect(idx.Resolve("Note")).To(BeTrue())

			// Should resolve by aliases (case-insensitive)
			Expect(idx.Resolve("AI")).To(BeTrue())
			Expect(idx.Resolve("ai")).To(BeTrue())
			Expect(idx.Resolve("Artificial Intelligence")).To(BeTrue())
			Expect(idx.Resolve("artificial intelligence")).To(BeTrue())
		})

		It("handles files with no aliases", func() {
			file := filepath.Join(tempDir, "Simple.md")
			Expect(os.WriteFile(file, []byte("No frontmatter"), 0600)).To(Succeed())

			idx, err := builder.Build(ctx, tempDir, []string{file})
			Expect(err).NotTo(HaveOccurred())

			Expect(idx.Resolve("Simple")).To(BeTrue())
		})

		It("returns false for non-existent targets", func() {
			file := filepath.Join(tempDir, "Exists.md")
			Expect(os.WriteFile(file, []byte("content"), 0600)).To(Succeed())

			idx, err := builder.Build(ctx, tempDir, []string{file})
			Expect(err).NotTo(HaveOccurred())

			Expect(idx.Resolve("DoesNotExist")).To(BeFalse())
			Expect(idx.Resolve("Missing")).To(BeFalse())
		})

		It("handles empty file list", func() {
			idx, err := builder.Build(ctx, tempDir, []string{})
			Expect(err).NotTo(HaveOccurred())

			Expect(idx.Resolve("Anything")).To(BeFalse())
		})

		It("handles filenames with special characters", func() {
			file := filepath.Join(tempDir, "File With Spaces.md")
			Expect(os.WriteFile(file, []byte("content"), 0600)).To(Succeed())

			idx, err := builder.Build(ctx, tempDir, []string{file})
			Expect(err).NotTo(HaveOccurred())

			Expect(idx.Resolve("File With Spaces")).To(BeTrue())
			Expect(idx.Resolve("file with spaces")).To(BeTrue())
		})
	})
})
