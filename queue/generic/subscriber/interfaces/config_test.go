package interfaces

import (
	"testing"
	"time"
)

type legacyConfig struct {
	readTimeout       time.Duration
	maxConcurrentSize uint64
	concurrentSize    int
}

func (c legacyConfig) GetReadTimeout() time.Duration {
	return c.readTimeout
}

func (c legacyConfig) GetMaxConcurrentSize() uint64 {
	return c.maxConcurrentSize
}

func (c legacyConfig) GetConcurrentSize() int {
	return c.concurrentSize
}

type aliasConfig struct {
	legacyConfig

	aliasReadTimeout       time.Duration
	aliasMaxConcurrentSize uint64
	aliasConcurrentSize    int
}

func (c aliasConfig) ReadTimeout() time.Duration {
	return c.aliasReadTimeout
}

func (c aliasConfig) MaxConcurrentSize() uint64 {
	return c.aliasMaxConcurrentSize
}

func (c aliasConfig) ConcurrentSize() int {
	return c.aliasConcurrentSize
}

func TestConfigValueAliases(t *testing.T) {
	cases := []struct {
		name                  string
		config                Config
		wantReadTimeout       time.Duration
		wantMaxConcurrentSize uint64
		wantConcurrentSize    int
	}{
		{
			name: "falls back to deprecated getters",
			config: legacyConfig{
				readTimeout:       time.Second,
				maxConcurrentSize: 3,
				concurrentSize:    2,
			},
			wantReadTimeout:       time.Second,
			wantMaxConcurrentSize: 3,
			wantConcurrentSize:    2,
		},
		{
			name: "prefers value aliases",
			config: aliasConfig{
				legacyConfig: legacyConfig{
					readTimeout:       time.Second,
					maxConcurrentSize: 3,
					concurrentSize:    2,
				},
				aliasReadTimeout:       2 * time.Second,
				aliasMaxConcurrentSize: 5,
				aliasConcurrentSize:    4,
			},
			wantReadTimeout:       2 * time.Second,
			wantMaxConcurrentSize: 5,
			wantConcurrentSize:    4,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := ReadTimeout(test.config); got != test.wantReadTimeout {
				t.Fatalf("read timeout = %s, want %s", got, test.wantReadTimeout)
			}

			if got := MaxConcurrentSize(test.config); got != test.wantMaxConcurrentSize {
				t.Fatalf("max concurrent size = %d, want %d", got, test.wantMaxConcurrentSize)
			}

			if got := ConcurrentSize(test.config); got != test.wantConcurrentSize {
				t.Fatalf("concurrent size = %d, want %d", got, test.wantConcurrentSize)
			}
		})
	}
}
