package prometheus

import (
	"net/http"
	"net/http/httptest"
	"runtime"
	"sort"
	"strings"
	"testing"

	dto "github.com/prometheus/client_model/go"

	"github.com/InsideGallery/core/metrics"
)

const protobufMetricFamilyAccept = "application/vnd.google.protobuf; " +
	"proto=io.prometheus.client.MetricFamily; encoding=delimited"

func TestHTTPHandlerWithoutActiveProcessor(t *testing.T) {
	resetActiveProcessor(t)

	w := httptest.NewRecorder()
	HTTPHandler(w, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	if got := w.Header().Get("Content-Type"); got != contentType {
		t.Fatalf("Content-Type = %q, want %q", got, contentType)
	}

	if body := w.Body.String(); body != "" {
		t.Fatalf("body = %q, want empty", body)
	}
}

func TestHTTPHandlerRendersActiveProcessor(t *testing.T) {
	resetActiveProcessor(t)

	rawProcessor, err := New(metrics.Config{}, "test-svc")
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	t.Cleanup(func() {
		if err := rawProcessor.Close(); err != nil {
			t.Fatalf("Close() error: %v", err)
		}
	})

	if err := rawProcessor.Count("sample.requests", 2, []string{"kind:test", "status:ok"}); err != nil {
		t.Fatalf("Count() error: %v", err)
	}

	w := httptest.NewRecorder()
	HTTPHandler(w, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	body := w.Body.String()
	for _, want := range []string{
		"# TYPE sample_requests counter",
		`sample_requests{kind="test",service="test-svc",status="ok"} 2`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q in:\n%s", want, body)
		}
	}
}

func TestCountRejectsNegativeValue(t *testing.T) {
	resetActiveProcessor(t)

	rawProcessor, err := New(metrics.Config{}, "test-svc")
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	t.Cleanup(func() {
		if err := rawProcessor.Close(); err != nil {
			t.Fatalf("Close() error: %v", err)
		}
	})

	if err := rawProcessor.Count("sample.requests", -1, nil); err == nil {
		t.Fatal("expected negative counter error")
	}
}

func TestCloseClearsOnlyActiveProcessor(t *testing.T) {
	resetActiveProcessor(t)

	firstRaw, err := New(metrics.Config{}, "first-svc")
	if err != nil {
		t.Fatalf("New(first) error: %v", err)
	}

	secondRaw, err := New(metrics.Config{}, "second-svc")
	if err != nil {
		t.Fatalf("New(second) error: %v", err)
	}

	first, ok := firstRaw.(*processor)
	if !ok {
		t.Fatalf("first processor type = %T", firstRaw)
	}

	second, ok := secondRaw.(*processor)
	if !ok {
		t.Fatalf("second processor type = %T", secondRaw)
	}

	if currentActiveProcessor() != second {
		t.Fatal("expected second processor to be active")
	}

	if err := first.Close(); err != nil {
		t.Fatalf("Close(first) error: %v", err)
	}

	if currentActiveProcessor() != second {
		t.Fatal("closing inactive processor should not clear active processor")
	}

	if err := second.Close(); err != nil {
		t.Fatalf("Close(second) error: %v", err)
	}

	if currentActiveProcessor() != nil {
		t.Fatal("expected active processor to be cleared")
	}
}

func TestLabelsFromTagsNormalizeAndSanitize(t *testing.T) {
	labels := labelsFromTags([]string{
		"status code:200",
		"method:GET",
		"method:POST",
		"1bad:value",
		"ignored",
	})

	wantNames := []string{"_1bad", "method", "status_code"}
	wantValues := []string{"value", "GET", "200"}

	if len(labels.names) != len(wantNames) {
		t.Fatalf("label names = %v, want %v", labels.names, wantNames)
	}

	for i := range wantNames {
		if labels.names[i] != wantNames[i] {
			t.Fatalf("label names = %v, want %v", labels.names, wantNames)
		}

		if labels.values[i] != wantValues[i] {
			t.Fatalf("label values = %v, want %v", labels.values, wantValues)
		}
	}
}

func TestNewRegistersRuntimeAndProcessCollectors(t *testing.T) {
	resetActiveProcessor(t)

	rawProcessor, err := New(metrics.Config{}, "test-svc")
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	t.Cleanup(func() {
		if err := rawProcessor.Close(); err != nil {
			t.Fatalf("Close() error: %v", err)
		}
	})

	processor, ok := rawProcessor.(*processor)
	if !ok {
		t.Fatalf("processor type = %T", rawProcessor)
	}

	families, err := processor.registry.Gather()
	if err != nil {
		t.Fatalf("Gather() error: %v", err)
	}

	requireMetricFamilies(t, families,
		"go_gc_duration_seconds",
		"go_goroutines",
		"go_memstats_heap_alloc_bytes",
		"go_memstats_stack_inuse_bytes",
		"go_threads",
	)

	if processFamilies := supportedProcessMetricFamilies(); len(processFamilies) > 0 {
		requireMetricFamilies(t, families, processFamilies...)
	}
}

func TestHTTPHandlerScrapesRuntimeAndProcessMetrics(t *testing.T) {
	resetActiveProcessor(t)

	rawProcessor, err := New(metrics.Config{}, "test-svc")
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	t.Cleanup(func() {
		if err := rawProcessor.Close(); err != nil {
			t.Fatalf("Close() error: %v", err)
		}
	})

	w := httptest.NewRecorder()
	HTTPHandler(w, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	body := w.Body.String()
	for _, want := range []string{
		"# TYPE go_goroutines gauge",
		`go_goroutines{service="test-svc"}`,
		"# TYPE go_memstats_heap_alloc_bytes gauge",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q in:\n%s", want, body)
		}
	}

	if len(supportedProcessMetricFamilies()) == 0 {
		return
	}

	for _, want := range []string{
		"# TYPE process_cpu_seconds_total counter",
		`process_cpu_seconds_total{service="test-svc"}`,
		"# TYPE process_resident_memory_bytes gauge",
		"# TYPE process_start_time_seconds gauge",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q in:\n%s", want, body)
		}
	}
}

