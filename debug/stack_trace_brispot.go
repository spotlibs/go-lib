package debug

import (
	"fmt"
	"runtime"
	"strings"
)

// GetStackTraceOnDebug return information from GetStackTraceInString if the
// internal state of debug is enabled. It can be enabled with EnableDebug.
//
// May be used when need to get the stack trace only if the debug is enabled,
// since calling runtime.Caller is expensive.
func GetStackTraceOnDebug(pick ...int) string {
	if isDebug.Load() {
		return GetStackTraceInString(pick...)
	}
	return ""
}

// GetStackTraceInString return the stack trace utilizing runtime.Caller but
// only pick the first line that has `/app/`.
//
// Will always return the stack trace even when DisableDebug already called.
// May be used when just want to get the stack trace without caring the debug
// state.
func GetStackTraceInString(pick ...int) string {
	stack := make([]uintptr, 2<<6)      // 128
	length := runtime.Callers(3, stack) // skip the first 3 frames

	var pickAll bool
	// set default to capture the first found line
	if len(pick) < 1 {
		pickAll = true
	}

	trackPicked := 1
	var allStackTrace strings.Builder
	for i := 0; i < length; i++ {
		funcPtr := runtime.FuncForPC(stack[i])
		file, line := funcPtr.FileLine(stack[i])
		if strings.Contains(file, "/app/") {

			s := fmt.Sprintf("%s:%d\n", file, line)
			// capture the matched pick
			if !pickAll && trackPicked == pick[0] {
				return s
			}

			allStackTrace.WriteString(s)

			trackPicked++

		}
	}

	return allStackTrace.String()
}
