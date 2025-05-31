package bunt

import (
	"encoding/json"

	"github.com/tidwall/buntdb"
)

// Wrapper wrapper of buntdb
type Wrapper struct {
	*buntdb.DB
}

// GetConnection return new connection to buntdb
func GetConnection() (*Wrapper, error) {
	cfg, err := GetConnectionConfigFromEnv(EncPrefixDB)
	if err != nil {
		return nil, err
	}

	db, err := buntdb.Open(cfg.Filename)
	if err != nil {
		return nil, err
	}

	return &Wrapper{DB: db}, nil
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
