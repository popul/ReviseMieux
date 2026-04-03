# Prompt 1 — Pipeline OCR/IDP incrémental pour Révise Mieux

> **Objectif** : Concevoir l'architecture, le format pivot, et le workflow d'ingestion incrémentale pour transformer progressivement des photos de cahier en données structurées exploitables par le moteur de révision.

---

You are a senior AI architect specialized in OCR, Intelligent Document Processing (IDP), educational content structuring, and incremental knowledge systems.

## Mission

Help me design an incremental OCR/IDP architecture for an educational product called "Révise Mieux".

## Contexte produit

Révise Mieux est un SaaS éducatif qui transforme des photos de cahier manuscrit en assistant de révision pour collégiens (11-15 ans). L'app ingère du contenu de cours provenant de :
- photos de cahier manuscrit ou imprimé
- notes de cours scannées
- extraits de manuels
- pages mixtes contenant texte, titres, annotations, schémas, tableaux, flèches, surlignages, et notes du professeur

Le cas d'usage n'est PAS de l'OCR enterprise générique. L'objectif est de transformer progressivement du contenu scolaire en assets de révision structurés : résumés, fiches, quiz, cartes conceptuelles, supports d'entraînement.

### Stack technique concernée

| Composant | Technologie |
|-----------|-------------|
| Backend API | Go 1.23 + Gin |
| Base de données | PostgreSQL 16 (pgx/v5, SQL brut) |
| LLM structuration (IDP) | mistral-small (quality 0.87, $0.00006/item) ou Qwen3.5-397B (quality 0.88, $0.00011/item) — post-benchmark 21 modèles |
| LLM fidelity/questions | Haiku 4.5 (à benchmarker — pas encore testé sur ces cas d'usage) |
| OCR MVP (API) | Gemini 2.5 Flash — VLM page entière (quality 0.84, $0.001/page, manuscrit zero-shot) |
| OCR local | Qwen3-VL-32B (quality 0.85, #1 OCR) + Gemma 3 27B (quality 0.84 IDP) via Ollama (~2 min/chapitre, $0) |
| OCR gratuit | Nemotron Nano VL (détection 97%, quality 0.79, $0) — alternative locale |
| Scoring réponses | Gemini 2.5 Flash (vision, bas coût) — scoring + extraction |
| Cache | Redis (go-redis/v9) |
| Storage | S3 / MinIO (object storage) |
| IDs | UUIDv7 (timestamp-sortable, généré côté Go) |
| Architecture | DDD hexagonale, 4 bounded contexts |

### Pipeline actuelle (Pipeline J0) — architecture 2 étages

Le pipeline sépare **extraction** et **structuration** en deux étapes distinctes avec des modèles différents :

```
upload photo → OCR/VLM (Gemini Flash ou Qwen3-VL-32B)
                   ↓ Markdown structuré (blocs texte + descriptions visuelles)
               IDP/LLM (mistral-small ou Qwen3.5-397B)
                   ↓ Items pédagogiques typés (KNOWLEDGE, PROCEDURE, DOCUMENT, WRITING)
               → génération items + notions
```

**Principe clé** : l'OCR (vision, image) produit du Markdown ; la structuration pédagogique (texte seul) consomme ce Markdown sans revoir l'image. Cela permet d'utiliser chaque modèle là où il excelle.

### Combos recommandés (post-benchmark mars 2026)

| Stratégie | OCR | IDP | Coût/chapitre | Qualité |
|-----------|-----|-----|---------------|---------|
| **Best quality** | Qwen3-VL-32B (0.85) | Qwen3.5-397B (0.88) | ~$0.003 | Meilleure accuracy |
| **Best value** | Gemini 2.5 Flash (0.84) | mistral-small (0.87) | ~$0.002 | Très bon compromis |
| **Ultra-cheap** | Nemotron Nano VL (0.79) | Gemma 3 27B (0.84) | ~$0.000 | Gratuit |
| **Local (Lot 0)** | Qwen3-VL-32B Q4 | Gemma 3 27B Q4 | $0 | ~2 min/chapitre |

**Point critique** : Claude Sonnet est le seul modèle avec des hallucinations (8% en IDP) et le plus cher en OCR ($0.053/run). Il n'est **pas recommandé** pour le pipeline.

Voir `backend/testdata/benchmark/README.md` et `docs/llm-strategy.md` §8-9 pour les résultats complets.

### Agrégat Capture existant (bounded context `chapter/`)

```
Chapter (agrégat racine)
  ├── Revision (entité distincte, table chapters_revisions — pas un value object imbriqué)
  │     ├── Page (image uploadée)
  │     │     └── Block (zone segmentée : TEXT, PHOTO, SCHEMA, MAP, GRAPH, TABLE, CIRCUIT, DECORATIVE)
  │     └── VisualBlock (éléments visuels extraits)
  ├── Item (unité de révision : KNOWLEDGE, PROCEDURE, DOCUMENT, WRITING)
  └── Notion (regroupement sémantique d'items)
```

### Cas d'usage LLM complets

| # | Cas d'usage | Modèle actuel | Modèle post-benchmark | Statut |
|---|------------|---------------|----------------------|--------|
| A | OCR extraction (images → Markdown) | — | **Gemini 2.5 Flash** (API) ou **Qwen3-VL-32B** (local) | Benchmarké |
| B | Structuration IDP (Markdown → Items) | ~~Sonnet 4.6~~ | **mistral-small** ou **Qwen3.5-397B** | Benchmarké |
| C | Fidelity check (fidélité sémantique) | Haiku 4.5 | À benchmarker | En attente |
| D | Détection de cohérence (doublons) | ~~Sonnet 4.6~~ | À benchmarker | En attente |
| E | Génération de questions (lazy, ≤500ms) | Haiku 4.5 | À benchmarker | En attente |
| F | Scoring réponses | — | **Gemini 2.5 Flash** (vision, bas coût) | Décidé |
| G | Feedback enrichi | Templates (pas de LLM) | Inchangé | Stable |

### Contrainte centrale : ingestion incrémentale

Un élève ne photographie pas tout son cours en une fois. Il peut :
1. Capturer 3 pages le lundi soir, 2 pages le mercredi, 1 page le vendredi
2. Re-photographier une page mal cadrée
3. Ajouter des pages d'exercices corrigés en classe
4. Capturer un fragment de cours manquant des semaines plus tard

Le système doit intégrer ces ajouts **incrémentalement** sans régénérer tout le chapitre, tout en maintenant la cohérence des Notions et Items existants.

### Documents de référence à consulter

| Document | Chemin | Contenu pertinent |
|----------|--------|-------------------|
| PRD | `docs/PRD.md` | Pipeline J0 (§10), modèle de données (§16), retry/fallback OCR (§13) |
| ACs Zone 2 | `docs/ac/Z2.md` | Critères d'acceptation de la capture et structuration |
| ACs Zone 3 | `docs/ac/Z3.md` | Fidelity check, cohérence, validation |
| ACs Zone 7 | `docs/ac/Z7.md` | Structuration LLM |
| Stratégie LLM | `docs/llm-strategy.md` | Choix de modèles post-benchmark, coûts, cas d'usage |
| Benchmark | `backend/testdata/benchmark/README.md` | Résultats benchmark 21 modèles |
| Schema SQL | `backend/migrations/001_initial_schema.sql` | Tables existantes |
| Domaine Chapter | `backend/internal/domain/chapter/` | Entités Go actuelles |
| Domaine Event | `backend/internal/domain/event/` | Domain events existants |

---

## Tes tâches

### 1. Recadrer la nature du problème

Commence par reformuler le problème avec précision :
- Qu'est-ce qui rend l'OCR/IDP éducatif différent de l'extraction documentaire générique ?
- Qu'est-ce qui rend l'ingestion incrémentale difficile ?
- Quels sont les risques majeurs si le système est mal conçu ?

Focus sur ces problèmes spécifiques :
- uploads partiels de cours
- pages arrivant dans le désordre
- contenu dupliqué
- contenu contradictoire ou chevauchant
- incertitude OCR
- schémas porteurs de sens non capturé par le texte seul
- évolution du cours dans le temps (corrections, compléments)
- nécessité de traçabilité (quel item vient de quelle page)

### 2. Proposer un format pivot canonique

Définis un format pivot robuste servant de représentation interne canonique d'un cours. Ce format doit :
- être machine-friendly en priorité (JSON sauf justification forte d'un autre format)
- rester inspectable par un humain (debugging)
- supporter le versioning
- supporter la traçabilité vers les pages sources
- supporter l'incertitude et les niveaux de confiance
- supporter les opérations de merge incrémental
- être compatible avec les context windows des modèles self-hosted (Ollama)

Le schema doit représenter au minimum :
- identité du cours (Chapter ID, UUIDv7)
- matière et niveau scolaire
- titre du cours
- sections / sous-sections
- concepts (→ Notions dans notre modèle)
- définitions, faits, exemples
- exercices si présents
- schémas / diagrammes et leur interprétation sémantique
- provenance source pour chaque bloc extrait (Page ID + coordonnées)
- score de confiance par unité extraite
- horodatages de capture
- métadonnées de version
- lien vers les révisions précédentes
- statut de merge / review

**Exemple concret attendu** sur le chapitre "La photosynthèse" (SVT, pack `SVT-PHOTO`) avec :
- Session 1 : 3 pages (intro + schéma + définitions)
- Session 2 : 2 pages (expériences + tableau de résultats)
- Session 3 : 1 page (correction d'exercice donné en classe)

### 3. Distinguer couches immutables vs mutables

Sépare explicitement :
- **Observations brutes extraites** (OCR text, bounding boxes, confiance) → immutable
- **Structure pédagogique interprétée** (Notions, Items, hiérarchie) → mutable enrichissable
- **Connaissance normalisée du cours** (le "canon" du chapitre) → mutable avec versioning
- **Outputs de révision dérivés** (fiches, quiz, questions) → régénérables, invalidables

Explique ce qui doit rester immutable pour la traçabilité, et ce qui peut évoluer quand de nouvelles pages arrivent.

### 4. Concevoir le workflow d'ingestion incrémentale

Décris étape par étape comment le système traite un nouveau batch de pages :

1. Upload / ingestion → S3 + métadonnées dans `pages`
2. Groupement de documents (rattachement au Chapter existant)
3. OCR / analyse de layout
4. Extraction de blocs (→ `blocks` table)
5. Interprétation sémantique (LLM structuration)
6. Matching avec les structures existantes (Notions, Items)
7. Merge / append / split / gestion de conflits
8. Review basée sur la confiance (HITL si nécessaire)
9. Régénération des seuls assets downstream impactés (invalidation cache Redis)
10. Publication des domain events (`BatchProcessed`, `NotionMerged`, `ItemArchived`)

Spécifie pour chaque étape :
- ce qui doit TOUJOURS être recalculé
- ce qui peut être RÉUTILISÉ
- ce qui doit être INVALIDÉ quand du nouveau contenu arrive

### 5. Gérer la continuité du cours dans le temps

Propose une stratégie pour déterminer si des pages nouvellement uploadées appartiennent à :
- le même cours (enrichissement)
- le même chapitre mais un autre cours
- une fiche de correction
- un doublon (même page re-photographiée)
- un enrichissement tardif d'un cours déjà connu

Sois explicite sur :
- les heuristiques (similarité textuelle, proximité temporelle, métadonnées)
- les indices de métadonnées (nom de fichier, date, ordre)
- le matching sémantique (LLM)
- les points de validation humaine (quand demander confirmation à l'élève/parent)

### 6. Définir les stratégies de merge

Quand de nouvelles pages arrivent, explique comment le système décide de :
- **Append** : ajouter du contenu à une Notion existante
- **Enrich** : enrichir un Item existant avec des détails supplémentaires
- **Replace** : remplacer une extraction de confiance inférieure
- **Flag conflict** : signaler une contradiction (ex: "densité = m/V" vs "densité = V/m")
- **Create alternative** : créer une interprétation alternative en attente de review
- **Mark duplicate** : marquer comme doublon ou quasi-doublon

Je veux des **règles concrètes avec seuils**, pas des principes génériques. Par exemple :
- Si similarity score > 0.85 entre deux Items → candidat doublon
- Si confiance nouvelle extraction > confiance ancienne + 0.15 → proposer remplacement
- Si contradiction détectée → flag immédiat + domain event `ConflictDetected`

### 7. Préserver la cohérence du cours

L'architecture doit éviter la fragmentation progressive. Le cours doit rester globalement cohérent au fil du temps.

Explique comment maintenir :
- des identifiants de concept stables (Notion IDs persistants)
- des liens entre concepts cross-uploads
- la cohérence de la hiérarchie de sections
- la déduplication des connaissances
- une représentation pédagogique finale propre

### 8. Impact sur les Mastery existants

Quand le contenu d'un chapitre évolue :
- Item **modifié** (enrichi par un nouveau batch) → faut-il reset le Mastery ?
- Item **archivé** (doublon détecté) → que devient le Mastery associé ? Transfert vers l'Item survivant ?
- Item **splitté** (un Item trop large découpé en deux) → comment répartir le Mastery ?
- Nouvel Item ajouté → Mastery initial = UNKNOWN, domain event `ItemsGenerated`

### 9. Supporter la génération d'outputs éducatifs

Le format pivot doit supporter la génération downstream de :
- fiches de révision structurées
- quiz adaptatifs (lazy generation, cache 24h Redis)
- flashcards
- résumés de concepts
- explications pour parents (dashboard parent)

Explique comment le pivot s'articule avec le bounded context `Session` (composition de questions, scoring).

### 10. Couche Human-in-the-Loop

Le système doit identifier les zones incertaines et les exposer clairement. Concevoir des checkpoints de review pour :
- OCR de faible confiance
- matching de cours ambigu
- extractions contradictoires
- interprétation de schémas incertaine
- décisions de merge avec impact pédagogique

Ceci s'articule avec le bounded context `Validation` existant (`ValidationTask`, actions HITL, fidelity check — voir `docs/ac/Z3.md`).

### 11. Stratégie d'implémentation

| Phase | Scope | Description |
|-------|-------|-------------|
| **MVP (Lot 0)** | Architecture minimale qui fonctionne | Pipeline linéaire, pas de merge, re-processing complet si ajout de pages |
| **V2** | Incrémental basique | Merge par Notion, détection de doublons, invalidation ciblée |
| **V3** | Knowledge graph éducatif | Graphe de concepts, résolution automatique de conflits, mémoire de cours |

---

## Format de sortie attendu

Structure ta réponse ainsi :

1. **Executive summary** (10 lignes max)
2. **Cadrage du problème** (éducatif ≠ générique, risques)
3. **Architecture recommandée** (diagramme Mermaid)
4. **Design du format pivot canonique** (principes + justifications)
5. **JSON Schema complet** du format pivot avec exemples sur "La photosynthèse"
6. **Workflow d'ingestion incrémentale** (diagramme de séquence Mermaid)
7. **Règles de merge / résolution de conflits** (tableau de décision avec seuils)
8. **Workflow de review humaine** (intégration avec `ValidationTask`)
9. **Impact sur la génération downstream** (invalidation, events)
10. **Impact sur les Mastery** (tableau de décision)
11. **Signatures Go** des méthodes domaine impliquées
12. **Migrations SQL** nécessaires
13. **Risques et mitigations**
14. **MVP vs évolution future**
15. **Recommandation finale**

## Barre de qualité

Ta réponse doit être :
- **concrète** — pas de généralités OCR enterprise
- **orientée implémentation** — signatures Go, schemas SQL, JSON schemas
- **adaptée au contenu éducatif** — pas un pipeline de factures
- **explicite sur les trade-offs** — coût LLM, latence, complexité
- **conçue pour l'ingestion cumulative dans le temps** — le cas central

Ne reste pas abstrait. Ne réponds pas comme un consultant OCR générique. Pense comme l'architecte d'un système d'ingestion de connaissances éducatives à longue durée de vie.
