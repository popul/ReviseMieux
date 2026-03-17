# Benchmark LLM — Révise Mieux

> Framework d'évaluation comparative des modèles LLM pour les cas d'usage Révise Mieux.
> Deux benchmarks indépendants : **OCR** (images → blocs texte) et **IDP** (blocs texte → items structurés).
> Exécutable à la demande via `make bench-idp` / `make bench-ocr` depuis `backend/`.
> Dernière mise à jour : 2026-03-12.

---

> **Mise à jour mars 2026 — Résultats réels disponibles**
>
> Ce document a été rédigé **avant** l'exécution du benchmark. Les résultats estimés, les tableaux de scores fictifs et les recommandations de modèles ci-dessous sont **obsolètes**.
>
> Les **résultats réels** (21 modèles IDP, 17 modèles OCR) sont dans [`backend/testdata/benchmark/README.md`](../backend/testdata/benchmark/README.md).
>
> Changements majeurs par rapport aux prédictions de ce document :
> - **Claude Sonnet n'est plus recommandé** — seul modèle avec des hallucinations (8%), et 30-50x plus cher que les alternatives à qualité égale ou supérieure.
> - **Top 3 IDP** : Qwen3.5-397B (quality 0.88), mistral-small (0.87), gemini-2.5-flash (0.87).
> - **Top 3 OCR** : Qwen3-VL-32B (quality 0.85), gemini-2.5-flash (0.84), gpt-4o (0.83).
> - **Exécution locale viable** : Qwen3-VL-32B + Gemma 3 27B via Ollama sur MacBook Pro M5 Max (~2 min/chapitre, coût $0).
> - **Modèles open-source via OpenRouter** (Qwen, Llama 4, Gemma) sont compétitifs avec les APIs propriétaires.
>
> L'architecture du benchmark (métriques, runner, structure des cas) décrite ci-dessous reste valide.

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

### 2.1 Modèles testés

> **Mis à jour mars 2026.** 21 modèles IDP, 17 modèles OCR testés. Voir [`backend/testdata/benchmark/README.md`](../backend/testdata/benchmark/README.md) pour les résultats complets.

| Provider | Modèle | Vision | Accès | Benchmarks |
|----------|--------|--------|-------|------------|
| **Anthropic** | Claude Sonnet 4.6 | Oui | Direct | OCR + IDP |
| **Anthropic** | Claude Haiku 4.5 | Oui | Direct | OCR + IDP |
| **OpenAI** | GPT-4o, GPT-4o Mini, o3-mini | Oui/Oui/Non | Direct | OCR + IDP |
| **Google** | Gemini 2.5 Pro, Gemini 2.5 Flash | Oui | Direct | OCR + IDP |
| **Mistral** | Mistral Large, Mistral Small, Mistral Small 3.2, Pixtral 12B | Oui | Direct | OCR + IDP |
| **DeepSeek** | deepseek-chat, deepseek-reasoner | Non | Direct | IDP |
| **Qwen** | Qwen3.5-397B, Qwen3.5-9B, Qwen3-VL-235B, Qwen3-VL-32B | Oui (VL) | OpenRouter | OCR + IDP |
| **Meta** | Llama 4 Maverick, Llama 4 Scout | Oui | OpenRouter | OCR + IDP |
| **Google** | Gemma 3 27B | Oui | OpenRouter | OCR + IDP |
| **NVIDIA** | Nemotron Nano VL 12B | Oui | OpenRouter (gratuit) | OCR + IDP |
| **MiniMax** | MiniMax M2.5 | Non | OpenRouter | IDP |
| **StepFun** | Step 3.5 Flash | Non | OpenRouter | IDP |

Les modèles open-source (Qwen, Llama, Gemma) sont testés via OpenRouter et peuvent aussi tourner en local via Ollama.

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

Cas de test disponibles (mars 2026) :

| # | ID | Matière | Niveau | Contenu | Items | Statut |
|---|-----|---------|--------|---------|-------|--------|
| 10 | `10_SVT_cours_louis` | SVT | 5e | Prélèvement de matière par les végétaux | 13 (4K + 3P + 6D) | Validé (golden corrigé manuellement) |

> K = KNOWLEDGE, P = PROCEDURE, D = DOCUMENT

Cas prioritaires à ajouter pour rendre le benchmark fiable (diversité des matières et difficultés OCR) :

