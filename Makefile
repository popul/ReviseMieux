# ============================================================
# Révise Mieux — Makefile racine (monorepo)
# ============================================================
#
# Hiérarchie :
#   Makefile              ← orchestration (ce fichier)
#   backend/Makefile      ← Go API, migrations, tests
#   mobile/Makefile       ← Expo, TypeScript, lint
#
# Usage :
#   make help             — liste toutes les cibles
#   make setup            — installe tout (deps + DB + migrate)
#   make dev              — lance backend + mobile en parallèle
#   make test             — lance tous les tests
#   make lint             — lint backend + mobile
#   make clean            — nettoyage
#
# ============================================================

.PHONY: help setup dev stop test test-unit test-integration lint fmt clean \
        backend-% mobile-% infra-up infra-down infra-reset db-migrate db-reset \
        bench bench-all bench-report design-system docs-serve docs-deploy

.DEFAULT_GOAL := help

# ------------------------------------------------------------
# Variables
# ------------------------------------------------------------

COMPOSE     := docker compose
BACKEND_DIR := backend
MOBILE_DIR  := mobile

# Couleurs (désactivées si pas de terminal)
ifneq ($(TERM),)
  GREEN  := \033[32m
  CYAN   := \033[36m
  YELLOW := \033[33m
  RESET  := \033[0m
else
  GREEN  :=
  CYAN   :=
  YELLOW :=
  RESET  :=
endif

# ------------------------------------------------------------
# Help
# ------------------------------------------------------------

help: ## Affiche cette aide
	@echo ""
	@echo "$(CYAN)Révise Mieux$(RESET) — Monorepo"
	@echo ""
	@echo "$(YELLOW)Cibles principales :$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""

# ------------------------------------------------------------
# Setup
# ------------------------------------------------------------

setup: infra-up deps db-migrate ## Setup complet : infra + deps + migrations
	@echo "$(GREEN)Setup terminé.$(RESET)"

deps: ## Installe les dépendances backend + mobile
	@$(MAKE) -C $(BACKEND_DIR) deps
	@$(MAKE) -C $(MOBILE_DIR) deps

# ------------------------------------------------------------
# Infrastructure (Docker Compose)
# ------------------------------------------------------------

infra-up: ## Démarre PostgreSQL + Redis
	@$(COMPOSE) up -d postgres redis
	@echo "$(GREEN)Infra up$(RESET) — Postgres :5432, Redis :6379"

infra-down: ## Arrête l'infra
	@$(COMPOSE) down

infra-reset: ## Détruit volumes et recrée l'infra
	@$(COMPOSE) down -v
	@$(MAKE) infra-up

# ------------------------------------------------------------
# Database
# ------------------------------------------------------------

db-migrate: ## Applique les migrations SQL
	@$(MAKE) -C $(BACKEND_DIR) db-migrate

db-reset: infra-reset db-migrate ## Reset complet DB (destroy + migrate)
	@echo "$(GREEN)DB reset terminé.$(RESET)"

# ------------------------------------------------------------
# Dev
# ------------------------------------------------------------

dev: infra-up ## Lance backend + mobile (ctrl-c pour arrêter)
	@echo "$(CYAN)Lancement backend + mobile...$(RESET)"
	@$(MAKE) -j2 backend-dev mobile-dev

backend-dev: ## Lance le backend (hot reload)
	@$(MAKE) -C $(BACKEND_DIR) dev

mobile-dev: ## Lance Expo
	@$(MAKE) -C $(MOBILE_DIR) dev

stop: infra-down ## Arrête tout

# ------------------------------------------------------------
# Tests
# ------------------------------------------------------------

test: test-unit test-integration ## Tous les tests

test-unit: ## Tests unitaires (backend + mobile)
	@$(MAKE) -C $(BACKEND_DIR) test-unit
	@$(MAKE) -C $(MOBILE_DIR) test

test-integration: infra-up ## Tests d'intégration (nécessite DB)
	@$(MAKE) -C $(BACKEND_DIR) test-integration

# ------------------------------------------------------------
# Qualité
# ------------------------------------------------------------

lint: ## Lint backend + mobile
	@$(MAKE) -C $(BACKEND_DIR) lint
	@$(MAKE) -C $(MOBILE_DIR) lint

fmt: ## Formatte le code
	@$(MAKE) -C $(BACKEND_DIR) fmt

check: ## Gate CI : format + vet + imports domaine + tests
	@$(MAKE) -C $(BACKEND_DIR) check
	@$(MAKE) -C $(MOBILE_DIR) test

# ------------------------------------------------------------
# Build
# ------------------------------------------------------------

build: ## Build backend + mobile
	@$(MAKE) -C $(BACKEND_DIR) build
	@$(MAKE) -C $(MOBILE_DIR) build

# ------------------------------------------------------------
# Benchmark LLM
# ------------------------------------------------------------

bench: ## Lance le benchmark LLM (voir make bench sans args pour l'aide)
	@$(MAKE) -C $(BACKEND_DIR) bench $(if $(ALL),ALL=$(ALL)) $(if $(PROVIDER),PROVIDER=$(PROVIDER)) $(if $(MODELS),MODELS=$(MODELS)) $(if $(CASE),CASE=$(CASE)) $(if $(RUNS),RUNS=$(RUNS)) $(if $(OUTPUT),OUTPUT=$(OUTPUT))

bench-all: ## Benchmark tous les modèles
	@$(MAKE) -C $(BACKEND_DIR) bench-all

bench-report: ## Génère le rapport HTML depuis les derniers résultats
	@$(MAKE) -C $(BACKEND_DIR) bench-report $(if $(RUN),RUN=$(RUN)) $(if $(REPORT_OUTPUT),REPORT_OUTPUT=$(REPORT_OUTPUT))

# ------------------------------------------------------------
# Documentation
# ------------------------------------------------------------

docs-serve: ## Prévisualisation du site docs en local (http://localhost:8000)
	@pip install mkdocs-material -q 2>/dev/null
	@mkdocs serve

docs-deploy: ## Déploie le site docs sur GitHub Pages
	@pip install mkdocs-material -q 2>/dev/null
	@mkdocs gh-deploy --force

# ------------------------------------------------------------
# Fiches de révision (skill /study-guide)
# ------------------------------------------------------------

FICHES_ROOT ?= backend/testdata/benchmark/cases
FICHE_PORT ?= 8080

fiche-serve: ## Sert toutes les fiches de révision (FICHES_ROOT=... FICHE_PORT=8080)
	@printf '<!doctype html><meta charset=utf-8><title>Fiches de révision</title><style>body{font-family:-apple-system,sans-serif;max-width:720px;margin:40px auto;padding:0 20px;background:#FBF8F3;color:#1b1b1b}h1{color:#1A4D4D;border-bottom:3px solid #F5C542;padding-bottom:8px}a{display:block;padding:14px 18px;margin:8px 0;background:#fff;border-left:4px solid #E85D4C;border-radius:6px;text-decoration:none;color:#1A4D4D;box-shadow:0 1px 3px rgba(0,0,0,.05)}a:hover{background:#fff8f6}</style><h1>📚 Fiches de révision disponibles</h1>' > $(FICHES_ROOT)/index.html
	@find $(FICHES_ROOT) -mindepth 2 -maxdepth 2 -name fiche-revision.html | sort | sed "s|$(FICHES_ROOT)/||" | while read f; do \
		dir=$$(dirname "$$f"); \
		printf '<a href="%s">%s</a>' "$$f" "$$dir" >> $(FICHES_ROOT)/index.html; \
	done
	@echo "Index : http://localhost:$(FICHE_PORT)/"
	@LAN_IP=$$(ipconfig getifaddr en0 2>/dev/null || ipconfig getifaddr en1 2>/dev/null); \
		if [ -n "$$LAN_IP" ]; then echo "LAN : http://$$LAN_IP:$(FICHE_PORT)/"; fi
	@cd $(FICHES_ROOT) && python3 -m http.server $(FICHE_PORT) --bind 0.0.0.0

fiche-stop: ## Arrête le serveur fiche (tue le process sur FICHE_PORT)
	@lsof -ti:$(FICHE_PORT) | xargs kill 2>/dev/null && echo "Serveur arrêté." || echo "Aucun serveur sur :$(FICHE_PORT)"

# ------------------------------------------------------------
# Clean
# ------------------------------------------------------------

clean: ## Nettoyage artefacts
	@$(MAKE) -C $(BACKEND_DIR) clean
	@$(MAKE) -C $(MOBILE_DIR) clean
	@echo "$(GREEN)Clean terminé.$(RESET)"

# ------------------------------------------------------------
# Passthrough : make backend-<cible> / mobile-<cible>
# ------------------------------------------------------------

backend-%:
	@$(MAKE) -C $(BACKEND_DIR) $*

mobile-%:
	@$(MAKE) -C $(MOBILE_DIR) $*
