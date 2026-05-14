# profiler

Import path: `github.com/InsideGallery/core/profiler`

## Overview

`profiler` exposes health checks, Kubernetes-style probes, Prometheus metrics, and pprof handlers through a
standalone HTTP monitor server.

## Main APIs

- `State` owns health checks and startup/readiness flags.
- `NewState` creates isolated profiler state; `DefaultState` returns package-level compatibility state.
- `AddHealthCheck`, `CheckHealth`, `ExecuteHealthCheck`, and `Monitor` operate on the default state.
- `(*State).AddHealthCheck`, `CheckHealth`, `Reset`, `SetStarted`, `IsStarted`, `SetReady`, `IsReady`, and
  `Monitor` operate on explicit state.
- `Started` and `Ready` are package-level atomic probe flags used by compatibility helpers.
- `ErrServiceIsOffline` is a reusable health-check error value.

## Usage

```go
state := profiler.NewState()
state.AddHealthCheck(func() error {
	return nil
})

shutdown := state.Monitor(":8011")
defer shutdown()

state.SetStarted(true)
state.SetReady(true)
```

## Endpoints And Operations

`Monitor(addr)` is a no-op when `addr` is empty. Otherwise it starts an HTTP server with `/metrics`, `/healthz`,
`/readyz`, `/livez`, `/startupz`, and `/debug/pprof/*` endpoints, and returns a shutdown function.

Health checks run concurrently and are joined with `errors.Join`. `/healthz` and `/readyz` return HTTP 503 when
checks fail; `/readyz` also requires the ready flag. `/livez` returns OK when the process can respond.
`/startupz` depends on the started flag.
