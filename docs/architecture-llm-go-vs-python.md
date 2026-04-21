# Architecture LLM Pipeline — Go vs Agent Python

## Contexte

Le backend Révise Mieux est en Go (Gin, pgx, architecture hexagonale). Le pipeline LLM actuel (OCR → structuration) est implémenté comme un adaptateur Go dans `internal/infra/`. L'outil de benchmark (12 modèles, 3 approches, 5 types de bench) a mis en évidence des frictions récurrentes dans l'orchestration LLM en Go.

Ce document évalue l'intérêt d'introduire un agent Python (Pydantic + instructor) pour le pipeline LLM, et sous quelle forme.

---

## Problèmes concrets rencontrés en Go

Lors de la phase de benchmark (avril 2026), les points de friction suivants ont été identifiés :

### 1. Parsing JSON fragile

Les LLMs retournent du JSON enveloppé dans des markdown fences, tronqué quand `max_tokens` est insuffisant, ou placé dans `reasoning_content` au lieu de `content` pour les modèles reasoning.

Solutions implémentées en Go :
- `StripMarkdownFences()` (parsing regex)
- Fallback `reasoning_content` → `content`
- Augmentation manuelle de `max_tokens` par modèle

En Python avec instructor + Pydantic : géré automatiquement. L'output est validé contre un schéma Pydantic, et instructor retry automatiquement si le JSON est invalide.

### 2. Multi-provider boilerplate

Chaque provider LLM nécessite son propre package Go :
- `internal/infra/anthropic/` — client HTTP custom, types spécifiques
- `internal/infra/openaicompat/` — compatible OpenAI mais quirks par provider
- `internal/infra/mistral/` — API OCR dédiée, format différent

Total : ~800 lignes de code HTTP/JSON par provider.

En Python avec litellm : un seul appel `completion(model="gemini/gemini-2.5-flash", ...)` couvre tous les providers. Le switch de provider est un changement de string.

### 3. Retry et rate limiting

Le retry avec backoff exponentiel a été ajouté manuellement dans le runner du benchmark (429 Mistral, timeouts LM Studio). Chaque type de bench (IDP, OCR, E2E, hybrid) a sa propre boucle de retry.

En Python avec tenacity ou instructor : `@retry(wait=wait_exponential(), retry=retry_if_exception_type(RateLimitError))` — une ligne.

### 4. Image encoding et redimensionnement

L'envoi d'images aux VLMs a nécessité :
- Encodage base64 par image
- Redimensionnement adaptatif (max 2048px, réduit si budget tokens dépassé)
- Mode per-page pour RolmOCR (1 image par requête)
- Re-encodage JPEG qualité 85

Total : ~80 lignes de code Go avec `image/jpeg`. En Python : `PIL.Image.open().resize()` en 3 lignes.

### 5. Vitesse d'itération

Chaque modification du pipeline benchmark nécessite `go build` + relancer (~5s). Un test sur 3 cas prend 2-10 minutes. Le cycle d'itération sur un prompt ou un paramètre est lent.

En Python : modification du fichier → exécution immédiate. Les notebooks Jupyter permettent de tester un cas isolé en quelques secondes.

---

## Forces de Go pour ce projet

### 1. Architecture hexagonale en place

Les appels LLM sont isolés derrière des interfaces du domaine :
- `chapter.LLMService` — structuration
- `chapter.OCRService` — extraction OCR
- `session.Scorer` — scoring des réponses

Le domaine métier (Mastery, Session, Validation) ne connaît pas les providers. Changer d'implémentation LLM ne touche pas le domaine.

### 2. Domaine métier mature

Le coeur métier est en Go : machine à états Mastery, répétition espacée, pipeline J0, HITL, lazy generation. Réécrire tout ça en Python serait une régression.

### 3. Déploiement simple

Un binaire Go statique, sans runtime, sans virtualenv, sans gestion de dépendances Python. Le déploiement est un `docker build` + `docker run`.

### 4. Performance serveur

Pour servir l'API mobile (Expo/React Native), Go est nettement plus performant que Python (FastAPI) en latence et en consommation mémoire.

---

## Analyse par composant

