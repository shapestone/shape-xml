// Package xml provides a user-friendly DOM API for XML manipulation.
//
// The DOM API provides type-safe, fluent interfaces for building and manipulating
// XML documents without requiring type assertions or working with raw AST nodes.
//
// # Element Type
//
// Element represents an XML element with chainable methods:
//
//	elem := xml.NewElement("user").
//		Attr("id", "123").
//		Text("Alice")
//
// # Nested Structures
//
// Build complex nested XML structures fluently:
//
//	doc := xml.NewElement("user").
//		Attr("id", "123").
//		Child("name", xml.NewElement("name").Text("Alice")).
//		Child("email", xml.NewElement("email").Text("alice@example.com"))
package xml

import (
	"fmt"
)

// Element represents an XML element with a fluent API for manipulation.
// All setter methods return *Element to enable method chaining.
type Element struct {
	data map[string]interface{}
}

// NewElement creates a new Element.
// The element name is not stored in the Element itself but is used when rendering.
// This is just a data container following the XML AST convention.
func NewElement() *Element {
	return &Element{data: make(map[string]interface{})}
}

// ParseElement parses XML string into an Element with a fluent API.
// Returns an error if the input is not valid XML.
func ParseElement(input string) (*Element, error) {
	// Parse XML to AST
	node, err := Parse(input)
	if err != nil {
		return nil, err
	}

	// Convert AST to map[string]interface{}
	value := NodeToInterface(node)
	data, ok := value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected XML element, got %T", value)
	}
	return &Element{data: data}, nil
}

// ============================================================================
// Element Builder Methods (fluent setters that return *Element)
// ============================================================================

// Set sets a generic value and returns the Element for chaining.
func (e *Element) Set(key string, value interface{}) *Element {
	e.data[key] = value
	return e
}

// Attr sets an attribute and returns the Element for chaining.
// Attributes are stored with "@" prefix following XML AST convention.
func (e *Element) Attr(name, value string) *Element {
	e.data["@"+name] = value
	return e
}

// Text sets the text content and returns the Element for chaining.
// Text content is stored as "#text" following XML AST convention.
func (e *Element) Text(value string) *Element {
	e.data["#text"] = value
	return e
}

// CDATA sets CDATA content and returns the Element for chaining.
// CDATA content is stored as "#cdata" following XML AST convention.
func (e *Element) CDATA(value string) *Element {
	e.data["#cdata"] = value
	return e
}

// Child adds a child element and returns the parent Element for chaining.
// The name is the element name (e.g., "name", "email").
func (e *Element) Child(name string, child *Element) *Element {
	e.data[name] = child.data
	return e
}

// ChildText adds a child element with text content and returns the parent Element for chaining.
// This is a convenience method equivalent to Child(name, NewElement().Text(text)).
func (e *Element) ChildText(name, text string) *Element {
	e.data[name] = map[string]interface{}{"#text": text}
	return e
}

// ============================================================================
// Element Getter Methods (type-safe access)
// ============================================================================

// Get gets a value as interface{}. Returns nil if not found.
func (e *Element) Get(key string) (interface{}, bool) {
	val, ok := e.data[key]
	return val, ok
}

// GetAttr gets an attribute value. Returns empty string and false if not found.
func (e *Element) GetAttr(name string) (string, bool) {
	if val, ok := e.data["@"+name]; ok {
		if str, ok := val.(string); ok {
			return str, true
		}
	}
	return "", false
}

// GetText gets the text content. Returns empty string and false if not found.
func (e *Element) GetText() (string, bool) {
	if val, ok := e.data["#text"]; ok {
		if str, ok := val.(string); ok {
			return str, true
		}
	}
	return "", false
}

// GetCDATA gets the CDATA content. Returns empty string and false if not found.
func (e *Element) GetCDATA() (string, bool) {
	if val, ok := e.data["#cdata"]; ok {
		if str, ok := val.(string); ok {
			return str, true
		}
	}
	return "", false
}

// GetChild gets a child element. Returns nil and false if not found or wrong type.
func (e *Element) GetChild(name string) (*Element, bool) {
	if val, ok := e.data[name]; ok {
		if m, ok := val.(map[string]interface{}); ok {
			return &Element{data: m}, true
		}
	}
	return nil, false
}

// Has checks if a key exists.
func (e *Element) Has(key string) bool {
	_, ok := e.data[key]
	return ok
}

// HasAttr checks if an attribute exists.
func (e *Element) HasAttr(name string) bool {
	_, ok := e.data["@"+name]
	return ok
}

// Remove removes a key and returns the Element for chaining.
func (e *Element) Remove(key string) *Element {
	delete(e.data, key)
	return e
}

// RemoveAttr removes an attribute and returns the Element for chaining.
func (e *Element) RemoveAttr(name string) *Element {
	delete(e.data, "@"+name)
	return e
}

// Keys returns all keys in the Element (including @-prefixed and #-prefixed).
func (e *Element) Keys() []string {
	keys := make([]string, 0, len(e.data))
	for k := range e.data {
		keys = append(keys, k)
	}
	return keys
}

// Attrs returns all attribute names (without @ prefix).
func (e *Element) Attrs() []string {
	attrs := make([]string, 0)
	for k := range e.data {
		if len(k) > 0 && k[0] == '@' {
			attrs = append(attrs, k[1:])
		}
	}
	return attrs
}

// Children returns names of all child elements (excluding attributes and text/cdata).
func (e *Element) Children() []string {
	children := make([]string, 0)
	for k := range e.data {
		if len(k) > 0 && k[0] != '@' && k[0] != '#' {
			children = append(children, k)
		}
	}
	return children
}

// ToMap returns the underlying map[string]interface{}.
func (e *Element) ToMap() map[string]interface{} {
	return e.data
}

// XML marshals the Element to an XML string with the given element name.
//
// Example:
//
//	elem := NewElement().Attr("id", "123").Text("Alice")
//	xml, _ := elem.XML("user")
//	// Returns: <user id="123">Alice</user>
func (e *Element) XML(elementName string) (string, error) {
	// Convert map to AST
	node, err := InterfaceToNode(e.data)
	if err != nil {
		return "", err
	}

	// Render AST to XML
	bytes, err := Render(node)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// XMLIndent returns a pretty-printed XML string representation with indentation.
// The prefix is written at the beginning of each line, and indent specifies the indentation string.
//
// Common usage:
//   - XMLIndent("user", "", "  ") - 2-space indentation
//   - XMLIndent("user", "", "\t") - tab indentation
//
// Example:
//
//	elem := NewElement().
//	    Attr("id", "123").
//	    ChildText("name", "Alice")
//	pretty, _ := elem.XMLIndent("user", "", "  ")
//	// Output:
//	// <user id="123">
//	//   <name>Alice</name>
//	// </user>
func (e *Element) XMLIndent(elementName, prefix, indent string) (string, error) {
	// Convert map to AST
	node, err := InterfaceToNode(e.data)
	if err != nil {
		return "", err
	}

	// Render AST to XML with indentation
	bytes, err := RenderIndent(node, prefix, indent)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
