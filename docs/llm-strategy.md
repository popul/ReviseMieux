# Stratégie LLM — Révise Mieux

> Document technique de référence pour les choix de modèles, l'optimisation des coûts, la stratégie de test et le monitoring des appels LLM.
> Dernière mise à jour : 2026-03-12.

---

> **Mise à jour mars 2026 — Choix de modèles révisés après benchmark**
>
> Les recommandations de modèles dans ce document (Sonnet pour structuration, Haiku pour fidelity/questions) sont **obsolètes**. Le benchmark réel sur 21 modèles a montré :
>
> - **Sonnet hallucine** (8%, seul modèle concerné) et coûte 30-50x plus cher que les alternatives.
> - **mistral-small** et **Qwen3.5-397B** offrent une meilleure qualité (0.87-0.88) à $0.00006-0.00011/item.
> - **L'exécution locale est viable** pour le Lot 0 (Qwen3-VL-32B + Gemma 3 27B via Ollama, ~2 min/chapitre, $0).
>
> Voir [`backend/testdata/benchmark/README.md`](../backend/testdata/benchmark/README.md) pour les résultats complets et les combos recommandés.
>
> L'architecture (ports, prompts versionnés, logging, monitoring, tests) décrite ci-dessous reste valide — seuls les choix de modèles changent.

---

## 1. Cas d'usage LLM

| # | Cas d'usage | AC | Interface Go | Modèle (initial) | Modèle (post-benchmark) | Latence cible | Temps réel ? |
|---|------------|-----|-------------|--------|--------|---------------|-------------|
| A | Structuration OCR → Items + Notions | Z2-AC01, Z7-AC15 | `chapter.LLMService.StructureBlocks()` | ~~Sonnet 4.6~~ | **Qwen3.5-397B** ou **mistral-small** | ≤ 60s/page | Non (pipeline J0) |
| B | Fidelity check (fidélité sémantique) | Z3-AC10 | À créer | **Haiku 4.5** | À benchmarker | ≤ 5s/item | Non (pipeline J0) |
| C | Détection de cohérence (doublons, contradictions) | Z3-AC11 | À créer | ~~Sonnet 4.6~~ | À benchmarker | ≤ 10s/chapitre | Non (pipeline J0) |
| D | Génération de questions (lazy generation) | Z4-AC01 | À créer | **Haiku 4.5** | À benchmarker | ≤ 500ms | Oui (session) |
| E | Feedback enrichi | Z4-AC09 | `feedback.go` | **Templates** (pas de LLM) | Inchangé | ≤ 200ms | Oui |
| F | Multimodal visuel (schémas, graphiques) | Z4-AC17 | Post-MVP | ~~Sonnet 4.6~~ | À benchmarker | ≤ 2s | Non |

---

## 2. Choix des modèles — Justification détaillée

### 2.1 Sonnet 4.6 — Structuration, cohérence, multimodal

**Tarifs** : $3.00 / 1M tokens input, $15.00 / 1M tokens output.

**Pourquoi Sonnet pour la structuration (cas A) :**
- L'extraction demande de la **synthèse multi-source** : plusieurs blocs OCR doivent être combinés en items cohérents.
- La **classification nuancée** entre types (KNOWLEDGE vs PROCEDURE vs DOCUMENT vs WRITING) nécessite une compréhension sémantique fine.
- La **fidélité au texte source** est critique — les hallucinations produisent des items faux étudiés par des collégiens.
- Le **JSON structuré** produit doit respecter un schéma strict (`StructuredItem` avec `Type`, `Term`, `Keywords`, `Steps`, `NotionName`, `Confidence`).
- Sonnet 4.6 est quasi au niveau d'Opus pour la compréhension de documents structurés, à 40% du prix.

**Pourquoi pas Haiku :** Haiku est "pragmatique" — il produit du JSON valide mais avec plus d'approximations sur les classifications et un risque accru d'hallucinations subtiles (terme absent du texte source mais plausible).

**Pourquoi pas Opus :** Le surcoût (+67%) ne se justifie pas pour de l'extraction structurée. Sonnet ferme le gap sur les benchmarks de compréhension de documents.

**Pourquoi Sonnet pour la cohérence (cas C) :**
- Détecter des **contradictions sémantiques** entre items ("densité = m/V" vs erreur "densité = V/m") nécessite un raisonnement multi-step.
- Faible volume (1 appel/chapitre) → le surcoût vs Haiku est négligeable.

