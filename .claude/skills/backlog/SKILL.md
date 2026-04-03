---
name: backlog
description: "Analyse complète du projet et alimentation du backlog GitHub avec priorisation"
disable-model-invocation: true
hooks:
  Stop:
    - type: command
      command: "${CLAUDE_SKILL_DIR}/scripts/generate-report.sh"
---

# Analyse complète et alimentation du backlog — `/backlog`

Tu es un product/tech lead qui analyse l'état complet du projet Révise Mieux pour identifier les lacunes, la dette technique, et les opportunités. Tu produis un backlog structuré d'issues GitHub.

## Instructions

Exécute les 5 phases séquentiellement.

---

### PHASE 1 — Collecter l'état du monde

#### 1A — Lire les specs et la vision produit

Lis intégralement :
- `CLAUDE.md` — conventions, architecture, règles
- `docs/PRD.md` — PRD complet (personas, epics, pipeline, algorithmes, SLA, machines à états)
- `docs/MVP-scope.md` — classification 171 ACs, 53 Lot 0
- `docs/lot0-tracker.md` — état d'avancement
- `docs/llm-strategy.md` — stratégie LLM, modèles, monitoring
- `docs/openapi.yaml` — spec API

#### 1B — Lire les docs de conception récentes

Lis tous les fichiers dans :
- `docs/architecture/` — OCR/IDP incrémental, merge rules
- `docs/ux/` — user flow, personas, écrans, edge cases, décisions ouvertes
- `docs/design/` — design system, couleurs, composants
- `docs/process/` — feedback loop
- `docs/admin/` — spec IHM admin
- `docs/e2e-feasibility.md` — étude Maestro

#### 1C — Lire l'historique des issues

```bash
# Issues ouvertes
gh issue list --repo popul/ReviseMieux --state open --json number,title,labels,body

# Issues fermées (pour le contexte)
gh issue list --repo popul/ReviseMieux --state closed --json number,title,labels

# Rapports de feedback-loop
ls reports/feedback-loop-*.md
cat reports/decision-log.md
```

#### 1D — Lire l'état du code

```bash
# Structure backend
ls backend/internal/domain/ backend/internal/app/ backend/internal/infra/ backend/internal/http/handler/

# Tests et coverage
cd backend && go test -coverprofile=coverage.out ./internal/... 2>&1 | grep "coverage:" | sort -t: -k2 -n
go tool cover -func=coverage.out | tail -1

# Tests mobile
cd mobile && npx jest --ci 2>&1 | tail -5

# Git log récent
git log --oneline -20

# Fichiers non testés
find backend/internal -name "*.go" ! -name "*_test.go" | while read f; do
  dir=$(dirname "$f")
  base=$(basename "$f" .go)
  test_file="${dir}/${base}_test.go"
  if [ ! -f "$test_file" ]; then
    echo "NO_TEST: $f"
  fi
done | head -30
```

#### 1E — Lire les ACs non couvertes

```bash
# ACs dans le tracker sans test dédié
comm -23 \
  <(grep -oE "Z[0-9]+-AC[0-9]+[a-z]*" docs/lot0-tracker.md | sort -u) \
  <(cd backend && go test ./... -v -count=1 2>&1 | grep -oE "Z[0-9]+AC[0-9]+[a-z]*" | sed 's/Z\([0-9]*\)AC/Z\1-AC/' | sort -u)
```

---

### PHASE 2 — Identifier les lacunes

Croise toutes les données collectées pour identifier :

#### Catégorie 1 — Dette technique

- Packages à 0% coverage (infra/, middleware, config)
- Code sans tests d'intégration
- TODO/FIXME/HACK dans le code
- Dépendances obsolètes ou vulnérables
- Imports non utilisés ou code mort
- Violations d'architecture (imports interdits dans domain/)
- Erreurs non wrappées
- Constructeurs sans validation d'invariants

#### Catégorie 2 — Gaps fonctionnels

- ACs Lot 0 marquées `[x]` mais mal implémentées
- ACs hors Lot 0 mais nécessaires pour la v1
- Fonctionnalités du PRD non encore planifiées
- Endpoints API manquants (spec admin, openapi vs router)
- Écrans mobile manquants ou incomplets
- Edge cases documentés dans `docs/ux/edge-cases.md` mais non implémentés

#### Catégorie 3 — Infra et DevOps

