package webserver

import (
	"context"
	"errors"
	"log/slog"
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
	Address         string        `env:"_ADDR" envDefault:":8080"`
	Host            string        `env:"_HOST" envDefault:"localhost:8080"`
	Scheme          string        `env:"_SCHEME" envDefault:"http"`
	Name            string        `env:"_NAME" envDefault:"server"`
	MonitorAddr     string        `env:"_MONITOR_ADDR" envDefault:":8011"`
	ShutdownTimeout time.Duration `env:"_SHUTDOWN_TIMEOUT" envDefault:"10s"`
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

// New creates a new HTTP server with the given configuration.
func New(cfg *Config) *Server {
	return &Server{
		App: NewFiberApp(cfg.Name),
		cfg: cfg,
	}
}

// RegisterHealthz adds probe endpoints to the router.
//
// Deprecated: use RegisterProbes.
func RegisterHealthz(router fiber.Router) {
	RegisterProbes(router)
}

// RegisterProbes adds GET /healthz, /readyz, /livez, and /startupz endpoints.
func RegisterProbes(router fiber.Router) {
	router.Get("/healthz", func(c fiber.Ctx) error {
		if err := profiler.CheckHealth(); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"online": false,
				"error":  err.Error(),
			})
		}

		return c.JSON(fiber.Map{"online": true})
	})

	router.Get("/readyz", func(c fiber.Ctx) error {
		ready := profiler.Ready.Load()
		status := fiber.StatusOK

		if err := profiler.CheckHealth(); err != nil || !ready {
			status = fiber.StatusServiceUnavailable
		}

		return c.Status(status).JSON(fiber.Map{"ready": ready})
	})

	router.Get("/livez", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"live": true})
	})

	router.Get("/startupz", func(c fiber.Ctx) error {
		started := profiler.Started.Load()
		status := fiber.StatusOK

		if !started {
			status = fiber.StatusServiceUnavailable
		}

		return c.Status(status).JSON(fiber.Map{"started": started})
	})
}

// Run starts the Fiber server and blocks until shutdown.
func (s *Server) Run(ctx context.Context) error {
	shutdown := func() {
		profiler.Ready.Store(false)
		slog.Default().Info("shutting down server", "name", s.cfg.Name)

		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout())
		defer cancel()

		if err := s.App.ShutdownWithContext(shutdownCtx); err != nil && !errors.Is(err, fiber.ErrNotRunning) {
			slog.Default().Error("shutdown server", "err", err)
		}
	}

	listener := oslistener.Get()
	listener.Append(syscall.SIGINT, shutdown)
	listener.Append(syscall.SIGTERM, shutdown)
	listener.Append(syscall.SIGQUIT, shutdown)

	oslistener.Start(ctx, listener)

	return s.App.Listen(s.cfg.Address, fiber.ListenConfig{
		GracefulContext:   ctx,
		ShutdownTimeout:   s.shutdownTimeout(),
		EnablePrintRoutes: os.Getenv("DEPLOYMENT_ENVIRONMENT") == "",
		BeforeServeFunc: func(_ *fiber.App) error {
			profiler.Ready.Store(true)

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
