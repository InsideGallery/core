package prometheus

import (
	"strconv"
	"strings"
	"sync"
	"testing"

	dto "github.com/prometheus/client_model/go"

	"github.com/InsideGallery/core/metrics"
)

func TestMemoizedRecordingMatchesRegistry(t *testing.T) {
	cases := []struct {
		name    string
		metric  string
		tags    []string
		record  func(metrics.Processor, string, []string) error
		verify  func(*testing.T, *processor, string)
		records int
	}{
		{
			name:   "count sums every record",
			metric: "cached.requests",
			tags:   []string{"status:200", "method:GET"},
			record: func(processor metrics.Processor, metric string, tags []string) error {
				return processor.Count(metric, 2, tags)
			},
			records: 3,
			verify: func(t *testing.T, processor *processor, metric string) {
				t.Helper()

				if got := counterValue(t, processor, metric); got != 6 {
					t.Fatalf("counter = %v, want 6", got)
				}
			},
		},
		{
			name:   "gauge keeps the last record",
			metric: "cached.sessions",
			tags:   []string{"site:42"},
			record: func(processor metrics.Processor, metric string, tags []string) error {
				return processor.Gauge(metric, 7, tags)
			},
			records: 3,
			verify: func(t *testing.T, processor *processor, metric string) {
				t.Helper()

				if got := gaugeValue(t, processor, metric); got != 7 {
					t.Fatalf("gauge = %v, want 7", got)
				}
			},
		},
		{
			name:   "distribution observes every record",
			metric: "cached.duration",
			tags:   []string{"route:lookup"},
			record: func(processor metrics.Processor, metric string, tags []string) error {
				return processor.Distribution(metric, 12.5, tags)
			},
			records: 3,
			verify: func(t *testing.T, processor *processor, metric string) {
				t.Helper()

				families := gather(t, processor)
				histogram := requireHistogram(t, families, metric)

				if histogram.GetSampleCount() != 3 {
					t.Fatalf("SampleCount = %d, want 3", histogram.GetSampleCount())
				}

				if histogram.GetSampleSum() != 37.5 {
					t.Fatalf("SampleSum = %v, want 37.5", histogram.GetSampleSum())
				}
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			processor := newTestProcessor(t)

			for range testCase.records {
				if err := testCase.record(processor, testCase.metric, testCase.tags); err != nil {
					t.Fatalf("record error: %v", err)
				}
			}

			sanitized := sanitizeName(testCase.metric)

			requireMetricWithLabels(t, gather(t, processor), sanitized, expectedLabels(testCase.tags))
			testCase.verify(t, processor, sanitized)
		})
	}
}

// TestRecordingIsAllocationFreeWhenMemoized is the reason the memo exists:
// recording an already-seen tuple must not allocate, so a caller on a data path
// pays no label-resolution cost per operation.
func TestRecordingIsAllocationFreeWhenMemoized(t *testing.T) {
	const runs = 1000

	tags := []string{"status:200", "method:GET", "route:lookup"}

	cases := []struct {
		name   string
		record func(metrics.Processor) error
	}{
		{
			name: "count",
			record: func(processor metrics.Processor) error {
				return processor.Count("alloc.requests", 1, tags)
			},
		},
		{
			name: "gauge",
			record: func(processor metrics.Processor) error {
				return processor.Gauge("alloc.sessions", 3, tags)
			},
		},
		{
			name: "distribution",
			record: func(processor metrics.Processor) error {
				return processor.Distribution("alloc.duration", 5, tags)
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			processor := newTestProcessor(t)

			// Resolve and memoize the tuple; only repeat records are measured.
			if err := testCase.record(processor); err != nil {
				t.Fatalf("warm record error: %v", err)
			}

			allocs := testing.AllocsPerRun(runs, func() {
				if err := testCase.record(processor); err != nil {
					t.Fatalf("record error: %v", err)
				}
			})

			if allocs != 0 {
				t.Fatalf("allocations = %v, want 0", allocs)
			}
		})
	}
}

