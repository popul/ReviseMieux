# Boucle d'amelioration continue — Specs / Code / Tests / Feedback

> Document de conception du workflow circulaire recurrent pour Revise Mieux.
> Derniere mise a jour : 2026-04-03.

---

## 1. Executive summary

Ce document formalise une boucle d'amelioration continue qui relie specifications, code, tests et retours terrain. Le modele est circulaire : chaque cycle produit des findings qui enrichissent les specs, mais uniquement apres arbitrage humain du product owner.

Le mecanisme est operationnel via la slash command `/feedback-loop` dans Claude Code, avec des rapports persistes dans `reports/`.

---

## 2. Pourquoi le top-down seul est insuffisant

L'approche PRD -> ACs -> Code -> Tests est necessaire pour la direction initiale, mais elle traite les specs comme des documents statiques. En pratique, l'implementation revele des informations que la specification ne pouvait pas anticiper.

### Exemples concrets dans Revise Mieux

| Situation | Ce que le top-down rate |
|-----------|------------------------|
| **Seuil de transition Mastery** | Le PRD dit 0.7, mais les tests montrent qu'un seuil different est utilise dans le code. Sans boucle, cette incoherence persiste indefiniment. |
| **Modele LLM obsolete** | La strategie LLM originale recommandait Sonnet pour la structuration, mais le benchmark a montre qu'il hallucine a 8%. Le code utilise-t-il deja le modele recommande post-benchmark (mistral-small) ou traîne-t-il encore sur Sonnet ? |
| **Espacement 24h et timezone** | Z1-AC03 dit "24h minimum" mais ne precise pas le timezone. Le code doit faire un choix — lequel a-t-il fait ? La spec doit-elle etre amendee ? |
| **Edge cases session** | Le PRD definit la composition 70/20/10 mais ne precise pas le comportement quand il y a moins de 10 items disponibles. L'implementation revele ce trou. |
| **Pipeline OCR en local** | Le benchmark montre que l'execution locale (Qwen3-VL-32B + Gemma 3 27B) est viable pour Lot 0. Cette opportunite n'existait pas au moment de la redaction du PRD. |

Le top-down produit une specification figee. La boucle la transforme en document vivant qui apprend de l'implementation.

---

## 3. Modele circulaire recommande

```
    +---------------------------------------------+
    |                                             |
    v                                             |
+----------+    +----------+    +----------+    +----------+
|  SPECS   |--->|   CODE   |--->|  TESTS   |--->| RETOURS  |
|          |    |          |    |          |    | TERRAIN  |
| PRD.md   |    | backend/ |    | go test  |    |          |
| Z*.md    |    | mobile/  |    | jest     |    | Incoher. |
| tracker  |    | migrat.  |    | E2E      |    | Opportun.|
| llm-strat|    |          |    | coverage |    | Features |
+----------+    +----------+    +----------+    +----------+
    ^                                             |
    |         ARBITRAGE HUMAIN                    |
    |         (validation PO)                     |
    +---------------------------------------------+
```

### Principes

1. **Evidence-based** : chaque finding est ancre dans du code, un test, ou une execution reelle.
2. **Human-in-the-loop** : le systeme propose, le PO decide. Aucune spec n'est modifiee automatiquement.
3. **Incremental** : chaque cycle produit un delta, pas une revision complete.
4. **Tracable** : chaque decision est loguee avec son contexte et sa justification.

---

## 4. Cycle operationnel step-by-step

| Etape | Action | Inputs | Outputs | Executant |
|-------|--------|--------|---------|-----------|
| 1 | Lire les specs | `docs/PRD.md`, `docs/ac/Z*.md`, `docs/lot0-tracker.md`, `docs/MVP-scope.md`, `docs/llm-strategy.md` | Baseline de reference | Spec Reader |
| 2 | Inspecter le code | `backend/internal/`, `mobile/`, `migrations/` | Etat du code | Code Reviewer |
| 3 | Executer les tests | `go test ./...`, `npx jest`, `go build` | Resultats + coverage | Test Runner |
| 4 | Croiser specs vs code vs tests | Baseline + code + resultats | Findings classifies | Improvement Synthesizer |
| 5 | Challenger les specs | Findings + specs originales | Propositions de changement | Decision Preparer |
| 6 | **Demander arbitrage PO** | Propositions formatees | -- | STOP (humain) |
| 7 | Appliquer les decisions | Decisions validees | Specs amendees, issues creees | Agent |
| 8 | Mettre a jour le tracker | Decisions + etat | `lot0-tracker.md` mis a jour | Agent |

