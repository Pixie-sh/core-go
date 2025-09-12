package validators

import (
	"reflect"
	"testing"
)

type FakeFieldLevel struct {
	fieldStr interface{}
	paramStr string
}

func (fs FakeFieldLevel) Top() reflect.Value {
	// TODO implement me
	panic("implement me")
}

func (fs FakeFieldLevel) Parent() reflect.Value {
	// TODO implement me
	panic("implement me")
}

func (fs FakeFieldLevel) FieldName() string {
	// TODO implement me
	panic("implement me")
}

func (fs FakeFieldLevel) StructFieldName() string {
	// TODO implement me
	panic("implement me")
}

func (fs FakeFieldLevel) GetTag() string {
	// TODO implement me
	panic("implement me")
}

func (fs FakeFieldLevel) ExtractType(field reflect.Value) (value reflect.Value, kind reflect.Kind, nullable bool) {
	// TODO implement me
	panic("implement me")
}

func (fs FakeFieldLevel) GetStructFieldOK() (reflect.Value, reflect.Kind, bool) {
	// TODO implement me
	panic("implement me")
}

func (fs FakeFieldLevel) GetStructFieldOKAdvanced(val reflect.Value, namespace string) (reflect.Value, reflect.Kind, bool) {
	// TODO implement me
	panic("implement me")
}

func (fs FakeFieldLevel) GetStructFieldOK2() (reflect.Value, reflect.Kind, bool, bool) {
	// TODO implement me
	panic("implement me")
}

func (fs FakeFieldLevel) GetStructFieldOKAdvanced2(val reflect.Value, namespace string) (reflect.Value, reflect.Kind, bool, bool) {
	// TODO implement me
	panic("implement me")
}

func (f FakeFieldLevel) Field() reflect.Value {
	return reflect.ValueOf(f.fieldStr)
}

func (f FakeFieldLevel) Param() string {
	return f.paramStr
}

func TestMapContainsRuleValidator_Validate(t *testing.T) {
	validator := MapContainsRuleValidator.ValidatorFn

	tests := []struct {
		name string
		fl   FakeFieldLevel
		want bool
	}{
		{
			name: "NonMapField",
			fl: FakeFieldLevel{
				fieldStr: "NotAMap",
				paramStr: "Key1|Key2",
			},
			want: false,
		},
		{
			name: "MapNoMatchingKey",
			fl: FakeFieldLevel{
				fieldStr: map[string]int{"OtherKey": 1, "AnotherKey": 2},
				paramStr: "Key1|Key2",
			},
			want: false,
		},
		{
			name: "MapMatchingKey",
			fl: FakeFieldLevel{
				fieldStr: map[string]int{"Key1": 1, "Key2": 1, "AnotherKey": 2},
				paramStr: "Key1|Key2",
			},
			want: true,
		},
		{
			name: "MapMatchingKey2",
			fl: FakeFieldLevel{
				fieldStr: map[string]int{"Key1": 1, "AnotherKey": 2},
				paramStr: "Key1",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validator(tt.fl); got != tt.want {
				t.Errorf("mapContainsRuleValidator.Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}
