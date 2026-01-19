package debug

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/spotlibs/go-lib/helper"
)

// GetStackTraceOnDebug return information from GetStackTraceInString if the
// internal state of debug is enabled. It can be enabled with EnableDebug.
//
// May be used when need to get the stack trace only if the debug is enabled,
// since calling runtime.Caller is expensive.
func GetStackTraceOnDebug(pick ...int) string {
	return GetStackTraceInString(pick...)
}

// GetStackTraceInString return the stack trace utilizing runtime.Caller but
// only pick the first line that has `/app/`.
//
// Will always return the stack trace even when DisableDebug already called.
// May be used when just want to get the stack trace without caring the debug
// state.
func GetStackTraceInString(pick ...int) string {
	stack := make([]uintptr, 2<<6)
	length := runtime.Callers(0, stack)

	var pickAll bool
	if len(pick) < 1 {
		pickAll = true
	}

	trackPicked := 1
	var traces []string
	seen := helper.NewStringSet()
	for i := 0; i < length; i++ {
		funcPtr := runtime.FuncForPC(stack[i])
		file, line := funcPtr.FileLine(stack[i])
		if strings.Contains(file, "/app/") {
			s := fmt.Sprintf("[%s:%d]", file, line)
			if !seen.Exists(s) {
				seen.Add(s)
				if !pickAll && trackPicked == pick[0] {
					return s
				}
				traces = append(traces, s)
				trackPicked++
			}
		}
	}

	return strings.Join(traces, " ")
}