**Point d'arret critique** : l'etape 6 est bloquante. Le systeme ne passe jamais a l'etape 7 sans validation humaine explicite.

---

## 5. Taxonomie des feedback

| Type | Description | Exemple Revise Mieux | Severite |
|------|-------------|---------------------|----------|
| **Incoherence spec/code** | Le code contredit la spec | "Le seuil de transition est 0.7 dans le PRD mais 0.8 dans le code" | Haute |
| **Drift modele LLM** | Le code utilise un modele non recommande post-benchmark | "Le pipeline utilise encore Sonnet pour l'IDP alors que le benchmark recommande mistral-small" | Haute |
| **Faiblesse de spec** | La spec est vague ou contradictoire | "Z1-AC03 dit '24h minimum' mais ne precise pas le timezone" | Moyenne |
| **Bug d'implementation** | Le code est incorrect | "La regression SOLID->OK ne reset pas consecutive_successes" | Haute |
| **Probleme UX revele** | Le comportement reel est frustrant | "L'eleve ne comprend pas pourquoi il reste OK malgre une reussite" | Moyenne |
| **Couverture test insuffisante** | AC sans test ou test vide | "Z1-AC07b n'a pas de test dedie" | Moyenne |
| **Opportunite d'amelioration** | Pas un bug, mais pourrait etre mieux | "Le feedback sur echec pourrait etre plus nuance par type d'item" | Basse |
| **Feature potentielle** | Idee qui emerge du terrain | "Un mode 'revision flash 5 min' pour les matins presses" | Basse |
| **Decision produit non resolue** | Necessite un choix humain | "Doit-on permettre de reviser un chapitre pas encore valide HITL ?" | Haute |

---

## 6. Workflow de decision / arbitrage

Quand la boucle a des findings a remonter, elle les presente dans ce format :

```markdown
## Arbitrage requis -- [N] points

### Point 1 : [Titre court]
- **Type** : incoherence / faiblesse spec / opportunite / feature / drift LLM
- **Trouve dans** : [fichier:ligne]
- **Spec dit** : [citation]
- **Realite** : [ce qu'on observe]
- **Ma recommandation** : [ce que je propose]
- **Impact si on ne fait rien** : [consequence]
- **Ta decision** : [ ] Accepter / [ ] Modifier / [ ] Rejeter / [ ] Reporter
```

### Regles d'arbitrage

1. Les findings de severite **haute** sont presentes en priorite.
2. Les findings de severite **basse** sont groupes en fin de liste.
3. Le PO peut accepter, modifier, rejeter, ou reporter chaque point individuellement.
4. Toute decision est loguee dans `reports/decision-log.md`.

---

## 7. Artifacts persistants

| Artifact | Chemin | Contenu | Cycle de vie |
|----------|--------|---------|-------------|
| **Rapport de boucle** | `reports/feedback-loop-YYYY-MM-DD-HHmm.md` | Resultat complet d'un cycle | Un par execution |
| **Registre d'issues** | `reports/issue-register.md` | Findings cumules cross-cycles | Append-only, mis a jour a chaque cycle |
| **Log de decisions** | `reports/decision-log.md` | Decisions prises a chaque arbitrage | Append-only |
| **Backlog d'amelioration** | `reports/improvement-backlog.md` | Opportunites et features en attente | Mis a jour apres arbitrage |
| **Templates** | `reports/templates/` | Structures vides pour chaque artifact | Statiques |

---

## 8. Decomposition en sous-agents

| Role | Responsabilite | Inputs | Output | Parallelisable |
|------|---------------|--------|--------|---------------|
| **Spec Reader** | Lit et indexe tous les docs de spec | `docs/` | Baseline structuree (ACs, etats, regles) | Oui |
| **Code Reviewer** | Verifie coherence code vs baseline | `backend/internal/`, `mobile/`, `migrations/` | Findings code | Oui |
| **Test Runner** | Execute tests, analyse resultats | Commandes shell | Resultats + coverage | Oui |
| **Improvement Synthesizer** | Fusionne findings, classifie, deduplique | Outputs des 3 agents | Rapport unifie | Non (sequentiel apres les 3) |
| **Decision Preparer** | Formule les points d'arbitrage | Rapport unifie | Questions formatees pour le PO | Non (sequentiel) |

