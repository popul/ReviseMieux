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
MODEL       ?= claude-opus-4-6-20250219
MAX_TOKENS  ?= 8192
CASE        ?= 10_SVT_cours_louis
CASES_DIR   := cases
CASE_DIR    := $(CASES_DIR)/$(CASE)
IMAGES_DIR  := $(CASE_DIR)/images
OUTPUT_DIR  := $(CASE_DIR)
SCRIPT      := scripts/generate_json.py

# Vérifie que la clé API est définie
check-api-key:
	@if [ -z "$$ANTHROPIC_API_KEY" ]; then \
		echo "❌ ANTHROPIC_API_KEY non définie. Export-la d'abord :"; \
		echo "   export ANTHROPIC_API_KEY=sk-ant-..."; \
		exit 1; \
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
generate: generate-input generate-golden generate-metadata ## Générer les 3 fichiers JSON pour CASE
	@echo ""
	@echo "✅ Génération terminée pour $(CASE)."
	@echo "   → $(OUTPUT_DIR)/input.json"
	@echo "   → $(OUTPUT_DIR)/golden_output.json"
	@echo "   → $(OUTPUT_DIR)/metadata.json"
	@echo ""
	@echo "⚠️  Ces fichiers sont des BROUILLONS. Relis-les et corrige-les à la main."

.PHONY: generate-input
generate-input: check-api-key check-case ## Générer input.json (transcription OCR) pour CASE
	@echo "🔄 Génération de input.json pour $(CASE)..."
	@python3 $(SCRIPT) --case $(CASE) --target input --model $(MODEL) --max-tokens $(MAX_TOKENS) --cases-dir $(CASES_DIR)
	@echo "   ✅ $(OUTPUT_DIR)/input.json créé ($$(jq '.blocks | length' $(OUTPUT_DIR)/input.json) blocs)"

.PHONY: generate-golden
generate-golden: check-api-key check-case ## Générer golden_output.json (items pédagogiques) pour CASE
	@echo "🔄 Génération de golden_output.json pour $(CASE)..."
	@python3 $(SCRIPT) --case $(CASE) --target golden --model $(MODEL) --max-tokens $(MAX_TOKENS) --cases-dir $(CASES_DIR)
	@echo "   ✅ $(OUTPUT_DIR)/golden_output.json créé ($$(jq '.items | length' $(OUTPUT_DIR)/golden_output.json) items)"

.PHONY: generate-metadata
generate-metadata: check-api-key check-case ## Générer metadata.json pour CASE
	@echo "🔄 Génération de metadata.json pour $(CASE)..."
	@python3 $(SCRIPT) --case $(CASE) --target metadata --model $(MODEL) --max-tokens $(MAX_TOKENS) --cases-dir $(CASES_DIR)
	@echo "   ✅ $(OUTPUT_DIR)/metadata.json créé"

##@ Validation

.PHONY: validate
validate: ## Valider la structure JSON des fichiers générés pour CASE
	@echo "🔍 Validation des fichiers pour $(CASE)..."
	@errors=0; \
	for f in input.json golden_output.json metadata.json; do \
		path="$(OUTPUT_DIR)/$$f"; \
		if [ ! -f "$$path" ]; then \
			echo "   ⚠️  $$f manquant"; \
			errors=$$((errors+1)); \
			continue; \
		fi; \
		if ! jq empty "$$path" 2>/dev/null; then \
			echo "   ❌ $$f : JSON invalide"; \
			errors=$$((errors+1)); \
		else \
			echo "   ✅ $$f : JSON valide"; \
		fi; \
	done; \
	if [ "$$errors" -gt 0 ]; then \
		echo ""; \
		echo "❌ $$errors erreur(s) trouvée(s)."; \
		exit 1; \
	else \
		echo ""; \
		echo "✅ Tous les fichiers sont valides."; \
	fi
	@echo ""
	@echo "Résumé $(CASE) :"
	@[ -f "$(OUTPUT_DIR)/input.json" ] && echo "   input.json        : $$(jq '.blocks | length' $(OUTPUT_DIR)/input.json) blocs" || true
	@[ -f "$(OUTPUT_DIR)/golden_output.json" ] && echo "   golden_output.json: $$(jq '.items | length' $(OUTPUT_DIR)/golden_output.json) items, $$(jq '.notions | length' $(OUTPUT_DIR)/golden_output.json) notions" || true

.PHONY: list-cases
list-cases: ## Lister tous les cas disponibles avec leur statut
	@echo "Cas disponibles :"
	@for d in $(CASES_DIR)/*/; do \
		case_name=$$(basename "$$d"); \
		[ "$$case_name" = "*" ] && continue; \
		imgs=$$(ls "$$d/images/"*.jpeg "$$d/images/"*.jpg 2>/dev/null | wc -l); \
		has_input="-"; [ -f "$$d/input.json" ] && has_input="✅"; \
		has_golden="-"; [ -f "$$d/golden_output.json" ] && has_golden="✅"; \
		has_meta="-"; [ -f "$$d/metadata.json" ] && has_meta="✅"; \
		printf "   %-25s  📷 %d  input:%s  golden:%s  meta:%s\n" "$$case_name" "$$imgs" "$$has_input" "$$has_golden" "$$has_meta"; \
	done

##@ Nettoyage

.PHONY: clean
clean: ## Supprimer les JSON générés pour CASE (garde les images)
	@echo "🗑  Suppression des JSON pour $(CASE)..."
	@rm -f $(OUTPUT_DIR)/input.json $(OUTPUT_DIR)/golden_output.json $(OUTPUT_DIR)/metadata.json
	@echo "   ✅ Nettoyé."

##@ Aide

.PHONY: help
help: ## Afficher cette aide
	@echo ""
	@echo "Makefile — Génération JSON benchmark via Claude Opus 4.6"
	@echo ""
	@echo "Génère des brouillons de input.json, golden_output.json et metadata.json"
	@echo "à partir des photos de cahier, en utilisant l'API Anthropic (Claude vision)."
	@echo ""
	@echo "⚠️  Les fichiers générés sont des BROUILLONS à relire et corriger manuellement."
	@echo ""
	@echo "Pré-requis : ANTHROPIC_API_KEY, python3, jq"
	@echo ""
	@echo "Usage : make <target> [CASE=nom_du_cas] [MODEL=claude-opus-4-6-20250219]"
	@echo ""
	@awk '/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } \
		/^[a-zA-Z_-]+:.*##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, substr($$0, index($$0, "##") + 3) }' $(MAKEFILE_LIST)
	@echo ""
	@echo "Exemples :"
	@echo "  make generate CASE=10_SVT_cours_louis        # Générer les 3 JSON"
	@echo "  make generate-input CASE=10_SVT_cours_louis  # Seulement input.json"
	@echo "  make validate CASE=10_SVT_cours_louis        # Valider les JSON"
	@echo "  make list-cases                              # Voir tous les cas"
	@echo ""