// TestTagOrderVariantsShareOneSeries locks in that the memo is a lookup cache,
// not a second source of truth: tags supplied in a different order are still
// normalized to the same series.
func TestTagOrderVariantsShareOneSeries(t *testing.T) {
	processor := newTestProcessor(t)

	for _, tags := range [][]string{
		{"status:200", "method:GET"},
		{"method:GET", "status:200"},
	} {
		if err := processor.Count("order.requests", 1, tags); err != nil {
			t.Fatalf("Count() error: %v", err)
		}
	}

	families := gather(t, processor)

	requireMetricWithLabels(t, families, "order_requests", map[string]string{
		"method":  "GET",
		"status":  "200",
		"service": "test-svc",
	})

	if got := seriesCount(t, processor, "order_requests"); got != 1 {
		t.Fatalf("series = %d, want 1", got)
	}

	if got := counterValue(t, processor, "order_requests"); got != 2 {
		t.Fatalf("counter = %v, want 2", got)
	}
}

// TestWideTagListsRecordWithoutMemo covers the fallback: a tag list too wide for
// the fixed-size cache key still records, it just resolves its labels per call.
func TestWideTagListsRecordWithoutMemo(t *testing.T) {
	processor := newTestProcessor(t)

	tags := make([]string, 0, maxCachedTagCount+1)
	for index := range maxCachedTagCount + 1 {
		tags = append(tags, "label"+strconv.Itoa(index)+":value"+strconv.Itoa(index))
	}

	for range 2 {
		if err := processor.Count("wide.requests", 1, tags); err != nil {
			t.Fatalf("Count() error: %v", err)
		}
	}

	if got := processor.counterHandles.size(); got != 0 {
		t.Fatalf("memoized handles = %d, want 0", got)
	}

	if got := counterValue(t, processor, "wide_requests"); got != 2 {
		t.Fatalf("counter = %v, want 2", got)
	}
}

// TestHandleCacheEvictsAtCapacity keeps caller-supplied tag values from growing
// the memo without bound, and keeps the bound from sealing the memo shut: at the
// limit a newly recorded tuple displaces a resident instead of being refused
// forever. The limit is small, so its admission interval is one attempt and every
// store evicts.
func TestHandleCacheEvictsAtCapacity(t *testing.T) {
	t.Parallel()

	const stores = 5

	var cache handleCache[int]

	cache.limit = 2

	for index := range stores {
		cache.store("metric", []string{"index:" + strconv.Itoa(index)}, index)
	}

	if got := cache.size(); got != cache.limit {
		t.Fatalf("memoized handles = %d, want %d", got, cache.limit)
	}

	// The last tuple stored is the one certain to be resident: it was just
	// published, and nothing has been stored since to displace it.
	if _, ok := cache.load("metric", []string{"index:" + strconv.Itoa(stores-1)}); !ok {
		t.Fatal("expected the most recently stored tuple to be memoized")
	}

	if got, want := cache.Stats().Evictions, int64(stores-cache.limit); got != want {
		t.Fatalf("Evictions = %d, want %d", got, want)
	}

	if got := cache.Stats().BypassFull; got != 0 {
		t.Fatalf("BypassFull = %d, want 0 while every attempt is admitted", got)
	}
}

// TestHandleCacheRationsAdmission covers the other half of the policy: a full
// memo whose rebuild is expensive turns most first-seen tuples away, so an
// unbounded label cannot make every record rebuild the map. Only the attempt that
// exhausts the interval is admitted.
func TestHandleCacheRationsAdmission(t *testing.T) {
	t.Parallel()

	var cache handleCache[int]

	cache.limit = 2 * admissionIntervalDivisor

	interval := cache.admissionInterval()
	if interval < 2 {
		t.Fatalf("admissionInterval = %d, want at least 2 for this test", interval)
	}

	for index := range cache.limit {
		cache.store("metric", []string{"index:" + strconv.Itoa(index)}, index)
	}

	// One attempt short of the interval: every attempt is turned away, and the
	// memo has published nothing new.
	for attempt := range interval - 1 {
		cache.store("metric", []string{"turned:" + strconv.Itoa(attempt)}, attempt)
	}

	stats := cache.Stats()

	if got, want := stats.BypassFull, int64(interval-1); got != want {
		t.Fatalf("BypassFull = %d, want %d", got, want)
	}

	if stats.Evictions != 0 {
		t.Fatalf("Evictions = %d, want 0 before the interval is exhausted", stats.Evictions)
	}

	admitted := []string{"admitted:0"}
	cache.store("metric", admitted, -1)

	if _, ok := cache.load("metric", admitted); !ok {
		t.Fatal("expected the attempt that exhausts the interval to be memoized")
	}

	if got := cache.Stats().Evictions; got != 1 {
		t.Fatalf("Evictions = %d, want 1", got)
	}

	if got := cache.size(); got != cache.limit {
		t.Fatalf("memoized handles = %d, want %d: an admission replaces, it does not grow", got, cache.limit)
	}
}

