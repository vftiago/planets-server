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
  ├── game/                     # Game domain
  │   ├── handlers/
  │   ├── models.go             # Game, GameConfig, GameStats structs
  │   ├── repository.go         # Game database operations
  │   └── service.go            # Game business logic
  ├── universe/                 # Universe generation and management
  │   ├── handlers/
  │   ├── models.go             # Universe, UniverseConfig, GenerationProgress structs
  │   ├── repository.go         # Universe database operations
  │   └── service.go            # Universe generation orchestration
  ├── galaxy/                   # Galaxy domain
  │   ├── models.go             # Galaxy struct
  │   ├── repository.go         # Galaxy database operations
  │   └── service.go            # Galaxy business logic
  ├── sector/                   # Sector domain
  │   ├── models.go             # Sector struct
  │   ├── repository.go         # Sector database operations
  │   └── service.go            # Sector business logic
  ├── system/                   # System domain
  │   ├── models.go             # System struct
  │   ├── repository.go         # System database operations
  │   └── service.go            # System business logic
  ├── planet/                   # Planet domain
  │   ├── models.go             # Planet struct with types and enums
  │   ├── repository.go         # Planet database operations
  │   └── service.go            # Planet business logic
  ├── middleware/               # HTTP middleware
  │   ├── auth.go
  │   ├── admin.go
  │   ├── cors.go
  │   └── rate_limit.go
  ├── player/                   # Player domain
  │   ├── handlers/
  │   │   ├── me.go
  │   │   └── players.go
  │   ├── errors.go             # Player-specific errors
  │   ├── models.go             # Player, PlayerAuthProvider structs
  │   ├── repository.go         # Player database operations
  │   └── service.go            # Player business logic
  ├── server/                   # HTTP server setup
  │   ├── handlers/
  │   │   └── health.go
  │   └── routes.go             # Route definitions
  └── shared/                   # Common utilities and infrastructure
      ├── config/               # Configuration management
      │   └── config.go
      ├── cookies/
      │   └── cookies.go        # Cookie helpers
      ├── database/             # Database connection, migrations
      │   ├── connection.go
      │   └── migrations.go
      ├── logger/               # Logging setup
      │   └── logger.go
      └── utils/                # Generic utilities
          └── env.go
```

### Service Layer Pattern

All domain services follow a consistent pattern with logger injection and repository delegation:

- **Constructor**: `NewService(repo *Repository, logger *slog.Logger)` with component-specific logging
- **Repository Delegation**: Direct pass-through methods like `GetByID`, `Create`, etc.
- **Business Logic**: Complex operations with structured logging using service logger
- **Error Handling**: Contextual error logging and wrapping

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
- **Structured Logging**: All repositories have logger injection for consistent logging

### Configuration Management

- **Environment-based**: Uses `.env` file and environment variables
- **Global Config**: Centralized configuration through `internal/shared/config`
- **Validation**: Configuration validation on startup with required field checks
- **Environment Detection**: Automatic production/development mode detection
- **Admin Configuration**: Admin user setup via environment variables

### Game Universe Architecture

The game uses a hierarchical spatial structure:

- **Universe**: Root container, defines overall game space parameters
- **Galaxy**: Contains sectors arranged in a grid pattern
- **Sector**: Contains systems arranged in a grid pattern
- **System**: Contains planets with random generation
- **Planet**: Individual game objects with types, sizes, and populations

Universe generation is orchestrated by the universe service, which coordinates galaxy, sector, system, and planet services to create the complete game world.

### HTTP Layer

- **Standard Library**: Uses `net/http` with `http.ServeMux`
- **CORS Support**: Configurable CORS middleware for frontend integration
- **Rate Limiting**: Token bucket rate limiting with configurable limits
- **Structured Logging**: `slog`-based logging with contextual information
- **Route Organization**: Routes defined in `internal/server/routes.go`
- **Middleware Stack**: CORS → Rate Limiting → Authentication → Admin → Handler

### Key Components

**Models**: Domain-specific structs with JSON tags for API serialization

**Handlers**: HTTP endpoint handlers with request/response handling

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
- **Universe**: `universes`, `galaxies`, `sectors`, `systems`, `planets` - Hierarchical spatial data

## Environment Configuration

Key environment variables (see `internal/shared/config/config.go` for complete list):

- `JWT_SECRET`: Required, minimum 32 characters
- `DB_*`: Database connection parameters
- `GOOGLE_CLIENT_ID/SECRET`: Google OAuth credentials
- `GITHUB_CLIENT_ID/SECRET`: GitHub OAuth credentials
- `FRONTEND_URL`: Frontend URL for CORS
- `SERVER_URL`: Server URL for OAuth redirects
- `ADMIN_EMAIL`: Admin user email for role assignment
- `UNIVERSE_*`: Universe generation parameters

## Development Notes

- OAuth providers are optional - server runs without OAuth credentials but warns in logs
- Database migrations run automatically on startup
- Configuration validation prevents startup with missing required settings
- Graceful shutdown handling with configurable timeouts
- Structured logging with component-based context throughout
- Clean separation between public, authenticated, and admin API endpoints
- Universe generation uses configurable parameters for procedural content creation
- All services follow the same constructor and logging patterns for consistency
