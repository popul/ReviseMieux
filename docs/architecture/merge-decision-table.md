# Tableau de décision -- Merge incrémental

> Référence pour le `MergeEngine` (domaine pur, `backend/internal/domain/chapter/merge.go`).
> Chaque règle est exécutée dans l'ordre de priorité. La première règle qui matche est appliquée.

---

## Métriques utilisées

| Métrique | Définition | Calcul |
|----------|-----------|--------|
| **CanonicalKey match** | Les deux items ont le même `CanonicalKey(term, item_type, pack_id)` | `identity.go` : lowercase + NFD accent removal + trim |
| **Jaccard(keywords)** | Coefficient de Jaccard entre les ensembles de keywords | `|A ∩ B| / |A ∪ B|` |
| **Confidence delta** | Différence de confiance entre le nouvel item et l'existant | `new.confidence - existing.confidence` |
| **Text similarity** | Similarité textuelle entre les `ocr_text` des blocs sources | Jaccard sur les mots normalisés (après stopwords removal) |

---

## Tableau de décision

| # | Condition | Action | Automatique ? | Domain Event | Impact Mastery | Référence AC |
|---|-----------|--------|---------------|--------------|----------------|-------------|
| **R1** | CanonicalKey match **ET** Jaccard(keywords) > 0.85 | **ARCHIVE_DUP** : archiver l'item de plus faible confiance | Oui | `ItemArchived` | Transfert Mastery vers survivant (le plus avancé) | Z3-AC11 |
| **R2** | CanonicalKey match **ET** Jaccard(keywords) 0.50-0.85 **ET** `new.confidence > existing.confidence + 0.15` | **REPLACE** : remplacer l'ancien par le nouveau (meilleure extraction) | Oui | `ItemReplaced` | Transfert Mastery vers nouveau | Z2-AC14 |
| **R3** | CanonicalKey match **ET** Jaccard(keywords) 0.50-0.85 **ET** confidence delta < 0.15 | **ENRICH** : fusionner les keywords du nouveau dans l'existant, ajouter `source_refs` | Oui | `ItemsEnriched` | Inchangé | Z2-AC14 |
| **R4** | CanonicalKey match **ET** Jaccard(keywords) < 0.50 | **FLAG_CONFLICT** : possible contradiction sémantique (même terme, contenu très différent) | Non (HITL) | `ConflictDetected` | Gelé (pas de transition) | Z3-AC11 |
| **R5** | CanonicalKey miss **ET** Jaccard(keywords) > 0.70 **ET** termes proches (edit distance < 3) | **FLAG_CONFLICT** : probable doublon avec formulation différente | Non (HITL si confiance < 0.7) | `ConflictDetected` | Gelé | Z3-AC11 |
| **R6** | CanonicalKey miss **ET** Jaccard(keywords) 0.50-0.70 | **CREATE_ALT** : créer comme item alternatif en attente de review | Non (HITL) | Aucun | Aucun (pas de Mastery créé tant que non validé) | Z3-AC11 |
| **R7** | CanonicalKey miss **ET** Jaccard(keywords) < 0.50 | **APPEND** : nouvel item, nouvelle connaissance | Oui | `ItemsGenerated` | Nouveau Mastery UNKNOWN | Z2-AC14 |

---

## Diagramme de décision

```
Nouvel item extrait
        │
        ▼
CanonicalKey match avec un item existant ?
        │
   ┌────┴────┐
   │ OUI     │ NON
   ▼         ▼
Jaccard    Jaccard(keywords) avec tous les existants
keywords?       │
   │       ┌───┼───┐───────┐
   │       │   │   │       │
   │     >0.70 │ 0.50-0.70 │ <0.50
   │       │   │   │       │
   │       ▼   │   ▼       ▼
   │     R5:   │  R6:     R7:
   │    FLAG   │ CREATE   APPEND
   │  CONFLICT │  ALT      │
   │       │   │   │    ItemsGenerated
   │       │   │   │    Mastery UNKNOWN
   │       │   │   │
   ├───────┤   │   │
   │       │   │   │
   │ >0.85 │ 0.50-0.85   <0.50
   │       │       │        │
   ▼       ▼       ▼        ▼
  R1:     ∆conf   R4:
 ARCHIVE  > 0.15? FLAG
  DUP     │       CONFLICT
   │   ┌──┴──┐
   │   OUI  NON
   │    │    │
   ▼    ▼    ▼
  R1   R2:  R3:
      REPLACE ENRICH
```

---

## Détails par action

### R1 : ARCHIVE_DUP (doublon exact)

**Quand** : même terme normalisé, même type, et keywords quasi-identiques.

