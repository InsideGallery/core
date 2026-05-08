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

// ErrServiceIsOffline indicates that a health check dependency is offline.
var ErrServiceIsOffline = errors.New("service is offline")

// Probe flags for Kubernetes. Set by the application lifecycle (main.go).
var (
	// Started is true once the app has initialized successfully (DB connected, etc.).
	Started atomic.Bool
	// Ready is true once the server is accepting traffic (Fiber listening / NATS subscribed).
	Ready atomic.Bool
)

var defaultState = &State{ //nolint:gochecknoglobals // compatibility state backing legacy package functions
	started: &Started,
	ready:   &Ready,
}

// State owns profiler health checks and probe flags.
type State struct {
	healthChecks []func() error
	started      *atomic.Bool
	ready        *atomic.Bool
	mu           sync.Mutex
}

// NewState returns isolated profiler health and probe state.
func NewState() *State {
	return &State{
		started: new(atomic.Bool),
		ready:   new(atomic.Bool),
	}
}

// DefaultState returns the package-level compatibility profiler state.
func DefaultState() *State {
	return defaultState
}

// AddHealthCheck registers a function called on /healthz, /readyz, and /livez.
// If any check returns an error the service is considered unhealthy.
func (s *State) AddHealthCheck(f func() error) {
	if s == nil {
		DefaultState().AddHealthCheck(f)

		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.healthChecks = append(s.healthChecks, f)
}

// CheckHealth runs every registered health check concurrently and returns
// a joined error if any fail.
func (s *State) CheckHealth() error {
	if s == nil {
		return DefaultState().CheckHealth()
	}

	return s.executeHealthChecks()
}

// Reset clears health checks and probe flags.
func (s *State) Reset() {
	if s == nil {
		DefaultState().Reset()

		return
	}

	s.mu.Lock()
	s.healthChecks = nil
	s.mu.Unlock()

	s.SetStarted(false)
	s.SetReady(false)
}

// SetStarted stores the startup probe flag.
func (s *State) SetStarted(started bool) {
	s.startedProbe().Store(started)
}

// IsStarted returns the startup probe flag.
func (s *State) IsStarted() bool {
	return s.startedProbe().Load()
}

// SetReady stores the readiness probe flag.
func (s *State) SetReady(ready bool) {
	s.readyProbe().Store(ready)
}

// IsReady returns the readiness probe flag.
func (s *State) IsReady() bool {
	return s.readyProbe().Load()
}

// AddHealthCheck registers a function called on /healthz, /readyz, and /livez.
// If any check returns an error the service is considered unhealthy.
//
// Deprecated: use NewState and State.AddHealthCheck for explicit ownership.
func AddHealthCheck(f func() error) {
	DefaultState().AddHealthCheck(f)
}

// CheckHealth runs every registered health check concurrently and returns
// a joined error if any fail. Exported so the main app server can expose
// /healthz on the Traefik-facing port without auth.
//
// Deprecated: use NewState and State.CheckHealth for explicit ownership.
func CheckHealth() error {
	return DefaultState().CheckHealth()
}

// ExecuteHealthCheck runs every registered health check.
//
// Deprecated: use CheckHealth.
func ExecuteHealthCheck() error {
	return CheckHealth()
}

// Monitor starts a standalone HTTP server on addr exposing health probes and pprof.
// It should be called as early as possible in main(), before DB/NATS/app init.
// Returns a shutdown function that should be deferred.
//
// Usage:
//
//	shutdown := profiler.Monitor(":8011")
//	defer shutdown()
//
// Deprecated: use NewState and State.Monitor for explicit ownership.
func Monitor(addr string) func() {
	return DefaultState().Monitor(addr)
}

// Monitor starts a standalone HTTP server on addr using this state.
func (s *State) Monitor(addr string) func() {
	if addr == "" {
		return func() {}
	}

	state := validState(s)
	mux := http.NewServeMux()

	// K8s probes
	mux.HandleFunc("/metrics", prometheus.HTTPHandler)
	mux.HandleFunc("/healthz", state.healthzHandler)
	mux.HandleFunc("/readyz", state.readyzHandler)
	mux.HandleFunc("/livez", state.livezHandler)
	mux.HandleFunc("/startupz", state.startupzHandler)

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
	DefaultState().healthzHandler(w, nil)
}

func (s *State) healthzHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	msg := map[string]any{"online": true}
	status := http.StatusOK

	if err := s.CheckHealth(); err != nil {
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
	DefaultState().readyzHandler(w, nil)
}

func (s *State) readyzHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ready := s.IsReady()
	msg := map[string]any{"ready": ready}
	status := http.StatusOK

	if err := s.CheckHealth(); err != nil {
		slog.Error("Readiness check failed", "err", err)

		msg["ready"] = false
		msg["error"] = err.Error()
		status = http.StatusServiceUnavailable
	}

	if !ready {
		status = http.StatusServiceUnavailable
	}

	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(msg)
}

// livezHandler checks if the service process is alive.
// This is lightweight — no dependency checks. If the HTTP server can respond, the process is alive.
// K8s uses this to detect deadlocks/hangs. Dependency health is checked by /healthz and /readyz.
func livezHandler(w http.ResponseWriter, _ *http.Request) {
	DefaultState().livezHandler(w, nil)
}

func (s *State) livezHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"live": true})
}

// startupzHandler checks if the service has completed initialization.
func startupzHandler(w http.ResponseWriter, _ *http.Request) {
	DefaultState().startupzHandler(w, nil)
}

func (s *State) startupzHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	started := s.IsStarted()
	status := http.StatusOK

	if !started {
		status = http.StatusServiceUnavailable
	}

	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"started": started})
}

func (s *State) executeHealthChecks() error {
	s.mu.Lock()
	checks := make([]func() error, len(s.healthChecks))
	copy(checks, s.healthChecks)
	s.mu.Unlock()

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

func validState(state *State) *State {
	if state == nil {
		return DefaultState()
	}

	return state
}

func (s *State) startedProbe() *atomic.Bool {
	if s == nil || s.started == nil {
		return &Started
	}

	return s.started
}

func (s *State) readyProbe() *atomic.Bool {
	if s == nil || s.ready == nil {
		return &Ready
	}

	return s.ready
}
