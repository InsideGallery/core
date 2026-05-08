package webserver

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/gofiber/fiber/v3"

	"github.com/InsideGallery/core/oslistener"
	"github.com/InsideGallery/core/profiler"
)

const (
	// EnvPrefix is the default environment variable prefix for HTTP server config.
	EnvPrefix = "APP"

	// DefaultShutdownTimeout is the default graceful shutdown timeout.
	DefaultShutdownTimeout = 10 * time.Second
)

// Config holds HTTP server configuration.
type Config struct {
	Address          string                     `env:"_ADDR" envDefault:":8080"`
	Host             string                     `env:"_HOST" envDefault:"localhost:8080"`
	Scheme           string                     `env:"_SCHEME" envDefault:"http"`
	Name             string                     `env:"_NAME" envDefault:"server"`
	MonitorAddr      string                     `env:"_MONITOR_ADDR" envDefault:":8011"`
	ShutdownTimeout  time.Duration              `env:"_SHUTDOWN_TIMEOUT" envDefault:"10s"`
	ShutdownListener *oslistener.SignalListener `env:"-"`
	ProfilerState    *profiler.State            `env:"-"`
}

// Options is the core-owned input for creating a web server runtime.
type Options struct {
	Address          string
	Host             string
	Scheme           string
	Name             string
	MonitorAddr      string
	ShutdownTimeout  time.Duration
	ShutdownListener *oslistener.SignalListener
	ProfilerState    *profiler.State
	InitRoutes       RouteInitializer
}

// RouteRequest is the core-owned inbound HTTP request shape for route callbacks.
type RouteRequest struct {
	Method      string
	Path        string
	OriginalURL string
	Header      map[string][]string
	Query       map[string]string
	Body        []byte
}

// RouteResponse is the core-owned outbound HTTP response shape for route callbacks.
type RouteResponse struct {
	StatusCode int
	Header     map[string][]string
	Body       []byte
}

// RouteHandler handles inbound HTTP requests without exposing Fiber context values.
type RouteHandler func(ctx context.Context, req RouteRequest) (RouteResponse, error)

// Router registers HTTP routes without exposing Fiber router values.
type Router interface {
	Handle(method string, path string, handler RouteHandler)
}

// RouteInitializer configures routes through a core-owned router.
type RouteInitializer func(ctx context.Context, router Router) error

// RunResult is the core-owned result for a completed web server run.
type RunResult struct {
	Name    string
	Address string
}

// GetEnvConfig reads HTTP server configuration from environment variables.
func GetEnvConfig(prefix ...string) (*Config, error) {
	p := EnvPrefix
	if len(prefix) > 0 && prefix[0] != "" {
		p = prefix[0]
	}

	cfg := new(Config)
	if err := env.ParseWithOptions(cfg, env.Options{
		Prefix: strings.ToUpper(p),
	}); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Server wraps fiber.App with graceful shutdown support.
type Server struct {
	App *fiber.App
	cfg *Config
}

// FiberRouter adapts Fiber routers to the core-owned Router contract.
type FiberRouter struct {
	router fiber.Router
}

// Runtime wraps the Fiber-backed server behind core-owned lifecycle methods.
type Runtime struct {
	server     *Server
	initRoutes RouteInitializer
}

// New creates a new HTTP server with the given configuration.
func New(cfg *Config) *Server {
	return &Server{
		App: NewFiberApp(cfg.Name),
		cfg: cfg,
	}
}

// NewFiberRouter wraps a Fiber router with the core-owned Router contract.
func NewFiberRouter(router fiber.Router) *FiberRouter {
	return &FiberRouter{router: router}
}

// NewRuntime creates a server runtime from core-owned options.
func NewRuntime(options Options) *Runtime {
	return &Runtime{
		server: New(&Config{
			Address:          options.Address,
			Host:             options.Host,
			Scheme:           options.Scheme,
			Name:             options.Name,
			MonitorAddr:      options.MonitorAddr,
			ShutdownTimeout:  options.ShutdownTimeout,
			ShutdownListener: options.ShutdownListener,
			ProfilerState:    options.ProfilerState,
		}),
		initRoutes: options.InitRoutes,
	}
}

// Handle registers one route on the wrapped Fiber router.
func (r *FiberRouter) Handle(method string, path string, handler RouteHandler) {
	r.router.Add([]string{method}, path, r.handle(handler))
}

func (r *FiberRouter) handle(handler RouteHandler) fiber.Handler {
	return func(c fiber.Ctx) error {
		response, err := handler(c.Context(), routeRequest(c))
		if err != nil {
			return err
		}

		for key, values := range response.Header {
			for _, value := range values {
				c.Append(key, value)
			}
		}

		status := response.StatusCode
		if status == 0 {
			status = http.StatusOK
		}

		return c.Status(status).Send(response.Body)
	}
}

// Run starts the server and returns a core-owned result when it stops.
func (r *Runtime) Run(ctx context.Context) (RunResult, error) {
	if r.initRoutes != nil {
		if err := r.initRoutes(ctx, NewFiberRouter(r.server.App)); err != nil {
			return RunResult{}, err
		}
	}

	if err := r.server.Run(ctx); err != nil {
		return RunResult{}, err
	}

	return RunResult{
		Name:    r.server.cfg.Name,
		Address: r.server.cfg.Address,
	}, nil
}

// RegisterHealthz adds probe endpoints to the router.
//
// Deprecated: use RegisterProbes.
func RegisterHealthz(router fiber.Router) {
	RegisterProbes(router)
}

// RegisterProbes adds GET /healthz, /readyz, /livez, and /startupz endpoints.
func RegisterProbes(router fiber.Router) {
	RegisterProbesWithState(router, profiler.DefaultState())
}

// RegisterProbesWithState adds probe endpoints backed by explicit profiler state.
func RegisterProbesWithState(router fiber.Router, state *profiler.State) {
	if state == nil {
		state = profiler.DefaultState()
	}

	router.Get("/healthz", func(c fiber.Ctx) error {
		if err := state.CheckHealth(); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"online": false,
				"error":  err.Error(),
			})
		}

		return c.JSON(fiber.Map{"online": true})
	})

	router.Get("/readyz", func(c fiber.Ctx) error {
		ready := state.IsReady()
		status := fiber.StatusOK

		if err := state.CheckHealth(); err != nil || !ready {
			status = fiber.StatusServiceUnavailable
		}

		return c.Status(status).JSON(fiber.Map{"ready": ready})
	})

	router.Get("/livez", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"live": true})
	})

	router.Get("/startupz", func(c fiber.Ctx) error {
		started := state.IsStarted()
		status := fiber.StatusOK

		if !started {
			status = fiber.StatusServiceUnavailable
		}

		return c.Status(status).JSON(fiber.Map{"started": started})
	})
}