// TestHandleCacheStatsCountBypasses pins each counter to the one path it stands
// for, without going through a processor.
func TestHandleCacheStatsCountBypasses(t *testing.T) {
	t.Parallel()

	var cache handleCache[int]

	cache.limit = 1

	if _, ok := cache.load("metric", make([]string, maxCachedTagCount+1)); ok {
		t.Fatal("expected a tag list wider than the key to miss the cache")
	}

	cache.store("metric", []string{"index:0"}, 0)
	cache.store("metric", []string{"index:1"}, 1)

	stats := cache.Stats()

	if stats.Size != cache.limit {
		t.Fatalf("Size = %d, want %d", stats.Size, cache.limit)
	}

	if stats.BypassWidth != 1 {
		t.Fatalf("BypassWidth = %d, want 1", stats.BypassWidth)
	}

	// A capacity of one admits every attempt, so nothing is turned away and the
	// second store evicts the first.
	if stats.BypassFull != 0 {
		t.Fatalf("BypassFull = %d, want 0", stats.BypassFull)
	}

	if stats.Evictions != 1 {
		t.Fatalf("Evictions = %d, want 1", stats.Evictions)
	}
}

// TestReStoringAResidentTupleEvictsNothing covers the race the memo has always
// had: two records of a first-seen tuple can both miss load and both reach
// store. The loser must not evict a resident to re-publish an entry that is
// already there.
func TestReStoringAResidentTupleEvictsNothing(t *testing.T) {
	t.Parallel()

	var cache handleCache[int]

	cache.limit = 2

	resident := []string{"index:0"}

	cache.store("metric", resident, 0)
	cache.store("metric", []string{"index:1"}, 1)
	cache.store("metric", resident, 0)

	if got := cache.size(); got != cache.limit {
		t.Fatalf("memoized handles = %d, want %d", got, cache.limit)
	}

	if got := cache.Stats().Evictions; got != 0 {
		t.Fatalf("Evictions = %d, want 0", got)
	}

	for _, tags := range [][]string{resident, {"index:1"}} {
		if _, ok := cache.load("metric", tags); !ok {
			t.Fatalf("expected %v to stay memoized", tags)
		}
	}
}

// TestEvictedTupleRecordsAndIsReMemoized is the behaviour the policy change
// exists for: a tuple the memo dropped keeps recording correctly, and comes back
// into the memo instead of resolving its labels for the life of the process.
// Every record must be counted exactly once across the eviction.
func TestEvictedTupleRecordsAndIsReMemoized(t *testing.T) {
	const (
		floodLimit = 100
		hotRecords = 3
	)

	processor := newTestProcessor(t)

	// A small bound so its admission interval is one attempt: each flooding
	// record evicts, and the hot tuple is displaced within a few of them.
	processor.counterHandles.limit = 2

	hot := []string{"route:hot"}

	if err := processor.Count("evicted.requests", 1, hot); err != nil {
		t.Fatalf("Count() error: %v", err)
	}

	flooded := 0

	for ; flooded < floodLimit; flooded++ {
		if _, resident := processor.counterHandles.load("evicted.requests", hot); !resident {
			break
		}

		if err := processor.Count("flood.requests", 1, []string{"index:" + strconv.Itoa(flooded)}); err != nil {
			t.Fatalf("Count() error: %v", err)
		}
	}

	if flooded == floodLimit {
		t.Fatalf("the hot tuple survived %d evictions; expected it to be displaced", flooded)
	}

	for range hotRecords - 1 {
		if err := processor.Count("evicted.requests", 1, hot); err != nil {
			t.Fatalf("Count() error: %v", err)
		}
	}

	if _, resident := processor.counterHandles.load("evicted.requests", hot); !resident {
		t.Fatal("expected the evicted tuple to be memoized again")
	}

	if got := counterValue(t, processor, "evicted_requests"); got != hotRecords {
		t.Fatalf("counter = %v, want %d: records must survive eviction exactly once each", got, hotRecords)
	}

	if got := seriesCount(t, processor, "evicted_requests"); got != 1 {
		t.Fatalf("series = %d, want 1", got)
	}
}

