# fastlog/all

Import path: `github.com/InsideGallery/core/fastlog/all`

`fastlog/all` is a bundle import package. It imports every in-tree fastlog handler so their `init` functions register
with `github.com/InsideGallery/core/fastlog/handlers`.

## Main APIs

This package has no exported API. Use it for side effects:

```go
package main

import _ "github.com/InsideGallery/core/fastlog/all"
```

## Registered Handlers

In the default build, the bundle registers:

- `datadog`
- `nop`
- `otel`
- `stderr`

## Build Tags

The default file is built with `!fastlog_minimal`. Building with `-tags fastlog_minimal` keeps the import path available
but omits all handler imports, so no handlers are registered by this bundle.

Use this package when a binary should select any in-tree log output from `LOG_OUTPUTS`.
