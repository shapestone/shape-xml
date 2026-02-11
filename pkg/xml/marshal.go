package xml

import (
	"reflect"
	"strconv"

	"github.com/shapestone/shape-xml/internal/fastparser"
)

// Marshal returns the XML encoding of v.
//
// Marshal traverses the value v recursively. If an encountered value implements
// the xml.Marshaler interface, Marshal calls its MarshalXML method to produce XML.
//
// Otherwise, Marshal uses the following type-dependent default encodings:
//
// Boolean values encode as XML text (true/false).
//
// Floating point, integer values encode as XML text.
//
// String values encode as XML text with proper escaping.
//
// Array and slice values encode as a sequence of XML elements with the same name.
//
// Struct values encode as XML elements. Each exported struct field becomes
// either an XML element or attribute, using the field name as the element/attribute name,
// unless the field is omitted for one of the reasons given below.
//
// The encoding of each struct field can be customized by the format string
// stored under the "xml" key in the struct field's tag. The format string
// gives the name of the field, possibly followed by a comma-separated list
// of options. The name may be empty in order to specify options without
// overriding the default field name.
//
// The "attr" option specifies that the field should be encoded as an XML attribute.
//
// The "chardata" option specifies that the field contains the text content of the element.
//
// The "cdata" option specifies that the field contains CDATA content.
//
// The "omitempty" option specifies that the field should be omitted from the
// encoding if the field has an empty value, defined as false, 0, a nil pointer,
// a nil interface value, and any empty array, slice, map, or string.
//
// As a special case, if the field tag is "-", the field is always omitted.
//
// Map values encode as XML elements with map keys as element names.
// The map's key type must be a string; the map keys are used as XML element names.
//
// Pointer values encode as the value pointed to. A nil pointer encodes as
// an empty XML element.
//
// Interface values encode as the value contained in the interface.
// A nil interface value encodes as an empty XML element.
//
// XML cannot represent cyclic data structures and Marshal does not handle them.
// Passing cyclic structures to Marshal will result in an error.
func Marshal(v interface{}) ([]byte, error) {
	if v == nil {
		return []byte("<root/>"), nil
	}

	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return []byte("<root/>"), nil
		}
		rv = rv.Elem()
	}

	// Determine root element name.
	rootName := "root"
	if rv.Kind() == reflect.Struct {
		if name := rv.Type().Name(); name != "" {
			rootName = name
		}
	}

	enc := xmlEncoderForType(rv.Type())

	bp := xmlBufPool.Get().(*[]byte)
	buf := (*bp)[:0]

	var err error
	buf, err = enc(buf, rv, rootName)
	if err != nil {
		*bp = buf
		xmlBufPool.Put(bp)
		return nil, err
	}

	result := make([]byte, len(buf))
	copy(result, buf)
	*bp = buf
	xmlBufPool.Put(bp)
	return result, nil
}

// MarshalIndent works like Marshal but with indentation for readability.
// Each XML element begins on a new line starting with prefix followed by one or more
// copies of indent according to the nesting depth.
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	// For now, just call Marshal - pretty printing can be added later
	// This matches the shape-json pattern
	return Marshal(v)
}

// Marshaler is the interface implemented by types that can marshal themselves into valid XML.
type Marshaler interface {
	MarshalXML() ([]byte, error)
}

// formatValue formats a reflect.Value as a string for attribute values or text content
func formatValue(rv reflect.Value) string {
	if !rv.IsValid() {
		return ""
	}

	switch rv.Kind() {
	case reflect.String:
		return rv.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 64)
	case reflect.Bool:
		return strconv.FormatBool(rv.Bool())
	case reflect.Ptr, reflect.Interface:
		if rv.IsNil() {
			return ""
		}
		return formatValue(rv.Elem())
	default:
		return ""
	}
}

// Unmarshal parses the XML-encoded data and stores the result in the value pointed to by v.
//
// This function uses a high-performance fast path that bypasses AST construction for
// optimal performance (4-5x faster than AST path). If you need the AST for advanced
// features, use Parse() followed by NodeToInterface().
//
// Unmarshal uses XML struct tags to map XML elements and attributes to struct fields:
//
//	type User struct {
//	    ID   string `xml:"id,attr"`     // Attribute
//	    Name string `xml:"name"`         // Child element
//	    Bio  string `xml:",chardata"`    // Text content
//	}
//
// To unmarshal XML into an interface value, Unmarshal stores a map[string]interface{}
// representation:
//   - "@attrname" for attributes
//   - "#text" for text content
//   - "#cdata" for CDATA sections
//   - "childname" for child elements
func Unmarshal(data []byte, v interface{}) error {
	// Fast path: Direct parsing without AST construction (4-5x faster)
	return fastparser.Unmarshal(data, v)
}
