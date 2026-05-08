package datadog

import (
	"io"
	"log/slog"
	"testing"

	"github.com/InsideGallery/core/fastlog/handlers"
)

func TestSetup(t *testing.T) {
	cases := []struct {
		name     string
		registry *handlers.Registry
		factory  handlers.HandlerFactory
		wantErr  bool
	}{
		{
			name:     "registers explicit handler factory",
			registry: handlers.NewRegistry(),
			factory: func() slog.Handler {
				return slog.NewTextHandler(io.Discard, nil)
			},
		},
		{
			name: "rejects nil registry",
			factory: func() slog.Handler {
				return slog.NewTextHandler(io.Discard, nil)
			},
			wantErr: true,
		},
		{
			name:     "rejects nil factory",
			registry: handlers.NewRegistry(),
			wantErr:  true,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			err := Setup(test.registry, test.factory)
			if (err != nil) != test.wantErr {
				t.Fatalf("Setup error = %v, wantErr %t", err, test.wantErr)
			}

			if test.wantErr {
				return
			}

			err = Setup(test.registry, test.factory)
			if err != nil {
				t.Fatalf("repeat setup: %v", err)
			}

			handler, err := test.registry.Get(OutKind, handlers.FormatJSON, slog.LevelInfo)
			if err != nil {
				t.Fatalf("get handler: %v", err)
			}

			if handler == nil {
				t.Fatal("handler is nil")
			}
		})
	}
}
