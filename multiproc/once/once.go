// Package once provides retryable one-time execution helpers.
//
// New code should import this package instead of the legacy multiproc/sync path:
//
//	import "github.com/InsideGallery/core/multiproc/once"
//
// Compatibility: github.com/InsideGallery/core/multiproc/sync remains available
// for existing consumers. Prefer Once from this package so call sites avoid a
// local name collision with the standard-library sync package.
package once

import legacy "github.com/InsideGallery/core/multiproc/sync"

// Once performs one action successfully once and can be reset.
type Once = legacy.Once
