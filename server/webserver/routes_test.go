package webserver

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFiberRouterHandle(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "handles route with core-owned callback",
			method:     http.MethodPost,
			path:       "/events?trace=true",
			body:       "payload",
			wantStatus: http.StatusAccepted,
			wantBody:   "POST:/events:payload",
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			app := NewFiberApp("routes")
			NewFiberRouter(app).Handle(test.method, "/events", func(_ context.Context, req RouteRequest) (RouteResponse, error) {
				if req.Query["trace"] != "true" {
					t.Fatalf("query trace = %q, want true", req.Query["trace"])
				}

				return RouteResponse{
					StatusCode: test.wantStatus,
					Header:     map[string][]string{"x-route": {"core"}},
					Body:       []byte(req.Method + ":" + req.Path + ":" + string(req.Body)),
				}, nil
			})

			request := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))

			response, err := app.Test(request)
			if err != nil {
				t.Fatalf("route request: %v", err)
			}
			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}

			if response.StatusCode != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.StatusCode, test.wantStatus)
			}

			if response.Header.Get("x-route") != "core" {
				t.Fatalf("x-route = %q, want core", response.Header.Get("x-route"))
			}

			if string(body) != test.wantBody {
				t.Fatalf("body = %q, want %q", body, test.wantBody)
			}
		})
	}
}
