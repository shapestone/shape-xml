// Package fastparser implements a high-performance XML validator without AST construction.
//
// This parser is optimized for validation only - checking that XML is well-formed.
// It bypasses AST construction, validating directly from bytes for minimal overhead.
//
// Performance targets (vs AST parser):
//   - 4-5x faster validation
//   - 5-6x less memory
//   - 4-5x fewer allocations
package fastparser

import (
	"errors"
	"fmt"
)

// Parser implements a zero-allocation XML validator that checks well-formedness without AST.
type Parser struct {
	data   []byte
	pos    int
	length int
}

// NewParser creates a new fast parser for the given data.
func NewParser(data []byte) *Parser {
	return &Parser{
		data:   data,
		pos:    0,
		length: len(data),
	}
}

// Parse validates the XML data and returns true if valid.
// This is used by Validate and ValidateReader.
func (p *Parser) Parse() (bool, error) {
	p.skipWhitespace()
	if p.pos >= p.length {
		return false, errors.New("unexpected end of XML input")
	}

	// Skip optional XML declaration
	if p.peekString("<?xml") {
		if err := p.skipXMLDeclaration(); err != nil {
			return false, err
		}
	}

	p.skipWhitespace()

	// Skip any comments before root element
	p.skipComments()

	// Validate root element
	if err := p.validateElement(); err != nil {
		return false, err
	}

	// Skip trailing comments and whitespace
	p.skipCommentsAndWhitespace()

	// After parsing the root element, we should be at EOF
	if p.pos < p.length {
		return false, fmt.Errorf("unexpected content after root element at position %d", p.pos)
	}

	return true, nil
}

// validateElement validates an XML element without building AST.
// Checks for:
//   - Opening tag with valid name
//   - Matching closing tag (for non-self-closing elements)
//   - Proper nesting of child elements
//   - Valid attribute syntax
func (p *Parser) validateElement() error {
	// Expect '<'
	if !p.consume('<') {
		return fmt.Errorf("expected '<' at position %d", p.pos)
	}

	// Read element name
	elementName := p.readName()
	if elementName == "" {
		return fmt.Errorf("expected element name at position %d", p.pos)
	}

	// Read attributes
	for {
		p.skipWhitespace()

		// Check for end of opening tag
		if p.pos >= p.length {
			return fmt.Errorf("unexpected end of input in element %q", elementName)
		}

		// Self-closing tag: />
		if p.peekString("/>") {
			p.pos += 2
			return nil
		}

		// Regular closing: >
		if p.peek() == '>' {
			p.pos++
			break
		}

		// Must be an attribute
		if err := p.validateAttribute(); err != nil {
			return fmt.Errorf("in element %q: %w", elementName, err)
		}
	}

	// Parse content (text, CDATA, child elements)
	for {
		p.skipWhitespace()

		if p.pos >= p.length {
			return fmt.Errorf("unexpected end of input, expected closing tag for %q", elementName)
		}

		// Check for closing tag
		if p.peekString("</") {
			p.pos += 2

			closingName := p.readName()
			if closingName != elementName {
				return fmt.Errorf("mismatched tags: opening %q, closing %q at position %d",
					elementName, closingName, p.pos)
			}

			p.skipWhitespace()
			if !p.consume('>') {
				return fmt.Errorf("expected '>' in closing tag for element %q at position %d",
					elementName, p.pos)
			}

			return nil
		}

		// Check for comment
		if p.peekString("<!--") {
			if err := p.skipComment(); err != nil {
				return err
			}
			continue
		}

		// Check for CDATA
		if p.peekString("<![CDATA[") {
			if err := p.skipCData(); err != nil {
				return err
			}
			continue
		}

		// Check for child element
		if p.peek() == '<' {
			if err := p.validateElement(); err != nil {
				return fmt.Errorf("in element %q: %w", elementName, err)
			}
			continue
		}

		// Otherwise, it's text content - skip until next tag
		if err := p.skipText(); err != nil {
			return err
		}
	}
}

// validateAttribute validates an attribute without storing it.
// Attribute = Name "=" String
func (p *Parser) validateAttribute() error {
	// Read attribute name
	attrName := p.readName()
	if attrName == "" {
		return fmt.Errorf("expected attribute name at position %d", p.pos)
	}

	p.skipWhitespace()

	// Expect '='
	if !p.consume('=') {
		return fmt.Errorf("expected '=' after attribute name %q at position %d", attrName, p.pos)
	}

	p.skipWhitespace()

	// Read string value
	if err := p.validateString(); err != nil {
		return fmt.Errorf("invalid value for attribute %q: %w", attrName, err)
	}

	return nil
}

