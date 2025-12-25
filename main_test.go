// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/bborbe/obsidian-lint/pkg/formatter"
	"github.com/bborbe/obsidian-lint/pkg/index"
	"github.com/bborbe/obsidian-lint/pkg/parser"
	"github.com/bborbe/obsidian-lint/pkg/resolver"
	"github.com/bborbe/obsidian-lint/pkg/scanner"
	"github.com/bborbe/obsidian-lint/pkg/validator"
)

var _ = Describe("Main", func() {
	It("Compiles", func() {
		var err error
		_, err = gexec.Build(".", "-mod=mod")
		Expect(err).NotTo(HaveOccurred())
	})
})

var _ = Describe("Integration", func() {
	var (
		ctx     context.Context
		tempDir string
		err     error
	)

	BeforeEach(func() {
		ctx = context.Background()
		tempDir, err = os.MkdirTemp("", "integration-test")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if tempDir != "" {
			_ = os.RemoveAll(tempDir)
		}
	})

	It("detects broken links in a real vault", func() {
		// Create vault structure
		note1 := filepath.Join(tempDir, "ValidNote.md")
		note2 := filepath.Join(tempDir, "BrokenLinks.md")
		note3 := filepath.Join(tempDir, "WithAlias.md")
		image := filepath.Join(tempDir, "image.png")

		Expect(os.WriteFile(note1, []byte("This is a valid note."), 0600)).To(Succeed())
		Expect(os.WriteFile(note2, []byte(`# Broken Links Test

This links to [[ValidNote]] which exists.
This links to [[DoesNotExist]] which is broken.
This embeds ![[image.png]] which exists.
This embeds ![[missing.png]] which is broken.
This links to [[MyAlias]] which resolves via alias.
This links to [[ValidNote#Heading]] which is valid (heading not checked).
`), 0600)).To(Succeed())
		Expect(os.WriteFile(note3, []byte(`---
aliases: [MyAlias, Another]
---
Content here.`), 0600)).To(Succeed())
		Expect(os.WriteFile(image, []byte("fake image data"), 0600)).To(Succeed())

		// Run validation
		s := scanner.New()
		p := parser.New()
		b := index.New(p)
		r := resolver.New()
		v := validator.New(s, p, b, r)

		result, err := v.Validate(ctx, tempDir)
		Expect(err).NotTo(HaveOccurred())

		// Verify results
		Expect(result.BrokenLinks).To(HaveLen(1))
		Expect(result.BrokenLinks[note2]).To(HaveLen(2))

		brokenLinks := result.BrokenLinks[note2]
		Expect(brokenLinks[0].Link).To(Equal("[[DoesNotExist]]"))
		Expect(brokenLinks[0].Line).To(Equal(4))
		Expect(brokenLinks[1].Link).To(Equal("![[missing.png]]"))
		Expect(brokenLinks[1].Line).To(Equal(6))

		// Test text formatter
		textFormatter := formatter.NewTextFormatter()
		textOutput, err := textFormatter.Format(ctx, result)
		Expect(err).NotTo(HaveOccurred())
		Expect(textOutput).To(ContainSubstring("Broken links found"))
		Expect(textOutput).To(ContainSubstring("BrokenLinks.md"))
		Expect(textOutput).To(ContainSubstring("[[DoesNotExist]]"))
		Expect(textOutput).To(ContainSubstring("![[missing.png]]"))

		// Test JSON formatter
		jsonFormatter := formatter.NewJSONFormatter()
		jsonOutput, err := jsonFormatter.Format(ctx, result)
		Expect(err).NotTo(HaveOccurred())
		Expect(jsonOutput).To(ContainSubstring("DoesNotExist"))
		Expect(jsonOutput).To(ContainSubstring("missing.png"))
	})

	It("returns no broken links for valid vault", func() {
		// Create vault with only valid links
		note1 := filepath.Join(tempDir, "Note1.md")
		note2 := filepath.Join(tempDir, "Note2.md")

		Expect(os.WriteFile(note1, []byte("Link to [[Note2]]"), 0600)).To(Succeed())
		Expect(os.WriteFile(note2, []byte("Link to [[Note1]]"), 0600)).To(Succeed())

		// Run validation
		s := scanner.New()
		p := parser.New()
		b := index.New(p)
		r := resolver.New()
		v := validator.New(s, p, b, r)

		result, err := v.Validate(ctx, tempDir)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.BrokenLinks).To(BeEmpty())

		// Test formatter
		f := formatter.NewTextFormatter()
		output, err := f.Format(ctx, result)
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(Equal("No broken links found.\n"))
	})
})

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Main Suite")
}
