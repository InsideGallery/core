package logfile

import (
	"log/slog"
	"os"
	"testing"
)

func TestGetConfigFromEnv(t *testing.T) {
	cases := []struct {
		name      string
		file      string
		level     string
		wantLevel slog.Level
		wantErr   bool
	}{
		{
			name:      "defaults",
			wantLevel: slog.LevelInfo,
		},
		{
			name:      "custom values",
			file:      "/tmp/core-logfile-test.log",
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
			unsetEnv(t, "LOGFILE_NAME")
			unsetEnv(t, "LOGFILE_LEVEL")

			if test.file != "" {
				t.Setenv("LOGFILE_NAME", test.file)
			}

			if test.level != "" {
				t.Setenv("LOGFILE_LEVEL", test.level)
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
		name    string
		level   string
		wantErr bool
	}{
		{
			name:  "opens configured file",
			level: "WARN",
		},
		{
			name:    "invalid level returns error",
			level:   "bad",
			wantErr: true,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			fileName := t.TempDir() + "/log.txt"
			t.Setenv("LOGFILE_NAME", fileName)
			t.Setenv("LOGFILE_LEVEL", test.level)

			writer, opts, err := New()
			if test.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}

				return
			}

			if err != nil {
				t.Fatalf("new logfile writer: %v", err)
			}

			if opts == nil || opts.Level == nil {
				t.Fatal("handler options are incomplete")
			}

			if _, err := writer.Write([]byte("hello")); err != nil {
				t.Fatalf("write log: %v", err)
			}

			if closer, ok := writer.(interface{ Close() error }); ok {
				if err := closer.Close(); err != nil {
					t.Fatalf("close writer: %v", err)
				}
			}
		})
	}
}

func TestNewFromConfig(t *testing.T) {
	cases := []struct {
		name      string
		level     slog.Level
		wantLevel slog.Level
	}{
		{
			name:      "opens explicit file",
			level:     slog.LevelDebug,
			wantLevel: slog.LevelDebug,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			writer, opts, err := NewFromConfig(Config{
				Name:  t.TempDir() + "/log.txt",
				Level: test.level,
			})
			if err != nil {
				t.Fatalf("new from config: %v", err)
			}

			if opts == nil || opts.Level == nil {
				t.Fatal("handler options are incomplete")
			}

			if got := opts.Level.Level(); got != test.wantLevel {
				t.Fatalf("level = %v, want %v", got, test.wantLevel)
			}

			if closer, ok := writer.(interface{ Close() error }); ok {
				if err := closer.Close(); err != nil {
					t.Fatalf("close writer: %v", err)
				}
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
