# Boucle d'amelioration continue — `/feedback-loop`

Tu es un agent d'amelioration continue pour le projet Revise Mieux. Tu executes un cycle complet d'audit specs-vs-code-vs-tests et produis un rapport structure avec des points d'arbitrage pour le product owner.

## Contexte

Revise Mieux est un SaaS educatif construit en architecture DDD hexagonale (Go + React Native + PostgreSQL). Le projet possede un corpus de specs riche dans `docs/` et un tracker de progression dans `docs/lot0-tracker.md`.

## Instructions

Execute les 5 phases ci-dessous **sequentiellement**. Ne saute aucune phase.

---

### PHASE 1 — Lecture des specifications (Spec Reader)

Lis integralement les documents suivants pour construire la baseline de reference :

1. `docs/PRD.md` — source de verite produit
2. `docs/ac/Z1.md` a `docs/ac/Z8.md` — criteres d'acceptation Given/When/Then
3. `docs/MVP-scope.md` — classification des 53 ACs Lot 0
4. `docs/lot0-tracker.md` — etat d'avancement actuel
5. `docs/llm-strategy.md` — pipeline LLM et choix de modeles post-benchmark
6. `CLAUDE.md` — conventions architecture et DDD

Pour chaque AC du Lot 0 (P1 et P2), note :
- L'identifiant (ex: Z1-AC01)
- Le comportement attendu (Given/When/Then)
- Le statut dans le tracker (`[x]`, `[~]`, `[ ]`)
- La priorite (P1 ou P2)

**Output** : baseline structuree en memoire.

---

### PHASE 2 — Inspection du code et execution des tests (Code Reviewer + Test Runner)

#### 2A — Inspection du code

Inspecte le code source pour verifier la coherence avec la baseline :

- `backend/internal/domain/` — entites, value objects, transitions mastery
- `backend/internal/app/` — services applicatifs, orchestration
- `backend/internal/infra/` — adaptateurs (postgres, redis, s3, LLM, OCR)
- `backend/internal/http/` — handlers, DTOs, middleware
- `backend/migrations/` — schema SQL
- `mobile/` — composants React Native (si present)
- `frontend/` — composants frontend (si present)

Pour chaque AC marquee `[x]` dans le tracker, verifie :
1. Le code correspondant existe
2. Le comportement implemente correspond a la spec
3. Les tests couvrent l'AC

Pour chaque AC marquee `[ ]` ou `[~]`, note l'absence d'implementation.

#### 2B — Execution des tests

Execute les commandes suivantes et analyse les resultats :

```bash
# Compilation Go
cd backend && go build ./... 2>&1

# Tests unitaires Go (avec verbose pour voir les noms de tests)
cd backend && go test ./... -v -count=1 2>&1

# Coverage Go (si les tests passent)
cd backend && go test -coverprofile=/tmp/coverage.out ./... 2>&1 && go tool cover -func=/tmp/coverage.out 2>&1

# Lint Go (optionnel -- ne pas bloquer si golangci-lint n'est pas installe)
cd backend && golangci-lint run 2>&1 || echo "golangci-lint non disponible, skip"

# Tests mobile/frontend (si package.json existe)
if [ -f mobile/package.json ]; then cd mobile && npx jest --verbose 2>&1; fi
if [ -f frontend/package.json ]; then cd frontend && npm test 2>&1 || echo "Tests frontend non configures"; fi
```

**Output** : resultats de compilation, tests (passes/echoues), coverage, lint.

---

### PHASE 3 — Synthese et classification des findings (Improvement Synthesizer)

Croise la baseline (Phase 1) avec l'etat du code et des tests (Phase 2). Produis une liste de findings classifies selon cette taxonomie :

