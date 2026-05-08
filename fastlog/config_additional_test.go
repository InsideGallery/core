package fastlog

import (
	"context"
	"io"
	"log/slog"
	"os"
	"reflect"
	"testing"

	"github.com/InsideGallery/core/fastlog/handlers"
)

func TestGetConfigFromEnv(t *testing.T) {
	cases := []struct {
		name      string
		env       map[string]string
		unset     []string
		want      *Config
		wantErr   bool
		wantLevel slog.Level
	}{
		{
			name:  "defaults",
			unset: []string{"LOG_OUTPUTS", "LOG_LEVEL", "LOG_CALLER", "LOG_ERROR_FORMATING"},
			want: &Config{
				Outputs: []string{"stderr:json"},
				Level:   slog.LevelInfo,
				Caller:  true,
			},
		},
		{
			name: "custom values",
			env: map[string]string{
				"LOG_OUTPUTS":         "stdout:text,nop:json",
				"LOG_LEVEL":           "DEBUG",
				"LOG_CALLER":          "false",
				"LOG_ERROR_FORMATING": "true",
			},
			want: &Config{
				Outputs:         []string{"stdout:text", "nop:json"},
				Level:           slog.LevelDebug,
				Caller:          false,
				ErrorFormatting: true,
			},
		},
		{
			name: "invalid level",
			env: map[string]string{
				"LOG_LEVEL": "bad",
			},
			wantErr: true,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			for _, key := range test.unset {
				unsetEnv(t, key)
			}

			for key, value := range test.env {
				t.Setenv(key, value)
			}

			cfg, err := GetConfigFromEnv()
			if test.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}

				return
			}

			if err != nil {
				t.Fatalf("config from env: %v", err)
			}

			if !reflect.DeepEqual(cfg, test.want) {
				t.Fatalf("config = %#v, want %#v", cfg, test.want)
			}
		})
	}
}

func TestConfigGetHandler(t *testing.T) {
	cases := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "invalid output shape falls back to nop",
			cfg: Config{
				Outputs: []string{"broken"},
				Level:   slog.LevelInfo,
			},
		},
		{
			name: "missing output returns combined error with fallback",
			cfg: Config{
				Outputs: []string{"missing:json"},
				Level:   slog.LevelInfo,
			},
			wantErr: true,
		},
		{
			name: "registered output returns handler",
			cfg: Config{
				Outputs: []string{"stdout:text"},
				Level:   slog.LevelInfo,
			},
		},
		{
			name: "default stderr output returns handler",
			cfg: Config{
				Outputs: []string{"stderr:json"},
				Level:   slog.LevelInfo,
			},
		},
		{
			name: "configured middleware is installed",
			cfg: Config{
				Outputs:         []string{"stdout:json"},
				Level:           slog.LevelInfo,
				Caller:          true,
				ErrorFormatting: true,
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			handler, err := test.cfg.GetHandler()
			if test.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
			} else if err != nil {
				t.Fatalf("get handler: %v", err)
			}

			if handler == nil {
				t.Fatal("handler is nil")
			}

			if !handler.Enabled(context.Background(), slog.LevelInfo) {
				t.Fatal("handler is not enabled for info level")
			}
		})
	}
}

func TestConfigGetHandlerFromRegistry(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "explicit registry resolves output",
			run: func(t *testing.T) {
				t.Helper()

				registry := handlers.NewRegistry()
				registry.RegisterWriter("scoped", func() (io.Writer, *slog.HandlerOptions, error) {
					return io.Discard, nil, nil
				})

				cfg := Config{
					Outputs: []string{"scoped:json"},
					Level:   slog.LevelInfo,
				}

				handler, err := cfg.GetHandlerFromRegistry(registry)
				if err != nil {
					t.Fatalf("get handler from registry: %v", err)
				}

				if handler == nil {
					t.Fatal("handler is nil")
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()

	oldValue, exists := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset %s: %v", key, err)
	}

	t.Cleanup(func() {
		if exists {
			if err := os.Setenv(key, oldValue); err != nil {
				t.Fatalf("restore %s: %v", key, err)
			}

			return
		}

		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("cleanup %s: %v", key, err)
		}
	})
}
