package stdout

import (
	"log/slog"
	"os"
	"testing"
)

func TestGetConfigFromEnv(t *testing.T) {
	cases := []struct {
		name      string
		level     string
		wantLevel slog.Level
		wantErr   bool
	}{
		{
			name:      "default level",
			wantLevel: slog.LevelInfo,
		},
		{
			name:      "custom level",
			level:     "ERROR",
			wantLevel: slog.LevelError,
		},
		{
			name:    "invalid level",
			level:   "bad",
			wantErr: true,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if test.level == "" {
				unsetEnv(t, "STDOUT_LEVEL")
			} else {
				t.Setenv("STDOUT_LEVEL", test.level)
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

			if cfg.Level != test.wantLevel {
				t.Fatalf("level = %v, want %v", cfg.Level, test.wantLevel)
			}
		})
	}
}

func TestNew(t *testing.T) {
	cases := []struct {
		name     string
		level    string
		wantOpts bool
	}{
		{
			name:     "configured writer",
			level:    "WARN",
			wantOpts: true,
		},
		{
			name:     "invalid config falls back to stdout without options",
			level:    "bad",
			wantOpts: false,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("STDOUT_LEVEL", test.level)

			writer, opts, err := New()
			if err != nil {
				t.Fatalf("new writer: %v", err)
			}

			if writer != os.Stdout {
				t.Fatal("writer is not stdout")
			}

			if got := opts != nil; got != test.wantOpts {
				t.Fatalf("opts present = %v, want %v", got, test.wantOpts)
			}
		})
	}
}

func TestNewFromConfig(t *testing.T) {
	cases := []struct {
		name      string
		cfg       Config
		wantLevel slog.Level
	}{
		{
			name: "explicit config",
			cfg: Config{
				Level: slog.LevelWarn,
			},
			wantLevel: slog.LevelWarn,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			writer, opts, err := NewFromConfig(test.cfg)
			if err != nil {
				t.Fatalf("new from config: %v", err)
			}

			if writer != os.Stdout {
				t.Fatal("writer is not stdout")
			}

			if opts == nil || opts.Level == nil {
				t.Fatal("handler options are incomplete")
			}

			if got := opts.Level.Level(); got != test.wantLevel {
				t.Fatalf("level = %v, want %v", got, test.wantLevel)
			}
		})
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
