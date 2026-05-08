package mongodb

import (
	"context"
	"errors"
	"testing"
)

func TestClientStore(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "set and get explicit client",
			run: func(t *testing.T) {
				t.Helper()

				store := NewClientStore(nil)
				client := &MongoClient{}

				store.Set(client)

				got, err := store.Get()
				if err != nil {
					t.Fatalf("get client: %v", err)
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
		{
			name: "close clears nil sdk client",
			run: func(t *testing.T) {
				t.Helper()

				store := NewClientStore(&MongoClient{})
				if err := store.Close(context.Background()); err != nil {
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

func TestDefaultClientCompatibility(t *testing.T) {
	cases := []struct {
		name string
	}{
		{name: "get returns package-level client"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			client := &MongoClient{}
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
