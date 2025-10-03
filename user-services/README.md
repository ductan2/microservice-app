# user-services

User service for authentication, sessions, profiles, password reset, and MFA. Go + Gin. Default port 8001.

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

---

## Authentication

Protected endpoints require an Authorization header with a Bearer JWT. The token is validated and the session is checked in Redis.

Header:
- Authorization: Bearer <access_token>

On successful auth, middleware attaches per-request context (also surfaced by /api/v1/profile/check-auth):
- X-User-ID
- X-User-Email
- X-Session-ID

## Common response format

Most endpoints return this envelope:

```json path=null start=null
{
  "status": "success" | "error",
  "message": "optional message",
  "data": { /* endpoint-specific */ },
  "error": { /* optional error details */ }
}
```

Exceptions:
- GET /health returns raw: { "status": "ok" }
- Password endpoints currently respond with raw JSON like {"message": ...} or {"error": ...}
- 204 responses have no body

## Data models (DTOs)

- PublicUser
```json path=null start=null
{
  "id": "uuid",
  "email": "string",
  "email_verified": true,
  "status": "active|locked|disabled|deleted",
  "profile": {
    "display_name": "string",
    "avatar_url": "string",
    "locale": "string",
    "time_zone": "string",
    "updated_at": "RFC3339 timestamp"
  },
  "roles": ["string"],
  "created_at": "RFC3339 timestamp",
  "updated_at": "RFC3339 timestamp"
}
```

- UserProfile
```json path=null start=null
{
  "display_name": "string",
  "avatar_url": "string",
  "locale": "string",
  "time_zone": "string",
  "updated_at": "RFC3339 timestamp"
}
```

- SessionResponse
```json path=null start=null
{
  "id": "uuid",
  "user_agent": "string",
  "ip_addr": "string",
  "created_at": "RFC3339 timestamp",
  "expires_at": "RFC3339 timestamp",
  "is_current": true
}
```

- MFA objects
```json path=null start=null
{
  "id": "uuid",
  "type": "totp|webauthn",
  "label": "string",
  "secret": "string",
  "qr_code_url": "data:image/png;base64,...",
  "added_at": "RFC3339 timestamp"
}
```

---

## API Reference

Base: /api/v1

### Health
- GET /health
  - 200
  ```json path=null start=null
  { "status": "ok" }
  ```

### Auth

- POST /api/v1/register
  - Request
  ```json path=null start=null
  { "email": "user@example.com", "name": "Jane", "password": "Str0ngP@ssword" }
  ```
  - 201
  ```json path=null start=null
  {
    "status": "success",
    "data": {
      "message": "Registration successful! Please check your email to verify your account.",
      "email": "user@example.com"
    }
  }
  ```
  - 400/500: error envelope

- POST /api/v1/login
  - Request
  ```json path=null start=null
  { "email": "user@example.com", "password": "Str0ngP@ssword", "mfa_code": "123456" }
  ```
  - 200
  ```json path=null start=null
  {
    "status": "success",
    "data": {
      "access_token": "jwt...",
      "refresh_token": "jwt...",
      "expires_at": "RFC3339 timestamp",
      "mfa_required": false,
      "user": {
        "id": "uuid",
        "email": "user@example.com",
        "email_verified": true,
        "status": "active",
        "profile": {
          "display_name": "Jane",
          "avatar_url": "https://...",
          "locale": "en",
          "time_zone": "UTC",
          "updated_at": "RFC3339 timestamp"
        },
        "created_at": "RFC3339 timestamp",
        "updated_at": "RFC3339 timestamp"
      }
    }
  }
  ```
  - 401
  ```json path=null start=null
  { "status": "error", "message": "Invalid email or password" }
  ```
  - 401 (MFA)
  ```json path=null start=null
  { "status": "error", "message": "Invalid MFA code" }
  ```

- POST /api/v1/logout
  - 200
  ```json path=null start=null
  { "status": "success", "data": { "message": "Logged out successfully" } }
  ```

