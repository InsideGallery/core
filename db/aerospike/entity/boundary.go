package entity

import (
	"context"
	"errors"

	aero "github.com/InsideGallery/core/db/aerospike"
	coreerrors "github.com/InsideGallery/core/errors"
)

const boundaryKind = "aerospike entity"

// ErrStoreNotSet reports a nil entity store dependency.
var ErrStoreNotSet = errors.New("entity store is not set")

// Store wraps record-level entity helpers behind core-owned Aerospike contracts.
type Store struct {
	namespace aero.NamespaceStore
	key       aero.Key
}

// BinOptions is the core-owned input for reading one entity bin.
type BinOptions struct {
	Name string
}

// BinResult is the core-owned result for reading one entity bin.
type BinResult struct {
	Found bool
	Value any
}

// ExistsResult is the core-owned result for an entity existence check.
type ExistsResult struct {
	Exists bool
}

// RecordStore is the core-owned contract for common Aerospike entity helpers.
type RecordStore interface {
	Put(ctx context.Context, bins map[string]any) (aero.Result, error)
	Get(ctx context.Context, bins ...string) (aero.RecordResult, error)
	GetBin(ctx context.Context, options BinOptions) (BinResult, error)
	Exists(ctx context.Context) (ExistsResult, error)
	Delete(ctx context.Context) (aero.Result, error)
}

// NewStore creates entity helpers from a core-owned namespace store and key.
func NewStore(namespace aero.NamespaceStore, key aero.Key) *Store {
	return &Store{
		namespace: namespace,
		key:       key,
	}
}

// Put writes entity bins through a core-owned namespace store.
func (s *Store) Put(ctx context.Context, bins map[string]any) (aero.Result, error) {
	if s == nil || s.namespace == nil {
		return aero.Result{}, ErrStoreNotSet
	}

	if ctx == nil {
		ctx = context.Background()
	}

	result, err := s.namespace.PutRecord(ctx, aero.PutOptions{
		Key:  s.key,
		Bins: bins,
	})
	if err != nil {
		return aero.Result{}, coreerrors.WrapBoundary(boundaryKind, "put", err)
	}

	return result, nil
}

// Get reads an entity record through a core-owned namespace store.
func (s *Store) Get(ctx context.Context, bins ...string) (aero.RecordResult, error) {
	if s == nil || s.namespace == nil {
		return aero.RecordResult{}, ErrStoreNotSet
	}

	if ctx == nil {
		ctx = context.Background()
	}

	result, err := s.namespace.GetRecord(ctx, aero.GetOptions{
		Key:      s.key,
		BinNames: bins,
	})
	if err != nil {
		return aero.RecordResult{}, coreerrors.WrapBoundary(boundaryKind, "get", err)
	}

	return result, nil
}

// GetBin reads one entity bin through a core-owned namespace store.
func (s *Store) GetBin(ctx context.Context, options BinOptions) (BinResult, error) {
	result, err := s.Get(ctx, options.Name)
	if err != nil {
		return BinResult{}, err
	}

	if !result.Found {
		return BinResult{}, nil
	}

	value, found := result.Record.Bins[options.Name]

	return BinResult{Found: found, Value: value}, nil
}

// Exists checks whether an entity exists through a core-owned namespace store.
func (s *Store) Exists(ctx context.Context) (ExistsResult, error) {
	result, err := s.Get(ctx)
	if err != nil {
		return ExistsResult{}, err
	}

	return ExistsResult{Exists: result.Found}, nil
}

// Delete deletes an entity through a core-owned namespace store.
func (s *Store) Delete(ctx context.Context) (aero.Result, error) {
	if s == nil || s.namespace == nil {
		return aero.Result{}, ErrStoreNotSet
	}

	if ctx == nil {
		ctx = context.Background()
	}

	result, err := s.namespace.DeleteRecord(ctx, aero.DeleteOptions{Key: s.key})
	if err != nil {
		return aero.Result{}, coreerrors.WrapBoundary(boundaryKind, "delete", err)
	}

	return result, nil
}
