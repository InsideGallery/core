package webserver

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestStandardClientErrorBranches(t *testing.T) {
	cases := []struct {
		name   string
		client HTTPClient
		req    HTTPRequest
	}{
		{
			name: "invalid request url",
			client: roundTripClient(func(*http.Request) (*http.Response, error) {
				t.Fatal("client should not be called")

				return nil, nil
			}),
			req: HTTPRequest{Method: http.MethodGet, URL: "://bad"},
		},
		{
			name: "execute error",
			client: roundTripClient(func(*http.Request) (*http.Response, error) {
				return nil, errors.New("execute failed")
			}),
			req: HTTPRequest{Method: http.MethodGet, URL: "https://inside.test"},
		},
		{
			name: "read error",
			client: roundTripClient(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       errReadCloser{readErr: errors.New("read failed")},
				}, nil
			}),
			req: HTTPRequest{Method: http.MethodGet, URL: "https://inside.test"},
		},
		{
			name: "close error",
			client: roundTripClient(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       errReadCloser{reader: strings.NewReader("ok"), closeErr: errors.New("close failed")},
				}, nil
			}),
			req: HTTPRequest{Method: http.MethodGet, URL: "https://inside.test"},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if _, err := NewStandardClient(test.client).Do(context.Background(), test.req); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestRuntimeRunReturnsListenError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runtime := NewRuntime(Options{
		Address:         "127.0.0.1:-1",
		Name:            "unit",
		ShutdownTimeout: time.Millisecond,
	})

	if _, err := runtime.Run(ctx); err == nil {
		t.Fatal("expected listen error")
	}
}

type errReadCloser struct {
	reader   io.Reader
	readErr  error
	closeErr error
}

func (r errReadCloser) Read(p []byte) (int, error) {
	if r.readErr != nil {
		return 0, r.readErr
	}

	return r.reader.Read(p)
}

func (r errReadCloser) Close() error {
	return r.closeErr
}
