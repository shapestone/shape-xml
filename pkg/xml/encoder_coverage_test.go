package xml

import (
	"strings"
	"testing"
)

// ---------- Uint types ----------

func TestMarshalEncoder_UintTypes(t *testing.T) {
	type Uints struct {
		A uint   `xml:"a"`
		B uint8  `xml:"b"`
		C uint16 `xml:"c"`
		D uint32 `xml:"d"`
		E uint64 `xml:"e"`
	}
	v := Uints{A: 1, B: 2, C: 3, D: 4, E: 5}
	out, err := Marshal(v)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	s := string(out)
	for _, want := range []string{"<a>1</a>", "<b>2</b>", "<c>3</c>", "<d>4</d>", "<e>5</e>"} {
		if !strings.Contains(s, want) {
			t.Errorf("expected %s in output %s", want, s)
		}
	}
}

// ---------- Pointer fields ----------

func TestMarshalEncoder_PointerFields(t *testing.T) {
	type WithPtrs struct {
		Name *string `xml:"name"`
		Age  *int    `xml:"age"`
	}

	t.Run("non-nil pointers", func(t *testing.T) {
		name := "Alice"
		age := 30
		v := WithPtrs{Name: &name, Age: &age}
		out, err := Marshal(v)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}
		s := string(out)
		if !strings.Contains(s, "<name>Alice</name>") {
			t.Errorf("expected <name>Alice</name>, got %s", s)
		}
		if !strings.Contains(s, "<age>30</age>") {
			t.Errorf("expected <age>30</age>, got %s", s)
		}
	})

	t.Run("nil pointers", func(t *testing.T) {
		v := WithPtrs{}
		out, err := Marshal(v)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}
		s := string(out)
		if !strings.Contains(s, "<name/>") {
			t.Errorf("expected <name/> for nil *string, got %s", s)
		}
		if !strings.Contains(s, "<age/>") {
			t.Errorf("expected <age/> for nil *int, got %s", s)
		}
	})
}

// ---------- Interface fields ----------

func TestMarshalEncoder_InterfaceFields(t *testing.T) {
	type WithIface struct {
		A interface{} `xml:"a"`
		B interface{} `xml:"b"`
		C interface{} `xml:"c"`
	}

	t.Run("non-nil values", func(t *testing.T) {
		v := WithIface{A: "hello", B: 42, C: true}
		out, err := Marshal(v)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}
		s := string(out)
		if !strings.Contains(s, "<a>hello</a>") {
			t.Errorf("expected <a>hello</a>, got %s", s)
		}
		if !strings.Contains(s, "<b>42</b>") {
			t.Errorf("expected <b>42</b>, got %s", s)
		}
		if !strings.Contains(s, "<c>true</c>") {
			t.Errorf("expected <c>true</c>, got %s", s)
		}
	})

	t.Run("nil interface", func(t *testing.T) {
		v := WithIface{}
		out, err := Marshal(v)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}
		s := string(out)
		if !strings.Contains(s, "<a/>") {
			t.Errorf("expected <a/> for nil interface, got %s", s)
		}
	})
}

// ---------- Marshaler interface ----------

type testMarshaler struct {
	val string
}

func (m testMarshaler) MarshalXML() ([]byte, error) {
	return []byte("<custom>" + m.val + "</custom>"), nil
}

type testMarshalerErr struct{}

func (m testMarshalerErr) MarshalXML() ([]byte, error) {
	return nil, errTestMarshaler
}

var errTestMarshaler = &marshalerError{"test marshaler error"}

type marshalerError struct{ msg string }

func (e *marshalerError) Error() string { return e.msg }

