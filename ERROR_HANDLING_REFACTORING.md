# Error Handling Refactoring Guide

## Overview

This document describes the new consistent error handling pattern implemented across the codebase. The refactoring follows these key principles:

1. **Log only at boundaries** (HTTP handlers, cron jobs)
2. **Wrap errors with context** as they bubble up through layers
3. **Use custom error types** to categorize different error scenarios
4. **Map errors to HTTP codes** at the handler level
5. **Return actionable messages** to clients

## Architecture

### Error Types (`internal/shared/errors/errors.go`)

Custom error types with helper functions:

- **NotFound**: Resource not found (404)
- **Validation**: Invalid input data (400)
- **Conflict**: Conflict with existing data (409)
- **Unauthorized**: Authentication failure (401)
- **Forbidden**: Insufficient permissions (403)
- **Internal**: Internal server error (500)
- **External**: External service error (503)

### HTTP Response Helper (`internal/shared/response/error.go`)

Centralized error logging and response generation:

```go
// Log error and send JSON response
response.Error(w, r, logger, err)

// Log with custom client message
response.ErrorWithMessage(w, r, logger, err, "User-friendly message")

// Send success response
response.Success(w, http.StatusOK, data)
```

## Refactoring Pattern

### Repository Layer

**Remove:**
- All logger initialization and calls
- The logger field from repository structs
- `nil, nil` returns for not found cases

**Add:**
- Custom error types for different scenarios
- `errors.NotFoundf()` instead of `nil, nil`
- `errors.WrapInternal()` for database errors
- `errors.Validation()` for invalid input

**Example:**

```go
// BEFORE
func (r *Repository) GetByID(ctx context.Context, id int) (*Entity, error) {
    logger := slog.With("component", "repo", "operation", "get_by_id")
    logger.Debug("Getting entity by ID")

    var entity Entity
    err := r.db.QueryRowContext(ctx, query, id).Scan(&entity.ID, &entity.Name)
    if err != nil {
        if err == sql.ErrNoRows {
            logger.Debug("Entity not found")
            return nil, nil
        }
        logger.Error("Database error", "error", err)
        return nil, fmt.Errorf("database error: %w", err)
    }

    logger.Debug("Entity retrieved", "id", entity.ID)
    return &entity, nil
}

// AFTER
func (r *Repository) GetByID(ctx context.Context, id int) (*Entity, error) {
    var entity Entity
    err := r.db.QueryRowContext(ctx, query, id).Scan(&entity.ID, &entity.Name)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, errors.NotFoundf("entity not found with id: %d", id)
        }
        return nil, errors.WrapInternal("failed to get entity by id", err)
    }
    return &entity, nil
}
```

### Service Layer

**Remove:**
- All logger initialization and calls (except in rare debug scenarios)
- The logger field from service structs
- Logger parameter from NewService constructors

**Keep:**
- Error wrapping with additional business context
- Use `fmt.Errorf("context: %w", err)` or custom error types

**Example:**

```go
// BEFORE
func (s *Service) CreateEntity(ctx context.Context, name string) (*Entity, error) {
    logger := s.logger.With("component", "service", "operation", "create")
    logger.Info("Creating entity", "name", name)

    entity, err := s.repo.Create(ctx, name)
    if err != nil {
        logger.Error("Failed to create entity", "error", err)
        return nil, fmt.Errorf("failed to create entity: %w", err)
    }

    logger.Info("Entity created successfully", "id", entity.ID)
    return entity, nil
}

// AFTER
func (s *Service) CreateEntity(ctx context.Context, name string) (*Entity, error) {
    entity, err := s.repo.Create(ctx, name)
    if err != nil {
        return nil, fmt.Errorf("failed to create entity: %w", err)
    }
    return entity, nil
}
```

### Handler Layer

**Replace:**
- `http.Error()` calls with `response.Error()`
- Manual JSON encoding with `response.Success()`
- `nil` checks with proper error type checking
- Multiple logger calls with single logger at handler level

**Add:**
- Custom error types for validation failures
- `response.Error()` for all error responses
- `response.Success()` for all success responses

**Example:**

