package handlers

import (
	"fmt"
	"io"
	"log/slog"
)

const (
	FormatText = "text"
	FormatJSON = "json"
)

var (
	writers  = map[string]func() (io.Writer, *slog.HandlerOptions, error){}
	handlers = map[string]slog.Handler{}
)

func RegisterWriter(kind string, writer func() (io.Writer, *slog.HandlerOptions, error)) {
	writers[kind] = writer
}

func RegisterHandler(kind string, handler slog.Handler) {
	handlers[kind] = handler
}

func Get(kind, format string, defaultLogLevel slog.Level) (slog.Handler, error) {
	handler, ok := handlers[kind]
	if ok {
		return handler, nil
	}

	h, ok := writers[kind]
	if !ok {
		return nil, fmt.Errorf("%w: kind: %s", ErrNotFoundHandler, kind)
	}

	w, opts, err := h()
	if err != nil {
		return nil, err
	}

	if opts == nil {
		opts = &slog.HandlerOptions{
			Level: defaultLogLevel,
		}
	}

	switch format {
	case FormatText:
		handler = slog.NewTextHandler(w, opts)
	default:
		handler = slog.NewJSONHandler(w, opts)
	}

	return handler, nil
}
