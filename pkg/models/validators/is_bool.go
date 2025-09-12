package validators

var IsBool = ValidatorPair{
	ValidatorTag: "isBool",
	ValidatorFn: func(fl Field) bool {
		_, ok := fl.Field().Interface().(bool)
		return ok
	},
}
