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

## Migrations

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
