# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Révise mieux** is an AI-powered study assistant for middle and high school students. The app transforms handwritten or printed course notes (via OCR) into study materials: revision cards, interactive quizzes, and mind maps. Students can also scan corrected exams for personalized error analysis.

**Current status**: Frontend prototypes complete, backend infrastructure in progress.

## Repository Structure (Monorepo)

```
/
├── Makefile              # Central Makefile: build, test, lint, docker commands
├── docker-compose.yml    # Local development environment
├── backend/              # Go backend API
│   ├── Dockerfile
│   ├── Makefile
│   ├── cmd/              # Entry points
│   ├── internal/         # Private application code
│   │   ├── api/          # HTTP handlers
│   │   ├── llm/          # LLM adapter interface + implementations (OpenAI, Mistral, local)
│   │   ├── ocr/          # OCR adapter interface + implementations
│   │   └── store/        # PostgreSQL repositories
│   └── go.mod
├── frontend/             # TypeScript frontend
│   ├── Dockerfile
│   ├── Makefile
│   ├── package.json
│   └── src/
├── helm/                 # Helm charts for Kubernetes deployment
│   └── revisemieux/
├── docs/
│   ├── PRD.md            # Product Requirements Document (French)
│   └── wireframe.html    # Original UI/UX prototype
└── .github/
    └── workflows/        # GitHub Actions CI
```

## Tech Stack

| Component | Technology |
|-----------|------------|
| Backend | Go |
| Frontend | TypeScript |
| Database | PostgreSQL |
| LLM/OCR | OpenAI API (primary), Mistral adapter, local models adapter (planned) |
| Deployment | Helm charts on Kubernetes |
| CI/CD | GitHub Actions |
| Container Registry | GitHub Container Registry (ghcr.io) |
| Local Dev | Docker Compose |

## Local Development

### Prerequisites

- Docker & Docker Compose
- Go 1.21+ (for local backend dev)
- Node.js 20+ (for local frontend dev)
- Make

### Quick Start (Docker)

```bash
# Start everything (database + backend + frontend)
make dev

# Services will be available at:
# - Frontend: http://localhost:3000
# - Backend API: http://localhost:8080
# - PostgreSQL: localhost:5432
```

### Development Commands

```bash
# Full environment
make dev              # Start all services with Docker Compose
make docker-down      # Stop all services
make docker-logs      # View logs from all containers
make docker-clean     # Stop and remove all containers + volumes

# Database only (for local backend/frontend development)
make db-up            # Start only PostgreSQL
make db-down          # Stop PostgreSQL
make db-reset         # Reset database (delete all data)

# Local development (without Docker for app code)
make install          # Install all dependencies (go mod + npm install)
make dev-backend      # Run Go backend locally (requires db-up)
make dev-frontend     # Run frontend dev server locally
```

### Environment Variables

Create a `.env` file at the root for API keys:

```bash
OPENAI_API_KEY=sk-...
MISTRAL_API_KEY=...   # Optional
```

## Build Commands

```bash
# Central Makefile (from root)
make build          # Build all components
make test           # Test all components
make lint           # Lint all components
make docker-build   # Build all Docker images

# Backend (from backend/)
make build          # go build
make test           # go test ./...
make lint           # golangci-lint run

# Frontend (from frontend/)
make build          # npm run build
make test           # npm test
make lint           # npm run lint
```

## Architecture Notes

**LLM/OCR Adapter Pattern**: The backend uses an adapter interface for LLM and OCR providers. Currently implements OpenAI, with Mistral adapter planned. Design for easy addition of local model support.

**Anonymous Mode**: The site supports anonymous usage. Courses can be stored without user authentication.

**Performance targets from PRD:**
- OCR processing: < 3 seconds/page
- Content generation: < 20 seconds for cards/quizzes
- OCR accuracy: > 92% (printed), > 85% (handwritten)

## Key Documentation

- **docs/PRD.md**: Complete product specs, user personas, feature priorities (P0/P1/P2), success metrics, 3-phase roadmap
- **docs/wireframe.html**: Original interactive prototype
- **frontend/*.html**: Current frontend design (landing, dashboard, scan, cards, quiz)

## Important Constraints

- **Language**: All user-facing content is in French
- **GDPR/RGPD compliance**: Working with minors' data requires strict data protection from day one
- **Accessibility**: Must support adjustable text size, colorblind mode, and high contrast
- **Design system**: Warm editorial palette with coral (#E85D4C), teal (#1A4D4D), gold (#F5C542), cream (#FBF8F3). Typography: Fraunces (display) + DM Sans (body).

## Frontend Pages

| Page | File | Description |
|------|------|-------------|
| Landing | `frontend/landing.html` | Marketing page, hero, features, CTA |
| Dashboard | `frontend/dashboard.html` | Student home, stats, recent courses |
| Scan | `frontend/scan.html` | Upload/OCR flow with progress |
| Cards | `frontend/cards.html` | Flashcard revision interface |
| Quiz | `frontend/quiz.html` | Interactive quiz with feedback |
