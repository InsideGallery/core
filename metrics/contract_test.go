package metrics //nolint:revive // package name matches directory/domain usage

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
)

func TestPackageStaysTransportAgnostic(t *testing.T) {
	disallowed := []string{
		"http.request.",
		"nats.publisher.",
		"queue.subscriptions.",
		"github.com/gofiber/fiber",
		"github.com/FrogoAI/mq-balancer",
	}

	err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		contents := string(data)
		for _, snippet := range disallowed {
			if strings.Contains(contents, snippet) {
				t.Fatalf("pkg/metrics must stay transport-agnostic: %s contains %q", path, snippet)
			}
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walk pkg/metrics: %v", err)
	}
}

// Optional handle API contract (MET-11). HandleProvider is additive, so the
// property under test is what happens to a backend that does NOT implement it:
// resolving a handle off *Client must keep recording through the Processor
// methods, and a mixed set of processors must all see the same record.

// sample is one recorded metric as a processor saw it.
type sample struct {
	kind  string
	name  string
	tags  []string
	value float64
}

// recordingProcessor is a Processor with no handle support: it captures whatever
// reaches Count, Gauge and Distribution.
type recordingProcessor struct {
	mu      sync.Mutex
	samples []sample
	err     error
}

func (p *recordingProcessor) Close() error { return nil }

func (p *recordingProcessor) Count(name string, value int64, tags []string) error {
	return p.record("count", name, float64(value), tags)
}

func (p *recordingProcessor) Gauge(name string, value float64, tags []string) error {
	return p.record("gauge", name, value, tags)
}

func (p *recordingProcessor) Distribution(name string, value float64, tags []string) error {
	return p.record("distribution", name, value, tags)
}

func (p *recordingProcessor) record(kind, name string, value float64, tags []string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.samples = append(p.samples, sample{
		kind:  kind,
		name:  name,
		value: value,
		tags:  slices.Clone(tags),
	})

	return p.err
}

func (p *recordingProcessor) recorded() []sample {
	p.mu.Lock()
	defer p.mu.Unlock()

	return slices.Clone(p.samples)
}

func (p *recordingProcessor) requireSample(t *testing.T, want sample) {
	t.Helper()

	for _, got := range p.recorded() {
		if got.kind == want.kind && got.name == want.name &&
			got.value == want.value && slices.Equal(got.tags, want.tags) {
			return
		}
	}

	t.Fatalf("missing %+v in %+v", want, p.recorded())
}

// providerProcessor stands in for a backend that DOES resolve handles (as the
// Prometheus processor does). Its handles tag their samples "handle-*" so a test
// can tell handle-recorded samples from Processor-recorded ones.
type providerProcessor struct {
	*recordingProcessor
	err error
}

//nolint:ireturn // handle API returns the abstraction by design
func (p providerProcessor) CounterHandle(name string, tags []string) (Counter, error) {
	if p.err != nil {
		return nil, p.err
	}

	return providerHandle{processor: p.recordingProcessor, name: name, tags: tags}, nil
}

//nolint:ireturn // handle API returns the abstraction by design
func (p providerProcessor) GaugeHandle(name string, tags []string) (Gauge, error) {
	if p.err != nil {
		return nil, p.err
	}

	return providerHandle{processor: p.recordingProcessor, name: name, tags: tags}, nil
}

//nolint:ireturn // handle API returns the abstraction by design
func (p providerProcessor) DistributionHandle(name string, tags []string) (Observer, error) {
	if p.err != nil {
		return nil, p.err
	}

	return providerHandle{processor: p.recordingProcessor, name: name, tags: tags}, nil
}

type providerHandle struct {
	processor *recordingProcessor
	name      string
	tags      []string
}

func (h providerHandle) Add(value int64) {
	_ = h.processor.record("handle-count", h.name, float64(value), h.tags)
}

func (h providerHandle) Set(value float64) {
	_ = h.processor.record("handle-gauge", h.name, value, h.tags)
}

func (h providerHandle) Observe(value float64) {
	_ = h.processor.record("handle-distribution", h.name, value, h.tags)
}

