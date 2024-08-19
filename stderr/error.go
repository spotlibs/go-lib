package stderr

import (
	"errors"
	"net/http"
)

// err standard object that hold any information about the error.
type err struct {
	httpCode int
	code     string
	msg      string
	metadata any
}

// Error implement error interface.
func (e err) Error() string {
	return e.code + " " + e.msg
}

// Err set std error by given error code, message and any desired http code.
//
// Each value can be retrieved by helper func such as GetCode, GetMsg, GetMeta.
func Err(code string, msg string, httpCode int) error {
	return err{code: code, msg: msg, httpCode: httpCode}
}

// ErrValidation error in validation, it's not recommended to be used directly.
//
// Currently, this func only used by validation pkg.
func ErrValidation(msg string, validation []string) error {
	if msg == "" {
		msg = ERROR_DESC_PARAMETER
	}
	return err{code: ERROR_CODE_PARAMETER, msg: msg, metadata: validation, httpCode: http.StatusOK}
}

// ErrParam error in validation such as in the parameter input coming from
// request.
func ErrParam(msg ...string) error {
	if len(msg) < 1 {
		msg = append(msg, ERROR_DESC_PARAMETER)
	}
	return Err(ERROR_CODE_PARAMETER, msg[0], http.StatusOK)
}

// ErrNotFound error when the 'expected' data not found, this may exist in both
// usecase and repository layer, or even in the service layer.
func ErrNotFound(msg ...string) error {
	if len(msg) < 1 {
		msg = append(msg, ERROR_DESC_DATA_NOT_FOUND)
	}
	return Err(ERROR_CODE_DATA_NOT_FOUND, msg[0], http.StatusOK)
}

// ErrPermission error permission denied, this may commonly be used in brispot
// gateway, or any other middleware that need to return non-200 http status
// code.
//
// This error will set the response http code to 403 via stderr.WithErr.
func ErrPermission(msg ...string) error {
	if len(msg) < 1 {
		msg = append(msg, ERROR_DESC_PERMISSION)
	}
	return Err(ERROR_CODE_ACCESS_PERMISSION, msg[0], http.StatusForbidden)
}

// ErrInvHeader error invalid header, this may commonly be used in brispot
// gateway.
//
// This error will set the response http code to 400 via stderr.WithErr.
func ErrInvHeader(msg ...string) error {
	if len(msg) < 1 {
		msg = append(msg, ERROR_DESC_INVALID_HEADER)
	}
	return Err(ERROR_CODE_INVALID_HEADER, msg[0], http.StatusBadRequest)
}

// ErrInvRule any process that does not meet with the requirement set in the
// usecase layer.
func ErrInvRule(msg ...string) error {
	if len(msg) < 1 {
		msg = append(msg, ERROR_DESC_INVALID_RULE)
	}
	return Err(ERROR_CODE_INVALID_RULE, msg[0], http.StatusOK)
}

// ErrThirdParty error triggered by surrounding service. It's up to developer
// whether they should propagate the 'error' from surrounding to the response,
// or to masking it and mark it as 3rd-party error.
func ErrThirdParty(msg ...string) error {
	if len(msg) < 1 {
		msg = append(msg, ERROR_DESC_THIRD_PARTY)
	}
	return Err(ERROR_CODE_THIRD_PARTY, msg[0], http.StatusOK)
}

// ErrWaiting error when there is still process occurring in the background.
func ErrWaiting(msg ...string) error {
	if len(msg) < 1 {
		msg = append(msg, ERROR_DESC_WAITING_STATUS)
	}
	return Err(ERROR_CODE_WAITING_STATUS, msg[0], http.StatusOK)
}

// ErrUnsupported error when the incoming request can not be supported.
//
// e.g. when incoming request with method GET but the endpoint expect POST.
func ErrUnsupported(msg ...string) error {
	if len(msg) < 1 {
		msg = append(msg, ERROR_DESC_UNSUPPORTED)
	}
	return Err(ERROR_CODE_UNSUPPORTED, msg[0], http.StatusOK)
}

// ErrRuntime unexpected error that may occur in the runtime.
//
// e.g. error when doing json Marshal/Unmarshal.
func ErrRuntime(msg ...string) error {
	if len(msg) < 1 {
		msg = append(msg, ERROR_DESC_SYSTEM)
	}
	return Err(ERROR_CODE_SYSTEM, msg[0], http.StatusOK)
}

// ErrDataNotFound error ErrNotFound with message 'data tidak ditemukan'.
func ErrDataNotFound() error {
	return ErrNotFound("data tidak ditemukan")
}

// IsErrNotFound return true if the given error is created using ErrNotFound
// or ErrDataNotFound.
func IsErrNotFound(e error) bool {
	var stdErr err
	if errors.As(e, &stdErr) {
		return stdErr.code == ERROR_CODE_DATA_NOT_FOUND
	}
	return false
}

// IsStdError check whether given error is created using this stderr pkg.
func IsStdError(e error) bool {
	return errors.As(e, &err{})
}

// GetHttpCode get the http status code reside in the error. Will return the
// correct http code if the error coming from this stderr pkg, otherwise will
// return 500.
func GetHttpCode(e error) int {
	var stdErr err
	if errors.As(e, &stdErr) {
		return stdErr.httpCode
	}
	return http.StatusInternalServerError
}

// GetCode get the error code if the given error if it's error coming from this
// stderr pkg, otherwise will return the default ERROR_CODE_SYSTEM.
func GetCode(e error) string {
	var stdErr err
	if errors.As(e, &stdErr) {
		return stdErr.code
	}
	return ERROR_CODE_SYSTEM
}

// GetMsg get the error message if the given error if it's error coming
// from this pkg, otherwise will return the default ERROR_DESC_SYSTEM.
func GetMsg(e error) string {
	var stdErr err
	if errors.As(e, &stdErr) {
		return stdErr.msg
	}
	return ERROR_DESC_SYSTEM
}

// GetMeta get the error metadata if the given error is coming from this stderr
// pkg, otherwise will return nil.
func GetMeta(e error) any {
	var stdErr err
	if errors.As(e, &stdErr) {
		return stdErr.metadata
	}
	return nil
}

// GetValidationErrorMsg return the validation message injected in
// ErrValidation, otherwise will just return nil.
func GetValidationErrorMsg(e error) []string {
	msg, ok := GetMeta(e).([]string)
	if ok {
		return msg
	}
	return nil
}
