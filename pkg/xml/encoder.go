package xml

import (
	"fmt"
	"reflect"
	"sort"
	"sync"
	"sync/atomic"
)

// xmlEncoderFunc appends XML encoding of rv to buf with the given element name.
type xmlEncoderFunc func(buf []byte, rv reflect.Value, elemName string) ([]byte, error)

// Encoder cache using copy-on-write pattern for lock-free reads.
var xmlEncoderCache atomic.Value
var xmlEncoderMu sync.Mutex

func init() {
	xmlEncoderCache.Store(make(map[reflect.Type]xmlEncoderFunc))
}

var xmlMarshalerType = reflect.TypeOf((*Marshaler)(nil)).Elem()

// xmlBufPool pools []byte slices for the compiled-encoder fast path.
var xmlBufPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 0, 1024)
		return &b
	},
}

// xmlEncoderForType returns a cached encoder for the given type, creating one if needed.
// Uses a copy-on-write map with a placeholder for recursive types.
func xmlEncoderForType(t reflect.Type) xmlEncoderFunc {
	// Fast path: check cache without lock.
	cache := xmlEncoderCache.Load().(map[reflect.Type]xmlEncoderFunc)
	if enc, ok := cache[t]; ok {
		return enc
	}

	// Slow path: build encoder under lock.
	xmlEncoderMu.Lock()

	// Double-check after acquiring lock.
	cache = xmlEncoderCache.Load().(map[reflect.Type]xmlEncoderFunc)
	if enc, ok := cache[t]; ok {
		xmlEncoderMu.Unlock()
		return enc
	}

	// Insert a placeholder to handle recursive types.
	// The placeholder will forward calls to the real encoder once it's built.
	var realEnc xmlEncoderFunc
	placeholder := func(buf []byte, rv reflect.Value, elemName string) ([]byte, error) {
		return realEnc(buf, rv, elemName)
	}

	// COW: copy the map, add placeholder, store.
	newCache := make(map[reflect.Type]xmlEncoderFunc, len(cache)+1)
	for k, v := range cache {
		newCache[k] = v
	}
	newCache[t] = placeholder
	xmlEncoderCache.Store(newCache)

	// Release lock before building so that nested calls to xmlEncoderForType
	// (e.g., for struct child fields) do not deadlock.
	xmlEncoderMu.Unlock()

	// Build the actual encoder. This may recursively call xmlEncoderForType
	// for child types; those calls will find the placeholder in the cache.
	realEnc = buildXMLEncoder(t)

	// Replace placeholder with real encoder under lock.
	xmlEncoderMu.Lock()
	cache = xmlEncoderCache.Load().(map[reflect.Type]xmlEncoderFunc)
	newCache = make(map[reflect.Type]xmlEncoderFunc, len(cache))
	for k, v := range cache {
		newCache[k] = v
	}
	newCache[t] = realEnc
	xmlEncoderCache.Store(newCache)
	xmlEncoderMu.Unlock()

	return realEnc
}

// buildXMLEncoder builds an encoder function for the given type.
func buildXMLEncoder(t reflect.Type) xmlEncoderFunc {
	// Check if the type itself implements Marshaler.
	if t.Implements(xmlMarshalerType) {
		return xmlMarshalerEnc
	}

	// Check if pointer-to-type implements Marshaler.
	if t.Kind() != reflect.Ptr && reflect.PointerTo(t).Implements(xmlMarshalerType) {
		return buildXMLAddrMarshalerEnc(t)
	}

	switch t.Kind() {
	case reflect.Ptr:
		return buildXMLPtrEncoder(t)
	case reflect.Interface:
		return xmlInterfaceEnc
	case reflect.String:
		return xmlStringEnc
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return xmlIntEnc
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return xmlUintEnc
	case reflect.Float32, reflect.Float64:
		return xmlFloatEnc
	case reflect.Bool:
		return xmlBoolEnc
	case reflect.Struct:
		return buildXMLStructEncoder(t)
	case reflect.Map:
		return buildXMLMapEncoder(t)
	case reflect.Slice:
		return buildXMLSliceEncoder(t)
	case reflect.Array:
		return buildXMLArrayEncoder(t)
	default:
		return xmlUnsupportedEnc(t)
	}
}

