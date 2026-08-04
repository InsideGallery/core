package prometheus

import (
	"sync"
	"sync/atomic"

	stdprom "github.com/prometheus/client_golang/prometheus"

	"github.com/InsideGallery/core/metrics"
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
	// otherwise make every snapshot publication copy an ever-growing map. At the
	// bound the memo starts evicting rather than sealing itself shut, so a
	// cardinality burst costs the tuples it displaces per-call label resolution
	// until they are recorded often enough to be re-admitted, instead of for the
	// life of the process. See admissionInterval for the rate.
	maxCachedHandles = 4096

	// admissionIntervalDivisor sets how often a full memo admits a first-seen
	// tuple: one attempt in capacity/admissionIntervalDivisor. See
	// admissionInterval.
	admissionIntervalDivisor = 8
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

	// Slow-path counters, kept so that a cache which has quietly stopped
	// memoizing, or is churning, is visible from a scrape instead of only from a
	// profile. All are written on paths that already resolve labels per record,
	// and never on a cache hit; they are declared last so the hot-path fields
	// keep their offsets.
	bypassWidth atomic.Int64
	bypassFull  atomic.Int64
	evictions   atomic.Int64

	// admissionPressure counts first-seen tuples the full memo has turned away
	// since its last admission. Read and written only under mu.
	admissionPressure int
}

// CacheStats reports one handle cache's occupancy and its cumulative slow-path
// events. BypassWidth counts records whose tag list was too wide for the
// fixed-size key; BypassFull counts records of a tuple the full memo turned
// away; Evictions counts residents dropped to admit one. BypassFull no longer
// means "locked out for the life of the process" — a turned-away tuple is
// admitted once it accumulates enough attempts — so the pair reads as churn:
// BypassFull climbing with Evictions is a cardinality burst being absorbed. All
// three are monotonic for the life of the processor.
type CacheStats struct {
	Size        int
	BypassWidth int64
	BypassFull  int64
	Evictions   int64
}

// Stats reads the cache's occupancy and slow-path counters. The values are
// loaded independently, so a concurrent record can land between them; they are
// intended for observability, not for exact cross-field arithmetic.
func (c *handleCache[T]) Stats() CacheStats {
	return CacheStats{
		Size:        c.size(),
		BypassWidth: c.bypassWidth.Load(),
		BypassFull:  c.bypassFull.Load(),
		Evictions:   c.evictions.Load(),
	}
}

