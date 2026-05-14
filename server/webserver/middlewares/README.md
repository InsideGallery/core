# server/webserver/middlewares

Import path: `github.com/InsideGallery/core/server/webserver/middlewares`

This package contains HTTP and Fiber middleware for CORS preflight responses,
panic recovery, JWE request/response handling, metrics, OpenTelemetry tracing,
timing, and URL normalization.

## Main APIs

- `CORSMiddleware(methods...)`: `net/http` preflight handler with wildcard CORS.
- `Recover(next)` and `RecoverFiber(next)`: panic recovery for `net/http` and
  Fiber.
- `NewJWE(keyGetter)` and `JWE.DecryptMiddleware`: decrypt compact JWE request
  bodies, store plaintext in `DecryptValueKey`, and optionally encrypt
  `ResponseValueKey` as a compact JWE response.
- `EncryptResponse` and `GetSessionKey`: JWE response encryption and HKDF-derived
  session keys.
- `Metrics(client)`: Fiber middleware that records request duration, count, and
  server error metrics.
- `Telemetry()`: Fiber middleware backed by OpenTelemetry `otelhttp`; it sets
  `X-Trace-ID` when a span context is available.
- `Timing`, `TimingStats`, and `StartTimingReporter`: request timing collection
  and periodic logging.
- `URLWithoutQuery(r)`: returns an opaque or escaped path without query values.

## Usage

```go
app := fiber.New()
app.Use(middlewares.RecoverFiber)
app.Use(middlewares.Metrics(metricsClient))
app.Use(middlewares.Telemetry())
```

For encrypted requests, handlers read decrypted bytes from
`c.Locals(middlewares.DecryptValueKey)` and set encrypted response bytes with
`c.Locals(middlewares.ResponseValueKey, payload)`.

## Operational Notes

The package depends on Fiber, `go-jose`, and OpenTelemetry HTTP instrumentation.
Metrics clients only need `Count` and `Distribution` methods; recorder errors are
logged and do not fail the request.
