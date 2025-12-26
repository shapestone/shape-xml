package xml

import (
	"strings"
	"testing"
)

func TestMarshal_String(t *testing.T) {
	type TestStruct struct {
		Value string
	}
	s := TestStruct{Value: "hello"}

	bytes, err := Marshal(s)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	result := string(bytes)
	if !strings.Contains(result, "hello") {
		t.Errorf("Expected 'hello', got: %s", result)
	}
}

func TestMarshal_Attributes(t *testing.T) {
	type User struct {
		ID   string `xml:"id,attr"`
		Name string `xml:"name,attr"`
	}
	user := User{ID: "123", Name: "Alice"}

	bytes, err := Marshal(user)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	result := string(bytes)
	if !strings.Contains(result, `id="123"`) {
		t.Errorf("Expected id attribute, got: %s", result)
	}
	if !strings.Contains(result, `name="Alice"`) {
		t.Errorf("Expected name attribute, got: %s", result)
	}
}

func TestMarshal_TextContent(t *testing.T) {
	type Message struct {
		Content string `xml:",chardata"`
	}
	msg := Message{Content: "Hello, World!"}

	bytes, err := Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	result := string(bytes)
	if !strings.Contains(result, "Hello, World!") {
		t.Errorf("Expected text content, got: %s", result)
	}
}

func TestMarshal_AttributesAndText(t *testing.T) {
	type User struct {
		ID   string `xml:"id,attr"`
		Name string `xml:",chardata"`
	}
	user := User{ID: "123", Name: "Alice"}

	bytes, err := Marshal(user)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	result := string(bytes)
	if !strings.Contains(result, `id="123"`) {
		t.Errorf("Expected id attribute, got: %s", result)
	}
	if !strings.Contains(result, "Alice") {
		t.Errorf("Expected text content, got: %s", result)
	}
}

func TestMarshal_NestedStructs(t *testing.T) {
	type Address struct {
		City string
		Zip  string
	}
	type User struct {
		Name    string
		Address Address
	}
	user := User{
		Name: "Alice",
		Address: Address{
			City: "NYC",
			Zip:  "10001",
		},
	}

	bytes, err := Marshal(user)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	result := string(bytes)
	if !strings.Contains(result, "Alice") {
		t.Errorf("Expected Name, got: %s", result)
	}
	if !strings.Contains(result, "NYC") {
		t.Errorf("Expected City, got: %s", result)
	}
	if !strings.Contains(result, "10001") {
		t.Errorf("Expected Zip, got: %s", result)
	}
}

func TestMarshal_OmitEmpty(t *testing.T) {
	type User struct {
		Name  string
		Email string `xml:",omitempty"`
	}
	user := User{Name: "Alice", Email: ""}

	bytes, err := Marshal(user)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	result := string(bytes)
	if strings.Contains(result, "Email") {
		t.Errorf("Expected Email to be omitted, got: %s", result)
	}
	if !strings.Contains(result, "Alice") {
		t.Errorf("Expected Name, got: %s", result)
	}
}

func TestMarshal_Skip(t *testing.T) {
	type User struct {
		Name     string
		Password string `xml:"-"`
	}
	user := User{Name: "Alice", Password: "secret"}

	bytes, err := Marshal(user)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	result := string(bytes)
	if strings.Contains(result, "secret") || strings.Contains(result, "Password") {
		t.Errorf("Expected Password to be skipped, got: %s", result)
	}
}

func TestMarshal_Numbers(t *testing.T) {
	type Stats struct {
		Count int
		Score float64
	}
	stats := Stats{Count: 42, Score: 3.14}

	bytes, err := Marshal(stats)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	result := string(bytes)
	if !strings.Contains(result, "42") {
		t.Errorf("Expected Count, got: %s", result)
	}
	if !strings.Contains(result, "3.14") {
		t.Errorf("Expected Score, got: %s", result)
	}
}

func TestMarshal_Bool(t *testing.T) {
	type Config struct {
		Enabled bool
	}
	config := Config{Enabled: true}

	bytes, err := Marshal(config)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	result := string(bytes)
	if !strings.Contains(result, "true") {
		t.Errorf("Expected 'true', got: %s", result)
	}
}

func TestMarshal_Slice(t *testing.T) {
	type List struct {
		Items []string
	}
	list := List{Items: []string{"apple", "banana", "cherry"}}

	bytes, err := Marshal(list)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	result := string(bytes)
	if !strings.Contains(result, "apple") {
		t.Errorf("Expected 'apple', got: %s", result)
	}
	if !strings.Contains(result, "banana") {
		t.Errorf("Expected 'banana', got: %s", result)
	}
}

func TestMarshal_Map(t *testing.T) {
	m := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	bytes, err := Marshal(m)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	result := string(bytes)
	if !strings.Contains(result, "key1") {
		t.Errorf("Expected 'key1', got: %s", result)
	}
	if !strings.Contains(result, "value1") {
		t.Errorf("Expected 'value1', got: %s", result)
	}
}

func TestUnmarshal_Simple(t *testing.T) {
	input := `<user id="123">Alice</user>`
	var result map[string]interface{}

	err := Unmarshal([]byte(input), &result)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Check for @id attribute
	if id, ok := result["@id"]; !ok {
		t.Error("Expected @id in result")
	} else if idStr, ok := id.(string); !ok || idStr != "123" {
		t.Errorf("Expected @id='123', got %v", id)
	}
}

func TestUnmarshal_Interface(t *testing.T) {
	input := `<user id="123">Alice</user>`
	var result interface{}

	err := Unmarshal([]byte(input), &result)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Should be a map
	if m, ok := result.(map[string]interface{}); !ok {
		t.Errorf("Expected map, got %T", result)
	} else if len(m) == 0 {
		t.Error("Expected non-empty map")
	}
}

func TestUnmarshal_Invalid(t *testing.T) {
	input := `<invalid`
	var result interface{}

	err := Unmarshal([]byte(input), &result)
	if err == nil {
		t.Error("Expected error for invalid XML")
	}
}

func TestUnmarshal_NotPointer(t *testing.T) {
	input := `<user>Alice</user>`
	var result map[string]interface{}

	err := Unmarshal([]byte(input), result) // Not a pointer
	if err == nil {
		t.Error("Expected error when not passing pointer")
	}
}
