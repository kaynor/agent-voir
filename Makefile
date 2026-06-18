SHELL := /bin/bash

.PHONY: help dev-up dev-down run-gateway run-api run-web fmt test lint clean

help:
	@echo "AgentVoir development commands"
	@echo "  make dev-up        Start local dependencies"
	@echo "  make dev-down      Stop local dependencies"
	@echo "  make run-gateway   Run Go gateway"
	@echo "  make run-api       Run Go registry API"
	@echo "  make run-web       Run Next.js web app"
	@echo "  make fmt           Format code"
	@echo "  make test          Run tests"
	@echo "  make lint          Run lint checks"
	@echo "  make clean         Remove build outputs"

dev-up:
	docker compose -f deployments/docker/docker-compose.yml up -d

dev-down:
	docker compose -f deployments/docker/docker-compose.yml down

run-gateway:
	cd apps/gateway && go run ./cmd/gateway

run-api:
	cd apps/registry-api && go run ./cmd/registry-api

run-web:
	cd apps/web && npm install && npm run dev

fmt:
	gofmt -w apps/gateway apps/registry-api packages/sdk-go || true
	cd apps/web && npm run format || true
	cd packages/sdk-typescript && npm run format || true
	cd packages/sdk-python && python -m ruff format . || true

test:
	cd apps/gateway && go test ./...
	cd apps/registry-api && go test ./...
	cd packages/sdk-go && go test ./...
	cd packages/sdk-python && python -m pytest || true
	cd packages/sdk-typescript && npm test || true

lint:
	cd apps/web && npm run lint || true
	cd packages/sdk-typescript && npm run lint || true
	cd packages/sdk-python && python -m ruff check . || true

clean:
	rm -rf bin dist build .data
