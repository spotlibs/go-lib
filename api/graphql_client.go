package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"

	"github.com/goravel/framework/facades"
	"github.com/spotlibs/go-lib/log"
)

// GraphQLBody is the JSON body sent with every GraphQL request.
type GraphQLBody struct {
	Query         string         `json:"query"`
	Variables     map[string]any `json:"variables,omitempty"`
	OperationName *string        `json:"operationName,omitempty"`
}

// GraphQLSurroundingLog holds the surrounding-log payload for a GraphQL call.
type GraphQLSurroundingLog struct {
	Host         string                 `json:"host"`
	Url          string                 `json:"url"`
	GraphQL      GraphQLLogDetail       `json:"graphql"`
	Request      SurroundingLogRequest  `json:"request"`
	Response     SurroundingLogResponse `json:"response"`
	ResponseTime time.Duration          `json:"response_time"`
	MemoryUsage  uint64                 `json:"memory_usage"`
}

// GraphQLLogDetail captures the GQL-specific parts of the request for the log.
type GraphQLLogDetail struct {
	Query         string         `json:"query"`
	Variables     map[string]any `json:"variables,omitempty"`
	OperationName *string        `json:"operationName,omitempty"`
}

// GraphQLClient is a thin, 1:1 Go equivalent of the PHP GraphQLClient library.
// It always posts to a fixed endpoint and supports basic auth, bearer token,
// and arbitrary extra headers.
//
// Usage:
//
//	client := api.NewGraphQLClient()
//	client.SetEndpoint("https://api.example.com/graphql")
//	client.SetBearerToken("my-jwt")
//	resp, err := client.Query(ctx, `{ users { id } }`, nil, nil)
type GraphQLClient struct {
	cl            *http.Client
	endpoint      string
	headers       map[string]string
	basicAuthUser string
	basicAuthPass string
}

// NewGraphQLClient returns a new GraphQLClient with no endpoint set.
// Call SetEndpoint before issuing queries.
// Default timeout is 10 s, TLS verification is skipped (matching the PHP default).
func NewGraphQLClient() *GraphQLClient {
	var trans http.Transport
	trans.MaxConnsPerHost = 50
	trans.MaxIdleConnsPerHost = 15
	trans.MaxIdleConns = 50
	trans.IdleConnTimeout = 10 * time.Second
	trans.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true, //nolint:gosec // intentional, matches PHP verify=false default
	}

	var cl http.Client
	cl.Transport = &trans
	cl.Timeout = 10 * time.Second // PHP default timeout is 10 s

	return &GraphQLClient{
		cl: &cl,
		headers: map[string]string{
			"Content-Type": "application/json",
			"Accept":       "application/json",
		},
	}
}

// SetEndpoint configure endpoint
// Returns the receiver so calls can be chained.
func (g *GraphQLClient) SetEndpoint(endpoint string) *GraphQLClient {
	g.endpoint = endpoint
	return g
}

// SetBasicAuth configures HTTP Basic authentication for every request.
// Returns the receiver so calls can be chained.
func (g *GraphQLClient) SetBasicAuth(username, password string) *GraphQLClient {
	g.basicAuthUser = username
	g.basicAuthPass = password
	return g
}

// SetBearerToken sets the Authorization header to "Bearer <token>".
// Returns the receiver so calls can be chained.
func (g *GraphQLClient) SetBearerToken(token string) *GraphQLClient {
	g.headers["Authorization"] = "Bearer " + token
	return g
}

// SetHeaders merges the supplied headers into the client's default header set.
// Duplicate keys will be overwritten by the supplied values.
// Returns the receiver so calls can be chained.
func (g *GraphQLClient) SetHeaders(headers map[string]string) *GraphQLClient {
	for k, v := range headers {
		g.headers[k] = v
	}
	return g
}

// Query executes a GraphQL query/mutation against the configured endpoint.
// variables and operationName are optional (pass nil to omit them).
// The surrounding log is written when APP_DEBUG is "true".
func (g *GraphQLClient) Query(
	ctx context.Context,
	query string,
	variables map[string]any,
	operationName *string,
) (HTTPResponse, error) {
	startTime := time.Now()

	// Build request body
	body := GraphQLBody{Query: query}
	if variables != nil {
		body.Variables = variables
	}
	if operationName != nil {
		body.OperationName = operationName
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("graphql: failed to marshal request body: %w", err)
	}

	// Build HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("graphql: failed to create HTTP request: %w", err)
	}

	// Set headers
	for k, v := range g.headers {
		req.Header.Set(k, v)
	}

	// Set basic auth if configured (overrides bearer token if both are set)
	if g.basicAuthUser != "" || g.basicAuthPass != "" {
		encoded := base64.StdEncoding.EncodeToString(
			[]byte(g.basicAuthUser + ":" + g.basicAuthPass),
		)
		req.Header.Set("Authorization", "Basic "+encoded)
	}

	// Execute request
	res, err := g.cl.Do(req)
	if err != nil {
		return nil, fmt.Errorf("graphql: connection error: %w", err)
	}
	defer res.Body.Close()

	var resp response
	resp.body, _ = io.ReadAll(res.Body)
	resp.statusCode = res.StatusCode
	resp.header = make(map[string][]string)
	resp.header = res.Header

	// Surrounding log (mirrors PHP APP_DEBUG behaviour)
	elapsed := time.Since(startTime)
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	if facades.Config().GetString("APP_DEBUG", "false") == "true" {
		const msgTooLong = "more than 5000 characters"

		var bodyReqLog any
		if json.Unmarshal(jsonBody, &bodyReqLog) != nil {
			bodyReqLog = string(jsonBody)
		}
		var bodyRespLog any
		if json.Unmarshal(resp.body, &bodyRespLog) != nil {
			bodyRespLog = string(resp.body)
		}

		// Truncate if over 5000 characters
		if s, ok := bodyReqLog.(string); ok && len(s) > 5000 {
			bodyReqLog = msgTooLong
		} else if b, err2 := json.Marshal(bodyReqLog); err2 == nil && len(b) > 5000 {
			bodyReqLog = msgTooLong
		}
		if s, ok := bodyRespLog.(string); ok && len(s) > 5000 {
			bodyRespLog = msgTooLong
		} else if b, err2 := json.Marshal(bodyRespLog); err2 == nil && len(b) > 5000 {
			bodyRespLog = msgTooLong
		}

		logData := GraphQLSurroundingLog{
			Host: req.URL.Host,
			Url:  req.URL.Path,
			GraphQL: GraphQLLogDetail{
				Query:         query,
				Variables:     variables,
				OperationName: operationName,
			},
			Request: SurroundingLogRequest{
				Method: req.Method,
				Header: req.Header,
				Body:   bodyReqLog,
			},
			Response: SurroundingLogResponse{
				HttpCode: res.StatusCode,
				Header:   res.Header,
				Body:     bodyRespLog,
			},
			ResponseTime: elapsed,
			MemoryUsage:  m.Alloc,
		}

		log.Activity(ctx).Info(g.graphQLCallLog(logData))
	}

	return &resp, nil
}

func (g *GraphQLClient) graphQLCallLog(logData GraphQLSurroundingLog) log.Map {
	return log.Map{
		"host":         logData.Host,
		"url":          logData.Url,
		"graphql":      logData.GraphQL,
		"request":      logData.Request,
		"response":     logData.Response,
		"responseTime": logData.ResponseTime.Milliseconds(),
		"memoryUsage":  logData.MemoryUsage,
	}
}
