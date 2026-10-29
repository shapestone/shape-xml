package fastparser

import (
	stdxml "encoding/xml"
	"go/parser"
	"go/token"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"testing"
)

// stdValid is the reference oracle: a token loop over encoding/xml's Decoder
// (Strict, no CharsetReader) until io.EOF.
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

// TestValidate_DifferentialRandom compares Validate with encoding/xml on
// pseudo-random concatenations of XML fragments, including malformed ones.
// The seed is fixed so a failure reproduces.
func TestValidate_DifferentialRandom(t *testing.T) {
	frags := []string{
		"<", ">", "/", "a", "b", "ab", "a:b", " ", "\n", "\t", "\r", "=", `"`, "'", "&", ";", "#", "x", "1", "0",
		"lt", "amp", "bogus", "!", "-", "--", "?", "[", "]", "]]>", "CDATA", "<![CDATA[", "<!--", "-->", "<?", "?>",
		"xml", "version", `"1.0"`, `encoding="utf-8"`, `<?xml version="1.0"?>`, `<?xml version="1.1"?>`,
		`<?xml encoding="latin1"?>`, "<?x y?>", "\x00", "\x01", "\x7f", "\xff", "\xc3", "\xed\xa0\x80",
		"é", "日", "😀", "\xef\xbb\xbf", "\xef\xbf\xbe", "DOCTYPE", "ENTITY", "SYSTEM", "<!DOCTYPE", "<!DOCTYPE a [",
		"]>", `<!ENTITY x "y">`, `"file:///etc/passwd"`, "<!-- c -->", "<![CDATA[x]]>",
		"<a>", "</a>", "<a/>", "<b>", "</b>", `<a x="1">`, "<a x=", "&amp;", "&#65;", "&#x41;", "&#0;", "&#65",
		"&#xZ;", "&#99999999999999999999;", "&#x10FFFF;", "&#x110000;", "&#xFFFE;", "&#x1;",
	}
	r := rand.New(rand.NewSource(7))
	for i := 0; i < 300000; i++ {
		var sb strings.Builder
		for j, n := 0, 1+r.Intn(16); j < n; j++ {
			sb.WriteString(frags[r.Intn(len(frags))])
		}
		s := sb.String()
		got, want := Validate([]byte(s)) == nil, stdValid(s)
		if got != want {
			t.Fatalf("Validate(%q) valid=%v, encoding/xml valid=%v (err: %v)", s, got, want, Validate([]byte(s)))
		}
	}
}

// encoding/xml turns a reference to a surrogate (&#xD800;) into U+FFFD and
// accepts it. XML 1.0 forbids surrogates, and the brief is to reject invalid
// character references, so Validate deliberately disagrees here.
func TestValidate_SurrogateReferencesRejected(t *testing.T) {
	for _, in := range []string{"<a>&#xD800;</a>", "<a>&#xDFFF;</a>", `<a x="&#55296;"/>`} {
		if Validate([]byte(in)) == nil {
			t.Errorf("Validate(%q) accepted a surrogate character reference", in)
		}
		if !stdValid(in) {
			t.Logf("encoding/xml no longer accepts %q; the divergence is gone", in)
		}
	}
}

// TestNameTablesMatchEncodingXML regenerates the name tables by probing
// encoding/xml with every rune and compares them with names.go.
func TestNameTablesMatchEncodingXML(t *testing.T) {
	if testing.Short() {
		t.Skip("sweeps every rune")
	}
	for r := rune(0x80); r <= 0x10FFFF; r++ {
		if r >= 0xD800 && r <= 0xDFFF {
			continue
		}
		s := string(r)
		if got, want := isNameStart(r), stdValid("<"+s+"/>"); got != want {
			t.Fatalf("isNameStart(%U) = %v, encoding/xml accepts %v", r, got, want)
		}
		if got, want := isNameRest(r), stdValid("<a"+s+"/>"); got != want {
			t.Fatalf("isNameRest(%U) = %v, encoding/xml accepts %v", r, got, want)
		}
	}
}

func TestValidate_DeepNestingDoesNotOverflowStack(t *testing.T) {
	const depth = 1000000
	doc := strings.Repeat("<a>", depth) + strings.Repeat("</a>", depth)
	if err := Validate([]byte(doc)); err != nil {
		t.Fatalf("balanced %d-deep document: %v", depth, err)
	}
	if err := Validate([]byte(doc[:len(doc)-4])); err == nil {
		t.Fatal("unclosed deep document accepted")
	}
}

func TestValidate_ErrorHasLineAndColumn(t *testing.T) {
	err := Validate([]byte("<a>\n  <b>&bogus;</b>\n</a>"))
	if err == nil || !strings.Contains(err.Error(), "line 2, column 6") {
		t.Fatalf("want error at line 2, column 6, got %v", err)
	}
}

// The validator must never be able to read files or use the network: it may
// import nothing but a few pure standard-library packages.
func TestValidator_ImportsArePure(t *testing.T) {
	allowed := map[string]bool{"bytes": true, "fmt": true, "sort": true, "strings": true, "unicode/utf8": true}
	for _, file := range []string{"validate.go", "names.go"} {
		f, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range f.Imports {
			path, _ := strconv.Unquote(imp.Path.Value)
			if !allowed[path] {
				t.Errorf("%s imports %q; the validator must stay free of I/O packages", file, path)
			}
		}
	}
}
