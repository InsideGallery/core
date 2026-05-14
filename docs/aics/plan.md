# Core Library Conformance Plan

Executable task queue for `github.com/InsideGallery/core`. Format and lifecycle rules are defined in
[`docs/aics/plan-format.md`](./plan-format.md). Each task block below is consumed by `./agent.sh`; do not place runner
instructions in this file.

Source standards used for the gap register that produced this plan:

- `docs/source/Go Library.md`
- `docs/source/Clean Code.md`
- `docs/source/Go Best Practice.md`
- `docs/source/Engineering Principles.md`
- `docs/source/solid.md`
- `docs/source/Twelve-Factor App.md`
- `docs/source/tbd.md`
- `docs/source/mdca.md`
- `docs/source/mdca_standard.md`

Layer keys in `WHERE` are kept verbatim per format (`domain`, `repository`, `handler`, `Tests`, `Swagger/docs`).
For this vendor library a few layers are not applicable; those entries are marked `n/a (vendor library)`.

Current observed metrics (informative): unit-test coverage `56.8%` of statements; `.testcoverage.yml` threshold is
`90%`. `Go Library.md` §8 requires ≥`70%`. Several packages are explicitly excluded (require external
infrastructure) and are not part of these tasks.

---

### TODO CORE-LIB-01 (MISSING): Lift `fastlog` package family to ≥70% unit-test coverage

- WHAT: Add table-driven unit tests for `fastlog` config and the in-process handlers (`stdout`, `stderr`, `nop`,
  `logfile`), middlewares, and the handler registry. Tests must run without external services. Use the existing
  `testutils` helpers and `t.TempDir()` for `logfile`. Network-bound handlers (`otel`, `datadog`, `logstash`) stay in
  the coverage-exclusion list and are out of scope.
- WHERE:
  Layer `domain`: `fastlog/config.go`, `fastlog/middlewares/*.go`.
  Layer `repository`: `fastlog/handlers/registry.go`, `fastlog/handlers/{stdout,stderr,nop,logfile}/*.go`.
  Layer `handler`: n/a (vendor library).
  Tests: `fastlog/config_test.go`, `fastlog/middlewares/*_test.go`,
  `fastlog/handlers/{registry,stdout,stderr,nop,logfile}_test.go`.
  Swagger/docs: n/a (vendor library).
- WHY: Engineering Principles forbid coverage decrease; Go Library.md §8 requires ≥70%; current `fastlog/*` family is
  near 0% and is on the import path of nearly every consumer service. Untested logging code is the worst place to
  ship regressions because it masks observability for everything else.
- References: `docs/source/Go Library.md`, `docs/source/Clean Code.md`, `docs/source/Engineering Principles.md`.
- Backward compatibility: Yes (additive/parity-preserving change).
- Status: DONE.

### TODO CORE-LIB-02 (MISSING): Lift `server/jwt` family to ≥70% unit-test coverage

- WHAT: Add table-driven tests for `server/jwt` token generation/validation, `server/jwt/model` scope parsing, and
  `server/jwt/middlewares` happy and failure paths. Cover boundary conditions: empty scope, malformed token, expired
  token, unsigned token, missing header. Use stdlib `crypto/*` — do not introduce new dependencies. Mocks live in
  the `_test` package per Clean Code §6.
- WHERE:
  Layer `domain`: `server/jwt/jwt.go`, `server/jwt/model/scope.go`.
  Layer `repository`: n/a (vendor library).
  Layer `handler`: `server/jwt/middlewares/*.go`.
  Tests: `server/jwt/jwt_test.go`, `server/jwt/model/scope_test.go`,
  `server/jwt/middlewares/*_test.go`.
  Swagger/docs: n/a (vendor library).
- WHY: JWT is security-sensitive; LSP and ISP demand that every implementation honor the same contract. Without
  tests, contract drift between scope parsing and middleware enforcement is silent.
- References: `docs/source/solid.md`, `docs/source/Clean Code.md`, `docs/source/Go Library.md`.
- Backward compatibility: Yes (additive/parity-preserving change).
- Status: DONE.

### TODO CORE-LIB-03 (MISSING): Lift `memory/{linkedlist,registry,sortedset}` to ≥70% unit-test coverage

