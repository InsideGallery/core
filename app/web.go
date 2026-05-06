package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	_ "github.com/InsideGallery/core/fastlog/handlers/stderr"       // register default log handler
	_ "github.com/InsideGallery/core/metrics/processors/prometheus" // register default metrics processor

	"github.com/gofiber/fiber/v3"

	"github.com/InsideGallery/core/fastlog"
	"github.com/InsideGallery/core/metrics"
	"github.com/InsideGallery/core/oslistener"
	"github.com/InsideGallery/core/profiler"
	"github.com/InsideGallery/core/server/webserver"
	webmiddlewares "github.com/InsideGallery/core/server/webserver/middlewares"
)

const shutdownTimeout = 10 * time.Second

type InitRouter func(ctx context.Context, app *fiber.App, met *metrics.Client) error

func WebMain(ctx context.Context, port string, serverName string, initRouter InitRouter) {
	fastlog.SetupDefaultLog()
	profiler.Started.Store(false)
	profiler.Ready.Store(false)

	shutdownMonitor := profiler.Monitor(os.Getenv("MONITOR_ADDR"))
	defer shutdownMonitor()
	defer recoverPanic("Recovered request panic")

	metricsClient, closeMetrics, err := newMetricsClient(serverName)
	if err != nil {
		slog.Default().Error("init metrics", "err", err)

		return
	}

	var closeMetricsOnce sync.Once

	closeMetricsFunc := func() {
		closeMetricsOnce.Do(func() {
			if err := closeMetrics(); err != nil {
				slog.Default().Error("close metrics", "err", err)
			}
		})
	}
	defer closeMetricsFunc()

	app := webserver.NewFiberApp(serverName)
	if metricsClient != nil {
		app.Use(webmiddlewares.Metrics(metricsClient))
	}

	var appStopped atomic.Bool

	profiler.AddHealthCheck(func() error {
		if appStopped.Load() {
			return fmt.Errorf("app stopped: %w", profiler.ErrServiceIsOffline)
		}

		return nil
	})

	if initRouter != nil {
		if err = initRouter(ctx, app, metricsClient); err != nil {
			slog.Default().Error("init routers", "err", err)

			return
		}
	}

	profiler.Started.Store(true)

	var shutdownOnce sync.Once

	shutdown := func() {
		shutdownOnce.Do(func() {
			appStopped.Store(true)
			profiler.Ready.Store(false)
			slog.Default().Info("shutting down web server", "service", serverName)

			shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
			defer cancel()

			if err := app.ShutdownWithContext(shutdownCtx); err != nil && !errors.Is(err, fiber.ErrNotRunning) {
				slog.Default().Error("stop fiber", "err", err)
			}

			closeMetricsFunc()
		})
	}

	registerShutdown(shutdown)
	startSignalListener(ctx)

	go func() {
		<-ctx.Done()
		shutdown()
	}()

	err = app.Listen(port, fiber.ListenConfig{
		GracefulContext:   ctx,
		ShutdownTimeout:   shutdownTimeout,
		EnablePrintRoutes: os.Getenv("DEPLOYMENT_ENVIRONMENT") == "",
		BeforeServeFunc: func(_ *fiber.App) error {
			profiler.Ready.Store(true)

			return nil
		},
	})

	profiler.Ready.Store(false)

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Default().Error("server stopped with error", "err", err)
	}
}

func registerShutdown(shutdown func()) {
	listener := oslistener.Get()
	listener.Append(syscall.SIGTERM, shutdown)
	listener.Append(syscall.SIGINT, shutdown)
	listener.Append(syscall.SIGQUIT, shutdown)
	listener.Append(syscall.SIGHUP, shutdown)
}

func startSignalListener(ctx context.Context) {
	oslistener.Start(ctx, oslistener.Get())
}

func recoverPanic(message string) {
	if value := recover(); value != nil {
		slog.Default().Error(message, "value", value)
	}
}
