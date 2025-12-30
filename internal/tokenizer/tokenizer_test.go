package tokenizer

import (
	"testing"

	"github.com/shapestone/shape-core/pkg/tokenizer"
)

func TestNewTokenizer(t *testing.T) {
	tok := NewTokenizer()
	// Just verify it doesn't panic
	_ = tok
}

func TestNewTokenizerWithStream(t *testing.T) {
	stream := tokenizer.NewStream("<root/>")
	tok := NewTokenizerWithStream(stream)
	// Just verify it doesn't panic
	_ = tok
}

func TestCommentMatcher(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOk  bool
		wantLen int
	}{
		{
			name:    "valid comment",
			input:   "<!-- comment -->",
			wantOk:  true,
			wantLen: 4, // "<!--"
		},
		{
			name:   "not a comment",
			input:  "<root>",
			wantOk: false,
		},
		{
			name:   "partial comment start",
			input:  "<!-",
			wantOk: false,
		},
		{
			name:   "unterminated comment",
			input:  "<!-- comment",
			wantOk: false,
		},
		{
			name:    "comment with dashes",
			input:   "<!-- some-text -->",
			wantOk:  true,
			wantLen: 4,
		},
		{
			name:    "empty comment",
			input:   "<!---->",
			wantOk:  true,
			wantLen: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := tokenizer.NewStream(tt.input)
			matcher := CommentMatcher()
			token := matcher(stream)

			if tt.wantOk {
				if token == nil {
					t.Errorf("expected token, got nil")
					return
				}
				if token.Kind() != TokenCommentStart {
					t.Errorf("expected TokenCommentStart, got %s", token.Kind())
				}
				if len(token.Value()) != tt.wantLen {
					t.Errorf("expected value length %d, got %d", tt.wantLen, len(token.Value()))
				}
			} else {
				if token != nil {
					t.Errorf("expected nil token, got %v", token)
				}
			}
		})
	}
}

func TestCDataMatcher(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		wantOk bool
	}{
		{
			name:   "valid CDATA start",
			input:  "<![CDATA[content]]>",
			wantOk: true,
		},
		{
			name:   "not CDATA",
			input:  "<root>",
			wantOk: false,
		},
		{
			name:   "partial CDATA",
			input:  "<![CDATA",
			wantOk: false,
		},
		{
			name:   "wrong case",
			input:  "<![cdata[",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := tokenizer.NewStream(tt.input)
			matcher := CDataMatcher()
			token := matcher(stream)

			if tt.wantOk {
				if token == nil {
					t.Errorf("expected token, got nil")
					return
				}
				if token.Kind() != TokenCDataStart {
					t.Errorf("expected TokenCDataStart, got %s", token.Kind())
				}
			} else {
				if token != nil {
					t.Errorf("expected nil token, got %v", token)
				}
			}
		})
	}
}

func TestPIAndXMLDeclMatcher(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantOk    bool
		wantKind  string
	}{
		{
			name:     "XML declaration",
			input:    "<?xml version=\"1.0\"?>",
			wantOk:   true,
			wantKind: TokenXMLDeclStart,
		},
		{
			name:     "processing instruction",
			input:    "<?target data?>",
			wantOk:   true,
			wantKind: TokenPIStart,
		},
		{
			name:   "not PI or decl",
			input:  "<root>",
			wantOk: false,
		},
		{
			name:   "partial PI",
			input:  "<?",
			wantOk: true,
			wantKind: TokenPIStart,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := tokenizer.NewStream(tt.input)
			matcher := PIAndXMLDeclMatcher()
			token := matcher(stream)

			if tt.wantOk {
				if token == nil {
					t.Errorf("expected token, got nil")
					return
				}
				if token.Kind() != tt.wantKind {
					t.Errorf("expected %s, got %s", tt.wantKind, token.Kind())
				}
			} else {
				if token != nil {
					t.Errorf("expected nil token, got %v", token)
				}
			}
		})
	}
}

func TestEndTagOpenMatcher(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		wantOk bool
	}{
		{
			name:   "valid end tag",
			input:  "</root>",
			wantOk: true,
		},
		{
			name:   "not end tag",
			input:  "<root>",
			wantOk: false,
		},
		{
			name:   "partial",
			input:  "<",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := tokenizer.NewStream(tt.input)
			matcher := EndTagOpenMatcher()
			token := matcher(stream)

			if tt.wantOk {
				if token == nil {
					t.Errorf("expected token, got nil")
					return
				}
				if token.Kind() != TokenEndTagOpen {
					t.Errorf("expected TokenEndTagOpen, got %s", token.Kind())
				}
			} else {
				if token != nil {
					t.Errorf("expected nil token, got %v", token)
				}
			}
		})
	}
}

