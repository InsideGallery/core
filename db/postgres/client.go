package postgres

import (
	"errors"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5"        //nolint:revive
	_ "github.com/jackc/pgx/v5/stdlib" //nolint:revive

	"github.com/jmoiron/sqlx"
)

var (
	mu     sync.RWMutex
	client *sqlx.DB
)

// Set postgres client
func Set(r *sqlx.DB) {
	mu.Lock()
	client = r
	mu.Unlock()
}

// Get return postgres DB wrapped in sqlx
func Get() (*sqlx.DB, error) {
	mu.RLock()
	defer mu.RUnlock()

	if client == nil {
		return nil, ErrConnectionIsNotSet
	}

	return client, nil
}

// Default return DB type - but not interface - adhering to go idiom.
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

		db, err := sqlx.Open("pgx", config.GetDSN())
		if err != nil {
			return nil, err
		}

		db.SetMaxOpenConns(config.MaxOpenConns)
		db.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetime))

		Set(c)

		c = db
	}

	return c, nil
}
