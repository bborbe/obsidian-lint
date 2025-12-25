// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package resolver_test

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/obsidian-lint/pkg/index"
	"github.com/bborbe/obsidian-lint/pkg/model"
	"github.com/bborbe/obsidian-lint/pkg/parser"
	"github.com/bborbe/obsidian-lint/pkg/resolver"
)

var _ = Describe("Resolver", func() {
	var (
		ctx     context.Context
		r       resolver.Resolver
		idx     *index.VaultIndex
		tempDir string
		err     error
	)

	BeforeEach(func() {
		ctx = context.Background()
		r = resolver.New()

		tempDir, err = os.MkdirTemp("", "resolver-test")
		Expect(err).NotTo(HaveOccurred())

		// Create test vault
		note1 := filepath.Join(tempDir, "Note1.md")
		note2 := filepath.Join(tempDir, "Note2.md")
		noteWithAlias := filepath.Join(tempDir, "AliasNote.md")

		Expect(os.WriteFile(note1, []byte("Content"), 0600)).To(Succeed())
		Expect(os.WriteFile(note2, []byte("Content"), 0600)).To(Succeed())
		Expect(os.WriteFile(noteWithAlias, []byte(`---
aliases: [MyAlias, Another Alias]
---
Content`), 0600)).To(Succeed())

		// Build index
		p := parser.New()
		builder := index.New(p)
		idx, err = builder.Build(ctx, tempDir, []string{note1, note2, noteWithAlias})
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if tempDir != "" {
			_ = os.RemoveAll(tempDir)
		}
	})

	Context("Resolve", func() {
		It("resolves link to existing note", func() {
			link := &model.Link{
				Target: "Note1",
			}

			exists := r.Resolve(ctx, link, idx)
			Expect(exists).To(BeTrue())
		})

		It("resolves link case-insensitively", func() {
			link := &model.Link{
				Target: "note1",
			}

			exists := r.Resolve(ctx, link, idx)
			Expect(exists).To(BeTrue())
		})

		It("resolves link via alias", func() {
			link := &model.Link{
				Target: "MyAlias",
			}

			exists := r.Resolve(ctx, link, idx)
			Expect(exists).To(BeTrue())
		})

		It("resolves link via alias case-insensitively", func() {
			link := &model.Link{
				Target: "another alias",
			}

			exists := r.Resolve(ctx, link, idx)
			Expect(exists).To(BeTrue())
		})

		It("returns false for non-existent target", func() {
			link := &model.Link{
				Target: "DoesNotExist",
			}

			exists := r.Resolve(ctx, link, idx)
			Expect(exists).To(BeFalse())
		})

		It("ignores heading when resolving", func() {
			link := &model.Link{
				Target:  "Note1",
				Heading: "SomeHeading",
			}

			// Should return true even though we don't validate heading
			exists := r.Resolve(ctx, link, idx)
			Expect(exists).To(BeTrue())
		})

		It("ignores alias field when resolving", func() {
			link := &model.Link{
				Target: "Note1",
				Alias:  "Display Text",
			}

			exists := r.Resolve(ctx, link, idx)
			Expect(exists).To(BeTrue())
		})
	})
})
