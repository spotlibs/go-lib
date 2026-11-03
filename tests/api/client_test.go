package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spotlibs/go-lib/api"
	"github.com/stretchr/testify/require"
)

func TestHTTPClient_Call_WaitsForHeadersWithinCustomTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10*time.Second + 250*time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	require.NoError(t, err)
	resp, err := api.NewHTTPClient().Call(req, 15*time.Second)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.GetStatusCode())
}

func TestHTTPClient_Call_EnforcesCustomTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	require.NoError(t, err)
	_, err = api.NewHTTPClient().Call(req, 50*time.Millisecond)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}
