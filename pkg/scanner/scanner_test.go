// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scanner_test

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/obsidian-lint/pkg/scanner"
)

var _ = Describe("Scanner", func() {
	var (
		ctx     context.Context
		s       scanner.Scanner
		tempDir string
		err     error
	)

	BeforeEach(func() {
		ctx = context.Background()
		s = scanner.New()

		tempDir, err = os.MkdirTemp("", "scanner-test")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if tempDir != "" {
			_ = os.RemoveAll(tempDir)
		}
	})

	Context("Scan", func() {
		It("returns .md files in vault", func() {
			// Create test files
			mdFile1 := filepath.Join(tempDir, "note1.md")
			mdFile2 := filepath.Join(tempDir, "note2.md")
			txtFile := filepath.Join(tempDir, "file.txt")

			Expect(os.WriteFile(mdFile1, []byte("content"), 0600)).To(Succeed())
			Expect(os.WriteFile(mdFile2, []byte("content"), 0600)).To(Succeed())
			Expect(os.WriteFile(txtFile, []byte("content"), 0600)).To(Succeed())

			files, err := s.Scan(ctx, tempDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(HaveLen(2))
			Expect(files).To(ContainElements(mdFile1, mdFile2))
		})

		It("returns .md files in nested directories", func() {
			// Create nested structure
			subDir := filepath.Join(tempDir, "folder1", "folder2")
			Expect(os.MkdirAll(subDir, 0755)).To(Succeed())

			mdFile1 := filepath.Join(tempDir, "root.md")
			mdFile2 := filepath.Join(tempDir, "folder1", "nested.md")
			mdFile3 := filepath.Join(subDir, "deep.md")

			Expect(os.WriteFile(mdFile1, []byte("content"), 0600)).To(Succeed())
			Expect(os.WriteFile(mdFile2, []byte("content"), 0600)).To(Succeed())
			Expect(os.WriteFile(mdFile3, []byte("content"), 0600)).To(Succeed())

			files, err := s.Scan(ctx, tempDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(HaveLen(3))
			Expect(files).To(ContainElements(mdFile1, mdFile2, mdFile3))
		})

		It("excludes non-.md files", func() {
			// Create various file types
			mdFile := filepath.Join(tempDir, "note.md")
			txtFile := filepath.Join(tempDir, "file.txt")
			pngFile := filepath.Join(tempDir, "image.png")

			Expect(os.WriteFile(mdFile, []byte("content"), 0600)).To(Succeed())
			Expect(os.WriteFile(txtFile, []byte("content"), 0600)).To(Succeed())
			Expect(os.WriteFile(pngFile, []byte("content"), 0600)).To(Succeed())

			files, err := s.Scan(ctx, tempDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(HaveLen(1))
			Expect(files[0]).To(Equal(mdFile))
		})

		It("returns empty slice when no .md files exist", func() {
			txtFile := filepath.Join(tempDir, "file.txt")
			Expect(os.WriteFile(txtFile, []byte("content"), 0600)).To(Succeed())

			files, err := s.Scan(ctx, tempDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(BeEmpty())
		})

		It("returns error for non-existent directory", func() {
			nonExistentPath := filepath.Join(tempDir, "does-not-exist")

			files, err := s.Scan(ctx, nonExistentPath)
			Expect(err).To(HaveOccurred())
			Expect(files).To(BeNil())
		})
	})
})
