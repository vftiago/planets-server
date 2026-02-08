# Railway Deployment

## 1. Create the Railway project

- Go to [railway.app](https://railway.app) and create a new project
- Choose **"Deploy from GitHub Repo"** and connect your `planets-server` repo
- This gives you auto-deploy on push to master

## 2. Add services

In your Railway project, add two services alongside your app:

- **PostgreSQL** — click "New" → "Database" → "PostgreSQL"
- **Redis** — click "New" → "Database" → "Redis"

Railway provisions these instantly and gives you connection variables.

## 3. Set environment variables

In your app service's **Variables** tab, add:

```
# Railway provides these as reference variables — link them from your Postgres service
DB_HOST=${{Postgres.PGHOST}}
DB_PORT=${{Postgres.PGPORT}}
DB_USER=${{Postgres.PGUSER}}
DB_PASSWORD=${{Postgres.PGPASSWORD}}
DB_NAME=${{Postgres.PGDATABASE}}
DB_SSLMODE=require

# Link from your Redis service
REDIS_ENABLED=true
REDIS_URL=${{Redis.REDIS_URL}}

# Generate locally with: openssl rand -hex 32
JWT_SECRET=<paste your generated secret>

# Server
ENVIRONMENT=production
SERVER_PORT=8080
SERVER_URL=https://<your-app>.up.railway.app

# Frontend
FRONTEND_CLIENT_URL=https://<your-client-domain>
FRONTEND_ADMIN_URL=https://<your-admin-domain>

# OAuth (whichever providers you use)
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
GITHUB_CLIENT_ID=
GITHUB_CLIENT_SECRET=
DISCORD_CLIENT_ID=
DISCORD_CLIENT_SECRET=

# Admin
ADMIN_EMAIL=<your email>
```

The `${{Service.VAR}}` syntax is Railway's reference variables — they auto-resolve to the actual values from your Postgres/Redis services.

## 4. Things to fix before deploying

### A. Railway sets `PORT`, not `SERVER_PORT`

Railway injects the port via a `PORT` env var. Your config reads `SERVER_PORT`. Easiest fix: set `SERVER_PORT=${{PORT}}` in Railway variables, or update your config to fall back to `PORT`.

### B. Migration path (item #25 in IMPROVEMENT_PLAN)

The relative `filepath.WalkDir("migrations", ...)` may not find the migrations directory depending on Railway's working directory. Verify this works after the first deploy — if it doesn't, use `go:embed` or set the working directory explicitly.

## 5. Deploy settings

In your app service's **Settings** tab:

- **Build command**: `go build -o planets-server cmd/server/main.go`
- **Start command**: `./planets-server`
- **Root directory**: `/` (default)

Railway auto-detects Go projects, but setting these explicitly avoids surprises.

## 6. Generate a domain

In **Settings** → **Networking** → **Generate Domain**. This gives you the `*.up.railway.app` URL to use for `SERVER_URL` and OAuth redirect URLs.

## 7. Update OAuth redirect URLs

In each OAuth provider's console (Google, GitHub, Discord), add the new callback URLs:

- `https://<your-app>.up.railway.app/auth/google/callback`
- `https://<your-app>.up.railway.app/auth/github/callback`
- `https://<your-app>.up.railway.app/auth/discord/callback`
