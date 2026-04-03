# Prompt 5 — Boucle circulaire d'amélioration continue (Specs ↔ Code ↔ Tests ↔ Feedback)

> **Objectif** : Formaliser un workflow circulaire récurrent où les spécifications nourrissent le code, le code nourrit les tests, les tests produisent des retours terrain, et les retours enrichissent les spécifications. Chaque cycle demande l'arbitrage du product owner avant de modifier les specs.

---

You are a systems-thinking product architect and continuous improvement strategist specialized in turning software delivery into an iterative learning loop.

## Mission

Help me formalize a recurring continuous-improvement workflow that links: specifications, produced code, execution results, end-to-end test outcomes, field feedback, and human product judgment.

## Contexte produit

Révise Mieux est un SaaS éducatif (Go + React Native + PostgreSQL) construit en architecture DDD hexagonale. Le projet a un corpus de specs riche :

| Document | Chemin | Rôle |
|----------|--------|------|
| PRD | `docs/PRD.md` | Source de vérité produit |
| ACs | `docs/ac/Z1.md` à `Z8.md` | 171 critères d'acceptation Given/When/Then |
| Scope | `docs/MVP-scope.md` | Classification 53 ACs Lot 0 |
| Tracker | `docs/lot0-tracker.md` | Suivi `[x]/[~]/[ ]` |
| User journeys | `docs/user-journeys/` | Parcours utilisateurs |
| LLM strategy | `docs/llm-strategy.md` | Pipeline 2 étages OCR/IDP, choix modèles post-benchmark (§8-9), monitoring (§6.3) |
| Benchmark LLM | `backend/testdata/benchmark/README.md` | Résultats 21 modèles, combos recommandés, exécution locale |
| API spec | `docs/openapi.yaml` | Endpoints |
| Conventions | `CLAUDE.md` | Architecture, DDD, testing |

### Pourquoi une approche circulaire

L'approche top-down (PRD → code) est nécessaire mais insuffisante. En pratique, une fois le code produit et les tests exécutés, de **nouvelles informations** apparaissent :
- incohérences entre spec et réalité
- faiblesses du PRD original
- edge cases imprévus
- opportunités d'amélioration
- features potentielles qui émergent du terrain
- endroits où l'intention produit devrait évoluer

Je veux que le système **apprenne de l'implémentation et de l'exécution**, pas qu'il se contente de valider contre des documents statiques.

### Modèle visé

```
    ┌─────────────────────────────────────────────┐
    │                                             │
    ▼                                             │
┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
│  SPECS   │───►│   CODE   │───►│  TESTS   │───►│ RETOURS  │
│          │    │          │    │          │    │ TERRAIN  │
│ PRD      │    │ Backend  │    │ Unitaire │    │          │
│ ACs      │    │ Mobile   │    │ Intég.   │    │ Incoher. │
│ Flows    │    │ Infra    │    │ E2E      │    │ Opportun.│
│          │    │          │    │          │    │ Features │
└──────────┘    └──────────┘    └──────────┘    └──────────┘
    ▲                                             │
    │         MON ARBITRAGE                       │
    │         (validation humaine)                │
    └─────────────────────────────────────────────┘
```

---

## Tes tâches

### 1. Recadrer le modèle

Explique les limites d'un review purement top-down. Puis explique pourquoi un modèle circulaire est plus approprié pour un produit en construction. Sois concret avec des exemples tirés du contexte Révise Mieux (pas de théorie agile générique).

### 2. Définir la boucle complète

Conçois le cycle opérationnel end-to-end. Pas un schéma générique — un cycle précis et exécutable :

| Étape | Action | Inputs | Outputs |
|-------|--------|--------|---------|
| 1 | Lire les spécifications | `docs/PRD.md`, `docs/ac/Z*.md`, `docs/lot0-tracker.md` | Baseline de référence |
| 2 | Inspecter l'implémentation | `backend/internal/`, `mobile/`, `migrations/` | État du code |
| 3 | Vérifier la cohérence specs↔code | Baseline + état code | Findings d'audit (prompt 04) |
| 4 | Exécuter les tests | `go test`, `jest`, E2E si dispo | Résultats tests |
| 5 | Extraire les retours du terrain | Résultats tests + audit | Feedback classifié |
| 6 | Challenger les specs | Feedback + specs originales | Propositions de changement |
| 7 | **Demander mon arbitrage** | Propositions | Décisions validées |
| 8 | Mettre à jour les specs | Décisions | Specs amendées |
| 9 | Relancer un tour | — | Nouveau cycle |

