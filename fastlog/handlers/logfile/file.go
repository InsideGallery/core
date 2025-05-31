package logfile

import (
	"io"
	"log/slog"
	"os"

	"github.com/InsideGallery/core/fastlog/handlers"
)

const (
	OutKind = "file"
	perm    = 0o666
)

func init() {
	handlers.RegisterWriter(OutKind, New)
}

func New() (io.Writer, *slog.HandlerOptions, error) {
	cfg, err := GetConfigFromEnv()
	if err != nil {
		return nil, nil, err
	}

	w, err := os.OpenFile(cfg.Name, os.O_RDWR|os.O_CREATE|os.O_APPEND, perm)

	return w, &slog.HandlerOptions{
		Level: cfg.Level,
	}, err
}
