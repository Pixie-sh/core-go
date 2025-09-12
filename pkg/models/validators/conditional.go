package validators

import (
	"fmt"
	"reflect"
	"strings"
)

func Conditional() ValidatorPair {
	return ValidatorPair{
		ValidatorTag: "conditional",
		ValidatorFn: func(fl Field) bool {
			// Example tag: "conditional=Role:admin->email"
			tag := fl.Param()

			parts := strings.Split(tag, "->")
			if len(parts) < 2 {
				// Invalid format, should have a condition and a validation
				return false
			}

			// Parse condition (e.g., "Role:admin")
			condition := parts[0]
			conditionParts := strings.Split(condition, ":")
			if len(conditionParts) != 2 {
				// Invalid condition format
				return false
			}

			otherFieldName := conditionParts[0]
			requiredValue := conditionParts[1]

			// Parse the actual validation to apply (e.g., "email")
			validation := strings.Join(parts[1:], ",")

			// Get the value of the other field
			structValue := fl.Top()
			if structValue.Kind() == reflect.Ptr {
				if structValue.IsNil() {
					// If the pointer is nil, you can't dereference it
					return false
				}
				structValue = structValue.Elem() // Dereference the pointer
			}

			otherField := structValue.FieldByName(otherFieldName)
			if !otherField.IsValid() {
				// Other field does not exist
				return false
			}

			// Check if the condition is met
			if fmt.Sprintf("%v", otherField.Interface()) == requiredValue {
				// If the condition is met, apply the specified validation
				// Temporarily create a new Validator to validate this field
				err := V.Var(fl.Field().Interface(), validation)
				if err != nil {
					return false
				}
			}
			// If the condition is not met, the field passes regardless
			return true
		},
	}
}