- GET /api/v1/verify-email?token=...
  - 200
  ```json path=null start=null
  { "status": "success", "data": { "message": "Email verified successfully! You can now login." } }
  ```
  - 400
  ```json path=null start=null
  { "status": "error", "message": "Verification token is required" }
  ```

### Profile (requires Authorization: Bearer <token>)

- GET /api/v1/profile
  - 200
  ```json path=null start=null
  { "status": "success", "data": { "display_name": "Jane", "avatar_url": "...", "locale": "en", "time_zone": "UTC", "updated_at": "RFC3339" } }
  ```
  - 401 error envelope

- PUT /api/v1/profile
  - Request
  ```json path=null start=null
  { "display_name": "Jane", "avatar_url": "https://...", "locale": "en", "time_zone": "UTC" }
  ```
  - 200: updated profile in envelope

- GET /api/v1/profile/check-auth
  - 200: empty data; headers include X-User-ID, X-User-Email, X-Session-ID

### Password

- POST /api/v1/password/reset/request
  - Request
  ```json path=null start=null
  { "email": "user@example.com" }
  ```
  - 200
  ```json path=null start=null
  { "message": "If the email exists, a password reset link has been sent" }
  ```
  - 400/500
  ```json path=null start=null
  { "error": "..." }
  ```

- POST /api/v1/password/reset/confirm
  - Request
  ```json path=null start=null
  { "token": "reset-token", "new_password": "NewStr0ngP@ss" }
  ```
  - 200
  ```json path=null start=null
  { "message": "Password has been reset successfully" }
  ```
  - 400
  ```json path=null start=null
  { "error": "..." }
  ```

- POST /api/v1/password/change (requires Authorization)
  - Request
  ```json path=null start=null
  { "old_password": "OldPass", "new_password": "NewStr0ngP@ss" }
  ```
  - 200
  ```json path=null start=null
  { "message": "Password changed successfully" }
  ```
  - 400/401/500 with {"error": "..."}

### MFA (requires Authorization)

- POST /api/v1/mfa/setup
  - Request
  ```json path=null start=null
  { "type": "totp", "label": "Authenticator" }
  ```
  - 200: MFA setup info (TOTP includes secret and QR code data URL)

- POST /api/v1/mfa/verify
  - Request
  ```json path=null start=null
  { "method_id": "uuid", "code": "123456" }
  ```
  - 200
  ```json path=null start=null
  { "status": "success", "data": { "message": "MFA verified successfully" } }
  ```

- POST /api/v1/mfa/disable
  - Request
  ```json path=null start=null
  { "method_id": "uuid", "password": "Str0ngP@ssword" }
  ```
  - 200
  ```json path=null start=null
  { "status": "success", "data": { "message": "MFA disabled" } }
  ```

- GET /api/v1/mfa/methods
  - 200: array of MFA methods in envelope

### Sessions (requires Authorization)

- GET /api/v1/sessions
  - 200
  ```json path=null start=null
  { "status": "success", "data": [ { "id": "uuid", "user_agent": "...", "ip_addr": "...", "created_at": "...", "expires_at": "...", "is_current": true } ] }
  ```

- DELETE /api/v1/sessions/:id
  - 204 No Content

- POST /api/v1/sessions/revoke-all
  - 204 No Content

---

## Curl quickstart

- Register
```bash path=null start=null
curl -s -X POST http://localhost:8001/api/v1/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","name":"Jane","password":"Str0ngP@ssword"}'
```

- Login
```bash path=null start=null
curl -s -X POST http://localhost:8001/api/v1/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"Str0ngP@ssword"}'
```

- Get profile
```bash path=null start=null
TOKEN=<access_token>
curl -s http://localhost:8001/api/v1/profile \
  -H "Authorization: Bearer $TOKEN"
```

---

## Notes
- Role and audit endpoints exist in code scaffolding but are not currently registered in the router.
- Refresh token DTOs exist but a refresh endpoint is not exposed in routes.
- Session validation depends on Redis being available and seeded by login flow.

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

