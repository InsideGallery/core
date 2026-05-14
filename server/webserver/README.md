# server/webserver

Import path: `github.com/InsideGallery/core/server/webserver`

`webserver` provides Fiber-backed HTTP server helpers plus core-owned request,
response, route, middleware, and outbound client contracts.

## Main APIs

- `Config` and `GetEnvConfig(prefix...)`: HTTP server configuration. Defaults use
  the `APP` prefix.
- `New(cfg)`: creates a `Server` with a configured Fiber app.
- `Server.Run(ctx)` and `MustRun(ctx)`: run Fiber with graceful shutdown hooks.
- `Options`, `NewRuntime`, and `Runtime.Run(ctx)`: core-owned runtime wrapper.
- `Router`, `RouteHandler`, `RouteRequest`, and `RouteResponse`: route contracts
  that avoid exposing Fiber to route callbacks.
- `NewFiberRouter(router)`: adapts a Fiber router to the core-owned router.
- `Middleware` and `RouteMiddleware`: chain Fiber handlers or core-owned route
  handlers.
- `NewFiberApp(name)`: creates a Fiber app with the package error handler.
- `RegisterProbes` and `RegisterProbesWithState`: install `/healthz`,
  `/readyz`, `/livez`, and `/startupz`.
- `Response`, `ErrorResponse`, `Pagination`, and response helper functions.
- `NewStandardClient(client)`: adapts a net/http-compatible client to the
  core-owned `Client` contract.

## Usage

```go
cfg, err := webserver.GetEnvConfig()
if err != nil {
	return err
}

server := webserver.New(cfg)
webserver.RegisterProbes(server.App)
server.App.Get("/ping", func(c fiber.Ctx) error {
	return c.JSON(webserver.GetSuccessResponse("ok"))
})

return server.Run(ctx)
```

## Configuration

Default environment variables are `APP_ADDR`, `APP_HOST`, `APP_SCHEME`,
`APP_NAME`, `APP_MONITOR_ADDR`, and `APP_SHUTDOWN_TIMEOUT`. Pass a custom prefix
to `GetEnvConfig("api")` to read variables such as `API_ADDR`.

## Operational Notes

This package exposes Fiber for server composition. `Server.Run` registers
SIGINT, SIGTERM, and SIGQUIT shutdown handlers through `oslistener`, marks the
profiler state ready before serving, and uses `DefaultShutdownTimeout` when no
timeout is configured.
