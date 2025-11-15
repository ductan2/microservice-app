SHELL := /bin/bash

APP_NAME := order-services
PKG := ./...

.PHONY: run build tidy test fmt lint migrate compose-up compose-down compose-logs

run:
	go run ./cmd/server

build:
	go build -o bin/$(APP_NAME) ./cmd/server

tidy:
	go mod tidy

fmt:
	gofmt -s -w .

test:
	go test $(PKG) -v

migrate:
	go run ./cmd/migrate

compose-up:
	docker compose up -d

compose-down:
	docker compose down -v

compose-logs:
	docker compose logs -f --tail=200
