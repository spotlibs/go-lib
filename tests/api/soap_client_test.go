package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/spotlibs/go-lib/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSOAPClient_Default(t *testing.T) {
	client := api.NewSOAPClient()
	assert.NotNil(t, client)
}

func TestNewSOAPClient_WithOptions(t *testing.T) {
	client := api.NewSOAPClient(
		api.WithSOAPVersion(api.SOAP12),
		api.WithTargetNamespace("http://custom.namespace.org/"),
		api.WithWSSecurity(api.WSSecurity{
			Username:     "user",
			Password:     "pass",
			PasswordType: api.WSPasswordPlainText,
			AddTimestamp: true,
			AddNonce:     true,
		}),
	)
	assert.NotNil(t, client)
}

func TestSOAPClient_Call_SOAP11_Success(t *testing.T) {
	// Create test server that responds with SOAP 1.1 response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify SOAP 1.1 headers
		assert.Equal(t, "text/xml; charset=utf-8", r.Header.Get("Content-Type"))
		assert.Contains(t, r.Header.Get("SOAPAction"), "inquiryUserLAS")
		assert.Equal(t, "POST", r.Method)

		// Return SOAP 1.1 response
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		response := `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <inquiryUserLASResponse xmlns="http://tempuri.org/">
      <inquiryUserLASResult>{"statusCode":"01","statusDesc":"Success","items":[{"pn":"00123456","name":"Test User"}]}</inquiryUserLASResult>
    </inquiryUserLASResponse>
  </soap:Body>
</soap:Envelope>`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client := api.NewSOAPClient()

	params := map[string]string{"PN": "00123456"}
	result, err := client.Call(context.Background(), server.URL, "inquiryUserLAS", params)

	require.NoError(t, err)
	assert.NotNil(t, result)

	// Result should be the JSON string
	var parsed struct {
		StatusCode string `json:"statusCode"`
		StatusDesc string `json:"statusDesc"`
		Items      []struct {
			PN   string `json:"pn"`
			Name string `json:"name"`
		} `json:"items"`
	}
	err = json.Unmarshal(result, &parsed)
	require.NoError(t, err)
	assert.Equal(t, "01", parsed.StatusCode)
	assert.Equal(t, "Success", parsed.StatusDesc)
	assert.Len(t, parsed.Items, 1)
	assert.Equal(t, "00123456", parsed.Items[0].PN)
}

func TestSOAPClient_Call_SOAP12_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify SOAP 1.2 headers
		contentType := r.Header.Get("Content-Type")
		assert.Contains(t, contentType, "application/soap+xml")
		assert.Contains(t, contentType, "action=")
		// SOAP 1.2 should NOT have SOAPAction header
		assert.Empty(t, r.Header.Get("SOAPAction"))

		w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		response := `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Body>
    <inquiryBankResponse xmlns="http://tempuri.org/">
      <inquiryBankResult>{"statusCode":"01","statusDesc":"Success","items":[{"code":"002","name":"BRI"}]}</inquiryBankResult>
    </inquiryBankResponse>
  </soap:Body>
</soap:Envelope>`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client := api.NewSOAPClient(api.WithSOAPVersion(api.SOAP12))

	result, err := client.Call(context.Background(), server.URL, "inquiryBank", nil)

	require.NoError(t, err)
	assert.Contains(t, string(result), "BRI")
}

func TestSOAPClient_Call_WithParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read request body and verify params are in XML
		buf := make([]byte, 4096)
		n, _ := r.Body.Read(buf)
		body := string(buf[:n])

		assert.Contains(t, body, "<PN>00123456</PN>")
		assert.Contains(t, body, "<kode_cabang>0001</kode_cabang>")

		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		response := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <updatePDWKResponse xmlns="http://tempuri.org/">
      <updatePDWKResult>{"statusCode":"01","statusDesc":"Success"}</updatePDWKResult>
    </updatePDWKResponse>
  </soap:Body>
