// Package bunt provides BuntDB connection helpers.
//
// New code should pass explicit configuration or an explicit environment prefix:
//
//	import "github.com/InsideGallery/core/db/bunt"
//
//	db, err := bunt.Open(&bunt.ConnectionConfig{Filename: ":memory:"})
//
// Compatibility: GetConnection remains available for consumers that still rely
// on the default DB environment prefix. Prefer Open or OpenFromEnv so ownership
// of configuration is visible at the call site.
package bunt

import (
	"encoding/json"
	"errors"

	"github.com/tidwall/buntdb"
)

var errConnectionConfigIsNotSet = errors.New("connection config is not set")

// Wrapper wrapper of buntdb
type Wrapper struct {
	*buntdb.DB
}

// Open creates a BuntDB wrapper from explicit config.
func Open(config *ConnectionConfig) (*Wrapper, error) {
	if config == nil {
		return nil, errConnectionConfigIsNotSet
	}

	db, err := buntdb.Open(config.Filename)
	if err != nil {
		return nil, err
	}

	return &Wrapper{DB: db}, nil
}

// OpenFromEnv creates a BuntDB wrapper from an explicit environment prefix.
func OpenFromEnv(prefix string) (*Wrapper, error) {
	cfg, err := GetConnectionConfigFromEnv(prefix)
	if err != nil {
		return nil, err
	}

	return Open(cfg)
}

// GetConnection return new connection to buntdb
//
// Deprecated: use Open or OpenFromEnv with explicit config ownership.
func GetConnection() (*Wrapper, error) {
	return OpenFromEnv(EncPrefixDB)
}

// Get get value by key
func (w *Wrapper) Get(name string, value interface{}) error {
	return w.View(func(tx *buntdb.Tx) error {
		content, err := tx.Get(name)
		if err != nil {
			return err
		}

		err = json.Unmarshal([]byte(content), &value)

		return err
	})
}

// Set set value by key
func (w *Wrapper) Set(name string, value interface{}) error {
	return w.Update(func(tx *buntdb.Tx) error {
		content, err := json.Marshal(value)
		if err != nil {
			return err
		}

		_, _, err = tx.Set(name, string(content), nil)

		return err
	})
}
