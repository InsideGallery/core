# Go Best Practice

**Source**: [Effective Go](https://go.dev/doc/effective_go), Go standard library conventions, [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
**Purpose**: Comprehensive Go idioms and conventions. Complements Clean Code.md (design principles) and Go Server/Library/Client.md (project structure). This document covers HOW to write Go code; those cover HOW to organize it.

---

## 1. Formatting

There is exactly one rule: **use `gofmt`**. No exceptions, no arguments.

- Indentation: tabs (not spaces).
- No line length limit. If a line feels too long, wrap and indent with an extra tab.
- No parentheses around `if`, `for`, `switch` conditions.
- `gofmt` handles vertical alignment of struct fields automatically.

---

## 2. Naming

> Names have semantic effect in Go. Exported = capitalized. Unexported = lowercase.

### Packages

| Rule | Example |
|------|---------|
| Lowercase, single-word, no underscores, no mixedCaps | `http`, `bufio`, `sync` |
| Short, concise, evocative | `fmt` not `format`, `os` not `operatingsystem` |
| Package name is part of the call site -- no stutter | `http.Server` not `http.HTTPServer` |
| Never `util`, `common`, `helpers`, `misc` | These names carry zero information |
| `import .` only in tests that must run outside the package | Everywhere else, explicit package prefix |

### Identifiers

| Rule | Example |
|------|---------|
| `MixedCaps` / `mixedCaps`, never `snake_case` | `readTimeout`, `MaxRetries` |
| Acronyms are all-caps | `HTTPClient`, `userID`, `xmlParser` |
| Length matches scope | `i` in 3-line loop, `retryCount` in 50-line function |
| Getters: `Owner()` not `GetOwner()` | `user.Name()` |
| Setters: `SetOwner()` | `user.SetName(n)` |
| Interfaces: method + `-er` | `Reader`, `Writer`, `Stringer`, `Closer` |
| Honor canonical names | `Read`, `Write`, `Close`, `String` have specific signatures -- use the same name only if the signature and meaning match |
| Bool functions/methods: `is`/`has`/`can` prefix or question form | `IsValid()`, `HasPermission()`, `CanRetry()` |

### Constants

| Rule | Example |
|------|---------|
| Use `iota` for enumerations | `const (Pending Status = iota; Active; Closed)` |
| Add `String()` method to iota types | Human-readable logging and debugging |
| Named constants for magic numbers | `const maxRetries = 3`, not bare `3` |

---

## 3. Control Structures

### If

```go
// GOOD: init statement + early return, no else
if err := file.Chmod(0664); err != nil {
    return fmt.Errorf("chmod: %w", err)
}
// success path continues here, unindented

// BAD: unnecessary else
if err != nil {
    return err
} else {
    process()
}
```

**Rule**: When the `if` body ends with `return`, `break`, `continue`, or `goto`, drop the `else`. Error cases return early; the happy path runs down the page.

### For

```go
// Three forms
for i := 0; i < n; i++ { }    // C-like
for condition { }               // while
for { }                         // infinite

// Range
for key, value := range m { }  // both
for key := range m { }         // key only
for _, value := range m { }    // value only
```

**Rule**: Range over strings iterates runes (Unicode code points), not bytes. Erroneous UTF-8 produces `U+FFFD`.

### Switch

```go
// No automatic fall-through (unlike C)
// Comma-separated cases
switch c {
case ' ', '\t', '\n':
    return true
}

// Expression-less switch = if-else chain
switch {
case x < 0:
    return -1
case x > 0:
    return 1
default:
    return 0
}
```

### Type Switch

```go
switch v := value.(type) {
case string:
    fmt.Println(v) // v is string here
case int:
    fmt.Println(v) // v is int here
default:
    fmt.Printf("unexpected type %T\n", v)
}
```

---

## 4. Functions

### Multiple Returns

Always return `(result, error)`. Never return only the error with the result as an output argument.

```go
// GOOD
func ReadConfig(path string) (Config, error)

// BAD: output argument
func ReadConfig(path string, cfg *Config) error
```

### Named Return Values

Use for documentation, not for naked returns in long functions.

```go
// GOOD: documents what n and err mean
func ReadFull(r io.Reader, buf []byte) (n int, err error)

// BAD: naked return in a 40-line function -- unclear what's returned
func process() (result string, err error) {
    // ... 40 lines ...
    return // what values?
}
```

### Defer

```go
f, err := os.Open(name)
if err != nil {
    return err
}
defer f.Close() // always right after successful open

mu.Lock()
defer mu.Unlock()
// ... critical section ...
```

**Rules**:
- Place `defer` immediately after the resource acquisition.
- Arguments are evaluated at `defer` time, not at execution time.
- Deferred calls execute in LIFO order.
- Prefer `defer` over manual cleanup -- never forget to close/unlock.

---

## 5. Data

### new vs make

| Function | Creates | Returns | Initializes |
|----------|---------|---------|-------------|
| `new(T)` | Any type | `*T` (pointer to zero value) | Zeros memory |
| `make(T)` | Slice, map, channel ONLY | `T` (not pointer) | Initializes internal structure |

```go
s := make([]int, 10, 100)  // slice: len=10, cap=100
m := make(map[string]int)  // map: empty, ready to use
ch := make(chan int, 5)     // channel: buffered, cap=5

p := new(bytes.Buffer)     // *bytes.Buffer, zero value = empty buffer, ready to use
```

**Rule**: Design types so the zero value is useful without constructors. `sync.Mutex{}` is unlocked. `bytes.Buffer{}` is an empty buffer. `[]int(nil)` is a valid empty slice.

### Slices

| Rule | Practice |
|------|---------|
| Prefer slices over arrays | Arrays are values (copying), slices are references |
| Always capture append result | `s = append(s, item)` -- the backing array may change |
| Append slice to slice | `s = append(s, other...)` with `...` |
| Nil slice is valid | `var s []int` -- `len(s)==0`, `append` works, no allocation until needed |
| Use `copy` for safe duplication | `dst := make([]T, len(src)); copy(dst, src)` |

### Maps

| Rule | Practice |
|------|---------|
| Check existence with comma-ok | `v, ok := m[key]` |
| Delete is safe on missing keys | `delete(m, key)` -- no error if absent |
| Nil map reads return zero values | But writing to a nil map panics -- always `make` or use literal |
| Maps as sets | `seen := map[string]bool{}; seen[x] = true` |

### Composite Literals

```go
// Struct with named fields (order doesn't matter, missing fields zero-valued)
return &File{fd: fd, name: name}

// Slice literal
primes := []int{2, 3, 5, 7, 11}

// Map literal
m := map[string]int{
    "one":   1,
    "two":   2,
    "three": 3,
}
```

---

## 6. Methods

### Pointer vs Value Receiver

| Receiver | When to Use |
|----------|-------------|
| **Pointer** `(t *T)` | Method modifies the receiver. Large struct (avoids copy). Consistency if any method uses pointer. |
| **Value** `(t T)` | Method is read-only. Small struct (int, small struct). |

**Rules**:
- Value methods can be called on both values and pointers.
- Pointer methods can only be called on pointers (but Go auto-inserts `&` for addressable values).
- If ANY method has a pointer receiver, ALL methods should use pointer receiver (consistency).
- Methods can be defined on any named type, not just structs.

---

## 7. Interfaces

### Design Rules

| Rule | Practice |
|------|---------|
| Small interfaces | 1-2 methods ideal. `io.Reader`, `io.Writer`, `error`. |
| Accept interfaces, return structs | Parameters = abstract, returns = concrete. |
| Define at the consumer | The package that USES the behavior defines the interface it needs. Not the provider. |
| Don't export the type if it only implements an interface | Export the interface, return it from constructors. Hide the concrete type. |
| Implicit satisfaction | No `implements` keyword. A type satisfies an interface by having the methods. |

### Type Assertions

```go
// Safe form (always use this)
s, ok := val.(string)
if !ok {
    // val is not a string
}

// Unsafe form (panics on failure -- avoid)
s := val.(string)
```

### Interface Conversions

```go
// Convert to access different method set
type Sequence []int

func (s Sequence) String() string {
    sort.IntSlice(s).Sort() // convert to sort.IntSlice to use Sort()
    return fmt.Sprint([]int(s)) // convert to []int to use default formatting
}
```

---

## 8. Embedding

| Rule | Practice |
|------|---------|
| Embed for composition, not inheritance | Go has no subclassing. Embed to "borrow" implementation. |
| Promoted methods | Embedded type's methods are available on the outer type directly. |
| Receiver is the inner type | When calling an embedded method, the receiver is the inner type, not outer. |
| Access embedded field by type name | `job.Logger.Printf(...)` for explicit access. |
| Shadowing resolves conflicts | Outer field/method hides inner field/method with the same name. |
| Don't embed to satisfy an interface unless the interface is needed | Embedding `io.Reader` in a struct only makes sense if the struct should be an `io.Reader`. |

```go
// Embedding in struct
type ReadWriter struct {
    *Reader  // has Read method
    *Writer  // has Write method
}
// ReadWriter now has both Read and Write methods

// Embedding in interface
type ReadWriter interface {
    Reader
    Writer
}
```

---

## 9. Concurrency

### Core Maxim

> "Do not communicate by sharing memory; share memory by communicating."

### Goroutines

| Rule | Practice |
|------|---------|
| Cheap to create | Stacks start small (~2KB), grow as needed. No thread-per-goroutine. |
| Launch with `go` | `go func() { ... }()` -- note the `()` to invoke. |
| Closures capture variables | Be careful with loop variables -- capture by parameter, not by closure. |
| Always ensure goroutines terminate | Leaked goroutines = leaked memory. Use `context.Context` for cancellation. |

### Channels

| Pattern | Channel Type | Description |
|---------|-------------|-------------|
| Synchronization | `chan struct{}` (unbuffered) | Signal completion. Send = done. |
| Data transfer | `chan T` (unbuffered) | One goroutine produces, one consumes. |
| Buffering | `chan T` (buffered) | Decouple producer/consumer speed. |
| Semaphore | `chan struct{}` (buffered N) | Limit concurrent work to N. |
| Fan-out | Send work to N workers via shared channel | Workers read from same channel. |
| Fan-in | N producers send to one channel | Consumer reads from one channel. |
| Request-reply | Channel of channels | Include `resultChan chan T` in the request struct. |

```go
// Semaphore pattern: limit to MaxOutstanding concurrent operations
sem := make(chan struct{}, MaxOutstanding)
for req := range requests {
    sem <- struct{}{}        // acquire
    go func(r Request) {
        defer func() { <-sem }() // release
        process(r)
    }(req)
}

// Fixed worker pool (preferred)
for range MaxOutstanding {
    go func() {
        for req := range requests {
            process(req)
        }
    }()
}
```

### Select

```go
select {
case msg := <-ch1:
    handle(msg)
case ch2 <- result:
    // sent
case <-ctx.Done():
    return ctx.Err()
default:
    // non-blocking: no channel ready
}
```

**Rule**: Use `default` in `select` only when you need non-blocking behavior (e.g., leaky buffer pattern). Without `default`, `select` blocks until a case is ready.

---

## 10. Error Handling

### Conventions

| Rule | Practice |
|------|---------|
| Return `error` as last value | `func Do() (Result, error)` |
| Check immediately | `if err != nil { return ..., fmt.Errorf("do: %w", err) }` |
| Error strings are lowercase, no punctuation | `"open file"` not `"Open file."` |
| Prefix with context | `fmt.Errorf("save user %d: %w", id, err)` -- each layer adds its context |
| Sentinel errors at package boundary | `var ErrNotFound = errors.New("not found")` |
| Use `errors.Is` / `errors.As` for checking | Not `==` (wrapping breaks equality) |
| Custom error types for rich info | Struct implementing `Error() string` with fields like Op, Path, Err |

```go
// Wrapping
if err != nil {
    return fmt.Errorf("save user %d: %w", id, err)
}

// Checking
if errors.Is(err, ErrNotFound) { ... }

// Extracting
var pathErr *os.PathError
if errors.As(err, &pathErr) {
    fmt.Println(pathErr.Path)
}
```

### Panic and Recover

| Rule | Practice |
|------|---------|
| Don't panic in library code | Return errors. Let the caller decide. |
| Panic only for truly unrecoverable states | Programmer bugs, impossible conditions, failed critical init. |
| Recover only in deferred functions | `recover()` returns nil outside deferred context. |
| Convert panic to error within a package | Use a local error type, recover in the exported function, re-panic on unexpected types. |
| Never expose panics to callers | Exported API returns errors, never panics. |

```go
// Internal panic-to-error conversion pattern
type parseError struct{ msg string }

func (e parseError) Error() string { return e.msg }

func Parse(input string) (result Result, err error) {
    defer func() {
        if r := recover(); r != nil {
            if e, ok := r.(parseError); ok {
                err = e // convert to error
            } else {
                panic(r) // re-panic unexpected types
            }
        }
    }()
    return doParse(input) // may panic with parseError internally
}
```

---

## 11. Initialization

### Zero Values

Design types so the zero value is useful:

| Type | Zero Value | Useful? |
|------|-----------|---------|
| `int`, `float64` | `0` | Yes |
| `string` | `""` | Yes |
| `bool` | `false` | Yes |
| `*T` | `nil` | Depends (check before use) |
| `[]T` | `nil` | Yes (`len`=0, `append` works) |
| `map[K]V` | `nil` | Reads work, writes panic -- use `make` |
| `sync.Mutex` | Unlocked | Yes (no constructor needed) |
| `bytes.Buffer` | Empty buffer | Yes (ready to Write) |

### Init Functions

| Rule | Practice |
|------|---------|
| Use sparingly | Prefer explicit initialization in `main` or constructors. |
| No I/O in init | Don't read files, make network calls, or access databases. |
| Safe uses | Register drivers, set defaults, validate compile-time constants. |
| Multiple init per file | Allowed, but keep it simple. |
| Init order | Imported packages first, then package-level vars, then `init()`. |

---

## 12. Printing

### Format Verbs

| Verb | Use |
|------|-----|
| `%v` | Default format for any value |
| `%+v` | Struct fields with names |
| `%#v` | Full Go syntax representation |
| `%T` | Type of the value |
| `%q` | Quoted string |
| `%x` | Hex (works on strings, byte slices, ints) |
| `%d` | Integer (signedness from type, not format) |
| `%s` | String or `[]byte` |

### String() Method

```go
func (t MyType) String() string {
    return fmt.Sprintf("MyType{name: %s, value: %d}", t.name, t.value)
}
```

**Danger**: Never call `Sprintf` with `%s` on the receiver inside `String()` -- infinite recursion. Convert to the base type first: `fmt.Sprintf("%s", string(t))`.

---

## Cross-References

| Topic | This Doc | See Also |
|-------|----------|----------|
| Naming conventions | Section 2 | Clean Code.md §1 (Naming) |
| Function design | Section 4 | Clean Code.md §2 (Functions) |
| Error handling | Section 10 | Clean Code.md §2 (Error Handling) |
| Concurrency | Section 9 | Clean Code.md §7 (Concurrency) |
| Project structure | -- | Go Server.md, Go Library.md, Go Client.md |
| Package design | Section 2 (Packages) | Go Server.md (MDCA), Go Library.md (vendor structure) |
| Interface design | Section 7 | Go Server.md (ports), Go Library.md (repositories/storage.go) |
| 12-Factor compliance | -- | Twelve-Factor App.md |
| Code smells | -- | Clean Code.md §10 (Smells Reference) |
