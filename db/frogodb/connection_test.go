package frogodb

import (
	"errors"
	"testing"
)

const testConnectionName = "unit"

func TestConnectionRegistry(t *testing.T) {
	t.Parallel()

	registry := NewConnectionRegistry()

	_, err := registry.Get(testConnectionName)
	if !errors.Is(err, ErrConnectionIsNotSet) {
		t.Fatalf("Get() error = %v, want %v", err, ErrConnectionIsNotSet)
	}

	registry.Set(testConnectionName, nil)

	client, err := registry.Get(testConnectionName)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}

	if client != nil {
		t.Fatal("Get() client should be nil")
	}

	if err := registry.Close(); err != nil {
		t.Fatalf("Close() error: %v", err)
	}

	_, err = registry.Get(testConnectionName)
	if !errors.Is(err, ErrConnectionIsNotSet) {
		t.Fatalf("Get() after Close() error = %v, want %v", err, ErrConnectionIsNotSet)
	}
}

func TestNewConnectionRejectsNilConfig(t *testing.T) {
	t.Parallel()

	_, err := NewConnection(nil)
	if !errors.Is(err, ErrConnectionConfigIsNotSet) {
		t.Fatalf("NewConnection(nil) error = %v, want %v", err, ErrConnectionConfigIsNotSet)
	}
}
