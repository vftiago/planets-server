# Game/Universe Merge Plan

## Current State

"Universe" is not a separate entity. There's no `Universe` table or struct. It exists only as:

1. **Two cosmetic fields on the `games` table** — `universe_name` and `universe_description` (unused flavor text)
2. **`UniverseConfig` struct** — generation parameters (galaxy count, sectors per galaxy, etc.) used once at game creation and then discarded

The actual spatial hierarchy (galaxies > sectors > systems > planets) lives in `spatial_entities` and `planets` tables, not in any universe table.

## What Game Actually Does

Tracks the match/session state:
- `Status` (creating/active/paused/completed)
- `CurrentTurn`, `NextTurnAt`, `TurnIntervalHours` (turn progression)
- `MaxPlayers`, `PlanetCount` (game rules/limits)

## What Universe Actually Does

Nothing at runtime. `UniverseConfig` is a bag of parameters for the world generation algorithm:
- `GalaxyCount`
- `SectorsPerGalaxy`
- `SystemsPerSector`
- `MinPlanetsPerSystem`
- `MaxPlanetsPerSystem`

These are consumed by `Service.generateUniverse()` during `CreateGame` and never referenced again.

## Proposed Changes

### Remove from `Game` struct and `games` table
- `description` — unused
- `universe_name` — unused flavor text
- `universe_description` — unused flavor text

### Remove from `GameConfig`
- `description`
- `universe_name`
- `universe_description`

### Keep `UniverseConfig` as-is (for now)
It's already a separate parameter to `CreateGame(ctx, gameConfig, universeConfig)`. It doesn't need to be merged into `GameConfig` — it serves a different purpose (generation params vs game rules). It could be renamed to `GenerationConfig` or similar to clarify that it's not a persistent entity.

### Remove `description` from spatial entities
- `SpatialEntity.Description` in `spatial/models.go` — always empty string
- `BatchInsertRequest.Description` in `spatial/repository.go` — always empty string
- `description` column in `spatial_entities` table

### Files affected

**Server (planets-server):**
- `internal/game/models.go` — remove fields from `Game`, `GameConfig`
- `internal/game/repository.go` — update 3 SQL queries (CreateGame, GetGameByID, GetAllGames) and all Scan/param lists
- `internal/game/handlers/game.go` — remove `UniverseName` default in CreateGame handler
- `internal/spatial/models.go` — remove `Description` from `SpatialEntity`
- `internal/spatial/repository.go` — remove `Description` from `BatchInsertRequest` and SQL insert
- `internal/spatial/service.go` — remove `Description: ""` from entity generation
- `migrations/` — new migration to drop columns

**Frontend (planets-admin):**
- `src/api/games.ts` — remove `description` and `universe_name` from `Game` type
