# Prompt 0 — Plan d'exécution : orchestration des 5 chantiers en sub-agents + PRs

> **Objectif** : Lancer les 5 prompts de conception via des sub-agents Claude Code, chacun travaillant dans un worktree isolé et produisant une PR dédiée. Ce prompt EST le chef d'orchestre.

---

You are the orchestrator agent for the "Révise Mieux" project. Your job is to launch, coordinate, and supervise 5 specialized sub-agents, each working on a distinct design/architecture chantier. Each sub-agent produces des livrables concrets dans une branche dédiée, et crée une PR à la fin.

## Vue d'ensemble

```
                        ┌─────────────────────┐
                        │   ORCHESTRATEUR      │
                        │   (ce prompt)        │
                        └──────┬──────────────┘
                               │
              ┌────────────────┼────────────────┐
              │                │                │
    ┌─────────▼──────┐ ┌──────▼───────┐ ┌──────▼───────┐
    │ VAGUE 1        │ │ VAGUE 1      │ │ VAGUE 1      │
    │ (parallèle)    │ │ (parallèle)  │ │ (parallèle)  │
    │                │ │              │ │              │
    │ Agent 1: OCR   │ │ Agent 2: UX  │ │ Agent 3: DS  │
    │ PR #1          │ │ PR #2        │ │ PR #3        │
    └────────────────┘ └──────────────┘ └──────────────┘
                               │
                        (attendre vague 1)
                               │
              ┌────────────────┼────────────────┐
              │                                 │
    ┌─────────▼──────┐               ┌──────────▼─────┐
    │ VAGUE 2        │               │ VAGUE 2        │
    │ (parallèle)    │               │ (parallèle)    │
    │                │               │                │
    │ Agent 4: Audit │               │ Agent 5: Loop  │
    │ PR #4          │               │ PR #5          │
    └────────────────┘               └────────────────┘
```

## Pourquoi 2 vagues

- **Vague 1** (agents 1, 2, 3) : chantiers de conception indépendants. Aucune dépendance entre eux. Se lancent en parallèle.
- **Vague 2** (agents 4, 5) : chantiers d'audit et de process. L'agent 4 (audit) bénéficie des livrables de la vague 1 pour vérifier la cohérence. L'agent 5 (feedback loop) s'appuie sur le format du rapport de l'agent 4 et sur les livrables des agents 1-3 pour définir ce que la boucle doit surveiller.

---

## Instructions d'exécution

### Pré-requis

Avant de lancer les agents :

1. **Vérifier l'état du repo** :
   ```bash
   git status          # doit être clean
   git branch          # noter la branche courante
   ```

2. **S'assurer que la branche de base est à jour** :
   ```bash
   git pull origin reboot
   ```

3. **Créer le dossier reports s'il n'existe pas** :
   ```bash
   mkdir -p reports
   ```

---

### VAGUE 1 — Lancement en parallèle (3 agents)

Lancer ces 3 agents **simultanément** avec `isolation: "worktree"` pour que chacun travaille sur sa propre copie du repo.

---

#### Agent 1 : Architecture OCR/IDP incrémental

**Branche** : `design/ocr-idp-incremental`

**Prompt à donner à l'agent** :

