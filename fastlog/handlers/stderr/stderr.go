package stderr

import (
	"io"
	"log/slog"
	"os"

	"github.com/InsideGallery/core/fastlog/handlers"
)

// OutKind is the registry key for the stderr handler.
const OutKind = "stderr"

func init() {
	handlers.RegisterWriter(OutKind, New)
}

// New returns os.Stderr as the writer with level from env config.
func New() (io.Writer, *slog.HandlerOptions, error) {
	cfg, err := getConfigFromEnv()
	if err != nil {
		return os.Stderr, nil, nil //nolint:nilerr
	}

	return os.Stderr, &slog.HandlerOptions{
		Level: cfg.Level,
	}, nil
}