| Matière | Intérêt | Difficulté OCR |
|---------|---------|---------------|
| Histoire / Français | Beaucoup de texte, peu de schémas | Faible |
| Maths / Physique | Formules, calculs, tableaux de mesures | Élevé |
| Cahier brouillon | Écriture difficile, ratures | Élevé |

Voir [`backend/testdata/benchmark/cases/README.md`](../backend/testdata/benchmark/cases/README.md) pour le guide de création de cas.

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

Exemple de sortie console (résultats réels, mars 2026) :

```
╔══════════════════════════╦═══════╦═══════╦═══════╦═════════╦════════╦═══════════╗
║ Modèle                   ║ Compl.║ Fidél.║ Hallu.║ $/item  ║ Lat.ms ║ Score     ║
╠══════════════════════════╬═══════╬═══════╬═══════╬═════════╬════════╬═══════════╣
║ mistral-small-latest     ║  0.85 ║  1.00 ║  0.00 ║ 0.00006 ║   8393 ║    0.9055 ║
║ gemini-2.5-flash         ║  0.85 ║  1.00 ║  0.00 ║ 0.00012 ║  25538 ║    0.8967 ║
║ deepseek-chat            ║  0.85 ║  1.00 ║  0.00 ║ 0.00021 ║    392 ║    0.8740 ║
║ claude-haiku-4-5         ║  0.77 ║  1.00 ║  0.00 ║ 0.00105 ║   9302 ║    0.8060 ║
║ claude-sonnet-4-6        ║  0.77 ║  0.92 ║  0.08 ║ 0.00311 ║  20162 ║    0.5470 ║
╚══════════════════════════╩═══════╩═══════╩═══════╩═════════╩════════╩═══════════╝
```

Voir [`backend/testdata/benchmark/README.md`](../backend/testdata/benchmark/README.md) pour les résultats complets des 21 modèles.

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

> **Baseline humaine** : toutes les métriques OCR comparent la sortie d'un modèle au `input.json`, qui est une **transcription humaine experte** du cahier (`"ocr_source": "human_expert"` dans metadata.json). C'est la meilleure baseline possible car :
> - Un humain qui lit la photo transcrit **fidèlement** le contenu, y compris les fautes d'orthographe de l'élève
> - Un humain segmente en **blocs sémantiques** cohérents (pas des blocs géométriques arbitraires)
> - Un humain décrit les **schémas et graphiques** de manière pertinente
> - Le score `confidence: 1.0` reflète la certitude totale du transcripteur
>
> Un modèle OCR parfait obtiendrait O1=1.0, O2=1.0, O3=1.0 : il produirait exactement la même transcription qu'un humain expert.

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

### Q6 — Cohérence du groupement en Notions

```
score = (count_score × 0.4) + (name_score × 0.3) + (assignment_score × 0.3)

count_score = 1 - |notions_produites - notions_attendues| / notions_attendues
              clampé à [0, 1]

name_score = avg(max_similarity(notion_name_llm, notion_names_golden))
             pour chaque notion produite
             similarity = Jaccard sur les mots significatifs du nom

assignment_score = items_correctly_assigned / items_matched
                   correctly_assigned = l'item est dans la même notion
                                        que dans le golden output
                                        (match par nom de notion, fuzzy)
```

**Pourquoi** : les items doivent être regroupés en `Notion`s pédagogiquement cohérentes. Un modèle qui met "théorème de Pythagore" et "équation du second degré" dans la même notion produit des sessions de révision incohérentes. Cette métrique est **semi-automatique** : le scoring algorithmique ci-dessus capture les cas évidents, mais les cas ambigus (deux découpages valides) nécessitent un jugement humain.

**Composantes** :
- **count_score** (40%) : le nombre de notions produites est-il dans la fourchette attendue ? Trop de notions = fragmentation, trop peu = confusion des concepts.
- **name_score** (30%) : les noms des notions sont-ils pertinents et proches du golden ? "Densité et masse volumique" vs "Propriétés de la matière" — les deux sont acceptables mais le premier est plus précis.
- **assignment_score** (30%) : chaque item est-il assigné à la bonne notion ? Un item "ρ = m/V" dans la notion "Unités de mesure" au lieu de "Masse volumique" est une erreur d'assignation.

