package tokenizer

import (
	"github.com/shapestone/shape-core/pkg/tokenizer"
)

// NewTokenizer creates a tokenizer for XML format.
// The tokenizer uses a state-based approach to handle XML's context-sensitive nature.
//
// XML requires different tokenization depending on context:
// 1. Outside tags: look for <, text content
// 2. Inside tags: look for element names, attributes, >, />
// 3. Inside CDATA: look for ]]>
// 4. Inside comments: look for -->
func NewTokenizer() tokenizer.Tokenizer {
	return tokenizer.NewTokenizer(
		// Comments (must be before < to avoid conflict)
		CommentMatcher(),

		// CDATA sections
		CDataMatcher(),

		// Processing instructions and XML declaration
		PIAndXMLDeclMatcher(),

		// Processing instruction end
		tokenizer.StringMatcherFunc(TokenPIEnd, "?>"),

		// Tag structures
		EndTagOpenMatcher(),         // </ (before <)
		TagSelfCloseMatcher(),       // />
		tokenizer.StringMatcherFunc(TokenTagOpen, "<"),
		tokenizer.StringMatcherFunc(TokenTagClose, ">"),
		tokenizer.StringMatcherFunc(TokenEquals, "="),

		// Strings (attribute values)
		StringMatcher(),

		// Names (element/attribute names)
		// Names can only appear after < or = or whitespace within tags
		// For simplicity, match names before text
		NameMatcher(),

		// Text content (must be last, matches everything else)
		TextMatcher(),
	)
}

// NewTokenizerWithStream creates a tokenizer for XML format using a pre-configured stream.
// This is used internally to support streaming from io.Reader.
func NewTokenizerWithStream(stream tokenizer.Stream) tokenizer.Tokenizer {
	tok := NewTokenizer()
	tok.InitializeFromStream(stream)
	return tok
}

// CommentMatcher creates a matcher for XML comments.
// Matches: <!-- ... -->
func CommentMatcher() tokenizer.Matcher {
	return func(stream tokenizer.Stream) *tokenizer.Token {
		// Check for <!--
		if !matchString(stream, "<!--") {
			return nil
		}

		// Find -->
		var content []rune
		for {
			r, ok := stream.PeekChar()
			if !ok {
				return nil // Unterminated comment
			}

			// Check for -->
			if r == '-' {
				checkpoint := stream.Clone()
				if matchString(stream, "-->") {
					// Return comment token (we could return the content if needed)
					return tokenizer.NewToken(TokenCommentStart, []rune("<!--"))
				}
				// Reset and continue
				stream.Match(checkpoint)
			}

			stream.NextChar()
			content = append(content, r)
		}
	}
}

// CDataMatcher creates a matcher for CDATA sections.
// Matches: <![CDATA[ ... ]]>
func CDataMatcher() tokenizer.Matcher {
	return func(stream tokenizer.Stream) *tokenizer.Token {
		if !matchString(stream, "<![CDATA[") {
			return nil
		}

		// Return CDATA start token
		return tokenizer.NewToken(TokenCDataStart, []rune("<![CDATA["))
	}
}

// PIAndXMLDeclMatcher creates a matcher for processing instructions and XML declarations.
// Matches: <?xml ... ?> or <? ... ?>
func PIAndXMLDeclMatcher() tokenizer.Matcher {
	return func(stream tokenizer.Stream) *tokenizer.Token {
		checkpoint := stream.Clone()

		if !matchString(stream, "<?") {
			return nil
		}

		// Check if it's an XML declaration
		if matchString(stream, "xml") {
			stream.Match(checkpoint)
			if matchString(stream, "<?xml") {
				return tokenizer.NewToken(TokenXMLDeclStart, []rune("<?xml"))
			}
		}

		// Reset and return as PI start
		stream.Match(checkpoint)
		if matchString(stream, "<?") {
			return tokenizer.NewToken(TokenPIStart, []rune("<?"))
		}

		return nil
	}
}

