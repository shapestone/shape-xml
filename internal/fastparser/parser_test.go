package fastparser

import (
	"testing"
)

func TestParseValidXML(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "simple element",
			input: `<root></root>`,
		},
		{
			name:  "self-closing element",
			input: `<root/>`,
		},
		{
			name:  "element with text content",
			input: `<root>hello world</root>`,
		},
		{
			name:  "element with single attribute",
			input: `<root id="123"></root>`,
		},
		{
			name:  "element with multiple attributes",
			input: `<root id="123" name="test"></root>`,
		},
		{
			name:  "nested elements",
			input: `<root><child>value</child></root>`,
		},
		{
			name:  "multiple child elements",
			input: `<root><child1>value1</child1><child2>value2</child2></root>`,
		},
		{
			name:  "deeply nested elements",
			input: `<root><level1><level2><level3>value</level3></level2></level1></root>`,
		},
		{
			name:  "element with attributes and children",
			input: `<root id="1"><child name="test">value</child></root>`,
		},
		{
			name:  "XML with declaration",
			input: `<?xml version="1.0" encoding="UTF-8"?><root></root>`,
		},
		{
			name:  "XML with declaration and whitespace",
			input: `<?xml version="1.0"?> <root></root>`,
		},
		{
			name:  "XML with comment before root",
			input: `<!-- comment --><root></root>`,
		},
		{
			name:  "XML with comment after root",
			input: `<root></root><!-- comment -->`,
		},
		{
			name:  "XML with multiple comments",
			input: `<!-- comment1 --><!-- comment2 --><root></root>`,
		},
		{
			name:  "XML with CDATA",
			input: `<root><![CDATA[some data]]></root>`,
		},
		{
			name:  "XML with CDATA containing special chars",
			input: `<root><![CDATA[<>&"']]></root>`,
		},
		{
			name:  "element with single quoted attribute",
			input: `<root id='123'></root>`,
		},
		{
			name:  "element with mixed quote styles",
			input: `<root id="123" name='test'></root>`,
		},
		{
			name:  "element with namespace",
			input: `<ns:root xmlns:ns="http://example.com"></ns:root>`,
		},
		{
			name:  "element with empty text",
			input: `<root>   </root>`,
		},
		{
			name:  "complex real-world example",
			input: `<?xml version="1.0"?>
<users>
	<user id="1" active="true">
		<name>Alice</name>
		<email>alice@example.com</email>
		<roles>
			<role>admin</role>
			<role>user</role>
		</roles>
	</user>
	<user id="2" active="false">
		<name>Bob</name>
		<email>bob@example.com</email>
	</user>
</users>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser([]byte(tt.input))
			result, err := p.Parse()
			if err != nil {
				t.Errorf("Parse() error = %v, want nil", err)
				return
			}
			if result == nil {
				t.Errorf("Parse() returned nil, want non-nil map")
			}
		})
	}
}

func TestParseInvalidXML(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty input",
			input: ``,
		},
		{
			name:  "whitespace only",
			input: `   `,
		},
		{
			name:  "missing opening tag",
			input: `</root>`,
		},
		{
			name:  "missing closing tag",
			input: `<root>`,
		},
		{
			name:  "mismatched tags",
			input: `<root></wrong>`,
		},
		{
			name:  "unclosed tag",
			input: `<root><child></root>`,
		},
		{
			name:  "invalid tag name - starts with digit",
			input: `<1root></1root>`,
		},
		{
			name:  "missing attribute value",
			input: `<root id></root>`,
		},
		{
			name:  "missing equals in attribute",
			input: `<root id"123"></root>`,
		},
		{
			name:  "unterminated attribute value",
			input: `<root id="123></root>`,
		},
		{
			name:  "unterminated comment",
			input: `<!-- comment <root></root>`,
		},
		{
			name:  "unterminated CDATA",
			input: `<root><![CDATA[data</root>`,
		},
		{
			name:  "unterminated XML declaration",
			input: `<?xml version="1.0"<root></root>`,
		},
		{
			name:  "extra content after root",
			input: `<root></root><extra></extra>`,
		},
		{
			name:  "text before root element",
			input: `text<root></root>`,
		},
		{
			name:  "multiple root elements",
			input: `<root1></root1><root2></root2>`,
		},
		{
			name:  "invalid character in tag name",
			input: `<root@></root@>`,
		},
		{
			name:  "missing tag name",
			input: `<></root>`,
		},
		{
			name:  "missing closing bracket",
			input: `<root`,
		},
		{
			name:  "invalid self-closing syntax",
			input: `<root/ >`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser([]byte(tt.input))
			_, err := p.Parse()
			if err == nil {
				t.Errorf("Parse() expected error for invalid input %q", tt.input)
			}
		})
	}
}

