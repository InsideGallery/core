# Instruction: How to Build a Go Vendor Library

Blueprint for reusable Go modules (`go get`-able packages) imported by other projects. This is **not** an application -- there is no `main()`, no `cmd/`, no HTTP server, no CLI, no deployment. The consumer application owns all of that.

---

## 1. What a Vendor Library Is

A vendor library is a self-contained Go module that:

- Is imported by applications via `go get github.com/Org/mylib`
- Exports types, functions, and interfaces -- never a `main` package
- Has **zero knowledge** of the consuming application (no framework imports, no routing, no DI containers)
- Manages its own tests, linting, benchmarks, and CI independently
- Ships a `README.md` that shows how to import and use it

---

## 2. Folder Structure

```
mylib/
  go.mod
  go.sum

  # --- Root package = public API surface ---
  service.go                      # Exported orchestrator type the consumer instantiates
  config.go                       # Config struct with env tags + direct construction
  error.go                        # All sentinel errors
  <engine>.go                     # Core algorithm / computation (internal to the lib)
  utils.go                        # Stateless pure helpers

  # --- Domain model ---
  model/
    <entity>.go                   # Plain data structs. ZERO dependencies.

  # --- Persistence abstraction ---
  repositories/
    storage.go                    # Interface ONLY
    memory/                       # In-memory impl (ships with the lib for testing)
      memory.go
      memory_test.go
    <provider>/                   # Optional: production impl (database, cache, etc.)
      connector.go
      connector_test.go           # Integration tests (build-tag gated)
      benchmark_test.go

  # --- Tests (colocated, same package) ---
  service_test.go
  <engine>_test.go
  utils_test.go
  config_test.go
  integration_test.go             # Build-tag gated

  # --- Tooling ---
  Makefile
  .golangci.yml
  .testcoverage.yml
  .gitignore
  .github/
    workflows/go.yml
    dependabot.yml

  # --- Metadata ---
  README.md
  AGENTS.md
  CODEOWNERS
  LICENSE

  # --- Test fixtures ---
  testdata/
  benchmarks/
    baseline.txt                  # Committed
```

### What is NOT here

- No `cmd/` -- this is not an application.
- No `main.go` -- there is nothing to run.
- No `internal/` -- everything is either exported (root package) or a sub-package. The consumer decides what to use.
- No `pkg/` -- the root package IS the package.
- No `docker-compose.yml` or deployment configs -- the consuming app owns infrastructure.

---

## 3. Consumer Experience

The library must be usable in 3 lines:

```go
import "github.com/Org/mylib"
import "github.com/Org/mylib/repositories/memory"

repo := memory.NewRepository()
svc  := mylib.NewService(repo, &mylib.Config{...})
result, err := svc.DoWork(ctx, input)
```

### What the consumer provides

- A `repositories.Storage` implementation (or uses the bundled `memory` one)
- A `Config` (either from env or a struct literal)
- A `context.Context`

### What the library provides

- The service type with its public methods
- The interface for storage
- A ready-to-use in-memory implementation
- Optionally, a production-grade implementation (e.g., database connector)
- Config struct with env parsing
- Sentinel errors for the consumer to check with `errors.Is`

---

## 4. Architecture

```
Consumer App
  |
  v
[service.go]       <-- Exported type. The only entry point.
  |         |
  v         v
[engine]   [Storage interface]
  |            |           |
  v            v           v
[utils]   [memory/]   [provider/]
```

### Dependency rules

**Allowed:**
```
service  -->  engine, utils, model, repositories (interface only)
engine   -->  utils, model
memory   -->  model
provider -->  model
```

**Forbidden:**
```
model       -/->  anything (zero imports)
engine      -/->  service, repositories
repositories -/-> service
memory      -/->  provider (or vice versa)
ANY file    -/->  framework, HTTP, CLI, DI container, or application-level packages
```

### Key constraint: no infrastructure leakage

The root package and `model/` must never import:
- Web frameworks (fiber, gin, echo)
- Database drivers directly (pgx, mongo) -- only through the `repositories/` interface
- Message queues, observability SDKs, or any application-level infra

These belong in the `repositories/<provider>/` sub-package or in the consuming application.

---

## 5. File Responsibilities

| File | Contains | Does NOT contain |
|---|---|---|
| `service.go` | Exported constructor, public methods, orchestration | Algorithms, config parsing, error definitions |
| `config.go` | Config struct, env parsing, derived-value methods | Business logic |
| `error.go` | `var Err... = errors.New(...)` | Logic, types, functions |
| `<engine>.go` | Core algorithm, internal computation type | IO, storage calls, orchestration |
| `utils.go` | Pure functions (no receiver, no state, no side effects) | Anything that needs state |
| `model/<entity>.go` | Plain structs | Methods with external dependencies, imports |
| `repositories/storage.go` | Interface definition | Implementations |

---

## 6. Configuration

### Two construction paths (both must work)

```go
// Path 1: Environment variables (production)
cfg, err := mylib.GetConfigFromEnv()

// Path 2: Struct literal (tests, programmatic use)
cfg := &mylib.Config{ParamA: 20, ParamB: 0.6, Seed: 42}
```

### Rules

- Every field has a sensible `envDefault`.
- Derived values are methods on Config, not standalone functions.
- Config is a plain struct -- no hidden initialization, no `init()`.

---

## 7. Error Handling

All sentinel errors live in `error.go`:

```go
var (
    ErrEmptyInput    = errors.New("empty input")
    ErrInvalidConfig = errors.New("invalid configuration")
)
```