// TestScrapeExportsEveryHandleCacheKind locks the collector's registration, its
// types, and its label set: all three caches report on every scrape from the
// first one, so a dashboard can rely on the series existing before saturation.
func TestScrapeExportsEveryHandleCacheKind(t *testing.T) {
	newTestProcessor(t)

	body := scrapeBody(t)

	for _, want := range []string{
		"# TYPE " + handleCacheSizeMetric + " gauge",
		"# TYPE " + handleCacheBypassMetric + " counter",
		"# TYPE " + handleCacheEvictionsMetric + " counter",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q in:\n%s", want, body)
		}
	}

	for _, kind := range []string{cacheKindCounter, cacheKindGauge, cacheKindHistogram} {
		kindFragment := kindLabel + `="` + kind + `"`

		if got := scrapedValue(t, body, handleCacheSizeMetric, kindFragment, `service="test-svc"`); got != 0 {
			t.Fatalf("%s{kind=%s} = %v, want 0", handleCacheSizeMetric, kind, got)
		}

		if got := scrapedValue(t, body, handleCacheEvictionsMetric, kindFragment, `service="test-svc"`); got != 0 {
			t.Fatalf("%s{kind=%s} = %v, want 0", handleCacheEvictionsMetric, kind, got)
		}

		for _, reason := range []string{bypassReasonWidth, bypassReasonFull} {
			reasonFragment := reasonLabel + `="` + reason + `"`

			if got := scrapedValue(t, body, handleCacheBypassMetric, kindFragment, reasonFragment); got != 0 {
				t.Fatalf("%s{kind=%s,reason=%s} = %v, want 0", handleCacheBypassMetric, kind, reason, got)
			}
		}
	}
}

// TestScrapeExportsHandleCacheSaturation is why the counters exist: once the memo
// fills, the tuples it turns away resolve their labels on every record until they
// win an admission, and nothing about that is observable from outside the package
// unless the scrape says so. Recording exactly one admission interval past the
// bound pins both halves of the policy at the real capacity: the interval's worth
// of turned-away records, then one admission that replaces a resident rather than
// growing the memo.
func TestScrapeExportsHandleCacheSaturation(t *testing.T) {
	processor := newTestProcessor(t)

	interval := processor.gaugeHandles.admissionInterval()

	for index := range maxCachedHandles + interval {
		if err := processor.Gauge("saturating.sessions", 1, []string{"index:" + strconv.Itoa(index)}); err != nil {
			t.Fatalf("Gauge() error: %v", err)
		}
	}

	body := scrapeBody(t)
	gaugeKind := kindLabel + `="` + cacheKindGauge + `"`

	if got := scrapedValue(t, body, handleCacheSizeMetric, gaugeKind); got != maxCachedHandles {
		t.Fatalf("%s{kind=gauge} = %v, want %d", handleCacheSizeMetric, got, maxCachedHandles)
	}

	fullReason := reasonLabel + `="` + bypassReasonFull + `"`

	if got, want := scrapedValue(t, body, handleCacheBypassMetric, gaugeKind, fullReason), float64(interval-1); got != want {
		t.Fatalf("%s{kind=gauge,reason=full} = %v, want %v", handleCacheBypassMetric, got, want)
	}

	if got := scrapedValue(t, body, handleCacheEvictionsMetric, gaugeKind); got != 1 {
		t.Fatalf("%s{kind=gauge} = %v, want 1", handleCacheEvictionsMetric, got)
	}
}

