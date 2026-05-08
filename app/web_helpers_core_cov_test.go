package app

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/InsideGallery/core/profiler"
)

func TestWebSupportHelpers(t *testing.T) {
	explicitState := profiler.NewState()
	existingErr := errors.New("existing")

	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "explicit profiler state is used",
			run: func(t *testing.T) {
				t.Helper()

				if got := webProfilerState(WebOptions{ProfilerState: explicitState}); got != explicitState {
					t.Fatal("explicit profiler state was not returned")
				}
			},
		},
		{
			name: "default profiler state is available",
			run: func(t *testing.T) {
				t.Helper()

				if got := webProfilerState(WebOptions{}); got == nil {
					t.Fatal("default profiler state is nil")
				}
			},
		},
		{
			name: "configured shutdown timeout is used",
			run: func(t *testing.T) {
				t.Helper()

				if got := webShutdownTimeout(WebOptions{ShutdownTimeout: time.Second}); got != time.Second {
					t.Fatalf("shutdown timeout = %v, want %v", got, time.Second)
				}
			},
		},
		{
			name: "default shutdown timeout is used",
			run: func(t *testing.T) {
				t.Helper()

				if got := webShutdownTimeout(WebOptions{}); got != shutdownTimeout {
					t.Fatalf("shutdown timeout = %v, want %v", got, shutdownTimeout)
				}
			},
		},
		{
			name: "bootstrap panic is returned",
			run: func(t *testing.T) {
				t.Helper()

				var err error
				func() {
					defer recoverBootstrapPanic(&err, "web bootstrap panic")

					panic("boom")
				}()

				if err == nil || !strings.Contains(err.Error(), "web bootstrap panic: boom") {
					t.Fatalf("panic error = %v, want bootstrap panic", err)
				}
			},
		},
		{
			name: "bootstrap panic joins existing error",
			run: func(t *testing.T) {
				t.Helper()

				err := existingErr
				func() {
					defer recoverBootstrapPanic(&err, "web bootstrap panic")

					panic("boom")
				}()

				if !strings.Contains(err.Error(), existingErr.Error()) {
					t.Fatalf("panic error = %v, want existing error", err)
				}

				if !strings.Contains(err.Error(), "web bootstrap panic: boom") {
					t.Fatalf("panic error = %v, want bootstrap panic", err)
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}
