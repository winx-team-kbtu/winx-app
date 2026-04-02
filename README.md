# Winx Project

Small microservice setup with:

- `auth`
- `notification`
- `profile`
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

## Migrations

After containers are up, run migrations for services that use Postgres.

Important:

- `api-gateway` has no database migrations.
- `profile` uses both `Postgres` and `Mongo`, but only `Postgres` needs migrations.
- `profile` migrations depend on `auth` migrations because `profile` tables reference `users(id)`.
- `profile` location support uses `PostGIS`, so `postgres-winx` must run from the `postgis/postgis` image from `docker-compose.yml`.

Recommended `up` order:

1. `auth`
2. `notification`
3. `profile`

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

If you want the full migration bootstrap in the correct order:

```bash
docker compose exec auth-winx go run cmd/migration/main.go -cmd up
docker compose exec notification-winx go run cmd/migration/main.go -cmd up
docker compose exec profile-winx go run cmd/migration/main.go -cmd up
```

Recommended `reset` order is the reverse:

1. `profile`
2. `notification`
3. `auth`

```bash
docker compose exec profile-winx go run cmd/migration/main.go -cmd reset
docker compose exec notification-winx go run cmd/migration/main.go -cmd reset
docker compose exec auth-winx go run cmd/migration/main.go -cmd reset
```

If you already had an older local Postgres volume before switching to PostGIS, recreate the DB container/volume before running `profile` migrations, otherwise `CREATE EXTENSION postgis` can fail or the geography type may be missing.