### 2.2 Haiku 4.5 — Fidelity check, génération de questions

**Tarifs** : $1.00 / 1M tokens input, $5.00 / 1M tokens output.

**Pourquoi Haiku pour le fidelity check (cas B) :**
- Tâche **bien cadrée** : comparer un item à son texte OCR source → score 0-1 + justification courte.
- Essentiellement de la **classification binaire** avec explication.
- **Volume élevé** : 1 appel par item généré = 5-20 appels/page. Le coût est multiplié par le volume.
- Haiku excelle sur les tâches de classification à prompt bien structuré.

**Pourquoi Haiku pour la génération de questions (cas D) :**
- **Latence critique** (≤ 500ms) — Haiku est 3-5x plus rapide que Sonnet.
- Tâche bien structurée : item + template → question rendue avec réponse attendue.
- Les questions sont **cachées 24h** → les erreurs rares sont corrigées au cycle suivant.
- Types de questions standardisés (CLOZE, MCQ, DEF_SHORT) → peu d'ambiguïté.

### 2.3 Templates — Feedback enrichi (pas de LLM)

**Pourquoi ne pas utiliser de LLM pour le feedback (cas E) :**
- Latence < 200ms requise — incompatible avec un appel LLM réseau.
- Le feedback doit être **factuel, jamais culpabilisant** (Z4-AC09). Un LLM ajoute de la variance indésirable.
- Cohérence parfaite : même feedback pour la même erreur = prévisibilité pédagogique.
- Coût nul.

**Évolution possible :** Si les feedbacks template s'avèrent trop génériques après le MVP, migrer vers Haiku avec un prompt très contraint + guardrails de tonalité. Décision basée sur les métriques d'engagement post-lancement.

---

## 3. Estimation des coûts

### 3.1 Coût par appel (sans optimisation)

| Cas d'usage | Tokens input | Tokens output | Modèle | Coût/appel |
|------------|-------------|---------------|--------|-----------|
| Structuration | ~1 500 | ~800 | Sonnet | ~$0.017 |
| Fidelity check | ~500 | ~100 | Haiku | ~$0.001 |
| Cohérence | ~3 000 | ~500 | Sonnet | ~$0.017 |
| Questions | ~300 | ~200 | Haiku | ~$0.001 |

### 3.2 Coût mensuel estimé

**Hypothèses** : 50 élèves actifs, 2 uploads/semaine/élève (10 pages chacun), 4 sessions/semaine/élève (15 questions/session).

| Cas d'usage | Volume mensuel | Coût brut | Avec optimisations | Coût optimisé |
|------------|---------------|-----------|--------------------|----|
| Structuration | 4 000 pages | $68 | Batch (-50%) + cache (-90% system) | **~$20** |
| Fidelity check | 40 000 items | $40 | Batch (-50%) + cache (-90% system) | **~$6** |
| Cohérence | 400 chapitres | $7 | Cache (-90% system) | **~$4** |
| Questions | 12 000 questions | $12 | Cache (-90% system) | **~$4** |
| **Total** | | **$127** | | **~$34/mois** |

**Scaling** : pour 500 élèves → ~$340/mois. Pour 5 000 élèves → ~$3 400/mois.

### 3.3 Optimisations de coût appliquées

#### Batch API (-50%)

