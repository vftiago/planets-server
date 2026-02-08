# Environment Variable Cleanup

Analysis of `.env.example` for variables that could be hardcoded or removed.

## Done

| Variable                       | Resolution                          |
| ------------------------------ | ----------------------------------- |
| `DB_MIGRATIONS_PATH`           | Hardcoded (was also unused)         |
| `DB_MAX_OPEN_CONNS`            | Hardcoded `25`                      |
| `DB_MAX_IDLE_CONNS`            | Hardcoded `5`                       |
| `DB_CONN_MAX_LIFETIME_MINUTES` | Hardcoded `5`                       |
| `SERVER_READ_TIMEOUT_SECONDS`  | Hardcoded `15`                      |
| `SERVER_WRITE_TIMEOUT_SECONDS` | Hardcoded `15`                      |
| `SERVER_IDLE_TIMEOUT_SECONDS`  | Hardcoded `60`                      |
| `COOKIE_SAME_SITE`             | Derived from `ENVIRONMENT` (`none` in production, `lax` in development) |
| `LOG_FORMAT`                   | Removed (was unused — format already derived from `ENVIRONMENT`)        |
| `RATE_LIMIT_TRUST_PROXY`       | Derived from `ENVIRONMENT` (`true` in production, `false` in development) |
| `RATE_LIMIT_ENABLED`            | Hardcoded `true`                    |
| `RATE_LIMIT_REQUESTS_PER_SECOND` | Hardcoded `10`                     |
| `RATE_LIMIT_BURST_SIZE`         | Hardcoded `20`                      |

## Already unused

The `UniverseConfig` has a TODO noting these aren't used yet — they're meant to become validation caps for the game creation endpoint later.

| Variable                 | Default |
| ------------------------ | ------- |
| `GALAXIES_PER_UNIVERSE`  | `1`     |
| `SECTORS_PER_GALAXY`     | `16`    |
| `SYSTEMS_PER_SECTOR`     | `16`    |
| `MIN_PLANETS_PER_SYSTEM` | `3`     |
| `MAX_PLANETS_PER_SYSTEM` | `12`    |

## Kept as env vars

| Variable | Reason |
| -------- | ------ |
| `CORS_DEBUG` | Debug flag, useful to toggle without code changes |

## What should remain

Variables that genuinely vary between environments:

- `ENVIRONMENT`
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`
- `REDIS_ENABLED`, `REDIS_URL`, `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`
- `JWT_SECRET`
- `SERVER_PORT`, `SERVER_URL`
- `FRONTEND_CLIENT_URL`, `FRONTEND_ADMIN_URL`, `CORS_DEBUG`
- `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`
- `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`
- `DISCORD_CLIENT_ID`, `DISCORD_CLIENT_SECRET`
- `ADMIN_EMAIL`, `ADMIN_USERNAME`, `ADMIN_DISPLAY_NAME`
- `LOG_LEVEL`
