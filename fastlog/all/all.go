//go:build !fastlog_minimal

// Package all imports every in-tree fastlog handler so each handler registers
// with the default fastlog handler registry through its init hook. Blank import
// this package when a binary should select stdout, stderr, nop, file, Logstash,
// OpenTelemetry, or Datadog logging from configuration; build with the
// fastlog_minimal tag to omit the bundle imports.
package all

import (
	_ "github.com/InsideGallery/core/fastlog/handlers/datadog" // register datadog handler
	//nolint:staticcheck // bundle intentionally imports deprecated compatibility handler
	_ "github.com/InsideGallery/core/fastlog/handlers/logfile"  // register logfile handler
	_ "github.com/InsideGallery/core/fastlog/handlers/logstash" // register logstash handler
	_ "github.com/InsideGallery/core/fastlog/handlers/nop"      // register nop handler
	_ "github.com/InsideGallery/core/fastlog/handlers/otel"     // register otel handler
	_ "github.com/InsideGallery/core/fastlog/handlers/stderr"   // register stderr handler
	_ "github.com/InsideGallery/core/fastlog/handlers/stdout"   // register stdout handler
)
