package validators

import (
	"time"

	"github.com/go-playground/validator/v10"
)

var TodayOrBefore = ValidatorPair{
	ValidatorTag: "today_or_before",
	ValidatorFn: func(fl validator.FieldLevel) bool {
		// Handle both time.Time and *time.Time
		switch date := fl.Field().Interface().(type) {
		case time.Time:
			now := time.Now().Truncate(24 * time.Hour)
			return date.Before(now) || date.Equal(now)
		case *time.Time:
			if date == nil {
				return true // return true for nil, as empty values may be valid
			}
			now := time.Now().Truncate(24 * time.Hour)
			return date.Before(now) || date.Equal(now)
		default:
			// Return false for unsupported types
			return false
		}
	},
}
