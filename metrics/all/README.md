# metrics/all

Import path: `github.com/InsideGallery/core/metrics/all`

`metrics/all` is a bundle import package. It imports every in-tree metrics processor so their `init` functions register
with `github.com/InsideGallery/core/metrics`.

## Main APIs

This package has no exported API. Use it for side effects:

```go
package main

import _ "github.com/InsideGallery/core/metrics/all"
```

## Registered Processors

In the default build, the bundle registers:

- `datadog`
- `otel`
- `prometheus`
- `statsd`

## Build Tags

The default file is built with `!metrics_minimal`. Building with `-tags metrics_minimal` keeps the import path available
but omits all processor imports, so no processors are registered by this bundle.

Use this package when a binary should select any in-tree metrics processor from `METRICS_PROCESSORS`.
