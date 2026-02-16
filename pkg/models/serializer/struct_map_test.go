package serializer

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/pixie-sh/core-go/pkg/types"
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

type metadataWithPointerTime struct {
	ID        string     `json:"id"`
	CreatedAt *time.Time `json:"created_at"`
}

type metadataWithNonPointerTime struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
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

func TestStructToMapValueTimeField(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	metadata := metadataWithTime{
		BlockID:                        "123",
		BlockVersion:                   "v1",
		BlockLastStatusDate:            now,
		BlockLastStatusChangedByUserID: 1,
	}

	result, err := StructToMap(metadata)
	if err != nil {
		t.Fatalf("StructToMap failed: %v", err)
	}

	raw, exists := result["block_last_status_date"]
	if !exists {
		t.Fatal("block_last_status_date key missing from result map")
	}

	expected := now.Format(time.RFC3339)

	switch m := raw.(type) {
	case map[string]string:
		if m["RFC3339"] != expected {
			t.Errorf("expected RFC3339=%q, got %q", expected, m["RFC3339"])
		}
	case map[string]interface{}:
		val, _ := m["RFC3339"].(string)
		if val != expected {
			t.Errorf("expected RFC3339=%q, got %q", expected, val)
		}
	default:
		t.Fatalf("expected map with RFC3339 key for time field, got %T: %v", raw, raw)
	}
}

func TestRoundTripPointerTimeField(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	original := metadataWithPointerTime{
		ID:        "abc",
		CreatedAt: &now,
	}

	payload, err := StructToMap[map[string]interface{}](original)
	if err != nil {
		t.Fatalf("StructToMap failed: %v", err)
	}

	raw, exists := payload["created_at"]
	if !exists {
		t.Fatal("created_at key missing from result map")
	}

	// Verify the serialized form has the RFC3339 key
	switch m := raw.(type) {
	case map[string]string:
		if m["RFC3339"] == "" {
			t.Fatal("RFC3339 key empty in serialized time map")
		}
	case map[string]interface{}:
		if m["RFC3339"] == nil {
			t.Fatal("RFC3339 key missing in serialized time map")
		}
	default:
		t.Fatalf("expected map with RFC3339 key for time field, got %T: %v", raw, raw)
	}

	var restored metadataWithNonPointerTime
	err = ToStruct(payload, &restored)
	if err != nil {
		t.Fatalf("ToStruct failed: %v", err)
	}

	if types.IsEmpty(restored.CreatedAt) {
		t.Fatal("restored CreatedAt is nil")
	}

	if restored.CreatedAt.UTC().Format(time.RFC3339) != now.Format(time.RFC3339) {
		t.Errorf("expected %q, got %q", now.Format(time.RFC3339), restored.CreatedAt.UTC().Format(time.RFC3339))
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
