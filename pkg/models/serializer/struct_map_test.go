package serializer

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

type metadataWithTime struct {
	BlockID                        string    `json:"block_id" validate:"required"`
	BlockVersion                   string    `json:"block_version" validate:"required"`
	BlockLastStatusDate            time.Time `json:"block_last_status_date" validate:"required"`
	BlockLastStatusChangedByUserID uint64    `json:"block_last_status_changed_by_user_id" validate:"required,numeric"`
}

type structWithPointers struct {
	RequiredField string  `json:"required_field" validate:"required"`
	OptionalStr   *string `json:"optional_str,omitempty"`
	OptionalInt   *int    `json:"optional_int,omitempty"`
}

func TestValidateMetadataWithTime(t *testing.T) {
	metadata := metadataWithTime{
		BlockID:                        "123",
		BlockVersion:                   "v1",
		BlockLastStatusDate:            time.Now(),
		BlockLastStatusChangedByUserID: 1,
	}

	metadataPayload, err := ToJSONB(metadata, true)
	if err != nil {
		t.Errorf("Validation failed for valid input: %v", err)
	}

	fmt.Println("result ", metadataPayload)

	var metadata2 metadataWithTime
	err = FromJSONB(metadataPayload, &metadata2, true)
	if err != nil {
		t.Errorf("Mapping failed: %v", err)
	}
	fmt.Println("results ", metadata, metadata2)

	if metadata.BlockLastStatusDate.UTC().Format(time.RFC3339) != metadata2.BlockLastStatusDate.UTC().Format(time.RFC3339) {
		t.Errorf("Expected metadata and metadata2 to be equal, but they are not")
	}
}

func TestNilPointerFields(t *testing.T) {
	testStr := structWithPointers{
		RequiredField: "test",
		OptionalStr:   nil,
		OptionalInt:   nil,
	}

	result, err := ToJSONB(testStr, false)
	if err != nil {
		t.Errorf("Failed to convert struct: %v", err)
	}

	if _, exists := result["optional_str"]; exists {
		t.Error("Nil pointer OptionalStr should not appear in result map")
	}

	if _, exists := result["optional_int"]; exists {
		t.Error("Nil pointer OptionalInt should not appear in result map")
	}

	if val, exists := result["required_field"]; !exists || val != "test" {
		t.Error("Required field should exist in result map with correct value")
	}
}

func TestDataFromAny(t *testing.T) {

	// Define a test struct
	type TestStruct struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	// Test cases
	tests := []struct {
		name            string
		input           any
		withValidations bool
		expectedOutput  TestStruct
		expectError     bool
	}{
		{
			name: "Valid map conversion",
			input: map[string]interface{}{
				"id":   1,
				"name": "test",
			},
			withValidations: false,
			expectedOutput: TestStruct{
				ID:   1,
				Name: "test",
			},
			expectError: false,
		},
		{
			name:            "Invalid input type",
			input:           "invalid input",
			withValidations: false,
			expectedOutput:  TestStruct{},
			expectError:     true,
		},
		{
			name: "With validations enabled",
			input: map[string]interface{}{
				"id":   1,
				"name": "test",
			},
			withValidations: true,
			expectedOutput: TestStruct{
				ID:   1,
				Name: "test",
			},
			expectError: false,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FromAny[TestStruct](tt.input, tt.withValidations)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// If no error is expected, compare the result
			if !tt.expectError {
				if !reflect.DeepEqual(result, tt.expectedOutput) {
					t.Errorf("Expected %+v but got %+v", tt.expectedOutput, result)
				}
			}
		})
	}
}
