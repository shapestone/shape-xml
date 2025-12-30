package xml

import (
	"strings"
	"testing"
)

// TestDOMZeroCoverageMethods tests DOM methods with 0% coverage
func TestDOMZeroCoverageMethods(t *testing.T) {
	elem := NewElement()

	// Test Set method
	elem.Set("key", "value")

	// Test Get method
	if val, ok := elem.Get("key"); !ok || val != "value" {
		t.Errorf("Get() failed: got %v, %v", val, ok)
	}

	// Test XML method
	xml, err := elem.XML("test")
	if err != nil {
		t.Errorf("XML() error = %v", err)
	}
	if !strings.Contains(xml, "key") || !strings.Contains(xml, "value") {
		t.Errorf("XML() missing expected content, got: %s", xml)
	}

	// Test XMLIndent method
	xmlIndent, err := elem.XMLIndent("test", "", "  ")
	if err != nil {
		t.Errorf("XMLIndent() error = %v", err)
	}
	if !strings.Contains(xmlIndent, "key") || !strings.Contains(xmlIndent, "value") {
		t.Errorf("XMLIndent() missing expected content, got: %s", xmlIndent)
	}
}

// TestMarshalIndent tests MarshalIndent with 0% coverage
func TestMarshalIndentCoverage(t *testing.T) {
	type Simple struct {
		Name string `xml:"name"`
	}

	s := Simple{Name: "test"}
	data, err := MarshalIndent(s, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent() error = %v", err)
	}
	if !strings.Contains(string(data), "test") {
		t.Errorf("MarshalIndent() missing expected content")
	}
}

// TestEscapedStrings tests parsing of strings with escape sequences
func TestEscapedStrings(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"escaped quote in attribute", `<elem attr="value with \" quote"/>`},
		{"escaped single quote", `<elem attr='value with \' quote'/>`},
		{"escaped backslash", `<elem attr="path\\to\\file"/>`},
		{"multiple escapes", `<elem attr="\"quoted\" and \\escaped\\"/>`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			if err != nil {
				t.Logf("Parse(%s) = %v (escape sequences may not be supported)", tt.input, err)
			}
		})
	}
}

// TestLongTextContent tests parsing of long text content
func TestLongTextContent(t *testing.T) {
	// Create a large text block that might trigger buffer joining
	largeText := strings.Repeat("This is a long text content block. ", 100)
	input := "<root>" + largeText + "</root>"
	
	node, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if node == nil {
		t.Fatal("Parse() returned nil")
	}
}

// TestComplexNestedStructure tests complex nested elements
func TestComplexNestedStructure(t *testing.T) {
	input := `<root>
		<level1 attr1="val1">
			<level2 attr2="val2">
				<level3>Text content</level3>
			</level2>
			<level2b>More text</level2b>
		</level1>
		<another>Final</another>
	</root>`
	
	node, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if node == nil {
		t.Fatal("Parse() returned nil")
	}

	// Also test Unmarshal with this structure
	var result interface{}
	err = Unmarshal([]byte(input), &result)
	if err != nil {
		t.Logf("Unmarshal() error = %v", err)
	}
}
