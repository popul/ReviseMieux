# Registre des issues — Audit 2026-04-03

## Résumé

| Sévérité | Nombre |
|----------|--------|
| Majeur | 5 |
| Mineur | 4 |
| **Total** | **9** |

---

## Issues majeures

### ISS-001 — Prompts non externalisés

| Champ | Valeur |
|-------|--------|
| Sévérité | MAJEUR |
| Fichiers | `backend/internal/infra/anthropic/prompts.go` |
| Spec | `docs/llm-strategy.md` §9 : prompts en fichiers `.txt` + `//go:embed` |
| Réalité | Prompts en constantes Go dans `prompts.go` |
| Recommandation | Externaliser dans `backend/internal/infra/llm/prompts/*.txt` avec `//go:embed` |

### ISS-002 — Pas de provider mistral-small pour l'IDP

| Champ | Valeur |
|-------|--------|
| Sévérité | MAJEUR |
| Fichiers | `backend/internal/config/config.go`, `backend/cmd/server/main.go` |
| Spec | `backend/testdata/benchmark/README.md` : mistral-small = best value IDP (quality 0.87, $0.00006/item) |
| Réalité | Seuls Gemini et Anthropic sont configurés. Pas de client Mistral direct. |
| Recommandation | Ajouter un provider mistral-small via `openaicompat` (API compatible OpenAI) |

### ISS-003 — PipelineService non câblé dans main.go

| Champ | Valeur |
|-------|--------|
| Sévérité | MAJEUR |
| Fichiers | `backend/cmd/server/main.go:100-102` |
| Réalité | Commenté : "requires Storage + OCR adapters (not yet implemented)" |
| Impact | Les endpoints `/chapters/:id/upload` et `/revisions/:id/progress` retournent nil |
| Recommandation | Implémenter les adaptateurs S3 et OCR puis câbler le PipelineService |

### ISS-004 — Tracker mobile sous-estimé (Phase 7)

| Champ | Valeur |
|-------|--------|
| Sévérité | MAJEUR |
| Fichiers | `docs/lot0-tracker.md` Phase 7 |
| Réalité | Phase 7 marquée `[ ]` mais 7 écrans Expo Router + 122 tests mobiles existent |
| Recommandation | Mettre à jour le tracker : marquer les tâches 7.1, 7.3-7.7 comme `[x]` ou `[~]` |

### ISS-005 — Frontend web non documenté

| Champ | Valeur |
|-------|--------|
| Sévérité | MAJEUR |
| Fichiers | `frontend/` |
| Réalité | Un frontend TypeScript/Vite existe mais n'est mentionné ni dans CLAUDE.md ni dans le tracker |
| Recommandation | Arbitrage : documenter comme backoffice admin, ou archiver si obsolète |

---

## Issues mineures

### ISS-006 — Z8-AC03 sans test backend dédié

| Champ | Valeur |
|-------|--------|
| Sévérité | MINEUR |
| AC | Z8-AC03 (Recovery si premier OCR échoue) |
| Tracker | `[x]` |
| Recommandation | Ajouter `TestZ8AC03_RecoveryOnFirstOCRFailure` |

### ISS-007 — Z7-AC15 sans test explicite

| Champ | Valeur |
|-------|--------|
| Sévérité | MINEUR |
| AC | Z7-AC15 (Structuration LLM → Items + Notions) |
| Tracker | `[x]` |
| Réalité | Testé indirectement via `TestPipelineService_*` |
| Recommandation | Ajouter `TestZ7AC15_StructurationLLM` explicite |

### ISS-008 — 2 ACs MVP Hardening marquées [x]

| Champ | Valeur |
|-------|--------|
| Sévérité | MINEUR |
| ACs | Z2-AC15 (visual_blocks), Z4-AC17 (multimodal visuel) |
| Tracker | `[x]` mais classées MVP Hardening/Post-MVP dans MVP-scope.md |
| Réalité | Schéma SQL prêt mais implémentation différée |
| Recommandation | Changer en `[~]` avec note "schéma prêt, implémentation différée" |

### ISS-009 — Z8-AC04 sans test backend

| Champ | Valeur |
|-------|--------|
| Sévérité | MINEUR |
| AC | Z8-AC04 (Écran progression pipeline J0) |
| Tracker | `[x]` |
| Réalité | Couvert côté mobile (`app/processing.tsx`) mais pas de test backend |
| Recommandation | Acceptable — l'AC est UI-oriented |
