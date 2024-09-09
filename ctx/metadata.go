package ctx

import (
	"context"

	"github.com/goravel/framework/contracts/http"
)

// keyMetadata custom type that can prevent collision.
type keyMetadata int

const keyMetadataCtx keyMetadata = iota

// metadataKey key to identify that a value in context is set and get from this
// package.
const metadataKey = "spotlibs-metadata-key"

// Metadata holds any request-scoped shared data within brispot microservice.
type Metadata struct {
	Authorization   string
	UserAgent       string
	CacheControl    string
	ForwardedFor    string
	RequestFrom     string
	DeviceId        string
	App             string
	VersionApp      string
	ReqId           string
	ReqTags         string
	ReqUser         string
	ReqNama         string
	ReqKodeJabatan  string
	ReqNamaJabatan  string
	ReqKodeUker     string
	ReqNamaUker     string
	ReqJenisUker    string
	ReqKodeMainUker string
	ReqKodeRegion   string
	PathGateway     string
	ApiKey          string
}

// Set inject given Metadata to context with custom key to make sure that the
// value is correct.
func Set(ctx context.Context, mt Metadata) context.Context {
	return context.WithValue(ctx, keyMetadataCtx, mt)
}

// Get retrieve Metadata from given context with key from this pkg.
func Get(ctx context.Context) Metadata {
	if mt, ok := ctx.Value(keyMetadataCtx).(Metadata); ok {
		return mt
	}
	return Metadata{}
}

// PassToContext pass Metadata from http.Context to context.
func PassToContext(c http.Context) context.Context {
	return Set(c, ParseRequest(c))
}

// ParseRequest return Metadata from given http context but return empty data
// instead if no data were found.
func ParseRequest(c http.Context) Metadata {
	mt, ok := c.Value(metadataKey).(Metadata)
	if !ok {
		mt = Metadata{}
	}
	return mt
}

// SetFromRequestHeader set any available metadata from given http context in
// the request header.
func SetFromRequestHeader(c http.Context) {
	mt := Metadata{
		Authorization:   c.Request().Header("Authorization"),
		UserAgent:       c.Request().Header("User-Agent"),
		CacheControl:    c.Request().Header("Cache-Control"),
		ApiKey:          c.Request().Header("X-Api-Key"),
		ForwardedFor:    c.Request().Header("X-Forwarded-For"),
		RequestFrom:     c.Request().Header("X-Request-From"),
		DeviceId:        c.Request().Header("X-Device-Id"),
		App:             c.Request().Header("X-App"),
		VersionApp:      c.Request().Header("X-Version-App"),
		ReqId:           c.Request().Header("X-Request-Id"),
		ReqTags:         c.Request().Header("X-Request-Tags"),
		ReqUser:         c.Request().Header("X-Request-User"),
		ReqNama:         c.Request().Header("X-Request-Nama"),
		ReqKodeJabatan:  c.Request().Header("X-Request-Kode-Jabatan"),
		ReqNamaJabatan:  c.Request().Header("X-Request-Nama-Jabatan"),
		ReqKodeUker:     c.Request().Header("X-Request-Kode-Uker"),
		ReqNamaUker:     c.Request().Header("X-Request-Nama-Uker"),
		ReqKodeMainUker: c.Request().Header("X-Request-Kode-MainUker"),
		ReqKodeRegion:   c.Request().Header("X-Request-Kode-Region"),
		ReqJenisUker:    c.Request().Header("X-Request-Jenis-Uker"),
		PathGateway:     c.Request().Header("X-Path-Gateway"),
	}
	c.WithValue(metadataKey, mt)
}

// GetReqId shortcut of ParseRequest with ReqId to get request id from given
// http context.
func GetReqId(c http.Context) string {
	return ParseRequest(c).ReqId
}