- WHAT: Add table-driven tests covering boundary conditions (empty, single element, duplicate keys, large key range,
  concurrent reads where the type is documented as concurrency-safe). Each test must `t.Parallel()` when independent.
  Do not change exported APIs.
- WHERE:
  Layer `domain`: `memory/linkedlist/list.go`, `memory/registry/registry.go`,
  `memory/sortedset/sortedset.go`.
  Layer `repository`: n/a (vendor library).
  Layer `handler`: n/a (vendor library).
  Tests: `memory/linkedlist/list_test.go`, `memory/registry/registry_test.go`,
  `memory/sortedset/sortedset_test.go`.
  Swagger/docs: n/a (vendor library).
- WHY: These data structures are used in hot paths by consumers; reaching parity with the rest of `memory/*`
  prevents coverage regression and validates LSP for the `Comparable`/`Iterator` contracts.
- References: `docs/source/Clean Code.md`, `docs/source/solid.md`, `docs/source/Go Best Practice.md`.
- Backward compatibility: Yes (additive/parity-preserving change).
- Status: DONE.

### TODO CORE-LIB-04 (MISSING): Add explicit-init wrappers for `init()` side-effect packages

- WHAT: For packages whose `init()` mutates global state, expose an exported, idempotent setup function (e.g.
  `Setup(ctx context.Context) error`) that performs the same configuration explicitly. Keep the existing `init()`
  for compatibility, but document it as deprecated and have it call the new function. Concretely:
  `db/aerospike/connection.go` (`buffer.Arch64Bits/Arch32Bits`), `db/gremlin/syntax.go` (env-driven syntax),
  `fastlog/handlers/otel/otel.go`, `fastlog/handlers/datadog/datadog.go`. The wrappers must accept dependencies as
  arguments and never read `os.Getenv` directly — env parsing belongs in the existing config structs.
- WHERE:
  Layer `domain`: `db/aerospike/connection.go`, `db/gremlin/syntax.go`,
  `fastlog/handlers/otel/otel.go`, `fastlog/handlers/datadog/datadog.go`.
  Layer `repository`: n/a (vendor library).
  Layer `handler`: n/a (vendor library).
  Tests: `db/aerospike/connection_test.go`, `db/gremlin/syntax_test.go`,
  `fastlog/handlers/otel/otel_test.go`, `fastlog/handlers/datadog/datadog_test.go`.
  Swagger/docs: n/a (vendor library).
- WHY: Twelve-Factor App §III/IV and Go Best Practice §11, plus SOLID DIP and Clean Code Go2, all forbid I/O and
  hidden state in `init()`. Keeping `init()` as a thin shim around an exported `Setup` is parity-preserving while
  giving consumers a deterministic composition root.
- References: `docs/source/Twelve-Factor App.md`, `docs/source/Go Best Practice.md`, `docs/source/solid.md`,
  `docs/source/mdca.md`.
- Backward compatibility: Yes (additive/parity-preserving change).
- Status: DONE.

### TODO CORE-LIB-05 (MISSING): Add value-returning aliases for `Get`-prefixed getters

- WHAT: Per Clean Code §1 and Go Best Practice §2, getters must not use the `Get` prefix. Introduce
  Go-idiomatic alias methods (`Subject()`, `Header()`, `ReadTimeout()`, `MaxConcurrentSize()`, `FS()`, `ID()`, etc.)
  that delegate to the existing `Get*` methods. Do not remove the legacy methods in this task; mark them
  `// Deprecated: use <NewName>` so consumers can migrate without a breaking change. Affected callsites within the
  module must use the new names.
- WHERE:
  Layer `domain`: `github.com/FrogoAI/mq-balancer/subscriber/driver`,
  `github.com/FrogoAI/mq-balancer/subscriber/mq`, `embedded/resources.go`, `ticker/ticker.go`.
  Layer `repository`: n/a (vendor library).
  Layer `handler`: n/a (vendor library).
  Tests: `github.com/FrogoAI/mq-balancer` subscriber tests, `embedded/resources_test.go`,
  `ticker/ticker_test.go`.
  Swagger/docs: n/a (vendor library).
