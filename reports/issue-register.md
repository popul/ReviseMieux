# Issue Register -- Audit 2026-04-03 (v2 corrige)

## Legende severite

| Severite | Description |
|----------|-------------|
| CRITIQUE | Code DDD supprime, ACs non presentes dans le code actuel |
| MAJEUR | Ecart significatif entre specs et code, necessite correction |
| MINEUR | Ecart de convention ou de completude, non bloquant |
| INFO | Observation, contexte utile |

---

## Issues

### CRITIQUE

| # | Titre | Description | Fichiers concernes | AC | Recommandation |
|---|-------|-------------|--------------------|----|----------------|
| C1 | Code DDD supprime | Les 5 bounded contexts (`domain/mastery`, `domain/chapter`, `domain/session`, `domain/validation`, `domain/event`) ont ete supprimes entre `e543dea` et `ec69500` lors d'un merge avec une branche de dev parallele. 12,674 lignes supprimees au total (domain: 2,172, app: 4,721, infra: 3,945, http: 1,836). | `backend/internal/domain/**`, `backend/internal/app/**`, `backend/internal/infra/**`, `backend/internal/http/**` | Z1-Z8 (53 ACs) | `git checkout e543dea -- backend/internal/domain/ backend/internal/app/` pour restaurer |
| C2 | 85 tests TestZ*AC* supprimes | Tous les tests nommes TestZ\*AC\* couvrant les 53 ACs Lot 0 ont ete supprimes avec le code DDD. Les 25 tests restants couvrent uniquement config, OCR hybride, et utils. | `backend/internal/domain/mastery/mastery_test.go`, `backend/internal/domain/session/scoring_test.go`, `backend/internal/app/*_test.go`, `backend/internal/infra/postgres/*_test.go` | Z1-Z8 | Restaurer depuis `e543dea` |
| C3 | Machine a etats Mastery absente | L'entite Mastery (UNKNOWN/FRAGILE/OK/SOLID) avec transitions, espacement 24h, regressions, resserrement exam, plafond n'existe plus dans le code actuel. | `backend/internal/domain/mastery/mastery.go` (supprime) | Z1-AC01 a Z1-AC13 | Restaurer `domain/mastery/` (205 lignes) |
| C4 | Pipeline J0 absent | Le pipeline en 7+ etapes (segmentation, OCR, structuration, items, notions, fidelity check) n'existe plus. Remplace par un OCR simple + generation directe. | `backend/internal/app/pipeline_service.go` (supprime) | Z2-AC01 a Z2-AC10 | Restaurer ou reimplementer |
| C5 | Sessions de revision absentes | L'agregat Session (composition, lazy generation, attempts, scoring par type) n'existe plus. Le quiz actuel est statique, pas adaptatif. | `backend/internal/domain/session/`, `backend/internal/app/session_service.go` (supprimes) | Z4-AC04 a Z4-AC09 | Restaurer ou reimplementer |
| C6 | Validation HITL absente | L'agregat ValidationTask (Confirmer, Corriger, Ignorer, Je ne sais pas) n'existe plus. | `backend/internal/domain/validation/`, `backend/internal/app/validation_service.go` (supprimes) | Z3-AC01 a Z3-AC10 | Restaurer ou reimplementer |
| C7 | Tracker et specs supprimes du working tree | `docs/lot0-tracker.md`, `docs/ac/Z1.md`-`Z8.md`, `docs/MVP-scope.md` n'existent plus dans le code actuel. | `docs/` | Toutes les ACs | Restaurer depuis `e543dea` |
| C8 | Schema DB non conforme | Les tables actuelles (`cours`, `fiches`, `quiz`, etc.) ne correspondent pas au schema DDD (`chapters`, `masteries`, `sessions`, `items`, etc.). | `backend/migrations/001-020` | Tout le modele | Planifier migration DB |

### MAJEUR

