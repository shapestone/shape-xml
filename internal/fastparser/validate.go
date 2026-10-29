package fastparser

import (
	"bytes"
	"fmt"
	"strings"
	"unicode/utf8"
)

// Validate checks that data is well-formed XML and returns the first problem
// found, with its line and column. It builds no tree and keeps no state beyond
// a stack of open element names, so deep nesting costs heap, not call stack.
//
// The verdicts aim to follow a token loop over encoding/xml's Decoder (Strict
// mode, no CharsetReader): the input is a stream of tokens, and Validate does
// not require exactly one root element, nor any element at all. The deliberate
// deviation is that references to surrogate characters are rejected. See
// pkg/xml.Validate for the full list.
//
// Accepted, as in encoding/xml: any number of root elements, empty input, text
// outside the root, duplicate attributes, undeclared namespace prefixes, a
// leading byte order mark, and a DOCTYPE (or any other <!...> declaration).
//
// Rejected: unknown entity references (only lt, gt, amp, apos and quot exist),
// character references to characters XML forbids (&#0;, surrogates, ...),
// characters outside the XML Char production (NUL, most C0 controls, U+FFFE,
// U+FFFF) and invalid UTF-8 in text, CDATA and attribute values, '<' in an
// attribute value, unquoted attributes, mismatched or unclosed tags, a
// version other than 1.0, and an XML declaration naming an encoding other than
// UTF-8 (there is no charset hook).
//
// DOCTYPE and other declarations are skipped, never interpreted: internal
// subset entities are not defined, parameter entities are not expanded, and
// SYSTEM/PUBLIC identifiers are not read, so no file or network access can
// happen. A reference to an entity declared in a DOCTYPE is an unknown entity.
func Validate(data []byte) error {
	v := validator{data: data}
	return v.run()
}

type validator struct {
	data  []byte
	pos   int
	stack [][2]int // [start, end) of each open element's name in data
}

func (v *validator) run() error {
	n := len(v.data)
	for v.pos < n {
		if v.data[v.pos] != '<' {
			if err := v.text(-1); err != nil {
				return err
			}
			continue
		}
		v.pos++
		if v.pos >= n {
			return v.errorf("unexpected EOF after '<'")
		}
		var err error
		switch v.data[v.pos] {
		case '/':
			v.pos++
			err = v.endTag()
		case '?':
			v.pos++
			err = v.procInst()
		case '!':
			v.pos++
			err = v.bang()
		default:
			err = v.startTag()
		}
		if err != nil {
			return err
		}
	}
	if len(v.stack) > 0 {
		top := v.stack[len(v.stack)-1]
		return v.errorf("unexpected EOF: element <%s> is not closed", v.data[top[0]:top[1]])
	}
	return nil
}

func (v *validator) startTag() error {
	nameStart := v.pos
	name, ok := v.readQName()
	if !ok {
		return v.errorf("expected element name after <")
	}
	selfClosing := false
	for {
		v.skipSpace()
		if v.pos >= len(v.data) {
			return v.errorf("unexpected EOF in element <%s>", name)
		}
		c := v.data[v.pos]
		if c == '/' {
			v.pos++
			if v.pos >= len(v.data) || v.data[v.pos] != '>' {
				return v.errorf("expected /> in element <%s>", name)
			}
			v.pos++
			selfClosing = true
			break
		}
		if c == '>' {
			v.pos++
			break
		}
		if _, ok := v.readQName(); !ok {
			return v.errorf("expected attribute name in element <%s>", name)
		}
		v.skipSpace()
		if v.pos >= len(v.data) || v.data[v.pos] != '=' {
			return v.errorf("attribute name without = in element <%s>", name)
		}
		v.pos++
		v.skipSpace()
		if v.pos >= len(v.data) || (v.data[v.pos] != '"' && v.data[v.pos] != '\'') {
			return v.errorf("unquoted or missing attribute value in element <%s>", name)
		}
		quote := int(v.data[v.pos])
		v.pos++
		if err := v.text(quote); err != nil {
			return err
		}
	}
	if !selfClosing {
		v.stack = append(v.stack, [2]int{nameStart, nameStart + len(name)})
	}
	return nil
}