### 3. Distinguer les types de feedback

Le workflow doit classifier les findings en :

| Type | Description | Exemple Révise Mieux |
|------|-------------|---------------------|
| **Incohérence spec/code** | Le code contredit la spec | "Le seuil de transition est 0.7 dans le PRD mais 0.8 dans le code" |
| **Drift modèle LLM** | Le code utilise un modèle non recommandé | "Le pipeline utilise encore Claude Sonnet pour l'IDP alors que le benchmark recommande mistral-small (30x moins cher, 0 hallucination vs 8%)" |
| **Faiblesse de spec** | La spec est vague ou contradictoire | "Z1-AC03 dit '24h minimum' mais ne précise pas le timezone" |
| **Bug d'implémentation** | Le code est incorrect | "La régression SOLID→OK ne reset pas consecutive_successes" |
| **Problème UX révélé** | Le comportement réel est frustrant | "L'élève ne comprend pas pourquoi il reste OK malgré une réussite" |
| **Couverture de test insuffisante** | AC sans test ou test qui ne teste rien | "Z1-AC07b n'a pas de test dédié" |
| **Opportunité d'amélioration** | Pas un bug, mais pourrait être mieux | "Le feedback sur échec pourrait être plus nuancé par type d'item" |
| **Feature potentielle** | Idée qui émerge du terrain | "Un mode 'révision flash 5 min' pour les matins pressés" |
| **Décision produit non résolue** | Nécessite un choix humain | "Doit-on permettre de réviser un chapitre pas encore validé HITL ?" |

### 4. Inclure la phase "demande mon avis"

**C'est critique.** Le système ne doit PAS modifier silencieusement les specs. Il doit explicitement me remonter :
- ce qu'il a trouvé
- ce qui est faux ou flou
- ce qui semble améliorable
- les nouvelles idées qui émergent

Puis **attendre mon arbitrage** avant toute modification de document de référence.

Format de la demande d'arbitrage :

```markdown
## Arbitrage requis — [N] points

### Point 1 : [Titre court]
- **Type** : incohérence / faiblesse spec / opportunité / feature
- **Trouvé dans** : [fichier:ligne]
- **Spec dit** : [citation]
- **Réalité** : [ce qu'on observe]
- **Ma recommandation** : [ce que je propose]
- **Impact si on ne fait rien** : [conséquence]
- **Ta décision** : [ ] Accepter / [ ] Modifier / [ ] Rejeter / [ ] Reporter

### Point 2 : ...
```

### 5. Définir les artifacts persistants

Chaque boucle produit ou met à jour :

| Artifact | Chemin | Contenu |
|----------|--------|---------|
| **Rapport de boucle** | `reports/feedback-loop-YYYY-MM-DD-HHmm.md` | Résultat complet du cycle (créer `reports/` si absent) |
| **Registre d'issues** | `reports/issue-register.md` | Findings cumulés cross-cycles |
| **Log de décisions** | `reports/decision-log.md` | Décisions prises à chaque arbitrage |
| **Backlog d'amélioration** | `reports/improvement-backlog.md` | Opportunités et features en attente |
| **Candidats de mise à jour** | Diffs proposés dans le rapport | Modifications de PRD/ACs à valider |

### 6. Connecter la boucle à l'evidence d'exécution

La boucle doit être ancrée dans des preuves, pas dans des impressions :

| Source d'evidence | Commande | Ce qu'on en tire |
|-------------------|----------|-----------------|
| Tests unitaires Go | `cd backend && go test ./... -v -count=1` | ACs validées, bugs |
| Tests intégration | `cd backend && go test -tags=integration ./...` | Comportement réel DB |
| Tests mobile | `cd mobile && npx jest --verbose` | Couverture UI |
| Compilation | `cd backend && go build ./...` | Sanity check |
| Lint | `cd backend && golangci-lint run` | Qualité code |
| Coverage | `cd backend && go test -coverprofile=coverage.out ./...` | Taux couverture |

**Note** : La commande `golangci-lint` est optionnelle (peut ne pas être installée localement). Le slash command doit la tenter mais ne pas bloquer si absente.

### 7. Définir les modes de déclenchement

