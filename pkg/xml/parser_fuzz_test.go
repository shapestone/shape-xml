package xml

import (
	"testing"
)

// FuzzParse fuzzes the Parse function with random XML input
func FuzzParse(f *testing.F) {
	// Seed corpus with valid XML examples
	f.Add("<root></root>")
	f.Add("<user id=\"123\">Alice</user>")
	f.Add("<empty/>")
	f.Add("<?xml version=\"1.0\"?><root/>")
	f.Add("<nested><child><grandchild/></child></nested>")

	f.Fuzz(func(t *testing.T, input string) {
		// Just ensure Parse doesn't panic
		// Errors are expected for invalid input
		_, _ = Parse(input)
	})
}

// FuzzValidate fuzzes the Validate function with random XML input
func FuzzValidate(f *testing.F) {
	// Seed corpus with valid and invalid XML examples
	f.Add("<root></root>")
	f.Add("<user id=\"123\">Alice</user>")
	f.Add("<empty/>")
	f.Add("invalid")
	f.Add("<unclosed")

	f.Fuzz(func(t *testing.T, input string) {
		// Just ensure Validate doesn't panic
		// Errors are expected for invalid input
		_ = Validate(input)
	})
}

// FuzzRender fuzzes the Render function by parsing and rendering
func FuzzRender(f *testing.F) {
	// Seed corpus with valid XML
	f.Add("<root></root>")
	f.Add("<user id=\"123\">Alice</user>")
	f.Add("<empty/>")

	f.Fuzz(func(t *testing.T, input string) {
		// Parse first
		node, err := Parse(input)
		if err != nil {
			// Invalid XML, skip rendering
			return
		}

		// Ensure Render doesn't panic on valid AST
		_, _ = Render(node)
	})
}

// FuzzMarshal fuzzes the Marshal function
func FuzzMarshal(f *testing.F) {
	// Seed with simple string values
	f.Add("hello")
	f.Add("world")
	f.Add("")

	f.Fuzz(func(t *testing.T, input string) {
		// Create a simple struct
		type TestStruct struct {
			Value string
		}
		s := TestStruct{Value: input}

		// Ensure Marshal doesn't panic
		_, _ = Marshal(s)
	})
}
