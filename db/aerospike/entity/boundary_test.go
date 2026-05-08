package entity

import (
	"context"
	"errors"
	"testing"

	aero "github.com/InsideGallery/core/db/aerospike"
)

func TestStoreBoundary(t *testing.T) {
	t.Parallel()

	store := NewStore(&fakeNamespaceStore{
		record: aero.RecordResult{
			Found: true,
			Record: aero.Record{
				Bins: map[string]any{"name": "inside"},
			},
		},
	}, aero.Key{Set: "users", Value: "1"})

	cases := []struct {
		name   string
		assert func(t *testing.T)
	}{
		{
			name: "store implements entity store",
			assert: func(t *testing.T) {
				t.Helper()

				var _ RecordStore = (*Store)(nil)
			},
		},
		{
			name: "get bin returns core-owned result",
			assert: func(t *testing.T) {
				t.Helper()

				got, err := store.GetBin(context.Background(), BinOptions{Name: "name"})
				if err != nil {
					t.Fatalf("GetBin() error: %v", err)
				}

				if !got.Found {
					t.Fatal("Found = false, want true")
				}

				if got.Value != "inside" {
					t.Fatalf("Value = %v, want inside", got.Value)
				}
			},
		},
		{
			name: "nil store returns stable error",
			assert: func(t *testing.T) {
				t.Helper()

				_, err := (*Store)(nil).Get(context.Background())
				if !errors.Is(err, ErrStoreNotSet) {
					t.Fatalf("err = %v, want %v", err, ErrStoreNotSet)
				}
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

type fakeNamespaceStore struct {
	record aero.RecordResult
}

func (s *fakeNamespaceStore) PutRecord(context.Context, aero.PutOptions) (aero.Result, error) {
	return aero.Result{Affected: 1}, nil
}

func (s *fakeNamespaceStore) GetRecord(context.Context, aero.GetOptions) (aero.RecordResult, error) {
	return s.record, nil
}

func (s *fakeNamespaceStore) DeleteRecord(context.Context, aero.DeleteOptions) (aero.Result, error) {
	return aero.Result{Affected: 1, Deleted: true}, nil
}
