// Package postgres provides Postgres connection and database helpers.
//
// New code should use explicit construction and core-owned SQL operation types:
//
//	import "github.com/InsideGallery/core/db/postgres"
//
//	db, err := postgres.NewDatabaseFromOptions(postgres.DatabaseOptions{
//		Host: "localhost",
//		Database: "app",
//	})
//
// Prefer Database, DatabaseClient, Statement, DatabaseOptions, and CommandResult
// when consumer code should not depend directly on sqlx.
//
// Compatibility: NewClient, Set, Get, and Default remain available for existing
// sqlx-shaped callers. Prefer NewDatabase, NewDatabaseFromOptions, DefaultDatabase,
// or NewClientStore with explicit lifecycle ownership in new code.
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	coreerrors "github.com/InsideGallery/core/errors"
)

// Statement is the core-owned input for SQL operations.
type Statement struct {
	Query string
	Args  []any
}

// DatabaseOptions is the core-owned input for constructing a Postgres database client.
type DatabaseOptions struct {
	Host            string
	Port            string
	User            string
	Password        string
	Database        string
	ApplicationName string
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}

// CommandResult is the core-owned result for SQL commands.
type CommandResult struct {
	RowsAffected int64
}

// Database is the core-owned Postgres contract for new consumers.
type Database interface {
	Ping(ctx context.Context) error
	Exec(ctx context.Context, statement Statement) (CommandResult, error)
	Query(ctx context.Context, statement Statement) (*sql.Rows, error)
	QueryRow(ctx context.Context, statement Statement) *sql.Row
	Close() error
}

// DatabaseClient wraps sqlx behind core-owned operation inputs and results.
type DatabaseClient struct {
	db *sqlx.DB
}

// NewDatabase creates a Postgres database client from core-owned config.
func NewDatabase(config *ConnectionConfig) (*DatabaseClient, error) {
	db, err := NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("postgres open: %w", err)
	}

	return &DatabaseClient{db: db}, nil
}

// NewDatabaseFromOptions creates a Postgres database client from core-owned construction options.
func NewDatabaseFromOptions(options DatabaseOptions) (*DatabaseClient, error) {
	return NewDatabase(options.config())
}

// DefaultDatabase returns the default Postgres client behind the core-owned DatabaseClient API.
func DefaultDatabase() (*DatabaseClient, error) {
	db, err := Default()
	if err != nil {
		return nil, err
	}

	return WrapDatabase(db), nil
}

// WrapDatabase adapts an existing sqlx DB to DatabaseClient.
func WrapDatabase(db *sqlx.DB) *DatabaseClient {
	return &DatabaseClient{db: db}
}

// SQLDB returns the standard-library database handle.
func (d *DatabaseClient) SQLDB() *sql.DB {
	return d.db.DB
}

// Ping verifies connectivity.
func (d *DatabaseClient) Ping(ctx context.Context) error {
	if err := d.db.PingContext(ctx); err != nil {
		return coreerrors.WrapBoundary("postgres", "ping", err)
	}

	return nil
}

// Exec runs a command with core-owned options.
func (d *DatabaseClient) Exec(ctx context.Context, statement Statement) (CommandResult, error) {
	result, err := d.db.ExecContext(ctx, statement.Query, statement.Args...)
	if err != nil {
		return CommandResult{}, coreerrors.WrapBoundary("postgres", "exec", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return CommandResult{}, coreerrors.WrapBoundary("postgres", "rows affected", err)
	}

	return CommandResult{RowsAffected: rowsAffected}, nil
}

// Query runs a query with core-owned options.
func (d *DatabaseClient) Query(ctx context.Context, statement Statement) (*sql.Rows, error) {
	rows, err := d.db.QueryContext(ctx, statement.Query, statement.Args...)
	if err != nil {
		return nil, coreerrors.WrapBoundary("postgres", "query", err)
	}

	return rows, nil
}

// QueryRow runs a single-row query with core-owned options.
func (d *DatabaseClient) QueryRow(ctx context.Context, statement Statement) *sql.Row {
	return d.db.QueryRowContext(ctx, statement.Query, statement.Args...)
}

// Close closes the database client.
func (d *DatabaseClient) Close() error {
	if err := d.db.Close(); err != nil {
		return coreerrors.WrapBoundary("postgres", "close", err)
	}

	return nil
}

func (o DatabaseOptions) config() *ConnectionConfig {
	return &ConnectionConfig{
		Host:            o.Host,
		Port:            o.Port,
		User:            o.User,
		Password:        o.Password,
		DB:              o.Database,
		ApplicationName: o.ApplicationName,
		MaxOpenConns:    o.MaxOpenConns,
		ConnMaxLifetime: int64(o.ConnMaxLifetime),
	}
}
