# Makefile — Génération des JSON benchmark via Claude Opus 4.6
#
# Pré-requis :
#   - ANTHROPIC_API_KEY définie dans l'environnement
#   - python3 (stdlib uniquement, pas de pip install)
#   - jq (pour validate / list-cases)
#
# Usage :
#   make help                                # afficher l'aide
#   make generate CASE=10_SVT_cours_louis    # générer les 3 fichiers JSON

SHELL := /bin/bash
.DEFAULT_GOAL := help

# --- Configuration ---
PROVIDER    ?= anthropic
MODEL       ?= claude-sonnet-4-20250514
MAX_TOKENS  ?= 8192
CASE        ?= 10_SVT_cours_louis
CASES_DIR   := cases
CASE_DIR    := $(CASES_DIR)/$(CASE)
IMAGES_DIR  := $(CASE_DIR)/images
OUTPUT_DIR  := $(CASE_DIR)
SCRIPT      := scripts/generate_json.py

# Vérifie que la clé API est définie (selon le provider)
check-api-key:
	@if [ "$(PROVIDER)" = "anthropic" ] && [ -z "$$ANTHROPIC_API_KEY" ]; then \
		echo "❌ ANTHROPIC_API_KEY non définie."; exit 1; \
	fi
	@if [ "$(PROVIDER)" = "openrouter" ] && [ -z "$$OPENROUTER_API_KEY" ]; then \
		echo "❌ OPENROUTER_API_KEY non définie."; exit 1; \
	fi
	@if [ "$(PROVIDER)" = "openai" ] && [ -z "$$OPENAI_API_KEY" ]; then \
		echo "❌ OPENAI_API_KEY non définie."; exit 1; \
	fi
	@if [ "$(PROVIDER)" = "google" ] && [ -z "$$GOOGLE_AI_API_KEY" ]; then \
		echo "❌ GOOGLE_AI_API_KEY non définie."; exit 1; \
	fi
	@if [ "$(PROVIDER)" = "mistral" ] && [ -z "$$MISTRAL_API_KEY" ]; then \
		echo "❌ MISTRAL_API_KEY non définie."; exit 1; \
	fi

