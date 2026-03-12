# Benchmark LLM — Révise Mieux

> Framework d'évaluation comparative des modèles LLM pour les cas d'usage Révise Mieux.
> Deux benchmarks indépendants : **OCR** (images → blocs texte) et **IDP** (blocs texte → items structurés).
> Exécutable à la demande via `go run ./cmd/benchmark/`.
> Dernière mise à jour : 2026-03-12.

---

## 1. Objectifs

1. **Comparer objectivement** les modèles LLM sur les tâches spécifiques de Révise Mieux (pas sur des benchmarks académiques génériques).
2. **Décider du meilleur rapport qualité/coût** par cas d'usage.
3. **Détecter les régressions** quand un provider met à jour son modèle.
4. **Réexécuter à la demande** quand un nouveau modèle sort ou quand les prix changent.

### 1.1 Deux benchmarks, deux étapes du pipeline

Le pipeline Révise Mieux comporte deux étapes LLM distinctes qui nécessitent chacune leur propre benchmark :

| Benchmark | Étape pipeline | Entrée | Sortie | Question clé |
|-----------|---------------|--------|--------|--------------|
| **OCR** | Photo → Blocs texte | Images JPEG de cahier | Blocs OCR (`text`, `block_type`, `confidence`) | Quel modèle vision lit le mieux l'écriture manuscrite d'un collégien ? |
| **IDP** | Blocs texte → Items | Blocs OCR (JSON) | Items structurés (`type`, `term`, `keywords`, `notion`) | Quel modèle structure le mieux les connaissances extraites ? |

Le fichier `input.json` de chaque cas de test sert de **charnière** : c'est le **golden output** du benchmark OCR ET l'**input** du benchmark IDP.

---

## 2. Concurrents à évaluer

### 2.1 Modèles retenus pour le benchmark

| Provider | Modèle | Catégorie | Input $/M | Output $/M | Vision | Benchmarks | Pourquoi le tester |
|----------|--------|-----------|-----------|------------|--------|------------|-------------------|
| **Anthropic** | Claude Sonnet 4.6 | Premium | $3.00 | $15.00 | Oui | OCR + IDP | Baseline actuel pour structuration |
| **Anthropic** | Claude Haiku 4.5 | Économique | $1.00 | $5.00 | Oui | OCR + IDP | Baseline actuel pour fidelity/questions |
| **OpenAI** | GPT-4o | Premium | $2.50 | $10.00 | Oui | OCR + IDP | Concurrent direct, vision forte |
| **OpenAI** | GPT-4o Mini | Économique | $0.15 | $0.60 | Oui | OCR + IDP | Alternative économique avec vision |
| **OpenAI** | o3-mini | Raisonnement | $1.10 | $4.40 | Non | IDP | Raisonnement avancé, pas de vision |
| **Google** | Gemini 2.5 Pro | Premium | $1.25 | $10.00 | Oui | OCR + IDP | Multimodal fort, même prix que GPT-4o |
| **Google** | Gemini 2.5 Flash | Économique | $0.15 | $0.60 | Oui | OCR + IDP | Ultra-compétitif en prix, context 1M |
| **Mistral** | Mistral Large | Mid-range | $2.00 | $6.00 | Oui | OCR + IDP | Souveraineté EU, bon ratio qualité/prix |
| **Mistral** | Mistral Small | Économique | $0.10 | $0.30 | Oui | OCR + IDP | Budget EU |
| **DeepSeek** | deepseek-chat | Budget | $0.27 | $1.10 | Non | IDP | 10-20x moins cher, qualité GPT-4 class |
| **DeepSeek** | deepseek-reasoner | Raisonnement | $0.55 | $2.19 | Non | IDP | Raisonnement avancé, prix agressif |

### 2.2 Modèles exclus (et pourquoi)

