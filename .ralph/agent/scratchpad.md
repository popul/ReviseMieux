# Scratchpad - Session 2026-02-04

## Analysis

After exploring the codebase, the project is ~90% feature-complete for P0 MVP. All major features exist:

**Fully Working:**
- OCR processing (images + PDF) ✅
- Flashcard generation & review ✅
- Adaptive quiz with session tracking ✅
- Mindmap visualization ✅
- Dashboard with statistics ✅
- Accessibility compliance ✅
- Rate limiting & quotas ✅
- E2E tests ✅

**Key Gaps Identified:**

1. **Course CRUD handlers are stubs** - `handlers.go` lines 71-93 return empty/placeholder responses, but the repository (`cours_repo.go`) has full implementation. Need to connect them.

2. **Frontend has no "Cours" page** - Dashboard shows recent courses but clicking them doesn't work since there's no course detail page and lister endpoint returns empty.

3. **Resources UI missing** - Backend has `GenererRessourcesHandler` and `ObtenirRessourcesHandler` working, but no frontend component displays them.

## Priority

The course CRUD stubs are the most critical gap - they break the core flow:
- User can't view their course list
- User can't view a specific course details
- This blocks navigation in the Dashboard "Voir tous" link

## Plan

1. **Implement Course CRUD Handlers** - Connect the stubbed handlers to the repository
2. **Add Course List/Detail Pages** - Frontend pages to view courses
3. **Add Resources Display** - Component to show generated resources in course detail

Starting with: Course CRUD Handlers (backend first, then frontend)

## Iteration 1 - Completed

**Task:** Implement Course CRUD Handlers

**Done:**
- Added `coursRepo` field to Handlers struct
- Implemented `ListerCoursHandler` with pagination (page, limite)
- Implemented `CreerCoursHandler` for manual course creation
- Implemented `ObtenirCoursHandler` to get a single course by ID
- Added `SupprimerCoursHandler` for course deletion
- Added DELETE route to routes.go
- All handlers follow the same error response pattern as other endpoints

**Verified:**
- Build succeeds
- All tests pass (31 passed, 19 skipped - skipped tests are API tests without backend)

**Committed:** feat(api): implement course CRUD handlers

**Next iteration:** Frontend needs to be updated to use these endpoints (Cours list page, course detail view)

## Iteration 2 - Completed

**Task:** Frontend: Cours list page

**Done:**
- Created `frontend/src/pages/Cours.tsx` with full course list and detail views
- Course list shows all courses with pagination, icons per matière, OCR confidence
- Each course card has quick links to Fiches, Quiz, Mindmap
- Course detail view shows full text content and zones incertaines
- Added course deletion with confirmation dialog
- Updated App.tsx to use Cours page instead of PageEnConstruction placeholder
- Updated Dashboard "Voir tous" link to navigate to /cours instead of /fiches

**Verified:**
- TypeScript compiles without errors
- All tests pass (31 passed, 19 skipped)

**Committed:** feat(frontend): add Cours list page with pagination and detail view

**Next iteration:** The MVP P0 features are largely complete. Remaining gaps:
- Resources UI is missing (backend has handlers but no frontend component)
- Could improve by adding Resources display in course detail
- Could add E2E tests for the new Cours page

## Iteration 3 - Completed

**Task:** Add Resources UI to course detail page

**Done:**
- Added `Ressource` interface and `ReponseRessources` type to `frontend/src/services/api.ts`
- Added `obtenirRessourcesCours(coursId)` to fetch existing resources
- Added `genererRessources(coursId)` to generate new resources via LLM
- Added Resources section to `CoursDetail` component:
  - Displays existing resources with type icons (video 🎬, article 📄, exercice ✏️, cours 📖, autre 🔗)
  - "Générer des ressources" button when no resources exist
  - Warning about AI-generated links ("Les liens suggérés sont générés par IA...")
  - Loading spinner during fetch/generation
  - Error handling with user-friendly messages
  - External links open in new tab with noopener

**Verified:**
- TypeScript compiles without errors
- ESLint passes on modified files
- (E2E tests require dev server, not running in CI)

**Committed:** feat(frontend): add Resources UI to course detail page

**P0 MVP Status:**
All P0 features from PRD are now implemented:
- ✅ OCR de cours (images + PDF)
- ✅ Génération de fiches de révision
- ✅ Quiz adaptatifs
- ✅ Enrichissement de contenu (Resources UI) ← Just completed
- ✅ Tableau de bord simple

## P1 Feature Assessment - 2026-02-04