func TestMarshalEncoder_Marshaler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		type Wrap struct {
			M testMarshaler `xml:"m"`
		}
		v := Wrap{M: testMarshaler{val: "data"}}
		out, err := Marshal(v)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}
		s := string(out)
		if !strings.Contains(s, "<custom>data</custom>") {
			t.Errorf("expected custom marshaler output, got %s", s)
		}
	})

	t.Run("error", func(t *testing.T) {
		type Wrap struct {
			M testMarshalerErr `xml:"m"`
		}
		v := Wrap{M: testMarshalerErr{}}
		_, err := Marshal(v)
		if err == nil {
			t.Fatal("expected error from MarshalXML")
		}
		if !strings.Contains(err.Error(), "test marshaler error") {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

// ---------- Addr Marshaler (pointer receiver) ----------

type addrMarshaler struct {
	val string
}

func (m *addrMarshaler) MarshalXML() ([]byte, error) {
	return []byte("<addr>" + m.val + "</addr>"), nil
}

func TestMarshalEncoder_AddrMarshaler(t *testing.T) {
	type Wrap struct {
		M addrMarshaler `xml:"m"`
	}
	// Must pass pointer so struct fields are addressable (CanAddr() == true).
	v := &Wrap{M: addrMarshaler{val: "ptr"}}
	out, err := Marshal(v)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "<addr>ptr</addr>") {
		t.Errorf("expected addr marshaler output, got %s", s)
	}
}

func TestMarshalEncoder_AddrMarshalerFallback(t *testing.T) {
	// When passed by value, CanAddr() is false, so it falls through to
	// buildXMLEncoderNoMarshaler (the fallback path).
	type Wrap struct {
		M addrMarshaler `xml:"m"`
	}
	v := Wrap{M: addrMarshaler{val: "ptr"}}
	out, err := Marshal(v)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	s := string(out)
	// Fallback path treats it as a struct with no exported fields → self-closing
	if !strings.Contains(s, "<m/>") {
		t.Errorf("expected <m/> fallback, got %s", s)
	}
}

// ---------- Fixed array ----------

func TestMarshalEncoder_FixedArray(t *testing.T) {
	type WithArray struct {
		Items [3]string `xml:"item"`
	}
	v := WithArray{Items: [3]string{"a", "b", "c"}}
	out, err := Marshal(v)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	s := string(out)
	if strings.Count(s, "<item>") != 3 {
		t.Errorf("expected 3 <item> elements, got %s", s)
	}
	for _, want := range []string{"<item>a</item>", "<item>b</item>", "<item>c</item>"} {
		if !strings.Contains(s, want) {
			t.Errorf("expected %s in output %s", want, s)
		}
	}
}

// ---------- Unsupported type ----------

func TestMarshalEncoder_UnsupportedType(t *testing.T) {
	type WithChan struct {
		Ch chan int `xml:"ch"`
	}
	v := WithChan{Ch: make(chan int)}
	_, err := Marshal(v)
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
	if !strings.Contains(err.Error(), "unsupported type") {
		t.Errorf("expected 'unsupported type' error, got: %v", err)
	}
}

// ---------- Escape characters ----------

func TestMarshalEncoder_EscapeChars(t *testing.T) {
	type Text struct {
		Value string `xml:"value"`
	}
	tests := []struct {
		input string
		want  string
	}{
		{"a&b", "&amp;"},
		{"a<b", "&lt;"},
		{"a>b", "&gt;"},
		{`a"b`, "&#34;"},
		{"a'b", "&#39;"},
		{"<>&\"'", "&lt;&gt;&amp;&#34;&#39;"},
		{"noescape", "<value>noescape</value>"},
	}

	for _, tt := range tests {
		v := Text{Value: tt.input}
		out, err := Marshal(v)
		if err != nil {
			t.Fatalf("Marshal(%q) failed: %v", tt.input, err)
		}
		s := string(out)
		if !strings.Contains(s, tt.want) {
			t.Errorf("Marshal(%q): expected %q in output %s", tt.input, tt.want, s)
		}
	}
}

// ---------- OmitEmpty all types ----------

func TestMarshalEncoder_OmitEmptyAllTypes(t *testing.T) {
	type Full struct {
		S   string            `xml:"s,omitempty"`
		I   int               `xml:"i,omitempty"`
		U   uint              `xml:"u,omitempty"`
		F   float64           `xml:"f,omitempty"`
		B   bool              `xml:"b,omitempty"`
		P   *string           `xml:"p,omitempty"`
		Ifc interface{}       `xml:"ifc,omitempty"`
		Sl  []string          `xml:"sl,omitempty"`
		M   map[string]string `xml:"m,omitempty"`
	}

	t.Run("all zero", func(t *testing.T) {
		v := Full{}
		out, err := Marshal(v)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}
		s := string(out)
		// Should be self-closing with no children
		if !strings.HasSuffix(s, "/>") {
			t.Errorf("expected self-closing tag for all-empty struct, got %s", s)
		}
		for _, field := range []string{"<s>", "<i>", "<u>", "<f>", "<b>", "<p>", "<ifc>", "<sl>", "<m>"} {
			if strings.Contains(s, field) {
				t.Errorf("expected %s to be omitted, got %s", field, s)
			}
		}
	})

	t.Run("all populated", func(t *testing.T) {
		str := "x"
		v := Full{
			S:   "hello",
			I:   1,
			U:   2,
			F:   3.14,
			B:   true,
			P:   &str,
			Ifc: "val",
			Sl:  []string{"a"},
			M:   map[string]string{"k": "v"},
		}
		out, err := Marshal(v)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}
		s := string(out)
		for _, field := range []string{"<s>", "<i>", "<u>", "<f>", "<b>", "<p>", "<ifc>", "<sl>", "<m>"} {
			if !strings.Contains(s, field) {
				t.Errorf("expected %s to be present, got %s", field, s)
			}
		}
	})
}

