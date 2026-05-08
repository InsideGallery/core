package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	_ "github.com/InsideGallery/core/metrics/processors/prometheus" // register default metrics processor

	"github.com/gofiber/fiber/v3"

	"github.com/InsideGallery/core/metrics"
	"github.com/InsideGallery/core/oslistener"
	"github.com/InsideGallery/core/profiler"
	"github.com/InsideGallery/core/server/webserver"
	webmiddlewares "github.com/InsideGallery/core/server/webserver/middlewares"
)

const shutdownTimeout = 10 * time.Second

// InitRoutes configures routes through a core-owned router.
type InitRoutes = webserver.RouteInitializer

// InitRouter configures routes with the legacy Fiber app callback.
//
// Deprecated: use WebOptions.InitRoutes for new code.
type InitRouter func(ctx context.Context, app *fiber.App, met *metrics.Client) error

// WebMain starts the web bootstrap flow and logs failures.
//
// Deprecated: use app.RunWeb with app.Config.
func WebMain(ctx context.Context, port string, serverName string, initRouter InitRouter) {
	cfg, err := webConfigFromEnv(port, serverName)
	if err != nil {
		slog.Default().Error("create web bootstrap options", "err", err)

		return
	}

	if err := RunWeb(ctx, cfg, initRouter); err != nil {
		slog.Default().Error("run web bootstrap", "err", err)
	}
}

func runWebOptions(ctx context.Context, options WebOptions) (runErr error) {
	if ctx == nil {
		return fmt.Errorf("context is not set")
	}

	runCtx, stopRun := context.WithCancel(ctx)
	defer stopRun()

	profilerState := webProfilerState(options)
	profilerState.SetStarted(false)
	profilerState.SetReady(false)

	shutdownMonitor := profilerState.Monitor(options.MonitorAddr)
	defer shutdownMonitor()
	defer recoverBootstrapPanic(&runErr, "web bootstrap panic")

	metricsOptions := options.Metrics
	if metricsOptions.ServiceName == "" {
		metricsOptions.ServiceName = options.ServerName
	}

	if metricsOptions.HealthState == nil {
		metricsOptions.HealthState = profilerState
	}

	metricsRuntime, err := NewMetricsClient(metricsOptions)
	if err != nil {
		return fmt.Errorf("init metrics: %w", err)
	}

	metricsClient := metricsRuntime.Client()

	var (
		closeMetricsOnce  sync.Once
		closeMetricsErr   error
		closeMetricsErrMu sync.Mutex
	)

	closeMetricsFunc := func() {
		closeMetricsOnce.Do(func() {
			if err := metricsRuntime.Close(); err != nil {
				closeMetricsErrMu.Lock()
				closeMetricsErr = errors.Join(closeMetricsErr, err)
				closeMetricsErrMu.Unlock()
			}
		})
	}

	defer func() {
		closeMetricsFunc()

		closeMetricsErrMu.Lock()
		runErr = errors.Join(runErr, closeMetricsErr)
		closeMetricsErrMu.Unlock()
	}()

	app := webserver.NewFiberApp(options.ServerName)
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

	if options.InitRouter != nil {
		if err = options.InitRouter(runCtx, app, metricsClient); err != nil {
			return fmt.Errorf("init routers: %w", err)
		}
	}

	if options.InitRoutes != nil {
		if err = options.InitRoutes(runCtx, webserver.NewFiberRouter(app)); err != nil {
			return fmt.Errorf("init routes: %w", err)
		}
	}

	profilerState.SetStarted(true)

	var shutdownOnce sync.Once

	var (
		shutdownErr   error
		shutdownErrMu sync.Mutex
	)

	recordShutdownErr := func(err error) {
		if err == nil {
			return
		}

		shutdownErrMu.Lock()
		defer shutdownErrMu.Unlock()

		shutdownErr = errors.Join(shutdownErr, err)
	}

	shutdown := func() {
		shutdownOnce.Do(func() {
			appStopped.Store(true)
			profilerState.SetReady(false)
			slog.Default().Info("shutting down web server", "service", options.ServerName)

			shutdownCtx, cancel := context.WithTimeout(context.Background(), webShutdownTimeout(options))
			defer cancel()

			if err := app.ShutdownWithContext(shutdownCtx); err != nil && !errors.Is(err, fiber.ErrNotRunning) {
				recordShutdownErr(fmt.Errorf("stop fiber: %w", err))
			}

			closeMetricsFunc()
		})
	}

	registerShutdown(options.ShutdownListener, shutdown)
	startSignalListener(runCtx, options.ShutdownListener)

	go func() {
		<-runCtx.Done()
		shutdown()
	}()

	err = app.Listen(options.Port, fiber.ListenConfig{
		GracefulContext:   runCtx,
		ShutdownTimeout:   webShutdownTimeout(options),
		EnablePrintRoutes: options.Environment.EnablePrintRoutes,
		BeforeServeFunc: func(_ *fiber.App) error {
			profilerState.SetReady(true)

			return nil
		},
	})

	profilerState.SetReady(false)
	shutdown()

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		runErr = errors.Join(runErr, fmt.Errorf("server stopped: %w", err))
	}

	shutdownErrMu.Lock()
	runErr = errors.Join(runErr, shutdownErr)
	shutdownErrMu.Unlock()

	return runErr
}

// WebOptionsFromEnv builds compatibility web options from environment-derived values.
func WebOptionsFromEnv(port string, serverName string, initRouter InitRouter) (WebOptions, error) {
	cfg, err := webConfigFromEnv(port, serverName)
	if err != nil {
		return WebOptions{}, err
	}

	return webOptionsFromConfig(cfg, initRouter), nil
}

func webProfilerState(options WebOptions) *profiler.State {
	if options.ProfilerState != nil {
		return options.ProfilerState
	}

	return profiler.DefaultState()
}

func webShutdownTimeout(options WebOptions) time.Duration {
	if options.ShutdownTimeout > 0 {
		return options.ShutdownTimeout
	}

	return shutdownTimeout
}

func registerShutdown(listener *oslistener.SignalListener, shutdown func()) {
	if listener == nil {
		return
	}

	listener.Append(syscall.SIGTERM, shutdown)
	listener.Append(syscall.SIGINT, shutdown)
	listener.Append(syscall.SIGQUIT, shutdown)
	listener.Append(syscall.SIGHUP, shutdown)
}

func startSignalListener(ctx context.Context, listener *oslistener.SignalListener) {
	if listener == nil {
		return
	}

	oslistener.Start(ctx, listener)
}

func recoverPanic(message string) {
	if value := recover(); value != nil {
		slog.Default().Error(message, "value", value)
	}
}

func recoverBootstrapPanic(err *error, message string) {
	if value := recover(); value != nil {
		*err = errors.Join(*err, fmt.Errorf("%s: %v", message, value))
	}
}
