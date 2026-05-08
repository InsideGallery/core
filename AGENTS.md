# Core Library Guide for AI Agents

This file provides guidance to AI coding assistants (Claude Code, Copilot, Cursor, etc.) when working with this repository.

## 1. Project Overview

`core` is a shared Go vendor library for InsideGallery projects. It provides reusable infrastructure packages imported via `go get github.com/InsideGallery/core`. There is no `main()` -- this is a library, not an application.

- **Module**: `github.com/InsideGallery/core`
- **Go version**: 1.24.0
- **Type**: Vendor library (`go get`-able, no `cmd/`, no `main.go`)

## 2. Directory Structure

| Path | Description |
|------|-------------|
| `antibot/` | Anti-bot utilities (proof of work) |
| `app/` | Application bootstrap helpers (web server, NATS) |
| `commands/` | Command handler pattern |
| `dataconv/` | Data conversion utilities (binary, IP, merge) |
| `db/` | Database connectors: aerospike, bunt, elasticsearch, gremlin, mongodb, neo4j, postgres, redis |
| `ecs/` | Entity-Component-System framework |
| `embedded/` | Embedded resources |
| `errors/` | Error utilities |
| `fastlog/` | Structured logging: pluggable slog handlers (stderr, stdout, datadog, logfile, logstash, otel, nop), metrics |
| `fixtures/` | Test fixtures |
| `machielearning/` | Machine learning (neural networks via Gorgonia) |
| `mathutils/` | Math utilities |
| `memory/` | In-memory data structures (sorted sets, fuzzy search, etc.) |
| `multiproc/` | Multi-process coordination |
| `oslistener/` | OS signal listener |
| `pki/` | Public key infrastructure, encryption (AES, JWT, etc.) |
| `queue/` | Message queue clients (NATS) |
| `server/` | Server utilities (JWT, HTTP helpers) |
| `testutils/` | Shared test utilities |
| `ticker/` | Periodic task runner |
| `utils/` | General-purpose utilities (strings, phone, etc.) |

## 3. Mandatory Post-Change Verification

**After EVERY code change, you MUST run both tests and linter before considering the work done.**

```bash
# Run tests (mandatory after every change)
go test ./...

# Run tests with race detection
go test -race -count=1 ./...

# Run linter (mandatory after every change)
golangci-lint run ./...

# Auto-fix formatting issues
golangci-lint run --fix ./...
```

Do NOT skip these steps. Do NOT consider a task complete until both tests pass and the linter reports no errors.

## 4. Engineering Principles

### Core Design Principles

