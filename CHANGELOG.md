# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- `xml.Validate` and `xml.ValidateReader` now use a dedicated validator that aims to give the same verdicts as a token loop over `encoding/xml`'s Decoder (Strict mode, no CharsetReader). It is checked by a differential test against `encoding/xml` (seed documents, random fragments and mutations, including colons, quoted attributes and `]]>`) and by `FuzzValidate`; that is strong evidence, not a proof of identical verdicts for every input.
  - **Now rejected:** unknown entities (`&bogus;`; only `lt`, `gt`, `amp`, `apos`, `quot` exist), invalid character references (`&#0;`, `&#x1;`, out-of-range values, and surrogates such as `&#xD800;`), characters XML forbids (NUL and other control characters, U+FFFE, U+FFFF) and invalid UTF-8 in text, CDATA and attribute values, `<` in an attribute value, an unescaped `]]>` in element content (it is allowed inside an attribute value), element and attribute names with more than one colon (`<a:b:c>`), an XML declaration with a version other than 1.0, and an XML declaration naming an encoding other than UTF-8 (`ISO-8859-1`, `UTF-16`; there is no charset hook).
  - **Now accepted:** a leading UTF-8 byte order mark, and a DOCTYPE. The DOCTYPE, including any internal subset, is skipped without being processed. Entities it declares are not defined, parameter entities are not expanded, and `SYSTEM`/`PUBLIC` identifiers are never read, so XXE and entity-expansion payloads cannot read files, use the network or blow up memory. A reference to an entity declared in a DOCTYPE is an unknown entity and is rejected.
  - **Behaviour change to check before upgrading: empty input, whitespace-only input, several root elements and text outside the root element are now accepted.** Earlier versions rejected all of them. `encoding/xml`'s token loop accepts all three, so Validate now follows it. A caller that used `Validate` to demand exactly one root element will get false accepts and must check that itself.
  - **Deviations from `encoding/xml`:** a character reference to a surrogate (`&#xD800;` to `&#xDFFF;`) is rejected, while `encoding/xml` maps it to U+FFFD and accepts it; XML 1.0 forbids surrogates. Error messages differ: they carry `line N, column M` instead of a byte offset.
  - Unchanged and matching `encoding/xml`: duplicate attributes and undeclared namespace prefixes are accepted; mismatched or unclosed tags and unquoted attributes are rejected.
  - Deeply nested input no longer risks a stack overflow in `Validate`: it keeps a heap-allocated stack of element names, 16 bytes per open element (a 15 MB input of `<a>` allocated about 417 MB in total).
- `xml.Validate` does not limit input size or nesting depth; limit untrusted input before calling it.

### Added
- `FuzzValidate` in `internal/fastparser`, comparing `Validate` with `encoding/xml`.
- A test that every non-test file in `internal/fastparser` imports only pure packages (no `os`, `net`, `io/ioutil`, `encoding/xml`), so no file or network access can be added unnoticed.

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