```go
// BEFORE
func (h *Handler) GetEntity(w http.ResponseWriter, r *http.Request) {
    logger := slog.With("handler", "get_entity", "remote_addr", r.RemoteAddr)
    logger.Debug("Entity requested")

    idStr := r.URL.Query().Get("id")
    if idStr == "" {
        logger.Error("Missing ID parameter")
        http.Error(w, "ID is required", http.StatusBadRequest)
        return
    }

    id, err := strconv.Atoi(idStr)
    if err != nil {
        logger.Error("Invalid ID", "error", err)
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    entity, err := h.service.GetByID(ctx, id)
    if err != nil {
        logger.Error("Failed to get entity", "error", err)
        http.Error(w, "Failed to get entity", http.StatusInternalServerError)
        return
    }

    if entity == nil {
        logger.Debug("Entity not found")
        http.Error(w, "Entity not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(entity); err != nil {
        logger.Error("Failed to encode response", "error", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    logger.Debug("Entity retrieved successfully")
}

// AFTER
func (h *Handler) GetEntity(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    logger := slog.With("handler", "get_entity")

    idStr := r.URL.Query().Get("id")
    if idStr == "" {
        response.Error(w, r, logger, errors.Validation("ID is required"))
        return
    }

    id, err := strconv.Atoi(idStr)
    if err != nil {
        response.Error(w, r, logger, errors.WrapValidation("invalid ID format", err))
        return
    }

    entity, err := h.service.GetByID(ctx, id)
    if err != nil {
        response.Error(w, r, logger, err)
        return
    }

    response.Success(w, http.StatusOK, entity)
}
```

## Completed Refactoring

### ✅ Infrastructure
- `internal/shared/errors/errors.go` - Custom error types
- `internal/shared/response/error.go` - HTTP error response helper

### ✅ Repositories
- `internal/player/repository.go`
- `internal/auth/repository.go`
- `internal/game/repository.go`

### ✅ Services
- `internal/game/service.go`

### ✅ Handlers
- `internal/game/handlers/game.go`

### ✅ Main
- `cmd/server/main.go` - Updated service initialization

## Remaining Work

### Repositories to Refactor
Apply the repository pattern to:
- `internal/spatial/repository.go`
- `internal/planet/repository.go`
- `internal/universe/repository.go` (if exists)

**Steps:**
1. Remove logger field and all logger calls
2. Replace `nil, nil` returns with `errors.NotFoundf()`
3. Replace `fmt.Errorf()` with `errors.WrapInternal()`
4. Use `errors.Validation()` for invalid input

### Services to Refactor
Apply the service pattern to:
- `internal/player/service.go`
- `internal/auth/service.go`
- `internal/spatial/service.go`
- `internal/planet/service.go`

**Steps:**
1. Remove logger field from struct
2. Remove logger parameter from NewService
3. Remove all logger calls
4. Keep error wrapping with business context
5. Update all calling code to remove logger parameter

### Handlers to Refactor
Apply the handler pattern to:
- `internal/player/handlers/*.go`
- `internal/auth/handlers/*.go`
- `internal/game/handlers/status.go`
- `internal/server/handlers/*.go`

**Steps:**
1. Replace `http.Error()` with `response.Error()`
2. Replace JSON encoding with `response.Success()`
3. Create validation errors with `errors.Validation()`
4. Remove `nil` checking (use error types instead)
5. Keep only one logger declaration per handler

## Testing

After refactoring each component:

1. **Build test:**
   ```bash
   go build ./...
   ```

2. **Run tests:**
   ```bash
   go test ./...
   ```

3. **Manual testing:**
   - Start the server
   - Test API endpoints
   - Verify error responses are JSON formatted
   - Check logs contain proper context

## Benefits

1. **Consistency**: All errors handled the same way
2. **Observability**: Errors logged once at boundaries with full context
3. **Client Experience**: Consistent JSON error responses
4. **Maintainability**: Clear error flow through layers
5. **Type Safety**: Error types help distinguish scenarios
6. **Reduced Noise**: No duplicate logging at multiple layers

## Error Response Format

All HTTP errors now return consistent JSON:

```json
{
  "error": "not_found",
  "message": "game not found with id: 123",
  "code": 404
}
```

Error types map to HTTP status codes:
- `not_found` → 404
- `validation` → 400
- `conflict` → 409
- `unauthorized` → 401
- `forbidden` → 403
- `external` → 503
- `internal` → 500
