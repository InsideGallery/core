# server/backoff

Import path: `github.com/InsideGallery/core/server/backoff`

`backoff` wraps an `http.RoundTripper` with retry behavior for outbound HTTP
requests.

## Main APIs

- `PoliticType`: retry policy enum. Supported values are `NoBackoff`,
  `ExponentialBackoff`, and `ConstantBackoff`.
- `HTTPTransport`: an `http.RoundTripper` implementation that retries failed
  transport calls and HTTP responses with status code `>= 400`.
- `NewTransport(delay, retries, backoff, tripper)`: creates a retrying
  transport. Non-positive delay and retry values fall back to package defaults.
- `SetupClientBackoff(client, delay, retries, backoff)`: replaces
  `client.Transport` with a retrying transport.

## Usage

```go
client := &http.Client{}
backoff.SetupClientBackoff(client, 250*time.Millisecond, 3, backoff.ExponentialBackoff)

resp, err := client.Get("https://api.example.test/resource")
```

If the wrapped transport is nil, `HTTPTransport` uses `http.DefaultTransport`.
Constant backoff sleeps the configured delay between attempts; exponential
backoff doubles the delay up to `DefaultMaxInterval`.

## Operational Notes

Each retry sets `req.Close = true`, so callers should expect connections to be
closed between attempts. The transport serializes `RoundTrip` calls with a mutex.
