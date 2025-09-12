package validators

import (
	"reflect"
	"strings"
)

// IfField TODO: to be finished, subsequent validations are failing
func IfField() ValidatorPair {
	return ValidatorPair{
		ValidatorTag: "if_field",
		ValidatorFn: func(fl Field) bool {
			field := fl.Field()

			if !field.IsValid() || !field.CanInterface() {
				return false
			}

			values := strings.Split(fl.Param(), ":")
			if len(values) != 2 {
				panic("invalid if_field configuration format; only two parts are allowed")
			}
			fieldName := values[0]
			validations := strings.Split(values[1], "->")
			if len(validations) != 2 {
				panic("invalid if_field configuration format; there must be two parts for subsequent validations")
			}

			// Retrieve the parent struct
			parent := fl.Top()
			parentValue := reflect.ValueOf(parent.Interface())

			// Get the field to be checked
			fieldToCheck := parentValue.FieldByName(fieldName)

			if !fieldToCheck.IsValid() || !fieldToCheck.CanInterface() {
				return false
			}

			// Check if the field matches the specified value
			if fieldToCheck.Interface() == validations[0] {
				tags := strings.Split(validations[1], ",")

				// Apply each validation tag using the validator instance
				for _, t := range tags {
					// Skip the custom if_field tag itself
					if strings.HasPrefix(t, "if_field") {
						continue
					}

					err := V.Var(field.Interface(), t)
					if err != nil {
						return false
					}
				}
			}

			return fieldToCheck.Interface() == validations[0]
		},
	}
}
