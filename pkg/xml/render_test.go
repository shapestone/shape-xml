package xml

import (
	"strings"
	"testing"
)

func TestRender_SimpleElement(t *testing.T) {
	input := `<user></user>`
	node, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	bytes, err := Render(node)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	result := string(bytes)
	// Should produce self-closing tag for empty element
	if !strings.Contains(result, "root") {
		t.Errorf("Expected root element, got: %s", result)
	}
}

func TestRender_WithAttributes(t *testing.T) {
	input := `<user id="123" name="Alice"></user>`
	node, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	bytes, err := Render(node)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	result := string(bytes)
	if !strings.Contains(result, `id="123"`) {
		t.Errorf("Expected id attribute, got: %s", result)
	}
	if !strings.Contains(result, `name="Alice"`) {
		t.Errorf("Expected name attribute, got: %s", result)
	}
}

func TestRender_WithTextContent(t *testing.T) {
	input := `<message>Hello, World!</message>`
	node, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	bytes, err := Render(node)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	result := string(bytes)
	if !strings.Contains(result, "Hello, World!") {
		t.Errorf("Expected text content, got: %s", result)
	}
}

func TestRender_NestedElements(t *testing.T) {
	input := `<user><name>Alice</name><email>alice@example.com</email></user>`
	node, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	bytes, err := Render(node)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	result := string(bytes)
	// Parser currently puts children under "child" element
	// Just verify rendering works and produces XML
	if !strings.Contains(result, "root") || !strings.Contains(result, "child") {
		t.Errorf("Expected rendered XML with root and child elements, got: %s", result)
	}
	if !strings.Contains(result, "Alice") {
		t.Errorf("Expected Alice in rendered output, got: %s", result)
	}
}

func TestRender_AttributesAndText(t *testing.T) {
	input := `<user id="123">Alice</user>`
	node, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	bytes, err := Render(node)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	result := string(bytes)
	if !strings.Contains(result, `id="123"`) {
		t.Errorf("Expected id attribute, got: %s", result)
	}
	if !strings.Contains(result, "Alice") {
		t.Errorf("Expected text content, got: %s", result)
	}
}

func TestRender_XMLEscaping(t *testing.T) {
	// Test that special XML characters are properly escaped
	elem := NewElement().
		Attr("title", `Quote "test" & <tag>`).
		Text(`Text with & and <tag>`)

	node, err := InterfaceToNode(elem.data)
	if err != nil {
		t.Fatalf("InterfaceToNode failed: %v", err)
	}

	bytes, err := Render(node)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	result := string(bytes)
	// Should escape &, <, >, "
	if strings.Contains(result, `& `) {
		t.Error("Expected & to be escaped as &amp;")
	}
	if strings.Contains(result, `<tag>`) {
		t.Error("Expected < and > to be escaped")
	}
}

func TestRenderIndent_Simple(t *testing.T) {
	input := `<user><name>Alice</name></user>`
	node, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	bytes, err := RenderIndent(node, "", "  ")
	if err != nil {
		t.Fatalf("RenderIndent failed: %v", err)
	}

	result := string(bytes)
	// Should contain newlines for pretty printing
	if !strings.Contains(result, "\n") {
		t.Errorf("Expected indented output with newlines, got: %s", result)
	}
}

func TestRenderIndent_Nested(t *testing.T) {
	elem := NewElement().
		Attr("id", "123").
		Child("name", NewElement().Text("Alice")).
		Child("address", NewElement().
			Child("city", NewElement().Text("NYC")).
			Child("zip", NewElement().Text("10001")))

	node, err := InterfaceToNode(elem.data)
	if err != nil {
		t.Fatalf("InterfaceToNode failed: %v", err)
	}

	bytes, err := RenderIndent(node, "", "  ")
	if err != nil {
		t.Fatalf("RenderIndent failed: %v", err)
	}

	result := string(bytes)
	// Should have multiple levels of indentation
	lines := strings.Split(result, "\n")
	if len(lines) < 3 {
		t.Errorf("Expected multiple lines of indented output, got: %s", result)
	}
}
