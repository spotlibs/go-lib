package middleware

import (
	"fmt"

	"github.com/brispot/go-lib/debug"
	"github.com/brispot/go-lib/stderr"
	"github.com/brispot/go-lib/stdresp"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

// Recover do recover when panic occurring anywhere in the stack. Also give
// back standardized response by utilizing stderr & stdresp pkg.
func Recover(c http.Context) {
	defer func() {
		// grab any panic occurring
		if r := recover(); r != nil {
			prefixMsg := "Panic Runtime Error - "
			suffixMsg := "Terjadi kesalahan, silahkan hubungi IT Helpdesk" // use masked message as the default
			if facades.Config().GetBool("APP_DEBUG") {
				// replace with the debug info if its enabled
				suffixMsg = fmt.Sprint(r) + " - " + debug.GetStackTraceInString(1)
			}

			err := stderr.Err(stderr.ERROR_CODE_SYSTEM, prefixMsg+suffixMsg, http.StatusOK)
			_ = stdresp.Error(c, stdresp.WithErr(err)).Render()
		}
	}()

	c.Request().Next()
}
