package profiler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func resetState() {
	DefaultState().Reset()
}

func TestHealthzOnline(t *testing.T) {
	resetState()

	w := httptest.NewRecorder()
	healthzHandler(w, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result map[string]any

	body, _ := io.ReadAll(w.Result().Body)
	_ = json.Unmarshal(body, &result)

	if result["online"] != true {
		t.Errorf("expected online=true, got %v", result["online"])
	}
}

func TestHealthzWithPassingCheck(t *testing.T) {
	resetState()
	AddHealthCheck(func() error { return nil })

	w := httptest.NewRecorder()
	healthzHandler(w, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHealthzUnhealthy(t *testing.T) {
	resetState()
	AddHealthCheck(func() error { return errors.New("db down") })

	w := httptest.NewRecorder()
	healthzHandler(w, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}

	var result map[string]any

	body, _ := io.ReadAll(w.Result().Body)
	_ = json.Unmarshal(body, &result)

	if result["online"] != false {
		t.Errorf("expected online=false, got %v", result["online"])
	}

	if result["error"] == nil {
		t.Error("expected error field")
	}
}

func TestHealthzMultipleChecks(t *testing.T) {
	resetState()
	AddHealthCheck(func() error { return nil })
	AddHealthCheck(func() error { return errors.New("redis down") })
	AddHealthCheck(func() error { return errors.New("queue full") })

	w := httptest.NewRecorder()
	healthzHandler(w, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestReadyzNotReady(t *testing.T) {
	resetState()

	w := httptest.NewRecorder()
	readyzHandler(w, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestReadyzReady(t *testing.T) {
	resetState()
	Ready.Store(true)

	w := httptest.NewRecorder()
	readyzHandler(w, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestLivezHealthy(t *testing.T) {
	resetState()

	w := httptest.NewRecorder()
	livezHandler(w, httptest.NewRequest(http.MethodGet, "/livez", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result map[string]any

	body, _ := io.ReadAll(w.Result().Body)
	_ = json.Unmarshal(body, &result)

	if result["live"] != true {
		t.Errorf("expected live=true, got %v", result["live"])
	}
}

func TestLivezAlwaysOK(t *testing.T) {
	resetState()
	// livez is lightweight — always 200 if process can respond. No dependency checks.
	AddHealthCheck(func() error { return errors.New("nats disconnected") })

	w := httptest.NewRecorder()
	livezHandler(w, httptest.NewRequest(http.MethodGet, "/livez", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 (process alive despite unhealthy deps), got %d", w.Code)
	}
}

func TestStartupzNotStarted(t *testing.T) {
	resetState()

	w := httptest.NewRecorder()
	startupzHandler(w, httptest.NewRequest(http.MethodGet, "/startupz", nil))

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestStartupzStarted(t *testing.T) {
	resetState()
	Started.Store(true)

	w := httptest.NewRecorder()
	startupzHandler(w, httptest.NewRequest(http.MethodGet, "/startupz", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestWriteJSONHandlesEncodeError(t *testing.T) {
	t.Helper()

	writeJSON(failingResponseWriter{}, map[string]any{"online": true})
}

type failingResponseWriter struct{}

func (failingResponseWriter) Header() http.Header {
	return http.Header{}
}

func (failingResponseWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func (failingResponseWriter) WriteHeader(int) {}

func TestMonitorDisabled(_ *testing.T) {
	shutdown := Monitor("")
	shutdown() // should be a no-op
}
