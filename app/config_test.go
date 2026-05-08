package app

import (
	"reflect"
	"testing"
	"time"

	"github.com/InsideGallery/core/fastlog"
	"github.com/InsideGallery/core/metrics"
	"github.com/InsideGallery/core/queue/nats/client"
)

func TestConfigFromEnv(t *testing.T) {
	cases := []struct {
		name   string
		env    map[string]string
		assert func(t *testing.T, cfg *Config)
	}{
		{
			name: "parses app and dependency config",
			env: map[string]string{
				"SERVICE_NAME":           "service-unit",
				"SERVER_NAME":            "web-unit",
				"NATS_SERVICE_NAME":      "worker-unit",
				"PORT":                   "127.0.0.1:0",
				"MONITOR_ADDR":           ":9191",
				"DEPLOYMENT_ENVIRONMENT": "production",
				"SHUTDOWN_TIMEOUT":       "3s",
				"LOG_OUTPUTS":            "nop:json",
				"LOG_LEVEL":              "DEBUG",
				"LOG_CALLER":             "false",
				"LOG_ERROR_FORMATING":    "true",
				"METRICS_PROCESSORS":     "disabled",
				"NATS_ADDR":              "nats://127.0.0.1:4223",
				"NATS_DRAIN_TIMEOUT":     "2s",
				"NATS_CONCURRENT_SIZE":   "7",
			},
			assert: func(t *testing.T, cfg *Config) {
				t.Helper()

				wantLog, err := fastlog.GetConfigFromEnv()
				if err != nil {
					t.Fatalf("log config: %v", err)
				}

				if !reflect.DeepEqual(cfg.Log, wantLog) {
					t.Fatalf("log config = %#v, want %#v", cfg.Log, wantLog)
				}

				wantMetrics, err := metrics.GetEnvConfig()
				if err != nil {
					t.Fatalf("metrics config: %v", err)
				}

				if !reflect.DeepEqual(*cfg.Metrics, wantMetrics) {
					t.Fatalf("metrics config = %#v, want %#v", *cfg.Metrics, wantMetrics)
				}

				wantNATS, err := client.GetNATSConnectionConfigFromEnv()
				if err != nil {
					t.Fatalf("nats config: %v", err)
				}

				if !reflect.DeepEqual(cfg.NATS, wantNATS) {
					t.Fatalf("nats config = %#v, want %#v", cfg.NATS, wantNATS)
				}

				if cfg.ServiceName != "service-unit" {
					t.Fatalf("service name = %q, want service-unit", cfg.ServiceName)
				}

				if cfg.ServerName != "web-unit" {
					t.Fatalf("server name = %q, want web-unit", cfg.ServerName)
				}

				if cfg.NATSServiceName != "worker-unit" {
					t.Fatalf("nats service name = %q, want worker-unit", cfg.NATSServiceName)
				}

				if cfg.Port != "127.0.0.1:0" {
					t.Fatalf("port = %q, want 127.0.0.1:0", cfg.Port)
				}

				if cfg.MonitorAddr != ":9191" {
					t.Fatalf("monitor addr = %q, want :9191", cfg.MonitorAddr)
				}

				if cfg.ShutdownTimeout != 3*time.Second {
					t.Fatalf("shutdown timeout = %v, want 3s", cfg.ShutdownTimeout)
				}

				if cfg.EnablePrintRoutes {
					t.Fatal("print routes enabled, want disabled for deployment environment")
				}
			},
		},
		{
			name: "empty deployment environment enables route printing",
			env: map[string]string{
				"DEPLOYMENT_ENVIRONMENT": "",
				"LOG_OUTPUTS":            "nop:json",
				"METRICS_PROCESSORS":     "disabled",
			},
			assert: func(t *testing.T, cfg *Config) {
				t.Helper()

				if !cfg.EnablePrintRoutes {
					t.Fatal("print routes disabled, want enabled")
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			setConfigTestEnv(t)

			for key, value := range test.env {
				t.Setenv(key, value)
			}

			cfg, err := ConfigFromEnv()
			if err != nil {
				t.Fatalf("config from env: %v", err)
			}

			test.assert(t, cfg)
		})
	}
}

func TestConfigFromEnvReturnsAppParseError(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
	}{
		{
			name: "invalid shutdown timeout",
			env: map[string]string{
				"SHUTDOWN_TIMEOUT": "bad",
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			setConfigTestEnv(t)

			for key, value := range test.env {
				t.Setenv(key, value)
			}

			if _, err := ConfigFromEnv(); err == nil {
				t.Fatal("config from env error = nil, want error")
			}
		})
	}
}

func setConfigTestEnv(t *testing.T) {
	t.Helper()

	defaults := map[string]string{
		"SERVICE_NAME":           "",
		"SERVER_NAME":            "",
		"NATS_SERVICE_NAME":      "",
		"PORT":                   ":8080",
		"MONITOR_ADDR":           "",
		"DEPLOYMENT_ENVIRONMENT": "",
		"SHUTDOWN_TIMEOUT":       "10s",
		"LOG_OUTPUTS":            "nop:json",
		"LOG_LEVEL":              "INFO",
		"LOG_CALLER":             "false",
		"LOG_ERROR_FORMATING":    "false",
		"METRICS_PROCESSORS":     "disabled",
		"NATS_ADDR":              "nats://127.0.0.1:4222",
		"NATS_DRAIN_TIMEOUT":     "1s",
		"NATS_CONCURRENT_SIZE":   "10",
	}

	for key, value := range defaults {
		t.Setenv(key, value)
	}
}
