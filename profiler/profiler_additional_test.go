package profiler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMonitorStartsAndShutsDown(_ *testing.T) {
	shutdown := Monitor("127.0.0.1:0")
	time.Sleep(10 * time.Millisecond)
	shutdown()
}

func TestReadyzWithFailingCheck(t *testing.T) {
	resetState()
	AddHealthCheck(func() error { return errors.New("dependency down") })

	w := httptest.NewRecorder()
	readyzHandler(w, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}