- WHY: Naming consistency is enforced by Clean Code G11 and Go Best Practice §2. Adding aliases avoids breaking
  existing consumers while letting new code adopt idiomatic names; the deprecation comment is the migration plan
  required by Engineering Principles §"API Development & Versioning".
- References: `docs/source/Clean Code.md`, `docs/source/Go Best Practice.md`, `docs/source/Engineering Principles.md`.
- Backward compatibility: Yes (additive/parity-preserving change).
- Status: DONE.

### TODO CORE-LIB-06 (MISSING): Replace silent fallbacks with explicit errors at boundaries

- WHAT: Convert silent fallbacks/log-and-continue paths into returned errors at package boundaries while preserving
  the current default-success path via wrapper helpers (e.g. `MustX` callers stay, but the lower-level function
  surfaces the error). Specifically: the removed local NATS proxy no longer owns Redis lock retries;
  `github.com/FrogoAI/mq-balancer/subscriber/driver/client` owns NATS auth option construction;
  `pki/aes/aes.go::NewAES` returns `(*AES, error)` for unknown sizes (keep `NewAES` signature, add `NewAESStrict`
  as the explicit-error variant).
- WHERE:
  Layer `domain`: `github.com/FrogoAI/mq-balancer/subscriber/driver/client`,
  `pki/aes/aes.go`.
  Layer `repository`: n/a (vendor library).
  Layer `handler`: n/a (vendor library).
  Tests: `github.com/FrogoAI/mq-balancer` client tests,
  `pki/aes/aes_test.go`.
  Swagger/docs: n/a (vendor library).
- WHY: Clean Code §2 (Error Handling), Go Best Practice §10, and SOLID LSP all require that contracts be honored;
  silent failure is undefined behavior at a boundary and produces ghost bugs in callers. Adding strict variants is
  additive and backward-compatible.
- References: `docs/source/Clean Code.md`, `docs/source/Go Best Practice.md`, `docs/source/solid.md`.
- Backward compatibility: Yes (additive/parity-preserving change).
- Status: DONE.

### TODO CORE-LIB-07 (MISSING): Normalize error string conventions across packages

- WHAT: Apply Go Best Practice §10 to all error strings: lowercase first letter, no trailing punctuation, no
  redundant `error ...` prefix (the wrapping caller already says "error"). Fix the typo "responset" → "response" in
  `db/elasticsearch/elasticsearch.go` (4 occurrences), the capitalized prefixes in
  `fastlog/handlers/{otel,datadog}/*.go`, `fastlog/metrics/metrics.go`, the grammar of
  `db/aerospike/connection.go`, the order in `db/aerospike/entity/errors.go` (`attribute not found`), and the
  redundant `error ...` prefix in `dataconv/errors.go`.
- WHERE:
  Layer `domain`: `db/elasticsearch/elasticsearch.go`, `fastlog/handlers/otel/otel.go`,
  `fastlog/handlers/datadog/datadog.go`, `fastlog/metrics/metrics.go`,
  `db/aerospike/connection.go`, `db/aerospike/entity/errors.go`, `dataconv/errors.go`.
  Layer `repository`: n/a (vendor library).
  Layer `handler`: n/a (vendor library).
  Tests: `dataconv/errors_test.go` (assert exact strings if asserted; otherwise no test changes needed).
  Swagger/docs: n/a (vendor library).
- WHY: Inconsistent error strings break log-aggregation queries downstream and signal to readers that the codebase
  has no enforced conventions (Clean Code G11, G24).
- References: `docs/source/Go Best Practice.md`, `docs/source/Clean Code.md`.
- Backward compatibility: Yes (additive/parity-preserving change).
- Status: DONE.

### TODO CORE-LIB-08 (MISSING): Replace remaining magic numbers with named constants

- WHAT: Introduce named constants for the remaining magic literals flagged in the root-level review:
  `0xFFFF` in `memory/sortedset/sortedset.go:43` (e.g. `const maxSegmentMask = 0xFFFF`), and the hardcoded
  `time.Second` debug-tick interval in `ticker/ticker.go:125` (e.g. `const debugTickInterval = time.Second`). Do not
  change observable behavior.
