# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.9.0] - 2025-12-29

### Added
- Full-featured XML parser with universal AST support
- Dual-path parser pattern (fast validation + full AST construction)
- Streaming parser via ParseReader for large files
- Fluent DOM API for programmatic XML construction
- Round-trip fidelity (Parse/Render/Marshal/Unmarshal)
- XML conventions: @ for attributes, #text, #cdata
- Namespace support
- Comprehensive test coverage (80.0%+)
- Grammar documentation (docs/grammar/xml.ebnf)
- GitHub Actions CI/CD pipeline
- Grammar verification tests (ADR 0005)
- Rune matcher tests for Unicode/emoji support
- Fastparser unmarshal tests
- Fuzz testing
- Benchmark suite (1.7-2.9x faster than stdlib)
- Thread-safe concurrent operations

### Changed
- Validate() now uses internal fast parser instead of encoding/xml
- Improved test coverage from initial 55% to 80%+

### Fixed
- Array detection logic in convert.go
- Unused variable cleanup in tokenizer

[Unreleased]: https://github.com/shapestone/shape-xml/compare/v0.9.0...HEAD
[0.9.0]: https://github.com/shapestone/shape-xml/releases/tag/v0.9.0
