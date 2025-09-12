package validators

import "github.com/go-playground/validator/v10"

var NotEqualValidator = ValidatorPair{
	ValidatorTag: "not_equal",
	ValidatorFn: func(fl validator.FieldLevel) bool {
		field := fl.Field()
		param := fl.Param()

		compareField := fl.Parent().FieldByName(param)
		if !compareField.IsValid() {
			return false
		}

		return field.Interface() != compareField.Interface()
	},
}
