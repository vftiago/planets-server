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
- **Shared Components**: Common utilities in `internal/shared/` (config, database, logger, utils)

```text
internal/
  ├── auth/                     # Authentication domain
  │   ├── handlers/
  │   ├── providers/
  │   ├── models.go             # Auth-specific models (Claims, etc.)
  │   ├── repository.go         # Auth-specific database operations
  │   ├── service.go            # Auth business logic
  │   ├── jwt.go
  │   ├── oauth.go
  │   └── state.go
  ├── handlers/                 # HTTP handlers
  │   ├── health.go             # Health check handler
  │   ├── logout.go             # Logout handler
  │   └── status.go             # Game status handler
  ├── middleware/               # HTTP middleware
  │   ├── auth.go
  │   ├── cors.go
  │   └── rate_limit.go
  ├── player/                   # Player domain
  │   ├── models.go             # Player, PlayerAuthProvider structs
  │   ├── repository.go         # Player database operations
  │   ├── service.go            # Player business logic
  │   ├── handlers.go           # Player HTTP handlers
  │   └── errors.go             # Player-specific errors
  ├── server/                   # HTTP server setup
  │   ├── routes.go             # Route definitions
  │   ├── server.go             # HTTP server setup and start
  │   └── middleware.go         # Middleware setup
  ├── game/                     # Game domain (future)
  │   ├── models.go             # Game, Turn, Fleet structs
  │   ├── repository.go         # Game database operations
  │   ├── service.go            # Game business logic (turn processing, etc.)
  │   ├── handlers.go           # Game HTTP handlers
  │   ├── engine/               # Game engine components
  │   │   ├── combat.go
  │   │   ├── movement.go
  │   │   ├── production.go
  │   │   └── turn_processor.go
  │   └── events/               # Game events system
  │       ├── publisher.go
  │       └── handlers.go
  ├── shared/                   # Common utilities and infrastructure
  │   └── config/               # Configuration management
  │   │   └── config.go
  │   ├── cookies/
  │   │   └── cookies.go        # Cookie helpers
  │   ├── database/             # Database connection, migrations
  │   │   ├── connection.go
  │   │   ├── migrations.go
  │   │   └── transaction.go    # Transaction helpers
  |   ├── logger/               # Logging setup
  │   │   └── logger.go
  │   ├── utils/                # Generic utilities
  │   │   ├── env.go
  │   │   ├── validation.go
  │   │   └── crypto.go
  │   ├── errors/               # Common error types and handling (future)
  │   │   ├── types.go
  │   │   └── handler.go
  └── api/                      # API layer (future - for API versioning)
      └── v1/
          ├── auth.go
          ├── player.go
          └── game.go
```

### Authentication System

- **JWT-based**: Cookie-based JWT authentication with configurable expiration
- **OAuth Providers**: Google and GitHub OAuth with automatic account linking
- **State Validation**: OAuth state parameter validation for CSRF protection
- **Account Linking**: Users can link multiple OAuth providers to same account
- **Middleware**: `JWTMiddleware` protects authenticated endpoints

### Database Layer

- **PostgreSQL**: Primary database with connection pooling
- **Migration System**: File-based migrations in `migrations/` directory with automatic execution
- **Repository Pattern**: `PlayerRepository` for data access abstraction
- **Transaction Support**: Database operations wrapped in transactions where needed

### Configuration Management

- **Environment-based**: Uses `.env` file and environment variables
- **Global Config**: Centralized configuration through `internal/shared/config`
- **Validation**: Configuration validation on startup with required field checks
- **Environment Detection**: Automatic production/development mode detection

### HTTP Layer

- **Standard Library**: Uses `net/http` with `http.ServeMux`
- **CORS Support**: Configurable CORS middleware for frontend integration
- **Structured Logging**: `slog`-based logging with contextual information
- **Route Organization**: Routes defined in `internal/server/routes.go`

### Key Components

**Models (`internal/models/`)**:

- `Player`: Core user entity with OAuth provider linking
- `PlayerRepository`: Database operations with automatic OAuth account linking

**Handlers (`internal/handlers/`)**:

- Health check, game status, player management endpoints
- OAuth handlers separated by provider in `internal/auth/handlers/`

**Middleware (`internal/middleware/`)**:

- JWT authentication middleware with context injection
- CORS middleware with environment-based configuration

**Auth System (`internal/auth/`)**:

- OAuth provider abstraction with Google and GitHub implementations
- JWT creation/validation with configurable claims
- Provider-specific user data mapping

## Database Schema

The system uses two main tables:

- `players`: Core user accounts (id, username, email, display_name, avatar_url)
- `player_auth_providers`: OAuth provider linkages (supports multiple providers per user)

## Environment Configuration

Key environment variables (see `internal/shared/config/config.go` for complete list):

- `JWT_SECRET`: Required, minimum 32 characters
- `DB_*`: Database connection parameters
- `GOOGLE_CLIENT_ID/SECRET`: Google OAuth credentials
- `GITHUB_CLIENT_ID/SECRET`: GitHub OAuth credentials
- `FRONTEND_URL`: Frontend URL for CORS
- `SERVER_URL`: Server URL for OAuth redirects

## Development Notes

- OAuth providers are optional - server runs without OAuth credentials but warns in logs
- Database migrations run automatically on startup
- Configuration validation prevents startup with missing required settings
- Graceful shutdown handling with configurable timeouts
- Structured logging with component-based context throughout
- Clean separation between public and protected API endpoints