```
Tu es un architecte IA senior spécialisé en OCR et IDP éducatif.

## Ta mission
Exécuter le prompt `prompts/01-ocr-idp-incremental.md` et produire des livrables concrets.

## Étapes

1. Lis intégralement `prompts/01-ocr-idp-incremental.md` — c'est ton brief détaillé.
2. Lis les documents de référence mentionnés dans le prompt :
   - `docs/PRD.md` (sections §10 pipeline, §16 modèle de données, §13 retry OCR)
   - `docs/ac/Z2.md`, `docs/ac/Z3.md`, `docs/ac/Z7.md`
   - `docs/llm-strategy.md` (sections §8-9 : pipeline 2 étages OCR/IDP, combos recommandés)
   - `backend/testdata/benchmark/README.md` (résultats 21 modèles OCR + IDP)
   - `backend/migrations/001_initial_schema.sql`
   - `backend/internal/domain/chapter/` (entités existantes)
   - `backend/internal/domain/event/` (events existants)
3. Produis les livrables suivants dans le repo :

### Livrables à écrire

| Fichier | Contenu |
|---------|---------|
| `docs/architecture/ocr-idp-incremental.md` | Document d'architecture complet (sections 1-15 du prompt) |
| `docs/architecture/pivot-format-schema.json` | JSON Schema du format pivot |
| `docs/architecture/pivot-format-example-photosynthese.json` | Exemple concret sur "La photosynthèse" |
| `docs/architecture/merge-decision-table.md` | Tableau de décision merge avec seuils |
| `docs/architecture/incremental-sequence.mermaid` | Diagramme de séquence du workflow |

4. Crée une PR avec :
   - Titre : "docs: architecture OCR/IDP incrémental"
   - Body structuré (summary + livrables + questions ouvertes)
```

---

#### Agent 2 : User Flow détaillé

**Branche** : `design/ux-user-flow`

**Prompt à donner à l'agent** :

```
Tu es un principal product designer spécialisé en apps éducatives mobile.

## Ta mission
Exécuter le prompt `prompts/02-ux-user-flow-detail.md` et produire des livrables concrets.

## Étapes

1. Lis intégralement `prompts/02-ux-user-flow-detail.md` — c'est ton brief détaillé.
2. Lis les documents de référence :
   - `docs/PRD.md` (§4 personas, §5 parcours utilisateurs, §15 fonctionnalités)
   - `docs/ac/Z1.md` à `docs/ac/Z8.md` (tous les ACs)
   - `docs/MVP-scope.md` (Lot 0, 53 ACs)
   - `docs/user-journeys/first-connection.md`
   - `docs/openapi.yaml`
   - `mobile/` (explorer les écrans existants)
3. Produis les livrables suivants :

### Livrables à écrire

| Fichier | Contenu |
|---------|---------|
| `docs/ux/user-flow-complete.md` | Document UX complet (sections 1-12 du prompt) |
| `docs/ux/personas.md` | 4 personas détaillés avec états émotionnels |
| `docs/ux/screen-inventory.md` | Inventaire de tous les écrans avec route, intention, CTA, mapping ACs |
| `docs/ux/state-model.md` | Modèle des 11 états par écran critique |
| `docs/ux/edge-cases.md` | Catalogue edge cases avec traitement et recovery |
| `docs/ux/open-decisions.md` | Décisions produit à prendre avant implémentation |

4. Crée une PR avec :
   - Titre : "docs: user flow complet élève + parent"
   - Body structuré
```

---

#### Agent 3 : Design System

**Branche** : `design/design-system`

**Prompt à donner à l'agent** :

```
Tu es un architecte design system senior spécialisé en produits éducatifs pour adolescents.

## Ta mission
Exécuter le prompt `prompts/03-design-system-intergenerational.md` et produire des livrables concrets.

## Étapes

1. Lis intégralement `prompts/03-design-system-intergenerational.md` — c'est ton brief détaillé.
2. Lis les documents de référence :
   - `docs/PRD.md` (§4 personas, §2.4 principes pédagogiques, §19 KPIs)
   - `docs/ac/Z5.md` (dashboard élève)
   - `docs/ac/Z6.md` (dashboard parent)
   - `mobile/` (composants existants)
3. Produis les livrables suivants :

### Livrables à écrire

| Fichier | Contenu |
|---------|---------|
| `docs/design/design-system.md` | Document complet (sections 1-15 du prompt) |
| `docs/design/design-principles.md` | 5-7 principes avec exemples/contre-exemples |
| `docs/design/color-system.md` | Palette complète avec hex, usages, couleurs Mastery |
| `docs/design/typography.md` | Échelle typo, fonts recommandées, traitement contenu pédagogique |
| `docs/design/component-philosophy.md` | Philosophie de chaque famille de composants |
| `docs/design/do-dont.md` | Exemples concrets trop enfantin / trop froid / bien équilibré |
| `docs/design/moodboard.md` | 5 apps de référence avec ce qu'on emprunte et évite |
| `docs/design/scalability.md` | Stratégie d'évolution collège → lycée → post-bac |

4. Crée une PR avec :
   - Titre : "docs: design system intergénérationnel"
   - Body structuré
```

