# Benchmark LLM — Révise Mieux

Benchmark comparatif des modèles LLM pour les deux étapes du pipeline :
1. **OCR** : images de cahier → blocs de texte structurés
2. **IDP** (structuration) : blocs OCR → items pédagogiques (KNOWLEDGE, PROCEDURE, DOCUMENT)

## Structure

```
benchmark/
├── Makefile                  # Génération assistée des golden files
├── scripts/generate_json.py  # Script de génération (utilise les prompts partagés)
├── cases/                    # Cas de test (photos + golden files)
│   └── 10_SVT_cours_louis/   # Exemple complet validé
└── results/                  # Résultats des runs
    ├── idp/                  # Résultats structuration
    └── ocr/                  # Résultats OCR
```

## Lancer un benchmark

```bash
# Depuis backend/
make bench-idp ALL=1                              # IDP tous modèles, tous cas
make bench-ocr ALL=1                              # OCR tous modèles vision
make bench-idp ALL=1 CASE=10_SVT_cours_louis      # IDP un seul cas
make bench-idp PROVIDER=anthropic                  # Un provider
make bench-idp MODELS=gpt-4o,claude-sonnet-4-6     # Modèles spécifiques
make bench-models                                  # Lister les modèles disponibles
make bench-report                                  # Rapport HTML depuis derniers résultats
```

## Clés API

Configurer dans `.env.bench` à la racine du projet :

```
ANTHROPIC_API_KEY=sk-ant-...
OPENAI_API_KEY=sk-...
GOOGLE_AI_API_KEY=AI...
MISTRAL_API_KEY=...
DEEPSEEK_API_KEY=sk-...
OPENROUTER_API_KEY=sk-or-...
```

---

## Resultats — Mars 2026

Benchmark exécuté sur le cas `10_SVT_cours_louis` (3 photos, 29 blocs OCR, 13 items golden).

### IDP — Structuration (blocs OCR → items pédagogiques)

| # | Modèle | Provider | Items | Compl. | Fidél. | Hallu. | Quality | $/item | Latence |
|---|--------|----------|-------|--------|--------|--------|---------|--------|---------|
| 1 | Qwen3.5-397B-A17B | OpenRouter | 12 | 0.85 | 1.00 | 0.00 | **0.88** | $0.00011 | 2.7s |
| 2 | mistral-large | Mistral | 12 | 0.85 | 1.00 | 0.00 | **0.88** | $0.00120 | 31.8s |
| 3 | mistral-small | Mistral | 13 | 0.85 | 1.00 | 0.00 | **0.87** | $0.00006 | 8.4s |
| 4 | gemini-2.5-flash | Google | 12 | 0.85 | 1.00 | 0.00 | **0.87** | $0.00012 | 25.5s |
| 5 | deepseek-chat | DeepSeek | 13 | 0.85 | 1.00 | 0.00 | 0.85 | $0.00021 | 0.4s |
| 6 | Llama 4 Maverick | OpenRouter | 10 | 0.69 | 1.00 | 0.00 | 0.85 | $0.00007 | 0.7s |
| 7 | o3-mini | OpenAI | 9 | 0.69 | 1.00 | 0.00 | 0.85 | $0.00221 | 34.5s |
| 8 | claude-haiku-4-5 | Anthropic | 11 | 0.77 | 1.00 | 0.00 | 0.86 | $0.00105 | 9.3s |
| 9 | Gemma 3 27B | OpenRouter | 11 | 0.77 | 1.00 | 0.00 | 0.84 | $0.00003 | 0.3s |
| 10 | gemini-2.5-pro | Google | 12 | 0.77 | 1.00 | 0.00 | 0.84 | $0.00163 | 43.4s |
| 11 | MiniMax M2.5 | OpenRouter | 11 | 0.69 | 1.00 | 0.00 | 0.84 | $0.00020 | 5.2s |
| 12 | claude-sonnet-4-6 | Anthropic | 12 | 0.77 | 0.92 | **0.08** | 0.83 | $0.00311 | 20.2s |
| 13 | Pixtral 12B | Mistral | 14 | 0.77 | 1.00 | 0.00 | 0.83 | $0.00004 | 11.3s |
| 14 | deepseek-reasoner | DeepSeek | 10 | 0.69 | 1.00 | 0.00 | 0.80 | $0.00082 | 0.3s |
| 15 | gpt-4o | OpenAI | 8 | 0.54 | 1.00 | 0.00 | 0.78 | $0.00179 | 9.3s |
| 16 | Qwen3-VL-235B | OpenRouter | 9 | 0.54 | 1.00 | 0.00 | 0.76 | $0.00017 | 0.8s |
| 17 | Step 3.5 Flash | OpenRouter | 10 | 0.62 | 0.80 | **0.20** | 0.75 | $0.00008 | 2.1s |
| 18 | Qwen3-VL-32B | OpenRouter | 10 | 0.54 | 1.00 | 0.00 | 0.73 | $0.00009 | 0.7s |
| 19 | gpt-4o-mini | OpenAI | 7 | 0.46 | 1.00 | 0.00 | 0.72 | $0.00011 | 17.8s |
| 20 | Qwen3.5-9B | OpenRouter | 10 | 0.46 | 1.00 | 0.00 | 0.72 | $0.00006 | 0.9s |
| 21 | Llama 4 Scout | OpenRouter | 6 | 0.38 | 1.00 | 0.00 | 0.72 | $0.00005 | 0.2s |

