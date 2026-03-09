# Révise Mieux

Assistant de révision personnalisé pour collégiens (11–15 ans), propulsé par l'IA.

**Révise Mieux** transforme des photos de cahier (manuscrit, schémas, documents) en entraînements adaptatifs : carte de leçon structurée, exercices en rappel actif avec répétition espacée, contrôles blancs, et reporting parent basé sur des preuves de maîtrise.

## Statut du projet

Le dépôt contient actuellement les **spécifications complètes** du produit. L'implémentation n'a pas encore démarré.

## Documentation

| Document | Description |
|---|---|
| [PRD](docs/PRD.md) | Product Requirements Document complet — personas, parcours, pipeline, architecture, modèle de données, algorithmes, SLA |
| [MVP Scope](docs/MVP-scope.md) | Classification des 166 critères d'acceptation (MVP Core / Hardening / Post-MVP) et périmètre Lot 0 |
| [Critères d'acceptation](docs/ac/README.md) | 166 ACs en format Given/When/Then, répartis en 8 zones |

### Zones de critères d'acceptation

| Zone | Sujet | ACs |
|---|---|---|
| [Z1](docs/ac/Z1.md) | Transitions Mastery & répétition espacée | 28 |
| [Z2](docs/ac/Z2.md) | Pipeline J0 — erreurs & timeouts | 18 |
| [Z3](docs/ac/Z3.md) | Validation HITL (Human-in-the-loop) | 17 |
| [Z4](docs/ac/Z4.md) | Lazy generation, concurrence & cache | 17 |
| [Z5](docs/ac/Z5.md) | Versioning chapitre & identité item | 10 |
| [Z6](docs/ac/Z6.md) | Emploi du temps, notifications & engagement parent | 43 |
| [Z7](docs/ac/Z7.md) | Routine de soirée & orchestration | 25 |
| [Z8](docs/ac/Z8.md) | Onboarding & première utilisation | 8 |

## Chapitres pilotes (MVP)

| Matière | Chapitre | Pack |
|---|---|---|
| Histoire-Géographie | Les inégalités dans le monde | HG-INEG |
| Histoire-Géographie | La société féodale | HG-FEOD |
| SVT | La photosynthèse | SVT-PHOTO |
| Physique-Chimie | Masse, volume et densité | PC-MVD |

## Stack cible

| Composant | Technologie |
|---|---|
| API | Node.js ou Go |
| OCR & Vision | Python (workers async) |
| LLM | Anthropic Claude API |
| Base de données | PostgreSQL |
| Cache | Redis |
| File de messages | BullMQ / SQS |
| Stockage | S3 |
| Frontend | React / React Native (mobile-first) |

## Licence

Projet privé — tous droits réservés.
