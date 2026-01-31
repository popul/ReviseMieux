# Codebase Analysis

## Current State Summary

### Frontend (Static HTML Prototypes)
**Status**: Complete UI prototypes, no backend integration

| Page | File | Description |
|------|------|-------------|
| Landing | `landing.html` | Marketing page with hero, features, CTA |
| Dashboard | `dashboard.html` | Student home with stats, recent courses, activity feed |
| Scan | `scan.html` | Upload/OCR flow with drag-drop, progress simulation |
| Cards | `cards.html` | Flashcard revision interface |
| Quiz | `quiz.html` | Interactive quiz with feedback |

**Tech stack:**
- Pure HTML/CSS/JS (no framework)
- Served via `npx serve`
- Design system fully implemented (Fraunces + DM Sans fonts, coral/teal/gold/cream palette)

**Key observations:**
- All interactions are simulated (no API calls)
- Files preview, progress bar, success states all work but with mock data
- UI is production-ready in terms of design quality

### Backend (Go API Skeleton)
**Status**: Routes defined, all handlers return placeholder responses

**Implemented:**
- chi router with middleware (Logger, Recoverer, RequestID, CORS)
- Environment loading via godotenv
- Route structure:
  ```
  GET  /health
  GET  /
  GET  /api/status
  GET  /api/courses
  POST /api/courses
  GET  /api/courses/{id}
  POST /api/ocr
  POST /api/generate/cards
  POST /api/generate/quiz
  POST /api/generate/mindmap
  ```

**Not implemented:**
- `internal/api/` - HTTP handlers (currently in main.go)
- `internal/llm/` - LLM adapter interface + OpenAI implementation
- `internal/ocr/` - OCR adapter interface + implementation
- `internal/store/` - PostgreSQL repositories
- Database migrations
- Request/response types

**Dependencies (go.mod):**
- chi/v5 (router)
- chi/cors
- godotenv

### Infrastructure
**Docker Compose:**
- PostgreSQL 16 Alpine (with healthcheck)
- Backend (builds from Dockerfile)
- Frontend (builds from Dockerfile)

**Missing:**
- Database migrations
- Backend Dockerfile likely incomplete (needs to be verified)

---

## Gap Analysis: MVP Requirements vs Current State

### PRD P0 Features

| Feature | Frontend | Backend | Status |
|---------|----------|---------|--------|
| OCR de cours | UI complete | Route only | **TODO** |
| Génération de fiches | UI complete | Route only | **TODO** |
| Quiz adaptatifs | UI complete | Route only | **TODO** |
| Enrichissement contenu | Not in UI | Not implemented | **P1 deprioritize for MVP** |
| Tableau de bord | UI complete | Routes only | **TODO** |

### Critical Path for MVP

1. **Database schema** - Store courses, cards, quizzes
2. **OCR integration** - Call OpenAI Vision API
3. **LLM integration** - Generate cards and quizzes from text
4. **API handlers** - Connect frontend to backend
5. **Frontend API integration** - Replace mock data with real calls

---

## Architecture Notes

### Adapter Pattern (from CLAUDE.md)
Backend should use adapter interfaces for:
- **LLM**: OpenAI (primary), Mistral (future), local models (future)
- **OCR**: OpenAI Vision API (primary)

This allows swapping providers without changing business logic.

### Anonymous Mode
- No authentication required for MVP
- Courses stored without user association
- Simplifies initial implementation

### Performance Targets (from PRD)
- OCR processing: < 3 seconds/page
- Content generation: < 20 seconds for cards/quizzes
- OCR accuracy: > 92% (printed), > 85% (handwritten)
