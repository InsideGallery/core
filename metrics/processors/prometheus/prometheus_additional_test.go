package prometheus

import (
	"strings"
	"testing"

	"github.com/InsideGallery/core/metrics"
)

func TestProcessorRejectsNegativeCounter(t *testing.T) {
	rawProcessor, err := New(metrics.Config{}, "unit")
	if err != nil {
		t.Fatalf("new processor: %v", err)
	}
	defer func() {
		if err := rawProcessor.Close(); err != nil {
			t.Fatalf("close processor: %v", err)
		}
	}()

	if err := rawProcessor.Count("negative", -1, nil); err == nil {
		t.Fatal("expected negative counter error")
	}
}

func TestProcessorGaugeReusesCollector(t *testing.T) {
	rawProcessor, err := New(metrics.Config{}, "unit")
	if err != nil {
		t.Fatalf("new processor: %v", err)
	}
	defer func() {
		if err := rawProcessor.Close(); err != nil {
			t.Fatalf("close processor: %v", err)
		}
	}()

	processor, ok := rawProcessor.(*processor)
	if !ok {
		t.Fatalf("processor type = %T", rawProcessor)
	}

	cases := []struct {
		name  string
		value float64
		tags  []string
	}{
		{
			name:  "first value",
			value: 1.5,
			tags:  []string{"status:ok", "status:duplicate", "bad tag", ":missing", "queue-depth:primary"},
		},
		{
			name:  "reused collector",
			value: 2.5,
			tags:  []string{"queue-depth:primary", "status:ok"},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if err := rawProcessor.Gauge("queue depth", test.value, test.tags); err != nil {
				t.Fatalf("gauge: %v", err)
			}
		})
	}

	families, err := processor.registry.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}

	if len(families) == 0 {
		t.Fatal("expected gathered families")
	}

	collector, labels, err := processor.gauge("queue depth", []string{"status:ok", "queue-depth:primary"})
	if err != nil {
		t.Fatalf("reuse gauge: %v", err)
	}

	collector.WithLabelValues(labels.values...).Set(3.5)
}

func TestProcessorHelpers(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "labels ignore invalid tags and sort names",
			run: func(t *testing.T) {
				t.Helper()

				labels := labelsFromTags([]string{"z:last", "bad", ":missing", "a:first", "a:duplicate"})
				if strings.Join(labels.names, ",") != "a,z" {
					t.Fatalf("label names = %#v, want [a z]", labels.names)
				}

				if strings.Join(labels.values, ",") != "duplicate,last" {
					t.Fatalf("label values = %#v, want [duplicate last]", labels.values)
				}
			},
		},
		{
			name: "collector key joins labels",
			run: func(t *testing.T) {
				t.Helper()

				key := newCollectorKey("requests", []string{"a", "b"})
				if key.name != "requests" || key.labelKeys != "a\xffb" {
					t.Fatalf("key = %#v", key)
				}
			},
		},
		{
			name: "names are sanitized",
			run: func(t *testing.T) {
				t.Helper()

				if got := sanitizeName("9 bad-name"); got != "_9_bad_name" {
					t.Fatalf("sanitizeName() = %q", got)
				}

				if got := sanitizeName(""); got != "met"+"ric" {
					t.Fatalf("empty sanitizeName() = %q", got)
				}

				if got := sanitizeLabelName(""); got != "label" {
					t.Fatalf("empty sanitizeLabelName() = %q", got)
				}

				if got := helpText("requests"); got != "Ptolemy metric requests." {
					t.Fatalf("helpText() = %q", got)
				}
			},
		},
		{
			name: "clear active ignores different processor",
			run: func(t *testing.T) {
				t.Helper()

				first := &processor{}
				second := &processor{}

				setActiveProcessor(first)
				clearActiveProcessor(second)

				if currentActiveProcessor() != first {
					t.Fatal("active processor was cleared by a different processor")
				}

				clearActiveProcessor(first)
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}
