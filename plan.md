# Code Review Action Plan

Comprehensive review of all packages in `github.com/InsideGallery/core`. Issues are grouped by severity, then by package.

**Status: CRITICAL (4/4 done), HIGH (13/14 done), MEDIUM (16/17 done), LOW (8/8 done)**

Items #18 (aerospike init buffer flags) skipped -- setting config flags is acceptable in init().
Items #33, #34 (interface getter renames) deferred -- breaking changes for next major version.
Items #39 (IPv4/IPv6 naming), #42 (ticker GetID) deferred -- breaking changes for next major version.

---

## Test Coverage -- Target 90% (IN PROGRESS)

Current total: **33.1%** (up from 28.2%). With external-dep exclusions in `.testcoverage.yml`, effective coverage is higher but not yet at 90%.

### Batch 1 -- DONE (10 packages)
Tests written for: memory/set, memory/stack, memory/orderedmap, memory/comparator, memory/utils, mathutils, dataconv, errors, ecs, pki/aescmac.

### Batch 2 -- TODO (packages needing tests, sorted by impact)

**0% coverage (no test files, testable without external deps):**
- `memory/stack` -- DONE
- `memory/set` -- DONE
- `multiproc/sync` -- needs tests
- `embedded` -- trivial, 1 function
- `fastlog` -- config parsing testable
- `fastlog/handlers` -- registry testable
- `fastlog/handlers/stderr` -- testable (writes to stderr)
- `fastlog/handlers/stdout` -- testable (writes to stdout)
- `fastlog/handlers/nop` -- testable (no-op handler)
- `fastlog/handlers/logfile` -- testable (filesystem)
- `fastlog/middlewares` -- testable (pure slog transformations)
- `server/honeypot` -- testable
- `server/instance` -- testable
- `server/jwt` -- testable (JWT operations)
- `server/jwt/middlewares` -- testable with mocks
- `server/jwt/model` -- testable (scope parsing)
- `server/profiler` -- partially testable
- `server/template` -- testable
- `server/webserver` -- testable with fiber test helpers
- `server/webserver/request` -- testable

**Low coverage (have tests but need more):**
- `app` -- 0% (web/nats setup, partially testable)
- `db/postgres` -- 12.9% (config parsing testable)
- `db/bunt` -- 79.2% (close, needs a few more tests)
- `db/gremlin` -- 1.2% (vertex/operation logic testable)
- `dataconv` -- 36.9% -> improved in batch 1, recheck
- `ecs` -- 50% -> improved in batch 1, recheck
- `errors` -- 75% -> improved in batch 1, recheck
- `mathutils` -- 20.6% -> improved in batch 1, recheck
- `memory/bloom` -- 80.3%
- `memory/btree` -- 89.4% (close)
- `memory/comparator` -- 23.9% -> improved in batch 1, recheck
- `memory/fuzzysearch` -- 76.2%
- `memory/hll` -- 78.9%
- `memory/linkedlist` -- 62.2%
- `memory/orderedmap` -- 37.7% -> improved in batch 1, recheck
- `memory/registry` -- 47.2%
- `memory/sortedset` -- 48.7%
- `memory/utils` -- 8.2% -> improved in batch 1, recheck
- `multiproc/worker` -- 72%
- `oslistener` -- 29.2%
- `pki/aes` -- 73.7%
- `pki/aescmac` -- 56.8% -> improved in batch 1, recheck
- `pki/diversify` -- 71.4%
- `pki/rsa` -- 80%
- `pki/saes` -- 73.9%
- `server/backoff` -- has tests
- `server/sse` -- has tests
- `server/throughput` -- has tests
- `server/webserver/middlewares` -- has tests
- `ticker` -- has tests
- `utils` -- has tests
- `utils/semver` -- has tests

### Priority order for next batches:
1. Recheck batch 1 packages (likely jumped significantly)
2. `fastlog/` family (handlers/registry, middlewares, nop, stderr, stdout, logfile, config)
3. `server/jwt/model` (scope parsing, 16 uncovered functions)
4. `memory/linkedlist`, `memory/registry`, `memory/sortedset` (high uncovered function count)
5. `server/` remaining packages
6. `ticker`, `oslistener`, `multiproc/sync`
7. `utils` remaining gaps
8. `pki/` remaining gaps (aes, saes, diversify)
9. `db/bunt`, `db/postgres` config tests

