# Shape-XML Implementation Plan

**Status:** Planning
**Goal:** Transform shape-xml from minimal validator to full-featured XML parser aligned with shape-core ecosystem
**Target Completion:** TBD

---

## Executive Summary

Shape-xml is currently a minimal 18-line validator wrapping Go's `encoding/xml`. This plan outlines transforming it into a production-ready XML parser that:
- Returns universal AST (aligned with shape-core)
- Implements XML conventions (attributes, namespaces, text content)
- Provides dual API (Parse→AST, Unmarshal→Go types)
- Supports XPath queries
- Handles XML-specific features (CDATA, processing instructions, etc.)

---

## Current State Analysis

### What Exists
- ✅ Basic repository structure (`pkg/xml/`)
- ✅ Single `Validate()` function
- ✅ Basic tests
- ✅ README with minimal documentation

### What's Missing
- ❌ Git repository initialization
- ❌ shape-core dependency
- ❌ Tokenizer implementation
- ❌ Parser returning AST
- ❌ XML-specific features (attributes, namespaces, CDATA, etc.)
- ❌ Dual API (Parse + Unmarshal)
- ❌ XPath query engine
- ❌ Comprehensive tests
- ❌ Grammar documentation (EBNF)
- ❌ CI/CD pipeline

---

## Architecture Overview

### Design Principles

Following shape-core PARSER_IMPLEMENTATION_GUIDE.md and AST_CONVENTIONS.md:

1. **Universal AST** - Use shape-core's format-agnostic AST
2. **XML Conventions** - Follow established conventions:
   - `@prefix` for attributes (e.g., `@id`, `@class`)
   - `#text` for element text content
   - `#cdata` for CDATA sections
   - `ns:element` for namespaced elements
   - `@xmlns:prefix` for namespace declarations
3. **Dual API** - Primary (`Parse`) returns AST, secondary (`Unmarshal`) returns Go types
4. **Position Tracking** - Preserve source positions for error messages

### XML → AST Mapping

```xml
<user id="123" xmlns:custom="http://example.com">
    <name>Alice</name>
    <custom:role>Admin</custom:role>
    <bio><![CDATA[Uses <tags>]]></bio>
</user>
```

**Maps to:**
```go
*ast.ObjectNode{
    properties: {
        "@id": *ast.LiteralNode{value: "123"},
        "@xmlns:custom": *ast.LiteralNode{value: "http://example.com"},
        "name": *ast.ObjectNode{
            properties: {
                "#text": *ast.LiteralNode{value: "Alice"},
            }
        },
        "custom:role": *ast.ObjectNode{
            properties: {
                "#text": *ast.LiteralNode{value: "Admin"},
            }
        },
        "bio": *ast.ObjectNode{
            properties: {
                "#cdata": *ast.LiteralNode{value: "Uses <tags>"},
            }
        }
    }
}
```

---

## Implementation Phases

### Phase 1: Foundation & Setup

**Goal:** Establish project infrastructure
**Duration:** 1-2 days

#### Tasks

1. **Initialize Git Repository**
   ```bash
   git init
   git remote add origin git@github.com:shapestone/shape-xml.git
   ```

2. **Update go.mod**
   ```go
   module github.com/shapestone/shape-xml

   go 1.25

   replace github.com/shapestone/shape-core => ../shape-core

   require github.com/shapestone/shape-core v0.9.2
   ```

3. **Create Project Structure**
   ```
   shape-xml/
   ├── README.md
   ├── LICENSE
   ├── go.mod
   ├── go.sum
   ├── Makefile
   ├── .github/
   │   └── workflows/
   │       └── ci.yml
   ├── docs/
   │   ├── grammar/
   │   │   └── xml.ebnf
   │   └── examples/
   ├── pkg/
   │   └── xml/
   │       ├── parser.go      # Public API
   │       ├── parser_test.go
   │       ├── unmarshal.go   # Go struct unmarshaling
   │       ├── marshal.go     # Go struct marshaling
   │       └── render.go      # AST → XML string
   ├── internal/
   │   ├── tokenizer/
   │   │   ├── tokenizer.go
   │   │   └── tokenizer_test.go
   │   └── parser/
   │       ├── parser.go      # Core parser logic
   │       └── parser_test.go
   └── examples/
       └── main.go
   ```

