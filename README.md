# InsideGallery Core

`github.com/InsideGallery/core` is a reusable Go library for InsideGallery
projects. It provides shared utilities, in-memory data structures, service
adapters, logging, metrics, PKI, queue, and server-support packages that
consumer applications import with Go modules.

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
	"github.com/InsideGallery/core/mathx"
	"github.com/InsideGallery/core/memory/set"
)

func example() bool {
	ids := set.NewGenericDataSet[string]("a", "b")
	pair := mathx.CantorPair(7, 11)

	return ids.Contains("a") && pair > 0
}
```

The Go version follows `go.mod`.

## Quick Start

Application binaries should read config once, install the process logger, and
then let downstream packages emit through `slog.Default()`. Blank-import the
bundle packages in the binary so logging and metrics backends are selected by
configuration as attached resources:

```go
package main

import (
	"context"
	"errors"

	"github.com/InsideGallery/core/app"
	"github.com/InsideGallery/core/fastlog"

	_ "github.com/InsideGallery/core/fastlog/all"
	_ "github.com/InsideGallery/core/metrics/all"
)

func run(ctx context.Context, initRoutes app.InitRouter) (runErr error) {
	cfg, err := app.ConfigFromEnv()
	if err != nil {
		return err
	}

	closeDefault, err := fastlog.SetupDefault(ctx, cfg.Log)
	if err != nil {
		return err
	}
	defer func() {
		runErr = errors.Join(runErr, closeDefault())
	}()

	cfg.Log = nil

	return app.RunWeb(ctx, cfg, initRoutes)
}
```

Keep new backends behind the same bundle-import pattern instead of hardcoding
provider choices in application code. This preserves Twelve-Factor
[Factor IV](docs/source/Twelve-Factor%20App.md#iv-backing-services) by keeping
backing services attached through config, and
[Factor XI](docs/source/Twelve-Factor%20App.md#xi-logs) by treating logs as
stdout/stderr event streams.

## Package Catalog

### Domain and Utility Packages

| Package path | Use |
|--------------|-----|
| `antibot/pow` | Proof-of-work helpers for anti-bot checks. |
| `commands` | Command and event handler registration/execution helpers. |
| `dataconv` | Binary encoding/decoding, IP conversion, and merge helpers. |
| `ecs` | Entity-component-system primitives. |
| `errors` | Error construction and combination utilities. |
| `mathx` | Cantor pairing, random, probability, and math helpers. |
| `mathutils` | Legacy compatibility path for `mathx`. |
| `memory/bloom` | Bloom and counting Bloom filters. |
| `memory/btree` | Generic B-tree implementation. |
| `memory/comparator` | Comparators shared by ordered structures. |
| `memory/concurrent` | Concurrency-safe map and list containers. |
| `memory/fuzzysearch` | In-process fuzzy search and Levenshtein helpers. |
| `memory/hll` | HyperLogLog helpers. |
| `memory/linkedlist` | Generic linked list. |
| `memory/lru` | In-memory LRU cache. |
| `memory/orderedmap` | Ordered map implementation. |
| `memory/order` | Ordering helpers for memory data structures. |
| `memory/registry` | Generic grouped registry. |
| `memory/set` | Generic set and ordered set helpers. |
| `memory/sortedset` | Sorted set implementation. |
| `memory/stack` | Stack data structure. |
| `memory/utils` | Legacy compatibility path for safe map/list and sorting helpers. |
| `multiproc/buffer` | Buffered delayed execution helpers. |
| `multiproc/once` | Retryable once-style coordination helpers. |
| `multiproc/sync` | Legacy compatibility path for `multiproc/once`. |
| `multiproc/worker` | Worker, worker pool, and aggregator helpers. |
| `oslistener` | OS signal listener helpers. |
| `pki/cryptor` | Shared cipher contracts. |
| `pki` | Legacy compatibility path for `pki/cryptor`. |
| `pki/aesgcm`, `pki/rsaoaep`, `pki/aescmac`, `pki/diversify`, `pki/saes` | Crypto helpers. |
| `pki/aes`, `pki/rsa` | Legacy compatibility paths for AES-GCM and RSA-OAEP helpers. |
| `ticker` | Periodic task and delayed execution helpers. |
| `utils` | Legacy aggregate for byte, context, hash, password, slice, string, semver, and tokenizer helpers. |

### Adapter and External-Service Packages

These packages are optional. They are used only when a consumer imports them and
supplies the backing service configuration.

| Package path | External dependency / role |
|--------------|----------------------------|
| `db/aerospike` | Aerospike client helpers, entity helpers, geospatial, and HLL support. |
| `db/bunt` | BuntDB connection helper. |
| `db/elasticsearch` | Elasticsearch client helper. |
| `db/gremlin` | Gremlin client, cache, and graph operation helpers. |
| `db/mongodb` | MongoDB client and filter helpers. |
| `db/neo4j` | Neo4j client configuration helpers. |
| `db/postgres` | Postgres connection helpers. |
| `db/redis` | Redis connection helpers. |
| `queue/generic/subscriber` | Generic subscriber contracts and helpers. |
| `queue/generic/subscriber/nats` | NATS adapter for generic subscribers. |
| `queue/nats` | NATS client, middleware, publisher, subscriber, proxy, and propagation helpers. |
| `metrics` | Metrics client and backend-agnostic processor selection. |
| `metrics/processors/datadog` | Datadog metrics processor. |
| `metrics/processors/otel` | OpenTelemetry metrics processor. |
| `metrics/processors/prometheus` | Prometheus metrics processor. |
| `metrics/processors/statsd` | StatsD metrics processor. |
| `fastlog/handlers/datadog` | Datadog `slog` handler support. |
| `fastlog/handlers/logfile` | Legacy opt-in log file handler for compatibility only. |
| `fastlog/handlers/logstash` | Logstash handler support. |
| `fastlog/handlers/nop` | No-op handler support. |
| `fastlog/handlers/otel` | OpenTelemetry log handler support. |
| `fastlog/handlers/stderr` | Structured stderr event-stream handler support. |
| `fastlog/handlers/stdout` | Structured stdout event-stream handler support. |

### Server-Support Packages

These packages help applications build servers, but the application still owns
the `main()`, routes, ports, TLS, auth policy, and graceful shutdown.

| Package path | Use |
|--------------|-----|
| `app` | Application bootstrap helpers for web, metrics, and NATS composition. |
| `fastlog` | Structured `log/slog` configuration, handler fanout, and middleware. |
| `fastlog/middlewares` | Caller, error formatting, and GDPR log middleware. |
| `server/backoff` | HTTP transport retry/backoff helpers. |
| `server/honeypot` | Honeypot helpers. |
| `server/instance` | Runtime instance helpers. |
| `server/jwt` | JWT service, config, models, and Fiber middleware. |
| `server/sse` | Server-sent event listener and pool helpers. |
| `server/view` | Embedded HTML template parsing helpers. |
| `server/template` | Legacy compatibility path for `server/view`. |
| `server/throughput` | Throughput tracking with memory storage. |
| `server/webserver` | Fiber app/server helpers, config, middleware, and request helpers. |

### Test, Resource, and Specialized Packages

| Package path | Use |
|--------------|-----|
| `embedded` | Embedded resource access. |
| `fixtures` | Shared test fixtures. |
| `testassert` | Shared assertions and test helpers. |
| `testutils` | Legacy compatibility path for `testassert`. |
| `profiler`, `server/profiler` | Profiling and health-check support. |
| `machielearning/nn` | Neural-network helpers. |

## Legacy Package Names

Some import paths predate the current package-naming rules. They remain
available for compatibility, but new code should prefer the focused
replacement paths below. Direct in-place renames are reserved for a future
major-version plan. Package docs at each affected path include the detailed
import examples and API-specific migration notes.

| Legacy path | Preferred path |
|-------------|----------------|
| `mathutils` | `mathx` |
| `memory/utils` | `memory/concurrent` for safe containers, `memory/order` for sorting |
| `multiproc/sync` | `multiproc/once` |
| `pki` | `pki/cryptor` |
| `pki/aes` | `pki/aesgcm` |
| `pki/rsa` | `pki/rsaoaep` |
| `queue/generic/subscriber/driver` | `queue/generic/subscriber/nats` |
| `server/template` | `server/view` |
| `testutils` | `testassert` |
| `utils` | Keep existing imports; place new helpers in focused owning packages. |

## Logging Defaults

`fastlog` defaults to structured JSON logs on `stderr` through `LOG_OUTPUTS=stderr:json`.
Applications can choose `stdout:json` for standard-output event streams, and both
stdout/stderr handlers are registered by the `fastlog` package by default.

The `fastlog/handlers/logfile` package is a legacy compatibility handler. It is
not registered by `fastlog` or the `app` bootstrap helpers; consumers must import
it and select `LOG_OUTPUTS=file:json` explicitly when they need local file logging
for an existing deployment.

## Configuration

Consumer applications own configuration. Packages use two compatible paths:

1. Environment parsing helpers for production wiring.
2. Direct struct literals for tests and programmatic composition.

Common environment prefixes:

| Area | Prefix / helper |
|------|-----------------|
| Aerospike | `AEROSPIKE` or caller-supplied prefix through `db/aerospike.GetConnectionConfigFromEnv`. |
| BuntDB | Caller-supplied prefix, default helper usage uses `DB`. |
| Gremlin | `GREMLIN`. |
| MongoDB | `MONGO`. |
| Neo4j | `NEO4J`. |
| Postgres | `POSTGRES`. |
| Redis | `REDIS`. |
| NATS | `NATS` or caller-supplied prefix through `queue/nats/client.GetNATSConnectionConfigFromEnv`. |
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
defined in `.testcoverage.yml`; the current total threshold is `90%`, with
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
  `queue/nats.Publisher`, `metrics.Recorder`, `server/webserver.Client`,
  `server/webserver.Runtime`, and JWT/PKI option/result helpers.
- Additional adapter boundaries cover helper surfaces that used to require
  direct SDK imports: `db/aerospike/entity.RecordStore`,
  `db/aerospike/hll.Counter`, `db/gremlin.GraphStore`,
  `db/mongodb.Filter`, `queue/nats.Subscriber`,
  `queue/nats/middleware.MessageMiddleware`,
  `queue/generic/subscriber.Message`, Redis proxy storage options, and
  `server/webserver.RouteInitializer` / `server/webserver.RouteMiddleware`.
- Legacy SDK-shaped clients remain available for existing consumers. New code
  should prefer the core-owned option/result types so vendor SDKs stay behind
  adapter packages.
- Runtime state now has explicit ownership APIs for new code: use
  `commands.NewEventManager`, `ecs.NewEntityFactory`,
  `oslistener.NewSignalListener`, `profiler.NewState`,
  `metrics.NewRegistry`, `metrics.InstallDefault`, database client stores such
  as `db/postgres.NewClientStore` and `db/redis.NewConnectionStore`,
  `queue/nats/client.ConnectClient`, `fastlog.NewLoggerWithRegistry`,
  config-based log handler constructors, scoped default registry handles, and
  `server/template.NewTemplateWithDir`.
  Legacy package-level `Default`, `Set`, `Get`, and env-reading wrappers remain
  available as deprecated compatibility paths.
- App bootstrap now has error-returning APIs for explicit composition:
  `app.RunWeb` and `app.RunNATS` accept option structs for logger, metrics,
  profiler state, monitor, shutdown listener, routing or subscription setup,
  and env-derived runtime values. `app.WebMain`, `app.WebMainWithOptions`, and
  `app.NATSMain` remain as deprecated logging shims.
