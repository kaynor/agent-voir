SHELL := /bin/bash

.PHONY: help dev-up dev-up-all dev-down dev-logs dev-docs onebox-up onebox-up-build onebox-down onebox-logs onebox-reset onebox-smoke quickstart run-gateway run-api run-token-accounting run-web db-migrate fmt test lint clean

ONEBOX_COMPOSE := deployments/docker/docker-compose.onebox.yml
ONEBOX_BUILD_COMPOSE := deployments/docker/docker-compose.onebox.build.yml
ONEBOX_ENV := deployments/docker/.env.onebox

help:
	@echo "AgentVoir development commands"
	@echo "  make onebox-up            Pull pre-built images and start onebox (end users)"
	@echo "  make onebox-up-build      Build app images locally and start onebox (contributors)"
	@echo "  make onebox-down          Stop onebox stack"
	@echo "  make onebox-logs          Follow onebox stack logs"
	@echo "  make onebox-reset         Stop onebox stack and delete onebox volumes"
	@echo "  make onebox-smoke         Run health checks against onebox stack"
	@echo "  make quickstart           End-to-end demo (onebox + agent + cache + usage)"
	@echo "  make demo-policy          OPA policy denial demo (HTTP 403)"
	@echo "  make demo-budget          Budget enforcement demo (HTTP 429)"
	@echo "  make showcase             quickstart + governance demos"
	@echo "  make run-web              Admin console at http://localhost:3000"
	@echo "  make dev-docs             Start local Swagger UI for API specs (:8089)"
	@echo "  make dev-up               Start local infrastructure (Postgres, Redis, ClickHouse, ...)"
	@echo "  make dev-up-all           Start infrastructure + AgentVoir apps in Docker"
	@echo "  make dev-down             Stop Docker Compose services"
	@echo "  make dev-logs             Follow Docker Compose logs"
	@echo "  make db-migrate           Apply PostgreSQL metadata migrations"
	@echo "  make run-gateway          Run Go gateway locally"
	@echo "  make run-api              Run Go registry API locally"
	@echo "  make run-token-accounting Run usage event ingestion service locally"
	@echo "  make run-web              Run Next.js web app"
	@echo "  make fmt           Format code"
	@echo "  make test          Run tests"
	@echo "  make lint          Run lint checks"
	@echo "  make clean         Remove build outputs"

dev-up:
	docker compose -f deployments/docker/docker-compose.yml up -d

dev-up-all:
	docker compose -f deployments/docker/docker-compose.yml --profile apps up -d --build

dev-down:
	docker compose -f deployments/docker/docker-compose.yml --profile apps down

dev-logs:
	docker compose -f deployments/docker/docker-compose.yml --profile apps logs -f

dev-docs:
	docker compose -f deployments/docker/docker-compose.yml --profile docs up -d

onebox-up:
	@cp -n deployments/docker/.env.onebox.example $(ONEBOX_ENV) || true
	docker compose --env-file $(ONEBOX_ENV) -f $(ONEBOX_COMPOSE) pull
	docker compose --env-file $(ONEBOX_ENV) -f $(ONEBOX_COMPOSE) up -d

onebox-up-build:
	@cp -n deployments/docker/.env.onebox.example $(ONEBOX_ENV) || true
	docker compose --env-file $(ONEBOX_ENV) -f $(ONEBOX_COMPOSE) -f $(ONEBOX_BUILD_COMPOSE) up -d --build

onebox-down:
	docker compose --env-file $(ONEBOX_ENV) -f $(ONEBOX_COMPOSE) down

onebox-logs:
	docker compose --env-file $(ONEBOX_ENV) -f $(ONEBOX_COMPOSE) logs -f

onebox-reset:
	docker compose --env-file $(ONEBOX_ENV) -f $(ONEBOX_COMPOSE) down -v

onebox-smoke:
	@set -a && source $(ONEBOX_ENV) && set +a && \
	GATEWAY_PORT=$${AGENTVOIR_GATEWAY_PORT:-8080} && \
	REGISTRY_PORT=$${AGENTVOIR_REGISTRY_PORT:-8081} && \
	USAGE_PORT=$${AGENTVOIR_USAGE_PORT:-8082} && \
	API_KEY=$${GATEWAY_API_KEY:-agentvoir-onebox-key} && \
	echo "==> registry /healthz" && curl -fsS "http://localhost:$$REGISTRY_PORT/healthz" && echo && \
	echo "==> usage /healthz" && curl -fsS "http://localhost:$$USAGE_PORT/healthz" && echo && \
	echo "==> gateway /healthz" && curl -fsS "http://localhost:$$GATEWAY_PORT/healthz" && echo && \
	echo "==> gateway /v1/models" && curl -fsS "http://localhost:$$GATEWAY_PORT/v1/models" -H "Authorization: Bearer $$API_KEY" | head -c 200 && echo

quickstart:
	@chmod +x scripts/quickstart.sh
	@./scripts/quickstart.sh

demo-policy:
	@chmod +x scripts/demo-policy-denial.sh
	@./scripts/demo-policy-denial.sh

demo-budget:
	@chmod +x scripts/demo-budget-block.sh
	@./scripts/demo-budget-block.sh

showcase:
	@chmod +x scripts/quickstart.sh scripts/demo-policy-denial.sh scripts/demo-budget-block.sh
	@./scripts/quickstart.sh --no-start || ./scripts/quickstart.sh
	@./scripts/demo-policy-denial.sh
	@./scripts/demo-budget-block.sh

	db-migrate:
	cd apps/registry-api && go run ./cmd/migrate

seed-demo:
	@chmod +x scripts/seed-demo.sh
	@./scripts/seed-demo.sh

wait-for-onebox:
	@chmod +x scripts/wait-for-onebox.sh
	@./scripts/wait-for-onebox.sh

test-migrations:
	@chmod +x scripts/test-migrations.sh
	@./scripts/test-migrations.sh

run-gateway:
	cd apps/gateway && go run ./cmd/gateway

run-api:
	cd apps/registry-api && go run ./cmd/registry-api

run-token-accounting:
	cd services/token-accounting && go run ./cmd/token-accounting

run-web:
	cd apps/web && npm install && REGISTRY_API_URL=http://localhost:8081 TOKEN_ACCOUNTING_URL=http://localhost:8082 npm run dev

fmt:
	gofmt -w apps/gateway apps/registry-api services/token-accounting packages/sdk-go || true
	cd apps/web && npm run format || true
	cd packages/sdk-typescript && npm run format || true
	cd packages/sdk-python && python -m ruff format . || true

test:
	cd apps/gateway && go test ./...
	cd apps/registry-api && go test ./...
	cd services/token-accounting && go test ./...
	cd packages/sdk-go && go test ./...
	cd packages/sdk-python && python -m pytest || true
	cd packages/sdk-typescript && npm test || true

lint:
	cd apps/web && npm run lint || true
	cd packages/sdk-typescript && npm run lint || true
	cd packages/sdk-python && python -m ruff check . || true

clean:
	rm -rf bin dist build .data