**Cas limites** :
- Un modèle qui met tous les items dans une seule notion → count_score ≈ 0, assignment_score = 1 (trivial)
- Un modèle qui crée une notion par item → count_score ≈ 0, name_score variable
- Deux découpages également valides → score potentiellement injuste → signaler en review manuelle

### Q7 — Conformité du schéma JSON

```
score = 1.0 si le JSON est valide ET conforme au schéma attendu
        0.0 sinon

# Vérifications :
# 1. JSON parseable sans erreur
# 2. Tous les champs requis présents (items[].type, items[].term, etc.)
# 3. Enums respectés (type ∈ {KNOWLEDGE, PROCEDURE, DOCUMENT})
# 4. Types corrects (term = string, keywords = []string)
```

**Pourquoi** : un modèle qui retourne du JSON invalide ou non conforme au schéma bloque le pipeline. C'est un critère binaire (pass/fail) mais essentiel. Un modèle qui retourne du texte libre au lieu de JSON est inutilisable, quelle que soit la qualité du contenu.

### P6 — Coût par item extrait

```
cost_per_item = total_cost / items_extracted

total_cost = (tokens_input × price_input / 1M)
           + (tokens_output × price_output / 1M)
```

**Pourquoi** : le vrai indicateur économique. Un modèle 3x plus cher mais qui extrait 2x plus d'items est plus rentable par item.

---

## 8. Seuils d'acceptation et scénarios de décision

### 8.1 Seuils go/no-go pour la production

Un modèle doit **passer tous les seuils éliminatoires** pour être considéré. Les seuils de qualité sont des **minimums** — en dessous, le modèle est exclu quelle que soit sa performance sur les autres axes.

#### IDP — Seuils éliminatoires

| Métrique | Seuil minimum | Justification |
|----------|---------------|---------------|
| **Q5 (hallucinations)** | ≤ 0.05 (5%) | Un item hallucine = un faux enseigné à un collégien. Non négociable. |
| **Q7 (schéma JSON)** | = 1.0 (100%) | Un JSON invalide bloque le pipeline. Binaire : pass ou fail. |
| **Q3 (fidélité)** | ≥ 0.80 | En dessous, le contenu est déformé — les items ne correspondent plus au cahier. |
| **Q1 (complétude)** | ≥ 0.70 | En dessous, trop d'items sont perdus — les révisions seront incomplètes. |

#### IDP — Seuils de confort (souhaitables, pas éliminatoires)

| Métrique | Seuil souhaité | Commentaire |
|----------|----------------|-------------|
| Q1 (complétude) | ≥ 0.85 | Bon niveau de couverture |
| Q3 (fidélité) | ≥ 0.90 | Haute fidélité au contenu original |
| Q4 (keywords) | ≥ 0.70 | Keywords suffisamment précis pour générer des questions |
| Q6 (notions) | ≥ 0.60 | Groupement cohérent — en dessous, review manuelle nécessaire |
| P6 (coût/item) | ≤ $0.005 | Budget soutenable pour un SaaS éducatif |
| P7 (latence) | ≤ 5000ms | UX acceptable pour un upload de chapitre |

#### OCR — Seuils éliminatoires

| Métrique | Seuil minimum | Justification |
|----------|---------------|---------------|
| **O2 (précision texte)** | ≥ 0.70 | En dessous, l'IDP n'a pas assez de matière pour travailler |
| **O1 (détection blocs)** | ≥ 0.60 | Blocs manqués = items perdus en cascade |
| **O3 (ordre lecture)** | ≥ 0.80 | Un mauvais ordre → regroupements notions incohérents |

#### Critères non-benchmark (éliminatoires)

| Critère | Condition éliminatoire | Modèles concernés |
|---------|----------------------|-------------------|
| **RGPD** | Données transférées hors UE sans DPA valide | DeepSeek (Chine — rédhibitoire) |
| **SLA disponibilité** | Uptime < 99.5% sur les 3 derniers mois | Vérifier status pages |
| **Rate limits** | < 100 RPM au tier actuel | Modèles en preview/beta |

### 8.2 Algorithme de recommandation automatique

Le rapport inclut une **recommandation automatique** basée sur l'algorithme suivant :

