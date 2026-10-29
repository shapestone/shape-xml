// Package xml provides XML format parsing and AST generation.
//
// This package implements a complete XML parser that parses XML data into
// Shape's unified AST representation.
//
// # Architecture
//
// This parser uses LL(1) recursive descent parsing (see Shape ADR 0004).
// The grammar is defined in docs/grammar/xml.ebnf and verified through
// automated tests (see Shape ADR 0005: Grammar-as-Verification).
//
// # Thread Safety
//
// All functions in this package are safe for concurrent use by multiple goroutines.
// Each function call creates its own parser instance with no shared mutable state.
//
// # Parsing APIs
//
// The package provides two parsing functions:
//
//   - Parse(string) - Parses XML from a string in memory
//   - ParseReader(io.Reader) - Parses XML from any io.Reader with streaming support
//
// Use Parse() for small XML documents that are already in memory as strings.
// Use ParseReader() for large files, network streams, or any io.Reader source.
//
// # Example usage with Parse:
//
//	xmlStr := `<user id="123"><name>Alice</name></user>`
//	node, err := xml.Parse(xmlStr)
//	if err != nil {
//	    // handle error
//	}
//	// node is now a *ast.ObjectNode representing the XML data
//
// # Example usage with ParseReader:
//
//	file, err := os.Open("data.xml")
//	if err != nil {
//	    // handle error
//	}
//	defer file.Close()
//
//	node, err := xml.ParseReader(file)
//	if err != nil {
//	    // handle error
//	}
//	// node is now a *ast.ObjectNode representing the XML data
package xml

import (
	"io"

	"github.com/shapestone/shape-core/pkg/ast"
	"github.com/shapestone/shape-core/pkg/tokenizer"
	"github.com/shapestone/shape-xml/internal/fastparser"
	"github.com/shapestone/shape-xml/internal/parser"
)

// Parse parses XML format into an AST from a string.
//
// The input is a complete XML document with a root element.
//
// Returns an ast.SchemaNode representing the parsed XML:
//   - *ast.ObjectNode for elements
//   - Properties prefixed with "@" for attributes
//   - "#text" property for text content
//   - "#cdata" property for CDATA sections
//
// For parsing large files or streaming data, use ParseReader instead.
//
// Example:
//
//	node, err := xml.Parse(`<user id="123"><name>Alice</name></user>`)
//	obj := node.(*ast.ObjectNode)
//	idNode, _ := obj.GetProperty("@id")
//	id := idNode.(*ast.LiteralNode).Value().(string) // "123"
func Parse(input string) (ast.SchemaNode, error) {
	p := parser.NewParser(input)
	return p.Parse()
}

// ParseReader parses XML format into an AST from an io.Reader.
//
// This function is designed for parsing large XML files or streaming data with
// constant memory usage. It uses a buffered stream implementation that reads data
// in chunks, making it suitable for files that don't fit entirely in memory.
//
// The reader can be any io.Reader implementation:
//   - os.File for reading from files
//   - strings.Reader for reading from strings
//   - bytes.Buffer for reading from byte slices
//   - Network streams, compressed streams, etc.
//
// Returns an ast.SchemaNode representing the parsed XML:
//   - *ast.ObjectNode for elements
//   - Properties prefixed with "@" for attributes
//   - "#text" property for text content
//   - "#cdata" property for CDATA sections
//
// Example parsing from a file:
//
//	file, err := os.Open("data.xml")
//	if err != nil {
//	    // handle error
//	}
//	defer file.Close()
//
//	node, err := xml.ParseReader(file)
//	if err != nil {
//	    // handle error
//	}
//	// node is now a *ast.ObjectNode representing the XML data
func ParseReader(reader io.Reader) (ast.SchemaNode, error) {
	stream := tokenizer.NewStreamFromReader(reader)
	p := parser.NewParserFromStream(stream)
	return p.Parse()
}

// Format returns the format identifier for this parser.
// Returns "XML" to identify this as the XML data format parser.
func Format() string {
	return "XML"
}

// Validate checks if the given string is well-formed XML, with the same
// verdicts as a token loop over encoding/xml's Decoder (Strict mode, no
// CharsetReader). It uses a dedicated validator that builds no tree.
//
// Returns nil if the input is valid XML.
// Returns an error with details (line and column) about why the XML is invalid.
//
// Rejected: unknown entities such as &bogus; (only lt, gt, amp, apos, quot are
// defined), invalid character references (&#0;), characters XML forbids (NUL
// and other control characters, U+FFFE, U+FFFF), invalid UTF-8, '<' inside an
// attribute value, unquoted attributes, mismatched or unclosed tags, a
// version other than 1.0, and an XML declaration naming an encoding other than
// UTF-8.
//
// Accepted: a leading byte order mark, and a DOCTYPE. A DOCTYPE (including
// its internal subset) is skipped without being processed: entities declared
// there are not defined, parameter entities are not expanded, and external
// SYSTEM/PUBLIC identifiers are never loaded. A reference to an entity that a
// DOCTYPE declares is therefore an unknown entity and is rejected.
//
// Like encoding/xml, Validate does not require exactly one root element:
// empty input, several root elements and text outside the root are accepted.
//
// This is the idiomatic Go approach - check the error:
//
//	if err := xml.Validate(input); err != nil {
//	    // Invalid XML
//	    fmt.Println("Invalid XML:", err)
//	}
//	// Valid XML - err is nil
//
// Validate does not limit input size or element depth; limit the input before
// calling it if it is untrusted. See also ValidateReader.
func Validate(input string) error {
	return fastparser.Validate([]byte(input))
}

// ValidateReader checks if the XML from an io.Reader is well-formed, with the
// same rules as Validate. The whole input is read into memory first; wrap the
// reader in io.LimitReader to cap its size.
//
// Returns nil if the input is valid XML.
// Returns an error with details about why the XML is invalid.
//
// Example validating from a file:
//
//	file, err := os.Open("data.xml")
//	if err != nil {
//	    // handle error
//	}
//	defer file.Close()
//
//	if err := xml.ValidateReader(file); err != nil {
//	    // Invalid XML
//	    fmt.Println("Invalid XML:", err)
//	}
//	// Valid XML - err is nil
func ValidateReader(reader io.Reader) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	return fastparser.Validate(data)
}