**Algorithme** :
1. Comparer `CanonicalKey(new) == CanonicalKey(existing)`.
2. Calculer `Jaccard(new.keywords, existing.keywords)`.
3. Si Jaccard > 0.85, c'est un doublon.
4. Archiver celui avec la confiance la plus basse (`item.Archive(now)`).
5. Si confiances égales (delta < 0.05), archiver le plus récent (conserver l'original).

**Exemple** :
- Existant : `term="Chloroplaste"`, keywords=["organite", "cellule végétale", "chlorophylle"], confidence=0.90
- Nouveau : `term="chloroplaste"`, keywords=["organite", "cellule", "chlorophylle", "photosynthèse"], confidence=0.85
- CanonicalKey match : oui (`chloroplaste|KNOWLEDGE|SVT-PHOTO`)
- Jaccard = |{organite, chlorophylle}| / |{organite, cellule végétale, cellule, chlorophylle, photosynthèse}| = 2/5 = 0.40
- Jaccard < 0.85 donc **pas R1**, on passe à R3 (ENRICH).

**Contre-exemple** (vrai doublon) :
- Existant : keywords=["organite", "cellule végétale", "chlorophylle", "photosynthèse"]
- Nouveau : keywords=["organite", "cellule végétale", "chlorophylle"]
- Jaccard = 3/4 = 0.75 -- toujours pas > 0.85.
- Pour atteindre R1, il faut vraiment que les keywords soient quasi-identiques : Jaccard > 0.85 signifie que le nouveau n'apporte essentiellement rien de neuf.

### R2 : REPLACE (meilleure extraction)

**Quand** : même terme, keywords partiellement similaires, mais la nouvelle extraction est significativement meilleure.

**Algorithme** :
1. CanonicalKey match.
2. Jaccard(keywords) entre 0.50 et 0.85.
3. `new.confidence - existing.confidence > 0.15`.
4. L'ancien item est archivé, le nouveau prend sa place.
5. Le Mastery est transféré.

**Cas typique** : l'élève re-photographie une page floue (confiance 0.55) avec une meilleure photo (confiance 0.88). Delta = 0.33 > 0.15 donc REPLACE.

**Seuil de 0.15** : calibré pour éviter les remplacements sur du bruit. Un delta de 0.05-0.10 peut être dû à la variance OCR sur la même page.

### R3 : ENRICH (enrichissement)

**Quand** : même terme, keywords partiellement similaires, confiances proches.

**Algorithme** :
1. CanonicalKey match.
2. Jaccard(keywords) entre 0.50 et 0.85.
3. Delta confiance <= 0.15.
4. Les keywords du nouveau qui ne sont pas dans l'existant sont ajoutés.
5. Les `source_refs` du nouveau sont ajoutés à l'existant.
6. La confiance est mise à jour : `max(existing.confidence, new.confidence)`.

**Exemple** :
- Existant : `term="Stomate"`, keywords=["orifice", "feuille", "échanges gazeux"]
- Nouveau (depuis correction d'exercice) : `term="Stomate"`, keywords=["orifice", "feuille", "CO₂", "O₂", "face inférieure"]
- Jaccard = |{orifice, feuille}| / |{orifice, feuille, échanges gazeux, CO₂, O₂, face inférieure}| = 2/6 = 0.33
- Jaccard < 0.50, donc c'est en fait R4 (FLAG_CONFLICT), pas R3.
- Mais les CanonicalKeys matchent et le contenu n'est pas contradictoire (il enrichit).
- **Correction** : quand le CanonicalKey matche, si les keywords sont complémentaires (pas contradictoires) et que le delta de confiance est faible, on fait ENRICH même avec un Jaccard < 0.50. La règle R4 ne s'applique que quand il y a une **contradiction sémantique détectée** (via le CoherenceChecker LLM), pas juste un faible Jaccard.

### R3 ajusté : ENRICH (version corrigée)

**Condition réelle** :
1. CanonicalKey match.
2. Pas de contradiction détectée par le CoherenceChecker.
3. Delta confiance <= 0.15.
4. -> ENRICH (fusionner keywords + source_refs).

**Condition R4 réelle** :
1. CanonicalKey match.
2. Contradiction détectée par le CoherenceChecker (contenu sémantiquement incompatible).
3. -> FLAG_CONFLICT.

Le Jaccard sert de **pré-filtre** pour décider s'il faut appeler le CoherenceChecker LLM (coûteux). Si Jaccard > 0.85, pas besoin de checker (doublon évident). Si Jaccard < 0.50 avec CanonicalKey match, le CoherenceChecker est appelé pour distinguer enrichissement de contradiction.

### R4 : FLAG_CONFLICT (contradiction)

**Quand** : même terme, mais le CoherenceChecker LLM détecte une contradiction sémantique.

**Exemple** :
- Existant : `term="Densité"`, definition: "masse divisée par le volume" (correct)
- Nouveau : `term="Densité"`, definition: "volume divisé par la masse" (incorrect)
- CanonicalKey match.
- CoherenceChecker retourne : contradiction, severity=0.9.
- Les deux items sont flaggés `coherence_flag='contradiction'`, `validation_required=true`.
- ValidationTask créée (Z3-AC11).

### R5 : FLAG_CONFLICT (doublon avec formulation différente)

**Quand** : termes différents mais keywords très proches.

**Exemple** :
- Existant : `term="Masse volumique"`, keywords=["masse", "volume", "rho", "kg/m³"]
- Nouveau : `term="Densité"`, keywords=["masse", "volume", "rho", "eau"]
- CanonicalKey miss (termes différents).
- Jaccard = |{masse, volume, rho}| / |{masse, volume, rho, kg/m³, eau}| = 3/5 = 0.60
- Wait, 0.60 < 0.70 donc c'est R6, pas R5.
- Pour R5, il faudrait Jaccard > 0.70 : par exemple deux formulations d'un même concept avec les mêmes keywords.

### R6 : CREATE_ALT (alternative en attente)

**Quand** : pas de match exact, mais similarité modérée. Possible enrichissement ou possible concept distinct.

**Algorithme** :
1. Créer le nouvel item avec `merge_status='alternative'`.
2. Ne pas créer de Mastery (item non validé).
3. Si confiance >= 0.70, pas de HITL (l'alternative est simplement conservée).
4. Si confiance < 0.70, créer une ValidationTask pour que l'élève/parent décide.

### R7 : APPEND (nouveau concept)

**Quand** : aucune similarité significative avec les items existants.

**Algorithme** :
1. Créer l'item avec `merge_status='original'`.
2. Rattacher à une Notion existante (par nom) ou créer une nouvelle Notion.
3. Émettre `ItemsGenerated`.
4. Créer un Mastery UNKNOWN pour l'utilisateur.

---

## Ordre d'évaluation

Les règles sont évaluées dans l'ordre R1-R7 pour chaque item entrant. La première règle qui matche est appliquée.

Pour chaque item entrant, le MergeEngine :
1. Calcule le CanonicalKey du nouvel item.
2. Cherche un item existant avec le même CanonicalKey.
3. Si trouvé : calcule Jaccard, applique R1/R2/R3/R4.
4. Si non trouvé : calcule Jaccard avec tous les existants, applique R5/R6/R7.

**Complexité** : O(n*m) où n = items entrants, m = items existants. Pour un chapitre typique (5-20 items), c'est négligeable.

---

## Seuils récapitulatifs

| Seuil | Valeur | Justification |
|-------|--------|---------------|
| Jaccard doublon | > 0.85 | Très peu de keywords différents = même contenu |
| Jaccard enrichissement | 0.50-0.85 | Keywords complémentaires = enrichissement probable |
| Jaccard matching cross-term | > 0.70 | Même concept, formulation différente |
| Jaccard modéré cross-term | 0.50-0.70 | Zone grise, nécessite review |
| Jaccard pas de match | < 0.50 | Concepts distincts |
| Confidence delta replace | > 0.15 | Écart significatif, pas du bruit OCR |
| Confidence HITL obligatoire | < 0.50 | Extraction trop incertaine |
| Confidence pas de HITL | >= 0.85 | Extraction fiable (Z3-AC09) |
| OCR confidence floue | < 0.30 | Page floue (Z2-AC03) |
| Text similarity re-photo | > 0.80 | Même page re-photographiée |
| Edit distance termes proches | < 3 | Typo ou variante orthographique |

---

## Cas limites

### Cas 1 : Item PROCEDURE avec même terme mais étapes différentes

Un item PROCEDURE "Calcul de la densité" peut avoir des étapes différentes selon la source (méthode du prof vs. méthode du manuel). Dans ce cas :
- Le CanonicalKey matche.
- Les keywords matchent (Jaccard > 0.50).
- Mais les `steps` divergent.

**Décision** : le CoherenceChecker compare les steps. Si les steps sont compatibles (un est un sous-ensemble de l'autre), c'est un ENRICH. Si les steps sont contradictoires, c'est un FLAG_CONFLICT.

### Cas 2 : Item DOCUMENT (schéma) re-extrait

Un schéma re-photographié produit un nouvel item DOCUMENT avec un visual_block_id différent mais le même `term`. Dans ce cas :
- CanonicalKey matche.
- On compare les `ocr_confidence` des blocs sources.
- Si le nouveau crop est de meilleure qualité (confidence > old + 0.15), REPLACE.
- Sinon, on garde l'existant (le premier crop suffit).

### Cas 3 : Fiche de correction qui enrichit des items existants

Une correction d'exercice (Session 3 dans l'exemple photosynthèse) reprend des définitions déjà extraites. Dans ce cas :
- Les CanonicalKeys matchent pour les items repris.
- La correction confirme les définitions existantes (pas de contradiction).
- Action = ENRICH : on ajoute la page de correction comme `source_ref` supplémentaire.
- La confiance peut augmenter (cross-validation par une source supplémentaire).
