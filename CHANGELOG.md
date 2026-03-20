# Changelog

All notable changes to this project will be documented in this file.

Please choose versions by [Semantic Versioning](http://semver.org/).

* MAJOR version when you make incompatible API changes,
* MINOR version when you add functionality in a backwards-compatible manner, and
* PATCH version when you make backwards-compatible bug fixes.

## v0.2.1

- chore: verify project health — all tests pass, linting clean, no vulnerabilities

## v0.2.0

- Update Go to 1.26.0
- Update GitHub Actions workflows to use checkout@v6
- Switch Claude Code review workflow to label-triggered activation
- Update multiple dependency versions including bborbe packages, Google OSV scanner, and golang.org/x packages
- Upgrade golangci-lint from v1 to v2
- Standardize Makefile: add .PHONY declarations, multiline trivy, mocks mkdir
- Update .golangci.yml to v2 format
- Setup dark-factory config

## v0.1.3

- Update GitHub workflows to v1 plugin system
- Simplify Claude Code action with inline conditions
- Add ready_for_review and reopened triggers

## v0.1.2

- Update Go from 1.25.5 to 1.25.6
- Update github.com/bborbe/* dependencies to latest versions
- Update github.com/onsi/ginkgo/v2 from v2.26.0 to v2.28.1
- Update github.com/onsi/gomega from v1.38.2 to v1.39.1
- Update golang.org/x toolchain dependencies

## v0.1.1

- Remove unused Gemini CLI GitHub Actions workflows and commands

## v0.1.0

- Add CLI tool for detecting broken wiki links in Obsidian vaults
- Add support for all Obsidian link formats: [[Note]], [[Note|alias]], [[Note#Heading]], ![[embeds]]
- Add case-insensitive link resolution with alias support from YAML frontmatter
- Add validation for embedded files (images, PDFs, etc.)
- Add text and JSON output formats
- Add comprehensive test suite with 81-100% coverage
- Add exit codes: 0 (no broken links), 1 (broken links found)
