package postgres

import (
	"testing"
	"time"
)

func TestDatabaseContract(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		assert func(t *testing.T)
	}{
		{
			name: "database client implements database",
			assert: func(t *testing.T) {
				t.Helper()

				var _ Database = (*DatabaseClient)(nil)
			},
		},
		{
			name: "database options convert to config",
			assert: func(t *testing.T) {
				t.Helper()

				options := DatabaseOptions{
					Host:            "localhost",
					Port:            "5432",
					User:            "inside",
					Password:        "secret",
					Database:        "core",
					ApplicationName: "tests",
					MaxOpenConns:    10,
					ConnMaxLifetime: time.Second,
				}

				got := options.config()
				if got.DB != options.Database {
					t.Fatalf("DB = %q, want %q", got.DB, options.Database)
				}

				if got.ConnMaxLifetime != int64(time.Second) {
					t.Fatalf("ConnMaxLifetime = %d, want %d", got.ConnMaxLifetime, int64(time.Second))
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
