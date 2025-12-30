package tokenizer

import (
	"testing"

	"github.com/shapestone/shape-core/pkg/tokenizer"
)

// TestRuneMatchers tests the rune-based fallback matchers with non-ASCII input.
// These are used when the ByteStream optimization isn't available.
func TestRuneMatchersWithNonASCII(t *testing.T) {
	// Create a stream that will use rune-based processing
	// Use characters that require multi-byte UTF-8 encoding
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "unicode string",
			input: `"hello 世界"`,
		},
		{
			name:  "unicode name",
			input: `element世界`,
		},
		{
			name:  "unicode text",
			input: `Text with 日本語 characters`,
		},
		{
			name:  "emoji in string",
			input: `"test 😀 emoji"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := tokenizer.NewStream(tt.input)

			// Try string matcher
			stringMatcher := StringMatcher()
			if token := stringMatcher(stream); token != nil {
				t.Logf("String matcher matched: %s", string(token.Value()))
			}

			// Reset stream
			stream = tokenizer.NewStream(tt.input)

			// Try name matcher
			nameMatcher := NameMatcher()
			if token := nameMatcher(stream); token != nil {
				t.Logf("Name matcher matched: %s", string(token.Value()))
			}

			// Reset stream
			stream = tokenizer.NewStream(tt.input)

			// Try text matcher
			textMatcher := TextMatcher()
			if token := textMatcher(stream); token != nil {
				t.Logf("Text matcher matched: %s", string(token.Value()))
			}
		})
	}
}

// TestStringMatcherRune tests the rune-based string matcher directly
func TestStringMatcherRune(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		wantOk bool
		want   string
	}{
		{
			name:   "double quoted with escape",
			input:  `"hello\"world"`,
			wantOk: true,
			want:   `"hello\"world"`,
		},
		{
			name:   "single quoted with escape",
			input:  `'hello\'world'`,
			wantOk: true,
			want:   `'hello\'world'`,
		},
		{
			name:   "backslash escape",
			input:  `"path\\to\\file"`,
			wantOk: true,
			want:   `"path\\to\\file"`,
		},
		{
			name:   "unicode characters",
			input:  `"日本語"`,
			wantOk: true,
			want:   `"日本語"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := tokenizer.NewStream(tt.input)
			token := stringMatcherRune(stream)

			if tt.wantOk {
				if token == nil {
					t.Errorf("expected token, got nil")
					return
				}
				got := string(token.Value())
				if got != tt.want {
					t.Errorf("got %q, want %q", got, tt.want)
				}
			} else {
				if token != nil {
					t.Errorf("expected nil, got %v", token)
				}
			}
		})
	}
}

// TestNameMatcherRune tests the rune-based name matcher directly
func TestNameMatcherRune(t *testing.T) {
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
			name:   "namespaced",
			input:  "ns:element",
			wantOk: true,
			want:   "ns:element",
		},
		{
			name:   "with dash",
			input:  "my-element",
			wantOk: true,
			want:   "my-element",
		},
		{
			name:   "starts with colon",
			input:  ":element",
			wantOk: true,
			want:   ":element",
		},
		{
			name:   "starts with underscore",
			input:  "_private",
			wantOk: true,
			want:   "_private",
		},
		{
			name:   "empty stream",
			input:  "",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := tokenizer.NewStream(tt.input)
			token := nameMatcherRune(stream)

			if tt.wantOk {
				if token == nil {
					t.Errorf("expected token, got nil")
					return
				}
				got := string(token.Value())
				if got != tt.want {
					t.Errorf("got %q, want %q", got, tt.want)
				}
			} else {
				if token != nil {
					t.Errorf("expected nil, got %v", token)
				}
			}
		})
	}
}

// TestTextMatcherRune tests the rune-based text matcher directly
func TestTextMatcherRune(t *testing.T) {
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
			input:  "text<tag>",
			wantOk: true,
			want:   "text",
		},
		{
			name:   "starts with tag",
			input:  "<tag>",
			wantOk: false,
		},
		{
			name:   "unicode text",
			input:  "こんにちは",
			wantOk: true,
			want:   "こんにちは",
		},
		{
			name:   "empty",
			input:  "",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := tokenizer.NewStream(tt.input)
			token := textMatcherRune(stream)

			if tt.wantOk {
				if token == nil {
					t.Errorf("expected token, got nil")
					return
				}
				got := string(token.Value())
				if got != tt.want {
					t.Errorf("got %q, want %q", got, tt.want)
				}
			} else {
				if token != nil {
					t.Errorf("expected nil, got %v", token)
				}
			}
		})
	}
}
