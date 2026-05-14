package prometheus

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/InsideGallery/core/metrics"
)

var benchmarkHTTPStatus int

func BenchmarkProcessorRecord(b *testing.B) {
	tags := []string{
		"status:200",
		"method:GET",
		"route:/v2/notifyapi/notifications",
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
