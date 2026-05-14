# Aerospike HLL

Import path: `github.com/InsideGallery/core/db/aerospike/hll`

This package counts existing Aerospike HyperLogLog bins. It supports direct
legacy helpers and a newer core-owned `Counter` boundary.

## Main APIs

- `Operator` is the minimal Aerospike operation interface required by the counters.
- `NewCounter(operator)` adapts an `Operator` to the core-owned `Counter` interface.
- `CountOptions`, `CountResult`, and `CountMode` describe core-owned count requests.
- `CountModeIntersection` counts values present in all requested HLL bins.
- `CountModeUnion` counts values present in any requested HLL bin.
- `CountHLL`, `CountIntersection`, `CountUnion`, `CountHLLBin`, and `CountHLLByBinName` are legacy helpers.
- `ErrCounterNotSet` reports a nil counter or operator dependency.

The default bin name is `aerospike.HLLBin` (`hll`) when `CountOptions.Bin` is
empty or when the legacy helpers without an explicit bin name are used.

## Usage

```go
package example

import (
	"context"

	"github.com/InsideGallery/core/db/aerospike/hll"
)

func countShared(ctx context.Context, operator hll.Operator) (int64, error) {
	counter := hll.NewCounter(operator)

	result, err := counter.Count(ctx, hll.CountOptions{
		Namespace: "transactions",
		Set:       "standard",
		Keys: []string{
			"account_email:ada@example.com",
			"transaction_currency:USD",
		},
		Mode: hll.CountModeIntersection,
	})
	if err != nil {
		return 0, err
	}

	return result.Count, nil
}
```

This package counts HLL values that already exist. Callers are responsible for
initializing HLL bins and adding values with Aerospike HLL operations.

## Operational Notes

An empty key list returns zero without calling Aerospike. Missing Aerospike keys
also count as zero. For multiple keys, the helper reads HLL bytes from all keys
except the last, then asks Aerospike to compute the union or intersection count
against the final key.

Integration tests are build-tagged with `integration`, require `AEROSPIKE_HOST`,
and load fixture data from `fixtures/aerospike/hll_count_test_data.json`.
