# Suivi Lot 0 — Révise Mieux

| | |
|---|---|
| **Périmètre** | 53 ACs (33 P1 + 20 P2) · usage local père-fils · 4 chapitres pilotes |
| **Démarrage** | 2026-03-11 |

---

## Légende statut

| Icône | Statut |
|---|---|
| `[ ]` | À faire |
| `[~]` | En cours |
| `[x]` | Terminé |
| `[—]` | Bloqué / en attente |

---

## Phase 0 — Spécifications & Schéma

| # | Tâche | Statut | ACs couvertes | Livrables |
|---|---|---|---|---|
| 0.1 | PRD, ACs, MVP-scope | `[x]` | — | `docs/PRD.md`, `docs/ac/Z*.md`, `docs/MVP-scope.md` |
| 0.2 | Schéma SQL initial (migration 001) | `[x]` | — | `backend/migrations/001_initial_schema.sql` |
| 0.3 | Ajout `visual_blocks` et refs visuelles | `[x]` | Z2-AC15/16/17/18, Z4-AC17 (schéma prêt, implémentation MVP Hardening) | migration 001 mise à jour |

---

## Phase 1 — Fondations backend Go

Prérequis de toutes les phases suivantes.

| # | Tâche | Statut | Livrables |
|---|---|---|---|
| 1.1 | Structure projet Go (DDD : `domain/`, `app/`, `infra/`, `http/`, `config/`, `db/`) | `[x]` | arborescence `backend/internal/` |
| 1.2 | Connexion PostgreSQL (pgx pool) + config env | `[x]` | `internal/db/pool.go`, `internal/config/config.go` |
| 1.3 | Runner de migrations (Go, transactionnel, table schema_migrations) | `[x]` | `internal/db/migrate.go` |
| 1.4 | Entités domaine Go (4 bounded contexts + events + value objects) | `[x]` | `internal/domain/{mastery,chapter,session,validation,event}/` |
| 1.5 | Middleware auth JWT + token generation (usage local) | `[x]` | `internal/http/middleware/auth.go`, `token.go` |
| 1.6 | Docker Compose (Postgres + Redis + MinIO) | `[x]` | `docker-compose.yml` |
| 1.7 | Makefile hiérarchique (racine + backend + mobile) | `[x]` | `Makefile`, `backend/Makefile`, `mobile/Makefile` |
| 1.8 | Swagger auto-généré (swaggo/swag + gin-swagger) | `[x]` | `internal/docs/`, `make swagger` |
| 1.9 | Router Gin + health + wiring main.go | `[x]` | `internal/http/router.go`, `cmd/server/main.go` |

---

## Phase 2 — Moteur Mastery (Z1)

Coeur du produit. Machine à états + espacement.

| # | Tâche | Statut | ACs couvertes | Priorité |
|---|---|---|---|---|
| 2.1 | `MasteryService` : transitions UNKNOWN→FRAGILE→OK→SOLID | `[x]` | Z1-AC01, Z1-AC02, Z1-AC03 | P1 |
| 2.2 | Blocage OK→SOLID sans espacement 24h | `[x]` | Z1-AC04 | P1 |
| 2.3 | Régressions SOLID→OK, OK→FRAGILE, FRAGILE reste FRAGILE | `[x]` | Z1-AC05, Z1-AC06, Z1-AC07 | P1 |
| 2.4 | Récupération après régression (cs=0 et cs<2) | `[x]` | Z1-AC07b, Z1-AC07c | P1 |
| 2.5 | Indépendance mastery entre items | `[x]` | Z1-AC09 | P1 |
| 2.6 | Échec sur item UNKNOWN | `[x]` | Z1-AC11 | P1 |
| 2.7 | Maintien SOLID sur succès continu | `[x]` | Z1-AC12 | P1 |
| 2.8 | Resserrement proportionnel si exam posé | `[—]` | Z1-AC08 | P2 |
| 2.9 | Scoring par type de question (RUBRIC, NUMERIC, KEYWORDS) | `[—]` | Z1-AC10 | P2 |
| 2.10 | Plafond maîtrise OK pour items restreints | `[x]` | Z1-AC13 | P2 |
| 2.11 | Tests unitaires exhaustifs (≥ 12 cas de transition) | `[x]` | Z1-AC01→AC13 | P1 |
| 2.12 | Repository `masteries` (CRUD + requêtes due/state) | `[x]` | — | P1 |

---

## Phase 3 — Pipeline J0 minimal (Z2 + Z7-AC15)

Upload → OCR → Structuration → Items.

