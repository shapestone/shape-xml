package xml_test

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strconv"
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

// BenchmarkStdXML_Unmarshal_Large benchmarks standard library XML unmarshaling on large file
func BenchmarkStdXML_Unmarshal_Large(b *testing.B) {
	if err := loadBenchmarkData(); err != nil {
		b.Fatalf("Failed to load benchmark data: %v", err)
	}

	type Category struct {
		Value string `xml:",chardata"`
	}

	type Price struct {
		Currency string `xml:"currency,attr"`
		Value    string `xml:",chardata"`
	}

	type Product struct {
		ID          string     `xml:"id,attr"`
		Name        string     `xml:"name"`
		Description string     `xml:"description"`
		Price       Price      `xml:"price"`
		InStock     string     `xml:"inStock"`
		Categories  []Category `xml:"categories>category"`
	}

	type Catalog struct {
		Products []Product `xml:"products>product"`
	}

	b.SetBytes(int64(len(largeXML)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var catalog Catalog
		err := xml.Unmarshal([]byte(largeXML), &catalog)
		if err != nil {
			b.Fatal(err)
		}
		// Prevent compiler optimization
		_ = catalog
	}
}

// ================================
// Marshal Benchmarks
// ================================

// BenchmarkStdXML_Marshal_Small benchmarks standard library XML marshaling
func BenchmarkStdXML_Marshal_Small(b *testing.B) {
	type User struct {
		ID     string `xml:"id,attr"`
		Active string `xml:"active,attr"`
		Name   string `xml:",chardata"`
	}

	user := User{
		ID:     "123",
		Active: "true",
		Name:   "Alice",
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bytes, err := xml.Marshal(&user)
		if err != nil {
			b.Fatal(err)
		}
		_ = bytes
	}
}

// BenchmarkStdXML_Marshal_Medium benchmarks standard library XML marshaling on medium struct
func BenchmarkStdXML_Marshal_Medium(b *testing.B) {
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

	users := []User{
		{
			ID:    "1",
			Name:  "Alice",
			Email: "alice@example.com",
			Address: Address{
				Street: "123 Main St",
				City:   "Springfield",
				State:  "IL",
				Zip:    "62701",
			},
			Tags: []Tag{
				{Value: "admin"},
				{Value: "user"},
			},
		},
		{
			ID:    "2",
			Name:  "Bob",
			Email: "bob@example.com",
			Address: Address{
				Street: "456 Oak Ave",
				City:   "Springfield",
				State:  "IL",
				Zip:    "62702",
			},
			Tags: []Tag{
				{Value: "user"},
			},
		},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bytes, err := xml.Marshal(&users)
		if err != nil {
			b.Fatal(err)
		}
		_ = bytes
	}
}

// BenchmarkShapeXML_Marshal_Small benchmarks shape-xml marshaling
func BenchmarkShapeXML_Marshal_Small(b *testing.B) {
	type User struct {
		ID     string `xml:"id,attr"`
		Active string `xml:"active,attr"`
		Name   string `xml:",chardata"`
	}

	user := User{
		ID:     "123",
		Active: "true",
		Name:   "Alice",
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bytes, err := shapexml.Marshal(&user)
		if err != nil {
			b.Fatal(err)
		}
		_ = bytes
	}
}

// BenchmarkShapeXML_Marshal_Medium benchmarks shape-xml marshaling on medium struct
func BenchmarkShapeXML_Marshal_Medium(b *testing.B) {
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

	users := []User{
		{
			ID:    "1",
			Name:  "Alice",
			Email: "alice@example.com",
			Address: Address{
				Street: "123 Main St",
				City:   "Springfield",
				State:  "IL",
				Zip:    "62701",
			},
			Tags: []Tag{
				{Value: "admin"},
				{Value: "user"},
			},
		},
		{
			ID:    "2",
			Name:  "Bob",
			Email: "bob@example.com",
			Address: Address{
				Street: "456 Oak Ave",
				City:   "Springfield",
				State:  "IL",
				Zip:    "62702",
			},
			Tags: []Tag{
				{Value: "user"},
			},
		},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bytes, err := shapexml.Marshal(&users)
		if err != nil {
			b.Fatal(err)
		}
		_ = bytes
	}
}

// ================================
// Unmarshal Benchmarks (Fast Path)
// ================================

// BenchmarkShapeXML_Unmarshal_Small benchmarks fast-path unmarshaling
func BenchmarkShapeXML_Unmarshal_Small(b *testing.B) {
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
		err := shapexml.Unmarshal([]byte(smallXML), &user)
		if err != nil {
			b.Fatal(err)
		}
		// Prevent compiler optimization
		_ = user
	}
}

