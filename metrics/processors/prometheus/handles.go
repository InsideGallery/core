package prometheus

import (
	"sync"
	"sync/atomic"
)

// Label resolution, not the actual increment, is what recording costs: Count,
// Gauge, and Distribution each re-ran NormalizeTags, sanitizeLabelName, and a
// fresh names/values build for label sets the process had already resolved
// thousands of times before (ten allocations per record on a three-tag metric).
// Callers on a data path record the same bounded (name, tags) tuples over and
// over, so the resolved child metric is memoized per tuple: a cache hit is one
// atomic load, one map lookup, and the Add/Set/Observe itself, with nothing
// allocated and no lock taken.
//
// A sync.Map cannot back this: its Load takes an `any` key, so every lookup
// would box the key and allocate exactly what the cache exists to avoid. The
// memo is therefore a copy-on-write map read through an atomic.Pointer.
// Snapshots are never mutated after publication, so readers need no lock and
// only a first-seen tuple takes the mutex to publish a replacement.
const (
	// maxCachedTagCount is the widest tag list the fixed-size cache key can
	// hold. Wider calls still record correctly, they just resolve their labels
	// every time. The key is hashed per lookup, and unused slots are not free
	// (a six-slot key costs ~65ns per record against ~81ns for eight), so the
	// width is kept at twice the widest tag list any current caller records.
	maxCachedTagCount = 6

	// maxCachedHandles bounds the memo. Tag values are caller-supplied, so an
	// unintended high-cardinality label (a request ID, a timestamp) would
	// otherwise make every snapshot publication copy an ever-growing map. Once
	// the bound is reached the memo stops growing and further tuples resolve
	// per call, which is the behavior that existed before the cache.
	maxCachedHandles = 4096
)

// handleKey identifies one recorded (metric name, tag list) tuple. It is a
// fixed-size comparable value so that hot-path lookups copy the key onto the
// stack instead of allocating a composite one. Tag lists differing only in
// order resolve to identical labels but occupy separate entries; that costs one
// extra entry and keeps key construction allocation-free.
type handleKey struct {
	name string
	tags [maxCachedTagCount]string
}

func newHandleKey(name string, tags []string) handleKey {
	key := handleKey{name: name}
	copy(key.tags[:], tags)

	return key
}

// handleCache memoizes resolved metric children per recorded tuple.
type handleCache[T any] struct {
	// limit bounds the number of memoized handles. Zero, the value the
	// processor uses, means maxCachedHandles.
	limit int

	mu       sync.Mutex
	snapshot atomic.Pointer[map[handleKey]T]
}

// load returns the memoized handle for the tuple. It is the recording hot path:
// no allocation, no lock.
func (c *handleCache[T]) load(name string, tags []string) (T, bool) {
	var zero T

	if len(tags) > maxCachedTagCount {
		return zero, false
	}

	snapshot := c.snapshot.Load()
	if snapshot == nil {
		return zero, false
	}

	handle, ok := (*snapshot)[newHandleKey(name, tags)]

	return handle, ok
}

// store memoizes the handle for the tuple by publishing a snapshot that
// includes it. It is only reached the first time a tuple is recorded.
func (c *handleCache[T]) store(name string, tags []string, handle T) {
	if len(tags) > maxCachedTagCount {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	current := c.snapshot.Load()

	size := 1

	if current != nil {
		if len(*current) >= c.capacity() {
			return
		}

		size += len(*current)
	}

	updated := make(map[handleKey]T, size)

	if current != nil {
		for key, cached := range *current {
			updated[key] = cached
		}
	}

	updated[newHandleKey(name, tags)] = handle
	c.snapshot.Store(&updated)
}

// capacity is the memo bound in effect.
func (c *handleCache[T]) capacity() int {
	if c.limit > 0 {
		return c.limit
	}

	return maxCachedHandles
}

// size reports how many handles are memoized.
func (c *handleCache[T]) size() int {
	snapshot := c.snapshot.Load()
	if snapshot == nil {
		return 0
	}

	return len(*snapshot)
}
