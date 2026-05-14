package otel

import (
	"testing"

	"github.com/InsideGallery/core/metrics"
)

func TestProcessorRecordsMetricsAndNormalizesAttributes(t *testing.T) {
	rawProcessor, err := New(metrics.Config{}, "test-svc")
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	processor, ok := rawProcessor.(*processor)
	if !ok {
		t.Fatalf("processor type = %T", rawProcessor)
	}

	if err := rawProcessor.Count("requests total", 2, []string{"status code:200"}); err != nil {
		t.Fatalf("Count() error: %v", err)
	}

	if err := rawProcessor.Gauge("active connections", 3, []string{"status code:200"}); err != nil {
		t.Fatalf("Gauge() error: %v", err)
	}

	if err := rawProcessor.Distribution("wait seconds", 1.5, []string{"status code:200"}); err != nil {
		t.Fatalf("Distribution() error: %v", err)
	}

	attrs := processor.attributes([]string{"status code:200", "loose"})
	if len(attrs) != 3 {
		t.Fatalf("attributes = %v, want service plus 2 tags", attrs)
	}

	if string(attrs[0].Key) != "service" || attrs[0].Value.AsString() != "test-svc" {
		t.Fatalf("service attribute = %v", attrs[0])
	}

	if string(attrs[1].Key) != "tag" || attrs[1].Value.AsString() != "loose" {
		t.Fatalf("loose tag attribute = %v", attrs[1])
	}

	if string(attrs[2].Key) != "status_code" || attrs[2].Value.AsString() != "200" {
		t.Fatalf("status attribute = %v", attrs[2])
	}
}

func TestSanitizeNameFallback(t *testing.T) {
	if got := sanitizeName(""); got != "metric" {
		t.Fatalf("sanitizeName(\"\") = %q, want metric", got)
	}
}