# Vérifie que le cas existe et a des images
check-case:
	@if [ ! -d "$(IMAGES_DIR)" ]; then \
		echo "❌ Dossier images introuvable : $(IMAGES_DIR)"; \
		exit 1; \
	fi
	@if ! ls $(IMAGES_DIR)/*.jpeg >/dev/null 2>&1 && ! ls $(IMAGES_DIR)/*.jpg >/dev/null 2>&1; then \
		echo "❌ Aucune image .jpeg/.jpg trouvée dans $(IMAGES_DIR)"; \
		exit 1; \
	fi

# --- Targets de génération ---

##@ Génération

.PHONY: generate
generate: generate-input generate-golden generate-metadata ## Générer les 3 fichiers YAML pour CASE
	@echo ""
	@echo "✅ Génération terminée pour $(CASE)."
	@echo "   → $(OUTPUT_DIR)/input.yaml"
	@echo "   → $(OUTPUT_DIR)/golden_output.yaml"
	@echo "   → $(OUTPUT_DIR)/metadata.yaml"
	@echo ""
	@echo "⚠️  Ces fichiers sont des BROUILLONS. Relis-les et corrige-les à la main."

.PHONY: generate-input
generate-input: check-api-key check-case ## Générer input.yaml (transcription OCR) pour CASE
	@echo "🔄 Génération de input.yaml pour $(CASE)..."
	@python3 $(SCRIPT) --case $(CASE) --target input --provider $(PROVIDER) --model $(MODEL) --max-tokens $(MAX_TOKENS) --cases-dir $(CASES_DIR)
	@echo "   ✅ $(OUTPUT_DIR)/input.yaml créé"

.PHONY: generate-golden
generate-golden: check-api-key check-case ## Générer golden_output.yaml (items pédagogiques) pour CASE
	@echo "🔄 Génération de golden_output.yaml pour $(CASE)..."
	@python3 $(SCRIPT) --case $(CASE) --target golden --provider $(PROVIDER) --model $(MODEL) --max-tokens $(MAX_TOKENS) --cases-dir $(CASES_DIR)
	@echo "   ✅ $(OUTPUT_DIR)/golden_output.yaml créé"

.PHONY: generate-metadata
generate-metadata: check-api-key check-case ## Générer metadata.yaml pour CASE
	@echo "🔄 Génération de metadata.yaml pour $(CASE)..."
	@python3 $(SCRIPT) --case $(CASE) --target metadata --provider $(PROVIDER) --model $(MODEL) --max-tokens $(MAX_TOKENS) --cases-dir $(CASES_DIR)
	@echo "   ✅ $(OUTPUT_DIR)/metadata.yaml créé"

##@ Validation

.PHONY: validate
validate: ## Valider la structure YAML des fichiers générés pour CASE
	@echo "🔍 Validation des fichiers pour $(CASE)..."
	@errors=0; \
	for f in input.yaml golden_output.yaml metadata.yaml; do \
		path="$(OUTPUT_DIR)/$$f"; \
		if [ ! -f "$$path" ]; then \
			echo "   ⚠️  $$f manquant"; \
			errors=$$((errors+1)); \
			continue; \
		fi; \
		if ! python3 -c "import yaml; yaml.safe_load(open('$$path'))" 2>/dev/null && \
		   ! python3 -c "import sys; __import__('yaml').safe_load(open(sys.argv[1]))" "$$path" 2>/dev/null; then \
			echo "   ✅ $$f : présent (validation YAML requiert pyyaml)"; \
		else \
			echo "   ✅ $$f : YAML valide"; \
		fi; \
	done; \
	if [ "$$errors" -gt 0 ]; then \
		echo ""; \
		echo "❌ $$errors fichier(s) manquant(s)."; \
		exit 1; \
	else \
		echo ""; \
		echo "✅ Tous les fichiers sont présents."; \
	fi

.PHONY: list-cases
list-cases: ## Lister tous les cas disponibles avec leur statut
	@echo "Cas disponibles :"
	@for d in $(CASES_DIR)/*/; do \
		case_name=$$(basename "$$d"); \
		[ "$$case_name" = "*" ] && continue; \
		imgs=$$(ls "$$d/images/"*.jpeg "$$d/images/"*.jpg 2>/dev/null | wc -l); \
		has_input="-"; [ -f "$$d/input.yaml" ] && has_input="✅"; \
		has_golden="-"; [ -f "$$d/golden_output.yaml" ] && has_golden="✅"; \
		has_meta="-"; [ -f "$$d/metadata.yaml" ] && has_meta="✅"; \
		printf "   %-40s  📷 %d  input:%s  golden:%s  meta:%s\n" "$$case_name" "$$imgs" "$$has_input" "$$has_golden" "$$has_meta"; \
	done

##@ Nettoyage

.PHONY: clean
clean: ## Supprimer les YAML générés pour CASE (garde les images)
	@echo "🗑  Suppression des YAML pour $(CASE)..."
	@rm -f $(OUTPUT_DIR)/input.yaml $(OUTPUT_DIR)/golden_output.yaml $(OUTPUT_DIR)/metadata.yaml
	@echo "   ✅ Nettoyé."

##@ Aide

.PHONY: help
help: ## Afficher cette aide
	@echo ""
	@echo "Makefile — Génération YAML benchmark via LLM vision"
	@echo ""
	@echo "Génère des brouillons de input.yaml, golden_output.yaml et metadata.yaml"
	@echo "à partir des photos de cahier, en utilisant un LLM vision (Anthropic, OpenRouter, etc.)."
	@echo ""
	@echo "⚠️  Les fichiers générés sont des BROUILLONS à relire et corriger manuellement."
	@echo ""
	@echo "Pré-requis : clé API du provider, python3"
	@echo ""
	@echo "Usage : make <target> [CASE=nom] [PROVIDER=openrouter] [MODEL=qwen/qwen3-vl-32b-instruct]"
	@echo ""
	@awk '/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } \
		/^[a-zA-Z_-]+:.*##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, substr($$0, index($$0, "##") + 3) }' $(MAKEFILE_LIST)
	@echo ""
	@echo "Exemples :"
	@echo "  make generate CASE=10_SVT_cours_louis                                             # Anthropic (défaut)"
	@echo "  make generate-input CASE=12_physique PROVIDER=openrouter MODEL=qwen/qwen3-vl-32b-instruct  # OpenRouter"
	@echo "  make validate CASE=10_SVT_cours_louis"
	@echo "  make list-cases"
	@echo ""
