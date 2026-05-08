package fastlog

import (
	"context"
	"log/slog"
	"testing"

	slogmulti "github.com/samber/slog-multi"
)

func TestNewLogger(t *testing.T) {
	cases := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "explicit config creates logger",
			cfg: &Config{
				Outputs: []string{"nop:json"},
				Level:   slog.LevelInfo,
			},
		},
		{
			name: "default stderr config creates logger",
			cfg: &Config{
				Outputs: []string{"stderr:json"},
				Level:   slog.LevelInfo,
			},
		},
		{
			name:    "missing config returns error",
			wantErr: true,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			logger, err := NewLogger(test.cfg)
			if test.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}

				return
			}

			if err != nil {
				t.Fatalf("new logger: %v", err)
			}

			if logger == nil {
				t.Fatal("logger is nil")
			}
		})
	}
}

func TestSetupDefaultLogger(t *testing.T) {
	cases := []struct {
		name string
		cfg  *Config
	}{
		{
			name: "installs and restores default logger",
			cfg: &Config{
				Outputs: []string{"nop:json"},
				Level:   slog.LevelInfo,
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			previous := slog.Default()

			handle, err := SetupDefaultLogger(test.cfg)
			if err != nil {
				t.Fatalf("setup default logger: %v", err)
			}

			if slog.Default() == previous {
				t.Fatal("default logger was not replaced")
			}

			if err := handle.Close(); err != nil {
				t.Fatalf("close default logger handle: %v", err)
			}

			if slog.Default() != previous {
				t.Fatal("default logger was not restored")
			}
		})
	}
}

func TestSetupDefault(t *testing.T) {
	cases := []struct {
		name          string
		cfg           *Config
		withCloseHook bool
		wantErr       bool
	}{
		{
			name:    "missing config returns error",
			wantErr: true,
		},
		{
			name: "installs and restores default logger",
			cfg: &Config{
				Outputs: []string{"nop:json"},
				Level:   slog.LevelInfo,
			},
		},
		{
			name: "close flushes handler hook",
			cfg: &Config{
				Outputs: []string{"nop:json"},
				Level:   slog.LevelInfo,
			},
			withCloseHook: true,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			previous := slog.Default()
			closed := false
			middlewares := closeTrackingMiddlewares(test.withCloseHook, &closed)

			closeDefault, err := SetupDefault(context.Background(), test.cfg, middlewares...)
			if test.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}

				if slog.Default() != previous {
					t.Fatal("default logger was changed")
				}

				return
			}

			if err != nil {
				t.Fatalf("setup default: %v", err)
			}

			if closeDefault == nil {
				t.Fatal("close default is nil")
			}

			if slog.Default() == previous {
				t.Fatal("default logger was not replaced")
			}

			if err := closeDefault(); err != nil {
				t.Fatalf("close default: %v", err)
			}

			if slog.Default() != previous {
				t.Fatal("default logger was not restored")
			}

			if test.withCloseHook && !closed {
				t.Fatal("handler close hook was not called")
			}
		})
	}
}

type closeTrackingHandler struct {
	slog.Handler
	closed *bool
}

func (h closeTrackingHandler) Close() error {
	*h.closed = true

	return nil
}

func closeTrackingMiddlewares(enabled bool, closed *bool) []slogmulti.Middleware {
	if !enabled {
		return nil
	}

	return []slogmulti.Middleware{
		func(next slog.Handler) slog.Handler {
			return closeTrackingHandler{
				Handler: next,
				closed:  closed,
			}
		},
	}
}
