# Winx Project

Small microservice setup with:

- `auth`
- `notification`
- `api-gateway`
- `postgres`
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

Auth:

```bash
docker compose exec auth-winx go run cmd/migration/main.go -cmd up
```

Notification:

```bash
docker compose exec notification-winx go run cmd/migration/main.go -cmd up
```