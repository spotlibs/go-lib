package api

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
)

// SOAPVersion represents the SOAP protocol version.
type SOAPVersion int

const (
	// SOAP11 uses namespace http://schemas.xmlsoap.org/soap/envelope/
	// Content-Type: text/xml; charset=utf-8
	// Action via SOAPAction header.
	SOAP11 SOAPVersion = iota

	// SOAP12 uses namespace http://www.w3.org/2003/05/soap-envelope
	// Content-Type: application/soap+xml; charset=utf-8; action="..."
	// Action embedded in Content-Type.
	SOAP12
)

const (
	soap11Namespace = "http://schemas.xmlsoap.org/soap/envelope/"
	soap12Namespace = "http://www.w3.org/2003/05/soap-envelope"
	defaultTargetNS = "http://tempuri.org/"
)

// SOAPFault represents a parsed SOAP fault from either version.
type SOAPFault struct {
	Code   string // e.g. "soap:Server" (1.1) or "env:Sender" (1.2)
	Reason string // human-readable error message
	Detail string // optional detail content
}

// Error implements the error interface.
func (f *SOAPFault) Error() string {
	if f.Detail != "" {
		return fmt.Sprintf("SOAP Fault [%s]: %s (detail: %s)", f.Code, f.Reason, f.Detail)
	}
	return fmt.Sprintf("SOAP Fault [%s]: %s", f.Code, f.Reason)
}

// buildSOAPEnvelope constructs a complete SOAP envelope XML.
func buildSOAPEnvelope(version SOAPVersion, action string, targetNS string, params map[string]string, securityHeader []byte) ([]byte, error) {
	if targetNS == "" {
		targetNS = defaultTargetNS
	}

	ns := soap11Namespace
	if version == SOAP12 {
		ns = soap12Namespace
	}

	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	buf.WriteString(fmt.Sprintf(`<soap:Envelope xmlns:soap="%s">`, ns))

	// Header (only if WS-Security is provided)
	if len(securityHeader) > 0 {
		buf.WriteString("<soap:Header>")
		buf.Write(securityHeader)
		buf.WriteString("</soap:Header>")
	}

	// Body
	buf.WriteString("<soap:Body>")
	buf.WriteString(fmt.Sprintf(`<%s xmlns="%s">`, action, targetNS))

	// Parameters as XML elements
	for key, value := range params {
		buf.WriteString(fmt.Sprintf("<%s>%s</%s>", key, xmlEscape(value), key))
	}

	buf.WriteString(fmt.Sprintf("</%s>", action))
	buf.WriteString("</soap:Body>")
	buf.WriteString("</soap:Envelope>")

	return buf.Bytes(), nil
}

// extractSOAPResult extracts the content of <{action}Result> from the SOAP
// response body. Returns the inner text/XML as raw bytes.
func extractSOAPResult(responseBody []byte, action string) ([]byte, error) {
	resultTag := action + "Result"

	// Find opening tag
	openTag := "<" + resultTag + ">"
	openTagAlt := "<" + resultTag + " " // handle attributes
	closeTag := "</" + resultTag + ">"

	bodyStr := string(responseBody)

	var startIdx int
	if idx := strings.Index(bodyStr, openTag); idx != -1 {
		startIdx = idx + len(openTag)
	} else if idx := strings.Index(bodyStr, openTagAlt); idx != -1 {
		// Find the closing > of the opening tag
		rest := bodyStr[idx:]
		closeAngle := strings.Index(rest, ">")
		if closeAngle == -1 {
			return nil, fmt.Errorf("malformed XML: unclosed opening tag for %s", resultTag)
		}
		startIdx = idx + closeAngle + 1
	} else {
		return nil, fmt.Errorf("result tag <%s> not found in response", resultTag)
	}

	endIdx := strings.Index(bodyStr[startIdx:], closeTag)
	if endIdx == -1 {
		return nil, fmt.Errorf("closing tag </%s> not found in response", resultTag)
	}

	content := bodyStr[startIdx : startIdx+endIdx]
	return []byte(content), nil
}

// parseSOAPFault detects and parses a SOAP fault from the response body.
// Returns nil if no fault is present.
func parseSOAPFault(responseBody []byte, version SOAPVersion) *SOAPFault {
	bodyStr := string(responseBody)

	// Quick check: does the response contain a Fault element?
	if !strings.Contains(bodyStr, "Fault") {
		return nil
	}

	if version == SOAP12 {
		return parseSoap12Fault(responseBody)
	}
	return parseSoap11Fault(responseBody)
}

