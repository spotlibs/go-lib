package api_test

import (
	"context"
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

func TestMain(m *testing.M) {
	// Set minimal required environment variables for tests
	os.Setenv("APP_NAME", "spotlib-go-test")
	os.Setenv("APP_ENV", "development")
	os.Setenv("APP_KEY", "JzLRl2YHe1Ec7MCGqkAJ4byaF08uKLfs")

	// Run tests
	code := m.Run()

	// Cleanup
	os.Unsetenv("APP_NAME")
	os.Unsetenv("APP_ENV")
	os.Unsetenv("APP_KEY")

	os.Exit(code)
}

func TestHTTPClientExternal_Call_Success(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}))
	defer server.Close()

	// Create client
	client := api.NewHTTPClientExternal()

	// Create request
	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)

	ctx := context.Background()

	// Make call
	resp, err := client.Call(ctx, req)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.GetStatusCode())
	assert.Contains(t, string(resp.GetBody()), "success")
	assert.Equal(t, "application/json", resp.GetHeader("Content-Type"))
}

func TestHTTPClientExternal_Call_WithTimeout(t *testing.T) {
	// Create slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "delayed"}`))
	}))
	defer server.Close()

	client := api.NewHTTPClientExternal()
	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)

	ctx := context.Background()

	// Test with custom timeout
	resp, err := client.Call(ctx, req, 200*time.Millisecond)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.GetStatusCode())
}

func TestHTTPClientExternal_Call_TimeoutError(t *testing.T) {
	// Create very slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := api.NewHTTPClientExternal()
	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)

	ctx := context.Background()

	// Test with very short timeout
	_, err = client.Call(ctx, req, 10*time.Millisecond)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection error on HTTP request")
}

func TestHTTPClientExternal_Call_PostRequest(t *testing.T) {
	// Create test server that expects POST
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": 123}`))
	}))
	defer server.Close()

	client := api.NewHTTPClientExternal()

	body := strings.NewReader(`{"name": "test"}`)
	req, err := http.NewRequest("POST", server.URL, body)
	require.NoError(t, err)

	ctx := context.Background()
	resp, err := client.Call(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.GetStatusCode())
	assert.Contains(t, string(resp.GetBody()), "123")
}

func TestHTTPClientExternal_Call_DefaultHeaders(t *testing.T) {
	// Create test server to check headers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check default headers are set
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "go-http-client/1.0", r.Header.Get("User-Agent"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "gzip, deflate", r.Header.Get("Accept-Encoding"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok": true}`))
	}))
	defer server.Close()

	client := api.NewHTTPClientExternal()
	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)

	ctx := context.Background()
	resp, err := client.Call(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.GetStatusCode())
}

func TestHTTPClientExternal_Call_CustomHeaders(t *testing.T) {
	// Create test server to check custom headers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Custom headers should override defaults
		assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))
		assert.Equal(t, "custom-agent", r.Header.Get("User-Agent"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`OK`))
	}))
	defer server.Close()

	client := api.NewHTTPClientExternal()
	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)

	// Set custom headers
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("User-Agent", "custom-agent")

	ctx := context.Background()
	resp, err := client.Call(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.GetStatusCode())
}

func TestResponse_Methods(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "value1")
		w.Header().Add("X-Custom-Header", "value2")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"test": true}`))
	}))
	defer server.Close()

	client := api.NewHTTPClientExternal()
	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)

	ctx := context.Background()
	resp, err := client.Call(ctx, req)
	require.NoError(t, err)

	// Test response methods
	assert.Equal(t, http.StatusCreated, resp.GetStatusCode())
	assert.Equal(t, `{"test": true}`, string(resp.GetBody()))
	assert.Equal(t, "application/json", resp.GetHeader("Content-Type"))
	assert.Equal(t, "value1,value2", resp.GetHeader("X-Custom-Header"))

	headers := resp.GetHeaders()
	assert.Equal(t, "application/json", headers["Content-Type"])
	assert.Equal(t, "value1,value2", headers["X-Custom-Header"])
}

func TestToObject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"name": "test", "id": 123}`))
	}))
	defer server.Close()

	client := api.NewHTTPClientExternal()
	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)

	ctx := context.Background()
	resp, err := client.Call(ctx, req)
	require.NoError(t, err)

	// Test ToObject function
	type TestStruct struct {
		Name string `json:"name"`
		ID   int    `json:"id"`
	}

	obj, err := api.ToObject[TestStruct](resp)
	assert.NoError(t, err)
	assert.Equal(t, "test", obj.Name)
	assert.Equal(t, 123, obj.ID)
}

func TestToObject_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	client := api.NewHTTPClientExternal()
	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)

	ctx := context.Background()
	resp, err := client.Call(ctx, req)
	require.NoError(t, err)

	type TestStruct struct {
		Name string `json:"name"`
	}

	_, err = api.ToObject[TestStruct](resp)
	assert.Error(t, err)
}
