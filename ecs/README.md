# ecs

Import path: `github.com/InsideGallery/core/ecs`

## Overview

`ecs` provides lightweight Entity-Component-System primitives: entity IDs, entity versioning, and small update
interfaces for components and systems.

## Main APIs

- `Registry` owns isolated entity ID generation state.
- `NewRegistry`, `(*Registry).NewBaseEntity`, `(*Registry).NewBaseEntityWithID`, and `(*Registry).LatestID`
  create and inspect registry-scoped entities.
- `BaseEntity` stores an ID and version with `GetID`, `SetID`, `GetVersion`, `SetVersion`, and `UpVersion`.
- `Entity`, `Versionable`, `Component`, and `System` are small contracts for consumers.
- `InstallDefaultEntityFactory` temporarily replaces the package-level compatibility registry.
- `EntityFactory`, `NewEntityFactory`, `DefaultEntityFactory`, `NewBaseEntity`, and `NewBaseEntityWithID` remain
  for backward compatibility.

## Usage

```go
registry := ecs.NewRegistry()

entity := registry.NewBaseEntity()
entity.UpVersion()

id := entity.GetID()
version := entity.GetVersion()

_ = id
_ = version
```

## Notes

Prefer `NewRegistry` for new code so ID state is scoped to the caller. The package-level helpers use the global
default registry for compatibility. `BaseEntity` version operations use atomics; `UpVersion` follows `uint64`
overflow behavior.
