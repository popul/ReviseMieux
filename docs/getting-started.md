# Guide de demarrage rapide

Ce guide couvre l'installation et le lancement de Revise Mieux en local (backend Go + mobile Expo).

---

## Prerequis

| Outil | Version | Verification |
|-------|---------|--------------|
| Go | 1.23+ | `go version` |
| Node.js | 20+ | `node --version` |
| Docker + Docker Compose | recent | `docker compose version` |
| PostgreSQL client (psql) | 16+ | `psql --version` |
| Expo CLI | via npx | `npx expo --version` |

Optionnel :
- `golangci-lint` pour le linting Go (`make backend-tools`)
- `swag` pour la generation Swagger (`make backend-tools`)

---

## Installation

```bash
# Cloner le depot
git clone https://github.com/popul/revisemieux.git
cd revisemieux

# Setup complet : demarre l'infra Docker, installe les deps, applique les migrations
make setup
```

La commande `make setup` execute dans l'ordre :
1. `infra-up` -- demarre PostgreSQL et Redis via Docker Compose
2. `deps` -- `go mod download` (backend) + `npm install` (mobile)
3. `db-migrate` -- applique les migrations SQL

---

## Variables d'environnement

Le backend charge sa configuration depuis les variables d'environnement. Les valeurs par defaut permettent de demarrer sans fichier `.env` pour le developpement local (sauf `JWT_SECRET` qui est obligatoire).

### Obligatoires

| Variable | Description | Exemple |
|----------|-------------|---------|
| `JWT_SECRET` | Cle secrete pour la signature des tokens JWT | `une-cle-secrete-longue-et-aleatoire` |

### Base de donnees et cache

| Variable | Description | Defaut |
|----------|-------------|--------|
| `DATABASE_URL` | URL de connexion PostgreSQL | `postgres://revisemieux:revisemieux@localhost:5432/revisemieux?sslmode=disable` |
| `REDIS_URL` | URL de connexion Redis | `redis://localhost:6379/0` |
| `MIGRATIONS_DIR` | Chemin vers les fichiers de migration SQL | `migrations` |

### Serveur HTTP

| Variable | Description | Defaut |
|----------|-------------|--------|
| `PORT` | Port d'ecoute du serveur | `8080` |
| `GIN_MODE` | Mode Gin (`debug` ou `release`) | `debug` |
| `JWT_EXPIRY` | Duree de validite des tokens JWT | `24h` |

### Storage S3 / MinIO

| Variable | Description | Defaut |
|----------|-------------|--------|
| `S3_ENDPOINT` | Endpoint S3 (MinIO en local) | `http://localhost:9000` |
| `S3_BUCKET` | Nom du bucket | `revisemieux` |
| `S3_ACCESS_KEY` | Cle d'acces S3 | (vide -- pipeline desactive si absent) |
| `S3_SECRET_KEY` | Cle secrete S3 | (vide -- pipeline desactive si absent) |
| `S3_REGION` | Region S3 | `us-east-1` |

### LLM (structuration et scoring)

| Variable | Description | Defaut |
|----------|-------------|--------|
| `LLM_PROVIDER` | Provider LLM actif : `gemini`, `anthropic`, `mistral` | `gemini` |
| `GOOGLE_AI_API_KEY` | Cle API Google AI (Gemini) -- sert aussi pour l'OCR | (vide) |
| `GEMINI_STRUCT_MODEL` | Modele Gemini pour la structuration | `gemini-2.5-flash` |
| `ANTHROPIC_API_KEY` | Cle API Anthropic | (vide) |
| `ANTHROPIC_STRUCT_MODEL` | Modele Anthropic pour la structuration | `claude-sonnet-4-6` |
| `ANTHROPIC_FIDELITY_MODEL` | Modele Anthropic pour le fidelity check | `claude-haiku-4-5` |
| `MISTRAL_API_KEY` | Cle API Mistral | (vide) |
| `MISTRAL_STRUCT_MODEL` | Modele Mistral pour la structuration | `mistral-small-latest` |

Le pipeline complet (upload -> OCR -> structuration -> items) n'est actif que si le storage S3, le service OCR (Google AI API key) et le LLM sont tous configures.

---

## Lancement en developpement

```bash
# Lance le backend (hot reload via go run) + le mobile (Expo) en parallele
# Demarre automatiquement l'infra Docker si necessaire
make dev
```

Pour lancer les composants separement :

```bash
# Backend seul
make backend-dev

# Mobile seul
make mobile-dev

# Infrastructure Docker seule
make infra-up
```

---

## Tests

```bash
# Tests unitaires (domaine + app, sans DB)
make test-unit

# Tests d'integration (necessite PostgreSQL)
make test-integration

# Tous les tests
make test

# Tests avec couverture
make backend-test-cover
```

Les tests unitaires du domaine n'ont besoin d'aucun mock (logique pure). Les tests d'integration utilisent une vraie base PostgreSQL.

---

## Qualite du code

```bash
# Lint (golangci-lint + go vet)
make lint

# Formatage du code Go
make fmt

# Lint + tests (CI gate)
make check
```

---

## Acces aux services

| Service | URL | Notes |
|---------|-----|-------|
| API backend | `http://localhost:8080` | Health check : `GET /health` |
| Swagger UI | `http://localhost:8080/swagger/index.html` | Disponible en mode `debug` |
| Token dev | `GET /dev/token` | Disponible en mode `debug`, genere un JWT de test |
| MinIO Console | `http://localhost:9001` | Login : `minioadmin` / `minioadmin` (profil `storage` requis) |
| PostgreSQL | `localhost:5432` | User : `revisemieux` / `revisemieux` |
| Redis | `localhost:6379` | Pas d'authentification en local |

Pour activer MinIO (necessaire pour le pipeline d'upload) :

```bash
docker compose --profile storage up -d
```

---

## Commandes utiles

| Commande | Description |
|----------|-------------|
| `make help` | Liste toutes les cibles disponibles |
| `make setup` | Installation complete |
| `make dev` | Backend + mobile en parallele |
| `make test` | Tous les tests |
| `make lint` | Lint complet |
| `make db-migrate` | Applique les migrations |
| `make db-reset` | Detruit et recree la DB |
| `make infra-reset` | Detruit les volumes Docker et redemarre |
| `make backend-swagger` | Regenere la doc Swagger |
| `make bench` | Lance le benchmark LLM |
| `make clean` | Nettoyage des artefacts |
