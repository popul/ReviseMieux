# Makefile — Génération des JSON benchmark via Claude Opus 4.6
#
# Pré-requis :
#   - ANTHROPIC_API_KEY définie dans l'environnement
#   - curl, jq, base64 installés
#
# Usage :
#   make help                          # afficher l'aide
#   make generate CASE=10_SVT_cours_louis  # générer les 3 fichiers JSON

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

# --- Fonctions internes ---

# Construit le tableau JSON des images encodées en base64 pour l'API
define build_images_json
$(shell \
	imgs=""; \
	for f in $(IMAGES_DIR)/*.jpeg $(IMAGES_DIR)/*.jpg; do \
		[ -f "$$f" ] || continue; \
		b64=$$(base64 -w0 "$$f"); \
		mt="image/jpeg"; \
		imgs="$$imgs{\"type\":\"image\",\"source\":{\"type\":\"base64\",\"media_type\":\"$$mt\",\"data\":\"$$b64\"}},"; \
	done; \
	echo "$$imgs" \
)
endef

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
	@IMAGES_CONTENT='$(build_images_json)' && \
	jq -n --argjson imgs "[$$IMAGES_CONTENT{\"type\":\"text\",\"text\":\"dummy\"}]" '$$imgs | del(.[-1])' > /dev/null 2>&1; \
	PROMPT='Tu es un expert en transcription de cahiers scolaires. Analyse ces photos de cahier et produis une transcription structurée en blocs.\n\nRègles :\n- Transcris mot à mot, y compris les fautes.\n- Chaque unité de sens = un bloc séparé.\n- Types de blocs : TEXT (texte, titres, questions, réponses), DIAGRAM (schémas, graphiques), TABLE (tableaux).\n- Pour les éléments visuels, décris entre crochets : [Schéma : description détaillée]\n- Préserve la structure (numéros, tirets, retours à la ligne).\n- Mots illisibles : [illisible]\n- confidence: 1.0 pour tous les blocs.\n\nRéponds UNIQUEMENT avec le JSON valide, sans markdown, sans commentaire :\n{\n  \"blocks\": [\n    {\"text\": \"...\", \"block_type\": \"TEXT|DIAGRAM|TABLE\", \"confidence\": 1.0}\n  ]\n}'; \
	IMAGES_JSON=$$(python3 -c " \
import base64, json, glob, sys; \
imgs = sorted(glob.glob('$(IMAGES_DIR)/*.jpeg') + glob.glob('$(IMAGES_DIR)/*.jpg')); \
content = []; \
[content.append({'type':'image','source':{'type':'base64','media_type':'image/jpeg','data':base64.b64encode(open(f,'rb').read()).decode()}}) for f in imgs]; \
content.append({'type':'text','text':'''$$PROMPT'''}); \
payload = {'model':'$(MODEL)','max_tokens':$(MAX_TOKENS),'messages':[{'role':'user','content':content}]}; \
json.dump(payload, sys.stdout)") && \
	RESPONSE=$$(curl -s https://api.anthropic.com/v1/messages \
		-H "content-type: application/json" \
		-H "x-api-key: $$ANTHROPIC_API_KEY" \
		-H "anthropic-version: 2023-06-01" \
		-d "$$IMAGES_JSON") && \
	echo "$$RESPONSE" | jq -r '.content[0].text' | sed 's/^```json//;s/^```//' | jq '.' > $(OUTPUT_DIR)/input.json && \
	echo "   ✅ $(OUTPUT_DIR)/input.json créé ($$(jq '.blocks | length' $(OUTPUT_DIR)/input.json) blocs)"

.PHONY: generate-golden
generate-golden: check-api-key check-case ## Générer golden_output.json (items pédagogiques) pour CASE
	@echo "🔄 Génération de golden_output.json pour $(CASE)..."
	@if [ ! -f "$(OUTPUT_DIR)/input.json" ]; then \
		echo "⚠️  input.json absent — génération depuis les images directement"; \
	fi
	@IMAGES_JSON=$$(python3 -c " \
import base64, json, glob, sys; \
imgs = sorted(glob.glob('$(IMAGES_DIR)/*.jpeg') + glob.glob('$(IMAGES_DIR)/*.jpg')); \
content = []; \
[content.append({'type':'image','source':{'type':'base64','media_type':'image/jpeg','data':base64.b64encode(open(f,'rb').read()).decode()}}) for f in imgs]; \
input_ctx = ''; \
try: \
    with open('$(OUTPUT_DIR)/input.json') as f: input_ctx = '\n\nTranscription déjà réalisée (input.json) :\n' + f.read() \
except: pass; \
prompt = '''Tu es un expert pédagogique. À partir de ces photos de cahier scolaire (cours de SVT), extrais les items de révision.\n\nTypes d items :\n- KNOWLEDGE : fait, définition, règle (champs: type, term, keywords, notion)\n- PROCEDURE : méthode, protocole (champs: type, term, keywords, steps, notion)\n- DOCUMENT : schéma/graphique à savoir lire (champs: type, term, keywords, notion)\n\nRègles :\n- 3 à 5 keywords par item\n- notion = concept chapeau (2-4 notions par chapitre)\n- steps uniquement pour PROCEDURE (3-6 étapes)\n- Vise 6-10 items\n- notions en bas = liste des notions uniques''' + input_ctx + '''\n\nRéponds UNIQUEMENT avec le JSON valide, sans markdown :\n{\n  \"items\": [{\"type\": \"...\", \"term\": \"...\", \"keywords\": [...], \"notion\": \"...\"}],\n  \"notions\": [\"Notion1\", \"Notion2\"]\n}'''; \
content.append({'type':'text','text': prompt}); \
payload = {'model':'$(MODEL)','max_tokens':$(MAX_TOKENS),'messages':[{'role':'user','content':content}]}; \
json.dump(payload, sys.stdout)") && \
	RESPONSE=$$(curl -s https://api.anthropic.com/v1/messages \
		-H "content-type: application/json" \
		-H "x-api-key: $$ANTHROPIC_API_KEY" \
		-H "anthropic-version: 2023-06-01" \
		-d "$$IMAGES_JSON") && \
	echo "$$RESPONSE" | jq -r '.content[0].text' | sed 's/^```json//;s/^```//' | jq '.' > $(OUTPUT_DIR)/golden_output.json && \
	echo "   ✅ $(OUTPUT_DIR)/golden_output.json créé ($$(jq '.items | length' $(OUTPUT_DIR)/golden_output.json) items)"

.PHONY: generate-metadata
generate-metadata: check-api-key check-case ## Générer metadata.json pour CASE
	@echo "🔄 Génération de metadata.json pour $(CASE)..."
	@PAGES=$$(ls $(IMAGES_DIR)/*.jpeg $(IMAGES_DIR)/*.jpg 2>/dev/null | wc -l); \
	IMAGES_JSON=$$(python3 -c " \
import base64, json, glob, sys; \
imgs = sorted(glob.glob('$(IMAGES_DIR)/*.jpeg') + glob.glob('$(IMAGES_DIR)/*.jpg')); \
content = []; \
[content.append({'type':'image','source':{'type':'base64','media_type':'image/jpeg','data':base64.b64encode(open(f,'rb').read()).decode()}}) for f in imgs]; \
golden_ctx = ''; \
try: \
    with open('$(OUTPUT_DIR)/golden_output.json') as f: golden_ctx = '\n\nItems extraits (golden_output.json) :\n' + f.read() \
except: pass; \
prompt = '''Analyse ces photos de cahier scolaire et génère le fichier metadata.json.\n\nChamps à remplir :\n- id: \"$(CASE)\"\n- subject: déduis la matière des photos\n- level: classe (6e, 5e, 4e, 3e)\n- topic: titre du chapitre\n- difficulty: low/medium/high (pour la difficulté OCR)\n- has_images: true\n- expected_item_count: nombre d items dans golden_output si disponible\n- expected_types: décompte par type (KNOWLEDGE, PROCEDURE, DOCUMENT)\n- ocr_source: \"human_expert\"\n- ocr_author: \"Claude Opus 4.6 (draft)\"\n- ocr_date: \"$(shell date +%Y-%m-%d)\"\n- ocr_metadata.handwriting_quality: soigné/normal/brouillon\n- ocr_metadata.ink_types: [blue_pen, red_pen, black_pen, pencil...]\n- ocr_metadata.photo_quality: good/average/poor\n- ocr_metadata.content_types: [text, diagram, table, formula, map]\n- ocr_metadata.support: pure_notebook/notebook_with_printout/printout_annotated\n- ocr_metadata.pages: $(shell ls $(IMAGES_DIR)/*.jpeg $(IMAGES_DIR)/*.jpg 2>/dev/null | wc -l | tr -d " ")\n- notes: description libre''' + golden_ctx + '''\n\nRéponds UNIQUEMENT avec le JSON valide, sans markdown.'''; \
content.append({'type':'text','text': prompt}); \
payload = {'model':'$(MODEL)','max_tokens':$(MAX_TOKENS),'messages':[{'role':'user','content':content}]}; \
json.dump(payload, sys.stdout)") && \
	RESPONSE=$$(curl -s https://api.anthropic.com/v1/messages \
		-H "content-type: application/json" \
		-H "x-api-key: $$ANTHROPIC_API_KEY" \
		-H "anthropic-version: 2023-06-01" \
		-d "$$IMAGES_JSON") && \
	echo "$$RESPONSE" | jq -r '.content[0].text' | sed 's/^```json//;s/^```//' | jq '.' > $(OUTPUT_DIR)/metadata.json && \
	echo "   ✅ $(OUTPUT_DIR)/metadata.json créé"

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
	@echo "Pré-requis : ANTHROPIC_API_KEY, curl, jq, python3"
	@echo ""
	@echo "Usage : make <target> [CASE=nom_du_cas] [MODEL=claude-opus-4-6-20250219]"
	@echo ""
	@awk '/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } \
		/^[a-zA-Z_-]+:.*##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, substr($$0, index($$0, "##") + 3) }' $(MAKEFILE_LIST)
	@echo ""
	@echo "Exemples :"
	@echo "  make generate CASE=10_SVT_cours_louis    # Générer les 3 JSON"
	@echo "  make generate-input CASE=10_SVT_cours_louis  # Seulement input.json"
	@echo "  make validate CASE=10_SVT_cours_louis    # Valider les JSON"
	@echo "  make list-cases                          # Voir tous les cas"
	@echo ""
