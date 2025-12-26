package xml

import (
	"strings"
	"testing"
)

// Coverage tests ensure comprehensive code coverage across all major functions

// TestCoverage_Parse tests various parse scenarios
func TestCoverage_Parse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty element", "<root/>", false},
		{"with attributes", "<user id=\"123\" name=\"Alice\"/>", false},
		{"with text", "<message>Hello</message>", false},
		{"nested elements", "<user><name>Alice</name></user>", false},
		{"with declaration", "<?xml version=\"1.0\"?><root/>", false},
		{"empty input", "", true},
		{"invalid xml", "<invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestCoverage_ParseReader tests ParseReader scenarios
func TestCoverage_ParseReader(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"simple", "<root/>", false},
		{"complex", "<user><name>Alice</name></user>", false},
		{"invalid", "<invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			_, err := ParseReader(reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseReader() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestCoverage_Validate tests validation scenarios
func TestCoverage_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "<root/>", false},
		{"valid complex", "<user id=\"123\">Alice</user>", false},
		{"empty", "", true},
		{"invalid", "<unclosed", true},
		{"mismatched", "<open></close>", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestCoverage_ValidateReader tests ValidateReader scenarios
func TestCoverage_ValidateReader(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "<root/>", false},
		{"invalid", "<invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			err := ValidateReader(reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateReader() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestCoverage_Render tests rendering scenarios
func TestCoverage_Render(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", "<root/>"},
		{"with attributes", "<user id=\"123\"/>"},
		{"with text", "<message>Hello</message>"},
		{"nested", "<user><name>Alice</name></user>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			bytes, err := Render(node)
			if err != nil {
				t.Errorf("Render() error = %v", err)
			}

			if len(bytes) == 0 {
				t.Error("Render() returned empty bytes")
			}
		})
	}
}

// TestCoverage_RenderIndent tests indented rendering
func TestCoverage_RenderIndent(t *testing.T) {
	input := "<user><name>Alice</name></user>"
	node, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	bytes, err := RenderIndent(node, "", "  ")
	if err != nil {
		t.Errorf("RenderIndent() error = %v", err)
	}

	if !strings.Contains(string(bytes), "\n") {
		t.Error("RenderIndent() should contain newlines")
	}
}

// TestCoverage_Marshal tests marshaling scenarios
func TestCoverage_Marshal(t *testing.T) {
	type SimpleStruct struct {
		Value string
	}

	type ComplexStruct struct {
		ID   string `xml:"id,attr"`
		Name string `xml:",chardata"`
	}

	tests := []struct {
		name  string
		input interface{}
	}{
		{"simple struct", SimpleStruct{Value: "test"}},
		{"complex struct", ComplexStruct{ID: "123", Name: "Alice"}},
		{"string", "test"},
		{"int", 42},
		{"bool", true},
		{"slice", []string{"a", "b"}},
		{"map", map[string]string{"key": "value"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytes, err := Marshal(tt.input)
			if err != nil {
				t.Errorf("Marshal() error = %v", err)
			}

			if len(bytes) == 0 {
				t.Error("Marshal() returned empty bytes")
			}
		})
	}
}

// TestCoverage_Unmarshal tests unmarshaling scenarios
func TestCoverage_Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"to map", "<user id=\"123\">Alice</user>", false},
		{"to interface", "<user>Alice</user>", false},
		{"invalid xml", "<invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result interface{}
			err := Unmarshal([]byte(tt.input), &result)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestCoverage_DOM tests DOM API scenarios
func TestCoverage_DOM(t *testing.T) {
	// Test Element creation and manipulation
	elem := NewElement().
		Attr("id", "123").
		Text("content").
		ChildText("name", "Alice")

	if !elem.HasAttr("id") {
		t.Error("Expected attribute 'id'")
	}

	if text, ok := elem.GetText(); !ok || text != "content" {
		t.Errorf("Expected text='content', got %v", text)
	}

	// Test ParseElement
	_, err := ParseElement("<user id=\"123\">Alice</user>")
	if err != nil {
		t.Errorf("ParseElement() error = %v", err)
	}

	// Test ParseElement error case
	_, err = ParseElement("<invalid")
	if err == nil {
		t.Error("Expected error for invalid XML")
	}
}

// TestCoverage_Convert tests conversion functions
func TestCoverage_Convert(t *testing.T) {
	// Test InterfaceToNode
	data := map[string]interface{}{
		"@id":  "123",
		"#text": "Alice",
	}

	node, err := InterfaceToNode(data)
	if err != nil {
		t.Errorf("InterfaceToNode() error = %v", err)
	}

	// Test NodeToInterface
	result := NodeToInterface(node)
	if result == nil {
		t.Error("NodeToInterface() returned nil")
	}

	// Test ReleaseTree
	ReleaseTree(node) // Should not panic
}

// TestCoverage_Format tests Format function
func TestCoverage_Format(t *testing.T) {
	format := Format()
	if format != "XML" {
		t.Errorf("Format() = %v, want XML", format)
	}
}
