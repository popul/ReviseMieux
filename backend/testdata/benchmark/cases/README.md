# Guide de remplissage des cas benchmark

Ce guide explique comment remplir chaque dossier de cas pour créer la baseline humaine du benchmark OCR + IDP (structuration).

## Vue d'ensemble

Chaque cas = un chapitre de cours photographié + transcription humaine + items pédagogiques attendus.

```
cases/11_histoire/
├── images/           ← photos de cahier (.jpeg)
├── input.json        ← transcription OCR (golden)
├── golden_output.json ← items pédagogiques attendus (golden)
└── metadata.json     ← infos sur le cas (matière, niveau, etc.)
```

- `input.json` sert de golden pour le **benchmark OCR** (images → blocs texte)
- `golden_output.json` sert de golden pour le **benchmark IDP** (blocs texte → items)

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

Un bloc = une unité de sens autonome sur la page :

| Sur la page | Bloc |
|-------------|------|
| Titre du chapitre | 1 bloc TEXT |
| Un paragraphe de cours | 1 bloc TEXT |
| Une question numérotée | 1 bloc TEXT séparé |
| Réponse manuscrite de l'élève | 1 bloc TEXT séparé |
| Un schéma avec ses légendes | 1 bloc DIAGRAM |
| Un graphique avec ses axes | 1 bloc DIAGRAM |
| Une photo collée (microscope, dispositif) | 1 bloc DIAGRAM |
| Un tableau de mesures | 1 bloc TABLE |
| Un encadré "à retenir" | 1 bloc TEXT séparé |

Ne pas fusionner plusieurs unités dans un même bloc. Ne pas découper une phrase en plusieurs blocs. Séparer quand la source change (polycopié vs manuscrit) ou quand le sujet change.

### Règles de transcription

1. **Mot à mot.** Transcris exactement ce qui est écrit, y compris les fautes de l'élève ("staumates" reste "staumates", pas "stomates").

2. **Visuels entre crochets.** Pour les schémas, graphiques et photos, décris le contenu en détail. Inclure :
   - Le type de document (schéma, graphique, photo microscopique, dispositif expérimental...)
   - Les éléments visibles (légendes, axes, courbes, structures annotées)
   - Les valeurs numériques et unités lisibles
   - Les couleurs ou annotations significatives
   ```
   [Schéma anatomique : Coupe transversale d'une feuille montrant les cellules de la feuille,
   cellule épidermique, stomate, CO2, O2, air atmosphérique, et face inférieure de la feuille]
   ```
   ```
   [Graphique : Concentration en CO2 (unités arbitraires) en fonction du temps (en minutes),
   montrant deux courbes - 'Sans feuilles' stable à 1200 unités et 'Avec feuilles' diminuant
   de 1200 à 400 unités sur 10 minutes. Légende: 'Evolution de la concentration en CO2
   dans une enceinte éclairée contenant ou non des feuilles de blé.']
   ```

3. **Préserve la structure.** Garde les numéros de questions, les retours à la ligne, les tirets.

4. **Mots illisibles.** Si un mot est vraiment impossible à lire, mets `[illisible]`.

5. **Confidence à 1.0.** Tu es l'expert humain, tous tes blocs ont une confiance de 1.0.

### Valeurs de `block_type`

| Type | Quand l'utiliser |
|------|------------------|
| `TEXT` | Texte imprimé, manuscrit, questions, réponses, titres, encadrés |
| `DIAGRAM` | Schémas, graphiques, dessins, photos collées, photos microscopiques, cartes |
| `TABLE` | Tableaux de mesures, tableaux comparatifs |

---

## Etape 3 — Items pédagogiques (`golden_output.json`)

Relis ta transcription et extrais les items que l'élève doit retenir pour ses révisions.

### Types d'items

**KNOWLEDGE** — un fait, une définition, un concept à mémoriser :
```json
{
  "type": "KNOWLEDGE",
  "term": "Les poils absorbants représentent une surface d'absorption de 400 m²",
  "keywords": ["poils absorbants", "surface d'absorption", "400 m²", "racines"],
  "notion": "Prélèvement de matière minérale"
}
```

**PROCEDURE** — un protocole expérimental, une méthode avec des étapes ordonnées :
```json
{
  "type": "PROCEDURE",
  "term": "Mesure de la concentration en CO2 avec ExAO",
  "keywords": ["ExAO", "CO2", "feuilles", "enceinte éclairée"],
  "steps": [
    "Couper des feuilles de blé en fragments",
    "Placer dans enceinte éclairée",
    "Mesurer CO2 avec ExAO pendant 10 minutes",
    "Renouveler sans feuilles",
    "Comparer les résultats"
  ],
  "notion": "Échanges gazeux"
}
```

**DOCUMENT** — un document visuel distinct à savoir analyser :
```json
{
  "type": "DOCUMENT",
  "term": "Graphique d'évolution de la concentration en CO2 avec et sans feuilles",
  "keywords": ["graphique", "concentration CO2", "courbe avec feuilles", "courbe sans feuilles", "1200 unités", "400 unités", "diminution"],
  "notion": "Échanges gazeux"
}
```

### Règles importantes pour les DOCUMENT

