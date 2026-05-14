# server/throughput

Import path: `github.com/InsideGallery/core/server/throughput`

`throughput` tracks per-key request throughput and can reject requests that
exceed per-second or 30-day rolling limits.

## Main APIs

- Tier constants: `Tier0` through `Tier3`.
- Limit constants: `Tier0RPS`, `Tier1RPS`, `Tier2RPS`, `Tier3RPS` and matching
  `Tier*RPM` values.
- `GetRPS(tier)` and `GetRPM(tier)`: return configured limits.
- `Storage`: interface for counters, tiers, increments, and resets.
- `MemoryStorage`: in-memory implementation backed by ordered maps and atomic
  counters.
- `New(ctx, storage)`: creates a `Throughput` validator.
- `Throughput.Validate(name)`: increments and returns whether the key is within
  limits.
- `Throughput.Loop()`: resets per-second counters every second until the context
  is cancelled.
- `Throughput.Middleware(parameter)`: Fiber middleware that reads a string from
  `c.Locals(parameter)` and returns HTTP 429 when validation fails.

## Usage

```go
storage := throughput.NewMemoryStorage()
storage.Add("client-1", throughput.Tier1)

limiter := throughput.New(ctx, storage)
go limiter.Loop()

if !limiter.Validate("client-1") {
	return errors.New("too many requests")
}
```

## Operational Notes

The Fiber middleware logs high latency and rejected requests with the short
instance ID from `server/instance`. Callers must set the configured local value
to a string before the middleware runs.
