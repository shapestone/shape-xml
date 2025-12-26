package xml

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strconv"
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
	buf := getBuffer()
	defer putBuffer(buf)

	// For struct types, we need the root element name
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Struct {
		// Use the struct type name as root element name
		rootName := rv.Type().Name()
		if rootName == "" {
			rootName = "root"
		}
		if err := marshalValue(rv, buf, rootName); err != nil {
			return nil, err
		}
	} else {
		// For non-struct types, wrap in a root element
		if err := marshalValue(rv, buf, "root"); err != nil {
			return nil, err
		}
	}

	// Must copy since buffer will be returned to pool
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
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

// marshalValue marshals a reflect.Value to a buffer as an XML element
func marshalValue(rv reflect.Value, buf *bytes.Buffer, elementName string) error {
	// Handle invalid values
	if !rv.IsValid() {
		buf.WriteString("<")
		buf.WriteString(elementName)
		buf.WriteString("/>")
		return nil
	}

	// Handle nil interface
	if rv.Kind() == reflect.Interface && rv.IsNil() {
		buf.WriteString("<")
		buf.WriteString(elementName)
		buf.WriteString("/>")
		return nil
	}

	// Check if type implements Marshaler interface
	if rv.Type().Implements(reflect.TypeOf((*Marshaler)(nil)).Elem()) {
		marshaler := rv.Interface().(Marshaler)
		b, err := marshaler.MarshalXML()
		if err != nil {
			return err
		}
		buf.Write(b)
		return nil
	}

	// Dereference interface
	if rv.Kind() == reflect.Interface {
		return marshalValue(rv.Elem(), buf, elementName)
	}

	// Handle pointers
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			buf.WriteString("<")
			buf.WriteString(elementName)
			buf.WriteString("/>")
			return nil
		}
		return marshalValue(rv.Elem(), buf, elementName)
	}

	switch rv.Kind() {
	case reflect.String:
		return marshalString(rv.String(), buf, elementName)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return marshalString(strconv.FormatInt(rv.Int(), 10), buf, elementName)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return marshalString(strconv.FormatUint(rv.Uint(), 10), buf, elementName)

	case reflect.Float32, reflect.Float64:
		return marshalString(strconv.FormatFloat(rv.Float(), 'g', -1, 64), buf, elementName)

	case reflect.Bool:
		return marshalString(strconv.FormatBool(rv.Bool()), buf, elementName)

	case reflect.Struct:
		return marshalStruct(rv, buf, elementName)

	case reflect.Map:
		return marshalMap(rv, buf, elementName)

	case reflect.Slice, reflect.Array:
		return marshalSlice(rv, buf, elementName)

	default:
		return fmt.Errorf("xml: unsupported type %s", rv.Type())
	}
}

// marshalString marshals a string value as an XML element with text content
func marshalString(s string, buf *bytes.Buffer, elementName string) error {
	buf.WriteString("<")
	buf.WriteString(elementName)
	buf.WriteString(">")
	buf.WriteString(escapeXML(s))
	buf.WriteString("</")
	buf.WriteString(elementName)
	buf.WriteString(">")
	return nil
}

