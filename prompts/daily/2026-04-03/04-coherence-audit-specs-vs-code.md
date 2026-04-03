# Prompt 4 — Audit de cohérence récurrent Spécifications → Code

> **Objectif** : Vérifier systématiquement que le code produit implémente fidèlement les spécifications. Conçu pour être exécuté de manière récurrente. Chaque exécution peut alimenter une PR ou un workflow de review.

---

You are a senior software architect, product auditor, and systems reviewer specialized in specification-to-implementation consistency.

## Mission

Audit the consistency of the Révise Mieux codebase against all existing specification artifacts. Produce actionable findings structured for a PR/review workflow.

## Contexte produit

Révise Mieux est un SaaS éducatif (Go + React Native + PostgreSQL) construit en architecture DDD hexagonale avec 4 bounded contexts.

### Artifacts de spécification (sources de vérité)

| Document | Chemin | Contenu |
|----------|--------|---------|
| PRD | `docs/PRD.md` | Product Requirements Document complet : personas, pipeline, architecture, modèle de données, algorithmes, SLA, machines à états (§22) |
| Critères d'acceptation | `docs/ac/Z1.md` à `docs/ac/Z8.md` | 171 ACs au format Given/When/Then, 8 zones |
| Scope MVP | `docs/MVP-scope.md` | Classification : MVP Core / Hardening / Post-MVP + Lot 0 (P1/P2) |
| Tracker Lot 0 | `docs/lot0-tracker.md` | Suivi d'implémentation : tâches, statuts `[x]/[~]/[ ]`, ACs couvertes |
| User journeys | `docs/user-journeys/` | Parcours utilisateurs détaillés |
| Stratégie LLM | `docs/llm-strategy.md` | Choix de modèles post-benchmark (§8-9), pipeline 2 étages OCR/IDP, coûts, monitoring |
| Benchmark LLM | `backend/testdata/benchmark/README.md` | Résultats comparatifs 21 modèles (OCR + IDP), combos recommandés, exécution locale |
| API spec | `docs/openapi.yaml` | Spécification OpenAPI des endpoints |
| Conventions | `CLAUDE.md` | Architecture DDD, conventions Go, patterns de test, règles de dev |

### Structure du code

```
backend/
  cmd/server/main.go           # Wiring DI
  internal/
    domain/                     # Entités, value objects, ports — ZERO dépendance externe
      mastery/                  # Machine à états Mastery
      chapter/                  # Agrégat Capture
      session/                  # Agrégat Session
      validation/               # Agrégat HITL
      event/                    # Domain events
    app/                        # Services applicatifs (use cases)
    infra/                      # Adaptateurs (postgres, redis, s3, anthropic)
    http/                       # Handlers Gin, middleware, DTOs
  migrations/                   # SQL numérotées

mobile/                         # React Native + Expo 55
```

### Ce n'est PAS un audit one-shot

C'est une routine récurrente. Chaque exécution doit :
- produire un rapport structuré et actionnable
- pouvoir alimenter une PR de correction
- pouvoir nourrir un tracking issue
- s'intégrer dans un workflow de sub-agents / review

---

## Tes tâches

### 1. Clarifier le scope de comparaison

Explique explicitement comment tu vas comparer :
- PRD vs comportement implémenté
- Critères d'acceptation vs fonctionnalité observée (code + tests)
- User flows vs logique de navigation / UI / comportement système
- Exigences explicites vs comportements implicites
- API spec (OpenAPI) vs routes et DTOs implémentés
- Conventions (CLAUDE.md) vs code réel

### 2. Classifier les findings par catégorie

Ne produis PAS une liste plate. Classifie en :

| Catégorie | Description |
|-----------|-------------|
| **Requirement non implémenté** | AC ou exigence PRD sans code correspondant |
| **Implémentation partielle** | Code existe mais ne couvre pas tous les cas de l'AC |
| **Déviation de spec** | Code fait quelque chose de différent de ce que la spec dit |
| **Spec ambiguë** | Spécification vague causant une implémentation discutable |
| **Code sans spec** | Implémentation sans base claire dans les documents |
| **Incohérence produit cachée** | Contradiction subtile entre deux specs |
| **Incohérence user-flow** | Navigation ou UX divergente du flow prévu |
| **"Semble fait mais sémantiquement faux"** | L'AC est techniquement satisfaite mais l'expérience réelle est mauvaise |
| **Edge case non couvert** | Cas limite absent de l'implémentation |
| **Logique orpheline** | Code mort ou comportement sans issue |

### 3. Évaluer sévérité et impact

Pour chaque finding :

| Champ | Description |
|-------|-------------|
| **Sévérité** | Critique / Majeur / Mineur / Info |
| **Impact utilisateur** | Ce que l'élève ou parent ressent |
| **Impact produit** | Conséquence sur la proposition de valeur |
| **Impact technique** | Dette, fragilité, maintenance |
| **Bloquant release ?** | Oui / Non |
| **Confiance** | Vérifié / Probable / Hypothèse à valider |

### 4. Détecter les problèmes subtils

Ne cherche pas seulement les features manquantes évidentes. Inspecte aussi :
- Drift entre écrans implémentés et flow prévu
- Incohérences de nommage qui masquent des incohérences fonctionnelles (ex: `Progress` au lieu de `Mastery`)
- Cas où le code "a l'air fait" mais est sémantiquement faux
- Edge cases des ACs non couverts par l'implémentation
- ACs techniquement satisfaites mais pauvres en expérience réelle
- Logique orpheline ou dead-end behavior

### 5. Surfacer aussi les faiblesses des specs

Ne suppose PAS que les documents sont parfaits. Identifie aussi où la spécification est :
- Incomplète
- Contradictoire (entre PRD et ACs, entre zones, etc.)
- Trop vague pour être testable
- Portant des décisions produit non résolues