- WHERE:
  Layer `domain`: `memory/sortedset/sortedset.go`, `ticker/ticker.go`.
  Layer `repository`: n/a (vendor library).
  Layer `handler`: n/a (vendor library).
  Tests: `memory/sortedset/sortedset_test.go`, `ticker/ticker_test.go`.
  Swagger/docs: n/a (vendor library).
- WHY: Clean Code G16/G25 — magic numbers obscure intent and break searchability when the value needs to change.
- References: `docs/source/Clean Code.md`, `docs/source/Go Best Practice.md`.
- Backward compatibility: Yes (additive/parity-preserving change).
- Status: DONE.

### TODO CORE-LIB-09 (MISSING): Refactor `memory/sortedset.GetByKeyRange` into smaller functions

- WHAT: Extract sub-functions from `GetByKeyRange` so each is under 40 lines, removing the existing
  `nolint:gocyclo`. Preserve the exact public API and behavior; this is an internal decomposition. Add unit tests
  that assert behavior on boundary cases (empty range, range below min, range above max, range crossing both ends,
  duplicate keys).
- WHERE:
  Layer `domain`: `memory/sortedset/sortedset.go`.
  Layer `repository`: n/a (vendor library).
  Layer `handler`: n/a (vendor library).
  Tests: `memory/sortedset/sortedset_test.go`.
  Swagger/docs: n/a (vendor library).
- WHY: Clean Code §2 (Functions: small, one-thing) and G34 (one level of abstraction). A 119-line function with a
  cyclomatic-complexity suppression is the textbook smell.
- References: `docs/source/Clean Code.md`, `docs/source/Go Best Practice.md`.
- Backward compatibility: Yes (additive/parity-preserving change).
- Status: DONE.

### TODO CORE-LIB-10 (MISSING): Document and harden `memory/orderedmap.Iterator` thread-safety

- WHAT: `Iterator()` spawns a goroutine that streams the current map content, but mutation during iteration is
  unspecified. Either snapshot the keys/values into a local slice before sending (preferred — preserves the
  existing channel-based API), or add a documented warning on the method that the caller MUST not mutate during
  iteration plus a `t.Run` table-test that asserts the documented behavior under `-race`.
- WHERE:
  Layer `domain`: `memory/orderedmap/orderedmap.go`.
  Layer `repository`: n/a (vendor library).
  Layer `handler`: n/a (vendor library).
  Tests: `memory/orderedmap/orderedmap_test.go`.
  Swagger/docs: n/a (vendor library).
- WHY: Clean Code §7 and Go Best Practice §9 require that concurrency contracts be explicit. An undocumented race
  is an LSP violation: any consumer assuming snapshot semantics is wrong, and any consumer assuming live-view
  semantics is also wrong.
- References: `docs/source/Clean Code.md`, `docs/source/Go Best Practice.md`, `docs/source/solid.md`.
- Backward compatibility: Yes (additive/parity-preserving change).
- Status: DONE.

### TODO CORE-LIB-11 (MISSING): Remove debug `fmt.Println` calls from test files

- WHAT: Replace remaining `fmt.Println` calls in tests with `t.Logf`, or delete them. Files flagged in the
  root-level review: `utils/hash_test.go`, `utils/semver/semver_test.go`, `utils/tokenizer/prepare_test.go`,
  `utils/strings_test.go`, `mathutils/helper_test.go`, `db/mongodb/mongodb_test.go`,
  `db/aerospike/hll/count_test.go`. Test output should travel through `testing.T` so `-v` and CI redaction work
  uniformly.
- WHERE:
  Layer `domain`: n/a (test-only change).
  Layer `repository`: n/a (test-only change).
  Layer `handler`: n/a (test-only change).
  Tests: `utils/hash_test.go`, `utils/semver/semver_test.go`, `utils/tokenizer/prepare_test.go`,
  `utils/strings_test.go`, `mathutils/helper_test.go`, `db/mongodb/mongodb_test.go`,
  `db/aerospike/hll/count_test.go`.
  Swagger/docs: n/a (vendor library).
