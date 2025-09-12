package validators

import (
	"reflect"
	"time"

	"github.com/go-playground/validator/v10"
)

var TodayOrLater = ValidatorPair{
	ValidatorTag: "today_or_later",
	ValidatorFn: func(fl validator.FieldLevel) bool {
		field := fl.Field()

		// Handle nil pointers
		if field.Kind() == reflect.Ptr && field.IsNil() {
			return true // nil is considered valid
		}

		var date time.Time

		// Dereference pointer fields and handle time.Time directly
		switch field.Kind() {
		case reflect.Ptr:
			date = field.Elem().Interface().(time.Time) // Dereference pointer
		case reflect.Struct:
			date = field.Interface().(time.Time) // Use directly if it's a struct
		default:
			return false // Invalid type
		}

		// Get the start of today
		now := time.Now().Truncate(24 * time.Hour)
		return !date.Before(now)
	},
}
