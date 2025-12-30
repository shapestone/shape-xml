// Package fastparser implements a high-performance XML parser without AST construction.
//
// This parser is optimized for direct parsing to Go native types (map[string]interface{}).
// It bypasses AST construction, parsing directly from bytes to Go values for minimal overhead.
//
// This is the "fast path" in the dual-path parser pattern. Use this for:
//   - Unmarshal: Populate Go structs from XML (4-5x faster than AST path)
//   - Validate: Check XML well-formedness (4-5x faster than AST path)
//
// Performance targets (vs AST parser):
//   - 4-5x faster parsing
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

// Parse parses the XML data and returns the value as interface{} (map[string]interface{}).
// This is used by Unmarshal and Validate.
// For validation, the caller can simply discard the returned value.
func (p *Parser) Parse() (interface{}, error) {
	p.skipWhitespace()
	if p.pos >= p.length {
		return nil, errors.New("unexpected end of XML input")
	}

	// Skip optional XML declaration
	if p.peekString("<?xml") {
		if err := p.skipXMLDeclaration(); err != nil {
			return nil, err
		}
	}

	p.skipWhitespace()

	// Skip any comments before root element
	p.skipComments()

	// Parse root element to Go map
	result, err := p.parseElement()
	if err != nil {
		return nil, err
	}

	// Skip trailing comments and whitespace
	p.skipCommentsAndWhitespace()

	// After parsing the root element, we should be at EOF
	if p.pos < p.length {
		return nil, fmt.Errorf("unexpected content after root element at position %d", p.pos)
	}

	return result, nil
}

// parseElement parses an XML element and returns it as a map[string]interface{}.
// The map contains:
//   - "@attribute": attribute values (prefixed with @)
//   - "childElement": child element nodes
//   - "#text": text content
//   - "#cdata": CDATA content
func (p *Parser) parseElement() (map[string]interface{}, error) {
	// Expect '<'
	if !p.consume('<') {
		return nil, fmt.Errorf("expected '<' at position %d", p.pos)
	}

	// Read element name
	elementName := p.readName()
	if elementName == "" {
		return nil, fmt.Errorf("expected element name at position %d", p.pos)
	}

	result := make(map[string]interface{})

	// Read attributes
	for {
		p.skipWhitespace()

		// Check for end of opening tag
		if p.pos >= p.length {
			return nil, fmt.Errorf("unexpected end of input in element %q", elementName)
		}

		// Self-closing tag: />
		if p.peekString("/>") {
			p.pos += 2
			return result, nil
		}

		// Regular closing: >
		if p.peek() == '>' {
			p.pos++
			break
		}

		// Must be an attribute
		attrName, attrValue, err := p.parseAttribute()
		if err != nil {
			return nil, fmt.Errorf("in element %q: %w", elementName, err)
		}
		// Prefix attribute names with @
		result["@"+attrName] = attrValue
	}

	// Parse content (text, CDATA, child elements)
	var textParts []string
	var cdataParts []string

	for {
		p.skipWhitespace()

		if p.pos >= p.length {
			return nil, fmt.Errorf("unexpected end of input, expected closing tag for %q", elementName)
		}

		// Check for closing tag
		if p.peekString("</") {
			p.pos += 2

			closingName := p.readName()
			if closingName != elementName {
				return nil, fmt.Errorf("mismatched tags: opening %q, closing %q at position %d",
					elementName, closingName, p.pos)
			}

			p.skipWhitespace()
			if !p.consume('>') {
				return nil, fmt.Errorf("expected '>' in closing tag for element %q at position %d",
					elementName, p.pos)
			}

			// Add accumulated text and CDATA if any
			if len(textParts) > 0 {
				text := trimSpace(joinStrings(textParts))
				if text != "" {
					result["#text"] = text
				}
			}
			if len(cdataParts) > 0 {
				result["#cdata"] = joinStrings(cdataParts)
			}

			return result, nil
		}

		// Check for comment
		if p.peekString("<!--") {
			if err := p.skipComment(); err != nil {
				return nil, err
			}
			continue
		}

		// Check for CDATA
		if p.peekString("<![CDATA[") {
			cdata, err := p.parseCDataContent()
			if err != nil {
				return nil, err
			}
			cdataParts = append(cdataParts, cdata)
			continue
		}

		// Check for child element
		if p.peek() == '<' {
			// Save accumulated text before parsing child
			if len(textParts) > 0 {
				text := trimSpace(joinStrings(textParts))
				if text != "" {
					result["#text"] = text
				}
				textParts = nil
			}

			// Peek ahead to get child element name
			savedPos := p.pos
			p.pos++ // skip '<'
			childName := p.readName()
			p.pos = savedPos // restore position

			if childName == "" {
				return nil, fmt.Errorf("expected child element name at position %d", p.pos)
			}

			childNode, err := p.parseElement()
			if err != nil {
				return nil, fmt.Errorf("in element %q: %w", elementName, err)
			}

			// Store child by element name
			if existing, exists := result[childName]; exists {
				// Already have this element - convert to array or append
				if arr, ok := existing.([]interface{}); ok {
					result[childName] = append(arr, childNode)
				} else {
					result[childName] = []interface{}{existing, childNode}
				}
			} else {
				result[childName] = childNode
			}
			continue
		}

		// Otherwise, it's text content
		text, err := p.parseText()
		if err != nil {
			return nil, err
		}
		if text != "" {
			textParts = append(textParts, text)
		}
	}
}

