# InsideGallery Core

`github.com/InsideGallery/core` is a reusable Go library for InsideGallery
projects. It provides shared utilities, service adapters, logging, metrics,
PKI, queue, and server-support packages that consumer applications import with
Go modules.

This repository is not an application. It has no `main()`, no `cmd/`, no owned
network port, and no deployment topology. Consumers own process startup,
configuration, routes, backing-service endpoints, secrets, shutdown, and
deployment.

## Install

```bash
go get github.com/InsideGallery/core
```

Import only the package paths you need:

```go
package example

import (
	"github.com/InsideGallery/core/stdx/maths"
)

func example() bool {
	pair := maths.CantorPair(7, 11)

	return pair > 0
}
```

The Go version follows `go.mod`.

Memory data structures formerly under `memory/*` now live in
`github.com/FrogoAI/memory`; set helpers live in `github.com/FrogoAI/set`.

## Quick Start

Application binaries should keep startup simple. `app.WebMain` and
`app.NATSMain` install logging and metrics through the in-tree bundle imports,
then call service-specific initialization callbacks:

```go
package main

import (
	"context"

	"github.com/gofiber/fiber/v3"

	"github.com/InsideGallery/core/app"
	"github.com/InsideGallery/core/server/webserver"
)

func run() error {
	cfg, err := webserver.GetEnvConfig()
	if err != nil {
		return err
	}

	app.WebMain("api", cfg, func(_ context.Context, router *fiber.App) error {
		router.Get("/ping", func(c fiber.Ctx) error {
			return c.SendString("ok")
		})

		return nil
	})

	return nil
}
```

