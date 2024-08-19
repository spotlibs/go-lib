package stdresp

import "github.com/brispot/go-lib/stderr"

// StdOpt option signature that accept and modify Std response object.
type StdOpt func(std *Std)

// WithDesc embed given string data to the standard response as field `responseDesc`.
func WithDesc(desc string) StdOpt {
	return func(s *Std) {
		s.ResponseDesc = desc
	}
}

// WithData embed given data object to the standard response as field `responseData`.
func WithData(data any) StdOpt {
	return func(s *Std) {
		s.ResponseData = data
	}
}

// WithErr embed given error to the standard response.
func WithErr(e error) StdOpt {
	return func(s *Std) {
		// set default response code and description
		s.ResponseCode = stderr.ERROR_CODE_SYSTEM
		s.ResponseDesc = stderr.ERROR_DESC_SYSTEM

		// check if the error is created using stderr pkg
		if stderr.IsStdError(e) {
			// override the code and description
			s.ResponseCode = stderr.GetCode(e)
			s.ResponseDesc = stderr.GetMsg(e)

			// also if any, get the validation message too
			s.ResponseValidation = stderr.GetValidationErrorMsg(e)

			// get the http code
			s.httpCode = stderr.GetHttpCode(e)
		}
	}
}
