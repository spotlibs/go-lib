package api

import (
	"net/http"
	"time"

	"github.com/bytedance/sonic"
)

// DEFAULT_TIMEOUT the default timeout of team specification.
var DEFAULT_TIMEOUT = 10 * time.Second

// Response object that may be returned by Agent.
type Response struct {
	StatusCode int
	Body       []byte
	Header     map[string][]string
}

// ToObject transform Response.Body to any object using json encoding.
func ToObject[T any](obj Response) (T, error) {
	var t T
	return t, sonic.ConfigFastest.Unmarshal(obj.Body, &t)
}

type HTTPAgent interface {
	// Do send given request using http, and optionally set custom timeout if
	// provided, otherwise will use DEFAULT_TIMEOUT.
	//
	// This function also help setting any necessary metadata for spotlibs using
	// ctx pkg that also come from this lib.
	Do(req *http.Request, timeouts ...time.Duration) (Response, error)
}