// ---------- Nil slice and map ----------

func TestMarshalEncoder_NilSliceAndMap(t *testing.T) {
	type WithNils struct {
		Items []string          `xml:"items"`
		Props map[string]string `xml:"props"`
	}
	v := WithNils{} // nil slice and nil map
	out, err := Marshal(v)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "<items/>") {
		t.Errorf("expected <items/> for nil slice, got %s", s)
	}
	if !strings.Contains(s, "<props/>") {
		t.Errorf("expected <props/> for nil map, got %s", s)
	}
}

// ---------- Nil input ----------

func TestMarshalEncoder_NilInput(t *testing.T) {
	t.Run("nil interface", func(t *testing.T) {
		out, err := Marshal(nil)
		if err != nil {
			t.Fatalf("Marshal(nil) failed: %v", err)
		}
		if string(out) != "<root/>" {
			t.Errorf("expected <root/>, got %s", string(out))
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		type S struct{ Name string }
		var p *S
		out, err := Marshal(p)
		if err != nil {
			t.Fatalf("Marshal(nil ptr) failed: %v", err)
		}
		if string(out) != "<root/>" {
			t.Errorf("expected <root/>, got %s", string(out))
		}
	})
}

// ---------- Uint attribute ----------

func TestMarshalEncoder_UintAttr(t *testing.T) {
	type WithAttr struct {
		Count uint `xml:"count,attr"`
		Name  string
	}
	v := WithAttr{Count: 42, Name: "test"}
	out, err := Marshal(v)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, `count="42"`) {
		t.Errorf("expected count=\"42\" attribute, got %s", s)
	}
}

// ---------- CDATA field ----------

func TestMarshalEncoder_CData(t *testing.T) {
	type WithCData struct {
		Data string `xml:",cdata"`
	}
	v := WithCData{Data: "some <raw> data"}
	out, err := Marshal(v)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "<![CDATA[some <raw> data]]>") {
		t.Errorf("expected CDATA section, got %s", s)
	}
}

// ---------- Map with non-string key ----------

func TestMarshalEncoder_MapNonStringKey(t *testing.T) {
	m := map[int]string{1: "a"}
	_, err := Marshal(m)
	if err == nil {
		t.Fatal("expected error for non-string map key")
	}
	if !strings.Contains(err.Error(), "unsupported map key type") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------- Nested struct with all features ----------

func TestMarshalEncoder_ComplexStruct(t *testing.T) {
	type Address struct {
		City string `xml:"city"`
		Zip  string `xml:"zip"`
	}
	type User struct {
		ID      int     `xml:"id,attr"`
		Name    string  `xml:"name"`
		Active  bool    `xml:"active"`
		Score   float64 `xml:"score"`
		Address Address `xml:"address"`
	}
	v := User{ID: 1, Name: "Alice", Active: true, Score: 9.5, Address: Address{City: "NYC", Zip: "10001"}}
	out, err := Marshal(v)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	s := string(out)
	for _, want := range []string{`id="1"`, "<name>Alice</name>", "<active>true</active>", "<score>9.5</score>", "<city>NYC</city>", "<zip>10001</zip>"} {
		if !strings.Contains(s, want) {
			t.Errorf("expected %s in output %s", want, s)
		}
	}
}

// ---------- Empty slice (non-nil) ----------

func TestMarshalEncoder_EmptySlice(t *testing.T) {
	type WithSlice struct {
		Items []string `xml:"items"`
	}
	v := WithSlice{Items: []string{}}
	out, err := Marshal(v)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	s := string(out)
	// Empty non-nil slice should NOT produce any <items> elements (length 0)
	if strings.Contains(s, "<items>") {
		t.Errorf("expected no <items> elements for empty slice, got %s", s)
	}
}

// ---------- Map with interface values ----------

func TestMarshalEncoder_MapInterfaceValues(t *testing.T) {
	m := map[string]interface{}{
		"name": "Alice",
		"age":  30,
	}
	out, err := Marshal(m)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "<name>Alice</name>") {
		t.Errorf("expected <name>Alice</name>, got %s", s)
	}
	if !strings.Contains(s, "<age>30</age>") {
		t.Errorf("expected <age>30</age>, got %s", s)
	}
}

// ---------- Pointer attribute (appendFormatValue Ptr/Interface branches) ----------

func TestMarshalEncoder_PtrAttribute(t *testing.T) {
	type WithPtrAttr struct {
		Name *string `xml:"name,attr"`
		Body string
	}
	name := "Alice"
	v := WithPtrAttr{Name: &name, Body: "content"}
	out, err := Marshal(v)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, `name="Alice"`) {
		t.Errorf("expected name=\"Alice\" attribute, got %s", s)
	}
}

