package middleware

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/spotlibs/go-lib/ctx"
)

// MetadataHeader set metadata information come from request header to current
// context.
func MetadataHeader(c http.Context) {
	ctx.SetFromRequestHeader(c)
	c.Request().Next()
}
