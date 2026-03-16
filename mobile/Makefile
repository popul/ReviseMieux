# ============================================================
# Révise Mieux — Mobile Makefile (Expo + React Native)
# ============================================================

.PHONY: help deps dev start android ios web test lint typecheck \
        build clean

.DEFAULT_GOAL := help

# ------------------------------------------------------------
# Variables
# ------------------------------------------------------------

NPM     := npm
NPXE    := npx expo
TSC     := npx tsc

# Couleurs
GREEN  := \033[32m
CYAN   := \033[36m
YELLOW := \033[33m
RESET  := \033[0m

# ------------------------------------------------------------
# Help
# ------------------------------------------------------------

help: ## Affiche cette aide
	@echo ""
	@echo "$(CYAN)Mobile$(RESET) — Expo + React Native"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-22s$(RESET) %s\n", $$1, $$2}'
	@echo ""

# ------------------------------------------------------------
# Dependencies
# ------------------------------------------------------------

deps: ## Installe les dépendances npm
	@echo "$(CYAN)npm install...$(RESET)"
	@$(NPM) install
	@echo "$(GREEN)Deps OK$(RESET)"

# ------------------------------------------------------------
# Dev
# ------------------------------------------------------------

dev: ## Lance Expo (mode dev)
	@$(NPXE) start

start: dev ## Alias de dev

android: ## Lance sur Android
	@$(NPXE) start --android

ios: ## Lance sur iOS
	@$(NPXE) start --ios

web: ## Lance sur web
	@$(NPXE) start --web

# ------------------------------------------------------------
# Tests
# ------------------------------------------------------------

test: ## Lance les tests
	@echo "$(CYAN)Tests...$(RESET)"
	@$(NPM) test -- --passWithNoTests
	@echo "$(GREEN)Tests OK$(RESET)"

# ------------------------------------------------------------
# Qualité
# ------------------------------------------------------------

typecheck: ## Vérifie les types TypeScript
	@echo "$(CYAN)TypeScript check...$(RESET)"
	@$(TSC) --noEmit
	@echo "$(GREEN)Types OK$(RESET)"

lint: typecheck ## Lint (TypeScript check)
	@echo "$(GREEN)Lint OK$(RESET)"

# ------------------------------------------------------------
# Build
# ------------------------------------------------------------

build: ## Build Expo (EAS)
	@echo "$(CYAN)Expo export...$(RESET)"
	@$(NPXE) export
	@echo "$(GREEN)Build OK$(RESET)"

# ------------------------------------------------------------
# Clean
# ------------------------------------------------------------

clean: ## Supprime node_modules et caches
	@rm -rf node_modules .expo dist
	@echo "$(GREEN)Clean OK$(RESET)"