| Rule | Detail |
|---|---|
| Never ignore errors | `_ = fn()` is forbidden |
| Critical errors | Return to caller |
| Non-critical errors | `slog.Warn` with structured fields, continue |
| Logging | `log/slog` only. No `fmt.Print*`, no `log.Print*` |
| Wrapping | `fmt.Errorf("context: %w", err)` |

The consumer checks errors with `errors.Is(err, mylib.ErrEmptyInput)`.

---

## 8. Testing

### Table-driven tests (mandatory, no exceptions)

```go
cases := []struct {
    name    string
    input   string
    want    string
    wantErr error
}{...}

for _, tc := range cases {
    t.Run(tc.name, func(t *testing.T) { ... })
}
```

### Test types

| Type | Build tag | Backend | CI runs it |
|---|---|---|---|
| Unit | None | `memory.NewRepository()` | Always |
| Integration | `//go:build integration` | Real database | On demand |
| Benchmark | Optional | Either | Smoke-run in CI |

### The in-memory implementation is part of the library

`repositories/memory/` ships with the library. This means:
- Consumers can use it in **their** tests too
- Unit tests never need external services
- Every code path is testable without Docker

### Spy wrappers for behavior verification

```go
type spyRepo struct {
    *memory.Repository
    saveCalls int64
}

func (s *spyRepo) SaveRecord(r model.Record) error {
    atomic.AddInt64(&s.saveCalls, 1)
    return s.Repository.SaveRecord(r)
}
```

### Benchmarks

- Sub-benchmarks via `b.Run` for different scenarios.
- Committed `benchmarks/baseline.txt`, gitignored `current.txt`.
- Compare with `benchstat`.

### Coverage

- Threshold: 70%+ in `.testcoverage.yml`.
- Exclude: generated code, mocks, integration-only packages.

---

## 9. Linting

### .golangci.yml

```yaml
version: "2"
linters:
  enable:
    - errcheck, govet, staticcheck, gosec, revive
    - misspell, lll, forbidigo, mnd, wsl_v5
formatters:
  enable:
    - gofmt, goimports, gci
```

### Import order (enforced by `gci`)

```go
import (
    "context"                              // 1. Standard library

    "github.com/third-party/pkg"           // 2. Third-party

    "github.com/YourOrg/mylib/model"       // 3. Your org
)
```

### Rules

- No `fmt.Print*` / `log.Print*` in production code.
- Line length <= 120.
- `// nolint:lintername` -- always name the linter.
- Tests exempt from: `errcheck`, `lll`, `mnd`, `wsl`, `dupl`.

---

## 10. Makefile

```makefile
.PHONY: test test-integration lint coverage bench bench-baseline bench-compare ci

test:
    go test -race -count=1 ./...

test-integration:
    go test -tags=integration -race -count=1 -v ./...

lint:
    golangci-lint run ./...

coverage:
    go test -coverprofile=coverage.out -cover -race ./...
    go-test-coverage --config=./.testcoverage.yml

bench:
    go test -bench=. -benchmem -count=6 -run="^$$" ./... | tee benchmarks/current.txt

bench-baseline:
    go test -bench=. -benchmem -count=6 -run="^$$" ./... | tee benchmarks/baseline.txt

bench-compare:
    benchstat benchmarks/baseline.txt benchmarks/current.txt

ci: lint coverage
```

---

## 11. CI Pipeline

```yaml
jobs:
  lint:
    steps: [checkout, setup-go, golangci-lint run ./...]

  test:
    steps:
      - checkout
      - setup-go
      - go test -coverprofile=coverage.out -cover -race ./...
      - go-test-coverage --config=./.testcoverage.yml
      - go test -bench=. -run="^$$" ./...
```

---

## 12. .gitignore

```gitignore
*.exe
*.dll
*.so
*.dylib
*.test
*.out
coverage.*
go.work
go.work.sum
.idea/
.vscode/
.env
benchmarks/*.txt
!benchmarks/baseline.txt
```

---

## 13. Metadata Files

| File | Purpose |
|---|---|
| `README.md` | `go get` install, 3-line quick start, config reference |
| `AGENTS.md` | Architecture, commands, test conventions, code rules (for AI/LLM agents) |
| `CODEOWNERS` | Review ownership |
| `LICENSE` | Apache 2.0, MIT, etc. |
| `dependabot.yml` | Weekly `gomod` updates |

---

## 14. Principles

| Principle | What it means for a vendor library |
|---|---|
| **No `main()`** | This is imported, never executed directly |
| **Root package = public API** | Consumer imports one path and gets the service, config, errors |
| **One file, one concern** | service / config / error / engine / utils -- never mixed |
| **Model is inert** | `model/` has zero imports. It's the shared vocabulary. |
| **Interface at the consumer boundary** | `repositories/storage.go` defines the contract. Implementations are pluggable. |
| **Memory impl ships with the lib** | Consumers use it in their own tests. No Docker needed. |
| **No infrastructure in domain** | Root package never imports DB drivers, frameworks, or infra SDKs |
| **Two config paths** | Env vars for production, struct literals for tests |
| **Errors are exported sentinels** | Consumer checks `errors.Is(err, mylib.ErrX)` |
| **Table-driven tests only** | `cases` + `t.Run`, no exceptions |
| **Integration behind build tags** | CI runs only unit tests. Integration needs explicit opt-in. |
| **Strict linting as CI gate** | No print statements, enforced imports, magic numbers flagged |
| **Benchmarks are tracked** | Baseline committed, regressions caught with `benchstat` |
