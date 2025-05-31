package stderr

import (
	"io"
	"log/slog"
	"os"

	"github.com/InsideGallery/core/fastlog/handlers"
)

const OutKind = "stderr"

func init() {
	handlers.RegisterWriter(OutKind, New)
}

func New() (io.Writer, *slog.HandlerOptions, error) {
	cfg, err := GetConfigFromEnv()
	if err != nil {
		return os.Stderr, nil, nil
	}

	return os.Stderr, &slog.HandlerOptions{
		Level: cfg.Level,
	}, nil
}
