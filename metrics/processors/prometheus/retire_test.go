package prometheus

import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"testing"

	dto "github.com/prometheus/client_model/go"

	"github.com/InsideGallery/core/metrics"
)

// The series shapes retirement exists for: a set of metrics whose label space grows
// with the addresses a process has talked to rather than with the code, one of them
// split by a label the caller cannot enumerate (`reason`), alongside a metric that
// carries no peer label at all and must survive every retirement.
const (
	connectionErrors = "fabric_connection_errors_total"
	sendFailures     = "fabric_send_failures_total"
	poolActive       = "fabric_pool_active"
	sendSeconds      = "fabric_send_seconds"
	sendBytes        = "fabric_send_bytes_total"

	peerLabel    = "peer"
	departedPeer = "10.0.0.7:3001"
	stayingPeer  = "10.0.0.8:3001"
)

// peerSeriesNames are the metrics recordPeerSeries populates for one peer, in the
// three collector kinds, so a retirement has to reach all of them.
var peerSeriesNames = []string{connectionErrors, sendFailures, poolActive, sendSeconds}

func TestDeleteSeriesMatchingRetiresOneSubjectAndNothingElse(t *testing.T) {
	processor := newTestProcessor(t)

	recordPeerSeries(t, processor, departedPeer)
	recordPeerSeries(t, processor, stayingPeer)
	requireRecorded(t, processor.Count(sendBytes, 512, []string{"channel:rw"}))

	retirePeer(t, processor, departedPeer)

	for _, name := range peerSeriesNames {
		if got := seriesWithLabel(t, processor, name, peerLabel, departedPeer); got != 0 {
			t.Errorf("%s exports %d series for the departed peer, want 0", name, got)
		}

		if got := seriesWithLabel(t, processor, name, peerLabel, stayingPeer); got == 0 {
			t.Errorf("%s retired the series of a peer that did not depart", name)
		}
	}

	// send_failures_total is split by an open-ended `reason`, which is why the match
	// is partial: both of the departed peer's reasons had to go, and both of the
	// staying peer's had to stay.
	if got := seriesWithLabel(t, processor, sendFailures, peerLabel, stayingPeer); got != 2 {
		t.Errorf("send failures for the staying peer = %d, want 2", got)
	}

	if got := len(exportedSeries(t, processor, sendBytes)); got != 1 {
		t.Errorf("series without a peer label = %d, want 1: retirement swept a series it does not own", got)
	}
}

// A retired subject that comes back — a peer rejoining at the same address — must
// record into a fresh series. Recreating the child is not enough on its own: the memo
// still holds the deleted one, and a record that hits the memo would be absorbed by a
// child no scrape can reach.
func TestRetiredSeriesIsRecreatedFromZero(t *testing.T) {
	cases := []struct {
		name   string
		metric string
		record func(t *testing.T, processor *processor, metric string, value float64)
		read   func(t *testing.T, processor *processor, metric string) float64
	}{
		{
			name:   "counter",
			metric: connectionErrors,
			record: func(t *testing.T, processor *processor, metric string, value float64) {
				t.Helper()
				requireRecorded(t, processor.Count(metric, int64(value), peerTags(departedPeer)))
			},
			read: func(t *testing.T, processor *processor, metric string) float64 {
				t.Helper()

				return counterValue(t, processor, metric)
			},
		},
		{
			name:   "gauge",
			metric: poolActive,
			record: func(t *testing.T, processor *processor, metric string, value float64) {
				t.Helper()
				requireRecorded(t, processor.Gauge(metric, value, peerTags(departedPeer)))
			},
			read: func(t *testing.T, processor *processor, metric string) float64 {
				t.Helper()

				return gaugeValue(t, processor, metric)
			},
		},
		{
			name:   "histogram",
			metric: sendSeconds,
			record: func(t *testing.T, processor *processor, metric string, value float64) {
				t.Helper()
				requireRecorded(t, processor.Distribution(metric, value, peerTags(departedPeer)))
			},
			read: func(t *testing.T, processor *processor, metric string) float64 {
				t.Helper()

				return float64(requireSingleMetric(t, processor, metric).GetHistogram().GetSampleCount())
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			processor := newTestProcessor(t)

			// Twice, so a value that survived the delete is distinguishable from a
			// value recorded after it.
			testCase.record(t, processor, testCase.metric, 1)
			testCase.record(t, processor, testCase.metric, 1)

			if err := processor.DeleteSeriesMatching(testCase.metric, peerTags(departedPeer)); err != nil {
				t.Fatalf("DeleteSeriesMatching() error: %v", err)
			}

			if got := len(exportedSeries(t, processor, testCase.metric)); got != 0 {
				t.Fatalf("series after retirement = %d, want 0", got)
			}

			testCase.record(t, processor, testCase.metric, 1)

			if got := testCase.read(t, processor, testCase.metric); got != 1 {
				t.Fatalf("value after the subject returned = %v, want 1: stale residue from before the retirement", got)
			}
		})
	}
}