// TestScrapeExportsWideTagBypass covers the other fallback: a tag list too wide
// for the fixed-size key records correctly but is never memoized, so it pays
// full label resolution on every record.
func TestScrapeExportsWideTagBypass(t *testing.T) {
	const records = 2

	processor := newTestProcessor(t)

	tags := make([]string, 0, maxCachedTagCount+1)
	for index := range maxCachedTagCount + 1 {
		tags = append(tags, "label"+strconv.Itoa(index)+":value"+strconv.Itoa(index))
	}

	for range records {
		if err := processor.Count("wide.bypass.requests", 1, tags); err != nil {
			t.Fatalf("Count() error: %v", err)
		}
	}

	body := scrapeBody(t)
	counterKind := kindLabel + `="` + cacheKindCounter + `"`
	widthReason := reasonLabel + `="` + bypassReasonWidth + `"`

	if got := scrapedValue(t, body, handleCacheBypassMetric, counterKind, widthReason); got != records {
		t.Fatalf("%s{kind=counter,reason=width} = %v, want %d", handleCacheBypassMetric, got, records)
	}
}

func TestHandleCacheDefaultCapacity(t *testing.T) {
	t.Parallel()

	var cache handleCache[int]

	if got := cache.capacity(); got != maxCachedHandles {
		t.Fatalf("capacity = %d, want %d", got, maxCachedHandles)
	}
}

// TestConcurrentRecordingKeepsEveryRecord exercises the memo the way a data path
// does — many goroutines racing on the same first-seen tuples — so that a
// snapshot published while others are reading cannot lose a record.
func TestConcurrentRecordingKeepsEveryRecord(t *testing.T) {
	const (
		writers          = 8
		recordsPerWriter = 200
	)

	processor := newTestProcessor(t)

	tagSets := [][]string{
		{"status:200", "method:GET"},
		{"status:500", "method:POST"},
	}

	var waitGroup sync.WaitGroup

	for writer := range writers {
		waitGroup.Add(1)

		go func(writer int) {
			defer waitGroup.Done()

			tags := tagSets[writer%len(tagSets)]

			for range recordsPerWriter {
				if err := processor.Count("concurrent.requests", 1, tags); err != nil {
					t.Errorf("Count() error: %v", err)

					return
				}
			}
		}(writer)
	}

	waitGroup.Wait()

	if got := seriesCount(t, processor, "concurrent_requests"); got != len(tagSets) {
		t.Fatalf("series = %d, want %d", got, len(tagSets))
	}

	var total float64

	for _, series := range familyMetrics(t, processor, "concurrent_requests") {
		total += series.GetCounter().GetValue()
	}

	if want := float64(writers * recordsPerWriter); total != want {
		t.Fatalf("recorded total = %v, want %v", total, want)
	}
}

// TestConcurrentRecordingSurvivesEviction runs the same race with a memo far too
// small for the tuple space, so readers are looking up handles while store keeps
// republishing snapshots that drop entries. Every record must still land exactly
// once: a snapshot must never be mutated after publication.
func TestConcurrentRecordingSurvivesEviction(t *testing.T) {
	const (
		writers          = 8
		recordsPerWriter = 200
		tupleSpace       = 32
		cacheLimit       = 4
	)

	processor := newTestProcessor(t)
	processor.counterHandles.limit = cacheLimit

	var waitGroup sync.WaitGroup

	for writer := range writers {
		waitGroup.Add(1)

		go func(writer int) {
			defer waitGroup.Done()

			for record := range recordsPerWriter {
				index := (writer*recordsPerWriter + record) % tupleSpace

				if err := processor.Count("evicting.requests", 1, []string{"index:" + strconv.Itoa(index)}); err != nil {
					t.Errorf("Count() error: %v", err)

					return
				}
			}
		}(writer)
	}

	waitGroup.Wait()

	if got := seriesCount(t, processor, "evicting_requests"); got != tupleSpace {
		t.Fatalf("series = %d, want %d", got, tupleSpace)
	}

	var total float64

	for _, series := range familyMetrics(t, processor, "evicting_requests") {
		total += series.GetCounter().GetValue()
	}

	if want := float64(writers * recordsPerWriter); total != want {
		t.Fatalf("recorded total = %v, want %v", total, want)
	}

	if got := processor.counterHandles.size(); got > cacheLimit {
		t.Fatalf("memoized handles = %d, want at most %d", got, cacheLimit)
	}

	if processor.counterHandles.Stats().Evictions == 0 {
		t.Fatal("expected the memo to have evicted under this tuple space")
	}
}