| # | Tâche | Statut | ACs couvertes | Priorité |
|---|---|---|---|---|
| 3.1 | Endpoint upload photo → Page + Block | `[x]` | — | P1 |
| 3.2 | Intégration OCR (service externe, stub pour dev local) | `[x]` | — | P1 |
| 3.3 | Structuration LLM → Items + Notions | `[x]` | Z7-AC15 | P1 |
| 3.4 | Carte leçon partielle si OCR en cours | `[x]` | Z2-AC01 | P1 |
| 3.5 | Aucun item généré sur une page (gestion gracieuse) | `[x]` | Z2-AC04 | P1 |
| 3.6 | Échec génération items (erreur LLM, retry + fallback) | `[x]` | Z2-AC05 | P1 |
| 3.7 | Diagnostic impossible si 0 items valides | `[x]` | Z2-AC07 | P1 |
| 3.8 | Streaming carte leçon (première page dispo) | `[x]` | Z2-AC10 | P1 |
| 3.9 | Timeout OCR sur page intermédiaire | `[ ]` | Z2-AC02 | P2 |
| 3.10 | Photo floue détectée (confidence < 0.3) | `[ ]` | Z2-AC03 | P2 |
| 3.11 | Idempotence pipeline au restart | `[ ]` | Z2-AC06 | P2 |
| 3.12 | Bloc SCHEMA/MAP conservé comme Document image | `[ ]` | Z2-AC08 | P2 |
| 3.13 | File validation plafonnée à 8 | `[ ]` | Z2-AC09 | P2 |
| 3.14 | Clé d'identité canonique de l'item | `[x]` | Z5-AC01 | P1 |
| 3.15 | Révision courante unique par chapitre | `[x]` | Z5-AC04 | P1 |
| 3.16 | Items archivés non proposés en session | `[x]` | Z5-AC06 | P1 |

---

## Phase 4 — Sessions & Questions (Z4 + Z6)

Composition, lazy generation, feedback.

| # | Tâche | Statut | ACs couvertes | Priorité |
|---|---|---|---|---|
| 4.1 | Session evening_first déclenchée après upload | `[x]` | Z6-AC04 | P1 |
| 4.2 | evening_first : 100% UNKNOWN, difficulté 1 | `[x]` | Z6-AC05 | P1 |
| 4.3 | Composition session avec contraintes pack | `[—]` | Z4-AC05 | P1 |
| 4.4 | Pool vide : dégradation gracieuse | `[x]` | Z4-AC06 | P1 |
| 4.5 | Reprise session interrompue | `[x]` | Z4-AC04 | P1 |
| 4.6 | Feedback enrichi après réponse incorrecte | `[x]` | Z4-AC09 | P1 |
| 4.7 | Lazy generation de questions (templates → questions) | `[x]` | — | P1 |
| 4.8 | Mode dégradé sans emploi du temps | `[x]` | Z6-AC11 | P1 |
| 4.9 | Pas de pénalité maîtrise pour items en retard | `[x]` | Z6-AC13 | P1 |
| 4.10 | Mock exam non bloqué par session daily active | `[ ]` | Z4-AC07 | P2 |
| 4.11 | Exam multi-chapitre : création et liaison | `[ ]` | Z6-AC10 | P2 |
| 4.12 | Vue chapitre par notion (accordéon + maîtrise) | `[ ]` | Z7-AC16 | P2 |

---

## Phase 5 — Validation HITL (Z3)

Boucle de qualité parent/admin.

| # | Tâche | Statut | ACs couvertes | Priorité |
|---|---|---|---|---|
| 5.1 | Chapitre utilisable avec 0 validations faites | `[ ]` | Z3-AC06 | P1 |
| 5.2 | Gabarits bloqués sur item validation_required non résolu | `[ ]` | Z3-AC01 | P2 |
| 5.3 | Action Confirmer sur ValidationTask | `[ ]` | Z3-AC02 | P2 |
| 5.4 | Action Corriger sur ValidationTask | `[ ]` | Z3-AC03 | P2 |
| 5.5 | Action « Je ne sais pas » sur ValidationTask | `[ ]` | Z3-AC04 | P2 |
| 5.6 | Action Ignorer sur ValidationTask | `[ ]` | Z3-AC05 | P2 |
| 5.7 | Pas de ValidationTask si confidence > 0.85 | `[ ]` | Z3-AC09 | P2 |
| 5.8 | Vérification croisée LLM (fidelity check) | `[ ]` | Z3-AC10 | P2 |

---

## Phase 6 — Onboarding (Z8)

Première expérience utilisateur.

| # | Tâche | Statut | ACs couvertes | Priorité |
|---|---|---|---|---|
| 6.1 | Chapitre démo pré-chargé (cold start) | `[ ]` | Z8-AC01 | P1 |
| 6.2 | Empty state guidé avant premier upload | `[ ]` | Z8-AC02 | P1 |
| 6.3 | Écran de progression pendant le traitement J0 | `[ ]` | Z8-AC04 | P1 |
| 6.4 | UX de recovery si le premier OCR échoue | `[ ]` | Z8-AC03 | P2 |
| 6.5 | Séquence d'onboarding déterministe (Day 0) | `[ ]` | Z8-AC08 | P2 |

---

## Phase 7 — Mobile (React Native + Expo)

En parallèle dès Phase 3. Expo Router, composants métier.

