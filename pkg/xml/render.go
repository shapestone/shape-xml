// Package xml provides AST rendering to XML bytes.
//
// This file implements the core XML rendering functionality, converting
// Shape AST nodes back into XML byte representations.
package xml

import (
	"bytes"
	"fmt"
	"html"
	"sort"
	"strings"
	"sync"

	"github.com/shapestone/shape-core/pkg/ast"
)

// bufferPool is a pool of bytes.Buffer instances to reduce allocations during rendering.
// Buffers are returned to the pool after use to minimize GC pressure.
var bufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 1024))
	},
}

// getBuffer retrieves a buffer from the pool and resets it for use.
func getBuffer() *bytes.Buffer {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

// putBuffer returns a buffer to the pool if it's not too large.
// Buffers larger than 64KB are not pooled to avoid holding excessive memory.
func putBuffer(buf *bytes.Buffer) {
	if buf.Cap() <= 64*1024 {
		bufferPool.Put(buf)
	}
}

// Render converts an AST node to compact XML bytes.
//
// The node should be the result of Parse() or ParseReader().
// Returns XML bytes with no unnecessary whitespace.
//
// The XML structure uses Shape's conventions:
//   - Properties prefixed with "@" are attributes
//   - Property "#text" contains text content
//   - Property "#cdata" contains CDATA sections
//   - Other properties are child elements
//
// Example:
//
//	node, _ := xml.Parse(`<user id="123"><name>Alice</name></user>`)
//	bytes, _ := xml.Render(node)
//	// bytes: <user id="123"><name>Alice</name></user>
func Render(node ast.SchemaNode) ([]byte, error) {
	buf := getBuffer()
	defer putBuffer(buf)

	if err := renderNode(node, buf, false, "", "", "root"); err != nil {
		return nil, err
	}

	// Must copy since buffer will be returned to pool
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	return result, nil
}

// RenderIndent converts an AST node to pretty-printed XML bytes with indentation.
//
// The prefix is added to the beginning of each line, and indent specifies
// the indentation string (typically spaces or tabs).
//
// Common usage:
//   - RenderIndent(node, "", "  ") - 2-space indentation
//   - RenderIndent(node, "", "\t") - tab indentation
//   - RenderIndent(node, ">>", "  ") - prefix each line with ">>"
//
// Example:
//
//	node, _ := xml.Parse(`<user id="123"><name>Alice</name></user>`)
//	bytes, _ := xml.RenderIndent(node, "", "  ")
//	// Output:
//	// <user id="123">
//	//   <name>Alice</name>
//	// </user>
func RenderIndent(node ast.SchemaNode, prefix, indent string) ([]byte, error) {
	buf := getBuffer()
	defer putBuffer(buf)

	if err := renderNode(node, buf, true, prefix, indent, "root"); err != nil {
		return nil, err
	}

	// Must copy since buffer will be returned to pool
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	return result, nil
}

// renderNode recursively renders an AST node to the buffer.
//
// Parameters:
//   - node: The AST node to render
//   - buf: The output buffer
//   - prettyPrint: Whether to add whitespace for readability
//   - prefix: String to add at the start of each line
//   - indent: Indentation string (spaces or tabs)
//   - elementName: The name of the XML element to render
func renderNode(node ast.SchemaNode, buf *bytes.Buffer, prettyPrint bool, prefix, indent, elementName string) error {
	return renderNodeWithDepth(node, buf, prettyPrint, prefix, indent, 0, elementName)
}

// renderNodeWithDepth renders a node with tracking of indentation depth.
func renderNodeWithDepth(node ast.SchemaNode, buf *bytes.Buffer, prettyPrint bool, prefix, indent string, depth int, elementName string) error {
	if node == nil {
		// Render self-closing tag for nil nodes
		if prettyPrint && depth > 0 {
			buf.WriteString(prefix)
			buf.WriteString(strings.Repeat(indent, depth))
		}
		buf.WriteString("<")
		buf.WriteString(elementName)
		buf.WriteString("/>")
		if prettyPrint {
			buf.WriteString("\n")
		}
		return nil
	}

	switch n := node.(type) {
	case *ast.ObjectNode:
		return renderElement(n, buf, prettyPrint, prefix, indent, depth, elementName)
	case *ast.ArrayDataNode:
		return renderArrayElements(n, buf, prettyPrint, prefix, indent, depth, elementName)
	case *ast.LiteralNode:
		// Literal nodes should be rendered as text content within an element
		if prettyPrint && depth > 0 {
			buf.WriteString(prefix)
			buf.WriteString(strings.Repeat(indent, depth))
		}
		buf.WriteString("<")
		buf.WriteString(elementName)
		buf.WriteString(">")
		buf.WriteString(escapeXML(fmt.Sprintf("%v", n.Value())))
		buf.WriteString("</")
		buf.WriteString(elementName)
		buf.WriteString(">")
		if prettyPrint {
			buf.WriteString("\n")
		}
		return nil
	default:
		return fmt.Errorf("unknown node type: %T", node)
	}
}

// renderElement renders an ObjectNode as an XML element.
func renderElement(node *ast.ObjectNode, buf *bytes.Buffer, prettyPrint bool, prefix, indent string, depth int, elementName string) error {
	props := node.Properties()

	// Add indentation if pretty printing
	if prettyPrint && depth > 0 {
		buf.WriteString(prefix)
		buf.WriteString(strings.Repeat(indent, depth))
	}

	// Start opening tag
	buf.WriteString("<")
	buf.WriteString(elementName)

	// Render attributes (properties starting with @)
	attrs := make([]string, 0)
	for key := range props {
		if strings.HasPrefix(key, "@") {
			attrs = append(attrs, key)
		}
	}
	sort.Strings(attrs) // Sort for consistent output

	for _, attrKey := range attrs {
		attrName := attrKey[1:] // Remove @ prefix
		attrNode := props[attrKey]
		if literal, ok := attrNode.(*ast.LiteralNode); ok {
			buf.WriteString(" ")
			buf.WriteString(attrName)
			buf.WriteString("=\"")
			buf.WriteString(escapeXML(fmt.Sprintf("%v", literal.Value())))
			buf.WriteString("\"")
		}
	}

	// Check for text content or CDATA
	textNode, hasText := props["#text"]
	cdataNode, hasCDATA := props["#cdata"]

	// Get child elements (properties not starting with @ or #)
	childKeys := make([]string, 0)
	for key := range props {
		if !strings.HasPrefix(key, "@") && !strings.HasPrefix(key, "#") {
			childKeys = append(childKeys, key)
		}
	}
	sort.Strings(childKeys) // Sort for consistent output

	hasChildren := len(childKeys) > 0

	// If no text, no CDATA, and no children, render as self-closing tag
	if !hasText && !hasCDATA && !hasChildren {
		buf.WriteString("/>")
		if prettyPrint {
			buf.WriteString("\n")
		}
		return nil
	}

	// Close opening tag
	buf.WriteString(">")

	// Render text content (no newline before/after text)
	if hasText {
		if literal, ok := textNode.(*ast.LiteralNode); ok {
			buf.WriteString(escapeXML(fmt.Sprintf("%v", literal.Value())))
		}
	}

	// Render CDATA content
	if hasCDATA {
		if literal, ok := cdataNode.(*ast.LiteralNode); ok {
			buf.WriteString("<![CDATA[")
			buf.WriteString(fmt.Sprintf("%v", literal.Value()))
			buf.WriteString("]]>")
		}
	}

	// Render child elements
	if hasChildren {
		if prettyPrint && !hasText {
			buf.WriteString("\n")
		}

		for _, childKey := range childKeys {
			childNode := props[childKey]
			if err := renderNodeWithDepth(childNode, buf, prettyPrint, prefix, indent, depth+1, childKey); err != nil {
				return err
			}
		}

		if prettyPrint && !hasText {
			buf.WriteString(prefix)
			buf.WriteString(strings.Repeat(indent, depth))
		}
	}

	// Close tag
	buf.WriteString("</")
	buf.WriteString(elementName)
	buf.WriteString(">")
	if prettyPrint {
		buf.WriteString("\n")
	}

	return nil
}

// renderArrayElements renders an ArrayDataNode as multiple XML elements.
func renderArrayElements(node *ast.ArrayDataNode, buf *bytes.Buffer, prettyPrint bool, prefix, indent string, depth int, elementName string) error {
	elements := node.Elements()

	for _, elem := range elements {
		if err := renderNodeWithDepth(elem, buf, prettyPrint, prefix, indent, depth, elementName); err != nil {
			return err
		}
	}

	return nil
}

// escapeXML escapes special XML characters.
//
// Handles:
//   - & → &amp;
//   - < → &lt;
//   - > → &gt;
//   - " → &quot;
//   - ' → &apos;
func escapeXML(s string) string {
	// Use html.EscapeString which handles &, <, >, ", and '
	return html.EscapeString(s)
}
