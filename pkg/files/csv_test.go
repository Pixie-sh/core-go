package files

import (
	"reflect"
	"testing"
)

// Test structs
type UserMapToStrucsTest struct {
	InstagramUsername string
	Email             string
	Name              string
}

type SimpleUserMapToStructsTest struct {
	Username string
	Email    string
}

type SingleFieldTest struct {
	Value string
}

func TestCSVDataMapToStructsGeneric(t *testing.T) {
	tests := []struct {
		name        string
		input       CSVData
		expected    []UserMapToStrucsTest
		expectError bool
	}{
		{
			name: "Basic mapping with all fields",
			input: CSVData{
				Headers: []string{"Instagram username", "email", "full_name"},
				Records: []CSVRecord{
					{
						"Instagram username": "margoulette",
						"email":              "test@testemail.fr",
						"full_name":          "Marie Goulette",
					},
					{
						"Instagram username": "pitounet",
						"email":              "test2@testagain.fr",
						"full_name":          "Pierre Pitou",
					},
				},
			},
			expected: []UserMapToStrucsTest{
				{
					InstagramUsername: "margoulette",
					Email:             "test@testemail.fr",
					Name:              "Marie Goulette",
				},
				{
					InstagramUsername: "pitounet",
					Email:             "test2@testagain.fr",
					Name:              "Pierre Pitou",
				},
			},
			expectError: false,
		},
		{
			name: "Empty records",
			input: CSVData{
				Headers: []string{"Instagram username", "email", "full_name"},
				Records: []CSVRecord{},
			},
			expected:    []UserMapToStrucsTest{},
			expectError: false,
		},
		{
			name: "Single record",
			input: CSVData{
				Headers: []string{"Instagram username", "email", "full_name"},
				Records: []CSVRecord{
					{
						"Instagram username": "single_user",
						"email":              "single@test.com",
						"full_name":          "Single UserMapToStrucsTest",
					},
				},
			},
			expected: []UserMapToStrucsTest{
				{
					InstagramUsername: "single_user",
					Email:             "single@test.com",
					Name:              "Single UserMapToStrucsTest",
				},
			},
			expectError: false,
		},
		{
			name: "Partial data - missing values",
			input: CSVData{
				Headers: []string{"Instagram username", "email", "full_name"},
				Records: []CSVRecord{
					{
						"Instagram username": "partial_user",
						"email":              "",
						"full_name":          "Partial UserMapToStrucsTest",
					},
					{
						"Instagram username": "",
						"email":              "email@test.com",
						"full_name":          "",
					},
				},
			},
			expected: []UserMapToStrucsTest{
				{
					InstagramUsername: "partial_user",
					Email:             "",
					Name:              "Partial UserMapToStrucsTest",
				},
				{
					InstagramUsername: "",
					Email:             "email@test.com",
					Name:              "",
				},
			},
			expectError: false,
		},
		{
			name: "More CSV columns than struct fields",
			input: CSVData{
				Headers: []string{"Instagram username", "email", "full_name", "extra_column", "another_extra"},
				Records: []CSVRecord{
					{
						"Instagram username": "user1",
						"email":              "user1@test.com",
						"full_name":          "UserMapToStrucsTest One",
						"extra_column":       "extra_value",
						"another_extra":      "more_extra",
					},
				},
			},
			expected: []UserMapToStrucsTest{
				{
					InstagramUsername: "user1",
					Email:             "user1@test.com",
					Name:              "UserMapToStrucsTest One",
					// Extra columns are ignored
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CSVDataToStructs[UserMapToStrucsTest](&tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("MapToStructsGeneric() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("MapToStructsGeneric() unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("MapToStructsGeneric() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestCSVDataMapToStructsGeneric_SimpleUser(t *testing.T) {
	tests := []struct {
		name     string
		input    CSVData
		expected []SimpleUserMapToStructsTest
	}{
		{
			name: "Two field struct",
			input: CSVData{
				Headers: []string{"username", "email"},
				Records: []CSVRecord{
					{
						"username": "john_doe",
						"email":    "john@example.com",
					},
					{
						"username": "jane_doe",
						"email":    "jane@example.com",
					},
				},
			},
			expected: []SimpleUserMapToStructsTest{
				{
					Username: "john_doe",
					Email:    "john@example.com",
				},
				{
					Username: "jane_doe",
					Email:    "jane@example.com",
				},
			},
		},
		{
			name: "Fewer CSV columns than struct fields",
			input: CSVData{
				Headers: []string{"username"},
				Records: []CSVRecord{
					{
						"username": "incomplete_user",
					},
				},
			},
			expected: []SimpleUserMapToStructsTest{
				{
					Username: "incomplete_user",
					Email:    "", // Second field remains empty
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CSVDataToStructs[SimpleUserMapToStructsTest](&tt.input)
			if err != nil {
				t.Errorf("MapToStructsGeneric() unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("MapToStructsGeneric() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestCSVDataMapToStructsGeneric_SingleField(t *testing.T) {
	tests := []struct {
		name     string
		input    CSVData
		expected []SingleFieldTest
	}{
		{
			name: "Single field struct",
			input: CSVData{
				Headers: []string{"value"},
				Records: []CSVRecord{
					{"value": "first"},
					{"value": "second"},
					{"value": "third"},
				},
			},
			expected: []SingleFieldTest{
				{Value: "first"},
				{Value: "second"},
				{Value: "third"},
			},
		},
		{
			name: "Single field with empty values",
			input: CSVData{
				Headers: []string{"value"},
				Records: []CSVRecord{
					{"value": "not_empty"},
					{"value": ""},
					{"value": "also_not_empty"},
				},
			},
			expected: []SingleFieldTest{
				{Value: "not_empty"},
				{Value: ""},
				{Value: "also_not_empty"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CSVDataToStructs[SingleFieldTest](&tt.input)
			if err != nil {
				t.Errorf("MapToStructsGeneric() unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("MapToStructsGeneric() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestCSVDataMapToStructsGeneric_EdgeCases(t *testing.T) {
	t.Run("Empty headers", func(t *testing.T) {
		input := CSVData{
			Headers: []string{},
			Records: []CSVRecord{{}},
		}

		result, err := CSVDataToStructs[UserMapToStrucsTest](&input)
		if err != nil {
			t.Errorf("MapToStructsGeneric() unexpected error: %v", err)
			return
		}

		expected := []UserMapToStrucsTest{{}}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("MapToStructsGeneric() = %+v, want %+v", result, expected)
		}
	})

	t.Run("Nil CSVData", func(t *testing.T) {
		var csvData *CSVData = nil

		// This should panic or be handled gracefully
		defer func() {
			if r := recover(); r != nil {
				// Expected behavior for nil pointer
			}
		}()

		_, err := CSVDataToStructs[UserMapToStrucsTest](csvData)
		if err == nil {
			t.Errorf("MapToStructsGeneric() expected error for nil input")
		}
	})
}