```
Pour chaque modèle :
  1. Vérifier les seuils éliminatoires → si un seuil échoue : EXCLU
  2. Vérifier RGPD / SLA → si non conforme : EXCLU
  3. Calculer le score composite (pondéré, §4)
  4. Classer les modèles éligibles par score composite décroissant

Recommandations :
  - "RECOMMANDÉ" = meilleur score composite parmi les éligibles
  - "ALTERNATIVE BUDGET" = meilleur score avec coût < 50% du recommandé
  - "NON RECOMMANDÉ" = exclu par un seuil éliminatoire (indiquer lequel)
  - "INSUFFISANT" = aucun modèle ne passe tous les seuils → intervention humaine
```

#### Recommandations issues du benchmark (mars 2026)

| Stratégie | OCR | IDP | Coût/chapitre |
|-----------|-----|-----|---------------|
| **Best quality** | Qwen3-VL-32B (quality 0.85) | Qwen3.5-397B (quality 0.88) | ~$0.003 |
| **Best value** | gemini-2.5-flash (quality 0.84) | mistral-small (quality 0.87) | ~$0.002 |
| **Ultra-cheap** | Nemotron Nano VL (gratuit) | Gemma 3 27B (quality 0.84) | ~$0.000 |
| **Local (Ollama)** | Qwen3-VL-32B Q8 | Gemma 3 27B Q8 | $0 (~2 min/chapitre) |

Voir [`backend/testdata/benchmark/README.md`](../backend/testdata/benchmark/README.md) pour le classement complet.

### 8.3 Scénarios de décision

Le benchmark alimente des **décisions contextuelles** :

| Scénario | Critère de décision | Action |
|----------|-------------------|--------|
| Modèle A a Q1 > 0.90 et coûte 10x moins que le baseline | Qualité suffisante, économie massive | Migrer vers A |
| Modèle A a Q5 > 0.05 (hallucinations) | **Éliminé** par seuil go/no-go | Ne pas utiliser — rapport l'indique en rouge |
| Modèle A a Q3 < 0.80 (fidélité basse) | **Éliminé** par seuil go/no-go | Exclure — risque pour les élèves |
| Tous les modèles se valent en qualité | Décision sur le coût + RGPD | Choisir le moins cher conforme |
| Aucun modèle ne passe les seuils | **INSUFFISANT** | Améliorer les prompts (§15.6) avant de rechoisir |
| Nouveau modèle sorti | Relancer le benchmark | `go run ./cmd/benchmark/ --models=new_model` |
| Provider augmente ses prix | Relancer le benchmark | Vérifier si un switch est rentable |
| Résultats très variables entre runs | Modèle instable | Augmenter `--runs=5`, vérifier `temperature=0` |
| **OCR** : Modèle A a O2 < 0.70 (texte) | **Éliminé** par seuil go/no-go | Tester l'approche hybride (OCR dédié + LLM) — voir §12.3 |
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

**Matières prioritaires** : Histoire, Géographie, Physique, Chimie, SVT.

| Phase | Cas | Matières couvertes | Type de contenu | Objectif |
|-------|-----|-------------------|-----------------|----------|
| **Phase 1** (actuel) | 1 cas, 3 images | SVT | Manuscrit + polycopié + schéma | Preuve de concept |
| **Phase 2** (court terme) | 4-5 cas, ~12-15 images | SVT, Physique, Chimie, Histoire ou Géo | Variété de contenus + golden humain | Première comparaison fiable |
| **Phase 3** (moyen terme) | 8-10 cas, ~25-30 images | Toutes matières prioritaires + Maths, Français | Variété complète | Benchmark représentatif |

**Cas concrets à créer en Phase 2** :

| ID | Matière | Contenu attendu | Difficulté OCR |
|----|---------|-----------------|----------------|
| `11_histoire_xxx` | Histoire | Dates, frises chronologiques, texte dense | Moyenne (surtout du texte) |
| `12_geographie_xxx` | Géographie | Cartes annotées, légendes, croquis | Haute (schémas + légendes manuscrites) |
| `13_physique_xxx` | Physique-Chimie | Formules, schémas de circuits/montages, tableaux de mesures | Haute (formules + symboles) |
| `14_svt_xxx` | SVT | 2e cas SVT avec écriture différente du cas 10 | Moyenne (teste la généralisation) |

> **Note** : remplacer `xxx` par un descriptif court du chapitre (ex: `11_histoire_revolution`, `13_physique_optique`).

### 13.3 Axes de diversité à couvrir

