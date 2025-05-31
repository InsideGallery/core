package metrics

import (
	"context"
	"log/slog"
)

type Wrapper struct {
	chart Chart
}

func NewWrapper(met *OTLPMetric, histogram string) Wrapper {
	chart, err := met.Histogram(histogram)
	if err != nil {
		slog.Default().Error("error getting histogram", "err", err)

		return Wrapper{}
	}

	return Wrapper{chart: chart}
}

func (w Wrapper) Execute(ctx context.Context, value int64, subject string, keyValues ...string) {
	if w.chart == nil {
		return
	}

	err := w.chart.Execute(ctx, value, subject, keyValues...)
	if err != nil {
		slog.Default().Error("error executing histogram", "err", err)

		return
	}
}
