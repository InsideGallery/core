package nop

import (
	"io"
	"log/slog"

	"github.com/InsideGallery/core/fastlog/handlers"
)

// OutKind is the registry key for the nop handler.
const OutKind = "nop"

func init() {
	handlers.RegisterWriter(OutKind, New)
}

// W is a no-op writer that discards all output.
type W struct{}

func (W) Write(p []byte) (int, error) {
	return len(p), nil
}

// New returns a no-op writer with default handler options.
func New() (io.Writer, *slog.HandlerOptions, error) {
	return W{}, &slog.HandlerOptions{}, nil
}
