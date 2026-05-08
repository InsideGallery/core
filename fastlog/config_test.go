package fastlog

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/InsideGallery/core/fastlog/handlers"
	"github.com/InsideGallery/core/fastlog/handlers/nop"
	"github.com/InsideGallery/core/testutils"
)

func TestConfigGetHandlerFromRegistryFallback(t *testing.T) {
	cases := []struct {
		name        string
		setup       func(*handlers.Registry)
		cfg         Config
		wantErr     bool
		wantHandler bool
	}{
		{
			name: "malformed output falls back to registered nop handler",
			setup: func(registry *handlers.Registry) {
				registry.RegisterWriter(nop.OutKind, func() (io.Writer, *slog.HandlerOptions, error) {
					return io.Discard, nil, nil
				})
			},
			cfg: Config{
				Outputs: []string{"broken"},
				Level:   slog.LevelInfo,
			},
			wantHandler: true,
		},
		{
			name: "malformed output returns error when fallback is unavailable",
			cfg: Config{
				Outputs: []string{"broken"},
				Level:   slog.LevelInfo,
			},
			wantErr: true,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			registry := handlers.NewRegistry()
			if test.setup != nil {
				test.setup(registry)
			}

			handler, err := test.cfg.GetHandlerFromRegistry(registry)
			if test.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}

				return
			}

			if err != nil {
				t.Fatalf("get handler from registry: %v", err)
			}

			testutils.Equal(t, handler != nil, test.wantHandler)
			testutils.Equal(t, handler.Enabled(context.Background(), slog.LevelInfo), true)
		})
	}
}
