// Package app provides reusable application entrypoints for HTTP and NATS services.
// Modeled after github.com/InsideGallery/core/app.
package app

import (
	"context"
	"log/slog"
	"os"
	"syscall"

	_ "github.com/InsideGallery/core/fastlog/all" // register supported log handlers
	_ "github.com/InsideGallery/core/metrics/all" // register supported metrics processors

	"github.com/gofiber/fiber/v3"

	"github.com/InsideGallery/core/fastlog"
	"github.com/InsideGallery/core/metrics"
	"github.com/InsideGallery/core/oslistener"
	"github.com/InsideGallery/core/profiler"
	httpserver "github.com/InsideGallery/core/server/webserver"
	webmiddlewares "github.com/InsideGallery/core/server/webserver/middlewares"
)

// InitRouter is a closure that wires service-specific dependencies (DB, auth, etc.)
// and registers routes on the Fiber app. If it returns nil, all setup succeeded.
type InitRouter func(ctx context.Context, app *fiber.App) error

// WebMain is the complete entrypoint for an HTTP service.
// It handles: logging → profiler → init closure → signals → listen.
// The caller provides only the service-specific wiring in initRouter.
func WebMain(name string, cfg *httpserver.Config, initRouter InitRouter) {
	setupLogging(name)

	ctx := context.Background()
	cfg.Name = name

	defer profiler.Monitor(cfg.MonitorAddr)()

	app := fiber.New(fiber.Config{
		AppName:      name,
		ServerHeader: name,
	})

	metricsCfg, err := appMetricsConfig()
	if err != nil {
		slog.Error("Metrics config failed", "service", name, "err", err)
		os.Exit(1) //nolint:gocritic // intentional — metrics misconfiguration is fatal
	}

	mc, err := metrics.New(metricsCfg, name)
	if err != nil {
		slog.Error("Metrics init failed", "service", name, "err", err)
		os.Exit(1) //nolint:gocritic // intentional — metrics misconfiguration is fatal
	}

	metrics.SetDefault(mc)
	app.Use(webmiddlewares.Metrics(mc))

	if err := initRouter(ctx, app); err != nil {
		slog.Error("Init failed", "service", name, "err", err)
		os.Exit(1) //nolint:gocritic // intentional — init failure is fatal
	}

	profiler.Started.Store(true)

	shutdown := func() {
		profiler.Ready.Store(false)
		slog.Info("Shutting down", "service", name)

		if err := mc.Close(); err != nil {
			slog.Error("Metrics close error", "err", err)
		}

		if err := app.Shutdown(); err != nil {
			slog.Error("Shutdown error", "err", err)
		}
	}

	listener := oslistener.Get()
	listener.Append(syscall.SIGINT, shutdown)
	listener.Append(syscall.SIGTERM, shutdown)
	listener.Append(syscall.SIGQUIT, shutdown)

	oslistener.Start(ctx, listener)

	profiler.Ready.Store(true)

	if err := app.Listen(cfg.Address); err != nil {
		slog.Error("Server stopped", "service", name, "err", err)
		os.Exit(1) //nolint:gocritic // intentional
	}
}

func appMetricsConfig() (metrics.Config, error) {
	metricsCfg, err := metrics.GetEnvConfig()
	if err != nil {
		return metrics.Config{}, err
	}

	return metricsCfg, nil
}

func setupLogging(service string) {
	logConfig, err := fastlog.GetConfigFromEnv()
	if err != nil {
		slog.Error("Logging config failed", "service", service, "err", err)
		os.Exit(1) //nolint:gocritic // intentional — logging misconfiguration is fatal
	}

	if err = fastlog.SetupDefaultLogger(logConfig); err != nil {
		slog.Error("Logging init failed", "service", service, "err", err)
		os.Exit(1) //nolint:gocritic // intentional — logging misconfiguration is fatal
	}
}