func newTestProcessor(t *testing.T) *processor {
	t.Helper()

	resetActiveProcessor(t)

	rawProcessor, err := New(metrics.Config{}, "test-svc")
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	t.Cleanup(func() {
		if err := rawProcessor.Close(); err != nil {
			t.Fatalf("Close() error: %v", err)
		}

		resetActiveProcessor(t)
	})

	processor, ok := rawProcessor.(*processor)
	if !ok {
		t.Fatalf("processor type = %T", rawProcessor)
	}

	return processor
}

func gather(t *testing.T, processor *processor) []*dto.MetricFamily {
	t.Helper()

	families, err := processor.registry.Gather()
	if err != nil {
		t.Fatalf("Gather() error: %v", err)
	}

	return families
}

// expectedLabels turns "name:value" tags into the label set the processor should
// emit for them, including the constant service label.
func expectedLabels(tags []string) map[string]string {
	labels := map[string]string{"service": "test-svc"}

	for _, tag := range tags {
		name, value, ok := strings.Cut(tag, ":")
		if !ok {
			continue
		}

		labels[sanitizeLabelName(name)] = value
	}

	return labels
}

func familyMetrics(t *testing.T, processor *processor, name string) []*dto.Metric {
	t.Helper()

	for _, family := range gather(t, processor) {
		if family.GetName() == name {
			return family.GetMetric()
		}
	}

	t.Fatalf("missing metric family %q", name)

	return nil
}

func requireSingleMetric(t *testing.T, processor *processor, name string) *dto.Metric {
	t.Helper()

	series := familyMetrics(t, processor, name)
	if len(series) != 1 {
		t.Fatalf("series for %q = %d, want 1", name, len(series))
	}

	return series[0]
}

func counterValue(t *testing.T, processor *processor, name string) float64 {
	t.Helper()

	return requireSingleMetric(t, processor, name).GetCounter().GetValue()
}

func gaugeValue(t *testing.T, processor *processor, name string) float64 {
	t.Helper()

	return requireSingleMetric(t, processor, name).GetGauge().GetValue()
}

func seriesCount(t *testing.T, processor *processor, name string) int {
	t.Helper()

	for _, family := range gather(t, processor) {
		if family.GetName() == name {
			return len(family.GetMetric())
		}
	}

	t.Fatalf("missing metric family %q", name)

	return 0
}

func TestSanitizeNameKeepsAlreadyValidNames(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
		want  string
	}{
		{name: "already valid", input: "frogodb_ops_total", want: "frogodb_ops_total"},
		{name: "dotted", input: "sample.requests", want: "sample_requests"},
		{name: "leading digit", input: "1bad", want: "_1bad"},
		{name: "empty", input: "", want: fallbackMetricName},
		{name: "multi byte rune", input: "héllo", want: "h_llo"},
		{name: "mixed case and digits", input: "Sample9_total", want: "Sample9_total"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := sanitizeName(testCase.input); got != testCase.want {
				t.Fatalf("sanitizeName(%q) = %q, want %q", testCase.input, got, testCase.want)
			}
		})
	}
}

// Resolved handles (MET-11). The properties that matter are that a handle records
// into the SAME series the Count/Gauge/Distribution path would have produced, that
// it allocates nothing, and that resolving one does not consume a memo slot the
// hot path needs.

