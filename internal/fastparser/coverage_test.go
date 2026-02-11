package fastparser

import (
	"reflect"
	"strings"
	"testing"
)

// ---------- parseStringWithEscapes ----------

func TestParse_EscapedAttributes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string // expected attribute value
	}{
		{
			name:  "escaped double quote",
			input: `<root attr="val\"ue"></root>`,
			want:  `val"ue`,
		},
		{
			name:  "escaped backslash",
			input: `<root attr="path\\to"></root>`,
			want:  `path\to`,
		},
		{
			name:  "escaped newline",
			input: `<root attr="line\nbreak"></root>`,
			want:  "line\nbreak",
		},
		{
			name:  "escaped tab",
			input: `<root attr="tab\there"></root>`,
			want:  "tab\there",
		},
		{
			name:  "escaped carriage return",
			input: `<root attr="cr\rreturn"></root>`,
			want:  "cr\rreturn",
		},
		{
			name:  "unknown escape preserved",
			input: `<root attr="unknown\xchar"></root>`,
			want:  `unknown\xchar`,
		},
		{
			name:  "escaped single quote in single-quoted attr",
			input: `<root attr='val\'ue'></root>`,
			want:  `val'ue`,
		},
		{
			name:  "multiple escapes",
			input: `<root attr="a\\b\nc"></root>`,
			want:  "a\\b\nc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser([]byte(tt.input))
			result, err := p.Parse()
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			m, ok := result.(map[string]interface{})
			if !ok {
				t.Fatalf("expected map, got %T", result)
			}
			got, ok := m["@attr"]
			if !ok {
				t.Fatal("expected @attr key in result")
			}
			if got != tt.want {
				t.Errorf("attr value = %q, want %q", got, tt.want)
			}
		})
	}
}

// ---------- joinStrings multi-part path ----------

func TestParse_MultiPartTextContent(t *testing.T) {
	input := `<root>hello<![CDATA[ world]]></root>`
	p := NewParser([]byte(input))
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	// Text part should be "hello"
	if text, ok := m["#text"]; ok {
		if text != "hello" {
			t.Errorf("text = %q, want %q", text, "hello")
		}
	}
	// CDATA part should be " world"
	if cdata, ok := m["#cdata"]; ok {
		if cdata != " world" {
			t.Errorf("cdata = %q, want %q", cdata, " world")
		}
	} else {
		t.Error("expected #cdata key in result")
	}
}

// ---------- joinStrings unit tests ----------

