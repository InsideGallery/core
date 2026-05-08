# Go Server Useful Packages — Architecture Blueprints

Reusable `pkg/` packages for Go microservices. These are not copied code — they are **architecture blueprints** describing what to build, the public API contract, file structure, and key design decisions.

When starting a new project, implement these packages following the patterns described here. Adapt to your specific stack (database, broker, framework) while keeping the architecture consistent.

---

## 1. pkg/fastlog — Structured Logging with Pluggable Handlers

**Purpose:** Production-grade structured logging with configurable multi-handler output and middleware support.

**Public API:**
```go
func SetupDefaultLog(m ...slogmulti.Middleware)
func GetConfigFromEnv() (*Config, error)

type Config struct {
    Outputs         []string   // LOG_OUTPUTS e.g. "stderr:json,otel:json"
    Level           slog.Level // LOG_LEVEL default "INFO"
    Caller          bool       // LOG_CALLER default true
    ErrorFormatting bool       // LOG_ERROR_FORMATTING default false
}
```

**File Structure:**
```
pkg/fastlog/
├── log.go                          # SetupDefaultLog — reads env, sets slog.Default
├── config.go                       # Config struct, GetConfigFromEnv, composite handler factory
├── handlers/
│   ├── registry.go                 # Plugin registry: RegisterWriter(), RegisterHandlerFunc(), Get()
│   ├── errors.go                   # Sentinel errors for unknown handler/format
│   ├── stderr/stderr.go            # Default stderr handler (text/json) — registered via init()
│   ├── otel/otel.go                # OpenTelemetry slog bridge — registered via init()
│   ├── datadog/datadog.go          # Datadog APM handler — registered via init()
│   └── nop/nop.go                  # Discard handler (testing) — registered via init()
└── middlewares/
    ├── caller.go                   # Adds source file:line to log records
    └── error.go                    # Formats error chain stack traces
```

**Key Dependencies:** `log/slog` (stdlib), `github.com/samber/slog-multi` (fanout + middleware chaining), `github.com/caarlos0/env/v10` (env parsing)

**Design Decisions:**
- init-time handler registration via blank imports (`_ "pkg/fastlog/handlers/stderr"`)
- Multi-handler fanout: `slogmulti.Fanout([]slog.Handler{})`
- Middleware chaining: `slogmulti.Pipe(middlewares...).Handler(handler)`
- Env-driven — no hardcoded log backends
- Each handler outputs the same records independently

---

## 2. pkg/oslistener — OS Signal Handling

**Purpose:** Graceful signal routing with ordered callback chains for shutdown sequences.

**Public API:**
```go
type OsListener interface {
    SignalsToSubscribe() OsSignalsList
    ReceiveSignal(os.Signal)
}

func Start(ctx context.Context, listener OsListener)
func Raise(sig os.Signal) error

type SignalListener struct { /* sync.Mutex protected */ }
func Get() *SignalListener                          // package-level singleton
func NewSignalListener() *SignalListener
func (l *SignalListener) Append(signal os.Signal, fn func())   // add callback (runs last)
func (l *SignalListener) Prepend(signal os.Signal, fn func())  // add callback (runs first)
func (l *SignalListener) Set(signal os.Signal, fn func())      // replace all callbacks
func (l *SignalListener) Reset(signal os.Signal)               // remove all callbacks
```

**File Structure:**
```
pkg/oslistener/
├── listener.go    # OsListener interface, Start goroutine, Raise
└── signal.go      # SignalListener struct, callback map, Get singleton
```

**Dependencies:** Go stdlib only (`os`, `os/signal`, `sync`)

**Design Decisions:**
- Singleton via `Get()` for app-level signal handling
- Ordered callbacks: `Prepend` for cleanup that must run first, `Append` for everything else
- Thread-safe callback map protected by `sync.Mutex`
- `Start()` spawns a goroutine blocking on `signal.Notify()` until `ctx.Done()`

---

## 3. pkg/app — Service Entrypoints (WebMain / WorkerMain)

**Purpose:** Complete lifecycle management for HTTP and worker services. Owns ALL setup — caller provides only a business-logic closure.

**Public API:**
```go
// HTTP service entrypoint
func WebMain(name string, cfg *httpserver.Config, initRouter InitRouter)
type InitRouter func(ctx context.Context, app *fiber.App) error

// Worker service entrypoint (message broker consumer)
func WorkerMain(name, monitorAddr string, initSubs InitSubscriptions)
type InitSubscriptions func(ctx context.Context, sub *Subscriber) error
```

**File Structure:**
```
pkg/app/
├── web.go     # WebMain: logging → profiler → init closure → signals → listen
└── worker.go  # WorkerMain: logging → profiler → broker connect → init closure → signals → wait
```

**Dependencies:** `pkg/fastlog`, `pkg/httpserver`, `pkg/oslistener`, `pkg/profiler`, HTTP framework, message broker client

