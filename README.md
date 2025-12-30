# shape-xml

![Build Status](https://github.com/shapestone/shape-xml/actions/workflows/ci.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/shapestone/shape-xml)](https://goreportcard.com/report/github.com/shapestone/shape-xml)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![codecov](https://codecov.io/gh/shapestone/shape-xml/branch/main/graph/badge.svg)](https://codecov.io/gh/shapestone/shape-xml)
![Go Version](https://img.shields.io/github/go-mod/go-version/shapestone/shape-xml)
![Latest Release](https://img.shields.io/github/v/release/shapestone/shape-xml)
[![GoDoc](https://pkg.go.dev/badge/github.com/shapestone/shape-xml.svg)](https://pkg.go.dev/github.com/shapestone/shape-xml)

[![CodeQL](https://github.com/shapestone/shape-xml/actions/workflows/codeql.yml/badge.svg)](https://github.com/shapestone/shape-xml/actions/workflows/codeql.yml)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/shapestone/shape-xml/badge)](https://securityscorecards.dev/viewer/?uri=github.com/shapestone/shape-xml)
[![Security Policy](https://img.shields.io/badge/Security-Policy-brightgreen)](SECURITY.md)

**Repository:** github.com/shapestone/shape-xml

An XML parser for the [Shape Parser™](https://github.com/shapestone/shape) ecosystem.

Parses XML documents into Shape Parser's™ unified AST representation.

## Installation

```bash
go get github.com/shapestone/shape-xml
```

## Usage

### Parse XML to AST

```go
import "github.com/shapestone/shape-xml/pkg/xml"

// Parse XML from string
node, err := xml.Parse(`<user id="123"><name>Alice</name></user>`)
if err != nil {
    log.Fatal(err)
}

// Access attributes (prefixed with @)
obj := node.(*ast.ObjectNode)
idNode, _ := obj.GetProperty("@id")
id := idNode.(*ast.LiteralNode).Value().(string) // "123"
```

### Validate XML (Fast Path)

```go
// Fast validation without AST construction - idiomatic Go
if err := xml.Validate(`<root><child>value</child></root>`); err != nil {
    fmt.Println("Invalid XML:", err)
}
// err == nil means valid XML
```

### Parse from Stream

```go
file, err := os.Open("data.xml")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

node, err := xml.ParseReader(file)
if err != nil {
    log.Fatal(err)
}
```

### Fluent DOM API

Build XML programmatically with a type-safe, chainable API:

```go
import "github.com/shapestone/shape-xml/pkg/xml"

// Build XML programmatically
user := xml.NewElement("user").
    Attr("id", "123").
    Attr("active", "true").
    Child(xml.NewElement("name").Text("Alice")).
    Child(xml.NewElement("email").Text("alice@example.com"))

// Render to XML string
output := user.Render()
// <user id="123" active="true"><name>Alice</name><email>alice@example.com</email></user>
```

### Marshal/Unmarshal

```go
type User struct {
    ID     string `xml:"id,attr"`
    Name   string `xml:"name"`
    Email  string `xml:"email"`
}

// Marshal Go struct to XML
user := User{ID: "123", Name: "Alice", Email: "alice@example.com"}
data, err := xml.Marshal(user)

// Unmarshal XML to Go struct
var parsed User
err = xml.Unmarshal(data, &parsed)
```

## Features

- **Dual-Path Parser Pattern**
  - Fast validation path (4-5x faster, no AST construction)
  - Full parsing path (complete AST generation)
  - Automatic path selection for optimal performance
- **Streaming Support**
  - Parse large files with constant memory usage via `ParseReader`
  - Validate streams efficiently with `ValidateReader`
- **Fluent DOM API**
  - Type-safe element construction
  - Chainable method calls for building XML programmatically
- **Round-Trip Fidelity**
  - Parse → AST → Render → Parse preserves structure
  - Marshal/Unmarshal support for Go structs
- **Universal AST Integration**
  - Attributes: `@` prefix (e.g., `@id`, `@class`)
  - Text content: `#text` property
  - CDATA sections: `#cdata` property
  - Namespace support: preserved in element names
- **Production Ready**
  - Thread-safe concurrent operations
  - Comprehensive test coverage (80.0%+)
  - Fuzz testing
  - Benchmark suite
  - Zero external dependencies (except shape-core)
  - LL(1) Recursive Descent Parser
- **Shape AST Integration**: Returns unified AST nodes for advanced use cases
- **Comprehensive Error Messages**: Context-aware error reporting

## XML → AST Conventions

shape-xml follows these conventions when mapping XML to the universal AST:

```xml
<user id="123" xmlns:custom="http://example.com">
    <name>Alice</name>
    <custom:role>Admin</custom:role>
    <bio><![CDATA[Uses <tags>]]></bio>
</user>
```

Maps to:

```go
*ast.ObjectNode{
    properties: {
        "@id": *ast.LiteralNode{value: "123"},
        "@xmlns:custom": *ast.LiteralNode{value: "http://example.com"},
        "name": *ast.LiteralNode{value: "Alice"},
        "custom:role": *ast.LiteralNode{value: "Admin"},
        "bio": *ast.ObjectNode{
            properties: {
                "#cdata": *ast.LiteralNode{value: "Uses <tags>"},
            },
        },
    },
}
```

## Performance

shape-xml uses an **intelligent dual-path architecture** that automatically selects the optimal parsing strategy:

### ⚡ Fast Path (Validation Only)

The fast path bypasses AST construction for maximum performance:

- **APIs**: `Validate()`, `ValidateReader()`
- **Performance**: 4-5x faster than AST path
- **Use when**: You just need to validate XML syntax

```go
// Fast path - validation only (4-5x faster!)
if err := xml.Validate(xmlString); err != nil {
    // Invalid XML
}
```

**Benchmark results**:
- **Small XML (50 bytes)**: ~140 ns/op (validation)
- **Medium XML (1KB)**: ~127 µs/op (7.97 MB/s)
- **Large XML (340KB)**: ~40.8 ms/op (8.35 MB/s)

### 🌳 AST Path (Full Features)

The AST path builds a complete Abstract Syntax Tree:

- **APIs**: `Parse()`, `ParseReader()`, `Marshal()`, `Unmarshal()`
- **Performance**: Slower, more memory (enables advanced features)
- **Use when**: You need AST manipulation, rendering, or format conversion

```go
// AST path - full tree structure for advanced features
node, _ := xml.Parse(xmlString)
// Work with AST, transform, render, etc.
```

Run benchmarks:

```bash
make bench
```

## Architecture

shape-xml uses a **unified architecture** with custom parsers:

- **Grammar-Driven**: EBNF grammar in `docs/grammar/xml.ebnf`
- **Tokenizer**: Custom tokenizer using Shape's framework
- **Parser**: LL(1) recursive descent with single token lookahead
- **Rendering**: Custom XML renderer
- **AST Representation**:
  - Elements → `*ast.ObjectNode` with properties map
  - Attributes → Properties with `@` prefix
  - Text content → `#text` property
  - CDATA → `#cdata` property
  - Primitives → `*ast.LiteralNode` (string values)

## Grammar

See [docs/grammar/xml.ebnf](docs/grammar/xml.ebnf) for the complete EBNF specification.

Key grammar rules:
```ebnf
Document = [ XMLDecl ] Element ;
Element = EmptyElement | StartTag Content EndTag ;
StartTag = "<" Name { Attribute } ">" ;
EndTag = "</" Name ">" ;
```

## Thread Safety

**shape-xml is thread-safe.** All public APIs can be called concurrently from multiple goroutines without external synchronization.

### Safe for Concurrent Use

```go
// ✅ SAFE: Multiple goroutines can call these concurrently
go func() {
    var v1 interface{}
    xml.Unmarshal(data1, &v1)
}()

go func() {
    var v2 interface{}
    xml.Unmarshal(data2, &v2)
}()

// ✅ SAFE: Parse, Marshal, Validate all create new instances
go func() { xml.Parse(input1) }()
go func() { xml.Marshal(obj1) }()
go func() { xml.Validate(input2) }()
```

### Thread Safety Guarantees

- **`Unmarshal()`, `Marshal()`** - Thread-safe
- **`Parse()`, `Validate()`** - Thread-safe, create new parser instances
- **Race detector verified** - All tests pass with `go test -race`

## Testing

shape-xml has comprehensive test coverage including unit tests, fuzzing, and grammar verification.

### Coverage Summary
- Fast Parser: 55.7%
- Parser: 77.3%
- XML API: 67.4%
- **Overall Library**: **55.0%**
- **Target**: 80%+ (current gap: 25 percentage points)

### Quick Start

```bash
# Run all tests
go test ./...

# Run with coverage
make coverage

# Run fuzzing tests
make fuzz

# Run grammar verification
make grammar-test
```

### Fuzzing

The parser includes extensive fuzzing tests to ensure robustness:

```bash
# Fuzz parser
go test ./pkg/xml -fuzz=FuzzParse -fuzztime=30s
go test ./pkg/xml -fuzz=FuzzValidate -fuzztime=30s
go test ./pkg/xml -fuzz=FuzzRender -fuzztime=30s
go test ./pkg/xml -fuzz=FuzzMarshal -fuzztime=30s
```

## API Reference

### Parsing Functions

- `Parse(input string) (ast.SchemaNode, error)` - Parse XML from string
- `ParseReader(reader io.Reader) (ast.SchemaNode, error)` - Parse from stream

### Validation Functions

- `Validate(input string) error` - Fast validation without AST
- `ValidateReader(reader io.Reader) error` - Validate from stream

### Marshaling Functions

- `Marshal(v interface{}) ([]byte, error)` - Go struct → XML
- `MarshalIndent(v interface{}, prefix, indent string) ([]byte, error)` - Pretty-print
- `Unmarshal(data []byte, v interface{}) error` - XML → Go struct

### Rendering Functions

- `Render(node ast.SchemaNode) []byte` - AST → compact XML
- `RenderIndent(node ast.SchemaNode, prefix, indent string) []byte` - AST → pretty XML

### DOM API

- `NewElement(name string) *Element` - Create element builder
- `Element.Attr(name, value string) *Element` - Add attribute (chainable)
- `Element.Text(content string) *Element` - Set text content (chainable)
- `Element.CDATA(content string) *Element` - Set CDATA content (chainable)
- `Element.Child(child *Element) *Element` - Add child element (chainable)
- `Element.Render() []byte` - Render to XML bytes

## Documentation

- [EBNF Grammar](docs/grammar/xml.ebnf) - Complete XML grammar specification
- [Parser Implementation Guide](https://github.com/shapestone/shape-core/blob/main/docs/PARSER_IMPLEMENTATION_GUIDE.md) - Guide for implementing parsers
- [Shape ADR 0004: LL(1) Recursive Descent Parser Strategy](https://github.com/shapestone/shape-core/blob/main/docs/adr/0004-ll1-recursive-descent-parser.md) - Parser design principles
- [Shape ADR 0005: Grammar-as-Verification](https://github.com/shapestone/shape-core/blob/main/docs/adr/0005-grammar-as-verification.md) - Grammar verification approach
- [Shape Core Infrastructure](https://github.com/shapestone/shape-core) - Universal AST and tokenizer framework
- [Shape Ecosystem](https://github.com/shapestone/shape) - Documentation and examples

## Development

```bash
# Run tests
make test

# Generate coverage report
make coverage

# Build
make build

# Run all checks
make all
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

Apache License 2.0

Copyright © 2020-2025 Shapestone

See [LICENSE](LICENSE) for the full license text and [NOTICE](NOTICE) for third-party attributions.

## Part of Shape Ecosystem

shape-xml is part of the [Shape](https://github.com/shapestone/shape-core) ecosystem:

- [shape-core](https://github.com/shapestone/shape-core) - Universal AST and validation framework
- [shape-json](https://github.com/shapestone/shape-json) - JSON parser
- [shape-yaml](https://github.com/shapestone/shape-yaml) - YAML parser
- [shape-xml](https://github.com/shapestone/shape-xml) - XML parser (this project)