func TestJoinStrings_Direct(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  string
	}{
		{"nil", nil, ""},
		{"empty", []string{}, ""},
		{"single", []string{"hello"}, "hello"},
		{"multiple", []string{"a", "b", "c"}, "abc"},
		{"with spaces", []string{"hello ", "world"}, "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := joinStrings(tt.input)
			if got != tt.want {
				t.Errorf("joinStrings() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ---------- unmarshalString to interface ----------

func TestUnmarshal_StringToInterface(t *testing.T) {
	input := `<root>text content</root>`
	var result interface{}
	err := Unmarshal([]byte(input), &result)
	if err != nil {
		t.Fatalf("Unmarshal error = %v", err)
	}
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	text, ok := m["#text"]
	if !ok {
		t.Fatal("expected #text key")
	}
	if text != "text content" {
		t.Errorf("text = %q, want %q", text, "text content")
	}
}

// ---------- unmarshalMap key type mismatch ----------

func TestUnmarshal_MapKeyTypeMismatch(t *testing.T) {
	m := map[string]interface{}{"key": "value"}
	target := make(map[int]string)
	rv := reflect.ValueOf(&target).Elem()
	err := unmarshalMap(m, rv)
	if err == nil {
		t.Fatal("expected error for map key type mismatch")
	}
	if !strings.Contains(err.Error(), "map key type mismatch") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------- unmarshalArray bounds ----------

func TestUnmarshal_ArrayBounds(t *testing.T) {
	arr := []interface{}{"a", "b", "c"}
	var target [2]string
	rv := reflect.ValueOf(&target).Elem()
	err := unmarshalArray(arr, rv)
	if err != nil {
		t.Fatalf("unmarshalArray error = %v", err)
	}
	if target[0] != "a" || target[1] != "b" {
		t.Errorf("got %v, want [a b]", target)
	}
}

func TestUnmarshal_ArrayNotSliceOrArray(t *testing.T) {
	arr := []interface{}{"a"}
	var target string
	rv := reflect.ValueOf(&target).Elem()
	err := unmarshalArray(arr, rv)
	if err == nil {
		t.Fatal("expected error for non-slice/array target")
	}
	if !strings.Contains(err.Error(), "cannot unmarshal array") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------- skipXMLDeclaration EOF branch ----------

func TestParse_UnterminatedDeclaration(t *testing.T) {
	input := `<?xml version`
	p := NewParser([]byte(input))
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected error for unterminated XML declaration")
	}
	if !strings.Contains(err.Error(), "unterminated XML declaration") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------- skipComment EOF branch ----------

func TestParse_UnterminatedComment(t *testing.T) {
	// Unterminated comment at top level: skipComments swallows the error,
	// but parsing still fails because no root element is found.
	input := `<!-- no end`
	p := NewParser([]byte(input))
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected error for unterminated comment")
	}

	// Unterminated comment inside an element: skipComment error is propagated.
	input2 := `<root><!-- no end`
	p2 := NewParser([]byte(input2))
	_, err2 := p2.Parse()
	if err2 == nil {
		t.Fatal("expected error for unterminated comment inside element")
	}
	if !strings.Contains(err2.Error(), "unterminated comment") {
		t.Errorf("unexpected error: %v", err2)
	}
}

// ---------- unmarshalValue with unexpected type ----------

func TestUnmarshalValue_UnexpectedType(t *testing.T) {
	var target string
	rv := reflect.ValueOf(&target).Elem()
	err := unmarshalValue(123, rv) // int, not string/map/slice
	if err == nil {
		t.Fatal("expected error for unexpected value type")
	}
	if !strings.Contains(err.Error(), "unexpected value type") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------- unmarshalValue map to unsupported type ----------

func TestUnmarshalValue_MapToUnsupportedType(t *testing.T) {
	m := map[string]interface{}{"key": "value"}
	var target int
	rv := reflect.ValueOf(&target).Elem()
	err := unmarshalValue(m, rv)
	if err == nil {
		t.Fatal("expected error for map to int")
	}
	if !strings.Contains(err.Error(), "cannot unmarshal object") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------- Unmarshal nil pointer ----------

func TestUnmarshal_NilPointer(t *testing.T) {
	err := Unmarshal([]byte(`<root/>`), (*map[string]interface{})(nil))
	if err == nil {
		t.Fatal("expected error for nil pointer")
	}
	if !strings.Contains(err.Error(), "nil") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------- Escaped backslash at end of string ----------

func TestParse_EscapeAtEndOfString(t *testing.T) {
	input := `<root attr="value\"></root>`
	p := NewParser([]byte(input))
	_, err := p.Parse()
	// The backslash escapes the quote, so the string runs to end of input
	if err == nil {
		t.Fatal("expected error for backslash escaping the closing quote")
	}
}

// ---------- Multiple text segments joined via joinStrings multi-part ----------

func TestParse_TextCDataText(t *testing.T) {
	// This exercises the joinStrings multi-part path for CDATA
	input := `<root><![CDATA[first]]><![CDATA[ second]]></root>`
	p := NewParser([]byte(input))
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	cdata, ok := m["#cdata"]
	if !ok {
		t.Fatal("expected #cdata key")
	}
	if cdata != "first second" {
		t.Errorf("cdata = %q, want %q", cdata, "first second")
	}
}

// ---------- unmarshalValue with pointer target ----------

func TestUnmarshalValue_PointerTarget(t *testing.T) {
	var target *string
	rv := reflect.ValueOf(&target).Elem()
	err := unmarshalValue("hello", rv)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if target == nil || *target != "hello" {
		t.Errorf("got %v, want pointer to 'hello'", target)
	}
}