// validateString validates a quoted string (single or double quotes).
func (p *Parser) validateString() error {
	if p.pos >= p.length {
		return errors.New("expected string")
	}

	quote := p.data[p.pos]
	if quote != '"' && quote != '\'' {
		return fmt.Errorf("expected quote at position %d", p.pos)
	}
	p.pos++ // skip opening quote

	// Find closing quote
	for p.pos < p.length {
		c := p.data[p.pos]
		p.pos++

		if c == quote {
			return nil
		}

		// Handle escape sequences
		if c == '\\' {
			if p.pos >= p.length {
				return errors.New("unexpected end of string after backslash")
			}
			p.pos++ // skip escaped character
		}
	}

	return errors.New("unterminated string")
}

// skipXMLDeclaration skips the XML declaration.
// <?xml version="1.0" encoding="UTF-8"?>
func (p *Parser) skipXMLDeclaration() error {
	if !p.peekString("<?xml") {
		return nil
	}
	p.pos += 5

	// Find ?>
	for p.pos < p.length-1 {
		if p.data[p.pos] == '?' && p.data[p.pos+1] == '>' {
			p.pos += 2
			return nil
		}
		p.pos++
	}

	return errors.New("unterminated XML declaration")
}

// skipComment skips an XML comment: <!-- ... -->
func (p *Parser) skipComment() error {
	if !p.peekString("<!--") {
		return nil
	}
	p.pos += 4

	// Find -->
	for p.pos < p.length-2 {
		if p.data[p.pos] == '-' && p.data[p.pos+1] == '-' && p.data[p.pos+2] == '>' {
			p.pos += 3
			return nil
		}
		p.pos++
	}

	return errors.New("unterminated comment")
}

// skipCData skips a CDATA section: <![CDATA[ ... ]]>
func (p *Parser) skipCData() error {
	if !p.peekString("<![CDATA[") {
		return nil
	}
	p.pos += 9

	// Find ]]>
	for p.pos < p.length-2 {
		if p.data[p.pos] == ']' && p.data[p.pos+1] == ']' && p.data[p.pos+2] == '>' {
			p.pos += 3
			return nil
		}
		p.pos++
	}

	return errors.New("unterminated CDATA section")
}

// skipText skips text content until the next tag or special sequence.
func (p *Parser) skipText() error {
	for p.pos < p.length {
		c := p.data[p.pos]
		if c == '<' {
			return nil
		}
		p.pos++
	}
	return nil
}

// skipComments skips multiple consecutive comments.
func (p *Parser) skipComments() {
	for p.peekString("<!--") {
		p.skipComment()
		p.skipWhitespace()
	}
}

// skipCommentsAndWhitespace skips comments and whitespace.
func (p *Parser) skipCommentsAndWhitespace() {
	for {
		if p.pos >= p.length {
			return
		}

		if p.peekString("<!--") {
			p.skipComment()
			continue
		}

		c := p.data[p.pos]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			p.pos++
			continue
		}

		return
	}
}

// skipWhitespace skips whitespace characters (space, tab, LF, CR).
func (p *Parser) skipWhitespace() {
	for p.pos < p.length {
		c := p.data[p.pos]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			p.pos++
		} else {
			break
		}
	}
}

// readName reads an XML name (element or attribute name).
// Names match: [A-Za-z_:][A-Za-z0-9_:.-]*
func (p *Parser) readName() string {
	if p.pos >= p.length {
		return ""
	}

	start := p.pos

	// First character: letter, underscore, or colon
	c := p.data[p.pos]
	if !isNameStartChar(c) {
		return ""
	}
	p.pos++

	// Subsequent characters: letters, digits, underscore, colon, dot, hyphen
	for p.pos < p.length {
		c = p.data[p.pos]
		if !isNameChar(c) {
			break
		}
		p.pos++
	}

	return string(p.data[start:p.pos])
}

// peek returns the current character without advancing.
func (p *Parser) peek() byte {
	if p.pos >= p.length {
		return 0
	}
	return p.data[p.pos]
}

// peekString checks if the next characters match the given string.
func (p *Parser) peekString(s string) bool {
	if p.pos+len(s) > p.length {
		return false
	}
	return string(p.data[p.pos:p.pos+len(s)]) == s
}

// consume attempts to consume the expected character.
// Returns true if successful, false otherwise.
func (p *Parser) consume(expected byte) bool {
	if p.pos >= p.length || p.data[p.pos] != expected {
		return false
	}
	p.pos++
	return true
}

// isNameStartChar returns true if c can start an XML name.
func isNameStartChar(c byte) bool {
	return (c >= 'A' && c <= 'Z') ||
		(c >= 'a' && c <= 'z') ||
		c == '_' ||
		c == ':'
}

// isNameChar returns true if c can appear in an XML name.
func isNameChar(c byte) bool {
	return isNameStartChar(c) ||
		(c >= '0' && c <= '9') ||
		c == '.' ||
		c == '-'
}
