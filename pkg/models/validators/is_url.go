package validators

import (
	"reflect"
	"regexp"
)

var urlRegexp = regexp.MustCompile(`^(https?|ftp)://[^\s/$.?#].[^\s]*$`)
var IsURL = ValidatorPair{
	ValidatorTag: "is_url",
	ValidatorFn: func(fl Field) bool {
		field := fl.Field()
		if !field.IsValid() || !field.CanInterface() || field.Kind() != reflect.String {
			return false
		}

		url, ok := field.Interface().(string)
		if !ok {
			return false
		}

		return urlRegexp.MatchString(url)
	},
}