### OCR — Vision (images → blocs texte)

| # | Modèle | Provider | Blocs | Détec. | Texte | Types | Quality | Coût | Latence |
|---|--------|----------|-------|--------|-------|-------|---------|------|---------|
| 1 | Qwen3-VL-32B | OpenRouter | 28/29 | **0.97** | **0.74** | 0.96 | **0.85** | $0.002 | 4.6s |
| 2 | gemini-2.5-flash | Google | 29/29 | 1.00 | 0.73 | 0.88 | 0.84 | $0.001 | 18.6s |
| 3 | gpt-4o | OpenAI | 29/29 | **1.00** | 0.67 | 0.96 | 0.83 | $0.021 | 18.9s |
| 4 | gemini-2.5-pro | Google | 33/29 | 0.88 | 0.71 | 0.93 | 0.80 | $0.024 | 42.3s |
| 5 | Qwen3.5-397B | OpenRouter | 37/29 | 0.78 | 0.74 | 0.96 | 0.80 | $0.005 | 5.9s |
| 6 | Nemotron Nano VL | OpenRouter | 28/29 | **0.97** | 0.62 | 0.95 | 0.79 | **gratuit** | 6.0s |
| 7 | gpt-4o-mini | OpenAI | 29/29 | 1.00 | 0.60 | 0.90 | 0.78 | $0.012 | 25.3s |
| 8 | claude-sonnet-4-6 | Anthropic | 37/29 | 0.78 | 0.71 | 0.93 | 0.78 | $0.053 | 40.8s |
| 9 | Qwen3-VL-235B | OpenRouter | 25/29 | 0.86 | 0.70 | 0.83 | 0.78 | $0.003 | 10.9s |
| 10 | Llama 4 Maverick | OpenRouter | 34/29 | 0.85 | 0.67 | 0.92 | 0.78 | $0.002 | 3.9s |
| 11 | claude-haiku-4-5 | Anthropic | 31/29 | 0.83 | 0.64 | 0.88 | 0.75 | $0.017 | 20.2s |
| 12 | mistral-small (3.2) | Mistral | 42/29 | 0.69 | 0.65 | 0.92 | 0.72 | $0.001 | 18.3s |
| 13 | mistral-small | Mistral | 42/29 | 0.69 | 0.65 | 0.92 | 0.72 | $0.001 | 16.5s |
| 14 | Pixtral 12B | Mistral | 37/29 | 0.78 | 0.61 | 0.88 | 0.71 | $0.001 | 19.5s |
| 15 | mistral-large | Mistral | 40/29 | 0.72 | 0.61 | 0.93 | 0.71 | $0.027 | 33.0s |
| 16 | Llama 4 Scout | OpenRouter | 17/29 | 0.59 | 0.64 | 0.93 | 0.68 | $0.001 | 1.4s |
| 17 | Gemma 3 27B | OpenRouter | 17/29 | 0.59 | 0.61 | 0.75 | 0.63 | $0.000 | 4.1s |

---

## Conclusions

### Meilleurs modèles par catégorie

**IDP — Meilleure qualité :** Qwen3.5-397B (quality 0.88, zéro hallucination)
**IDP — Meilleur rapport qualité/prix :** mistral-small (quality 0.87, $0.00006/item)
**IDP — Plus rapide :** deepseek-chat (quality 0.85, 400ms) ou Gemma 3 27B (quality 0.84, 300ms)

**OCR — Meilleure qualité :** Qwen3-VL-32B (texte 0.74, détection 97%)
**OCR — Meilleur rapport qualité/prix :** gemini-2.5-flash (texte 0.73, $0.001)
**OCR — Gratuit :** Nemotron Nano VL (détection 97%, quality 0.79, $0)

### Combos recommandés pour la prod

| Stratégie | OCR | IDP | Coût/chapitre | Commentaire |
|-----------|-----|-----|---------------|-------------|
| **Best quality** | Qwen3-VL-32B | Qwen3.5-397B | ~$0.003 | Meilleure accuracy texte + structuration |
| **Best value** | gemini-2.5-flash | mistral-small | ~$0.002 | Très bon compromis, APIs stables |
| **Ultra-cheap** | Nemotron Nano VL | Gemma 3 27B | ~$0.000 | Gratuit, quality correcte |
| **Actuel (prod)** | claude-sonnet | claude-sonnet | ~$0.090 | 30x plus cher que les alternatives |

