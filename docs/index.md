# Revise Mieux -- Documentation

Revise Mieux est un SaaS educatif qui transforme des photos de cahier en assistant de revision personnalise pour collegiens (11-15 ans). A partir d'un upload de cours manuscrit, le produit genere une carte de lecon structuree, des entrainements adaptatifs bases sur la repetition espacee, et des controles blancs. Les parents sont integres via un reporting base sur des preuves de maitrise reelles (boucle HITL).

---

## Stack technique

| Composant | Technologie |
|-----------|-------------|
| Backend API | Go 1.23 + Gin |
| Base de donnees | PostgreSQL 16 (pgx/v5, SQL brut) |
| Mobile | React Native + Expo 55 + Expo Router |
| LLM | Gemini Flash (defaut), Anthropic Claude Sonnet, Mistral |
| OCR | Gemini Flash VLM |
| Cache | Redis 7 (go-redis/v9) |
| Storage | S3 / MinIO (dev) |
| IDs | UUIDv7 (google/uuid) |

---

## Etat du Lot 0

Le Lot 0 couvre 53 criteres d'acceptation (33 P1 + 20 P2) pour un usage local pere-fils sur 4 chapitres pilotes.

| Phase | Description | Statut |
|-------|-------------|--------|
| Phase 0 | Specifications et schema SQL | Terminee |
| Phase 1 | Fondations backend Go (DDD, auth, infra) | Terminee |
| Phase 2 | Moteur Mastery (Z1) -- machine a etats | Terminee |
| Phase 3 | Pipeline J0 (upload, OCR, structuration LLM) | Terminee |
| Phase 4 | Sessions et questions (composition, lazy gen) | Terminee |
| Phase 5 | Validation HITL (boucle qualite parent) | Terminee |
| Phase 6 | Onboarding (cold start, chapitre demo) | Terminee |
| Phase 7 | Mobile React Native + Expo | En cours |

**Progression** : 52/53 ACs implementees (backend complet, mobile en finalisation).

---

## Sections de la documentation

### Demarrage et developpement

- [Guide de demarrage rapide](getting-started.md) -- prerequis, installation, configuration, lancement
- [Guide de contribution](contributing.md) -- workflow TDD, conventions Go, checklist PR

### Architecture

- [Vue d'ensemble architecture](architecture/index.md) -- bounded contexts, hexagonal, packages Go
- [Decision table merge OCR/IDP](architecture/merge-decision-table.md)
- [Pipeline OCR/IDP incremental](architecture/ocr-idp-incremental.md)

### Specifications produit

- [PRD complet](PRD.md) -- personas, pipeline, architecture, modele de donnees, algorithmes
- [Perimetre MVP (Lot 0)](MVP-scope.md) -- classification des 171 ACs
- [Suivi Lot 0](lot0-tracker.md) -- progression detaillee par phase et matrice AC

### Criteres d'acceptation

- [Z1 -- Mastery](ac/Z1.md) -- machine a etats, transitions, espacement
- [Z2 -- Pipeline J0](ac/Z2.md) -- upload, OCR, structuration
- [Z3 -- Validation HITL](ac/Z3.md) -- boucle qualite parent/admin
- [Z4 -- Sessions](ac/Z4.md) -- composition, questions, feedback
- [Z5 -- Gestion items](ac/Z5.md) -- identite, archivage, CRUD
- [Z6 -- Planification](ac/Z6.md) -- calendrier, exams, sessions
- [Z7 -- Carte lecon](ac/Z7.md) -- affichage, notions, visuels
- [Z8 -- Onboarding](ac/Z8.md) -- cold start, chapitre demo

### LLM et benchmarks

- [Strategie LLM](llm-strategy.md) -- choix de modeles, providers
- [Benchmark LLM](llm-benchmark.md) -- resultats comparatifs IDP/OCR

### Design

- [Design system](design/index.md) -- principes, couleurs, typographie, composants

### UX

- [Parcours utilisateurs](user-journeys/first-connection.md)
- [Personas](ux/personas.md)
- [Inventaire ecrans](ux/screen-inventory.md)
- [Modele d'etats](ux/state-model.md)
- [Flux utilisateur complet](ux/user-flow-complete.md)
- [Cas limites](ux/edge-cases.md)
- [Decisions ouvertes](ux/open-decisions.md)

### Deploiement

- [Guide de deploiement](deployment.md) -- Docker Compose, variables d'environnement, health check

### Processus

- [Boucle de feedback](process/feedback-loop.md)
