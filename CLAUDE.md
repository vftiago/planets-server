# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

**Run the server:**

```bash
go run cmd/server/main.go
```

**Build the server:**

```bash
go build -o planets-server cmd/server/main.go
```

**Run tests:**

```bash
go test ./...
```

**Format code:**

```bash
go fmt ./...
```

**Lint and vet:**

```bash
go vet ./...
```

**Generate secure JWT secret (for setup):**

```bash
openssl rand -hex 32
```

## Architecture Overview

This is a Go web server for the "Planets!" turn-based space strategy game with the following key architectural components:

### Core Structure

- **Entry Point**: `cmd/server/main.go` - Main application with graceful shutdown
- **Internal Package**: All business logic under `internal/` to prevent external imports
- **Shared Components**: Common utilities in `internal/shared/` (config, database, errors, response, logger, etc.)

```text
internal/
  ├── auth/                     # Authentication domain
  │   ├── handlers/
  │   │   ├── google.go         # Google OAuth handler
  │   │   ├── github.go         # GitHub OAuth handler
  │   │   ├── discord.go        # Discord OAuth handler
  │   │   ├── logout.go         # Logout handler
  │   │   └── utils.go          # Handler utilities (redirectWithError)
  │   ├── providers/
  │   │   ├── google.go         # Google OAuth provider
  │   │   ├── github.go         # GitHub OAuth provider
  │   │   └── discord.go        # Discord OAuth provider
  │   ├── models.go             # Auth-specific models (Claims, etc.)
  │   ├── repository.go         # Auth-specific database operations
  │   ├── service.go            # Auth business logic
  │   ├── jwt.go                # JWT creation/validation
  │   ├── oauth.go              # OAuth orchestration
  │   └── state.go              # OAuth state parameter handling
  ├── game/                     # Game domain
  │   ├── handlers/
  │   │   ├── game.go           # Game CRUD endpoints
  │   │   └── status.go         # Game status endpoint
  │   ├── models.go             # Game, GameConfig structs
  │   ├── repository.go         # Game database operations
  │   └── service.go            # Game business logic, universe generation
  ├── spatial/                  # Unified spatial hierarchy (galaxy, sector, system)
  │   ├── models.go             # SpatialEntity base type + Galaxy, Sector, System aliases
  │   ├── repository.go         # Spatial entity database operations
  │   └── service.go            # Spatial entity business logic
  ├── planet/                   # Planet domain
  │   ├── models.go             # Planet struct with types and enums
  │   ├── repository.go         # Planet database operations
  │   └── service.go            # Planet business logic
  ├── player/                   # Player domain
  │   ├── handlers/
  │   │   ├── me.go             # Current user profile endpoint
  │   │   └── players.go        # Player list endpoint
  │   ├── models.go             # Player, PlayerAuthProvider structs
  │   ├── repository.go         # Player database operations
  │   └── service.go            # Player business logic
  ├── middleware/               # HTTP middleware
  │   ├── auth.go               # JWT authentication
  │   ├── admin.go              # Admin authorization
  │   ├── cors.go               # CORS handling
  │   └── rate_limit.go         # Token bucket rate limiting
  ├── server/                   # HTTP server setup
  │   ├── handlers/
  │   │   └── health.go         # Health check endpoint
  │   └── routes.go             # Route definitions
  └── shared/                   # Common utilities and infrastructure
      ├── config/
      │   └── config.go         # Configuration management
      ├── cookies/
      │   └── cookies.go        # Cookie helpers
      ├── database/
      │   ├── connection.go     # Database connection pooling
      │   └── migrations.go     # Migration execution
      ├── errors/
      │   └── errors.go         # Custom error types (NotFound, Validation, etc.)
      ├── response/
      │   └── error.go          # HTTP error/success response helpers
      ├── logger/
      │   └── logger.go         # slog-based logging setup
      ├── redis/
      │   └── connection.go     # Redis connection
      └── utils/
          └── env.go            # Environment variable utilities
```

### Service Layer Pattern

All domain services follow a consistent pattern:

- **Constructor**: `NewService(repo *Repository)` - no logger injection
- **Repository Delegation**: Direct pass-through methods like `GetByID`, `Create`, etc.
- **Business Logic**: Complex operations that coordinate multiple repositories
- **Error Handling**: Returns custom error types from `internal/shared/errors`

### Error Handling Pattern

The codebase uses a consistent error handling approach:

1. **Custom Error Types** (`internal/shared/errors/errors.go`):
   - `NotFoundf()`, `Validation()`, `Conflictf()`, `Unauthorized()`, `External()`
   - `WrapInternal()`, `WrapValidation()`, `WrapExternal()` for wrapping errors

