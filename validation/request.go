package validation

import (
	"encoding/json"

	httpValidate "github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/validation"
	"github.com/spotlibs/go-lib/stderr"
)

// ValidateRequest validate request data with given rules.
func ValidateRequest[T any](rules map[string]string, data map[string]any, obj *T, opts ...httpValidate.Option) error {
	// If no options provided, use default validationMessages
	if len(opts) == 0 {
		messagesCopy := make(map[string]string, len(validationMessages))
		for k, v := range validationMessages {
			messagesCopy[k] = v
		}
		opts = append(opts, validation.Messages(messagesCopy))
	}

	val, err := facades.Validation().Make(data, rules, opts...)
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

	by, _ := json.Marshal(data)
	_ = json.Unmarshal(by, obj)

	return nil
}