Explored codebase to determine P1 feature status:

| Feature | Backend | Frontend | Status |
|---------|---------|----------|--------|
| Génération de mindmaps | ✅ | ✅ | **DONE** |
| OCR de copies corrigées | ❌ | ❌ | Missing |
| Analyse des erreurs | ❌ | ❌ | Missing |
| Recommandations personnalisées | ❌ | ❌ | Missing |
| Historique de progression | ⚠️ | ❌ | Partial (basic stats exist, detailed view missing) |

**Priority for P1 Implementation:**

1. **Historique de progression (complete)** - Foundation already exists:
   - Quiz sessions store scores
   - Backend `/api/statistiques` returns basic stats
   - Dashboard shows recent activity
   - Need: Create `/progression` page with trends over time, per-subject breakdown

2. **OCR de copies corrigées + Analyse des erreurs** - These are coupled:
   - Need: Database schema for exam copies
   - Need: Backend handlers for exam upload and OCR
   - Need: Error classification service (compréhension, méthode, inattention)
   - Need: Frontend pages for upload and analysis display

3. **Recommandations personnalisées** - Depends on error analysis:
   - Once errors are tracked, can generate targeted suggestions
   - Requires significant LLM work for personalized recommendations

## Plan for P1

Starting with **Historique de progression** since infrastructure exists and it provides immediate value.

## Iteration 4 - Completed

**Task:** Implement Progression page frontend

**Done:**
Backend:
- Added `SessionAvecDetails` struct and `ListerSessionsCompletes()` to `quiz_repo.go`
- Added `Progression`, `HistoriqueQuiz`, `StatistiquesParMatiere` types to `statistiques.go`
- Added `ObtenirProgression()` service method that aggregates history and per-subject stats
- Added `GET /api/progression` endpoint with handlers

Frontend:
- Created `Progression.tsx` page with:
  - Global stats (quiz completed, average score, number of subjects)
  - Per-subject breakdown cards showing average and best scores
  - Quiz history list with score color-coding (green >= 80%, gold >= 60%, orange >= 40%, red < 40%)
  - Empty state when no quiz completed
- Added API types and `obtenirProgression()` function to `api.ts`
- Updated `App.tsx` to use Progression page instead of placeholder

**Verified:**
- Backend builds successfully
- Frontend builds successfully (tsc + vite)
- All E2E tests pass (31 passed, 19 skipped)

**Committed:** feat: add Progression page with quiz history and per-subject stats

**P1 Feature Status Update:**

| Feature | Backend | Frontend | Status |
|---------|---------|----------|--------|
| Génération de mindmaps | ✅ | ✅ | **DONE** |
| Historique de progression | ✅ | ✅ | **DONE** ← Just completed |
| OCR de copies corrigées | ❌ | ❌ | Missing |
| Analyse des erreurs | ❌ | ❌ | Missing |
| Recommandations personnalisées | ❌ | ❌ | Missing |

**Next:** The remaining P1 features (OCR copies, error analysis, recommendations) are coupled and require significant new infrastructure

## Iteration 5 - Planning OCR Copies + Error Analysis

### Exploration Summary

Backend follows clean patterns:
- **Layer 1:** HTTP handlers (Gin) - request/response
- **Layer 2:** Services - business logic & LLM interaction
- **Layer 3:** Repositories - database CRUD
- **Layer 4:** LLM Adapters - external integrations

Key files to follow as templates:
- OCR: `backend/internal/services/ocr.go`
- Generation: `backend/internal/services/generation.go`
- Handlers: `backend/internal/api/handlers_ocr.go`, `handlers_generation.go`
- Repos: `backend/internal/store/cours_repo.go`

### Implementation Plan for OCR Copies + Error Analysis

**Task Breakdown:**

1. **Database migration** - Create `copies_examens` and `erreurs_analyse` tables
2. **Copie Repository** - CRUD for exam copies
3. **Erreur Repository** - CRUD for error analysis results
4. **Error Analysis Service** - LLM-based error classification
5. **OCR Copie Handlers** - POST /api/copies/ocr to upload and scan
6. **Error Analysis Handlers** - POST /api/copies/{id}/analyser to trigger analysis
7. **Frontend: Upload Copie page** - Similar to scan page but for exam copies
8. **Frontend: Analysis display** - Show error breakdown and recommendations

Starting with: Database migration (foundation for everything else)

### Task Completed: Database Migration

Created 4 migration files:
- `009_create_copies_examens.up.sql` / `.down.sql`
- `010_create_erreurs_analyse.up.sql` / `.down.sql`