func TestParseMalformedTags(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "tag not closed",
			input: `<root`,
		},
		{
			name:  "closing tag not closed",
			input: `<root></root`,
		},
		{
			name:  "self-closing tag space before slash",
			input: `<root / >`,
		},
		{
			name:  "nested unclosed tags",
			input: `<root><child>`,
		},
		{
			name:  "wrong nesting order",
			input: `<root><child></root></child>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser([]byte(tt.input))
			_, err := p.Parse()
			if err == nil {
				t.Errorf("Parse() expected error for malformed input %q", tt.input)
			}
		})
	}
}

func TestParseAttributes(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid single attribute",
			input:   `<root attr="value"></root>`,
			wantErr: false,
		},
		{
			name:    "valid multiple attributes",
			input:   `<root attr1="val1" attr2="val2" attr3="val3"></root>`,
			wantErr: false,
		},
		{
			name:    "attribute with spaces around equals",
			input:   `<root attr = "value"></root>`,
			wantErr: false,
		},
		{
			name:    "attribute with empty value",
			input:   `<root attr=""></root>`,
			wantErr: false,
		},
		{
			name:    "attribute with special characters",
			input:   `<root attr="value with &lt;special&gt; chars"></root>`,
			wantErr: false,
		},
		{
			name:    "missing equals sign",
			input:   `<root attr"value"></root>`,
			wantErr: true,
		},
		{
			name:    "missing value",
			input:   `<root attr=></root>`,
			wantErr: true,
		},
		{
			name:    "unquoted value",
			input:   `<root attr=value></root>`,
			wantErr: true,
		},
		{
			name:    "unterminated quoted value",
			input:   `<root attr="value></root>`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser([]byte(tt.input))
			_, err := p.Parse()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseComments(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "comment before root",
			input:   `<!-- comment --><root></root>`,
			wantErr: false,
		},
		{
			name:    "comment after root",
			input:   `<root></root><!-- comment -->`,
			wantErr: false,
		},
		{
			name:    "multiple comments",
			input:   `<!-- c1 --><!-- c2 --><root></root><!-- c3 -->`,
			wantErr: false,
		},
		{
			name:    "comment with special characters",
			input:   `<!-- <>&"' --><root></root>`,
			wantErr: false,
		},
		{
			name:    "unterminated comment",
			input:   `<!-- comment <root></root>`,
			wantErr: true,
		},
		{
			name:    "comment inside element",
			input:   `<root><!-- comment --></root>`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser([]byte(tt.input))
			_, err := p.Parse()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseCData(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "simple CDATA",
			input:   `<root><![CDATA[data]]></root>`,
			wantErr: false,
		},
		{
			name:    "CDATA with special characters",
			input:   `<root><![CDATA[<>&"']]></root>`,
			wantErr: false,
		},
		{
			name:    "CDATA with XML-like content",
			input:   `<root><![CDATA[<child>value</child>]]></root>`,
			wantErr: false,
		},
		{
			name:    "empty CDATA",
			input:   `<root><![CDATA[]]></root>`,
			wantErr: false,
		},
		{
			name:    "unterminated CDATA",
			input:   `<root><![CDATA[data</root>`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser([]byte(tt.input))
			_, err := p.Parse()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseWhitespace(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "leading whitespace",
			input: `   <root></root>`,
		},
		{
			name:  "trailing whitespace",
			input: `<root></root>   `,
		},
		{
			name:  "whitespace around elements",
			input: `  <root>  <child>  value  </child>  </root>  `,
		},
		{
			name:  "newlines and tabs",
			input: "\t\n<root>\n\t<child>value</child>\n</root>\n\t",
		},
		{
			name:  "whitespace in attributes",
			input: `<root   attr1="val1"   attr2="val2"   ></root>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser([]byte(tt.input))
			result, err := p.Parse()
			if err != nil {
				t.Errorf("Parse() error = %v, want nil", err)
				return
			}
			if result == nil {
				t.Errorf("Parse() returned nil, want non-nil map")
			}
		})
	}
}

func TestParseNamespaces(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "element with namespace prefix",
			input: `<ns:root></ns:root>`,
		},
		{
			name:  "element with xmlns declaration",
			input: `<root xmlns="http://example.com"></root>`,
		},
		{
			name:  "element with namespace prefix and xmlns",
			input: `<ns:root xmlns:ns="http://example.com"></ns:root>`,
		},
		{
			name:  "multiple namespace declarations",
			input: `<root xmlns:ns1="http://ex1.com" xmlns:ns2="http://ex2.com"></root>`,
		},
		{
			name:  "child with different namespace",
			input: `<root xmlns:a="http://a.com"><a:child></a:child></root>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser([]byte(tt.input))
			result, err := p.Parse()
			if err != nil {
				t.Errorf("Parse() error = %v, want nil", err)
				return
			}
			if result == nil {
				t.Errorf("Parse() returned nil, want non-nil map")
			}
		})
	}
}