| Composant | Go | Python + Pydantic |
|-----------|----|--------------------|
| Domain logic (Mastery, Session) | Bien adapté | Non pertinent |
| API HTTP / handlers | Bien adapté | Overkill |
| SQL / repositories | Bien adapté (pgx) | Non pertinent |
| **Orchestration LLM (pipeline)** | **Verbose et fragile** | **Excellente DX** |
| **Structured output validation** | **Manuel (json.Unmarshal)** | **Automatique (Pydantic)** |
| **Multi-provider switching** | **1 package par provider** | **1 ligne (litellm)** |
| **Benchmark / expérimentation** | **Itération lente** | **Itération rapide** |
| Déploiement production | 1 binaire | Service supplémentaire |

Le problème n'est pas Go vs Python. C'est que le code d'orchestration LLM (retry, structured output, multi-provider, image handling) est un domaine où Python a un avantage d'écosystème majeur, tandis que Go excelle sur tout le reste.

---

## Options architecturales

### Option A — Tout en Go (amélioration de l'existant)

```
┌──────────────────────────────────────────┐
│              Go backend                   │
│                                          │
│  ┌─────────┐  ┌──────────┐  ┌────────┐ │
│  │ Domaine │←─│ App svc  │←─│ HTTP   │ │
│  │         │  │          │  │        │ │
│  └─────────┘  └────┬─────┘  └────────┘ │
│                    │                     │
│              ┌─────▼──────┐              │
│              │ LLM Adapter│              │
│              │ (amélioré) │              │
│              └─────┬──────┘              │
│                    │                     │
└────────────────────┼─────────────────────┘
                     ▼
              APIs LLM (cloud)
```

Améliorations :
- Factoriser les providers en un seul adapter configurable
- Ajouter retry/backoff dans l'adapter
- Structured output : écrire un helper `UnmarshalLLMResponse[T any]()` avec strip fences + retry

**Effort** : ~3 jours.
**Avantage** : pas de changement d'architecture, 1 seul binaire.
**Inconvénient** : toujours verbose, chaque nouveau modèle ou quirk nécessite du code Go.

### Option B — Microservice Python pour le LLM pipeline

```
┌──────────────────┐     ┌──────────────────────┐
│   Go backend     │     │  Python LLM service   │
│                  │     │                      │
│  ┌────────────┐  │     │  instructor          │
│  │ Domaine    │  │     │  + litellm           │
│  │ + App svc  │──┼────▶│  + Pydantic schemas  │
│  │ + HTTP     │  │     │  + image processing  │
│  └────────────┘  │     │                      │
│                  │     │  POST /extract        │
│                  │     │  POST /ocr            │
└──────────────────┘     └──────────┬───────────┘
                                    │
                                    ▼
                             APIs LLM (cloud)
```

Le service Python expose 2-3 endpoints :
- `POST /ocr` — images → blocs texte (RolmOCR ou cloud)
- `POST /extract` — images + blocs → items structurés (pipeline hybrid)
- `POST /score` — question + réponse → score

Le Go backend appelle ce service comme un adaptateur HTTP (même pattern que les providers actuels).

```python
# Exemple : structuration avec instructor + Pydantic
from instructor import from_litellm
from litellm import completion
from pydantic import BaseModel

class Item(BaseModel):
    type: str  # KNOWLEDGE, PROCEDURE, DOCUMENT, WRITING
    term: str
    keywords: list[str]
    steps: list[str] = []
    notion_name: str
    confidence: float

class ExtractionResult(BaseModel):
    items: list[Item]
    notions: list[str]

client = from_litellm(completion)

result = client.chat.completions.create(
    model="gemini/gemini-2.5-flash",
    response_model=ExtractionResult,  # Pydantic validation automatique
    messages=[
        {"role": "system", "content": system_prompt},
        {"role": "user", "content": [
            {"type": "image_url", "image_url": {"url": f"data:image/jpeg;base64,{img}"}},
            {"type": "text", "text": user_prompt},
        ]},
    ],
    max_retries=3,  # retry automatique si JSON invalide
)
# result est un objet Python typé, validé par Pydantic
```

**Effort** : ~2 jours.
**Avantage** : itération rapide, validation automatique, multi-provider trivial, retry built-in.
**Inconvénient** : 1 service de plus (déploiement, monitoring, healthcheck).

### Option C — Agent Python pour le benchmark uniquement

```
┌──────────────────┐
│   Go backend     │  ← production, inchangé
│   (inchangé)     │
└──────────────────┘

┌──────────────────────┐
│  Python benchmark    │  ← expérimentation, pas en prod
│                      │
│  instructor          │
│  + litellm           │
│  + Pydantic schemas  │
│  + notebooks Jupyter │
│                      │
│  bench.py --type=e2e │
└──────────────────────┘
```