// EndTagOpenMatcher creates a matcher for end tag opening.
// Matches: </
func EndTagOpenMatcher() tokenizer.Matcher {
	return func(stream tokenizer.Stream) *tokenizer.Token {
		if matchString(stream, "</") {
			return tokenizer.NewToken(TokenEndTagOpen, []rune("</"))
		}
		return nil
	}
}

// TagSelfCloseMatcher creates a matcher for self-closing tag syntax.
// Matches: />
func TagSelfCloseMatcher() tokenizer.Matcher {
	return func(stream tokenizer.Stream) *tokenizer.Token {
		if matchString(stream, "/>") {
			return tokenizer.NewToken(TokenTagSelfClose, []rune("/>"))
		}
		return nil
	}
}

// StringMatcher creates a matcher for XML string literals (attribute values).
// Matches: "..." or '...'
func StringMatcher() tokenizer.Matcher {
	return func(stream tokenizer.Stream) *tokenizer.Token {
		r, ok := stream.PeekChar()
		if !ok {
			return nil
		}

		// Check for opening quote
		var quote rune
		if r == '"' || r == '\'' {
			quote = r
		} else {
			return nil
		}

		var value []rune
		value = append(value, quote)
		stream.NextChar()

		// Read until closing quote
		for {
			r, ok := stream.NextChar()
			if !ok {
				return nil // Unterminated string
			}

			value = append(value, r)

			if r == quote {
				return tokenizer.NewToken(TokenString, value)
			}

			// Handle escape sequences if needed
			if r == '\\' {
				escaped, ok := stream.NextChar()
				if !ok {
					return nil
				}
				value = append(value, escaped)
			}
		}
	}
}

// NameMatcher creates a matcher for XML names (element and attribute names).
// Matches: [A-Za-z_:][A-Za-z0-9_:.-]*
// Supports namespaces with colon (e.g., "ns:element")
func NameMatcher() tokenizer.Matcher {
	return func(stream tokenizer.Stream) *tokenizer.Token {
		r, ok := stream.PeekChar()
		if !ok {
			return nil
		}

		// First character must be letter, underscore, or colon
		if !isNameStartChar(r) {
			return nil
		}

		var value []rune
		for {
			r, ok := stream.PeekChar()
			if !ok {
				break
			}

			if !isNameChar(r) {
				break
			}

			stream.NextChar()
			value = append(value, r)
		}

		if len(value) == 0 {
			return nil
		}

		return tokenizer.NewToken(TokenName, value)
	}
}

// TextMatcher creates a matcher for text content between tags.
// Matches any text until < is encountered.
func TextMatcher() tokenizer.Matcher {
	return func(stream tokenizer.Stream) *tokenizer.Token {
		r, ok := stream.PeekChar()
		if !ok {
			return nil
		}

		// Text cannot start with <
		if r == '<' {
			return nil
		}

		var value []rune
		for {
			r, ok := stream.PeekChar()
			if !ok {
				break
			}

			// Stop at tag opening
			if r == '<' {
				break
			}

			stream.NextChar()
			value = append(value, r)
		}

		if len(value) == 0 {
			return nil
		}

		return tokenizer.NewToken(TokenText, value)
	}
}

// Helper functions

// matchString attempts to match a specific string at the current position.
// Returns true and advances if match succeeds, returns false otherwise.
func matchString(stream tokenizer.Stream, s string) bool {
	checkpoint := stream.Clone()

	for _, expected := range s {
		r, ok := stream.NextChar()
		if !ok || r != expected {
			stream.Match(checkpoint)
			return false
		}
	}

	return true
}

// isNameStartChar returns true if r can start an XML name.
// XML spec: [A-Za-z_:] plus Unicode letters
func isNameStartChar(r rune) bool {
	return (r >= 'A' && r <= 'Z') ||
		(r >= 'a' && r <= 'z') ||
		r == '_' ||
		r == ':'
}

// isNameChar returns true if r can appear in an XML name.
// XML spec: NameStartChar plus [0-9.-]
func isNameChar(r rune) bool {
	return isNameStartChar(r) ||
		(r >= '0' && r <= '9') ||
		r == '.' ||
		r == '-'
}
