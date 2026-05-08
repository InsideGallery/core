package client

import (
	"errors"
	"strings"
	"testing"

	"github.com/nats-io/nkeys"

	"github.com/InsideGallery/core/testutils"
)

func TestGetNATSConnectionConfigFromEnv(t *testing.T) {
	t.Setenv("CUSTOM_NATS_ADDR", "test")

	// success with custom variable
	c, err := GetNATSConnectionConfigFromEnv("custom_nats")
	testutils.Equal(t, err, nil)
	testutils.Equal(t, c.Addr, "test")

	// success with unknown/invalid prefix
	_, err = GetNATSConnectionConfigFromEnv("unknown_nats")
	testutils.Equal(t, err, nil)
}

func TestConfigGetOptionsStrict(t *testing.T) {
	keyPair, err := nkeys.CreateUser()
	if err != nil {
		t.Fatalf("create user key: %v", err)
	}

	seed, err := keyPair.Seed()
	if err != nil {
		t.Fatalf("get seed: %v", err)
	}

	cases := []struct {
		name        string
		config      *Config
		wantErr     bool
		wantWrapped string
	}{
		{
			name:   "without seed returns options",
			config: &Config{},
		},
		{
			name:   "with valid seed returns options",
			config: &Config{Seed: string(seed)},
		},
		{
			name:        "with invalid seed returns wrapped error",
			config:      &Config{Seed: "invalid"},
			wantErr:     true,
			wantWrapped: "get key from seed",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			options, err := test.config.GetOptionsStrict()
			if test.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}

				if errors.Unwrap(err) == nil {
					t.Fatalf("err = %v, want wrapped error", err)
				}

				if !strings.Contains(err.Error(), test.wantWrapped) {
					t.Fatalf("err = %v, want context %q", err, test.wantWrapped)
				}

				if options != nil {
					t.Fatal("options should be nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("get options strict: %v", err)
			}

			if len(options) == 0 {
				t.Fatal("options should not be empty")
			}
		})
	}
}

func TestConfigGetOptionsCompatibility(t *testing.T) {
	cases := []struct {
		name    string
		config  *Config
		wantNil bool
	}{
		{
			name:   "valid config returns options",
			config: &Config{},
		},
		{
			name:    "invalid seed keeps legacy nil options",
			config:  &Config{Seed: "invalid"},
			wantNil: true,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			options := test.config.GetOptions()
			if test.wantNil {
				if options != nil {
					t.Fatal("options should be nil")
				}

				return
			}

			if len(options) == 0 {
				t.Fatal("options should not be empty")
			}
		})
	}
}
