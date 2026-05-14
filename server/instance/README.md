# server/instance

Import path: `github.com/InsideGallery/core/server/instance`

`instance` exposes stable identifiers for the current process instance.

## Main APIs

- `GetInstanceID() string`: returns the package-level unique instance ID.
- `GetShortInstanceID() string`: lazily returns a short instance ID.

## Usage

```go
slog.Info("serving request",
	"instance_id", instance.GetInstanceID(),
	"siid", instance.GetShortInstanceID(),
)
```

## Operational Notes

The full instance ID is initialized when the package is loaded. The short ID is
initialized once on first use with `sync.Once`; if short ID generation fails, the
error is logged and the returned value can be empty.