Le benchmark actuel (~1500 lignes Go, `cmd/benchmark/`) est réécrit en Python (~200 lignes). Le Go backend reste inchangé en production. Le Python sert uniquement à itérer sur les prompts, tester de nouveaux modèles, et comparer les approches.

Une fois le pipeline stabilisé (prompt final, modèle choisi, approche validée), les paramètres optimaux sont reportés dans la config Go.

**Effort** : ~1 jour.
**Avantage** : itération 10x plus rapide, zéro impact prod, notebooks pour l'analyse.
**Inconvénient** : le code benchmark diverge du code prod (risque de comportement différent).

---

## Estimation de l'écart de code

Pour illustrer la différence, voici la même opération (pipeline hybrid) en Go vs Python :

### Go actuel (~120 lignes, 5 fichiers)

```go
// benchmark_e2e.go — provider E2E
// benchmark_hybrid.go — provider Hybrid (supprimé/refactoré)
// benchmark_ocr_raw.go — provider OCR raw text
// ocr.go — OCR processor (vision request, per-page, doVisionRequest)
// benchmark.go — IDP provider (chat request, retry)
// + main.go runner (~200 lignes pour le type hybrid)
```

### Python avec instructor + litellm (~30 lignes)

```python
import instructor, litellm
from pydantic import BaseModel

class Item(BaseModel):
    type: str
    term: str
    keywords: list[str]
    steps: list[str] = []
    notion_name: str
    confidence: float

class Result(BaseModel):
    items: list[Item]
    notions: list[str]

def extract(images: list[str], ocr_blocks: str, model: str) -> Result:
    client = instructor.from_litellm(litellm.completion)
    content = [{"type": "image_url", "image_url": {"url": img}} for img in images]
    content.append({"type": "text", "text": f"Blocs OCR:\n{ocr_blocks}\n\nExtrais les items."})
    
    return client.chat.completions.create(
        model=model,
        response_model=Result,
        messages=[
            {"role": "system", "content": SYSTEM_PROMPT},
            {"role": "user", "content": content},
        ],
        max_retries=3,
    )
```

Le ratio est d'environ **5:1** en lignes de code pour la même fonctionnalité. La différence vient principalement de la validation JSON automatique, du multi-provider intégré, et du retry built-in.

---

## Recommandation

### Court terme (maintenant)

**Option C** — Agent Python pour le benchmark uniquement.

Réécrire le benchmark en Python pour itérer rapidement sur :
- Les hypothèses H1-H10 identifiées dans le benchmark
- Les prompts (few-shot, spécialisé par matière)
- Les nouveaux modèles (cycle de release = 2-4 semaines)

Le Go backend reste inchangé en production.

### Moyen terme (pipeline validé)

**Option B** — Microservice Python pour le pipeline LLM.

Une fois l'approche, le modèle, et les prompts stabilisés, extraire le pipeline LLM dans un service Python dédié :
- Endpoint `/extract` appelé par le Go backend
- Pydantic pour la validation, instructor pour le retry
- litellm pour le multi-provider + failover
- Déployé comme un sidecar ou un service ECS séparé

Le Go backend conserve le domaine métier, l'API HTTP, et le SQL. Il appelle le service Python comme un adaptateur externe (même pattern que S3 ou Redis).

### Ce qu'il ne faut pas faire

- Réécrire le domaine métier Go en Python (régression, perte de la type safety, architecture hexagonale cassée)
- Ajouter Python comme dépendance dans le binaire Go (CGo, subprocess = fragilité)
- Maintenir deux implémentations du même pipeline (Go prod + Python bench) sur le long terme — l'une des deux doit être la source de vérité

---

## Impact sur le déploiement (Option B)

| | Aujourd'hui (Go seul) | Avec service Python |
|---|---|---|
| **Containers** | 1 (Go API) | 2 (Go API + Python LLM) |
| **RAM** | ~50MB | ~50MB + ~200MB (Python + modèle) |
| **Healthcheck** | 1 endpoint | 2 endpoints |
| **CI/CD** | 1 pipeline | 2 pipelines (ou monorepo multi-stage) |
| **Logs** | 1 source | 2 sources (corrélation par request ID) |
| **Complexité ops** | Simple | Moyenne |