func TestHandlesRecordThroughProcessorsWithoutHandleSupport(t *testing.T) {
	processor := &recordingProcessor{}
	client := &Client{processors: []Processor{processor}, service: "test-svc"}

	tags := []string{"op:get", "namespace:ns"}

	counter, err := client.CounterHandle("requests_total", tags)
	if err != nil {
		t.Fatalf("CounterHandle() error: %v", err)
	}

	gauge, err := client.GaugeHandle("queue_depth", tags)
	if err != nil {
		t.Fatalf("GaugeHandle() error: %v", err)
	}

	observer, err := client.DistributionHandle("duration_seconds", tags)
	if err != nil {
		t.Fatalf("DistributionHandle() error: %v", err)
	}

	counter.Add(2)
	gauge.Set(7)
	observer.Observe(0.5)

	processor.requireSample(t, sample{kind: "count", name: "requests_total", value: 2, tags: tags})
	processor.requireSample(t, sample{kind: "gauge", name: "queue_depth", value: 7, tags: tags})
	processor.requireSample(t, sample{kind: "distribution", name: "duration_seconds", value: 0.5, tags: tags})
}

func TestHandlesFanOutToProviderAndNonProviderProcessors(t *testing.T) {
	plain := &recordingProcessor{}
	provider := providerProcessor{recordingProcessor: &recordingProcessor{}}
	client := &Client{processors: []Processor{plain, provider}, service: "test-svc"}

	tags := []string{"op:put"}

	counter, err := client.CounterHandle("requests_total", tags)
	if err != nil {
		t.Fatalf("CounterHandle() error: %v", err)
	}

	counter.Add(3)

	// The provider resolved a handle; the plain processor kept the Count path.
	provider.requireSample(t, sample{kind: "handle-count", name: "requests_total", value: 3, tags: tags})
	plain.requireSample(t, sample{kind: "count", name: "requests_total", value: 3, tags: tags})
}

// A handle outlives the call that resolved it, so it must not alias a caller's
// label array — internal/metrics in fdb-server resolves from a stack array.
func TestHandlesCopyTheCallerTags(t *testing.T) {
	processor := &recordingProcessor{}
	client := &Client{processors: []Processor{processor}, service: "test-svc"}

	tags := [2]string{"op:get", "namespace:ns"}

	counter, err := client.CounterHandle("requests_total", tags[:])
	if err != nil {
		t.Fatalf("CounterHandle() error: %v", err)
	}

	tags[1] = "namespace:reused"
	counter.Add(1)

	processor.requireSample(t, sample{
		kind:  "count",
		name:  "requests_total",
		value: 1,
		tags:  []string{"op:get", "namespace:ns"},
	})
}

func TestNilClientHandlesDiscard(t *testing.T) {
	var client *Client

	counter, err := client.CounterHandle("requests_total", nil)
	if err != nil {
		t.Fatalf("CounterHandle() error: %v", err)
	}

	gauge, err := client.GaugeHandle("queue_depth", nil)
	if err != nil {
		t.Fatalf("GaugeHandle() error: %v", err)
	}

	observer, err := client.DistributionHandle("duration_seconds", nil)
	if err != nil {
		t.Fatalf("DistributionHandle() error: %v", err)
	}

	// A nil client records nothing; the handles must still be safe to call.
	counter.Add(1)
	gauge.Set(1)
	observer.Observe(1)
}

func TestHandleResolutionErrorIsReturned(t *testing.T) {
	failure := errors.New("resolve failed")
	provider := providerProcessor{recordingProcessor: &recordingProcessor{}, err: failure}
	client := &Client{processors: []Processor{provider}, service: "test-svc"}

	if _, err := client.CounterHandle("requests_total", nil); !errors.Is(err, failure) {
		t.Fatalf("CounterHandle() error = %v, want %v", err, failure)
	}

	if _, err := client.GaugeHandle("queue_depth", nil); !errors.Is(err, failure) {
		t.Fatalf("GaugeHandle() error = %v, want %v", err, failure)
	}

	if _, err := client.DistributionHandle("duration_seconds", nil); !errors.Is(err, failure) {
		t.Fatalf("DistributionHandle() error = %v, want %v", err, failure)
	}
}