// The memo is what makes a repeat record cheap, so it is also what holds the
// strongest reference to a deleted child. Retirement has to leave it holding only
// the tuples that still exist.
func TestRetirementForgetsExactlyTheRetiredTuples(t *testing.T) {
	processor := newTestProcessor(t)

	recordPeerSeries(t, processor, departedPeer)
	recordPeerSeries(t, processor, stayingPeer)

	// Per peer: one connection-error tuple and two send-failure reasons in the
	// counter memo, one pool gauge, one duration observation.
	const (
		counterTuplesPerPeer   = 3
		gaugeTuplesPerPeer     = 1
		histogramTuplesPerPeer = 1
	)

	requireMemoSizes(t, processor, "before retirement",
		2*counterTuplesPerPeer, 2*gaugeTuplesPerPeer, 2*histogramTuplesPerPeer)

	retirePeer(t, processor, departedPeer)

	requireMemoSizes(t, processor, "after retirement",
		counterTuplesPerPeer, gaugeTuplesPerPeer, histogramTuplesPerPeer)
}

// An empty match selects every child of a metric, so it is refused. The case that
// makes this matter is not a caller typing nil: it is a caller assembling the tag
// from an empty subject — a blank address, a missing tenant — and wiping the whole
// metric family for every subject at once.
func TestDeleteSeriesMatchingRefusesAnEmptyMatch(t *testing.T) {
	cases := []struct {
		name string
		tags []string
	}{
		{name: "nil tags"},
		{name: "empty tags", tags: []string{}},
		{name: "tag without a value separator", tags: []string{"peer"}},
		{name: "tag without a name", tags: []string{":10.0.0.7:3001"}},
		{name: "empty tag", tags: []string{""}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			processor := newTestProcessor(t)

			recordPeerSeries(t, processor, departedPeer)
			recordPeerSeries(t, processor, stayingPeer)

			err := processor.DeleteSeriesMatching(connectionErrors, testCase.tags)
			if !errors.Is(err, metrics.ErrEmptySeriesMatch) {
				t.Fatalf("DeleteSeriesMatching() error = %v, want %v", err, metrics.ErrEmptySeriesMatch)
			}

			if got := len(exportedSeries(t, processor, connectionErrors)); got != 2 {
				t.Fatalf("series after a refused match = %d, want 2", got)
			}
		})
	}
}

// Retirement runs from lifecycle events, which fire whether or not the subject ever
// recorded anything and can fire twice for the same departure. Neither is a failure.
func TestDeleteSeriesMatchingIsIdempotent(t *testing.T) {
	processor := newTestProcessor(t)

	recordPeerSeries(t, processor, departedPeer)
	recordPeerSeries(t, processor, stayingPeer)

	if err := processor.DeleteSeriesMatching("metric_never_recorded_total", peerTags(departedPeer)); err != nil {
		t.Fatalf("DeleteSeriesMatching() on an unrecorded metric error: %v", err)
	}

	if err := processor.DeleteSeriesMatching(connectionErrors, peerTags("10.0.0.99:3001")); err != nil {
		t.Fatalf("DeleteSeriesMatching() on an unrecorded subject error: %v", err)
	}

	retirePeer(t, processor, departedPeer)
	retirePeer(t, processor, departedPeer)

	if got := seriesWithLabel(t, processor, connectionErrors, peerLabel, stayingPeer); got != 1 {
		t.Fatalf("series for the staying peer = %d, want 1", got)
	}
}

// The memo keys on the caller's raw tags while the vec keys on resolved labels, so
// both sides resolve before they compare. A tag name that sanitizes reaches the same
// label from two different spellings: the delete must forget the memoized child it
// just deleted, not the one whose raw string happens to be identical.
func TestRetirementMatchesResolvedLabelsNotRawTags(t *testing.T) {
	processor := newTestProcessor(t)

	// "peer.id" and "peer_id" sanitize to the same label name.
	requireRecorded(t, processor.Count(connectionErrors, 4, []string{"peer.id:" + departedPeer}))

	if err := processor.DeleteSeriesMatching(connectionErrors, []string{"peer_id:" + departedPeer}); err != nil {
		t.Fatalf("DeleteSeriesMatching() error: %v", err)
	}

	if got := len(exportedSeries(t, processor, connectionErrors)); got != 0 {
		t.Fatalf("series after retirement = %d, want 0", got)
	}

	requireRecorded(t, processor.Count(connectionErrors, 1, []string{"peer.id:" + departedPeer}))

	if got := counterValue(t, processor, connectionErrors); got != 1 {
		t.Fatalf("counter after the subject returned = %v, want 1: the memo kept the deleted child", got)
	}
}

