//go:build !fastlog_minimal

// Package all imports every in-tree fastlog handler so each handler registers
// with the default fastlog handler registry through its init hook.
//
// Blank import this package when a binary should select any supported log
// output from configuration. Build with the fastlog_minimal tag to omit the
// bundle imports while keeping the import path available.
package all

import (
	_ "github.com/InsideGallery/core/fastlog/handlers/datadog" // register datadog handler
	_ "github.com/InsideGallery/core/fastlog/handlers/nop"     // register nop handler
	_ "github.com/InsideGallery/core/fastlog/handlers/otel"    // register otel handler
	_ "github.com/InsideGallery/core/fastlog/handlers/stderr"  // register stderr handler
)
