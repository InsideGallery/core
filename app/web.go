package app

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"syscall"

	"github.com/gofiber/fiber/v2"

	"github.com/InsideGallery/core/fastlog"
	"github.com/InsideGallery/core/fastlog/handlers/otel"
	"github.com/InsideGallery/core/fastlog/metrics"
	"github.com/InsideGallery/core/oslistener"
	"github.com/InsideGallery/core/server/profiler"
	"github.com/InsideGallery/core/server/webserver"
)

type InitRouter func(ctx context.Context, app *fiber.App, met *metrics.OTLPMetric) error

func WebMain(ctx context.Context, port string, serverName string, initRouter InitRouter) {
	fastlog.SetupDefaultLog()

	defer otel.Default(ctx).Shutdown()
	defer profiler.GOPS()()
	defer func() {
		if rval := recover(); rval != nil {
			slog.Default().Error("Recovered request panic", "rval", rval)
		}
	}()

	met, err := metrics.Default(ctx)
	if err != nil {
		slog.Default().Error("Error getting metrics", "err", err)
		return
	}

	defer met.Shutdown()

	app := webserver.NewFiberApp(serverName)

	if initRouter != nil {
		err = initRouter(ctx, app, met)
		if err != nil {
			slog.Default().Error("Error init routers", "err", err)
			return
		}
	}

	var appStopped int32
	app.Hooks().OnShutdown(func() error {
		atomic.StoreInt32(&appStopped, 1)
		return nil
	})

	shutdown := func() {
		atomic.StoreInt32(&appStopped, 1)

		err := app.Shutdown()
		if err != nil {
			slog.Default().Error("Error stop fiber", "err", err)
		}
	}

	oslistener.Get().Append(syscall.SIGTERM, shutdown)
	oslistener.Get().Append(syscall.SIGINT, shutdown)
	oslistener.Get().Append(syscall.SIGQUIT, shutdown)
	oslistener.Get().Append(syscall.SIGHUP, shutdown)

	oslistener.Start(ctx, oslistener.Get())

	profiler.AddHealthCheck(func() error {
		if atomic.LoadInt32(&appStopped) == 1 {
			return fmt.Errorf("app just stopped: %w", profiler.ErrServiceIsOffline)
		}

		return nil
	})
	defer profiler.Monitor(ctx)()

	err = app.Listen(port)
	if err != nil {
		slog.Default().Error("Server has been stopped with error", "err", err)
	}
}
