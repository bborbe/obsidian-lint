// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"os"

	libsentry "github.com/bborbe/sentry"
	"github.com/bborbe/service"

	"github.com/bborbe/obsidian-lint/pkg/formatter"
	"github.com/bborbe/obsidian-lint/pkg/index"
	"github.com/bborbe/obsidian-lint/pkg/parser"
	"github.com/bborbe/obsidian-lint/pkg/resolver"
	"github.com/bborbe/obsidian-lint/pkg/scanner"
	"github.com/bborbe/obsidian-lint/pkg/validator"
)

func main() {
	app := &application{}
	os.Exit(service.Main(context.Background(), app, &app.SentryDSN, &app.SentryProxy))
}

type application struct {
	SentryDSN   string `required:"false" arg:"sentry-dsn"   env:"SENTRY_DSN"   usage:"SentryDSN (optional)"      display:"length"`
	SentryProxy string `required:"false" arg:"sentry-proxy" env:"SENTRY_PROXY" usage:"Sentry Proxy"`
	Vault       string `required:"true"  arg:"vault"        env:"VAULT"        usage:"vault directory path"`
	Format      string `required:"false" arg:"format"       env:"FORMAT"       usage:"output format (text|json)" default:"text"`
}

func (a *application) Run(ctx context.Context, sentryClient libsentry.Client) error {
	// Build dependencies
	s := scanner.New()
	p := parser.New()
	b := index.New(p)
	r := resolver.New()
	v := validator.New(s, p, b, r)

	// Validate vault
	result, err := v.Validate(ctx, a.Vault)
	if err != nil {
		return err
	}

	// Format output
	var f formatter.Formatter
	switch a.Format {
	case "json":
		f = formatter.NewJSONFormatter()
	case "text":
		f = formatter.NewTextFormatter()
	default:
		return fmt.Errorf("invalid format: %s (must be 'text' or 'json')", a.Format)
	}

	output, err := f.Format(ctx, result)
	if err != nil {
		return err
	}

	// Print output
	fmt.Print(output)

	// Exit with non-zero if broken links found
	if len(result.BrokenLinks) > 0 {
		os.Exit(1)
	}

	return nil
}
