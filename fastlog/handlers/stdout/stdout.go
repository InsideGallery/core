package stdout

import (
	"io"
	"log/slog"
	"os"

	"github.com/InsideGallery/core/fastlog/handlers"
)

const OutKind = "stdout"

func init() {
	handlers.RegisterWriter(OutKind, New)
}

func New() (io.Writer, *slog.HandlerOptions, error) {
	cfg, err := GetConfigFromEnv()
	if err != nil {
		return os.Stdout, nil, nil
	}

	return os.Stdout, &slog.HandlerOptions{
		Level: cfg.Level,
	}, nil
}
