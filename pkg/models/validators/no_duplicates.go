package validators

import (
	"reflect"
)

var NoDuplicates = ValidatorPair{
	ValidatorTag: "no_duplicates",
	ValidatorFn: func(fl Field) bool {
		// Get the field value
		field := fl.Field()

		// Ensure the field is valid and can be interfaced
		if !field.IsValid() || !field.CanInterface() {
			return false
		}

		// Check if the field is a slice
		if field.Kind() != reflect.Slice {
			return false
		}

		// Create a map to track seen values
		seen := make(map[interface{}]bool)

		// Iterate over the slice
		for i := 0; i < field.Len(); i++ {
			item := field.Index(i).Interface()
			// Check for duplicate values
			if seen[item] {
				return false
			}
			seen[item] = true
		}

		return true
	},
}
