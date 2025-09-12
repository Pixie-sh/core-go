package base64

import (
	"bytes"
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "empty string",
			input:    []byte(""),
			expected: "",
		},
		{
			name:     "hello world",
			input:    []byte("hello world"),
			expected: "aGVsbG8gd29ybGQ=",
		},
		{
			name:     "special characters",
			input:    []byte("!@#$%^&*()"),
			expected: "IUAjJCVeJiooKQ==",
		},
		{
			name:     "binary data",
			input:    []byte{0, 1, 2, 3, 4, 5},
			expected: "AAECAwQF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MustEncode(tt.input)
			if result != tt.expected {
				t.Errorf("Encode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []byte
		wantErr  bool
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []byte(""),
			wantErr:  false,
		},
		{
			name:     "hello world",
			input:    "aGVsbG8gd29ybGQ=",
			expected: []byte("hello world"),
			wantErr:  false,
		},
		{
			name:     "special characters",
			input:    "IUAjJCVeJiooKQ==",
			expected: []byte("!@#$%^&*()"),
			wantErr:  false,
		},
		{
			name:     "binary data",
			input:    "AAECAwQF",
			expected: []byte{0, 1, 2, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "invalid base64",
			input:    "invalid!base64",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Decode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !bytes.Equal([]byte(result), tt.expected) {
				t.Errorf("Decode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "empty",
			input: []byte(""),
		},
		{
			name:  "hello world",
			input: []byte("hello world"),
		},
		{
			name:  "longer text",
			input: []byte("The quick brown fox jumps over the lazy dog"),
		},
		{
			name:  "binary data",
			input: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := MustEncode(tt.input)
			decoded, err := Decode(encoded)
			if err != nil {
				t.Errorf("Failed to decode: %v", err)
				return
			}
			if !bytes.Equal([]byte(decoded), tt.input) {
				t.Errorf("Round trip failed. Got %v, want %v", decoded, tt.input)
			}
		})
	}
}

// Benchmark tests if needed
func BenchmarkEncode(b *testing.B) {
	data := bytes.Repeat([]byte("benchmark"), 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Encode(data)
	}
}

func BenchmarkDecode(b *testing.B) {
	data := MustEncode(bytes.Repeat([]byte("benchmark"), 100))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Decode(data)
	}
}