// ---------- Marshaler encoders ----------

func xmlMarshalerEnc(buf []byte, rv reflect.Value, elemName string) ([]byte, error) {
	marshaler := rv.Interface().(Marshaler)
	b, err := marshaler.MarshalXML()
	if err != nil {
		return buf, err
	}
	return append(buf, b...), nil
}

func buildXMLAddrMarshalerEnc(t reflect.Type) xmlEncoderFunc {
	return func(buf []byte, rv reflect.Value, elemName string) ([]byte, error) {
		if rv.CanAddr() {
			marshaler := rv.Addr().Interface().(Marshaler)
			b, err := marshaler.MarshalXML()
			if err != nil {
				return buf, err
			}
			return append(buf, b...), nil
		}
		// Can't take address; fall back to non-marshaler encoding.
		fallback := buildXMLEncoderNoMarshaler(t)
		return fallback(buf, rv, elemName)
	}
}

// buildXMLEncoderNoMarshaler builds an encoder skipping the Marshaler check.
// Used as fallback when we cannot take the address.
func buildXMLEncoderNoMarshaler(t reflect.Type) xmlEncoderFunc {
	switch t.Kind() {
	case reflect.Ptr:
		return buildXMLPtrEncoder(t)
	case reflect.Interface:
		return xmlInterfaceEnc
	case reflect.String:
		return xmlStringEnc
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return xmlIntEnc
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return xmlUintEnc
	case reflect.Float32, reflect.Float64:
		return xmlFloatEnc
	case reflect.Bool:
		return xmlBoolEnc
	case reflect.Struct:
		return buildXMLStructEncoder(t)
	case reflect.Map:
		return buildXMLMapEncoder(t)
	case reflect.Slice:
		return buildXMLSliceEncoder(t)
	case reflect.Array:
		return buildXMLArrayEncoder(t)
	default:
		return xmlUnsupportedEnc(t)
	}
}

// ---------- Pointer / Interface encoders ----------

func buildXMLPtrEncoder(t reflect.Type) xmlEncoderFunc {
	elemEnc := xmlEncoderForType(t.Elem())
	return func(buf []byte, rv reflect.Value, elemName string) ([]byte, error) {
		if rv.IsNil() {
			buf = append(buf, '<')
			buf = append(buf, elemName...)
			buf = append(buf, '/', '>')
			return buf, nil
		}
		return elemEnc(buf, rv.Elem(), elemName)
	}
}

func xmlInterfaceEnc(buf []byte, rv reflect.Value, elemName string) ([]byte, error) {
	if rv.IsNil() {
		buf = append(buf, '<')
		buf = append(buf, elemName...)
		buf = append(buf, '/', '>')
		return buf, nil
	}
	// Resolve the concrete type at runtime and dispatch.
	elem := rv.Elem()
	enc := xmlEncoderForType(elem.Type())
	return enc(buf, elem, elemName)
}

// ---------- Primitive encoders ----------

func xmlStringEnc(buf []byte, rv reflect.Value, elemName string) ([]byte, error) {
	buf = append(buf, '<')
	buf = append(buf, elemName...)
	buf = append(buf, '>')
	buf = appendEscapeXML(buf, rv.String())
	buf = append(buf, '<', '/')
	buf = append(buf, elemName...)
	buf = append(buf, '>')
	return buf, nil
}

func xmlIntEnc(buf []byte, rv reflect.Value, elemName string) ([]byte, error) {
	buf = append(buf, '<')
	buf = append(buf, elemName...)
	buf = append(buf, '>')
	buf = appendFormatValue(buf, rv)
	buf = append(buf, '<', '/')
	buf = append(buf, elemName...)
	buf = append(buf, '>')
	return buf, nil
}

func xmlUintEnc(buf []byte, rv reflect.Value, elemName string) ([]byte, error) {
	buf = append(buf, '<')
	buf = append(buf, elemName...)
	buf = append(buf, '>')
	buf = appendFormatValue(buf, rv)
	buf = append(buf, '<', '/')
	buf = append(buf, elemName...)
	buf = append(buf, '>')
	return buf, nil
}

