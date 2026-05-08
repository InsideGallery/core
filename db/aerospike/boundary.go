// Package aerospike provides Aerospike client and namespace helpers.
//
// New code should depend on core-owned operation contracts where possible:
//
//	import "github.com/InsideGallery/core/db/aerospike"
//
// Use NamespaceStore with Key, PutOptions, GetOptions, DeleteOptions, Record,
// RecordResult, and Result for record operations that should not expose
// Aerospike SDK types. Use NewConnection or ConnectionRegistry.Default when the
// consuming application owns client lifecycle explicitly.
//
// Compatibility: the legacy Aerospike and Namespace interfaces, NamespaceInstance
// SDK-shaped methods, and package-level Default connection helper remain
// available for existing consumers.
package aerospike

import (
	"context"

	aero "github.com/aerospike/aerospike-client-go/v7"

	coreerrors "github.com/InsideGallery/core/errors"
)

// Key is a core-owned Aerospike record identity.
type Key struct {
	Set   string
	Value any
}

// PutOptions is the core-owned input for writing a record.
type PutOptions struct {
	Key  Key
	Bins map[string]any
}

// GetOptions is the core-owned input for reading a record.
type GetOptions struct {
	Key      Key
	BinNames []string
}

// DeleteOptions is the core-owned input for deleting a record.
type DeleteOptions struct {
	Key Key
}

// Record is the core-owned Aerospike record result.
type Record struct {
	Key        Key
	Bins       map[string]any
	Generation uint32
	Expiration uint32
}

// RecordResult reports a record lookup result.
type RecordResult struct {
	Found  bool
	Record Record
}

// Result reports a write/delete operation result.
type Result struct {
	Affected int64
	Deleted  bool
}

// NamespaceStore is the core-owned namespace contract for new consumers.
type NamespaceStore interface {
	PutRecord(ctx context.Context, options PutOptions) (Result, error)
	GetRecord(ctx context.Context, options GetOptions) (RecordResult, error)
	DeleteRecord(ctx context.Context, options DeleteOptions) (Result, error)
}

// PutRecord writes a record through a core-owned input and result shape.
func (ni *NamespaceInstance) PutRecord(ctx context.Context, options PutOptions) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, coreerrors.WrapBoundary("aerospike", "put record", err)
	}

	bins := make(aero.BinMap, len(options.Bins))
	for name, value := range options.Bins {
		bins[name] = value
	}

	if err := ni.Put(nil, options.Key.Set, options.Key.Value, bins); err != nil {
		return Result{}, coreerrors.WrapBoundary("aerospike", "put record", err)
	}

	return Result{Affected: 1}, nil
}

// GetRecord reads a record through a core-owned input and result shape.
func (ni *NamespaceInstance) GetRecord(ctx context.Context, options GetOptions) (RecordResult, error) {
	if err := ctx.Err(); err != nil {
		return RecordResult{}, coreerrors.WrapBoundary("aerospike", "get record", err)
	}

	record, err := ni.Get(nil, options.Key.Set, options.Key.Value, options.BinNames...)
	if err != nil {
		return RecordResult{}, coreerrors.WrapBoundary("aerospike", "get record", err)
	}

	if record == nil {
		return RecordResult{}, nil
	}

	return RecordResult{
		Found:  true,
		Record: newRecord(options.Key, record),
	}, nil
}

// DeleteRecord deletes a record through a core-owned input and result shape.
func (ni *NamespaceInstance) DeleteRecord(ctx context.Context, options DeleteOptions) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, coreerrors.WrapBoundary("aerospike", "delete record", err)
	}

	deleted, err := ni.Delete(nil, options.Key.Set, options.Key.Value)
	if err != nil {
		return Result{}, coreerrors.WrapBoundary("aerospike", "delete record", err)
	}

	if !deleted {
		return Result{}, nil
	}

	return Result{Affected: 1, Deleted: true}, nil
}

func newRecord(key Key, record *aero.Record) Record {
	bins := make(map[string]any, len(record.Bins))
	for name, value := range record.Bins {
		bins[name] = value
	}

	return Record{
		Key:        key,
		Bins:       bins,
		Generation: record.Generation,
		Expiration: record.Expiration,
	}
}
