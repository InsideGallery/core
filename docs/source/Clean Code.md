# Clean Code in Go

**Source**: Robert C. Martin, "Clean Code: A Handbook of Agile Software Craftsmanship" (2008)
**Adaptation**: Go-specific practices, aligned with existing Engineering Principles (POL-ENG-001), Go Server/Library/Client architecture instructions, and Twelve-Factor App methodology.

---

## Philosophy

Clean code is code that is easy to read, easy to change, and does what the reader expects. It looks like it was written by someone who cares.

**Core beliefs:**
- The reading-to-writing ratio is 10:1. Optimizing for readability IS optimizing for productivity.
- "Later equals never" (LeBlanc's Law). Clean now or it stays dirty.
- The only way to go fast is to keep code clean at all times.
- Getting code to work and making code clean are two different activities. Do both.

**The Boy Scout Rule**: Leave the code cleaner than you found it. Every commit should improve the surrounding code a little -- rename a variable, extract a function, simplify a conditional.

**Kent Beck's Four Rules of Simple Design** (priority order):
1. Runs all the tests
2. Contains no duplication (DRY)
3. Expresses the intent of the programmer
4. Minimizes the number of entities (packages, types, functions)

---

## 1. Naming

> Names are 90% of what makes software readable. Spend time choosing them.

### Rules

| Rule | Go Adaptation |
|------|---------------|
| **Use intention-revealing names** | If a name needs a comment, it doesn't reveal its intent. `elapsedDays` not `d`. |
| **Avoid disinformation** | Don't call it `accountList` unless it's a `[]Account`. Prefer `accounts`. |
| **Make meaningful distinctions** | If names must differ, they must mean different things. No `data` vs `info`, no `a1`/`a2`. |
| **Use pronounceable names** | `generationTimestamp` not `genymdhms`. Code is discussed aloud. |
| **Use searchable names** | Name length should match scope size. Single-letter names only in tiny scopes (`for i, v := range items`). |
| **Pick one word per concept** | Don't mix `Get`/`Fetch`/`Retrieve` for the same abstract operation across the codebase. |
| **Don't pun** | If `Add` means "combine two values" in one package, don't reuse `Add` to mean "append to collection" in another. Use `Append` or `Insert`. |
| **Use solution domain names** | `JobQueue`, `EventBus`, `Repository` -- programmers know these. |
| **Use problem domain names** | When no technical term exists, use the business term so maintainers can ask domain experts. |

### Go-Specific Conventions

- **Exported = Capitalized**. This is Go's access control. Choose exported names as carefully as a public API.
- **MixedCaps / mixedCaps**, never `snake_case` (except in test files for `Test_` subtests and in generated code).
- **Short names for short scopes**. `r` for a `*http.Request` in a 5-line handler is fine. `request` in a 50-line function is better.
- **Acronyms are all-caps**: `HTTPClient`, `userID`, `xmlParser`. Not `HttpClient`, `userId`.
- **Interface names**: single-method interfaces use the method name + `er` suffix (`Reader`, `Writer`, `Closer`, `Stringer`). Multi-method interfaces describe the capability (`ReadWriter`, `Storage`).
- **No `Get` prefix on getters**: `user.Name()` not `user.GetName()`. Setters use `Set`: `user.SetName(n)`.
- **Package names**: short, lowercase, singular nouns. The package name is part of the call site (`http.Get`, not `httputil.HTTPGet`). Never `util`, `common`, `helpers`, `misc` -- these are meaningless. See **Go Server.md** and **Go Library.md** for package structure conventions.
- **Avoid stuttering**: `http.HTTPServer` stutters. `http.Server` is clean.

### Anti-Patterns

- `Manager`, `Processor`, `Handler`, `Data`, `Info` as type names -- these are signs of unclear responsibility (unless `Handler` is an HTTP handler, which is an established Go convention).
- Hungarian notation, member prefixes (`m_`, `f_`), type encodings -- never in Go.
- `IReader` for interfaces -- Go doesn't prefix interfaces. The concrete type gets the descriptive name; the interface stays clean.

---

## 2. Functions

> The first rule: small. The second rule: smaller than that.

### Rules

| Rule | Go Adaptation |
|------|---------------|
| **Small** | Aim for 5-20 lines. If a function is over 40 lines, it likely does too much. |
| **Do one thing** | If you can extract a sub-function with a name that isn't a restatement of the parent, the parent does more than one thing. |
| **One level of abstraction per function** | Don't mix `service.ProcessOrder()` (high) with `buf.WriteString("\n")` (low) in the same function. |
| **The Stepdown Rule** | Read top-to-bottom. High-level functions call mid-level, which call low-level. Matches Go Server.md's vertical ordering convention. |
| **Few arguments** | 0-2 ideal. 3 maximum. More than 3 → group into a struct (Options pattern). |
| **No flag arguments** | A `bool` parameter means the function does two things. Split into two functions. |
| **No side effects** | A function named `CheckPassword` must not also initialize a session. Name must match behavior exactly (N7). |
| **Command-Query Separation** | A function either changes state or returns a value, not both. |
| **DRY** | Every piece of knowledge has one authoritative representation. See Engineering Principles POL-ENG-001. |

### Go-Specific Practices

- **Error returns replace exceptions**. Go's `(result, error)` pattern naturally separates happy path from error handling. Wrap errors with `fmt.Errorf("operation: %w", err)` for context.
- **Options pattern** for 3+ configuration values:
  ```go
  type Options struct {
      Timeout  time.Duration
      RetryMax int
      Logger   *slog.Logger
  }
  func NewService(repo Repository, opts Options) *Service
  ```
- **Functional options** (`With...` functions) for libraries where zero-value defaults matter (see **Go Library.md**).
- **Named return values** only for documentation in godoc, not for naked returns (naked returns obscure intent).
- **`defer` for cleanup** -- replaces try/finally. Keep defer close to the resource acquisition.

### Error Handling (Chapter 7 adapted)

Go doesn't have exceptions. It has error values. This is cleaner -- but requires discipline:

| Rule | Go Practice |
|------|------------|
| **Don't ignore errors** | Always check `err`. Use `errcheck` linter. |
| **Wrap with context** | `fmt.Errorf("save user %d: %w", id, err)` -- every layer adds context. |
| **Don't return nil error with nil result** | Return a zero value or a special-case object. |
| **Sentinel errors sparingly** | Define `var ErrNotFound = errors.New("not found")` only at package boundaries. Prefer error wrapping. |
| **Error handling is one thing** | If a block handles an error, it should do nothing else. Extract the happy path. |
| **Don't panic** | `panic` is not for error handling. Reserve for truly unrecoverable states (init failures, programmer bugs). |
| **Special Case pattern** | Return a "null object" instead of nil. `EmptyResult{}` that satisfies the interface with no-op behavior. |

```go
// BAD: error handling mixed with logic
func Process(id int) error {
    user, err := repo.Get(id)
    if err != nil {
        log.Error("failed", "err", err)
        return fmt.Errorf("process: %w", err)
    }
    if user.IsActive() {
        // 40 lines of business logic...
    }
    return nil
}

// GOOD: separated concerns
func Process(id int) error {
    user, err := repo.Get(id)
    if err != nil {
        return fmt.Errorf("process user %d: %w", id, err)
    }
    return processActiveUser(user)
}
```

---

## 3. Comments

> Comments are a failure to express intent in code. Prefer self-documenting code.

### Good Comments (acceptable in Go)

| Type | Example |
|------|---------|
| **Godoc on exported symbols** | Required. First sentence is the summary. `// Server handles incoming HTTP requests.` |
| **Explanation of intent** | Why a non-obvious decision was made. `// We use a pool of 10 because the upstream API rate-limits at 10 concurrent requests.` |
| **Warning of consequences** | `// This is not thread-safe. Caller must hold mu.` |
| **TODO** | `// TODO(username): remove after migration completes (DAF-456)` -- with a tracking reference. |
| **Legal headers** | Copyright notice at file top (keep brief). |

### Bad Comments (never in Go)

| Type | Why |
|------|-----|
| **Redundant** | `// GetUser returns the user` -- the name already says this. |
| **Journal / changelog** | Git history does this. |
| **Commented-out code** | Delete it. Git remembers. |
| **Closing brace comments** | `} // end if` means your function is too long. |
| **Attribution** | `// Added by John` -- `git blame` does this. |
| **Noise** | Comments that restate the obvious. |

### Go Rule

`gofmt` and `golint`/`staticcheck` enforce that exported symbols have comments. But **don't write comments just to satisfy the linter**. If the comment adds nothing, improve the name instead.

---

## 4. Formatting

> `gofmt` is not optional. It IS the formatting standard.

### Vertical

| Rule | Practice |
|------|---------|
| **Keep files small** | 200-500 lines. If a file exceeds 500 lines, look for a package split. |
| **Newspaper metaphor** | Package-level vars and types at top. Exported functions next. Private functions below. Low-level helpers at bottom. Matches **Go Server.md** stepdown convention. |
| **Blank lines separate concepts** | One blank line between functions. One blank line between logical sections inside a function. No blank lines between tightly related declarations. |
| **Vertical proximity** | Variables declared near first use. Caller above callee. |
| **Dependent functions close** | If `handleRequest` calls `validateInput`, keep them adjacent. |

### Horizontal

| Rule | Practice |
|------|---------|
| **`gofmt` handles it** | Tabs for indentation, spaces for alignment. Automated. |
| **Line length** | No hard limit in Go, but stay under 120 characters for readability. |
| **No horizontal alignment** | Don't align struct field names or assignments in columns -- `gofmt` doesn't enforce it and it creates noisy diffs. |

---

## 5. Structs, Interfaces, and Packages

Go has no classes. The equivalent concepts are **structs** (data + methods), **interfaces** (behavior contracts), and **packages** (modules/namespaces).

### Struct Design

| Rule | Practice |
|------|---------|
| **Small** | A struct should have one responsibility. If the name needs `And` or a weasel word (`Manager`, `Processor`), it's too big. |
| **Single Responsibility Principle** | One reason to change. If a struct handles both HTTP routing and database queries, split it. |
| **Cohesion** | Every field should be used by most methods. Fields used by only a subset of methods signal a struct that should be split. |
| **Exported fields = data structure (DTO)** | If fields are exported, it's a data transfer object. Don't add business logic methods to it. See **Go Server.md** `model/` convention. |
| **Unexported fields + methods = object** | Hides data, exposes behavior through methods. Business logic lives here. |
| **Don't create hybrids** | A struct with both exported fields AND business methods is the worst of both worlds. |

### Interface Design

| Rule | Practice |
|------|---------|
| **Small interfaces** | Prefer 1-2 method interfaces. `io.Reader`, `io.Writer` are the gold standard. |
| **Accept interfaces, return structs** | Function parameters should be interfaces; return values should be concrete types. |
| **Define interfaces at the consumer, not the provider** | Go interfaces are satisfied implicitly. The package that USES the behavior defines the interface it needs. See **Go Server.md** ports convention. |
| **Law of Demeter** | A method should only call methods on: its receiver, its arguments, objects it creates, its receiver's fields. No train wrecks (`a.GetB().GetC().Do()`). |

### Package Design

| Rule | Practice |
|------|---------|
| **One purpose per package** | A package is the Go equivalent of a "class" in terms of responsibility scope. |
| **Internal packages for implementation details** | Use `internal/` to hide packages that external consumers should not depend on. |
| **Wrap third-party dependencies** | Don't leak third-party types through your API. Wrap them behind your own interfaces (Boundary pattern). See **Go Library.md** `repositories/storage.go` pattern. |
| **Adapter pattern for undefined APIs** | When a dependency's API doesn't exist yet, define the interface you wish you had and write an adapter later. |

---

## 6. Testing

> Test code is just as important as production code. If tests rot, code rots.

### Rules

| Rule | Practice |
|------|---------|
| **F.I.R.S.T.** | **F**ast, **I**ndependent, **R**epeatable, **S**elf-validating, **T**imely. |
| **One concept per test** | Each test function tests one behavioral scenario. |
| **Build-Operate-Check** | Three phases: setup, action, assertion. In Go: arrange/act/assert. |
| **Table-driven tests** | Go's idiomatic way to test multiple scenarios without duplication. |
| **Test boundary conditions** | Off-by-one, empty collections, nil inputs, max values, timeouts. |
| **Exhaustively test near bugs** | Bugs cluster. When you find one, test the surrounding function thoroughly. |
| **Don't skip trivial tests** | Their documentation value exceeds their cost. |
| **Tests should be fast** | A slow test is a test that won't get run. Use build-tag gated integration tests (see **Go Library.md**). |

### Go-Specific Practices

- **`testing` package**: Use stdlib. No test framework required for most cases.
- **`testify/assert`**: Acceptable for assertion readability in larger projects.
- **`t.Helper()`**: Mark helper functions so stack traces point to the failing test, not the helper.
- **`t.Parallel()`**: Run independent tests in parallel for speed.
- **Test file naming**: `service_test.go` next to `service.go`. Same package for white-box tests, `_test` package for black-box tests.
- **Integration tests**: Build-tag gated (`//go:build integration`). Separated from unit tests per Engineering Principles (POL-ENG-001).
- **Coverage**: `go test -cover`. Use as a heuristic, not a target. Uncovered code is a question, not a verdict.

```go
// Table-driven test (Go idiom for Build-Operate-Check)
func TestCalculateScore(t *testing.T) {
    tests := []struct {
        name     string
        input    Event
        expected float64
    }{
        {"normal user", Event{Amount: 100}, 0.5},
        {"high value", Event{Amount: 10000}, 0.95},
        {"zero amount", Event{Amount: 0}, 0.0},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := CalculateScore(tt.input)
            assert.InDelta(t, tt.expected, got, 0.01)
        })
    }
}
```

---

## 7. Concurrency

> Go's concurrency model is fundamentally different from Java's. The principles remain; the mechanisms change.

### Core Principle

**"Don't communicate by sharing memory; share memory by communicating."** -- Go Proverb

### Rules (adapted from Clean Code Ch. 13)

| Rule | Go Practice |
|------|------------|
| **Separate concurrency code from business logic** | Business logic in pure functions/methods. Goroutine orchestration in a thin coordination layer. |
| **Limit shared mutable state** | Prefer channels over mutexes. When mutex is needed, keep critical sections minimal. |
| **Use copies of data** | Send copies through channels rather than sharing pointers. |
| **Goroutines should be as independent as possible** | Each goroutine owns its data. No shared state unless explicitly coordinated. |
| **Know your patterns** | Fan-out/fan-in, pipeline, worker pool, pub/sub. These replace Producer-Consumer, Readers-Writers, Dining Philosophers. |
| **Fast startup, graceful shutdown** | Handle `context.Done()` and OS signals. See **Twelve-Factor App** Factor IX (Disposability). |
| **Don't ignore spurious failures** | A race condition that happens once in a million runs is still a bug. Use `-race` flag. |
| **Get non-concurrent code working first** | Test business logic without goroutines. Then add concurrency. |

### Go-Specific Tools

- **`go test -race`**: Run always in CI. Non-negotiable.
- **`context.Context`**: First parameter of every function that may block or be cancelled.
- **`sync.WaitGroup`**: For coordinating goroutine completion.
- **`sync.Once`**: For safe lazy initialization.
- **`sync.Mutex` / `sync.RWMutex`**: When channels aren't appropriate. Keep lock scope minimal.
- **`errgroup`** (`golang.org/x/sync/errgroup`): For parallel operations with error propagation.

---

## 8. Boundaries

> Depend on something you control, not something that controls you.

### Rules

| Rule | Go Practice |
|------|------------|
| **Wrap third-party APIs** | Define your own interface. Write an adapter. Third-party types never appear in your domain. See **Go Library.md** `repositories/storage.go` pattern. |
| **Write learning tests** | When adopting a new library, write tests that verify your assumptions about its behavior. Run them on library upgrades. |
| **Don't pass boundary types around** | A `*sql.DB` should not leak past the repository layer. Wrap it. |
| **Adapter pattern for future APIs** | Define the interface you need now. Write the adapter when the real API arrives. |
| **Keep boundary references minimal** | As few import points as possible for external dependencies. |

---

## 9. Systems

> Separate construction from use. See **Twelve-Factor App** Factor V (Build, Release, Run).

### Rules

| Rule | Go Practice |
|------|------------|
| **Separation of `main`** | `cmd/service/main.go` wires everything together. Business logic packages have zero knowledge of `main`. Matches **Go Server.md** `cmd/` convention. |
| **Constructor injection** | Pass dependencies as arguments to `New` functions. No global state, no service locators. |
| **No framework coupling** | Business logic is pure Go. No ORM types in domain. See **Go Server.md** MDCA principle: domain depends on nothing. |
| **Start simple, evolve** | Don't over-architect. Add complexity only when justified by real requirements. See Engineering Principles: KISS > DRY > MDCA. |
| **Cross-cutting concerns via middleware** | HTTP middleware for logging, auth, metrics. NATS middleware for message tracing. Not scattered across business logic. |

---

## 10. Smells and Heuristics (Quick Reference)

Adapted from Clean Code Chapter 17. Java-specific items (J1-J3) replaced with Go equivalents. IDs preserved for cross-reference.

### Comments
| ID | Smell | Rule |
|----|-------|------|
| C1 | Inappropriate information | No metadata in comments (authors, dates, changelogs). Use git. |
| C2 | Obsolete comment | Update or delete immediately. |
| C3 | Redundant comment | Don't restate what the code says. |
| C4 | Poorly written comment | If worth writing, write it well. |
| C5 | Commented-out code | Delete. Git remembers. |

### Environment
| ID | Smell | Rule |
|----|-------|------|
| E1 | Build requires more than one step | `go build ./...` or `make build`. One command. |
| E2 | Tests require more than one step | `go test ./...`. One command. |

### Functions
| ID | Smell | Rule |
|----|-------|------|
| F1 | Too many arguments | 0-2 ideal. 3 max. Use options struct. |
| F2 | Output arguments | Return values instead. Go's multiple returns make this natural. |
| F3 | Flag arguments | Split into two functions. |
| F4 | Dead function | Delete unreferenced functions. |

### General
| ID | Smell | Rule |
|----|-------|------|
| G1 | Multiple languages in one file | One language per file. Minimize `//go:generate` inline templates. |
| G2 | Obvious behavior unimplemented | Principle of Least Surprise. |
| G3 | Incorrect behavior at boundaries | Test every boundary condition. |
| G4 | Overridden safeties | Never disable linters, skip tests, or bypass `//nolint` without justification. |
| G5 | Duplication | Extract. Use polymorphism (interfaces) for repeated switch/if patterns. |
| G6 | Code at wrong level of abstraction | High-level in interfaces/exported functions, low-level in unexported helpers. |
| G7 | Base depending on derivative | In Go: an interface package must not import its implementations. |
| G8 | Too much information | Small interfaces, few exported symbols, minimal public API surface. |
| G9 | Dead code | Delete unreachable code. `staticcheck` catches this. |
| G10 | Vertical separation | Declare variables near use. Private functions below their first caller. |
| G11 | Inconsistency | Same pattern everywhere. If one repo uses `New`, all repos use `New`. |
| G12 | Clutter | Remove unused vars, empty functions, meaningless comments. |
| G13 | Artificial coupling | Don't put constants in the wrong package. |
| G14 | Feature envy | If a function uses another struct's fields more than its own, it belongs on that struct. |
| G15 | Selector arguments | No bool/enum/int to select behavior. Split into multiple functions. |
| G16 | Obscured intent | No magic numbers, no overly clever one-liners. |
| G17 | Misplaced responsibility | Code goes where readers expect it. |
| G18 | Inappropriate static | In Go: prefer methods on types over package-level functions when behavior might vary. |
| G19 | Use explanatory variables | Break complex expressions into named intermediates. |
| G20 | Function names say what they do | If you must read the body, rename. |
| G21 | Understand the algorithm | Don't stop at "tests pass." Understand why it works. |
| G22 | Make logical dependencies physical | If module A depends on a value from B, A should explicitly receive it from B. |
| G23 | Prefer polymorphism to if/else chains | Use interfaces and strategy pattern instead of switching on type. |
| G24 | Follow standard conventions | Use `gofmt`, follow Go conventions, be consistent. |
| G25 | Replace magic numbers with named constants | `const maxRetries = 3`, not bare `3`. |
| G26 | Be precise | Check errors, validate assumptions, handle nil, use correct types. |
| G27 | Structure over convention | Interfaces enforce contracts. Conventions can be ignored. |
| G28 | Encapsulate conditionals | `if shouldRetry(err)` not `if err != nil && !errors.Is(err, ErrFatal) && retries < max`. |
| G29 | Avoid negative conditionals | `if isValid` not `if !isInvalid`. |
| G30 | Functions do one thing | If a function has sections, it does multiple things. |
| G31 | Hidden temporal couplings | Make ordering explicit via return values (bucket brigade). |
| G32 | Don't be arbitrary | Have a reason for every structural decision. Communicate it. |
| G33 | Encapsulate boundary conditions | `nextLevel := level + 1`, not scattered `+1`s. |
| G34 | One level of abstraction per function | Don't mix high and low in the same function body. |
| G35 | Configurable data at high levels | Config values in `main` or config structs, not buried in low-level code. See **Twelve-Factor App** Factor III. |
| G36 | Avoid transitive navigation (Law of Demeter) | No `a.GetB().GetC().Do()`. Talk to direct collaborators only. |

### Go-Specific (replacing Java J1-J3)
| ID | Smell | Rule |
|----|-------|------|
| Go1 | Dot imports | Never use `import . "pkg"` outside tests. It obscures where symbols come from. |
| Go2 | Init functions with side effects | `init()` should only register things. Never do I/O or mutate global state in `init()`. |
| Go3 | Bare string/int constants | Use typed constants or `iota` enums for categorization, not raw strings or ints. |

### Tests
| ID | Smell | Rule |
|----|-------|------|
| T1 | Insufficient tests | Test everything that could break. |
| T2 | Use a coverage tool | `go test -cover`. |
| T3 | Don't skip trivial tests | Documentation value exceeds cost. |
| T4 | Ignored test = question about ambiguity | `t.Skip("unclear requirement: ...")`. |
| T5 | Test boundary conditions | Off-by-one, nil, empty, max. |
| T6 | Exhaustively test near bugs | Bugs cluster. |
| T7 | Patterns of failure are revealing | Order tests to expose patterns. |
| T8 | Coverage patterns are revealing | Uncovered code reveals why tests fail. |
| T9 | Tests should be fast | Slow tests don't get run. |