Chaque axe doit être représenté par au moins 2 cas de test :

| Axe | Valeurs à couvrir | Priorité |
|-----|-------------------|----------|
| **Qualité écriture** | Soigné, Normal, Brouillon | Haute |
| **Type d'encre** | Bleu, Noir, Rouge, Crayon, Mixte | Moyenne |
| **Qualité photo** | Bonne (bien éclairé, droit), Moyenne (léger angle), Mauvaise (sombre, flou) | Haute |
| **Contenu** | Texte seul, Texte + schéma, Formules, Tableaux, Cartes | Haute |
| **Support** | Cahier manuscrit pur, Polycopié collé + annotations, Imprimé annoté | Moyenne |
| **Matière** | Histoire, Géographie, Physique-Chimie, SVT (prioritaires) ; Maths, Français (phase 3) | Haute |

### 13.4 Protocole de création d'un cas OCR

#### Étape 1 — Photographier (5 min)

1. Choisir 2-4 pages consécutives d'un cahier réel
2. Photographier avec un smartphone en conditions réalistes (pas de studio photo)
3. Nommer les fichiers `IMG_XXXX.jpeg` et les placer dans `cases/<id>/images/`

**Conditions photo recommandées** :
- Éclairage naturel ou lampe de bureau (pas de flash direct)
- Cahier posé à plat, pas tenu en main
- Cadrage : la page entière visible, marges incluses
- Résolution native du téléphone (ne pas compresser)
- OK si un peu de l'autre page est visible sur le bord — c'est réaliste

#### Étape 2 — Transcrire manuellement (golden OCR humain, 30-60 min)

C'est l'étape **la plus importante**. Le `input.json` que tu crées est la **référence absolue** contre laquelle tous les modèles seront évalués. La qualité du benchmark dépend directement de la qualité de cette transcription.

**Règles de transcription** :

| Règle | Explication | Exemple |
|-------|-------------|---------|
| **Fidélité totale** | Transcrire exactement ce qui est écrit, y compris les fautes d'orthographe de l'élève | "les staumates" → garder "staumates" (pas "stomates") |
| **Segmenter en blocs sémantiques** | Un bloc = une unité de sens cohérente (titre, paragraphe, activité, schéma) | Le titre "Chapitre 3 : ..." est un bloc séparé |
| **Décrire les visuels** | Pour les schémas/graphiques, décrire entre `[crochets]` ce qui est visible | `[Schéma : coupe de feuille montrant...]` |
| **Respecter l'ordre de lecture** | Les blocs doivent être dans l'ordre de lecture naturel, pas dans l'ordre physique (haut→bas, gauche→droite) | Si une note en marge se rapporte au §2, la mettre après le §2 |
| **Ne rien inventer** | Ne pas compléter les phrases coupées en bas de page, ne pas deviner les mots illisibles | Un mot illisible → `[illisible]` |
| **Préserver la mise en forme** | Listes numérotées, retours à la ligne, tirets — garder la structure | "1 – L'eau\n2 – Les sels minéraux" |

**Workflow recommandé** :

```
1. Ouvrir les photos sur un grand écran (pas sur téléphone)
2. Ouvrir un éditeur de texte à côté
3. Parcourir la photo de haut en bas
4. Pour chaque bloc identifié :
   a. Déterminer le type : TEXT, DIAGRAM, ou TABLE
   b. Transcrire le contenu mot à mot
   c. Pour les schémas : décrire le contenu visuel entre [crochets]
5. Relire la transcription en comparant phrase par phrase avec la photo
6. Vérifier l'ordre : les blocs sont-ils dans l'ordre de lecture ?
```

**Décisions de découpage** :

Le découpage en blocs est un jugement humain. Voici les heuristiques :

