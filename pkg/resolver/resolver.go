// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package resolver

import (
	"context"

	"github.com/bborbe/obsidian-lint/pkg/index"
	"github.com/bborbe/obsidian-lint/pkg/model"
)

//counterfeiter:generate -o ../../mocks/resolver.go --fake-name Resolver . Resolver

// Resolver resolves wiki links against a vault index
type Resolver interface {
	Resolve(ctx context.Context, link *model.Link, index *index.VaultIndex) bool
}

// New creates a new Resolver
func New() Resolver {
	return &resolver{}
}

type resolver struct{}

// Resolve checks if a link target exists in the vault index
func (r *resolver) Resolve(ctx context.Context, link *model.Link, index *index.VaultIndex) bool {
	// We only check if the target note/file exists
	// We don't validate headings (per user requirement)
	return index.Resolve(link.Target)
}
