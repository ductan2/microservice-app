# Repository Guidelines

## Project Structure & Module Organization
Each microservice lives in its own folder (`user-services`, `lesson-services`, `content-services`, `notification-services`, `bff-services`). Entry points stay under `cmd/server`, shared logic belongs in `internal/`, and transport glue in `server/`. Tests sit beside the code as `_test.go`; place fixtures under `internal/<package>/testdata` if needed. `infrastructure/` hosts Docker Compose, Terraform, monitoring configs, and helper scripts (`dev.sh`, `start.sh`, `stop.sh`). Markdown specs such as `ACTIVITY_SESSION_API.md` and `FRONTEND_INTEGRATION_README.md` document contracts and must be updated with API changes.

## Build, Test, and Development Commands
- `make -C infrastructure dev` — build and start the full stack with `docker-compose.env`.
- `make -C infrastructure logs` or `status` — follow container output or check health.
- `make -C user-services run` (swap directory name per service) — execute a single service locally.
- `make -C user-services test` or `go test ./...` — run that service’s unit tests.
- `make -C infrastructure clean` — stop everything and prune volumes when a reset is needed.
Create a local copy of the provided `docker-compose.env` (or prod variant) before invoking any Compose target.

## Coding Style & Naming Conventions
This repo is Go-first: use tabs, run `make fmt` (gofmt) before committing, and keep imports grouped stdlib / third-party / internal. Name packages with short lowercase nouns (`session`, `mailer`) and exported types with PascalCase (`StartSessionService`). Align filenames with their feature (`progress_controller.go`), and follow the `*.env` / `.example` suffix pattern for configuration—never commit live secrets.

## Testing Guidelines
Favor table-driven tests named `Test<Thing>` in the package they exercise. Changes touching multiple services need coverage in each service’s `internal/...` package plus a `go test ./...` run. Use `make -C infrastructure dev` to boot Postgres, Redis, and RabbitMQ before executing integration tests. Keep critical handlers (auth, lesson completion, notifications) under regression tests and document any temporary gaps in the pull request.

## Commit & Pull Request Guidelines
Recent history (`git log`) shows imperative, descriptive commit subjects such as “Add quiz attempt service…”; follow that style and keep the first line ≤72 characters. Describe the affected service or subsystem up front (e.g., “user: add MFA enrollment API”). Pull requests should link the tracking issue, summarize behavior changes, list test commands run, and include screenshots or curl examples for new endpoints. Reference updated docs (`ACTIVITY_SESSION_API.md`, etc.) so reviewers can verify contract changes quickly.