func xmlFloatEnc(buf []byte, rv reflect.Value, elemName string) ([]byte, error) {
	buf = append(buf, '<')
	buf = append(buf, elemName...)
	buf = append(buf, '>')
	buf = appendFormatValue(buf, rv)
	buf = append(buf, '<', '/')
	buf = append(buf, elemName...)
	buf = append(buf, '>')
	return buf, nil
}

func xmlBoolEnc(buf []byte, rv reflect.Value, elemName string) ([]byte, error) {
	buf = append(buf, '<')
	buf = append(buf, elemName...)
	buf = append(buf, '>')
	buf = appendFormatValue(buf, rv)
	buf = append(buf, '<', '/')
	buf = append(buf, elemName...)
	buf = append(buf, '>')
	return buf, nil
}

// ---------- Struct encoder ----------

// xmlAttrField holds pre-computed metadata for a struct attribute field.
type xmlAttrField struct {
	index       int    // field index in the struct
	name        string // attribute name for sorting
	prefixBytes []byte // pre-encoded ` name="` (space + name + =")
}

// xmlChildField holds pre-computed metadata for a struct child element field.
type xmlChildField struct {
	index     int
	name      string
	encoder   xmlEncoderFunc
	omitEmpty bool
}

// xmlFieldRef references a struct field by index.
type xmlFieldRef struct {
	index int
}

// xmlStructEncoder holds all pre-computed struct encoding metadata.
type xmlStructEncoder struct {
	attrs    []xmlAttrField
	chardata *xmlFieldRef
	cdata    *xmlFieldRef
	children []xmlChildField
}

func buildXMLStructEncoder(t reflect.Type) xmlEncoderFunc {
	se := &xmlStructEncoder{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields.
		if field.PkgPath != "" {
			continue
		}

		info := getFieldInfo(field)

		// Skip fields with "-" tag.
		if info.skip {
			continue
		}

		if info.attr {
			// Pre-encode attribute prefix: ` name="`
			prefix := make([]byte, 0, 1+len(info.name)+2)
			prefix = append(prefix, ' ')
			prefix = append(prefix, info.name...)
			prefix = append(prefix, '=', '"')

			se.attrs = append(se.attrs, xmlAttrField{
				index:       i,
				name:        info.name,
				prefixBytes: prefix,
			})
			continue
		}

		if info.chardata {
			se.chardata = &xmlFieldRef{index: i}
			continue
		}

		if info.cdata {
			se.cdata = &xmlFieldRef{index: i}
			continue
		}

		// Regular child element - resolve encoder.
		childEnc := xmlEncoderForType(field.Type)

		se.children = append(se.children, xmlChildField{
			index:     i,
			name:      info.name,
			encoder:   childEnc,
			omitEmpty: info.omitEmpty,
		})
	}

	// Sort attributes by name for deterministic output.
	sort.Slice(se.attrs, func(i, j int) bool {
		return se.attrs[i].name < se.attrs[j].name
	})

	return func(buf []byte, rv reflect.Value, elemName string) ([]byte, error) {
		// Start opening tag: `<elemName`
		buf = append(buf, '<')
		buf = append(buf, elemName...)

		// Write sorted attributes.
		for _, attr := range se.attrs {
			fv := rv.Field(attr.index)
			attrVal := formatValue(fv)
			if attrVal != "" {
				buf = append(buf, attr.prefixBytes...)
				buf = appendEscapeXML(buf, attrVal)
				buf = append(buf, '"')
			}
		}

		// Check if there is any content.
		hasContent := false

		if se.chardata != nil {
			fv := rv.Field(se.chardata.index)
			if formatValue(fv) != "" {
				hasContent = true
			}
		}

		if !hasContent && se.cdata != nil {
			fv := rv.Field(se.cdata.index)
			if formatValue(fv) != "" {
				hasContent = true
			}
		}

		if !hasContent {
			for _, child := range se.children {
				fv := rv.Field(child.index)
				if child.omitEmpty && isEmptyValue(fv) {
					continue
				}
				hasContent = true
				break
			}
		}

		if !hasContent {
			buf = append(buf, '/', '>')
			return buf, nil
		}

		// Close opening tag.
		buf = append(buf, '>')

		// Write chardata content.
		if se.chardata != nil {
			fv := rv.Field(se.chardata.index)
			val := formatValue(fv)
			if val != "" {
				buf = appendEscapeXML(buf, val)
			}
		}

		// Write CDATA content.
		if se.cdata != nil {
			fv := rv.Field(se.cdata.index)
			val := formatValue(fv)
			if val != "" {
				buf = append(buf, "<![CDATA["...)
				buf = append(buf, val...)
				buf = append(buf, "]]>"...)
			}
		}

		// Write child elements.
		var err error
		for _, child := range se.children {
			fv := rv.Field(child.index)
			if child.omitEmpty && isEmptyValue(fv) {
				continue
			}
			buf, err = child.encoder(buf, fv, child.name)
			if err != nil {
				return buf, err
			}
		}

		// Close element.
		buf = append(buf, '<', '/')
		buf = append(buf, elemName...)
		buf = append(buf, '>')

		return buf, nil
	}
}

