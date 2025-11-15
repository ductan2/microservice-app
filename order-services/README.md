# order-services

Order service for managing orders. Go + Gin. Default port 8004.

- Base API URL: /api/v1
- Health URL: /health

## Run

```
make run
```

Or directly:

```
go run ./cmd/server
```

Set a custom port via `PORT` environment variable (defaults to 8004):

```
PORT=9000 go run ./cmd/server
```

## Database migrations

- SQL migrations live in `migrations/` and follow the `0001_description.up.sql` pattern.
- Apply them locally with `make migrate` (or `go run ./cmd/migrate`); rerunning only applies new files via the `schema_migrations` ledger.
- Override the lookup directory when needed: `MIGRATIONS_DIR=/custom/path make migrate`.

## Infrastructure (Docker Compose)

This repo includes a `docker-compose.yml` that provisions:
- PostgreSQL (16-alpine) on 5432
- Redis (7-alpine) on 6379
- RabbitMQ (3-management) on 5672 (AMQP) and 15672 (HTTP UI)

Start services:
```
make compose-up
```

Stop and remove:
```
make compose-down
```

Tail logs:
```
make compose-logs
```

---

## Notes

This service is currently under development.
