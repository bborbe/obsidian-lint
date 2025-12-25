# Changelog

All notable changes to this project will be documented in this file.

Please choose versions by [Semantic Versioning](http://semver.org/).

* MAJOR version when you make incompatible API changes,
* MINOR version when you add functionality in a backwards-compatible manner, and
* PATCH version when you make backwards-compatible bug fixes.

## v0.1.0

- Add CLI tool for detecting broken wiki links in Obsidian vaults
- Add support for all Obsidian link formats: [[Note]], [[Note|alias]], [[Note#Heading]], ![[embeds]]
- Add case-insensitive link resolution with alias support from YAML frontmatter
- Add validation for embedded files (images, PDFs, etc.)
- Add text and JSON output formats
- Add comprehensive test suite with 81-100% coverage
- Add exit codes: 0 (no broken links), 1 (broken links found)
