package parser

import (
	"testing"
)

// FuzzParser fuzzes the internal parser with random XML input
func FuzzParser(f *testing.F) {
	// Seed corpus with valid XML examples
	f.Add("<root></root>")
	f.Add("<user id=\"123\">Alice</user>")
	f.Add("<empty/>")
	f.Add("<?xml version=\"1.0\"?><root/>")
	f.Add("<nested><child><grandchild/></child></nested>")
	f.Add("<![CDATA[some data]]>")
	f.Add("<!-- comment --><root/>")

	f.Fuzz(func(t *testing.T, input string) {
		// Just ensure parser doesn't panic
		// Errors are expected for invalid input
		p := NewParser(input)
		_, _ = p.Parse()
	})
}
