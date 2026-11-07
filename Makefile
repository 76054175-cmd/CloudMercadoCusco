# Comandos comunes. Solo requieren Docker (y make).
UID := $(shell id -u)
GID := $(shell id -g)
export UID GID

GO_IMAGE := golang:1.27-alpine
GO_RUN := docker run --rm -v "$(CURDIR)":/app -v qhatudigital_go-mod:/go/pkg/mod \
	-v qhatudigital_go-build:/root/.cache/go-build -w /app/backend $(GO_IMAGE)
NODE_RUN := docker run --rm -u $(UID):$(GID) -e HOME=/tmp -e COREPACK_HOME=/tmp/corepack \
	-e COREPACK_ENABLE_DOWNLOAD_PROMPT=0 -e CI=true -e npm_config_store_dir=/app/.pnpm-store -v "$(CURDIR)":/app -w /app/apps/web node:24-bookworm-slim

.PHONY: help env dev down logs reset seed seed-check seed-update tidy check build

help: ## Muestra esta ayuda
	@grep -E '^[a-z-]+:.*##' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  make %-12s %s\n", $$1, $$2}'

env: ## Crea .env desde .env.example (si no existe)
	@test -f .env || cp .env.example .env && echo ".env listo"

dev: env ## Levanta todo el entorno de desarrollo
	docker compose up

down: ## Detiene el entorno
	docker compose down

logs: ## Muestra logs del core
	docker compose logs -f core

reset: ## Borra la base de datos de desarrollo (pb_data) y vuelve a sembrar al iniciar
	docker compose down
	docker volume rm qhatudigital_pb-data || true

seed: ## Importa data/seed en la base de desarrollo (solo crea lo que falta)
	docker compose exec core go run ./core seed --dir=/pb_data

seed-update: ## Importa data/seed SOBRESCRIBIENDO registros existentes
	docker compose exec core go run ./core seed --dir=/pb_data --update

seed-check: env ## Valida data/seed sin guardar (usa una base temporal)
	$(GO_RUN) sh -c 'APP_ENV=dev SEED_DEV_PASSWORD=seed-check-only QHATU_SEED_DIR=/app/data/seed go run ./core seed --dry-run --dir=/tmp/pb_check'

tidy: ## go mod tidy en todos los módulos
	$(GO_RUN) sh -c 'for m in core shared ms-*; do GOWORK=off go mod tidy -C $$m; done; chown -R $(UID):$(GID) /app/backend'

check: ## go vet + svelte-check
	$(GO_RUN) go vet ./core/... ./shared/... ./ms-auth-trust ./ms-payments-qr ./ms-logistics-dispatch ./ms-ratings-aggregator
	$(NODE_RUN) sh -c 'corepack pnpm install --frozen-lockfile && corepack pnpm check'

build: ## Construye la imagen de producción (un solo binario con la web embebida)
	docker build -t qhatudigital:latest .
