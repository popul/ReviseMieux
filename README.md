# Révise mieux

Assistant d'étude IA pour collégiens et lycéens. Transforme vos notes de cours en fiches de révision, quiz interactifs et cartes mentales.

## Fonctionnalités

- **OCR intelligent** : Scannez vos notes manuscrites ou imprimées avec détection des zones incertaines
- **Fiches de révision** : Génération automatique de flashcards question/réponse
- **Quiz interactifs** : QCM personnalisés avec feedback et explications
- **Cartes mentales** : Visualisation des concepts clés et leurs relations
- **Ressources complémentaires** : Suggestions de vidéos et articles pour approfondir
- **Accessibilité** : Mode daltonien, contraste élevé, taille de texte ajustable

## Prérequis

- Docker et Docker Compose
- Clé API OpenAI (requis) ou Mistral (optionnel, fallback)

## Démarrage rapide

```bash
# 1. Cloner le dépôt
git clone https://github.com/votre-utilisateur/revisemieux.git
cd revisemieux

# 2. Configurer les variables d'environnement
cp .env.example .env
# Éditer .env et ajouter votre clé OPENAI_API_KEY

# 3. Lancer l'environnement de développement
make dev
```

Services disponibles :
- Frontend : http://localhost:3000
- Backend API : http://localhost:8080
- Base de données : localhost:5432

## Commandes disponibles

### Développement

```bash
make dev              # Démarre tous les services (Docker Compose)
make dev-frontend     # Lance le serveur de développement frontend uniquement
make dev-backend      # Lance le serveur backend uniquement
make install          # Installe les dépendances (Go + npm)
```

### Base de données

```bash
make db-up            # Démarre PostgreSQL uniquement
make db-down          # Arrête PostgreSQL
make db-reset         # Réinitialise la base de données
make db-migrate       # Applique les migrations
make db-migrate-down  # Annule la dernière migration
```

### Build et tests

```bash
make build            # Compile backend et frontend
make test             # Lance tous les tests
make lint             # Vérifie le code (Go + TypeScript)
```

### Docker

```bash
make docker-build     # Construit les images Docker
make docker-up        # Démarre les conteneurs
make docker-down      # Arrête les conteneurs
make docker-logs      # Affiche les logs
make docker-clean     # Supprime conteneurs et volumes
```

## Architecture

```
revisemieux/
├── backend/                 # API Go (Gin)
│   ├── cmd/server/         # Point d'entrée
│   ├── internal/
│   │   ├── api/            # Handlers HTTP et routes
│   │   ├── config/         # Configuration
│   │   ├── llm/            # Adaptateurs LLM (OpenAI, Mistral)
│   │   ├── services/       # Logique métier (OCR, génération, quotas)
│   │   └── store/          # Repositories PostgreSQL
│   └── migrations/         # Migrations SQL
├── frontend/               # Application React (Vite + Tailwind)
│   ├── src/
│   │   ├── components/     # Composants réutilisables
│   │   ├── pages/          # Pages (Dashboard, Scanner, Fiches, Quiz, Mindmap)
│   │   ├── services/       # Client API
│   │   └── contexte/       # Contextes React (accessibilité)
│   └── e2e/                # Tests Playwright
└── docker-compose.yml      # Configuration Docker
```

## Configuration

### Variables d'environnement

Créez un fichier `.env` à la racine du projet :

```bash
# Base de données
POSTGRES_USER=revisemieux
POSTGRES_PASSWORD=revisemieux
POSTGRES_DB=revisemieux

# Backend
DATABASE_URL=postgres://revisemieux:revisemieux@localhost:5432/revisemieux?sslmode=disable
PORT=8080

# LLM (au moins une clé requise)
OPENAI_API_KEY=sk-...
MISTRAL_API_KEY=          # Optionnel, utilisé en fallback

# Quotas journaliers (optionnel)
QUOTA_OCR=50              # Pages OCR par jour
QUOTA_GENERATION=100      # Générations par jour
```

## API Endpoints

### Santé

| Méthode | Endpoint | Description |
|---------|----------|-------------|
| GET | `/health` | État du serveur |
| GET | `/api/statut` | État avec connexion DB |

### Cours

| Méthode | Endpoint | Description |
|---------|----------|-------------|
| GET | `/api/cours` | Liste des cours |
| GET | `/api/cours/recents` | Cours récents avec stats |
| GET | `/api/cours/:id` | Détails d'un cours |
| POST | `/api/ocr` | Upload et OCR de fichiers |

### Génération

| Méthode | Endpoint | Description |
|---------|----------|-------------|
| POST | `/api/generer/fiches` | Générer des fiches |
| POST | `/api/generer/quiz` | Générer un quiz |
| POST | `/api/generer/mindmap` | Générer une mindmap |
| POST | `/api/generer/ressources` | Générer des ressources |

### Consultation

| Méthode | Endpoint | Description |
|---------|----------|-------------|
| GET | `/api/cours/:id/fiches` | Fiches d'un cours |
| GET | `/api/cours/:id/mindmap` | Mindmap d'un cours |
| GET | `/api/cours/:id/ressources` | Ressources d'un cours |
| GET | `/api/quiz/:id` | Détails d'un quiz |

### Quiz sessions

| Méthode | Endpoint | Description |
|---------|----------|-------------|
| POST | `/api/quiz/:id/demarrer` | Démarrer une session |
| POST | `/api/quiz/:id/session/:sid/repondre` | Répondre à une question |
| POST | `/api/quiz/:id/session/:sid/terminer` | Terminer et obtenir le score |

### Statistiques

| Méthode | Endpoint | Description |
|---------|----------|-------------|
| GET | `/api/statistiques` | Stats globales |
| GET | `/api/quotas` | Quotas restants |

## Tests

```bash
# Tests backend (Go)
cd backend && make test

# Tests frontend (Vitest)
cd frontend && npm test

# Tests E2E (Playwright)
cd frontend && npm run test:e2e

# Tests E2E en mode UI
cd frontend && npm run test:e2e:ui
```

## Design System

- **Couleurs** : Coral (#E85D4C), Teal (#1A4D4D), Gold (#F5C542), Cream (#FBF8F3)
- **Typographie** : Fraunces (titres), DM Sans (corps)
- **Accessibilité** : WCAG 2.1 AA (taille texte, daltonisme, contraste)

## Licence

Ce projet est sous licence privée. Tous droits réservés.