func TestTagSelfCloseMatcher(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		wantOk bool
	}{
		{
			name:   "valid self-close",
			input:  "/>",
			wantOk: true,
		},
		{
			name:   "not self-close",
			input:  ">",
			wantOk: false,
		},
		{
			name:   "wrong order",
			input:  "</",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := tokenizer.NewStream(tt.input)
			matcher := TagSelfCloseMatcher()
			token := matcher(stream)

			if tt.wantOk {
				if token == nil {
					t.Errorf("expected token, got nil")
					return
				}
				if token.Kind() != TokenTagSelfClose {
					t.Errorf("expected TokenTagSelfClose, got %s", token.Kind())
				}
			} else {
				if token != nil {
					t.Errorf("expected nil token, got %v", token)
				}
			}
		})
	}
}

func TestStringMatcher(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		wantOk bool
		want   string
	}{
		{
			name:   "double quoted string",
			input:  `"hello"`,
			wantOk: true,
			want:   `"hello"`,
		},
		{
			name:   "single quoted string",
			input:  `'world'`,
			wantOk: true,
			want:   `'world'`,
		},
		{
			name:   "empty string",
			input:  `""`,
			wantOk: true,
			want:   `""`,
		},
		{
			name:   "string with spaces",
			input:  `"hello world"`,
			wantOk: true,
			want:   `"hello world"`,
		},
		{
			name:   "not a string",
			input:  `hello`,
			wantOk: false,
		},
		{
			name:   "unterminated string",
			input:  `"hello`,
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := tokenizer.NewStream(tt.input)
			matcher := StringMatcher()
			token := matcher(stream)

			if tt.wantOk {
				if token == nil {
					t.Errorf("expected token, got nil")
					return
				}
				if token.Kind() != TokenString {
					t.Errorf("expected TokenString, got %s", token.Kind())
				}
				got := string(token.Value())
				if got != tt.want {
					t.Errorf("expected %q, got %q", tt.want, got)
				}
			} else {
				if token != nil {
					t.Errorf("expected nil token, got %v", token)
				}
			}
		})
	}
}

func TestNameMatcher(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		wantOk bool
		want   string
	}{
		{
			name:   "simple name",
			input:  "element",
			wantOk: true,
			want:   "element",
		},
		{
			name:   "name with underscore",
			input:  "_private",
			wantOk: true,
			want:   "_private",
		},
		{
			name:   "namespaced element",
			input:  "ns:element",
			wantOk: true,
			want:   "ns:element",
		},
		{
			name:   "name with dash",
			input:  "my-element",
			wantOk: true,
			want:   "my-element",
		},
		{
			name:   "name with dot",
			input:  "element.name",
			wantOk: true,
			want:   "element.name",
		},
		{
			name:   "name with numbers",
			input:  "element123",
			wantOk: true,
			want:   "element123",
		},
		{
			name:   "starts with number - invalid",
			input:  "123element",
			wantOk: false,
		},
		{
			name:   "starts with dash - invalid",
			input:  "-element",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := tokenizer.NewStream(tt.input)
			matcher := NameMatcher()
			token := matcher(stream)

			if tt.wantOk {
				if token == nil {
					t.Errorf("expected token, got nil")
					return
				}
				if token.Kind() != TokenName {
					t.Errorf("expected TokenName, got %s", token.Kind())
				}
				got := string(token.Value())
				if got != tt.want {
					t.Errorf("expected %q, got %q", tt.want, got)
				}
			} else {
				if token != nil {
					t.Errorf("expected nil token, got %v", token)
				}
			}
		})
	}
}