- **KISS (Keep It Simple and Smart)**: Systems work best when kept simple. Avoid unnecessary complexity.
- **DRY (Don't Repeat Yourself)**: Every piece of knowledge has one authoritative representation.
- **Performance by Design**: All design must prioritize performance from the start.
- **Clean Code**: Code should be easy to read, easy to change, and do what the reader expects. The reading-to-writing ratio is 10:1 -- optimize for readability.
- **Boy Scout Rule**: Leave the code cleaner than you found it. Every commit should improve surrounding code.
- **"Later equals never"** (LeBlanc's Law): Clean now or it stays dirty.

### Kent Beck's Four Rules of Simple Design (priority order)

1. Runs all the tests
2. Contains no duplication (DRY)
3. Expresses the intent of the programmer
4. Minimizes the number of entities (packages, types, functions)

### Architecture Principles

- **No `main()`**: This is imported, never executed directly.
- **Root packages = public API**: Consumer imports the package path and gets the service, config, errors.
- **One file, one concern**: service / config / error / engine / utils -- never mixed.
- **Interface at the consumer boundary**: Interfaces define contracts. Implementations are pluggable.
- **No infrastructure leakage**: Root packages and models must never import web frameworks or infra SDKs directly.
- **Dependency direction**: High-level modules must not depend on low-level modules. Both depend on abstractions.
- **Ports at consumer side**: Interfaces are defined where they are USED, not where they are IMPLEMENTED. Go's implicit interface satisfaction.
- **Command-Query Separation**: A function either changes state or returns a value, not both.
- **Wrap third-party dependencies**: Don't leak third-party types through your API. Wrap them behind your own interfaces.

## 5. Coding Standards

### Formatting & Linting

- **Formatter**: `gofumpt` (strict superset of gofmt).
- **Linter**: `golangci-lint v2` (config in `.golangci.yml`).
- **Import ordering** (enforced by gci): standard library, blank, dot, third-party, then `github.com/InsideGallery/core` packages.
- **Line length**: 120 characters max.
- **WSL (whitespace linter)** is enabled -- follow its blank-line conventions around blocks, declarations, and returns.
- **Forbidden**: `fmt.Print*`, `log.Print*`, bare `print*`. The linter will reject them.
- **Nolint directives** must specify the linter: `// nolint:gosec`.

### Naming

- **MixedCaps / mixedCaps**, never `snake_case` (except in test files for subtests).
- **Acronyms all-caps**: `HTTPClient`, `userID`, `xmlParser` (not `HttpClient`, `userId`).
- **No `Get` prefix on getters**: `user.Name()` not `user.GetName()`. Setters use `Set`: `user.SetName(n)`.
- **Full words for variables**: `httpConfig` not `httpCfg`.
- **Package names**: short, lowercase, singular nouns. Never `util`, `common`, `helpers`, `misc`.
- **Avoid stuttering**: `http.Server` not `http.HTTPServer`.
- **Interface names**: single-method interfaces use the method name + `er` suffix (`Reader`, `Writer`). Multi-method interfaces describe the capability.
- **Doc comments start with name**: `// Order represents...` not `// Represents an order...`.

### Functions

- **Small**: Aim for 5-20 lines. Over 40 lines signals the function does too much.
- **Few arguments**: 0-2 ideal. 3 maximum. More than 3 -- group into an options struct.
- **No flag arguments**: A `bool` parameter means the function does two things. Split into two functions or use an options struct.
- **Early return, no else**: Error cases return early; happy path stays left-aligned.
- **One level of abstraction per function**: Don't mix high-level orchestration with low-level byte manipulation.

### Error Handling

- **Never ignore errors**: Always check and return errors. `_ = fn()` is forbidden.
- **Error strings**: Lowercase, no punctuation -- `"open file"` not `"Open file."`.
- **Wrap with context**: `fmt.Errorf("context: %w", err)` -- every layer adds context.
- **Use `errors.Is` / `errors.As`** for checking, not `==`.
- **Sentinel errors at package boundaries**: `var ErrNotFound = errors.New("not found")`.
- **Critical errors**: Return to caller.
- **Non-critical errors**: Log with `slog.Warn` and structured fields, continue execution.
- **Don't panic in library code**: Return errors. Let the caller decide.

### Strings & Performance

- **Avoid `fmt.Sprintf` for string building** in hot paths -- uses reflect, slow. Prefer `+` concatenation or `strings.Join`. `fmt.Sprintf` is fine for one-time calls (errors, logs).
- **No commented-out code**: Delete it. Git remembers.
- **No `init()` I/O**: `init()` should only register things (handlers, drivers). Never read files or make network calls.
- **Zero value should be useful**: Design types so the zero value is valid.

### Concurrency

- **"Don't communicate by sharing memory; share memory by communicating."**
- **Separate concurrency code from business logic**: Business logic in pure functions/methods. Goroutine orchestration in a thin coordination layer.
- **Always ensure goroutines terminate**: Leaked goroutines = leaked memory. Use `context.Context` for cancellation.
- **`go test -race`**: Must pass. Non-negotiable.
- **`context.Context`**: First parameter of every function that may block or be cancelled.
- **Graceful shutdown**: Handle `SIGTERM`, drain in-flight work, close connections cleanly.

## 6. Testing Conventions

- **Table-driven tests only**: Define a `cases` (or `testcases`) slice of structs, iterate with `t.Run`.
- **F.I.R.S.T.**: Fast, Independent, Repeatable, Self-validating, Timely.
- **Test boundary conditions**: Off-by-one, empty collections, nil inputs, max values, timeouts.
- **One concept per test**: Each test function tests one behavioral scenario.
- **`t.Helper()`**: Mark helper functions so stack traces point to the failing test.
- **`t.Parallel()`**: Run independent tests in parallel for speed.
- **Integration tests**: Build-tag gated (`//go:build integration`). CI runs only unit tests.
- **Never ignore errors in tests**: Use `t.Fatalf` on setup errors.
- Tests are exempt from: `bodyclose`, `dogsled`, `dupl`, `errcheck`, `lll`, `mnd`, `wsl`.

## 7. Clean Code Smells Quick Reference

| ID | Smell | Rule |
|----|-------|------|
| C5 | Commented-out code | Delete. Git remembers. |
| E1 | Build requires more than one step | `go build ./...`. One command. |
| E2 | Tests require more than one step | `go test ./...`. One command. |
| F1 | Too many arguments | 0-2 ideal. 3 max. Use options struct. |
| F3 | Flag arguments | Split into two functions. |
| F4 | Dead function | Delete unreferenced functions. |
| G2 | Obvious behavior unimplemented | Principle of Least Surprise. |
| G4 | Overridden safeties | Never disable linters or skip tests without justification. |
| G5 | Duplication | Extract. Use interfaces for repeated switch/if patterns. |
| G9 | Dead code | Delete unreachable code. `staticcheck` catches this. |
| G11 | Inconsistency | Same pattern everywhere. |
| G16 | Obscured intent | No magic numbers, no overly clever one-liners. |
| G25 | Replace magic numbers with named constants | `const maxRetries = 3`, not bare `3`. |
| G28 | Encapsulate conditionals | `if shouldRetry(err)` not complex boolean expressions. |
| G29 | Avoid negative conditionals | `if isValid` not `if !isInvalid`. |
| G34 | One level of abstraction per function | Don't mix high and low in the same function body. |
| G36 | Law of Demeter | No `a.GetB().GetC().Do()`. Talk to direct collaborators only. |

## 8. Twelve-Factor Compliance

These apply when this library is used in applications:

| Factor | Rule |
|--------|------|
| II. Dependencies | All deps in `go.mod`. No implicit reliance on system tools. |
| III. Config | Config via environment variables. Two paths: env vars (production) and struct literals (tests). |
| IV. Backing Services | All service connections configured via env. Use interfaces for swappable implementations. |
| VI. Processes | Stateless, share-nothing. No local filesystem writes for persistent state. |
| IX. Disposability | Fast startup, graceful shutdown. Handle `SIGTERM`. Idempotent operations. |
| XI. Logs | Log to stdout/stderr only. No log files, no log rotation in the app. Structured logging (JSON). |

## 9. CTO Decision Framework

- **Standardize by default**: Use existing patterns and technologies. Research new ones only when the improvement is **at least 10x** on one dimension without degrading others.
- **Trust, but verify**: Blind trust is not a management technique. Review data, talk to people doing the work, investigate inconsistencies immediately.
- **Architecture changes require discussion**: Any changes to architecture require documentation via an Architecture Communication Canvas before implementation.
- **Over-engineering is prohibited**: Don't introduce unnecessary complexity. KISS > DRY > abstraction.

## 10. Rules

1. **Use `#AI-assisted` in commit messages** for AI-assisted code.
2. **Run tests and linter after every change** -- `go test ./...` and `golangci-lint run ./...`.
3. **Never ignore errors** -- every `error` return must be checked.
4. **No `fmt.Print*` / `log.Print*`** in production code -- use `log/slog`.
5. **Import order enforced by gci**: standard -> blank -> dot -> default -> `github.com/InsideGallery/core` prefix.
6. **Line length <= 120 characters**.
7. **Table-driven tests only** -- `cases` + `t.Run`, no exceptions.
8. **Interfaces at the consumer, not the provider** -- Go implicit satisfaction.
9. **Wrap third-party dependencies** -- don't leak external types through your API.
10. **No commented-out code** -- delete it, Git remembers.
11. **Test coverage must not decrease** in any merge request.
12. **Consult source documents** before making architectural or API design decisions.
13. **Use the logging and bundle workflow** -- application bootstraps install logging through `app.WebMain`
    or `app.NATSMain`; downstream code emits through `slog.Default()`. New logging and metrics
    backends must register through their package and be exposed through `fastlog/all` or `metrics/all` for
    blank imports, preserving Twelve-Factor
    [Factor IV](docs/source/Twelve-Factor%20App.md#iv-backing-services) and
    [Factor XI](docs/source/Twelve-Factor%20App.md#xi-logs).
