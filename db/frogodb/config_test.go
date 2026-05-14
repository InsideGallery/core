package frogodb

import (
	"reflect"
	"testing"
	"time"
)

const (
	testPrefix       = "UNIT_FDB"
	testSeedPrimary  = "node1:3000"
	testSeedReplica  = "node2:3000"
	testInvalidValue = "invalid"
)

func TestDefaultConnectionConfig(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		seeds     []string
		wantSeeds []string
	}{
		{
			name:      "uses local seed by default",
			wantSeeds: []string{defaultSeed},
		},
		{
			name:      "uses explicit seeds",
			seeds:     []string{testSeedPrimary, testSeedReplica},
			wantSeeds: []string{testSeedPrimary, testSeedReplica},
		},
	}

	for _, test := range cases {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			config := DefaultConnectionConfig(test.seeds...)
			if !reflect.DeepEqual(config.Seeds, test.wantSeeds) {
				t.Fatalf("Seeds = %v, want %v", config.Seeds, test.wantSeeds)
			}

			if config.ConnectionTimeout != defaultConnectionTimeout {
				t.Fatalf("ConnectionTimeout = %v, want %v", config.ConnectionTimeout, defaultConnectionTimeout)
			}

			clientConfig := config.clientConfig()
			if !reflect.DeepEqual(clientConfig.Seeds, test.wantSeeds) {
				t.Fatalf("clientConfig.Seeds = %v, want %v", clientConfig.Seeds, test.wantSeeds)
			}
		})
	}
}

func TestGetConnectionConfigFromEnv(t *testing.T) {
	t.Setenv(testPrefix+"_SEEDS", testSeedPrimary+","+testSeedReplica)
	t.Setenv(testPrefix+"_CONNECTION_TIMEOUT", "3s")
	t.Setenv(testPrefix+"_IDLE_TIMEOUT", "4s")
	t.Setenv(testPrefix+"_TEND_INTERVAL", "5ms")
	t.Setenv(testPrefix+"_POOL_SIZE_PER_NODE", "7")
	t.Setenv(testPrefix+"_MAX_CONNS_PER_NODE", "9")
	t.Setenv(testPrefix+"_MAX_ERROR_RATE", "11")
	t.Setenv(testPrefix+"_ERROR_RATE_WINDOW", "12s")
	t.Setenv(testPrefix+"_MULTIPLEXING", "true")
	t.Setenv(testPrefix+"_MULTIPLEX_CONNS_PER_NODE", "13")
	t.Setenv(testPrefix+"_MULTIPLEX_MIN_CONNS_PER_NODE", "2")

	config, err := GetConnectionConfigFromEnv(testPrefix)
	if err != nil {
		t.Fatalf("GetConnectionConfigFromEnv() error: %v", err)
	}

	if !reflect.DeepEqual(config.Seeds, []string{testSeedPrimary, testSeedReplica}) {
		t.Fatalf("Seeds = %v", config.Seeds)
	}

	if config.ConnectionTimeout != 3*time.Second {
		t.Fatalf("ConnectionTimeout = %v", config.ConnectionTimeout)
	}

	if config.IdleTimeout != 4*time.Second {
		t.Fatalf("IdleTimeout = %v", config.IdleTimeout)
	}

	if config.TendInterval != 5*time.Millisecond {
		t.Fatalf("TendInterval = %v", config.TendInterval)
	}

	if config.PoolSizePerNode != 7 {
		t.Fatalf("PoolSizePerNode = %d", config.PoolSizePerNode)
	}

	if config.MaxConnsPerNode != 9 {
		t.Fatalf("MaxConnsPerNode = %d", config.MaxConnsPerNode)
	}

	if config.MaxErrorRate != 11 {
		t.Fatalf("MaxErrorRate = %d", config.MaxErrorRate)
	}

	if config.ErrorRateWindow != 12*time.Second {
		t.Fatalf("ErrorRateWindow = %v", config.ErrorRateWindow)
	}

	if !config.Multiplexing {
		t.Fatal("Multiplexing = false, want true")
	}

	if config.MultiplexConnsPerNode != 13 {
		t.Fatalf("MultiplexConnsPerNode = %d", config.MultiplexConnsPerNode)
	}

	if config.MultiplexMinConnsPerNode != 2 {
		t.Fatalf("MultiplexMinConnsPerNode = %d", config.MultiplexMinConnsPerNode)
	}
}

func TestGetConnectionConfigFromEnvInvalidEnv(t *testing.T) {
	cases := []struct {
		name  string
		key   string
		value string
	}{
		{
			name:  "invalid duration",
			key:   testPrefix + "_CONNECTION_TIMEOUT",
			value: testInvalidValue,
		},
		{
			name:  "invalid int",
			key:   testPrefix + "_POOL_SIZE_PER_NODE",
			value: testInvalidValue,
		},
		{
			name:  "invalid bool",
			key:   testPrefix + "_MULTIPLEXING",
			value: testInvalidValue,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv(test.key, test.value)

			_, err := GetConnectionConfigFromEnv(testPrefix)
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