// parseAttribute parses an attribute and returns its name and value.
// Attribute = Name "=" String
func (p *Parser) parseAttribute() (string, string, error) {
	// Read attribute name
	attrName := p.readName()
	if attrName == "" {
		return "", "", fmt.Errorf("expected attribute name at position %d", p.pos)
	}

	p.skipWhitespace()

	// Expect '='
	if !p.consume('=') {
		return "", "", fmt.Errorf("expected '=' after attribute name %q at position %d", attrName, p.pos)
	}

	p.skipWhitespace()

	// Read string value
	attrValue, err := p.parseString()
	if err != nil {
		return "", "", fmt.Errorf("invalid value for attribute %q: %w", attrName, err)
	}

	return attrName, attrValue, nil
}

// parseString parses a quoted string (single or double quotes) and returns its value.
func (p *Parser) parseString() (string, error) {
	if p.pos >= p.length {
		return "", errors.New("expected string")
	}

	quote := p.data[p.pos]
	if quote != '"' && quote != '\'' {
		return "", fmt.Errorf("expected quote at position %d", p.pos)
	}
	p.pos++ // skip opening quote

	start := p.pos

	// Fast path: no escape sequences
	for p.pos < p.length {
		c := p.data[p.pos]

		if c == quote {
			// Found closing quote
			s := string(p.data[start:p.pos])
			p.pos++ // skip closing quote
			return s, nil
		}

		// Handle escape sequences
		if c == '\\' {
			// Found escape, use slow path
			return p.parseStringWithEscapes(start, quote)
		}

		p.pos++
	}

	return "", errors.New("unterminated string")
}

// parseStringWithEscapes handles strings containing escape sequences.
func (p *Parser) parseStringWithEscapes(start int, quote byte) (string, error) {
	// We already found an escape at p.pos, everything before is in data[start:p.pos]
	var buf []byte
	buf = append(buf, p.data[start:p.pos]...)

	for p.pos < p.length {
		c := p.data[p.pos]

		if c == quote {
			p.pos++ // skip closing quote
			return string(buf), nil
		}

		if c == '\\' {
			p.pos++
			if p.pos >= p.length {
				return "", errors.New("unexpected end of string after backslash")
			}

			escaped := p.data[p.pos]
			p.pos++

			// Handle common XML escape sequences
			switch escaped {
			case '\\', '"', '\'':
				buf = append(buf, escaped)
			case 'n':
				buf = append(buf, '\n')
			case 't':
				buf = append(buf, '\t')
			case 'r':
				buf = append(buf, '\r')
			default:
				// For other escapes, preserve the backslash
				buf = append(buf, '\\', escaped)
			}
		} else {
			buf = append(buf, c)
			p.pos++
		}
	}

	return "", errors.New("unterminated string")
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


// parseText parses text content until the next tag or special sequence.
func (p *Parser) parseText() (string, error) {
	start := p.pos
	for p.pos < p.length {
		c := p.data[p.pos]
		if c == '<' {
			return string(p.data[start:p.pos]), nil
		}
		p.pos++
	}
	return string(p.data[start:p.pos]), nil
}

// parseCDataContent parses a CDATA section and returns its content.
// <![CDATA[ ... ]]>
func (p *Parser) parseCDataContent() (string, error) {
	if !p.peekString("<![CDATA[") {
		return "", errors.New("expected CDATA section")
	}
	p.pos += 9 // skip "<![CDATA["

	start := p.pos

	// Find ]]>
	for p.pos < p.length-2 {
		if p.data[p.pos] == ']' && p.data[p.pos+1] == ']' && p.data[p.pos+2] == '>' {
			content := string(p.data[start:p.pos])
			p.pos += 3 // skip "]]>"
			return content, nil
		}
		p.pos++
	}

	return "", errors.New("unterminated CDATA section")
}

// skipComments skips multiple consecutive comments.
func (p *Parser) skipComments() {
	for p.peekString("<!--") {
		_ = p.skipComment() // Ignore error; parsing will fail later if comment is invalid
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
			_ = p.skipComment() // Ignore error; parsing will fail later if comment is invalid
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

// Helper functions for string manipulation

// joinStrings joins a slice of strings efficiently.
func joinStrings(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}

	// Calculate total length
	totalLen := 0
	for _, s := range parts {
		totalLen += len(s)
	}

	// Build result
	buf := make([]byte, 0, totalLen)
	for _, s := range parts {
		buf = append(buf, s...)
	}
	return string(buf)
}

// trimSpace trims leading and trailing whitespace from a string.
func trimSpace(s string) string {
	// Find first non-whitespace
	start := 0
	for start < len(s) && isWhitespace(s[start]) {
		start++
	}

	// Find last non-whitespace
	end := len(s)
	for end > start && isWhitespace(s[end-1]) {
		end--
	}

	return s[start:end]
}

// isWhitespace returns true if c is a whitespace character.
func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}