4. **Create Makefile**
   ```makefile
   .PHONY: test lint build all

   test:
       go test -v -race ./...

   lint:
       golangci-lint run

   build:
       go build ./...

   coverage:
       go test -coverprofile=coverage.out ./...
       go tool cover -html=coverage.out -o coverage.html

   all: test lint build
   ```

5. **Create Basic CI/CD Pipeline**
   - GitHub Actions workflow for tests
   - Multi-platform testing (Ubuntu, macOS, Windows)
   - Code coverage reporting

#### Deliverables
- ✅ Git repository initialized
- ✅ Project structure created
- ✅ Makefile with standard targets
- ✅ CI/CD pipeline configured

---

### Phase 2: XML Grammar & Tokenizer

**Goal:** Define XML grammar and implement tokenizer
**Duration:** 2-3 days

#### Tasks

1. **Document XML Grammar (EBNF)**

Create `docs/grammar/xml.ebnf`:

```ebnf
(* Simplified XML Grammar - W3C XML 1.0 compliant *)

(* Document *)
Document = Prolog Element Misc* ;

Prolog = XMLDecl? Misc* (Doctypedecl Misc*)? ;

XMLDecl = "<?xml" VersionInfo EncodingDecl? SDDecl? S? "?>" ;

(* Elements *)
Element = EmptyElemTag | STag Content ETag ;

STag = "<" Name (S Attribute)* S? ">" ;
ETag = "</" Name S? ">" ;
EmptyElemTag = "<" Name (S Attribute)* S? "/>" ;

(* Attributes *)
Attribute = Name Eq AttValue ;
AttValue = '"' ([^<&"] | Reference)* '"'
         | "'" ([^<&'] | Reference)* "'" ;

(* Content *)
Content = CharData? ((Element | Reference | CDSect | PI | Comment) CharData?)* ;

CDSect = CDStart CData CDEnd ;
CDStart = "<![CDATA[" ;
CData = (Char* - (Char* "]]>" Char*)) ;
CDEnd = "]]>" ;

(* Character Data *)
CharData = [^<&]* - ([^<&]* "]]>" [^<&]*) ;

(* Comments *)
Comment = "<!--" ((Char - '-') | ('-' (Char - '-')))* "-->" ;

(* Processing Instructions *)
PI = "<?" PITarget (S (Char* - (Char* "?>" Char*)))? "?>" ;
PITarget = Name - (("X"|"x")("M"|"m")("L"|"l")) ;

(* Names and Tokens *)
Name = NameStartChar NameChar* ;
NameStartChar = ":" | [A-Z] | "_" | [a-z] | [#xC0-#xD6] | [#xD8-#xF6] ;
NameChar = NameStartChar | "-" | "." | [0-9] | #xB7 ;

(* References *)
Reference = EntityRef | CharRef ;
EntityRef = "&" Name ";" ;
CharRef = "&#" [0-9]+ ";" | "&#x" [0-9a-fA-F]+ ";" ;

(* Whitespace *)
S = (#x20 | #x9 | #xD | #xA)+ ;

(* Misc *)
Misc = Comment | PI | S ;

Eq = S? "=" S? ;
```

2. **Implement Tokenizer**

Create `internal/tokenizer/tokenizer.go`:

