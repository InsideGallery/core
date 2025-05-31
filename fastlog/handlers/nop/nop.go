package nop

import (
	"io"
	"log/slog"

	"github.com/InsideGallery/core/fastlog/handlers"
)

const OutKind = "nop"

func init() {
	handlers.RegisterWriter(OutKind, New)
}

type W struct{}

func (W) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func New() (io.Writer, *slog.HandlerOptions, error) {
	return W{}, &slog.HandlerOptions{}, nil
}
