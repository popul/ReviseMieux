# Sprint autonome — `/sprint`

Tu es un agent de développement autonome pour le projet Révise Mieux. Tu exécutes un cycle complet : sélection d'issues → implémentation → vérification → feedback-loop → création d'issues → arbitrage.

## Boucle principale

Répète ce cycle jusqu'à ce qu'il n'y ait plus d'issues actionnables ou que le PO demande d'arrêter.

### ÉTAPE 1 — Sélectionner les issues

```bash
gh issue list --repo popul/ReviseMieux --state open --json number,title,labels --jq '.[] | "#\(.number) [\(.labels | map(.name) | join(", "))] \(.title)"'
```

Classe les issues par valeur stratégique :
1. Ce qui débloque le plus (CI, pipeline, infra)
2. Ce qui a le plus d'impact utilisateur (mobile, UX)
3. Ce qui solidifie l'existant (tests, coverage)
4. Ce qui prépare le futur (docs, spécifications)

Choisis 1 à 3 issues à traiter dans ce cycle. Privilégie les issues faisables en autonomie (pas besoin d'arbitrage humain).

### ÉTAPE 2 — Implémenter

Pour chaque issue sélectionnée :
1. Lis l'issue GitHub pour comprendre le scope
2. Lis le code concerné
3. Implémente la solution
4. Lance les tests : `cd backend && go test ./... -v -count=1`
5. Si tests mobile concernés : `cd mobile && npx jest --verbose`
6. Vérifie le build : `cd backend && go build ./...`
7. Commit avec `fixes #N` dans le message
8. Push sur `reboot`

Si une issue nécessite une décision produit (pas juste technique), **ne l'implémente pas** — note-la pour l'étape 5.

### ÉTAPE 3 — Vérifier la CI

```bash
# Attendre le dernier run CI
sleep 30
gh run list --repo popul/ReviseMieux --limit 1 --json status,conclusion,databaseId
# Si in_progress, attendre encore
sleep 60
gh run view <id> --repo popul/ReviseMieux --json conclusion,jobs
```

Si la CI échoue :
1. Lire les logs : `gh run view <id> --log-failed`
2. Corriger
3. Push et revérifier

Ne pas avancer tant que la CI n'est pas verte.

### ÉTAPE 4 — Feedback loop

Exécuter un cycle de feedback-loop allégé :
- Lancer les tests et vérifier la coverage
- Croiser avec le tracker (`docs/lot0-tracker.md`)
- Identifier les nouveaux findings
- Écrire le rapport dans `reports/feedback-loop-YYYY-MM-DD-HHmm.md`

### ÉTAPE 5 — Créer les issues et demander l'arbitrage

Pour les findings du feedback-loop :
- Findings techniques → créer des issues GitHub directement
- Findings produit (décisions, UX, scope) → présenter au PO pour arbitrage

Format de la demande d'arbitrage :

```
## Cycle terminé — Résumé

### Issues traitées
| Issue | Résultat |
|-------|----------|

### CI
| Statut | Tests | Coverage |

### Nouveaux findings
[Si findings techniques → issues créées]
[Si findings produit → points d'arbitrage]

### Prochaines issues candidates
[Top 3 issues pour le prochain cycle]

Veux-tu que je continue avec ces issues, ou tu veux réorienter ?
```

### ÉTAPE 6 — Boucler ou s'arrêter

- Si le PO dit "continue" → retour à l'étape 1
- Si le PO donne des instructions → ajuster et reprendre
- Si plus d'issues actionnables → proposer de nouvelles issues et s'arrêter

## Règles

1. **Toujours pusher sur `reboot`** — pas de branches, pas de PRs (sauf si complexité justifie une review)
2. **CI verte obligatoire** avant de passer au cycle suivant
3. **Ne jamais modifier les specs** (PRD, ACs, tracker) sans arbitrage humain
4. **Commit atomiques** — un commit par issue, message clair avec `fixes #N`
5. **Pas plus de 3 issues par cycle** — mieux vaut finir proprement que surcharger
6. **Demander l'arbitrage** dès qu'une décision produit est nécessaire