| Modèle | Raison d'exclusion |
|--------|-------------------|
| Claude Opus 4.6 | Trop cher ($5/$25) pour le gain marginal vs Sonnet sur nos tâches |
| GPT-5.2 / GPT-5.4 | Prix premium ($1.75-$2.50/$14-$20), pas de gain attendu sur de l'extraction structurée |
| GPT-5 Nano | Trop léger (0.05/0.40), risque élevé d'hallucinations |
| Gemini Flash-Lite | Trop basique pour la structuration |
| Llama / open-source self-hosted | Hors scope MVP (infrastructure d'hébergement à gérer) |

### 2.3 Considérations non-techniques

| Critère | Anthropic | OpenAI | Google | Mistral | DeepSeek |
|---------|-----------|--------|--------|---------|----------|
| **RGPD / données mineurs** | US, DPA disponible | US, DPA disponible | US/EU, DPA disponible | **EU (France)** | **Chine** |
| **Data retention** | 0 jours (API) | 0 jours (API) | Variable | 0 jours (API) | 30 jours |
| **SOC 2 / ISO 27001** | Oui | Oui | Oui | Oui | Non |
| **Risque géopolitique** | Faible | Faible | Faible | **Très faible (EU)** | **Élevé** |
| **Batch API** | Oui (-50%) | Oui (-50%) | Oui (-50%) | Non | Non |
| **Prompt caching** | Oui (-90%) | Oui (-50 à 90%) | Oui (-90%) | Non | Oui (-90%) |

> **Note RGPD** : Révise Mieux traite des données de collégiens (mineurs). DeepSeek pose un risque de conformité RGPD élevé (données potentiellement transférées en Chine, retention 30 jours). Mistral est le choix le plus sûr pour la souveraineté des données. Ce critère peut être **éliminatoire** indépendamment des résultats du benchmark.

---

## 3. Indicateurs d'évaluation

### 3.1 Indicateurs de qualité — Benchmark IDP (structuration)

| # | Indicateur | Description | Méthode de mesure | Poids |
|---|-----------|-------------|-------------------|-------|
| Q1 | **Complétude d'extraction** | % d'items attendus effectivement extraits | Comparaison avec la référence humaine (golden items) | 25% |
| Q2 | **Précision de classification** | % d'items avec le bon `ItemType` (KNOWLEDGE vs PROCEDURE vs DOCUMENT) | Match exact avec la référence | 15% |
| Q3 | **Fidélité source** | % de termes extraits réellement présents dans le texte OCR source | Vérification automatique : chaque `Term` doit être traçable dans le texte source | 20% |
| Q4 | **Qualité des keywords** | Pertinence et couverture des keywords extraits | Score Jaccard vs keywords de référence (pondéré par importance) | 10% |
| Q5 | **Anti-hallucination** | % d'items dont le `Term` est absent du texte source (= hallucination) | Détection automatique : recherche du terme dans le texte OCR | 20% |
| Q6 | **Cohérence du groupement en Notions** | Les items sont-ils regroupés de manière pédagogiquement cohérente | Évaluation semi-automatique : nombre de notions dans la fourchette attendue + noms pertinents | 5% |
| Q7 | **Conformité schéma JSON** | La réponse est-elle un JSON valide conforme au schéma `StructurationResult` | Parsing + validation automatique | 5% |

### 3.2 Indicateurs de qualité — Benchmark OCR (extraction vision)

#### Métriques principales (implémentées)

| # | Indicateur | Description | Méthode de mesure | Poids |
|---|-----------|-------------|-------------------|-------|
| O1 | **Détection de blocs** | Ratio entre le nombre de blocs détectés et le nombre attendu | `min(trouvés, attendus) / max(trouvés, attendus)` | 25% |
| O2 | **Précision textuelle** | Fidélité du texte extrait par rapport au golden | Word overlap : fraction des mots significatifs du golden retrouvés dans la sortie | 35% |
| O3 | **Classification des blocs** | % de blocs avec le bon `block_type` (TEXT vs DIAGRAM) | Match exact après appariement des blocs | 15% |
| O4 | **Calibration de la confidence** | Corrélation entre le score `confidence` et la qualité réelle | Corrélation de Pearson entre confidence et word_overlap par bloc | 10% |
| O5 | **Préservation de l'ordre** | Les blocs sont-ils dans l'ordre de lecture du cahier | Score de Kendall tau entre l'ordre golden et l'ordre produit | 10% |
| O6 | **Diacritiques et caractères spéciaux** | Fidélité des accents, cédilles et caractères français | Levenshtein normalisé sur les mots contenant des diacritiques | 5% |

#### O4 — Calibration de la confidence (nouveau)

```
score = pearson_correlation(confidence_scores, actual_quality_scores)

actual_quality = word_overlap(golden_block, produced_block)
                 pour chaque bloc apparié
```

**Pourquoi** : un modèle qui donne systématiquement `confidence: 0.95` même quand il se trompe est **dangereux** pour le pipeline. La confidence sert à décider si un bloc nécessite une validation humaine (HITL). Si elle n'est pas calibrée, le filtrage HITL ne fonctionne pas.

**Interprétation** :
- `> 0.7` : bien calibré — la confidence est utilisable pour le tri HITL
- `0.3 - 0.7` : modérément calibré — utiliser avec prudence
- `< 0.3` : mal calibré — ignorer le champ confidence, appliquer un seuil uniforme

#### O5 — Préservation de l'ordre (nouveau)

```
score = (kendall_tau(ordre_golden, ordre_produit) + 1) / 2

# Normalisé de [-1, 1] à [0, 1]
# kendall_tau mesure la concordance entre deux rankings
```

**Pourquoi** : l'ordre des blocs reflète la structure du cours (titre → définition → exercice). Si l'OCR mélange l'ordre, la structuration IDP en aval produira des regroupements incohérents. Un modèle qui lit de droite à gauche ou mélange les pages est inutilisable.

#### O6 — Diacritiques et caractères spéciaux (nouveau)

```
score = 1 - avg(levenshtein_norm(mot_golden, mot_produit))
         pour chaque mot contenant un diacritique (é, è, ê, ë, à, ç, ô, û, î, ï, ù, œ, æ)

levenshtein_norm(a, b) = levenshtein(a, b) / max(len(a), len(b))
```

**Pourquoi** : les accents sont critiques en français. "élève" ≠ "eleve", "matière" ≠ "matiere". Un OCR qui supprime systématiquement les accents produit du texte dégradé qui peut changer le sens ("ou" vs "où", "a" vs "à", "du" vs "dû"). Les collégiens écrivent souvent les accents de manière ambiguë — le modèle doit faire le bon choix.

**Mots critiques en contexte scolaire** : température, matière, résumé, schéma, théorème, hypothèse, phénomène, expérience, équation, molécule, génétique, géographie, littérature.

> **Note** : Le benchmark OCR est conçu pour des cas de test ayant un dossier `images/`. Les cas de test texte-only (pas de `"has_images": true` dans metadata.json) sont ignorés par le benchmark OCR.

### 3.3 Indicateurs de performance (communs OCR + IDP)

| # | Indicateur | Description | Méthode de mesure |
|---|-----------|-------------|-------------------|
| P1 | **Latence (TTFT)** | Time to first token | Mesuré côté client |
| P2 | **Latence totale** | Temps entre envoi de la requête et réception complète | Mesuré côté client |
| P3 | **Tokens input** | Nombre de tokens consommés en input | Retourné par l'API |
| P4 | **Tokens output** | Nombre de tokens consommés en output | Retourné par l'API |
| P5 | **Coût par appel** | Coût calculé : tokens × prix/token du modèle | Calculé |
| P6 | **Coût par item extrait** | Coût total / nombre d'items extraits (IDP uniquement) | Calculé |
| P7 | **Taux d'erreur** | % de requêtes en erreur (timeout, 429, 500, JSON invalide) | Compté |

### 3.4 Score composite

Le **score global** combine qualité et coût (même formule pour les deux benchmarks) :

```
Score = (Q_weighted × 0.7) + (Cost_efficiency × 0.3)

Q_weighted = Σ(Qi × poids_i)   // pondéré selon §3.1 (IDP) ou §3.2 (OCR)
Cost_efficiency = 1 - (coût_modèle / coût_modèle_le_plus_cher)
```

Ce score permet de comparer directement les modèles sur un axe unique tout en conservant la traçabilité des sous-scores.

---

## 4. Dataset de benchmark

### 4.1 Golden inputs

10 cas de test couvrant la diversité des matières et des difficultés :

| # | Matière | Niveau | Contenu | Difficulté | Items attendus |
|---|---------|--------|---------|-----------|----------------|
| 1 | Physique-Chimie | 5e | Densité et masse volumique | Moyen | 8 (6K + 2P) |
| 2 | Physique-Chimie | 4e | Circuit électrique (schéma + formules) | Élevé | 10 (5K + 3P + 2D) |
| 3 | SVT | 5e | Cellule animale vs végétale | Moyen | 7 (6K + 1D) |
| 4 | SVT | 3e | Génétique et chromosomes | Élevé | 12 (9K + 2P + 1D) |
| 5 | Histoire | 4e | Révolution française (dates + concepts) | Moyen | 9 (8K + 1D) |
| 6 | Géographie | 3e | Urbanisation mondiale (données + carte) | Élevé | 8 (5K + 1P + 2D) |
| 7 | Mathématiques | 4e | Théorème de Pythagore | Moyen | 6 (2K + 4P) |
| 8 | Mathématiques | 3e | Fonctions affines (graphiques + formules) | Élevé | 10 (3K + 5P + 2D) |
| 9 | Français | 5e | Conjugaison passé simple | Faible | 5 (2K + 3P) |
| 10 | Français | 3e | Figures de style | Moyen | 8 (7K + 1P) |

> K = KNOWLEDGE, P = PROCEDURE, D = DOCUMENT

### 4.2 Référence humaine (golden output)

Chaque cas de test a une **référence humaine validée** (`golden_output.json`) contenant :
- Les items attendus avec leur type, terme, keywords et steps
- Les notions attendues
- Le mapping item → notion

Cette référence est créée manuellement une fois, puis utilisée comme base de comparaison pour tous les modèles.

### 4.3 Stockage

```
backend/testdata/benchmark/
├── cases/
│   ├── 01_physique_densite/         # IDP only (pas d'images)
│   │   ├── metadata.json           # Matière, niveau, difficulté
│   │   ├── input.json              # OCR blocks — input IDP + golden OCR
│   │   └── golden_output.json      # Référence humaine — golden IDP
│   ├── 10_SVT_cours_louis/          # OCR + IDP (avec images)
│   │   ├── metadata.json           # ... + "has_images": true
│   │   ├── images/                 # Input OCR benchmark
│   │   │   ├── IMG_3167.jpeg
│   │   │   ├── IMG_3168.jpeg
│   │   │   └── IMG_3171.jpeg
│   │   ├── input.json              # Golden OCR + Input IDP (charnière)
│   │   └── golden_output.json      # Golden IDP
│   └── ...
├── results/
│   ├── idp/                        # Résultats IDP par run
│   │   └── 2026-03-12_14h30/
│   │       ├── summary.json
│   │       └── report.html
│   └── ocr/                        # Résultats OCR par run
│       └── 2026-03-12_14h30/
│           └── summary.json
└── README.md
```

> **`input.json` est la charnière** entre les deux benchmarks : c'est le golden output du benchmark OCR ET l'input du benchmark IDP. Cela garantit que les deux benchmarks sont évalués sur les mêmes données de référence.

---

## 5. Architecture du benchmark runner

### 5.1 Vue d'ensemble

Le runner supporte deux modes via le flag `--type` :

```
cmd/benchmark/main.go --type=idp
    │
    ├── Charge les cas de test (testdata/benchmark/cases/)
    ├── Pour chaque IDP provider configuré :
    │   ├── Pour chaque cas de test :
    │   │   ├── Envoie les blocs OCR (input.json) au LLM
    │   │   ├── Mesure latence, tokens, coût
    │   │   ├── Parse la réponse JSON (items structurés)
    │   │   └── Évalue vs golden_output.json (Q1-Q7)
    │   └── Agrège les résultats
    ├── Calcule les scores composites
    └── Sauvegarde dans results/idp/

cmd/benchmark/main.go --type=ocr
    │
    ├── Charge les cas de test avec images ("has_images": true)
    ├── Pour chaque OCR provider configuré (vision) :
    │   ├── Pour chaque cas de test :
    │   │   ├── Envoie les images (images/*.jpeg) au LLM vision
    │   │   ├── Mesure latence, tokens, coût
    │   │   ├── Parse la réponse JSON (blocs OCR)
    │   │   └── Évalue vs input.json (O1-O6)
    │   └── Agrège les résultats
    ├── Calcule les scores composites
    └── Sauvegarde dans results/ocr/
```

### 5.2 Interfaces provider

```go
// Provider abstracts an LLM API for IDP benchmarking (structuration).
type Provider interface {
    Name() string
    ModelID() string
    PricePerMInput() float64
    PricePerMOutput() float64
    StructureBlocks(ctx context.Context, systemPrompt, userPrompt string) (*Response, error)
}

// OCRProvider abstracts an OCR/vision API for OCR benchmarking.
type OCRProvider interface {
    Name() string
    ModelID() string
    PricePerMInput() float64
    PricePerMOutput() float64
    ExtractBlocks(ctx context.Context, imagePaths []string, subject string) (*Response, error)
}

type Response struct {
    RawJSON      []byte
    TokensInput  int
    TokensOutput int
    LatencyMs    int64
    ModelVersion string
}
```

Chaque provider (Anthropic, OpenAI, Google, Mistral, DeepSeek) implémente `Provider` pour l'IDP. Ceux qui supportent la vision implémentent aussi `OCRProvider`.

### 5.3 Évaluation automatique

```go
type EvalResult struct {
    CaseID              string
    Provider            string
    Model               string

    // Qualité
    CompletenessScore   float64  // Q1: items trouvés / items attendus
    ClassificationScore float64  // Q2: types corrects / total items
    FidelityScore       float64  // Q3: termes traçables dans source
    KeywordScore        float64  // Q4: Jaccard keywords
    HallucinationRate   float64  // Q5: termes inventés / total items
    NotionScore         float64  // Q6: cohérence groupement
    SchemaCompliance    bool     // Q7: JSON valide + schéma respecté

    // Performance
    LatencyMs           int64    // P2
    TokensInput         int      // P3
    TokensOutput        int      // P4
    CostUSD             float64  // P5
    CostPerItem         float64  // P6

    // Composite
    QualityScore        float64  // Pondéré
    CompositeScore      float64  // Qualité + coût
}
```

### 5.4 Rapport de sortie

Le benchmark produit :

1. **Console** : tableau comparatif trié par score composite
2. **JSON** : résultats détaillés par cas de test et par provider
3. **CSV** : export pour analyse dans un tableur

Exemple de sortie console :

```
╔══════════════════════════╦═══════╦═══════╦═══════╦═══════╦════════╦═══════════╗
║ Modèle                  ║ Q1    ║ Q3    ║ Q5    ║ Coût  ║ Lat.   ║ Score     ║
║                          ║ Compl.║ Fidél.║ Hallu.║ $/item║ ms     ║ composite ║
╠══════════════════════════╬═══════╬═══════╬═══════╬═══════╬════════╬═══════════╣
║ Claude Sonnet 4.6        ║ 0.92  ║ 0.95  ║ 0.02  ║ 0.002 ║ 3200   ║ 0.87      ║
║ GPT-5                    ║ 0.89  ║ 0.93  ║ 0.04  ║ 0.001 ║ 2800   ║ 0.86      ║
║ Gemini 2.5 Pro           ║ 0.88  ║ 0.91  ║ 0.05  ║ 0.001 ║ 2500   ║ 0.85      ║
║ Mistral Medium 3         ║ 0.83  ║ 0.88  ║ 0.07  ║ 0.0004║ 2100   ║ 0.84      ║
║ Gemini 2.5 Flash         ║ 0.81  ║ 0.86  ║ 0.08  ║ 0.0003║ 1500   ║ 0.83      ║
║ DeepSeek V3.2            ║ 0.80  ║ 0.85  ║ 0.09  ║ 0.0001║ 3500   ║ 0.82      ║
║ GPT-5 Mini               ║ 0.78  ║ 0.84  ║ 0.10  ║ 0.0002║ 1800   ║ 0.79      ║
║ Claude Haiku 4.5         ║ 0.76  ║ 0.82  ║ 0.11  ║ 0.0001║ 1200   ║ 0.78      ║
║ DeepSeek R1              ║ 0.85  ║ 0.90  ║ 0.06  ║ 0.0003║ 8000   ║ 0.77      ║
╚══════════════════════════╩═══════╩═══════╩═══════╩═══════╩════════╩═══════════╝
```

*(valeurs fictives à titre d'illustration)*

---

## 6. Exécution

### 6.1 Commande

```bash
# --- Benchmark IDP (structuration LLM) ---
# Tous les modèles
go run ./cmd/benchmark/ --type=idp --all
make bench-idp ALL=1

# Un seul provider
go run ./cmd/benchmark/ --type=idp --provider=anthropic
make bench-idp PROVIDER=anthropic

# Comparer deux modèles spécifiques
go run ./cmd/benchmark/ --type=idp --models=claude-sonnet-4-6,gpt-4o

# Un seul cas de test
go run ./cmd/benchmark/ --type=idp --case=01_physique_densite

# --- Benchmark OCR (extraction vision) ---
# Tous les modèles vision
go run ./cmd/benchmark/ --type=ocr --all
make bench-ocr ALL=1

# Un seul provider
go run ./cmd/benchmark/ --type=ocr --provider=anthropic
make bench-ocr PROVIDER=anthropic

# --- Options communes ---
# N répétitions (pour mesurer la variance)
go run ./cmd/benchmark/ --type=idp --all --runs=3

# Export CSV
go run ./cmd/benchmark/ --type=idp --all --output=csv

# Rapport HTML (depuis les derniers résultats IDP)
go run ./cmd/benchmark/ --report
make bench-report

# Lister les modèles disponibles (indique [vision] pour les modèles OCR)
go run ./cmd/benchmark/ --list-models
make bench-models
```

> **Note** : `--type=idp` est le défaut. `make bench` est un alias pour `make bench-idp`.

### 6.2 Variables d'environnement

```bash
ANTHROPIC_API_KEY=sk-ant-xxx
OPENAI_API_KEY=sk-xxx
GOOGLE_AI_API_KEY=xxx
MISTRAL_API_KEY=xxx
DEEPSEEK_API_KEY=xxx
```

Seuls les providers avec une clé configurée sont exécutés. Les autres sont ignorés avec un warning.

### 6.3 Coût estimé d'un run complet

**Benchmark IDP** (10 cas texte) :

| Provider | Cas × runs | Coût estimé |
|----------|-----------|-------------|
| Claude Sonnet 4.6 | 10 × 1 | ~$0.17 |
| Claude Haiku 4.5 | 10 × 1 | ~$0.01 |
| GPT-4o | 10 × 1 | ~$0.07 |
| GPT-4o Mini | 10 × 1 | ~$0.01 |
| Gemini 2.5 Pro | 10 × 1 | ~$0.07 |
| Gemini 2.5 Flash | 10 × 1 | ~$0.02 |
| Mistral Large | 10 × 1 | ~$0.02 |
| DeepSeek chat | 10 × 1 | ~$0.005 |
| **Total IDP** | | **~$0.40** |

**Benchmark OCR** (1 cas avec 3 images de cahier ~1500×2000px) :

> Voir §11 pour le détail de la tokenisation images par provider.

| Provider | Cas × runs | Tokens input estimés | Coût estimé |
|----------|-----------|---------------------|-------------|
| Claude Sonnet 4.6 | 1 × 1 | ~5 300 | ~$0.022 |
| Claude Haiku 4.5 | 1 × 1 | ~5 300 | ~$0.008 |
| GPT-4o (high detail) | 1 × 1 | ~2 795 | ~$0.012 |
| GPT-4o Mini (high detail) | 1 × 1 | ~2 795 | ~$0.001 |
| Gemini 2.5 Pro | 1 × 1 | ~5 144 | ~$0.011 |
| Gemini 2.5 Flash | 1 × 1 | ~5 144 | ~$0.002 |
| Mistral Pixtral Large | 1 × 1 | ~9 911 | ~$0.023 |
| **Total OCR** | | | **~$0.08** |

Un run complet des deux benchmarks coûte ~$0.48. Avec 3 répétitions : ~$1.44.

> **Note** : le coût OCR varie d'un facteur ~20× entre providers. Il augmentera linéairement avec le nombre de cas images ajoutés (voir §13). Voir §11 pour le détail de la tokenisation.

---

## 7. Métriques détaillées — Comment elles sont calculées

### Benchmark OCR

#### O1 — Détection de blocs

```
score = min(blocs_trouvés, blocs_attendus) / max(blocs_trouvés, blocs_attendus)
```

**Pourquoi** : un modèle qui fusionne 5 blocs en 2, ou qui en segmente 2 en 10, perd de l'information structurelle. Le ratio symétrique pénalise à la fois la sur-segmentation et la sous-segmentation.

**Cas limites** :
- Un modèle qui retourne 0 blocs → score = 0 (échec complet)
- Un modèle qui retourne exactement le bon nombre → score = 1 (mais le contenu peut être faux)
- Un modèle qui retourne 10 blocs pour 5 attendus → score = 0.5 (sur-segmentation)

#### O2 — Précision textuelle

```
score = avg(word_overlap(golden_block, produced_block))
        pour chaque paire de blocs appariés

word_overlap(golden, produced) = mots_significatifs_retrouvés / mots_significatifs_golden
```

**Pourquoi** : c'est la métrique la plus critique du benchmark OCR. Si le texte extrait ne correspond pas au texte réel du cahier, tout le pipeline aval (structuration, questions) sera faussé. L'OCR doit lire fidèlement l'écriture manuscrite d'un collégien, y compris les fautes d'orthographe.

**Algorithme d'appariement** : le matching des blocs golden/produits utilise un algorithme glouton. Pour chaque bloc golden, on trouve le bloc produit non encore utilisé qui maximise le word_overlap (seuil minimum : 0.2). Ce n'est pas optimal (un algorithme hongrois donnerait le matching parfait), mais c'est suffisant pour nos cas avec 3-10 blocs.

**Mots significatifs** : les stop words français sont retirés avant le calcul (le, la, les, de, du, des, un, une, et, en, dans, pour, sur, avec, par, est, sont, qui, que, ce, se, ne, pas, plus, tout, cette, ces, son, sa, ses, leur, leurs, autre, même, aussi, très, bien, peu, trop, mais, ou, donc, car, comme, quand, si). Le texte est normalisé en minuscules.

#### O3 — Classification des blocs

```
score = blocs_correctement_typés / blocs_appariés

correctly_typed = le bloc apparié a le même block_type (TEXT ou DIAGRAM)
                  que son correspondant dans le golden
```

**Pourquoi** : distinguer un schéma (DIAGRAM) d'un texte (TEXT) est essentiel pour le traitement aval. Un schéma mal classifié en texte générera des items de révision absurdes (essayer de structurer la "légende d'un schéma" comme un fait à mémoriser).

**Catégories actuelles** : `TEXT`, `DIAGRAM`. Extension future possible : `TABLE`, `FORMULA`, `HEADER`.

#### O4 — Calibration de la confidence

```
score = max(0, pearson_r(confidence[], word_overlap[]))

# pearson_r négatif ou NaN → score = 0
# Nécessite au moins 3 blocs appariés pour être significatif
```

**Pourquoi** : le champ `confidence` retourné par le modèle doit être fiable pour alimenter le filtre HITL. Un modèle "surconfiant" (confidence toujours à 0.9 même sur du texte illisible) ne permet pas de prioriser la validation humaine.

**Seuils d'action** :
- Pearson > 0.7 → la confidence est fiable, on peut l'utiliser pour le seuil HITL
- Pearson 0.3-0.7 → ajuster le seuil HITL conservativement (0.5 au lieu de 0.7)
- Pearson < 0.3 → ignorer la confidence, envoyer tout en HITL

#### O5 — Préservation de l'ordre

```
score = (kendall_tau + 1) / 2

# kendall_tau(golden_order, produced_order) ∈ [-1, 1]
# Normalisé en [0, 1] : 1.0 = ordre parfait, 0.5 = aléatoire, 0.0 = inversé
```

**Pourquoi** : l'ordre des blocs encode la structure pédagogique du cours (titre → contenu → exercice → correction). La structuration IDP en aval s'appuie sur cette séquence pour regrouper les items en notions cohérentes.

**Calcul** : on compare l'indice de chaque bloc apparié dans l'ordre golden vs l'ordre produit. Un bloc golden[0] matché avec produced[2] et golden[1] matché avec produced[0] constitue une inversion.

#### O6 — Diacritiques et caractères spéciaux

```
score = 1 - avg(levenshtein_norm(mot_golden, mot_produit))
         restreint aux mots du golden contenant ≥1 diacritique

# Si aucun mot avec diacritique → score = 1.0 (pas applicable)
```

**Pourquoi** : en français scolaire, les accents sont omniprésents et porteurs de sens. Un OCR qui les perd systématiquement dégrade la qualité perçue et peut causer des erreurs de compréhension pour l'élève.

---

### Benchmark IDP

### Q1 — Complétude d'extraction

```
score = items_matched / items_expected

items_matched = nombre d'items de la golden output
                qui ont un match dans la sortie LLM
                (match = Jaccard(term_golden, term_llm) > 0.6
                 OU substring match insensible à la casse)
```

**Pourquoi c'est important** : un modèle qui n'extrait que 4 items sur 8 manque la moitié du cours. L'élève ne sera jamais interrogé sur ce qu'il manque.

### Q2 — Précision de classification

```
score = items_correctly_typed / items_matched

correctly_typed = l'item matché a le même ItemType
                  que son correspondant dans la golden output
```

**Pourquoi** : classifier une formule (PROCEDURE) en fait (KNOWLEDGE) produit des questions inadaptées. "Récite ρ = m/V" au lieu de "Calcule la masse volumique".

### Q3 — Fidélité source

```
score = items_traceable / total_items

traceable = le Term de l'item (ou ses mots clés principaux)
            se retrouvent dans le texte OCR source
            (fuzzy match, tolérance accents + casse)
```

**Pourquoi** : l'indicateur le plus critique. Un item "fidèle" enseigne ce que le prof a écrit. Un item infidèle enseigne autre chose.

### Q5 — Anti-hallucination

```
rate = items_hallucinated / total_items

hallucinated = aucun mot significatif du Term
               ne se retrouve dans le texte OCR source
               (après normalisation + stop words removal)
```

**Pourquoi** : une hallucination est un item faux enseigné à un collégien. C'est le pire scénario. Score idéal = 0%.

### Q4 — Qualité des keywords

```
score = avg(Jaccard(keywords_golden, keywords_llm))
        pour chaque item matché

Jaccard(A, B) = |A ∩ B| / |A ∪ B|
```

**Pourquoi** : les keywords servent à générer les questions (cloze, MCQ). Des keywords imprécis → des questions imprécises.

### P6 — Coût par item extrait

```
cost_per_item = total_cost / items_extracted

total_cost = (tokens_input × price_input / 1M)
           + (tokens_output × price_output / 1M)
```

**Pourquoi** : le vrai indicateur économique. Un modèle 3x plus cher mais qui extrait 2x plus d'items est plus rentable par item.

---

## 8. Scénarios de décision

Le benchmark ne donne pas un gagnant absolu. Il alimente des **décisions contextuelles** :

| Scénario | Critère de décision | Action |
|----------|-------------------|--------|
| Modèle A a Q1 > 0.90 et coûte 10x moins que le baseline | Qualité suffisante, économie massive | Migrer vers A |
| Modèle A a Q5 > 0.15 (hallucinations) | Disqualifié | Ne pas utiliser pour la structuration |
| Modèle A a Q3 < 0.80 (fidélité basse) | Risque pour les élèves | Exclure ou restreindre aux fidelity-checked items |
| Tous les modèles se valent en qualité | Décision sur le coût + RGPD | Choisir le moins cher conforme |
| Nouveau modèle sorti | Relancer le benchmark | `go run ./cmd/benchmark/ --models=new_model` |
| Provider augmente ses prix | Relancer le benchmark | Vérifier si un switch est rentable |
| Résultats très variables entre runs | Modèle instable | Augmenter `--runs=5`, vérifier `temperature=0` |
| **OCR** : Modèle A a O2 < 0.70 (texte) | OCR inutilisable seul | Tester l'approche hybride (OCR dédié + LLM) — voir §12.3 |
| **OCR** : Modèle A a O4 < 0.30 (confidence) | Confidence non fiable | Envoyer tout en HITL, ignorer le score de confidence |
| **OCR** : Google Vision ≈ LLM Vision en O2 | OCR dédié aussi bon | Migrer vers Google Vision (100x moins cher) — voir §11 |
| **OCR** : Manuscrit brouillon < 0.50 en O2 | Cas dégradé | Proposer la saisie manuelle comme fallback — voir §14 |
| **OCR** : Coût OCR > 50% du pipeline total | Budget déséquilibré | Utiliser un modèle flash pour l'OCR, premium pour l'IDP |

---

## 9. Fréquence d'exécution recommandée

| Événement | Action |
|-----------|--------|
| Nouveau modèle majeur (ex: GPT-6, Claude 5) | Run complet |
| Mise à jour de modèle existant (ex: Sonnet 4.6 → 4.7) | Run sur le provider concerné |
| Changement de prompt | Run complet (régression prompt) |
| Changement de prix | Recalcul des scores composites (pas besoin de re-run) |
| Trimestriel | Run complet de routine |
| Pré-production | Run complet avant chaque release majeure |

---

## 10. Deep dive OCR — Défis spécifiques aux cahiers de collégiens

### 10.1 Caractéristiques du contenu à traiter

Les photos de cahiers de collégiens présentent des défis uniques qui les distinguent radicalement des benchmarks OCR académiques (IAM Handwriting, MNIST, etc.) :

| Défi | Description | Impact sur l'OCR |
|------|-------------|-----------------|
| **Écriture manuscrite adolescente** | Écriture en cours de maturation, souvent rapide, parfois brouillonne. Mélange de cursive et script. | Taux d'erreur caractère (CER) typiquement 15-30% vs 2-5% sur du texte imprimé |
| **Contenu bilingue/mixte** | Termes scientifiques (latin, grec), formules mathématiques, noms propres | Le modèle doit gérer du vocabulaire hors distribution |
| **Multi-encre** | Stylo bleu + rouge + crayon + surligneur. Corrections au blanc/barrés | Certaines encres sont peu contrastées, le surlignage masque le texte |
| **Mélange imprimé/manuscrit** | Polycopiés collés + annotations manuscrites de l'élève | Le modèle doit traiter les deux modes dans la même image |
| **Schémas annotés** | Schémas biologiques, circuits électriques, cartes avec légendes manuscrites | Frontière floue entre TEXT et DIAGRAM |
| **Qualité photo variable** | Pris au téléphone, angle variable, éclairage non contrôlé, ombres, doigts | Flou, distorsion perspective, luminosité non uniforme |
| **Pages multiples** | Un chapitre = 2-6 pages de cahier, parfois recto-verso | Cohérence cross-page, risque de duplication |

### 10.2 Matrice de difficulté par matière

| Matière | Difficulté OCR | Raison principale |
|---------|---------------|-------------------|
| SVT | Élevée | Schémas biologiques complexes + vocabulaire latin |
| Physique-Chimie | Élevée | Formules (ρ = m/V), unités, schémas de circuits |
| Mathématiques | Très élevée | Formules, symboles (√, ∑, ∫, ≤), graphiques, fractions |
| Histoire-Géographie | Moyenne | Texte principalement, quelques cartes et frises |
| Français | Faible-Moyenne | Texte majoritaire, citations, conjugaisons tabulaires |

### 10.3 Modes de dégradation observés

D'après l'analyse de notre cas pilote (10_SVT_cours_louis, 3 pages), les modèles LLM vision présentent les dégradations suivantes :

1. **Hallucination de mots** : le modèle "devine" un mot illisible au lieu de signaler une faible confidence
2. **Fusion de blocs** : deux sections distinctes fusionnées en un seul bloc (perte de structure)
3. **Omission sélective** : les petites annotations marginales (corrections, flèches) sont ignorées
4. **Normalisation excessive** : le modèle corrige l'orthographe de l'élève au lieu de la transcrire fidèlement
5. **Confusion schéma/texte** : une légende de schéma est extraite comme texte sans le contexte visuel

> **Risque critique** : la dégradation n°4 (normalisation) est la plus insidieuse. Le prompt OCR dit explicitement "extraire FIDÈLEMENT y compris les fautes", mais certains modèles ont un biais fort vers la correction. Cela fausse ensuite le score de fidélité Q3 du benchmark IDP.

---

## 11. Tokenisation des images par provider

### 11.1 Comment chaque provider facture les images

La tokenisation des images varie d'un provider à l'autre. Cela impacte directement le coût OCR, qui est dominé par les tokens d'entrée (images).

| Provider | Méthode de tokenisation | Formule | Tokens pour 1 photo cahier (~1500×2000px) | Coût input/image |
|----------|------------------------|---------|------------------------------------------|-----------------|
| **Anthropic** | Pixel area / 750, auto-downscale si > 1568px long edge. | `(w×h)/750` après resize | **~1 600 tokens** (résolution réduite à ~951×1268) | $0.0048 (Sonnet) / $0.0016 (Haiku) |
| **OpenAI** | `detail: high` = tiles 512×512 + overhead. Shortest side → 768px. | `85 + 170 × tiles` | **~765 tokens** (4 tiles après resize 768×1024) | $0.0019 (4o) / $0.0001 (4o-mini) |
| **Google** | Tiles 768×768, 258 tokens/tile. < 384px = 1 tile. | `258 × tiles` | **~1 548 tokens** (6 tiles) | $0.0019 (Pro) / $0.0005 (Flash) |
| **Mistral** | Pixtral : patches 16×16 à résolution native (max 1024 long edge). | `(w/16)×(h/16) + h/16 + 1` | **~3 137 tokens** (après resize 768×1024) | $0.0063 (Pixtral Large) |

### 11.2 Impact sur le coût OCR (3 images de cahier)

| Provider | Modèle | Tokens input (3 img + prompt) | Coût input | Tokens output (~500) | Coût output | **Total** |
|----------|--------|------------------------------|-----------|---------------------|------------|-----------|
| Anthropic | Sonnet 4.6 | ~5 300 | $0.016 | 500 | $0.008 | **$0.022** |
| Anthropic | Haiku 4.5 | ~5 300 | $0.005 | 500 | $0.003 | **$0.008** |
| OpenAI | GPT-4o (high) | ~2 795 | $0.007 | 500 | $0.005 | **$0.012** |
| OpenAI | GPT-4o Mini (high) | ~2 795 | $0.0004 | 500 | $0.0003 | **$0.001** |
| Google | Gemini 2.5 Pro | ~5 144 | $0.006 | 500 | $0.005 | **$0.011** |
| Google | Gemini 2.5 Flash | ~5 144 | $0.002 | 500 | $0.0003 | **$0.002** |
| Mistral | Pixtral Large | ~9 911 | $0.020 | 500 | $0.003 | **$0.023** |

> **Constat** : le coût OCR varie d'un facteur **~20x** entre GPT-4o Mini ($0.001) et Mistral Pixtral Large ($0.023) pour les mêmes 3 images. Les modèles économiques (GPT-4o Mini, Gemini Flash) sont très compétitifs — le benchmark doit déterminer si la qualité suit.
>
> **Prompt caching** : Anthropic propose un cache à 0.1× le coût input. Si les mêmes images sont analysées plusieurs fois (runs, itérations), le coût tombe à ~$0.004 par appel sur Sonnet — compétitif avec les modèles flash.

### 11.3 Projections à l'échelle

Pour un utilisateur typique (4 chapitres × 4 pages × 3 images/page = 48 images, traitées en 16 appels de 3 images) :

| Provider | Modèle | Coût pipeline OCR complet | Avec prompt caching |
|----------|--------|--------------------------|---------------------|
| Anthropic | Sonnet 4.6 | ~$0.35 | ~$0.06 (cache 0.1×) |
| Anthropic | Haiku 4.5 | ~$0.13 | ~$0.02 (cache 0.1×) |
| OpenAI | GPT-4o | ~$0.19 | ~$0.10 (cache 0.5×) |
| OpenAI | GPT-4o Mini | ~$0.01 | ~$0.007 |
| Google | Gemini 2.5 Pro | ~$0.18 | ~$0.03 (cache 0.1×) |
| Google | Gemini 2.5 Flash | ~$0.03 | ~$0.006 |
| Mistral | Pixtral Large | ~$0.37 | N/A (pas de cache) |

> **Conclusion** : le coût OCR est raisonnable pour tous les providers sur le volume MVP (48 images, < $0.40). L'écart se creuse à l'échelle (10K pages/mois) : $48/mois sur Sonnet vs $1/mois sur GPT-4o Mini. Le prompt caching réduit significativement l'écart entre les modèles premium et économiques.

---

## 12. Alternatives : services OCR dédiés vs LLM vision

### 12.1 Vue d'ensemble

L'approche actuelle utilise des **LLM vision** (modèles généralistes avec capacités vision) pour l'OCR. Il existe des **services OCR dédiés** potentiellement plus adaptés :

| Service | Type | Pricing | Manuscrit | Français | RGPD | Output structuré |
|---------|------|---------|-----------|----------|------|-----------------|
| **Google Cloud Vision** | API OCR cloud | $1.50/1000 pages | Oui (modéré) | Oui | US/EU, DPA | Blocs + positions |
| **Google Document AI** | OCR enterprise | $1.50/1000 pages (OCR), $30/1000 (forms) | Oui (avancé, 50 langues) | Oui | US/EU | Blocs + layout + formules + **Math OCR (LaTeX)** |
| ~~**AWS Textract**~~ | ~~API OCR cloud~~ | ~~$1.50/1000 pages~~ | ~~Manuscrit anglais only~~ | ~~Non (manuscrit)~~ | ~~US/EU, DPA~~ | ~~Blocs + tables + forms~~ |
| **Azure Doc Intelligence** | API OCR cloud | $1.50/1000 pages | Oui (avancé) | Oui | US/EU/Global, DPA | Blocs + styles + positions |
| **Apple Vision** | On-device (iOS) | Gratuit | Oui (modéré) | Oui | On-device (RGPD ++) | Texte + bounding boxes |
| **LLM Vision** (actuel) | LLM généraliste | $0.001-$0.10/page | Oui (variable) | Oui | Selon provider | JSON structuré (prompt-dépendant) |

> **AWS Textract éliminé** : la reconnaissance manuscrite de Textract est **anglais uniquement**. Les 5 autres langues supportées ne couvrent que le texte imprimé. C'est disqualifiant pour notre cas d'usage (manuscrit français de collégiens).

### 12.2 Comparaison qualitative

| Critère | OCR dédié (Google/Azure) | LLM Vision | Gagnant |
|---------|-------------------------|------------|---------|
| **Précision texte imprimé** | 98-99% | 95-98% | OCR dédié |
| **Précision manuscrit propre** | 85-92% | 80-95% | Variable |
| **Précision manuscrit brouillon** | 60-75% | 70-90% | **LLM Vision** |
| **Compréhension contextuelle** | Aucune | Forte (corrige par contexte) | **LLM Vision** |
| **Détection de structure** | Blocs géométriques | Blocs sémantiques | **LLM Vision** |
| **Classification TEXT/DIAGRAM** | Limitée (heuristiques) | Native (comprend le contenu) | **LLM Vision** |
| **Latence** | 1-3s | 3-15s | OCR dédié |
| **Coût par page** | ~$0.0015 | $0.001-$0.02 | Variable |
| **Stabilité output** | Très stable | Variable (température) | OCR dédié |
| **Bounding boxes** | Oui (pixel-level) | Non | OCR dédié |

> **Données benchmark 2025** : les LLM vision **surpassent** désormais les OCR dédiés sur le manuscrit. Gemini 3 Pro atteint ~100% sur des benchmarks de cursive ; les OCR traditionnels plafonnent à ~64% sur le manuscrit (source : arXiv:2510.10138). L'avantage LLM vient de la compréhension contextuelle : un LLM utilise le contexte des mots voisins pour résoudre les lettres ambiguës, ce qu'un OCR dédié ne peut pas faire.

### 12.3 Approche hybride : OCR dédié + LLM structuration

```
┌──────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Photo   │────►│  OCR dédié   │────►│  Post-traite-│────►│  LLM IDP     │
│  cahier  │     │  (Cloud      │     │  ment (merge, │     │  (structura- │
│          │     │   Vision)    │     │  clean)       │     │  tion)       │
└──────────┘     └──────────────┘     └──────────────┘     └──────────────┘
                  ~$0.0015/page        code local           ~$0.01-0.05/page
                  ~2s                  ~50ms                 ~3-10s
```

**Avantages** :
- Coût OCR quasi nul ($0.0015 vs $0.01-0.10 par page en LLM)
- Latence OCR réduite (2s vs 5-15s)
- Bounding boxes disponibles (utile pour surligner l'image dans l'app)
- Plus stable (pas de variabilité liée à la température LLM)

**Inconvénients** :
- Perte de la compréhension contextuelle (l'OCR dédié ne comprend pas que "ρ = m/V" est une formule de physique)
- Pas de classification TEXT/DIAGRAM intelligente
- Nécessite un post-traitement pour fusionner les blocs géométriques en blocs sémantiques
- Une dépendance de plus (Google Vision + LLM au lieu de LLM seul)
- Le manuscrit brouillon est souvent mieux lu par les LLM (qui devinent par contexte)

### 12.4 Approche on-device : Apple Vision + LLM

```
┌──────────┐     ┌──────────────┐     ┌──────────────┐
│  Photo   │────►│  Apple Vision│────►│  API backend  │
│  (prise  │     │  (on-device) │     │  LLM IDP     │
│  sur iOS)│     │  OCR gratuit │     │  (structura-  │
│          │     │              │     │  tion)         │
└──────────┘     └──────────────┘     └──────────────┘
                  gratuit, ~1s          ~$0.01-0.05
                  RGPD parfait          ~3-10s
```

**Avantages** :
- **Gratuit** et **zéro donnée transmise** pour l'étape OCR
- Latence inférieure à 1s
- Conformité RGPD maximale (aucune image envoyée à un service tiers)
- Plugins React Native/Expo existants (`@bear-block/vision-camera-ocr`, `expo-ocr`)

**Inconvénients** :
- Qualité manuscrit inférieure aux LLM vision (pas de compréhension contextuelle)
- iOS only (Android utilise ML Kit, qualité différente)
- Pas de contrôle sur le format de sortie (texte brut, pas de JSON structuré)
- Pas de classification TEXT/DIAGRAM
- Pas de champ `confidence` granulaire

### 12.5 Recommandation : intégrer les OCR dédiés dans le benchmark

**Décision** : ajouter au minimum **Google Cloud Vision** et **Apple Vision (on-device)** comme providers OCR dans le benchmark, à côté des LLM vision.

Cela permettra de mesurer objectivement :
1. Le delta de qualité LLM vs OCR dédié sur nos images réelles
2. Le ratio coût/qualité pour décider si le surcoût LLM est justifié
3. La viabilité de l'approche hybride (OCR dédié pour le texte + LLM pour la classification)

**Implémentation** : créer un `CloudVisionOCRProvider` et un `AppleVisionOCRProvider` (mock pour les tests serveur) implémentant `benchmark.OCRProvider`.

### 12.6 Recommandation architecturale

Étant donné le cas d'usage (manuscrit français de collégiens, contenu mixte texte/schéma/formules) :

| Option | Architecture | Lot 0 (MVP) | À l'échelle | Risque |
|--------|-------------|-------------|-------------|--------|
| **A — LLM Vision pur** (actuel) | Photo → LLM Vision OCR → LLM IDP | **Recommandé** : simple, meilleur manuscrit | Coût élevé ($0.35-1.50/chapitre) | Pipeline simple, un seul point de défaillance |
| **B — Hybride OCR+LLM** | Photo → Google Document AI → LLM IDP | Complexité non justifiée | Intéressant si qualité manuscrit OCR suffisante | Pipeline double, perte compréhension contextuelle |
| **C — On-device + cloud** | Photo → Apple Vision (preview) → Cloud LLM | UX enhancement possible | Indépendant du choix backend | iOS only, qualité insuffisante seule |

**Décision Lot 0** : **Option A** (LLM Vision pur). Le benchmark OCR validera ce choix. Si aucun LLM ne passe les seuils MVP (§14.4), basculer vers l'option B.

**Bonus Option C** : indépendamment du choix backend, Apple Vision peut servir de **preview en temps réel** pendant la prise de photo (overlay du texte détecté), sans coût ni transfert de données. C'est un enhancement UX à considérer pour le Lot 1.

---

## 13. Stratégie d'expansion du dataset OCR

### 13.1 État actuel : 1 cas de test avec images

Le benchmark OCR ne dispose actuellement que d'**un seul cas de test avec images** (10_SVT_cours_louis, 3 photos de cahier). C'est insuffisant pour tirer des conclusions fiables.

### 13.2 Plan de couverture cible

| Phase | Cas | Matières couvertes | Type de contenu | Objectif |
|-------|-----|-------------------|-----------------|----------|
| **Phase 1** (actuel) | 1 cas, 3 images | SVT | Manuscrit + polycopié + schéma | Preuve de concept |
| **Phase 2** (court terme) | 5 cas, ~15 images | SVT, Physique, Maths | Variété de contenus | Première comparaison fiable |
| **Phase 3** (moyen terme) | 10 cas, ~30 images | Toutes matières | Variété complète | Benchmark représentatif |

### 13.3 Axes de diversité à couvrir

Chaque axe doit être représenté par au moins 2 cas de test :

| Axe | Valeurs à couvrir | Priorité |
|-----|-------------------|----------|
| **Qualité écriture** | Soigné, Normal, Brouillon | Haute |
| **Type d'encre** | Bleu, Noir, Rouge, Crayon, Mixte | Moyenne |
| **Qualité photo** | Bonne (bien éclairé, droit), Moyenne (léger angle), Mauvaise (sombre, flou) | Haute |
| **Contenu** | Texte seul, Texte + schéma, Formules, Tableaux, Cartes | Haute |
| **Support** | Cahier manuscrit pur, Polycopié collé + annotations, Imprimé annoté | Moyenne |
| **Matière** | SVT, Physique-Chimie, Maths, Histoire-Géo, Français | Haute |

### 13.4 Protocole de création d'un cas OCR

1. **Photographier** : 2-4 pages d'un cahier réel avec un smartphone (conditions réalistes)
2. **Annoter manuellement** : créer le `input.json` (golden OCR output) en transcrivant fidèlement chaque bloc
3. **Classifier** : attribuer `block_type` et `confidence` de référence à chaque bloc
4. **Valider** : faire relire par une deuxième personne (cross-validation humaine)
5. **Métadonnées** : renseigner `metadata.json` avec les caractéristiques (qualité écriture, type d'encre, etc.)

### 13.5 Format metadata.json étendu pour les cas OCR

```json
{
  "id": "11_physique_circuit",
  "subject": "Physique-Chimie",
  "level": "4e",
  "topic": "Circuit électrique en série et en dérivation",
  "difficulty": "high",
  "has_images": true,
  "expected_item_count": 10,
  "ocr_metadata": {
    "handwriting_quality": "normal",
    "ink_types": ["blue_pen", "red_pen"],
    "photo_quality": "good",
    "content_types": ["text", "diagram", "formula"],
    "support": "notebook_with_printout",
    "pages": 3,
    "notes": "Mix of handwritten notes and pasted circuit diagrams with handwritten labels"
  }
}
```

### 13.6 Variance et reproductibilité

Contrairement au benchmark IDP (entrée texte déterministe), le benchmark OCR peut présenter de la **variance** même avec `temperature=0` car :
- Les images sont riches en informations ambiguës
- L'encodage base64 peut varier selon l'implémentation
- Le padding/resize avant tokenisation peut différer

**Mitigation** : exécuter chaque cas OCR avec `--runs=3` minimum et reporter l'écart-type en plus de la moyenne.

---

## 14. Stratégie de dégradation et plan B

### 14.1 Scénarios de dégradation

| Scénario | Seuil d'alerte | Action |
|----------|---------------|--------|
| **Aucun modèle > 0.7 en O2 (texte)** | Précision textuelle trop basse pour être exploitable | Passer à l'approche hybride (OCR dédié + LLM) |
| **Manuscrit brouillon < 0.5 en O2** | Écriture trop difficile pour l'OCR automatique | Proposer à l'utilisateur de retaper le texte (fallback manuel) |
| **Confidence mal calibrée (O4 < 0.3)** | Le filtre HITL ne fonctionne pas | Envoyer tout en HITL, ignorer le score de confidence |
| **Coût OCR > 50% du coût total pipeline** | Budget déséquilibré | Migrer vers un modèle flash ou OCR dédié pour l'étape OCR |
| **Formules mathématiques illisibles** | Maths = cas le plus difficile | Utiliser un OCR spécialisé maths (Google Document AI Math OCR, LaTeX) |
| **Schémas non détectés** | Perte de contenu DOCUMENT | Ajouter une passe de détection d'objets avant l'OCR |

### 14.2 Architecture de fallback progressive

```
Niveau 1 : LLM Vision (actuel)
    ↓ si O2 < 0.7
Niveau 2 : LLM Vision + prompt amélioré (few-shot avec exemples de cahier)
    ↓ si O2 < 0.7
Niveau 3 : Approche hybride (Google Cloud Vision + LLM post-traitement)
    ↓ si O2 < 0.6
Niveau 4 : OCR dédié + validation humaine systématique
    ↓ si échec total
Niveau 5 : Saisie manuelle (l'élève/parent retape le texte)
```

### 14.3 Optimisations du prompt OCR

Si les résultats du benchmark montrent des faiblesses, voici les leviers d'optimisation du prompt :

| Levier | Description | Quand l'utiliser |
|--------|-------------|-----------------|
| **Few-shot examples** | Ajouter 1-2 exemples de transcription (image → JSON) dans le prompt | Si le modèle ne comprend pas le format attendu |
| **Subject-specific hints** | "Cette page contient des formules de physique, note-les en notation mathématique" | Si les formules sont mal transcrites |
| **Multi-pass** | Premier pass = extraction brute, second pass = nettoyage/validation | Si le taux d'erreur est élevé mais le modèle comprend le contexte |
| **Zoom regions** | Envoyer des crops zoomés sur les zones difficiles | Si les petites annotations sont manquées |
| **Image preprocessing** | Binarisation, correction de perspective, augmentation de contraste | Si la qualité photo est systématiquement mauvaise |

### 14.4 Seuils de qualité minimaux pour la production

| Métrique | Seuil minimum (MVP) | Seuil cible (production) |
|----------|---------------------|-------------------------|
| O1 — Détection | ≥ 0.70 | ≥ 0.85 |
| O2 — Précision texte | ≥ 0.75 | ≥ 0.90 |
| O3 — Classification | ≥ 0.80 | ≥ 0.90 |
| O4 — Confidence | ≥ 0.30 | ≥ 0.60 |
| O5 — Ordre | ≥ 0.70 | ≥ 0.85 |
| O6 — Diacritiques | ≥ 0.80 | ≥ 0.90 |
| **QualityScore composite** | **≥ 0.70** | **≥ 0.85** |

Si un modèle ne passe pas les seuils MVP, il est **éliminé** du benchmark pour la production.

### 14.5 Métriques de bout en bout (OCR → IDP)

Le vrai test est le **pipeline complet** : est-ce que des blocs OCR de qualité X produisent des items IDP de qualité Y ?

```
Pipeline_score = f(OCR_quality, IDP_quality)

# Mesure proposée : exécuter le pipeline complet (images → OCR → IDP) sur les cas ayant des images
# et comparer le golden_output.json IDP final vs ce que le pipeline complet produit.
# Cela capture l'effet de propagation d'erreurs OCR vers l'IDP.
```

| OCR quality | IDP quality attendue | Commentaire |
|-------------|---------------------|-------------|
| O2 ≥ 0.90 | Q1 ~ Q1_baseline | Pas de dégradation visible |
| O2 0.75-0.90 | Q1 réduit de 5-15% | Quelques items manqués à cause de texte mal lu |
| O2 0.60-0.75 | Q1 réduit de 20-40% | Dégradation significative, HITL nécessaire |
| O2 < 0.60 | Pipeline inutilisable | Fallback nécessaire |

Ce test de bout en bout sera implémenté comme un troisième mode du benchmark : `--type=pipeline`.
