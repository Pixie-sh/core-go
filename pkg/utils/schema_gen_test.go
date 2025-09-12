package utils

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email,omitempty"`
}

type NestedStruct struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type ComplexStruct struct {
	ID      string       `json:"id"`
	Nested  NestedStruct `json:"nested"`
	Tags    []string     `json:"tags"`
	Active  bool         `json:"active"`
	Count   int          `json:"count"`
	Balance float64      `json:"balance"`
}

func TestGenerateJSONSchema(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected map[string]interface{}
	}{
		{
			name:  "Simple struct",
			input: TestStruct{},
			expected: map[string]interface{}{
				"schema": []byte("{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"properties\":{\"age\":{\"type\":\"integer\"},\"email\":{\"type\":\"string\"},\"id\":{\"type\":\"string\"},\"name\":{\"type\":\"string\"}},\"required\":[\"id\",\"name\",\"age\"],\"type\":\"object\"}"),
			},
		},
		{
			name:  "Complex struct",
			input: ComplexStruct{},
			expected: map[string]interface{}{
				"schema": []byte("{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"properties\":{\"active\":{\"type\":\"boolean\"},\"balance\":{\"type\":\"number\"},\"count\":{\"type\":\"integer\"},\"id\":{\"type\":\"string\"},\"nested\":{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"properties\":{\"body\":{\"type\":\"string\"},\"title\":{\"type\":\"string\"}},\"required\":[\"title\",\"body\"],\"type\":\"object\"},\"tags\":{\"items\":{\"type\":\"string\"},\"type\":\"array\"}},\"required\":[\"id\",\"nested\",\"tags\",\"active\",\"count\",\"balance\"],\"type\":\"object\"}"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := SchemaJSON(tt.input)
			bbbb, _ := json.Marshal(schema)
			assert.Equal(t, tt.expected["schema"], bbbb)
		})
	}
}
