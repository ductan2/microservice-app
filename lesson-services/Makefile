PYTHON := python3

.PHONY: run dev test lint fmt migrate migrate-status migrate-create

run:
	uvicorn main:app --reload

dev: run

test:
	pytest

lint:
	ruff check .

fmt:
	ruff check --fix .

migrate:
	./migrate.sh upgrade

migrate-status:
	./migrate.sh status

migrate-create:
	@if [ -z "$(MSG)" ]; then echo "Usage: make migrate-create MSG=\"Description\""; exit 1; fi
	./migrate.sh create "$(MSG)"
