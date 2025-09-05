//go:generate easyjson -all count.go
package hll

import (
	"fmt"

	as "github.com/aerospike/aerospike-client-go/v7"
	"github.com/aerospike/aerospike-client-go/v7/types"

	aero "github.com/InsideGallery/core/db/aerospike"
)

// CountHLL calculates amount of unique transactions by passed keys.
func CountHLL(client Operator, namespace, set string, by []string, union bool) (int64, error) {
	if len(by) == 0 {
		return 0, nil
	}

	if len(by) == 1 {
		return CountHLLBin(client, namespace, set, by[0])
	}

	var (
		lastKey *as.Key
		hlls    []as.HLLValue
	)

	for i, attr := range by {
		key, err := as.NewKey(namespace, set, attr)
		if err != nil {
			return 0, fmt.Errorf("failed to create key for %s: %w", attr, err)
		}

		if i >= len(by)-1 {
			lastKey = key
			break
		}

		record, err := client.Operate(nil, key, as.GetBinOp(aero.HLLBin))
		if err != nil {
			if err.Matches(types.KEY_NOT_FOUND_ERROR) {
				return 0, nil
			}

			return 0, fmt.Errorf("failed to get hll bin for %s: %w", attr, err)
		}

		hllValue, ok := record.Bins[aero.HLLBin].(as.HLLValue)
		if !ok {
			return 0, fmt.Errorf("failed to get hll bin, wrong type %s", attr)
		}

		hllBytes, ok := hllValue.GetObject().([]byte)
		if !ok {
			return 0, fmt.Errorf("failed to get hll bin, bytes is empty %s", attr)
		}

		hlls = append(hlls, hllBytes)
	}

	if lastKey == nil {
		return 0, nil
	}

	var (
		record *as.Record
		err    as.Error
	)

	if union {
		record, err = client.Operate(nil, lastKey, as.HLLGetUnionCountOp(aero.HLLBin, hlls))
	} else {
		record, err = client.Operate(nil, lastKey, as.HLLGetIntersectCountOp(aero.HLLBin, hlls))
	}

	if err != nil {
		if err.Matches(types.KEY_NOT_FOUND_ERROR) {
			return 0, nil
		}

		return 0, fmt.Errorf("failed to get hll intersect count: %w, %v", err, lastKey.String())
	}

	counter, ok := record.Bins[aero.HLLBin].(int64)
	if !ok {
		return 0, nil
	}

	return counter, nil
}

func CountIntersection(client Operator, namespace, set string, by []string) (int64, error) {
	return CountHLL(client, namespace, set, by, false)
}

func CountUnion(client Operator, namespace, set string, by []string) (int64, error) {
	return CountHLL(client, namespace, set, by, true)
}

func CountHLLBin(client Operator, namespace, set, by string) (int64, error) {
	return CountHLLByBinName(client, namespace, set, by, aero.HLLBin)
}

func CountHLLByBinName(client Operator, namespace, set, by, hllBinName string) (int64, error) {
	key, err := as.NewKey(namespace, set, by)
	if err != nil {
		return 0, fmt.Errorf("failed to create key for %s: %w", by, err)
	}

	record, err := client.Operate(nil, key, as.HLLGetCountOp(hllBinName))
	if err != nil {
		if err.Matches(types.KEY_NOT_FOUND_ERROR) {
			return 0, nil
		}

		return 0, fmt.Errorf("failed to get hll count for %s: %w", by, err)
	}

	count, ok := record.Bins[hllBinName].(int64)
	if !ok {
		return 0, nil
	}

	return count, nil
}
