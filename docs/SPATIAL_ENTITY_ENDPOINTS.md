# Spatial Entity Endpoints

## Context

Game creation and deletion work. Now we need read endpoints to browse a game's spatial hierarchy. Currently **no read endpoints exist** for spatial entities or planets — only batch-insert methods for universe generation.

We're introducing a new top-level spatial entity type `universe` (level 0) that connects a game to its spatial hierarchy. This makes the hierarchy fully uniform: every entity has a parent except the universe, and `GET /api/spatial/{id}/children` works for all spatial levels.

These endpoints must be accessible to both admins and regular players who are members of the game. The `game_players` table already exists.

## API Design

Three pure endpoints — each queries exactly one table:

| Method | Endpoint | Description | Returns |
|--------|----------|-------------|---------|
| GET | `/api/spatial/{id}/children` | Child spatial entities | `SpatialEntity[]` |
| GET | `/api/spatial/{id}/ancestors` | Ancestor chain, root-first | `SpatialEntity[]` |
| GET | `/api/spatial/{id}/planets` | Planets in a system | `Planet[]` |

The entry point is the game's `universe_id` field — the frontend fetches the game, gets `universe_id`, then browses from there. The frontend knows when to call `/planets` vs `/children` because the entity it clicked has `entity_type` in its data.

## Changes

### 1. Migration — `migrations/002_add_universe_entity.sql` (new file)
- Add `'universe'` to the `entity_type` enum
- Add `universe_id` column to `games` table (references `spatial_entities`, nullable initially)
- Update CHECK constraints on `spatial_entities`:
  - `CHECK (level >= 0)` (was `> 0`)
  - `CHECK ((level = 0 AND parent_id IS NULL) OR (level > 0 AND parent_id IS NOT NULL))` (was level 1 special case)
- For existing games: insert a universe entity per game, update galaxies' `parent_id`, set `games.universe_id`
- Make `universe_id` NOT NULL after backfill

### 2. Spatial model — `internal/spatial/models.go`
- Add `EntityTypeUniverse EntityType = "universe"`
- Add `EntityTypeUniverse: 0` to `EntityLevels`
- Add `Universe = SpatialEntity` type alias

### 3. Game model — `internal/game/models.go`
- Add `UniverseID int` field to `Game` struct with `json:"universe_id"`

### 4. Game repository — `internal/game/repository.go`
- Update `CreateGame`, `GetGameByID`, `GetAllGames` queries to include `universe_id`
- Add `SetUniverseID(ctx, gameID, universeID, tx)` method

### 5. Game service — `internal/game/service.go`
- Update `generateUniverse` to:
  1. Create a single universe entity (level 0, parent_id NULL, game_id, coords 0,0)
  2. Use its ID as parent for galaxies (instead of `nil`)
  3. Call `SetUniverseID` to link game → universe
- Update `BuildGenerationPlan` or generation loop accordingly

### 6. Spatial repository — `internal/spatial/repository.go`
Add two methods:
- `GetChildren(ctx, parentID)` — `WHERE parent_id = $1 ORDER BY x_coord, y_coord`
- `GetAncestors(ctx, entityID)` — recursive CTE walking `parent_id` up to root, `ORDER BY level ASC`

### 7. Planet repository — `internal/planet/repository.go`
Add one method:
- `GetBySystemID(ctx, systemID)` — `WHERE system_id = $1 ORDER BY planet_index`

### 8. Spatial service — `internal/spatial/service.go`
Add pass-through methods: `GetChildren`, `GetAncestors`

### 9. Planet service — `internal/planet/service.go`
Add pass-through method: `GetBySystemID`

### 10. Game access middleware — `internal/middleware/game_access.go` (new file)
`GameAccessMiddleware` struct with `db *database.DB`:
- `Require(next http.Handler) http.Handler` method
- Wraps `JWTMiddleware`
- Reads `{id}` path value as spatial entity ID
- Looks up `game_id` from `spatial_entities` table: `SELECT game_id FROM spatial_entities WHERE id = $1`
- Gets claims from context via `GetUserFromContext`
- If `claims.Role == "admin"` → pass through
- Otherwise, checks `game_players` for `(game_id, player_id)` match
- If no match → `errors.Forbidden("game access required")`

### 11. Spatial handler — `internal/spatial/handlers/spatial.go` (new file)
`SpatialHandler` struct with `spatialService`. Two methods:
- `GetChildren` — parses `{id}`, calls `spatialService.GetChildren`, returns `SpatialEntity[]`
- `GetAncestors` — parses `{id}`, calls `spatialService.GetAncestors`, returns `SpatialEntity[]`

### 12. Planet handler — `internal/planet/handlers/planet.go` (new file)
`PlanetHandler` struct with `planetService`. One method:
- `GetBySystemID` — parses `{id}`, calls `planetService.GetBySystemID`, returns `Planet[]`

### 13. Routes — `internal/server/routes.go`
- Add `spatialService *spatial.Service` and `planetService *planet.Service` to `Routes` struct and `NewRoutes`
- Create `GameAccessMiddleware` with DB reference
- Register three routes:
  ```
  mux.Handle("/api/spatial/{id}/children", gameAccess.Require(http.HandlerFunc(spatialHandler.GetChildren)))
  mux.Handle("/api/spatial/{id}/ancestors", gameAccess.Require(http.HandlerFunc(spatialHandler.GetAncestors)))
  mux.Handle("/api/spatial/{id}/planets", gameAccess.Require(http.HandlerFunc(planetHandler.GetBySystemID)))
  ```

### 14. Wire services — `cmd/server/main.go`
Pass `spatialService` and `planetService` to `server.NewRoutes()` (already instantiated on lines 80-81).

## Verification
1. Run migration, start server: `go run cmd/server/main.go`
2. Create a game — verify it gets a `universe_id`
3. `GET /api/spatial/{universeId}/children` → returns galaxies
4. `GET /api/spatial/{galaxyId}/children` → returns sectors
5. `GET /api/spatial/{sectorId}/children` → returns systems
6. `GET /api/spatial/{systemId}/planets` → returns planets
7. `GET /api/spatial/{systemId}/ancestors` → returns [universe, galaxy, sector, system]
8. `go test ./...`
