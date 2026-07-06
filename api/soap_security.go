package api

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1" //nolint:gosec // SHA1 required by WS-Security spec
	"encoding/base64"
	"fmt"
	"time"
)

const (
	wsseNS = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd"
	wsuNS  = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd"

	passwordTextType   = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-username-token-profile-1.0#PasswordText"
	passwordDigestType = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-username-token-profile-1.0#PasswordDigest"
	encodingBase64     = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-soap-message-security-1.0#Base64Binary"
)

// WSPasswordType represents the password encoding type for WS-Security.
type WSPasswordType int

const (
	// WSPasswordPlainText sends password as plain text.
	WSPasswordPlainText WSPasswordType = iota

	// WSPasswordDigest sends password as Base64(SHA1(Nonce + Created + Password)).
	WSPasswordDigest
)

// WSSecurity holds the configuration for WS-Security UsernameToken.
type WSSecurity struct {
	Username     string
	Password     string
	PasswordType WSPasswordType
	AddTimestamp bool          // Add wsu:Timestamp to Security header
	AddNonce     bool          // Add Nonce to UsernameToken (always true for Digest)
	TimestampTTL time.Duration // Timestamp validity duration (default: 5 minutes)
}

// buildWSSecurityHeader constructs the <wsse:Security> XML header.
// Returns nil if ws is nil (WS-Security disabled).
func buildWSSecurityHeader(ws *WSSecurity) ([]byte, error) {
	if ws == nil {
		return nil, nil
	}

	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf(
		`<wsse:Security xmlns:wsse="%s" xmlns:wsu="%s">`,
		wsseNS, wsuNS,
	))

	// Timestamp (optional)
	if ws.AddTimestamp {
		ttl := ws.TimestampTTL
		if ttl == 0 {
			ttl = 5 * time.Minute
		}
		buf.Write(buildTimestamp(ttl))
	}

	// UsernameToken
	token, err := buildUsernameToken(ws)
	if err != nil {
		return nil, fmt.Errorf("failed to build UsernameToken: %w", err)
	}
	buf.Write(token)

	buf.WriteString("</wsse:Security>")

	return buf.Bytes(), nil
}

// buildUsernameToken constructs the <wsse:UsernameToken> element.
func buildUsernameToken(ws *WSSecurity) ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString("<wsse:UsernameToken>")
	buf.WriteString(fmt.Sprintf("<wsse:Username>%s</wsse:Username>", xmlEscape(ws.Username)))

	now := time.Now().UTC()
	created := now.Format(time.RFC3339)

	switch ws.PasswordType {
	case WSPasswordDigest:
		nonce, err := generateNonce()
		if err != nil {
			return nil, err
		}

		// Digest = Base64(SHA1(Nonce + Created + Password))
		digest := computePasswordDigest(nonce, created, ws.Password)

		buf.WriteString(fmt.Sprintf(
			`<wsse:Password Type="%s">%s</wsse:Password>`,
			passwordDigestType, digest,
		))
		buf.WriteString(fmt.Sprintf(
			`<wsse:Nonce EncodingType="%s">%s</wsse:Nonce>`,
			encodingBase64, base64.StdEncoding.EncodeToString(nonce),
		))
		buf.WriteString(fmt.Sprintf("<wsu:Created>%s</wsu:Created>", created))

	case WSPasswordPlainText:
		buf.WriteString(fmt.Sprintf(
			`<wsse:Password Type="%s">%s</wsse:Password>`,
			passwordTextType, xmlEscape(ws.Password),
		))

		// Nonce (optional for PlainText)
		if ws.AddNonce {
			nonce, err := generateNonce()
			if err != nil {
				return nil, err
			}
			buf.WriteString(fmt.Sprintf(
				`<wsse:Nonce EncodingType="%s">%s</wsse:Nonce>`,
				encodingBase64, base64.StdEncoding.EncodeToString(nonce),
			))
			buf.WriteString(fmt.Sprintf("<wsu:Created>%s</wsu:Created>", created))
		}
	}

	buf.WriteString("</wsse:UsernameToken>")

	return buf.Bytes(), nil
}

// buildTimestamp constructs the <wsu:Timestamp> element.
func buildTimestamp(ttl time.Duration) []byte {
	now := time.Now().UTC()
	created := now.Format(time.RFC3339)
	expires := now.Add(ttl).Format(time.RFC3339)

	var buf bytes.Buffer
	buf.WriteString("<wsu:Timestamp>")
	buf.WriteString(fmt.Sprintf("<wsu:Created>%s</wsu:Created>", created))
	buf.WriteString(fmt.Sprintf("<wsu:Expires>%s</wsu:Expires>", expires))
	buf.WriteString("</wsu:Timestamp>")

	return buf.Bytes()
}

// generateNonce generates a 16-byte cryptographic random nonce.
func generateNonce() ([]byte, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}
	return nonce, nil
}

// computePasswordDigest computes Base64(SHA1(Nonce + Created + Password))
// as specified by WS-Security UsernameToken Profile 1.0.
func computePasswordDigest(nonce []byte, created, password string) string {
	// Concatenate: nonce bytes + created string + password string
	h := sha1.New() //nolint:gosec // SHA1 required by WS-Security spec
	h.Write(nonce)
	h.Write([]byte(created))
	h.Write([]byte(password))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
