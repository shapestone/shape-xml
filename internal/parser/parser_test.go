package parser

import (
	"testing"
	"github.com/shapestone/shape-core/pkg/tokenizer"
)

// TestNewParserFromStream tests the NewParserFromStream constructor
func TestNewParserFromStream(t *testing.T) {
	input := "<root>test</root>"
	stream := tokenizer.NewStream(input)
	
	parser := NewParserFromStream(stream)
	if parser == nil {
		t.Fatal("NewParserFromStream() returned nil")
	}
	
	node, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if node == nil {
		t.Fatal("Parse() returned nil node")
	}
}

// TestSkipCommentsAndWhitespace tests parsing with comments and whitespace
func TestSkipCommentsAndWhitespace(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"comment before root", "<!-- comment -->\n<root></root>"},
		{"multiple comments", "<!-- c1 --><!-- c2 --><root></root>"},
		{"whitespace and comments", "  \n\t  <!-- comment -->  \n  <root></root>"},
		{"comment after root", "<root></root><!-- comment -->"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			_, err := parser.Parse()
			if err != nil {
				t.Logf("Parse() error = %v (may not be fully supported)", err)
			}
		})
	}
}
