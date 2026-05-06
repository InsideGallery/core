package prometheus

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	dto "github.com/prometheus/client_model/go"

	"github.com/InsideGallery/core/metrics"
)

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

	if err := rawProcessor.Count("http.requests", 2, []string{"status:200", "method:GET"}); err != nil {
		t.Fatalf("Count() error: %v", err)
	}

	w := httptest.NewRecorder()
	HTTPHandler(w, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	body := w.Body.String()
	for _, want := range []string{
		"# TYPE http_requests counter",
		`http_requests{method="GET",service="test-svc",status="200"} 2`,
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
		if err := rawProcessor.Distribution("http.request.duration", value, []string{
			"route:/users/:id",
			"method:GET",
		}); err != nil {
			t.Fatalf("Distribution() error: %v", err)
		}
	}

	families, err := processor.registry.Gather()
	if err != nil {
		t.Fatalf("Gather() error: %v", err)
	}

	histogram := requireHistogram(t, families, "http_request_duration")
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

	if err := rawProcessor.Distribution("http.request.duration", 25, []string{"method:GET"}); err != nil {
		t.Fatalf("Distribution() error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Accept", "application/vnd.google.protobuf; proto=io.prometheus.client.MetricFamily; encoding=delimited")

	w := httptest.NewRecorder()
	HTTPHandler(w, req)

	if got := w.Header().Get("Content-Type"); !strings.Contains(got, "application/vnd.google.protobuf") {
		t.Fatalf("Content-Type = %q, want protobuf", got)
	}
}

func resetActiveProcessor(t *testing.T) {
	t.Helper()

	activeMu.Lock()
	activeProcessor = nil
	activeMu.Unlock()
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