| # | Titre | Description | Fichiers concernes | Recommandation |
|---|-------|-------------|--------------------|----|
| M1 | Architecture plate vs hexagonale | Le code actuel suit `api/services/store/llm/ocr` au lieu de `domain/app/infra/http`. Pas de ports, pas d'adapters, logique metier dans services et handlers. | `backend/internal/` | Migrer progressivement vers hexagonal |
| M2 | Pas de Gemini pour OCR | Le benchmark recommandait Gemini Flash pour l'OCR. Le code utilise OpenAI GPT (primaire) + Mistral (fallback). | `backend/internal/llm/` | Ajouter un adaptateur Gemini |
| M3 | Pas de pipeline 2 etages | Les specs prescrivent OCR (vision) + IDP (texte) separes. Le code fait tout en 1 passe. | `backend/internal/services/ocr.go` | Separer les etapes |
| M4 | Prompts en dur dans le code Go | Les specs prescrivent des fichiers `.txt` avec `go:embed`. Les prompts sont des strings dans `services/generation.go`, `services/concepts.go`, etc. | `backend/internal/services/generation.go`, `backend/internal/services/concepts.go` | Externaliser prompts |
| M5 | Frontend web vs mobile RN | Le CLAUDE.md specs prescrit React Native + Expo. Le code contient un frontend TypeScript/Vite/Tailwind. | `frontend/` | Arbitrage produit necessaire |
| M6 | Pas d'auth JWT | Les routes sont sans authentification. Les specs prescrivent JWT Bearer sur `/api/v1/*`. | `backend/internal/api/routes.go` | Implementer middleware auth |
| M7 | Pas de prefix /api/v1 | Routes sous `/api` sans versioning. | `backend/internal/api/routes.go` | Ajouter prefix v1 |
| M8 | UUID v4 au lieu de UUIDv7 | Les migrations utilisent `uuid_generate_v4()`. Les specs prescrivent UUIDv7 genere cote Go. | `backend/migrations/001_create_cours.up.sql`, `backend/internal/store/cours_repo.go` | Migrer vers google/uuid v7 |
| M9 | 2 CLAUDE.md contradictoires | Le CLAUDE.md racine decrit l'architecture actuelle (plate, TS frontend, OpenAI). Les instructions projet CLAUDE.md decrivent l'architecture DDD cible. | `CLAUDE.md` | Unifier ou separer clairement etat actuel vs cible |

### MINEUR

| # | Titre | Description | Fichiers concernes | Recommandation |
|---|-------|-------------|--------------------|----|
| m1 | Nommage en francais vs anglais | Le code utilise des noms francais (`Cours`, `Fiches`, `Charger`, `NouveauGestionnaire`). Les specs prescrivent le langage ubiquitaire en anglais (`Mastery`, `Item`, `Chapter`). | Tout le backend | Aligner sur le choix final (francais OU anglais) |
| m2 | Entites DB exposees en HTTP | `store.Cours`, `store.Quiz` sont directement serialises dans les reponses. Pas de DTOs. | `backend/internal/api/handlers*.go` | Ajouter une couche DTO |
| m3 | Interfaces cote implementeur | `CoursRepository`, `QuizRepository` etc. sont definies dans `store/` (cote implementeur). L'hexagonal prescrit de les definir cote consommateur (`domain/`). | `backend/internal/store/cours_repo.go` | Deplacer les interfaces |
| m4 | Pas de wrapping d'erreurs systematique | Les erreurs ne sont pas wrappees avec contexte (`fmt.Errorf("fn: %w", err)`). Erreurs sentinelles definies comme pointeurs mutables. | `backend/internal/services/`, `backend/internal/store/` | Ajouter wrapping |
| m5 | Go 1.24 vs 1.23 spec | `go.mod` indique Go 1.24. Le CLAUDE.md specs dit Go 1.23. | `backend/go.mod` | Insignifiant, go 1.24 est un upgrade |
| m6 | lib/pq vs pgx/v5 | Le code utilise `lib/pq` + `database/sql`. Les specs prescrivent `pgx/v5`. | `backend/go.mod` | Migrer si necessaire |
| m7 | Pas de Redis | Les specs mentionnent Redis (go-redis/v9) pour le cache. Aucune utilisation dans le code. | N/A | Ajouter si lazy generation implementee |
| m8 | Pas de S3/Object storage | Les specs mentionnent S3. Le stockage actuel est un service fichier local (`ServiceStorage`). | `backend/internal/services/storage.go` | Ajouter adaptateur S3 |

### INFO

| # | Titre | Description |
|---|-------|-------------|
| I1 | Code DDD recoverable | Le code DDD complet est present dans l'historique git au commit `e543dea`. Il peut etre restaure avec `git checkout e543dea -- <paths>`. |
| I2 | Agent Ralph | Le commit `f047a63` ("auto-commit before merge (loop primary)") contient des fichiers `.ralph/` suggerant l'utilisation d'un agent autonome qui a produit l'architecture plate. |
| I3 | Dates de commit non-monotones | Le commit `e543dea` (DDD, dernier) date du 2026-03-28, tandis que les commits qui le remplacent (`ec69500` etc.) datent de 2026-02-11. Cela indique un rebase ou un merge non-chronologique. |
| I4 | Frontend fonctionnel | Le frontend TypeScript contient 13 pages et 20 composants, couvre scanner, dashboard, cours, fiches, quiz, mindmap, lexique, examen blanc, plans de revision, progression, analyse. C'est un produit utilisable. |
| I5 | Tests e2e Playwright | Des tests e2e existent dans `frontend/e2e/` pour valider le frontend, bien qu'ils ne couvrent pas les ACs Z1-Z8. |

---

## Statistiques

| Severite | Nombre |
|----------|--------|
| CRITIQUE | 8 |
| MAJEUR | 9 |
| MINEUR | 8 |
| INFO | 5 |
| **Total** | **30** |
