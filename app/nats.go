package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	_ "github.com/InsideGallery/core/metrics/processors/prometheus" // register default metrics processor

	"github.com/InsideGallery/core/metrics"
	"github.com/InsideGallery/core/profiler"
	"github.com/InsideGallery/core/queue/nats/client"
	"github.com/InsideGallery/core/queue/nats/middleware"
	"github.com/InsideGallery/core/queue/nats/subscriber"
)

type InitSubscriptions func(
	ctx context.Context, met *metrics.Client, m *middleware.Middleware, handler *subscriber.Subscriber,
) error

const defaultNATSServiceName = "nats"

// NATSMain starts the NATS bootstrap flow and logs failures.
//
// Deprecated: use app.RunNATS with app.Config.
func NATSMain(ctx context.Context, initSubscriptions InitSubscriptions) {
	cfg, err := ConfigFromEnv()
	if err != nil {
		slog.Default().Error("create nats bootstrap options", "err", err)

		return
	}

	if err := RunNATS(ctx, cfg, initSubscriptions); err != nil {
		slog.Default().Error("run nats bootstrap", "err", err)
	}
}

func runNATSOptions(ctx context.Context, options NATSOptions) (runErr error) {
	if ctx == nil {
		return fmt.Errorf("context is not set")
	}

	runCtx, stopRun := context.WithCancel(ctx)
	defer stopRun()

	profilerState := natsProfilerState(options)
	profilerState.SetStarted(false)
	profilerState.SetReady(false)

	shutdownMonitor := profilerState.Monitor(options.MonitorAddr)
	defer shutdownMonitor()
	defer recoverBootstrapPanic(&runErr, "nats bootstrap panic")

	metricsOptions := options.Metrics
	if metricsOptions.ServiceName == "" {
		metricsOptions.ServiceName = natsServiceName(options)
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

	natsRuntime, err := newNATSRuntime(runCtx, options, slog.Default())
	if err != nil {
		return err
	}

	natsClient := natsRuntime.client
	natsConnection := natsRuntime.subscriber
	m := natsMiddleware(options.Middleware, metricsClient)

	var appStopped atomic.Bool

	profilerState.AddHealthCheck(func() error {
		conn := natsConnection.Conn()
		if !conn.IsConnected() {
			return fmt.Errorf("nats connection is closed: %w", profiler.ErrServiceIsOffline)
		}

		if _, err := conn.RTT(); err != nil {
			return fmt.Errorf("nats ping: %w", err)
		}

		if appStopped.Load() {
			return fmt.Errorf("app stopped: %w", profiler.ErrServiceIsOffline)
		}

		return nil
	})

	if options.InitSubscriptions != nil {
		if err = options.InitSubscriptions(runCtx, metricsClient, m, natsConnection); err != nil {
			return fmt.Errorf("init subscriptions: %w", err)
		}
	}

	profilerState.SetStarted(true)
	profilerState.SetReady(true)

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
			slog.Default().Info("shutting down nats worker")

			if err := natsConnection.Close(); err != nil {
				recordShutdownErr(fmt.Errorf("stop nats subscriptions: %w", err))
			}

			if natsRuntime.closeClient {
				if err := natsClient.Close(); err != nil {
					recordShutdownErr(fmt.Errorf("stop nats connection: %w", err))
				}
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

	if err = natsConnection.Wait(); err != nil {
		runErr = errors.Join(runErr, fmt.Errorf("run nats handler: %w", err))
	}

	shutdown()

	shutdownErrMu.Lock()
	runErr = errors.Join(runErr, shutdownErr)
	shutdownErrMu.Unlock()

	return runErr
}

// NATSOptionsFromEnv builds compatibility NATS options from environment-derived values.
func NATSOptionsFromEnv(initSubscriptions InitSubscriptions) (NATSOptions, error) {
	cfg, err := ConfigFromEnv()
	if err != nil {
		return NATSOptions{}, err
	}

	return natsOptionsFromConfig(cfg, initSubscriptions), nil
}

type natsRuntime struct {
	client      *client.Client
	subscriber  *subscriber.Subscriber
	closeClient bool
}

func newNATSRuntime(ctx context.Context, options NATSOptions, logger client.Logger) (*natsRuntime, error) {
	if options.Subscriber != nil {
		return &natsRuntime{
			client:      options.NATSClient,
			subscriber:  options.Subscriber,
			closeClient: options.NATSClient != nil && options.CloseNATSClient,
		}, nil
	}

	natsClient := options.NATSClient
	closeClient := options.CloseNATSClient

	if natsClient == nil {
		if options.NATSConfig == nil {
			return nil, fmt.Errorf("nats config is not set")
		}

		var err error

		natsClient, err = client.ConnectClient(ctx, options.NATSConfig, logger)
		if err != nil {
			return nil, fmt.Errorf("get nats client: %w", err)
		}

		closeClient = true
	}

	return &natsRuntime{
		client:      natsClient,
		subscriber:  subscriber.NewSubscriber(natsClient),
		closeClient: closeClient,
	}, nil
}

func natsMiddleware(m *middleware.Middleware, metricsClient *metrics.Client) *middleware.Middleware {
	if m != nil {
		return m
	}

	return middleware.NewMiddleware(
		middleware.GetChains(
			middleware.NewTracer().Call,
			middleware.NewMetrics(middleware.CreateMeasuresWithClient(metricsClient)).Call,
		)...,
	)
}

func natsServiceName(options NATSOptions) string {
	if options.ServiceName != "" {
		return options.ServiceName
	}

	return defaultNATSServiceName
}

func natsProfilerState(options NATSOptions) *profiler.State {
	if options.ProfilerState != nil {
		return options.ProfilerState
	}

	return profiler.DefaultState()
}