- WHY: Clean Code §3 (no debug noise) and Go Library.md §9 forbid `fmt.Print*` in production code; even in tests
  these calls escape coverage redirection and leak through CI logs.
- References: `docs/source/Clean Code.md`, `docs/source/Go Library.md`.
- Backward compatibility: Yes (additive/parity-preserving change).
- Status: DONE.

### TODO CORE-LIB-12 (MISSING): Refactor remaining `else`-after-return blocks to early returns

- WHAT: Apply Go Best Practice §3 to the call sites flagged in the root-level review:
  `dataconv/ip.go:162`, `memory/fuzzysearch/levenshtein.go:38-45`,
  `memory/linkedlist/list.go:52-56`, `mathutils/helper.go:25-34`,
  `fastlog/middlewares/gdpr.go:45-46`,
  `fastlog/middlewares/error.go:33-38`, `pki/aescmac/aescmac.go:73,82`. Each `if … return` block drops its `else`
  and the happy path stays left-aligned.
- WHERE:
  Layer `domain`: `dataconv/ip.go`, `memory/fuzzysearch/levenshtein.go`,
  `memory/linkedlist/list.go`, `mathutils/helper.go`,
  `fastlog/middlewares/gdpr.go`, `fastlog/middlewares/error.go`,
  `pki/aescmac/aescmac.go`.
  Layer `repository`: n/a (vendor library).
  Layer `handler`: n/a (vendor library).
  Tests: existing unit tests must continue to pass; no new tests required because behavior is preserved.
  Swagger/docs: n/a (vendor library).
- WHY: Clean Code G29 and Go Best Practice §3 — early return keeps the happy path readable; trailing `else` adds
  indentation without information.
- References: `docs/source/Go Best Practice.md`, `docs/source/Clean Code.md`.
- Backward compatibility: Yes (additive/parity-preserving change).
- Status: DONE.

### TODO CORE-LIB-13 (MISSING): Drop redundant `err` returns from env-config helpers

- WHAT: For every `GetConfigFromEnv()` whose only error path is the env-decoder result already returned earlier,
  return `c, nil` explicitly (not `c, err` with `err` known nil). Files flagged: `db/bunt/config.go:28`,
  `db/mongodb/config.go:37`, `db/aerospike/config.go:35`, `db/postgres/config.go:36`, `db/redis/config.go:29`,
  `db/neo4j/config.go:42`, `fastlog/config.go:42`, the seven `fastlog/handlers/*/config.go`,
  `fastlog/metrics/config.go:27`. The function signatures stay the same (`(Config, error)`), only the body is
  cleaned up.
- WHERE:
  Layer `domain`: `db/bunt/config.go`, `db/mongodb/config.go`, `db/aerospike/config.go`,
  `db/postgres/config.go`, `db/redis/config.go`, `db/neo4j/config.go`, `fastlog/config.go`,
  `fastlog/handlers/datadog/config.go`, `fastlog/handlers/logfile/config.go`,
  `fastlog/handlers/logstash/config.go`, `fastlog/handlers/nop/config.go`,
  `fastlog/handlers/otel/config.go`, `fastlog/handlers/stderr/config.go`,
  `fastlog/handlers/stdout/config.go`, `fastlog/metrics/config.go`.
  Layer `repository`: n/a (vendor library).
  Layer `handler`: n/a (vendor library).
  Tests: existing config tests; add cases for invalid env values where missing.
  Swagger/docs: n/a (vendor library).
- WHY: Clean Code G26 (be precise) and Go Best Practice §10. A guaranteed-nil error confuses readers and creates a
  false branch in static analysis.
- References: `docs/source/Clean Code.md`, `docs/source/Go Best Practice.md`,
  `docs/source/Twelve-Factor App.md`.
- Backward compatibility: Yes (additive/parity-preserving change).
- Status: DONE.

### TODO CORE-LIB-14 (MISSING): Remove or implement commented-out blocks flagged in review

- WHAT: Delete the commented-out validation block in `server/jwt/model/scope.go:62-73` (or implement it under a
  documented contract and add a unit test). Delete the commented-out `SetFilter` calls in
  the removed local NATS proxy storage and `utils/bytes.go:72-73`. Each removed block stays
  documented in git history per Clean Code C5 — no inline TODO comments left behind without a tracking ticket.
