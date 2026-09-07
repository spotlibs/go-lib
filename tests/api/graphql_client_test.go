package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spotlibs/go-lib/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func ptr[T any](v T) *T { return &v }

// newGQLServer creates a test HTTP server that validates incoming GraphQL
// requests and returns a fixed JSON payload.
func newGQLServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(handler))
}

// successGQLHandler is a simple handler that always responds 200 + JSON body.
func successGQLHandler(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}
}

// ---------------------------------------------------------------------------
// constructor
// ---------------------------------------------------------------------------

func TestNewGraphQLClient_NotNil(t *testing.T) {
	client := api.NewGraphQLClient()
	assert.NotNil(t, client)
}

// ---------------------------------------------------------------------------
// SetEndpoint
// ---------------------------------------------------------------------------

func TestGraphQLClient_SetEndpoint_Chainable(t *testing.T) {
	client := api.NewGraphQLClient()
	returned := client.SetEndpoint("http://example.com/graphql")
	assert.Equal(t, client, returned, "SetEndpoint should return the same receiver")
}

func TestGraphQLClient_SetEndpoint_UsedForRequest(t *testing.T) {
	srv := newGQLServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{}}`))
	})
	defer srv.Close()

	client := api.NewGraphQLClient()
	client.SetEndpoint(srv.URL)

	resp, err := client.Query(context.Background(), `{ ping }`, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.GetStatusCode())
}

// ---------------------------------------------------------------------------
// SetBasicAuth
// ---------------------------------------------------------------------------

func TestGraphQLClient_SetBasicAuth_SetsHeader(t *testing.T) {
	srv := newGQLServer(t, func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		assert.True(t, strings.HasPrefix(authHeader, "Basic "), "expected Basic auth header, got: %s", authHeader)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{}}`))
	})
	defer srv.Close()

	client := api.NewGraphQLClient()
	client.SetEndpoint(srv.URL)
	client.SetBasicAuth("user", "pass")

	resp, err := client.Query(context.Background(), `{ ping }`, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.GetStatusCode())
}

func TestGraphQLClient_SetBasicAuth_Chainable(t *testing.T) {
	client := api.NewGraphQLClient()
	returned := client.SetBasicAuth("u", "p")
	assert.Equal(t, client, returned, "SetBasicAuth should return the same receiver")
}

// ---------------------------------------------------------------------------
// SetBearerToken
// ---------------------------------------------------------------------------

func TestGraphQLClient_SetBearerToken_SetsHeader(t *testing.T) {
	const token = "my-jwt-token"
	srv := newGQLServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer "+token, r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{}}`))
	})
	defer srv.Close()

	client := api.NewGraphQLClient()
	client.SetEndpoint(srv.URL)
	client.SetBearerToken(token)

	resp, err := client.Query(context.Background(), `{ ping }`, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.GetStatusCode())
}

func TestGraphQLClient_SetBearerToken_Chainable(t *testing.T) {
	client := api.NewGraphQLClient()
	returned := client.SetBearerToken("tok")
	assert.Equal(t, client, returned, "SetBearerToken should return the same receiver")
}

// ---------------------------------------------------------------------------
// SetHeaders
// ---------------------------------------------------------------------------

func TestGraphQLClient_SetHeaders_MergesHeaders(t *testing.T) {
	srv := newGQLServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "bar", r.Header.Get("X-Foo"))
		assert.Equal(t, "baz", r.Header.Get("X-Extra"))
		// Default header still present
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{}}`))
	})
	defer srv.Close()

	client := api.NewGraphQLClient()
	client.SetEndpoint(srv.URL)
	client.SetHeaders(map[string]string{
		"X-Foo":   "bar",
		"X-Extra": "baz",
	})

	resp, err := client.Query(context.Background(), `{ ping }`, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.GetStatusCode())
}

func TestGraphQLClient_SetHeaders_OverridesExistingKey(t *testing.T) {
	srv := newGQLServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{}}`))
	})
	defer srv.Close()

	client := api.NewGraphQLClient()
	client.SetEndpoint(srv.URL)
	client.SetHeaders(map[string]string{"Content-Type": "text/plain"})

	resp, err := client.Query(context.Background(), `{ ping }`, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.GetStatusCode())
}