func (v *validator) endTag() error {
	name, ok := v.readQName()
	if !ok {
		return v.errorf("expected element name after </")
	}
	v.skipSpace()
	if v.pos >= len(v.data) || v.data[v.pos] != '>' {
		return v.errorf("invalid characters between </%s and >", name)
	}
	v.pos++
	if len(v.stack) == 0 {
		return v.errorf("unexpected end element </%s>", name)
	}
	top := v.stack[len(v.stack)-1]
	if open := v.data[top[0]:top[1]]; !bytes.Equal(open, name) {
		return v.errorf("element <%s> closed by </%s>", open, name)
	}
	v.stack = v.stack[:len(v.stack)-1]
	return nil
}

// procInst handles <?target ...?>, with v.pos just after "<?".
func (v *validator) procInst() error {
	target, ok := v.readName()
	if !ok {
		return v.errorf("expected target name after <?")
	}
	start := v.pos
	var b0 byte
	for {
		if v.pos >= len(v.data) {
			return v.errorf("unexpected EOF in processing instruction")
		}
		b := v.data[v.pos]
		v.pos++
		if b0 == '?' && b == '>' {
			break
		}
		b0 = b
	}
	if string(target) != "xml" {
		return nil
	}
	content := string(v.data[start : v.pos-2])
	if ver := procInstParam("version", content); ver != "" && ver != "1.0" {
		return v.errorf("unsupported version: %s; only version 1.0 is supported", ver)
	}
	if enc := procInstParam("encoding", content); enc != "" && !equalFoldASCII(enc, "utf-8") {
		return v.errorf("unsupported encoding %q: only UTF-8 is supported", enc)
	}
	return nil
}

// procInstParam extracts a pseudo-attribute from an XML declaration the way
// encoding/xml does: it looks for param followed by '=' and a quote character,
// and keeps searching past occurrences where the '=' is not followed by a quote
// (`version=1.version="1.0"` yields "1.0"). It is not a full parse of the
// declaration.
func procInstParam(param, s string) string {
	param += "="
	i := 0
	var quote byte
	for i < len(s) {
		sub := s[i:]
		k := strings.Index(sub, param)
		if k < 0 || len(param)+k >= len(sub) {
			return ""
		}
		i += len(param) + k + 1
		if c := sub[len(param)+k]; c == '\'' || c == '"' {
			quote = c
			break
		}
	}
	if quote == 0 {
		return ""
	}
	end := strings.IndexByte(s[i:], quote)
	if end < 0 {
		return ""
	}
	return s[i : i+end]
}

func equalFoldASCII(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		x, y := a[i], b[i]
		if 'A' <= x && x <= 'Z' {
			x += 'a' - 'A'
		}
		if 'A' <= y && y <= 'Z' {
			y += 'a' - 'A'
		}
		if x != y {
			return false
		}
	}
	return true
}

// bang handles <!...>, with v.pos just after "<!".
func (v *validator) bang() error {
	if v.pos >= len(v.data) {
		return v.errorf("unexpected EOF after <!")
	}
	switch v.data[v.pos] {
	case '-':
		if !bytes.HasPrefix(v.data[v.pos:], []byte("--")) {
			return v.errorf("invalid sequence <!- not part of <!--")
		}
		v.pos += 2
		return v.comment()
	case '[':
		if !bytes.HasPrefix(v.data[v.pos:], []byte("[CDATA[")) {
			return v.errorf("invalid <![ sequence")
		}
		v.pos += len("[CDATA[")
		return v.cdata()
	}
	return v.directive()
}

// comment skips a comment body: "--" may only appear as the start of "-->".
func (v *validator) comment() error {
	var b0, b1 byte
	for {
		if v.pos >= len(v.data) {
			return v.errorf("unexpected EOF in comment")
		}
		b := v.data[v.pos]
		v.pos++
		if b0 == '-' && b1 == '-' {
			if b != '>' {
				return v.errorf(`invalid sequence "--" not allowed in comments`)
			}
			return nil
		}
		b0, b1 = b1, b
	}
}

