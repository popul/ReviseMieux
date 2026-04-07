#!/usr/bin/env bash
# check-skill-prompts.sh — Garantie exécutable que le skill `study-guide`
# reste cohérent avec les prompts figés dans backend/prompts/.
#
# Vérifie deux invariants :
#   1. Existence : chaque chemin `${CLAUDE_PROJECT_DIR}/backend/prompts/.../*.md`
#      cité dans SKILL.md pointe vers un fichier réel.
#   2. Anti-duplication : aucune ligne-signature d'un prompt figé n'apparaît
#      dans SKILL.md (sinon le contenu a été re-recopié, créant une dette de
#      duplication silencieuse).
#
# Usage : bash backend/scripts/check-skill-prompts.sh
# Retourne 0 si tout est OK, 1 sinon.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SKILL_FILE="$REPO_ROOT/.claude/skills/study-guide/SKILL.md"
PROMPTS_DIR="$REPO_ROOT/backend/prompts"

RED='\033[31m'
GREEN='\033[32m'
YELLOW='\033[33m'
RESET='\033[0m'

errors=0

if [[ ! -f "$SKILL_FILE" ]]; then
    echo -e "${RED}SKILL.md introuvable : $SKILL_FILE${RESET}"
    exit 1
fi

# ------------------------------------------------------------
# 1. Existence des références
# ------------------------------------------------------------
echo -e "${YELLOW}[1/2] Vérification des références prompts dans SKILL.md...${RESET}"

# Extrait toutes les références du type ${CLAUDE_PROJECT_DIR}/backend/prompts/.../*.md
refs=$(grep -oE '\$\{CLAUDE_PROJECT_DIR\}/backend/prompts/[^ )`*]+\.md' "$SKILL_FILE" | sort -u || true)

if [[ -z "$refs" ]]; then
    echo -e "${RED}  ✗ aucune référence à backend/prompts/ trouvée dans SKILL.md${RESET}"
    echo -e "${RED}    le skill devrait pointer vers les prompts figés${RESET}"
    errors=$((errors + 1))
else
    while IFS= read -r ref; do
        # Strip ${CLAUDE_PROJECT_DIR}/ prefix → relative repo path
        rel_path="${ref#\$\{CLAUDE_PROJECT_DIR\}/}"
        abs_path="$REPO_ROOT/$rel_path"
        if [[ -f "$abs_path" ]]; then
            echo -e "  ${GREEN}✓${RESET} $rel_path"
        else
            echo -e "  ${RED}✗${RESET} $rel_path ${RED}(fichier introuvable)${RESET}"
            errors=$((errors + 1))
        fi
    done <<< "$refs"
fi

# ------------------------------------------------------------
# 2. Anti-duplication : aucune ligne-signature des prompts figés dans SKILL.md
# ------------------------------------------------------------
echo -e "${YELLOW}[2/2] Vérification anti-duplication (lignes-signatures)...${RESET}"

# Pour chaque prompt figé, on prend la première ligne non vide non triviale
# (> 40 chars) comme signature. Si elle apparaît dans SKILL.md, c'est qu'on a
# re-recopié le contenu du prompt → drift.
while IFS= read -r prompt_file; do
    rel="${prompt_file#$REPO_ROOT/}"
    # Première ligne non vide d'au moins 40 caractères
    signature=$(awk 'NF && length($0) >= 40 { print; exit }' "$prompt_file")
    if [[ -z "$signature" ]]; then
        continue
    fi
    # Recherche littérale (fixed string) dans SKILL.md
    if grep -Fq "$signature" "$SKILL_FILE"; then
        echo -e "  ${RED}✗${RESET} $rel ${RED}: contenu dupliqué dans SKILL.md${RESET}"
        echo -e "    ${RED}signature trouvée :${RESET} \"${signature:0:80}...\""
        echo -e "    ${YELLOW}→ retire cette duplication, le skill doit seulement référencer le prompt${RESET}"
        errors=$((errors + 1))
    else
        echo -e "  ${GREEN}✓${RESET} $rel (pas de duplication)"
    fi
done < <(find "$PROMPTS_DIR" -type f -name 'system.md' | sort)

# ------------------------------------------------------------
# Résultat
# ------------------------------------------------------------
if [[ $errors -gt 0 ]]; then
    echo -e "${RED}check-skill-prompts: $errors erreur(s)${RESET}"
    exit 1
fi
echo -e "${GREEN}check-skill-prompts: OK${RESET}"