Les 3 premiers agents tournent en parallele. Les 2 derniers sont sequentiels.

### MVP loop simplifiee

Pour le MVP, un seul agent execute les 5 roles sequentiellement. La parallelisation viendra en mature loop.

---

## 9. Strategie de declenchement

| Frequence | Quand | Scope | Commande |
|-----------|-------|-------|----------|
| **A la demande (1-2x/jour)** | Slash command | Boucle complete | `/feedback-loop` |
| **Apres chaque PR** | Manuel post-merge | Mini-audit sur les ACs touchees | `/feedback-loop` |
| **Hebdomadaire** | Lundi matin | Boucle complete + revision priorites | `/feedback-loop` |
| **Apres un jalon** | Manuel | Boucle complete + challenge du scope | `/feedback-loop` |

---

## 10. MVP loop vs mature loop

| Aspect | MVP loop (maintenant) | Mature loop (plus tard) |
|--------|----------------------|------------------------|
| **Declenchement** | Manuel (`/feedback-loop`) | Automatique post-PR + schedule |
| **Tests** | `go test` unitaires + `go build` | Unitaires + integration + E2E |
| **Analyse** | 1 agent sequentiel | 5 sous-agents (3 paralleles + 2 sequentiels) |
| **Rapport** | Markdown dans `reports/` | Markdown + GitHub issue auto-creee |
| **Arbitrage** | Dans la conversation Claude | PR avec review request |
| **Historique** | Fichiers locaux | Decision log versionne + dashboards |
| **Coverage** | Coverage basique si disponible | Coverage tracking + trends |

---

## 11. Sources d'evidence

| Source | Commande | Ce qu'on en tire |
|--------|----------|-----------------|
| Tests unitaires Go | `cd backend && go test ./... -v -count=1` | ACs validees, bugs, regressions |
| Tests integration | `cd backend && go test -tags=integration ./...` | Comportement reel DB |
| Tests mobile | `cd mobile && npx jest --verbose` | Couverture UI |
| Compilation | `cd backend && go build ./...` | Sante du code |
| Lint | `cd backend && golangci-lint run` | Qualite code (optionnel) |
| Coverage | `cd backend && go test -coverprofile=coverage.out ./...` | Taux couverture |

---

## 12. Risques et garde-fous

| Risque | Garde-fou |
|--------|-----------|
| Modification silencieuse des specs | Arret obligatoire a l'etape 6, pas de modification sans validation |
| Bruit : trop de findings non pertinents | Classification par severite + filtre sur les findings nouveaux uniquement |
| Fatigue d'arbitrage : le PO ignore les demandes | Limiter a 5-7 points par cycle, prioriser |
| Cycle infini : un finding revient sans arret | Le decision log trace les rejets, un finding rejete 2 fois est archive |
| False positives dans l'audit | Chaque finding doit citer le fichier et la ligne concrets |
| Perte de contexte entre cycles | Le registre d'issues cumule, le decision log historise |

---

## 13. Recommandation operationnelle finale

### Cadence recommandee pour Lot 0

1. **Quotidien** : lancer `/feedback-loop` en fin de journee de dev (1x/jour).
2. **Apres chaque PR mergee** : lancer un cycle rapide pour verifier la coherence.
3. **Hebdomadaire (lundi)** : cycle complet avec revision des priorites et du tracker.

### Actions immediates

1. Le slash command `/feedback-loop` est pret a l'emploi dans `.claude/commands/feedback-loop.md`.
2. Les templates de rapport sont dans `reports/templates/`.
3. Le premier cycle devrait etre lance immediatement apres merge de cette PR pour etablir la baseline.

### Evolution

- Phase 1 (maintenant) : MVP loop manuelle, 1 agent sequentiel
- Phase 2 (Lot 0 termine) : parallelisation des sous-agents
- Phase 3 (post-MVP) : automatisation post-PR via GitHub Actions
