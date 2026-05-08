package postgres

import (
	"errors"
	"testing"
)

func TestClientStore(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "get or create explicit client",
			run: func(t *testing.T) {
				t.Helper()

				store := NewClientStore(nil)
				client, err := store.GetOrCreate(&ConnectionConfig{
					Host:            "localhost",
					Port:            "5432",
					User:            "user",
					Password:        "pass",
					DB:              "db",
					MaxOpenConns:    2,
					ConnMaxLifetime: 1,
				})
				if err != nil {
					t.Fatalf("get or create: %v", err)
				}
				defer client.Close()

				got, err := store.Get()
				if err != nil {
					t.Fatalf("get: %v", err)
				}

				if got != client {
					t.Fatal("store returned a different client")
				}
			},
		},
		{
			name: "missing client returns sentinel",
			run: func(t *testing.T) {
				t.Helper()

				_, err := NewClientStore(nil).Get()
				if !errors.Is(err, ErrConnectionIsNotSet) {
					t.Fatalf("err = %v, want %v", err, ErrConnectionIsNotSet)
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}

func TestNewClient(t *testing.T) {
	cases := []struct {
		name string
		cfg  *ConnectionConfig
	}{
		{
			name: "opens sql handle without connecting",
			cfg: &ConnectionConfig{
				Host:            "localhost",
				Port:            "5432",
				User:            "user",
				Password:        "pass",
				DB:              "db",
				MaxOpenConns:    4,
				ConnMaxLifetime: 1,
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			client, err := NewClient(test.cfg)
			if err != nil {
				t.Fatalf("new client: %v", err)
			}
			defer client.Close()

			if got := client.Stats().MaxOpenConnections; got != test.cfg.MaxOpenConns {
				t.Fatalf("max open connections = %d, want %d", got, test.cfg.MaxOpenConns)
			}
		})
	}
}