2. **Response Helpers** (`internal/shared/response/error.go`):
   - `response.Error(w, r, logger, err)` - logs and sends JSON error response
   - `response.Success(w, statusCode, data)` - sends JSON success response

3. **Error Flow**:
   - Repositories return custom error types (no logging)
   - Services pass through or wrap errors (no logging)
   - Handlers log errors at the boundary using `response.Error()`

### Authentication System

- **JWT-based**: Cookie-based JWT authentication with configurable expiration
- **OAuth Providers**: Google and GitHub OAuth with automatic account linking
- **State Validation**: OAuth state parameter validation for CSRF protection
- **Account Linking**: Users can link multiple OAuth providers to same account
- **Admin System**: Role-based authorization with admin middleware
- **Middleware**: `JWTMiddleware` and `RequireAdmin` protect endpoints

### Database Layer

- **PostgreSQL**: Primary database with connection pooling
- **Migration System**: File-based migrations in `migrations/` directory with automatic execution
- **Repository Pattern**: Domain-specific repositories for data access abstraction
- **Transaction Support**: Database operations wrapped in transactions where needed
- **Custom Errors**: Repositories return typed errors (`errors.NotFoundf()`, `errors.WrapInternal()`)

### Configuration Management

- **Environment-based**: Uses `.env` file and environment variables
- **Global Config**: Centralized configuration through `internal/shared/config`
- **Validation**: Configuration validation on startup with required field checks
- **Environment Detection**: Automatic production/development mode detection
- **Admin Configuration**: Admin user setup via environment variables

### Game Universe Architecture

The game uses a hierarchical spatial structure managed by the unified `spatial` package:

- **Galaxy**: Top-level container, contains sectors arranged in a grid pattern
- **Sector**: Contains systems arranged in a grid pattern
- **System**: Contains planets with random generation
- **Planet**: Individual game objects with types, sizes, and populations

The `spatial` package uses a single `spatial_entities` table with an `entity_type` column to distinguish between galaxies, sectors, and systems. Type aliases (`Galaxy`, `Sector`, `System`) provide semantic clarity in code.

Universe generation is orchestrated by the game service, which coordinates the spatial and planet services to create the complete game world.

### HTTP Layer

- **Standard Library**: Uses `net/http` with `http.ServeMux`
- **CORS Support**: Configurable CORS middleware for frontend integration
- **Rate Limiting**: Token bucket rate limiting with configurable limits
- **Structured Logging**: `slog`-based logging with contextual information
- **Route Organization**: Routes defined in `internal/server/routes.go`
- **Middleware Stack**: CORS → Rate Limiting → Authentication → Admin → Handler
- **JSON Responses**: All endpoints use `response.Success()` and `response.Error()` for consistent JSON formatting

### Key Components

**Models**: Domain-specific structs with JSON tags for API serialization

**Handlers**: HTTP endpoint handlers using `response.Error()` and `response.Success()`

**Middleware**:

- JWT authentication middleware with context injection
- Admin role authorization middleware
- CORS middleware with environment-based configuration
- Rate limiting middleware with token bucket algorithm

**Auth System**:

- OAuth provider abstraction with Google and GitHub implementations
- JWT creation/validation with configurable claims
- Provider-specific user data mapping
- Admin role assignment via configuration

## Database Schema

The system uses multiple tables organized by domain:

- **Players**: `players`, `player_auth_providers` - User accounts with OAuth linking
- **Games**: `games` - Game instances with turn management
- **Spatial**: `spatial_entities` - Unified table for galaxies, sectors, and systems with `entity_type` discriminator
- **Planets**: `planets` - Individual planets linked to systems

## Environment Configuration

Key environment variables (see `internal/shared/config/config.go` for complete list):

- `JWT_SECRET`: Required, minimum 32 characters
- `DB_*`: Database connection parameters
- `GOOGLE_CLIENT_ID/SECRET`: Google OAuth credentials
- `GITHUB_CLIENT_ID/SECRET`: GitHub OAuth credentials
- `FRONTEND_ADMIN_URL`: Admin dashboard URL for CORS
- `FRONTEND_CLIENT_URL`: Player client URL for CORS
- `SERVER_URL`: Server URL for OAuth redirects
- `ADMIN_EMAIL`: Admin user email for role assignment

## Development Notes

- OAuth providers are optional - server runs without OAuth credentials but warns in logs
- Database migrations run automatically on startup
- Configuration validation prevents startup with missing required settings
- Graceful shutdown handling with configurable timeouts
- Structured logging with `slog` at handler boundaries only
- Clean separation between public, authenticated, and admin API endpoints
- Error handling follows "log at boundaries" pattern - repositories and services don't log
- All HTTP responses use the centralized `response` package for consistency
