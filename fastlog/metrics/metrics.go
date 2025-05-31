package metrics

import (
	"context"
	"log/slog"
	"strings"
	"sync"

	"go.opentelemetry.io/contrib/instrumentation/host"
	rt "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"github.com/InsideGallery/core/server/instance"
)

const (
	ServiceNameKey       = attribute.Key("service")
	ServiceInstanceIDKey = attribute.Key("service_instance_id")
)

type OTLPMetric struct {
	ctx           context.Context
	metric        metric.Meter
	charts        map[string]Chart
	mu            *sync.Mutex
	meterProvider *sdkmetric.MeterProvider
	prefix        string
}

func Default(ctx context.Context) (*OTLPMetric, error) {
	cfg, err := GetConfigFromEnv()
	if err != nil {
		return nil, err
	}

	m, err := otlpmetricgrpc.New(ctx)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			// The service name used to display traces in backends
			ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceVersionKey.String(cfg.Version),
			ServiceInstanceIDKey.String(instance.GetShortInstanceID()),
		),
		resource.WithFromEnv(),
		resource.WithHost(),
	)
	if err != nil {
		return nil, err
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(m)),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)

	// runtime exported
	err = rt.Start()
	if err != nil {
		return nil, err
	}

	// host metrics exporter
	err = host.Start()
	if err != nil {
		return nil, err
	}

	return NewOTLPMetric(
		ctx,
		meterProvider,
		otel.Meter(cfg.Namespace, cfg.GetOptions()...),
		cfg.Prefix,
	), nil
}

func NewOTLPMetric(
	ctx context.Context,
	meterProvider *sdkmetric.MeterProvider,
	meter metric.Meter,
	prefix string,
) *OTLPMetric {
	return &OTLPMetric{
		ctx:           ctx,
		meterProvider: meterProvider,
		charts:        map[string]Chart{},
		prefix:        prefix,
		metric:        meter,
		mu:            &sync.Mutex{},
	}
}

func (o *OTLPMetric) GetMetric() metric.Meter {
	return o.metric
}

func (o *OTLPMetric) Shutdown() {
	err := o.meterProvider.Shutdown(o.ctx)
	if err != nil {
		slog.Default().Warn("Error shutdown execution", "err", err)
	}
}

func (o *OTLPMetric) Histogram(chartName string) (Chart, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	v, ok := o.charts[chartName]
	if ok {
		return v, nil
	}

	name := chartName
	if o.prefix != "" {
		name = strings.Join([]string{o.prefix, "_", chartName}, "")
	}

	scoringHistogram, err := o.metric.Int64Histogram(name)
	if err != nil {
		return nil, err
	}

	o.charts[chartName] = NewHistogramChart(scoringHistogram)

	return o.charts[chartName], nil
}
