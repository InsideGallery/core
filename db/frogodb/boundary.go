// Package frogodb provides FrogoDB connection and record helpers.
//
// New code should use explicit connection ownership and core-owned operation
// shapes:
//
//	import "github.com/InsideGallery/core/db/frogodb"
//
//	store, err := frogodb.NewDatabase(frogodb.DefaultConnectionConfig("localhost:3000"))
//
// Prefer Database, DatabaseClient, Key, PutOptions, GetOptions, DeleteOptions,
// Record, RecordResult, and Result when consumer code should not depend
// directly on the FrogoDB client package.
//
// Compatibility: NewConnection, NewConnectionFromEnv, Set, Get, and Default
// expose the FrogoDB SDK-shaped client for advanced callers and existing
// low-level integration points.
package frogodb

import (
	"context"
	"errors"

	fdbclient "github.com/FrogoAI/fdb-client/pkg/client"

	coreerrors "github.com/InsideGallery/core/errors"
)

const writeOptionCapacity = 9

// Key is a core-owned FrogoDB record identity.
type Key struct {
	Namespace string
	Set       string
	Value     string
}

// WriteOptions is a core-owned subset of FrogoDB write behavior.
type WriteOptions struct {
	TTLSeconds   uint32
	Generation   uint32
	MergeBins    bool
	ReplaceBins  bool
	CreateOnly   bool
	Replace      bool
	PreserveTTL  bool
	ClearTTL     bool
	CommitMaster bool
}

// PutOptions is the core-owned input for writing a record.
type PutOptions struct {
	Key   Key
	Bins  map[string]any
	Write WriteOptions
}

// GetOptions is the core-owned input for reading a record.
type GetOptions struct {
	Key      Key
	BinNames []string
}

// DeleteOptions is the core-owned input for deleting a record.
type DeleteOptions struct {
	Key   Key
	Write WriteOptions
}

// CountOptions is the core-owned input for counting records.
type CountOptions struct {
	Namespace string
	Set       string
	AllNodes  bool
}

// Record is the core-owned FrogoDB record result.
type Record struct {
	Key        Key
	Bins       map[string]any
	Generation uint32
}

// RecordResult reports a record lookup result.
type RecordResult struct {
	Found  bool
	Record Record
}

// Result reports a write/delete/count operation result.
type Result struct {
	Affected int64
	Deleted  bool
}

// Database is the core-owned FrogoDB contract for new consumers.
type Database interface {
	Ping(ctx context.Context) error
	PutRecord(ctx context.Context, options PutOptions) (Result, error)
	GetRecord(ctx context.Context, options GetOptions) (RecordResult, error)
	DeleteRecord(ctx context.Context, options DeleteOptions) (Result, error)
	CountRecords(ctx context.Context, options CountOptions) (Result, error)
	Close() error
}

// DatabaseClient wraps the FrogoDB smart client behind core-owned operation inputs and results.
type DatabaseClient struct {
	client *fdbclient.Client
}

// NewDatabase creates a FrogoDB database client from core-owned config.
func NewDatabase(config *ConnectionConfig) (*DatabaseClient, error) {
	client, err := NewConnection(config)
	if err != nil {
		return nil, err
	}

	return WrapClient(client), nil
}

// DefaultDatabase returns the default FrogoDB client behind the core-owned DatabaseClient API.
func DefaultDatabase(names ...string) (*DatabaseClient, error) {
	client, err := Default(names...)
	if err != nil {
		return nil, err
	}

	return WrapClient(client), nil
}

// WrapClient adapts an existing FrogoDB client to DatabaseClient.
func WrapClient(client *fdbclient.Client) *DatabaseClient {
	return &DatabaseClient{client: client}
}

// Client returns the wrapped FrogoDB smart client for advanced operations.
func (c *DatabaseClient) Client() *fdbclient.Client {
	if c == nil {
		return nil
	}

	return c.client
}

// Ping verifies FrogoDB connectivity.
func (c *DatabaseClient) Ping(ctx context.Context) error {
	client, err := c.requireClient()
	if err != nil {
		return err
	}

	if err := client.Ping(ctx); err != nil {
		return coreerrors.WrapBoundary("frogodb", "ping", err)
	}

	return nil
}

