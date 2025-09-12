package validators

import (
	"fmt"
)

func IsEnumValidator[T comparable](slugSufix string, validValues []T) ValidatorPair {
	return ValidatorPair{

		ValidatorTag: fmt.Sprintf("isEnum:%s", slugSufix),
		ValidatorFn: func(fl Field) bool {
			field := fl.Field()

			if !field.IsValid() || !field.CanInterface() || field.IsZero() {
				return true
			}

			value, ok := field.Interface().(T)
			if !ok {
				return false
			}

			for _, validValue := range validValues {
				if value == validValue {
					return true
				}
			}

			return false
		},
	}
}