- **Chaque document visuel distinct** (schéma, graphique, image microscopique, photo d'expérience...) doit produire un item DOCUMENT séparé.
- **Ne PAS fusionner** un document avec un KNOWLEDGE ou PROCEDURE qu'il illustre : savoir analyser un graphique est une compétence distincte de connaître le concept ou la procédure.
- Le **term** décrit précisément le type de document et son contenu (ex: "Graphique d'évolution de la concentration en CO2 avec et sans feuilles", pas juste "Graphique CO2").
- Les **keywords** incluent les données clés visibles : valeurs numériques, légendes, axes, structures annotées.
- Ne pas confondre **PROCEDURE** (protocole à reproduire) et **DOCUMENT** (document visuel à analyser) : un schéma d'expérience est un DOCUMENT, le protocole de cette expérience est un PROCEDURE.

### Conseils

- **3 à 8 keywords** par item, incluant les données factuelles du cours (chiffres, unités, noms propres)
- **`notion`** = le concept chapeau qui regroupe plusieurs items (2 à 7 notions par chapitre)
- **`notions`** en bas du fichier = la liste de toutes les notions uniques utilisées
- **`steps`** uniquement pour les PROCEDURE (3 à 6 étapes)
- Un bloc DIAGRAM dans `input.json` devrait typiquement correspondre à un item DOCUMENT dans `golden_output.json`

---

## Etape 4 — Métadonnées (`metadata.json`)

```json
{
  "id": "11_histoire",
  "subject": "Histoire",
  "level": "4e",
  "topic": "La Révolution française (1789-1799)",
  "difficulty": "medium",
  "has_images": true,
  "expected_item_count": 8,
  "expected_types": {
    "KNOWLEDGE": 4,
    "PROCEDURE": 1,
    "DOCUMENT": 3
  },
  "ocr_source": "human_expert",
  "ocr_author": "Ton prénom",
  "ocr_date": "2026-03-15",
  "ocr_metadata": {
    "handwriting_quality": "normal",
    "ink_types": ["blue_pen"],
    "photo_quality": "good",
    "content_types": ["text", "diagram"],
    "support": "notebook_with_printout",
    "pages": 3
  },
  "notes": "Description libre du contenu."
}
```

| Champ | Valeurs |
|-------|---------|
| `difficulty` | `low` (surtout imprimé), `medium` (mix manuscrit/imprimé), `high` (beaucoup de manuscrit, formules, brouillon) |
| `handwriting_quality` | `soigné`, `normal`, `brouillon` |
| `ink_types` | `blue_pen`, `red_pen`, `black_pen`, `pencil` |
| `photo_quality` | `good`, `average`, `poor` |
| `content_types` | `text`, `diagram`, `table`, `formula`, `map` |
| `support` | `pure_notebook`, `notebook_with_printout`, `printout_annotated` |

---

## Checklist par cas

Avant de considérer un cas comme terminé :

- [ ] 2-4 photos dans `images/` (format .jpeg)
- [ ] `input.json` : tous les blocs transcrits, `confidence: 1.0`, `block_type` correct
- [ ] `input.json` : les blocs DIAGRAM ont des descriptions détaillées (valeurs, légendes, structures)
- [ ] `golden_output.json` : chaque item a `type`, `term`, `keywords`, `notion`
- [ ] `golden_output.json` : les PROCEDURE ont un champ `steps`
- [ ] `golden_output.json` : chaque bloc DIAGRAM a un item DOCUMENT correspondant
- [ ] `golden_output.json` : `notions` en bas contient toutes les notions uniques
- [ ] `metadata.json` : `expected_item_count` correspond au nombre réel d'items
- [ ] Relecture finale : comparer la transcription aux photos une dernière fois

---

## Génération assistée par LLM (Makefile)

Le Makefile à la racine de `benchmark/` génère un **brouillon** des 3 fichiers JSON à partir des photos via l'API Anthropic.

Les prompts utilisés sont les **memes que ceux de la prod** (source unique dans `backend/internal/infra/anthropic/prompts/`).

### Pré-requis

- `ANTHROPIC_API_KEY` définie dans l'environnement
- `python3`, `jq` installés

### Commandes rapides

```bash
# Depuis backend/testdata/benchmark/
make help                                                      # Afficher toutes les targets
make generate CASE=11_histoire MODEL=claude-sonnet-4-20250514  # Générer les 3 JSON
make generate-input CASE=11_histoire                           # Seulement la transcription
make validate CASE=11_histoire                                 # Vérifier la validité JSON
make list-cases                                                # Voir le statut de tous les cas
```

### Workflow recommandé

1. `make generate CASE=11_histoire MODEL=claude-sonnet-4-20250514` — génère les brouillons
2. Ouvre chaque JSON et **corrige à la main** (le LLM fait des erreurs, surtout sur le manuscrit et les DOCUMENT)
3. `make validate CASE=11_histoire` — vérifie que le JSON est bien formé
4. Passe la checklist ci-dessus

> **Les fichiers générés sont des brouillons.** Le LLM peut inventer du contenu, mal lire l'écriture manuscrite, fusionner des documents visuels, ou rater des blocs. La relecture humaine est indispensable.

---

## Cas prioritaires à ajouter

Pour rendre le benchmark fiable, il faut couvrir la diversité des cahiers :

| Cas | Matière | Interet |
|-----|---------|---------|
| Existant : `10_SVT_cours_louis` | SVT 5e | Schémas, graphiques, protocoles |
| A ajouter | Histoire/Français | Beaucoup de texte, peu de schémas |
| A ajouter | Maths/Physique | Formules, calculs, tableaux de mesures |
| A ajouter | Ecriture difficile | Brouillon, ratures — tester la robustesse OCR |

## Exemple complet de référence

Le cas `10_SVT_cours_louis/` est un exemple complet et validé. En cas de doute sur le format ou le niveau de détail attendu, consulte ses fichiers.