- WHERE:
  Layer `domain`: `server/jwt/model/scope.go`,
  `utils/bytes.go`.
  Layer `repository`: n/a (vendor library).
  Layer `handler`: n/a (vendor library).
  Tests: `server/jwt/model/scope_test.go` (covers any newly-implemented validation),
  `utils/bytes_test.go`.
  Swagger/docs: n/a (vendor library).
- WHY: Clean Code C5/G9, tbd.md "what's forbidden on trunk", and the Boy Scout Rule. Commented code is dead code
  that pretends to be live.
- References: `docs/source/Clean Code.md`, `docs/source/tbd.md`.
- Backward compatibility: Yes (additive/parity-preserving change).
- Status: DONE.

### TODO CORE-LIB-15 (MISSING): Add `Setup`-style composition root for `ecs` global registry

- WHAT: Wrap the package-level `store` in `ecs/base.go` behind a `Registry` struct with `NewRegistry()` and
  expose a default registry for backward compatibility (`var Default = NewRegistry()`). Existing package-level
  functions delegate to `Default`. New consumers can pass a `*Registry` explicitly. No call site needs to change.
- WHERE:
  Layer `domain`: `ecs/base.go`.
  Layer `repository`: n/a (vendor library).
  Layer `handler`: n/a (vendor library).
  Tests: `ecs/base_test.go` (cover `NewRegistry`, isolation between registries, parallel tests).
  Swagger/docs: n/a (vendor library).
- WHY: SOLID DIP and MDCA composition-root rule forbid undecorated package-level state; offering a `Registry` makes
  ECS testable in isolation without removing the legacy global. This is the same pattern used for
  `db/aerospike/ConnectionRegistry`.
- References: `docs/source/solid.md`, `docs/source/mdca.md`,
  `docs/source/Twelve-Factor App.md`.
- Backward compatibility: Yes (additive/parity-preserving change).
- Status: DONE.

### TODO CORE-LIB-16 (MISSING): Add CI gate that fails when test coverage drops

- WHAT: Update `.github/workflows/go.yml` to run `go-test-coverage --config=./.testcoverage.yml` as a
  required job, so any merge that lowers coverage below the configured threshold fails before review. Verify the
  threshold value in `.testcoverage.yml` matches the team's stated target (Go Library.md §8 = 70% minimum) and
  document the per-package exclusions already enumerated in the root-level `plan.md`.
- WHERE:
  Layer `domain`: n/a (CI/tooling change).
  Layer `repository`: n/a (CI/tooling change).
  Layer `handler`: n/a (CI/tooling change).
  Tests: `.github/workflows/go.yml`, `.testcoverage.yml`.
  Swagger/docs: n/a (vendor library).
- WHY: Engineering Principles §"Development & Testing" forbid coverage decrease; tbd.md §"CI Requirements" requires
  the green-trunk invariant be machine-enforced. Without a gate the rule is aspirational.
- References: `docs/source/Engineering Principles.md`, `docs/source/tbd.md`,
  `docs/source/Go Library.md`.
- Backward compatibility: Yes (additive/parity-preserving change).
- Status: DONE.

### TODO CORE-LIB-17 (MISSING): Ship `fastlog/all` and `metrics/all` plugin bundles

- WHAT: Create two new packages whose only purpose is to import every in-tree plugin so its `init()` registers
  with the default registry. `fastlog/all/all.go` does `import _` for `stderr`, `stdout`, `nop`, `logfile`,
  `logstash`, `otel`, `datadog`. `metrics/all/all.go` does `import _` for `datadog`, `otel`, `prometheus`,
  `statsd`. Each bundle file gets a build tag `//go:build !<bundle>_minimal` so consumers that need a smaller
  binary can opt out (`go build -tags fastlog_minimal,metrics_minimal`). The per-handler import path is
  preserved; bundle imports are additive. Document the pattern in the new `fastlog/all/doc.go` and
  `metrics/all/doc.go` (one short paragraph each — these earn their `doc.go` keep).
