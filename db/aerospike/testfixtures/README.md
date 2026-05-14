# Aerospike Test Fixtures

Import path: `github.com/InsideGallery/core/db/aerospike/testfixtures`

This package provides helpers for integration tests that load and clean up
Aerospike records from JSON fixture data.

## Main APIs

- `AerospikeFixture` describes one fixture record with `namespace`, `set`, `key`, and `bins` JSON fields.
- `LoadAerospikeFixtures(client, fixturesData)` unmarshals fixture JSON and writes each record.
- `CleanupAerospikeFixtures(client, fixtures)` deletes the records described by loaded fixtures.

## Usage

```go
package example_test

import (
	"os"
	"testing"

	as "github.com/aerospike/aerospike-client-go/v7"

	"github.com/InsideGallery/core/db/aerospike/testfixtures"
)

func TestWithAerospikeFixtures(t *testing.T) {
	client, err := as.NewClient("127.0.0.1", 3000)
	if err != nil {
		t.Fatalf("create aerospike client: %v", err)
	}
	defer client.Close()

	data, err := os.ReadFile("testdata/aerospike.json")
	if err != nil {
		t.Fatalf("read fixtures: %v", err)
	}

	fixtures, err := testfixtures.LoadAerospikeFixtures(client, data)
	if err != nil {
		t.Fatalf("load fixtures: %v", err)
	}
	t.Cleanup(func() {
		if err := testfixtures.CleanupAerospikeFixtures(client, fixtures); err != nil {
			t.Fatalf("cleanup fixtures: %v", err)
		}
	})

	// Run assertions against the loaded Aerospike records.
}
```

## Operational Notes

`LoadAerospikeFixtures` uses an Aerospike write policy with `SendKey` set to
true. The package requires a live Aerospike client and is intended for tests,
not production data loading.
