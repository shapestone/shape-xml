# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.9.0] - 2025-12-26

### Added
- Full-featured XML parser with universal AST support
- Dual-path parser pattern (fast validation + full AST construction)
- Streaming parser via ParseReader for large files
- Fluent DOM API for programmatic XML construction
- Round-trip fidelity (Parse/Render/Marshal/Unmarshal)
- XML conventions: @ for attributes, #text, #cdata
- Namespace support
- Comprehensive test coverage (64%+)
- Fuzz testing
- Benchmark suite
- Thread-safe concurrent operations

### Changed
- Validate() now uses internal fast parser instead of encoding/xml

[Unreleased]: https://github.com/shapestone/shape-xml/compare/v0.9.0...HEAD
[0.9.0]: https://github.com/shapestone/shape-xml/releases/tag/v0.9.0
