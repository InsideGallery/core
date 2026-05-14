package frogodb

import (
	"context"
	"errors"
	"testing"

	fdbclient "github.com/FrogoAI/fdb-client/pkg/client"
)

const (
	testNamespace = "test-ns"
	testSet       = "test-set"
	testKey       = "test-key"
	testBinName   = "count"
)

func TestDatabaseClientRequiresConnection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	database := WrapClient(nil)

	cases := []struct {
		name   string
		assert func(t *testing.T)
	}{
		{
			name: "ping",
			assert: func(t *testing.T) {
				t.Helper()

				err := database.Ping(ctx)
				assertMissingConnection(t, err)
			},
		},
		{
			name: "put",
			assert: func(t *testing.T) {
				t.Helper()

				_, err := database.PutRecord(ctx, PutOptions{})
				assertMissingConnection(t, err)
			},
		},
		{
			name: "get",
			assert: func(t *testing.T) {
				t.Helper()

				_, err := database.GetRecord(ctx, GetOptions{})
				assertMissingConnection(t, err)
			},
		},
		{
			name: "delete",
			assert: func(t *testing.T) {
				t.Helper()

				_, err := database.DeleteRecord(ctx, DeleteOptions{})
				assertMissingConnection(t, err)
			},
		},
		{
			name: "count",
			assert: func(t *testing.T) {
				t.Helper()

				_, err := database.CountRecords(ctx, CountOptions{})
				assertMissingConnection(t, err)
			},
		},
		{
			name: "close",
			assert: func(t *testing.T) {
				t.Helper()

				err := database.Close()
				assertMissingConnection(t, err)
			},
		},
	}

	for _, test := range cases {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			test.assert(t)
		})
	}
}

func TestDatabaseClientNilClientAccessor(t *testing.T) {
	t.Parallel()

	var database *DatabaseClient
	if database.Client() != nil {
		t.Fatal("Client() should return nil for nil receiver")
	}
}

func TestNewRecordCopiesBins(t *testing.T) {
	t.Parallel()

	key := Key{Namespace: testNamespace, Set: testSet, Value: testKey}
	source := &fdbclient.Record{
		Generation: 7,
		Bins: map[string]any{
			testBinName: int64(42),
		},
	}

	record := newRecord(key, source)
	source.Bins[testBinName] = int64(24)

	if record.Key != key {
		t.Fatalf("Key = %#v, want %#v", record.Key, key)
	}

	if record.Generation != source.Generation {
		t.Fatalf("Generation = %d, want %d", record.Generation, source.Generation)
	}

	if record.Bins[testBinName] != int64(42) {
		t.Fatalf("Bins[%q] = %v", testBinName, record.Bins[testBinName])
	}
}

func TestWriteOptionsClientOptions(t *testing.T) {
	t.Parallel()

	options := WriteOptions{
		TTLSeconds:   60,
		Generation:   2,
		MergeBins:    true,
		ReplaceBins:  true,
		CreateOnly:   true,
		Replace:      true,
		PreserveTTL:  true,
		ClearTTL:     true,
		CommitMaster: true,
	}.clientOptions()

	if len(options) != writeOptionCapacity {
		t.Fatalf("clientOptions length = %d, want %d", len(options), writeOptionCapacity)
	}
}

func assertMissingConnection(t *testing.T, err error) {
	t.Helper()

	if !errors.Is(err, ErrConnectionIsNotSet) {
		t.Fatalf("error = %v, want %v", err, ErrConnectionIsNotSet)
	}
}
