package validators

import (
	"reflect"

	"github.com/go-playground/validator/v10"
)

var NotRepeated = ValidatorPair{
	ValidatorTag: "not_repeated",
	ValidatorFn: func(fl validator.FieldLevel) bool {
		field := fl.Field()

		if field.Kind() != reflect.Slice {
			return false
		}

		seen := make(map[interface{}]bool)
		fieldName := fl.Param()

		for i := 0; i < field.Len(); i++ {
			item := field.Index(i)
			fieldValue := item.FieldByName(fieldName)
			if !fieldValue.IsValid() {
				return false
			}

			value := fieldValue.Interface()
			if seen[value] {
				return false
			}

			seen[value] = true
		}

		return true
	},
}
