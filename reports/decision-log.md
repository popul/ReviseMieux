# Log de decisions — Feedback Loop

> Historique de toutes les decisions prises lors des arbitrages de la boucle d'amelioration continue.
> Chaque entree est immutable une fois ecrite. Pas de modification retroactive.

---

## Decisions

### DEC-2026-04-03-001 — Fermer issue #15 (prompts déjà externalisés)

| | |
|---|---|
| **Date** | 2026-04-03 |
| **Cycle** | #1 |
| **Finding** | FL-2026-04-03-001 |
| **Type** | OPPORTUNITE |
| **Decision** | Accepter |

**Contexte** : L'issue #15 demandait d'externaliser les prompts LLM en fichiers `.txt` + `//go:embed`. Or `backend/internal/infra/llm/prompts.go` utilise déjà `//go:embed` avec SHA-256, et les fichiers `.txt` existent dans `infra/llm/prompts/`.

**Proposition** : Fermer l'issue comme résolue.

**Decision PO** : Accepter.

**Action** : Issue #15 fermée via `gh issue close 15`.

---

### DEC-2026-04-03-002 — Mettre à jour tracker Phase 7 mobile

| | |
|---|---|
| **Date** | 2026-04-03 |
| **Cycle** | #1 |
| **Finding** | FL-2026-04-03-004 |
| **Type** | INCOHERENCE |
| **Decision** | Accepter |

**Contexte** : Le tracker marquait Phase 7 (Mobile) entièrement `[ ]` alors que 7 écrans Expo Router existent avec 122 tests.

**Proposition** : Passer 7.1, 7.3-7.7, 7.9 à `[x]`, 7.2 à `[~]`, 7.8 reste `[ ]`.

**Decision PO** : Accepter.

**Action** : `docs/lot0-tracker.md` Phase 7 mis à jour.

---

### DEC-2026-04-03-003 — Ajouter tests Z8-AC03 et Z7-AC15

| | |
|---|---|
| **Date** | 2026-04-03 |
| **Cycle** | #1 |
| **Finding** | FL-2026-04-03-002, FL-2026-04-03-003 |
| **Type** | COUVERTURE |
| **Decision** | Accepter |

**Contexte** : 2 ACs sans test dédié `TestZ*AC*` : Z8-AC03 (recovery OCR) et Z7-AC15 (structuration LLM).

**Proposition** : Ajouter les tests. Issue #20 existe.

**Decision PO** : Accepter.

**Action** : Issue #20 reste ouverte pour implémentation.

---

### DEC-2026-04-04-004 — Workflow CI intégration avec PostgreSQL

| | |
|---|---|
| **Date** | 2026-04-04 |
| **Cycle** | #4 |
| **Finding** | FL-2026-04-04-030 |
| **Type** | COUVERTURE |
| **Decision** | Accepter |

**Contexte** : 30+ tests tagués `//go:build integration` ne tournaient pas en CI (pas de PostgreSQL, pas de tag).

**Proposition** : Créer `.github/workflows/integration.yml` avec PostgreSQL 16 en service, sur PR + nightly.

**Decision PO** : Accepter.

**Action** : Workflow créé et pushé.

---

## Statistiques

| Métrique | Valeur |
|----------|--------|
| Total decisions | 4 |
| Acceptées | 4 |
| Modifiées | 0 |
| Rejetées | 0 |
| Reportées | 0 |
