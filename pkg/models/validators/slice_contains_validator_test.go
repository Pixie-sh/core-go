package validators

import (
	"reflect"
	"testing"
)

// Mock FieldLevel for testing purposes
type mockFieldLevel struct {
	field reflect.Value
	param string
}

func (m *mockFieldLevel) GetStructFieldOK() (reflect.Value, reflect.Kind, bool) {
	// TODO implement me
	panic("implement me")
}

func (m *mockFieldLevel) GetStructFieldOKAdvanced(val reflect.Value, namespace string) (reflect.Value, reflect.Kind, bool) {
	// TODO implement me
	panic("implement me")
}

func (m *mockFieldLevel) GetStructFieldOK2() (reflect.Value, reflect.Kind, bool, bool) {
	// TODO implement me
	panic("implement me")
}

func (m *mockFieldLevel) GetStructFieldOKAdvanced2(val reflect.Value, namespace string) (reflect.Value, reflect.Kind, bool, bool) {
	// TODO implement me
	panic("implement me")
}

func (m *mockFieldLevel) FieldName() string {

	// TODO implement me
	panic("implement me")
}

func (m *mockFieldLevel) StructFieldName() string {
	// TODO implement me
	panic("implement me")
}

func (m *mockFieldLevel) ExtractType(field reflect.Value) (value reflect.Value, kind reflect.Kind, nullable bool) {
	// TODO implement me
	panic("implement me")
}

func (m *mockFieldLevel) Field() reflect.Value {
	return m.field
}

func (m *mockFieldLevel) Param() string {
	return m.param
}

func (m *mockFieldLevel) GetTag() string {
	return ""
}

func (m *mockFieldLevel) Parent() reflect.Value {
	return reflect.Value{}
}

func (m *mockFieldLevel) Top() reflect.Value {
	return reflect.Value{}
}

func TestValidateSliceContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    interface{}
		param    string
		expected bool
	}{
		{"IntSliceContains", []int{1, 2, 3, 4}, "3|2", true},
		{"IntSliceNotContains", []int{1, 2, 3, 4}, "5|6", false},
		{"BoolSliceContains", []bool{true, false, true}, "true|false", true},
		{"BoolSliceNotContains", []bool{true, true, true}, "false", false},
		{"StringSliceContains", []string{"apple", "banana", "cherry"}, "banana|apple", true},
		{"StringSliceNotContains", []string{"apple", "banana", "cherry"}, "date", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fl := mockFieldLevel{
				field: reflect.ValueOf(test.slice),
				param: test.param,
			}
			result := SliceContainsRuleValidator.ValidatorFn(&fl)
			if result != test.expected {
				t.Errorf("Failed %s: expected %v, got %v", test.name, test.expected, result)
			}
		})
	}
}