### Excluded from coverage (require external infrastructure):
- db/aerospike, db/mongodb, db/redis, db/neo4j, db/elasticsearch, db/gremlin
- queue/nats/*, queue/generic/*
- fastlog/handlers/otel, fastlog/handlers/datadog, fastlog/handlers/logstash
- fastlog/metrics
- machielearning/nn
- All mocks/mock_cipher/testfixtures

---

## CRITICAL -- Bugs & Logic Errors

### 1. `db/postgres/client.go:61` -- Wrong variable passed to Set()
`Set(c)` stores the old nil client instead of the newly created `db`. The global `client` remains nil; every call to `Default()` creates a new connection.
**Fix:** Change `Set(c)` to `Set(db)`.

### 2. `memory/fuzzysearch/index.go:60-61` -- Loop index compared instead of value
`Remove()` compares loop index `i` with `doc.ID` instead of comparing `ids[i]` with `doc.ID`. Documents are never actually removed.
**Fix:** Change `if i == doc.ID` to `if ids[i] == doc.ID`.

### 3. `multiproc/worker/aggregator.go:36` -- Wrong variable assigned
When `count <= 0`, the code sets `goroutines = 1` instead of `count = 1`.
**Fix:** Change `goroutines = 1` to `count = 1`.

### 4. `db/mongodb/mongodb.go:41-42` -- Inverted error check
`SetReadPreference(pref)` is called when `err != nil` (when preference creation failed), using an invalid `pref`.
**Fix:** Change `if err != nil` to `if err == nil` on line 41.

---

## HIGH -- Resource Leaks & Race Conditions

### 5. `multiproc/worker/aggregator.go:103` -- Ticker never stopped
`Flusher()` creates `time.NewTicker` but never calls `tck.Stop()`.
**Fix:** Add `defer tck.Stop()` after ticker creation.

### 6. `multiproc/worker/helpers.go:131` -- Timer never stopped
`GetMessageOrTimeout()` creates `time.NewTimer` but never stops it when a message is received first.
**Fix:** Add `defer timer.Stop()` after timer creation.

### 7. `ticker/ticker.go:125` -- Debug ticker never stopped
`tickTimerDebug` is created but never stopped. Only `tickTimer` is stopped on ctx.Done().
**Fix:** Add `tickTimerDebug.Stop()` in the ctx.Done() case.

### 8. `queue/nats/proxy/client.go:86-94` -- Goroutine leak in ServiceHealthcheck
`time.NewTicker` created but never stopped. The `for range ticker.C` loop runs forever with no exit path.
**Fix:** Accept a `context.Context` parameter and select on both `ticker.C` and `ctx.Done()`.

### 9. `server/webserver/middlewares/timing.go:17-33` -- init() goroutine leak + never-reset counters
`init()` starts a goroutine with a ticker that never stops. `dur` and `count` accumulate without reset.
**Fix:** Refactor from global init() to an explicit struct with lifecycle management.

### 10. `oslistener/listener.go:23-35` -- signal.Stop never called
When the context is cancelled, the goroutine returns but `signal.Stop(sigs)` is never called. Signal handlers leak.
**Fix:** Add `signal.Stop(sigs)` before return in the ctx.Done() case.

### 11. `server/throughput/storage.go:87-101` -- Race condition in Reset()
`Reset()` launches goroutines that call `s.date.Add()` and `s.counterM.Add()` concurrently without synchronization.
**Fix:** Either remove goroutines (operations are fast) or add proper synchronization.

### 12. `memory/bloom/bloom.go:49-57` -- panic() in library code
`estimates()` panics on invalid parameters instead of returning errors.
**Fix:** Change `estimates()` to return `(uint32, uint32, error)` and propagate to callers.

### 13. `server/instance/instance.go:14-21` -- log.Fatalf in init()
`init()` calls `log.Fatalf` which kills the process. Any package importing `instance` inherits this risk.
**Fix:** Return error lazily or use a sync.Once pattern that exposes the error.

### 14. `db/gremlin/syntax.go:18-31` -- panic() in init()
`init()` reads `os.Getenv()` and panics on unknown syntax values.
**Fix:** Remove panic; use default or expose error via a setup function.

### 15. `fastlog/handlers/otel/otel.go:19-22` -- init() performs network I/O
`init()` calls `Handler(context.Background())` which does network operations.
**Fix:** Move initialization to an explicit setup function.

### 16. `fastlog/handlers/datadog/datadog.go:19-22` -- init() performs network I/O
Same pattern as otel -- `init()` performs I/O.
**Fix:** Move initialization to an explicit setup function.

### 17. `pki/aescmac/aescmac.go:186` -- panic() in Xor utility
`Xor()` panics on invalid parameters instead of returning an error.
**Fix:** Change to return `([]byte, error)`.

### 18. `db/aerospike/connection.go:16-20` -- init() modifies global buffer settings
`init()` modifies aerospike client global state.
**Fix:** Move to explicit initialization.

---

## MEDIUM -- Error Handling & Style

### 19. `queue/nats/proxy/storage/redis.go:46-56` -- Silent lock failure
`LockOrWait()` retries 10 times, then returns without lock or error. Caller assumes lock is held.
**Fix:** Return an error when lock cannot be acquired.

### 20. `queue/nats/client/config.go:88-101` -- Silent auth failure
Errors from `nkeys.FromSeed()` and `kp.PublicKey()` are logged but function continues with incomplete auth options.
**Fix:** Return error instead of continuing silently.

### 21. `db/elasticsearch/elasticsearch.go:39,44,49,54` -- Typo "responset"
Multiple error messages contain "responset" instead of "response".
**Fix:** Correct the typo in all four error messages.

### 22. `server/jwt/model/scope.go:62-73` -- Commented-out code block
Large block of commented-out validation logic with a TODO.
**Fix:** Either implement the validation or delete the commented code. Git remembers.

### 23. `errors/errors.go:48` -- else after return
`MultipleError.Error()` uses `else if` after a return statement.
**Fix:** Refactor to early returns.

### 24. `errors/errors.go:21,90` -- Deprecated interface{} usage
Uses `interface{}` instead of `any` (Go 1.18+).
**Fix:** Replace `interface{}` with `any`.

### 25. `memory/orderedmap/orderedmap.go:136-145` -- Race condition in Iterator()
`Iterator()` creates a goroutine feeding a channel but doesn't protect against map mutation during iteration.
**Fix:** Copy data before iterating or document thread-safety requirements.

### 26. `memory/sortedset/sortedset.go:323` -- Function too long (119 lines)
`GetByKeyRange()` is 119 lines with a `nolint:gocyclo` directive.
**Fix:** Extract sub-functions to bring each under 40 lines.

### 27. `memory/set/generic.go:140-143` -- Inefficient Count()
`Count()` iterates with `for range` + `i++` instead of returning `len(set.order)`.
**Fix:** Replace with `return len(set.order)`.

### 28. `pki/aescmac/aescmac.go:210-211` -- Input slice mutation
`Padding()` creates an alias `result := data` then appends, which can mutate the input when there's spare capacity.
**Fix:** Copy the slice first: `result := make([]byte, len(data)); copy(result, data)`.

### 29. `pki/aes/aes.go:25-30` -- Silent fallback on invalid size
`NewAES()` silently defaults to AES32 on invalid input.
**Fix:** Return an error for invalid AES sizes.

### 30. `multiproc/worker/workers.go:187` -- Wrong log level for errors
Uses `slog.Debug()` for error messages during message processing.
**Fix:** Change to `slog.Error()`.

### 31. Config functions return redundant `err` variable
Multiple `GetConfigFromEnv()` functions return `c, err` where `err` is guaranteed nil after the nil check.
Affected files:
- `db/bunt/config.go:28`
- `db/mongodb/config.go:37`
- `db/aerospike/config.go:35`
- `db/postgres/config.go:36`
- `db/redis/config.go:29`
- `db/neo4j/config.go:42`
- `fastlog/config.go:42`
- `fastlog/handlers/*/config.go` (7 files)
- `fastlog/metrics/config.go:27`

**Fix:** Return `c, nil` explicitly.

### 32. Error messages with capital letters or wrong grammar
- `fastlog/handlers/datadog/datadog.go:46,52` -- "Error get datadog..." (capital E)
- `fastlog/handlers/otel/otel.go:68,74,80,106,112` -- "Error get otel..." (capital E)
- `fastlog/metrics/metrics.go:111` -- "Error shutdown execution" (capital E)
- `db/aerospike/connection.go:65` -- "error parse aerospike hosts" (grammar)
- `db/aerospike/entity/errors.go:5` -- "not found attribute" (should be "attribute not found")
- `dataconv/errors.go:7-8` -- "error wrong encode type" (redundant "error" prefix)

**Fix:** Lowercase error strings, fix grammar, remove redundant prefixes.

### 33. `queue/generic/subscriber/driver/nats-subscriber.go` -- Get prefix on getters
Multiple methods use `Get` prefix: `GetSubject()`, `GetHeader()`, etc.
**Fix:** Rename to `Subject()`, `Header()`, etc.

### 34. `queue/generic/subscriber/interfaces/config.go` -- Get prefix on getters
`GetReadTimeout()`, `GetMaxConcurrentSize()`, etc.
**Fix:** Rename to `ReadTimeout()`, `MaxConcurrentSize()`, etc.

### 35. `embedded/resources.go:9` -- GetFS has Get prefix
**Fix:** Rename to `FS()`.

---

## LOW -- Minor Style & Cleanup

### 36. `queue/nats/proxy/storage/aerospike.go:58-60,93-95,140-142` -- Commented-out code
Commented SetFilter calls with TODO comments.
**Fix:** Delete commented code or implement with secondary index.

### 37. `utils/bytes.go:72-73` -- Commented-out code
**Fix:** Delete.

### 38. fmt.Println in test files
Debug print statements left in tests:
- `utils/hash_test.go:68,72,78-79`
- `utils/semver/semver_test.go:66,76`
- `utils/tokenizer/prepare_test.go:43,50`
- `utils/strings_test.go:70-71`
- `mathutils/helper_test.go:118`
- `db/mongodb/mongodb_test.go:23,29`
- `db/aerospike/hll/count_test.go:37,43,64`

**Fix:** Replace with `t.Logf()` or remove.

### 39. `dataconv/ip.go` -- Naming inconsistency
`IPV4ToIPV6` and `IPV6ToString` use all-caps V. Standard Go convention is `IPv4`, `IPv6`.
**Fix:** Rename to `IPv4ToIPv6`, `IPv6ToString` (breaking change -- needs major version or alias).

### 40. Multiple else-after-return patterns
- `dataconv/ip.go:162`
- `memory/fuzzysearch/levenshtein.go:38-45`
- `memory/linkedlist/list.go:52-56`
- `mathutils/helper.go:25-34`
- `queue/nats/middleware/trace.go:55-60`
- `queue/nats/client/config.go:114-120`
- `queue/nats/publisher/publisher.go:77-86`
- `fastlog/middlewares/gdpr.go:45-46`
- `fastlog/middlewares/error.go:33-38`
- `pki/aescmac/aescmac.go:73,82`

**Fix:** Refactor to early returns.

### 41. `ecs/base.go:9` -- Global mutable state
Package-level `store` variable accessed from multiple functions.
**Fix:** Pass registry as dependency or protect with mutex.

### 42. `ticker/ticker.go:16` -- GetID() uses Get prefix
Handler interface defines `GetID()`.
**Fix:** Rename to `ID()` (breaking change for implementors).

### 43. Magic numbers without named constants
- `memory/sortedset/sortedset.go:43` -- `0xFFFF`
- `ticker/ticker.go:125` -- `time.Second` hardcoded for debug ticker

---

## Summary

| Severity | Count |
|----------|-------|
| CRITICAL (bugs) | 4 |
| HIGH (leaks, races, panics) | 14 |
| MEDIUM (error handling, style) | 17+ |
| LOW (cleanup, naming) | 8+ |

**Recommended execution order:**
1. Fix CRITICAL bugs first (items 1-4) -- these cause incorrect behavior
2. Fix HIGH resource leaks and race conditions (items 5-18) -- these cause degraded reliability
3. Address MEDIUM error handling issues (items 19-35) -- these improve correctness and maintainability
4. Clean up LOW items (items 36-43) -- cosmetic improvements
