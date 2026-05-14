package webserver

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestStandardClientContract(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name              string
		request           HTTPRequest
		wantRequestHeader string
		wantStatus        int
		wantBody          string
	}{
		{
			name: "executes request",
			request: HTTPRequest{
				Method: http.MethodPost,
				URL:    "https://inside.test/resource",
				Header: map[string][]string{"x-test": {"true"}},
				Body:   []byte("payload"),
			},
			wantRequestHeader: "true",
			wantStatus:        http.StatusAccepted,
			wantBody:          "accepted",
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var _ Client = (*StandardClient)(nil)

			var gotRequestHeader string
			httpClient := roundTripClient(func(req *http.Request) (*http.Response, error) {
				gotRequestHeader = req.Header.Get("x-test")

				return &http.Response{
					StatusCode: http.StatusAccepted,
					Header:     http.Header{"x-response": []string{"ok"}},
					Body:       io.NopCloser(strings.NewReader("accepted")),
				}, nil
			})

			got, err := NewStandardClient(httpClient).Do(context.Background(), test.request)
			if err != nil {
				t.Fatalf("Do() error: %v", err)
			}

			if gotRequestHeader != test.wantRequestHeader {
				t.Fatalf("request header = %q, want %q", gotRequestHeader, test.wantRequestHeader)
			}

			if got.StatusCode != test.wantStatus {
				t.Fatalf("StatusCode = %d, want %d", got.StatusCode, test.wantStatus)
			}

			if string(got.Body) != test.wantBody {
				t.Fatalf("Body = %q, want %q", got.Body, test.wantBody)
			}
		})
	}
}

type roundTripClient func(req *http.Request) (*http.Response, error)

func (c roundTripClient) Do(req *http.Request) (*http.Response, error) {
	return c(req)
}
