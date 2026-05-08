# Entity-Component-System (ECS) — Definition & Usage Guide

## What is ECS?

ECS is an architectural pattern that separates data from behavior:

- **Entity** = unique identifier (uint64). Has no data, no logic. Just an ID that links components.
- **Component** = pure data struct. No methods, no callbacks, no logic. Only fields.
- **System** = stateless function that reads/writes components. All logic lives here.

Key principle: **composition over inheritance**. An entity's capabilities are defined by which components it has, not by a class hierarchy. Add a `Draggable` component → entity becomes draggable. Remove it → it stops being draggable. No class changes needed.

## Rules

### Components MUST:
- Be plain structs with exported fields
- Contain only data (primitives, pointers, slices)
- Have single responsibility (Transform = position only, Sprite = image only)
- Be reusable across different entity types

### Components MUST NOT:
- Contain methods with logic
- Hold function callbacks
- Reference other components directly
- Import system packages

### Systems MUST:
- Be stateless — all state lives in components
- Have single responsibility (one system = one behavior)
- Operate on components by querying the registry
- Be independent of each other (no system calls another system)

### Systems MUST NOT:
- Store game state in struct fields
- Know about specific entity types
- Call other systems directly
- Access raw OS APIs (use scene accessor)

### Entities MUST:
- Be just an ID (uint64) in the registry
- Be composed of components, never subclassed
- Belong to a group (archetype) that defines their component set

## Our Implementation

### Registry as Indexed ECS
```
Registry[group string, id uint64, value any]
```
- **group** = archetype name ("cursor", "tray_piece", "button")
- **id** = entity identifier
- **value** = `*components.Entity` with component pointers

### Entity Container
```go
type Entity struct {
    Transform   *Transform    // position, size
    Sprite      *Sprite       // image reference
    Renderable  *Renderable   // rotation, alpha, z-order
    Button      *Button       // label, colors
    Clickable   *Clickable    // RTree spatial
    Scrollable  *Scrollable   // scroll offset
    TextBlock   *TextBlock    // multi-line text
    Avatar      *Avatar       // character portrait
    Cursor      *Cursor       // virtual cursor state
    State       *State        // game state machine
    Progress    *ProgressBar  // loading progress
    PuzzleGrid  *PuzzleGrid   // puzzle workspace
    PuzzlePiece *PuzzlePiece  // tray piece link
    Draggable   *Draggable    // drag-and-drop flag
    ImageLayer  *ImageLayerComp // TMX image layer
    TileLayer   *TileLayerComp  // TMX tile layer
}
```
Components are pointers — nil means entity doesn't have that component.

### System Execution Order
Systems register in `BaseScene.Systems` and run in order every frame:
```
1. StateSystem    — state machine, loading, transitions
2. CursorSystem   — virtual cursor from mouse/touch
3. InputSystem    — RTree hit testing, fires button clicks
4. ScrollSystem   — mouse wheel / swipe scrolling
5. DragDropSystem — piece pickup, rotation, placement
6. CameraSystem   — zoom keys / pinch gesture
7. RenderSystem   — ALL drawing
```
Each system has `Update(ctx)` (logic) and `Draw(ctx, screen)` (rendering).

### SceneAccessor Interface
Systems access scene data through an interface, never a concrete type:
```go
type SceneAccessor interface {
    GetRegistry() *Registry
    GetCamera() *Camera
    GetInputSystem() *InputSystem
    GetTileMap() *tilemap.Map
    GetCases() []*CaseConfig
    GetSelectedCase() int
    SetSelectedCase(int)
    GetCursorPos() (int, int)
    SetCursorPos(x, y int)
    GetCasesScroll() int
    SetCasesScroll(int)
    // ... etc
}
```

### Coords Helper
All map-to-screen coordinate conversion goes through `Coords`:
```go
co := CoordsFromScene(scene)
screenX := co.MapToScreenX(mapX)     // mapX * scale + offsetX
screenY := co.MapToScreenY(mapY)     // mapY * scale + offsetY
mapX := co.ScreenToMapX(screenX)     // (screenX - offsetX) / scale
pixels := co.MapToScreenSize(mapUnits) // mapUnits * scale
sx, sy, sw, sh := co.MapRect(x, y, w, h) // full rectangle
```
**Never** write `obj.X * scaleX + offsetX` directly.

### Assemblage Functions
Factories that create entity groups for a state:
```go
func CreateAppLayoutEntities(reg, cases, selectedCase) {
    CleanStateGroups(reg)  // remove old entities
    reg.Add("case_list", 0, &Entity{Scrollable: &Scrollable{...}})
    reg.Add("avatar", 0, &Entity{Avatar: &Avatar{...}})
    // ...
}
```
Called by StateSystem on state transitions.

## Adding a New Feature

1. **New data needed?** → New component in `components/`
2. **New behavior needed?** → New system in `fsystems/`
3. **New entity type?** → New group constant in `components/groups.go`
4. **Entities created/destroyed?** → Assemblage function in `entities/`
5. **Scene data access?** → New method on `SceneAccessor` + implement on GameScene
6. **Register system** in `GameScene.Init()` at correct execution position

## Anti-Patterns

| Don't | Do |
|-------|-----|
| Logic in components | Logic in systems |
| State in system fields | State in components or scene accessor |
| `obj.X * scaleX + offsetX` | `co.MapToScreenX(obj.X)` |
| `ebiten.CursorPosition()` | `scene.GetCursorPos()` |
| `s.state = StateXxx` | `s.setECSState(c.StateXxx)` |
| System calls another system | Systems communicate through components |
| One massive Update() function | Multiple focused systems |
