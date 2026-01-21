package validators

import "fmt"

func IsValidVATValidator(country string, predicate func(string) bool) ValidatorPair {
	return ValidatorPair{
		ValidatorTag: fmt.Sprintf("vat:%s", country),
		ValidatorFn: func(fl Field) bool {
			field := fl.Field()

			if !field.IsValid() || !field.CanInterface() || field.IsZero() {
				return true
			}

			value, ok := field.Interface().(string)
			if !ok {
				return false
			}

			return predicate(value)
		},
	}
}
