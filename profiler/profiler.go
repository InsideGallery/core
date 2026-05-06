package profiler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"sync"
	"sync/atomic"
	"time"

	"github.com/InsideGallery/core/metrics/processors/prometheus"
)

var (
	healthChecks []func() error
	mu           sync.Mutex

	// ErrServiceIsOffline indicates that a health check dependency is offline.
	ErrServiceIsOffline = errors.New("service is offline")
)

// Probe flags for Kubernetes. Set by the application lifecycle (main.go).
var (
	// Started is true once the app has initialized successfully (DB connected, etc.).
	Started atomic.Bool
	// Ready is true once the server is accepting traffic (Fiber listening / NATS subscribed).
	Ready atomic.Bool
)

// AddHealthCheck registers a function called on /healthz, /readyz, and /livez.
// If any check returns an error the service is considered unhealthy.
func AddHealthCheck(f func() error) {
	mu.Lock()
	defer mu.Unlock()

	healthChecks = append(healthChecks, f)
}

// CheckHealth runs every registered health check concurrently and returns
// a joined error if any fail. Exported so the main app server can expose
// /healthz on the Traefik-facing port without auth.
func CheckHealth() error {
	return executeHealthChecks()
}

// ExecuteHealthCheck runs every registered health check.
//
// Deprecated: use CheckHealth.
func ExecuteHealthCheck() error {
	return CheckHealth()
}

// executeHealthChecks runs every registered health check concurrently.
func executeHealthChecks() error {
	mu.Lock()
	checks := make([]func() error, len(healthChecks))
	copy(checks, healthChecks)
	mu.Unlock()

	var (
		wg   sync.WaitGroup
		errs []error
		m    sync.Mutex
	)

	wg.Add(len(checks))

	for _, f := range checks {
		go func() {
			defer wg.Done()

			if err := f(); err != nil {
				m.Lock()

				errs = append(errs, err)
				m.Unlock()
			}
		}()
	}

	wg.Wait()

	return errors.Join(errs...)
}

// Monitor starts a standalone HTTP server on addr exposing health probes and pprof.
// It should be called as early as possible in main(), before DB/NATS/app init.
// Returns a shutdown function that should be deferred.
//
// Usage:
//
//	shutdown := profiler.Monitor(":8011")
//	defer shutdown()
func Monitor(addr string) func() {
	if addr == "" {
		return func() {}
	}

	mux := http.NewServeMux()

	// K8s probes
	mux.HandleFunc("/metrics", prometheus.HTTPHandler)
	mux.HandleFunc("/healthz", healthzHandler)
	mux.HandleFunc("/readyz", readyzHandler)
	mux.HandleFunc("/livez", livezHandler)
	mux.HandleFunc("/startupz", startupzHandler)

	// pprof
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: time.Minute,
	}

	go func() {
		slog.Info("Profiler monitor started", "addr", addr)

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Profiler monitor error", "err", err)
		}
	}()

	return func() {
		const shutdownTimeout = 5 * time.Second

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("Profiler monitor shutdown error", "err", err)
		}
	}
}

// healthzHandler returns overall health status including all registered checks.
func healthzHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	msg := map[string]any{"online": true}
	status := http.StatusOK

	if err := executeHealthChecks(); err != nil {
		slog.Error("Health check failed", "err", err)

		msg["online"] = false
		msg["error"] = err.Error()
		status = http.StatusServiceUnavailable
	}

	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(msg)
}

// readyzHandler checks if the service is ready to accept traffic.
func readyzHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	msg := map[string]any{"ready": Ready.Load()}
	status := http.StatusOK

	if err := executeHealthChecks(); err != nil {
		slog.Error("Readiness check failed", "err", err)

		msg["ready"] = false
		msg["error"] = err.Error()
		status = http.StatusServiceUnavailable
	}

	if !Ready.Load() {
		status = http.StatusServiceUnavailable
	}

	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(msg)
}

// livezHandler checks if the service process is alive.
// This is lightweight — no dependency checks. If the HTTP server can respond, the process is alive.
// K8s uses this to detect deadlocks/hangs. Dependency health is checked by /healthz and /readyz.
func livezHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"live": true})
}

// startupzHandler checks if the service has completed initialization.
func startupzHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	started := Started.Load()
	status := http.StatusOK

	if !started {
		status = http.StatusServiceUnavailable
	}

	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"started": started})
}
