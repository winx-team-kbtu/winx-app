# Winx Project

Small microservice setup with:

- `auth`
- `notification`
- `profile`
- `match`
- `recommendation`
- `chat`
- `api-gateway`
- `postgres`
- `mongo`
- `redis`
- `kafka`

## Start

Requirements:

- Docker
- Docker Compose

Run the project:

```bash
docker compose up --build
```

## Local start without Docker (Windows + Ubuntu WSL)

Prerequisites: Go 1.25+, Node.js/npm, and an Ubuntu 24.04 WSL distribution named `Ubuntu`.
Run from the repository root in PowerShell:

```powershell
# First run: install infrastructure packages, build services, migrate, and start.
powershell -ExecutionPolicy Bypass -File .\scripts\start-local.ps1 -Setup

# Subsequent runs:
powershell -ExecutionPolicy Bypass -File .\scripts\start-local.ps1

# Stop the application and its infrastructure; preserve all data:
powershell -ExecutionPolicy Bypass -File .\scripts\stop-local.ps1
```

The scripts run PostgreSQL 16 + PostGIS, MongoDB 8.0, Redis, and Kafka 4.0.0 in WSL.
Go services and Angular run as background Windows processes. No Docker is used.
The initial setup downloads packages and may take several minutes.

- Frontend: http://localhost:4200
- API / Swagger: http://localhost:3000/swagger
- Local PostgreSQL: `127.0.0.1:55432`, database `winx-db`, user/password `winx`/`winx`.
- Go/frontend logs: `.local/logs/`.
- Infrastructure data/logs: `~/.local/share/winx-app/` inside Ubuntu.

Each service gets a git-ignored `config.dev.local.yaml`; `APP_ENV=local` selects it.
Docker configurations are kept as-is. Local gateway configuration enables the Social
routes (`pilot.mode: false`). Local configuration files are created only when absent,
so edits to them are retained. Stop services before rebuilding changed Go code.

All six services share one PostgreSQL database and the default Goose version table.
The local script combines their SQL migrations and applies them in timestamp order,
including the newer auth migrations. Do not run the per-service migration commands
below against this local database.

Graylog is optional for this mode; services keep logging to stderr and retry their
Graylog connection. Email delivery and S3 photo storage need valid SMTP/S3 settings
in the respective local configuration files.

Infrastructure installation follows the official
[MongoDB Ubuntu instructions](https://www.mongodb.com/docs/v8.0/tutorial/install-mongodb-on-ubuntu/),
[Redis WSL instructions](https://redis.io/docs/latest/operate/oss_and_stack/install/install-redis/install-redis-on-windows/),
and [Kafka quickstart](https://kafka.apache.org/40/getting-started/quickstart/).

## Social and Team Mode

The header switches between Social and Team Mode and remembers the selected mode
and last page. On phones the mode switch is at the top and navigation is at the bottom.

Team Mode provides an academic profile (courses, skills, preferred role), a course-specific
candidate deck with touch swipes and buttons, a saved shortlist, and course teams.
Create a team, share its six-character code, or join an existing team. Teams have
1–2 members while Forming, 3–4 while Ready, and 5 when Full. To invite a classmate
from a Team Mode chat, use **Пригласить в команду**. They join when they accept.

To populate academic demo data after running the existing recommendation seeder:

```powershell
$env:PGPASSWORD = 'winx'
psql -h 127.0.0.1 -p 55432 -U winx -d winx-db -v ON_ERROR_STOP=1 -f scripts/seed-team-demo.sql
Remove-Item Env:PGPASSWORD
```

This script touches only `@seed.local` demo accounts. It adds academic profiles,
course selections and three example teams in Algorithms. Existing academic edits
and existing team memberships are retained. Example join codes: `DEMA22` (Forming),
`DEMB33` (Ready), `DEMC55` (Full). A user must select Algorithms before joining.

Deferred UI follow-ups: automatic sign-in after registration, general Social error
presentation, and the mobile chat composer. These are separate from the Team Mode work.

## Docker migrations

After containers are up, run migrations for services that use Postgres.

Important:

- `api-gateway` has no database migrations.
- `profile` uses both `Postgres` and `Mongo`, but only `Postgres` needs migrations.
- `profile` migrations depend on `auth` migrations because `profile` tables reference `users(id)`.
- `profile` location support uses `PostGIS`, so `postgres-winx` must run from the `postgis/postgis` image from `docker-compose.yml`.
- `recommendation` scoring query references `swipes` and `matches` tables, so `match` migrations must run before `recommendation`.
- `chat` tables (`chats`, `messages`) depend on matches being present via Kafka events, so run `match` first.

Recommended `up` order:

1. `auth`
2. `notification`
3. `profile`
4. `match`
5. `recommendation`
6. `chat`

Auth:

```bash
docker compose exec auth-winx go run cmd/migration/main.go -cmd up
```

Notification:

```bash
docker compose exec notification-winx go run cmd/migration/main.go -cmd up
```

Profile:

```bash
docker compose exec profile-winx go run cmd/migration/main.go -cmd up
```

Match:

```bash
docker compose exec match-winx go run cmd/migration/main.go -cmd up
```

Recommendation:

```bash
docker compose exec recommendation-winx go run cmd/migration/main.go -cmd up
```

Chat:

```bash
docker compose exec chat-winx go run cmd/migration/main.go -cmd up
```

If you want the full migration bootstrap in the correct order:

```bash
docker compose exec auth-winx go run cmd/migration/main.go -cmd up
docker compose exec notification-winx go run cmd/migration/main.go -cmd up
docker compose exec profile-winx go run cmd/migration/main.go -cmd up
docker compose exec match-winx go run cmd/migration/main.go -cmd up
docker compose exec recommendation-winx go run cmd/migration/main.go -cmd up
docker compose exec chat-winx go run cmd/migration/main.go -cmd up
```

Recommended `reset` order is the reverse:

1. `chat`
2. `recommendation`
3. `match`
4. `profile`
5. `notification`
6. `auth`

```bash
docker compose exec chat-winx go run cmd/migration/main.go -cmd reset
docker compose exec recommendation-winx go run cmd/migration/main.go -cmd reset
docker compose exec match-winx go run cmd/migration/main.go -cmd reset
docker compose exec profile-winx go run cmd/migration/main.go -cmd reset
docker compose exec notification-winx go run cmd/migration/main.go -cmd reset
docker compose exec auth-winx go run cmd/migration/main.go -cmd reset
```

If you already had an older local Postgres volume before switching to PostGIS, recreate the DB container/volume before running `profile` migrations, otherwise `CREATE EXTENSION postgis` can fail or the geography type may be missing.

## Server deployment

See [the Winx deployment guide](docs/deploy.md) for `winx.emilkalimullin.com`, local Linux builds and server operations.