```go
package tokenizer

import "github.com/shapestone/shape-core/pkg/tokenizer"

const (
    TokenLAngle        = "LAngle"        // <
    TokenRAngle        = "RAngle"        // >
    TokenSlash         = "Slash"         // /
    TokenEquals        = "Equals"        // =
    TokenQuestion      = "Question"      // ?
    TokenExclamation   = "Exclamation"   // !
    TokenName          = "Name"          // element/attribute name
    TokenString        = "String"        // "value" or 'value'
    TokenText          = "Text"          // character data
    TokenCDATAStart    = "CDATAStart"    // <![CDATA[
    TokenCDATAEnd      = "CDATAEnd"      // ]]>
    TokenCommentStart  = "CommentStart"  // <!--
    TokenCommentEnd    = "CommentEnd"    // -->
    TokenPIStart       = "PIStart"       // <?
    TokenPIEnd         = "PIEnd"         // ?>
    TokenWhitespace    = "Whitespace"    // whitespace
)

func NewTokenizer() tokenizer.Tokenizer {
    return tokenizer.NewTokenizer(
        // Special sequences (must come before single chars)
        tokenizer.StringMatcherFunc(TokenCDATAStart, "<![CDATA["),
        tokenizer.StringMatcherFunc(TokenCDATAEnd, "]]>"),
        tokenizer.StringMatcherFunc(TokenCommentStart, "<!--"),
        tokenizer.StringMatcherFunc(TokenCommentEnd, "-->"),
        tokenizer.StringMatcherFunc(TokenPIStart, "<?"),
        tokenizer.StringMatcherFunc(TokenPIEnd, "?>"),

        // Single characters
        tokenizer.CharMatcherFunc(TokenLAngle, '<'),
        tokenizer.CharMatcherFunc(TokenRAngle, '>'),
        tokenizer.CharMatcherFunc(TokenSlash, '/'),
        tokenizer.CharMatcherFunc(TokenEquals, '='),
        tokenizer.CharMatcherFunc(TokenQuestion, '?'),
        tokenizer.CharMatcherFunc(TokenExclamation, '!'),

        // Strings (attribute values)
        StringMatcher(),

        // Names (element/attribute names)
        NameMatcher(),

        // Text content
        TextMatcher(),

        // Whitespace
        tokenizer.WhitespaceMatcherFunc(),
    )
}

func StringMatcher() tokenizer.Matcher {
    // Match "..." or '...'
    return tokenizer.RegexMatcherFunc(TokenString, `"[^"]*"|'[^']*'`)
}

func NameMatcher() tokenizer.Matcher {
    // XML Name: NameStartChar NameChar*
    // Simplified: [a-zA-Z_:][a-zA-Z0-9_:.-]*
    return tokenizer.RegexMatcherFunc(TokenName, `[a-zA-Z_:][a-zA-Z0-9_:.-]*`)
}

func TextMatcher() tokenizer.Matcher {
    // Character data (everything except <, &, ]]>)
    return tokenizer.RegexMatcherFunc(TokenText, `[^<&]+`)
}
```

3. **Write Tokenizer Tests**

Create comprehensive tests for all token types:
- XML declaration (`<?xml version="1.0"?>`)
- Elements (`<element>`, `</element>`, `<empty/>`)
- Attributes (`name="value"`, `name='value'`)
- Text content
- CDATA sections
- Comments
- Processing instructions
- Namespaces (`ns:element`)

#### Deliverables
- ✅ EBNF grammar documented
- ✅ Tokenizer implementation
- ✅ Tokenizer tests with >90% coverage

---

### Phase 3: Core Parser Implementation

**Goal:** Implement parser returning universal AST
**Duration:** 3-5 days

#### Tasks

1. **Implement Core Parser**

Create `internal/parser/parser.go`:

```go
package parser

import (
    "fmt"
    "github.com/shapestone/shape-core/pkg/ast"
    "github.com/shapestone/shape-xml/internal/tokenizer"
)

type Parser struct {
    tok     tokenizer.Tokenizer
    current tokenizer.Token
}

func NewParser(input string) *Parser {
    tok := tokenizer.NewTokenizer()
    stream := tokenizer.NewStream(input)
    tok.SetStream(stream)

    return &Parser{
        tok: tok,
    }
}

// Parse returns AST representing the XML document
func (p *Parser) Parse() (ast.SchemaNode, error) {
    // Skip XML declaration if present
    if err := p.skipProlog(); err != nil {
        return nil, err
    }

    // Parse root element
    return p.parseElement()
}

