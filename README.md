# shape-xml

Production-ready XML parser for the Shape ecosystem.

[![Go Version](https://img.shields.io/badge/go-1.23+-blue.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

## Overview

shape-xml is a complete XML parser that integrates with the Shape ecosystem's universal AST representation. It provides high-performance parsing, validation, and transformation capabilities for XML documents.

## Features

- **Dual-Path Parser Pattern**
  - Fast validation path (4-5x faster, no AST construction)
  - Full parsing path (complete AST generation)

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
  - Comprehensive test coverage (64%+)
  - Fuzz testing
  - Benchmark suite
  - Zero external dependencies (except shape-core)

## Installation

```bash
go get github.com/shapestone/shape-xml
```

## Quick Start

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
// Fast validation without AST construction
if err := xml.Validate(`<root><child>value</child></root>`); err != nil {
    log.Fatal("Invalid XML:", err)
}
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

Benchmark results on typical XML documents:

- **Small XML (50 bytes)**: ~6.6 µs/op
- **Medium XML (1KB)**: ~127 µs/op (7.97 MB/s)
- **Large XML (340KB)**: ~40.8 ms/op (8.35 MB/s)
- **Validation (fast path)**: ~140 ns/op for small XML

Run benchmarks:

```bash
make bench
```

## Testing

```bash
# Run all tests
make test

# Generate coverage report
make coverage

# Run fuzz tests
make fuzz
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

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

Apache License 2.0

Copyright 2020-2025 Shapestone

See [LICENSE](LICENSE) for details.

## Part of Shape Ecosystem

shape-xml is part of the [Shape](https://github.com/shapestone/shape-core) ecosystem:

- [shape-core](https://github.com/shapestone/shape-core) - Universal AST and validation framework
- [shape-json](https://github.com/shapestone/shape-json) - JSON parser
- [shape-yaml](https://github.com/shapestone/shape-yaml) - YAML parser
- [shape-xml](https://github.com/shapestone/shape-xml) - XML parser (this project)
