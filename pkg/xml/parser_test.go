package xml

import (
	"strings"
	"testing"

	"github.com/shapestone/shape-core/pkg/ast"
)

func TestParse_BasicElement(t *testing.T) {
	input := `<user></user>`
	node, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	obj, ok := node.(*ast.ObjectNode)
	if !ok {
		t.Fatalf("Expected *ast.ObjectNode, got %T", node)
	}

	if len(obj.Properties()) != 0 {
		t.Errorf("Expected empty object, got %d properties", len(obj.Properties()))
	}
}

func TestParse_SelfClosingElement(t *testing.T) {
	input := `<user/>`
	node, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	obj, ok := node.(*ast.ObjectNode)
	if !ok {
		t.Fatalf("Expected *ast.ObjectNode, got %T", node)
	}

	if len(obj.Properties()) != 0 {
		t.Errorf("Expected empty object, got %d properties", len(obj.Properties()))
	}
}

func TestParse_Attributes(t *testing.T) {
	input := `<user id="123" name="Alice"></user>`
	node, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	obj, ok := node.(*ast.ObjectNode)
	if !ok {
		t.Fatalf("Expected *ast.ObjectNode, got %T", node)
	}

	// Check @id attribute
	idNode, exists := obj.GetProperty("@id")
	if !exists {
		t.Fatal("Expected @id property")
	}
	idLiteral, ok := idNode.(*ast.LiteralNode)
	if !ok {
		t.Fatalf("Expected @id to be *ast.LiteralNode, got %T", idNode)
	}
	if idLiteral.Value() != "123" {
		t.Errorf("Expected @id='123', got %v", idLiteral.Value())
	}

	// Check @name attribute
	nameNode, exists := obj.GetProperty("@name")
	if !exists {
		t.Fatal("Expected @name property")
	}
	nameLiteral, ok := nameNode.(*ast.LiteralNode)
	if !ok {
		t.Fatalf("Expected @name to be *ast.LiteralNode, got %T", nameNode)
	}
	if nameLiteral.Value() != "Alice" {
		t.Errorf("Expected @name='Alice', got %v", nameLiteral.Value())
	}
}

func TestParse_TextContent(t *testing.T) {
	input := `<message>Hello, World!</message>`
	node, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	obj, ok := node.(*ast.ObjectNode)
	if !ok {
		t.Fatalf("Expected *ast.ObjectNode, got %T", node)
	}

	// Check #text property
	textNode, exists := obj.GetProperty("#text")
	if !exists {
		t.Fatal("Expected #text property")
	}
	textLiteral, ok := textNode.(*ast.LiteralNode)
	if !ok {
		t.Fatalf("Expected #text to be *ast.LiteralNode, got %T", textNode)
	}
	if textLiteral.Value() != "Hello, World!" {
		t.Errorf("Expected text='Hello, World!', got %v", textLiteral.Value())
	}
}

func TestParse_NestedElements(t *testing.T) {
	input := `<user><name>Alice</name><email>alice@example.com</email></user>`
	node, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	obj, ok := node.(*ast.ObjectNode)
	if !ok {
		t.Fatalf("Expected *ast.ObjectNode, got %T", node)
	}

	// For now, we expect child elements under "child" key
	// In the future, this should use actual element names
	if len(obj.Properties()) == 0 {
		t.Error("Expected child elements")
	}
}

func TestParse_AttributesAndText(t *testing.T) {
	input := `<user id="123">Alice</user>`
	node, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	obj, ok := node.(*ast.ObjectNode)
	if !ok {
		t.Fatalf("Expected *ast.ObjectNode, got %T", node)
	}

	// Check @id attribute
	idNode, exists := obj.GetProperty("@id")
	if !exists {
		t.Fatal("Expected @id property")
	}
	idLiteral, ok := idNode.(*ast.LiteralNode)
	if !ok {
		t.Fatalf("Expected @id to be *ast.LiteralNode, got %T", idNode)
	}
	if idLiteral.Value() != "123" {
		t.Errorf("Expected @id='123', got %v", idLiteral.Value())
	}

	// Check #text property
	textNode, exists := obj.GetProperty("#text")
	if !exists {
		t.Fatal("Expected #text property")
	}
	textLiteral, ok := textNode.(*ast.LiteralNode)
	if !ok {
		t.Fatalf("Expected #text to be *ast.LiteralNode, got %T", textNode)
	}
	if textLiteral.Value() != "Alice" {
		t.Errorf("Expected text='Alice', got %v", textLiteral.Value())
	}
}

func TestParse_XMLDeclaration(t *testing.T) {
	input := `<?xml version="1.0" encoding="UTF-8"?><root></root>`
	node, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	_, ok := node.(*ast.ObjectNode)
	if !ok {
		t.Fatalf("Expected *ast.ObjectNode, got %T", node)
	}
}

func TestParse_Namespaces(t *testing.T) {
	input := `<ns:user xmlns:ns="http://example.com"><ns:name>Alice</ns:name></ns:user>`
	node, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	_, ok := node.(*ast.ObjectNode)
	if !ok {
		t.Fatalf("Expected *ast.ObjectNode, got %T", node)
	}

	// Namespaces should be treated as part of the element name
	// and attributes
}

func TestParse_MismatchedTags(t *testing.T) {
	input := `<user><name>Alice</user></name>`
	_, err := Parse(input)
	if err == nil {
		t.Fatal("Expected error for mismatched tags")
	}
	if !strings.Contains(err.Error(), "mismatched") {
		t.Errorf("Expected 'mismatched' error, got: %v", err)
	}
}

func TestParse_UnterminatedTag(t *testing.T) {
	input := `<user`
	_, err := Parse(input)
	if err == nil {
		t.Fatal("Expected error for unterminated tag")
	}
}

func TestParse_InvalidXML(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"no root", "text without root"},
		{"unclosed tag", "<user>"},
		{"invalid attribute", "<user id=>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			if err == nil {
				t.Errorf("Expected error for input: %s", tt.input)
			}
		})
	}
}

func TestParseReader(t *testing.T) {
	input := `<user id="123"><name>Alice</name></user>`
	reader := strings.NewReader(input)

	node, err := ParseReader(reader)
	if err != nil {
		t.Fatalf("ParseReader failed: %v", err)
	}

	obj, ok := node.(*ast.ObjectNode)
	if !ok {
		t.Fatalf("Expected *ast.ObjectNode, got %T", node)
	}

	// Check @id attribute
	idNode, exists := obj.GetProperty("@id")
	if !exists {
		t.Fatal("Expected @id property")
	}
	idLiteral, ok := idNode.(*ast.LiteralNode)
	if !ok {
		t.Fatalf("Expected @id to be *ast.LiteralNode, got %T", idNode)
	}
	if idLiteral.Value() != "123" {
		t.Errorf("Expected @id='123', got %v", idLiteral.Value())
	}
}

func TestFormat(t *testing.T) {
	format := Format()
	if format != "XML" {
		t.Errorf("Expected format 'XML', got %q", format)
	}
}
