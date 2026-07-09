package paymentgateway

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"
)

type SDKHTTPRequest struct {
	// Method is the HTTP method, such as GET, POST, PUT, or DELETE.
	Method string
	// URL is the absolute request URL.
	URL string
	// Headers are sent as-is to the gateway.
	Headers map[string]string
	// Body is the JSON request body. It is blank for GET/DELETE calls without a body.
	Body string
	// ConnectTimeout is reserved for custom transports that separate connect and read timeouts.
	ConnectTimeout time.Duration
	// ReadTimeout is used by the default transport as the request context timeout.
	ReadTimeout time.Duration
}

// SDKHTTPResponse is the raw HTTP response returned by a transport.
type SDKHTTPResponse struct {
	StatusCode int
	Headers    map[string][]string
	Body       string
}

// HTTPTransport lets merchants or tests replace the default net/http transport.
type HTTPTransport interface {
	Execute(ctx context.Context, request SDKHTTPRequest) (*SDKHTTPResponse, error)
}

// NetHTTPTransport is the default transport implementation backed by net/http.
type NetHTTPTransport struct {
	Client *http.Client
}

// NewNetHTTPTransport creates a default net/http transport.
func NewNetHTTPTransport() *NetHTTPTransport {
	return &NetHTTPTransport{Client: &http.Client{}}
}

// Execute sends one HTTP request and returns the raw status, headers, and body.
func (t *NetHTTPTransport) Execute(ctx context.Context, request SDKHTTPRequest) (*SDKHTTPResponse, error) {
	client := t.Client
	if client == nil {
		client = &http.Client{}
	}
	timeout := request.ReadTimeout
	if timeout == 0 {
		timeout = HTTPReadTimeoutMS * time.Millisecond
	}
	requestCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var body io.Reader
	if request.Body != "" {
		body = bytes.NewBufferString(request.Body)
	}
	httpRequest, err := http.NewRequestWithContext(requestCtx, request.Method, request.URL, body)
	if err != nil {
		return nil, httpError("failed to create OpenAPI HTTP request", err)
	}
	for key, value := range request.Headers {
		httpRequest.Header.Set(key, value)
	}
	response, err := client.Do(httpRequest)
	if err != nil {
		return nil, httpError("OpenAPI HTTP request failed", err)
	}
	defer response.Body.Close()
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, httpError("failed to read OpenAPI HTTP response", err)
	}
	return &SDKHTTPResponse{
		StatusCode: response.StatusCode,
		Headers:    response.Header,
		Body:       string(responseBytes),
	}, nil
}
