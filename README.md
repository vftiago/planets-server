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

Copy `.env.example` to `.env` (if available) or create a `.env` file with the required variables. At minimum you need:

```bash
# Generate a secure JWT secret
openssl rand -hex 32
```

Set `JWT_SECRET` in `.env` to the generated value (minimum 32 characters).

### 4. Run the server

```bash
go run cmd/server/main.go
```

### Build

```bash
go build -o planets-server cmd/server/main.go
```
