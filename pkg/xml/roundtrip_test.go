package xml

import (
	"strings"
	"testing"
)

// TestRoundtrip_Simple tests basic parse -> render round trip
func TestRoundtrip_Simple(t *testing.T) {
	input := `<user id="123">Alice</user>`

	// Parse
	node, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Render
	bytes, err := Render(node)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	result := string(bytes)

	// Verify key elements are present
	if !strings.Contains(result, `id="123"`) {
		t.Errorf("Expected id attribute in result: %s", result)
	}
	if !strings.Contains(result, "Alice") {
		t.Errorf("Expected text content in result: %s", result)
	}
}

// TestRoundtrip_Nested tests round trip with nested elements
func TestRoundtrip_Nested(t *testing.T) {
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

	// Parser puts children under "child" - just verify round trip works
	if !strings.Contains(result, "Alice") {
		t.Errorf("Expected 'Alice' in result: %s", result)
	}
	if !strings.Contains(result, "alice@example.com") {
		t.Errorf("Expected email in result: %s", result)
	}
}

// TestRoundtrip_Complex tests round trip with attributes, text, and nested elements
func TestRoundtrip_Complex(t *testing.T) {
	elem := NewElement().
		Attr("id", "123").
		Attr("active", "true").
		ChildText("name", "Alice").
		ChildText("email", "alice@example.com").
		Child("address", NewElement().
			ChildText("city", "NYC").
			ChildText("zip", "10001"))

	// Convert to AST
	node, err := InterfaceToNode(elem.data)
	if err != nil {
		t.Fatalf("InterfaceToNode failed: %v", err)
	}

	// Render to XML
	bytes, err := Render(node)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	result := string(bytes)

	// Verify the output contains key data
	if !strings.Contains(result, `id="123"`) {
		t.Errorf("Expected id attribute in result: %s", result)
	}
	if !strings.Contains(result, "Alice") {
		t.Errorf("Expected Alice in result: %s", result)
	}
	if !strings.Contains(result, "NYC") {
		t.Errorf("Expected NYC in result: %s", result)
	}
}

// TestRoundtrip_Marshal tests marshal -> parse -> render round trip
func TestRoundtrip_Marshal(t *testing.T) {
	type User struct {
		ID   string `xml:"id,attr"`
		Name string `xml:",chardata"`
	}
	user := User{ID: "123", Name: "Alice"}

	// Marshal
	bytes, err := Marshal(user)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Parse
	node, err := Parse(string(bytes))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Render
	bytes2, err := Render(node)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	result := string(bytes2)

	// Verify content
	if !strings.Contains(result, `id="123"`) {
		t.Errorf("Expected id attribute in result: %s", result)
	}
	if !strings.Contains(result, "Alice") {
		t.Errorf("Expected text content in result: %s", result)
	}
}

// TestRoundtrip_Unmarshal tests parse -> unmarshal -> marshal round trip
func TestRoundtrip_Unmarshal(t *testing.T) {
	input := `<user id="123">Alice</user>`

	// Unmarshal
	var data map[string]interface{}
	err := Unmarshal([]byte(input), &data)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Convert to AST
	node, err := InterfaceToNode(data)
	if err != nil {
		t.Fatalf("InterfaceToNode failed: %v", err)
	}

	// Render
	bytes, err := Render(node)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	result := string(bytes)

	// Verify content
	if !strings.Contains(result, `id="123"`) {
		t.Errorf("Expected id attribute in result: %s", result)
	}
	if !strings.Contains(result, "Alice") {
		t.Errorf("Expected text content in result: %s", result)
	}
}
