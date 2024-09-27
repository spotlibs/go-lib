package middleware

import (
	"slices"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/spotlibs/go-lib/ctx"
	"github.com/spotlibs/go-lib/log"
)

// ActivityMonitor capture and log all request/response.
func ActivityMonitor(c http.Context) {
	now := time.Now()
	c.Request().Next()
	apiActivityRecorder(c, now)
}

func apiActivityRecorder(c http.Context, start time.Time) {
	ct := c.Request().Header("Content-Type")

	// check the content type and capture the request body according to it
	var req any
	switch {
	case hasPrefix(ct, "application/json", "application/x-www-form-urlencoded"):
		req = captureRequestMap(c)
	case hasPrefix(ct, "multipart/form-data"):
		req = captureRequestMultipart(c)
	default:
		req = captureRequestMap(c) // treat any unhandled content-type as map
	}

	// transform back response to an object before capturing
	var res map[string]any
	_ = sonic.ConfigFastest.Unmarshal(c.Response().Origin().Body().Bytes(), &res)

	// get metadata from context
	mt := ctx.Get(c)

	activityData := map[string]any{
		"app_name":     facades.Config().GetString("app.name", "Microservice"),
		"host":         c.Request().Host(),
		"path":         c.Request().Path(),
		"client_ip":    c.Request().Header("X-Forwarded-For", c.Request().Ip()),
		"client_app":   mt.App,
		"path_alias":   mt.PathGateway,
		"requestID":    mt.ReqId,
		"requestFrom":  mt.RequestFrom,
		"requestUser":  mt.ReqUser,
		"deviceID":     mt.DeviceId,
		"requestTags":  mt.ReqTags,
		"requestBody":  req,
		"responseBody": res,
		"responseTime": time.Since(start).Milliseconds(),
		"httpCode":     c.Response().Origin().Status(),
		"requestAt":    start.Format(time.RFC3339Nano),
		//"memoryUsage":  // coming soon
	}

	log.Activity(c).Info(activityData)
}

// captureRequestMap capture request as map and transform it to json string.
func captureRequestMap(c http.Context) any {
	return c.Request().All()
}

// captureRequestMultipart capture request multipart data and only get
// the information of that file such as the filename, size and extension.
func captureRequestMultipart(c http.Context) any {
	reqOrg := c.Request().Origin()
	_ = reqOrg.ParseMultipartForm(2 << 9) // 1024

	var bagOfForm []map[string]any
	// grab any available form-value
	for k, v := range reqOrg.MultipartForm.Value {
		bagOfForm = append(bagOfForm, map[string]any{
			"field": k,
			"value": v,
		})
	}
	// grab any available files
	for field, header := range reqOrg.MultipartForm.File {
		for _, headerFile := range header {
			if headerFile != nil {
				bagOfForm = append(bagOfForm, map[string]any{
					"field":    field,
					"filename": headerFile.Filename,
					"size":     int(headerFile.Size),
				})
			}
		}
	}

	return bagOfForm
}

// hasPrefix return true if the given s has at least one of the given prefixes.
func hasPrefix(s string, prefix ...string) bool {
	return slices.ContainsFunc(prefix, func(pre string) bool {
		return strings.HasPrefix(s, pre)
	})
}

// sniffMIMEType return the mime-type from given FileHeader instance by using
// helper provided by Goravel.
func sniffMIMEType(f *multipart.FileHeader) string {
	fl, err := filesystem.NewFileFromRequest(f)
	if err != nil {
		return "ERR-" + err.Error()
	}
	mt, err := fl.MimeType()
	if err != nil {
		return "ERR-" + err.Error()
	}
	return mt
}