Manual compositions that do not use `app` should blank-import
`github.com/InsideGallery/core/fastlog/all` and
`github.com/InsideGallery/core/metrics/all` before configuring logging or
metrics. Keep new backends behind the same bundle-import pattern instead of
hardcoding provider choices in application code. This preserves Twelve-Factor
[Factor IV](docs/source/Twelve-Factor%20App.md#iv-backing-services) by keeping
backing services attached through config, and
[Factor XI](docs/source/Twelve-Factor%20App.md#xi-logs) by treating logs as
stdout/stderr event streams.

## Package Catalog

### Domain and Utility Packages

| Package path | Use |
|--------------|-----|
| `antibot` | Proof-of-work helpers for anti-bot checks. |
| `dataconv` | Binary encoding/decoding, IP conversion, and merge helpers. |
| `ecs` | Entity-component-system primitives. |
| `errors` | Error construction and combination utilities. |
| `oslistener` | OS signal listener helpers. |
| `pki/cryptor` | Shared cipher contracts. |
| `pki` | Legacy compatibility path for `pki/cryptor`. |
| `pki/aesgcm`, `pki/rsaoaep`, `pki/aescmac`, `pki/diversify`, `pki/saes` | Crypto helpers. |
| `pki/aes`, `pki/rsa` | Legacy compatibility paths for AES-GCM and RSA-OAEP helpers. |
| `stdx/bytes` | Byte, bit, CRC, and XOR helpers. |
| `stdx/maths` | Cantor pairing, random, probability, and math helpers. |
| `stdx/slices` | Slice batching and shingling helpers. |
| `stdx/strings` | String normalization, hashing, password, and context-key helpers. |
| `ticker` | Periodic task and delayed execution helpers. |

### Adapter and External-Service Packages

These packages are optional. They are used only when a consumer imports them and
supplies the backing service configuration.

| Package path | External dependency / role |
|--------------|----------------------------|
| `db/aerospike` | Aerospike client helpers, entity helpers, geospatial, and HLL support. |
| `db/bunt` | BuntDB connection helper. |
| `db/elasticsearch` | Elasticsearch client helper. |
| `db/frogodb` | FrogoDB smart-client connection and record helpers. |
| `db/gremlin` | Gremlin client, cache, and graph operation helpers. |
| `db/mongodb` | MongoDB client and filter helpers. |
| `db/neo4j` | Neo4j client configuration helpers. |
| `db/postgres` | Postgres connection helpers. |
| `db/redis` | Redis connection helpers. |
| `metrics` | Metrics client and backend-agnostic processor selection. |
| `metrics/processors/datadog` | Datadog metrics processor. |
| `metrics/processors/otel` | OpenTelemetry metrics processor. |
| `metrics/processors/prometheus` | Prometheus metrics processor. |
| `metrics/processors/statsd` | StatsD metrics processor. |
| `fastlog/all` | Bundle import for every in-tree log handler. |
| `fastlog/handlers/datadog` | Datadog `slog` handler support. |
| `fastlog/handlers/nop` | No-op handler support. |
| `fastlog/handlers/otel` | OpenTelemetry log handler support. |
| `fastlog/handlers/stderr` | Structured stderr event-stream handler support. |

### Server-Support Packages

These packages help applications build servers, but the application still owns
the `main()`, routes, ports, TLS, auth policy, and graceful shutdown.

| Package path | Use |
|--------------|-----|
| `app` | Application bootstrap helpers for web, metrics, and NATS composition. |
| `fastlog` | Structured `log/slog` configuration, handler fanout, and middleware. |
| `fastlog/middlewares` | Caller, error formatting, and GDPR log middleware. |
| `profiler` | Health checks, readiness/liveness probes, and pprof support. |
| `server/backoff` | HTTP transport retry/backoff helpers. |
| `server/honeypot` | Honeypot helpers. |
| `server/instance` | Runtime instance helpers. |
| `server/jwt` | JWT service, config, models, and Fiber middleware. |
| `server/sse` | Server-sent event listener and pool helpers. |
| `server/template` | Embedded HTML template parsing helpers. |
| `server/throughput` | Throughput tracking with memory storage. |
| `server/webserver` | Fiber app/server helpers, config, middleware, and request helpers. |

### Test, Resource, and Specialized Packages

| Package path | Use |
|--------------|-----|
| `fixtures` | Shared test fixtures. |
| `machielearning/nn` | Neural-network helpers. |

## Legacy Package Names

Some import paths predate the current package-naming rules. They remain
available for compatibility, but new code should prefer the focused
replacement paths below. Direct in-place renames are reserved for a future
major-version plan. Package docs at each affected path include the detailed
import examples and API-specific migration notes.

| Legacy path | Preferred path |
|-------------|----------------|
| `pki` | `pki/cryptor` |
| `pki/aes` | `pki/aesgcm` |
| `pki/rsa` | `pki/rsaoaep` |

## Logging Defaults

`fastlog` defaults to structured JSON logs on `stderr` through `LOG_OUTPUTS=stderr:json`.
The `stderr` output is registered when a binary uses `app.WebMain`, `app.NATSMain`,
or blank-imports `github.com/InsideGallery/core/fastlog/all`.

Supported in-tree output kinds are `stderr`, `datadog`, `otel`, and `nop`. The
base `fastlog` package imports only the `nop` fallback directly, so manual
logging setup should import either `fastlog/all` or the specific handler packages
it wants to select from configuration.

## Configuration

Consumer applications own configuration. Packages use two compatible paths:

1. Environment parsing helpers for production wiring.
2. Direct struct literals for tests and programmatic composition.

Common environment prefixes:

| Area | Prefix / helper |
|------|-----------------|
| Aerospike | `AEROSPIKE` or caller-supplied prefix through `db/aerospike.GetConnectionConfigFromEnv`. |
| BuntDB | Caller-supplied prefix, default helper usage uses `DB`. |
| FrogoDB | `FDB` or caller-supplied prefix through `db/frogodb.GetConnectionConfigFromEnv`. |
| Gremlin | `GREMLIN`. |
| MongoDB | `MONGO`. |
| Neo4j | `NEO4J`. |
| Postgres | `POSTGRES`. |
| Redis | `REDIS`. |
| NATS | `NATS`, consumed by `app.NATSMain` through `github.com/FrogoAI/mq-balancer/subscriber/driver/client`. |
| Fastlog | `LOG`; handler packages also own prefixes such as `DATADOG`. |
| Metrics | `METRICS`, plus processor prefixes `METRICS_DATADOG`, `METRICS_OTEL`, `METRICS_PROMETHEUS`, and `METRICS_STATSD`. |
| JWT | `JWT`. |
| Webserver | `APP` or caller-supplied prefix through `server/webserver.GetEnvConfig`. |

Do not hardcode credentials, hostnames, ports, or key material in library code.
Consumers should source those values from their own environment or secret store.

## External Services

Database, queue, metrics, logging, and server-support packages may wrap
third-party SDKs, but they do not create a repository-owned runtime. Consumers
choose which adapters to import, provide endpoints and credentials, pass
contexts where operations can block, and close or drain resources during
shutdown.

Pure utility packages should stay independent of optional infrastructure
packages. New public contracts should prefer core-owned types or small
interfaces over exposing vendor SDK types.

## Verification

Run the same repository gate locally that CI runs:

```bash
make ci
```

The gate expands to:

```bash
go test ./...
go test -race -count=1 ./...
golangci-lint run ./...
go test -coverprofile=coverage.out -cover ./...
go-test-coverage --config=./.testcoverage.yml
```

Pure unit tests are intentionally untagged, so `go test ./...` is the complete
default unit lane. Tests that require external services use the `integration`
build tag, and developer-local checks use the `local_test` build tag.

Run integration tests only when the required backing services are available:

```bash
go test -tags=integration ./...
```

GitHub Actions exposes the same command as a manual workflow lane through the
`workflow_dispatch` input `run-integration`.

Formatting and import-order issues can usually be fixed with:

```bash
golangci-lint run --fix ./...
```

Smoke benchmarks are available through `make bench`; the target writes
`benchmarks/current.txt` for local comparison. The Go version follows `go.mod`
locally and in GitHub Actions.

Coverage must not decrease in a merge request. The repository coverage gate is
defined in `.testcoverage.yml`; the current total threshold is `70%`, with
generated files, mocks, fixtures, and external-service-heavy packages excluded
there.

## Compatibility and Deprecation

Exported package paths, types, functions, methods, constants, and variables are
public compatibility contracts. Compatible releases should be additive.

When a legacy API needs replacement:

1. Add the replacement API first.
2. Keep the old API compiling as a compatibility shim.
3. Mark the old API with Go's `Deprecated:` doc-comment convention.
4. Document the detailed migration path in package docs; keep this README to
   broadly used compatibility policy and summary tables.
5. Remove or rename exported APIs only in a deliberately planned SemVer major
   release.

Existing awkward package names or vendor-shaped APIs may stay in place to
protect downstream consumers. Prefer additive wrappers for new code.

Current additive replacements:

- `pki/diversify.Key` is preferred for key diversification. `pki/diversify.DiversifyKey`
  remains as a deprecated compatibility shim.
- Core-owned boundary contracts now exist for infrastructure adapters:
  `db/aerospike.NamespaceStore`, `db/mongodb.DocumentStore`,
  `db/postgres.Database`, `db/redis.KeyValueStore`, `db/neo4j.Graph`,
  `db/elasticsearch.Searcher`, `db/gremlin.VertexStore`,
  `metrics.Recorder`, `server/webserver.Client`,
  `server/webserver.Runtime`, and JWT/PKI option/result helpers.
- Additional adapter boundaries cover helper surfaces that used to require
  direct SDK imports: `db/aerospike/entity.RecordStore`,
  `db/aerospike/hll.Counter`, `db/gremlin.GraphStore`,
  `db/mongodb.Filter`, and
  `server/webserver.RouteInitializer` / `server/webserver.RouteMiddleware`.
- Queue worker balancing is provided by `github.com/FrogoAI/mq-balancer`; core
  keeps only `app.NATSMain` bootstrap wiring to that external library.
- Legacy SDK-shaped clients remain available for existing consumers. New code
  should prefer the core-owned option/result types so vendor SDKs stay behind
  adapter packages.
- Runtime state now has explicit ownership APIs for new code: use
  `ecs.NewRegistry`, `oslistener.NewSignalListener`, `profiler.NewState`,
  `metrics.InstallDefault`, database client stores such as
  `db/aerospike.NewConnectionRegistry`, `db/mongodb.NewClientStore`,
  `db/postgres.NewClientStore`, and `db/redis.NewConnectionStore`,
  `fastlog.SetupDefaultLogger`, `server/webserver.NewRuntime`, and
  `server/template.NewTemplateWithDir`.
  Legacy package-level `Default`, `Set`, `Get`, and env-reading wrappers remain
  available as deprecated compatibility paths.
- App bootstrap keeps the simple main-style entry points: `app.WebMain` and
  `app.NATSMain` initialize logging, metrics, profiler health, shutdown
  listeners, routes or subscriptions, and runtime config.