**copies_examens table:**
- id, cours_id (nullable FK), titre, matiere
- note_obtenue, note_totale (for grade tracking)
- texte_ocr, annotations_professeur (teacher corrections)
- confiance, zones_incertaines, fichiers_originaux (OCR metadata)
- date_examen, date_creation, date_modification
- Indexes on cours_id, date_creation, matiere

**erreurs_analyse table:**
- id, copie_id (FK CASCADE)
- type_erreur: 'comprehension', 'methode', 'inattention'
- texte_original, correction, explication, conseil
- severite: 'legere', 'moderate', 'grave'
- position_debut, position_fin (for highlighting)
- Indexes on copie_id, type_erreur

**Committed:** feat(db): add migrations for exam copies and error analysis

**Next iteration:** Create CopieExamen repository following cours_repo.go pattern

## Iteration 6 - Completed

**Task:** Create CopieExamen and ErreurAnalyse repositories

**Done:**
- Created `backend/internal/store/copie_repo.go` (568 lines)
- `CopieExamen` struct with all fields from migration (nullable grades, course link, exam date)
- `ErreurAnalyse` struct with type, severity, position highlighting
- `CopieExamenRepository` interface: Creer, ObtenirParID, Lister, ListerParCours, MettreAJour, Supprimer, Compter
- `ErreurAnalyseRepository` interface: CreerPlusieurs (batch with transaction), ListerParCopie, SupprimerParCopie, CompterParType
- Helper functions: nullFloat64, nullTime, nullInt for nullable field handling
- Follows same pattern as cours_repo.go (interface verification, error handling, JSON serialization)

**Verified:**
- Backend builds successfully (`make build`)
- No new lint issues on the new file

**Committed:** feat(store): add CopieExamen and ErreurAnalyse repositories

**Next iteration:** Create error analysis service (LLM-based error classification)

## Iteration 7 - Completed

**Task:** Create Error Analysis Service (LLM-based error classification)

**Done:**
- Created `backend/internal/services/analyse_erreurs.go` (~280 lines)
- `ServiceAnalyseErreurs` with dependencies on LLM, CopieExamenRepository, ErreurAnalyseRepository, CoursRepository
- `AnalyserCopie()` method:
  - Fetches the exam copy by ID
  - Optionally includes course context for better analysis
  - Calls LLM with a detailed prompt to classify errors
  - Parses response into `ErreurAnalyse` structs
  - Replaces old analysis (delete + insert) for re-analysis support
- Error types classified into: comprehension, methode, inattention
- Severity levels: legere, moderate, grave
- `ResultatAnalyseErreurs` includes:
  - List of errors with explanations and advice
  - Summary by error type (counts)
  - Global advice for improvement
  - Points forts / Points à améliorer lists
- `ObtenirErreursParCopie()` to retrieve existing analysis
- `CompterErreursParType()` for statistics

**LLM Prompt designed to:**
- Analyze student copy text + teacher annotations
- Classify each error by type and severity
- Provide pedagogical explanations
- Give actionable advice per error
- Identify strengths and areas for improvement

**Verified:**
- Backend builds successfully (`make build`)
- All tests pass (31 passed, 19 skipped)

**Next iteration:** Create HTTP handlers for exam copies (upload OCR + analysis endpoints)

## Iteration 8 - Completed

**Task:** Create HTTP handlers for exam copies (OCR + analysis endpoints)

**Done:**
- Created `backend/internal/api/handlers_copies.go` (~450 lines)
- `HandlersCopies` struct with dependencies on ServiceOCR, ServiceAnalyseErreurs, CopieExamenRepository, ErreurAnalyseRepository
- Implemented handlers:
  - `TraiterOCRCopieHandler` - POST /api/copies/ocr - OCR upload for exam copies
  - `ListerCopiesHandler` - GET /api/copies - List all copies with pagination
  - `ObtenirCopieHandler` - GET /api/copies/:id - Get single copy
  - `SupprimerCopieHandler` - DELETE /api/copies/:id - Delete copy
  - `AnalyserCopieHandler` - POST /api/copies/:id/analyser - Trigger LLM error analysis
  - `ObtenirErreursHandler` - GET /api/copies/:id/erreurs - Get existing error analysis
- Updated `handlers.go`:
  - Added `handlersCopies` field to Handlers struct
  - Updated `NouveauHandlers()` to accept serviceAnalyseErreurs, copieRepo, erreurRepo
  - Added delegation methods for all copies handlers
