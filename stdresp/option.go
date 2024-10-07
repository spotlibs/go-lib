package stdresp

import (
	"strings"

	"github.com/spotlibs/go-lib/debug"
	"github.com/spotlibs/go-lib/stderr"
)

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
		s.ResponseDesc = "Terjadi kesalahan, mohon coba beberapa saat lagi yaa... "

		// check if the error is created using stderr pkg
		if stderr.IsStdError(e) {
			// override the code and description
			s.ResponseCode = stderr.GetCode(e)

			// do trim space in case stacktrace is empty
			s.ResponseDesc = strings.TrimSpace(stderr.GetMsg(e) + " " + stderr.GetStackTrace(e))

			// also if any, get the validation message too
			s.ResponseValidation = stderr.GetValidationErrorMsg(e)

			// get the http code
			s.httpCode = stderr.GetHttpCode(e)
			return
		}

		// capture in case its random error that's not constructed with stderr pkg
		//  but only print if the debug flag is on
		if debug.IsOn() {
			s.ResponseDesc = e.Error()
		}
	}
}
