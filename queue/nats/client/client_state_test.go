package client

import (
	"context"
	"testing"
)

func TestConnectClient(t *testing.T) {
	cases := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "invalid address returns error",
			cfg: &Config{
				Addr: "://bad",
			},
			wantErr: true,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			client, err := ConnectClient(context.Background(), test.cfg, StubLogger{})
			if test.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}

				return
			}

			if err != nil {
				t.Fatalf("connect client: %v", err)
			}
			defer client.Close()
		})
	}
}

func TestDefaultCompatibility(t *testing.T) {
	cases := []struct {
		name    string
		addr    string
		wantErr bool
	}{
		{
			name:    "invalid env address returns error",
			addr:    "://bad",
			wantErr: true,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("NATS_ADDR", test.addr)

			client, err := Default(context.Background(), StubLogger{})
			if test.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}

				return
			}

			if err != nil {
				t.Fatalf("default: %v", err)
			}
			defer client.Close()
		})
	}
}