Le surcoût opérationnel est réel mais gérable avec un docker-compose ou un ECS task definition à 2 containers.

---

## Supervision, troubleshooting et exploitation

Ce point est souvent sous-estimé. En production, le pipeline LLM n'est pas un appel API statique — c'est un système vivant dont le comportement change à chaque mise à jour de modèle, chaque variation de prompt, et chaque nouveau type de document.

### Le problème en Go aujourd'hui

En Go, la supervision LLM se résume à `log.Printf`. Quand un appel échoue ou produit un résultat de mauvaise qualité :

1. **On ne sait pas pourquoi.** Le prompt envoyé, la réponse brute du LLM, les tokens consommés, le nombre de retries — tout ça est perdu sauf si on l'a loggé manuellement.
2. **On ne peut pas reproduire.** Sans le prompt exact + les images exactes + le modèle exact, le bug est non reproductible.
3. **On ne détecte pas la dérive.** Si Mistral met à jour `mistral-small-latest` et que la qualité baisse de 10%, on ne le voit qu'en production quand les utilisateurs se plaignent.
4. **On ne peut pas mesurer le coût réel.** Les tokens consommés par appel, par utilisateur, par chapitre — il faut les calculer manuellement.

### Ce que l'écosystème Python apporte

#### Observabilité (LangFuse, LangSmith, Braintrust)

Ces outils s'intègrent en **une ligne** avec litellm ou instructor et capturent automatiquement :

```python
# Activation LangFuse — chaque appel LLM est tracé automatiquement
litellm.success_callback = ["langfuse"]

# Chaque appel enregistre :
# - Le prompt exact envoyé (system + user + images)
# - La réponse brute du LLM
# - Les tokens input/output et le coût
# - La latence
# - Les retries (combien, pourquoi)
# - Les erreurs Pydantic (quels champs étaient invalides)
# - Le modèle et la version
# - Un trace ID corrélable avec le Go backend
```

En Go, instrumenter chaque appel LLM avec ce niveau de détail nécessiterait ~200 lignes de middleware custom par provider.

#### Troubleshooting en production

Quand un document produit un mauvais résultat :

| Étape | En Go | Avec LangFuse/Python |
|-------|-------|---------------------|
| Trouver l'appel | Chercher dans les logs texte | Dashboard : filtrer par document_id |
| Voir le prompt envoyé | Pas loggé (trop gros) | Affiché en entier avec les images |
| Voir la réponse brute | Pas loggée | Affichée + diff avec le JSON parsé |
| Comprendre pourquoi le JSON a échoué | `JSON parse error: unexpected end` | Détail : "champ `keywords` manquant, retry #2 a réussi après 3.2s" |
| Rejouer l'appel | Impossible (pas le prompt exact) | 1 clic : "Replay" dans le dashboard |
| Coût de cet appel | Pas tracé | $0.0034 (12K tokens in, 2K out) |

#### Détection de dérive qualité

Les providers cloud mettent à jour leurs modèles sans préavis. `mistral-small-latest` a changé de Mistral Small 3.1 (24B) à Mistral Small 4 (119B) pendant ce benchmark — sans aucune notification.

Avec un outil de supervision :

```python
# Eval automatique sur un dataset de référence
# Exécuté par un cron (quotidien ou hebdomadaire)
from langfuse import Langfuse

langfuse = Langfuse()

for case in reference_cases:
    result = extract(case.images, case.ocr_blocks, model="gemini/gemini-2.5-flash")
    score = evaluate(result, case.golden)
    
    langfuse.score(
        trace_id=result.trace_id,
        name="completeness",
        value=score.completeness,
    )
    # Alerte si le score moyen chute de >5% par rapport à la semaine précédente
```

En Go, il faudrait construire tout ce pipeline de monitoring from scratch.

#### Monitoring des coûts

```python
# litellm tracke automatiquement les coûts par modèle, par appel
# Accessible via callbacks ou dashboard LangFuse

# Budget alerting :
litellm.max_budget = 100.0  # $100/mois max
litellm.budget_duration = "monthly"
# → lève une exception si le budget est dépassé
```

En Go : calcul manuel `tokens * price_per_million / 1_000_000` dans chaque provider, agrégation en base, dashboard custom.

### Boucle de feedback (HITL → amélioration du pipeline)

