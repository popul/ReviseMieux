# Guide de contribution

Ce guide decrit le workflow de developpement, les conventions de code et la checklist de revue pour Revise Mieux.

---

## Workflow TDD

Chaque critere d'acceptation (AC) en Given/When/Then se traduit directement en un test Go. Le cycle est strict :

1. **Rouge** -- Ecrire le test. Il echoue car le code n'existe pas encore.
2. **Vert** -- Ecrire le minimum de code pour que le test passe.
3. **Bleu** -- Refactorer si necessaire, en verifiant que le test passe toujours.

### Processus pour une AC

1. Lire l'AC correspondante dans `docs/ac/Z*.md`
2. Ecrire le test en Given/When/Then (TDD red)
3. Implementer le minimum pour passer le test (TDD green)
4. Refactorer si necessaire (TDD refactor)
5. Mettre a jour `docs/lot0-tracker.md` (statut `[x]`)

---

## Conventions de nommage Go

### Packages

- Nom en **un mot minuscule** : `mastery`, `chapter`, `session` (pas `mastery_service`)
- Fichiers en `snake_case.go`

### Types

- Structs exportees en `PascalCase`
- Interfaces nommees d'apres le comportement : `Repository`, `Clock`, `Publisher` (pas `IRepository`)
- Enums metier comme Value Objects immuables : `MasteryState`, `ItemType`, `SessionStatus`

### Fonctions et methodes

- Constructeurs `New*` qui valident les invariants : `NewMastery(state)` refuse un state vide
- Premier parametre toujours `ctx context.Context`
- Errors wrapping avec contexte : `fmt.Errorf("mastery.RecordAttempt: %w", err)`

### Erreurs metier

Erreurs sentinelles definies dans le domaine :

```go
var (
    ErrTransitionBlocked   = errors.New("transition blocked: spacing requirement not met")
    ErrItemArchived        = errors.New("item is archived")
    ErrSessionNotResumable = errors.New("session cannot be resumed from current status")
)
```

Les handlers HTTP mappent ces erreurs vers les status codes HTTP correspondants.

---

## Patterns de test

### Nommage des tests

Nommer les tests d'apres l'AC couverte :

```go
func TestZ1AC01_UnknownToFragile(t *testing.T) { ... }
func TestZ1AC04_SpacingBlocksOKToSolid(t *testing.T) { ... }
func TestZ2AC05_LLMFailureFallback(t *testing.T) { ... }
```

Un test par AC minimum, plus de tests pour les cas limites.

### Table-driven tests

Utiliser des table-driven tests pour les transitions Mastery et le scoring :

```go
func TestMasteryTransition(t *testing.T) {
    tests := []struct {
        name     string
        initial  MasteryState
        score    float64
        want     MasteryState
        wantCS   int
    }{
        {"unknown_success", Unknown, 1.0, Fragile, 1},
        {"unknown_failure", Unknown, 0.0, Unknown, 0},
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

### Strategie de test par couche

| Couche | Type de test | Quoi tester | Quoi mocker |
|--------|-------------|-------------|-------------|
| `domain/` | Unitaire pur | Transitions, regles metier, validations | Rien -- pas de dependance |
| `app/` | Unitaire + mocks | Orchestration, enchainement d'appels | Repositories (interfaces), clients externes |
| `infra/postgres/` | Integration | Requetes SQL, mappings | Rien -- vraie DB |
| `http/handler/` | HTTP | Status codes, DTOs, validation input | Services (interfaces) |

### Regles

- `go test ./...` doit passer en moins de 30 secondes (tests unitaires)
- `go test -tags=integration ./...` pour les tests avec DB
- Pas de `time.Sleep` dans les tests -- injecter une `Clock` interface pour l'espacement 24h
- Ecrire un mock uniquement quand on ne peut pas tester avec la vraie implementation

---

## Regles d'architecture

### Ce qu'il ne faut PAS faire

- Importer Gin ou pgx dans `domain/`
- Retourner une entite domaine dans une reponse HTTP (utiliser un DTO dans `http/dto/`)
- Tester la logique metier via des tests HTTP end-to-end
- Ajouter une feature non listee dans le Lot 0 tracker
- Utiliser `interface{}` / `any` quand un type concret existe
- Mettre de la logique metier dans les handlers HTTP
- Modifier un `Item` directement -- passer par l'agregat `Chapter`
- Modifier un `Attempt` directement -- passer par l'agregat `Session`

### Sens des dependances

Les imports vont toujours vers l'interieur :
- `http/` importe `app/` et `domain/`
- `app/` importe `domain/`
- `infra/` importe `domain/` (implemente les interfaces)
- `domain/` n'importe rien du projet

---

## Checklist PR

Avant de soumettre une pull request, verifier :

- [ ] Le test existe et couvre l'AC
- [ ] Le domaine ne depend d'aucun package infra
- [ ] Les erreurs sont wrappees avec contexte
- [ ] Les transitions Mastery passent par l'entite, pas par SQL direct
- [ ] Les DTOs sont separes des entites domaine
- [ ] Les migrations sont idempotentes ou versionnees
- [ ] Le tracker Lot 0 est mis a jour

---

## Commandes utiles

| Commande | Description |
|----------|-------------|
| `make test` | Tous les tests (unitaires + integration) |
| `make test-unit` | Tests unitaires uniquement |
| `make test-integration` | Tests d'integration (necessite DB) |
| `make lint` | Lint complet (golangci-lint + go vet) |
| `make fmt` | Formatage du code Go |
| `make check` | Lint + tests (CI gate) |
| `make bench` | Benchmark LLM (IDP) |
| `make bench-ocr` | Benchmark OCR |
| `make bench-all` | Benchmark tous les modeles |
| `make bench-report` | Genere le rapport HTML des benchmarks |
| `make backend-test-mastery` | Tests du domaine Mastery uniquement |
| `make backend-swagger` | Regenere la doc Swagger |
| `make backend-tools` | Installe golangci-lint et swag |
