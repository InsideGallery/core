package redis

import (
	"errors"
	"testing"
)

func TestConnectionStore(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "get or create explicit connection",
			run: func(t *testing.T) {
				t.Helper()

				store := NewConnectionStore(nil)
				client, err := store.GetOrCreate(&ConnectionConfig{
					Host: "localhost",
					Port: "6379",
				})
				if err != nil {
					t.Fatalf("get or create: %v", err)
				}
				defer client.Stop()

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
			name: "missing connection returns sentinel",
			run: func(t *testing.T) {
				t.Helper()

				_, err := NewConnectionStore(nil).Get()
				if !errors.Is(err, ErrConnectionIsNotSet) {
					t.Fatalf("err = %v, want %v", err, ErrConnectionIsNotSet)
				}
			},
		},
		{
			name: "close clears connection",
			run: func(t *testing.T) {
				t.Helper()

				store := NewConnectionStore(NewRedisClient(&ConnectionConfig{
					Host: "localhost",
					Port: "6379",
				}))
				if err := store.Close(); err != nil {
					t.Fatalf("close: %v", err)
				}

				_, err := store.Get()
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

func TestDefaultCompatibility(t *testing.T) {
	cases := []struct {
		name string
	}{
		{name: "get returns package-level client"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			client := NewRedisClient(&ConnectionConfig{
				Host: "localhost",
				Port: "6379",
			})
			defer client.Stop()

			Set(client)
			t.Cleanup(func() {
				Set(nil)
			})

			got, err := Get()
			if err != nil {
				t.Fatalf("get: %v", err)
			}

			if got != client {
				t.Fatal("get did not return package-level client")
			}
		})
	}
}
