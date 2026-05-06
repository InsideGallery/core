package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"

	_ "github.com/InsideGallery/core/fastlog/handlers/stderr"       // register default log handler
	_ "github.com/InsideGallery/core/metrics/processors/prometheus" // register default metrics processor

	"github.com/InsideGallery/core/fastlog"
	"github.com/InsideGallery/core/metrics"
	"github.com/InsideGallery/core/profiler"
	"github.com/InsideGallery/core/queue/nats/client"
	"github.com/InsideGallery/core/queue/nats/middleware"
	"github.com/InsideGallery/core/queue/nats/subscriber"
)

type InitSubscriptions func(
	ctx context.Context, met *metrics.Client, m *middleware.Middleware, handler *subscriber.Subscriber,
) error

func NATSMain(ctx context.Context, initSubscriptions InitSubscriptions) {
	fastlog.SetupDefaultLog()
	profiler.Started.Store(false)
	profiler.Ready.Store(false)

	shutdownMonitor := profiler.Monitor(os.Getenv("MONITOR_ADDR"))
	defer shutdownMonitor()
	defer recoverPanic("Recovered nats panic")

	metricsClient, closeMetrics, err := newMetricsClient("nats")
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

	natsClient, err := client.Default(ctx, slog.Default())
	if err != nil {
		slog.Default().Error("get nats client", "err", err)

		return
	}

	natsConnection := subscriber.NewSubscriber(natsClient)
	m := middleware.NewMiddleware(
		middleware.GetChains(
			middleware.NewTracer().Call,
			middleware.NewMetrics(middleware.CreateMeasures()).Call,
		)...,
	)

	var appStopped atomic.Bool

	profiler.AddHealthCheck(func() error {
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

	profiler.Started.Store(true)
	profiler.Ready.Store(true)

	var shutdownOnce sync.Once

	shutdown := func() {
		shutdownOnce.Do(func() {
			appStopped.Store(true)
			profiler.Ready.Store(false)
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
}
