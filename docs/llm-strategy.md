# Stratégie LLM — Révise Mieux

> Document technique de référence pour les choix de modèles, l'optimisation des coûts, la stratégie de test et le monitoring des appels LLM.
> Dernière mise à jour : 2026-03-12.

---

## 1. Cas d'usage LLM

| # | Cas d'usage | AC | Interface Go | Modèle | Latence cible | Temps réel ? |
|---|------------|-----|-------------|--------|---------------|-------------|
| A | Structuration OCR → Items + Notions | Z2-AC01, Z7-AC15 | `chapter.LLMService.StructureBlocks()` | **Sonnet 4.6** | ≤ 60s/page | Non (pipeline J0) |
| B | Fidelity check (fidélité sémantique) | Z3-AC10 | À créer | **Haiku 4.5** | ≤ 5s/item | Non (pipeline J0) |
| C | Détection de cohérence (doublons, contradictions) | Z3-AC11 | À créer | **Sonnet 4.6** | ≤ 10s/chapitre | Non (pipeline J0) |
| D | Génération de questions (lazy generation) | Z4-AC01 | À créer | **Haiku 4.5** | ≤ 500ms | Oui (session) |
| E | Feedback enrichi | Z4-AC09 | `feedback.go` | **Templates** (pas de LLM) | ≤ 200ms | Oui |
| F | Multimodal visuel (schémas, graphiques) | Z4-AC17 | Post-MVP | **Sonnet 4.6** | ≤ 2s | Non |

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
├── prompts.go          # Prompt templates versionnés (constantes)
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

## 8. Résumé des décisions

| Décision | Choix | Raison |
|----------|-------|--------|
| Modèle structuration | Sonnet 4.6 | Fidélité source, classification nuancée, JSON structuré |
| Modèle fidelity check | Haiku 4.5 | Classification binaire, volume élevé, coût |
| Modèle questions | Haiku 4.5 | Latence critique ≤ 500ms, tâche bien structurée |
| Modèle cohérence | Sonnet 4.6 | Raisonnement sémantique multi-items |
| Feedback | Templates (pas de LLM) | Latence ≤ 200ms, cohérence, coût nul |
| Multimodal visuel | Sonnet 4.6 (post-MVP) | Vision intégrée, extraction de labels |
| Pipeline non temps-réel | Batch API | -50% sur les coûts |
| Prompts répétitifs | Prompt caching | -90% sur les system prompts |
| Tests CI | Golden files + contrat | Rapides, gratuits, chaque PR |
| Tests qualité | Régression nightly | Appels réels, détection drift |
| Monitoring | Logs structurés + alertes | Détection proactive des dégradations |
