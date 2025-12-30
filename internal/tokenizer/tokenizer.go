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
		for {
			r, ok := stream.PeekChar()
			if !ok {
				return nil // Unterminated comment
			}

			// Check for -->
			if r == '-' {
				savedLoc := stream.GetLocation()
				if matchString(stream, "-->") {
					// Return comment token
					return tokenizer.NewToken(TokenCommentStart, []rune("<!--"))
				}
				// Reset and continue
				stream.SetLocation(savedLoc)
			}

			stream.NextChar()
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
		savedLoc := stream.GetLocation()

		if !matchString(stream, "<?") {
			return nil
		}

		// Check if it's an XML declaration
		if matchString(stream, "xml") {
			stream.SetLocation(savedLoc)
			if matchString(stream, "<?xml") {
				return tokenizer.NewToken(TokenXMLDeclStart, []rune("<?xml"))
			}
		}

		// Reset and return as PI start
		stream.SetLocation(savedLoc)
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
// Uses ByteStream fast path with SWAR for optimal performance on ASCII strings.
func StringMatcher() tokenizer.Matcher {
	return func(stream tokenizer.Stream) *tokenizer.Token {
		// Try ByteStream fast path for ASCII strings
		if byteStream, ok := stream.(tokenizer.ByteStream); ok {
			return stringMatcherByte(byteStream)
		}

		// Fallback to rune-based matcher
		return stringMatcherRune(stream)
	}
}

// stringMatcherByte uses ByteStream for optimal string matching.
func stringMatcherByte(stream tokenizer.ByteStream) *tokenizer.Token {
	b, ok := stream.PeekByte()
	if !ok {
		return nil
	}

	// Check for opening quote
	var quote byte
	if b == '"' || b == '\'' {
		quote = b
	} else {
		return nil
	}

	startPos := stream.BytePosition()
	stream.NextByte() // consume opening quote

	// Use SWAR to find closing quote quickly
	remaining := stream.RemainingBytes()
	offset := tokenizer.FindByte(remaining, quote)

	if offset == -1 {
		// No closing quote found
		return nil
	}

	// Advance to the quote position
	for i := 0; i < offset; i++ {
		stream.NextByte()
	}

	// Consume the closing quote
	stream.NextByte()

	// Extract the string value
	value := stream.SliceFrom(startPos)
	return tokenizer.NewToken(TokenString, []rune(string(value)))
}

// stringMatcherRune is the fallback rune-based implementation.
func stringMatcherRune(stream tokenizer.Stream) *tokenizer.Token {
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

// NameMatcher creates a matcher for XML names (element and attribute names).
// Matches: [A-Za-z_:][A-Za-z0-9_:.-]*
// Supports namespaces with colon (e.g., "ns:element")
// Uses ByteStream fast path for optimal performance on ASCII names.
func NameMatcher() tokenizer.Matcher {
	return func(stream tokenizer.Stream) *tokenizer.Token {
		// Try ByteStream fast path for ASCII names
		if byteStream, ok := stream.(tokenizer.ByteStream); ok {
			return nameMatcherByte(byteStream)
		}

		// Fallback to rune-based matcher
		return nameMatcherRune(stream)
	}
}

// nameMatcherByte uses ByteStream for optimal name scanning.
func nameMatcherByte(stream tokenizer.ByteStream) *tokenizer.Token {
	b, ok := stream.PeekByte()
	if !ok {
		return nil
	}

	// First character must be letter, underscore, or colon
	if !isNameStartByte(b) {
		return nil
	}

	startPos := stream.BytePosition()

	// Consume name characters
	for {
		b, ok := stream.PeekByte()
		if !ok {
			break
		}

		if !isNameByte(b) {
			break
		}

		stream.NextByte()
	}

	// Extract the name value
	value := stream.SliceFrom(startPos)
	if len(value) == 0 {
		return nil
	}

	return tokenizer.NewToken(TokenName, []rune(string(value)))
}

// nameMatcherRune is the fallback rune-based implementation.
func nameMatcherRune(stream tokenizer.Stream) *tokenizer.Token {
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

// TextMatcher creates a matcher for text content between tags.
// Matches any text until < is encountered.
// Uses ByteStream fast path with SWAR for optimal performance on ASCII text.
func TextMatcher() tokenizer.Matcher {
	return func(stream tokenizer.Stream) *tokenizer.Token {
		// Try ByteStream fast path for ASCII text
		if byteStream, ok := stream.(tokenizer.ByteStream); ok {
			return textMatcherByte(byteStream)
		}

		// Fallback to rune-based matcher
		return textMatcherRune(stream)
	}
}

// textMatcherByte uses ByteStream + SWAR for optimal text scanning.
func textMatcherByte(stream tokenizer.ByteStream) *tokenizer.Token {
	b, ok := stream.PeekByte()
	if !ok {
		return nil
	}

	// Text cannot start with <
	if b == '<' {
		return nil
	}

	startPos := stream.BytePosition()

	// Use SWAR to find < delimiter quickly (8 bytes at a time)
	remaining := stream.RemainingBytes()
	offset := tokenizer.FindByte(remaining, '<')

	if offset == -1 {
		// No < found - consume all remaining bytes
		for {
			_, ok := stream.NextByte()
			if !ok {
				break
			}
		}
	} else if offset == 0 {
		// Starts with <, no text content
		return nil
	} else {
		// Advance to the < position
		for i := 0; i < offset; i++ {
			stream.NextByte()
		}
	}

	// Extract the text value
	value := stream.SliceFrom(startPos)
	if len(value) == 0 {
		return nil
	}

	return tokenizer.NewToken(TokenText, []rune(string(value)))
}

// textMatcherRune is the fallback rune-based implementation.
func textMatcherRune(stream tokenizer.Stream) *tokenizer.Token {
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

// Helper functions

// matchString attempts to match a specific string at the current position.
// Returns true and advances if match succeeds, returns false otherwise.
// Uses GetLocation/SetLocation instead of Clone() to avoid allocations.
func matchString(stream tokenizer.Stream, s string) bool {
	savedLoc := stream.GetLocation()

	for _, expected := range s {
		r, ok := stream.NextChar()
		if !ok || r != expected {
			stream.SetLocation(savedLoc)
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

// isNameStartByte returns true if b can start an XML name (byte version).
func isNameStartByte(b byte) bool {
	return (b >= 'A' && b <= 'Z') ||
		(b >= 'a' && b <= 'z') ||
		b == '_' ||
		b == ':'
}

// isNameByte returns true if b can appear in an XML name (byte version).
func isNameByte(b byte) bool {
	return isNameStartByte(b) ||
		(b >= '0' && b <= '9') ||
		b == '.' ||
		b == '-'
}