---

### POINT DE SYNCHRONISATION — Attendre la vague 1

Avant de lancer la vague 2, vérifier que les 3 PRs de la vague 1 sont créées. Résumer au product owner :

```markdown
## Vague 1 terminée

| Agent | Branche | PR | Statut |
|-------|---------|-----|--------|
| 1 - OCR/IDP | `design/ocr-idp-incremental` | #XX | Créée |
| 2 - UX Flow | `design/ux-user-flow` | #XX | Créée |
| 3 - Design System | `design/design-system` | #XX | Créée |

### Prêt pour la vague 2 ?
Les agents 4 et 5 vont lire les livrables de la vague 1 pour :
- Agent 4 : vérifier la cohérence entre les nouvelles specs et le code existant
- Agent 5 : intégrer les nouveaux documents dans la boucle de monitoring

Lancer la vague 2 ? [oui/non]
```

---

### VAGUE 2 — Lancement en parallèle (2 agents)

Lancer ces 2 agents **simultanément** avec `isolation: "worktree"`.

---

#### Agent 4 : Audit de cohérence

**Branche** : `audit/specs-vs-code`

**Prompt à donner à l'agent** :

```
Tu es un auditeur technique senior spécialisé en cohérence specs/implémentation.

## Ta mission
Exécuter le prompt `prompts/04-coherence-audit-specs-vs-code.md` et produire un rapport d'audit actionnable.

## Étapes

1. Lis intégralement `prompts/04-coherence-audit-specs-vs-code.md` — c'est ta procédure.
2. Exécute la procédure d'audit complète :
   - Phase A : ACs vs code (lire docs/ac/Z*.md, vérifier tests et implémentation)
   - Phase B : Domaine (machine à états, modèle de données, langage ubiquitaire)
   - Phase C : Architecture (hexagonale, conventions Go)
   - Phase D : API (OpenAPI vs router)
   - Phase E : Tracker sync
3. Lance les tests pour collecter l'evidence :
   ```bash
   cd backend && go test ./... -v -count=1 2>&1
   cd backend && go build ./... 2>&1
   ```
4. Lis aussi les nouveaux documents de la vague 1 (s'ils existent dans le worktree) :
   - `docs/architecture/ocr-idp-incremental.md`
   - `docs/ux/screen-inventory.md`
   - `docs/design/design-system.md`
   Vérifie leur cohérence avec le PRD et les ACs existants.

### Livrables à écrire

| Fichier | Contenu |
|---------|---------|
| `reports/audit-[DATE].md` | Rapport d'audit complet (format du prompt) |
| `reports/issue-register.md` | Registre initial des issues trouvées |

5. Crée une PR avec :
   - Titre : "audit: cohérence specs vs code — [DATE]"
   - Body = executive summary du rapport + nombre de findings par sévérité
```

---

#### Agent 5 : Feedback Loop setup

**Branche** : `tooling/feedback-loop`

**Prompt à donner à l'agent** :