| Type | Description |
|------|-------------|
| **INCOHERENCE** | Le code contredit la spec |
| **DRIFT_LLM** | Le code utilise un modele LLM non recommande post-benchmark |
| **FAIBLESSE_SPEC** | La spec est vague, contradictoire ou incomplete |
| **BUG** | Le code est incorrect (test echoue, comportement errone) |
| **UX** | Le comportement reel est frustrant ou peu clair |
| **COUVERTURE** | AC sans test ou test qui ne valide rien |
| **OPPORTUNITE** | Pas un bug, mais ameliorable |
| **FEATURE** | Idee nouvelle qui emerge de l'analyse |
| **DECISION** | Necessite un choix humain, pas une correction technique |

Pour chaque finding, documente :
- Identifiant unique (ex: FL-2026-04-03-001)
- Type (selon la taxonomie)
- Severite : HAUTE / MOYENNE / BASSE
- Fichier et ligne concernes
- Ce que la spec dit (citation)
- Ce que le code fait (observation)
- Recommandation
- Impact si on ne fait rien

---

### PHASE 4 — Challenge des specs (Decision Preparer)

Pour chaque finding de severite HAUTE ou MOYENNE, formule un point d'arbitrage :

```markdown
### Point [N] : [Titre court]
- **Type** : [taxonomie]
- **Severite** : [HAUTE/MOYENNE]
- **Trouve dans** : [fichier:ligne]
- **Spec dit** : [citation exacte]
- **Realite** : [observation dans le code/tests]
- **Ma recommandation** : [ce que je propose]
- **Impact si on ne fait rien** : [consequence concrete]
- **Ta decision** : [ ] Accepter / [ ] Modifier / [ ] Rejeter / [ ] Reporter
```

Les findings de severite BASSE sont listes en annexe sans demande d'arbitrage.

---

### PHASE 5 — Rapport et arret pour arbitrage

#### 5A — Ecrire le rapport

Genere le rapport complet dans `reports/feedback-loop-YYYY-MM-DD-HHmm.md` (utilise la date et l'heure actuelles).

Utilise le template `reports/templates/feedback-loop-template.md` comme structure.

Le rapport contient :
1. Resume executif (3-5 lignes)
2. Metriques cles (tests passes/echoues, coverage, ACs couvertes)
3. Findings classifies (tableau)
4. Points d'arbitrage detailles
5. Findings basse severite (annexe)
6. Prochaines actions recommandees

#### 5B — Mettre a jour le registre d'issues

Si le fichier `reports/issue-register.md` existe, ajoute les nouveaux findings.
Sinon, cree-le a partir du template `reports/templates/issue-register-template.md`.

#### 5C — ARRET : demander l'arbitrage

**STOP ICI.** Ne modifie aucun document de specification (`docs/PRD.md`, `docs/ac/Z*.md`, `docs/lot0-tracker.md`, `docs/llm-strategy.md`, `CLAUDE.md`).

Presente les points d'arbitrage au product owner dans la conversation et attends sa reponse.

Format de la demande :

```
## Arbitrage requis -- [N] points

[Liste des points d'arbitrage de la Phase 4]

Reponds pour chaque point : Accepter / Modifier / Rejeter / Reporter.
Je n'appliquerai aucune modification de spec sans ta validation.
```

---

### PHASE 6 (conditionnelle) — Application des decisions

**Uniquement apres reponse du PO.** Pour chaque decision :

- **Accepter** : appliquer la modification proposee dans le document concerne
- **Modifier** : appliquer la version modifiee par le PO
- **Rejeter** : loguer le rejet dans `reports/decision-log.md`
- **Reporter** : ajouter au backlog `reports/improvement-backlog.md`

Loguer toutes les decisions dans `reports/decision-log.md` (utiliser le template `reports/templates/decision-log-template.md`).

---

## Regles strictes

1. **JAMAIS modifier une spec sans arbitrage humain explicite.**
2. Chaque finding doit citer un fichier et une ligne concrets.
3. Ne pas inventer de problemes — se baser uniquement sur l'evidence (code, tests, execution).
4. Limiter les points d'arbitrage a 7 maximum par cycle (prioriser par severite).
5. Le rapport est ecrit dans `reports/` et versionne dans git.
6. Si les tests ne compilent pas ou ne s'executent pas, le signaler comme finding de severite HAUTE.
