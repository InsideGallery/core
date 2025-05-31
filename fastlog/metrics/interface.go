package metrics

import "context"

type Metric interface {
	Get(chartName string) Chart
}

type Chart interface {
	Execute(ctx context.Context, value int64, subject string, keyValues ...string) error
}
