package postgres

import "testing"

func TestConnectionConfigDSN(t *testing.T) {
	cases := []struct {
		name string
		cfg  ConnectionConfig
		want string
	}{
		{
			name: "without application name",
			cfg: ConnectionConfig{
				Host:     "localhost",
				Port:     "5432",
				User:     "user",
				Password: "pass",
				DB:       "db",
			},
			want: "port=5432 dbname=db user=user password=pass host=localhost sslmode=disable",
		},
		{
			name: "with application name",
			cfg: ConnectionConfig{
				Host:            "localhost",
				Port:            "5433",
				User:            "user",
				Password:        "pass",
				DB:              "db",
				ApplicationName: "app",
			},
			want: "port=5433 dbname=db user=user password=pass host=localhost fallback_application_name=app sslmode=disable",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := test.cfg.GetDSN(); got != test.want {
				t.Fatalf("dsn = %q, want %q", got, test.want)
			}
		})
	}
}

func TestDefaultWithoutConnecting(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "default opens configured sql handle",
			run: func(t *testing.T) {
				t.Helper()

				Set(nil)
				t.Cleanup(func() {
					Set(nil)
				})

				t.Setenv("POSTGRES_HOST", "localhost")
				t.Setenv("POSTGRES_PORT", "5432")
				t.Setenv("POSTGRES_USER", "user")
				t.Setenv("POSTGRES_PASSWORD", "pass")
				t.Setenv("POSTGRES_DB", "db")
				t.Setenv("POSTGRES_MAXOPENCONNS", "3")
				t.Setenv("POSTGRES_CONNMAXLIFETIME", "1")

				db, err := Default()
				if err != nil {
					t.Fatalf("default: %v", err)
				}

				if db == nil {
					t.Fatal("db is nil")
				}
				defer db.Close()

				stats := db.Stats()
				if stats.MaxOpenConnections != 3 {
					t.Fatalf("max open = %d, want 3", stats.MaxOpenConnections)
				}
			},
		},
		{
			name: "get returns set client",
			run: func(t *testing.T) {
				t.Helper()

				Set(nil)
				t.Cleanup(func() {
					Set(nil)
				})

				t.Setenv("POSTGRES_CONNMAXLIFETIME", "1")

				db, err := Default()
				if err != nil {
					t.Fatalf("default: %v", err)
				}
				defer db.Close()

				got, err := Get()
				if err != nil {
					t.Fatalf("get: %v", err)
				}

				if got != db {
					t.Fatal("get did not return default db")
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}