func (p *Parser) parseElement() (ast.SchemaNode, error) {
    startPos := p.tok.Position()

    // Check for empty element: <name attr="val"/>
    if p.isEmptyElement() {
        return p.parseEmptyElement()
    }

    // Parse start tag: <name attr="val">
    name, attributes, err := p.parseStartTag()
    if err != nil {
        return nil, err
    }

    // Parse content (text, child elements, CDATA, etc.)
    content, err := p.parseContent()
    if err != nil {
        return nil, err
    }

    // Parse end tag: </name>
    if err := p.parseEndTag(name); err != nil {
        return nil, err
    }

    // Build ObjectNode with attributes and content
    properties := make(map[string]ast.SchemaNode)

    // Add attributes with @ prefix
    for key, value := range attributes {
        properties["@"+key] = ast.NewLiteralNode(value, startPos)
    }

    // Add content
    for key, value := range content {
        properties[key] = value
    }

    return ast.NewObjectNode(properties, startPos), nil
}

func (p *Parser) parseStartTag() (name string, attributes map[string]string, err error) {
    // <name attr="val">
    p.expect(TokenLAngle)
    name = p.expect(TokenName)

    attributes = make(map[string]string)

    // Parse attributes
    for p.peek().Kind() == TokenName {
        attrName := p.consume().Lexeme()
        p.expect(TokenEquals)
        attrValue := p.parseAttributeValue()
        attributes[attrName] = attrValue
    }

    p.expect(TokenRAngle)
    return name, attributes, nil
}

func (p *Parser) parseContent() (map[string]ast.SchemaNode, error) {
    content := make(map[string]ast.SchemaNode)

    for {
        switch p.peek().Kind() {
        case TokenText:
            // Text content
            text := p.consume().Lexeme()
            content["#text"] = ast.NewLiteralNode(text, p.tok.Position())

        case TokenLAngle:
            if p.isEndTag() {
                return content, nil
            }
            if p.isCDATA() {
                cdata := p.parseCDATA()
                content["#cdata"] = ast.NewLiteralNode(cdata, p.tok.Position())
            } else if p.isComment() {
                p.parseComment() // Skip comments
            } else {
                // Child element
                child, err := p.parseElement()
                if err != nil {
                    return nil, err
                }
                // Use element name as key (handle duplicates later)
                childName := p.getElementName(child)
                content[childName] = child
            }

        default:
            return content, nil
        }
    }
}

func (p *Parser) parseCDATA() string {
    p.expect(TokenCDATAStart)

    var cdata string
    for p.peek().Kind() != TokenCDATAEnd {
        cdata += p.consume().Lexeme()
    }

    p.expect(TokenCDATAEnd)
    return cdata
}

// ... more parser methods
```

2. **Handle XML-Specific Features**

- **Attributes**: Store with `@` prefix
- **Text content**: Store with `#text` key
- **CDATA sections**: Store with `#cdata` key
- **Namespaces**: Include prefix in element/attribute names
- **Namespace declarations**: Store as `@xmlns:prefix`
- **Comments**: Skip or store in metadata
- **Processing instructions**: Skip or store with `?target` key
- **Entity references**: Resolve during parsing (`&lt;` → `<`)

3. **Handle Edge Cases**

- Mixed content (text + elements)
- Multiple child elements with same name (convert to array)
- Empty elements (`<empty/>`)
- Self-closing tags
- Whitespace handling
- Malformed XML (error recovery)

4. **Write Parser Tests**

Comprehensive tests covering:
- Simple elements
- Nested elements
- Attributes
- Namespaces
- Text content
- CDATA sections
- Comments
- Processing instructions
- Mixed content
- Entity references
- Error cases

#### Deliverables
- ✅ Parser implementation returning AST
- ✅ All XML features handled
- ✅ Parser tests with >90% coverage
- ✅ Error messages with source positions

---

### Phase 4: Public API & Dual API Pattern

**Goal:** Provide both Parse (AST) and Unmarshal (Go types) APIs
**Duration:** 2-3 days

#### Tasks

1. **Implement Primary API (Parse → AST)**

Create `pkg/xml/parser.go`:

```go
package xml

import (
    "github.com/shapestone/shape-core/pkg/ast"
    "github.com/shapestone/shape-xml/internal/parser"
    "io"
)

// Parse parses XML input and returns universal AST.
// This is the primary API for working with XML structure.
func Parse(input string) (ast.SchemaNode, error) {
    p := parser.NewParser(input)
    return p.Parse()
}

// ParseReader parses XML from an io.Reader (for large files).
// Uses constant memory via streaming.
func ParseReader(r io.Reader) (ast.SchemaNode, error) {
    p := parser.NewParserFromReader(r)
    return p.Parse()
}

// Validate checks if input is valid XML (legacy function, kept for compatibility).
func Validate(input string) error {
    _, err := Parse(input)
    return err
}
```

2. **Implement Secondary API (Unmarshal → Go types)**

Create `pkg/xml/unmarshal.go`:

```go
package xml

import (
    "fmt"
    "reflect"
    "github.com/shapestone/shape-core/pkg/ast"
)

// Unmarshal parses XML and unmarshals into Go struct.
// This is the secondary API compatible with encoding/xml.
func Unmarshal(data []byte, v interface{}) error {
    // 1. Parse to AST
    node, err := Parse(string(data))
    if err != nil {
        return err
    }

    // 2. Convert AST → Go struct
    return unmarshalNode(node, reflect.ValueOf(v))
}

func unmarshalNode(node ast.SchemaNode, val reflect.Value) error {
    // Handle pointer dereferencing
    if val.Kind() == reflect.Ptr {
        if val.IsNil() {
            val.Set(reflect.New(val.Type().Elem()))
        }
        val = val.Elem()
    }

    switch n := node.(type) {
    case *ast.LiteralNode:
        return unmarshalLiteral(n, val)

    case *ast.ObjectNode:
        return unmarshalObject(n, val)

    default:
        return fmt.Errorf("unsupported node type: %T", node)
    }
}

func unmarshalObject(node *ast.ObjectNode, val reflect.Value) error {
    if val.Kind() != reflect.Struct {
        return fmt.Errorf("cannot unmarshal object into %v", val.Kind())
    }

    props := node.Properties()

    // Iterate struct fields
    typ := val.Type()
    for i := 0; i < val.NumField(); i++ {
        field := typ.Field(i)
        fieldVal := val.Field(i)

        // Get XML tag or use field name
        tag := field.Tag.Get("xml")
        if tag == "-" {
            continue // Skip field
        }
        if tag == "" {
            tag = field.Name
        }

        // Handle attributes (,attr tag)
        isAttr := false
        if len(tag) > 5 && tag[len(tag)-5:] == ",attr" {
            tag = tag[:len(tag)-5]
            tag = "@" + tag  // Attributes have @ prefix in AST
            isAttr = true
        }

        // Get value from AST
        propNode, ok := props[tag]
        if !ok {
            continue // Field not present in XML
        }

        // Unmarshal recursively
        if err := unmarshalNode(propNode, fieldVal); err != nil {
            return err
        }
    }

    return nil
}

// ToGoValue converts AST node to Go interface{} types.
// Useful for dynamic XML processing without structs.
func ToGoValue(node ast.SchemaNode) interface{} {
    switch n := node.(type) {
    case *ast.LiteralNode:
        return n.Value()

    case *ast.ObjectNode:
        result := make(map[string]interface{})
        for key, val := range n.Properties() {
            result[key] = ToGoValue(val)
        }
        return result

    default:
        return nil
    }
}
```

3. **Implement Marshal (Go types → XML)**

Create `pkg/xml/marshal.go`:

```go
package xml

import (
    "fmt"
    "reflect"
    "github.com/shapestone/shape-core/pkg/ast"
)

// Marshal converts Go value to XML string.
// Compatible with encoding/xml API.
func Marshal(v interface{}) ([]byte, error) {
    // 1. Convert Go value → AST
    node, err := marshalValue(reflect.ValueOf(v))
    if err != nil {
        return nil, err
    }

    // 2. Render AST → XML string
    xml := Render(node)

    return []byte(xml), nil
}

func marshalValue(val reflect.Value) (ast.SchemaNode, error) {
    // Handle different Go types
    switch val.Kind() {
    case reflect.Struct:
        return marshalStruct(val)

    case reflect.String:
        return ast.NewLiteralNode(val.String(), ast.Position{}), nil

    case reflect.Int, reflect.Int64:
        return ast.NewLiteralNode(val.Int(), ast.Position{}), nil

    // ... other types

    default:
        return nil, fmt.Errorf("unsupported type: %v", val.Kind())
    }
}

func marshalStruct(val reflect.Value) (ast.SchemaNode, error) {
    properties := make(map[string]ast.SchemaNode)

    typ := val.Type()
    for i := 0; i < val.NumField(); i++ {
        field := typ.Field(i)
        fieldVal := val.Field(i)

        // Get XML tag
        tag := field.Tag.Get("xml")
        if tag == "-" {
            continue
        }
        if tag == "" {
            tag = field.Name
        }

        // Handle attributes
        if len(tag) > 5 && tag[len(tag)-5:] == ",attr" {
            tag = "@" + tag[:len(tag)-5]
        }

        // Marshal field
        node, err := marshalValue(fieldVal)
        if err != nil {
            return nil, err
        }

        properties[tag] = node
    }

    return ast.NewObjectNode(properties, ast.Position{}), nil
}
```

4. **Implement Render (AST → XML string)**

Create `pkg/xml/render.go`:

```go
package xml

import (
    "fmt"
    "strings"
    "github.com/shapestone/shape-core/pkg/ast"
)

// Render converts AST back to XML string.
func Render(node ast.SchemaNode) string {
    return renderNode(node, 0)
}

func renderNode(node ast.SchemaNode, indent int) string {
    switch n := node.(type) {
    case *ast.LiteralNode:
        return fmt.Sprintf("%v", n.Value())

    case *ast.ObjectNode:
        return renderElement(n, indent)

    default:
        return ""
    }
}

func renderElement(node *ast.ObjectNode, indent int) string {
    props := node.Properties()

    var name string
    var attributes []string
    var content []string

    // Separate attributes, text, and child elements
    for key, val := range props {
        if strings.HasPrefix(key, "@") {
            // Attribute
            attrName := key[1:] // Remove @ prefix
            if !strings.HasPrefix(attrName, "xmlns") {
                attrVal := renderNode(val, 0)
                attributes = append(attributes, fmt.Sprintf(`%s="%s"`, attrName, attrVal))
            }
        } else if key == "#text" {
            // Text content
            content = append(content, renderNode(val, 0))
        } else if key == "#cdata" {
            // CDATA
            content = append(content, fmt.Sprintf("<![CDATA[%v]]>", val.(*ast.LiteralNode).Value()))
        } else {
            // Child element
            name = key
            content = append(content, renderNode(val, indent+2))
        }
    }

    // Build XML
    var sb strings.Builder

    // Start tag
    sb.WriteString("<")
    sb.WriteString(name)
    for _, attr := range attributes {
        sb.WriteString(" ")
        sb.WriteString(attr)
    }

    if len(content) == 0 {
        // Self-closing tag
        sb.WriteString("/>")
    } else {
        sb.WriteString(">")
        for _, c := range content {
            sb.WriteString(c)
        }
        sb.WriteString("</")
        sb.WriteString(name)
        sb.WriteString(">")
    }

    return sb.String()
}
```

#### Deliverables
- ✅ Parse() API returning AST
- ✅ ParseReader() for streaming
- ✅ Unmarshal() API for Go structs
- ✅ Marshal() API from Go structs
- ✅ Render() AST → XML
- ✅ ToGoValue() helper
- ✅ Tests for all APIs

---

### Phase 5: XPath Query Engine (Optional, Future)

**Goal:** Enable XPath queries on parsed XML
**Duration:** 5-7 days (separate phase)

This phase is optional and can be deferred. Shape-json has JSONPath, so shape-xml should eventually have XPath for parity.

#### Tasks

1. Implement basic XPath parser
2. Support common XPath expressions
3. Query AST nodes using XPath
4. Provide `Query(node, xpath)` API

**Defer to Phase 6 or later.**

---

### Phase 6: Documentation & Examples

**Goal:** Comprehensive documentation and examples
**Duration:** 2-3 days

