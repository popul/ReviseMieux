# Guide de remplissage des cas benchmark OCR

Ce guide t'explique comment remplir chaque dossier de cas pour créer la baseline humaine du benchmark OCR.

## Vue d'ensemble

Chaque cas = un chapitre de cours photographié + ta transcription parfaite.

```
cases/11_histoire/
├── images/           ← tes photos de cahier (.jpeg)
├── input.json        ← ta transcription (golden OCR)
├── golden_output.json ← les items pédagogiques que tu en extrais
└── metadata.json     ← infos sur le cas (matière, niveau, etc.)
```

Le `input.json` est **la pièce maîtresse** : c'est contre lui que tous les modèles OCR seront comparés.

---

## Etape 1 — Photos

Prends 2 à 4 photos du cahier avec ton téléphone.

**Conditions** :
- Cahier posé à plat sur une table
- Eclairage naturel ou lampe (pas de flash)
- Toute la page visible, y compris les marges
- Résolution native (ne pas compresser)
- C'est OK si c'est un peu de travers ou si le bord de l'autre page apparait

**Nommage** : garde le nom du téléphone (`IMG_XXXX.jpeg`) et place les fichiers dans `images/`.

---

## Etape 2 — Transcription (`input.json`)

Ouvre les photos sur un grand écran et transcris bloc par bloc.

### Qu'est-ce qu'un bloc ?

Un bloc = une unité de sens sur la page. Exemples :

| Sur la page | Bloc |
|-------------|------|
| Titre du chapitre + intro du cours | 1 bloc TEXT |
| Enoncé d'une activité (questions) | 1 bloc TEXT |
| Réponses manuscrites de l'élève | 1 bloc TEXT séparé |
| Un schéma avec ses légendes | 1 bloc DIAGRAM |
| Un graphique avec ses axes | 1 bloc DIAGRAM |
| Un tableau de mesures | 1 bloc TABLE |
| Un encadré "à retenir" | 1 bloc TEXT séparé |

En cas de doute : mieux vaut trop de blocs que pas assez. Sépare quand la source change (polycopié vs manuscrit) ou quand le sujet change.

### Règles de transcription

1. **Mot à mot.** Transcris exactement ce qui est écrit, y compris les fautes de l'élève ("staumates" reste "staumates", pas "stomates").

2. **Visuels entre crochets.** Pour les schémas, graphiques et photos, décris ce que tu vois :
   ```
   [Schéma : coupe de feuille — Cellule épidermique, Stomate, CO2, O2]
   ```
   ```
   [Graphique : concentration en CO2 en fonction du temps. Courbe "avec feuilles" diminue de 800 à 200 en 10 min. Courbe "sans feuille" reste constante à 1800.]
   ```

3. **Préserve la structure.** Garde les numéros de questions, les retours à la ligne, les tirets.

4. **Mots illisibles.** Si un mot est vraiment impossible à lire, mets `[illisible]`.

5. **Confidence à 1.0.** Tu es l'expert humain, tous tes blocs ont une confiance de 1.0.

### Format

```json
{
  "blocks": [
    {
      "text": "Chapitre 5 : Les circuits électriques\n\nI – Circuit en série\n\nDans un circuit en série, les dipôles sont branchés les uns à la suite des autres.",
      "block_type": "TEXT",
      "confidence": 1.0
    },
    {
      "text": "[Schéma : circuit en série avec une pile, une lampe L1, une lampe L2 et un interrupteur. Flèche indiquant le sens du courant.]",
      "block_type": "DIAGRAM",
      "confidence": 1.0
    },
    {
      "text": "Activité 2 :\n1 – Que se passe-t-il si on dévisse la lampe L1 ?\n2 – L'autre lampe s'éteint aussi. Pourquoi ?\n\nRéponse : parce que le circuit est ouvert, le courant ne passe plus.",
      "block_type": "TEXT",
      "confidence": 1.0
    }
  ]
}
```

### Valeurs de `block_type`

| Type | Quand l'utiliser |
|------|------------------|
| `TEXT` | Texte imprimé, manuscrit, questions, réponses, titres, encadrés |
| `DIAGRAM` | Schémas, graphiques, dessins, photos collées, cartes |
| `TABLE` | Tableaux de mesures, tableaux comparatifs |

---

## Etape 3 — Items pédagogiques (`golden_output.json`)

Relis ta transcription et extrais les items que l'élève doit retenir pour ses révisions.

### Types d'items

**KNOWLEDGE** — un fait, une définition, une règle à mémoriser :
```json
{
  "type": "KNOWLEDGE",
  "term": "Dans un circuit en série, si un dipôle est défaillant, tous les autres s'éteignent car le circuit est ouvert.",
  "keywords": ["circuit en série", "dipôle", "circuit ouvert"],
  "notion": "Circuit en série"
}
```

**PROCEDURE** — une méthode, un protocole, des étapes à suivre :
```json
{
  "type": "PROCEDURE",
  "term": "Réaliser un circuit en série avec deux lampes et un interrupteur.",
  "keywords": ["circuit en série", "lampe", "interrupteur", "montage"],
  "steps": [
    "Connecter le fil à la borne + de la pile",
    "Relier la première lampe L1",
    "Relier la deuxième lampe L2",
    "Ajouter l'interrupteur",
    "Fermer le circuit en reliant à la borne - de la pile"
  ],
  "notion": "Circuit en série"
}
```

