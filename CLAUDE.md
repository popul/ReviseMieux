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
3. **Domain Events.** Les transitions inter-contextes passent par des events via le `SyncDispatcher`, pas par des appels directs :
   - `AttemptRecorded` → déclenche transition Mastery (Session → Mastery)
   - `ExamCreated` → déclenche resserrement intervalles (Exam → Mastery)
   - `ItemsGenerated` → déclenche création des `Mastery` UNKNOWN (Pipeline → Mastery)
   - `ValidationResolved` → déclenche invalidation cache questions (Validation → Session)
   
   **Chaque event DOIT avoir un consumer enregistré dans `cmd/server/main.go` via `dispatcher.On("event.name", handler)`.** Un event sans consumer est du code mort — le `SyncDispatcher` logge les events mais seuls les handlers enregistrés produisent des effets.
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
│   │   ├── eventbus/               # SyncDispatcher (event routing)
│   │   ├── s3/                     # Storage objets
│   │   ├── anthropic/              # Client LLM
│   │   ├── openaicompat/           # Client LLM OpenAI-compatible (Gemini, Mistral)
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

6. **Handlers HTTP (`http/handler/`) sont des adaptateurs driving.** Ils traduisent HTTP ↔ DTOs, appellent les services applicatifs, et mappent les erreurs domaine vers des status codes HTTP. Zéro logique métier. **Les handlers ne reçoivent JAMAIS de Repository en injection** — uniquement des services `app/`. Si un handler a besoin de données, ajouter une méthode au service correspondant.

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
- **CHECK constraints obligatoires** sur tout champ numérique à domaine borné :
  - Scores et confidences : `CHECK (score >= 0 AND score <= 1)`
  - Compteurs : `CHECK (consecutive_successes >= 0)`
  - Difficulté : `CHECK (difficulty BETWEEN 1 AND 5)`
  - Durées/tokens : `CHECK (duration_ms >= 0)`
- **Cohérence struct ↔ table** : chaque champ du struct Go domaine DOIT avoir sa colonne SQL correspondante. Si un champ est ajouté au struct, la migration ET le repository (SELECT + INSERT/UPDATE) doivent être mis à jour dans le même commit.

---

## Règles de développement

### Workflow

1. Lire l'AC correspondante dans `docs/ac/Z*.md`
2. Écrire le test en Given/When/Then (TDD red)
3. Implémenter le minimum pour passer le test (TDD green)
4. Refactorer si nécessaire (TDD refactor)
5. **Vérifier la Definition of Done** (voir ci-dessous)
6. Lancer `make check` (doit passer avant tout commit)
7. Mettre à jour `docs/lot0-tracker.md` (statut `[x]`)

### Definition of Done — quand une AC est VRAIMENT terminée

Une AC ne peut être marquée `[x]` dans le tracker que si **TOUS** ces critères sont remplis :

1. **Test domaine** : un test unitaire prouve la logique métier (ex: `TestZ1AC01_UnknownToFragile`)
2. **Persistence complète** : chaque champ de l'entité domaine a sa colonne SQL correspondante, et le repository le lit ET l'écrit
3. **Event consumer** : si l'AC déclenche un domain event, il existe un handler enregistré dans le dispatcher qui produit l'effet attendu (pas juste un `log.Println`)
4. **Bout en bout vérifiable** : un test `app/` ou `handler/` exerce le chemin complet (handler → service → repo/event)
5. **CHECK constraints SQL** : tout champ numérique avec un domaine de valeur (score ∈ [0,1], difficulty ∈ [1,5]) a un CHECK en DB
6. **`make check` passe** : format + vet + lint (golangci-lint) + imports domaine + tests unitaires

**Symptôme d'une AC faussement cochée** : le code existe mais il manque une migration SQL, un champ n'est pas persisté, un event n'a pas de consumer, ou le test ne couvre que le happy path.

### Garanties exécutables (CI gate)

```bash
make check      # format + vet + lint + domain imports + tests unitaires
make check-ci   # reproduit EXACTEMENT la CI (check + integration + garde-fou anti-skip)
make check-ci-act  # optionnel : exécute ci.yml + integration.yml via nektos/act (attrape les incompat d'actions tierces). Requiert act ≥ 0.2.87 et assez d'espace Docker (prévoir >5 GB libres, sinon `docker system prune -af`).
```

`make check` est le filet de sécurité minimal.

**Règle : tout nouveau gate CI (workflow job, filtre, action GitHub) doit avoir une contrepartie exécutable en local.** La logique vit dans `make check-ci`, les workflows `.github/workflows/*.yml` ne doivent faire que l'invoquer ou la dupliquer à l'identique. Historique : un garde-fou anti-skip trop strict et un `golangci-lint-action@v6` incompatible sont passés en CI parce qu'ils n'étaient testables qu'en poussant. Avant tout push touchant `.github/workflows/`, `Makefile`, ou `.golangci.yml` : lancer `make check-ci` (et idéalement `make check-ci-act`). Il est composé de :
- `fmt-check` : le code est formaté (`gofmt`)
- `vet` : `go vet ./...`
- `check-domain` : script `scripts/check-domain-imports.sh` vérifie que `domain/` n'importe jamais `infra/`, `http/`, `app/`, Gin, pgx, etc.
- `test-unit` : tests du domaine et des services