#### Tasks

1. **Update README.md**

```markdown
# shape-xml

Production-ready XML parser for the Shape ecosystem.

## Features

- ✅ **Universal AST** - Parse XML to format-agnostic AST
- ✅ **Dual API** - Primary (Parse→AST) + Secondary (Unmarshal→Go structs)
- ✅ **XML Conventions** - Attributes (`@prefix`), text (`#text`), CDATA (`#cdata`)
- ✅ **Namespaces** - Full namespace support
- ✅ **Streaming** - ParseReader for large files
- ✅ **Position Tracking** - Source positions for error messages
- ✅ **encoding/xml Compatible** - Drop-in replacement for stdlib

## Installation

```bash
go get github.com/shapestone/shape-xml
```

## Quick Start

### Parse to AST (Primary API)

```go
import "github.com/shapestone/shape-xml/pkg/xml"

// Parse XML to AST
node, err := xml.Parse(`
<user id="123">
    <name>Alice</name>
</user>
`)

// Access AST
obj := node.(*ast.ObjectNode)
userID := obj.Properties()["@id"].(*ast.LiteralNode).Value() // "123"
```

### Unmarshal to Go Struct (Secondary API)

```go
type User struct {
    ID   string `xml:"id,attr"`
    Name string `xml:"name"`
}

var user User
err := xml.Unmarshal(data, &user)
// user.ID == "123", user.Name == "Alice"
```

### Why AST Instead of Go Types?

- ✅ Source positions for error messages
- ✅ Supports XPath queries (future)
- ✅ Document diffing and transformations
- ✅ Programmatic construction
- ✅ Preserves XML-specific features

## XML Conventions

| XML Feature | AST Representation |
|-------------|-------------------|
| Attribute | Property with `@` prefix (`@id`) |
| Text content | Property with `#text` key |
| CDATA | Property with `#cdata` key |
| Namespace | Include prefix in name (`ns:element`) |
| Namespace declaration | `@xmlns:prefix` |

## Documentation

