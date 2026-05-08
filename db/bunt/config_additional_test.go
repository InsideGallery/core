package bunt

import (
	"errors"
	"os"
	"testing"

	"github.com/tidwall/buntdb"
)

func TestConnectionConfigFromEnv(t *testing.T) {
	cases := []struct {
		name         string
		prefix       string
		filename     string
		wantFilename string
	}{
		{
			name:         "default filename",
			prefix:       "UNIT_BUNT_DEFAULT",
			wantFilename: ":memory:",
		},
		{
			name:         "custom filename",
			prefix:       "UNIT_BUNT_CUSTOM",
			filename:     "/tmp/core-bunt-test.db",
			wantFilename: "/tmp/core-bunt-test.db",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if test.filename != "" {
				t.Setenv(test.prefix+"_FILENAME", test.filename)
			}

			cfg, err := GetConnectionConfigFromEnv(test.prefix)
			if err != nil {
				t.Fatalf("config from env: %v", err)
			}

			if cfg.Filename != test.wantFilename {
				t.Fatalf("filename = %q, want %q", cfg.Filename, test.wantFilename)
			}
		})
	}
}

func TestWrapperErrors(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "get missing key returns buntdb error",
			run: func(t *testing.T) {
				t.Helper()

				db, err := buntdb.Open(":memory:")
				if err != nil {
					t.Fatalf("open buntdb: %v", err)
				}
				defer db.Close()

				wrapper := &Wrapper{DB: db}
				var value map[string]string

				err = wrapper.Get("missing", &value)
				if !errors.Is(err, buntdb.ErrNotFound) {
					t.Fatalf("err = %v, want %v", err, buntdb.ErrNotFound)
				}
			},
		},
		{
			name: "set returns marshal error",
			run: func(t *testing.T) {
				t.Helper()

				db, err := buntdb.Open(":memory:")
				if err != nil {
					t.Fatalf("open buntdb: %v", err)
				}
				defer db.Close()

				wrapper := &Wrapper{DB: db}
				err = wrapper.Set("bad", map[string]func(){"bad": func() {}})
				if err == nil {
					t.Fatal("expected marshal error")
				}
			},
		},
		{
			name: "get connection uses configured filename",
			run: func(t *testing.T) {
				t.Helper()

				file, err := os.CreateTemp(t.TempDir(), "bunt-*.db")
				if err != nil {
					t.Fatalf("temp file: %v", err)
				}

				if err := file.Close(); err != nil {
					t.Fatalf("close temp file: %v", err)
				}

				t.Setenv("DB_FILENAME", file.Name())
				wrapper, err := GetConnection()
				if err != nil {
					t.Fatalf("get connection: %v", err)
				}
				defer wrapper.Close()
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}