**Design Decisions:**
- **No `run()` function pattern** — `WebMain`/`WorkerMain` are the entire `main()` body
- Closure returns error = init failed = fatal exit (`os.Exit(1)`)
- Profiler starts BEFORE init closure (K8s probes available during slow startup)
- `profiler.Started.Store(true)` after init succeeds; `profiler.Ready.Store(true)` after listen
- SIGINT/SIGTERM: set `Ready=false`, close server/broker, exit gracefully

---

## 4. pkg/profiler — K8s Health Probes + pprof

**Purpose:** Standalone HTTP server for Kubernetes probes and Go runtime profiling. Starts before main app.

**Public API:**
```go
var (
    Started atomic.Bool  // true after init succeeds
    Ready   atomic.Bool  // true after server listening
)

func AddHealthCheck(f func() error)  // register dependency check (DB, broker, etc.)
func CheckHealth() error             // runs all checks concurrently
func Monitor(addr string) func()     // starts probe server, returns shutdown func

// Endpoints:
// GET /healthz   → all dependency checks (200 or 503)
// GET /readyz    → Ready flag (200 or 503)
// GET /livez     → always 200 (process alive)
// GET /startupz  → Started flag (200 or 503)
// GET /debug/pprof/*  → Go profiling
```

**File Structure:**
```
pkg/profiler/
└── profiler.go    # Single file: probes, health checks, pprof server
```

**Dependencies:** Go stdlib only (`net/http`, `net/http/pprof`, `sync/atomic`)

**Design Decisions:**
- **Separate port** from main app (e.g., `:8011`) — probes work even if app crashes
- Call `Monitor()` as first thing in `main()`, before any init
- `AddHealthCheck()` for each infrastructure dependency (DB ping, broker ping, etc.)
- Health checks run **concurrently** with `errors.Join` aggregation
- `/livez` never checks dependencies — it only proves the process responds to HTTP

---

## 5. pkg/swagger — Swagger UI for Fiber

**Purpose:** Embedded Swagger/OpenAPI documentation served alongside APIs.

**Public API:**
```go
type Handler struct { /* unexported */ }
func NewHandler(host, specJSON, uiHTML string) *Handler
func (h *Handler) Register(router fiber.Router)  // mounts /swagger and /swagger.json
```

**File Structure:**
```
pkg/swagger/
└── handler.go    # Handler type, Register, UI/Spec endpoints
```

**Dependencies:** Fiber v3