**Règle : ne jamais committer si `make check` échoue.**

### Ce qu'il ne faut PAS faire

- Importer Gin ou pgx dans `domain/`
- Retourner une entité domaine dans une réponse HTTP (utiliser un DTO)
- Tester la logique métier via des tests HTTP end-to-end
- Ajouter une feature non listée dans le Lot 0 tracker
- Utiliser `interface{}` / `any` quand un type concret existe
- Écrire un mock quand on peut tester avec la vraie implémentation (domaine pur)
- Mettre de la logique métier dans les handlers HTTP
- **Injecter un Repository dans un handler HTTP.** Les handlers dépendent uniquement des services `app/`. Si un handler a besoin de données, ajouter une méthode au service, pas un repo en paramètre.
- **Publier un event sans consumer.** Chaque `publisher.Publish(event.Xxx{})` doit avoir un `dispatcher.On("xxx", handler)` correspondant dans `main.go`. Un event sans consumer est du code mort.
- **Ajouter un champ à une entité domaine sans migration SQL.** Si le struct Go a un champ, la table doit avoir la colonne, le repository doit le lire/écrire, et les valeurs numériques doivent avoir un CHECK constraint.
- **Marquer une AC `[x]` sans vérifier la persistence.** "Le code compile" ≠ "ça marche". Vérifier que le champ est dans le SELECT, l'INSERT, l'UPDATE du repository.
- **Ignorer les erreurs de `publisher.Publish()`.** Toujours vérifier le retour d'erreur.
- **Utiliser `t.Skip()` dans un test d'intégration.** Un skip silencieux masque des régressions (incident : `make test-integration` a passé en vert pendant une période alors qu'il skippait tous les tests à cause d'un mismatch `DATABASE_URL`/`TEST_DATABASE_URL`). Un test d'intégration doit échouer bruyamment si ses préconditions ne sont pas réunies — utiliser `t.Fatal()` avec un message d'action ("run: make test-db-up"). La CI refuse désormais tout `Action: "skip"` sur les tests integration-tagged.
- **Jamais de pansement.** Si un fix nécessite de contourner un mauvais design, corriger le design d'abord. La dette technique s'accumule silencieusement et coûte exponentiellement plus tard. Un refactoring propre maintenant vaut mieux qu'un workaround qui deviendra permanent.
- **Ne jamais sauter à une cause plausible sans lire les logs jusqu'au bout.** Quand une commande échoue, quand un container est "unhealthy", quand un test flaky — la première action est de **lire le message d'erreur complet** (`docker logs`, `go test -v`, stderr en entier), pas de proposer une explication qui "sonne bien". Historique : "postgres:16 unhealthy sous act" a été diagnostiqué à tort comme un problème d'émulation ARM/health check alors que la vraie cause était `No space left on device` dans Docker Desktop — visible en 2 secondes avec `docker logs`. Un diagnostic plausible mais faux coûte plus cher qu'un "je ne sais pas, je regarde". Règle : si tu n'as pas vu le message d'erreur avec tes yeux, tu n'as pas diagnostiqué — tu as deviné.
- **Ne jamais committer un target/gate/script sans l'avoir exécuté au moins une fois.** "Ça devrait marcher" n'est pas une vérification. Historique : `make check-ci-act` a été poussé sans test, puis a fallu deux itérations correctives parce que (a) `golangci-lint-action@v6` ne supporte pas golangci-lint v2, (b) `actions/upload-artifact@v4` réclame un token absent en local, (c) act avait besoin de `--container-architecture linux/amd64`. Tous ces problèmes auraient été vus au premier lancement local.

### Code review checklist

- [ ] Le test existe et couvre l'AC (domaine + service)
- [ ] Le domaine ne dépend d'aucun package infra (`make check-domain`)
- [ ] Les erreurs sont wrappées avec contexte
- [ ] Les transitions Mastery passent par l'entité, pas par SQL direct
- [ ] Les DTOs sont séparés des entités domaine
- [ ] Les migrations sont idempotentes ou versionnées
- [ ] Tout champ numérique a un CHECK constraint en SQL
- [ ] Tout champ du struct domaine est persisté (SELECT + INSERT/UPDATE)
- [ ] Les handlers n'injectent aucun Repository directement
- [ ] Tout event publié a un consumer enregistré dans le dispatcher
- [ ] `make check` passe
- [ ] Le tracker Lot 0 est mis à jour
