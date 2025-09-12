package validators

import (
	"reflect"
	"strconv"
	"strings"
)

const SliceContainsRuleValidatorTag = "slice_contains"

var SliceContainsRuleValidator = ValidatorPair{
	ValidatorTag: SliceContainsRuleValidatorTag,
	ValidatorFn: func(fl Field) bool {
		field := fl.Field()
		if field.Kind() != reflect.Slice {
			return false
		}

		// Parameter that contains the keys to be found
		keys := strings.Split(fl.Param(), "|")
		found := make([]bool, len(keys))

		// Iterate over the slice and check for each element
		for i := 0; i < field.Len(); i++ {
			value := field.Index(i)

			// Convert value to a comparable interface
			var comp interface{}
			switch value.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				comp = value.Int()
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				comp = value.Uint()
			case reflect.Bool:
				comp = value.Bool()
			case reflect.String:
				comp = value.String()
			default:
				// Unhandled types
				continue
			}

			// Check if the current value is one of the desired keys
			for j, key := range keys {
				// Attempt conversion of key to the same type as slice element for comparison
				var keyComp interface{}
				switch value.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					keyComp, _ = strconv.ParseInt(key, 10, 64)
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					keyComp, _ = strconv.ParseUint(key, 10, 64)
				case reflect.Bool:
					keyComp, _ = strconv.ParseBool(key)
				case reflect.String:
					keyComp = key
				}

				if comp == keyComp {
					found[j] = true
				}
			}
		}

		// Check if all keys were found in the slice
		for _, f := range found {
			if !f {
				return false
			}
		}

		return true
	},
}
