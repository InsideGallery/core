# Go Client Application Architecture

This document covers architecture, patterns, and conventions specific to
**UI client applications** built in Go. This is NOT a server, NOT an HTTP API,
NOT a microservice mesh. This is a desktop/mobile application with a game loop,
a render pipeline, and direct user interaction.

See `instruction.md` for general architecture (MDCA, DDD, Event-Driven, Go best
practices). This document extends it with client-specific concerns.

---

## Table of Contents

1. [Client vs Server — Mindset Shift](#client-vs-server--mindset-shift)
2. [Game Loop Architecture](#game-loop-architecture)
3. [Application Shell](#application-shell)
4. [Rendering Pipeline](#rendering-pipeline)
5. [Coordinate Systems](#coordinate-systems)
6. [Input Handling](#input-handling)
7. [HiDPI and Scaling](#hidpi-and-scaling)
8. [Asset Management](#asset-management)
9. [Scene and State Management](#scene-and-state-management)
10. [Camera](#camera)
11. [UI Primitives and Theming](#ui-primitives-and-theming)
12. [Window Management](#window-management)
13. [Platform Abstraction](#platform-abstraction)
14. [Configuration and Persistence](#configuration-and-persistence)
15. [Mobile Builds](#mobile-builds)
16. [Performance Budget](#performance-budget)
17. [Client Anti-Patterns](#client-anti-patterns)

---

## Client vs Server — Mindset Shift

Forget everything about request/response, HTTP handlers, middleware chains,
database connections, container orchestration, and load balancers. A client
application has fundamentally different concerns:

| Server Concept | Client Equivalent |
|---|---|
| HTTP handler | System.Update() per frame |
| Request routing | Scene Manager + state machine |
| Response rendering (JSON/HTML) | System.Draw() to screen image |
| Database connection pool | Resource Manager (asset cache) |
| Middleware chain | System execution order |
| Session state | Component data in Registry |
| Load balancer | N/A — single user, single instance |
| Container/k8s | Platform package (OS-specific window calls) |
| REST endpoint | Event Bus subscription |
| Background job queue | Goroutine with context cancellation |
| SQL migration | Config migration (old format detection) |

**Key difference:** A server processes independent requests concurrently. A
client runs a single-threaded game loop at 60 FPS where every frame is a
complete cycle of input processing, state update, and rendering.

---

## Game Loop Architecture

The application runs a **fixed game loop** driven by the framework (Ebiten).
Every frame (~16.6ms at 60 FPS), two methods are called in sequence:

```
┌──────────────────────────────────────────┐
│                 Frame N                  │
│                                          │
│  Update()                                │
│    ├── Process input (mouse, keyboard)   │
│    ├── Update state (timers, animations) │
│    ├── Handle events (Bus pub/sub)       │
│    └── Transition logic (scene/state)    │
│                                          │
│  Draw(screen)                            │
│    ├── Clear screen                      │
│    ├── Draw world content → World buffer │
│    ├── Apply camera transform to screen  │
│    └── Draw UI overlays → screen direct  │
│                                          │
│  Layout(outsideW, outsideH) → (w, h)    │
│    └── Return logical resolution         │
└──────────────────────────────────────────┘
```

### Rules

- **Update is for logic.** No drawing calls in Update. No GPU operations.
- **Draw is for rendering.** No state mutation in Draw. It is called with the
  state that Update produced. Draw may be skipped if the window is hidden.
- **Layout is for resolution.** Returns the logical screen size. Called before
  Draw. This is where HiDPI scaling is computed.
- **No blocking in Update or Draw.** Both run on the main thread. Blocking
  freezes the entire UI. Long operations (asset loading, file I/O) go in
  goroutines with progress reporting.

---

## Application Shell

The app shell (`pkg/app/`) is a thin wrapper around the game loop. It owns:

- Window configuration (size, title, decoration, transparency)
- Scene Manager (which scene is active)
- Event Bus (shared across all scenes)
- Plugin loader (external .so modules)
- Window drag handling (for undecorated windows)

The shell contains **zero business logic.** Products configure it via `Config`
and `SetupFunc`:

```go
game := app.New(app.Config{
    Width:       800,
    Height:      600,
    Title:       "My App",
    Transparent: false,
    Decorated:   true,
    DragEnabled: false,
    Setup: func(ctx context.Context, bus *event.Bus, manager *scene.Manager, switchScene func(string)) string {
        // Create scenes, register plugins
        // Return initial scene name
        mainScene := myscene.New(bus, switchScene)
        manager.Add(ctx, mainScene)
        return "main"
    },
})
```

### Shell Responsibilities

| Responsibility | Location | NOT Here |
|---|---|---|
| Window creation | app.Config | Scene code |
| Scene wiring | SetupFunc | Domain logic |
| Event Bus creation | app.New() | Per-scene buses |
| Plugin loading | app.initApp() | Scene code |
| Window drag | app.updateDrag() | Input systems |
| Frame dispatch | Update/Draw/Layout | Direct framework calls |

---

## Rendering Pipeline

Drawing follows a strict two-pass pipeline:

### Pass 1: World Space (via Camera)

Systems draw to an offscreen `World` image in world coordinates. This content
is affected by camera position, zoom, and rotation.

```
Systems.Draw(ctx, World)  →  World composited to screen via Camera.WorldMatrix()
```

### Pass 2: Screen Space (UI Overlays)

Systems implementing `SystemWindow` draw directly to screen in pixel
coordinates. This content is NOT affected by camera transforms.

```
SystemWindow.ScreenDraw(ctx, screen)  →  cursor, debug HUD, fixed UI
```

### Drawing Primitives

All drawing goes through `pkg/ui/` — never call framework drawing functions
directly from systems:

| Primitive | Function |
|---|---|
| Filled rounded rect | `ui.DrawRoundedRect(dst, x, y, w, h, radius, color)` |
| Rect outline | `ui.DrawRoundedRectStroke(dst, x, y, w, h, radius, strokeW, color)` |
| Circle | `ui.DrawCircle(dst, cx, cy, radius, color)` |
| Arc / ring segment | `ui.DrawArc(dst, cx, cy, outerR, innerR, start, end, color)` |
| Gradient arc | `ui.DrawGradientArc(dst, cx, cy, outerR, innerR, start, end, startClr, endClr)` |
| Filled polygon | `ui.DrawFilledPolygon(dst, points, color)` |
| Text (left-aligned) | `ui.DrawText(dst, str, face, x, y, color)` |
| Text (centered) | `ui.DrawTextCentered(dst, str, face, cx, y, color)` |
| Icons | `ui.DrawPlayIcon`, `DrawPauseIcon`, `DrawCloseIcon`, etc. |

### Image Drawing

```go
op := &ebiten.DrawImageOptions{}
op.GeoM.Translate(x, y)
op.GeoM.Scale(sx, sy)
screen.DrawImage(img, op)
```

### Vertex-Based Drawing

Complex shapes use `vector.Path` → triangulation → `DrawTriangles`. The `ui/`
package handles this internally. Systems call `ui.Draw*` functions and never
build vertex buffers directly.

---

## Coordinate Systems

A client application deals with multiple coordinate spaces. Mixing them is the
number one source of visual bugs.

### Three Spaces

| Space | Origin | Units | Used By |
|---|---|---|---|
| **Screen** | Top-left of window | Physical pixels (after HiDPI) | Layout(), ScreenDraw(), UI overlays |
| **World** | Map origin (0,0) | Map units (TMX pixels, game units) | Systems.Draw(), Camera, entity positions |
| **Logical** | Top-left of window | Density-independent pixels | Config values, design specs |

### Conversion Rules

```
Logical → Screen:    ui.S(logicalValue)           // multiply by UIScale
Screen  → World:     camera.ScreenToWorld(sx, sy)  // invert camera matrix
World   → Screen:    camera.WorldMatrix().Apply()   // apply camera matrix
Map     → Screen:    co.MapToScreenX(mapX)          // via Coords helper
Screen  → Map:       co.ScreenToMapX(screenX)       // via Coords helper
```

### Rules

- **Store positions in world/map coordinates.** Screen pixels change with
  zoom/pan/resize. World coordinates are stable.
- **Scale at draw time.** Convert from world to screen in the render system,
  not in component data.
- **Use Coords helper** for TMX-based conversions. Never write
  `obj.X * scaleX + offsetX` manually.
- **Use `ui.S()` for logical UI measurements.** Button sizes, padding, font
  sizes — always scale by device factor.

---

## Input Handling

All user input flows through a single pipeline. There are NO click handlers on
widgets, NO manual coordinate checks in render code.

### RTree-Based Input System

```
Mouse/Touch event
    ↓
InputSystem.Update()
    ↓
Cursor position → spatial query on RTree
    ↓
Find intersecting Zone(s)
    ↓
Priority sort (lowest = first)
    ↓
Dispatch: OnClick / OnDrag / OnHover
```

### Zone Registration

```go
inputSystem.AddZone(&systems.Zone{
    Spatial:  shapes.NewBox(x, y, x+w, y+h),
    OnClick:  func() { /* handle click */ },
    OnHover:  func(hovered bool) { /* visual feedback */ },
    Priority: 0, // lower = checked first
})
```

### Input Rules

| Rule | Rationale |
|---|---|
| ALL input through RTree InputSystem | Single source of truth for hit detection |
| No hit testing in Draw | Draw is pure rendering, no logic |
| No coordinate math in components | Components store data, not behavior |
| Zone priorities for overlapping areas | Explicit resolution, no z-fighting |
| CursorOverride for virtual cursors | Delta-based positioning for custom cursors |
| ScrollOffset for scrollable containers | Zones stay in content-space; offset converts screen to content |

### Click vs Drag

The InputSystem distinguishes:
- **Click** = press + release on the same zone (no drag handler registered)
- **Drag** = press on a zone with OnDragStart → OnDrag each frame → OnDragEnd

A zone with both OnClick and OnDragStart will only fire OnClick if no drag
movement occurred.

### Keyboard Input

Keyboard handling lives in dedicated systems, not in InputSystem. Check keys
directly in Update:

```go
if inpututil.IsKeyJustPressed(ebiten.KeyEscape) { ... }
if ebiten.IsKeyPressed(ebiten.KeyShift) { ... }
```

### Mouse Wheel / Scroll

ScrollSystem handles wheel input and updates Scrollable components:

```go
_, dy := ebiten.Wheel()
if dy != 0 {
    scrollable.Scroll += int(dy * -scrollSpeed)
    scrollable.Scroll = clamp(scrollable.Scroll, 0, max)
}
```

---

## HiDPI and Scaling

Desktop and mobile screens have different pixel densities. All UI must render
correctly on 1x, 1.25x, 1.5x, 2x, and higher scale factors.

### Architecture

```
Layout(outsideW, outsideH):
    scale = ebiten.Monitor().DeviceScaleFactor()
    ui.UIScale = scale
    return ceil(outsideW * scale), ceil(outsideH * scale)
```

### Scaling Functions

```go
ui.S(32)        // float32: 32 logical → 64 physical at 2x
ui.Sf(32.0)     // float64: same thing
ui.Face(false, 14) // font face at 14pt logical, scaled internally
```

### Rules

- **Design in logical pixels.** All constants, config values, and layout specs
  use density-independent values.
- **Scale at the boundary.** Call `ui.S()` when converting logical to physical
  for drawing. Store logical values, draw physical values.
- **Font faces auto-scale.** `ui.Face(bold, size)` multiplies size by UIScale
  internally. Do not multiply again.
- **Images scale via GeoM.** Use `op.GeoM.Scale(scaleX, scaleY)` in
  DrawImageOptions. Do not pre-scale images.

---

## Asset Management

Client applications load and cache assets: images, fonts, sounds, tile maps.
This is fundamentally different from server-side database access.

### Resource Manager

```go
resources := resources.NewManager()

// Synchronous (small assets, startup)
resources.LoadImageFromFS(embedFS, "icons/play.png", "icon-play")

// Asynchronous (large assets, with progress)
resources.LoadAsync([]resources.LoadTask{
    {Key: "bg-main", Load: func() (any, error) { return loadImage("bg.png") }},
    {Key: "bg-alt",  Load: func() (any, error) { return loadImage("alt.png") }},
})

// Progress tracking
loaded, total := resources.Progress()
isLoading := resources.IsLoading()

// Retrieval
img, ok := resources.GetImage("bg-main")
```

### Asset Sources

| Source | When | How |
|---|---|---|
| `embed.FS` | Core assets (fonts, icons) | `go:embed` + `LoadImageFromFS` |
| Disk files | Game assets, user content | `os.ReadFile` + `LoadImageFromBytes` |
| Generated | Runtime icons, tray images | In-memory creation |

### Rules

- **Load once, cache forever.** Resources are immutable after loading. No
  reload-on-change (this is not a web server with hot reload).
- **Async loading for large assets.** Show a progress bar. Never block the
  game loop waiting for disk I/O.
- **Key naming convention:** `"type-name"` — `"bg-main"`, `"avatar-suspect-1"`,
  `"font-regular"`.
- **Dispose on scene Unload** if the asset is scene-specific and large.
  Framework images are GPU resources — holding them costs VRAM.

### Embedded Assets

```go
//go:embed fonts/*.ttf
//go:embed icons/*.png
var Content embed.FS

//go:embed AppIcon.png
var AppIcon []byte
```

Embedded assets are compiled into the binary. They are always available, never
fail to load, and have zero disk I/O at runtime. Use for core UI assets.

---

## Scene and State Management

### Scenes

A scene is a self-contained screen with its own ECS (Systems, Registry, RTree,
Camera, Resources). Scenes are the top-level organizational unit.

```
Scene Manager
    ├── "main"     → MainScene (active)
    ├── "settings" → SettingsScene
    └── "mini"     → MiniScene
```

Transitions: `manager.SwitchSceneTo("settings")` — calls `Load()` on new,
`Unload()` on previous.

### State Machines (Within a Scene)

Complex scenes use an internal state machine instead of multiple scenes:

```go
const (
    StateLoading    GameState = iota
    StateMenu
    StateGameplay
    StateResults
)

type State struct {
    Current  GameState
    BootTick int
}
```

State transitions happen in a StateSystem. Entity creation/destruction is tied
to state transitions via assemblage functions:

```
StateSystem.Update():
    if state changed:
        CleanStateGroups(registry)       // remove old entities
        CreateNewStateEntities(registry) // build new entities
```

### Rules

- **One scene = one BaseScene** (own Systems, Registry, RTree, Bus, Camera,
  Resources).
- **Scenes communicate via Event Bus only.** Never import another scene.
- **State machines for intra-scene flow.** Multiple visual states within one
  scene — not one scene per state.
- **Clean up on Unload.** Remove systems, clear registry, clear RTree zones.
  Prevent stale state from leaking into the next Load.

---

## Camera

Camera provides world-to-screen transformation: pan, zoom, rotate.

```go
camera := core.NewCamera(shapes.NewPoint())
camera.SetViewPort(screenW, screenH)
camera.ZoomFactor = 50.0  // exponential: pow(1.01, zoom)
camera.Position = shapes.NewPoint(worldX, worldY)
```

### Two Rendering Modes

| Mode | Camera | Use Case |
|---|---|---|
| **World buffer** | `scene.World` is not nil | Full-screen games, scrollable maps. Systems draw to World, composited via `camera.WorldMatrix()` |
| **Direct to screen** | `scene.World` is nil | Simple UIs (timer, settings). No camera transform. |

### Screen-to-World Conversion

```go
worldX, worldY := camera.ScreenToWorld(screenX, screenY)
```

Required for input: convert mouse screen position to world coordinates before
querying RTree or checking entity positions.

---

## UI Primitives and Theming

### Theme System

Two themes: dark and light. Colors are package-level variables switched by
`ui.SetTheme()`:

```go
ui.SetTheme(ui.ThemeDark)
// Now ui.ColorBgPrimary, ui.ColorTextPrimary, etc. are dark-theme values
```

| Variable | Purpose |
|---|---|
| `ColorBgPrimary` | Main background |
| `ColorBgSecondary` | Card/panel background |
| `ColorTextPrimary` | Primary text |
| `ColorTextSecond` | Secondary/muted text |
| `ColorAccentFocus` | Primary action color |
| `ColorAccentBreak` | Secondary action color |
| `ColorAccentDanger` | Destructive action color |
| `ColorWindowBg` | Window background (with transparency) |
| `ColorCardBg` | Card background (with transparency) |

### Transparency

```go
ui.ApplyTransparency(0.3)  // 0.0 = opaque, 1.0 = fully transparent
// Adjusts alpha on WindowBg, CardBg, CardBorder
```

### Font Faces

```go
titleFace := ui.Face(true, 18)   // bold, 18pt logical
bodyFace  := ui.Face(false, 12)  // regular, 12pt logical
```

Font faces are created per use — they are lightweight wrappers around the
shared font source.

### Layout Constants

```go
const (
    RadiusCard   = 12  // logical pixels, use ui.S() to convert
    RadiusButton = 8
)
```

---

## Window Management

Client applications manage their own window — this is fundamentally different
from server processes.

### Window Properties

| Property | Set At | Via |
|---|---|---|
| Size | Startup / runtime | `ebiten.SetWindowSize(w, h)` |
| Title | Startup | `ebiten.SetWindowTitle(title)` |
| Decoration | Startup | `ebiten.SetWindowDecorated(bool)` |
| Resizable | Startup | `ebiten.SetWindowResizingMode(mode)` |
| Fullscreen | Startup / runtime | `ebiten.SetFullscreen(bool)` |
| Transparency | Startup | `RunGameOptions{ScreenTransparent: true}` |
| Icon | Startup | `ebiten.SetWindowIcon([]image.Image)` |
| Close handling | Startup | `ebiten.SetWindowClosingHandled(true)` |
| Position | Runtime | `ebiten.SetWindowPosition(x, y)` |

### Window Close Handling

Desktop applications may need to intercept window close (hide to tray instead
of quit):

```go
ebiten.SetWindowClosingHandled(true)  // prevent default close

// In Update():
if ebiten.IsWindowBeingClosed() {
    platform.HideWindow("AppTitle")  // hide instead of quit
}
```

### Undecorated Windows

For custom-chrome applications (no OS title bar):

```go
ebiten.SetWindowDecorated(false)  // no OS title bar
app.Config{DragEnabled: true}     // enable manual window drag
```

The app shell implements drag by tracking mouse delta and calling
`ebiten.SetWindowPosition`.

---

## Platform Abstraction

OS-specific operations are isolated in `pkg/platform/`. Build tags select the
implementation.

### Platform Interface

| Operation | Desktop | Mobile |
|---|---|---|
| Data directory | `~/.config/appname/` | Set by mobile entry point |
| Show/hide window | X11/macOS/Windows API | N/A |
| Raise window | X11/macOS/Windows API | N/A |
| Asset filesystem | `os.DirFS` or `embed.FS` | `embed.FS` only |

### Build Tags

```
platform_linux.go    // +build linux,!android
platform_darwin.go   // +build darwin
platform_windows.go  // +build windows
platform_android.go  // +build android
```

### Rules

- **All OS calls go through `pkg/platform/`.** Systems and scenes never call
  OS APIs directly.
- **Build tags for platform variants.** Never use runtime `GOOS` checks for
  code that only compiles on one platform.
- **Mobile entry points** set platform state before the game loop starts
  (`platform.SetDataDir()`).

---

## Configuration and Persistence

Client applications persist user preferences and state to local files.
No databases. No remote config services.

### Configuration

```go
cfg := config.Load()    // reads ~/.config/appname/config.json
cfg.Theme = "light"
config.Save(cfg)        // writes back
```

- **JSON format** — human-readable, editable, diffable.
- **Defaults via `Default()` function** — zero-value config is never used raw.
- **Migration** — detect old formats, convert in `Load()`.
- **Fail gracefully** — missing or corrupted config returns defaults, never
  panics.

### Application State

```go
state := config.LoadState()  // reads state.json
state.Round = 3
config.SaveState(state)      // writes back
```

State is separate from config. Config is user preferences (what they chose).
State is runtime data (where they left off).

### Rules

- **Config and state are separate files.** Config = user preferences.
  State = session data.
- **Load returns defaults on error.** Never panic on corrupted or missing files.
- **Save with restricted permissions** — `0o600` for user-only access.
- **No `init()` for loading config.** Load explicitly in `SetupFunc` or when
  the settings scene loads.

---

## Mobile Builds

The same codebase targets both desktop (native binary) and mobile (Android/iOS).

### Architecture

```
services/<name>/mobile/
    └── android/
        └── app/src/main/java/.../MainActivity.java

services/<name>/cmd/<name>/main.go          // desktop entry point
services/<name>/mobile/main_mobile.go       // mobile entry point (build-tagged)
```

### Mobile Differences

| Concern | Desktop | Mobile |
|---|---|---|
| Entry point | `func main()` | `func main()` with `ebitenmobile` binding |
| Window management | Full control | OS manages |
| System tray | Available | N/A |
| File paths | `os.UserHomeDir()` | Passed from Java/ObjC bridge |
| Input | Mouse + keyboard | Touch + virtual cursor |
| Scaling | `Monitor().DeviceScaleFactor()` | OS-reported density |

### Rules

- **Feature-gate platform-specific code** with build tags, not runtime checks.
- **Touch input** maps to mouse events via `CursorOverride` for virtual cursor.
- **No tray, no window decoration, no window positioning** on mobile — use
  build tags to exclude.

---

## Performance Budget

A client application has a hard real-time constraint: **16.6ms per frame at
60 FPS.** Every millisecond counts.

### Budget Allocation

```
Update:  < 5ms   (input, logic, events, state transitions)
Draw:    < 10ms  (rendering all systems)
Layout:  < 0.1ms (resolution calculation)
Margin:  ~1.5ms  (framework overhead, GC)
```

### Rules

- **No allocations in Draw.** Pre-allocate vertex buffers, reuse slices.
  GC pauses cause frame drops.
- **No blocking I/O in Update or Draw.** File reads, network calls — all in
  goroutines with progress reporting.
- **Batch draw calls.** Each `DrawImage` / `DrawTriangles` is a GPU call.
  Minimize the number of calls per frame.
- **Profile before optimizing.** Use `pprof` and frame time counters. Do not
  guess where time is spent.
- **Large image loading is async.** Show a loading screen with progress bar.
  Never freeze the UI.
- **RTree queries are O(log n).** Spatial indexing is fast for hundreds of
  interactive zones. Do not replace with O(n) linear scans.

---

## Client Anti-Patterns

| Anti-Pattern | Problem | Correct Approach |
|---|---|---|
| Blocking in Update/Draw | Freezes UI, drops frames | Goroutine + progress reporting |
| State mutation in Draw | Draw may be skipped; state becomes inconsistent | All state changes in Update |
| Hit detection in render code | Mixes concerns, breaks with camera/zoom | RTree InputSystem |
| Raw `ebiten.CursorPosition()` in systems | Ignores virtual cursor, scroll offset | `scene.GetCursorPos()` or CursorOverride |
| Manual pixel math for TMX objects | Breaks when scale/offset changes | Coords helper |
| Loading assets synchronously | Blocks game loop for seconds | `resources.LoadAsync()` |
| `os.Exit()` in business logic | Bypasses cleanup, loses state | Proper shutdown via scene Unload |
| Hardcoded pixel sizes | Breaks on HiDPI | `ui.S()` for all measurements |
| Creating font faces per frame | Unnecessary allocation pressure | Create in Load, reuse |
| Global mutable state | Race conditions, hidden dependencies | Components in Registry, events via Bus |
| One giant Draw function | Unmaintainable, no reuse | Multiple focused systems |
| HTTP/REST patterns in client code | Wrong paradigm entirely | Game loop + ECS + Event Bus |

---

## Quick Reference

| Area | Client Pattern |
|---|---|
| Main loop | Update() → Draw() → Layout() at 60 FPS |
| Input | RTree spatial queries, Zone-based dispatch |
| Rendering | Two-pass: World (camera) + Screen (overlays) |
| Coordinates | World space for storage, screen space for drawing |
| Scaling | `ui.S()` for all measurements, auto-scaled fonts |
| Assets | `resources.Manager` with async loading + progress |
| State | Components in Registry, transitions via StateSystem |
| Navigation | Scene Manager for top-level, state machines within |
| Communication | Event Bus between modules, direct calls within |
| Config | Local JSON files, defaults on error, separate from state |
| Platform | Build-tagged packages, never raw OS calls in scenes |
| Performance | 16.6ms budget, no allocs in Draw, no blocking I/O |
