package fastparser

import (
	"errors"
	"fmt"
	"reflect"
)

// Unmarshaler is the interface implemented by types that can unmarshal an XML description of themselves.
type Unmarshaler interface {
	UnmarshalXML([]byte) error
}

// Unmarshal parses XML and unmarshals it into the value pointed to by v.
// This is the fast path that bypasses AST construction.
func Unmarshal(data []byte, v interface{}) error {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() || v == nil {
		return errors.New("xml: Unmarshal(nil)")
	}

	if rv.Kind() != reflect.Ptr {
		return errors.New("xml: Unmarshal(non-pointer " + rv.Type().String() + ")")
	}

	if rv.IsNil() {
		return errors.New("xml: Unmarshal(nil " + rv.Type().String() + ")")
	}

	// Check if type implements Unmarshaler interface
	if rv.Type().Implements(reflect.TypeOf((*Unmarshaler)(nil)).Elem()) {
		unmarshaler := rv.Interface().(Unmarshaler)
		return unmarshaler.UnmarshalXML(data)
	}

	p := NewParser(data)
	// Parse to map[string]interface{}
	value, err := p.Parse()
	if err != nil {
		return err
	}

	// Unmarshal from the parsed map
	return unmarshalValue(value, rv.Elem())
}

// UnmarshalValue unmarshals a parsed value into a reflect.Value.
// This is exported for use by the AST path unmarshal function.
func UnmarshalValue(value interface{}, rv reflect.Value) error {
	return unmarshalValue(value, rv)
}

// unmarshalValue unmarshals a parsed value into a reflect.Value.
func unmarshalValue(value interface{}, rv reflect.Value) error {
	if value == nil {
		rv.Set(reflect.Zero(rv.Type()))
		return nil
	}

	// Handle interface{} specially
	if rv.Kind() == reflect.Interface && rv.NumMethod() == 0 {
		rv.Set(reflect.ValueOf(value))
		return nil
	}

	// Handle pointers
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		return unmarshalValue(value, rv.Elem())
	}

	// Route based on Go type
	switch v := value.(type) {
	case map[string]interface{}:
		// If target is a string and map has #text, extract text content
		if rv.Kind() == reflect.String {
			text := extractTextContent(v)
			return unmarshalString(text, rv)
		}
		switch rv.Kind() {
		case reflect.Struct:
			return unmarshalStruct(v, rv)
		case reflect.Map:
			return unmarshalMap(v, rv)
		default:
			return fmt.Errorf("xml: cannot unmarshal object into Go value of type %s", rv.Type())
		}
	case []interface{}:
		return unmarshalArray(v, rv)
	case string:
		return unmarshalString(v, rv)
	default:
		return fmt.Errorf("xml: unexpected value type %T", value)
	}
}

// unmarshalStruct unmarshals a map into a struct.
func unmarshalStruct(m map[string]interface{}, rv reflect.Value) error {
	structType := rv.Type()

	// Build field map
	fieldMap := make(map[string]int)
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if field.PkgPath != "" { // Skip unexported fields
			continue
		}

		// Check XML tag
		tag := field.Tag.Get("xml")
		if tag == "-" {
			continue
		}

		// Get XML name from tag or use field name
		xmlName := field.Name
		isAttr := false
		isCharData := false

		if tag != "" {
			// Parse tag: "name,attr" or ",chardata"
			for idx := 0; idx < len(tag); idx++ {
				if tag[idx] == ',' {
					if idx > 0 {
						xmlName = tag[:idx]
					}
					remainder := tag[idx+1:]
					if remainder == "attr" {
						isAttr = true
					} else if remainder == "chardata" {
						isCharData = true
					}
					break
				}
			}
			if !isAttr && !isCharData && tag[0] != ',' {
				xmlName = tag
			}
		}

		// Map XML name to field index
		if isAttr {
			fieldMap["@"+xmlName] = i
		} else if isCharData {
			fieldMap["#text"] = i
		} else {
			fieldMap[xmlName] = i
		}
	}

	// Populate struct fields from map
	for key, value := range m {
		if fieldIdx, ok := fieldMap[key]; ok {
			fieldValue := rv.Field(fieldIdx)
			if err := unmarshalValue(value, fieldValue); err != nil {
				return fmt.Errorf("field %s: %w", structType.Field(fieldIdx).Name, err)
			}
		}
	}

	return nil
}

// unmarshalMap unmarshals a map into a Go map.
func unmarshalMap(m map[string]interface{}, rv reflect.Value) error {
	if rv.IsNil() {
		rv.Set(reflect.MakeMap(rv.Type()))
	}

	keyType := rv.Type().Key()
	valueType := rv.Type().Elem()

	for k, v := range m {
		keyValue := reflect.ValueOf(k)
		if !keyValue.Type().AssignableTo(keyType) {
			return fmt.Errorf("xml: map key type mismatch: cannot assign %s to %s", keyValue.Type(), keyType)
		}

		elemValue := reflect.New(valueType).Elem()
		if err := unmarshalValue(v, elemValue); err != nil {
			return fmt.Errorf("map key %s: %w", k, err)
		}

		rv.SetMapIndex(keyValue, elemValue)
	}

	return nil
}

// unmarshalArray unmarshals an array into a Go slice.
func unmarshalArray(arr []interface{}, rv reflect.Value) error {
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return fmt.Errorf("xml: cannot unmarshal array into Go value of type %s", rv.Type())
	}

	if rv.Kind() == reflect.Slice {
		rv.Set(reflect.MakeSlice(rv.Type(), len(arr), len(arr)))
	}

	for i, elem := range arr {
		if i >= rv.Len() {
			break // Array is full
		}
		if err := unmarshalValue(elem, rv.Index(i)); err != nil {
			return fmt.Errorf("array index %d: %w", i, err)
		}
	}

	return nil
}

// unmarshalString unmarshals a string or map with #text into a Go value.
func unmarshalString(s string, rv reflect.Value) error {
	switch rv.Kind() {
	case reflect.String:
		rv.SetString(s)
		return nil
	case reflect.Interface:
		if rv.NumMethod() == 0 {
			rv.Set(reflect.ValueOf(s))
			return nil
		}
	}
	return fmt.Errorf("xml: cannot unmarshal string into Go value of type %s", rv.Type())
}

// Extract text content from a value that might be a string or map with #text
func extractTextContent(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case map[string]interface{}:
		if text, ok := v["#text"]; ok {
			return extractTextContent(text)
		}
	}
	return ""
}
