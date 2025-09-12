package validators

import "reflect"

var IsWeekday = ValidatorPair{
	ValidatorTag: "is_weekday",
	ValidatorFn: func(fl Field) bool {
		// Get the field value
		field := fl.Field()

		// Ensure the field is valid and can be interfaced
		if !field.IsValid() || !field.CanInterface() {
			return false
		}

		// Ensure the field is a slice
		if field.Kind() != reflect.Slice {
			return false
		}

		// Iterate over the slice and validate each element
		for i := 0; i < field.Len(); i++ {
			// Get the current element as an interface
			element := field.Index(i).Interface()

			// Convert the element to uint64 for validation
			var value uint64
			switch v := element.(type) {
			case uint64:
				value = v
			case uint32:
				value = uint64(v)
			case int:
				if v < 0 {
					return false // Negative integers are invalid
				}
				value = uint64(v)
			case int32:
				if v < 0 {
					return false
				}
				value = uint64(v)
			default:
				return false // Unsupported type
			}

			// Check if the value is within the valid range (0–6)
			if value < 0 || value > 6 {
				return false
			}
		}

		return true
	},
}