| Fréquence | Quand | Scope |
|-----------|-------|-------|
| **À la demande (1-2x/jour)** | Slash command `/feedback-loop` | Boucle complète |
| **Après chaque PR** | Hook post-merge ou manual | Mini-audit sur les ACs touchées |
| **Hebdomadaire** | Planifié | Boucle complète + révision priorités |
| **Après un jalon** | Manuel | Boucle complète + challenge du scope |
| **Après tests E2E échoués** | Automatique | Focus sur les failures |

### 8. Proposer une décomposition en sous-agents

Pour une exécution efficace dans Claude Code, propose une décomposition :

| Rôle | Responsabilité | Inputs | Output |
|------|---------------|--------|--------|
| **Spec Reader** | Lit et indexe tous les docs de spec | `docs/` | Baseline structurée |
| **Code Reviewer** | Vérifie cohérence code vs baseline | `backend/`, `mobile/` | Findings code |
| **Test Runner** | Exécute tests, analyse résultats | Commandes shell | Résultats + coverage |
| **Improvement Synthesizer** | Fusionne tous les findings, classifie | Outputs des 3 agents | Rapport unifié |
| **Decision Preparer** | Formule les points d'arbitrage | Rapport | Questions pour le PO |

Les agents Spec Reader, Code Reviewer, et Test Runner peuvent tourner **en parallèle**. Les deux derniers sont séquentiels.

### 9. Rendre le tout pratique et exécutable

**Ce prompt ne doit PAS rester un document théorique.**

Le livrable final est la **création d'un mécanisme exécutable** dans Claude Code :

#### Livrable 1 : Custom slash command

Un fichier `.claude/commands/feedback-loop.md` qui :
- orchestre les 5 phases séquentiellement
- utilise des sous-agents en parallèle où c'est possible
- écrit le rapport dans `reports/feedback-loop-YYYY-MM-DD-HHmm.md`
- s'arrête en phase d'arbitrage pour demander validation humaine
- est autonome : pas besoin de paramètres, il lit le contexte du projet

#### Livrable 2 : Templates de rapport

Les fichiers templates dans `reports/templates/` :
- `feedback-loop-template.md`
- `issue-register-template.md`
- `decision-log-template.md`

#### Livrable 3 (optionnel) : Agent schedulé

Configuration pour automatiser à heures fixes si souhaité :
```bash
# Lancement manuel
> /feedback-loop

# Ou planifié
> claude schedule "0 9,17 * * 1-5" /feedback-loop
```

### 10. MVP loop vs mature loop

| Aspect | MVP loop (maintenant) | Mature loop (plus tard) |
|--------|----------------------|------------------------|
| **Déclenchement** | Manuel (`/feedback-loop`) | Automatique post-PR + schedulé |
| **Tests** | `go test` unitaires seulement | Unitaires + intég + E2E |
| **Analyse** | 1 agent séquentiel | 4 sous-agents parallèles |
| **Rapport** | Markdown dans `reports/` | Markdown + GitHub issue auto-créée |
| **Arbitrage** | Dans la conversation Claude | PR avec review request |
| **Historique** | Fichiers locaux | Decision log versionné + trends |

---

## Format de sortie attendu

1. **Executive summary**
2. **Pourquoi le top-down seul est insuffisant** (avec exemples Révise Mieux)
3. **Modèle circulaire recommandé** (diagramme)
4. **Cycle opérationnel step-by-step** (tableau détaillé)
5. **Taxonomie des feedback** (tableau avec exemples concrets)
6. **Workflow de décision / arbitrage** (format de la demande)
7. **Artifacts persistants** (chemins, contenus, ownership)
8. **Décomposition en sous-agents** (tableau)
9. **Stratégie de déclenchement** (modes et fréquences)
10. **MVP loop vs mature loop** (tableau comparatif)
11. **Slash command `.claude/commands/feedback-loop.md`** (contenu complet, prêt à utiliser)
12. **Templates de rapport** (contenu complet)
13. **Risques et garde-fous**
14. **Recommandation opérationnelle finale**

## Barre de qualité

Ta réponse doit être :
- **systémique** — pas un checklist mais une boucle qui apprend
- **opérationnel** — exécutable, pas théorique
- **basé sur l'evidence** — ancré dans les tests et le code réel
- **centré sur le workflow récurrent** — conçu pour être relancé régulièrement
- **explicite sur le rôle humain** — le PO décide, le système propose

Ne réponds pas comme un coach agile générique. Pense comme l'architecte d'une boucle d'apprentissage produit/engineering qui réconcilie en continu les specs, l'implémentation, l'exécution et l'évolution stratégique.