**DOCUMENT** — un schéma ou graphique important à savoir lire/reproduire :
```json
{
  "type": "DOCUMENT",
  "term": "Schéma d'un circuit en série avec les symboles normalisés : pile, lampe, interrupteur, fils de connexion.",
  "keywords": ["schéma", "circuit en série", "symboles normalisés", "pile", "lampe"],
  "notion": "Circuit en série"
}
```

### Structure complète

```json
{
  "items": [
    { "type": "KNOWLEDGE", "term": "...", "keywords": [...], "notion": "..." },
    { "type": "PROCEDURE", "term": "...", "keywords": [...], "steps": [...], "notion": "..." },
    { "type": "DOCUMENT", "term": "...", "keywords": [...], "notion": "..." }
  ],
  "notions": [
    "Circuit en série",
    "Circuit en dérivation"
  ]
}
```

### Conseils

- **3 à 5 keywords** par item, choisis les termes que l'élève doit connaître
- **`notion`** = le concept chapeau qui regroupe plusieurs items (2 à 4 notions par chapitre)
- **`notions`** en bas du fichier = la liste de toutes les notions uniques utilisées
- **`steps`** uniquement pour les PROCEDURE (3 à 6 étapes)
- Vise **6 à 10 items** par cas, c'est la fourchette typique d'un chapitre

---

## Etape 4 — Métadonnées (`metadata.json`)

Complète les champs vides du squelette.

### Champs à remplir

```json
{
  "id": "11_histoire",                          // ne pas changer
  "subject": "Histoire",                        // ne pas changer
  "level": "4e",                                // ← classe : 6e, 5e, 4e ou 3e
  "topic": "La Révolution française (1789-1799)", // ← titre du chapitre
  "difficulty": "medium",                       // low, medium ou high
  "has_images": true,                           // ne pas changer
  "expected_item_count": 8,                     // ← nombre d'items dans golden_output.json
  "expected_types": {                           // ← décompte par type
    "KNOWLEDGE": 6,
    "PROCEDURE": 1,
    "DOCUMENT": 1
  },
  "ocr_source": "human_expert",                // ne pas changer
  "ocr_author": "Ton prénom",                  // ← qui a transcrit
  "ocr_date": "2026-03-15",                    // ← date de transcription (AAAA-MM-JJ)
  "ocr_metadata": {
    "handwriting_quality": "normal",            // ← soigné, normal ou brouillon
    "ink_types": ["blue_pen"],                  // ← blue_pen, red_pen, black_pen, pencil
    "photo_quality": "good",                    // ← good, average, poor
    "content_types": ["text", "diagram"],       // ← text, diagram, table, formula, map
    "support": "notebook_with_printout",        // ← pure_notebook, notebook_with_printout, printout_annotated
    "pages": 3                                  // ← nombre de photos
  },
  "notes": "Description libre du contenu."      // ← ce que tu veux
}
```

### Valeurs de `difficulty`

| Valeur | Signification pour l'OCR |
|--------|--------------------------|
| `low` | Surtout du texte imprimé, écriture soignée, peu de schémas |
| `medium` | Mix manuscrit/imprimé, quelques schémas ou formules |
| `high` | Beaucoup de manuscrit, formules, schémas complexes, écriture brouillon |

---

## Checklist par cas

Avant de considérer un cas comme terminé :

- [ ] 2-4 photos dans `images/` (format .jpeg)
- [ ] `input.json` : tous les blocs transcrits, `confidence: 1.0`, `block_type` correct
- [ ] `golden_output.json` : 6-10 items, chaque item a `type`, `term`, `keywords`, `notion`
- [ ] `golden_output.json` : les PROCEDURE ont un champ `steps`
- [ ] `golden_output.json` : `notions` en bas contient toutes les notions uniques
- [ ] `metadata.json` : tous les champs remplis, `expected_item_count` correspond au nombre réel d'items
- [ ] Relecture finale : comparer la transcription aux photos une dernière fois

---

## Génération assistée par LLM (Makefile)

Tu peux utiliser le Makefile à la racine de `benchmark/` pour générer un **brouillon** des 3 fichiers JSON à partir des photos, via l'API Anthropic (Claude Opus 4.6 en mode vision).

### Pré-requis

- `ANTHROPIC_API_KEY` définie dans l'environnement
- `curl`, `jq`, `python3` installés

### Commandes rapides

```bash
# Depuis backend/testdata/benchmark/
make help                                   # Afficher toutes les targets
make generate CASE=10_SVT_cours_louis       # Générer les 3 JSON d'un coup
make generate-input CASE=10_SVT_cours_louis # Seulement la transcription
make validate CASE=10_SVT_cours_louis       # Vérifier la validité JSON
make list-cases                             # Voir le statut de tous les cas
```

### Workflow recommandé

1. `make generate CASE=10_SVT_cours_louis` — génère les brouillons
2. Ouvre chaque JSON et **corrige à la main** (le LLM fait des erreurs, surtout sur le manuscrit)
3. `make validate CASE=10_SVT_cours_louis` — vérifie que le JSON est bien formé
4. Passe la checklist ci-dessous

> **⚠️ Les fichiers générés sont des brouillons.** Le LLM peut inventer du contenu, mal lire l'écriture manuscrite, ou rater des blocs. La relecture humaine est indispensable.

---

## Exemple complet de référence

Le cas `10_SVT_cours_louis/` est un exemple complet et validé. En cas de doute sur le format ou le niveau de détail attendu, consulte ses fichiers.
