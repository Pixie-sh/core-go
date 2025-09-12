package serializer

import (
	"encoding/json"
	"testing"

	gojson "github.com/goccy/go-json"
)

// Simple Struct
type SimpleStruct struct {
	Name    string `json:"name"`
	Age     int    `json:"age"`
	IsValid bool   `json:"isValid"`
}

// Complex Struct
type ComplexStruct struct {
	ID          int64             `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags"`
	Metadata    map[string]string `json:"metadata"`
	Author      struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"author"`
	Items []*SimpleStruct `json:"items"`
}

// EasyJSON generated marshaller (assuming you've run gojson -all structs.go)
// easyjson:json
type EasyJSONSimpleStruct struct {
	Name    string `json:"name"`
	Age     int    `json:"age"`
	IsValid bool   `json:"isValid"`
}

// easyjson:json
type EasyJSONComplexStruct struct {
	ID          int64                       `json:"id"`
	Title       string                      `json:"title"`
	Description string                      `json:"description"`
	Tags        []string                    `json:"tags"`
	Metadata    map[string]string           `json:"metadata"`
	Author      EasyJSONComplexStructAuthor `json:"author"`
	Items       []*EasyJSONSimpleStruct     `json:"items"`
}

// easyjson:json
type EasyJSONComplexStructAuthor struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func generateSimpleStruct() SimpleStruct {
	return SimpleStruct{
		Name:    "Example Name",
		Age:     30,
		IsValid: true,
	}
}

func generateComplexStruct() ComplexStruct {
	return ComplexStruct{
		ID:          12345,
		Title:       "Benchmark Test Data",
		Description: "This is a sample complex struct for benchmarking JSON serialization.",
		Tags:        []string{"benchmark", "json", "go"},
		Metadata: map[string]string{
			"version": "1.0",
			"type":    "test",
		},
		Author: struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}{
			Name:  "Benchmark Author",
			Email: "author@example.com",
		},
		Items: []*SimpleStruct{
			&SimpleStruct{Name: "Item 1", Age: 10, IsValid: true},
			&SimpleStruct{Name: "Item 2", Age: 20, IsValid: false},
		},
	}
}

func generateLargeSliceOfComplexStructs(n int) []ComplexStruct {
	slice := make([]ComplexStruct, n)
	for i := 0; i < n; i++ {
		slice[i] = generateComplexStruct()
	}
	return slice
}

// Benchmarks for encoding/json (Go 1.24)
func BenchmarkStdJSONMarshal_Simple_Go124(b *testing.B) {
	data := generateSimpleStruct()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(data)
	}
}

func BenchmarkStdJSONMarshal_Complex_Go124(b *testing.B) {
	data := generateComplexStruct()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(data)
	}
}

func BenchmarkStdJSONMarshal_LargeSlice_Go124(b *testing.B) {
	data := generateLargeSliceOfComplexStructs(1000) // Large slice of 1000 structs
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(data)
	}
}

// Benchmarks for gojson (easyjson)
func BenchmarkGoJSONMarshal_Simple(b *testing.B) {
	data := EasyJSONSimpleStruct{Name: "Example Name", Age: 30, IsValid: true}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gojson.Marshal(data)
	}
}

func BenchmarkGoJSONMarshal_Complex(b *testing.B) {
	data := generateEasyJSONComplexStruct()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gojson.Marshal(data)
	}
}

func BenchmarkGoJSONMarshal_LargeSlice(b *testing.B) {
	data := generateLargeSliceOfEasyJSONComplexStructs(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gojson.Marshal(data)
	}
}

func generateEasyJSONComplexStruct() EasyJSONComplexStruct {
	return EasyJSONComplexStruct{
		ID:          12345,
		Title:       "Benchmark Test Data",
		Description: "This is a sample complex struct for benchmarking JSON serialization.",
		Tags:        []string{"benchmark", "json", "go"},
		Metadata: map[string]string{
			"version": "1.0",
			"type":    "test",
		},
		Author: EasyJSONComplexStructAuthor{
			Name:  "Benchmark Author",
			Email: "author@example.com",
		},
		Items: []*EasyJSONSimpleStruct{
			&EasyJSONSimpleStruct{Name: "Item 1", Age: 10, IsValid: true},
			&EasyJSONSimpleStruct{Name: "Item 2", Age: 20, IsValid: false},
		},
	}
}

func generateLargeSliceOfEasyJSONComplexStructs(n int) []EasyJSONComplexStruct {
	slice := make([]EasyJSONComplexStruct, n)
	for i := 0; i < n; i++ {
		slice[i] = generateEasyJSONComplexStruct()
	}
	return slice
}
