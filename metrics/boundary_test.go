package metrics //nolint:revive // package name matches directory/domain usage

import (
	"context"
	"testing"
)

func TestRecorderContract(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		client    *Client
		operation func(context.Context, *Client, Metric) (RecordResult, error)
		metric    Metric
		wantKind  string
	}{
		{
			name: "count metric",
			client: &Client{
				processors: []Processor{&spyProcessor{}},
			},
			operation: func(ctx context.Context, client *Client, metric Metric) (RecordResult, error) {
				return client.CountMetric(ctx, metric)
			},
			metric:   Metric{Name: "events_total", Int: 1},
			wantKind: "count",
		},
		{
			name: "gauge metric",
			client: &Client{
				processors: []Processor{&spyProcessor{}},
			},
			operation: func(ctx context.Context, client *Client, metric Metric) (RecordResult, error) {
				return client.GaugeMetric(ctx, metric)
			},
			metric:   Metric{Name: "queue_depth", Float: 2},
			wantKind: "gauge",
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var _ Recorder = (*Client)(nil)

			got, err := test.operation(context.Background(), test.client, test.metric)
			if err != nil {
				t.Fatalf("operation() error: %v", err)
			}

			if got.Kind != test.wantKind {
				t.Fatalf("Kind = %q, want %q", got.Kind, test.wantKind)
			}

			if got.ProcessorCount != len(test.client.processors) {
				t.Fatalf("ProcessorCount = %d, want %d", got.ProcessorCount, len(test.client.processors))
			}
		})
	}
}