- WHERE:
  Layer `domain`: `fastlog/all/all.go`, `metrics/all/all.go`.
  Layer `repository`: n/a (vendor library).
  Layer `handler`: n/a (vendor library).
  Tests: `fastlog/all/all_test.go`, `metrics/all/all_test.go` (assert that every expected `OutKind`/processor
  is registered on `DefaultRegistry()` after a blank import).
  Swagger/docs: `fastlog/all/doc.go`, `metrics/all/doc.go`.
- WHY: Today every backend swap is `code change + config change` — the consumer must `import _ "<handler>"` AND
  set env. That violates Twelve-Factor IV (backing services swappable via config alone) and KISS. The
  `database/sql` driver pattern is the established Go answer; bundle packages collapse N imports into one and
  let the operator move between backends with env only.
- References: `docs/source/Twelve-Factor App.md`, `docs/source/Engineering Principles.md`,
  `docs/source/Go Library.md`, `docs/source/Go Best Practice.md`.
- Backward compatibility: Yes (additive/parity-preserving change).
- Status: DONE.

### TODO CORE-LIB-18 (MISSING): Add `fastlog.SetupDefault` one-call bootstrap

- WHAT: Add `fastlog.SetupDefault(ctx context.Context, cfg *Config, m ...slogmulti.Middleware) (func() error,
  error)` that builds the logger from `cfg`, calls `slog.SetDefault(logger)`, and returns a close function that
  restores the previous default and flushes the underlying handler. Internally compose the existing
  `NewLoggerWithRegistry` + `InstallDefaultLogger`. Mark `NewLogger` and the standalone `InstallDefaultLogger`
  for callers that need bespoke wiring; do not remove them. After this lands, the canonical bootstrap is two
  lines: `close, err := fastlog.SetupDefault(ctx, cfg)` then `defer close()` — and every downstream package
  uses `slog.Info(...)` directly.
- WHERE:
  Layer `domain`: `fastlog/log.go`.
  Layer `repository`: n/a (vendor library).
  Layer `handler`: n/a (vendor library).
  Tests: `fastlog/log_test.go` (table-driven: nil cfg → error; valid cfg → `slog.Default()` swapped; close
  restores previous default).
  Swagger/docs: n/a (vendor library).
- WHY: Twelve-Factor XI ("Treat logs as event streams") and Go Library §11 expect `slog` to be the universal
  logging contract. `Go Best Practice.md` §11 forbids hidden init outside bootstrap code; app startup should
  install the process logger once and downstream code should use `slog.Default()`.
- References: `docs/source/Twelve-Factor App.md`, `docs/source/Go Library.md`,
  `docs/source/Go Best Practice.md`, `docs/source/Engineering Principles.md`.
- Backward compatibility: Yes (additive/parity-preserving change).
- Status: DONE.

### TODO CORE-LIB-19 (MISSING): Keep app bootstrap in `WebMain` and `NATSMain`

- WHAT: Keep the app package focused on the two simple main-style bootstrap helpers: `app.WebMain` and
  `app.NATSMain`. They perform the full bootstrap directly: fastlog env config → `fastlog.SetupDefault`
  (CORE-LIB-18) → metrics env config → profiler monitor/probes → oslistener shutdown hooks → web routes or
  NATS subscriptions. Do not add alternate exported run entrypoints, app-level config structs, option structs,
  or an `any`/option-union switch.
- WHERE:
  Layer `domain`: `app/web.go`, `app/nats.go`.
  Layer `repository`: n/a (vendor library).
  Layer `handler`: n/a (vendor library).
  Tests: `app/web_test.go`, `go test ./app`, `go test ./...`, `go test -race -count=1 ./...`,
  `golangci-lint run ./...`.
  Swagger/docs: n/a (vendor library).
- WHY: The app package should remain a tiny bootstrap convenience layer. Extra config structs and run variants
  obscure the original API and violate KISS for a library that already has `WebMain` and `NATSMain`.
- References: `docs/source/Engineering Principles.md`, `docs/source/Clean Code.md`,
  `docs/source/solid.md`, `docs/source/mdca.md`, `docs/source/Twelve-Factor App.md`.
- Backward compatibility: Yes (additive/parity-preserving change).
- Status: DONE.

