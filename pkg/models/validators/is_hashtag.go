package validators

import (
	"reflect"
	"regexp"
)

var hashTagRegex = regexp.MustCompile(`^#[A-Za-z0-9_]+$`) // Starts with #, followed by letters, numbers, or underscores
var IsHashtag = ValidatorPair{
	ValidatorTag: "is_hashtag",
	ValidatorFn: func(fl Field) bool {
		// Validate the string field
		field := fl.Field()
		if !field.IsValid() || !field.CanInterface() || field.Kind() != reflect.String {
			return false
		}

		tag, ok := field.Interface().(string)
		if !ok {
			return false
		}

		return hashTagRegex.MatchString(tag)
	},
}
