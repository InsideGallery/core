package middlewares

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
)

type failingMetricRecorder struct{}

func (failingMetricRecorder) Count(string, int64, []string) error {
	return errors.New("count failed")
}

func (failingMetricRecorder) Distribution(string, float64, []string) error {
	return errors.New("distribution failed")
}

func TestCORSMiddleware(t *testing.T) {
	cases := []struct {
		name    string
		methods []string
		want    string
	}{
		{
			name:    "single method",
			methods: []string{http.MethodGet},
			want:    http.MethodGet,
		},
		{
			name:    "multiple methods",
			methods: []string{http.MethodGet, http.MethodPost},
			want:    "GET,POST",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			CORSMiddleware(test.methods...)(recorder, httptest.NewRequest(http.MethodOptions, "/", nil))

			if recorder.Code != http.StatusNoContent {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
			}

			if got := recorder.Header().Get("Access-Control-Allow-Methods"); got != test.want {
				t.Fatalf("methods = %q, want %q", got, test.want)
			}

			if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "*" {
				t.Fatalf("origin = %q, want *", got)
			}
		})
	}
}

func TestURLWithoutQuery(t *testing.T) {
	cases := []struct {
		name string
		url  *url.URL
		want string
	}{
		{
			name: "path without query",
			url:  mustParseURL(t, "http://example.test/users?id=1"),
			want: "/users",
		},
		{
			name: "empty path returns slash",
			url:  &url.URL{},
			want: "/",
		},
		{
			name: "opaque schemeless url gets scheme prefix",
			url: &url.URL{
				Scheme: "http",
				Opaque: "//example.test/users",
			},
			want: "http://example.test/users",
		},
		{
			name: "opaque path is returned",
			url: &url.URL{
				Opaque: "mailto:user@example.test",
			},
			want: "mailto:user@example.test",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			req := &http.Request{URL: test.url}
			if got := URLWithoutQuery(req); got != test.want {
				t.Fatalf("url = %q, want %q", got, test.want)
			}
		})
	}
}

func TestRecover(t *testing.T) {
	cases := []struct {
		name       string
		next       http.Handler
		wantStatus int
		wantBody   string
	}{
		{
			name: "passes through successful request",
			next: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusCreated)
				if _, err := w.Write([]byte("ok")); err != nil {
					t.Fatalf("write response: %v", err)
				}
			}),
			wantStatus: http.StatusCreated,
			wantBody:   "ok",
		},
		{
			name: "recovers panic",
			next: http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				panic("boom")
			}),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "Panic during the request",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			Recover(test.next).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, test.wantStatus)
			}

			if !strings.Contains(recorder.Body.String(), test.wantBody) {
				t.Fatalf("body = %q, want to contain %q", recorder.Body.String(), test.wantBody)
			}
		})
	}
}

func TestRecoverFiber(t *testing.T) {
	cases := []struct {
		name       string
		handler    fiber.Handler
		wantStatus int
		wantBody   string
	}{
		{
			name: "passes through successful request",
			handler: func(c fiber.Ctx) error {
				return c.Status(http.StatusCreated).SendString("ok")
			},
			wantStatus: http.StatusCreated,
			wantBody:   "ok",
		},
		{
			name: "recovers panic",
			handler: func(fiber.Ctx) error {
				panic("boom")
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "Panic during the request",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/", RecoverFiber(test.handler))

			resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
			if err != nil {
				t.Fatalf("app test: %v", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}

			if resp.StatusCode != test.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, test.wantStatus)
			}

			if !strings.Contains(string(body), test.wantBody) {
				t.Fatalf("body = %q, want to contain %q", string(body), test.wantBody)
			}
		})
	}
}

func TestTiming(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "stats are empty after reset",
			run: func(t *testing.T) {
				t.Helper()

				avg, count := TimingStats()
				if avg != 0 || count != 0 {
					t.Fatalf("avg = %v count = %d, want zero values", avg, count)
				}
			},
		},
		{
			name: "middleware records request timing",
			run: func(t *testing.T) {
				t.Helper()

				app := fiber.New()
				app.Get("/", Timing(func(c fiber.Ctx) error {
					return c.SendString("ok")
				}))

				resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
				if err != nil {
					t.Fatalf("app test: %v", err)
				}
				defer resp.Body.Close()

				_, count := waitForTimingStats(t)
				if count != 1 {
					t.Fatalf("count = %d, want 1", count)
				}
			},
		},
		{
			name: "reporter stops when context is cancelled",
			run: func(t *testing.T) {
				t.Helper()

				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				StartTimingReporter(ctx)
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}

func TestTelemetry(t *testing.T) {
	cases := []struct {
		name string
		path string
	}{
		{
			name: "request passes through telemetry middleware",
			path: "/users/123",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/users/:id", Telemetry()(func(c fiber.Ctx) error {
				return c.SendString("ok")
			}))

			resp, err := app.Test(httptest.NewRequest(http.MethodGet, test.path, nil))
			if err != nil {
				t.Fatalf("app test: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
			}
		})
	}
}

func TestMetricRecorders(t *testing.T) {
	cases := []struct {
		name   string
		client metricRecorder
	}{
		{
			name: "nil client is ignored",
		},
		{
			name:   "client errors are logged and ignored",
			client: failingMetricRecorder{},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(_ *testing.T) {
			recordMetricCount(test.client, "count", 1, []string{"tag:value"})
			recordMetricDistribution(test.client, "distribution", 1, []string{"tag:value"})
		})
	}
}

func TestJWEErrorBranches(t *testing.T) {
	expectedErr := errors.New("key failed")

	cases := []struct {
		name       string
		body       string
		keyGetter  func(fiber.Ctx) ([]byte, error)
		wantStatus int
	}{
		{
			name: "empty body passes through",
			keyGetter: func(fiber.Ctx) ([]byte, error) {
				return []byte("short"), nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "key getter error is returned",
			body: "not-empty",
			keyGetter: func(fiber.Ctx) ([]byte, error) {
				return nil, expectedErr
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "invalid jwe is returned",
			body: "not-jwe",
			keyGetter: func(fiber.Ctx) ([]byte, error) {
				return []byte("01234567890123456789012345678901"), nil
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			app := fiber.New()
			jwe := NewJWE(test.keyGetter)
			app.Use(jwe.DecryptMiddleware)
			app.Post("/", func(c fiber.Ctx) error {
				return c.SendStatus(http.StatusOK)
			})

			resp, err := app.Test(httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.body)))
			if err != nil {
				t.Fatalf("app test: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != test.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, test.wantStatus)
			}
		})
	}
}

func TestEncryptResponseErrors(t *testing.T) {
	cases := []struct {
		name string
		key  []byte
	}{
		{
			name: "invalid direct key size",
			key:  []byte("short"),
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			result, err := EncryptResponse(test.key, []byte("payload"))
			if err == nil {
				t.Fatal("expected error")
			}

			if result != "" {
				t.Fatalf("result = %q, want empty", result)
			}
		})
	}
}

func mustParseURL(t *testing.T, raw string) *url.URL {
	t.Helper()

	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}

	return parsed
}

func waitForTimingStats(t *testing.T) (time.Duration, int) {
	t.Helper()

	for i := 0; i < 20; i++ {
		avg, count := TimingStats()
		if count > 0 {
			return avg, count
		}

		time.Sleep(10 * time.Millisecond)
	}

	return TimingStats()
}