### TODO CORE-LIB-20 (MISSING): Drop redundant `Logger` accessor and option fields once `slog.Default` is canonical

- WHAT: After CORE-LIB-18 lands, keep app bootstrap logging on `fastlog.SetupDefault` and `slog.Default()`.
  Internal app code must not expose logger accessors or option structs. Other packages should use
  `slog.Default()` unless a function is genuinely a logger sink, e.g. middleware.
- WHERE:
  Layer `domain`: `app/web.go`, `app/nats.go`, `fastlog/log.go`,
  any `server/*` and `metrics/*` that currently take a `*slog.Logger` purely to forward it.
  Layer `repository`: n/a (vendor library).
  Layer `handler`: n/a (vendor library).
  Tests: existing tests must keep passing.
  Swagger/docs: n/a (vendor library).
- WHY: Twelve-Factor XI and the user's own constraint ("logger must be initiated and prepared in config; logs
  must be available via `slog`, not by getting a logger"). Two ways to do the same thing is a Clean Code G11
  inconsistency. Once a single configured `slog.Default()` is the contract, accessor methods become noise.
- References: `docs/source/Twelve-Factor App.md`, `docs/source/Clean Code.md`,
  `docs/source/Go Best Practice.md`.
- Backward compatibility: Yes (additive/parity-preserving change).
- Status: DONE.

### TODO CORE-LIB-21 (MISSING): Remove redundant `doc.go` files; keep only those with real prose

- WHAT: Audit every `doc.go` (currently 34 files). Rule: keep `doc.go` only when it has ≥30 lines of
  package-overview prose that genuinely helps a reader. Otherwise move the `// Package <name> ...` comment to
  the top of the package's main `.go` file (the one that introduces the primary type) and delete the empty
  shell. Expected outcome: ~9 `doc.go` files survive (those covering `db/*`, `fastlog`, `app`, `mdca`-heavy
  subsystems). The audit report (which files kept, which deleted, why) goes in the PR description, not into
  source. The app package keeps its package comment in `app/web.go`, not a separate `app/doc.go`.
- WHERE:
  Layer `domain`: every `**/doc.go` under the repository root (full list from
  `find . -name doc.go`).
  Layer `repository`: n/a (vendor library).
  Layer `handler`: n/a (vendor library).
  Tests: none (deletions only); CI must still pass with `go vet ./...` and `golangci-lint run ./...` since
  some linters require package comments.
  Swagger/docs: n/a (vendor library).
- WHY: Clean Code C3 (redundant comments) and G12 (clutter): a 4-line `doc.go` that restates the package name
  is noise. Keeping a separate file for two sentences violates KISS. The remaining `doc.go` files earn their
  keep by carrying real architectural overview.
- References: `docs/source/Clean Code.md`, `docs/source/Engineering Principles.md`,
  `docs/source/Go Library.md`.
- Backward compatibility: Yes (additive/parity-preserving change).
- Status: DONE.

### TODO CORE-LIB-22 (MISSING): Document the bundle + `slog.Default` workflow in `README.md` and `AGENTS.md`

- WHAT: Add a "Quick Start" section to `README.md` showing the canonical bootstrap (`app.WebMain` or
  `app.NATSMain`, blank-import the bundles), and a matching note in `AGENTS.md` §10
  ("Rules") that future contributors must use `slog.Default()` everywhere downstream and the bundle import
  pattern for new backends. Cross-link `docs/source/Twelve-Factor App.md` Factor IV and §11.
- WHERE:
  Layer `domain`: n/a (documentation change).
  Layer `repository`: n/a (documentation change).
  Layer `handler`: n/a (documentation change).
  Tests: n/a (documentation change).
  Swagger/docs: `README.md`, `AGENTS.md`.
- WHY: A pattern is only adopted if it is documented where contributors look first. Engineering Principles
  §"Documentation" requires visible architecture decisions; without this update CORE-LIB-17..20 land but new
  code keeps reproducing the old pattern.
- References: `docs/source/Engineering Principles.md`, `docs/source/Twelve-Factor App.md`,
  `docs/source/Go Library.md`.
- Backward compatibility: Yes (additive/parity-preserving change).
- Status: DONE.
