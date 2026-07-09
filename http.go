package paymentgateway

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"
)

type SDKHTTPRequest struct {
	Method         string
	URL            string
	Headers        map[string]string
	Body           string
	ConnectTimeout time.Duration
	ReadTimeout    time.Duration
}

type SDKHTTPResponse struct {
	StatusCode int
	Headers    map[string][]string
	Body       string
}

type HTTPTransport interface {
	Execute(ctx context.Context, request SDKHTTPRequest) (*SDKHTTPResponse, error)
}

type NetHTTPTransport struct {
	Client *http.Client
}

func NewNetHTTPTransport() *NetHTTPTransport {
	return &NetHTTPTransport{Client: &http.Client{}}
}

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
