# Extraction structurée de documents — Benchmark des architectures LLM

## Objectif

Évaluer les architectures LLM pour extraire des données structurées (JSON) depuis des photos de documents manuscrits. L'étude compare 3 approches sur 12 modèles, benchmarkées sur 3 cas réels (cahiers manuscrits français, 2 à 8 pages par document).

---

## Les 3 approches

### E2E (End-to-End)

```
┌──────────┐                      ┌──────────────┐
│  Images  │────▶  VLM cloud  ───▶│ JSON structuré│
└──────────┘      (1 appel)       └──────────────┘
```

Un seul modèle vision fait OCR + structuration en une passe.

**Avantages :**
- Architecture la plus simple (1 seul appel API)
- Latence minimale (pas d'étape intermédiaire)
- Pas d'infra OCR à maintenir

**Inconvénients :**
- Qualité dépendante d'un seul modèle
- Pas de cache intermédiaire (re-traitement complet à chaque changement de prompt)
- Coût plus élevé (images = beaucoup de tokens)

---

### Pipeline classique (OCR → IDP)

```
┌──────────┐     ┌─────────────────┐     ┌──────────────────┐     ┌──────────────┐
│  Images  │────▶│ OCR self-hosted  │────▶│ LLM cloud (texte)│────▶│ JSON structuré│
└──────────┘     │ (RolmOCR 7B)    │     │ (reçoit blocs)   │     └──────────────┘
                 └─────────────────┘     └──────────────────┘
```

Un OCR local extrait le texte, puis un LLM cloud structure les blocs texte (sans voir les images).

**Avantages :**
- Le structureur ne reçoit que du texte = peu de tokens, pas de rate limit images
- Cache OCR : changer de modèle de structuration ne nécessite pas de refaire l'OCR
- Compatible avec les modèles text-only (moins chers, plus de choix)
- Pas de contrainte de taille d'image

**Inconvénients :**
- Perte d'information : le structureur ne voit pas les schémas, graphiques, tableaux
- 2 étapes = 2 points de failure
- L'OCR peut introduire des erreurs non corrigibles en aval

---

### Hybrid (OCR → images + texte)

```
┌──────────┐     ┌─────────────────┐     ┌──────────────────────────┐     ┌──────────────┐
│  Images  │──┬─▶│ OCR self-hosted  │──┐  │ VLM cloud                │────▶│ JSON structuré│
└──────────┘  │  │ (RolmOCR 7B)    │  ├─▶│ reçoit images + blocs OCR│     └──────────────┘
              │  └─────────────────┘  │  └──────────────────────────┘
              └───────────────────────┘
```

Le structureur reçoit à la fois les images ET les blocs OCR. Il peut croiser les deux sources : utiliser le texte OCR pour structurer et les images pour vérifier visuellement.

**Avantages :**
- Meilleure qualité (le modèle vérifie l'OCR contre les images)
- Meilleure complétude (repère ce que l'OCR a raté via les images)
- Combinaison des forces des deux approches

**Inconvénients :**
- Plus de tokens (images + texte) = plus cher que le pipeline classique
- Contrainte de rate limit sur les APIs avec images (Mistral : 1M tokens/min)
- Nécessite un modèle vision (exclut les modèles text-only)
- Nécessite un redimensionnement adaptatif des images pour les gros documents

---

## Modèle OCR self-hosted

| Modèle | Taille | Licence | Spécialité | Déploiement |
|--------|--------|---------|------------|-------------|
| **RolmOCR** | 7B | Apache 2.0 | Fine-tune Qwen2.5-VL, 92% accuracy manuscrit | llama-server, 8GB RAM, CPU ARM |

RolmOCR traite les images **une par une** (mode per-page) et produit des blocs texte structurés en JSON. Il est utilisé dans les approches Pipeline classique et Hybrid.

---

## Résultats

### Benchmark complet — 3 cas réels, 3/3 cas réussis, tous modèles

| # | Modèle | E2E | Pipeline classique | Hybrid | Meilleur | Coût/doc |
|---|--------|-----|-------------------|--------|----------|----------|
| 1 | **Gemini 2.5 Flash** | 0.811 | 0.811 | **0.844** | Hybrid | $0.002 |
| 2 | **Qwen3-VL-32B** | 0.774 | **0.805** | 0.795 | Pipeline | $0.001 |
| 3 | **qwen3.5-397b** | — | **0.789** | — | Pipeline | $0.002 |
| 4 | **Llama 4 Maverick** | 0.724 | **0.788** | 0.695 | Pipeline | $0.001 |
| 5 | **GPT-4.1 Mini** | 0.666 | 0.752 | **0.772** | Hybrid | $0.009 |
| 6 | **Mistral Small 3.1** | 0.758 | **0.770** | 0.720 | Pipeline | $0.001 |
| 7 | **Pixtral 12B** | 0.758 | 0.741 | **0.766** | Hybrid | $0.002 |
| 8 | **Mistral Small 4** | 0.722 | **0.757** | 0.753 | Pipeline | $0.001 |
| 9 | **Llama 4 Maverick** | **0.724** | 0.788 | 0.695 | Pipeline | $0.001 |

Note : "—" = modèle non compatible avec cette approche (text-only pour E2E, ou non testé).

### Analyse par métriques détaillées (approche optimale par modèle)

| # | Combo | Score | Complétude | Fidélité | Halluc. | Coût/doc |
|---|-------|-------|------------|----------|---------|----------|
| 1 | RolmOCR + Gemini Flash (hybrid) | **0.844** | **80%** | **98%** | 2% | $0.002 |
| 2 | Gemini Flash seul (E2E) | **0.811** | 86% | 93% | 7% | $0.002 |
| 3 | RolmOCR + Qwen3-VL-32B (pipeline) | **0.805** | 64% | **99%** | **1%** | **$0.001** |
| 4 | RolmOCR + qwen3.5-397b (pipeline) | **0.789** | 55% | 93% | 7% | $0.002 |
| 5 | RolmOCR + Llama 4 Maverick (pipeline) | **0.788** | 47% | **100%** | **0%** | **$0.001** |
| 6 | RolmOCR + GPT-4.1 Mini (hybrid) | **0.772** | 77% | **100%** | **0%** | $0.009 |
| 7 | RolmOCR + Mistral Small 3.1 (pipeline) | **0.770** | 57% | 93% | 7% | **$0.001** |
| 8 | RolmOCR + Pixtral 12B (hybrid) | **0.766** | 62% | 89% | 11% | $0.002 |

---

## Analyse

### L'approche hybrid profite-t-elle à tous les modèles ?

Non. Seuls certains modèles bénéficient de recevoir les images en plus de l'OCR :

| Modèle | Gain hybrid vs pipeline | Interprétation |
|--------|------------------------|----------------|
| Gemini 2.5 Flash | **+4%** | Profite de la vérification visuelle |
| GPT-4.1 Mini | **+3%** | Idem |
| Pixtral 12B | +3% | Modèle vision, synergique |
| Qwen3-VL-32B | -1% | Déjà excellent en texte, images = bruit |
| Mistral Small 4 | -0.5% | Neutre |
| Llama 4 Maverick | **-12%** | Les images le perturbent fortement |
| Mistral Small 3.1 | **-6%** | Idem |

**Conclusion :** L'approche hybrid n'est bénéfique que pour Gemini et GPT-4.1 Mini. Pour tous les autres, le pipeline classique (texte seul) fait aussi bien ou mieux.

### Quel est le vrai apport de RolmOCR ?

Comparaison E2E (sans RolmOCR) vs Pipeline classique (avec RolmOCR, texte seul) :

| Modèle | E2E | Pipeline (+RolmOCR) | Gain |
|--------|-----|---------------------|------|
| Qwen3-VL-32B | 0.774 | **0.805** | **+4%** |
| Llama 4 Maverick | 0.724 | **0.788** | **+9%** |
| Mistral Small 3.1 | 0.758 | **0.770** | +2% |
| Mistral Small 4 | 0.722 | **0.757** | **+5%** |
| GPT-4.1 Mini | 0.666 | **0.752** | **+13%** |
| Gemini 2.5 Flash | 0.811 | 0.811 | 0% |
| Pixtral 12B | 0.758 | 0.741 | -2% |

RolmOCR améliore 5 modèles sur 7. Le gain est spectaculaire sur GPT-4.1 Mini (+13%) et Llama 4 Maverick (+9%) — des modèles qui ne sont pas spécialisés vision mais excellents en structuration texte. RolmOCR leur pré-mâche le travail visuel.

---

## Coût d'exploitation

### Infrastructure OCR (fixe)

| Option | Instance | Coût/mois | Capacité |
|--------|----------|-----------|----------|
| Minimal | AWS t4g.medium (ARM, 4GB) | $25 | ~100 docs/heure |
| Standard | AWS t4g.large (ARM, 8GB) | $50 | ~200 docs/heure |

### Coût total par approche (10 000 docs/mois)

| Approche | Infra OCR | API cloud | Total/mois | Coût/doc |
|----------|-----------|-----------|------------|----------|
| E2E (Gemini Flash) | $0 | $20 | **$20** | $0.002 |
| Pipeline (RolmOCR + Qwen3-VL-32B) | $25 | $10 | **$35** | $0.001 |
| Pipeline (RolmOCR + Llama 4 Maverick) | $25 | $10 | **$35** | $0.001 |
| Hybrid (RolmOCR + Gemini Flash) | $25 | $20 | **$45** | $0.002 |
| Hybrid (RolmOCR + GPT-4.1 Mini) | $25 | $90 | **$115** | $0.009 |

### Seuil de rentabilité RolmOCR self-hosted

Le coût fixe de l'infra OCR ($25/mois) est amorti à partir de **~4 000 docs/mois** par rapport à un OCR cloud.

En dessous, l'E2E sans infra est plus économique. Au-dessus, le pipeline avec RolmOCR est systématiquement moins cher.

---

## Stratégie de failover

Le pipeline est agnostique du provider de structuration. Le switch se fait par variable d'environnement :

```
Primary:    Gemini 2.5 Flash    ($0.15/$0.60)  — meilleur score hybrid
Fallback 1: Qwen3-VL-32B       ($0.10/$0.42)  — meilleur pipeline, 0% hallucination
Fallback 2: Llama 4 Maverick   ($0.20/$0.20)  — open-source, 0% hallucination
Fallback 3: GPT-4.1 Mini       ($0.40/$1.60)  — Azure, 0% hallucination
```

---

## Recommandation

| Critère prioritaire | Approche | Modèle | Score | Coût/doc |
|---------------------|----------|--------|-------|----------|
| **Meilleure qualité** | Hybrid | Gemini 2.5 Flash | 0.844 | $0.002 |
| **Meilleur rapport qualité/prix** | Pipeline classique | Qwen3-VL-32B | 0.805 | $0.001 |
| **Zéro hallucination** | Pipeline classique | Llama 4 Maverick | 0.788 | $0.001 |
| **Simplicité (pas d'infra)** | E2E | Gemini 2.5 Flash | 0.811 | $0.002 |
| **Écosystème Microsoft/Azure** | Hybrid | GPT-4.1 Mini | 0.772 | $0.009 |
| **Budget minimal (<4K docs)** | E2E | Gemini 2.5 Flash | 0.811 | $0.002 |
| **Open-source / multi-cloud** | Pipeline classique | Llama 4 Maverick | 0.788 | $0.001 |

---

## Contraintes techniques identifiées

| Contrainte | Impact | Mitigation |
|------------|--------|------------|
| Mistral rate limit 1M tokens/min | Impossible d'envoyer >5 images en hybrid | Redimensionnement adaptatif (2048→1280px) |
| RolmOCR crash LM Studio multi-images | OCR instable en batch | Mode per-page (1 image par requête) via llama-server |
| Modèles reasoning (tokens thinking) | Output tronqué, latence x10 | Désactivation thinking (`reasoning_effort: none`) |
| JSON output tronqué (gros documents) | Structuration incomplète | `max_tokens: 16384` au lieu de 8192 |

---

## Pistes d'ingénierie

L'analyse des résultats par cas révèle 4 bottlenecks principaux et 10 hypothèses d'amélioration.

### Bottlenecks identifiés

| # | Bottleneck | Impact | Données |
|---|-----------|--------|---------|
| 1 | **Complétude plafonnée à 80%** | 20% des items sont systématiquement ratés | Meilleur modèle (Gemini hybrid) : 80% complétude. Cas 10 (SVT) : la plupart des modèles ne trouvent que 8/13 items |
| 2 | **Sur-segmentation OCR** | Bruit + surcoût tokens | RolmOCR per-page : 47 blocs pour 29 attendus (x1.6), 98 pour 49 (x2). Ajoute ~50% de tokens input inutiles |
| 3 | **Coût dominé par les tokens input** | Le prompt OCR = 80% du coût | Pipeline classique : 16-43K tokens input vs 1-5K output. Réduire les blocs OCR réduirait directement la facture |
| 4 | **Latence OCR locale** | 30-50% du temps total | RolmOCR per-page : 35-110s. La structuration cloud rajoute 30-170s. L'OCR est le premier goulot |

### Hypothèses d'amélioration — Performance

**H1. Fusion des blocs OCR avant structuration**

RolmOCR en mode per-page traite chaque image indépendamment et produit des blocs redondants ou trop fragmentés (98 blocs pour 49 attendus). Une étape de post-processing légère (merge des blocs adjacents, dédoublonnage, filtrage <10 chars) pourrait réduire le bruit de 30-50%.

- Impact attendu : complétude +5-10%, tokens input -30%, hallucination -3%
- Coût : zéro (traitement local)
- Complexité : faible (heuristiques texte)

**H2. Few-shot prompting**

Le prompt de structuration actuel est zero-shot (instructions + format attendu, pas d'exemple concret). Ajouter 1-2 exemples d'input OCR → output JSON attendu dans le prompt pourrait significativement améliorer le suivi du format et la complétude.

- Impact attendu : complétude +10-15%, fidélité +5%
- Coût : +500-1000 tokens input par requête (~+$0.0001/doc)
- Complexité : faible (prompt engineering)

**H3. Chunked processing pour les gros documents**

Le cas 13 (8 pages, 98 blocs, 43K tokens) est systématiquement le plus faible. Découper en 2-3 chunks de pages (pages 1-3, 4-6, 7-8), structurer chaque chunk indépendamment, puis merger les items et dédupliquer.

- Impact attendu : complétude +10% sur les gros documents, fiabilité (pas de truncation JSON)
- Coût : x2-3 appels API mais sur des prompts plus petits (similaire au total)
- Complexité : moyenne (logique de merge + dédoublonnage d'items)

**H4. Double pass : extraction + validation**

Pass 1 : modèle cheap extrait tous les items (optimisé complétude, tolère l'hallucination).
Pass 2 : modèle cheap relit les items et flag les incohérents (optimisé précision).

- Impact attendu : hallucination -50% tout en gardant la complétude
- Coût : +$0.0003/doc (GPT-4.1 Nano pour la validation)
- Complexité : moyenne (second prompt, logique de filtrage)

**H5. Prompt spécialisé par matière**

Le prompt actuel est générique. Un prompt spécialisé SVT (avec vocabulaire scientifique, types de schémas attendus) ou Histoire (chronologie, personnages, concepts) pourrait améliorer la classification et la détection des items spécifiques.

- Impact attendu : classification +10%, complétude +5%
- Coût : zéro
- Complexité : faible (variations de prompt, A/B testing)

### Hypothèses d'amélioration — Coût

**H6. Routing par complexité**

Documents simples (1-3 pages, <20 blocs) → modèle cheap (Mistral Small 3.1, $0.001/doc).
Documents complexes (>5 pages, >50 blocs) → modèle premium (Gemini Flash, $0.002/doc).
Décision basée sur le nombre de blocs OCR extraits (disponible avant l'appel de structuration).

- Impact attendu : -30% coût moyen sans perte de qualité
- Coût : zéro
- Complexité : faible (if/else sur le nombre de blocs)

**H7. Cache OCR par hash d'image**

Les mêmes photos de cahier sont souvent re-traitées (re-structuration après changement de prompt, retry, etc.). Cacher les blocs OCR par hash SHA256 de l'image évite de relancer RolmOCR.

- Impact attendu : -100% latence OCR sur les re-traitements, -30% latence moyenne
- Coût : stockage négligeable (quelques KB par image)
- Complexité : faible (key-value store, déjà du Redis dans la stack)

**H8. Remplacement RolmOCR par OCR cloud cheap pour le bench de production**

RolmOCR (7B, self-hosted) fait 65% de détection. Gemini 2.5 Flash en mode OCR fait 78% de détection pour ~$0.0005/page. Le surcoût OCR cloud est marginal et pourrait améliorer toute la chaîne aval.

- Impact attendu : détection OCR +20%, complétude pipeline +5-10%
- Coût : +$0.001/doc
- Complexité : faible (déjà implémenté dans le bench)

### Hypothèses d'amélioration — Vitesse

**H9. OCR parallèle multi-slots**

llama-server supporte 4 slots parallèles. Actuellement on traite les pages séquentiellement (1 par 1). En les envoyant en parallèle (4 pages simultanées), la latence OCR d'un document 8 pages passerait de ~110s à ~30s.

- Impact attendu : latence OCR /3-4x
- Coût : zéro (même infra)
- Complexité : faible (requêtes HTTP parallèles)

**H10. Streaming structuration**

Commencer la structuration dès que les premières pages sont OCR-isées, sans attendre la fin complète de l'OCR. Le structureur reçoit les blocs en streaming et peut commencer à extraire les items des premières pages.

- Impact attendu : latence totale -30% (OCR et structuration en overlap)
- Coût : zéro
- Complexité : élevée (architecture streaming, restructuration du pipeline)

### Matrice impact / effort

```
        Impact élevé
            │
    H2      │  H3      H4
  Few-shot  │ Chunks   Double pass
            │
    H1      │  H8         H10
  Merge OCR │ OCR cloud   Streaming
            │
────────────┼──────────────────── Effort
            │
    H7      │  H6         H5
  Cache     │ Routing     Prompt/matière
            │
    H9      │
  Parallel  │
            │
        Impact faible
```

### Recommandation : quick wins à implémenter en premier

| Priorité | Hypothèse | Impact | Effort | Gain attendu |
|----------|-----------|--------|--------|-------------|
| 1 | **H1 — Merge blocs OCR** | Qualité + coût | 1 jour | Complétude +5%, tokens -30% |
| 2 | **H2 — Few-shot prompting** | Qualité | 2 heures | Complétude +10%, fidélité +5% |
| 3 | **H9 — OCR parallèle** | Vitesse | 2 heures | Latence OCR /4 |
| 4 | **H7 — Cache OCR** | Vitesse + coût | 4 heures | -100% latence sur re-traitement |
| 5 | **H6 — Routing complexité** | Coût | 2 heures | -30% coût moyen |

---

## Méthodologie

- **3 cas réels** : SVT (3 pages), Histoire féodale (3 pages), Histoire chrétienne (8 pages)
- **Métriques** : complétude, fidélité, hallucination, coût, latence
- **Score composite** : qualité × 0.85 + efficacité coût × 0.15
- **Critère de validité** : seuls les résultats avec 3/3 cas réussis sont retenus
- **OCR self-hosted** : RolmOCR 7B via llama-server, mode per-page, context 32K
- **Images** : redimensionnement adaptatif (max 2048px, réduit automatiquement si budget tokens dépassé)
- **Date** : avril 2026