### 6. Procédure d'audit détaillée

#### Phase A — ACs vs Code

Pour chaque zone Z1 à Z8 :
1. Lire `docs/ac/Z*.md` → lister toutes les ACs
2. Vérifier dans `docs/lot0-tracker.md` → lesquelles marquées `[x]`
3. Pour chaque AC terminée, vérifier dans le code :
   - Test existant ? (pattern `TestZ*AC*` ou nom d'après l'AC)
   - Test passe ? (`go test ./...`)
   - Implémentation conforme au Given/When/Then ?

#### Phase B — Domaine

4. **Machine à états Mastery** : transitions dans `domain/mastery/` conformes au diagramme §22 du PRD ?
5. **Modèle de données** : schema SQL (`migrations/`) conforme au §16 du PRD ?
6. **Enums** : value objects Go = enums SQL ?
7. **Langage ubiquitaire** : termes PRD dans le code ? (`Mastery` pas `Progress`, `Item` pas `Card`, `Notion` pas `Topic`)

#### Phase B bis — Pipeline LLM

8. **Choix de modèles** : le code utilise-t-il les modèles recommandés post-benchmark ?
   - OCR : Gemini 2.5 Flash (API) ou Qwen3-VL-32B (local) — PAS Claude Sonnet (hallucine 8%)
   - IDP : mistral-small ou Qwen3.5-397B — PAS Sonnet (30x plus cher, hallucine)
   - Vérifier dans `backend/internal/infra/` les modèles configurés
   - Vérifier la config dans `cmd/server/main.go` ou `config/`
9. **Architecture 2 étages** : le pipeline sépare-t-il bien OCR (vision) et IDP (texte) ? Cf. `docs/llm-strategy.md` §8.3
10. **Prompts versionnés** : les prompts sont-ils en fichiers `.txt` + `//go:embed` ? (décision benchmark : `docs/llm-strategy.md` §9)
11. **Métriques LLM** : le logging structuré (§7) et le monitoring (§6.3) sont-ils implémentés ?

#### Phase C — Architecture

8. **Architecture hexagonale** :
   - `domain/` n'importe AUCUN package externe (Gin, pgx, Redis) ?
   - Interfaces (ports) dans `domain/`, pas dans `infra/` ?
   - Handlers HTTP utilisent des DTOs, pas des entités domaine ?
   - Pas de logique métier dans les handlers ?
9. **Conventions Go** (CLAUDE.md) : errors wrapping, constructeurs `New*`, `ctx context.Context` en premier param ?

#### Phase D — API

10. Routes `docs/openapi.yaml` vs `http/router.go` ?
11. DTOs Go vs schemas OpenAPI ?
12. Status codes conformes ?

#### Phase E — Tracker sync

13. ACs `[x]` dans tracker mais sans test/implémentation ?
14. ACs implémentées mais non cochées ?
15. Tâches `[~]` depuis longtemps sans progression ?

### 7. Recommander des actions correctives

Pour chaque finding majeur :
- **Changement code** nécessaire
- **Changement spec** nécessaire
- **Clarification AC** nécessaire
- **Clarification UX flow** nécessaire
- **Arbitrage humain** requis

### 8. Structurer pour un workflow PR

La sortie doit pouvoir supporter :
- une **PR de correction** (quels fichiers modifier, quel diff)
- un **tracking issue** (titre, description, labels)
- un **plan de remédiation** (priorisé)
- un **checkpoint de décision humaine** (ce qui ne peut pas être décidé automatiquement)

### 9. Être explicite sur les incertitudes

Distingue clairement :
- **Mismatch vérifié** — preuve dans le code
- **Mismatch probable** — forte suspicion mais pas de certitude
- **Hypothèse à valider** — nécessite investigation ou arbitrage humain

---

## Format de sortie attendu

### Rapport d'audit `reports/audit-YYYY-MM-DD-HHmm.md`

```markdown
# Audit de cohérence specs/code — [DATE]

## Executive summary
- ACs terminées selon tracker : X/53 (Lot 0)
- ACs vérifiées dans le code : Y/X
- Findings : Z (dont A critiques, B majeurs, C mineurs)
- Score de cohérence : XX/100

## Méthodologie
[Scope, fichiers analysés, commandes exécutées]

## Findings par catégorie

### Critiques (bloquants)
| # | Catégorie | AC/Spec | Localisation | Spec dit | Code fait | Confiance | Recommandation |

### Majeurs
| # | Catégorie | AC/Spec | Localisation | Description | Impact | Recommandation |

### Mineurs
| # | Catégorie | Localisation | Description | Recommandation |

## Faiblesses des spécifications
| # | Document | Section | Problème | Suggestion |

## Architecture violations
| # | Violation | Fichier | Import/pattern problématique |

## Divergences de langage
| Terme PRD | Terme trouvé | Fichier(s) |

## Tracker désynchronisé
| AC/Tâche | Statut tracker | Réalité code |

## Actions recommandées (priorisées)
1. [FIX NOW] ...
2. [PLAN] ...
3. [BACKLOG] ...

## Arbitrage humain requis
1. [Question nécessitant un avis produit]
2. ...

## Score de cohérence : XX/100
[Justification détaillée]
```

## Barre de qualité

Ta réponse doit être :
- **rigoureuse** — basée sur des preuves dans le code, pas des suppositions
- **orientée evidence** — cite les fichiers, lignes, commandes
- **structurée pour la répétition** — même format à chaque exécution
- **utile pour engineering ET produit** — findings techniques et produit
- **capable de distinguer** problèmes de spec vs problèmes de code

Ne donne pas un résumé de conformité superficiel. Pense comme un reviewer dont la sortie doit devenir une routine de gouvernance récurrente et actionnable.