func TestHTTPHandlerRendersDistributionAsHistogram(t *testing.T) {
	resetActiveProcessor(t)

	rawProcessor, err := New(metrics.Config{}, "test-svc")
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	t.Cleanup(func() {
		if err := rawProcessor.Close(); err != nil {
			t.Fatalf("Close() error: %v", err)
		}
	})

	processor, ok := rawProcessor.(*processor)
	if !ok {
		t.Fatalf("processor type = %T", rawProcessor)
	}

	for _, value := range []float64{5, 20, 120} {
		if err := rawProcessor.Distribution("sample.duration", value, []string{
			"kind:test",
			"operation:lookup",
		}); err != nil {
			t.Fatalf("Distribution() error: %v", err)
		}
	}

	families, err := processor.registry.Gather()
	if err != nil {
		t.Fatalf("Gather() error: %v", err)
	}

	histogram := requireHistogram(t, families, "sample_duration")
	if histogram.GetSampleCount() != 3 {
		t.Fatalf("SampleCount = %d, want 3", histogram.GetSampleCount())
	}

	if histogram.GetSampleSum() != 145 {
		t.Fatalf("SampleSum = %v, want 145", histogram.GetSampleSum())
	}

	if histogram.Schema == nil {
		t.Fatal("expected native histogram schema")
	}

	if len(histogram.GetBucket()) != 0 {
		t.Fatalf("classic buckets = %d, want 0", len(histogram.GetBucket()))
	}

	if len(histogram.GetPositiveSpan()) == 0 && histogram.GetZeroCount() == 0 {
		t.Fatal("expected native positive or zero buckets")
	}
}

