package validation

import (
	"github.com/brispot/go-lib/stderr"
	"github.com/bytedance/sonic"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/validation"
)

// ValidateRequest validate request data with given rules.
func ValidateRequest[T any](rules map[string]string, data map[string]any, obj *T) error {
	val, err := facades.Validation().Make(data, rules, validation.Messages(validationMessages))
	if err != nil {
		return stderr.ErrRuntime(err.Error())
	}

	// return validation error if fail including what's fail
	if val.Fails() {
		var errorMessages []string
		for _, errs := range val.Errors().All() {
			for _, msg := range errs {
				errorMessages = append(errorMessages, msg)
			}
		}
		return stderr.ErrValidation(val.Errors().One(), errorMessages)
	}

	by, _ := sonic.ConfigFastest.Marshal(data)
	_ = sonic.ConfigFastest.Unmarshal(by, obj)

	return nil
}
