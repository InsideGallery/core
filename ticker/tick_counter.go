package ticker //nolint:mnd

import (
	"context"
	"sync/atomic"
)

var counter uint64

// Tick increase global tick counter
func Tick(_ context.Context) {
	atomic.AddUint64(&counter, 1)
}

// Get return global tick counter
func Get() uint64 {
	return atomic.LoadUint64(&counter)
}

// Reset reset ticker
func Reset() {
	atomic.StoreUint64(&counter, 0)
}