**Design Decisions:**
- Spec and UI HTML passed as **embedded strings** (via `go:embed` in each service's `docs/` package)
- Host substituted at runtime (dev: `localhost:8080`, prod: `api.example.com`)
- Each service embeds its own spec — no shared spec file

---

## 6. pkg/netutil — Network Utilities

**Purpose:** Extract real client IP from HTTP requests behind reverse proxies (K8s Ingress, Traefik, AWS ALB).

**Public API:**
```go
func IPFromRequest(c fiber.Ctx) (net.IP, error)
func IPStringFromRequest(c fiber.Ctx) string
```

**File Structure:**
```
pkg/netutil/
└── ip.go    # IP extraction with proxy header handling
```

**Dependencies:** Go stdlib + Fiber context

**Design Decisions:**
- Check `X-Real-Ip` first (set by Nginx/Traefik)
- Scan `X-Forwarded-For` list, skip private/loopback IPs, use first public IP
- Fallback to `c.IP()` (direct connection)
- Strip port from remote address if present

---

## 7. pkg/scope — Authorization Scope Engine

**Purpose:** Parse, validate, and match authorization scopes in `method:ownership:path` format. Zero external dependencies.

**Public API:**
```go
// Format: "method:ownership:path"
// Examples: "read:own:/v2/api/shifts", "write:admin:/v2/*", "*"

type Scope struct {
    Method    string  // "read", "write", "*"
    Ownership string  // "own", "admin", "*"
    Path      string  // "/v2/myapi/*", "*"
}

func Parse(s string) (Scope, error)
func ValidateScope(s string) error
func BuildScope(method, ownership, path string) string
func Match(s Scope, method, path string) bool
func CheckAccess(scopes []string, method, path string) (allowed bool, ownership string)
func MergeScopes(global, site []string) []string
func HasCapability(capabilities []string, cap string) bool
func MergeCapabilities(global, site []string) []string
```

**File Structure:**
```
pkg/scope/
└── scope.go    # Single file, pure Go stdlib
```

**Dependencies:** None — Go stdlib only

**Design Decisions:**
- **Pure logic, zero dependencies** — can be imported anywhere without pulling in frameworks
- Wildcard `*` matches everything (superadmin scope)
- Path matching: `/v2/api/*` matches `/v2/api/shifts`, `/v2/api/trades/123`
- `CheckAccess` returns **most permissive** ownership across all matching scopes
- Ownership hierarchy: `*` = `admin` > `own`
- Separate from auth middleware — scope logic is reusable without HTTP framework

---

## 8. pkg/auth — JWT Authentication & Middleware

**Purpose:** JWT token issuance/verification, HTTP middleware, claims management. Supports impersonation, per-site scopes, and legacy cookie compatibility.

**Public API:**
```go
// Config
type Config struct {
    JWTSecret    string        // AUTH_JWT_SECRET (required)
    TTL          time.Duration // AUTH_TTL default "1h"
    KeepAliveTTL time.Duration // AUTH_KEEP_ALIVE_TTL default "2160h"
}
func GetEnvConfig() (*Config, error)

// Interfaces
type TokenVerifier interface {
    Verify(ctx context.Context, token string) (*Claims, error)
}
type TokenIssuer interface {
    Issue(ctx context.Context, userID int, scopes []string) (string, error)
    IssueWithTTL(ctx, userID, scopes, ttl) (string, error)
    IssueImpersonation(ctx, realUserID, targetUserID, scopes) (string, error)
    IssueWithSites(ctx, userID, globalScopes, []SiteAccess) (string, error)
}

// Implementation
type HMACTokenService struct { /* HS256 */ }
func NewHMACTokenService(secret string, ttl time.Duration) (*HMACTokenService, error)

// Claims
type Claims struct {
    UserID         int
    ActingAsUserID *int        // non-nil = impersonating
    Scopes         []string    // global scopes
    Sites          []SiteAccess // per-site scopes + capabilities
    jwt.RegisteredClaims
}
type SiteAccess struct {
    SiteID       int
    Scopes       []string
    Capabilities []string
}

// Middleware
func AuthMiddleware(verifier TokenVerifier) fiber.Handler   // full: JWT + scopes + site
func JWTMiddleware(verifier TokenVerifier) fiber.Handler     // signature only (refresh)

// Context getters (after middleware)
func GetClaims(c fiber.Ctx) *Claims
func GetSiteID(c fiber.Ctx) int
func GetOwnership(c fiber.Ctx) string       // "own" or "admin"
func GetEffectiveUserID(c fiber.Ctx) int    // impersonation-aware
func GetRealUserID(c fiber.Ctx) int         // always the real user (for audit)
func GetCapabilities(c fiber.Ctx) []string

// Test helper
func SetTestContext(c fiber.Ctx, claims *Claims, ownership string, siteID int)
```

**File Structure:**
```
pkg/auth/
├── config.go       # Config + GetEnvConfig
├── hmac.go         # HMACTokenService (Issue*, Verify)
├── claims.go       # Claims struct + methods (EffectiveUserID, MergedScopes, etc.)
├── middleware.go    # AuthMiddleware, JWTMiddleware, context getters, SetTestContext
└── cookie.go       # (optional) Legacy cookie encoder for backward compatibility
```

**Dependencies:** `golang-jwt/jwt/v5`, `gofiber/fiber/v3`, `pkg/scope`, `caarlos0/env/v10`

**Design Decisions:**
- `pkg/scope` handles pure logic; `pkg/auth` handles HTTP + JWT concerns
- AuthMiddleware: verify JWT → read `X-Site-ID` header → merge global + site scopes → `scope.CheckAccess` → set context
- Impersonation: `ActingAsUserID` in claims; `GetEffectiveUserID()` returns target for authorization, `GetRealUserID()` for audit
- Per-site access: JWT carries `Sites[]` array with per-site scopes and capabilities
- `SetTestContext()` enables handler testing without real JWT infrastructure
- Cookie support is optional (for legacy system backward compatibility)

---

## Cross-Package Lifecycle Integration

```
main()
  ├── fastlog.SetupDefaultLog()              # 1. Logging first
  ├── defer profiler.Monitor(":8011")        # 2. Health probes second (before init)
  └── app.WebMain("myservice", cfg, func(ctx, fiberApp) error {
        ├── db := postgres.NewDB(ctx, cfg)   # 3. Connect dependencies
        ├── profiler.AddHealthCheck(db.Ping)  # 4. Register health checks
        ├── tokenSvc := auth.NewHMAC(secret)  # 5. Auth setup
        ├── services := app.NewServices(...)  # 6. Wire domain services
        ├── handler.SetupRoutes(fiberApp, services, tokenSvc)  # 7. Routes
        └── return nil                        # 8. Init complete → profiler.Started = true
      })
```

**Signal handling** (automatic via `app.WebMain`):
- SIGTERM/SIGINT → `profiler.Ready = false` → stop accepting → drain connections → exit

**Authorization flow** (per request):
```
Request → AuthMiddleware → verify JWT → read X-Site-ID → merge scopes
  → scope.CheckAccess(method, path) → set ownership + capabilities in context
  → Handler reads: GetEffectiveUserID(), GetOwnership(), GetSiteID()
```
