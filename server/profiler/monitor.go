package profiler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"
	"sync"
	"time"

	"github.com/InsideGallery/core/errors"
)

var (
	healthCheck []func() error
	mu          sync.Mutex

	ErrServiceIsOffline = errors.New("service is offline")
)

func AddHealthCheck(f func() error) {
	mu.Lock()
	defer mu.Unlock()

	healthCheck = append(healthCheck, f)
}

func ExecuteHealthCheck() error {
	mu.Lock()
	defer mu.Unlock()

	var errs []error

	for _, f := range healthCheck {
		err := f()
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Combine(errs...)
}

// Monitor run pprof
func Monitor(ctx context.Context) func() {
	addr := os.Getenv("MONITOR_ADDR")
	if addr == "" {
		return func() {}
	}

	mux := http.NewServeMux()
	mux.Handle("/health", http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Add("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)

		msg := map[string]any{"online": true}

		err := ExecuteHealthCheck()
		if err != nil {
			slog.Default().Error("Error get status of service", "err", err)
			msg["online"] = false
			msg["error"] = err.Error()

			writer.WriteHeader(http.StatusServiceUnavailable)
		} else {
			writer.WriteHeader(http.StatusOK)
		}

		wr := json.NewEncoder(writer)
		if err := wr.Encode(msg); err != nil {
			_, _ = writer.Write([]byte(err.Error()))
		}
	}))
	mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	mux.Handle("/debug/pprof/cmdline/", http.HandlerFunc(pprof.Cmdline))
	mux.Handle("/debug/pprof/profile/", http.HandlerFunc(pprof.Profile))
	mux.Handle("/debug/pprof/symbol/", http.HandlerFunc(pprof.Symbol))
	mux.Handle("/debug/pprof/trace/", http.HandlerFunc(pprof.Trace))
	server := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: time.Minute}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			slog.Default().Error("Error listen monitoring", "err", err)
		}
	}()

	return func() {
		err := server.Shutdown(ctx)
		if err != nil {
			slog.Default().Error("Error shutdown monitoring", "err", err)
		}
	}
}
