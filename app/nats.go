package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"

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

func NATSMain(ctx context.Context, initSubscriptions InitSubscriptions) {
	closeLogger, err := setupLogger(ctx)
	if err != nil {
		slog.Default().Error("init logger", "err", err)

		return
	}
	defer closeAndLog("close logger", closeLogger)

	profilerState := resetProfiler()

	shutdownMonitor := profilerState.Monitor(os.Getenv("MONITOR_ADDR"))
	defer shutdownMonitor()
	defer recoverPanic("Recovered nats panic")

	metricsClient, closeMetrics, err := setupMetrics(defaultNATSServiceName, profilerState)
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

	natsConfig, err := client.GetNATSConnectionConfigFromEnv()
	if err != nil {
		slog.Default().Error("get nats config", "err", err)

		return
	}

	natsClient, err := client.ConnectClient(ctx, natsConfig, slog.Default())
	if err != nil {
		slog.Default().Error("get nats client", "err", err)

		return
	}

	natsConnection := subscriber.NewSubscriber(natsClient)
	m := middleware.NewMiddleware(
		middleware.GetChains(
			middleware.NewTracer().Call,
			middleware.NewMetrics(middleware.CreateMeasuresWithClient(metricsClient)).Call,
		)...,
	)

	var appStopped atomic.Bool

	profilerState.AddHealthCheck(func() error {
		conn := natsClient.Conn()
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

	if initSubscriptions != nil {
		if err = initSubscriptions(ctx, metricsClient, m, natsConnection); err != nil {
			slog.Default().Error("init subscriptions", "err", err)

			return
		}
	}

	profilerState.SetStarted(true)
	profilerState.SetReady(true)

	var shutdownOnce sync.Once

	shutdown := func() {
		shutdownOnce.Do(func() {
			appStopped.Store(true)
			profilerState.SetReady(false)
			slog.Default().Info("shutting down nats worker")

			if err := natsConnection.Close(); err != nil {
				slog.Default().Error("stop nats subscriptions", "err", err)
			}

			if err := natsClient.Close(); err != nil {
				slog.Default().Error("stop nats connection", "err", err)
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

	if err = natsConnection.Wait(); err != nil {
		slog.Default().Error("run nats handler", "err", err)

		return
	}

	shutdown()
}
