package client

import (
	"testing"
	"time"
)

func TestConnectOptionsConfig(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		options ConnectOptions
		want    *Config
	}{
		{
			name: "maps core options to legacy config",
			options: ConnectOptions{
				Addr:                 "nats://127.0.0.1:4222",
				Username:             "user",
				Password:             "pass",
				DrainTimeout:         time.Second,
				MaxReconnects:        1,
				ReconnectWait:        time.Millisecond,
				RetryOnFailedConnect: true,
				ConcurrentSize:       2,
				MaxConcurrentSize:    3,
				ReadTimeout:          time.Second,
				IdleTimeout:          time.Second,
			},
			want: &Config{
				Addr:                 "nats://127.0.0.1:4222",
				Username:             "user",
				Password:             "pass",
				DrainTimeout:         time.Second,
				MaxReconnects:        1,
				ReconnectWait:        time.Millisecond,
				RetryOnFailedConnect: true,
				ConcurrentSize:       2,
				MaxConcurrentSize:    3,
				ReadTimeout:          time.Second,
				IdleTimeout:          time.Second,
			},
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := test.options.config()
			if got.Addr != test.want.Addr {
				t.Fatalf("Addr = %q, want %q", got.Addr, test.want.Addr)
			}

			if got.Username != test.want.Username {
				t.Fatalf("Username = %q, want %q", got.Username, test.want.Username)
			}

			if got.GetConcurrentSize() != test.want.ConcurrentSize {
				t.Fatalf("GetConcurrentSize() = %d, want %d", got.GetConcurrentSize(), test.want.ConcurrentSize)
			}
		})
	}
}
