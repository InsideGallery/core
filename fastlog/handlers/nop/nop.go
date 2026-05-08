package nop

import (
	"io"
	"log/slog"

	"github.com/InsideGallery/core/fastlog/handlers"
)

const OutKind = "nop"

func init() {
	handlers.DefaultRegistry().RegisterWriter(OutKind, New)
}

type W struct{}

func (W) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// NewFromConfig creates a no-op writer without hidden runtime state.
func NewFromConfig() (io.Writer, *slog.HandlerOptions, error) {
	return W{}, &slog.HandlerOptions{}, nil
}

// New creates a no-op writer.
func New() (io.Writer, *slog.HandlerOptions, error) {
	return NewFromConfig()
}