func TestMarshalEncoder_NilPtrAttribute(t *testing.T) {
	type WithPtrAttr struct {
		Name *string `xml:"name,attr"`
		Body string
	}
	v := WithPtrAttr{Body: "content"}
	out, err := Marshal(v)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	s := string(out)
	// Nil pointer attr produces empty string, so attr is skipped
	if strings.Contains(s, "name=") {
		t.Errorf("expected nil ptr attr to be omitted, got %s", s)
	}
}

// ---------- Interface attribute ----------

func TestMarshalEncoder_InterfaceAttribute(t *testing.T) {
	type WithIfaceAttr struct {
		Val interface{} `xml:"val,attr"`
		X   string
	}
	v := WithIfaceAttr{Val: "hello", X: "body"}
	out, err := Marshal(v)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, `val="hello"`) {
		t.Errorf("expected val=\"hello\" attribute, got %s", s)
	}
}

// ---------- Float attribute ----------

func TestMarshalEncoder_FloatTypes(t *testing.T) {
	type Floats struct {
		A float32 `xml:"a"`
		B float64 `xml:"b"`
	}
	v := Floats{A: 1.5, B: 2.5}
	out, err := Marshal(v)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "<a>1.5</a>") {
		t.Errorf("expected <a>1.5</a>, got %s", s)
	}
	if !strings.Contains(s, "<b>2.5</b>") {
		t.Errorf("expected <b>2.5</b>, got %s", s)
	}
}

// ---------- Bool attribute ----------

func TestMarshalEncoder_BoolAttr(t *testing.T) {
	type WithBoolAttr struct {
		Active bool `xml:"active,attr"`
		Body   string
	}
	v := WithBoolAttr{Active: true, Body: "x"}
	out, err := Marshal(v)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, `active="true"`) {
		t.Errorf("expected active=\"true\" attribute, got %s", s)
	}
}

// ---------- Int attribute ----------

func TestMarshalEncoder_IntAttr(t *testing.T) {
	type WithIntAttr struct {
		ID   int `xml:"id,attr"`
		Body string
	}
	v := WithIntAttr{ID: 99, Body: "x"}
	out, err := Marshal(v)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, `id="99"`) {
		t.Errorf("expected id=\"99\" attribute, got %s", s)
	}
}

// ---------- Float attribute ----------

func TestMarshalEncoder_FloatAttr(t *testing.T) {
	type WithFloatAttr struct {
		Score float64 `xml:"score,attr"`
		Body  string
	}
	v := WithFloatAttr{Score: 3.14, Body: "x"}
	out, err := Marshal(v)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, `score="3.14"`) {
		t.Errorf("expected score=\"3.14\" attribute, got %s", s)
	}
}

// ---------- Escape in attribute value ----------

func TestMarshalEncoder_EscapeInAttr(t *testing.T) {
	type WithAttr struct {
		Name string `xml:"name,attr"`
		Body string
	}
	v := WithAttr{Name: `a<b&c"d`, Body: "x"}
	out, err := Marshal(v)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, `name="a&lt;b&amp;c&#34;d"`) {
		t.Errorf("expected escaped attribute value, got %s", s)
	}
}

// ---------- Chardata with empty value (self-closing) ----------

func TestMarshalEncoder_EmptyChardata(t *testing.T) {
	type WithChardata struct {
		Text string `xml:",chardata"`
	}
	v := WithChardata{}
	out, err := Marshal(v)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	s := string(out)
	if !strings.HasSuffix(s, "/>") {
		t.Errorf("expected self-closing for empty chardata, got %s", s)
	}
}

// ---------- Empty CDATA (self-closing) ----------

func TestMarshalEncoder_EmptyCData(t *testing.T) {
	type WithCData struct {
		Data string `xml:",cdata"`
	}
	v := WithCData{}
	out, err := Marshal(v)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	s := string(out)
	if !strings.HasSuffix(s, "/>") {
		t.Errorf("expected self-closing for empty cdata, got %s", s)
	}
}

// ---------- Nested pointer to struct (double deref) ----------

func TestMarshalEncoder_DoublePointer(t *testing.T) {
	type Inner struct {
		Name string `xml:"name"`
	}
	type Outer struct {
		Inner **Inner `xml:"inner"`
	}
	inner := &Inner{Name: "Bob"}
	v := Outer{Inner: &inner}
	out, err := Marshal(v)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "<name>Bob</name>") {
		t.Errorf("expected <name>Bob</name>, got %s", s)
	}
}
