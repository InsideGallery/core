package prometheus

import (
	"testing"

	"github.com/InsideGallery/core/metrics"
)

func TestProcessorCounterAndHistogramReuseCollectors(t *testing.T) {
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
		name      string
		metric    string
		record    func(string) error
		cacheSize func() int
	}{
		{
			name:   "counter",
			metric: "requests total",
			record: func(name string) error {
				return rawProcessor.Count(name, 2, []string{"status:ok", "method:GET"})
			},
			cacheSize: func() int {
				return len(processor.counters)
			},
		},
		{
			name:   "histogram",
			metric: "request duration",
			record: func(name string) error {
				return rawProcessor.Distribution(name, 1.5, []string{"route:/v1/items"})
			},
			cacheSize: func() int {
				return len(processor.histograms)
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if err := test.record(test.metric); err != nil {
				t.Fatalf("record metric: %v", err)
			}

			if err := test.record(test.metric); err != nil {
				t.Fatalf("record cached metric: %v", err)
			}

			if got := test.cacheSize(); got != 1 {
				t.Fatalf("cache size = %d, want 1", got)
			}
		})
	}
}