### Exécution locale (MacBook Pro M5 Max 128 Go)

Avec 128 Go de RAM unifiée et 614 Go/s de bande passante, les modèles suivants tournent en local via Ollama.

#### Disponibilité des modèles

| Modèle | Ollama | Vision OK | Commande |
|--------|--------|-----------|----------|
| **Qwen3-VL-32B** | Officiel | Oui | `ollama run qwen3-vl:32b` |
| **Qwen3-VL-8B** | Officiel | Oui | `ollama run qwen3-vl:8b` |
| **Gemma 3 27B** | Officiel | Oui (officiel uniquement) | `ollama run gemma3:27b` |
| **Pixtral 12B** | Community | Oui | `ollama run hf.co/EnlistedGhost/Pixtral-12B-Ollama-GGUF` |

> **Piege Gemma 3** : les GGUF communautaires (unsloth, bartowski) **perdent la vision** dans Ollama. Utiliser impérativement la version officielle de la library Ollama.

#### Vitesse estimée

| Modèle | Quant | VRAM | Tok/s réaliste |
|--------|-------|------|----------------|
| **Qwen3-VL-32B** | Q4 | ~18 Go | 22-28 tok/s |
| **Qwen3-VL-32B** | Q8 | ~32 Go | 13-16 tok/s |
| **Gemma 3 27B** | Q4 | ~15 Go | 28-34 tok/s |
| **Gemma 3 27B** | Q8 | ~27 Go | 16-19 tok/s |
| **Qwen3-VL-8B** | Q8 | ~8 Go | 55-65 tok/s |
| **Pixtral 12B** | Q4 | ~7 Go | 60-70 tok/s |

Formule : tok/s ≈ bande_passante / taille_modèle × 0.7 (overhead KV cache, framework, encodeur vision).

#### Temps estimé par chapitre (3 photos)

**Combo principal : Qwen3-VL-32B (Q4) + Gemma 3 27B (Q4)** — 33 Go total

| Étape | Modèle | Traitement | Génération | Total |
|-------|--------|-----------|------------|-------|
| OCR | Qwen3-VL-32B | ~10s (encodage 3 images) | ~60s (~1500 tokens à 25 tok/s) | **~70s** |
| IDP | Gemma 3 27B | ~2s (prefill 2500 tokens) | ~35s (~1000 tokens à 30 tok/s) | **~37s** |
| | | | **Pipeline total** | **~2 min** |

**Combo rapide : Qwen3-VL-8B (Q8) + Pixtral 12B (Q4)** — 15 Go total

| Étape | Modèle | Total estimé |
|-------|--------|-------------|
| OCR | Qwen3-VL-8B | ~30s |
| IDP | Pixtral 12B | ~15s |
| | **Pipeline total** | **~45s** |

#### Comparaison local vs API

| | Local (combo principal) | API (mêmes modèles) | API (gemini-flash + mistral-small) |
|---|---|---|---|
| **Temps/chapitre** | ~2 min | ~5s | ~25s |
| **Coût/chapitre** | $0 | ~$0.003 | ~$0.002 |
| **Données** | 100% locales | Cloud | Cloud |

Le local est ~20x plus lent que les APIs, mais gratuit et privé. Pour le Lot 0 (4 chapitres pilotes), 2 minutes par chapitre est acceptable — le pipeline ne tourne qu'une fois par upload.

### Observations notables

- **Claude Sonnet est le seul modèle avec des hallucinations** (8%) en IDP, et le plus cher en OCR ($0.053/run). Pas recommandé.
- **Les modèles MoE (Qwen3.5-397B, Llama 4 Maverick) ne rentrent pas en 128 Go** (~200 Go en Q4) — API obligatoire.
- **Les modèles Qwen3-VL dominent l'OCR** grâce à leur optimisation spécifique pour le parsing de documents et l'OCR multilingue.
- **Step 3.5 Flash hallucine** (20%) car il mélange raisonnement interne et sortie JSON — à éviter.
- **Les modèles "reasoning" (o3-mini, deepseek-reasoner)** nécessitent `max_completion_tokens` au lieu de `max_tokens`, et certains mettent le contenu dans un champ `reasoning` séparé.

### Métriques

**IDP :** Completeness (items trouvés/attendus), Classification (types corrects), Fidelity (traçabilité dans le texte source), Keywords (Jaccard), Hallucination rate, Notion coherence, Schema compliance.

**OCR :** Detection score (blocs trouvés/attendus), Text accuracy (chevauchement mots), Type accuracy (TEXT/DIAGRAM/TABLE correct).

Le **score composite** combine qualité (85%) et coût-efficacité (15%).