func TestHandleRecordingMatchesTheRecordPath(t *testing.T) {
	const (
		counterMetric = "handled.requests"
		gaugeMetric   = "handled.sessions"
		histoMetric   = "handled.duration"
		records       = 3
	)

	tags := []string{"status:200", "method:GET"}

	recorded := newTestProcessor(t)
	handled := newTestProcessor(t)

	for range records {
		if err := recorded.Count(counterMetric, 2, tags); err != nil {
			t.Fatalf("Count() error: %v", err)
		}

		if err := recorded.Gauge(gaugeMetric, 7, tags); err != nil {
			t.Fatalf("Gauge() error: %v", err)
		}

		if err := recorded.Distribution(histoMetric, 0.5, tags); err != nil {
			t.Fatalf("Distribution() error: %v", err)
		}
	}

	counter, err := handled.CounterHandle(counterMetric, tags)
	if err != nil {
		t.Fatalf("CounterHandle() error: %v", err)
	}

	gauge, err := handled.GaugeHandle(gaugeMetric, tags)
	if err != nil {
		t.Fatalf("GaugeHandle() error: %v", err)
	}

	observer, err := handled.DistributionHandle(histoMetric, tags)
	if err != nil {
		t.Fatalf("DistributionHandle() error: %v", err)
	}

	for range records {
		counter.Add(2)
		gauge.Set(7)
		observer.Observe(0.5)
	}

	for _, metric := range []string{"handled_requests", "handled_sessions", "handled_duration"} {
		requireSameSeries(t, recorded, handled, metric)
	}
}

// requireSameSeries fails unless both processors emit the metric family with the
// same series: same label names, same label values, same recorded value. The
// created timestamp is excluded — it records when each processor registered its
// collector, not what was recorded.
func requireSameSeries(t *testing.T, want, got *processor, name string) {
	t.Helper()

	wantSeries := familyMetrics(t, want, name)
	gotSeries := familyMetrics(t, got, name)

	if len(wantSeries) != len(gotSeries) {
		t.Fatalf("%s series = %d, want %d", name, len(gotSeries), len(wantSeries))
	}

	for index := range wantSeries {
		wantText := seriesIdentity(wantSeries[index])

		if gotText := seriesIdentity(gotSeries[index]); gotText != wantText {
			t.Fatalf("%s series %d:\n got %s\nwant %s", name, index, gotText, wantText)
		}
	}
}

// seriesIdentity renders one series as its label set plus its recorded value.
func seriesIdentity(series *dto.Metric) string {
	var builder strings.Builder

	for _, label := range series.GetLabel() {
		builder.WriteString(label.GetName() + "=" + label.GetValue() + " ")
	}

	switch {
	case series.GetCounter() != nil:
		builder.WriteString("counter=" + strconv.FormatFloat(series.GetCounter().GetValue(), 'g', -1, 64))
	case series.GetGauge() != nil:
		builder.WriteString("gauge=" + strconv.FormatFloat(series.GetGauge().GetValue(), 'g', -1, 64))
	case series.GetHistogram() != nil:
		histogram := series.GetHistogram()
		builder.WriteString("histogram count=" + strconv.FormatUint(histogram.GetSampleCount(), 10) +
			" sum=" + strconv.FormatFloat(histogram.GetSampleSum(), 'g', -1, 64))

		for _, bucket := range histogram.GetBucket() {
			builder.WriteString(" le=" + strconv.FormatFloat(bucket.GetUpperBound(), 'g', -1, 64) +
				":" + strconv.FormatUint(bucket.GetCumulativeCount(), 10))
		}
	}

	return builder.String()
}

func TestHandleRecordingIsAllocationFree(t *testing.T) {
	const runs = 1000

	tags := []string{"status:200", "method:GET", "route:lookup"}

	processor := newTestProcessor(t)

	counter, err := processor.CounterHandle("handle.alloc.requests", tags)
	if err != nil {
		t.Fatalf("CounterHandle() error: %v", err)
	}

	gauge, err := processor.GaugeHandle("handle.alloc.sessions", tags)
	if err != nil {
		t.Fatalf("GaugeHandle() error: %v", err)
	}

	observer, err := processor.DistributionHandle("handle.alloc.duration", tags)
	if err != nil {
		t.Fatalf("DistributionHandle() error: %v", err)
	}

	cases := []struct {
		name   string
		record func()
	}{
		{name: "counter", record: func() { counter.Add(1) }},
		{name: "gauge", record: func() { gauge.Set(3) }},
		{name: "observer", record: func() { observer.Observe(5) }},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if allocs := testing.AllocsPerRun(runs, testCase.record); allocs != 0 {
				t.Fatalf("allocations = %v, want 0", allocs)
			}
		})
	}
}

