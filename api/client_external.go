package api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/goravel/framework/facades"
	"github.com/spotlibs/go-lib/ctx"
)

// NewHTTPClientExternal return HTTPClient implementer that also set some metadata header
// before sending the request.
func NewHTTPClientExternal() HTTPClient {
	var trans http.Transport
	trans.MaxConnsPerHost = 50
	trans.MaxIdleConnsPerHost = 15
	trans.MaxIdleConns = 50
	trans.IdleConnTimeout = 10 * time.Second
	trans.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	var client http.Client
	client.Transport = &trans
	client.Timeout = 30 * time.Second

	return &httpClientExternal{cl: &client}
}

type httpClientExternal struct {
	cl *http.Client
}

type MapRoute struct {
	Flag    bool   `json:"flag"`
	MockURL string `json:"mock_url"`
}

func (h *httpClientExternal) Call(req *http.Request, timeouts ...time.Duration) (HTTPResponse, error) {
	ctx.SetHTTPRequestHeader(req)
	reqTimeout := DEFAULT_TIMEOUT
	if len(timeouts) > 0 {
		reqTimeout = timeouts[0]
	}

	// Apply Mock
	reqUrl := req.URL.Scheme + "://" + req.URL.Host + req.URL.Path
	mapRoute, err := h.checkMock(reqUrl)
	if err == nil && mapRoute.Flag {
		mockURL, err := url.Parse(mapRoute.MockURL)
		if err == nil {
			req.URL = mockURL
			req.Host = mockURL.Host
		}
	}

	ctxWithTimeout, cancel := context.WithTimeout(req.Context(), reqTimeout)
	defer cancel()
	req = req.WithContext(ctxWithTimeout)

	var resp response
	res, err := h.cl.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connection error on HTTP request: %w", err)
	}
	defer res.Body.Close()

	resp.body, _ = io.ReadAll(res.Body)
	resp.statusCode = res.StatusCode
	resp.header = make(map[string][]string)
	resp.header = res.Header

	return &resp, nil
}

func (h *httpClientExternal) checkMock(url string) (*MapRoute, error) {
	if facades.Config().GetString("APP_ENV") == "production" {
		return nil, errors.New("cannot use mock in production environment")
	}

	key := "eksternal_mock_url_mapping:" + url
	mapRouteData := facades.Cache().GetString(key, "")

	if mapRouteData == "" {
		return &MapRoute{}, nil
	}

	var mapRoute MapRoute
	err := json.Unmarshal([]byte(mapRouteData), &mapRoute)
	if err != nil {
		return nil, err
	}

	return &mapRoute, nil
}
