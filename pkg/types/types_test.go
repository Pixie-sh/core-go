package types

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestType struct{}

type OTPService struct{}
type EmailOTPService struct{}

func TestPayloadTypeOf(t *testing.T) {
	intType := PayloadTypeOf[int]()
	expectedIntType := "int"
	assert.Equal(t, expectedIntType, intType.String(), "PayloadTypeOf[int] should return correct PayloadType")

	testType := PayloadTypeOf[TestType]()
	expectedTestType := "test_type"
	assert.Equal(t, expectedTestType, testType.String(), "PayloadTypeOf[TestType] should return correct PayloadType")

	testPtrType := PayloadTypeOf[*TestType]()
	expectedTestPtrType := "test_type"
	assert.Equal(t, expectedTestPtrType, testPtrType.String(), "PayloadTypeOf[*TestType] should return correct PayloadType")

	stringType := PayloadTypeOf[string]()
	expectedStringType := "string"
	assert.Equal(t, expectedStringType, stringType.String(), "PayloadTypeOf[string] should return correct PayloadType")

	sliceType := PayloadTypeOf[[]int]()
	expectedSliceType := "[]int"
	assert.Equal(t, expectedSliceType, sliceType.String(), "PayloadTypeOf[[]int] should return correct PayloadType")

	testPtrType = PayloadTypeOf[OTPService]()
	expectedTestPtrType = "otp_service"
	assert.Equal(t, expectedTestPtrType, testPtrType.String(), "PayloadTypeOf[OTPService] should return correct PayloadType")

	testPtrType = PayloadTypeOf[EmailOTPService]()
	expectedTestPtrType = "email_otp_service"
	assert.Equal(t, expectedTestPtrType, testPtrType.String(), "PayloadTypeOf[EmailOTPService] should return correct PayloadType")
}

type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}

func TestInstanceOf(t *testing.T) {

	// Test with a string
	if !InstanceOf[string]("hello") {
		t.Error("Expected 'hello' to be instance of string")
	}

	// Test with an int
	if !InstanceOf[int](42) {
		t.Error("Expected 42 to be instance of int")
	}

	// Test with a custom error type
	myErr := &testError{message: "test error"}
	if !InstanceOf[*testError](myErr) {
		t.Error("Expected myErr to be instance of *testError")
	}

	// Test with error interface
	if !InstanceOf[error](myErr) {
		t.Error("Expected myErr to be instance of error interface")
	}

	// Test with incorrect type
	if InstanceOf[float64](42) {
		t.Error("Expected 42 (int) not to be instance of float64")
	}

	// Test with nil
	var nilErr *testError = nil
	if InstanceOf[*testError](nilErr) {
		t.Error("Expected nil not to be instance of *testError")
	}
}

func TestStringToByteSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []byte
	}{
		{
			name:     "Basic string",
			input:    "Hello, World!",
			expected: []byte("Hello, World!"),
		},
		{
			name:     "Empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "Unicode string",
			input:    "こんにちは",
			expected: []byte("こんにちは"),
		},
		{
			name:     "String with spaces",
			input:    "  spaces  ",
			expected: []byte("  spaces  "),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UnsafeBytes(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("UnsafeBytes(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestStringToByteSliceImmutability(t *testing.T) {
	input := "Hello, World!"
	byteSlice := UnsafeBytes(input)

	// Check if the original string is modified
	expectedString := "Hello, World!" // The original string should remain unchanged
	if UnsafeString(byteSlice) != expectedString {
		t.Errorf("Original string was modified. Got %q, want %q", input, expectedString)
	}
}

func TestIntPtr(t *testing.T) {
	input := 1
	inputPtr := &input
	assert.Equal(t, "1", fmt.Sprintf("%d", *inputPtr))
}

func TestUint64IsEmpty(t *testing.T) {
	assert.True(t, IsEmpty(uint64(0)))

	var a uint64
	assert.True(t, IsEmpty(a))
}