// cdata checks a CDATA section body up to "]]>".
func (v *validator) cdata() error {
	for v.pos < len(v.data) {
		if bytes.HasPrefix(v.data[v.pos:], []byte("]]>")) {
			v.pos += 3
			return nil
		}
		if err := v.char(); err != nil {
			return err
		}
	}
	return v.errorf("unexpected EOF in CDATA section")
}

// directive skips <!DOCTYPE ...>, <!ENTITY ...> and similar declarations
// without interpreting them. It only tracks nested <> pairs, quoted strings
// and comments so it can find the closing '>'. Like encoding/xml it does not
// look at the first byte after "<!".
func (v *validator) directive() error {
	v.pos++ // first byte after "<!" is not examined
	var inquote byte
	depth := 0
	for {
		if v.pos >= len(v.data) {
			return v.errorf("unexpected EOF in declaration")
		}
		b := v.data[v.pos]
		v.pos++
		if inquote == 0 && b == '>' && depth == 0 {
			return nil
		}
		switch {
		case b == inquote:
			inquote = 0
		case inquote != 0:
		case b == '\'' || b == '"':
			inquote = b
		case b == '>':
			depth--
		case b == '<':
			if bytes.HasPrefix(v.data[v.pos:], []byte("!--")) {
				v.pos += 3
				var b0, b1 byte
				for {
					if v.pos >= len(v.data) {
						return v.errorf("unexpected EOF in comment inside declaration")
					}
					c := v.data[v.pos]
					v.pos++
					if b0 == '-' && b1 == '-' && c == '>' {
						break
					}
					b0, b1 = b1, c
				}
			} else {
				depth++
			}
		}
	}
}

// text checks character data. With quote < 0 it runs up to the next '<' (or
// EOF) and leaves v.pos on the '<'. With quote >= 0 it is an attribute value:
// it consumes the closing quote and rejects '<'.
func (v *validator) text(quote int) error {
	runStart := v.pos
	for v.pos < len(v.data) {
		b := v.data[v.pos]
		switch {
		case quote < 0 && b == '>' && v.pos-runStart >= 2 && v.data[v.pos-1] == ']' && v.data[v.pos-2] == ']':
			return v.errorf("unescaped ]]> not in CDATA section")
		case b == '<':
			if quote >= 0 {
				return v.errorf("unescaped < inside quoted string")
			}
			return nil
		case int(b) == quote:
			v.pos++
			return nil
		case b == '&':
			if err := v.entity(); err != nil {
				return err
			}
			runStart = v.pos
		default:
			if err := v.char(); err != nil {
				return err
			}
		}
	}
	if quote >= 0 {
		return v.errorf("unexpected EOF in quoted string")
	}
	return nil
}

// char consumes one character, which must be valid UTF-8 and an XML Char.
func (v *validator) char() error {
	b := v.data[v.pos]
	if b < utf8.RuneSelf {
		if !isXMLChar(rune(b)) {
			return v.errorf("illegal character code %U", rune(b))
		}
		v.pos++
		return nil
	}
	r, size := utf8.DecodeRune(v.data[v.pos:])
	if r == utf8.RuneError && size == 1 {
		return v.errorf("invalid UTF-8")
	}
	if !isXMLChar(r) {
		return v.errorf("illegal character code %U", r)
	}
	v.pos += size
	return nil
}

