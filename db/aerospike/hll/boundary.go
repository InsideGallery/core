package hll

import (
	"context"
	"errors"

	aero "github.com/InsideGallery/core/db/aerospike"
	coreerrors "github.com/InsideGallery/core/errors"
)

// ErrCounterNotSet reports a nil HLL counter dependency.
var ErrCounterNotSet = errors.New("hll counter is not set")

// CountMode identifies an HLL count operation.
type CountMode string

const (
	// CountModeIntersection counts values present in all HLL bins.
	CountModeIntersection CountMode = "intersection"
	// CountModeUnion counts values present in any HLL bin.
	CountModeUnion CountMode = "union"
)

// CountOptions is the core-owned input for HLL count helpers.
type CountOptions struct {
	Namespace string
	Set       string
	Keys      []string
	Bin       string
	Mode      CountMode
}

// CountResult is the core-owned result for HLL count helpers.
type CountResult struct {
	Count int64
}

// Counter is the core-owned HLL counting contract for new consumers.
type Counter interface {
	Count(ctx context.Context, options CountOptions) (CountResult, error)
}

// OperatorCounter adapts legacy Aerospike HLL operations to the Counter contract.
type OperatorCounter struct {
	operator Operator
}

// NewCounter wraps an Aerospike HLL operator with the core-owned Counter contract.
func NewCounter(operator Operator) *OperatorCounter {
	return &OperatorCounter{operator: operator}
}

// Count calculates HLL cardinality through core-owned options.
func (c *OperatorCounter) Count(ctx context.Context, options CountOptions) (CountResult, error) {
	if c == nil || c.operator == nil {
		return CountResult{}, ErrCounterNotSet
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if err := ctx.Err(); err != nil {
		return CountResult{}, coreerrors.WrapBoundary("aerospike hll", "count", err)
	}

	bin := options.Bin
	if bin == "" {
		bin = aero.HLLBin
	}

	count, err := countHLLByBinName(c.operator, options.Namespace, options.Set, options.Keys, bin, isUnion(options.Mode))
	if err != nil {
		return CountResult{}, coreerrors.WrapBoundary("aerospike hll", "count", err)
	}

	return CountResult{Count: count}, nil
}

func isUnion(mode CountMode) bool {
	return mode == CountModeUnion
}