// PutRecord writes a record through a core-owned input and result shape.
func (c *DatabaseClient) PutRecord(ctx context.Context, options PutOptions) (Result, error) {
	client, err := c.requireClient()
	if err != nil {
		return Result{}, err
	}

	if err := client.Put(
		ctx,
		options.Key.Namespace,
		options.Key.Set,
		options.Key.Value,
		options.Bins,
		options.Write.clientOptions()...,
	); err != nil {
		return Result{}, coreerrors.WrapBoundary("frogodb", "put record", err)
	}

	return Result{Affected: 1}, nil
}

// GetRecord reads a record through a core-owned input and result shape.
func (c *DatabaseClient) GetRecord(ctx context.Context, options GetOptions) (RecordResult, error) {
	client, err := c.requireClient()
	if err != nil {
		return RecordResult{}, err
	}

	record, err := client.Get(
		ctx,
		options.Key.Namespace,
		options.Key.Set,
		options.Key.Value,
		options.BinNames...,
	)
	if errors.Is(err, fdbclient.ErrKeyNotFound) {
		return RecordResult{}, nil
	}

	if err != nil {
		return RecordResult{}, coreerrors.WrapBoundary("frogodb", "get record", err)
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
func (c *DatabaseClient) DeleteRecord(ctx context.Context, options DeleteOptions) (Result, error) {
	client, err := c.requireClient()
	if err != nil {
		return Result{}, err
	}

	deleted, err := client.Delete(
		ctx,
		options.Key.Namespace,
		options.Key.Set,
		options.Key.Value,
		options.Write.clientOptions()...,
	)
	if err != nil {
		return Result{}, coreerrors.WrapBoundary("frogodb", "delete record", err)
	}

	if !deleted {
		return Result{}, nil
	}

	return Result{Affected: 1, Deleted: true}, nil
}

// CountRecords counts records through a core-owned input and result shape.
func (c *DatabaseClient) CountRecords(ctx context.Context, options CountOptions) (Result, error) {
	client, err := c.requireClient()
	if err != nil {
		return Result{}, err
	}

	count, err := countRecords(ctx, client, options)
	if err != nil {
		return Result{}, coreerrors.WrapBoundary("frogodb", "count records", err)
	}

	return Result{Affected: count}, nil
}

// Close closes the FrogoDB client.
func (c *DatabaseClient) Close() error {
	client, err := c.requireClient()
	if err != nil {
		return err
	}

	if err := client.Close(); err != nil {
		return coreerrors.WrapBoundary("frogodb", "close", err)
	}

	return nil
}

func (c *DatabaseClient) requireClient() (*fdbclient.Client, error) {
	if c == nil || c.client == nil {
		return nil, ErrConnectionIsNotSet
	}

	return c.client, nil
}

func countRecords(ctx context.Context, client *fdbclient.Client, options CountOptions) (int64, error) {
	if options.AllNodes {
		return client.CountAll(ctx, options.Namespace, options.Set)
	}

	return client.Count(ctx, options.Namespace, options.Set)
}

func newRecord(key Key, record *fdbclient.Record) Record {
	bins := make(map[string]any, len(record.Bins))
	for name, value := range record.Bins {
		bins[name] = value
	}

	return Record{
		Key:        key,
		Bins:       bins,
		Generation: record.Generation,
	}
}

func (o WriteOptions) clientOptions() []fdbclient.WriteOption {
	options := make([]fdbclient.WriteOption, 0, writeOptionCapacity)

	if o.MergeBins {
		options = append(options, fdbclient.WithMergeBins())
	}

	if o.ReplaceBins {
		options = append(options, fdbclient.WithReplaceBins())
	}

	if o.CreateOnly {
		options = append(options, fdbclient.WithCreateOnly())
	}

	if o.Replace {
		options = append(options, fdbclient.WithReplace())
	}

	if o.PreserveTTL {
		options = append(options, fdbclient.WithPreserveTTL())
	}

	if o.ClearTTL {
		options = append(options, fdbclient.WithClearTTL())
	}

	if o.TTLSeconds > 0 {
		options = append(options, fdbclient.WithTTL(o.TTLSeconds))
	}

	if o.Generation > 0 {
		options = append(options, fdbclient.WithGeneration(o.Generation))
	}

	if o.CommitMaster {
		options = append(options, fdbclient.WithCommitMaster())
	}

	return options
}