// parseSoap11Fault parses SOAP 1.1 fault structure:
// <soap:Fault><faultcode>...</faultcode><faultstring>...</faultstring><detail>...</detail></soap:Fault>
func parseSoap11Fault(body []byte) *SOAPFault {
	type fault11 struct {
		XMLName     xml.Name `xml:"Fault"`
		FaultCode   string   `xml:"faultcode"`
		FaultString string   `xml:"faultstring"`
		Detail      string   `xml:"detail"`
	}

	// Try to find and parse the Fault element
	bodyStr := string(body)
	faultStart := strings.Index(bodyStr, "<Fault")
	if faultStart == -1 {
		// Try with namespace prefix
		faultStart = strings.Index(bodyStr, ":Fault")
		if faultStart == -1 {
			return nil
		}
		// Walk back to find the <
		for faultStart > 0 && bodyStr[faultStart] != '<' {
			faultStart--
		}
	}

	faultEnd := strings.Index(bodyStr[faultStart:], "</")
	if faultEnd == -1 {
		return nil
	}
	// Find the actual end of the Fault element
	remaining := bodyStr[faultStart:]
	closeIdx := findClosingTag(remaining, "Fault")
	if closeIdx == -1 {
		return nil
	}
	faultXML := remaining[:closeIdx]

	var f fault11
	if err := xml.Unmarshal([]byte(faultXML), &f); err != nil {
		// Fallback: extract manually
		return &SOAPFault{
			Code:   extractTagContent(bodyStr, "faultcode"),
			Reason: extractTagContent(bodyStr, "faultstring"),
			Detail: extractTagContent(bodyStr, "detail"),
		}
	}

	return &SOAPFault{
		Code:   f.FaultCode,
		Reason: f.FaultString,
		Detail: f.Detail,
	}
}

// parseSoap12Fault parses SOAP 1.2 fault structure:
// <soap:Fault><soap:Code><soap:Value>...</soap:Value></soap:Code>
// <soap:Reason><soap:Text>...</soap:Text></soap:Reason><soap:Detail>...</soap:Detail></soap:Fault>
func parseSoap12Fault(body []byte) *SOAPFault {
	bodyStr := string(body)

	code := extractTagContent(bodyStr, "Value")
	if code == "" {
		code = extractNestedContent(bodyStr, "Code", "Value")
	}

	reason := extractTagContent(bodyStr, "Text")
	if reason == "" {
		reason = extractNestedContent(bodyStr, "Reason", "Text")
	}

	detail := extractTagContent(bodyStr, "Detail")

	if code == "" && reason == "" {
		return nil
	}

	return &SOAPFault{
		Code:   code,
		Reason: reason,
		Detail: detail,
	}
}

// xmlEscape escapes special XML characters in a string.
func xmlEscape(s string) string {
	var buf bytes.Buffer
	if err := xml.EscapeText(&buf, []byte(s)); err != nil {
		return s
	}
	return buf.String()
}

// extractTagContent extracts content between <tag> and </tag>, handling
// namespace prefixes (e.g., <soap:Value>).
func extractTagContent(body, tag string) string {
	// Try without prefix
	start := strings.Index(body, "<"+tag+">")
	if start != -1 {
		start += len("<" + tag + ">")
		end := strings.Index(body[start:], "</"+tag+">")
		if end != -1 {
			return body[start : start+end]
		}
	}

	// Try with namespace prefix pattern :<tag>
	pattern := ":" + tag + ">"
	start = strings.Index(body, pattern)
	if start != -1 {
		start += len(pattern)
		// Find closing tag with any prefix
		closePattern := ":" + tag + ">"
		closeSearch := "</"
		remaining := body[start:]
		closeIdx := strings.Index(remaining, closeSearch)
		for closeIdx != -1 {
			afterClose := remaining[closeIdx+2:]
			if strings.HasPrefix(afterClose, tag+">") || strings.Contains(afterClose[:min(len(afterClose), len(tag)+10)], closePattern) {
				return remaining[:closeIdx]
			}
			remaining = remaining[closeIdx+2:]
			closeIdx = strings.Index(remaining, closeSearch)
		}
	}

	return ""
}

// extractNestedContent extracts content from a nested pattern like
// <parent><child>content</child></parent>.
func extractNestedContent(body, parent, child string) string {
	parentContent := extractTagContent(body, parent)
	if parentContent == "" {
		return ""
	}
	return extractTagContent(parentContent, child)
}

// findClosingTag finds the end position (after closing tag) of an XML element.
func findClosingTag(xml string, tag string) int {
	// Look for </tag> or </*:tag>
	patterns := []string{"</" + tag + ">", ":Fault>"}
	for _, p := range patterns {
		closeTag := "</" // start of any closing tag
		idx := 0
		for {
			pos := strings.Index(xml[idx:], closeTag)
			if pos == -1 {
				break
			}
			absPos := idx + pos
			remaining := xml[absPos:]
			for _, pattern := range patterns {
				endPattern := strings.Index(remaining, pattern)
				if endPattern != -1 && endPattern < 20 { // within reasonable prefix length
					return absPos + endPattern + len(pattern)
				}
			}
			idx = absPos + 2
		}
		_ = p
	}
	return -1
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
