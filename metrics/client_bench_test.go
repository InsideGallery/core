package metrics //nolint:revive // package name matches directory/domain usage

import "testing"

var (
	benchmarkFanoutValue int64
	benchmarkTagSet      string
	benchmarkTags        []string
)

func BenchmarkNormalizeTags(b *testing.B) {
	tags := []string{
		"status:200",
		"method:GET",
		"route:/v2/notifyapi/notifications",
		"processor:prometheus",
	}

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		benchmarkTags = NormalizeTags(tags)
	}
}

func BenchmarkTagSet(b *testing.B) {
	tags := []string{
		"status:200",
		"method:GET",
		"route:/v2/notifyapi/notifications",
		"processor:prometheus",
	}

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		benchmarkTagSet = TagSet(tags)
	}
}

func BenchmarkClientFanout(b *testing.B) {
	tags := []string{
		"status:200",
		"method:GET",
		"route:/v2/notifyapi/notifications",
	}
	cases := []struct {
		name   string
		record func(*Client) error
	}{
		{
			name: "count_four_processors",
			record: func(c *Client) error {
				return c.Count("ptolemy_requests_total", 1, tags)
			},
		},
		{
			name: "gauge_four_processors",
			record: func(c *Client) error {
				return c.Gauge("ptolemy_active_sessions", 7, tags)
			},
		},
		{
			name: "distribution_four_processors",
			record: func(c *Client) error {
				return c.Distribution("ptolemy_request_duration_ms", 12.5, tags)
			},
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			first := &benchmarkProcessor{}
			c := &Client{
				processors: []Processor{
					first,
					&benchmarkProcessor{},
					&benchmarkProcessor{},
					&benchmarkProcessor{},
				},
				service: "bench-svc",
			}

			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				if err := tc.record(c); err != nil {
					b.Fatal(err)
				}
			}

			benchmarkFanoutValue = first.value
		})
	}
}

type benchmarkProcessor struct {
	value int64
}

func (p *benchmarkProcessor) Close() error {
	return nil
}

func (p *benchmarkProcessor) Count(name string, value int64, tags []string) error {
	p.value += value + int64(len(name)) + int64(len(tags))

	return nil
}

func (p *benchmarkProcessor) Gauge(name string, value float64, tags []string) error {
	p.value += int64(value) + int64(len(name)) + int64(len(tags))

	return nil
}

func (p *benchmarkProcessor) Distribution(name string, value float64, tags []string) error {
	p.value += int64(value) + int64(len(name)) + int64(len(tags))

	return nil
}
