package stderr

import "github.com/brispot/go-lib/debug"

// errWithDebug internal helper to construct err with setting stack trace level
// to 3.
func errWithDebug(code string, msg string, httpCode int, metadata ...string) error {
	return err{code: code, msg: msg, httpCode: httpCode, metadata: metadata, stackTrc: debug.GetStackTraceOnDebug(3)}
}