- Updated `routes.go`:
  - Added /api/copies group with all routes
  - OCR and analysis routes use quota middleware
- Updated `cmd/server/main.go`:
  - Create copieRepo and erreurRepo from store
  - Create serviceAnalyseErreurs with all dependencies
  - Pass all new dependencies to NouveauHandlers

**API Endpoints:**
- `GET /api/copies` - List copies with pagination
- `POST /api/copies/ocr` - Upload exam copy for OCR (with quota)
- `GET /api/copies/:id` - Get single copy details
- `DELETE /api/copies/:id` - Delete a copy
- `POST /api/copies/:id/analyser` - Analyze errors in copy (with quota)
- `GET /api/copies/:id/erreurs` - Get analysis results

**Verified:**
- Backend builds successfully (`make build`)
- Frontend builds successfully
- All E2E tests pass (31 passed, 19 skipped)

**Next iteration:** Create frontend pages for copies upload and analysis display

## Iteration 9 - Completed

**Task:** Create Analyser.tsx frontend page for exam copies and error analysis

**Done:**
Frontend:
- Created `frontend/src/pages/Analyser.tsx` (~790 lines) with complete workflow:
  - **Liste state**: Shows existing scanned exam copies with grades, matière, deletion
  - **Upload state**: File upload with ZoneUpload component, form for titre/matière/note/annotations
  - **OCR Processing state**: Loading spinner during extraction
  - **Copie Detail state**: Shows extracted text, OCR confidence, teacher annotations, "Analyser" button
  - **Analyse Processing state**: Loading spinner during LLM analysis
  - **Resultats state**: Full analysis display with:
    - Summary by error type (comprehension, méthode, inattention)
    - Points forts / Points à améliorer sections
    - Global conseil
    - Detailed error cards with original text, correction, explanation, and advice
    - Color-coded severity (légère, modérée, grave)
  - **Error state**: Error display with retry option
  - Confirmation modal for deletion
  - URL params sync (copie=id) for deep linking
- Added API types and functions to `api.ts`:
  - `CopieExamen`, `ErreurAnalyse`, `ResultatAnalyse` types
  - `ReponseCopie`, `ReponseCopies`, `ReponseAnalyse`, `ReponseErreurs` types
  - `envoyerOCRCopie()`, `listerCopies()`, `obtenirCopie()`, `supprimerCopie()`
  - `analyserCopie()`, `obtenirErreursCopie()`
- Updated `App.tsx`:
  - Replaced `PageEnConstruction` placeholder with real `Analyser` component
  - Removed unused PageEnConstruction function

**Verified:**
- Backend builds successfully (`make build`)
- Frontend builds successfully (tsc + vite)
- All E2E tests pass (31 passed, 19 skipped)

**Committed:** feat(frontend): add Analyser page for exam copies and error analysis

**P1 Feature Status Update:**

| Feature | Backend | Frontend | Status |
|---------|---------|----------|--------|
| Génération de mindmaps | ✅ | ✅ | **DONE** |
| Historique de progression | ✅ | ✅ | **DONE** |
| OCR de copies corrigées | ✅ | ✅ | **DONE** ← Just completed |
| Analyse des erreurs | ✅ | ✅ | **DONE** ← Just completed |
| Recommandations personnalisées | ❌ | ❌ | Remaining P1 feature |

**Summary:**
4 of 5 P1 features are now complete. The remaining feature (Recommandations personnalisées) would require additional LLM work to generate targeted study suggestions based on error patterns.

## Iteration 10 - In Progress

**Task:** Implement personalized recommendations backend service

**Done:**
Backend:
- Created `backend/internal/services/recommandations.go` (~320 lines)
  - `ServiceRecommandations` with dependencies on LLM, copieRepo, erreurRepo, coursRepo, quizRepo
  - `GenererRecommandations()` main method:
    - Collects data from error analysis and quiz results
    - Calls LLM with structured prompt to generate recommendations
    - Returns typed recommendations with priority ranking
  - `collecterDonnees()` aggregates:
    - Errors from exam copy analysis (type, severity counts)
    - Weak quiz results (score < 60%)
    - Points to improve from error analysis
  - `ResultatRecommandations` includes:
    - List of 3-5 prioritized recommendations
    - Domain, reason, severity, suggested action
    - Resume summary
    - Action plan for the week
    - Next quiz suggestion
    - Motivation message
  - `GenererRecommandationsPrioritaires()` for top 3 recommendations
