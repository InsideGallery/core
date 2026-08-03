package prometheus

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	stdprom "github.com/prometheus/client_golang/prometheus"

	"github.com/InsideGallery/core/metrics"
)

var benchmarkHTTPStatus int

func BenchmarkProcessorRecord(b *testing.B) {
	tags := []string{
		"status:200",
		"method:GET",
		"route:/v2/notifyapi/notifications",
	}
	wideTags := make([]string, 0, maxCachedTagCount+1)

	for index := range maxCachedTagCount + 1 {
		wideTags = append(wideTags, "label"+strconv.Itoa(index)+":value"+strconv.Itoa(index))
	}
	cases := []struct {
		name   string
		record func(metrics.Processor) error
	}{
		{
			name: "count_existing_collector",
			record: func(processor metrics.Processor) error {
				return processor.Count("ptolemy_requests_total", 1, tags)
			},
		},
		{
			name: "gauge_existing_collector",
			record: func(processor metrics.Processor) error {
				return processor.Gauge("ptolemy_active_sessions", 7, tags)
			},
		},
		{
			name: "distribution_existing_collector",
			record: func(processor metrics.Processor) error {
				return processor.Distribution("ptolemy_request_duration_ms", 12.5, tags)
			},
		},
		{
			// Tag lists too wide for the handle cache key resolve their labels
			// on every record: the fallback path, and what every record cost
			// before the cache existed.
			name: "count_uncached_wide_tags",
			record: func(processor metrics.Processor) error {
				return processor.Count("ptolemy_wide_total", 1, wideTags)
			},
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			processor := newBenchmarkProcessor(b)
			if err := tc.record(processor); err != nil {
				b.Fatal(err)
			}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				if err := tc.record(processor); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkHandleCachePolicy measures the memo's eviction policy against the
// freeze policy it replaces, because the fix must not cost more than the label
// resolution it avoids (AIC risk R4). Both arms run in one process, over the
// same tuple space, through the same real resolution path, and differ only in
// what store() does at capacity.
//
// The workloads are the three cases that behave differently:
//
//   - unsaturated_cycle stays below the bound, where no policy is exercised;
//     eviction must be free here.
//   - saturated_hot_set is the case the change exists for: the memo is filled by
//     a cardinality burst, and the tuples recorded afterwards are the hot ones.
//     Under freeze they never memoize again.
//   - oversized_cycle is the adversarial case: a genuinely unbounded label
//     recorded uniformly forever, which no policy can memoize. Freeze wins it by
//     refusing to try, so the number to watch is how much eviction loses by.
func BenchmarkHandleCachePolicy(b *testing.B) {
	const (
		burstMetric    = "policy_burst_total"
		workloadMetric = "policy_workload_total"
		hotTuples      = 3
		oversizedSlack = maxCachedHandles / 8
	)

	policies := []struct {
		name     string
		newCache func() benchmarkHandleCache
	}{
		{name: "freeze", newCache: func() benchmarkHandleCache { return &frozenHandleCache{} }},
		{name: "evict", newCache: func() benchmarkHandleCache { return &handleCache[stdprom.Counter]{} }},
	}

	workloads := []struct {
		name     string
		burst    int
		measured int
		warm     bool
	}{
		{name: "unsaturated_cycle", burst: 0, measured: maxCachedHandles / 2, warm: true},
		{name: "saturated_hot_set", burst: maxCachedHandles, measured: hotTuples, warm: false},
		{name: "oversized_cycle", burst: 0, measured: maxCachedHandles + oversizedSlack, warm: true},
	}

	for _, workload := range workloads {
		burst := benchmarkTupleSpace(workload.burst)
		measured := benchmarkTupleSpace(workload.measured)

		for _, policy := range policies {
			b.Run(workload.name+"/"+policy.name, func(b *testing.B) {
				processor := newBenchmarkPolicyProcessor(b)
				cache := policy.newCache()

				for _, tags := range burst {
					benchmarkRecord(b, processor, cache, burstMetric, tags)
				}

				if workload.warm {
					for _, tags := range measured {
						benchmarkRecord(b, processor, cache, workloadMetric, tags)
					}
				}

				b.ReportAllocs()
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					benchmarkRecord(b, processor, cache, workloadMetric, measured[i%len(measured)])
				}
			})
		}
	}
}

// benchmarkTupleSpace builds count distinct tag lists whose last label is the
// unbounded one, the shape that saturates the memo in the first place.
func benchmarkTupleSpace(count int) [][]string {
	space := make([][]string, 0, count)

	for index := range count {
		space = append(space, []string{"status:200", "method:GET", "index:" + strconv.Itoa(index)})
	}

	return space
}

// benchmarkHandleCache is the part of the memo a record uses, so a benchmark can
// swap the policy without the shipped cache carrying a benchmark-only mode.
type benchmarkHandleCache interface {
	load(name string, tags []string) (stdprom.Counter, bool)
	store(name string, tags []string, handle stdprom.Counter)
}

// benchmarkRecord is processor.countResolved with the memo swapped out: the same
// real label resolution, collector lookup, and child resolution, so the arms
// differ only in the policy under test.
func benchmarkRecord(b *testing.B, p *processor, cache benchmarkHandleCache, name string, tags []string) {
	if counter, ok := cache.load(name, tags); ok {
		counter.Add(1)

		return
	}

	collector, labels, err := p.counter(name, tags)
	if err != nil {
		b.Fatal(err)
	}

	counter, err := collector.GetMetricWithLabelValues(labels.values...)
	if err != nil {
		b.Fatal(err)
	}

	counter.Add(1)
	cache.store(name, tags, counter)
}

// frozenHandleCache is the pre-eviction memo, kept as the baseline the shipped
// policy is measured against: at capacity store() refused to memoize anything
// new, so every tuple first seen after saturation resolved its labels on every
// record for the life of the process. Everything except that refusal matches
// handleCache, so the comparison isolates the policy.
type frozenHandleCache struct {
	limit int

	mu       sync.Mutex
	snapshot atomic.Pointer[map[handleKey]stdprom.Counter]
}

// load mirrors handleCache.load, whose generic T is the Prometheus child
// interface the memo hands back.
//
//nolint:ireturn // matches the memoized handle type under test
func (c *frozenHandleCache) load(name string, tags []string) (stdprom.Counter, bool) {
	if len(tags) > maxCachedTagCount {
		return nil, false
	}

	snapshot := c.snapshot.Load()
	if snapshot == nil {
		return nil, false
	}

	handle, ok := (*snapshot)[newHandleKey(name, tags)]

	return handle, ok
}

func (c *frozenHandleCache) store(name string, tags []string, handle stdprom.Counter) {
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

	updated := make(map[handleKey]stdprom.Counter, size)

	if current != nil {
		for key, cached := range *current {
			updated[key] = cached
		}
	}

	updated[newHandleKey(name, tags)] = handle
	c.snapshot.Store(&updated)
}

func (c *frozenHandleCache) capacity() int {
	if c.limit > 0 {
		return c.limit
	}

	return maxCachedHandles
}

func BenchmarkHTTPHandler(b *testing.B) {
	b.Run("no_processor", func(b *testing.B) {
		resetActiveProcessorForBenchmark()

		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		writer := newBenchmarkResponseWriter()

		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			writer.reset()
			HTTPHandler(writer, req)
		}

		benchmarkHTTPStatus = writer.status
	})

	b.Run("active_text", func(b *testing.B) {
		processor := newBenchmarkProcessor(b)
		seedBenchmarkProcessor(b, processor)

		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		writer := newBenchmarkResponseWriter()

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			writer.reset()
			HTTPHandler(writer, req)
		}

		benchmarkHTTPStatus = writer.status
	})

	b.Run("active_openmetrics", func(b *testing.B) {
		processor := newBenchmarkProcessor(b)
		seedBenchmarkProcessor(b, processor)

		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		req.Header.Set("Accept", "application/openmetrics-text")

		writer := newBenchmarkResponseWriter()

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			writer.reset()
			HTTPHandler(writer, req)
		}

		benchmarkHTTPStatus = writer.status
	})
}

