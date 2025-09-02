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
	"runtime"
	"time"

	"github.com/goravel/framework/facades"
	"github.com/redis/go-redis/v9"
	"github.com/spotlibs/go-lib/ctx"
	"github.com/spotlibs/go-lib/databases"
	"github.com/spotlibs/go-lib/log"
)

// NewHTTPClientExternal return HTTPClient implementer that also set some metadata header
// before sending the request.
func NewHTTPClientExternal() HTTPClientExternal {
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

type ClientExternalSurroundingLog struct {
	AppName      string                 `json:"app_name"`
	Path         string                 `json:"path"`
	Host         string                 `json:"host"`
	Url          string                 `json:"url"`
	Request      SurroundingLogRequest  `json:"request"`
	Response     SurroundingLogResponse `json:"response"`
	ResponseTime time.Duration          `json:"response_time"`
	MemoryUsage  uint64                 `json:"memory_usage"`
}

type SurroundingLogRequest struct {
	Method string `json:"method"`
	Header any    `json:"header"`
}

type SurroundingLogResponse struct {
	HttpCode int `json:"http_code"`
	Header   any `json:"header"`
}

func (h *httpClientExternal) Call(requestCtx context.Context, req *http.Request, timeouts ...time.Duration) (HTTPResponse, error) {
	// Init
	startTime := time.Now()
	metadata := ctx.Get(req.Context())

	// Set Timeout
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

	// Prepare Request
	ctxWithTimeout, cancel := context.WithTimeout(req.Context(), reqTimeout)
	defer cancel()
	req = req.WithContext(ctxWithTimeout)

	// Push Default Header
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "go-http-client/1.0")
	}
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/json")
	}
	if req.Header.Get("Accept-Encoding") == "" {
		req.Header.Set("Accept-Encoding", "gzip, deflate")
	}

	// Call HTTP
	var resp response
	res, err := h.cl.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connection error on HTTP request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			log.Runtime(ctxWithTimeout).Error(map[string]any{
				"msg": fmt.Sprintf("error closing response body: %v", err),
			})
		}
	}(res.Body)

	// Responses
	resp.body, _ = io.ReadAll(res.Body)
	resp.statusCode = res.StatusCode
	resp.header = make(map[string][]string)
	resp.header = res.Header

	// Calculate elapsed time and memory usage
	elapsed := time.Since(startTime)
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Populate ClientExternalSurroundingLog
	logData := ClientExternalSurroundingLog{
		AppName: facades.Config().GetString("APP_NAME"),
		Path:    getIdentifierPath(metadata),
		Host:    req.URL.Host,
		Url:     req.URL.Path,
		Request: SurroundingLogRequest{
			Method: req.Method,
			Header: req.Header,
		},
		Response: SurroundingLogResponse{
			HttpCode: res.StatusCode,
			Header:   res.Header,
		},
		ResponseTime: elapsed,
		MemoryUsage:  m.Alloc,
	}

	// Record Surrounding Log
	log.Activity(requestCtx).Info(h.externalCallLog(logData))

	return &resp, nil
}

func (h *httpClientExternal) checkMock(url string) (*MapRoute, error) {
	if facades.Config().GetString("APP_ENV") == "production" {
		return nil, errors.New("cannot use mock in production environment")
	}

	redisClient := databases.GetRedisClient()

	key := "eksternal_mock_url_mapping:" + url
	mapRouteData, err := redisClient.Get(context.Background(), key).Result()
	if errors.Is(err, redis.Nil) {
		return &MapRoute{}, nil
	}
	if err != nil {
		return nil, err
	}

	var mapRoute MapRoute
	err = json.Unmarshal([]byte(mapRouteData), &mapRoute)
	if err != nil {
		return nil, err
	}

	return &mapRoute, nil
}

func getIdentifierPath(metadata ctx.Metadata) string {
	if metadata.UrlPath != "" {
		return metadata.UrlPath
	}
	if metadata.SignaturePath != "" {
		return metadata.SignaturePath
	}
	return ""
}

func (h *httpClientExternal) externalCallLog(logData ClientExternalSurroundingLog) log.Map {
	return log.Map{
		"app_name":      logData.AppName,
		"path":          logData.Path,
		"host":          logData.Host,
		"url":           logData.Url,
		"request":       logData.Request,
		"response":      logData.Response,
		"response_time": logData.ResponseTime.Milliseconds(),
		"memory_usage":  logData.MemoryUsage,
	}
}