// Run starts the Fiber server and blocks until shutdown.
func (s *Server) Run(ctx context.Context) error {
	profilerState := s.profilerState()

	shutdown := func() {
		profilerState.SetReady(false)
		slog.Default().Info("shutting down server", "name", s.cfg.Name)

		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout())
		defer cancel()

		if err := s.App.ShutdownWithContext(shutdownCtx); err != nil && !errors.Is(err, fiber.ErrNotRunning) {
			slog.Default().Error("shutdown server", "err", err)
		}
	}

	listener := s.shutdownListener()
	listener.Append(syscall.SIGINT, shutdown)
	listener.Append(syscall.SIGTERM, shutdown)
	listener.Append(syscall.SIGQUIT, shutdown)

	oslistener.Start(ctx, listener)

	return s.App.Listen(s.cfg.Address, fiber.ListenConfig{
		GracefulContext:   ctx,
		ShutdownTimeout:   s.shutdownTimeout(),
		EnablePrintRoutes: os.Getenv("DEPLOYMENT_ENVIRONMENT") == "",
		BeforeServeFunc: func(_ *fiber.App) error {
			profilerState.SetReady(true)

			return nil
		},
	})
}

// MustRun starts the server and exits the process on failure.
func (s *Server) MustRun(ctx context.Context) {
	if err := s.Run(ctx); err != nil {
		slog.Default().Error("server stopped with error", "err", err)
		os.Exit(1) //nolint:gocritic // application entrypoint helper
	}
}

func (s *Server) shutdownTimeout() time.Duration {
	if s.cfg.ShutdownTimeout > 0 {
		return s.cfg.ShutdownTimeout
	}

	return DefaultShutdownTimeout
}

func (s *Server) shutdownListener() *oslistener.SignalListener {
	if s.cfg.ShutdownListener != nil {
		return s.cfg.ShutdownListener
	}

	return oslistener.DefaultListener()
}

func (s *Server) profilerState() *profiler.State {
	if s.cfg.ProfilerState != nil {
		return s.cfg.ProfilerState
	}

	return profiler.DefaultState()
}

func routeRequest(c fiber.Ctx) RouteRequest {
	return RouteRequest{
		Method:      c.Method(),
		Path:        c.Path(),
		OriginalURL: c.OriginalURL(),
		Header:      cloneHeader(c.GetReqHeaders()),
		Query:       cloneQuery(c.Queries()),
		Body:        append([]byte(nil), c.Body()...),
	}
}

func cloneQuery(query map[string]string) map[string]string {
	if len(query) == 0 {
		return nil
	}

	cloned := make(map[string]string, len(query))
	for key, value := range query {
		cloned[key] = value
	}

	return cloned
}
