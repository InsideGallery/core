package handlers

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestDefaultRegistryHandleEdgeCases(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "nil registry install creates isolated registry",
			run: func(t *testing.T) {
				t.Helper()

				previous := DefaultRegistry()
				handle := InstallDefaultRegistry(nil)

				current := DefaultRegistry()
				if current == nil {
					t.Fatal("default registry is nil")
				}

				if current == previous {
					t.Fatal("default registry was not replaced")
				}

				if err := handle.Close(); err != nil {
					t.Fatalf("close default registry handle: %v", err)
				}

				testutils.Equal(t, DefaultRegistry() == previous, true)
			},
		},
		{
			name: "nil handle close is no-op",
			run: func(t *testing.T) {
				t.Helper()

				var handle *DefaultRegistryHandle
				if err := handle.Close(); err != nil {
					t.Fatalf("close nil default registry handle: %v", err)
				}
			},
		},
		{
			name: "package handler factory wrapper uses default registry",
			run: func(t *testing.T) {
				t.Helper()

				handle := InstallDefaultRegistry(NewRegistry())
				t.Cleanup(func() {
					if err := handle.Close(); err != nil {
						t.Fatalf("close default registry handle: %v", err)
					}
				})

				RegisterHandlerFactory("compat-factory", func() slog.Handler {
					return slog.NewTextHandler(io.Discard, nil)
				})

				handler, err := Get("compat-factory", FormatText, slog.LevelInfo)
				if err != nil {
					t.Fatalf("get compatibility handler factory: %v", err)
				}

				testutils.Equal(t, handler.Enabled(context.Background(), slog.LevelInfo), true)
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}
