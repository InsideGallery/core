# app

Import path: `github.com/InsideGallery/core/app`

## Overview

`app` provides process entrypoint helpers for InsideGallery HTTP services and NATS workers. The helpers install
logging and metrics bundle imports, start the profiler monitor, wire service-specific initialization callbacks,
register shutdown handlers, and then run the server or worker loop.

## Main APIs

- `InitRouter` is the HTTP setup callback: `func(context.Context, *fiber.App) error`.
- `WebMain(name string, cfg *webserver.Config, initRouter InitRouter)` starts a Fiber HTTP service.
- `InitSubscriptions` is the NATS setup callback: `func(context.Context, *subscriber.Subscriber) error`.
- `NATSMain(name, monitorAddr string, initSubs InitSubscriptions)` starts a NATS subscriber worker.

## Usage

```go
cfg, err := webserver.GetEnvConfig()
if err != nil {
	return err
}

app.WebMain("api", cfg, func(ctx context.Context, router *fiber.App) error {
	router.Get("/ping", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	return nil
})
```

## Configuration And Operations

`WebMain` reads logging and metrics configuration through `fastlog.GetConfigFromEnv` and
`metrics.GetEnvConfig`, starts `profiler.Monitor(cfg.MonitorAddr)`, installs the metrics middleware, and listens
on `cfg.Address`. The caller usually builds `cfg` with `server/webserver.GetEnvConfig`.

`NATSMain` reads logging and metrics configuration the same way, creates the default NATS client through
`mq-balancer`, registers a NATS health check, and waits on the subscriber. Its source documents the expected NATS
environment variables as `NATS_ADDR`, `NATS_CONCURRENT_SIZE`, and `NATS_READ_TIMEOUT`.

Both entrypoints register `SIGINT`, `SIGTERM`, and `SIGQUIT` callbacks through `oslistener`. Fatal logging,
metrics, initialization, listen, or worker errors call `os.Exit(1)`, so use lower-level packages directly when a
caller must handle startup errors itself.
