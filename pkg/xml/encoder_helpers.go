package xml

import (
	"reflect"
	"strconv"
)

// appendEscapeXML appends XML-escaped text to buf without allocating.
// Handles: & < > " '
// This matches the behavior of html.EscapeString used by escapeXML in render.go.
func appendEscapeXML(buf []byte, s string) []byte {
	start := 0
	for i := 0; i < len(s); i++ {
		var esc string
		switch s[i] {
		case '&':
			esc = "&amp;"
		case '<':
			esc = "&lt;"
		case '>':
			esc = "&gt;"
		case '"':
			esc = "&#34;"
		case '\'':
			esc = "&#39;"
		default:
			continue
		}
		buf = append(buf, s[start:i]...)
		buf = append(buf, esc...)
		start = i + 1
	}
	buf = append(buf, s[start:]...)
	return buf
}

// appendFormatValue appends a formatted reflect.Value to buf without allocating.
// Zero-alloc replacement for formatValue() which returns string.
func appendFormatValue(buf []byte, rv reflect.Value) []byte {
	if !rv.IsValid() {
		return buf
	}
	switch rv.Kind() {
	case reflect.String:
		return append(buf, rv.String()...)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.AppendInt(buf, rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.AppendUint(buf, rv.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.AppendFloat(buf, rv.Float(), 'g', -1, 64)
	case reflect.Bool:
		return strconv.AppendBool(buf, rv.Bool())
	case reflect.Ptr, reflect.Interface:
		if rv.IsNil() {
			return buf
		}
		return appendFormatValue(buf, rv.Elem())
	default:
		return buf
	}
}
