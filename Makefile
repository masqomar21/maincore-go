.PHONY: dev build migrate seed tidy help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTIDY=$(GOCMD) mod tidy
BINARY_NAME=server-build
MIGRATE_NAME=migrate-build
SEED_NAME=seed-build

# Air for hot reload
AIR_BIN=$(shell go env GOPATH)/bin/air

all: build

help:
	@echo "Available commands:"
	@echo "  make dev      - Run with hot reload (installs 'air' if missing)"
	@echo "  make build    - Build all binaries (api, migrate, seed)"
	@echo "  make migrate  - Run database migrations"
	@echo "  make seed     - Run database seeders"
	@echo "  make tidy     - Run go mod tidy"

install-air:
	@hash air 2>/dev/null || (echo "Installing air..." && go install github.com/air-verse/air@latest)

dev: install-air
	@if [ -f "$(AIR_BIN)" ]; then \
		$(AIR_BIN); \
	else \
		air; \
	fi

build:
	$(GOBUILD) -o $(BINARY_NAME) ./cmd/api
	$(GOBUILD) -o $(MIGRATE_NAME) ./cmd/migrate
	$(GOBUILD) -o $(SEED_NAME) ./cmd/seed

migrate:
	$(GOCMD) run cmd/migrate/main.go

seed:
	$(GOCMD) run cmd/seed/main.go

tidy:
	$(GOTIDY)
