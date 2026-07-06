package api

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"slices"
	"time"

	"github.com/goravel/framework/facades"
	"github.com/redis/go-redis/v9"
	"github.com/spotlibs/go-lib/ctx"
	"github.com/spotlibs/go-lib/databases"
	"github.com/spotlibs/go-lib/log"
)

// SOAPClient is the interface for making SOAP calls.
type SOAPClient interface {
	// Call sends a SOAP request to the given URL with the specified action
	// and parameters. Returns the raw content of <{Action}Result> element
	// (typically a JSON string for .NET services).
	//
	// The caller is responsible for unmarshaling the result bytes.
	Call(ctx context.Context, url string, action string, params map[string]string, timeouts ...time.Duration) ([]byte, error)
}

// SOAPOption is a functional option for configuring SOAPClient.
type SOAPOption func(*soapClient)

// WithSOAPVersion sets the SOAP protocol version. Default is SOAP11.
func WithSOAPVersion(v SOAPVersion) SOAPOption {
	return func(s *soapClient) {
		s.version = v
	}
}

// WithWSSecurity enables WS-Security with the given configuration.
func WithWSSecurity(ws WSSecurity) SOAPOption {
	return func(s *soapClient) {
		s.wsSecurity = &ws
	}
}

// WithTargetNamespace sets the target namespace for SOAP actions.
// Default is "http://tempuri.org/".
func WithTargetNamespace(ns string) SOAPOption {
	return func(s *soapClient) {
		s.targetNamespace = ns
	}
}

// NewSOAPClient creates a new SOAP client with the given options.
func NewSOAPClient(opts ...SOAPOption) SOAPClient {
	var trans http.Transport
	trans.MaxConnsPerHost = 50
	trans.MaxIdleConnsPerHost = 15
	trans.MaxIdleConns = 50
	trans.IdleConnTimeout = 10 * time.Second
	trans.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true, //nolint:gosec // internal services use self-signed certs
	}

	var client http.Client
	client.Transport = &trans
	client.Timeout = 30 * time.Second

	sc := &soapClient{
		cl:              &client,
		version:         SOAP11,
		targetNamespace: defaultTargetNS,
	}

	for _, opt := range opts {
		opt(sc)
	}

	return sc
}

type soapClient struct {
	cl              *http.Client
	version         SOAPVersion
	wsSecurity      *WSSecurity
	targetNamespace string
}

// SOAPSurroundingLog holds the data for surrounding log of SOAP calls.
type SOAPSurroundingLog struct {
	AppName      string                 `json:"app_name"`
	Path         string                 `json:"path"`
	Host         string                 `json:"host"`
	URL          string                 `json:"url"`
	SOAPAction   string                 `json:"soap_action"`
	SOAPVersion  string                 `json:"soap_version"`
	Request      SurroundingLogRequest  `json:"request"`
	Response     SurroundingLogResponse `json:"response"`
	ResponseTime time.Duration          `json:"response_time"`
	MemoryUsage  uint64                 `json:"memory_usage"`
}