Le pipeline Révise Mieux prévoit une validation humaine (parent/admin) des items générés. Cette validation est une mine d'or pour améliorer le pipeline — si on peut la capturer.

#### Le cycle vertueux

```
Photo cahier → OCR → Structuration → Items bruts
                                        │
                                        ▼
                                  Validation HITL
                                  (parent corrige)
                                        │
                                        ▼
                              ┌─────────────────────┐
                              │ Dataset de feedback  │
                              │                     │
                              │ - Items corrigés    │
                              │ - Items supprimés   │
                              │ - Items ajoutés     │
                              │ - Annotations       │
                              └────────┬────────────┘
                                       │
                          ┌────────────┼────────────┐
                          ▼            ▼            ▼
                    Few-shot      Eval auto    Fine-tuning
                    (H2)          (drift)      (futur)
```

#### En pratique avec l'écosystème Python

**1. Capturer le feedback**

Chaque correction HITL est stockée comme un couple (input, output_corrigé) :

```python
# Quand un parent corrige un item
langfuse.score(
    trace_id=original_trace_id,
    name="human_validation",
    value=0.0,  # rejeté
    comment="Item hallucinate - le texte n'apparaît pas dans le cahier"
)
```

**2. Construire un dataset d'évaluation**

Les corrections HITL deviennent des cas de test golden. Après 100 validations, on a un dataset de 100 documents avec la "bonne réponse" validée par un humain.

```python
# Exporter les corrections en dataset d'évaluation
dataset = langfuse.get_dataset("hitl-corrections")
# 100 cas avec input (images + OCR) + output attendu (items corrigés)
```

**3. Évaluer automatiquement les changements**

Avant de déployer un nouveau prompt ou un nouveau modèle :

```python
for case in dataset.items:
    result = extract(case.input, model="gemini/gemini-2.5-flash")
    scores = evaluate(result, case.expected_output)
    # Si le nouveau prompt dégrade le score sur les cas HITL → ne pas déployer
```

**4. Alimenter le few-shot (H2)**

Les meilleurs exemples HITL deviennent des few-shot dans le prompt :

```python
# Sélectionner les 2 cas HITL les plus représentatifs
best_examples = dataset.select(n=2, strategy="diverse")
# Les injecter dans le system prompt comme exemples
system_prompt = base_prompt + "\n\nExemples:\n" + format_examples(best_examples)
```

**5. Fine-tuning (futur)**

Quand le dataset atteint ~500 exemples validés, fine-tuner un modèle open-source (Qwen3-VL-32B, Mistral Small) sur les corrections HITL. Le modèle apprend les spécificités du format de cahier, les erreurs courantes d'OCR, et le style de structuration attendu.

En Go, chacune de ces étapes nécessiterait une implémentation custom. En Python, c'est un assemblage de briques existantes (LangFuse, instructor, litellm, HuggingFace datasets).

### Comparaison supervision Go vs Python

| Capacité | Go (custom) | Python (écosystème) |
|----------|-------------|---------------------|
| Logging des appels LLM | `log.Printf` | LangFuse/LangSmith auto-trace |
| Visualisation prompt/réponse | Logs texte | Dashboard web interactif |
| Replay d'un appel | Impossible | 1 clic |
| Coût par appel | Calcul manuel | Automatique (litellm) |
| Budget alerting | Custom | `litellm.max_budget` |
| Détection de dérive | Cron + SQL custom | Eval dataset + alerting |
| Capture feedback HITL | INSERT SQL | LangFuse scores + dataset |
| A/B testing prompts | Feature flags + code | LangFuse experiments |
| Dataset d'évaluation | JSON files | LangFuse datasets / HF datasets |
| Few-shot from feedback | Manuel | Automatisable |
| Fine-tuning | Non réaliste en Go | HuggingFace / Unsloth |
| **Effort total** | **~2-4 semaines** | **~2 jours** |

### Conclusion sur la supervision

La supervision n'est pas un nice-to-have. C'est ce qui transforme un prototype ("ça marche sur 3 cas de test") en un système de production fiable ("on détecte et corrige les problèmes avant les utilisateurs").

L'écosystème Python (LangFuse, instructor, litellm) fournit cette couche de supervision quasi gratuitement. En Go, chaque brique doit être construite from scratch, ce qui explique pourquoi la plupart des pipelines LLM en production utilisent Python pour l'orchestration, même quand le backend est en Go, Java ou Rust.