// BenchmarkShapeXML_Unmarshal_Medium benchmarks fast-path unmarshaling on medium file
func BenchmarkShapeXML_Unmarshal_Medium(b *testing.B) {
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
		err := shapexml.Unmarshal([]byte(mediumXML), &users)
		if err != nil {
			b.Fatal(err)
		}
		// Prevent compiler optimization
		_ = users
	}
}

// BenchmarkShapeXML_Unmarshal_Large benchmarks fast-path unmarshaling on large file
func BenchmarkShapeXML_Unmarshal_Large(b *testing.B) {
	if err := loadBenchmarkData(); err != nil {
		b.Fatalf("Failed to load benchmark data: %v", err)
	}

	type Category struct {
		Value string `xml:",chardata"`
	}

	type Price struct {
		Currency string `xml:"currency,attr"`
		Value    string `xml:",chardata"`
	}

	type Product struct {
		ID          string     `xml:"id,attr"`
		Name        string     `xml:"name"`
		Description string     `xml:"description"`
		Price       Price      `xml:"price"`
		InStock     string     `xml:"inStock"`
		Categories  []Category `xml:"categories>category"`
	}

	type Catalog struct {
		Products []Product `xml:"products>product"`
	}

	b.SetBytes(int64(len(largeXML)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var catalog Catalog
		err := shapexml.Unmarshal([]byte(largeXML), &catalog)
		if err != nil {
			b.Fatal(err)
		}
		// Prevent compiler optimization
		_ = catalog
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

// ================================
// Large Marshal Benchmarks
// ================================

// BenchmarkShapeXML_Marshal_Large benchmarks shape-xml marshaling on a large struct with 100 items
func BenchmarkShapeXML_Marshal_Large(b *testing.B) {
	type Item struct {
		ID    string `xml:"id,attr"`
		Name  string `xml:"name"`
		Price string `xml:"price"`
	}
	type Catalog struct {
		Items []Item `xml:"item"`
	}
	cat := Catalog{}
	for i := 0; i < 100; i++ {
		cat.Items = append(cat.Items, Item{
			ID:    strconv.Itoa(i),
			Name:  "Item " + strconv.Itoa(i),
			Price: strconv.FormatFloat(float64(i)*9.99, 'f', 2, 64),
		})
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out, err := shapexml.Marshal(&cat)
		if err != nil {
			b.Fatal(err)
		}
		_ = out
	}
}

// BenchmarkEncodingXML_Marshal_Large benchmarks encoding/xml marshaling on a large struct with 100 items
func BenchmarkEncodingXML_Marshal_Large(b *testing.B) {
	type Item struct {
		XMLName xml.Name `xml:"item"`
		ID      string   `xml:"id,attr"`
		Name    string   `xml:"name"`
		Price   string   `xml:"price"`
	}
	type Catalog struct {
		XMLName xml.Name `xml:"Catalog"`
		Items   []Item   `xml:"item"`
	}
	cat := Catalog{}
	for i := 0; i < 100; i++ {
		cat.Items = append(cat.Items, Item{
			ID:    strconv.Itoa(i),
			Name:  "Item " + strconv.Itoa(i),
			Price: strconv.FormatFloat(float64(i)*9.99, 'f', 2, 64),
		})
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out, err := xml.Marshal(&cat)
		if err != nil {
			b.Fatal(err)
		}
		_ = out
	}
}
