# Rapport de boucle feedback-loop

| | |
|---|---|
| **Date** | YYYY-MM-DD HH:mm |
| **Cycle** | #N |
| **Branche** | `branch-name` |
| **Declencheur** | Manuel / Post-PR / Hebdomadaire |

---

## 1. Resume executif

<!-- 3-5 lignes : qu'est-ce qui a ete trouve, quel est l'etat general -->

---

## 2. Metriques cles

| Metrique | Valeur |
|----------|--------|
| Tests Go passes | X / Y |
| Tests Go echoues | X |
| Coverage Go | X% |
| Tests mobile passes | X / Y |
| Compilation Go | OK / KO |
| Lint Go | OK / KO / N/A |
| ACs Lot 0 couvertes (code existe) | X / 53 |
| ACs Lot 0 testees (test existe) | X / 53 |

---

## 3. Findings classifies

| # | ID | Type | Severite | AC | Fichier | Resume |
|---|-----|------|----------|-----|---------|--------|
| 1 | FL-YYYY-MM-DD-001 | INCOHERENCE | HAUTE | Z1-AC01 | `path/file.go:42` | Description courte |
| 2 | FL-YYYY-MM-DD-002 | COUVERTURE | MOYENNE | Z1-AC07b | -- | Description courte |

---

## 4. Points d'arbitrage

<!-- Maximum 7 points, tries par severite -->

### Point 1 : [Titre court]
- **Type** : [taxonomie]
- **Severite** : HAUTE
- **Trouve dans** : `fichier:ligne`
- **Spec dit** : "[citation]"
- **Realite** : [observation]
- **Ma recommandation** : [proposition]
- **Impact si on ne fait rien** : [consequence]
- **Decision PO** : [ ] Accepter / [ ] Modifier / [ ] Rejeter / [ ] Reporter

### Point 2 : [Titre court]
<!-- meme format -->

---

## 5. Findings basse severite (annexe)

<!-- Findings informatifs, pas de demande d'arbitrage -->

| # | Type | Resume | Action suggeree |
|---|------|--------|-----------------|
| 1 | OPPORTUNITE | ... | ... |
| 2 | FEATURE | ... | ... |

---

## 6. Prochaines actions recommandees

- [ ] Action 1
- [ ] Action 2
- [ ] Action 3

---

## 7. Historique des decisions (ce cycle)

<!-- Rempli apres arbitrage PO -->

| Point | Decision | Justification PO |
|-------|----------|-------------------|
| 1 | Accepter / Modifier / Rejeter / Reporter | ... |