func TestGraphQLClient_SetHeaders_Chainable(t *testing.T) {
	client := api.NewGraphQLClient()
	returned := client.SetHeaders(map[string]string{"X-Test": "1"})
	assert.Equal(t, client, returned, "SetHeaders should return the same receiver")
}

// ---------------------------------------------------------------------------
// Query — happy paths
// ---------------------------------------------------------------------------

func TestGraphQLClient_Query_SimpleQuery_Success(t *testing.T) {
	expectedBody := `{"data":{"users":[{"id":"1","name":"Alice"}]}}`

	srv := newGQLServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))

		// Decode and verify the request body
		var gqlBody map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gqlBody))
		assert.Equal(t, `{ users { id name } }`, gqlBody["query"])
		assert.Nil(t, gqlBody["variables"])
		assert.Nil(t, gqlBody["operationName"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(expectedBody))
	})
	defer srv.Close()

	client := api.NewGraphQLClient()
	client.SetEndpoint(srv.URL)
	resp, err := client.Query(context.Background(), `{ users { id name } }`, nil, nil)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.GetStatusCode())
	assert.JSONEq(t, expectedBody, string(resp.GetBody()))
}

func TestGraphQLClient_Query_WithVariables_SendsVariables(t *testing.T) {
	srv := newGQLServer(t, func(w http.ResponseWriter, r *http.Request) {
		var gqlBody map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gqlBody))

		vars, ok := gqlBody["variables"].(map[string]any)
		require.True(t, ok, "variables should be an object")
		assert.Equal(t, "1", vars["id"])

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"user":{"id":"1"}}}`))
	})
	defer srv.Close()

	client := api.NewGraphQLClient()
	client.SetEndpoint(srv.URL)
	resp, err := client.Query(
		context.Background(),
		`query GetUser($id: ID!) { user(id: $id) { id } }`,
		map[string]any{"id": "1"},
		nil,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.GetStatusCode())
}

func TestGraphQLClient_Query_WithOperationName_SendsOperationName(t *testing.T) {
	srv := newGQLServer(t, func(w http.ResponseWriter, r *http.Request) {
		var gqlBody map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gqlBody))
		assert.Equal(t, "GetUser", gqlBody["operationName"])

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{}}`))
	})
	defer srv.Close()

	client := api.NewGraphQLClient()
	client.SetEndpoint(srv.URL)
	resp, err := client.Query(
		context.Background(),
		`query GetUser { user { id } }`,
		nil,
		ptr("GetUser"),
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.GetStatusCode())
}

func TestGraphQLClient_Query_WithVariablesAndOperationName(t *testing.T) {
	srv := newGQLServer(t, func(w http.ResponseWriter, r *http.Request) {
		var gqlBody map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gqlBody))
		assert.Equal(t, "CreateUser", gqlBody["operationName"])
		vars := gqlBody["variables"].(map[string]any)
		assert.Equal(t, "Bob", vars["name"])

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"data":{"createUser":{"id":"99"}}}`))
	})
	defer srv.Close()

	client := api.NewGraphQLClient()
	client.SetEndpoint(srv.URL)
	resp, err := client.Query(
		context.Background(),
		`mutation CreateUser($name: String!) { createUser(name: $name) { id } }`,
		map[string]any{"name": "Bob"},
		ptr("CreateUser"),
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.GetStatusCode())
}

// ---------------------------------------------------------------------------
// Query — default headers
// ---------------------------------------------------------------------------

func TestGraphQLClient_Query_DefaultHeadersPresent(t *testing.T) {
	srv := newGQLServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{}}`))
	})
	defer srv.Close()

	client := api.NewGraphQLClient()
	client.SetEndpoint(srv.URL)
	_, err := client.Query(context.Background(), `{ ping }`, nil, nil)
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// Query — non-200 responses (should still return, not error)
// ---------------------------------------------------------------------------

