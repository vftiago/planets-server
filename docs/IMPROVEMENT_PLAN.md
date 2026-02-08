# Improvement Plan

Findings from a full codebase review. Organized by priority.

## Bugs

### ~~1. Route path parameter not used~~ FIXED

### ~~2. `redirectWithError` missing URL encoding~~ FIXED

Resolved by removing messages entirely — `redirectWithError` now only sends an error code.

### ~~3. Method not allowed returns 400 instead of 405~~ FIXED

### ~~4. `DeleteAllGames` runs outside the transaction~~ FIXED

---

## Security

### ~~5. Rate limiter IP spoofing~~ FIXED

Added `RATE_LIMIT_TRUST_PROXY` config. Proxy headers are only trusted when explicitly enabled. Also fixed `X-Forwarded-For` comma parsing and port stripping from `RemoteAddr`.

### ~~6. Middleware ordering~~ FIXED

Reordered to CORS → Rate Limiter → Router. Refactored to sequential wrapping for clarity. Also renamed `SetupCORS` → `NewCORS` and `Handler` → `Middleware` for consistency with `RateLimiter`.

### ~~7. No request body size limit~~ FIXED

Added `MaxHeaderBytes` (1 MB) to HTTP server config and `http.MaxBytesReader` (1 MB) on the game creation endpoint.

### ~~8. No input validation on universe config~~ DEFERRED

Universe config env vars (`SECTORS_PER_GALAXY`, etc.) will be used as validation caps for the game creation endpoint when the admin dashboard is built.

### ~~9. Missing security headers~~ FIXED

Added `X-Content-Type-Options: nosniff` to all JSON responses via `setCommonHeaders` helper. `Strict-Transport-Security` can be added later when production deployment is configured.

---

## Inconsistencies

### ~~10. JWT/Admin middleware return plain text, not JSON~~ FIXED

Replaced `http.Error()` calls with `response.Error()` using typed errors. Also added missing `errors.Forbidden()` constructor.

### ~~11. Discord requires verified email, Google/GitHub don't~~ FIXED

### ~~12. GitHub user ID is `int`, Google/Discord are `string`~~ FIXED

Resolved together with #14 by introducing an `OAuthProvider` interface and `OAuthUser` struct. Each provider normalizes its API response internally (Discord sets `EmailVerified` from its `Verified` field; GitHub converts `int` ID via `strconv.Itoa`). The handler checks `!userInfo.EmailVerified` uniformly for all providers.

### ~~13. Handler registration mixes `Handle` and `HandleFunc`~~ FIXED

Standardized all route registrations to use `mux.Handle` with `http.HandlerFunc` wrappers.

---

## Code Quality / Duplication

### ~~14. OAuth handlers are ~95% identical~~ FIXED

Introduced `OAuthProvider` interface and `OAuthUser` struct in `internal/auth/providers/provider.go`. Each provider implements the interface, normalizing API-specific quirks internally. Replaced three handler files with a single generic `OAuthHandler` in `internal/auth/handlers/oauth.go`.

### 15. `getExecutor` duplicated across three repos

**Files**: `internal/game/repository.go`, `spatial/repository.go`, `planet/repository.go`

Same method copy-pasted. Could live in `shared/database`.

### 16. Row scanning duplicated in game repository

**File**: `internal/game/repository.go`

Same 12-column scan repeated in `GetGameByID`, `GetAllGames`, etc. A `scanGameRow` helper would reduce duplication.

### 17. Username generation can collide

**File**: `internal/player/service.go:70-75`

`generateUsernameFromEmail("john@a.com")` and `"john@b.com"` both produce `"john"`. The DB UNIQUE constraint on `username` causes an opaque internal error instead of a clear validation message.

---

## Potential Improvements

### 18. No pagination on list endpoints

**Files**: `player/repository.go:28`, `game/repository.go:95`

`GetAllPlayers` and `GetAllGames` load everything into memory.

### 19. Admin email comparison is case-sensitive

**File**: `internal/player/service.go:38`

`email == cfg.Admin.Email` should use `strings.EqualFold` since email addresses are practically case-insensitive.

### 20. Connection string doesn't escape password

**File**: `internal/shared/config/config.go:342-350`

Special characters in DB password (spaces, `=`, `'`) would break the connection string.

### 21. Graceful shutdown reuses `WriteTimeout`

**File**: `cmd/server/main.go:225`

Shutdown timeout should be independent and longer than write timeout.

### ~~22. `CreatePlanetsBatch` returns full objects when only count is used~~ FIXED

Removed `RETURNING` clause and replaced `QueryContext` + row scanning with `ExecContext` + `RowsAffected()`. Repository now returns `int` directly.

---

## Additional Findings

### 23. `DeleteAllGames` on every game creation

**File**: `internal/game/service.go:42-46`

The service deletes all existing games every time a new game is created, guarded only by a TODO comment about "development constraint". This is a data-loss risk if it reaches production.

### 24. Rate limiter cleanup is unreliable

**File**: `internal/middleware/rate_limit.go:53-66`

The cleanup goroutine checks if tokens are at full capacity to detect idle clients. This is not a reliable heuristic — an IP that made one request an hour ago would still have near-full tokens and avoid cleanup. Should use last-access timestamps instead.

### 25. Migration path is relative

**File**: `internal/shared/database/migrations.go`

`filepath.WalkDir("migrations", ...)` uses a relative path. If the binary is run from a different working directory, migrations won't be found. Should be configurable or use `go:embed`.

### 26. Admin email comparison is case-sensitive (promote from #19)

Should be in the Security section — a miscased email in the env var silently makes the admin a regular user. High impact for a trivial fix (`strings.EqualFold`).
