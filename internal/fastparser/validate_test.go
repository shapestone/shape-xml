package fastparser

import (
	"bytes"
	stdxml "encoding/xml"
	"go/parser"
	"go/token"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
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

// diffSeeds are documents and fragments that probe the places where a
// hand-written validator tends to diverge from encoding/xml: names with colons,
// namespace declarations, ]]> in different contexts, the XML declaration, <!
// constructs, entities, and truncated input.
var diffSeeds = []string{
	`<p:x xmlns:p="u" xmlns:q="u"></q:x>`, `<a xmlns:p="u"><p:x></p:x></a>`,
	`<a:b:c></a:b:c>`, `<a:b:c/>`, `<a xml:a:g="x"/>`, `<a b:c:d="1"/>`, `</a:b:c>`, `<:a></:a>`, `<a:></a:>`,
	`<xmlns:a></xmlns:a>`, `<a xmlns:xml="x"/>`, `<a xml:lang="x"/>`, `<?a:b:c d?><a/>`,
	`<a b="1"c="2"/>`, `<a b=1/>`, `<a b/>`, `<a/ >`, `<a b = '1' />`, `<a b="1" b="2"/>`,
	`<a b="]]>"/>`, `<a b="x]]>y"/>`, `<a b='x]]>y'/>`, `<a>]]></a>`, `<a>&amp;]]></a>`, `<a><![CDATA[ ]]]]> ]]></a>`,
	`<a b="<"/>`, `<a b="a&b"/>`, `<a b="&lt;&amp;&#65;"/>`, `<a b="&bogus;"/>`, `<a b="`, `<a b='x`,
	`<!DOCTYPE a [<!ENTITY e "<!--">]><a/>`, `<!DOCTYPE a [<!-- > -->]><a/>`, `<!--->`, `<!-- a -- b -->`, `<!---->`,
	`<?xml version="1.0" encoding="utf-8"?><a/>`, `<?xml encoding='UTF-8' ?><a/>`, `<?xml version="1.1"?>`,
	`<?xml version=1.0?>`, `<?xml version=1.version="1.0"?>`, `<?xml version=1.version="2.0"?>`,
	`<?xml encoding=x encoding="latin1"?>`, `<?xmlx version="9"?>`, `<?XML version="9"?>`, `<?xml?>`, `<??>`, `<? xml ?>`,
	"\xef\xbb\xbf<a/>", "<a>\x00</a>", "<a>\xff</a>",
	`<a>&#x;</a>`, `<a>&#;</a>`, `<a>&#X41;</a>`, `<a>&#0000000000000000065;</a>`, `<a>&#xFFFFFFFFFFFFFFFFFFFF;</a>`, `<a>&;</a>`,
	`<a>&amp</a>`, `<a>&#65</a>`, `<a>&lt;&gt;&amp;&apos;&quot;</a>`, `<a>&AMP;</a>`,
	"<a\x0b/>", "<a />", "<é/>", "<a·/>", "<·a/>",
	`</a>`, `<a></a></a>`, `<a><b></a></b>`, `<`, `<a`, `<a b`, `<a b=`, `<!`, `<!-`, `<!--`, `<![`, `<![CDATA[`, `<?`, `<?a`,
	`<!x>`, `<!>`, `<!"<>">`, `<!a<b>>`, `<!a<!--b>-->c>`,
}

var diffFrags = []string{
	"<", ">", "/", "?", "!", "-", "--", "[", "]", "]]>", "CDATA", "a", "b", "p:", ":", "a:b", "a:b:c", "x", "xmlns",
	"xmlns:p", "=", `"`, "'", ` b="1"`, ` b='x]]>y'`, ` b="<"`, ` xml:a:g="x"`, " ", "\n", "&", ";", "#", "amp", "lt",
	"<!--", "-->", "<![CDATA[", "<?xml", "?>", "<!DOCTYPE", "\x00", "\xff", "é", "u", "\xef\xbb\xbf",
	`version="1.0"`, `version=1.`, `encoding="UTF-8"`, `encoding="latin1"`, "/>", "</", "<a", "</a>", "<a>", "<p:a",
	"</p:a>", "<a:b:c>", "</a:b:c>", "&#65;", "&#0;", "&bogus;", "\r", "\t",
}

// checkAgainstStd compares Validate with encoding/xml on b. The one tolerated
// difference is a reference to a surrogate, which Validate rejects on purpose.
func checkAgainstStd(t testing.TB, b []byte) bool {
	t.Helper()
	got, want := Validate(b), stdErr(b)
	if (got == nil) == (want == nil) {
		return true
	}
	if want == nil && isSurrogateRefError(got) {
		return true
	}
	t.Errorf("Validate(%q) = %v; encoding/xml = %v", b, got, want)
	return false
}

func stdErr(b []byte) error {
	d := stdxml.NewDecoder(bytes.NewReader(b))
	for {
		if _, err := d.Token(); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

var surrogateRef = regexp.MustCompile(`invalid character reference &#(x[0-9a-fA-F]+|[0-9]+);`)

func isSurrogateRefError(err error) bool {
	m := surrogateRef.FindStringSubmatch(err.Error())
	if m == nil {
		return false
	}
	n, e := strconv.ParseUint(strings.TrimPrefix(m[1], "x"), map[bool]int{true: 16, false: 10}[strings.HasPrefix(m[1], "x")], 64)
	return e == nil && n >= 0xD800 && n <= 0xDFFF
}

// TestValidate_DifferentialRandom compares Validate with encoding/xml on the
// seed documents, on pseudo-random concatenations of fragments, and on
// mutations of the seeds. The random seed is fixed so a failure reproduces.
func TestValidate_DifferentialRandom(t *testing.T) {
	for _, s := range diffSeeds {
		checkAgainstStd(t, []byte(s))
	}
	r := rand.New(rand.NewSource(7))
	failures := 0
	for i := 0; i < 400000 && failures < 20; i++ {
		var b []byte
		if r.Intn(2) == 0 {
			b = []byte(diffSeeds[r.Intn(len(diffSeeds))])
			for k := r.Intn(3); k >= 0 && len(b) > 0; k-- {
				p := r.Intn(len(b))
				switch r.Intn(3) {
				case 0:
					b = append(b[:p], b[p+1:]...)
				case 1:
					b = append(b[:p], append([]byte(diffFrags[r.Intn(len(diffFrags))]), b[p:]...)...)
				case 2:
					b[p] = diffFrags[r.Intn(len(diffFrags))][0]
				}
			}
		} else {
			for k := 1 + r.Intn(14); k > 0; k-- {
				b = append(b, diffFrags[r.Intn(len(diffFrags))]...)
			}
		}
		if !checkAgainstStd(t, b) {
			failures++
		}
	}
}

// FuzzValidate checks Validate against encoding/xml. Run with:
//
//	go test ./internal/fastparser -fuzz=FuzzValidate -fuzztime=60s
func FuzzValidate(f *testing.F) {
	for _, s := range diffSeeds {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		checkAgainstStd(t, data)
	})
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

// Nothing in this package may be able to read files or use the network, so
// every non-test source file may import only these pure packages. A new file
// that imports os, net, io/ioutil, encoding/xml or similar fails this test.
func TestPackageImportsArePure(t *testing.T) {
	allowed := map[string]bool{
		"bytes": true, "errors": true, "fmt": true, "reflect": true, "sort": true,
		"strings": true, "unicode/utf8": true,
	}
	files, err := filepath.Glob("*.go")
	if err != nil || len(files) == 0 {
		t.Fatalf("no source files found: %v", err)
	}
	checked := 0
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		if _, err := os.Stat(file); err != nil {
			t.Fatal(err)
		}
		f, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		checked++
		for _, imp := range f.Imports {
			path, _ := strconv.Unquote(imp.Path.Value)
			if !allowed[path] {
				t.Errorf("%s imports %q; this package must stay free of I/O packages", file, path)
			}
		}
	}
	if checked < 4 {
		t.Errorf("only %d source files checked, expected validate.go, names.go, parser.go and unmarshal.go", checked)
	}
}
