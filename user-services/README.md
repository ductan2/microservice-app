# user-services

Minimal Go backend using Gin, listening on port 8001 by default.

## Run

```
make run
```

Or directly:

```
go run ./cmd/server
```

Set a custom port via `PORT` environment variable (defaults to 8001):

```
PORT=9000 go run ./cmd/server
```

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

### Default credentials (local dev)
Defined in `.env` (not committed):
```
POSTGRES_USER=user
POSTGRES_PASSWORD=password
POSTGRES_DB=userdb
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_SSLMODE=disable

REDIS_ADDR=localhost:6379
REDIS_PASSWORD=

RABBITMQ_USER=guest
RABBITMQ_PASSWORD=guest
RABBITMQ_VHOST=/
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
```

When running the app in Docker, set hosts to the service names (`postgres`, `redis`, `rabbitmq`). Example override:
```
POSTGRES_HOST=postgres
REDIS_ADDR=redis:6379
RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/
```

## Endpoints

- `GET /health` -> `{ "status": "ok" }`
- `POST /register` -> `{ "message" :"success | fail" }`
- `POST /login` -> `{ "token" : "....."}`

## SQL 
``` sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- users: canonical identity
CREATE TABLE users (
  id               uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  email            citext UNIQUE NOT NULL,
  email_normalized citext GENERATED ALWAYS AS (lower(email)) STORED,
  password_hash    text NOT NULL,
  email_verified   boolean NOT NULL DEFAULT false,
  status           text NOT NULL DEFAULT 'active' CHECK (status IN ('active','locked','disabled','deleted')),
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX users_email_norm_idx ON users(email_normalized);

-- profile (non-auth PII kept minimal here)
CREATE TABLE user_profiles (
  user_id    uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  display_name text,
  avatar_url text,
  locale    text DEFAULT 'en',
  time_zone text DEFAULT 'UTC',
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- sessions (server-side record of JWTs issued)
CREATE TABLE sessions (
  id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id      uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  user_agent   text,
  ip_addr      inet,
  created_at   timestamptz NOT NULL DEFAULT now(),
  expires_at   timestamptz NOT NULL,
  revoked_at   timestamptz
);
CREATE INDEX sessions_user_expires_idx ON sessions(user_id, expires_at);

-- refresh tokens (rotating)
CREATE TABLE refresh_tokens (
  id              uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  session_id      uuid NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
  token_hash      text NOT NULL UNIQUE,
  issued_at       timestamptz NOT NULL DEFAULT now(),
  expires_at      timestamptz NOT NULL,
  consumed_at     timestamptz,
  revoked_at      timestamptz
);
CREATE INDEX refresh_tokens_session_idx ON refresh_tokens(session_id, expires_at);

-- MFA: TOTP/WebAuthn
CREATE TABLE mfa_methods (
  id            uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id       uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  type          text NOT NULL CHECK (type IN ('totp','webauthn')),
  label         text,
  secret        text,        -- encrypted at rest (KMS)
  webauthn_pub  text,        -- for 'webauthn'
  added_at      timestamptz NOT NULL DEFAULT now(),
  last_used_at  timestamptz
);
CREATE UNIQUE INDEX one_totp_per_user ON mfa_methods(user_id) WHERE type='totp';

-- login attempts (throttling / anomaly detection)
CREATE TABLE login_attempts (
  id         bigserial PRIMARY KEY,
  user_id    uuid,
  email      citext,
  ip_addr    inet,
  success    boolean NOT NULL,
  reason     text,  -- e.g., invalid_password, locked, mfa_required
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX login_attempts_user_time_idx ON login_attempts(coalesce(user_id,'00000000-0000-0000-0000-000000000000'::uuid), created_at DESC);

-- password reset flow
CREATE TABLE password_resets (
  id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id      uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash   text NOT NULL UNIQUE,
  expires_at   timestamptz NOT NULL,
  consumed_at  timestamptz
);

-- RBAC (optional)
CREATE TABLE roles (
  id   uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  name text UNIQUE NOT NULL
);
CREATE TABLE user_roles (
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role_id uuid NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  PRIMARY KEY(user_id, role_id)
);

-- audit log (append-only)
CREATE TABLE audit_logs (
  id          bigserial PRIMARY KEY,
  user_id     uuid,
  actor_id    uuid, -- who performed action (user or system)
  action      text NOT NULL, -- e.g., user.login, user.update_profile
  ip_addr     inet,
  metadata    jsonb NOT NULL DEFAULT '{}',
  created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX audit_logs_user_time_idx ON audit_logs(coalesce(user_id,'00000000-0000-0000-0000-000000000000'::uuid), created_at DESC);

-- outbox (for cross-service events)
CREATE TABLE outbox (
  id           bigserial PRIMARY KEY,
  aggregate_id uuid NOT NULL,    -- users.id, sessions.id, etc.
  topic        text NOT NULL,    -- e.g., user.events
  type         text NOT NULL,    -- e.g., UserRegistered
  payload      jsonb NOT NULL,
  created_at   timestamptz NOT NULL DEFAULT now(),
  published_at timestamptz
);
CREATE INDEX outbox_unpublished_idx ON outbox(published_at) WHERE published_at IS NULL;

```

