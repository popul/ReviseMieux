# ============================================================
# Révise Mieux — Mobile Makefile (Expo + React Native)
# ============================================================

.PHONY: help deps dev start android ios web test test-watch lint typecheck \
        build build-dev build-dev-sim build-preview \
        device-register eas-login eas-whoami \
        clean prebuild nuke

.DEFAULT_GOAL := help

# ------------------------------------------------------------
# Variables
# ------------------------------------------------------------

NPM     := npm
NPXE    := npx expo
NPXEAS  := npx eas
TSC     := npx tsc

# Couleurs
ifneq ($(TERM),)
  GREEN  := \033[32m
  CYAN   := \033[36m
  YELLOW := \033[33m
  RED    := \033[31m
  RESET  := \033[0m
else
  GREEN  :=
  CYAN   :=
  YELLOW :=
  RED    :=
  RESET  :=
endif

# ------------------------------------------------------------
# Help
# ------------------------------------------------------------

help: ## Affiche cette aide
	@echo ""
	@echo "$(CYAN)Mobile$(RESET) — Expo 55 + React Native"
	@echo ""
	@echo "$(YELLOW)Démarrage rapide :$(RESET)"
	@echo "  1. make deps"
	@echo "  2. make build-dev          (build iPhone — première fois uniquement)"
	@echo "  3. make dev                (serveur Metro)"
	@echo ""
	@echo "$(YELLOW)Toutes les cibles :$(RESET)"
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

dev: ## Lance Expo (dev client, même réseau Wi-Fi)
	@$(NPXE) start --dev-client

start: dev ## Alias de dev

android: ## Lance sur Android
	@$(NPXE) start --android

ios: ## Lance sur iOS simulator
	@$(NPXE) start --ios

web: ## Lance sur web
	@$(NPXE) start --web

# ------------------------------------------------------------
# EAS — Authentification & Devices
# ------------------------------------------------------------

eas-login: ## Connexion au compte Expo (EAS)
	@$(NPXEAS) login

eas-whoami: ## Vérifie le compte Expo connecté
	@$(NPXEAS) whoami

device-register: ## Enregistre un iPhone pour les builds dev (QR code)
	@echo "$(CYAN)Scanne le QR code avec ton iPhone pour l'enregistrer...$(RESET)"
	@$(NPXEAS) device:create

# ------------------------------------------------------------
# EAS Build — Builds natifs
# ------------------------------------------------------------

build-dev: ## Build dev client pour iPhone physique (EAS Cloud)
	@echo "$(CYAN)Build dev client iOS (device)...$(RESET)"
	@$(NPXEAS) build --profile development --platform ios
	@echo "$(GREEN)Build terminé — installe l'app via le lien ci-dessus$(RESET)"

build-dev-sim: ## Build dev client pour simulateur iOS
	@echo "$(CYAN)Build dev client iOS (simulator)...$(RESET)"
	@$(NPXEAS) build --profile development-simulator --platform ios

build-preview: ## Build preview (TestFlight-ready, sans dev tools)
	@echo "$(CYAN)Build preview iOS...$(RESET)"
	@$(NPXEAS) build --profile preview --platform ios

build-android: ## Build dev client Android (EAS Cloud)
	@echo "$(CYAN)Build dev client Android...$(RESET)"
	@$(NPXEAS) build --profile development --platform android

# ------------------------------------------------------------
# Tests
# ------------------------------------------------------------

test: ## Lance les tests Jest
	@echo "$(CYAN)Tests...$(RESET)"
	@$(NPM) test -- --passWithNoTests
	@echo "$(GREEN)Tests OK$(RESET)"

test-watch: ## Tests en mode watch
	@$(NPM) test -- --watch

# ------------------------------------------------------------
# Simulateur
# ------------------------------------------------------------

SIM_DEVICE ?= iPhone 17 Pro

sim: ## Demarre le simulateur iOS
	@xcrun simctl boot "$(SIM_DEVICE)" 2>/dev/null || true
	@open -a Simulator
	@echo "$(GREEN)Simulateur $(SIM_DEVICE) ready$(RESET)"

sim-install: sim ## Build + installe l'app sur le simulateur
	@echo "$(CYAN)Expo prebuild + build iOS...$(RESET)"
	@API_URL=http://localhost:8080/api/v1 npx expo prebuild --platform ios --no-install 2>/dev/null
	@cd ios && pod install --silent 2>/dev/null && \
		xcodebuild -workspace RviseMieux.xcworkspace -scheme RviseMieux -configuration Debug \
		-destination "platform=iOS Simulator,name=$(SIM_DEVICE)" -derivedDataPath build build 2>&1 | tail -1
	@xcrun simctl install booted ios/build/Build/Products/Debug-iphonesimulator/RviseMieux.app
	@echo "$(GREEN)App installee sur $(SIM_DEVICE)$(RESET)"

sim-launch: ## Lance l'app sur le simulateur (via Metro)
	@npx expo start --ios --port 8083

# ------------------------------------------------------------
# E2E (Maestro)
# ------------------------------------------------------------

e2e: ## Lance tous les tests E2E Maestro
	@echo "$(CYAN)E2E Maestro...$(RESET)"
	@maestro test e2e/flows/
	@echo "$(GREEN)E2E OK$(RESET)"

e2e-flow: ## Lance un flow et genere la video (usage: make e2e-flow FLOW=01-onboarding)
	@if [ -z "$(FLOW)" ]; then echo "$(RED)Usage: make e2e-flow FLOW=01-onboarding$(RESET)"; exit 1; fi
	@mkdir -p e2e/videos
	@echo "$(CYAN)Maestro flow + video: $(FLOW)$(RESET)"
	@rm -f e2e/videos/$(FLOW).mp4
	@xcrun simctl io booted recordVideo --codec h264 "$$(pwd)/e2e/videos/$(FLOW).mp4" & \
		RPID=$$!; \
		sleep 2; \
		maestro test e2e/flows/$(FLOW).yaml || true; \
		sleep 1; \
		kill -INT $$RPID 2>/dev/null; \
		wait $$RPID 2>/dev/null
	@echo "$(GREEN)Video: e2e/videos/$(FLOW).mp4$(RESET)"
	@open e2e/videos/$(FLOW).mp4

e2e-record: ## Enregistre les videos de tous les flows
	@echo "$(CYAN)E2E Maestro (record all)...$(RESET)"
	@mkdir -p e2e/videos
	@for f in e2e/flows/*.yaml; do \
		name=$$(basename "$$f" .yaml); \
		echo "  Recording $$name..."; \
		maestro record --local "$$f" "e2e/videos/$$name.mp4" 2>&1 || true; \
	done
	@echo "$(GREEN)Videos dans e2e/videos/$(RESET)"

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
# Build local
# ------------------------------------------------------------

build: ## Export Expo (bundle web/static)
	@echo "$(CYAN)Expo export...$(RESET)"
	@$(NPXE) export
	@echo "$(GREEN)Build OK$(RESET)"

prebuild: ## Génère les projets natifs ios/ et android/ (pour debug natif)
	@echo "$(CYAN)Expo prebuild...$(RESET)"
	@$(NPXE) prebuild
	@echo "$(GREEN)Prebuild OK$(RESET)"

# ------------------------------------------------------------
# Clean
# ------------------------------------------------------------

clean: ## Supprime caches Expo et dist
	@rm -rf .expo dist
	@echo "$(GREEN)Clean OK$(RESET)"

nuke: ## Supprime tout (node_modules, caches, natifs)
	@rm -rf node_modules .expo dist ios android
	@echo "$(GREEN)Nuke OK — relance make deps$(RESET)"