// entity consumes a reference starting at '&'. Only the five predefined
// entities and references to valid characters are known.
func (v *validator) entity() error {
	start := v.pos
	v.pos++ // '&'
	data := v.data
	if v.pos < len(data) && data[v.pos] == '#' {
		v.pos++
		base := uint64(10)
		if v.pos < len(data) && data[v.pos] == 'x' {
			base = 16
			v.pos++
		}
		digits := v.pos
		var n uint64
		overflow := false
		for v.pos < len(data) {
			d, ok := digitValue(data[v.pos], base)
			if !ok {
				break
			}
			if n > 0x10FFFF {
				overflow = true
			} else {
				n = n*base + d
			}
			v.pos++
		}
		if v.pos >= len(data) || data[v.pos] != ';' || v.pos == digits {
			return v.badEntity(start)
		}
		v.pos++
		if overflow || n > 0x10FFFF || !isXMLChar(rune(n)) {
			return v.errorAt(start, "invalid character reference %s", data[start:v.pos])
		}
		return nil
	}
	for _, name := range [...]string{"lt;", "gt;", "amp;", "apos;", "quot;"} {
		if bytes.HasPrefix(data[v.pos:], []byte(name)) {
			v.pos += len(name)
			return nil
		}
	}
	return v.badEntity(start)
}

func (v *validator) badEntity(start int) error {
	end := start + 1
	for end < len(v.data) && end-start < 32 && (isNameByte(v.data[end]) || v.data[end] == '#') {
		end++
	}
	if end < len(v.data) && v.data[end] == ';' {
		end++
	}
	return v.errorAt(start, "invalid character entity %s", v.data[start:end])
}

func digitValue(c byte, base uint64) (uint64, bool) {
	switch {
	case '0' <= c && c <= '9':
		return uint64(c - '0'), true
	case base == 16 && 'a' <= c && c <= 'f':
		return uint64(c-'a') + 10, true
	case base == 16 && 'A' <= c && c <= 'F':
		return uint64(c-'A') + 10, true
	}
	return 0, false
}

// readQName reads an element or attribute name. Like encoding/xml's nsname it
// refuses a name with more than one colon ("a:b:c"). Processing-instruction
// targets use readName, which has no such rule.
func (v *validator) readQName() ([]byte, bool) {
	start := v.pos
	name, ok := v.readName()
	if ok && bytes.Count(name, []byte(":")) > 1 {
		v.pos = start
		return nil, false
	}
	return name, ok
}

// readName reads an XML name at v.pos and returns it (a slice of v.data).
func (v *validator) readName() ([]byte, bool) {
	start := v.pos
	for v.pos < len(v.data) {
		c := v.data[v.pos]
		if c < utf8.RuneSelf && !isNameByte(c) {
			break
		}
		v.pos++
	}
	name := v.data[start:v.pos]
	if len(name) == 0 || !isName(name) {
		v.pos = start
		return nil, false
	}
	return name, true
}

func (v *validator) skipSpace() {
	for v.pos < len(v.data) {
		switch v.data[v.pos] {
		case ' ', '\t', '\r', '\n':
			v.pos++
		default:
			return
		}
	}
}

func isNameByte(c byte) bool {
	return 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || '0' <= c && c <= '9' ||
		c == '_' || c == ':' || c == '.' || c == '-'
}

// isName reports whether b is a Name as encoding/xml defines it.
func isName(b []byte) bool {
	for i := 0; len(b) > 0; i++ {
		r, size := utf8.DecodeRune(b)
		if r == utf8.RuneError && size == 1 {
			return false
		}
		if i == 0 && !isNameStart(r) || i > 0 && !isNameRest(r) {
			return false
		}
		b = b[size:]
	}
	return true
}

// isXMLChar reports whether r is in the XML 1.0 Char production.
func isXMLChar(r rune) bool {
	return r == 0x09 || r == 0x0A || r == 0x0D ||
		0x20 <= r && r <= 0xD7FF ||
		0xE000 <= r && r <= 0xFFFD ||
		0x10000 <= r && r <= 0x10FFFF
}

func (v *validator) errorf(format string, args ...interface{}) error {
	return v.errorAt(v.pos, format, args...)
}

// errorAt formats an error with the 1-indexed line and column of byte offset off.
func (v *validator) errorAt(off int, format string, args ...interface{}) error {
	if off > len(v.data) {
		off = len(v.data)
	}
	line, col := 1, 1
	for i := 0; i < off; i++ {
		if v.data[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return fmt.Errorf("xml: line %d, column %d: %s", line, col, fmt.Sprintf(format, args...))
}