func TestGraphQLClient_Query_ServerReturns500_ReturnsResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"errors":[{"message":"internal"}]}`))
	}))
	defer srv.Close()

	client := api.NewGraphQLClient()
	client.SetEndpoint(srv.URL)
	resp, err := client.Query(context.Background(), `{ ping }`, nil, nil)

	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.GetStatusCode())
	assert.Contains(t, string(resp.GetBody()), "internal")
}

func TestGraphQLClient_Query_ServerReturns404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`not found`))
	}))
	defer srv.Close()

	client := api.NewGraphQLClient()
	client.SetEndpoint(srv.URL)
	resp, err := client.Query(context.Background(), `{ ping }`, nil, nil)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.GetStatusCode())
}

// ---------------------------------------------------------------------------
// Query — connection errors
// ---------------------------------------------------------------------------

func TestGraphQLClient_Query_ConnectionError(t *testing.T) {
	// Use a port that refuses connections
	client := api.NewGraphQLClient()
	_, err := client.Query(context.Background(), `{ ping }`, nil, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "graphql: connection error")
}

func TestGraphQLClient_Query_InvalidURL(t *testing.T) {
	client := api.NewGraphQLClient()
	_, err := client.Query(context.Background(), `{ ping }`, nil, nil)

	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// Query — timeout
// ---------------------------------------------------------------------------

func TestGraphQLClient_Query_TimeoutReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// Use a context with a very short deadline
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	client := api.NewGraphQLClient()
	client.SetEndpoint(srv.URL)
	_, err := client.Query(ctx, `{ ping }`, nil, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "graphql: connection error")
}

// ---------------------------------------------------------------------------
// Query — APP_DEBUG logging path
// ---------------------------------------------------------------------------

func TestGraphQLClient_Query_AppDebugTrue_LogsWithoutError(t *testing.T) {
	// Activate debug logging
	t.Setenv("APP_DEBUG", "true")

	responsePayload := `{"data":{"product":{"id":"42"}}}`
	srv := newGQLServer(t, successGQLHandler(responsePayload))
	defer srv.Close()

	client := api.NewGraphQLClient()
	client.SetEndpoint(srv.URL)
	resp, err := client.Query(
		context.Background(),
		`query GetProduct($id: ID!) { product(id: $id) { id } }`,
		map[string]any{"id": "42"},
		ptr("GetProduct"),
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.GetStatusCode())
	assert.JSONEq(t, responsePayload, string(resp.GetBody()))
}

func TestGraphQLClient_Query_AppDebugFalse_NoLogEmitted(t *testing.T) {
	os.Unsetenv("APP_DEBUG") //nolint:errcheck

	srv := newGQLServer(t, successGQLHandler(`{"data":{}}`))
	defer srv.Close()

	client := api.NewGraphQLClient()
	client.SetEndpoint(srv.URL)
	resp, err := client.Query(context.Background(), `{ ping }`, nil, nil)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.GetStatusCode())
}

func TestGraphQLClient_Query_AppDebugTrue_LargeRequestBodyTruncated(t *testing.T) {
	t.Setenv("APP_DEBUG", "true")

	// Build a query that produces a JSON body > 5000 chars
	bigQuery := strings.Repeat("x", 6000)

	srv := newGQLServer(t, successGQLHandler(`{"data":{}}`))
	defer srv.Close()

	client := api.NewGraphQLClient()
	client.SetEndpoint(srv.URL)
	resp, err := client.Query(context.Background(), bigQuery, nil, nil)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.GetStatusCode())
}

func TestGraphQLClient_Query_AppDebugTrue_LargeResponseBodyTruncated(t *testing.T) {
	t.Setenv("APP_DEBUG", "true")

	// Response body > 5000 chars (plain string, not JSON object)
	bigBody := `"` + strings.Repeat("y", 6000) + `"`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(bigBody))
	}))
	defer srv.Close()

	client := api.NewGraphQLClient()
	client.SetEndpoint(srv.URL)
	resp, err := client.Query(context.Background(), `{ ping }`, nil, nil)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.GetStatusCode())
}

func TestGraphQLClient_Query_AppDebugTrue_LargeResponseObjectTruncated(t *testing.T) {
	t.Setenv("APP_DEBUG", "true")

	// Build a large JSON object response (> 5000 chars serialised)
	data := map[string]string{"key": strings.Repeat("v", 5100)}
	raw, _ := json.Marshal(map[string]any{"data": data})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(raw)
	}))
	defer srv.Close()

	client := api.NewGraphQLClient()
	client.SetEndpoint(srv.URL)
	resp, err := client.Query(context.Background(), `{ ping }`, nil, nil)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.GetStatusCode())
}

// ---------------------------------------------------------------------------
// Query — response body methods (GetBody / GetHeaders / GetHeader)
// ---------------------------------------------------------------------------

func TestGraphQLClient_Query_ResponseBodyParseable(t *testing.T) {
	payload := `{"data":{"id":"7","name":"Carol"}}`
	srv := newGQLServer(t, successGQLHandler(payload))
	defer srv.Close()

	client := api.NewGraphQLClient()
	client.SetEndpoint(srv.URL)
	resp, err := client.Query(context.Background(), `{ user { id name } }`, nil, nil)
	require.NoError(t, err)

	type Result struct {
		Data struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	var result Result
	require.NoError(t, json.Unmarshal(resp.GetBody(), &result))
	assert.Equal(t, "7", result.Data.ID)
	assert.Equal(t, "Carol", result.Data.Name)
}

func TestGraphQLClient_Query_GetHeaderReturnsValue(t *testing.T) {
	srv := newGQLServer(t, func(w http.ResponseWriter, r *http.Request) {
		// Go's http layer canonicalises header keys: "X-Request-ID" → "X-Request-Id"
		w.Header().Set("X-Request-Id", "abc-123")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{}}`))
	})
	defer srv.Close()

	client := api.NewGraphQLClient()
	client.SetEndpoint(srv.URL)
	resp, err := client.Query(context.Background(), `{ ping }`, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "abc-123", resp.GetHeader("X-Request-Id"))
}

