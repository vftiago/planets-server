# Planets! Game Server

A turn-based space strategy game server built with Go.

## Quick Start

### 1. Start PostgreSQL

```bash
sudo service postgresql start
```

### 2. Start Redis

```bash
sudo service redis-server start
```

Or disable Redis by setting `REDIS_ENABLED=false` in your `.env`.

### 3. Configure environment

Create a `.env` file in the project root. The server reads it automatically on startup.

#### Admin Configuration

```bash
ADMIN_DISPLAY_NAME=Admin
ADMIN_EMAIL=admin@localhost          # The first user to log in with this email gets the admin role
ADMIN_USERNAME=admin
```

#### Database Configuration

```bash
DB_HOST=localhost
DB_NAME=planets
DB_PASSWORD=
DB_PORT=5432
DB_SSLMODE=disable
DB_USER=postgres
```

#### Environment

```bash
ENVIRONMENT=development
```

Setting `ENVIRONMENT=production` enables secure cookies, `SameSite=None`, JSON log format, and proxy-aware rate limiting.

Rate limiting is always enabled (10 req/s, burst 20). In production, the rate limiter automatically trusts proxy headers (`X-Forwarded-For`) to identify clients. Without this, all requests behind a reverse proxy appear to come from the proxy's IP, causing all users to share a single rate limit bucket.

#### Frontend Configuration

```bash
FRONTEND_ADMIN_URL=
FRONTEND_CLIENT_URL=
CORS_DEBUG=false                     # Logs CORS request details when true
```

#### JWT & Authentication Configuration

```bash
JWT_EXPIRATION_HOURS=24
JWT_SECRET=                          # Required, min 32 chars. Generate with: openssl rand -hex 32
```

Secure cookies and `SameSite=None` are enabled automatically when `ENVIRONMENT=production`.

#### Logging Configuration

```bash
LOG_LEVEL=debug
```

#### OAuth Configuration

The server runs without any OAuth configured but users won't be able to log in. Configure at least one provider.

```bash
# Discord - https://discord.com/developers/applications
DISCORD_CLIENT_ID=
DISCORD_CLIENT_SECRET=

# GitHub - https://github.com/settings/developers
GITHUB_CLIENT_ID=
GITHUB_CLIENT_SECRET=

# Google - https://console.cloud.google.com/apis/credentials
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
```

#### Redis (optional)

Used for OAuth state storage. Falls back to in-memory storage if disabled.

```bash
REDIS_ENABLED=true
REDIS_HOST=localhost
REDIS_PASSWORD=
REDIS_PORT=6379
REDIS_URL=                           # If set, used instead of host/port/password
```

#### Server Configuration

```bash
SERVER_PORT=8080                     # Required
SERVER_URL=http://localhost:8080     # Required, used for OAuth redirect URLs
```

#### Game Configuration

Defaults for game creation. All can be overridden per game via the admin API.

```bash
GALAXY_COUNT=1
MAX_PLANETS_PER_SYSTEM=12
MAX_PLAYERS=200
MIN_PLANETS_PER_SYSTEM=3
SECTORS_PER_GALAXY=16
SYSTEMS_PER_SECTOR=16
TURN_INTERVAL_HOURS=1
```

### Reset Database

Drop and recreate the database to start fresh. Migrations run automatically on next server start.

```bash
PGPASSWORD=postgres psql -h localhost -U postgres -c "DROP DATABASE IF EXISTS planets;"
PGPASSWORD=postgres psql -h localhost -U postgres -c "CREATE DATABASE planets;"
```

Adjust credentials to match your `DB_*` settings in `.env`.

### 4. Run the server

```bash
go run cmd/server/main.go
```

### Build

```bash
go build -o planets-server cmd/server/main.go
```