// Documented caveat, pinned so it changes deliberately: a handle resolved through
// metrics.HandleProvider holds the child directly, so retiring the series orphans the
// handle — records through it are accepted and exported nowhere. It is why a series a
// caller intends to retire belongs on the Count/Gauge/Distribution path.
func TestRetirementOrphansAHandleResolvedBeforeIt(t *testing.T) {
	processor := newTestProcessor(t)

	handle, err := processor.CounterHandle(connectionErrors, peerTags(departedPeer))
	if err != nil {
		t.Fatalf("CounterHandle() error: %v", err)
	}

	handle.Add(1)

	if got := counterValue(t, processor, connectionErrors); got != 1 {
		t.Fatalf("counter before retirement = %v, want 1", got)
	}

	retirePeer(t, processor, departedPeer)

	handle.Add(1)

	if got := len(exportedSeries(t, processor, connectionErrors)); got != 0 {
		t.Fatalf("series after recording through an orphaned handle = %d, want 0", got)
	}

	// The recording path is unaffected: it resolves per record and recreates the child.
	requireRecorded(t, processor.Count(connectionErrors, 3, peerTags(departedPeer)))

	if got := counterValue(t, processor, connectionErrors); got != 3 {
		t.Fatalf("counter after the recording path returned = %v, want 3", got)
	}
}

// Retirement runs on a lifecycle event while data paths keep recording, so the two
// races by construction. Under -race this asserts the only guarantees the design
// makes: no torn state, and a subject that keeps recording ends up exported.
func TestConcurrentRecordingSurvivesRetirement(t *testing.T) {
	processor := newTestProcessor(t)

	const (
		recorders  = 8
		iterations = 200
	)

	stop := make(chan struct{})

	var waiting sync.WaitGroup

	waiting.Add(recorders)

	for index := range recorders {
		go func(peer string) {
			defer waiting.Done()

			for range iterations {
				_ = processor.Count(connectionErrors, 1, peerTags(peer))
				_ = processor.Gauge(poolActive, 1, peerTags(peer))
			}
		}("10.0.1." + strconv.Itoa(index) + ":3001")
	}

	var retiring sync.WaitGroup

	retiring.Add(1)

	go func() {
		defer retiring.Done()

		for {
			select {
			case <-stop:
				return
			default:
				for index := range recorders {
					peer := "10.0.1." + strconv.Itoa(index) + ":3001"
					_ = processor.DeleteSeriesMatching(connectionErrors, peerTags(peer))
					_ = processor.DeleteSeriesMatching(poolActive, peerTags(peer))
				}
			}
		}
	}()

	waiting.Wait()
	close(stop)
	retiring.Wait()

	requireRecorded(t, processor.Count(connectionErrors, 1, peerTags(departedPeer)))

	if got := seriesWithLabel(t, processor, connectionErrors, peerLabel, departedPeer); got != 1 {
		t.Fatalf("series recorded after the churn = %d, want 1", got)
	}
}

// The Prometheus processor must be reachable as a series deleter through the client
// abstraction, and through the bare interface literal a consumer pinned to an older
// release asserts structurally.
func TestProcessorSatisfiesSeriesDeleter(t *testing.T) {
	processor := newTestProcessor(t)

	var asProcessor metrics.Processor = processor

	if _, ok := asProcessor.(metrics.SeriesDeleter); !ok {
		t.Fatal("processor does not implement metrics.SeriesDeleter")
	}

	if _, ok := asProcessor.(interface {
		DeleteSeriesMatching(name string, tags []string) error
	}); !ok {
		t.Fatal("processor does not satisfy the structural series-delete assertion")
	}
}