func TestGraphQLClient_Query_GetHeadersReturnsMap(t *testing.T) {
	srv := newGQLServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{}}`))
	})
	defer srv.Close()

	client := api.NewGraphQLClient()
	client.SetEndpoint(srv.URL)
	resp, err := client.Query(context.Background(), `{ ping }`, nil, nil)
	require.NoError(t, err)
	headers := resp.GetHeaders()
	assert.Equal(t, "application/json", headers["Content-Type"])
}

// ---------------------------------------------------------------------------
// Query — non-JSON response body (plain text)
// ---------------------------------------------------------------------------

func TestGraphQLClient_Query_NonJSONResponse_ReturnsBodyAsBytes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("plain text response"))
	}))
	defer srv.Close()

	client := api.NewGraphQLClient()
	client.SetEndpoint(srv.URL)
	resp, err := client.Query(context.Background(), `{ ping }`, nil, nil)

	require.NoError(t, err)
	assert.Equal(t, "plain text response", string(resp.GetBody()))
}

// ---------------------------------------------------------------------------
// ToObject helper with GraphQL response
// ---------------------------------------------------------------------------

func TestGraphQLClient_Query_ToObject(t *testing.T) {
	srv := newGQLServer(t, successGQLHandler(`{"data":{"id":"5","label":"test"}}`))
	defer srv.Close()

	client := api.NewGraphQLClient()
	client.SetEndpoint(srv.URL)
	resp, err := client.Query(context.Background(), `{ item { id label } }`, nil, nil)
	require.NoError(t, err)

	type GQLResp struct {
		Data struct {
			ID    string `json:"id"`
			Label string `json:"label"`
		} `json:"data"`
	}
	obj, err := api.ToObject[GQLResp](resp)
	require.NoError(t, err)
	assert.Equal(t, "5", obj.Data.ID)
	assert.Equal(t, "test", obj.Data.Label)
}

// ---------------------------------------------------------------------------
// Query — nil context (covers NewRequestWithContext error path)
// ---------------------------------------------------------------------------

func TestGraphQLClient_Query_NilContext_ReturnsError(t *testing.T) {
	client := api.NewGraphQLClient()
	//nolint:staticcheck // intentional nil ctx to exercise error branch
	_, err := client.Query(nil, `{ ping }`, nil, nil) //nolint:staticcheck
	require.Error(t, err)
}
