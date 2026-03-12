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

| # | Indicateur | Description | Méthode de mesure | Poids |
|---|-----------|-------------|-------------------|-------|
| O1 | **Détection de blocs** | Ratio entre le nombre de blocs détectés et le nombre attendu | `min(trouvés, attendus) / max(trouvés, attendus)` | 30% |
| O2 | **Précision textuelle** | Fidélité du texte extrait par rapport au golden | Word overlap : fraction des mots significatifs du golden retrouvés dans la sortie | 50% |
| O3 | **Classification des blocs** | % de blocs avec le bon `block_type` (TEXT vs DIAGRAM) | Match exact après appariement des blocs | 20% |

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
    │   │   └── Évalue vs input.json (O1-O3)
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

**Benchmark OCR** (1 cas avec 3 images, ~1500 tokens/image) :

| Provider | Cas × runs | Coût estimé |
|----------|-----------|-------------|
| Claude Sonnet 4.6 | 1 × 1 | ~$0.05 |
| Claude Haiku 4.5 | 1 × 1 | ~$0.01 |
| GPT-4o | 1 × 1 | ~$0.03 |
| Gemini 2.5 Pro | 1 × 1 | ~$0.02 |
| Gemini 2.5 Flash | 1 × 1 | ~$0.005 |
| **Total OCR** | | **~$0.12** |

Un run complet des deux benchmarks coûte ~$0.52. Avec 3 répétitions : ~$1.56.

> **Note** : le coût OCR augmentera significativement à mesure qu'on ajoutera des cas de test avec images. L'envoi d'images en base64 consomme beaucoup de tokens input.

---

## 7. Métriques détaillées — Comment elles sont calculées

### Benchmark OCR

#### O1 — Détection de blocs

```
score = min(blocs_trouvés, blocs_attendus) / max(blocs_trouvés, blocs_attendus)
```

**Pourquoi** : un modèle qui fusionne 5 blocs en 2, ou qui en segmente 2 en 10, perd de l'information structurelle. Le ratio symétrique pénalise à la fois la sur-segmentation et la sous-segmentation.

#### O2 — Précision textuelle

```
score = avg(word_overlap(golden_block, produced_block))
        pour chaque paire de blocs appariés

word_overlap(golden, produced) = mots_significatifs_retrouvés / mots_significatifs_golden
```

**Pourquoi** : c'est la métrique la plus critique du benchmark OCR. Si le texte extrait ne correspond pas au texte réel du cahier, tout le pipeline aval (structuration, questions) sera faussé. L'OCR doit lire fidèlement l'écriture manuscrite d'un collégien, y compris les fautes d'orthographe.

#### O3 — Classification des blocs

```
score = blocs_correctement_typés / blocs_appariés

correctly_typed = le bloc apparié a le même block_type (TEXT ou DIAGRAM)
                  que son correspondant dans le golden
```

**Pourquoi** : distinguer un schéma (DIAGRAM) d'un texte (TEXT) est essentiel pour le traitement aval. Un schéma mal classifié en texte générera des items de révision absurdes.

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