- [EBNF Grammar](docs/grammar/xml.ebnf)
- [Parser Implementation Guide](https://github.com/shapestone/shape-core/blob/main/docs/PARSER_IMPLEMENTATION_GUIDE.md)
- [AST Conventions](https://github.com/shapestone/shape-core/blob/main/docs/AST_CONVENTIONS.md)
- [Shape Core](https://github.com/shapestone/shape-core)
```

2. **Create Examples**

- `examples/parse/main.go` - Parse XML to AST
- `examples/unmarshal/main.go` - Unmarshal to structs
- `examples/marshal/main.go` - Marshal from structs
- `examples/streaming/main.go` - Parse large files

3. **Write Developer Guide**

- Contributing guidelines
- Architecture overview
- Testing strategy
- Release process

#### Deliverables
- ✅ Comprehensive README
- ✅ Code examples
- ✅ Developer documentation

---

### Phase 7: Testing & Quality

**Goal:** Ensure production quality
**Duration:** 2-3 days

#### Tasks

1. **Comprehensive Test Suite**

- Unit tests for all components
- Integration tests
- Edge case tests
- Error handling tests
- Benchmark tests

**Target Coverage:** >90% for all packages

2. **Grammar Verification Tests**

Following shape-core guide:

```go
func TestGrammarVerification(t *testing.T) {
    spec, err := grammar.ParseEBNF("../../docs/grammar/xml.ebnf")
    if err != nil {
        t.Fatal(err)
    }

    // Generate tests from grammar
    tests := spec.GenerateTests(grammar.TestOptions{
        MaxDepth:      3,
        CoverAllRules: true,
    })

    // Verify parser matches grammar
    for _, test := range tests {
        _, err := Parse(test.Input)
        if test.ShouldPass && err != nil {
            t.Errorf("Expected valid, got error: %v", err)
        }
        if !test.ShouldPass && err == nil {
            t.Errorf("Expected error, got valid parse")
        }
    }
}
```

3. **Linting & Code Quality**

- golangci-lint configuration
- Run linter on all code
- Fix all issues

4. **Performance Testing**

- Benchmark parsing speed
- Benchmark memory usage
- Compare with encoding/xml
- Optimize hot paths

#### Deliverables
- ✅ >90% test coverage
- ✅ Grammar verification tests
- ✅ Clean lint results
- ✅ Performance benchmarks

---

### Phase 8: CI/CD & Release

**Goal:** Automated testing and release process
**Duration:** 1-2 days

#### Tasks

1. **GitHub Actions Workflows**

- Test on push/PR (Ubuntu, macOS, Windows)
- Lint on push/PR
- Coverage reporting (codecov)
- Benchmark tracking
- Security scanning (CodeQL, gosec)

2. **Release Process**

- Semantic versioning
- Changelog generation
- GitHub releases
- Tag creation

3. **Community Health Files**

- CODE_OF_CONDUCT.md
- CONTRIBUTING.md
- SECURITY.md
- CHANGELOG.md

#### Deliverables
- ✅ CI/CD pipeline configured
- ✅ Release automation
- ✅ Community health files

---

## Success Criteria

### Must Have (MVP)
- ✅ Parse XML to universal AST
- ✅ Handle attributes, namespaces, text, CDATA
- ✅ Dual API (Parse + Unmarshal)
- ✅ >90% test coverage
- ✅ Source position tracking
- ✅ Error messages with line/column
- ✅ EBNF grammar documented
- ✅ CI/CD pipeline

### Should Have
- ✅ Marshal (Go → XML)
- ✅ Render (AST → XML)
- ✅ Streaming parser (ParseReader)
- ✅ encoding/xml compatibility
- ✅ Benchmark suite
- ✅ Comprehensive examples

### Nice to Have (Future)
- ⏳ XPath query engine
- ⏳ Schema validation (XSD)
- ⏳ XML namespaces advanced features
- ⏳ Pretty printing with indentation
- ⏳ Comments preservation

---

## Timeline Estimate

| Phase | Duration | Dependencies |
|-------|----------|--------------|
| 1. Foundation | 1-2 days | None |
| 2. Tokenizer | 2-3 days | Phase 1 |
| 3. Parser | 3-5 days | Phase 2 |
| 4. Public API | 2-3 days | Phase 3 |
| 5. XPath (deferred) | - | - |
| 6. Documentation | 2-3 days | Phase 4 |
| 7. Testing | 2-3 days | All above |
| 8. CI/CD | 1-2 days | Phase 7 |

**Total:** 13-21 days (2-4 weeks)

**Recommended Approach:** Work in phases, complete MVP (Phases 1-4, 7-8) first (~10-14 days), then add documentation (Phase 6) and nice-to-haves later.

---

## Risk Assessment

### Technical Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| XML complexity (mixed content, entities) | High | High | Follow W3C spec strictly, comprehensive tests |
| Namespace handling complexity | Medium | Medium | Use established conventions, test extensively |
| Performance concerns vs encoding/xml | Medium | Low | Optimize after correctness, benchmark regularly |
| AST doesn't fit XML well | Low | High | Conventions handle 95% of cases, proven by shape-json |

### Process Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Scope creep (XPath, XSD, etc.) | High | Medium | Defer advanced features to later phases |
| Time estimate too optimistic | Medium | Low | Prioritize MVP, iterate |
| Breaking changes to shape-core | Low | High | Version pin, coordinate with shape-core changes |

---

## Next Steps

1. **Review this plan** with stakeholders
2. **Prioritize phases** (MVP vs nice-to-have)
3. **Allocate resources** (developer time)
4. **Set milestones** (weekly check-ins)
5. **Begin Phase 1** (Foundation & Setup)

---

## Appendix: Reference Implementation

See **shape-json** as the reference implementation:
- Structure: `/Users/michaelsundell/Projects/shapestone/shape-eco/shape-json`
- Dual API pattern
- AST usage
- Test coverage
- CI/CD setup

Apply the same patterns to shape-xml with XML-specific adaptations.

---

**Document Version:** 1.0
**Last Updated:** 2025-12-14
**Author:** Shape Team
**Status:** Draft - Pending Approval
