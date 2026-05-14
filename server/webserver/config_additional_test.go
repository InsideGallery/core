package webserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/InsideGallery/core/profiler"
)

func TestGetEnvConfig(t *testing.T) {
	cases := []struct {
		name    string
		prefix  string
		env     map[string]string
		unset   []string
		want    *Config
		wantErr bool
	}{
		{
			name:  "defaults",
			unset: []string{"APP_ADDR", "APP_HOST", "APP_SCHEME", "APP_NAME", "APP_MONITOR_ADDR", "APP_SHUTDOWN_TIMEOUT"},
			want: &Config{
				Address:         ":8080",
				Host:            "localhost:8080",
				Scheme:          "http",
				Name:            "server",
				MonitorAddr:     ":8011",
				ShutdownTimeout: DefaultShutdownTimeout,
			},
		},
		{
			name:   "custom prefix",
			prefix: "api",
			env: map[string]string{
				"API_ADDR":             ":9090",
				"API_HOST":             "example.test",
				"API_SCHEME":           "https",
				"API_NAME":             "api",
				"API_MONITOR_ADDR":     ":9011",
				"API_SHUTDOWN_TIMEOUT": "2s",
			},
			want: &Config{
				Address:         ":9090",
				Host:            "example.test",
				Scheme:          "https",
				Name:            "api",
				MonitorAddr:     ":9011",
				ShutdownTimeout: 2 * time.Second,
			},
		},
		{
			name: "invalid duration",
			env: map[string]string{
				"APP_SHUTDOWN_TIMEOUT": "bad",
			},
			wantErr: true,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			for _, key := range test.unset {
				unsetEnv(t, key)
			}

			for key, value := range test.env {
				t.Setenv(key, value)
			}

			got, err := GetEnvConfig(test.prefix)
			if test.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}

				return
			}

			if err != nil {
				t.Fatalf("env config: %v", err)
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("config = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestServerHelpers(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "new server stores app and config",
			run: func(t *testing.T) {
				t.Helper()

				cfg := &Config{Name: "unit", ShutdownTimeout: time.Second}
				server := New(cfg)

				if server.App == nil {
					t.Fatal("app is nil")
				}

				if server.cfg != cfg {
					t.Fatal("config was not stored")
				}
			},
		},
		{
			name: "shutdown timeout falls back to default",
			run: func(t *testing.T) {
				t.Helper()

				server := &Server{cfg: &Config{}}
				if got := server.shutdownTimeout(); got != DefaultShutdownTimeout {
					t.Fatalf("timeout = %v, want %v", got, DefaultShutdownTimeout)
				}
			},
		},
		{
			name: "shutdown timeout uses configured value",
			run: func(t *testing.T) {
				t.Helper()

				server := &Server{cfg: &Config{ShutdownTimeout: time.Second}}
				if got := server.shutdownTimeout(); got != time.Second {
					t.Fatalf("timeout = %v, want %v", got, time.Second)
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}

func TestRegisterProbes(t *testing.T) {
	cases := []struct {
		name       string
		path       string
		ready      bool
		started    bool
		wantStatus int
	}{
		{
			name:       "health is online",
			path:       "/healthz",
			wantStatus: http.StatusOK,
		},
		{
			name:       "ready is unavailable when flag is false",
			path:       "/readyz",
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name:       "ready is ok when flag is true",
			path:       "/readyz",
			ready:      true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "live is ok",
			path:       "/livez",
			wantStatus: http.StatusOK,
		},
		{
			name:       "startup is unavailable when flag is false",
			path:       "/startupz",
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name:       "startup is ok when flag is true",
			path:       "/startupz",
			started:    true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "deprecated health registration delegates to probes",
			path:       "/healthz",
			wantStatus: http.StatusOK,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			profiler.Ready.Store(test.ready)
			profiler.Started.Store(test.started)
			t.Cleanup(func() {
				profiler.Ready.Store(false)
				profiler.Started.Store(false)
			})

			app := NewFiberApp("probe")
			if test.name == "deprecated health registration delegates to probes" {
				RegisterHealthz(app)
			} else {
				RegisterProbes(app)
			}

			resp, err := app.Test(httptest.NewRequest(http.MethodGet, test.path, nil))
			if err != nil {
				t.Fatalf("probe request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != test.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, test.wantStatus)
			}
		})
	}
}

func TestRegisterProbesWithState(t *testing.T) {
	cases := []struct {
		name       string
		path       string
		configure  func(*profiler.State)
		wantStatus int
	}{
		{
			name: "ready uses explicit state",
			path: "/readyz",
			configure: func(state *profiler.State) {
				state.SetReady(true)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "startup uses explicit state",
			path: "/startupz",
			configure: func(state *profiler.State) {
				state.SetStarted(true)
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			state := profiler.NewState()
			test.configure(state)

			app := NewFiberApp("probe")
			RegisterProbesWithState(app, state)

			resp, err := app.Test(httptest.NewRequest(http.MethodGet, test.path, nil))
			if err != nil {
				t.Fatalf("probe request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != test.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, test.wantStatus)
			}
		})
	}
}

func TestResponseHelpers(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "success response contains data",
			run: func(t *testing.T) {
				t.Helper()

				resp := GetSuccessResponse("ok")
				if !resp.Ok || resp.Data != "ok" || resp.Error != nil {
					t.Fatalf("response = %#v", resp)
				}
			},
		},
		{
			name: "nil error uses internal error",
			run: func(t *testing.T) {
				t.Helper()

				resp := GetResponseWithError(nil, http.StatusInternalServerError)
				if resp.Ok || resp.Error == nil {
					t.Fatalf("response = %#v", resp)
				}

				if resp.Error.Message != ErrInternal.Error() {
					t.Fatalf("message = %q, want %q", resp.Error.Message, ErrInternal.Error())
				}
			},
		},
		{
			name: "custom error response",
			run: func(t *testing.T) {
				t.Helper()

				expectedErr := errors.New("denied")
				resp := GetResponseWithError(expectedErr, http.StatusForbidden)
				if resp.Error.Message != expectedErr.Error() || resp.Error.Code != http.StatusForbidden {
					t.Fatalf("error = %#v", resp.Error)
				}
			},
		},
		{
			name: "list response calculates pages",
			run: func(t *testing.T) {
				t.Helper()

				resp := GetSuccessResponseList([]int{1, 2}, 5, 2, 2)
				if resp.Pagination.Pages != 3 {
					t.Fatalf("pages = %d, want 3", resp.Pagination.Pages)
				}
			},
		},
		{
			name: "list response handles zero totals",
			run: func(t *testing.T) {
				t.Helper()

				resp := GetSuccessResponseList([]int{}, 0, 1, 25)
				if resp.Pagination.Pages != 0 {
					t.Fatalf("pages = %d, want 0", resp.Pagination.Pages)
				}
			},
		},
		{
			name: "easyjson marshal round trip",
			run: func(t *testing.T) {
				t.Helper()

				resp := GetSuccessResponseList([]int{1}, 1, 1, 25)
				data, err := json.Marshal(resp.Pagination)
				if err != nil {
					t.Fatalf("marshal pagination: %v", err)
				}

				var pagination Pagination
				if err := json.Unmarshal(data, &pagination); err != nil {
					t.Fatalf("unmarshal pagination: %v", err)
				}

				if pagination.Total != 1 {
					t.Fatalf("total = %d, want 1", pagination.Total)
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()

	oldValue, exists := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset %s: %v", key, err)
	}

	t.Cleanup(func() {
		if exists {
			if err := os.Setenv(key, oldValue); err != nil {
				t.Fatalf("restore %s: %v", key, err)
			}

			return
		}

		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("cleanup %s: %v", key, err)
		}
	})
}
