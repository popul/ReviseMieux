# Guide de deploiement

Ce guide couvre la configuration et le deploiement de Revise Mieux.

---

## Infrastructure Docker Compose

Le fichier `docker-compose.yml` a la racine du projet definit 3 services :

### PostgreSQL 16

```yaml
services:
  postgres:
    image: postgres:16-alpine
    ports: 5432:5432
    volumes: pg_data:/var/lib/postgresql/data
```

| Variable | Defaut | Description |
|----------|--------|-------------|
| `POSTGRES_USER` | `revisemieux` | Utilisateur PostgreSQL |
| `POSTGRES_PASSWORD` | `revisemieux` | Mot de passe |
| `POSTGRES_DB` | `revisemieux` | Nom de la base |
| `POSTGRES_PORT` | `5432` | Port expose |

Health check : `pg_isready` toutes les 5 secondes.

### Redis 7

```yaml
services:
  redis:
    image: redis:7-alpine
    ports: 6379:6379
    volumes: redis_data:/data
```

| Variable | Defaut | Description |
|----------|--------|-------------|
| `REDIS_PORT` | `6379` | Port expose |

Health check : `redis-cli ping` toutes les 5 secondes.

### MinIO (S3-compatible, dev uniquement)

```yaml
services:
  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    ports: 9000:9000 (API), 9001:9001 (console)
    profiles: [storage]
```

MinIO est sur le profil `storage` -- il n'est pas demarre par defaut. Pour l'activer :

```bash
docker compose --profile storage up -d
```

| Variable | Defaut | Description |
|----------|--------|-------------|
| `MINIO_ROOT_USER` | `minioadmin` | Utilisateur admin MinIO |
| `MINIO_ROOT_PASSWORD` | `minioadmin` | Mot de passe admin |
| `MINIO_API_PORT` | `9000` | Port API S3 |
| `MINIO_CONSOLE_PORT` | `9001` | Port console web |

---

## Variables d'environnement du backend

### Configuration requise

| Variable | Obligatoire | Defaut | Description |
|----------|-------------|--------|-------------|
| `JWT_SECRET` | Oui | -- | Cle secrete pour la signature JWT. Le serveur refuse de demarrer sans. |
| `PORT` | Non | `8080` | Port d'ecoute HTTP |
| `GIN_MODE` | Non | `debug` | Mode Gin : `debug` (Swagger + /dev/token) ou `release` |
| `JWT_EXPIRY` | Non | `24h` | Duree de validite des tokens (format Go duration) |

### Base de donnees

| Variable | Defaut | Description |
|----------|--------|-------------|
| `DATABASE_URL` | `postgres://revisemieux:revisemieux@localhost:5432/revisemieux?sslmode=disable` | URL de connexion PostgreSQL |
| `REDIS_URL` | `redis://localhost:6379/0` | URL de connexion Redis |
| `MIGRATIONS_DIR` | `migrations` | Chemin vers les fichiers SQL de migration |

Les migrations sont appliquees automatiquement au demarrage du serveur (runner Go transactionnel avec table `schema_migrations`).

### Storage S3

| Variable | Defaut | Description |
|----------|--------|-------------|
| `S3_ENDPOINT` | `http://localhost:9000` | Endpoint S3 (MinIO en dev, AWS en prod) |
| `S3_BUCKET` | `revisemieux` | Nom du bucket |
| `S3_ACCESS_KEY` | (vide) | Cle d'acces |
| `S3_SECRET_KEY` | (vide) | Cle secrete |
| `S3_REGION` | `us-east-1` | Region |

Si `S3_ACCESS_KEY` ou `S3_SECRET_KEY` sont absents, le pipeline d'upload est desactive (le serveur demarre mais sans fonctionnalite de capture photo).

### Configuration LLM

Le backend supporte 3 providers LLM, selectionnes par la variable `LLM_PROVIDER` :

#### Gemini (defaut)

| Variable | Defaut | Description |
|----------|--------|-------------|
| `LLM_PROVIDER` | `gemini` | Provider actif |
| `GOOGLE_AI_API_KEY` | (vide) | Cle API Google AI |
| `GEMINI_STRUCT_MODEL` | `gemini-2.5-flash` | Modele pour la structuration |

La cle `GOOGLE_AI_API_KEY` sert aussi pour le service OCR (Gemini Flash VLM). Si elle est absente, l'OCR et la structuration sont desactives.

#### Anthropic

| Variable | Defaut | Description |
|----------|--------|-------------|
| `LLM_PROVIDER` | `anthropic` | |
| `ANTHROPIC_API_KEY` | (vide) | Cle API Anthropic |
| `ANTHROPIC_STRUCT_MODEL` | `claude-sonnet-4-6` | Modele pour la structuration |
| `ANTHROPIC_FIDELITY_MODEL` | `claude-haiku-4-5` | Modele pour le fidelity check HITL |

#### Mistral

| Variable | Defaut | Description |
|----------|--------|-------------|
| `LLM_PROVIDER` | `mistral` | |
| `MISTRAL_API_KEY` | (vide) | Cle API Mistral |
| `MISTRAL_STRUCT_MODEL` | `mistral-small-latest` | Modele pour la structuration |

Les providers Gemini et Mistral utilisent l'adaptateur OpenAI-compatible (`infra/openaicompat/`). Anthropic utilise son propre adaptateur (`infra/anthropic/`).

---

## Pipeline complet

Le pipeline J0 (photo -> items de revision) necessite que tous ces composants soient configures :

1. **Storage S3** -- `S3_ACCESS_KEY` + `S3_SECRET_KEY`
2. **OCR** -- `GOOGLE_AI_API_KEY` (Gemini Flash VLM)
3. **LLM structuration** -- la cle API du provider actif

Si l'un de ces composants manque, le `PipelineService` n'est pas initialise et les endpoints d'upload retournent une erreur.

---

## Health check

```
GET /health
```

Retourne `200 OK` avec la version du serveur. Disponible sans authentification.

---

## Demarrage du serveur

### Developpement

```bash
# Methode 1 : via Make (recommande)
make dev

# Methode 2 : directement
export JWT_SECRET="dev-secret-key"
cd backend && go run cmd/server/main.go
```

### Production

```bash
# Build
cd backend && make build

# Lancer
JWT_SECRET="..." \
DATABASE_URL="..." \
GIN_MODE=release \
LLM_PROVIDER=gemini \
GOOGLE_AI_API_KEY="..." \
S3_ENDPOINT="..." \
S3_ACCESS_KEY="..." \
S3_SECRET_KEY="..." \
./bin/revisemieux
```

Le binaire est compile sans CGO (`CGO_ENABLED=0`) pour un deploiement statique.

### Graceful shutdown

Le serveur intercepte `SIGINT` et `SIGTERM` et ferme proprement les connexions HTTP en cours (timeout 10 secondes). Le pool PostgreSQL est ferme apres l'arret du serveur.
