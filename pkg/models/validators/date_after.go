package validators

import (
	"reflect"
	"time"
)

var DateAfter = ValidatorPair{
	ValidatorTag: "date_after",
	ValidatorFn: func(fl Field) bool {
		// Get the current field value (end time)
		field := fl.Field()
		if !field.IsValid() || !field.CanInterface() {
			return false
		}

		// Ensure the field is a *time.Time type
		endTime, ok := field.Interface().(*time.Time)
		if !ok || endTime == nil {
			return true // Ignore nil values, assuming they pass validation
		}

		// Get the name of the other field from the tag parameter
		otherFieldName := fl.Param()
		if otherFieldName == "" {
			panic("date_after: missing field name to compare")
		}

		// Get the parent struct and the other field value (start time)
		parent := fl.Top()
		parentValue := reflect.ValueOf(parent.Interface())
		otherField := parentValue.FieldByName(otherFieldName)

		if !otherField.IsValid() || !otherField.CanInterface() {
			return false
		}

		// Ensure the other field is a *time.Time type
		startTime, ok := otherField.Interface().(*time.Time)
		if !ok || startTime == nil {
			return true // If no start time is provided, we assume the end time is valid
		}

		// Validate: endTime must be after startTime
		return endTime.After(*startTime)
	},
}