// load returns the memoized handle for the tuple. It is the recording hot path:
// no allocation, no lock.
func (c *handleCache[T]) load(name string, tags []string) (T, bool) {
	var zero T

	// Counted here rather than in store: every record calls load first, so this
	// sees each wide record exactly once, while store is reached only after the
	// labels have already been resolved.
	if len(tags) > maxCachedTagCount {
		c.bypassWidth.Add(1)

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
// includes it. It is reached the first time a tuple is recorded, and again if
// that tuple's entry was evicted in the meantime.
func (c *handleCache[T]) store(name string, tags []string, handle T) {
	if len(tags) > maxCachedTagCount {
		return
	}

	key := newHandleKey(name, tags)

	c.mu.Lock()
	defer c.mu.Unlock()

	current := c.snapshot.Load()
	evicting := false

	if c.isFull(current, key) {
		if !c.admit() {
			c.bypassFull.Add(1)

			return
		}

		evicting = true
	}

	updated := c.republish(current, evicting)

	updated[key] = handle
	c.snapshot.Store(&updated)

	if evicting {
		c.evictions.Add(1)
	}
}

// forget drops every memoized handle whose key matches, republishing the snapshot
// without them. It is the memo half of series retirement (see retire.go): the child
// a matching entry resolved to has been deleted, so the entry has to go with it or
// records of that tuple keep landing on a child no scrape can see.
//
// It walks every entry instead of keeping an index, because retirement happens on a
// lifecycle event while an index would have to be maintained per record — the one
// path this design keeps cheap. The snapshot is republished only when something
// matched, so a delete for a subject this cache never memoized costs nothing but the
// walk.
func (c *handleCache[T]) forget(matches func(key handleKey) bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	current := c.snapshot.Load()
	if current == nil {
		return
	}

	updated := make(map[handleKey]T, len(*current))

	for key, handle := range *current {
		if matches(key) {
			continue
		}

		updated[key] = handle
	}

	if len(updated) == len(*current) {
		return
	}

	c.snapshot.Store(&updated)
}

// isFull reports whether the memo is at its bound and this tuple would have to
// displace a resident to get in. A tuple already resident never displaces one:
// two records of a first-seen tuple can race into store, and the loser must not
// evict anything to re-publish an entry that is already there.
func (c *handleCache[T]) isFull(current *map[handleKey]T, key handleKey) bool {
	if current == nil || len(*current) < c.capacity() {
		return false
	}

	_, resident := (*current)[key]

	return !resident
}

// admit reports whether this attempt is the one the full memo lets in, and
// clears the pressure it accumulated when it is. Callers hold mu.
func (c *handleCache[T]) admit() bool {
	c.admissionPressure++

	if c.admissionPressure < c.admissionInterval() {
		return false
	}

	c.admissionPressure = 0

	return true
}

// republish copies the published snapshot into a fresh map, dropping one entry
// when the insert has to evict. Go randomizes where a map range starts, so the
// entry dropped is an arbitrary resident: random replacement needs no per-hit
// bookkeeping, which is what keeps a cache hit down to one atomic load and one
// map lookup — an LRU list or a clock bit would have to be maintained on every
// hit, on the one path this design exists to keep cheap.
func (c *handleCache[T]) republish(current *map[handleKey]T, evicting bool) map[handleKey]T {
	size := 1
	if current != nil {
		size += len(*current)
	}

	updated := make(map[handleKey]T, size)
	if current == nil {
		return updated
	}

	dropped := !evicting

	for cachedKey, cached := range *current {
		if !dropped {
			dropped = true

			continue
		}

		updated[cachedKey] = cached
	}

	return updated
}

// capacity is the memo bound in effect.
func (c *handleCache[T]) capacity() int {
	if c.limit > 0 {
		return c.limit
	}

	return maxCachedHandles
}

// admissionInterval is how many first-seen tuples a full memo turns away before
// it admits one, and it is why eviction is affordable at all. Publication is
// copy-on-write, so every admission rebuilds the whole map — an O(capacity) copy
// on a path whose whole point is to be cheap. Admitting every first-seen tuple
// turns an unbounded label, the exact defect the bound exists to survive, into a
// 176µs recording path against 153ns for refusing (BenchmarkHandleCachePolicy
// oversized_cycle, 4096-entry memo, tuple space one eighth past it). Evicting a
// batch per admission does not fix that: the batch changes how many entries a
// rebuild drops, not how often a rebuild happens. Rationing admission does, and
// it costs 465ns/op on the same workload — 3.0x refusing, and cheap in absolute
// terms because a record that misses the memo already pays ~0.6µs to resolve its
// labels.
//
// What this keeps from refuse-at-capacity is its cost; what it drops is its
// permanence. A tuple is no longer locked out for the life of the process, only
// until it is recorded often enough to win an admission — and since attempts are
// what wins admission, the tuples recorded most often win first. The interval
// scales with capacity because the rebuild it rations does.
func (c *handleCache[T]) admissionInterval() int {
	interval := c.capacity() / admissionIntervalDivisor
	if interval < 1 {
		return 1
	}

	return interval
}

// size reports how many handles are memoized.
func (c *handleCache[T]) size() int {
	snapshot := c.snapshot.Load()
	if snapshot == nil {
		return 0
	}

	return len(*snapshot)
}

// Resolved handles (metrics.HandleProvider). The memo above removes label
// resolution from a repeat record but not the lookup that finds the memoized
// child: the key is hashed per record, and on a caller's hot path that lookup is
// what is left to remove. A caller that records the same tuple for the life of the
// process resolves it once here instead and keeps the child.
//
// This path deliberately does NOT populate the memo. A handle never looks its
// tuple up again, so an entry for it would occupy one of the bounded slots without
// ever being read — and, once the memo is at capacity, would ration admission
// against tuples that do read it.
//
//nolint:ireturn // handle API returns the abstraction by design
func (p *processor) CounterHandle(name string, tags []string) (metrics.Counter, error) {
	counter, err := p.resolveCounter(name, tags)
	if err != nil {
		return nil, err
	}

	return counterHandle{counter: counter}, nil
}

// GaugeHandle resolves a gauge child. A Prometheus gauge already satisfies
// metrics.Gauge, so the child is returned as-is.
//
//nolint:ireturn // handle API returns the abstraction by design
func (p *processor) GaugeHandle(name string, tags []string) (metrics.Gauge, error) {
	gauge, err := p.resolveGauge(name, tags)
	if err != nil {
		return nil, err
	}

	return gauge, nil
}

// DistributionHandle resolves a histogram child. A Prometheus observer already
// satisfies metrics.Observer, so the child is returned as-is.
//
//nolint:ireturn // handle API returns the abstraction by design
func (p *processor) DistributionHandle(name string, tags []string) (metrics.Observer, error) {
	observer, err := p.resolveHistogram(name, tags)
	if err != nil {
		return nil, err
	}

	return observer, nil
}

// counterHandle adapts a Prometheus counter child to metrics.Counter, which counts
// in int64 like Processor.Count while Prometheus counts in float64. Gauges and
// histograms need no such adapter.
type counterHandle struct {
	counter stdprom.Counter
}

// Add records the increment. A negative value is dropped: Count reports it as an
// error, a handle has no error channel, and stdprom.Counter.Add panics on a
// negative — taking down the caller's record path over a sample.
func (h counterHandle) Add(value int64) {
	if value < 0 {
		return
	}

	h.counter.Add(float64(value))
}

// Names and label values of the processor's self-instrumentation. Saturating the
// memo degrades recording silently — the displaced tuples pay full label
// resolution until they are recorded again — so the condition is exported as
// ordinary series that an existing scrape already collects, rather than left to
// a profile.
const (
	handleCacheSizeMetric      = "metrics_handle_cache_size"
	handleCacheBypassMetric    = "metrics_handle_cache_bypass_total"
	handleCacheEvictionsMetric = "metrics_handle_cache_evictions_total"

	kindLabel   = "kind"
	reasonLabel = "reason"

	cacheKindCounter   = "counter"
	cacheKindGauge     = "gauge"
	cacheKindHistogram = "histogram"

	bypassReasonWidth = "width"
	bypassReasonFull  = "full"
)

// handleCacheCollector exports the processor's own handle caches. It reads them
// at scrape time instead of mirroring them into gauges as records happen, so
// the recording path carries none of the cost of being observable.
type handleCacheCollector struct {
	processor *processor

	sizeDesc      *stdprom.Desc
	bypassDesc    *stdprom.Desc
	evictionsDesc *stdprom.Desc
}

func newHandleCacheCollector(p *processor) *handleCacheCollector {
	return &handleCacheCollector{
		processor: p,
		sizeDesc: stdprom.NewDesc(
			handleCacheSizeMetric,
			"Resolved metric handles currently memoized by the recording cache.",
			[]string{kindLabel},
			nil,
		),
		bypassDesc: stdprom.NewDesc(
			handleCacheBypassMetric,
			"Records that bypassed the handle cache and resolved their labels per call.",
			[]string{kindLabel, reasonLabel},
			nil,
		),
		evictionsDesc: stdprom.NewDesc(
			handleCacheEvictionsMetric,
			"Memoized handles dropped to make room for a newly recorded tuple.",
			[]string{kindLabel},
			nil,
		),
	}
}

func (c *handleCacheCollector) Describe(descs chan<- *stdprom.Desc) {
	descs <- c.sizeDesc

	descs <- c.bypassDesc

	descs <- c.evictionsDesc
}

func (c *handleCacheCollector) Collect(collected chan<- stdprom.Metric) {
	caches := []struct {
		kind  string
		stats CacheStats
	}{
		{kind: cacheKindCounter, stats: c.processor.counterHandles.Stats()},
		{kind: cacheKindGauge, stats: c.processor.gaugeHandles.Stats()},
		{kind: cacheKindHistogram, stats: c.processor.histogramHandles.Stats()},
	}

	for _, cache := range caches {
		emitConstMetric(collected, c.sizeDesc, stdprom.GaugeValue, float64(cache.stats.Size), cache.kind)
		emitConstMetric(collected, c.bypassDesc, stdprom.CounterValue,
			float64(cache.stats.BypassWidth), cache.kind, bypassReasonWidth)
		emitConstMetric(collected, c.bypassDesc, stdprom.CounterValue,
			float64(cache.stats.BypassFull), cache.kind, bypassReasonFull)
		emitConstMetric(collected, c.evictionsDesc, stdprom.CounterValue,
			float64(cache.stats.Evictions), cache.kind)
	}
}

// emitConstMetric sends one constant sample, reporting a construction failure
// as an invalid metric so a malformed series fails its own scrape instead of
// panicking and taking the whole endpoint down with it.
func emitConstMetric(
	collected chan<- stdprom.Metric,
	desc *stdprom.Desc,
	valueType stdprom.ValueType,
	value float64,
	labelValues ...string,
) {
	metric, err := stdprom.NewConstMetric(desc, valueType, value, labelValues...)
	if err != nil {
		collected <- stdprom.NewInvalidMetric(desc, err)

		return
	}

	collected <- metric
}
