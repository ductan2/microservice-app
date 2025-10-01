SHELL := /bin/bash

APP_NAME := content-services
PKG := ./...

.PHONY: run build tidy test fmt lint generate setup

setup:
	@echo "Setting up content-services..."
	@./scripts/setup.sh

generate:
	@echo "Generating GraphQL code..."
	@go run github.com/99designs/gqlgen generate

run:
	@go run ./cmd/server

build:
	@go build -o bin/$(APP_NAME) ./cmd/server

tidy:
	@go mod tidy

fmt:
	@gofmt -s -w .

test:
	@go test $(PKG) -v
