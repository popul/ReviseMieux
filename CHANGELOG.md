# Changelog

Tous les changements notables du projet sont documentés ici.

Format : [Keep a Changelog](https://keepachangelog.com/fr/1.1.0/)

## [Unreleased]

_Rien pour l'instant._

## [0.1.0] — 2026-04-03

Lot 0 complet : 53 ACs implémentées (33 P1 + 20 P2), usage local père-fils sur 4 chapitres pilotes.

### Added

#### Infrastructure
- Setup DDD hexagonal (5 bounded contexts : mastery, chapter, session, validation, event)
- CI GitHub Actions : backend Go (build + tests + coverage) + mobile Jest
- CI integration tests avec PostgreSQL
- Site documentation MkDocs Material sur GitHub Pages
- Docker Compose (PostgreSQL, Redis, MinIO)
- Makefile hiérarchique (racine + backend + mobile)
- Swagger auto-généré (swaggo/swag + gin-swagger)

#### Backend
- Machine à états Mastery (UNKNOWN → FRAGILE → OK → SOLID) avec régressions et espacement 24h
- Scoring par type de question (RUBRIC, NUMERIC, KEYWORDS)
- Resserrement proportionnel des intervalles si exam posé
- Pipeline J0 (upload → OCR Gemini Flash → structuration LLM → items)
- Gestion OCR dégradé : timeout, photo floue, idempotence au restart
- Sessions de révision adaptatives (composition par pack, lazy generation, feedback enrichi)
- Reprise de session interrompue
- Mode dégradé sans emploi du temps, pas de pénalité maîtrise pour items en retard
- Validation HITL (confirm, correct, ignore, skip) avec seuil confidence et fidelity check LLM
- Onboarding (chapitre démo cold start, empty state guidé, écran progression J0)
- Adaptateurs S3/MinIO, OCR Gemini Flash, providers Gemini/Anthropic/Mistral
- Middleware auth JWT
- API REST /api/v1

#### Mobile
- 7 écrans Expo Router (dashboard, capture, session, chapter, onboarding, processing, settings)
- Design system appliqué (couleurs Mastery, typographie, MasteryBar animée)
- Client API avec auth et error handling
- 106 tests Jest

#### Documentation
- PRD complet, 171 ACs dont 53 Lot 0
- Architecture OCR/IDP incrémental
- User flow complet (20 écrans, 4 personas, 30 edge cases)
- Design system intergénérationnel (principes, couleurs, typo, composants, moodboard, scalabilité)
- Spécification fonctionnelle IHM admin (6 écrans, endpoints API, wireframes)
- Étude de faisabilité tests E2E mobile Maestro
- README principal avec décisions UX documentées

#### Tests
- 342 tests backend (coverage 46.3%)
- 106 tests mobile
- Tests unitaires domaine (mastery 100%, chapter 100%, session 100%, validation)
- Tests d'intégration cross-couches (HTTP → Service → DB)
- Tests infra S3 (MinIO) + OCR Gemini (contrat golden file)
- Tests middleware auth JWT (valide, expiré, malformé, absent)
- Tests HTTP handlers (mastery, session, validation, onboarding, pipeline, health)

#### Tooling
- Commande /sprint (cycle autonome sélection → implémentation → CI → feedback-loop)
- Commande /feedback-loop (boucle qualité post-sprint)
- Commande /backlog (analyse complète + alimentation backlog d'issues)
