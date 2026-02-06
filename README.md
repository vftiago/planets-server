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

#### Required

```bash
# Authentication - generate with: openssl rand -hex 32
JWT_SECRET=                          # Must be at least 32 characters

# Database
DB_HOST=localhost
DB_NAME=planets

# Server
SERVER_PORT=8080
SERVER_URL=http://localhost:8080     # Used for OAuth redirect URLs
SERVER_READ_TIMEOUT_SECONDS=15
SERVER_WRITE_TIMEOUT_SECONDS=15
SERVER_IDLE_TIMEOUT_SECONDS=60
```

#### Database

```bash
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=
DB_SSLMODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME_MINUTES=5
DB_MIGRATIONS_PATH=migrations
```

#### OAuth providers (configure at least one)

The server runs without any OAuth configured but users won't be able to log in. Configure at least one provider.

```bash
# Google - https://console.cloud.google.com/apis/credentials
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=

# GitHub - https://github.com/settings/developers
GITHUB_CLIENT_ID=
GITHUB_CLIENT_SECRET=

# Discord - https://discord.com/developers/applications
DISCORD_CLIENT_ID=
DISCORD_CLIENT_SECRET=
```

#### Auth settings

```bash
JWT_EXPIRATION_HOURS=24
COOKIE_SAME_SITE=lax                 # Default: lax (use "none" for cross-site)
```

Secure cookies are enabled automatically when `ENVIRONMENT=production`.

#### Admin

The first user to log in with this email gets the admin role.

```bash
ADMIN_EMAIL=admin@localhost
ADMIN_USERNAME=admin
ADMIN_DISPLAY_NAME=Admin
```

#### Frontend / CORS

```bash
FRONTEND_URL=http://localhost:3000
CORS_DEBUG=false                     # Logs CORS request details when true
```

#### Redis (optional)

Used for OAuth state storage. Falls back to in-memory storage if disabled.

```bash
REDIS_ENABLED=true
REDIS_URL=                           # If set, used instead of host/port/password
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
```

#### Rate limiting

```bash
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER_SECOND=10
RATE_LIMIT_BURST_SIZE=20
RATE_LIMIT_TRUST_PROXY=false
```

`RATE_LIMIT_TRUST_PROXY` controls how the rate limiter identifies clients. When `false`, it uses the direct network connection IP, which is correct when the server is exposed directly to the internet. Set to `true` when the server is behind a reverse proxy (nginx, a load balancer, etc.) so it reads the client IP from proxy headers (`X-Forwarded-For`) instead of seeing every request as coming from the proxy's IP.

#### Universe generation defaults

```bash
UNIVERSE_SECTOR_COUNT=16
UNIVERSE_SYSTEMS_PER_SECTOR=16
UNIVERSE_MIN_PLANETS_PER_SYSTEM=3
UNIVERSE_MAX_PLANETS_PER_SYSTEM=12
UNIVERSE_DEFAULT_GALAXY_NAME="Andromeda"
```

#### General

```bash
ENVIRONMENT=development              # "development" or "production"
LOG_LEVEL=debug
LOG_FORMAT=text                      # Default: text (JSON is forced in production)
```

Setting `ENVIRONMENT=production` enables secure cookies, JSON log format, and HTTPS-appropriate defaults.

### 4. Run the server

```bash
go run cmd/server/main.go
```

### Build

```bash
go build -o planets-server cmd/server/main.go
```
