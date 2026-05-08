// Package app provides simple application bootstrap helpers.
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

	_ "github.com/InsideGallery/core/metrics/processors/prometheus" // register default metrics processor

	"github.com/gofiber/fiber/v3"

	"github.com/InsideGallery/core/fastlog"
	"github.com/InsideGallery/core/metrics"
	"github.com/InsideGallery/core/oslistener"
	"github.com/InsideGallery/core/profiler"
	"github.com/InsideGallery/core/server/webserver"
	webmiddlewares "github.com/InsideGallery/core/server/webserver/middlewares"
)

const (
	defaultWebServiceName = "web"
	shutdownTimeout       = 10 * time.Second
)

type InitRouter func(ctx context.Context, app *fiber.App, met *metrics.Client) error

func WebMain(ctx context.Context, port string, serverName string, initRouter InitRouter) {
	closeLogger, err := setupLogger(ctx)
	if err != nil {
		slog.Default().Error("init logger", "err", err)

		return
	}
	defer closeAndLog("close logger", closeLogger)

	profilerState := resetProfiler()

	shutdownMonitor := profilerState.Monitor(os.Getenv("MONITOR_ADDR"))
	defer shutdownMonitor()
	defer recoverPanic("Recovered request panic")

	serviceName := webServiceName(serverName)

	metricsClient, closeMetrics, err := setupMetrics(serviceName, profilerState)
	if err != nil {
		slog.Default().Error("init metrics", "err", err)

		return
	}

	var closeMetricsOnce sync.Once

	closeMetricsFunc := func() {
		closeMetricsOnce.Do(func() {
			closeAndLog("close metrics", closeMetrics)
		})
	}
	defer closeMetricsFunc()

	app := webserver.NewFiberApp(serviceName)
	if metricsClient != nil {
		app.Use(webmiddlewares.Metrics(metricsClient))
	}

	var appStopped atomic.Bool

	profilerState.AddHealthCheck(func() error {
		if appStopped.Load() {
			return fmt.Errorf("app stopped: %w", profiler.ErrServiceIsOffline)
		}

		return nil
	})

	if initRouter != nil {
		if err := initRouter(ctx, app, metricsClient); err != nil {
			slog.Default().Error("init routers", "err", err)

			return
		}
	}

	profilerState.SetStarted(true)

	var shutdownOnce sync.Once

	shutdown := func() {
		shutdownOnce.Do(func() {
			appStopped.Store(true)
			profilerState.SetReady(false)
			slog.Default().Info("shutting down web server", "service", serviceName)

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
			profilerState.SetReady(true)

			return nil
		},
	})

	profilerState.SetReady(false)
	shutdown()

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Default().Error("server stopped with error", "err", err)
	}
}

func setupLogger(ctx context.Context) (func() error, error) {
	cfg, err := fastlog.GetConfigFromEnv()
	if err != nil {
		return nil, fmt.Errorf("log config: %w", err)
	}

	closeLogger, err := fastlog.SetupDefault(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("log setup: %w", err)
	}

	return closeLogger, nil
}

func setupMetrics(serviceName string, profilerState *profiler.State) (*metrics.Client, func() error, error) {
	cfg, err := metrics.GetEnvConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("metrics config: %w", err)
	}

	metricsClient, err := metrics.NewWithRegistry(nil, cfg, serviceName)
	if err != nil {
		return nil, nil, fmt.Errorf("metrics init: %w", err)
	}

	metricsHandle := metrics.InstallDefault(metricsClient)
	if metricsClient != nil {
		profilerState.AddHealthCheck(metricsClient.HealthCheck)
	}

	return metricsClient, metricsHandle.Close, nil
}

func closeAndLog(message string, closer func() error) {
	if closer == nil {
		return
	}

	if err := closer(); err != nil {
		slog.Default().Error(message, "err", err)
	}
}

func resetProfiler() *profiler.State {
	profilerState := profiler.DefaultState()
	profilerState.Reset()

	return profilerState
}

func registerShutdown(shutdown func()) {
	listener := oslistener.DefaultListener()
	listener.Append(syscall.SIGTERM, shutdown)
	listener.Append(syscall.SIGINT, shutdown)
	listener.Append(syscall.SIGQUIT, shutdown)
	listener.Append(syscall.SIGHUP, shutdown)
}

func startSignalListener(ctx context.Context) {
	oslistener.Start(ctx, oslistener.DefaultListener())
}

func webServiceName(serverName string) string {
	if serverName != "" {
		return serverName
	}

	return defaultWebServiceName
}

func recoverPanic(message string) {
	if value := recover(); value != nil {
		slog.Default().Error(message, "value", value)
	}
}