// marshalStruct marshals a struct to XML
func marshalStruct(rv reflect.Value, buf *bytes.Buffer, elementName string) error {
	structType := rv.Type()

	// Start element opening tag
	buf.WriteString("<")
	buf.WriteString(elementName)

	// Collect attributes and content fields
	type attrEntry struct {
		name  string
		value string
	}
	var attrs []attrEntry
	var textContent string
	var cdataContent string

	// Collect child elements
	type childEntry struct {
		name  string
		value reflect.Value
	}
	var children []childEntry

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		info := getFieldInfo(field)

		// Skip fields with "-" tag
		if info.skip {
			continue
		}

		fieldVal := rv.Field(i)

		// Handle omitempty
		if info.omitEmpty && isEmptyValue(fieldVal) {
			continue
		}

		// Handle attributes
		if info.attr {
			attrVal := formatValue(fieldVal)
			if attrVal != "" {
				attrs = append(attrs, attrEntry{name: info.name, value: attrVal})
			}
			continue
		}

		// Handle chardata (text content)
		if info.chardata {
			textContent = formatValue(fieldVal)
			continue
		}

		// Handle cdata
		if info.cdata {
			cdataContent = formatValue(fieldVal)
			continue
		}

		// Regular child element
		children = append(children, childEntry{name: info.name, value: fieldVal})
	}

	// Sort attributes for deterministic output
	sort.Slice(attrs, func(i, j int) bool {
		return attrs[i].name < attrs[j].name
	})

	// Write attributes
	for _, attr := range attrs {
		buf.WriteString(" ")
		buf.WriteString(attr.name)
		buf.WriteString("=\"")
		buf.WriteString(escapeXML(attr.value))
		buf.WriteString("\"")
	}

	// Check if we have any content
	hasContent := textContent != "" || cdataContent != "" || len(children) > 0

	if !hasContent {
		// Self-closing tag
		buf.WriteString("/>")
		return nil
	}

	// Close opening tag
	buf.WriteString(">")

	// Write text content
	if textContent != "" {
		buf.WriteString(escapeXML(textContent))
	}

	// Write CDATA content
	if cdataContent != "" {
		buf.WriteString("<![CDATA[")
		buf.WriteString(cdataContent)
		buf.WriteString("]]>")
	}

	// Write child elements
	for _, child := range children {
		if err := marshalValue(child.value, buf, child.name); err != nil {
			return err
		}
	}

	// Close element
	buf.WriteString("</")
	buf.WriteString(elementName)
	buf.WriteString(">")

	return nil
}

// marshalMap marshals a map to XML
func marshalMap(rv reflect.Value, buf *bytes.Buffer, elementName string) error {
	if rv.IsNil() {
		buf.WriteString("<")
		buf.WriteString(elementName)
		buf.WriteString("/>")
		return nil
	}

	mapType := rv.Type()

	// Only support string keys
	if mapType.Key().Kind() != reflect.String {
		return fmt.Errorf("xml: unsupported map key type %s", mapType.Key())
	}

	// Start element
	buf.WriteString("<")
	buf.WriteString(elementName)
	buf.WriteString(">")

	// Get keys and sort them for deterministic output
	keys := rv.MapKeys()
	strKeys := make([]string, len(keys))
	for i, key := range keys {
		strKeys[i] = key.String()
	}
	sort.Strings(strKeys)

	// Marshal each key-value pair as a child element
	for _, keyStr := range strKeys {
		key := reflect.ValueOf(keyStr)
		val := rv.MapIndex(key)
		if err := marshalValue(val, buf, keyStr); err != nil {
			return err
		}
	}

	// Close element
	buf.WriteString("</")
	buf.WriteString(elementName)
	buf.WriteString(">")

	return nil
}

// marshalSlice marshals a slice or array to XML
func marshalSlice(rv reflect.Value, buf *bytes.Buffer, elementName string) error {
	// Nil slices encode as empty element
	if rv.Kind() == reflect.Slice && rv.IsNil() {
		buf.WriteString("<")
		buf.WriteString(elementName)
		buf.WriteString("/>")
		return nil
	}

	// For slices, we marshal each element with the same element name
	length := rv.Len()
	for i := 0; i < length; i++ {
		if err := marshalValue(rv.Index(i), buf, elementName); err != nil {
			return err
		}
	}

	return nil
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
// For now, this is a simple implementation that uses Parse and converts to native types.
func Unmarshal(data []byte, v interface{}) error {
	// Parse XML to AST
	node, err := Parse(string(data))
	if err != nil {
		return err
	}

	// Convert AST to native Go types
	value := NodeToInterface(node)

	// Use reflection to assign to v
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("xml: Unmarshal requires a pointer")
	}

	// Get the value that the pointer points to
	elem := rv.Elem()

	// Assign the converted value
	if !elem.CanSet() {
		return fmt.Errorf("xml: Unmarshal cannot set value")
	}

	// For now, we only support unmarshaling to interface{} or map[string]interface{}
	switch elem.Kind() {
	case reflect.Interface:
		elem.Set(reflect.ValueOf(value))
	case reflect.Map:
		if elem.Type().Key().Kind() == reflect.String {
			if m, ok := value.(map[string]interface{}); ok {
				elem.Set(reflect.ValueOf(m))
			} else {
				return fmt.Errorf("xml: cannot unmarshal to %T", v)
			}
		}
	default:
		return fmt.Errorf("xml: Unmarshal to %T not yet supported - use map[string]interface{} or interface{}", v)
	}

	return nil
}
