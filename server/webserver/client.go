//go:generate mockgen -source=client.go -destination=mocks/client.go
package webserver

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

// HTTPClient is the legacy net/http client boundary.
//
// Deprecated: use StandardClient with Request and Response for a core-owned contract.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// HTTPRequest is the core-owned HTTP request shape for outbound web calls.
type HTTPRequest struct {
	Method string
	URL    string
	Header map[string][]string
	Body   []byte
}

// HTTPResponse is the core-owned HTTP response shape for outbound web calls.
type HTTPResponse struct {
	StatusCode int
	Header     map[string][]string
	Body       []byte
}

// Client executes HTTP requests without exposing net/http request and response values.
type Client interface {
	Do(ctx context.Context, req HTTPRequest) (HTTPResponse, error)
}

// StandardClient adapts a legacy HTTPClient to the core-owned Client contract.
type StandardClient struct {
	client HTTPClient
}

// NewStandardClient wraps a net/http-compatible client with the core-owned Client contract.
func NewStandardClient(client HTTPClient) *StandardClient {
	if client == nil {
		client = http.DefaultClient
	}

	return &StandardClient{client: client}
}

// Do executes a request and returns a core-owned response.
func (c *StandardClient) Do(ctx context.Context, req HTTPRequest) (HTTPResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, bytes.NewReader(req.Body))
	if err != nil {
		return HTTPResponse{}, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header = cloneHeader(req.Header)

	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return HTTPResponse{}, fmt.Errorf("execute request: %w", err)
	}

	body, err := io.ReadAll(httpResp.Body)
	closeErr := httpResp.Body.Close()

	if err != nil {
		return HTTPResponse{}, fmt.Errorf("read response: %w", err)
	}

	if closeErr != nil {
		return HTTPResponse{}, fmt.Errorf("close response: %w", closeErr)
	}

	return HTTPResponse{
		StatusCode: httpResp.StatusCode,
		Header:     cloneHeader(httpResp.Header),
		Body:       body,
	}, nil
}

func cloneHeader(header map[string][]string) map[string][]string {
	if len(header) == 0 {
		return nil
	}

	cloned := make(http.Header, len(header))
	for key, values := range header {
		for _, value := range values {
			cloned.Add(key, value)
		}
	}

	return cloned
}