func (s *soapClient) Call(requestCtx context.Context, soapURL string, action string, params map[string]string, timeouts ...time.Duration) ([]byte, error) {
	// Init
	startTime := time.Now()
	metadata := ctx.Get(requestCtx)

	// Build WS-Security header (if configured)
	var securityHeader []byte
	if s.wsSecurity != nil {
		var err error
		securityHeader, err = buildWSSecurityHeader(s.wsSecurity)
		if err != nil {
			return nil, fmt.Errorf("failed to build WS-Security header: %w", err)
		}
	}

	// Build SOAP envelope
	envelope, err := buildSOAPEnvelope(s.version, action, s.targetNamespace, params, securityHeader)
	if err != nil {
		return nil, fmt.Errorf("failed to build SOAP envelope: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, soapURL, bytes.NewReader(envelope))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers based on SOAP version
	s.setHeaders(req, action)

	// Store request body for logging
	var bodyDataLog interface{}
	bodyStr := string(envelope)
	bodyDataLog = bodyStr

	// Set timeout
	reqTimeout := DEFAULT_TIMEOUT
	if len(timeouts) > 0 {
		reqTimeout = timeouts[0]
	}

	// Apply mock
	reqURL := req.URL.Scheme + "://" + req.URL.Host + req.URL.Path
	mapRoute, mockErr := s.checkMock(reqURL)
	if mockErr == nil && mapRoute.Flag {
		mockURL, parseErr := url.Parse(mapRoute.MockURL)
		if parseErr == nil {
			req.URL = mockURL
			req.Host = mockURL.Host
		}
	}

	// Prepare request with timeout
	ctxWithTimeout, cancel := context.WithTimeout(req.Context(), reqTimeout)
	defer cancel()
	req = req.WithContext(ctxWithTimeout)

	// Execute HTTP call
	res, err := s.cl.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connection error on SOAP request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		if closeErr := Body.Close(); closeErr != nil {
			log.Runtime(ctxWithTimeout).Error(log.Map{
				"msg": fmt.Sprintf("error closing response body: %v", closeErr),
			})
		}
	}(res.Body)

	// Handle gzip
	var bodyReader io.Reader = res.Body
	if res.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, gzErr := gzip.NewReader(res.Body)
		if gzErr != nil {
			return nil, fmt.Errorf("gzip decompress error: %w", gzErr)
		}
		defer gzipReader.Close()
		bodyReader = gzipReader
	}

	// Read response
	respBody, _ := io.ReadAll(bodyReader)

	// Calculate elapsed time and memory usage
	elapsed := time.Since(startTime)
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Response body for logging
	var bodyResponseLog interface{}
	bodyResponseLog = string(respBody)

	// Log surrounding
	logData := SOAPSurroundingLog{
		AppName:     facades.Config().GetString("APP_NAME"),
		Path:        getSOAPIdentifierPath(metadata),
		Host:        req.URL.Host,
		URL:         req.URL.Path,
		SOAPAction:  action,
		SOAPVersion: s.versionString(),
		Request: SurroundingLogRequest{
			Method: req.Method,
			Header: req.Header,
			Body:   bodyDataLog,
		},
		Response: SurroundingLogResponse{
			HttpCode: res.StatusCode,
			Header:   res.Header,
			Body:     bodyResponseLog,
		},
		ResponseTime: elapsed,
		MemoryUsage:  m.Alloc,
	}

	const msgValidate = "more than 5000 characters"

	// Truncate request body log
	if facades.Config().GetString("CLIENT_DEBUG", "false") == "false" {
		if reqStr, ok := logData.Request.Body.(string); ok {
			if len(reqStr) > 5000 {
				logData.Request.Body = msgValidate
			}
		}
	}

	// Truncate response body log
	if facades.Config().GetString("CLIENT_DEBUG", "false") == "false" {
		if respStr, ok := logData.Response.Body.(string); ok {
			if len(respStr) > 5000 {
				logData.Response.Body = msgValidate
			}
		}
	}

	// Record surrounding log
	log.Activity(requestCtx).Info(s.soapSurroundingLog(logData))

	// Check for SOAP Fault
	if fault := parseSOAPFault(respBody, s.version); fault != nil {
		return nil, fault
	}

	// Extract result
	result, extractErr := extractSOAPResult(respBody, action)
	if extractErr != nil {
		return nil, fmt.Errorf("failed to extract SOAP result: %w", extractErr)
	}

	return result, nil
}

// setHeaders sets appropriate HTTP headers based on SOAP version.
func (s *soapClient) setHeaders(req *http.Request, action string) {
	switch s.version {
	case SOAP11:
		req.Header.Set("Content-Type", "text/xml; charset=utf-8")
		req.Header.Set("SOAPAction", fmt.Sprintf(`"%s%s"`, s.targetNamespace, action))
	case SOAP12:
		req.Header.Set("Content-Type", fmt.Sprintf(
			`application/soap+xml; charset=utf-8; action="%s%s"`,
			s.targetNamespace, action,
		))
	}

	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "go-soap-client/1.0")
	}
	if req.Header.Get("Accept-Encoding") == "" {
		req.Header.Set("Accept-Encoding", "gzip, deflate")
	}
}

// versionString returns a human-readable SOAP version string.
func (s *soapClient) versionString() string {
	if s.version == SOAP12 {
		return "1.2"
	}
	return "1.1"
}

// checkMock checks Redis for mock URL mapping (non-production only).
func (s *soapClient) checkMock(urlStr string) (*MapRoute, error) {
	disallowedEnv := []string{"production", "staging", "piloting"}
	if slices.Contains(disallowedEnv, facades.Config().GetString("APP_ENV")) {
		return nil, errors.New("cannot use mock in production environment")
	}

	redisClient := databases.GetRedisClient()

	key := "eksternal_mock_url_mapping:" + urlStr
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

// soapSurroundingLog converts SOAPSurroundingLog to log.Map.
func (s *soapClient) soapSurroundingLog(logData SOAPSurroundingLog) log.Map {
	return log.Map{
		"app_name":     logData.AppName,
		"path":         logData.Path,
		"host":         logData.Host,
		"url":          logData.URL,
		"soap_action":  logData.SOAPAction,
		"soap_version": logData.SOAPVersion,
		"request":      logData.Request,
		"response":     logData.Response,
		"responseTime": logData.ResponseTime.Milliseconds(),
		"memoryUsage":  logData.MemoryUsage,
	}
}

// getSOAPIdentifierPath extracts the path identifier from metadata.
func getSOAPIdentifierPath(metadata ctx.Metadata) string {
	if metadata.UrlPath != "" {
		return metadata.UrlPath
	}
	if metadata.SignaturePath != "" {
		return metadata.SignaturePath
	}
	return ""
}