// The whole chain a consumer wires in production: metrics.New builds a *Client over
// the registered processor, the consumer detects the capability structurally on that
// *Client — it cannot name metrics.SeriesDeleter while pinned to an older release —
// and the delete has to reach the registry the scrape endpoint reads. Every earlier
// test in this file works on the processor directly, so this is the one that proves
// the seam a caller actually holds.
func TestClientRetiresSeriesThroughTheScrapeEndpoint(t *testing.T) {
	resetActiveProcessor(t)
	t.Setenv("METRICS_PROCESSORS", ProcessorName)

	cfg, err := metrics.GetEnvConfig()
	if err != nil {
		t.Fatalf("GetEnvConfig() error: %v", err)
	}

	client, err := metrics.New(cfg, "test-svc")
	if err != nil {
		t.Fatalf("metrics.New() error: %v", err)
	}

	if client == nil {
		t.Fatal("metrics.New() configured no processor")
	}

	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Fatalf("Close() error: %v", err)
		}

		resetActiveProcessor(t)
	})

	deleter, ok := metrics.Processor(client).(interface {
		DeleteSeriesMatching(name string, tags []string) error
	})
	if !ok {
		t.Fatal("*Client does not satisfy the structural series-delete assertion")
	}

	retirable := []string{connectionErrors, sendFailures, poolActive}

	for _, peer := range []string{departedPeer, stayingPeer} {
		requireRecorded(t, client.Count(connectionErrors, 1, peerTags(peer)))
		requireRecorded(t, client.Count(sendFailures, 1, append(peerTags(peer), "reason:dial")))
		requireRecorded(t, client.Gauge(poolActive, 2, peerTags(peer)))
	}

	for _, name := range retirable {
		if err := deleter.DeleteSeriesMatching(name, []string{peerLabel + ":" + departedPeer}); err != nil {
			t.Fatalf("DeleteSeriesMatching(%q) error: %v", name, err)
		}
	}

	body := scrapeBody(t)

	for _, line := range strings.Split(body, "\n") {
		if strings.Contains(line, departedPeer) {
			t.Errorf("departed peer still exported: %s", line)
		}
	}

	if !strings.Contains(body, stayingPeer) {
		t.Error("the scrape lost a peer that did not depart")
	}

	// A subject that returns at the same identity counts from the next event: the
	// series is present again, and carries only what was recorded after retirement.
	requireRecorded(t, client.Count(connectionErrors, 1, peerTags(departedPeer)))

	if got := scrapedValue(t, scrapeBody(t), connectionErrors, `peer="`+departedPeer+`"`); got != 1 {
		t.Fatalf("counter after the subject returned = %v, want 1", got)
	}
}

func peerTags(peer string) []string {
	return []string{"channel:rw", peerLabel + ":" + peer}
}

// recordPeerSeries populates every peer-labelled metric for one peer: a counter, a
// counter split by an extra open-ended label, a gauge, and a histogram.
func recordPeerSeries(t *testing.T, target *processor, peer string) {
	t.Helper()

	tags := peerTags(peer)

	requireRecorded(t, target.Count(connectionErrors, 1, tags))
	requireRecorded(t, target.Count(sendFailures, 1, append(peerTags(peer), "reason:dial")))
	requireRecorded(t, target.Count(sendFailures, 1, append(peerTags(peer), "reason:timeout")))
	requireRecorded(t, target.Gauge(poolActive, 2, tags))
	requireRecorded(t, target.Distribution(sendSeconds, 0.5, tags))
}

// retirePeer retires every peer-labelled metric for one address, the way a peer
// teardown hook does.
func retirePeer(t *testing.T, target *processor, peer string) {
	t.Helper()

	for _, name := range peerSeriesNames {
		if err := target.DeleteSeriesMatching(name, []string{peerLabel + ":" + peer}); err != nil {
			t.Fatalf("DeleteSeriesMatching(%q) error: %v", name, err)
		}
	}
}

func requireRecorded(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("record error: %v", err)
	}
}

func requireMemoSizes(t *testing.T, target *processor, stage string, counters, gauges, histograms int) {
	t.Helper()

	cases := []struct {
		kind string
		got  int
		want int
	}{
		{kind: cacheKindCounter, got: target.counterHandles.Stats().Size, want: counters},
		{kind: cacheKindGauge, got: target.gaugeHandles.Stats().Size, want: gauges},
		{kind: cacheKindHistogram, got: target.histogramHandles.Stats().Size, want: histograms},
	}

	for _, testCase := range cases {
		if testCase.got != testCase.want {
			t.Fatalf("%s memo size %s = %d, want %d", testCase.kind, stage, testCase.got, testCase.want)
		}
	}
}

// exportedSeries returns the series of one metric as a scrape would see them. A
// metric whose every child was deleted exports no family at all, so an absent family
// is zero series rather than a failure — which is exactly what retirement is for.
func exportedSeries(t *testing.T, target *processor, name string) []*dto.Metric {
	t.Helper()

	for _, family := range gather(t, target) {
		if family.GetName() == name {
			return family.GetMetric()
		}
	}

	return nil
}

func seriesWithLabel(t *testing.T, target *processor, name, label, value string) int {
	t.Helper()

	matched := 0

	for _, series := range exportedSeries(t, target, name) {
		for _, pair := range series.GetLabel() {
			if pair.GetName() == label && pair.GetValue() == value {
				matched++

				break
			}
		}
	}

	return matched
}