- Created `backend/internal/api/handlers_recommandations.go` (~150 lines)
  - `GenererRecommandationsCopieHandler` - POST /api/copies/:id/recommandations
  - `ObtenirRecommandationsPrioritairesHandler` - GET /api/recommandations/prioritaires
  - `ObtenirRecommandationsParMatiereHandler` - GET /api/recommandations/par-matiere/:matiere
- Updated `handlers.go`:
  - Added `handlersRecommandations` field
  - Updated `NouveauHandlers()` to accept serviceRecommandations
  - Added delegation methods for all 3 recommendation handlers
- Updated `routes.go`:
  - Added POST route for /api/copies/:id/recommandations with quota middleware
  - Added /api/recommandations group with prioritaires and par-matiere routes
- Updated `cmd/server/main.go`:
  - Create serviceRecommandations with all dependencies
  - Pass to NouveauHandlers

**API Endpoints:**
- `POST /api/copies/:id/recommandations` - Generate recommendations for analyzed copy (with quota)
- `GET /api/recommandations/prioritaires?matiere=` - Get top 3 priority recommendations
- `GET /api/recommandations/par-matiere/:matiere` - Get recommendations for a specific subject

**Verified:**
- Backend builds successfully (`make build`)
- Frontend builds successfully
- All E2E tests pass (31 passed, 19 skipped)

**Next:** Frontend integration - add recommendations display to Analyser.tsx

## Iteration 10 - Completed

**Task:** Add recommendations UI to Analyser.tsx (frontend integration)

**Done:**
Frontend API (`api.ts`):
- Added `Recommandation` interface with domaine, raison, severiteMax, actionSuggerie, typeQuiz, priorite
- Added `ResultatRecommandations` interface with recommandations array, resume, planAction, prochainQuiz, motivation
- Added `ReponseRecommandations` interface for API response
- Added `genererRecommandationsCopie(copieId, options?)` function to POST /api/copies/:id/recommandations

Frontend Page (`Analyser.tsx`):
- Added `PRIORITES` config for priority display (1=Urgent/coral, 2=Important/gold, 3=Normal/teal, etc.)
- Added `recommandations` and `chargementRecommandations` state
- Added `genererRecommandations()` function to call API and update state
- Added "Recommandations personnalisees" section in results view:
  - "Generer des recommandations" button when no recommendations exist
  - Loading state with spinner
  - Motivation message (encouragement)
  - Resume summary
  - Prioritized recommendation cards with:
    - Priority number badge (1-5)
    - Domain name
    - Severity indicator
    - Reason explanation
    - Action suggérée
    - Quiz type suggestion
  - Plan d'action pour la semaine (weekly action plan)
  - Prochain quiz suggestion
- Reset recommendations when reinitializing page

**Verified:**
- TypeScript compiles without errors (`tsc -b`)
- No lint errors in modified files (Analyser.tsx, api.ts)
- Backend builds successfully (`make build`)

**P1 Feature Status - COMPLETE:**

| Feature | Backend | Frontend | Status |
|---------|---------|----------|--------|
| Génération de mindmaps | ✅ | ✅ | **DONE** |
| Historique de progression | ✅ | ✅ | **DONE** |
| OCR de copies corrigées | ✅ | ✅ | **DONE** |
| Analyse des erreurs | ✅ | ✅ | **DONE** |
| Recommandations personnalisées | ✅ | ✅ | **DONE** ← Just completed |

**All P1 features are now implemented!**

## Final Status - 2026-02-04

### P0 Features (MVP) - COMPLETE
- ✅ OCR de cours (scan/photo with text recognition)
- ✅ Génération de fiches de révision
- ✅ Quiz adaptatifs
- ✅ Enrichissement de contenu
- ✅ Tableau de bord simple

### P1 Features (Post-MVP) - COMPLETE
- ✅ Génération de mindmaps
- ✅ Historique de progression
- ✅ OCR de copies corrigées
- ✅ Analyse des erreurs
- ✅ Recommandations personnalisées

### Verification
- Backend: `make build` passes
- Frontend: `tsc -b && vite build` passes (69 modules, 347KB JS, 43KB CSS)
- E2E tests: 31 passed, 19 skipped

### Next Steps (P2 Features)
Per PRD, P2 features are "Nice to have":
- Mode collaboratif (file sharing between students)
- Flashcards (spaced repetition)
- Export multi-formats (PDF, print)
- Mode hors-ligne
- Intégration agenda

**OBJECTIVE COMPLETE: All P0 and P1 features from the PRD are now implemented.**
