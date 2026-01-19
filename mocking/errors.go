package mocking

import "errors"

const (
	ErrExpectedNetwork = "called error expected network"
	ErrExpectedInfra   = "called error expected infra"
)

func ProduceSampleError(errs ...error) error {
	var errMsg string
	if len(errs) > 0 {
		for _, err := range errs {
			errMsg += err.Error() + "\n"
		}
		return errors.New(errMsg)
	}
	return errors.New(ErrExpectedNetwork)
}