// ---------- Map encoder ----------

func buildXMLMapEncoder(t reflect.Type) xmlEncoderFunc {
	if t.Key().Kind() != reflect.String {
		return func(buf []byte, rv reflect.Value, elemName string) ([]byte, error) {
			return buf, fmt.Errorf("xml: unsupported map key type %s", t.Key())
		}
	}

	return func(buf []byte, rv reflect.Value, elemName string) ([]byte, error) {
		if rv.IsNil() {
			buf = append(buf, '<')
			buf = append(buf, elemName...)
			buf = append(buf, '/', '>')
			return buf, nil
		}

		// Opening tag.
		buf = append(buf, '<')
		buf = append(buf, elemName...)
		buf = append(buf, '>')

		// Sort keys for deterministic output.
		keys := rv.MapKeys()
		strKeys := make([]string, len(keys))
		for i, key := range keys {
			strKeys[i] = key.String()
		}
		sort.Strings(strKeys)

		// Encode each value. We resolve the encoder per-value because map values
		// can be interface{} and the concrete type may vary.
		for _, keyStr := range strKeys {
			val := rv.MapIndex(reflect.ValueOf(keyStr))
			// Resolve concrete type for interface values.
			actual := val
			for actual.Kind() == reflect.Interface && !actual.IsNil() {
				actual = actual.Elem()
			}
			enc := xmlEncoderForType(actual.Type())
			var err error
			buf, err = enc(buf, actual, keyStr)
			if err != nil {
				return buf, err
			}
		}

		// Close element.
		buf = append(buf, '<', '/')
		buf = append(buf, elemName...)
		buf = append(buf, '>')

		return buf, nil
	}
}

// ---------- Slice / Array encoder ----------

func buildXMLSliceEncoder(t reflect.Type) xmlEncoderFunc {
	elemEnc := xmlEncoderForType(t.Elem())

	return func(buf []byte, rv reflect.Value, elemName string) ([]byte, error) {
		// Nil slices encode as self-closing element.
		if rv.IsNil() {
			buf = append(buf, '<')
			buf = append(buf, elemName...)
			buf = append(buf, '/', '>')
			return buf, nil
		}

		// Encode each element with the same element name.
		length := rv.Len()
		for i := 0; i < length; i++ {
			var err error
			buf, err = elemEnc(buf, rv.Index(i), elemName)
			if err != nil {
				return buf, err
			}
		}

		return buf, nil
	}
}

func buildXMLArrayEncoder(t reflect.Type) xmlEncoderFunc {
	elemEnc := xmlEncoderForType(t.Elem())

	return func(buf []byte, rv reflect.Value, elemName string) ([]byte, error) {
		// Encode each element with the same element name.
		length := rv.Len()
		for i := 0; i < length; i++ {
			var err error
			buf, err = elemEnc(buf, rv.Index(i), elemName)
			if err != nil {
				return buf, err
			}
		}

		return buf, nil
	}
}

// ---------- Unsupported ----------

func xmlUnsupportedEnc(t reflect.Type) xmlEncoderFunc {
	return func(buf []byte, rv reflect.Value, elemName string) ([]byte, error) {
		return buf, fmt.Errorf("xml: unsupported type %s", t)
	}
}