| # | Tâche | Statut | Écrans / Composants |
|---|---|---|---|
| 7.1 | Navigation Expo Router (tabs, stack, modal) | `[ ]` | `app/_layout.tsx`, tabs |
| 7.2 | Client API (fetch, auth, error handling) | `[ ]` | `services/api.ts` |
| 7.3 | Écran upload photo (caméra + galerie) | `[ ]` | `app/upload.tsx` |
| 7.4 | Écran progression pipeline J0 | `[ ]` | `app/processing.tsx` |
| 7.5 | Carte leçon (notions, items, visuels) | `[ ]` | `app/chapter/[id].tsx` |
| 7.6 | Écran session de révision (question → réponse → feedback) | `[ ]` | `app/session/[id].tsx` |
| 7.7 | Dashboard élève (chapitres, mastery, exams) | `[ ]` | `app/(tabs)/index.tsx` |
| 7.8 | Interface validation HITL (parent) | `[ ]` | `app/validation.tsx` |
| 7.9 | Onboarding + chapitre démo | `[ ]` | `app/onboarding/` |

---

## Matrice de couverture ACs Lot 0

Vérification exhaustive : chaque AC Lot 0 est rattachée à une tâche.

### P1 (33 ACs)

| AC | Phase.Tâche | Statut |
|---|---|---|
| Z1-AC01 | 2.1 | `[x]` |
| Z1-AC02 | 2.1 | `[x]` |
| Z1-AC03 | 2.1 | `[x]` |
| Z1-AC04 | 2.2 | `[x]` |
| Z1-AC05 | 2.3 | `[x]` |
| Z1-AC06 | 2.3 | `[x]` |
| Z1-AC07 | 2.3 | `[x]` |
| Z1-AC07b | 2.4 | `[x]` |
| Z1-AC07c | 2.4 | `[x]` |
| Z1-AC09 | 2.5 | `[x]` |
| Z1-AC11 | 2.6 | `[x]` |
| Z1-AC12 | 2.7 | `[x]` |
| Z2-AC01 | 3.4 | `[x]` |
| Z2-AC04 | 3.5 | `[x]` |
| Z2-AC05 | 3.6 | `[x]` |
| Z2-AC07 | 3.7 | `[x]` |
| Z2-AC10 | 3.8 | `[x]` |
| Z3-AC06 | 5.1 | `[ ]` |
| Z4-AC04 | 4.5 | `[x]` |
| Z4-AC05 | 4.3 | `[—]` |
| Z4-AC06 | 4.4 | `[x]` |
| Z4-AC09 | 4.6 | `[x]` |
| Z5-AC01 | 3.14 | `[x]` |
| Z5-AC04 | 3.15 | `[x]` |
| Z5-AC06 | 3.16 | `[x]` |
| Z6-AC04 | 4.1 | `[x]` |
| Z6-AC05 | 4.2 | `[x]` |
| Z6-AC11 | 4.8 | `[x]` |
| Z6-AC13 | 4.9 | `[x]` |
| Z7-AC15 | 3.3 | `[ ]` |
| Z8-AC01 | 6.1 | `[ ]` |
| Z8-AC02 | 6.2 | `[ ]` |
| Z8-AC04 | 6.3 | `[ ]` |

### P2 (20 ACs)

| AC | Phase.Tâche | Statut |
|---|---|---|
| Z1-AC08 | 2.8 | `[—]` |
| Z1-AC10 | 2.9 | `[—]` |
| Z1-AC13 | 2.10 | `[x]` |
| Z2-AC02 | 3.9 | `[ ]` |
| Z2-AC03 | 3.10 | `[ ]` |
| Z2-AC06 | 3.11 | `[ ]` |
| Z2-AC08 | 3.12 | `[ ]` |
| Z2-AC09 | 3.13 | `[ ]` |
| Z3-AC01 | 5.2 | `[ ]` |
| Z3-AC02 | 5.3 | `[ ]` |
| Z3-AC03 | 5.4 | `[ ]` |
| Z3-AC04 | 5.5 | `[ ]` |
| Z3-AC05 | 5.6 | `[ ]` |
| Z3-AC09 | 5.7 | `[ ]` |
| Z3-AC10 | 5.8 | `[ ]` |
| Z4-AC07 | 4.10 | `[ ]` |
| Z6-AC10 | 4.11 | `[ ]` |
| Z7-AC16 | 4.12 | `[ ]` |
| Z8-AC03 | 6.4 | `[ ]` |
| Z8-AC08 | 6.5 | `[ ]` |

---

## Dépendances entre phases

```
Phase 0 (specs + schéma) ──[terminée]
    │
    ▼
Phase 1 (fondations Go)
    │
    ├──────────────────┐
    ▼                  ▼
Phase 2 (mastery)   Phase 3 (pipeline J0)
    │                  │
    └────────┬─────────┘
             ▼
         Phase 4 (sessions)
             │
             ├──────────────────┐
             ▼                  ▼
         Phase 5 (HITL)     Phase 6 (onboarding)

Phase 7 (mobile) ── démarrage parallèle dès Phase 3
```
