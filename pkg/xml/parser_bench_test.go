package xml_test

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"

	shapexml "github.com/shapestone/shape-xml/pkg/xml"
)

// Benchmark test data is loaded once and reused across all benchmarks
var (
	smallXML  string
	mediumXML string
	largeXML  string
)

// loadBenchmarkData loads test data files once during test initialization
func loadBenchmarkData() error {
	if smallXML != "" {
		return nil // already loaded
	}

	testdataDir := filepath.Join("..", "..", "testdata", "benchmarks")

	// Load small.xml
	smallBytes, err := os.ReadFile(filepath.Join(testdataDir, "small.xml"))
	if err != nil {
		return err
	}
	smallXML = string(smallBytes)

	// Load medium.xml
	mediumBytes, err := os.ReadFile(filepath.Join(testdataDir, "medium.xml"))
	if err != nil {
		return err
	}
	mediumXML = string(mediumBytes)

	// Load large.xml
	largeBytes, err := os.ReadFile(filepath.Join(testdataDir, "large.xml"))
	if err != nil {
		return err
	}
	largeXML = string(largeBytes)

	return nil
}

// ================================
// Shape-XML Benchmarks
// ================================

// BenchmarkShapeXML_Parse_Small benchmarks parsing of small XML (~100 bytes)
// using the shape-xml parser.
func BenchmarkShapeXML_Parse_Small(b *testing.B) {
	if err := loadBenchmarkData(); err != nil {
		b.Fatalf("Failed to load benchmark data: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node, err := shapexml.Parse(smallXML)
		if err != nil {
			b.Fatal(err)
		}
		// Prevent compiler optimization
		_ = node
	}
}

// BenchmarkShapeXML_Parse_Medium benchmarks parsing of medium XML (~1KB)
// using the shape-xml parser.
func BenchmarkShapeXML_Parse_Medium(b *testing.B) {
	if err := loadBenchmarkData(); err != nil {
		b.Fatalf("Failed to load benchmark data: %v", err)
	}

	b.SetBytes(int64(len(mediumXML)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node, err := shapexml.Parse(mediumXML)
		if err != nil {
			b.Fatal(err)
		}
		// Prevent compiler optimization
		_ = node
	}
}

// BenchmarkShapeXML_Parse_Large benchmarks parsing of large XML (~100KB)
// using the shape-xml parser.
func BenchmarkShapeXML_Parse_Large(b *testing.B) {
	if err := loadBenchmarkData(); err != nil {
		b.Fatalf("Failed to load benchmark data: %v", err)
	}

	b.SetBytes(int64(len(largeXML)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node, err := shapexml.Parse(largeXML)
		if err != nil {
			b.Fatal(err)
		}
		// Prevent compiler optimization
		_ = node
	}
}

// BenchmarkShapeXML_ParseReader_Small benchmarks ParseReader with small XML
func BenchmarkShapeXML_ParseReader_Small(b *testing.B) {
	if err := loadBenchmarkData(); err != nil {
		b.Fatalf("Failed to load benchmark data: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(smallXML)
		node, err := shapexml.ParseReader(reader)
		if err != nil {
			b.Fatal(err)
		}
		// Prevent compiler optimization
		_ = node
	}
}

// BenchmarkShapeXML_ParseReader_Medium benchmarks ParseReader with medium XML
func BenchmarkShapeXML_ParseReader_Medium(b *testing.B) {
	if err := loadBenchmarkData(); err != nil {
		b.Fatalf("Failed to load benchmark data: %v", err)
	}

	b.SetBytes(int64(len(mediumXML)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(mediumXML)
		node, err := shapexml.ParseReader(reader)
		if err != nil {
			b.Fatal(err)
		}
		// Prevent compiler optimization
		_ = node
	}
}

// BenchmarkShapeXML_ParseReader_Large benchmarks ParseReader with large XML
func BenchmarkShapeXML_ParseReader_Large(b *testing.B) {
	if err := loadBenchmarkData(); err != nil {
		b.Fatalf("Failed to load benchmark data: %v", err)
	}

	b.SetBytes(int64(len(largeXML)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(largeXML)
		node, err := shapexml.ParseReader(reader)
		if err != nil {
			b.Fatal(err)
		}
		// Prevent compiler optimization
		_ = node
	}
}

// ================================
// Validate Benchmarks
// ================================

// BenchmarkShapeXML_Validate_Small benchmarks fast validation of small XML
func BenchmarkShapeXML_Validate_Small(b *testing.B) {
	if err := loadBenchmarkData(); err != nil {
		b.Fatalf("Failed to load benchmark data: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := shapexml.Validate(smallXML)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkShapeXML_Validate_Medium benchmarks fast validation of medium XML
func BenchmarkShapeXML_Validate_Medium(b *testing.B) {
	if err := loadBenchmarkData(); err != nil {
		b.Fatalf("Failed to load benchmark data: %v", err)
	}

	b.SetBytes(int64(len(mediumXML)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := shapexml.Validate(mediumXML)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkShapeXML_Validate_Large benchmarks fast validation of large XML
func BenchmarkShapeXML_Validate_Large(b *testing.B) {
	if err := loadBenchmarkData(); err != nil {
		b.Fatalf("Failed to load benchmark data: %v", err)
	}

	b.SetBytes(int64(len(largeXML)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := shapexml.Validate(largeXML)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ================================
// Render Benchmarks
// ================================

// BenchmarkShapeXML_Render_Small benchmarks rendering small AST to XML
func BenchmarkShapeXML_Render_Small(b *testing.B) {
	if err := loadBenchmarkData(); err != nil {
		b.Fatalf("Failed to load benchmark data: %v", err)
	}

	// Parse once
	node, err := shapexml.Parse(smallXML)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bytes, err := shapexml.Render(node)
		if err != nil {
			b.Fatal(err)
		}
		// Prevent compiler optimization
		_ = bytes
	}
}

// BenchmarkShapeXML_Render_Medium benchmarks rendering medium AST to XML
func BenchmarkShapeXML_Render_Medium(b *testing.B) {
	if err := loadBenchmarkData(); err != nil {
		b.Fatalf("Failed to load benchmark data: %v", err)
	}

	// Parse once
	node, err := shapexml.Parse(mediumXML)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bytes, err := shapexml.Render(node)
		if err != nil {
			b.Fatal(err)
		}
		// Prevent compiler optimization
		_ = bytes
	}
}

// ================================
// Standard library Comparison
// ================================

// BenchmarkStdXML_Unmarshal_Small benchmarks standard library XML unmarshaling
func BenchmarkStdXML_Unmarshal_Small(b *testing.B) {
	if err := loadBenchmarkData(); err != nil {
		b.Fatalf("Failed to load benchmark data: %v", err)
	}

	type User struct {
		ID     string `xml:"id,attr"`
		Active string `xml:"active,attr"`
		Name   string `xml:",chardata"`
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var user User
		err := xml.Unmarshal([]byte(smallXML), &user)
		if err != nil {
			b.Fatal(err)
		}
		// Prevent compiler optimization
		_ = user
	}
}

// BenchmarkStdXML_Unmarshal_Medium benchmarks standard library XML unmarshaling
func BenchmarkStdXML_Unmarshal_Medium(b *testing.B) {
	if err := loadBenchmarkData(); err != nil {
		b.Fatalf("Failed to load benchmark data: %v", err)
	}

	type Tag struct {
		Value string `xml:",chardata"`
	}

	type Address struct {
		Street string `xml:"street"`
		City   string `xml:"city"`
		State  string `xml:"state"`
		Zip    string `xml:"zip"`
	}

	type User struct {
		ID      string  `xml:"id,attr"`
		Name    string  `xml:"name"`
		Email   string  `xml:"email"`
		Address Address `xml:"address"`
		Tags    []Tag   `xml:"tags>tag"`
	}

	type Users struct {
		Users []User `xml:"user"`
	}

	b.SetBytes(int64(len(mediumXML)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var users Users
		err := xml.Unmarshal([]byte(mediumXML), &users)
		if err != nil {
			b.Fatal(err)
		}
		// Prevent compiler optimization
		_ = users
	}
}

// ================================
// Round-trip Benchmarks
// ================================

// BenchmarkShapeXML_Roundtrip_Small benchmarks parse -> render round trip
func BenchmarkShapeXML_Roundtrip_Small(b *testing.B) {
	if err := loadBenchmarkData(); err != nil {
		b.Fatalf("Failed to load benchmark data: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node, err := shapexml.Parse(smallXML)
		if err != nil {
			b.Fatal(err)
		}

		bytes, err := shapexml.Render(node)
		if err != nil {
			b.Fatal(err)
		}

		// Prevent compiler optimization
		_ = bytes
	}
}

// BenchmarkShapeXML_Roundtrip_Medium benchmarks parse -> render round trip
func BenchmarkShapeXML_Roundtrip_Medium(b *testing.B) {
	if err := loadBenchmarkData(); err != nil {
		b.Fatalf("Failed to load benchmark data: %v", err)
	}

	b.SetBytes(int64(len(mediumXML)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node, err := shapexml.Parse(mediumXML)
		if err != nil {
			b.Fatal(err)
		}

		bytes, err := shapexml.Render(node)
		if err != nil {
			b.Fatal(err)
		}

		// Prevent compiler optimization
		_ = bytes
	}
}
