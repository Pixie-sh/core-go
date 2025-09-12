package templates

import (
	"sync"

	"github.com/mailgun/raymond/v2"
)

var once sync.Once
var GlobalRegisteredHelpersWithExample []GlobalRegisterHelper

// GlobalRegisterHelper struct that hold a Helper definition
// runtime enforcement the AnonymousFunction !! must return a string !!
// if not a runtime error will be thrown
type GlobalRegisterHelper struct {
	FunctionName      string      `json:"function_name" field_description:"FunctionName of helper function"`
	UsageExample      string      `json:"usage_example" field_description:"Helper usage example"`
	AnonymousFunction interface{} `json:"-"`
}

// RegisterHelpers panic if the same handler is registered twice
// Function is only allowed to run Once (sync.Once)
// runtime enforcement the GlobalRegisterHelper.AnonymousFunction !! must return a string !!
// if not a runtime error will be thrown
func RegisterHelpers(helpers ...GlobalRegisterHelper) {
	once.Do(func() {
		for _, helper := range helpers {
			raymond.RegisterHelper(helper.FunctionName, helper.AnonymousFunction)
		}

		GlobalRegisteredHelpersWithExample = helpers
	})
}

// SetPartial protects global partials of duplicates
// stupid raymond :\
func SetPartial(key string, val string) {
	raymond.RemovePartial(key)
	raymond.RegisterPartial(key, val)
}
