package handlers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
)

func TestRegistryGet(t *testing.T) {
	expectedErr := errors.New("writer failed")

	cases := []struct {
		name        string
		setup       func(*Registry)
		kind        string
		format      string
		wantErr     bool
		wantEnabled bool
	}{
		{
			name: "registered handler is returned",
			setup: func(registry *Registry) {
				registry.RegisterHandler("direct", slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{
					Level: slog.LevelWarn,
				}))
			},
			kind:        "direct",
			format:      FormatJSON,
			wantEnabled: false,
		},
		{
			name: "registered factory is returned",
			setup: func(registry *Registry) {
				registry.RegisterHandlerFactory("factory", func() slog.Handler {
					return slog.NewTextHandler(io.Discard, nil)
				})
			},
			kind:        "factory",
			format:      FormatText,
			wantEnabled: true,
		},
		{
			name: "nil factory falls back to writer",
			setup: func(registry *Registry) {
				registry.RegisterHandlerFactory("writer", func() slog.Handler {
					return nil
				})
				registry.RegisterWriter("writer", func() (io.Writer, *slog.HandlerOptions, error) {
					return new(bytes.Buffer), nil, nil
				})
			},
			kind:        "writer",
			format:      FormatJSON,
			wantEnabled: true,
		},
		{
			name: "writer error is returned",
			setup: func(registry *Registry) {
				registry.RegisterWriter("broken", func() (io.Writer, *slog.HandlerOptions, error) {
					return nil, nil, expectedErr
				})
			},
			kind:    "broken",
			format:  FormatJSON,
			wantErr: true,
		},
		{
			name:    "missing handler returns not found",
			setup:   func(*Registry) {},
			kind:    "missing",
			format:  FormatJSON,
			wantErr: true,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			registry := NewRegistry()
			test.setup(registry)

			handler, err := registry.Get(test.kind, test.format, slog.LevelInfo)
			if test.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}

				return
			}

			if err != nil {
				t.Fatalf("get handler: %v", err)
			}

			if handler == nil {
				t.Fatal("handler is nil")
			}

			if got := handler.Enabled(context.Background(), slog.LevelInfo); got != test.wantEnabled {
				t.Fatalf("enabled = %v, want %v", got, test.wantEnabled)
			}
		})
	}
}

func TestPackageRegistryCompatibility(t *testing.T) {
	cases := []struct {
		name string
	}{
		{name: "package-level wrappers use default registry"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			handle := InstallDefaultRegistry(NewRegistry())
			t.Cleanup(func() {
				if err := handle.Close(); err != nil {
					t.Fatalf("close default registry handle: %v", err)
				}
			})

			RegisterWriter("compat", func() (io.Writer, *slog.HandlerOptions, error) {
				return io.Discard, nil, nil
			})

			handler, err := Get("compat", FormatJSON, slog.LevelInfo)
			if err != nil {
				t.Fatalf("get compatibility handler: %v", err)
			}

			if handler == nil {
				t.Fatal("handler is nil")
			}
		})
	}
}

func TestDefaultRegistryHandle(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "restores previous registry",
			run: func(t *testing.T) {
				t.Helper()

				previous := DefaultRegistry()
				next := NewRegistry()
				handle := InstallDefaultRegistry(next)

				if got := DefaultRegistry(); got != next {
					t.Fatal("default registry was not installed")
				}

				if err := handle.Close(); err != nil {
					t.Fatalf("close default registry handle: %v", err)
				}

				if got := DefaultRegistry(); got != previous {
					t.Fatal("default registry was not restored")
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}