Les cas **non temps-réel** (structuration, fidelity check) utilisent le [Batch API](https://platform.claude.com/docs/en/about-claude/pricing) d'Anthropic :
- Soumission de lots de requêtes (max 100 000 requêtes ou 256 Mo).
- Résultats disponibles sous quelques minutes à quelques heures.
- **-50% sur tous les tokens** (input et output).
- Compatible avec le pipeline J0 qui n'est pas temps réel.

#### Prompt caching (-90% en lecture)

Le [prompt caching](https://platform.claude.com/docs/en/build-with-claude/prompt-caching) est utilisé pour tous les appels :
- Le **system prompt** (instructions d'extraction, schéma JSON, consignes pédagogiques) est identique pour tous les appels d'un même type.
- Première écriture : 1.25x du prix de base (cache 5 min) ou 2x (cache 1h).
- Lectures suivantes : **0.1x du prix de base** → économie de 90%.
- Break-even dès 2 appels avec le même préfixe.

**Implémentation** : ajouter `cache_control: {"type": "ephemeral"}` sur le bloc system du message API.

#### Combinaison batch + cache

Les deux optimisations se cumulent. Pour la structuration en batch avec cache :
- Input : $3.00 × 0.5 (batch) × 0.1 (cache read) = **$0.15 / 1M tokens**
- Output : $15.00 × 0.5 (batch) = **$7.50 / 1M tokens**

---

## 4. Architecture du client Anthropic

### 4.1 Structure des fichiers

```
backend/internal/infra/anthropic/
├── client.go           # Client HTTP Anthropic (auth, retry, timeout)
├── structurer.go       # Implémente chapter.LLMService (Sonnet)
├── fidelity.go         # Fidelity check (Haiku)
├── question_gen.go     # Génération de questions (Haiku)
├── prompts.go          # Réexporte les prompts depuis internal/infra/llm
└── client_test.go      # Tests de contrat avec golden files
```

### 4.2 Configuration

```go
type Config struct {
    APIKey              string  // ANTHROPIC_API_KEY
    StructurationModel  string  // "claude-sonnet-4-6-20250217"
    FidelityModel       string  // "claude-haiku-4-5-20251001"
    QuestionGenModel    string  // "claude-haiku-4-5-20251001"
    CoherenceModel      string  // "claude-sonnet-4-6-20250217"
    MaxRetries          int     // 2 (avec backoff exponentiel)
    StructurationTimeout time.Duration // 60s
    FidelityTimeout      time.Duration // 10s
    QuestionGenTimeout   time.Duration // 5s
}
```

### 4.3 Routing de modèle

Le client route automatiquement vers le bon modèle selon le type d'appel. Chaque appel est tagué avec :
- `llm_model_version` : identifiant exact du modèle (`claude-sonnet-4-6-20250217`)
- `prompt_template_version` : hash SHA-256 du prompt template utilisé

Conformément à **Z2-AC13**, ces métadonnées sont persistées sur chaque `Item` et `Question` générés.

### 4.4 Gestion des erreurs

| Erreur | Comportement | Référence AC |
|--------|-------------|-------------|
| Timeout structuration | Page marquée `items_generation_failed`, blocs OCR conservés | Z2-AC05 |
| Timeout fidelity check | Item conservé avec `fidelity_score = null` | Z3-AC10 |
| Quota dépassé | Retry x2 avec backoff, puis échec pipeline page | Z2-AC05 |
| Réponse malformée (JSON invalide) | Retry x1, puis échec pipeline page | Z2-AC05 |
| LLM indisponible (session) | Fallback mode "Relecture active" (flashcards statiques) | Z4-AC12 |

---

## 5. Versioning des prompts (Z2-AC13)

### 5.1 Principe

Chaque prompt template est versionné dans le code via une constante :

```go
const (
    StructurationPromptV1 = "v1.0.0"
    FidelityPromptV1      = "v1.0.0"
    QuestionGenPromptV1   = "v1.0.0"
)
```

Le hash SHA-256 du contenu du prompt est calculé au démarrage et utilisé comme `prompt_template_version` dans les résultats.

### 5.2 Traçabilité

Chaque résultat LLM persiste :
- Le modèle exact utilisé (champ `llm_model_version` sur `Item` et `Question`)
- La version du prompt (champ `prompt_template_version`)
- Le timestamp de l'appel

Cela permet de diagnostiquer les régressions : "depuis quand les fidelity_score baissent-ils ?" → corrélation avec un changement de modèle ou de prompt.

---

## 6. Stratégie de test

### 6.1 Niveau 1 — Tests de contrat (CI, chaque PR)

**Build tag** : `//go:build integration`
**Pas d'appel API réel** — utilisation de golden files.

**Objectif** : valider que le parsing des réponses LLM et la logique de mapping vers le domaine fonctionnent correctement.

**Ce qui est testé :**
- Le JSON de réponse est parseable
- Tous les champs obligatoires sont présents (`Type`, `Term`, `Keywords`)
- `Confidence` est dans [0, 1]
- `ItemType` est une valeur valide de l'enum
- Au moins 1 item retourné pour un input non-vide
- Les keywords sont non-vides pour les items KNOWLEDGE
- Les steps sont non-vides pour les items PROCEDURE

**Golden files** stockés dans `backend/testdata/llm/` :
```
testdata/llm/
├── structuration_physique_input.json    # OCR blocks d'entrée
├── structuration_physique_golden.json   # Réponse LLM de référence
├── structuration_svt_input.json
├── structuration_svt_golden.json
├── fidelity_check_golden.json
└── question_gen_golden.json
```

**Exemple de test :**
```go
func TestStructureBlocks_Contract(t *testing.T) {
    golden := loadGoldenFile(t, "structuration_physique_golden.json")
    result := parseStructurationResponse(golden)

    for _, item := range result.Items {
        require.NotEmpty(t, item.Type, "type required")
        require.NotEmpty(t, item.Term, "term required")
        require.True(t, item.Confidence >= 0 && item.Confidence <= 1)
        require.Contains(t, validItemTypes, item.Type)
    }
}
```

### 6.2 Niveau 2 — Tests de régression prompt (nightly, gated)

**Build tag** : `//go:build llm`
**Appels API réels** — nécessite `ANTHROPIC_API_KEY`.

**Objectif** : détecter les régressions de qualité quand le modèle ou le prompt change.

**Ce qui est testé :**
- **Assertions structurelles** : nombre d'items dans une fourchette (±30%), types d'items attendus présents, keywords non vides.
- **Assertions sémantiques** : le terme principal est présent dans le texte OCR source (anti-hallucination).
- **Cohérence inter-runs** : le même input produit des résultats structurellement similaires (±20% sur le nombre d'items).

**Dataset de référence** : 10-15 golden inputs couvrant les cas types :
- Physique-Chimie (formules, procédures de calcul)
- SVT (vocabulaire, schémas)
- Histoire-Géographie (dates, concepts)
- Mathématiques (théorèmes, procédures)

**Seuils de blocage** :
- Nombre d'items < 50% de la référence → FAIL
- Aucun item KNOWLEDGE détecté → FAIL
- `avg(confidence)` < 0.5 → FAIL
- Terme hallucine (absent du texte source) → WARNING

**Exécution** : CI nightly ou pré-release, jamais sur chaque PR (coût + latence).

### 6.3 Niveau 3 — Monitoring en production

**Objectif** : détecter les dégradations silencieuses en temps réel.

#### 6.3.1 Métriques collectées

Chaque appel LLM émet les métriques suivantes :

| Métrique | Type | Description |
|---------|------|-------------|
| `llm.call.duration_ms` | Histogram | Latence de l'appel |
| `llm.call.tokens_input` | Counter | Tokens input consommés |
| `llm.call.tokens_output` | Counter | Tokens output consommés |
| `llm.call.cost_usd` | Counter | Coût estimé de l'appel |
| `llm.call.status` | Counter | Succès / erreur / timeout (par type d'appel) |
| `llm.structuration.items_count` | Histogram | Nombre d'items générés par page |
| `llm.structuration.avg_confidence` | Gauge | Confiance moyenne des items générés |
| `llm.fidelity.score` | Histogram | Distribution des fidelity scores |
| `llm.fidelity.low_rate` | Gauge | % d'items avec fidelity_score < 0.5 |
| `llm.question.cache_hit_rate` | Gauge | % de questions servies depuis le cache |
| `llm.model_version` | Label | Version du modèle sur chaque métrique |
| `llm.prompt_version` | Label | Version du prompt sur chaque métrique |

#### 6.3.2 Alertes

| Alerte | Condition | Sévérité | Action |
|--------|-----------|----------|--------|
| Dégradation fidelity | `avg(fidelity_score) < 0.6` sur 24h glissantes | **CRITICAL** | Notification admin. Vérifier `llm_model_version` et `prompt_template_version`. Comparer avec les derniers résultats stables. |
| Hausse hallucinations | `rate(fidelity_score < 0.5) > 15%` sur 24h | **CRITICAL** | Pause automatique de la génération d'items. Admin doit valider manuellement. |
| Latence structuration | `p95(duration_ms) > 30 000` sur 1h | **WARNING** | Vérifier quotas Anthropic API. Potentiel rate limiting. |
| Latence questions | `p95(duration_ms) > 2 000` sur 1h | **WARNING** | Vérifier quotas. Impact direct sur l'UX élève. |
| Coût journalier excessif | `sum(cost_usd) > budget_daily * 1.5` | **WARNING** | Potentiel loop infini ou abus. Vérifier les logs d'appels. |
| Changement de modèle | `llm_model_version` différent du dernier connu | **INFO** | Log l'événement. Déclencher les tests de régression nightly. |
| Timeout rate | `rate(status=timeout) > 10%` sur 1h | **WARNING** | Vérifier la santé de l'API Anthropic. Considérer le fallback Z4-AC12. |

#### 6.3.3 Dashboard admin

Le dashboard admin expose les vues suivantes :

1. **Vue temps réel** : appels/min, latence p50/p95, taux d'erreur, coût cumulé du jour.
2. **Vue qualité** : courbe de `avg(fidelity_score)` par jour, segmentée par `llm_model_version`. Permet de détecter visuellement un changement de comportement après une mise à jour Anthropic.
3. **Vue coûts** : coût par type d'appel (structuration, fidelity, questions), tendance hebdomadaire, projection mensuelle.
4. **Vue drift** : comparaison automatique des métriques de qualité entre la version de prompt actuelle et la précédente.

#### 6.3.4 Job hebdomadaire de comparaison (Z2-AC13)

Un job CRON hebdomadaire :
1. Sélectionne les 50 derniers items générés avec la version de prompt courante.
2. Calcule : `avg(fidelity_score)`, `avg(confidence)`, distribution des types d'items.
3. Compare avec les mêmes métriques de la semaine précédente.
4. Si dégradation > 15% sur un indicateur → alerte admin `LLM_QUALITY_REGRESSION`.
5. Log complet persisté pour audit.

---

## 7. Implémentation du logging LLM

### 7.1 Structure du log

Chaque appel LLM produit un log structuré (JSON) :

```json
{
  "timestamp": "2026-03-12T14:30:00Z",
  "call_type": "structuration",
  "model": "claude-sonnet-4-6-20250217",
  "prompt_version": "v1.0.0",
  "prompt_hash": "sha256:abc123...",
  "tokens_input": 1523,
  "tokens_output": 847,
  "duration_ms": 3200,
  "cost_usd": 0.017,
  "status": "success",
  "items_count": 8,
  "avg_confidence": 0.82,
  "cache_hit": true,
  "batch": false,
  "user_id": "019ce...",
  "chapter_id": "019ce..."
}
```

### 7.2 Port Go

```go
// Dans domain/event/ ou un package dédié
type LLMCallLogger interface {
    LogCall(ctx context.Context, entry LLMCallEntry) error
}

type LLMCallEntry struct {
    Timestamp     time.Time
    CallType      string  // "structuration", "fidelity", "question_gen", "coherence"
    Model         string
    PromptVersion string
    PromptHash    string
    TokensInput   int
    TokensOutput  int
    DurationMs    int64
    CostUSD       float64
    Status        string  // "success", "error", "timeout"
    ItemsCount    int     // pour structuration
    AvgConfidence float64 // pour structuration
    CacheHit      bool
    Batch         bool
    UserID        string
    ChapterID     string
    Error         string  // si status != "success"
}
```

### 7.3 Stockage

**Option recommandée pour le MVP** : table PostgreSQL `llm_call_logs` avec les colonnes ci-dessus. Index sur `(call_type, timestamp)` et `(model, timestamp)` pour les requêtes de monitoring.

**Post-MVP** : migration vers un système de métriques dédié (Prometheus + Grafana) quand le volume justifie la séparation.

---

## 8. Strategie OCR/IDP — Pipeline d'extraction des cahiers

> Ajouté le 2 avril 2026 suite à l'évaluation des approches ABBYY, Tesseract, PaddleOCR et VLM.

### 8.1 Contenu des cahiers de collégiens

Les pages contiennent un mix hétérogène de :
- Texte imprimé (polycopiés collés)
- Écriture manuscrite (notes de l'élève, corrections)
- Tableaux (résultats d'expériences, données)
- Schémas et dessins (SVT, physique, techno)
- Formules mathématiques
- Photos collées

Le pipeline doit extraire et structurer **tout ça en contexte** pour générer des items de révision pertinents.

### 8.2 Approches évaluées

| Approche | Description | Coût/page | Qualité manuscrit | Compréhension images |
|----------|------------|-----------|-------------------|---------------------|
| **ABBYY Vantage** | Pipeline IDP entreprise (K8s, multi-modèles spécialisés) | $0.01-0.05 | Bon (ICR dédié) | Non (détection seule) |
| **Tesseract 5.5** | OCR open-source LSTM | ~$0 | Quasi nul | Non |
| **PaddleOCR-VL 1.5** | VLM OCR open-source 0.9B | ~$0 | Limité | Non (catégorise mais ne décrit pas) |
| **Layout + routing** | PP-DocLayoutV2 → PaddleOCR (texte) + VLM (manuscrit/images) | ~$0.0008 | Bon (via VLM) | Oui (via VLM) |
| **VLM direct (page entière)** | Gemini Flash sur l'image complète | ~$0.0013 | Très bon (zero-shot) | Oui |

### 8.3 Décision : VLM direct, pas de layout detection

**Choix retenu : envoyer la page entière à Gemini 2.5 Flash en un seul appel.**

**Raisons :**

1. **Contexte sémantique préservé.** Un VLM qui voit la page entière comprend le lien entre un schéma et le texte à côté. Le cropping par zones perd ce contexte — or c'est exactement ce qui fait la qualité pédagogique des items générés. Un schéma de circuit électrique avec des annotations manuscrites à côté d'un tableau de mesures doit être compris comme un ensemble.

2. **Gain économique négligeable.** L'écart entre le VLM direct ($0.0013/page) et le pipeline hybride ($0.0008/page) est ~$0.0005/page, soit ~$5/mois pour 10K pages. La complexité ajoutée (3 étapes au lieu d'1, routing, maintenance de 2 modèles) ne justifie pas cette économie.

3. **Robustesse sur les cahiers réels.** Les cahiers d'ados sont chaotiques — collages, dessins dans les marges, flèches, ratures, stickers. Un layout detector va régulièrement se tromper sur ces documents, causant des erreurs en cascade. Un VLM généraliste encaisse mieux le chaos.

4. **Manuscrit = point fort des VLMs.** Les benchmarks HTR (mars 2025) montrent que Claude et Gemini surpassent les OCR spécialisés (Transkribus, TrOCR) en zero-shot sur le manuscrit moderne. PaddleOCR-VL est bon sur le texte imprimé mais limité sur le manuscrit français.

5. **PaddleOCR-VL ne décrit pas les images.** Il catégorise les zones visuelles ("figure") mais ne les décrit pas sémantiquement. Pour un schéma de cellule en SVT, il retourne un placeholder, pas "schéma de la cellule animale avec noyau et mitochondries". Le VLM fait les deux.

6. **Simplicité pipeline.** Un seul appel API au lieu de 3 étapes (layout → OCR → VLM). Moins de code, moins de points de failure, moins d'infra à maintenir.

### 8.4 Pourquoi Gemini Flash et pas Claude pour l'OCR

| Critère | Gemini 2.5 Flash | Claude Haiku 4.5 | Claude Sonnet 4.6 |
|---------|-----------------|-----------------|-------------------|
| Coût input/Mtok | $0.30 | $1.00 | $3.00 |
| Tokens par image | ~258 | ~1,334 | ~1,334 |
| **Coût/page vision** | **$0.0013** | $0.0038 | $0.0115 |
| 10K pages/mois | **$13** | $38 | $115 |

Gemini tokenise les images ~5x plus efficacement que Claude. Pour l'OCR (extraction de texte/structure), Gemini Flash est 3x moins cher que Haiku et 9x moins cher que Sonnet, pour une qualité comparable sur cette tâche.

**Principe : utiliser chaque modèle là où il excelle.**
- **Gemini Flash** → extraction OCR + scoring réponses (vision, bas coût)
- **Claude Sonnet** → génération pédagogique (qualité rédaction française)

### 8.5 Quand reconsidérer

| Condition | Action |
|-----------|--------|
| > 100K pages/mois | Évaluer PaddleOCR-VL en pré-traitement pour le texte imprimé |
| Self-hosting requis | Qwen2.5-VL-7B + PaddleOCR-VL en local |
| Qualité insuffisante Gemini | Benchmark Haiku 4.5 ou Gemini 2.5 Pro sur les cas en échec |
| Baisse de prix LLM | Réévaluer — si l'API coûte 10x moins, le layout detection perd tout intérêt |

### 8.6 Comparatif solutions IDP évaluées

| Solution | Forces | Faiblesses pour Révise Mieux |
|----------|--------|------------------------------|
| **ABBYY Vantage** | Précision industrielle, 200+ langues, ICR manuscrit dédié | Over-engineered pour un MVP, propriétaire, ~350 MB/core, pas de description sémantique des images |
| **Tesseract 5.5** | Léger (~100 MB/page), gratuit, mature | Quasi inutilisable sur le manuscrit, pas d'analyse de layout, pas de compréhension structurelle |
| **PaddleOCR-VL** | 0.9B params, gratuit, SOTA sur benchmarks OCR texte | Ne décrit pas les images, manuscrit français limité, pas de compréhension sémantique |
| **Gemini 2.5 Flash** | Vision + compréhension en 1 pass, manuscrit zero-shot, $0.0013/page | Dépendance API externe, latence ~3-5s/page |

### 8.7 Estimations de coût consolidées

| Volume | Gemini Flash (VLM direct) | Pipeline hybride | Tout Claude Sonnet |
|--------|--------------------------|------------------|--------------------|
| 1K pages/mois | $1.3 | $0.8 | $11.5 |
| 10K pages/mois | $13 | $8 | $115 |
| 100K pages/mois | $130 | $80 | $1,150 |

### 8.8 Optimisations coût complémentaires

Les vrais leviers de réduction de coût ne sont pas dans le layout detection mais dans :

1. **Batch API** (-50%) : le pipeline J0 est asynchrone, parfait pour le batch
2. **Prompt caching** (-90% sur les cache reads) : le system prompt d'extraction est identique entre appels
3. **Séparation extraction/génération** : l'extraction (Gemini Flash, image) produit du Markdown ; la génération pédagogique (Claude Sonnet, texte seul) consomme ce Markdown sans revoir l'image

### 8.9 Benchmark à réaliser

Avant de finaliser le pipeline J0, benchmarker sur 20+ photos de vrais cahiers :

| Modèle | Type | Critères |
|--------|------|----------|
| Gemini 2.5 Flash | API vision | Précision texte imprimé, manuscrit, description schémas |
| Claude Haiku 4.5 | API vision | Idem (comparaison qualité) |
| PaddleOCR-VL 1.5 | Self-hosted | Texte imprimé uniquement (baseline gratuite) |

Le framework de benchmark existe dans `backend/cmd/benchmark/`. Il faut ajouter un type `ocr` avec des cas de test sur un corpus de cahiers réels.

---

## 9. Résumé des décisions

> **Mis à jour mars 2026** après benchmark réel sur 21 modèles. Voir [`backend/testdata/benchmark/README.md`](../backend/testdata/benchmark/README.md).

| Décision | Choix initial | Choix post-benchmark | Raison du changement |
|----------|--------------|---------------------|---------------------|
| Modèle structuration | ~~Sonnet 4.6~~ | **Qwen3.5-397B** ou **mistral-small** | Sonnet hallucine (8%), 30-50x plus cher. Qwen3.5-397B quality 0.88, mistral-small quality 0.87 à $0.00006/item |
| Modèle OCR / extraction | ~~Sonnet 4.6~~ | **Gemini 2.5 Flash** (VLM direct, page entière) | 5x plus efficace en tokens image, $0.0013/page, manuscrit zero-shot. Pas de layout detection — cf. section 8. |
| Modèle fidelity check | Haiku 4.5 | À benchmarker | Pas encore testé sur ce cas d'usage |
| Modèle questions | Haiku 4.5 | À benchmarker | Pas encore testé sur ce cas d'usage |
| Modèle cohérence | ~~Sonnet 4.6~~ | À benchmarker | Pas encore testé sur ce cas d'usage |
| Feedback | Templates (pas de LLM) | Inchangé | Latence ≤ 200ms, cohérence, coût nul |
| Exécution locale | Hors scope | **Qwen3-VL-32B + Gemma 3 27B via Ollama** | Viable sur MacBook Pro M5 Max 128 Go (~2 min/chapitre, $0) |
| Prompts | Constantes Go | **Fichiers .txt + `//go:embed`** | Source unique partagée entre prod et benchmark |
| Pipeline non temps-réel | Batch API | Batch API | Inchangé |
| Prompts répétitifs | Prompt caching | Prompt caching | Inchangé |
| Tests CI | Golden files + contrat | Golden files + contrat | Inchangé |
| Tests qualité | Régression nightly | Régression nightly | Inchangé |
| Monitoring | Logs structurés + alertes | Logs structurés + alertes | Inchangé |