// A handle never looks its tuple up again, so an entry for it would occupy one of
// the bounded memo slots without ever being read.
func TestHandleResolutionLeavesTheMemoEmpty(t *testing.T) {
	processor := newTestProcessor(t)

	tags := []string{"status:200"}

	counter, err := processor.CounterHandle("memo.free.requests", tags)
	if err != nil {
		t.Fatalf("CounterHandle() error: %v", err)
	}

	if _, err := processor.GaugeHandle("memo.free.sessions", tags); err != nil {
		t.Fatalf("GaugeHandle() error: %v", err)
	}

	if _, err := processor.DistributionHandle("memo.free.duration", tags); err != nil {
		t.Fatalf("DistributionHandle() error: %v", err)
	}

	counter.Add(1)

	for _, cache := range []struct {
		kind string
		size int
	}{
		{kind: cacheKindCounter, size: processor.counterHandles.size()},
		{kind: cacheKindGauge, size: processor.gaugeHandles.size()},
		{kind: cacheKindHistogram, size: processor.histogramHandles.size()},
	} {
		if cache.size != 0 {
			t.Fatalf("%s memo size = %d, want 0", cache.kind, cache.size)
		}
	}
}

// Count rejects a negative value with an error; a handle has no error channel and
// stdprom.Counter.Add panics on a negative, so the handle drops it.
func TestCounterHandleDropsNegativeValues(t *testing.T) {
	processor := newTestProcessor(t)

	counter, err := processor.CounterHandle("negative.requests", []string{"status:200"})
	if err != nil {
		t.Fatalf("CounterHandle() error: %v", err)
	}

	counter.Add(4)
	counter.Add(-1)

	if got := counterValue(t, processor, "negative_requests"); got != 4 {
		t.Fatalf("counter = %v, want 4", got)
	}
}

// The Prometheus processor must be usable as a handle provider through the client
// abstraction, since that is how every caller reaches it.
func TestProcessorSatisfiesHandleProvider(t *testing.T) {
	processor := newTestProcessor(t)

	if _, ok := metrics.Processor(processor).(metrics.HandleProvider); !ok {
		t.Fatal("processor does not implement metrics.HandleProvider")
	}
}

// BenchmarkResolvedHandleVersusMemoizedRecord is the A/B for what a resolved handle
// removes: both arms record the same tuple into the same processor in one process,
// one through Count (memo hit: hash a 112-byte key, map lookup, interface call)
// and one through a handle resolved up front.
//
//	go test -run=^$ -bench=BenchmarkResolvedHandleVersusMemoizedRecord -benchmem \
//	    ./metrics/processors/prometheus/
func BenchmarkResolvedHandleVersusMemoizedRecord(b *testing.B) {
	const metric = "bench.handle.requests"

	tags := []string{"status:200", "method:GET", "route:lookup"}

	resetActiveProcessorForBenchmark()

	rawProcessor, err := New(metrics.Config{}, "bench-svc")
	if err != nil {
		b.Fatalf("New() error: %v", err)
	}

	b.Cleanup(func() {
		if err := rawProcessor.Close(); err != nil {
			b.Fatalf("Close() error: %v", err)
		}

		resetActiveProcessorForBenchmark()
	})

	promProcessor, ok := rawProcessor.(*processor)
	if !ok {
		b.Fatalf("processor type = %T", rawProcessor)
	}

	// Warm the memo so the Count arm measures a hit, not a resolution.
	if err := promProcessor.Count(metric, 1, tags); err != nil {
		b.Fatalf("warm Count() error: %v", err)
	}

	counter, err := promProcessor.CounterHandle(metric, tags)
	if err != nil {
		b.Fatalf("CounterHandle() error: %v", err)
	}

	b.Run("memoized_count", func(b *testing.B) {
		b.ReportAllocs()

		for range b.N {
			if err := promProcessor.Count(metric, 1, tags); err != nil {
				b.Fatalf("Count() error: %v", err)
			}
		}
	})

	b.Run("resolved_handle", func(b *testing.B) {
		b.ReportAllocs()

		for range b.N {
			counter.Add(1)
		}
	})
}
