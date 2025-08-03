package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync/atomic"
	"syscall"

	"github.com/InsideGallery/core/fastlog"
	"github.com/InsideGallery/core/fastlog/handlers/otel"
	"github.com/InsideGallery/core/fastlog/metrics"
	"github.com/InsideGallery/core/oslistener"
	"github.com/InsideGallery/core/queue/nats/client"
	"github.com/InsideGallery/core/queue/nats/middleware"
	"github.com/InsideGallery/core/queue/nats/subscriber"
	"github.com/InsideGallery/core/server/profiler"
)

type InitSubscriptions func(
	ctx context.Context, met *metrics.OTLPMetric, m *middleware.Middleware, handler *subscriber.Subscriber,
) error

func NATSMain(ctx context.Context, initSubscriptions InitSubscriptions) {
	fastlog.SetupDefaultLog()

	defer otel.Default(ctx).Shutdown()
	defer profiler.GOPS()()
	defer func() {
		if rval := recover(); rval != nil {
			slog.Default().Error("Recovered request panic", "rval", rval)
			os.Exit(1)
		}
	}()

	met, err := metrics.Default(ctx)
	if err != nil {
		slog.Default().Error("Error getting metrics", "err", err)
		return
	}

	defer met.Shutdown()

	natsClient, err := client.Default(ctx, slog.Default())
	if err != nil {
		slog.Default().Error("Error getting NATS client", "err", err)
		return
	}

	natsConnection := subscriber.NewSubscriber(natsClient)
	natsConnection.WithMeter(met.GetMetric()) // Add metric to count

	m := middleware.NewMiddleware(
		middleware.GetChains(
			middleware.NewTracer().Call,
			middleware.NewMetrics(middleware.CreateMeasures()).Call,
		)...,
	)

	var appStopped int32

	shutdown := func() {
		atomic.StoreInt32(&appStopped, 1)

		err := natsConnection.Close()
		if err != nil {
			slog.Default().Error("Error stop nats", "err", err)
		}
	}

	oslistener.Get().Append(syscall.SIGTERM, shutdown)
	oslistener.Get().Append(syscall.SIGINT, shutdown)
	oslistener.Get().Append(syscall.SIGQUIT, shutdown)
	oslistener.Get().Append(syscall.SIGHUP, shutdown)

	oslistener.Start(ctx, oslistener.Get())

	profiler.AddHealthCheck(func() error {
		online := natsClient.Conn().IsConnected()
		if !online {
			return fmt.Errorf("nats connection is closed: %w", profiler.ErrServiceIsOffline)
		}

		if atomic.LoadInt32(&appStopped) == 1 {
			return fmt.Errorf("app just stopped: %w", profiler.ErrServiceIsOffline)
		}

		return nil
	})
	defer profiler.Monitor(ctx)()

	if initSubscriptions != nil {
		err = initSubscriptions(ctx, met, m, natsConnection)
		if err != nil {
			slog.Default().Error("Error init subscriptions", "err", err)
			return
		}
	}

	if err = natsConnection.Wait(); err != nil {
		slog.Default().Error("Error run nats handler", "err", err)
		return
	}
}