</soap:Envelope>`)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client := api.NewSOAPClient()

	params := map[string]string{
		"PN":           "00123456",
		"kode_cabang":  "0001",
	}
	result, err := client.Call(context.Background(), server.URL, "updatePDWK", params)

	require.NoError(t, err)
	assert.Contains(t, string(result), "Success")
}

func TestSOAPClient_Call_SOAPFault11(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		response := `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <soap:Fault>
      <faultcode>soap:Server</faultcode>
      <faultstring>Internal Server Error</faultstring>
      <detail>Service unavailable</detail>
    </soap:Fault>
  </soap:Body>
</soap:Envelope>`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client := api.NewSOAPClient()

	_, err := client.Call(context.Background(), server.URL, "inquiryBank", nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "SOAP Fault")
	assert.Contains(t, err.Error(), "Internal Server Error")
}

func TestSOAPClient_Call_SOAPFault12(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		response := `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Body>
    <soap:Fault>
      <soap:Code><soap:Value>soap:Sender</soap:Value></soap:Code>
      <soap:Reason><soap:Text>Invalid request parameters</soap:Text></soap:Reason>
      <soap:Detail>Parameter PN is required</soap:Detail>
    </soap:Fault>
  </soap:Body>
</soap:Envelope>`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client := api.NewSOAPClient(api.WithSOAPVersion(api.SOAP12))

	_, err := client.Call(context.Background(), server.URL, "inquiryUserLAS", nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "SOAP Fault")
	assert.Contains(t, err.Error(), "Invalid request parameters")
}

func TestSOAPClient_Call_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := api.NewSOAPClient()

	_, err := client.Call(context.Background(), server.URL, "slowAction", nil, 10*time.Millisecond)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection error on SOAP request")
}

func TestSOAPClient_Call_ConnectionError(t *testing.T) {
	client := api.NewSOAPClient()

	// Use an invalid URL
	_, err := client.Call(context.Background(), "http://localhost:99999", "action", nil, 1*time.Second)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection error on SOAP request")
}

func TestSOAPClient_Call_ResultNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		// Response with wrong action name in result
		response := `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <wrongActionResponse xmlns="http://tempuri.org/">
      <wrongActionResult>data</wrongActionResult>
    </wrongActionResponse>
  </soap:Body>
</soap:Envelope>`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client := api.NewSOAPClient()

	_, err := client.Call(context.Background(), server.URL, "inquiryUserLAS", nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found in response")
}

func TestSOAPClient_Call_WithWSSecurity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read request body and verify WS-Security header is present
		buf := make([]byte, 8192)
		n, _ := r.Body.Read(buf)
		body := string(buf[:n])

		assert.Contains(t, body, "wsse:Security")
		assert.Contains(t, body, "wsse:UsernameToken")
		assert.Contains(t, body, "testuser")
		assert.Contains(t, body, "wsu:Timestamp")
		assert.Contains(t, body, "wsu:Created")
		assert.Contains(t, body, "wsu:Expires")

		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		response := `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <secureActionResponse xmlns="http://tempuri.org/">
      <secureActionResult>{"status":"ok"}</secureActionResult>
    </secureActionResponse>
  </soap:Body>
</soap:Envelope>`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client := api.NewSOAPClient(
		api.WithWSSecurity(api.WSSecurity{
			Username:     "testuser",
			Password:     "testpass",
			PasswordType: api.WSPasswordPlainText,
			AddTimestamp: true,
			AddNonce:     true,
		}),
	)

	result, err := client.Call(context.Background(), server.URL, "secureAction", nil)

	require.NoError(t, err)
	assert.Contains(t, string(result), "ok")
}

func TestSOAPClient_Call_WithWSSecurityDigest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 8192)
		n, _ := r.Body.Read(buf)
		body := string(buf[:n])

		assert.Contains(t, body, "wsse:Security")
		assert.Contains(t, body, "PasswordDigest")
		assert.Contains(t, body, "wsse:Nonce")

		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		response := `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <secureActionResponse xmlns="http://tempuri.org/">
      <secureActionResult>{"authenticated":true}</secureActionResult>
    </secureActionResponse>
  </soap:Body>