func newBenchmarkProcessor(b *testing.B) metrics.Processor {
	b.Helper()

	resetActiveProcessorForBenchmark()

	processor, err := New(metrics.Config{}, "bench-svc")
	if err != nil {
		b.Fatal(err)
	}

	b.Cleanup(func() {
		if err := processor.Close(); err != nil {
			b.Fatal(err)
		}

		resetActiveProcessorForBenchmark()
	})

	return processor
}

// newBenchmarkPolicyProcessor returns the concrete processor, which the policy
// benchmark needs so it can reuse the real label-resolution path directly.
func newBenchmarkPolicyProcessor(b *testing.B) *processor {
	b.Helper()

	raw := newBenchmarkProcessor(b)

	concrete, ok := raw.(*processor)
	if !ok {
		b.Fatalf("processor type = %T", raw)
	}

	return concrete
}

func seedBenchmarkProcessor(b *testing.B, processor metrics.Processor) {
	b.Helper()

	if err := processor.Count("ptolemy_requests_total", 3, []string{"status:200", "method:GET"}); err != nil {
		b.Fatal(err)
	}

	if err := processor.Gauge("ptolemy_active_sessions", 7, []string{"site:42"}); err != nil {
		b.Fatal(err)
	}

	if err := processor.Distribution("ptolemy_request_duration_ms", 12.5, []string{"route:notifications"}); err != nil {
		b.Fatal(err)
	}
}

type benchmarkResponseWriter struct {
	header http.Header
	status int
	bytes  int
}

func newBenchmarkResponseWriter() *benchmarkResponseWriter {
	return &benchmarkResponseWriter{
		header: make(http.Header),
	}
}

func (w *benchmarkResponseWriter) Header() http.Header {
	return w.header
}

func (w *benchmarkResponseWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}

	w.bytes += len(p)

	return len(p), nil
}

func (w *benchmarkResponseWriter) WriteHeader(status int) {
	w.status = status
}

func (w *benchmarkResponseWriter) reset() {
	for key := range w.header {
		delete(w.header, key)
	}

	w.status = 0
	w.bytes = 0
}

func resetActiveProcessorForBenchmark() {
	activeMu.Lock()
	activeProcessor = nil
	activeMu.Unlock()
}