- CI/CD manquant (tests intégration, E2E, deploy)
- Monitoring non implémenté (métriques LLM, alertes)
- Logs structurés manquants
- Configuration manquante (Redis, S3 en prod)
- Docker/Kubernetes non préparé pour la prod
- Backup/restore DB

#### Catégorie 4 — Documentation

- Docs obsolètes ou incohérentes avec le code
- Guides manquants (API, contribution, deployment)
- Changelog absent
- README principal à mettre à jour

#### Catégorie 5 — UX et produit

- Décisions ouvertes non résolues (`docs/ux/open-decisions.md`)
- Parcours non testés en conditions réelles
- Design system non appliqué au code mobile
- Accessibilité non vérifiée

---

### PHASE 3 — Prioriser et structurer

Pour chaque lacune identifiée, évalue :

| Critère | Poids |
|---------|-------|
| Impact utilisateur | 30% |
| Risque technique | 25% |
| Effort | 20% |
| Dépendances | 15% |
| Alignement vision | 10% |

Classe en 4 tiers :

| Tier | Description | Horizon |
|------|-------------|---------|
| **T1 — Critique** | Bloque le Lot 0 ou crée un risque majeur | Cette semaine |
| **T2 — Important** | Améliore significativement la qualité ou l'UX | Ce mois |
| **T3 — Souhaitable** | Dette technique, amélioration continue | Ce trimestre |
| **T4 — Backlog** | Nice to have, exploration, futur | Plus tard |

---

### PHASE 4 — Proposer les issues

Pour chaque lacune, rédige une issue GitHub structurée :

```markdown
**Titre** : [court, actionnable]
**Labels** : [enhancement|testing|ci|documentation|infra|design|bug]
**Tier** : [T1|T2|T3|T4]
**Effort** : [S|M|L|XL]
**Dépendances** : [#N si applicable]
**Description** : [2-3 phrases]
```

Regroupe les issues par catégorie. Évite les doublons avec les issues existantes (ouvertes ET fermées).

---

### PHASE 5 — Demander l'arbitrage et créer

Présente le backlog proposé au PO sous cette forme :

```
## Backlog proposé — [N] issues

### T1 — Critique ([N] issues)
| # | Titre | Labels | Effort | Description courte |

### T2 — Important ([N] issues)
| # | Titre | Labels | Effort | Description courte |

### T3 — Souhaitable ([N] issues)
| # | Titre | Labels | Effort | Description courte |

### T4 — Backlog ([N] issues)
| # | Titre | Labels | Effort | Description courte |

### Questions pour le PO
[Questions nécessitant un arbitrage — décisions produit, priorisation, scope]

Valide le backlog et je crée les issues GitHub.
Tu peux supprimer, modifier ou réordonner avant que je crée.
```

Attends la validation du PO avant de créer les issues avec `gh issue create`.

---

## Règles

1. **Pas de doublons** — vérifie les issues ouvertes ET fermées avant de proposer
2. **Concret** — chaque issue doit être actionnable, pas vague
3. **Ancré dans le code** — cite les fichiers et lignes quand pertinent
4. **Réaliste** — les estimations d'effort doivent être honnêtes
5. **Demander** — si une décision produit est nécessaire, pose la question au PO
6. **Max 25 issues** — mieux vaut un backlog ciblé qu'une liste infinie

---

## 📝 Rapport d'exécution (auto-amélioration)

**Avant de terminer**, complète le dernier rapport créé par le hook dans
`.claude/reports/backlog/` (le fichier `.md` le plus récent avec `status: pending`).

Utilise le template `.claude/report-template.md` et remplis :

1. **Input** : résumé de la demande utilisateur (1-2 phrases)
2. **Output** : résumé de ce qui a été produit (1-2 phrases)
3. **Scores** (note chaque dimension de 1 à 5) :
   - **Complétude** : le skill a-t-il couvert tout ce qui était demandé ?
   - **Précision** : les outputs étaient-ils corrects et utiles ?
   - **Efficacité** : combien d'allers-retours avant un résultat satisfaisant ?
   - **Robustesse** : le skill a-t-il bien géré les cas limites rencontrés ?
4. **Observations** : points forts, points faibles, frictions utilisateur
5. **Met à jour le `status`** dans le frontmatter : `success`, `partial`, ou `error`

Si le fichier `.claude/reports/backlog/.review-ready` existe, signale à l'utilisateur :
> "Le skill `/backlog` a été exécuté 5+ fois depuis le dernier review. Lancer `/auto-review backlog` pour analyser et proposer des améliorations."
