package handlers

import (
	"io"
	"log/slog"
)

const (
	FormatText = "text"
	FormatJSON = "json"
)

// WriterFunc is a factory that returns an io.Writer and optional handler options.
type WriterFunc func() (io.Writer, *slog.HandlerOptions, error)

// HandlerFunc is a factory that returns a pre-built slog.Handler directly.
// Used by outputs (e.g. OTEL, Datadog) that manage their own handler construction.
type HandlerFunc func() (slog.Handler, error)

var (
	writers      = make(map[string]WriterFunc)
	handlerFuncs = make(map[string]HandlerFunc)
)

// RegisterWriter registers a writer factory for the given output kind.
// The writer will be wrapped in slog.JSONHandler or slog.TextHandler based on format.
func RegisterWriter(kind string, fn WriterFunc) {
	writers[kind] = fn
}

// RegisterHandlerFunc registers a handler factory for the given output kind.
// The factory produces a complete slog.Handler — format parameter is ignored.
func RegisterHandlerFunc(kind string, fn HandlerFunc) {
	handlerFuncs[kind] = fn
}

// Get returns a slog.Handler for the given kind and format.
// It first checks handler factories, then tries to build one from a registered writer.
func Get(kind, format string, level slog.Level) (slog.Handler, error) {
	if fn, ok := handlerFuncs[kind]; ok {
		return fn()
	}

	fn, ok := writers[kind]
	if !ok {
		return nil, ErrNotFoundHandler
	}

	w, opts, err := fn()
	if err != nil {
		return nil, err
	}

	if opts == nil {
		opts = &slog.HandlerOptions{Level: level}
	}

	switch format {
	case FormatText:
		return slog.NewTextHandler(w, opts), nil
	default:
		return slog.NewJSONHandler(w, opts), nil
	}
}
