package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertJSONNumbersToUint64(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "Integer as float64",
			input:    float64(42),
			expected: uint64(42),
		},
		{
			name:     "Float with decimal part",
			input:    float64(42.5),
			expected: float64(42.5),
		},
		{
			name:     "String value",
			input:    "test",
			expected: "test",
		},
		{
			name:     "Map with integer values",
			input:    map[string]interface{}{"a": float64(1), "b": float64(2.5)},
			expected: map[string]interface{}{"a": uint64(1), "b": float64(2.5)},
		},
		{
			name:     "Array with mixed values",
			input:    []interface{}{float64(1), "test", float64(3.5)},
			expected: []interface{}{uint64(1), "test", float64(3.5)},
		},
		{
			name:     "Nested structures",
			input:    map[string]interface{}{"a": []interface{}{float64(1), float64(2.5)}, "b": map[string]interface{}{"c": float64(3)}},
			expected: map[string]interface{}{"a": []interface{}{uint64(1), float64(2.5)}, "b": map[string]interface{}{"c": uint64(3)}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ConvertJSONNumbersToUint64(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
