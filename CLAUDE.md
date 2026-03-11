# CLAUDE.md — Révise Mieux

## Projet

Révise Mieux est un SaaS éducatif qui transforme des photos de cahier en assistant de révision pour collégiens.

## Stack technique

| Composant | Technologie |
|-----------|-------------|
| Backend API | Go 1.23 + Gin |
| Base de données | PostgreSQL 16 (pgx/v5, SQL brut — pas d'ORM) |
| Mobile | React Native + Expo 55 + Expo Router |
| LLM | Anthropic API (Claude Sonnet) |
| Cache | Redis (go-redis/v9) |
| Storage | S3 / Object storage |
| IDs | UUIDv7 (timestamp-sortable, généré côté Go via `google/uuid`) |

**Monorepo** : `backend/` (Go) + `mobile/` (Expo)

---

## Documents clés

| Document | Contenu |
|----------|---------|
| `docs/PRD.md` | PRD complet (personas, pipeline, architecture, modèle de données, algorithmes, SLA) |
| `docs/MVP-scope.md` | Classification des 171 ACs et périmètre Lot 0 (53 ACs) |
| `docs/ac/Z1.md` à `docs/ac/Z8.md` | Critères d'acceptation détaillés par zone (format Given/When/Then) |
| `docs/lot0-tracker.md` | Suivi d'implémentation Lot 0 — phases, tâches, matrice AC |

---

## Concepts métier essentiels

- **Mastery** : machine à états UNKNOWN → FRAGILE → OK → SOLID avec régressions sur échec
- **Répétition espacée** : transition OK → SOLID requiert 2 réussites espacées de 24h minimum
- **Pipeline J0** : upload photo → segmentation → OCR → structuration LLM → génération items
- **HITL** : validation humaine (parent/admin) des items générés par le LLM
- **Lazy generation** : questions générées à la demande, cache 24h, invalidation sur validation ou CRUD exam
- **Lot 0** : 53 ACs pré-MVP pour usage local père-fils sur 4 chapitres pilotes

---

## Architecture — Domain-Driven Design

### Bounded Contexts

Le domaine se décompose en 4 contextes bornés, chacun avec son agrégat racine :

```
┌─────────────────────────────────────────────────────────────┐
│                      RÉVISE MIEUX                           │
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │   CAPTURE    │  │   MASTERY    │  │     SESSION      │  │
│  │              │  │              │  │                  │  │
│  │ Chapter ◄──┐ │  │ Mastery      │  │ Session          │  │
│  │  Revision  │ │  │  (user+item) │  │  Questions       │  │
│  │  Page      │ │  │              │  │  Attempts        │  │
│  │  Block     │ │  │ Transitions  │  │                  │  │
│  │  Item      │ │  │ Espacement   │  │ Composition      │  │
│  │  Notion    │ │  │ Régression   │  │ Lazy generation  │  │
│  │  VisualBlk │ │  │              │  │ Scoring          │  │
│  └──────────────┘  └──────────────┘  └──────────────────┘  │
│                                                             │
│  ┌──────────────────┐                                       │
│  │   VALIDATION     │                                       │
│  │                  │                                       │
│  │ ValidationTask   │                                       │
│  │  HITL actions    │                                       │
│  │  Fidelity check  │                                       │
│  └──────────────────┘                                       │
└─────────────────────────────────────────────────────────────┘
```

### Règles DDD

1. **Agrégat racine = point d'entrée unique.** Ne jamais modifier un `Item` directement — passer par `Chapter`. Ne jamais modifier un `Attempt` directement — passer par `Session`.
2. **Entités vs Value Objects.** Les enums métier (`MasteryState`, `ItemType`, `SessionStatus`) sont des Value Objects immuables. Les transitions Mastery sont modélisées comme des méthodes sur l'entité `Mastery`, pas comme du code procédural dans un handler.
3. **Domain Events.** Les transitions inter-contextes passent par des events, pas par des appels directs :
   - `ItemsGenerated` → déclenche création des `Mastery` UNKNOWN
   - `AttemptRecorded` → déclenche transition Mastery
   - `ValidationResolved` → déclenche invalidation cache questions
   - `ExamCreated` → déclenche resserrement intervalles
4. **Le domaine ne dépend de rien.** Les packages `domain/` n'importent ni Gin, ni pgx, ni Redis. Les dépendances pointent vers l'intérieur (hexagonal).
5. **Langage ubiquitaire.** Utiliser les termes du PRD dans le code : `Mastery` (pas `Progress`), `Item` (pas `Card`), `Notion` (pas `Topic`), `Chapter` (pas `Course`).

### Structure des packages Go

```
backend/
├── cmd/server/main.go              # Point d'entrée, wiring DI
├── internal/
│   ├── domain/                     # Coeur métier — ZERO dépendance externe
│   │   ├── mastery/                # Entité Mastery, transitions, value objects
│   │   │   ├── mastery.go          # Entité + méthodes de transition
│   │   │   ├── state.go            # Value object MasteryState
│   │   │   ├── repository.go       # Interface (port)
│   │   │   └── mastery_test.go     # Tests unitaires du domaine
│   │   ├── chapter/                # Agrégat Chapter (+ Item, Notion, Block)
│   │   ├── session/                # Agrégat Session (+ Question, Attempt)
│   │   ├── validation/             # Agrégat ValidationTask
│   │   └── event/                  # Domain events partagés
│   │
│   ├── app/                        # Application services (use cases, orchestration)
│   │   ├── mastery_service.go      # Orchestre transitions + events
│   │   ├── pipeline_service.go     # Orchestre pipeline J0
│   │   ├── session_service.go      # Composition + lazy generation
│   │   └── validation_service.go   # Actions HITL
│   │
│   ├── infra/                      # Adaptateurs (implémentations des ports)
│   │   ├── postgres/               # Repositories pgx
│   │   ├── redis/                  # Cache
│   │   ├── s3/                     # Storage objets
│   │   ├── anthropic/              # Client LLM
│   │   └── ocr/                    # Client OCR externe
│   │
│   ├── http/                       # Couche HTTP (handlers Gin, middleware, DTOs)
│   │   ├── handler/
│   │   ├── middleware/
│   │   ├── dto/                    # Request/Response structs (jamais les entités domaine)
│   │   └── router.go
│   │
│   └── config/                     # Configuration env
│
├── migrations/                     # SQL migrations (numérotées)
└── testdata/                       # Fixtures, golden files
```

**Règle clé** : les imports vont **toujours vers l'intérieur**.
- `http/` → importe `app/` et `domain/`
- `app/` → importe `domain/`
- `infra/` → importe `domain/` (implémente les interfaces)
- `domain/` → **n'importe rien** du projet

### Architecture Hexagonale (Ports & Adapters)

Le projet suit strictement l'architecture hexagonale. Le domaine est au centre, protégé de toute dépendance technique par des **ports** (interfaces) et des **adaptateurs** (implémentations).

```
                         ┌──────────────────────────────┐
                         │      Adaptateurs Driving      │
                         │  (HTTP handlers, CLI, tests)  │
                         └──────────┬───────────────────┘
                                    │ appelle
                                    ▼
                         ┌──────────────────────────────┐
                         │     Application Services      │
                         │  (use cases, orchestration)   │
                         │  Dépend des PORTS (interfaces)│
                         └──────────┬───────────────────┘
                                    │ utilise
                                    ▼
┌───────────────────┐   ┌──────────────────────────────┐   ┌───────────────────┐
│ Adaptateurs Driven│◄──│         DOMAINE               │──►│ Adaptateurs Driven│
│  (postgres, redis │   │  Entités, Value Objects,      │   │  (S3, Anthropic,  │
│   implémentent    │   │  Règles métier, PORTS         │   │   OCR)            │
│   les ports)      │   │  (Repository, Clock, etc.)    │   │                   │
└───────────────────┘   └──────────────────────────────┘   └───────────────────┘
```

#### Principes stricts

1. **Le domaine est le centre.** Les packages `domain/` ne connaissent aucune technologie (pas de Gin, pgx, Redis, S3, HTTP). Ils définissent les **ports** (interfaces) que les adaptateurs implémentent.

2. **Ports = interfaces dans `domain/`.** Chaque bounded context expose ses ports :
   - `mastery.Repository` — persistance des Mastery
   - `chapter.Repository` — persistance de l'agrégat Chapter
   - `session.Repository` — persistance de l'agrégat Session
   - `event.Publisher` — publication des domain events
   - Les ports techniques (`Clock`, `IDGenerator`) vivent aussi dans `domain/`

3. **Adaptateurs = implémentations dans `infra/`.** Chaque technologie a son package :
   - `infra/postgres/` implémente les `Repository` interfaces avec pgx
   - `infra/redis/` implémente le cache
   - `infra/s3/` implémente le storage
   - `infra/anthropic/` implémente le client LLM

4. **Sens des dépendances : toujours vers l'intérieur.** Jamais `domain/` n'importe `infra/` ou `http/`. Le wiring (injection des adaptateurs dans les services) se fait uniquement dans `cmd/server/main.go`.

5. **Application services (`app/`) orchestrent.** Ils reçoivent les ports par injection de constructeur et coordonnent les appels entre domaine et infrastructure. Ils ne contiennent pas de logique métier — celle-ci vit dans les entités du domaine.

6. **Handlers HTTP (`http/handler/`) sont des adaptateurs driving.** Ils traduisent HTTP ↔ DTOs, appellent les services applicatifs, et mappent les erreurs domaine vers des status codes HTTP. Zéro logique métier.

7. **Testabilité par design.** Grâce aux ports :
   - Les tests unitaires du domaine n'ont besoin d'aucun mock (logique pure)
   - Les tests des services `app/` mockent les ports (interfaces)
   - Les tests d'intégration `infra/` utilisent une vraie DB
   - Les tests HTTP mockent les services

8. **Pas de fuite d'abstraction.** Les entités domaine ne sortent jamais dans les réponses HTTP — utiliser des DTOs (`http/dto/`). Les structures pgx ne remontent jamais au-delà de `infra/postgres/`.

---

## Test-Driven Development

### Philosophie

**Écrire le test AVANT le code.** Chaque AC en Given/When/Then se traduit directement en un test Go. Le cycle est :
1. Écrire le test (rouge)
2. Écrire le minimum de code pour passer (vert)
3. Refactorer (bleu)

### Patterns de test Go

**Table-driven tests** pour les transitions Mastery et le scoring :

```go
func TestMasteryTransition(t *testing.T) {
    tests := []struct {
        name     string
        initial  MasteryState
        score    float64
        want     MasteryState
        wantCS   int  // consecutive_successes attendu
    }{
        // Z1-AC01: UNKNOWN + succès → FRAGILE
        {"unknown_success", Unknown, 1.0, Fragile, 1},
        // Z1-AC11: UNKNOWN + échec → reste UNKNOWN
        {"unknown_failure", Unknown, 0.0, Unknown, 0},
        // Z1-AC05: SOLID + échec → OK
        {"solid_failure", Solid, 0.0, OK, 0},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            m := NewMastery(tt.initial)
            m.RecordAttempt(tt.score, time.Now())
            assert(t, m.State, tt.want)
            assert(t, m.ConsecutiveSuccesses, tt.wantCS)
        })
    }
}
```

**Tests d'intégration** avec une vraie DB PostgreSQL (testcontainers ou DB de test) :

```go
func TestMasteryRepository_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    db := setupTestDB(t)  // pool pgx vers DB de test
    repo := postgres.NewMasteryRepository(db)
    // ...
}
```

### Stratégie de test par couche

| Couche | Type de test | Quoi tester | Quoi mocker |
|--------|-------------|-------------|-------------|
| `domain/` | **Unitaire pur** | Transitions, règles métier, validations | Rien — pas de dépendance |
| `app/` | **Unitaire + mocks** | Orchestration, enchaînement d'appels | Repositories (interfaces), clients externes |
| `infra/postgres/` | **Intégration** | Requêtes SQL, mappings | Rien — vraie DB |
| `http/handler/` | **HTTP** | Status codes, DTOs, validation input | Services (interfaces) |

### Conventions

- `go test ./...` doit passer en < 30 secondes (tests unitaires)
- `go test -tags=integration ./...` pour les tests avec DB
- Nommer les tests d'après l'AC : `TestZ1AC01_UnknownToFragile`
- Un test par AC minimum, plus de tests pour les edge cases
- Pas de `time.Sleep` dans les tests — injecter une `Clock` interface pour l'espacement 24h

---

## Conventions Go

### Générales

- **Pas d'ORM.** SQL brut via pgx. Les requêtes vivent dans `infra/postgres/`.
- **Pas de framework DI.** Le wiring se fait dans `cmd/server/main.go` via constructeurs explicites.
- **Errors wrapping.** Utiliser `fmt.Errorf("mastery.RecordAttempt: %w", err)` avec le contexte.
- **Interfaces définies côté consommateur** (dans `domain/`), pas côté implémenteur (dans `infra/`).
- **Constructeurs `New*`** qui valident les invariants : `NewMastery(state)` refuse un state vide.

### Nommage

- Packages en un mot minuscule : `mastery`, `chapter`, `session` (pas `mastery_service`)
- Fichiers : `snake_case.go`
- Structs exported : `PascalCase`
- Interfaces : nom du comportement (`Repository`, `Clock`, `EventPublisher`), pas `IRepository`
- Contexte : toujours premier paramètre `ctx context.Context`

### Erreurs métier

Définir des erreurs sentinelles dans le domaine :

```go
var (
    ErrTransitionBlocked  = errors.New("transition blocked: spacing requirement not met")
    ErrItemArchived       = errors.New("item is archived")
    ErrSessionNotResumable = errors.New("session cannot be resumed from current status")
)
```

Les handlers HTTP mappent ces erreurs vers les status codes appropriés.

---

## Conventions Mobile (React Native + Expo)

- **Expo Router** pour le routing (file-based)
- **TypeScript strict** — `strict: true` dans tsconfig
- Composants fonctionnels + hooks uniquement
- État global minimal — React Context pour auth, sinon état local + props
- **Pas de state management lourd** (pas de Redux) — TanStack Query pour le cache serveur
- Nommage composants : `PascalCase.tsx`, hooks : `use*.ts`

---

## SQL & Migrations

- **Migrations numérotées** : `NNN_description.sql` (ex: `002_fix_trigger_column.sql`)
- **SQL brut** — pas de migration generator
- **UUIDv7** généré côté Go (pas `gen_random_uuid()` qui produit du v4)
- Enums PostgreSQL pour les valeurs finies connues, `TEXT` + `CHECK` pour les valeurs extensibles
- `TIMESTAMPTZ` partout (jamais `TIMESTAMP`)
- `ON DELETE CASCADE` sur les tables enfants d'un agrégat, `ON DELETE SET NULL` pour les refs cross-agrégat
- Index nommés `idx_{table}_{colonnes}`

---

## Règles de développement

### Workflow

1. Lire l'AC correspondante dans `docs/ac/Z*.md`
2. Écrire le test en Given/When/Then (TDD red)
3. Implémenter le minimum pour passer le test (TDD green)
4. Refactorer si nécessaire (TDD refactor)
5. Mettre à jour `docs/lot0-tracker.md` (statut `[x]`)

### Ce qu'il ne faut PAS faire

- Importer Gin ou pgx dans `domain/`
- Retourner une entité domaine dans une réponse HTTP (utiliser un DTO)
- Tester la logique métier via des tests HTTP end-to-end
- Ajouter une feature non listée dans le Lot 0 tracker
- Utiliser `interface{}` / `any` quand un type concret existe
- Écrire un mock quand on peut tester avec la vraie implémentation (domaine pur)
- Mettre de la logique métier dans les handlers HTTP

### Code review checklist

- [ ] Le test existe et couvre l'AC
- [ ] Le domaine ne dépend d'aucun package infra
- [ ] Les erreurs sont wrappées avec contexte
- [ ] Les transitions Mastery passent par l'entité, pas par SQL direct
- [ ] Les DTOs sont séparés des entités domaine
- [ ] Les migrations sont idempotentes ou versionnées
- [ ] Le tracker Lot 0 est mis à jour
