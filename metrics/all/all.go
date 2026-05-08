//go:build !metrics_minimal

// Package all imports every in-tree metrics processor so each processor
// registers with the default metrics registry through its init hook. Blank
// import this package when a binary should select Datadog, OpenTelemetry,
// Prometheus, or StatsD metrics from configuration; build with the
// metrics_minimal tag to omit the bundle imports.
package all

import (
	_ "github.com/InsideGallery/core/metrics/processors/datadog"    // register datadog processor
	_ "github.com/InsideGallery/core/metrics/processors/otel"       // register otel processor
	_ "github.com/InsideGallery/core/metrics/processors/prometheus" // register prometheus processor
	_ "github.com/InsideGallery/core/metrics/processors/statsd"     // register statsd processor
)
