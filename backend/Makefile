# ============================================================
# Révise Mieux — Backend Makefile (Go + Gin + pgx)
# ============================================================

.PHONY: help deps build run dev test test-unit test-integration \
        lint fmt vet db-migrate db-create-migration swagger clean \
        docker-build bench bench-idp bench-ocr bench-all bench-report

.DEFAULT_GOAL := help

# ------------------------------------------------------------
# Variables
# ------------------------------------------------------------

APP_NAME    := revisemieux
CMD_DIR     := ./cmd/server
BUILD_DIR   := ./bin
BINARY      := $(BUILD_DIR)/$(APP_NAME)
MIGRATIONS  := ./migrations

# Go
GO          := go
GOTEST      := $(GO) test
GOVET       := $(GO) vet
GOFMT       := gofmt
GOMOD       := $(GO) mod

# DB (configurable via .env ou env vars)
DATABASE_URL ?= postgres://revisemieux:revisemieux@localhost:5432/revisemieux?sslmode=disable

# Tools (installés dans GOBIN)
GOLANGCI_LINT := $(shell which golangci-lint 2>/dev/null)

# Couleurs
GREEN  := \033[32m
CYAN   := \033[36m
YELLOW := \033[33m
RED    := \033[31m
RESET  := \033[0m

# ------------------------------------------------------------
# Help
# ------------------------------------------------------------

help: ## Affiche cette aide
	@echo ""
	@echo "$(CYAN)Backend$(RESET) — Go API"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-22s$(RESET) %s\n", $$1, $$2}'
	@echo ""

# ------------------------------------------------------------
# Dependencies
# ------------------------------------------------------------

deps: ## Installe les dépendances Go
	@echo "$(CYAN)go mod download...$(RESET)"
	@$(GOMOD) download
	@$(GOMOD) tidy
	@echo "$(GREEN)Deps OK$(RESET)"

tools: ## Installe les outils de dev (golangci-lint, swag)
	@echo "$(CYAN)Installation golangci-lint + swag...$(RESET)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "$(GREEN)Tools OK$(RESET)"

# ------------------------------------------------------------
# Build
# ------------------------------------------------------------

swagger: ## Génère la doc Swagger depuis les annotations
	@echo "$(CYAN)swag init...$(RESET)"
	@swag init -g cmd/server/main.go -o internal/docs --parseDependency --parseInternal
	@echo "$(GREEN)Swagger OK$(RESET) — http://localhost:$(or $(PORT),8080)/swagger/index.html"

build: swagger ## Compile le binaire (regénère swagger)
	@mkdir -p $(BUILD_DIR)
	@echo "$(CYAN)Build $(BINARY)...$(RESET)"
	@CGO_ENABLED=0 $(GO) build -o $(BINARY) $(CMD_DIR)
	@echo "$(GREEN)Build OK$(RESET)"

run: build ## Compile et lance
	@$(BINARY)

dev: ## Lance avec hot-reload (go run)
	@echo "$(CYAN)Dev server...$(RESET)"
	@$(GO) run $(CMD_DIR)/main.go

# ------------------------------------------------------------
# Tests
# ------------------------------------------------------------

test: test-unit ## Alias : tests unitaires

test-unit: ## Tests unitaires (domaine + app, sans DB)
	@echo "$(CYAN)Tests unitaires...$(RESET)"
	@$(GOTEST) -short -race -count=1 ./internal/domain/... ./internal/app/...
	@echo "$(GREEN)Tests unitaires OK$(RESET)"

test-integration: ## Tests d'intégration (nécessite PostgreSQL)
	@echo "$(CYAN)Tests intégration...$(RESET)"
	@DATABASE_URL=$(DATABASE_URL) $(GOTEST) -race -count=1 -tags=integration ./internal/infra/...
	@echo "$(GREEN)Tests intégration OK$(RESET)"

test-all: test-unit test-integration ## Tous les tests

test-cover: ## Tests avec couverture
	@mkdir -p $(BUILD_DIR)
	@$(GOTEST) -short -race -coverprofile=$(BUILD_DIR)/coverage.out ./internal/...
	@$(GO) tool cover -func=$(BUILD_DIR)/coverage.out
	@echo "$(CYAN)Rapport HTML : $(BUILD_DIR)/coverage.html$(RESET)"
	@$(GO) tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html

test-mastery: ## Tests du domaine Mastery uniquement
	@echo "$(CYAN)Tests Mastery...$(RESET)"
	@$(GOTEST) -v -race -count=1 ./internal/domain/mastery/...

# ------------------------------------------------------------
# Qualité
# ------------------------------------------------------------

lint: vet ## Lint (golangci-lint + vet)
ifdef GOLANGCI_LINT
	@echo "$(CYAN)golangci-lint...$(RESET)"
	@golangci-lint run ./...
else
	@echo "$(YELLOW)golangci-lint non installé — make tools pour l'installer$(RESET)"
endif
	@echo "$(GREEN)Lint OK$(RESET)"

vet: ## Go vet
	@echo "$(CYAN)go vet...$(RESET)"
	@$(GOVET) ./...

fmt: ## Formatte le code Go
	@echo "$(CYAN)gofmt...$(RESET)"
	@$(GOFMT) -s -w .
	@echo "$(GREEN)Fmt OK$(RESET)"

fmt-check: ## Vérifie le formatage (CI)
	@test -z "$$($(GOFMT) -l .)" || (echo "$(RED)Fichiers non formatés :$(RESET)" && $(GOFMT) -l . && exit 1)

check-domain: ## Vérifie que domain/ n'importe pas d'infra/http/app
	@echo "$(CYAN)Vérification imports domaine...$(RESET)"
	@bash scripts/check-domain-imports.sh

check: fmt-check vet check-domain test-unit ## Gate CI : format + vet + imports domaine + tests
	@echo "$(GREEN)All checks passed$(RESET)"

# ------------------------------------------------------------
# Database & Migrations
# ------------------------------------------------------------

db-migrate: ## Applique toutes les migrations
	@echo "$(CYAN)Migrations...$(RESET)"
	@for f in $(MIGRATIONS)/*.sql; do \
		echo "  Applying $$(basename $$f)..."; \
		psql "$(DATABASE_URL)" -f "$$f" -v ON_ERROR_STOP=1 || exit 1; \
	done
	@echo "$(GREEN)Migrations OK$(RESET)"

db-create-migration: ## Crée un fichier de migration (usage: make db-create-migration NAME=add_foo)
ifndef NAME
	$(error NAME requis. Usage: make db-create-migration NAME=add_foo)
endif
	@NEXT=$$(printf "%03d" $$(( $$(ls -1 $(MIGRATIONS)/*.sql 2>/dev/null | wc -l) + 1 ))); \
	FILE="$(MIGRATIONS)/$${NEXT}_$(NAME).sql"; \
	echo "-- Migration: $${NEXT}_$(NAME).sql" > "$$FILE"; \
	echo "" >> "$$FILE"; \
	echo "$(GREEN)Créé : $$FILE$(RESET)"

db-psql: ## Ouvre psql sur la DB
	@psql "$(DATABASE_URL)"

# ------------------------------------------------------------
# Benchmark LLM
# ------------------------------------------------------------

BENCH_CMD := ./cmd/benchmark

# Charge .env.bench s'il existe (clés API pour les benchmarks)
BENCH_ENV := $(wildcard $(CURDIR)/../.env.bench)
ifneq ($(BENCH_ENV),)
  include $(BENCH_ENV)
  export
endif

bench: ## Lance le benchmark IDP (alias, usage: make bench PROVIDER=anthropic)
	@$(MAKE) bench-idp $(if $(ALL),ALL=$(ALL)) $(if $(PROVIDER),PROVIDER=$(PROVIDER)) $(if $(MODELS),MODELS=$(MODELS)) $(if $(CASE),CASE=$(CASE)) $(if $(RUNS),RUNS=$(RUNS)) $(if $(OUTPUT),OUTPUT=$(OUTPUT))

bench-idp: ## Benchmark IDP — structuration LLM (OCR blocks → items)
ifdef ALL
	@echo "$(CYAN)Benchmark IDP — tous les modèles...$(RESET)"
	@$(GO) run $(BENCH_CMD) --type=idp --all $(if $(CASE),--case=$(CASE)) $(if $(RUNS),--runs=$(RUNS)) $(if $(OUTPUT),--output=$(OUTPUT))
else ifdef MODELS
	@echo "$(CYAN)Benchmark IDP — $(MODELS)...$(RESET)"
	@$(GO) run $(BENCH_CMD) --type=idp --models=$(MODELS) $(if $(CASE),--case=$(CASE)) $(if $(RUNS),--runs=$(RUNS)) $(if $(OUTPUT),--output=$(OUTPUT))
else ifdef PROVIDER
	@echo "$(CYAN)Benchmark IDP — $(PROVIDER)...$(RESET)"
	@$(GO) run $(BENCH_CMD) --type=idp --provider=$(PROVIDER) $(if $(CASE),--case=$(CASE)) $(if $(RUNS),--runs=$(RUNS)) $(if $(OUTPUT),--output=$(OUTPUT))
else
	@echo "$(YELLOW)Usage :$(RESET)"
	@echo "  make bench-idp ALL=1                                    — tous les modèles"
	@echo "  make bench-idp PROVIDER=anthropic                       — tous les modèles d'un provider"
	@echo "  make bench-idp MODELS=gpt-4o,claude-sonnet-4-6          — sélection de modèles"
	@echo ""
	@echo "$(YELLOW)Modèles disponibles :$(RESET)"
	@$(GO) run $(BENCH_CMD) --list-models
endif

bench-ocr: ## Benchmark OCR — extraction vision (images → OCR blocks)
ifdef ALL
	@echo "$(CYAN)Benchmark OCR — tous les modèles vision...$(RESET)"
	@$(GO) run $(BENCH_CMD) --type=ocr --all $(if $(CASE),--case=$(CASE)) $(if $(RUNS),--runs=$(RUNS)) $(if $(OUTPUT),--output=$(OUTPUT))
else ifdef MODELS
	@echo "$(CYAN)Benchmark OCR — $(MODELS)...$(RESET)"
	@$(GO) run $(BENCH_CMD) --type=ocr --models=$(MODELS) $(if $(CASE),--case=$(CASE)) $(if $(RUNS),--runs=$(RUNS)) $(if $(OUTPUT),--output=$(OUTPUT))
else ifdef PROVIDER
	@echo "$(CYAN)Benchmark OCR — $(PROVIDER)...$(RESET)"
	@$(GO) run $(BENCH_CMD) --type=ocr --provider=$(PROVIDER) $(if $(CASE),--case=$(CASE)) $(if $(RUNS),--runs=$(RUNS)) $(if $(OUTPUT),--output=$(OUTPUT))
else
	@echo "$(YELLOW)Usage :$(RESET)"
	@echo "  make bench-ocr ALL=1                                    — tous les modèles vision"
	@echo "  make bench-ocr PROVIDER=anthropic                       — tous les modèles d'un provider"
	@echo "  make bench-ocr MODELS=gpt-4o,claude-sonnet-4-6          — sélection de modèles"
	@echo ""
	@echo "$(YELLOW)Modèles vision disponibles :$(RESET)"
	@$(GO) run $(BENCH_CMD) --list-models
endif

bench-all: ## Benchmark tous les modèles (IDP)
	@$(MAKE) bench-idp ALL=1

bench-models: ## Liste les modèles disponibles
	@$(GO) run $(BENCH_CMD) --list-models

bench-report: ## Génère le rapport HTML depuis les derniers résultats
	@echo "$(CYAN)Génération du rapport HTML...$(RESET)"
	@$(GO) run $(BENCH_CMD) --report $(if $(RUN),--report-run=$(RUN)) $(if $(REPORT_OUTPUT),--report-output=$(REPORT_OUTPUT))
	@echo "$(GREEN)Rapport généré.$(RESET)"

# ------------------------------------------------------------
# Docker
# ------------------------------------------------------------

docker-build: ## Build l'image Docker du backend
	docker build -t revisemieux-backend .

# ------------------------------------------------------------
# Clean
# ------------------------------------------------------------

clean: ## Supprime les artefacts de build
	@rm -rf $(BUILD_DIR)
	@echo "$(GREEN)Clean OK$(RESET)"
