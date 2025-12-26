// Package xml provides XML format parsing and AST generation.
//
// This package implements a complete XML parser.
// It parses XML data into Shape's unified AST representation.
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

// Validate checks if the given string is valid XML.
// It uses the fast parser for efficient validation without AST construction.
//
// Returns nil if the input is valid XML.
// Returns an error with details about why the XML is invalid.
//
// This is the idiomatic Go approach - check the error:
//
//	if err := xml.Validate(input); err != nil {
//	    // Invalid XML
//	    fmt.Println("Invalid XML:", err)
//	}
//	// Valid XML - err is nil
//
// For validating large files or streaming data, use ValidateReader instead.
func Validate(input string) error {
	parser := fastparser.NewParser([]byte(input))
	_, err := parser.Parse()
	return err
}

// ValidateReader checks if the XML from an io.Reader is valid.
// It uses the fast parser for efficient validation without AST construction.
//
// This function is designed for validating large XML files or streaming data
// without loading the entire content into memory.
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
	parser := fastparser.NewParser(data)
	_, err = parser.Parse()
	return err
}
