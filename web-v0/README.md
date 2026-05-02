# web-v0 — Lot -1 (one-shot revision sheet generator)

> Le **lot -1** de Revise Mieux : un site web one-page qui transforme des photos de cahier en fiche de révision HTML autonome. **Pas de suivi, pas de comptes, pas de base de données** — un formulaire, un appel LLM, un fichier HTML téléchargeable. Précède le Lot 0 (vrai backend Go + mobile React Native).

## Pourquoi un Lot -1

Avoir un produit utilisable **avant** le Lot 0 :

- valider l'utilité réelle de la fiche générée auprès d'un vrai utilisateur (Louis, 11-15 ans) ;
- alimenter la skill `study-guide` du dépôt principal en retours d'usage concrets ;
- itérer sur les critères d'acceptance pédagogiques (Sections 1→9, AC-CONT-*, AC-HTML-*, AC-20-*) sans bloquer le développement du backend Go.

## Stack

| Composant | Choix | Raison |
|---|---|---|
| Backend | FastAPI + Pydantic v2 + Uvicorn | minimal, async, pattern de jobs d'arrière-plan |
| Templating | Jinja2 (`shell.html.j2`) | shell HTML pré-baké, contenu dynamique en slots |
| LLM | Qwen 3.6 35B-A3B (vision + thinking) via LM Studio compatible OpenAI | tourne sur la machine de dev, pas de coût d'API |
| Empaquetage | Image Docker `ghcr.io/popul/revisemieux-web-v0` | pull simple côté homelab |

## Architecture

```
[Photos cahier] ──POST /jobs──┐
                              ▼
                      ┌───────────────┐
                      │  FastAPI app  │
                      │  (job queue)  │
                      └───────┬───────┘
                              │  HTTP /v1/chat/completions
                              ▼
                      ┌───────────────┐
                      │   LM Studio   │  Qwen 3.6 (vision + thinking)
                      │   sur Mac     │
                      └───────┬───────┘
                              │  HTML brut
                              ▼
                      ┌───────────────┐
                      │ Jinja shell + │  → fiche-revision.html
                      │ post-process  │     (autonome, ⌘P-friendly)
                      └───────────────┘
```

## Fichiers

| Fichier | Rôle |
|---|---|
| `main.py` | API FastAPI, job queue async, appel LLM, rendu Jinja |
| `schema.py` | Modèles Pydantic v2 décrivant la JSON attendue du LLM (Stage 2) |
| `prompt.txt` | Skill `study-guide` complète : pipeline OCR → structuration → fiche, critères d'acceptance |
| `shell.html.j2` | Shell HTML pré-baké (chrome stable) : drawer, boutons impression élève/parent, theme, mobile, `@media print` |
| `index.html` | Formulaire upload one-page |
| `fixture.json` | Fiche-jouet pour valider le shell sans LLM (`/preview`) |
| `requirements.txt` | Pin Python deps |
| `Dockerfile` | Image runtime |

## Endpoints

| Méthode | Route | Description |
|---|---|---|
| GET | `/` | Formulaire upload |
| POST | `/jobs` | Crée un job (multipart files + subject + target + extra), renvoie `{job_id}` |
| GET | `/jobs/{id}` | Statut du job (`pending`/`running`/`done`/`error`) |
| GET | `/jobs/{id}/result` | Télécharge la fiche HTML (200 si `done`, 409 sinon) |
| GET | `/preview` | Rend la `fixture.json` à travers le shell — **aucun appel LLM** (validation chrome) |
| GET | `/preview/raw` | Renvoie la fixture en JSON brut (inspection schéma) |
| GET | `/healthz` | Liveness |

## Dev local

```bash
cd web-v0
python -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
LLM_URL=http://localhost:1234/v1/chat/completions \
LLM_MODEL=qwen/qwen3.6-35b-a3b \
PROMPT_PATH=./prompt.txt \
INDEX_PATH=./index.html \
SHELL_PATH=./shell.html.j2 \
FIXTURE_PATH=./fixture.json \
uvicorn main:app --reload --port 8000
```

Puis http://localhost:8000 (formulaire) et http://localhost:8000/preview (fiche-jouet).

## Build local

```bash
docker build -t revisemieux-web-v0:dev web-v0
docker run --rm -p 8000:8000 \
  -e LLM_URL=http://host.docker.internal:1234/v1/chat/completions \
  revisemieux-web-v0:dev
```

## Build CI

Le workflow `.github/workflows/web-v0.yml` build et push sur GHCR à chaque push touchant `web-v0/**`. Tags publiés : `<sha>`, nom de branche, et `latest` sur `reboot`.

## Déploiement

Manifests Kubernetes (Deployment + Service + HTTPRoute) dans le dépôt **homelab** sous `services/revise/`. Le Deployment référence simplement l'image GHCR.

## Variables d'environnement

| Variable | Défaut | Rôle |
|---|---|---|
| `LLM_URL` | `http://macbook-pro-2.home:1234/v1/chat/completions` | endpoint OpenAI-compatible |
| `LLM_MODEL` | `qwen/qwen3.6-35b-a3b` | nom du modèle |
| `LLM_TIMEOUT` | `600` | secondes |
| `JOB_TTL_SECONDS` | `3600` | TTL des jobs en mémoire |
| `LOG_LEVEL` | `INFO` | niveau de log Python |

## Statut

- [x] Shell Jinja stable (drawer + impression + mobile + `@media print` validés via `/preview`)
- [x] Schéma Pydantic complet pour les Sections 1→9
- [x] Pipeline 1-stage HTML (legacy, conservé tant que Stage 2 n'est pas branché)
- [ ] Pipeline 2-stage JSON (rewrite `prompt.txt` en extraction → render Jinja) — en cours
- [ ] Tests d'intégration sur images réelles (HG, SVT, PC, français)
