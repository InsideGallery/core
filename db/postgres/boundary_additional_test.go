package postgres

import (
	"context"
	"testing"
	"time"
)

func TestDatabaseClientBoundary(t *testing.T) {
	cfg := &ConnectionConfig{
		Host:            "127.0.0.1",
		Port:            "1",
		User:            "user",
		Password:        "pass",
		DB:              "db",
		MaxOpenConns:    1,
		ConnMaxLifetime: int64(time.Second),
	}

	client, err := NewDatabase(cfg)
	if err != nil {
		t.Fatalf("new database: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			t.Fatalf("close database: %v", err)
		}
	}()

	if client.SQLDB() == nil {
		t.Fatal("standard db handle is nil")
	}

	wrapped := WrapDatabase(client.db)
	if wrapped.SQLDB() != client.SQLDB() {
		t.Fatal("wrapped database did not expose same handle")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	cases := []struct {
		name string
		run  func() error
	}{
		{
			name: "ping wraps connection error",
			run: func() error {
				return client.Ping(ctx)
			},
		},
		{
			name: "exec wraps connection error",
			run: func() error {
				_, err := client.Exec(ctx, Statement{Query: "select 1"})

				return err
			},
		},
		{
			name: "query wraps connection error",
			run: func() error {
				rows, err := client.Query(ctx, Statement{Query: "select 1"})
				if rows != nil {
					defer rows.Close()
				}

				return err
			},
		},
		{
			name: "query row returns row handle",
			run: func() error {
				row := client.QueryRow(ctx, Statement{Query: "select 1"})
				var value int

				return row.Scan(&value)
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if err := test.run(); err == nil {
				t.Fatal("expected connection error")
			}
		})
	}
}