| Situation | Découpage recommandé |
|-----------|---------------------|
| Titre + paragraphe de cours | 1 bloc (le titre contexte le paragraphe) |
| Activité avec questions numérotées | 1 bloc pour l'énoncé |
| Réponses manuscrites de l'élève | 1 bloc séparé (source différente : élève vs prof/polycopié) |
| Schéma avec légendes | 1 bloc DIAGRAM (description textuelle du schéma + légendes) |
| Graphique avec données | 1 bloc DIAGRAM (décrire les axes, courbes, valeurs clés) |
| Tableau | 1 bloc TABLE (transcrire le contenu en texte structuré) |
| Encadré "à retenir" | 1 bloc séparé (souvent l'item KNOWLEDGE le plus important) |

#### Étape 3 — Assembler le `input.json`

```json
{
  "blocks": [
    {
      "text": "Chapitre 3 : Prélèvement de matière...\n\nI – Le prélèvement...",
      "block_type": "TEXT",
      "confidence": 1.0
    },
    {
      "text": "[Schéma : expérience avec deux récipients...]",
      "block_type": "DIAGRAM",
      "confidence": 1.0
    }
  ]
}
```

> **`confidence: 1.0` pour tous les blocs** — c'est une transcription humaine parfaite, pas un OCR automatique. Le score de confiance est de 1.0 par définition.

#### Étape 4 — Créer le golden IDP (`golden_output.json`)

À partir de ta transcription, extraire les items pédagogiques :
- Identifier chaque fait/formule/procédure/schéma → créer un item
- Grouper les items en notions
- Suivre le format existant (voir `10_SVT_cours_louis/golden_output.json`)

#### Étape 5 — Renseigner `metadata.json`

```json
{
  "id": "11_physique_circuit",
  "subject": "Physique-Chimie",
  "level": "4e",
  "topic": "Circuit électrique en série et en dérivation",
  "difficulty": "high",
  "has_images": true,
  "expected_item_count": 10,
  "expected_types": {
    "KNOWLEDGE": 7,
    "PROCEDURE": 2,
    "DOCUMENT": 1
  },
  "ocr_source": "human_expert",
  "ocr_author": "Prénom (rôle)",
  "ocr_date": "2026-03-14",
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

### 14.5 Benchmark pipeline bout-en-bout (`--type=pipeline`)

Le vrai test est le **pipeline complet** : est-ce que des blocs OCR de qualité X produisent des items IDP de qualité Y ? Ce mode mesure la **propagation d'erreurs** de l'OCR vers l'IDP.

#### Architecture

```
cmd/benchmark/main.go --type=pipeline

    ├── Charge les cas de test avec images ("has_images": true)
    ├── Pour chaque combinaison (OCR_provider × IDP_provider) :
    │   ├── Pour chaque cas de test :
    │   │   ├── Étape 1 : Envoie les images au OCR_provider → blocs OCR bruts
    │   │   ├── Étape 2 : Envoie les blocs OCR bruts au IDP_provider → items structurés
    │   │   ├── Mesure latence totale, tokens totaux, coût total
    │   │   ├── Évalue les blocs OCR vs input.json (métriques O1-O6)
    │   │   ├── Évalue les items vs golden_output.json (métriques Q1-Q7)
    │   │   └── Calcule le delta de dégradation (voir ci-dessous)
    │   └── Agrège les résultats
    ├── Matrice de résultats OCR×IDP
    └── Sauvegarde dans results/pipeline/
```

#### Combinaisons testées

Le mode pipeline teste toutes les combinaisons pertinentes :

```
OCR providers :  [Sonnet, Haiku, GPT-4o, Flash, Cloud Vision, ...]
                          ×
IDP providers :  [Sonnet, Haiku, GPT-4o, Flash, ...]
```

Cela inclut les **combinaisons croisées** (OCR Haiku + IDP Sonnet, OCR Cloud Vision + IDP Sonnet) — exactement ce qu'il faut pour valider l'approche hybride (§12.3).

#### Métrique de dégradation

```
degradation = 1 - (Q1_pipeline / Q1_baseline)

Q1_baseline = Q1 du benchmark IDP seul (input = golden OCR, §5.1)
Q1_pipeline = Q1 du pipeline complet (input = images → OCR → IDP)

# Même calcul pour Q3, Q5, etc.
```

| Dégradation Q1 | Interprétation | Action |
|----------------|----------------|--------|
| < 5% | Négligeable — l'OCR ne dégrade pas l'IDP | OK pour production |
| 5-15% | Modérée — quelques items perdus par erreurs OCR | Acceptable si coût justifié |
| 15-30% | Significative — HITL nécessaire | Améliorer le prompt OCR ou changer de modèle |
| > 30% | Critique — pipeline inutilisable | Fallback (§14.2) |

#### Rapport de sortie spécifique

```
╔════════════════════════════════════════════════════════════════════════╗
║                    Pipeline OCR → IDP : Matrice                       ║
╠══════════════╦═══════════════╦═══════════════╦═══════════════════════╣
║ OCR ↓ IDP →  ║ Sonnet 4.6    ║ Haiku 4.5     ║ Gemini Flash         ║
╠══════════════╬═══════════════╬═══════════════╬═══════════════════════╣
║ Sonnet 4.6   ║ Q1=0.90 $0.07 ║ Q1=0.88 $0.05 ║ Q1=0.85 $0.05       ║
║ Haiku 4.5    ║ Q1=0.87 $0.04 ║ Q1=0.84 $0.02 ║ Q1=0.82 $0.02       ║
║ Cloud Vision ║ Q1=0.83 $0.03 ║ Q1=0.80 $0.01 ║ Q1=0.78 $0.01       ║
║ GPT-4o Mini  ║ Q1=0.85 $0.03 ║ Q1=0.82 $0.01 ║ Q1=0.80 $0.01       ║
╚══════════════╩═══════════════╩═══════════════╩═══════════════════════╝
(valeurs fictives à titre d'illustration)
```

Ce tableau permet de répondre directement à : **quelle combinaison OCR+IDP offre le meilleur rapport qualité/coût ?**

#### Commandes

```bash
# Toutes les combinaisons
go run ./cmd/benchmark/ --type=pipeline --all

# Une combinaison spécifique
go run ./cmd/benchmark/ --type=pipeline --ocr=anthropic:haiku --idp=anthropic:sonnet

# OCR dédié + LLM IDP (test de l'approche hybride)
go run ./cmd/benchmark/ --type=pipeline --ocr=google:cloud-vision --idp=anthropic:sonnet

make bench-pipeline ALL=1
```

#### Coût estimé

Un run pipeline complet teste N×M combinaisons. Avec 5 OCR providers × 4 IDP providers = 20 combinaisons × 1 cas image = ~$1.60. Avec 3 répétitions : ~$4.80. Le coût augmente linéairement avec le nombre de cas images (§13).

#### Prérequis

Le mode pipeline nécessite :
1. Au moins 1 cas de test avec `"has_images": true`
2. Au moins 1 OCR provider et 1 IDP provider configurés
3. Les clés API correspondantes dans les variables d'environnement

> **Note** : ce mode est le plus coûteux car il exécute deux appels API par combinaison×cas. Il est recommandé de l'exécuter après les benchmarks OCR et IDP séparés, une fois les modèles présélectionnés.

---

## 15. Gestion des prompts

### 15.1 Pourquoi c'est critique

Le prompt est une **variable silencieuse** du benchmark. À modèle identique, un changement de prompt peut faire varier Q1 de ±20%. Le benchmark doit donc :
1. **Fixer** le prompt utilisé par run pour que les résultats soient reproductibles
2. **Versionner** les prompts pour tracer les évolutions
3. **Stocker** le prompt exact utilisé dans chaque résultat

### 15.2 Stockage des prompts

```
backend/testdata/benchmark/
├── prompts/
│   ├── ocr/
│   │   ├── v1_baseline.txt          # Prompt OCR initial
│   │   ├── v2_fewshot.txt           # Avec exemples few-shot
│   │   └── current -> v1_baseline.txt  # Symlink vers le prompt actif
│   ├── idp/
│   │   ├── v1_baseline.txt          # Prompt IDP initial
│   │   ├── v2_subject_hints.txt     # Avec hints par matière
│   │   └── current -> v1_baseline.txt
│   └── README.md                    # Changelog des prompts
```

### 15.3 Contenu d'un fichier prompt

Chaque fichier prompt contient le **system prompt** et le **user prompt template** séparés par un délimiteur :

```
--- SYSTEM ---
Tu es un assistant spécialisé dans l'extraction de contenu pédagogique
à partir de photos de cahiers de collégiens français.

Extrais FIDÈLEMENT le contenu, y compris les fautes d'orthographe.
Ne corrige PAS l'orthographe de l'élève.
...

--- USER ---
Voici {{N_IMAGES}} photos d'un cahier de {{SUBJECT}}, niveau {{LEVEL}}.

Extrais chaque bloc de texte ou schéma dans l'ordre de lecture.
Retourne un JSON conforme au schéma suivant :
...
```

Les variables `{{...}}` sont interpolées par le runner au moment de l'exécution.

### 15.4 Versioning et traçabilité

Chaque résultat de benchmark (`results/*/summary.json`) inclut :

```json
{
  "run_id": "2026-03-13_10h00",
  "prompt_version": "ocr/v1_baseline",
  "prompt_hash": "sha256:a1b2c3...",
  "models": ["claude-sonnet-4-6", "gpt-4o"],
  "results": [...]
}
```

Le `prompt_hash` garantit qu'on peut détecter si un prompt a été modifié sans changer la version.

### 15.5 Règles

1. **Ne jamais modifier un prompt existant** — créer une nouvelle version (`v2_...`)
2. **Un run = un prompt** — le même prompt est utilisé pour tous les modèles dans un run, sinon la comparaison est invalide
3. **Documenter le changement** dans `prompts/README.md` (quoi a changé, pourquoi)
4. **Comparer les versions** : relancer le benchmark avec l'ancien et le nouveau prompt pour mesurer l'impact

```bash
# Comparer deux versions de prompt
go run ./cmd/benchmark/ --type=ocr --all --prompt=ocr/v1_baseline
go run ./cmd/benchmark/ --type=ocr --all --prompt=ocr/v2_fewshot

# Le rapport affiche le delta entre les deux runs
go run ./cmd/benchmark/ --compare=results/ocr/2026-03-13_v1,results/ocr/2026-03-13_v2
```

### 15.6 Optimisation itérative

Le cycle d'optimisation des prompts suit le même principe que le TDD :

```
1. Lancer le benchmark avec le prompt courant → résultats baseline
2. Identifier la métrique la plus faible (ex: O6 diacritiques = 0.72)
3. Modifier le prompt pour adresser cette faiblesse (ex: "Préserve les accents...")
4. Créer une nouvelle version du prompt (v2_diacritics)
5. Relancer le benchmark → comparer
6. Si amélioration sans régression → mettre à jour le symlink `current`
7. Si régression sur une autre métrique → itérer
```

---

## 16. Limites actuelles du dataset

> **AVERTISSEMENT** : le benchmark OCR ne dispose actuellement que d'**un seul cas de test avec images** (10_SVT_cours_louis, 3 photos). Les conclusions tirées du benchmark OCR sont **préliminaires** et ne doivent pas être considérées comme définitives.

### 16.1 Risques d'un dataset insuffisant

| Risque | Description | Impact |
|--------|-------------|--------|
| **Biais de matière** | Un seul cas SVT — aucune couverture maths, physique, français | Un modèle peut exceller en SVT et échouer en maths (formules) |
| **Biais d'écriture** | Une seule qualité d'écriture (celle de Louis) | Pas de généralisation possible à d'autres élèves |
| **Biais de photo** | Une seule qualité photo, un seul appareil | Pas de robustesse testée (angle, éclairage, résolution) |
| **Surapprentissage du prompt** | Optimiser le prompt sur 1 cas = overfitting | Le prompt fonctionnera sur ce cas mais pas sur d'autres |
| **Absence de cas difficiles** | Pas de manuscrit brouillon, pas de multi-encre | Le benchmark ne teste pas les cas limites |

### 16.2 Seuils de confiance par taille de dataset

| Nb de cas OCR | Confiance dans les résultats | Décisions possibles |
|---------------|------------------------------|---------------------|
| **1** (actuel) | Très faible — indicatif seulement | Aucune décision définitive. Utile pour valider le framework. |
| **3-5** | Modérée — tendances visibles | Éliminer les modèles clairement inadéquats |
| **8-10** | Bonne — résultats exploitables | Choisir le modèle OCR pour la production |
| **15+** | Haute — statistiquement significatif | Publier des benchmarks, optimiser les prompts |

### 16.3 Plan d'action

La priorité est d'atteindre **5 cas OCR** (Phase 2, §13.2) avant de tirer des conclusions sur le choix LLM vs OCR dédié. Le protocole de création est décrit en §13.4.

Matières prioritaires pour les prochains cas :
1. **Mathématiques** — le plus difficile (formules, symboles, graphiques)
2. **Physique-Chimie** — mixte (formules + schémas de circuits)
3. **Histoire-Géographie** — cas "facile" (texte majoritaire) pour établir un plafond
4. **Un second cas SVT** avec une écriture différente — pour tester la généralisation
