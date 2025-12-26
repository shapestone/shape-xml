package xml

import (
	"reflect"
	"strings"
)

// fieldInfo contains parsed information from a struct field's xml tag
type fieldInfo struct {
	name      string // XML field name (empty means use Go field name)
	attr      bool   // field is an XML attribute (attr option)
	cdata     bool   // field is CDATA content (cdata option)
	chardata  bool   // field is text content (chardata option)
	omitEmpty bool   // omitempty option
	skip      bool   // skip this field (tag is "-")
}

// parseTag parses a struct field's xml tag value
// Format: "fieldname" or "fieldname,option1,option2"
// Options: attr, cdata, chardata, omitempty
// Special: "-" means skip field
//
// XML tag conventions:
//   - attr: Field is an XML attribute
//   - chardata: Field contains text content
//   - cdata: Field contains CDATA content
//   - omitempty: Omit field if value is empty
func parseTag(tag string) fieldInfo {
	info := fieldInfo{}

	if tag == "-" {
		info.name = "-"
		info.skip = true
		return info
	}

	parts := strings.Split(tag, ",")
	if len(parts) > 0 {
		info.name = parts[0]
	}

	// Parse options
	for i := 1; i < len(parts); i++ {
		switch strings.TrimSpace(parts[i]) {
		case "attr":
			info.attr = true
		case "cdata":
			info.cdata = true
		case "chardata":
			info.chardata = true
		case "omitempty":
			info.omitEmpty = true
		}
	}

	return info
}

// getFieldInfo extracts field information from a struct field
// Returns fieldInfo with the XML name and options
func getFieldInfo(field reflect.StructField) fieldInfo {
	tag := field.Tag.Get("xml")

	info := parseTag(tag)

	// If no name specified in tag, use the Go field name
	if info.name == "" && !info.skip {
		info.name = field.Name
	}

	return info
}

// isEmptyValue reports whether v is empty according to omitempty rules
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}
