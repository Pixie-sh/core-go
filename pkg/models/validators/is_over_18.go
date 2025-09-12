package validators

import (
	"reflect"
	"time"
)

func IsOver18Validator() ValidatorPair {
	return ValidatorPair{
		ValidatorTag: "isOver18",
		ValidatorFn: func(fl Field) bool {
			field := fl.Field()

			if !field.IsValid() || !field.CanInterface() {
				return false
			}

			// We expect the value to be a time.Time value.
			if field.Kind() != reflect.TypeOf(time.Time{}).Kind() {
				return false
			}

			value, ok := field.Interface().(time.Time)
			if !ok {
				return false
			}

			// This is where we make sure the date is at least 18 years old
			if (time.Now().Year() - value.Year()) < 18 {
				return false
			}

			return true
		},
	}
}