```
Tu es un architecte de processus d'amélioration continue spécialisé en boucles produit/engineering.

## Ta mission
Exécuter le prompt `prompts/05-feedback-loop-continuous.md` et produire les artifacts opérationnels de la boucle.

## Étapes

1. Lis intégralement `prompts/05-feedback-loop-continuous.md` — c'est ton brief.
2. Lis les documents de référence pour comprendre ce que la boucle doit surveiller :
   - `CLAUDE.md` (conventions, architecture)
   - `docs/PRD.md` (scope produit)
   - `docs/lot0-tracker.md` (état d'avancement)
   - `docs/ac/Z1.md` à `Z8.md` (ACs à monitorer)
3. Si le rapport d'audit (agent 4) existe, lis-le pour calibrer la boucle sur les vrais problèmes trouvés.

### Livrables à écrire

| Fichier | Contenu |
|---------|---------|
| `docs/process/feedback-loop.md` | Document de conception de la boucle (sections 1-14 du prompt) |
| `.claude/commands/feedback-loop.md` | Slash command `/feedback-loop` prêt à l'emploi |
| `reports/templates/feedback-loop-template.md` | Template du rapport de boucle |
| `reports/templates/issue-register-template.md` | Template du registre d'issues |
| `reports/templates/decision-log-template.md` | Template du log de décisions |

4. Teste que le slash command fonctionne :
   - Vérifie que le fichier `.claude/commands/feedback-loop.md` est bien formé
   - Vérifie que les chemins référencés existent

5. Crée une PR avec :
   - Titre : "tooling: setup boucle feedback-loop + slash command"
   - Body structuré (résumé + comment utiliser + cadence recommandée)
```

---

### SYNTHÈSE FINALE

Après les 2 vagues, présenter au product owner un résumé complet :

```markdown
## Exécution terminée — 5 PRs créées

| # | Chantier | Branche | PR | Livrables | Statut |
|---|----------|---------|-----|-----------|--------|
| 1 | OCR/IDP incrémental | `design/ocr-idp-incremental` | #XX | Architecture + JSON Schema + merge rules | Créée |
| 2 | User Flow | `design/ux-user-flow` | #XX | Flow complet + personas + edge cases | Créée |
| 3 | Design System | `design/design-system` | #XX | Principes + tokens + composants | Créée |
| 4 | Audit cohérence | `audit/specs-vs-code` | #XX | Rapport + issue register | Créée |
| 5 | Feedback Loop | `tooling/feedback-loop` | #XX | Process doc + slash command + templates | Créée |

### Ordre de review recommandé

1. **PR #4 (Audit)** en premier — donne une photo de l'état actuel
2. **PR #1, #2, #3** en parallèle — les 3 chantiers de conception
3. **PR #5 (Loop)** en dernier — met en place le monitoring continu

### Actions immédiates après review

- [ ] Review et merge des PRs
- [ ] Lancer `/feedback-loop` pour vérifier que la boucle fonctionne
- [ ] Planifier la cadence de la boucle (1-2x/jour recommandé)
- [ ] Prioriser les findings de l'audit pour le prochain sprint

### Questions ouvertes remontées par les agents
[Consolidation des questions de chaque agent nécessitant un arbitrage humain]
```

---

## Gestion des erreurs

| Problème | Stratégie |
|----------|-----------|
| Un agent échoue en cours de route | Les autres continuent. L'orchestrateur note l'échec et propose un retry. |
| Un agent ne peut pas créer de PR (permissions) | Écrire les livrables dans la branche, lister les commandes PR à exécuter manuellement. |
| Les tests backend échouent | L'agent 4 documente les échecs dans son rapport au lieu de bloquer. |
| Un document de référence est manquant | L'agent note le manque dans son rapport et travaille avec ce qui est disponible. |
| Le worktree a des conflits | L'agent rebase sur main et résout, ou signale le conflit. |

## Contraintes

- Chaque agent travaille dans un **worktree isolé** — pas d'interférence entre agents
- Chaque agent lit le prompt complet correspondant dans `prompts/` — c'est son brief
- Chaque agent **ne modifie PAS** les documents de spec existants (PRD, ACs, etc.) — il produit de NOUVEAUX documents
- L'agent 4 est le seul qui **exécute des commandes** (tests, build) — les autres sont documentation-only
- Toutes les PRs ciblent la branche `reboot`
