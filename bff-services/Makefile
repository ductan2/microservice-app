SHELL := /bin/bash

APP_NAME := user-services
PKG := ./...

.PHONY: run build tidy test fmt lint compose-up compose-down compose-logs

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

compose-up:
	docker compose up -d

compose-down:
	docker compose down -v

compose-logs:
	docker compose logs -f --tail=200
