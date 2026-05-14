# fastlog/handlers

Import path: `github.com/InsideGallery/core/fastlog/handlers`

`handlers` is the process-wide registry behind `fastlog`. Handler packages register either writer factories or complete
`slog.Handler` factories, and `fastlog.Config` resolves configured output kinds through this registry.

## Main APIs

- `FormatJSON` and `FormatText` are the supported writer-backed output formats.
- `WriterFunc` returns an `io.Writer`, optional `*slog.HandlerOptions`, and an error.
- `HandlerFunc` returns a complete `slog.Handler`.
- `RegisterWriter(kind string, fn WriterFunc)` registers writer-backed handlers.
- `RegisterHandlerFunc(kind string, fn HandlerFunc)` registers complete handler factories.
- `Get(kind, format string, level slog.Level)` returns a handler or `ErrNotFoundHandler`.
- `ErrNotFoundHandler` reports an unknown output kind.

## Usage

```go
package example

import (
	"io"
	"log/slog"
	"os"

	"github.com/InsideGallery/core/fastlog/handlers"
)

func registerStdout() {
	handlers.RegisterWriter("stdout", func() (io.Writer, *slog.HandlerOptions, error) {
		return os.Stdout, nil, nil
	})
}

func newLogger() (*slog.Logger, error) {
	handler, err := handlers.Get("stdout", handlers.FormatJSON, slog.LevelInfo)
	if err != nil {
		return nil, err
	}

	return slog.New(handler), nil
}
```

## Operational Notes

`Get` checks registered `HandlerFunc` values first. For those handlers, the `format` and `level` parameters are handled
by the factory itself. Writer-backed handlers are wrapped in `slog.TextHandler` for `text`; all other formats use
`slog.JSONHandler`. If a writer factory returns nil options, `Get` applies the requested level.
