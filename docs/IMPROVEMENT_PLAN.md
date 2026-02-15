# Improvement Plan

Findings from a full codebase review. Organized by priority.

## Code Quality / Duplication

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

---

## Additional Findings

### 24. Rate limiter cleanup is unreliable

**File**: `internal/middleware/rate_limit.go:53-66`

The cleanup goroutine checks if tokens are at full capacity to detect idle clients. This is not a reliable heuristic — an IP that made one request an hour ago would still have near-full tokens and avoid cleanup. Should use last-access timestamps instead.

### 25. Migration path is relative

**File**: `internal/shared/database/migrations.go`

`filepath.WalkDir("migrations", ...)` uses a relative path. If the binary is run from a different working directory, migrations won't be found. Should be configurable or use `go:embed`.

### 26. Admin email comparison is case-sensitive (promote from #19)

Should be in the Security section — a miscased email in the env var silently makes the admin a regular user. High impact for a trivial fix (`strings.EqualFold`).