</soap:Envelope>`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client := api.NewSOAPClient(
		api.WithWSSecurity(api.WSSecurity{
			Username:     "digestuser",
			Password:     "digestpass",
			PasswordType: api.WSPasswordDigest,
			AddTimestamp: true,
		}),
	)

	result, err := client.Call(context.Background(), server.URL, "secureAction", nil)

	require.NoError(t, err)
	assert.Contains(t, string(result), "authenticated")
}

func TestSOAPClient_Call_GzipResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that client accepts gzip
		assert.Contains(t, r.Header.Get("Accept-Encoding"), "gzip")

		// Return uncompressed for simplicity (gzip encoding tested at transport level)
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		response := `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <testGzipResponse xmlns="http://tempuri.org/">
      <testGzipResult>{"compressed":true}</testGzipResult>
    </testGzipResponse>
  </soap:Body>
</soap:Envelope>`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client := api.NewSOAPClient()

	result, err := client.Call(context.Background(), server.URL, "testGzip", nil)

	require.NoError(t, err)
	assert.Contains(t, string(result), "compressed")
}

func TestSOAPClient_Call_CustomNamespace(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 4096)
		n, _ := r.Body.Read(buf)
		body := string(buf[:n])

		// Verify custom namespace
		assert.Contains(t, body, "http://custom.service.bri.co.id/")

		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		response := `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <customActionResponse xmlns="http://custom.service.bri.co.id/">
      <customActionResult>{"custom":true}</customActionResult>
    </customActionResponse>
  </soap:Body>
</soap:Envelope>`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client := api.NewSOAPClient(
		api.WithTargetNamespace("http://custom.service.bri.co.id/"),
	)

	result, err := client.Call(context.Background(), server.URL, "customAction", nil)

	require.NoError(t, err)
	assert.Contains(t, string(result), "custom")
}

func TestSOAPClient_Call_XMLEscaping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 4096)
		n, _ := r.Body.Read(buf)
		body := string(buf[:n])

		// Verify special characters are escaped
		assert.Contains(t, body, "&amp;")
		assert.Contains(t, body, "&lt;")
		assert.NotContains(t, body, "<script>")

		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		response := `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <escapeTestResponse xmlns="http://tempuri.org/">
      <escapeTestResult>{"escaped":true}</escapeTestResult>
    </escapeTestResponse>
  </soap:Body>
</soap:Envelope>`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client := api.NewSOAPClient()

	params := map[string]string{
		"data": `<script>alert("xss")</script> & "quotes"`,
	}
	result, err := client.Call(context.Background(), server.URL, "escapeTest", params)

	require.NoError(t, err)
	assert.Contains(t, string(result), "escaped")
}

func TestSOAPClient_Call_NoFaultOnSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		// Response that mentions "Fault" in data but is not a SOAP Fault
		response := `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <checkFaultResponse xmlns="http://tempuri.org/">
      <checkFaultResult>{"message":"Default Fault handling is active"}</checkFaultResult>
    </checkFaultResponse>
  </soap:Body>
</soap:Envelope>`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client := api.NewSOAPClient()

	result, err := client.Call(context.Background(), server.URL, "checkFault", nil)

	require.NoError(t, err)
	assert.Contains(t, string(result), "Default Fault handling is active")
}

func TestSOAPClient_Call_EmptyParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 4096)
		n, _ := r.Body.Read(buf)
		body := string(buf[:n])

		// Should have action element but no param children
		assert.Contains(t, body, "<inquiryBank")
		assert.Contains(t, body, "</inquiryBank>")
		// Should NOT have random param elements
		assert.Equal(t, -1, strings.Index(body, "<PN>"))

		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		response := `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <inquiryBankResponse xmlns="http://tempuri.org/">
      <inquiryBankResult>{"statusCode":"01"}</inquiryBankResult>
    </inquiryBankResponse>
  </soap:Body>
</soap:Envelope>`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client := api.NewSOAPClient()

	// Call with nil params
	result, err := client.Call(context.Background(), server.URL, "inquiryBank", nil)

	require.NoError(t, err)
	assert.Contains(t, string(result), "01")
}
