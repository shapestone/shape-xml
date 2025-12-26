package xml

import (
	"bytes"
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid XML - simple element",
			input:   `<root></root>`,
			wantErr: false,
		},
		{
			name:    "valid XML - self-closing",
			input:   `<root/>`,
			wantErr: false,
		},
		{
			name:    "valid XML - with text",
			input:   `<root><child>value</child></root>`,
			wantErr: false,
		},
		{
			name:    "valid XML - with attributes",
			input:   `<root attr="value"><child>text</child></root>`,
			wantErr: false,
		},
		{
			name:    "valid XML - with declaration",
			input:   `<?xml version="1.0"?><root></root>`,
			wantErr: false,
		},
		{
			name:    "valid XML - with comment",
			input:   `<!-- comment --><root></root>`,
			wantErr: false,
		},
		{
			name:    "valid XML - with CDATA",
			input:   `<root><![CDATA[data]]></root>`,
			wantErr: false,
		},
		{
			name:    "valid XML - nested elements",
			input:   `<root><level1><level2>value</level2></level1></root>`,
			wantErr: false,
		},
		{
			name:    "valid XML - multiple children",
			input:   `<root><child1>v1</child1><child2>v2</child2></root>`,
			wantErr: false,
		},
		{
			name:    "valid XML - complex",
			input:   `<?xml version="1.0"?><users><user id="1"><name>Alice</name></user></users>`,
			wantErr: false,
		},
		{
			name:    "invalid XML - empty string",
			input:   ``,
			wantErr: true,
		},
		{
			name:    "invalid XML - whitespace only",
			input:   `   `,
			wantErr: true,
		},
		{
			name:    "invalid XML - unclosed tag",
			input:   `<root><child>value</root>`,
			wantErr: true,
		},
		{
			name:    "invalid XML - malformed",
			input:   `<root><child>value`,
			wantErr: true,
		},
		{
			name:    "invalid XML - mismatched tags",
			input:   `<root></wrong>`,
			wantErr: true,
		},
		{
			name:    "invalid XML - missing closing tag",
			input:   `<root>`,
			wantErr: true,
		},
		{
			name:    "invalid XML - extra content after root",
			input:   `<root></root><extra></extra>`,
			wantErr: true,
		},
		{
			name:    "invalid XML - invalid attribute",
			input:   `<root attr=value></root>`,
			wantErr: true,
		},
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

func TestValidateReader(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid XML - simple element",
			input:   `<root></root>`,
			wantErr: false,
		},
		{
			name:    "valid XML - self-closing",
			input:   `<root/>`,
			wantErr: false,
		},
		{
			name:    "valid XML - with text",
			input:   `<root><child>value</child></root>`,
			wantErr: false,
		},
		{
			name:    "valid XML - with attributes",
			input:   `<root attr="value"><child>text</child></root>`,
			wantErr: false,
		},
		{
			name:    "valid XML - with declaration",
			input:   `<?xml version="1.0"?><root></root>`,
			wantErr: false,
		},
		{
			name:    "valid XML - with comment",
			input:   `<!-- comment --><root></root>`,
			wantErr: false,
		},
		{
			name:    "valid XML - with CDATA",
			input:   `<root><![CDATA[data]]></root>`,
			wantErr: false,
		},
		{
			name:    "valid XML - large document",
			input:   generateLargeXML(100),
			wantErr: false,
		},
		{
			name:    "invalid XML - empty",
			input:   ``,
			wantErr: true,
		},
		{
			name:    "invalid XML - unclosed tag",
			input:   `<root><child>value</root>`,
			wantErr: true,
		},
		{
			name:    "invalid XML - malformed",
			input:   `<root><child>value`,
			wantErr: true,
		},
		{
			name:    "invalid XML - mismatched tags",
			input:   `<root></wrong>`,
			wantErr: true,
		},
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

func TestValidateReaderWithBytesBuffer(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid XML from buffer",
			input:   `<root><child>value</child></root>`,
			wantErr: false,
		},
		{
			name:    "invalid XML from buffer",
			input:   `<root><child>value</root>`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bytes.NewBufferString(tt.input)
			err := ValidateReader(buffer)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateReader() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// generateLargeXML generates a large XML document for testing streaming.
func generateLargeXML(numElements int) string {
	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0"?>`)
	sb.WriteString(`<root>`)
	for i := 0; i < numElements; i++ {
		sb.WriteString(`<item id="`)
		sb.WriteString(strings.Repeat("a", 10))
		sb.WriteString(`">`)
		sb.WriteString(`<name>Item `)
		sb.WriteString(strings.Repeat("b", 10))
		sb.WriteString(`</name>`)
		sb.WriteString(`<value>`)
		sb.WriteString(strings.Repeat("c", 100))
		sb.WriteString(`</value>`)
		sb.WriteString(`</item>`)
	}
	sb.WriteString(`</root>`)
	return sb.String()
}

// ============================================================================
// Thread Safety Tests
// ============================================================================

func TestConcurrent_Parse(t *testing.T) {
	const numGoroutines = 100
	const numIterations = 10

	input := `<user id="123"><name>Alice</name><email>alice@example.com</email></user>`

	// Run Parse concurrently
	done := make(chan bool, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() { done <- true }()
			for j := 0; j < numIterations; j++ {
				_, err := Parse(input)
				if err != nil {
					t.Errorf("Concurrent Parse failed: %v", err)
				}
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

func TestConcurrent_Validate(t *testing.T) {
	const numGoroutines = 100
	const numIterations = 10

	input := `<user id="123"><name>Alice</name></user>`

	// Run Validate concurrently
	done := make(chan bool, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() { done <- true }()
			for j := 0; j < numIterations; j++ {
				err := Validate(input)
				if err != nil {
					t.Errorf("Concurrent Validate failed: %v", err)
				}
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

func TestConcurrent_ParseReader(t *testing.T) {
	const numGoroutines = 50
	const numIterations = 5

	input := `<user><name>Alice</name><email>alice@example.com</email></user>`

	// Run ParseReader concurrently
	done := make(chan bool, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() { done <- true }()
			for j := 0; j < numIterations; j++ {
				reader := strings.NewReader(input)
				_, err := ParseReader(reader)
				if err != nil {
					t.Errorf("Concurrent ParseReader failed: %v", err)
				}
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

func TestConcurrent_ValidateReader(t *testing.T) {
	const numGoroutines = 50
	const numIterations = 5

	input := `<user><name>Alice</name></user>`

	// Run ValidateReader concurrently
	done := make(chan bool, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() { done <- true }()
			for j := 0; j < numIterations; j++ {
				reader := strings.NewReader(input)
				err := ValidateReader(reader)
				if err != nil {
					t.Errorf("Concurrent ValidateReader failed: %v", err)
				}
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

func TestConcurrent_MixedOperations(t *testing.T) {
	const numGoroutines = 100

	input := `<user id="123"><name>Alice</name></user>`

	// Run mixed operations concurrently
	done := make(chan bool, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer func() { done <- true }()

			// Each goroutine performs different operations
			switch index % 4 {
			case 0:
				_, _ = Parse(input)
			case 1:
				_ = Validate(input)
			case 2:
				reader := strings.NewReader(input)
				_, _ = ParseReader(reader)
			case 3:
				reader := strings.NewReader(input)
				_ = ValidateReader(reader)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}