func TestTextMatcher(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		wantOk bool
		want   string
	}{
		{
			name:   "simple text",
			input:  "hello world",
			wantOk: true,
			want:   "hello world",
		},
		{
			name:   "text before tag",
			input:  "hello<tag>",
			wantOk: true,
			want:   "hello",
		},
		{
			name:   "starts with tag",
			input:  "<tag>hello",
			wantOk: false,
		},
		{
			name:   "empty text",
			input:  "",
			wantOk: false,
		},
		{
			name:   "whitespace text",
			input:  "   ",
			wantOk: true,
			want:   "   ",
		},
		{
			name:   "text with newlines",
			input:  "line1\nline2",
			wantOk: true,
			want:   "line1\nline2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := tokenizer.NewStream(tt.input)
			matcher := TextMatcher()
			token := matcher(stream)

			if tt.wantOk {
				if token == nil {
					t.Errorf("expected token, got nil")
					return
				}
				if token.Kind() != TokenText {
					t.Errorf("expected TokenText, got %s", token.Kind())
				}
				got := string(token.Value())
				if got != tt.want {
					t.Errorf("expected %q, got %q", tt.want, got)
				}
			} else {
				if token != nil {
					t.Errorf("expected nil token, got %v", token)
				}
			}
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	t.Run("isNameStartChar", func(t *testing.T) {
		tests := []struct {
			r    rune
			want bool
		}{
			{'A', true},
			{'Z', true},
			{'a', true},
			{'z', true},
			{'_', true},
			{':', true},
			{'0', false},
			{'-', false},
			{'.', false},
			{'@', false},
		}

		for _, tt := range tests {
			got := isNameStartChar(tt.r)
			if got != tt.want {
				t.Errorf("isNameStartChar(%q) = %v, want %v", tt.r, got, tt.want)
			}
		}
	})

	t.Run("isNameChar", func(t *testing.T) {
		tests := []struct {
			r    rune
			want bool
		}{
			{'A', true},
			{'a', true},
			{'_', true},
			{':', true},
			{'0', true},
			{'9', true},
			{'-', true},
			{'.', true},
			{'@', false},
			{'!', false},
		}

		for _, tt := range tests {
			got := isNameChar(tt.r)
			if got != tt.want {
				t.Errorf("isNameChar(%q) = %v, want %v", tt.r, got, tt.want)
			}
		}
	})

	t.Run("isNameStartByte", func(t *testing.T) {
		tests := []struct {
			b    byte
			want bool
		}{
			{'A', true},
			{'Z', true},
			{'a', true},
			{'z', true},
			{'_', true},
			{':', true},
			{'0', false},
			{'-', false},
		}

		for _, tt := range tests {
			got := isNameStartByte(tt.b)
			if got != tt.want {
				t.Errorf("isNameStartByte(%q) = %v, want %v", tt.b, got, tt.want)
			}
		}
	})

	t.Run("isNameByte", func(t *testing.T) {
		tests := []struct {
			b    byte
			want bool
		}{
			{'A', true},
			{'a', true},
			{'_', true},
			{':', true},
			{'0', true},
			{'9', true},
			{'-', true},
			{'.', true},
			{'@', false},
		}

		for _, tt := range tests {
			got := isNameByte(tt.b)
			if got != tt.want {
				t.Errorf("isNameByte(%q) = %v, want %v", tt.b, got, tt.want)
			}
		}
	})
}

func TestMatchString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		match string
		want  bool
	}{
		{
			name:  "exact match",
			input: "<?xml",
			match: "<?xml",
			want:  true,
		},
		{
			name:  "prefix match",
			input: "<?xml version",
			match: "<?xml",
			want:  true,
		},
		{
			name:  "no match",
			input: "<root>",
			match: "<?xml",
			want:  false,
		},
		{
			name:  "partial match fails",
			input: "<?xm",
			match: "<?xml",
			want:  false,
		},
		{
			name:  "empty match",
			input: "test",
			match: "",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := tokenizer.NewStream(tt.input)
			got := matchString(stream, tt.match)
			if got != tt.want {
				t.Errorf("matchString(%q, %q) = %v, want %v", tt.input, tt.match, got, tt.want)
			}

			// Verify stream position
			if tt.want {
				// Should have advanced
				remaining := ""
				for {
					r, ok := stream.NextChar()
					if !ok {
						break
					}
					remaining += string(r)
				}
				expected := tt.input[len(tt.match):]
				if remaining != expected {
					t.Errorf("stream position incorrect: got %q, want %q", remaining, expected)
				}
			}
		})
	}
}

func TestTokenizerIntegration(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantKinds []string
	}{
		{
			name:  "simple element",
			input: "<root/>",
			wantKinds: []string{TokenTagOpen, TokenName, TokenTagSelfClose},
		},
		{
			name:  "element with text",
			input: "<p>Hello</p>",
			// Note: Text "Hello" is tokenized as TokenName when inside tags, not TokenText
			wantKinds: []string{TokenTagOpen, TokenName, TokenTagClose, TokenName, TokenEndTagOpen, TokenName, TokenTagClose},
		},
		{
			name:  "element with attribute",
			input: `<div id="main">`,
			wantKinds: []string{TokenTagOpen, TokenName, TokenName, TokenEquals, TokenString, TokenTagClose},
		},
		{
			name:  "XML declaration",
			input: `<?xml version="1.0"?>`,
			wantKinds: []string{TokenXMLDeclStart, TokenName, TokenEquals, TokenString, TokenPIEnd},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := tokenizer.NewStream(tt.input)
			tok := NewTokenizerWithStream(stream)

			var gotKinds []string
			for {
				token, ok := tok.NextToken()
				if !ok {
					break
				}
				if token.Kind() == TokenEOF {
					break
				}
				// Skip whitespace tokens for cleaner test expectations
				if token.Kind() == "Whitespace" {
					continue
				}
				gotKinds = append(gotKinds, token.Kind())
			}

			if len(gotKinds) != len(tt.wantKinds) {
				t.Errorf("got %d tokens, want %d\nGot: %v\nWant: %v",
					len(gotKinds), len(tt.wantKinds), gotKinds, tt.wantKinds)
				return
			}

			for i, want := range tt.wantKinds {
				if gotKinds[i] != want {
					t.Errorf("token %d: got %s, want %s", i, gotKinds[i], want)
				}
			}
		})
	}
}
