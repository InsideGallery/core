package aerospike

import (
	"errors"
	"testing"

	aero "github.com/aerospike/aerospike-client-go/v7"
)

func TestConnectionRegistry(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "set and get explicit connection",
			run: func(t *testing.T) {
				t.Helper()

				registry := NewConnectionRegistry()
				client := &aero.Client{}

				registry.Set("unit", client)

				got, err := registry.Get("unit")
				if err != nil {
					t.Fatalf("get connection: %v", err)
				}

				if got != client {
					t.Fatal("registry returned a different client")
				}
			},
		},
		{
			name: "missing connection returns sentinel",
			run: func(t *testing.T) {
				t.Helper()

				_, err := NewConnectionRegistry().Get("missing")
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
		{name: "default returns package-level registered client"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			client := &aero.Client{}

			Set("unit-default", client)
			t.Cleanup(func() {
				defaultConnections = NewConnectionRegistry()
			})

			got, err := Default("unit-default")
			if err != nil {
				t.Fatalf("default: %v", err)
			}

			if got != client {
				t.Fatal("default did not return registered client")
			}
		})
	}
}
