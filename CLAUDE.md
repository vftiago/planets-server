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
- **Shared Components**: Common utilities in `internal/shared/` (config, database, logger)

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
- `BASE_URL`: Server base URL for OAuth redirects

## Development Notes

- OAuth providers are optional - server runs without OAuth credentials but warns in logs
- Database migrations run automatically on startup
- Configuration validation prevents startup with missing required settings
- Graceful shutdown handling with configurable timeouts
- Structured logging with component-based context throughout
- Clean separation between public and protected API endpoints