func TestHTTPHandlerNegotiatesProtobufForNativeHistograms(t *testing.T) {
	resetActiveProcessor(t)

	rawProcessor, err := New(metrics.Config{}, "test-svc")
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	t.Cleanup(func() {
		if err := rawProcessor.Close(); err != nil {
			t.Fatalf("Close() error: %v", err)
		}
	})

	if err := rawProcessor.Distribution("sample.duration", 25, []string{"kind:test"}); err != nil {
		t.Fatalf("Distribution() error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Accept", protobufMetricFamilyAccept)

	w := httptest.NewRecorder()
	HTTPHandler(w, req)

	if got := w.Header().Get("Content-Type"); !strings.Contains(got, "application/vnd.google.protobuf") {
		t.Fatalf("Content-Type = %q, want protobuf", got)
	}
}

func TestHTTPHandlerSupportsOpenMetricsScrapeContract(t *testing.T) {
	resetActiveProcessor(t)

	rawProcessor, err := New(metrics.Config{}, "test-svc")
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	t.Cleanup(func() {
		if err := rawProcessor.Close(); err != nil {
			t.Fatalf("Close() error: %v", err)
		}
	})

	if err := rawProcessor.Count("contract.requests", 1, []string{"outcome:ok"}); err != nil {
		t.Fatalf("Count() error: %v", err)
	}

	if err := rawProcessor.Distribution("contract.duration", 15, []string{"operation:validate"}); err != nil {
		t.Fatalf("Distribution() error: %v", err)
	}

	processor, ok := rawProcessor.(*processor)
	if !ok {
		t.Fatalf("processor type = %T", rawProcessor)
	}

	families, err := processor.registry.Gather()
	if err != nil {
		t.Fatalf("Gather() error: %v", err)
	}

	requireMetricWithLabels(t, families, "contract_requests", map[string]string{
		"outcome": "ok",
		"service": "test-svc",
	})
	requireMetricWithLabels(t, families, "contract_duration", map[string]string{
		"operation": "validate",
		"service":   "test-svc",
	})

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Accept", "application/openmetrics-text")

	w := httptest.NewRecorder()
	HTTPHandler(w, req)

	if got := w.Header().Get("Content-Type"); !strings.Contains(got, "application/openmetrics-text") {
		t.Fatalf("Content-Type = %q, want OpenMetrics text", got)
	}

	body := w.Body.String()
	for _, want := range []string{
		`contract_requests`,
		`contract_duration`,
		`service="test-svc"`,
		"# EOF",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("OpenMetrics body missing %q in:\n%s", want, body)
		}
	}
}

func resetActiveProcessor(t *testing.T) {
	t.Helper()

	activeMu.Lock()
	activeProcessor = nil
	activeMu.Unlock()
}

func supportedProcessMetricFamilies() []string {
	switch runtime.GOOS {
	case "linux":
		return []string{
			"process_cpu_seconds_total",
			"process_max_fds",
			"process_open_fds",
			"process_resident_memory_bytes",
			"process_start_time_seconds",
		}
	case "windows":
		return []string{
			"process_cpu_seconds_total",
			"process_resident_memory_bytes",
			"process_start_time_seconds",
		}
	default:
		return nil
	}
}

func requireMetricFamilies(t *testing.T, families []*dto.MetricFamily, names ...string) {
	t.Helper()

	available := make(map[string]struct{}, len(families))
	for _, family := range families {
		available[family.GetName()] = struct{}{}
	}

	for _, name := range names {
		if _, ok := available[name]; ok {
			continue
		}

		t.Fatalf("missing metric family %q; available families: %v", name, familyNames(families))
	}
}

func requireMetricWithLabels(
	t *testing.T,
	families []*dto.MetricFamily,
	name string,
	wantLabels map[string]string,
) {
	t.Helper()

	for _, family := range families {
		if family.GetName() != name {
			continue
		}

		for _, metric := range family.GetMetric() {
			if metricHasLabels(metric, wantLabels) {
				return
			}
		}
	}

	t.Fatalf("missing metric %q with labels %v", name, wantLabels)
}

func metricHasLabels(metric *dto.Metric, wantLabels map[string]string) bool {
	gotLabels := make(map[string]string, len(metric.GetLabel()))
	for _, label := range metric.GetLabel() {
		gotLabels[label.GetName()] = label.GetValue()
	}

	for name, want := range wantLabels {
		if gotLabels[name] != want {
			return false
		}
	}

	return true
}

func familyNames(families []*dto.MetricFamily) []string {
	names := make([]string, 0, len(families))
	for _, family := range families {
		names = append(names, family.GetName())
	}

	sort.Strings(names)

	return names
}

func requireHistogram(t *testing.T, families []*dto.MetricFamily, name string) *dto.Histogram {
	t.Helper()

	for _, family := range families {
		if family.GetName() != name {
			continue
		}

		metrics := family.GetMetric()
		if len(metrics) != 1 {
			t.Fatalf("metric count = %d, want 1", len(metrics))
		}

		histogram := metrics[0].GetHistogram()
		if histogram == nil {
			t.Fatalf("metric %q is not a histogram", name)
		}

		return histogram
	}

	t.Fatalf("missing metric family %q", name)

	return nil
}
