package types

import (
	goJson "encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/pixie-sh/di-go"
)

func InstanceOf[T any](i any) bool {
	if Nil(i) {
		return false
	}

	_, ok := i.(T)
	return ok
}

func IsPointer(i interface{}) bool {
	if i == nil {
		return false
	}

	return reflect.TypeOf(i).Kind() == reflect.Ptr
}

// ToSnakeCase converts a CamelCase string to snake_case.
func ToSnakeCase(str string) string {
	var result strings.Builder
	for i, r := range str {
		if unicode.IsUpper(r) {
			if i > 0 && (unicode.IsLower(rune(str[i-1])) || (i+1 < len(str) && unicode.IsLower(rune(str[i+1])))) {
				result.WriteRune('_')
			}
			r = unicode.ToLower(r)
		}
		result.WriteRune(r)
	}
	return result.String()
}

// NameOf returns the canonical name of the type T without package path
// and converts it to snake case.
func NameOf[T any](t T) string {
	typ := reflect.TypeOf(t)
	if typ == nil {
		return "nil"
	}

	typeName := typ.String()
	if idx := strings.LastIndex(typeName, "."); idx != -1 {
		typeName = typeName[idx+1:]
	}

	return ToSnakeCase(typeName)
}

func ParseUint64(s string) uint64 {
	val, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		panic(err)
	}

	return val
}

func ParseUint64WithError(s string) (uint64, error) {
	val, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse '%s' as uint64: %w", s, err)
	}

	return val, nil
}

func ParseFloat64(s string) (float64, error) {
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse '%s' as float64: %w", s, err)
	}

	return val, nil
}

func IsNumber(s string) bool {
	if _, err := strconv.Atoi(s); err == nil {
		return true
	}

	if _, err := strconv.ParseUint(s, 10, 64); err == nil {
		return true
	}

	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return true
	}

	return false
}

func Nil(i interface{}) bool {
	if i == nil {
		return true
	}

	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Chan, reflect.Slice, reflect.Func, reflect.Interface:
		return reflect.ValueOf(i).IsNil()
	case reflect.Array:
		// Arrays cannot be nil in Go, they are value types
		return false
	case reflect.Struct:
		return IsEmpty(i)
	default:
		//logger.Logger.Debug("reflection type %s not supported yet for nil check", reflect.TypeOf(i).Kind())
	}

	return false
}

// IsEmpty checks if the given struct of type T has zero values for all its fields.
func IsEmpty[T any](i T) bool {
	return reflect.DeepEqual(i, reflect.Zero(reflect.TypeOf(i)).Interface())
}

// IsJSON check if provided payload is a json
func IsJSON(blob []byte) bool {
	if len(blob) == 0 {
		return false
	}

	return goJson.Valid(blob)
}

// SafeCast attempts to safely cast an interface{} value to type T.
// It uses type assertion to convert the input value to the desired type.
// Parameters:
//   - i: interface{} value to be cast
//   - T: generic type parameter specifying the target type
//
// Returns:
//   - T: the cast value of type T if successful
//   - bool: true if the cast was successful, false otherwise
func SafeCast[T any](i interface{}) (T, bool) {
	return di.SafeTypeAssert[T](i)
}

type Numeric interface {
	int64 | int32 | uint64 | uint32 | int | uint | float32 | float64
}

func NumericPointerCast[T, R Numeric](pointer *T) *R {
	var resultPointer *R
	if pointer != nil {
		pointerValue := R(*pointer)
		resultPointer = &pointerValue
	}

	return resultPointer
}
