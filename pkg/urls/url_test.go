package urls

import (
	"testing"
)

func TestExtractHost(t *testing.T) {
	testCases := []struct {
		name     string
		url      string
		expected string
	}{
		{"SimpleURL", "https://example.com", "example.com"},
		{"WithSubDomain", "https://sub.example.com", "sub.example.com"},
		{"WithPort", "https://example.com:1234", "example.com"},
		{"WithPath", "https://example.com/path/anotherpath", "example.com"},
		{"WithQueryParams", "https://example.com?test=test", "example.com"},
		{"WithFragment", "https://example.com#section", "example.com"},
		{"IpAddress", "http://192.168.1.1:8080", "192.168.1.1"},
		{"Localhost", "http://192.168.1.1:8080", "192.168.1.1"},
		{"Localhost", "NotAnUrl", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ExtractHost(tc.url)
			if err != nil {
				t.Errorf("ExtractHost(%v) failed but shouldn't", tc.url)
			}
			if result != tc.expected {
				t.Errorf("ExtractHost(%v) = %s; want %s", tc.url, result, tc.expected)
			}
		})
	}
}
