package api

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/spotlibs/go-lib/ctx"
)

// NewHTTPAgent return HTTPAgent implementer that also set some metadata header
// before sending the request.
func NewHTTPAgent() HTTPAgent {
	var trans http.Transport
	trans.MaxConnsPerHost = 50
	trans.MaxIdleConnsPerHost = 15
	trans.MaxIdleConns = 50
	trans.IdleConnTimeout = 10 * time.Second

	var client http.Client
	client.Transport = &trans
	client.Timeout = 30 * time.Second

	return &httpAgent{cl: &client}
}

type httpAgent struct {
	cl *http.Client
}

func (h *httpAgent) Do(req *http.Request, timeouts ...time.Duration) (Response, error) {
	ctx.SetHTTPRequestHeader(req)
	reqTimeout := DEFAULT_TIMEOUT
	if len(timeouts) > 0 {
		reqTimeout = timeouts[0]
	}

	ctxWithTimeout, cancel := context.WithTimeout(req.Context(), reqTimeout)
	defer cancel()
	req = req.WithContext(ctxWithTimeout)

	var response Response
	res, err := h.cl.Do(req)
	if err != nil {
		return response, err
	}
	defer res.Body.Close()

	response.Body, _ = io.ReadAll(res.Body)
	response.StatusCode = res.StatusCode
	response.Header = make(map[string][]string)
	response.Header = res.Header

	return response, nil
}
