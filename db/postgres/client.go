package postgres

import (
	"errors"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5"        //nolint:revive
	_ "github.com/jackc/pgx/v5/stdlib" //nolint:revive

	"github.com/jmoiron/sqlx"
)

var defaultStore = NewClientStore(nil) //nolint:gochecknoglobals // compatibility store

// ClientStore owns a Postgres sqlx client for explicit application composition.
type ClientStore struct {
	mu     sync.RWMutex
	client *sqlx.DB
}

// NewClientStore creates a Postgres client store with an optional existing client.
func NewClientStore(client *sqlx.DB) *ClientStore {
	return &ClientStore{
		client: client,
	}
}

// Set stores a Postgres client in this store.
func (s *ClientStore) Set(client *sqlx.DB) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.client = client
}

// Get returns the Postgres client from this store.
func (s *ClientStore) Get() (*sqlx.DB, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.client == nil {
		return nil, ErrConnectionIsNotSet
	}

	return s.client, nil
}

// GetOrCreate returns or creates a Postgres client from explicit config.
func (s *ClientStore) GetOrCreate(config *ConnectionConfig) (*sqlx.DB, error) {
	client, err := s.Get()
	if err == nil {
		return client, nil
	}

	if !errors.Is(err, ErrConnectionIsNotSet) {
		return nil, err
	}

	client, err = NewClient(config)
	if err != nil {
		return nil, err
	}

	s.Set(client)

	return client, nil
}

// Close closes the stored Postgres client and clears this store.
func (s *ClientStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return nil
	}

	err := s.client.Close()
	s.client = nil

	return err
}

// NewClient creates a legacy sqlx Postgres client from explicit config.
func NewClient(config *ConnectionConfig) (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", config.GetDSN())
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetime))

	return db, nil
}

// Set stores the legacy sqlx Postgres client.
//
// Deprecated: use NewDatabase, DefaultDatabase, and DatabaseClient for new code.
func Set(r *sqlx.DB) {
	defaultStore.Set(r)
}

// Get returns the legacy sqlx Postgres client.
//
// Deprecated: use DefaultDatabase for new code.
func Get() (*sqlx.DB, error) {
	return defaultStore.Get()
}

// Default returns the legacy sqlx Postgres client.
//
// Deprecated: use DefaultDatabase for new code.
func Default() (*sqlx.DB, error) {
	c, err := Get()
	if err != nil {
		if !errors.Is(err, ErrConnectionIsNotSet) {
			return nil, err
		}

		config, err := GetConnectionConfigFromEnv()
		if err != nil {
			return nil, err
		}

		c, err = defaultStore.GetOrCreate(config)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}
