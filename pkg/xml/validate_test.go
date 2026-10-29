package xml

import (
	stdxml "encoding/xml"
	"io"
	"strings"
	"testing"
)

func stdValid(s string) bool {
	d := stdxml.NewDecoder(strings.NewReader(s))
	for {
		_, err := d.Token()
		if err == io.EOF {
			return true
		}
		if err != nil {
			return false
		}
	}
}

// validateCases: want is the Validate verdict. Every case is also checked
// against a token loop over encoding/xml, except where std is set to say the
// verdict differs on purpose.
var validateCases = []struct {
	name string
	in   string
	want bool
	std  string // non-empty: why Validate deliberately differs from encoding/xml
}{
	// Rejected before this change (the previous Validate accepted them).
	{"unknown entity", "<a>&bogus;</a>", false, ""},
	{"unknown entity nope", "<a>&nope;</a>", false, ""},
	{"entity without semicolon", "<a>&amp</a>", false, ""},
	{"bare ampersand", "<a>a & b</a>", false, ""},
	{"NUL in text", "<a>\x00</a>", false, ""},
	{"control char in text", "<a>\x01</a>", false, ""},
	{"U+FFFE in text", "<a>\xef\xbf\xbe</a>", false, ""},
	{"NUL in attribute", "<a x=\"\x00\"/>", false, ""},
	{"NUL in CDATA", "<a><![CDATA[\x00]]></a>", false, ""},
	{"char ref to NUL", "<a>&#0;</a>", false, ""},
	{"hex char ref to NUL", "<a>&#x0;</a>", false, ""},
	{"char ref to control", "<a>&#1;</a>", false, ""},
	{"char ref out of range", "<a>&#x110000;</a>", false, ""},
	{"char ref overflow", "<a>&#99999999999999999999;</a>", false, ""},
	{"char ref no digits", "<a>&#;</a>", false, ""},
	{"char ref no semicolon", "<a>&#65</a>", false, ""},
	{"char ref to surrogate", "<a>&#xD800;</a>", false, "encoding/xml maps it to U+FFFD and accepts"},
	{"invalid UTF-8 in text", "<a>\xff</a>", false, ""},
	{"invalid UTF-8 in attribute", "<a x=\"\xff\"/>", false, ""},
	{"invalid UTF-8 in CDATA", "<a><![CDATA[\xc3]]></a>", false, ""},
	{"less-than in attribute", `<a x="<"/>`, false, ""},
	{"latin1 encoding", `<?xml version="1.0" encoding="ISO-8859-1"?><a/>`, false, ""},
	{"UTF-16 encoding", `<?xml version="1.0" encoding="UTF-16"?><a/>`, false, ""},
	{"unsupported version", `<?xml version="1.1"?><a/>`, false, ""},
	{"]]> in text", "<a>]]></a>", false, ""},

	// Accepted before this change as well, still accepted or newly accepted.
	{"leading BOM", "\xef\xbb\xbf<a/>", true, ""},
	{"BOM before declaration", "\xef\xbb\xbf<?xml version=\"1.0\"?><a/>", true, ""},
	{"plain DOCTYPE", "<!DOCTYPE a><a/>", true, ""},
	{"DOCTYPE with SYSTEM id", `<!DOCTYPE a SYSTEM "a.dtd"><a/>`, true, ""},
	{"DOCTYPE with internal subset", `<!DOCTYPE a [<!ELEMENT a (#PCDATA)>]><a>x</a>`, true, ""},
	{"DOCTYPE after declaration", "<?xml version=\"1.0\"?>\n<!DOCTYPE html>\n<html/>", true, ""},
	{"DOCTYPE with comment", "<!DOCTYPE a [<!-- > --><!ENTITY x \"y\">]><a/>", true, ""},
	{"utf-8 encoding", `<?xml version="1.0" encoding="UTF-8"?><a/>`, true, ""},
	{"lowercase utf-8", `<?xml version="1.0" encoding="utf-8"?><a/>`, true, ""},
	{"predefined entities", "<a>&lt;&gt;&amp;&apos;&quot;</a>", true, ""},
	{"char refs", "<a>&#65;&#x41;&#x10FFFF;&#x9;</a>", true, ""},
	{"entities in attribute", `<a x="&lt;&amp;"/>`, true, ""},
	{"CDATA with markup", "<a><![CDATA[<b>&bogus;</b>]]></a>", true, ""},
	{"comment", "<!-- c --><a><!-- d --></a>", true, ""},
	{"PI", "<?pi data?><a/>", true, ""},
	{"unicode names", "<é日:ñ xé=\"1\"/>", true, ""},
	{"namespaces", `<a xmlns:n="u"><n:b/></a>`, true, ""},
	{"duplicate attributes", `<a x="1" x="2"/>`, true, ""},
	{"single-quoted attribute", `<a x='1'/>`, true, ""},
	{"whitespace around =", `<a x = "1" />`, true, ""},
	{"emoji in text", "<a>😀</a>", true, ""},
	{"]]> in double-quoted attribute", `<a b="x]]>y"/>`, true, ""},
	{"]]> in single-quoted attribute", `<a b='x]]>y'/>`, true, ""},
	{"two colons in element name", "<a:b:c></a:b:c>", false, ""},
	{"two colons in end tag", "<a:b:c></a:b:c >", false, ""},
	{"two colons in attribute name", `<a xml:a:g="x"/>`, false, ""},
	{"two colons in PI target", "<?a:b:c d?><a/>", true, ""},
	{"declaration param search continues", `<?xml version=1.version="1.0"?><a/>`, true, ""},
	{"declaration param search finds bad version", `<?xml version=1.version="2.0"?><a/>`, false, ""},

	// Structure: encoding/xml requires matching tags, but not a single root.
	{"empty input", "", true, ""},
	{"whitespace only", " \n\t", true, ""},
	{"two roots", "<a/><b/>", true, ""},
	{"text after root", "<a/>tail", true, ""},
	{"text only", "hello", true, ""},
	{"mismatched tags", "<a></b>", false, ""},
	{"unclosed element", "<a><b></a>", false, ""},
	{"unclosed at EOF", "<a>", false, ""},
	{"unexpected end tag", "</a>", false, ""},
	{"truncated tag", "<a", false, ""},
	{"unquoted attribute", "<a x=1/>", false, ""},
	{"attribute without value", "<a x/>", false, ""},
	{"missing element name", "< a/>", false, ""},
	{"unterminated comment", "<!-- x", false, ""},
	{"double dash in comment", "<!-- a -- b -->", false, ""},
	{"unterminated CDATA", "<a><![CDATA[x</a>", false, ""},
	{"unterminated PI", "<?pi x", false, ""},
	{"unterminated DOCTYPE", "<!DOCTYPE a [", false, ""},
	{"bad <! sequence", "<!x", false, ""},
}

func TestValidate_Table(t *testing.T) {
	for _, tc := range validateCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := Validate(tc.in); (err == nil) != tc.want {
				t.Fatalf("Validate(%q) = %v, want valid=%v", tc.in, err, tc.want)
			}
			if err := ValidateReader(strings.NewReader(tc.in)); (err == nil) != tc.want {
				t.Fatalf("ValidateReader(%q) = %v, want valid=%v", tc.in, err, tc.want)
			}
			std := stdValid(tc.in)
			switch {
			case tc.std == "" && std != tc.want:
				t.Errorf("encoding/xml says valid=%v for %q, table says %v", std, tc.in, tc.want)
			case tc.std != "" && std == tc.want:
				t.Errorf("case is marked as differing from encoding/xml (%s) but they agree", tc.std)
			}
		})
	}
}
