# Révise Mieux — PRD

> **Product Requirements Document — MVP Collège**
>
> | | |
> |---|---|
> | **Statut** | Draft — Review interne |

---

## Table des matières

1. [Résumé produit](#1-résumé-produit)
2. [Objectifs, promesse, principes](#2-objectifs-promesse-principes)
3. [Périmètre & séquencement](#3-périmètre--séquencement)
4. [Personas](#4-personas)
5. [Parcours utilisateurs](#5-parcours-utilisateurs)
6. [Tags — Spécification complète](#6-tags--spécification-complète)
7. [Gabarits — Bibliothèque générique + Packs](#7-gabarits--bibliothèque-générique--packs)
8. [Bibliothèque de gabarits (27 templates)](#8-bibliothèque-de-gabarits-27-templates-mvp)
9. [Packs matière](#9-packs-matière)
10. [Pipeline J0 — Flux de données détaillé](#10-pipeline-j0--flux-de-données-détaillé)
11. [SLA & contraintes de performance](#11-sla--contraintes-de-performance)
12. [Stratégie cache & génération lazy](#12-stratégie-cache--génération-lazy)
13. [Retry / Fallback OCR](#13-retry--fallback-ocr)
14. [Architecture cible](#14-architecture-cible-mvp)
15. [Fonctionnalités et exigences (Epics)](#15-fonctionnalités-et-exigences-epics)
16. [Modèle de données](#16-modèle-de-données-mvp)
17. [Algorithmes MVP](#17-algorithmes-mvp)
18. [Exigences non fonctionnelles](#18-exigences-non-fonctionnelles)
19. [KPIs](#19-kpis-mvp)
20. [Questions §16 tranchées](#20-questions-16-tranchées-v14)
21. [Risques & mitigations](#21-risques--mitigations)
22. [Machines à états formelles](#22-machines-à-états-formelles)

---

## 1. Résumé produit

Révise Mieux est un SaaS qui transforme des photos de cahier (manuscrit, schémas, documents) en un assistant de révision personnalisé pour collégiens. À partir d'un upload de cours, le produit génère une carte de leçon structurée, des entraînements adaptatifs (rappel actif, analyse documentaire, méthodes de calcul), des contrôles blancs, et met les parents dans la boucle via un reporting rassurant basé sur des preuves de maîtrise réelles.

Le produit est livré en **trois paliers** :

| Palier | Statut | Périmètre | ACs |
|---|---|---|---|
| **Lot -1** (web-v0) | en ligne | Site web one-page : photos → fiche de révision HTML autonome. Pas de comptes, pas de DB, pas de suivi. Cf. § 3.4. | 15 (Z0) |
| **Lot 0** (pré-MVP père-fils) | en cours | Boucle pédagogique locale, 4 packs pilotes, mastery + spaced rep. Cf. § 3.2. | 53 sur 171 |
| **MVP** (produit complet) | spec | Multi-utilisateur, ENT, parents actifs, admin. Cf. § 3.1. | 171 |

### Chapitres pilotes MVP

| Matière | Chapitres pilotes | Pack |
|---|---|---|
| Histoire-Géographie | Les inégalités dans le monde | HG-INEG |
| Histoire-Géographie | La société féodale | HG-FEOD |
| SVT | La photosynthèse | SVT-PHOTO |
| Physique-Chimie | Masse, volume et densité | PC-MVD |

---

## 2. Objectifs, promesse, principes

### 2.1 Objectifs utilisateur

| Persona | Besoin principal | Succès mesuré par |
|---|---|---|
| Élève (11–15 ans) | Réviser efficacement 10–20 min/jour, savoir quoi faire aujourd'hui, réduire les angles morts | Transitions FRAGILE→SOLID, score contrôle blanc |
| Parent | Signal fiable sur la maîtrise sans micro-manager, alerté seulement si risque réel | Taux ouverture digest, taux opt-out alertes < 5% |
| Admin produit | Gérer packs, lexiques, gabarits, superviser qualité de génération | Taux réussite par template_id, temps moyen/item |

### 2.2 Objectifs business (MVP)

- Valider l'adoption sur des cycles réels *contrôle dans 1–3 semaines*.
- Prouver la valeur : plan d'action quotidien + maîtrise mesurable + contrôles blancs.
- Mesurer la rétention sur au moins 2 chapitres consécutifs pour un même élève.

### 2.3 Promesse produit

> **Ne pas promettre « meilleure note garantie ».** Promesse : maximiser les chances de réussite par entraînement ciblé, maîtrise mesurée et préparation au format du contrôle.

### 2.4 Principes pédagogiques (non négociables)

- **Rappel actif > relecture.** Toute session génère des questions, jamais de simple re-lecture de carte.
- **Répétition espacée.** État SOLID uniquement après 2 réussites espacées d'au moins 24h.
- **Révision proactive dès J0.** Chaque cours capturé déclenche une première révision le soir même, sans attendre qu'un contrôle soit annoncé. Un élève qui révise un peu chaque soir est mieux armé face aux interros surprises.
- **Analyse documentaire standardisée :** décrire → prélever → expliquer → conclure.
- **Double codage (Paivio) & génération de documents originaux.** Le système ne se limite pas à exploiter les visuels extraits du cahier : il **génère des documents inédits** (tableaux de données, graphiques, schémas légendés, cartes simplifiées, diagrammes de classification) cohérents avec le cours mais contenant des valeurs et contextes différents. Ces documents servent de supports d'exercice pour renforcer l'encodage verbal + visuel. Formats : Markdown tables, Mermaid (xychart-beta, flowchart, mindmap, pie), SVG inline, ASCII art.
- **Human-in-the-loop :** validation rapide des zones incertaines critiques, plutôt que faux sentiment de certitude.
- **First value rapide :** l'élève doit pouvoir faire un premier exercice en moins de 5 min après upload. La qualité s'affine ensuite.
- **Anticipation permanente.** Le service connaît l'emploi du temps de l'élève et prépare une révision ciblée la veille de chaque cours, pour couvrir le risque d'interro surprise.

---

## 3. Périmètre & séquencement

### 3.0 Vue d'ensemble du séquencement

Le produit progresse en **trois paliers** complémentaires, chaque palier livrant une valeur autonome :

```
   Lot -1               Lot 0                    MVP
  (web-v0)         (pré-MVP local)          (produit complet)
     │                   │                        │
     ▼                   ▼                        ▼
 site one-page     boucle pédagogique       multi-utilisateur
 photos → HTML  →  père-fils, 4 packs   →   ENT, parents actifs
 pas de suivi     mastery + spaced rep      admin, RGPD J+30
     │                   │                        │
   15 ACs              53 ACs                  171 ACs
   (Z0)               (Z1-Z8 ⊂)              (Z1-Z8 complet)
     │                   │                        │
   en ligne           en cours                  spec
```

| | Lot -1 | Lot 0 | MVP |
|---|---|---|---|
| **Cible utilisateur** | Louis seul, à la dernière minute | binôme père-fils | familles, écoles |
| **Persistance** | aucune (mémoire process) | local (Postgres) | cloud + S3 + Redis |
| **Comptes** | aucun | 1 famille | multi-utilisateur |
| **Suivi** | aucun | mastery + spaced rep | mastery + reporting parent |
| **Stack** | FastAPI + Jinja + LM Studio | Go + RN + Postgres | Go + RN + Postgres + S3 + Redis |
| **Déploiement** | image Docker GHCR + K3s homelab | local docker-compose | cloud (à choisir) |
| **Domaine** | `revise-lab.musso.io` | localhost | TBD |
| **Détail** | § 3.4 | § 3.2 | § 3.1 |

> **Lot -1 est volontairement disjoint** des autres paliers : il n'est pas un sous-ensemble fonctionnel mais un **produit séparé** qui valide la qualité de la fiche générée. Aucune ligne de code n'est partagée avec Lot 0/MVP en dehors du **prompt système de la skill `study-guide`** (`prompts/study-guide/system.md`), qui est la source unique pour les deux familles de produits.

### 3.1 Périmètre MVP (produit complet)

| ✅ Inclus | ❌ Exclu |
|---|---|
| Création chapitre + upload multi-photos (1–30 photos, mobile/web). | Couverture toutes matières / tous chapitres. |
| Segmentation page en blocs typés (TEXT, PHOTO, SCHEMA, MAP, GRAPH, TABLE…). | Notation IA parfaite de longues rédactions (rubriques simples uniquement). |
| OCR + structuration (plan hiérarchique) + génération d'Items pédagogiques. | Mode enseignant/classe. |
| Auto-tagging + enrichissement contrôlé par pack. | OCR offline complet sur device. |
| Détection d'incertitudes → file de validation (max 8/chapter). | Apprentissage progressif des synonymes (admin uniquement en MVP). |
| Diagnostic initial 5–10 min → carte de maîtrise initiale. | Imports depuis ENT ou manuels numériques. |
| Sessions quotidiennes adaptatives (spaced repetition, 70/20/10). | Import automatique emploi du temps depuis ENT/Pronote. |
| Questions générées via gabarits génériques + packs (lazy generation). | |
| Contrôles blancs (≥1 par chapitre) avec correction guidée. | |
| Dashboard élève + reporting parent passif (digest hebdo + veille contrôle). | |
| Alertes parent actif (opt-in) sur risques détectés. | |
| Admin : CRUD packs, lexiques, gabarits, analytics template. | |
| Versioning chapitre (re-upload de pages → nouvelle révision). | |
| **Emploi du temps élève** (saisie manuelle) → notifications saisie cours + révision proactive. | |
| **Révision proactive dès le soir** après chaque prise de cours, même sans date de contrôle. | |
| **Anticipation interros surprises** : révision ciblée la veille de chaque cours. | |
| **Exams multi-chapitres** : un contrôle peut couvrir plusieurs leçons. | |

Le MVP complet représente **171 ACs** répartis en 8 zones de risque, classifiés en MVP Core (106), MVP Hardening (55) et Post-MVP (10). Voir `MVP-scope.md` pour le détail par AC.

### 3.2 Lot 0 — Pré-MVP (version locale père-fils)

> Version locale pour un binôme père-fils, validant la boucle pédagogique fondamentale sur les 4 packs pilotes. **53 ACs retenus sur 171 (31%)**, tous extraits de MVP Core.

| Zone | Retenus | Différés | Ratio |
|---|---|---|---|
| Z1 Mastery | 15 | 13 | 54% |
| Z2 Pipeline | 10 | 8 | 56% |
| Z3 HITL | 8 | 9 | 47% |
| Z4 Lazy gen | 5 | 12 | 29% |
| Z5 Revision | 3 | 8 | 27% |
| Z6 Schedule | 5 | 41 | 11% |
| Z7 Routine | 2 | 24 | 8% |
| Z8 Onboarding | 5 | 3 | 63% |
| **Total** | **53** | **118** | **31%** |

**Priorités internes :** P1 (33 ACs) = la boucle fonctionne · P2 (20 ACs) = expérience quotidienne complète.

**Coupé (reporté au MVP) :** multi-utilisateur, notifications push, emploi du temps, orchestration de soirée, RGPD J+30, admin backoffice, mode vacances, fiches PDF in-app.

> **Note :** la génération de fiches de révision complètes est disponible via le skill `/study-guide` (Claude Code). Ce skill analyse des photos de cours et produit une fiche Markdown autonome avec : résumé, notions clés, documents d'exercice générés (tableaux, graphiques, schémas, cartes — originaux, pas des copies du cahier), 2 sessions de 15 questions (J0 découverte + J2 consolidation), corrigé détaillé (3 composantes Z4-AC09) et plan de progression mastery. Les fiches PDF in-app (génération serveur, QR code, report papier) restent reportées au MVP.

Le tag Lot 0 (P1/P2/—) est porté par chaque AC dans les fichiers `docs/ac/Z*.md`.

### 3.3 Phases d'implémentation Lot 0

#### Phase 0 — Fondations

Créer les artefacts d'ancrage **avant toute fonctionnalité** :
1. **Schéma de base de données** — traduire le §16 en schéma exécutable, avec les corrections INC-2 à INC-10 déjà appliquées.
2. **Contrat d'interface** — les routes d'échange entre l'app mobile et le serveur.
3. **Données de référence** — les 27 templates, les 4 packs avec leurs lexiques et concept_tags.
4. **Prompts LLM versionnés** — les prompts du pipeline (extraction items, vérification fidélité, regroupement notions, instanciation questions).
5. **Données de test** — chapitre démo (8 items statiques), fixtures pour les edge cases mastery.

#### Phase 1 — Le moteur (Z1 P1 + Z4 core)

1. Machine à états mastery (Z1-AC01→AC07c, AC09, AC13, AC14)
2. Calcul de `next_due_at` (spaced repetition §17.2)
3. Scoring (KEYWORDS, MCQ, SHORT_ANSWER)
4. Composition de session (algorithme 70/20/10, §17.1)
5. Reprise de session interrompue (Z4-AC06)
6. Feedback après réponse (Z4-AC12 simplifié)

> **Critère de validation :** le fils peut jouer une session sur le chapitre démo et voir sa maîtrise évoluer correctement.

#### Phase 2 — Le pipeline (Z2 + Z5 + Z7-AC15)

1. Upload de photos → stockage
2. OCR par page → texte brut + confidence
3. Génération d'items par le LLM → Items structurés avec tags
4. Regroupement en Notions (Z7-AC15)
5. Streaming de la carte de leçon (Z2-AC10)
6. Gestion des erreurs : page en échec (Z2-AC02/04/05), 0 items (Z2-AC07)
7. Identité canonique des items (Z5-AC01)

> **Critère de validation :** le fils prend en photo une page de son cahier, la carte de leçon apparaît avec les items regroupés par notion.

#### Phase 3 — L'expérience complète (P2)

1. Validation HITL (Z3-AC01→AC05, AC09, AC10)
2. Plafond mastery items non validés (Z1-AC13)
3. Scoring avancé (Z1-AC10 : RUBRIC, NUMERIC avec unité, KEYWORDS N-1)
4. Exam + resserrement (Z1-AC08, Z6-AC10)
5. Mock exam (Z6-AC09, Z6-AC38)
6. Session evening_first (Z6-AC05/AC06)
7. Vue par notion (Z7-AC16)
8. Onboarding complet (Z8-AC01→AC04, AC08)

> **Critère de validation :** le fils utilise l'app quotidiennement pendant 1 semaine sur un vrai chapitre, avec un exam posé.

### 3.4 Lot -1 — One-shot generator (web-v0)

> **Statut :** en ligne sur `https://revise-lab.musso.io/`. **Code :** [`web-v0/`](../web-v0/). **ACs :** [Z0](ac/Z0.md) (15 critères, tous bloquants).
>
> Précède Lot 0. C'est un produit **autonome** qui sert deux objectifs : (1) livrer une valeur immédiate à un seul utilisateur (Louis) sans attendre la boucle pédagogique complète ; (2) valider la qualité opérationnelle de la skill `study-guide` (le prompt système qui fabrique la fiche) sur des données réelles, avant que le pipeline LLM soit ré-implémenté en Go côté Lot 0/MVP.

#### Périmètre

| ✅ Inclus | ❌ Exclu (différé Lot 0) |
|---|---|
| Upload multi-photos via formulaire web (1-30 photos, jpg/png/heic/webp). | Comptes utilisateurs / authentification. |
| Génération d'une **fiche HTML autonome unique** (auto-suffisante, lisible hors ligne). | Persistance des photos ou des fiches au-delà du TTL job (1 h). |
| Choix de la **note visée** (12-14 / 15-17 / 18-20) qui module la profondeur de la fiche. | Suivi de mastery, sessions répétées, spaced repetition. |
| Sections 1-6 + Section 8 (annexes) + Section 9 (contrôle blanc) selon la cible. | Reporting parent, notifications, emploi du temps. |
| **Drawer hamburger** + **deux boutons d'impression** (élève / parent) + mode `@media print`. | Stockage S3 ou base de données. |
| `/preview` pour tests sans LLM. | Multi-tenant, RGPD J+30, admin backoffice. |
| Image OCI publique (`ghcr.io/popul/revisemieux-web-v0`) déployée en K3s par ArgoCD. | Mobile native (Lot -1 reste 100% web responsive). |

#### Architecture

```
[Photos] ──POST /jobs──▶ FastAPI ──HTTP──▶ LM Studio (Qwen 3.6 35B-A3B vision)
   │                       │
   │                       ▼
   │                  Pydantic Fiche (schema.py)
   │                       │
   │                       ▼
   └──poll /jobs/{id}──▶ Jinja shell (shell.html.j2) ──▶ HTML autonome
                          ↑
                          └── /preview (fixture.json)
```

- **Stack** : Python 3.12, FastAPI, Pydantic v2, Jinja2, Uvicorn.
- **LLM** : Qwen 3.6 35B-A3B via LM Studio (compatible OpenAI), tournant sur la machine de dev locale du développeur. Endpoint configurable `LLM_URL`.
- **Pipeline** : un seul appel LLM en mode thinking activé (`chat_template_kwargs.enable_thinking=true`), avec rappel des critères d'acceptance ([AC-CONT-*], [AC-HTML-*], [AC-15-*], [AC-20-*]) en fin de prompt utilisateur pour saliency.
- **Source du prompt** : [`prompts/study-guide/system.md`](../prompts/study-guide/) — **partagée** avec la skill Claude Code `/study-guide` (cf. § 3.0).

#### Personas

Reprend le persona « Louis » du § 4 mais dans un usage solo, à la veille du contrôle. Le parent n'utilise pas le Lot -1 directement (sauf pour imprimer la version « Parent » et corriger avec Louis le lendemain).

#### KPIs Lot -1

| KPI | Cible | Mesure |
|---|---|---|
| Délai génération E2E (12 photos, cible 18-20) | P95 ≤ 5 min, P50 ≤ 3 min | logs `elapsed=…` Z0-AC10 |
| Conformité AC pédagogiques | 100% des [AC-CONT-*] et [AC-HTML-*] applicables | audit manuel + `grep` sur HTML |
| Taux de troncature | 0% (pas de `finish=length`) | logs `output_truncated` Z0-AC10 |
| Adoption | Louis utilise pour ≥ 1 contrôle / semaine | autodéclaration |
| Coût | 0 € (LLM local, GHCR public, K3s homelab) | — |

#### Risques & mitigations

| Risque | Impact | Mitigation |
|---|---|---|
| Le LLM tronque la fiche (output > `max_tokens`) | fiche partielle = inutilisable | logs warning `output_truncated`, alerte explicite, `max_tokens=32000` actuel |
| Le shell HTML diverge des fiches passées (chrome cassé) | drawer/print buttons KO | endpoint `/preview` + smoke test à chaque déploiement (cf. Z0-AC12) |
| Photos envoyées à un LLM tiers | leak privé | LLM tourne 100% en local sur la machine du dev (pas de cloud) |
| LM Studio offline | service down | sortie en erreur explicite côté pod, pas de fallback (acceptable pour Lot -1) |
| Le prompt source diverge entre Lot -1 et la skill Claude Code | régression silencieuse | `make sync-skill-prompt` + CI `skill-prompt-sync` (cf. § 3.0) |

#### Phases d'implémentation Lot -1

Cf. [`lot-minus-1-tracker.md`](lot-minus-1-tracker.md). Quatre phases, dont 0 et 1 sont livrées :

- **Phase 0** — Shell Jinja + schéma Pydantic + endpoint `/preview` ✓
- **Phase 1** — Pipeline 1-stage HTML (legacy actuel, conservé tant que Stage 2 n'est pas prêt) ✓
- **Phase 2** — Pipeline 2-stage JSON : LLM → JSON validé → render Jinja (en cours)
- **Phase 3** — Itération sur les 15 ACs Z0 + ouverture publique (rate-limit Cloudflare Access)

---

## 4. Personas

### Élève collège (11–15 ans)
- Motivation variable selon la proximité du contrôle
- Besoin de guidance et de micro-objectifs clairs
- Temps court : 10–20 min/jour réalistes
- Sensible au feedback positif et aux indicateurs de progression

### Parent
- Veut un signal fiable, lisible en < 30 s
- Peu intrusif par défaut, actionnable si sollicité
- Ne veut pas corriger lui-même les exercices
- Déclenche une intervention uniquement sur alerte rouge

### Admin produit
- Gère packs, lexiques, gabarits en backoffice
- Surveille la qualité des items générés
- Ajuste les paramètres par matière et par niveau
- Analyse les analytics par template_id et par tag

---

## 5. Parcours utilisateurs

### 5.1 Parcours élève (standard)

> **Principe first value rapide :** dès la fin de l'upload, l'élève peut accéder à une version préliminaire de la carte de leçon et lancer un premier exercice de reconnaissance (QCM flash) pendant que l'OCR complète son traitement en arrière-plan.

| Étape | Description |
|---|---|
| **Onboarding — Zone scolaire** | Saisie de la zone scolaire (A / B / C / DOM-TOM) pour intégrer le calendrier des vacances. Saisie unique, modifiable dans les paramètres. |
| **J0 — Emploi du temps contextuel** | Lors du premier upload d'une matière, l'app demande « Quand as-tu [Matière] ? » avec une grille jour × période (matin/après-midi). L'emploi du temps se construit progressivement, matière par matière, au rythme des uploads. Modifiable à tout moment dans les paramètres. |
| **J0 — Création** | Création chapitre : matière, classe, nom, date contrôle (fortement encouragée). Parcours < 60 s. Si date présente, plan calibré automatiquement. |
| **J0 — Upload** | Upload 1–30 photos. Segmentation immédiate (blocs typés). Carte de leçon préliminaire affichée dès la première page traitée (streaming). |
| **J0 — Validation** | 0 à 8 validations critiques proposées par ordre d'impact estimé sur la note. Le reste est marqué UNCERTAIN et non bloquant. |
| **J0 — Diagnostic** | Diagnostic initial 5–10 min → carte de maîtrise → plan personnalisé. Questions pré-générées (lazy) disponibles immédiatement après validation. |
| **J0 soir — 1re révision** | Le soir même après la prise de cours, micro-session de première appropriation (5–10 min). Focus : définitions, vocabulaire clé, structure du cours. Même sans date de contrôle connue. |
| **Soir après cours** | Notification push le soir après chaque journée de cours : « Tu as eu [Matière] aujourd'hui — saisis ton cours pour commencer à réviser ce soir ! ». Déclenchée via l'emploi du temps. |
| **Veille de cours** | Révision anticipation surprise : la veille de chaque cours d'une matière, le service propose une mini-session de rappel (5 min) sur les chapitres actifs de cette matière. Objectif : être prêt en cas d'interro surprise. |
| **J+1..J-3** | Micro-sessions quotidiennes adaptatives (70% dû/fragile, 20% consolidation, 10% découverte). Feedback immédiat + explication « pourquoi cet exercice ». |
| **J-3** | Contrôle blanc n°1 : format proche du vrai contrôle, incluant doc + exercice méthode. Correction guidée avec rubriques. Un contrôle peut couvrir plusieurs chapitres — le contrôle blanc les inclut tous. |
| **J-1** | Contrôle blanc n°2 (si date connue). Synthèse : points solides / fragiles / recommandations d'urgence. |

### 5.2 Parcours parent passif (défaut)

- **Digest hebdo (5 sections standardisées, AC Z6-AC27) :**
  1. **Résumé activité** : sessions complétées, temps total. Mention explicite si aucune activité.
  2. **Maîtrise par chapitre** : % OK+SOLID, items restants. Mention si items en vérification ou pages OCR échouées (AC Z6-AC28).
  3. **Alertes** : items à risque pré-exam, sessions manquées (notification lendemain AC Z6-AC18, résumé hebdo AC Z6-AC26), inactivité.
  4. **Prochaine action** : recommandation concrète et actionnable.
  5. **Score contrôle blanc** (si complété) : score + indicateur de confiance (AC Z1-AC16).
  - Labels de maîtrise traduits en langage parent : « Pas encore vu / En cours / Compris / Bien acquis » (AC Z6-AC29).
  - Lisible en < 30 s (max 150 mots hors titres).
- **Digest pré-contrôle (AC Z6-AC25) :** envoyé à J-3 avant chaque exam. Maîtrise par chapitre + items fragiles + recommandation.
- **Veille contrôle :** alerte 5 jours avant si maîtrise insuffisante sur des items critiques (définis par pack).
- **Score contrôle blanc :** résumé envoyé automatiquement après chaque contrôle blanc complété, avec caveat qualité si items non validés.

### 5.3 Parcours parent actif (opt-in)

- **Validations critiques :** le parent peut confirmer/corriger 1–3 items bloquants avec contexte visuel (photo du cahier + suggestion IA). UI simplifiée, jamais plus de 3 actions/semaine.
- **Script 3 minutes :** 2–3 questions à poser à l'enfant pour tester à l'oral les points les plus fragiles.
- **Alerte rouge :** déclenchée seulement si maîtrise < seuil critique à J-3 et progression stagnante depuis 5 jours.

---

## 6. Tags — Spécification complète

Un tag est une étiquette attachée à un Item (et optionnellement à un Document) servant de signal pour le matching des gabarits, la sélection des exercices en session, le paramétrage de la correction et le reporting.

### 6.1 Taxonomie (vocabulaire limité, ≤ 60 tags MVP)

| Famille | Tags | Rôle |
|---|---|---|
| **Skill** (génériques) | `definition` · `methode` · `document` · `redaction` · `calcul` · `unites` · `vocabulaire` · `piege` | Matching gabarits, paramétrage correction |
| **Doc** (génériques) | `map` · `graph` · `table` · `schema` · `photo` · `text` · `experiment` · `circuit` | Matching gabarits typés document |
| **Chapter** (par pack) | `inegalites` · `feodalite` · `photosynthese` · `densite` | Filtrage scope leçon |
| **Concept** (par pack) | `idh` · `pib_hab` · `pma` · `vassal` · `suzerain` · `chloroplaste` · `chlorophylle` · `co2` · `o2` · `rho` · `masse` · `volume` … | Enrichissement pack, analytics confusion |

### 6.2 Attribution des tags

- **Auto-tagging (principal) :** marqueurs de structure (« Définition : », « Méthode : »), type de bloc (map/graph/table…), lexique pack.
- **Pack enrichment (déterministe) :** mapping de termes prioritaires (ex. `idh`, `rho`).
- **Validation humaine (admin) :** ajustements en backoffice pour cas limites. Pas de tags libres côté élève en MVP.

### 6.3 Regroupement en Notions

Après le tagging (étape 6 du pipeline), les items sont regroupés en **Notions** — clusters sémantiques nommés (3–7 par chapitre). Le LLM analyse les `concept_tags` attribués et produit des regroupements lisibles (ex : tags `rho`, `masse`, `volume` → Notion « Masse volumique (ρ) »). Chaque Notion reçoit un nom pédagogique, un ordre d'affichage, et une liste d'items. L'entité `Notion` est définie en §16. Les Notions sont utilisées pour :

- **Vue chapitre** : accordéon par notion avec barre de maîtrise individuelle.
- **Sélection périmètre exam** : l'élève coche/décoche des Notions (pas des items).
- **Aide contextuelle** : le bouton « Voir ma leçon » (en session) affiche la section OCR de la Notion pertinente.

Le regroupement est déterministe pour un même contenu. En cas d'ajout de pages (upload incrémental), les nouveaux items sont intégrés dans les Notions existantes ou de nouvelles Notions sont créées si les `concept_tags` ne correspondent à aucun cluster existant.

### 6.4 Exploitation runtime

| Usage | Règle |
|---|---|
| Matching Template ↔ Item | `required_tags_any`, `required_tags_all`, `excluded_tags` |
| Sélection session | HG : ≥1 exercice document si disponible · PC : push `unites` si erreurs récurrentes |
| Correction `unites` | Unité obligatoire dans la réponse numérique |
| Correction `document` | Checklist documentaire (décrire→prélever→expliquer→conclure) |
| Correction `redaction` | Rubrique 4 critères max |
| Correction `piege` | QCM misconception avec distractor justifié |

---

## 7. Gabarits — Bibliothèque générique + Packs

Les gabarits décrivent la **forme** de l'exercice (réutilisable, indépendant de la leçon). Les packs décrivent **sur quoi** appliquer les gabarits (tags, paramètres, seuils). Cette séparation garantit l'extensibilité à de nouveaux chapitres sans dupliquer la logique.

### 7.1 Modèle Template (schéma JSON simplifié)

```json
{
  "template_id": "GEN.KNOW.DEF_SHORT",
  "name": "Définition — réponse courte",
  "version": 1,
  "question_type": "SHORT_ANSWER",
  "difficulty": 1,
  "eligibility": {
    "item_types": ["KNOWLEDGE"],
    "required_tags_any": ["definition"],
    "min_item_confidence": 0.6,
    "requires_validation_resolved": false
  },
  "variables": [
    {"key": "TERM", "source": "item.term", "required": true},
    {"key": "KEYWORDS", "source": "item.keywords", "required": true}
  ],
  "prompt_template": "Définis {TERM}.",
  "expected_answer": {
    "kind": "KEYWORDS",
    "min_keywords_required": 2,
    "synonyms": {}
  },
  "grading": {"mode": "AUTO", "max_points": 1}
}
```

### 7.2 Moteur d'exploitation (lazy generation)

> **Lazy generation :** les Questions ne sont plus pré-générées en masse après ingestion. Elles sont instanciées à la demande lors de la composition de session (ou du contrôle blanc). Un pool de candidats est mis en cache par chapitre avec TTL 24h. Cela réduit le coût LLM d'un facteur 5–10x et simplifie le pipeline J0.

1. **Matching :** filtre des templates éligibles via type + tags + confiance + état validation.
2. **Composition session :** sélection selon politique 70/20/10 + contraintes pack (1 doc max, 1 rédaction max).
3. **Génération de documents supports :** pour les questions documentaires, le système **génère des documents originaux inédits** (tableaux de données avec valeurs à calculer, graphiques avec données différentes du cours, schémas avec légendes vides, cartes avec zones à identifier). Ces documents ne sont pas des copies du cahier — ils utilisent des valeurs et contextes différents tout en testant les mêmes notions. Formats : Markdown tables, Mermaid, SVG inline, ASCII art. Minimum 1 document par session si des items DOCUMENT existent.
4. **Instanciation :** remplissage des variables → création Question (avec référence au document généré le cas échéant) → stockage en cache session.
5. **Correction :** AUTO (QCM/keywords/numeric) ou RUBRIC (doc/rédaction).
6. **Maîtrise :** UNKNOWN → FRAGILE → OK → SOLID (SOLID = 2 réussites espacées ≥ 24h).

---

## 8. Bibliothèque de gabarits (27 templates MVP)

### Core — toutes matières

| template_id | Type | Diff. | Tags requis |
|---|---|---|---|
| `GEN.KNOW.DEF_SHORT` | SHORT_ANSWER | 1 | `definition` |
| `GEN.KNOW.FLASH_MCQ` | MCQ | 1 | — |
| `GEN.KNOW.ASSOC_TERM_DEF` | MATCHING | 2 | `definition`, `vocabulaire` |
| `GEN.KNOW.CLOZE_KEYWORDS` | FILL_BLANK | 2 | — |
| `GEN.MISCONCEPTION.MCQ` | MCQ | 2 | `piege` |

### Documents — transversal

> **Génération de documents originaux :** les templates documentaires ci-dessous peuvent s'appuyer soit sur des visuels extraits du cahier (via `VisualBlock`), soit sur des **documents générés à la volée** par le système. Les documents générés utilisent des valeurs et contextes inédits (pas de copie du cahier) pour tester les mêmes notions sous un angle nouveau. Cela renforce le double codage et empêche l'apprentissage par simple reconnaissance visuelle du document original.
>
> **Types de documents générables :**
> - **Tableaux de données** (Markdown) — valeurs à calculer (`?`), unités variées, colonnes d'identification
> - **Graphiques** (Mermaid xychart-beta, SVG, ASCII) — courbes d'évolution, histogrammes, diagrammes circulaires
> - **Schémas légendés** (SVG, Mermaid flowchart) — légendes vides pour LABEL_COMPLETION, numérotation des éléments
> - **Cartes simplifiées** (SVG) — zones lettrées à identifier, contours schématiques
> - **Diagrammes de classification** (Mermaid mindmap) — hiérarchies avec éléments à compléter

| template_id | Type | Diff. | Tags requis |
|---|---|---|---|
| `GEN.DOC.PRESENT` | CHECKLIST | 2 | `document` |
| `GEN.DOC.DESCRIBE` | SHORT_ANSWER | 1 | `document` |
| `GEN.DOC.EXTRACT_EVIDENCE` | SHORT_ANSWER | 2 | `document` |
| `GEN.DOC.INTERPRET_WITH_CONCEPT` | SHORT_ANSWER | 3 | `document` |
| `GEN.DOC.CONCLUDE` | SHORT_ANSWER | 3 | `document` |

### Documents typés

| template_id | Type | Diff. | Tags requis |
|---|---|---|---|
| `GEN.DOC.MAP.READ_ZONES` | SHORT_ANSWER | 2 | `map` |
| `GEN.DOC.MAP.LABEL_COMPLETION` | FILL_BLANK | 2 | `map` |
| `GEN.DOC.GRAPH.READ_AXES` | SHORT_ANSWER | 1 | `graph` |
| `GEN.DOC.GRAPH.READ_VALUE` | SHORT_ANSWER | 2 | `graph` |
| `GEN.DOC.GRAPH.INTERPRET_TREND` | SHORT_ANSWER | 3 | `graph` |
| `GEN.DOC.TABLE.READ_VALUE` | SHORT_ANSWER | 2 | `table` |
| `GEN.DOC.TABLE.COMPLETE_CELL` | FILL_BLANK | 2 | `table` |
| `GEN.DOC.SCHEMA.LABEL_COMPLETION` | FILL_BLANK | 2 | `schema` |
| `GEN.DOC.SCHEMA.FUNCTION_MATCHING` | MATCHING | 3 | `schema` |
| `GEN.DOC.CIRCUIT.IDENTIFY_COMPONENT` | SHORT_ANSWER | 2 | `circuit` |
| `GEN.DOC.IMAGE.DESCRIBE_INTERPRET` | SHORT_ANSWER | 2 | `photo`, `schema` |

### Méthodes & rédaction

| template_id | Type | Diff. | Tags requis |
|---|---|---|---|
| `GEN.PROC.STEPS.ORDER` | ORDERING | 2 | `methode` |
| `GEN.PROC.STEPS.MISSING` | FILL_BLANK | 2 | `methode` |
| `GEN.WRITE.PARAGRAPH_GUIDED` | RUBRIC | 3 | `redaction` |
| `GEN.WRITE.BILAN_GUIDED` | RUBRIC | 3 | `redaction` |

### PC — Calcul (génériques discipline)

| template_id | Type | Diff. | Tags requis |
|---|---|---|---|
| `PC.UNITS.CONVERT` | NUMERIC | 2 | `unites`, `calcul` |
| `PC.FORMULA.APPLY` | NUMERIC | 2 | `calcul` |
| `PC.FORMULA.ISOLATE` | NUMERIC | 3 | `calcul` |
| `PC.MEASURE.DISPLACEMENT_VOLUME` | NUMERIC | 2 | `calcul` |
| `PC.TABLE.ID_MATERIAL` | MCQ | 2 | `document` |
| `PC.REASON.FLOAT_SINK` | SHORT_ANSWER | 2 | `calcul` |
| `PC.ERROR.UNITS_MCQ` | MCQ | 2 | `piege`, `unites` |
| `PC.PROBLEM.SOLUTION_RUBRIC` | RUBRIC | 3 | `redaction`, `calcul` |

---

## 9. Packs matière

### Pack HG-INEG — HG — Inégalités dans le monde

| | |
|---|---|
| **chapter_tag** | `inegalites` |
| **concept_tags** | `idh` · `pib_hab` · `pma` · `developpement` · `inegalites` |
| **templates activés** | `GEN.KNOW.DEF_SHORT` · `GEN.KNOW.FLASH_MCQ` · `GEN.KNOW.CLOZE_KEYWORDS` · `GEN.DOC.PRESENT` · `GEN.DOC.MAP.READ_ZONES` · `GEN.DOC.GRAPH.READ_AXES` · `GEN.DOC.GRAPH.READ_VALUE` · `GEN.DOC.TABLE.READ_VALUE` · `GEN.DOC.EXTRACT_EVIDENCE` · `GEN.DOC.INTERPRET_WITH_CONCEPT` · `GEN.DOC.CONCLUDE` · `GEN.WRITE.PARAGRAPH_GUIDED` · `GEN.MISCONCEPTION.MCQ` |
| **paramètres** | `min_evidence_required = 2` · `max_writing_per_session = 1` · `min_keywords_def = 2` · `session_must_include_doc = true` |

### Pack HG-FEOD — HG — Société féodale

| | |
|---|---|
| **chapter_tag** | `feodalite` |
| **concept_tags** | `suzerain` · `vassal` · `fief` · `hommage` · `feodalite` · `seigneurie` |
| **templates activés** | `GEN.KNOW.DEF_SHORT` · `GEN.KNOW.ASSOC_TERM_DEF` · `GEN.KNOW.FLASH_MCQ` · `GEN.DOC.PRESENT` · `GEN.DOC.IMAGE.DESCRIBE_INTERPRET` · `GEN.DOC.EXTRACT_EVIDENCE` · `GEN.DOC.CONCLUDE` · `GEN.WRITE.PARAGRAPH_GUIDED` · `GEN.MISCONCEPTION.MCQ` |
| **paramètres** | `min_evidence_required = 2` · `max_writing_per_session = 1` |

### Pack SVT-PHOTO — SVT — Photosynthèse

| | |
|---|---|
| **chapter_tag** | `photosynthese` |
| **concept_tags** | `chloroplaste` · `chlorophylle` · `co2` · `eau` · `lumiere` · `o2` · `matiere_organique` · `facteurs_limitants` |
| **templates activés** | `GEN.KNOW.DEF_SHORT` · `GEN.KNOW.ASSOC_TERM_DEF` · `GEN.KNOW.CLOZE_KEYWORDS` · `GEN.DOC.IMAGE.DESCRIBE_INTERPRET` · `GEN.DOC.GRAPH.READ_AXES` · `GEN.DOC.GRAPH.READ_VALUE` · `GEN.DOC.EXTRACT_EVIDENCE` · `GEN.DOC.INTERPRET_WITH_CONCEPT` · `GEN.DOC.CONCLUDE` · `GEN.WRITE.BILAN_GUIDED` · `GEN.MISCONCEPTION.MCQ` |
| **paramètres** | `min_keywords_def = 3` · `doc_checklist_requires_concept = true` |

### Pack PC-MVD — PC — Masse, volume et densité

| | |
|---|---|
| **chapter_tag** | `densite` |
| **concept_tags** | `masse` · `volume` · `rho` · `unites` · `deplacement_eau` · `materiaux` |
| **templates activés** | `GEN.KNOW.DEF_SHORT` · `PC.UNITS.CONVERT` · `PC.FORMULA.APPLY` · `PC.FORMULA.ISOLATE` · `PC.MEASURE.DISPLACEMENT_VOLUME` · `PC.TABLE.ID_MATERIAL` · `PC.REASON.FLOAT_SINK` · `PC.ERROR.UNITS_MCQ` · `PC.PROBLEM.SOLUTION_RUBRIC` |
| **paramètres** | `numeric_tolerance_percent = 2` · `unit_required = true` · `max_long_problem_per_session = 1` · `intermediate_rounding_allowed = true` (arrondi à 2 décimales) |

---

## 10. Pipeline J0 — Flux de données détaillé

> Ce pipeline explicite les étapes, la parallélisation possible, les points de latence critiques et les dépendances bloquantes vs non-bloquantes.

| Étape | Entrée | Sortie | Bloquant ? | SLA cible |
|---|---|---|---|---|
| 1. Upload & validation | Photos (JPEG/PNG/HEIC) | URLs stockage, métadonnées | Oui | < 3 s/photo |
| 2. Segmentation blocs | Photo | Crops + type + confidence + coordonnées + classification pédagogique | Non | < 1 s/page |
| 3. OCR parallèle | Blocs TEXT/SCHEMA | Texte brut + confidence | Non | < 2 s/bloc |
| 4. Reconstruction plan | Texte OCR | Plan hiérarchique JSON | Non | < 500 ms |
| 5. Génération Items | Plan + blocs | Items KNOWLEDGE/PROC/DOC | Non | < 3 s/page |
| 5b. Génération docs supports | Items DOCUMENT + blocs visuels | Documents originaux (tableaux, graphiques, schémas) | Non | < 2 s/item |
| 6. Auto-tagging | Items + lexique pack | Items taggés | Non | < 200 ms |
| 6b. Regroupement Notions | Items taggés | Notions nommées (3–7 clusters par chapitre) | Non | < 500 ms |
| 7. Détection incertains | Items + confidence | File validation (max 8) | Oui (si critiques) | < 500 ms |
| 7b. Vérification croisée LLM | Items + texte OCR source | Score fidélité sémantique | Non | < 2 s/page |
| 7c. Cohérence intra-chapitre | Tous items du chapitre | Doublons/contradictions flaggés | Non | < 1 s |
| 8. Carte leçon (draft) | Items + Notions + blocs OCR | Carte navigable (vue calculée, pas d'entité persistée) | Non (streaming) | Dès page 1 prête |
| 9. Validation HITL | File validation | Items validés/corrigés | Partiel | Élève, asynchrone |
| 10. Diagnostic initial | Items validés | Questions instanciées (lazy) | Oui | < 1 s |

**Parallélisation :** les étapes 2, 3, 4, 5, 6 peuvent être exécutées en parallèle par page (worker pool). La carte de leçon est affichée en streaming dès qu'une page est traitée. L'élève peut commencer un QCM flash pendant que les pages suivantes s'analysent.

### 10.1 Détail étape 2 — Segmentation et classification visuelle

> Les cahiers de collégiens contiennent typiquement 30-50% de contenu visuel (schémas SVT, graphiques, tableaux, circuits physique, cartes géographie). La segmentation doit distinguer et préserver ces éléments pour exploiter le dual coding (Paivio).

**Détection et classification des blocs :**

| Type de bloc | Exemples typiques | Traitement OCR | Sortie |
|---|---|---|---|
| TEXT | Paragraphe, définition, titre | OCR standard | `ocr_text` + confidence |
| TABLE | Tableau dessiné ou imprimé | OCR + extraction structure (lignes/colonnes) | `ocr_text` + `table_structure` (JSON rows×cols) |
| SCHEMA | Schéma annoté (cellule, circuit, cycle) | OCR légendes uniquement | `crop_url` + `labels[]` |
| MAP | Carte géographique légendée | OCR légendes uniquement | `crop_url` + `labels[]` |
| GRAPH | Graphique avec axes et courbes | OCR axes + valeurs-clés | `crop_url` + `labels[]` + `axis_labels{}` |
| CIRCUIT | Circuit électrique, schéma de forces | OCR composants/valeurs | `crop_url` + `labels[]` |
| PHOTO | Photographie collée (expérience, document historique) | Pas d'OCR | `crop_url` + description LLM |
| DECORATIVE | Gribouillage, marge, rature | Ignoré | — |

**Qualité des crops visuels :** chaque crop visuel (SCHEMA, MAP, GRAPH, TABLE, CIRCUIT, PHOTO) doit avoir une résolution minimale de 300×300 px (sinon redimensionné via upscaling conservatif) et un cadrage qui inclut les légendes et annotations environnantes (marge de 5% autour de la bounding box détectée). Les crops sont stockés en WebP (qualité 85, compression ~60% vs JPEG) avec l'URL dans `Block.crop_url`.

**Association bloc→Item→VisualBlock :** les blocs visuels pédagogiques sont encapsulés dans une entité `VisualBlock` (voir §14 modèle de données) qui associe le crop, les labels extraits et le type visuel. Chaque Item peut référencer 0-N `VisualBlock` via `Item.visual_block_ids[]`. Le LLM de structuration (étape 5) reçoit les blocs visuels en contexte multimodal et génère des gabarits visuels spécifiques (cf. Z4-AC17).

**Classification pédagogique vs décoratif :** le modèle de segmentation classe chaque bloc non-TEXT en pédagogique ou décoratif. En cas de doute (confidence classification < 0.6), le bloc est classé décoratif (faux négatif préférable au faux positif — un gribouillage classé « schéma » dégraderait l'expérience). La classification est vérifiable dans la carte leçon (Z2-AC15) : le parent/élève peut reclasser un bloc ignoré.

---

## 11. SLA & contraintes de performance

| Métrique | Cible MVP | Seuil d'alerte |
|---|---|---|
| Carte leçon + items (10 pages) | < 2 min (objectif < 90 s) | > 3 min |
| First exercise after upload | < 5 min | > 8 min |
| Composition session (lazy) | < 1 s | > 3 s |
| Contrôle blanc (génération) | < 10 s | > 30 s |
| OCR confidence moyenne | > 0.85 | < 0.70 |
| Taux items confidence >= 0.6 | > 80% | < 60% |
| Disponibilité pipeline J0 | > 99% (hors maintenance) | < 98% |

---

## 12. Stratégie cache & génération lazy

### 12.1 Niveaux de cache

| Niveau | Clé | TTL | Contenu |
|---|---|---|---|
| OCR results | `hash(photo_id + model_version)` | Permanent | Texte brut + confidence blocs |
| Item pool | `chapter_id + pack_version` | 24h | Liste items taggés instanciés |
| Question candidates | `chapter_id + session_type` | 24h | Pool questions éligibles |
| Session courante | `session_id` | 72h | Questions sélectionnées + état |
| Mastery state | `user_id + item_id` | Permanent | État UNKNOWN/FRAGILE/OK/SOLID |

### 12.2 Invalidation du cache

- **Re-upload de page** → invalide OCR results + Item pool + Question candidates du chapter concerné.
- **Validation HITL d'un item** → invalide uniquement les Question candidates liées à cet item.
- **Mise à jour de pack_version** → invalide Item pool + Question candidates de tous les chapters du pack.
- **Changement de mastery state** → pas d'invalidation (in-place update).
- **Création/modification/suppression d'Exam** → invalide Question candidates des chapters liés à l'Exam (AC Z4-AC03). Recalcul `next_due_at` si exam supprimé.

---

## 13. Retry / Fallback OCR

> L'OCR manuscrit est le point de fragilité principal du pipeline. Une stratégie explicite de retry et de fallback est nécessaire pour éviter des items incorrects qui contaminent la maîtrise de l'élève.

| Condition | Action | Impact utilisateur |
|---|---|---|
| confidence < 0.4 sur un bloc TEXT | Marquer bloc UNCERTAIN, ajouter à file validation si item critique | Badge jaune sur la carte leçon |
| confidence < 0.6 sur définition/formule/chiffre | `validation_required = true` sur l'Item | Validation demandée (max 8) |
| OCR timeout (> 10 s par bloc) | Retry x2 avec back-off 2 s, puis fallback modèle secondaire | Transparent (indicateur de chargement) |
| Bloc illisible après 3 tentatives | Item créé avec `content = null`, `flagged = true` | Message 'Zone illisible — à vérifier' |
| Page entière confidence < 0.3 | Notifier l'élève : 'Photo floue — reprendre si possible' | Suggestion retake, non-bloquant |
| confidence bloc SCHEMA/MAP < 0.5 | Conserver l'image brute comme `Document.source`, ne pas OCRiser | Document exploitable via gabarit image |
| Bloc visuel trop petit (< 300×300 px) | Upscaling conservatif (bilinear) jusqu'à 300×300 px min. Si < 100×100 px après crop → classé DECORATIVE | Visuel ignoré si trop petit (probablement un symbole/icône) |
| Bloc TABLE avec structure complexe (≥ 8 colonnes ou cellules fusionnées) | Fallback : capture comme image (SCHEMA) au lieu de table structurée | Tableau exploitable en mode image, pas en mode cellule |
| Labels/légendes extraites avec confidence < 0.4 | Labels conservés comme `uncertain`, affichés avec badge dans carte leçon | Le parent peut corriger les labels via HITL |

### 13.1 Vérification croisée LLM (fidélité sémantique)

> L'OCR peut être correct mais l'item généré par le LLM peut déformer le sens du cours. Un second appel LLM compare chaque item au texte OCR source.

- **Étape 7b du pipeline** : après génération des items, un prompt dédié évalue la fidélité sémantique de chaque item par rapport au texte OCR source.
- Le LLM produit un `fidelity_score` (0–1) et une `fidelity_reason` pour chaque item.
- **Seuils** : `fidelity_score < 0.5` → `validation_required = true` + `fidelity_flag = 'low'`. `fidelity_score 0.5–0.7` → item utilisable mais `fidelity_flag = 'medium'` (prioritaire pour validation HITL si la file n'est pas pleine). `fidelity_score ≥ 0.7` → item validé sémantiquement.
- Le prompt de vérification reçoit : le texte OCR brut du bloc source, l'item généré (term, keywords, steps), et le contexte du pack (matière, tags attendus).
- **Non-bloquant** : si le service LLM de vérification timeout, l'item est conservé avec `fidelity_score = null` (comportement dégradé = confiance OCR seule).

### 13.2 Cohérence intra-chapitre

> Les items d'un même chapitre doivent être cohérents entre eux : pas de doublons, pas de contradictions.

- **Étape 7c du pipeline** : après vérification croisée, tous les items du chapitre sont comparés deux à deux.
- **Doublons** : deux items avec un `term` identique ou une similarité cosinus des `keywords` > 0.9 → le doublon de plus faible confidence est archivé automatiquement.
- **Contradictions** : deux items du même chapitre avec des définitions contradictoires (détection LLM) → les deux sont flaggés `validation_required = true` avec `coherence_flag = 'contradiction'`.
- **Termes orphelins** : un item référençant un concept non défini par aucun autre item du chapitre → `coherence_flag = 'orphan_reference'` (informatif, non bloquant).

### 13.3 Feedback élève sur items

> L'élève peut signaler une erreur sur un item ou une question à tout moment pendant une session.

- **Bouton « Signaler une erreur »** visible sur chaque question et sur chaque item de la carte de leçon.
- Le signalement crée une `ValidationTask` avec : `source = 'student_report'`, `item_id`, `crop_url` (contexte visuel), texte libre optionnel de l'élève.
- Les signalements élèves sont prioritaires dans la file de validation (affichés en premier).
- **Déduplication** : si une `ValidationTask` existe déjà pour cet item (quelque soit la source), le signalement élève est ajouté comme note sur la tâche existante (pas de doublon).
- L'item signalé reste utilisable en mode dégradé (templates simples uniquement) jusqu'à résolution.

### 13.4 Détection par taux d'échec anormal

> Un item avec un taux d'échec anormalement élevé est probablement mal extrait.

- **Job quotidien** : pour chaque item avec `≥ 5 tentatives`, calcul du taux d'échec sur les 7 derniers jours.
- **Seuil** : taux d'échec `> 80%` ET `≥ 5 tentatives` → `validation_required = true` avec `anomaly_flag = 'high_failure_rate'`.
- Une `ValidationTask` est créée automatiquement avec `source = 'anomaly_detection'`, `suggestion = 'Taux d\'échec anormal (X%) — vérifier l\'item'`.
- **Exception** : les items en état UNKNOWN ne sont pas concernés (taux d'échec élevé attendu à la première exposition).
- **Auto-résolution** : si le taux d'échec repasse sous 50% après correction ou nouvelles tentatives, le flag est retiré automatiquement.

---

## 14. Architecture cible (MVP)

### 14.1 Vue d'ensemble des composants

| Composant | Techno | Rôle |
|---|---|---|
| Backend API | Go + Gin | Auth, routing, rate-limit, streaming SSE pour pipeline J0 |
| OCR Service | Go (async worker) | Segmentation + OCR parallèle par page (worker pool) |
| LLM Service | Anthropic API (Sonnet 4.6 + Haiku 4.5) | Structuration + cohérence (Sonnet) ; fidelity check + questions (Haiku). Cf. `docs/llm-strategy.md`. |
| Cache Layer | Redis | OCR results, item pool, question candidates, sessions |
| Base de données | PostgreSQL (pgx, SQL brut) | Users, Chapters, Items, Mastery, Attempts |
| Storage | S3 / Object storage | Photos originales (opt-in), crops indexés |
| Mobile | React Native + Expo + Expo Router | Mobile-first, SSE pour affichage streaming carte leçon |

### 14.2 Versioning des chapitres

> Un élève peut re-uploader des pages (correction d'une photo floue, ajout de pages manquantes). Chaque re-upload crée une **ChapterRevision** immutable. La révision courante est la dernière validée. Les Mastery states sont préservés entre révisions.

- **ChapterRevision** : `chapter_id` + `revision_number` + `created_at` + `pages[]`.
- Les Items créés dans une révision antérieure sont **archivés** (non supprimés) pour préserver l'historique de maîtrise.
- Un Item identique (même term + même pack) retrouvé dans une nouvelle révision hérite du Mastery state existant.
- En cas de conflit (OCR différent sur même bloc), l'item est re-soumis à validation HITL.

---

## 15. Fonctionnalités et exigences (Epics)

### Epic 1 — Onboarding & chapitres

Création chapitre : matière, classe, nom, date contrôle. Saisie de l'emploi du temps hebdomadaire lors de l'onboarding (une seule fois, modifiable).

**Critères d'acceptation :**
- Parcours création chapitre < 60 s.
- Saisie emploi du temps < 3 min (interface par matière + créneau).
- Date présente → plan calibré automatiquement.
- Sans date → révision proactive activée dès J0 soir + nudge UI répété à J+3 si toujours absente.

### Epic 2 — Upload & segmentation

Upload 1–30 photos mobile/web. Segmentation en blocs typés (TEXT, PHOTO, SCHEMA, MAP, GRAPH, TABLE, CIRCUIT, DECORATIVE). Classification pédagogique vs décoratif. Extraction des légendes, labels d'axes, structures de tableaux. Stockage des crops visuels en WebP. Streaming de la carte dès la première page.

**Critères d'acceptation :**
- Crops avec type + confidence + coordonnées (bounding box).
- Classification pédagogique/décoratif pour chaque bloc non-TEXT (cf. §10.1).
- Extraction labels/légendes pour blocs SCHEMA, MAP, GRAPH, CIRCUIT (cf. Z2-AC15).
- Extraction structure rows×cols pour blocs TABLE (cf. Z2-AC16).
- Résolution minimale 300×300 px par crop visuel, marge 5% (cf. Z2-AC17).
- Indicateur de progression temps réel.
- Suggestion retake si page floue (confidence < 0.3).

### Epic 3 — OCR & structuration

OCR manuscrit. Reconstruction plan hiérarchique. Stratégie retry/fallback (voir §13).

**Critères d'acceptation :**
- Carte de leçon navigable générée.
- Confidence par bloc + par item.
- Blocs illisibles flaggés clairement.

### Epic 4 — Items + tags

Génération Items KNOWLEDGE/PROCEDURE/DOCUMENT/WRITING. Auto-tagging + enrichissement pack.

**Critères d'acceptation :**
- Chaque item a `skill_tag` + `doc_tag` + `chapter_tag` + `concept_tag` si reconnu.
- Taux auto-tagging correct > 85% sur chapitres pilotes.

### Epic 5 — Validation HITL

Détection incertains critiques. File de validation : Confirmer / Corriger / Ignorer / Je ne sais pas. Max 8/chapter triés par impact estimé sur la note.

**Critères d'acceptation :**
- Régénération ciblée des questions liées à l'item modifié.
- UX < 2 min pour valider les 8 items.
- Parent peut valider (mode actif) avec contexte visuel photo.

### Epic 6 — Diagnostic & maîtrise

Diagnostic 5–10 min. Carte de maîtrise par item. États UNKNOWN/FRAGILE/OK/SOLID.

**Critères d'acceptation :**
- Historique tentatives consultable.
- Explicabilité : 2 lignes max — « pourquoi cet exercice ».
- SOLID = 2 réussites espacées ≥ 24h.

### Epic 7 — Génération & sessions

Lazy generation questions. Composition session 70/20/10. Contraintes pack : 1 doc/1 rédaction max. **Génération de documents originaux** : pour les exercices documentaires, le système crée des supports inédits (tableaux, graphiques, schémas, cartes) avec des valeurs et contextes différents du cahier, renforçant le double codage et évitant l'apprentissage par reconnaissance visuelle du document source.

**Critères d'acceptation :**
- Session 10–20 min. Reprise après interruption.
- Couverture tag document garantie si items disponibles (HG).
- Documents générés utilisent des données originales (pas de copie du cahier), en formats Markdown/Mermaid/SVG.
- Indicateur de couverture du chapitre visible.

### Epic 8 — Contrôles blancs

≥1 contrôle blanc par chapitre. Format proche du vrai contrôle. Correction guidée + rubriques.

**Critères d'acceptation :**
- Inclut doc + exercice méthode obligatoirement.
- Restitution : points solides / fragiles / recommandations.
- Score archivé pour comparaison J-3 vs J-1.

### Epic 9 — Parents

Mode passif : digest hebdo + veille contrôle. Mode actif opt-in : validations + alertes rares. Script 3 minutes.

**Critères d'acceptation :**
- Digest lisible < 30 s.
- Alertes actionnables (max 1/semaine en mode passif).
- UX validation parent simplifiée (≤3 clics).

### Epic 10 — Admin & qualité

CRUD packs (templates activés, lexiques tags, paramètres). Analytics par template_id + par tag.

**Critères d'acceptation :**
- Taux réussite / temps moyen / partial par template.
- Zones de confusion identifiées par tag (ex. `unites`, `document`).
- Export CSV des métriques qualité.

### Epic 11 — Emploi du temps, notifications & révision proactive

> Le service exploite l'emploi du temps de l'élève pour orchestrer deux boucles proactives : (1) notification saisie de cours le soir + révision immédiate, (2) révision d'anticipation la veille de chaque cours.

**Sous-fonctionnalités :**

**11.1 — Saisie de l'emploi du temps**
- Saisie manuelle lors de l'onboarding : matière + jour + créneau horaire (matin/après-midi suffit).
- Modifiable à tout moment dans les paramètres.
- L'emploi du temps déclenche les notifications et la planification des sessions.
- **Vue semaine avec actions contextuelles** : les créneaux récurrents sont affichés sur une grille semaine. Tap sur un créneau → menu contextuel : « Annuler cette semaine » / « Déplacer ». Ajout ponctuel via « + » sur un jour vide.

**11.1b — Exceptions ponctuelles (annulation / déplacement)**
- Un cours peut être **annulé** pour une date précise (ex. « pas de PC mardi 11 mars »). Le créneau récurrent reste inchangé pour les semaines suivantes.
- Un cours peut être **déplacé** vers un autre jour/période pour une semaine donnée (ex. « PC déplacé de mardi matin à jeudi après-midi cette semaine »). Cela crée une annulation sur le créneau d'origine + un ajout ponctuel sur le nouveau créneau.
- Les exceptions sont modélisées par l'entité `ScheduleException` (voir §16).
- **Impact notifications** : les notifications `capture_reminder` et `pre_class` s'ajustent automatiquement. Un cours annulé ne déclenche aucune notification. Un cours déplacé déclenche les notifications sur le nouveau jour.
- **Impact sessions** : la session `pre_class` est planifiée la veille du créneau effectif (après prise en compte des exceptions), pas du créneau récurrent.
- Les exceptions passées (date < aujourd'hui) sont ignorées par le scheduler. Nettoyage automatique des exceptions de plus de 30 jours.

**11.2 — Notification saisie de cours (soir après cours)**
- Le soir (heure configurable, défaut 18h30) après une journée où l'élève a eu cours dans une matière suivie, notification push : « Tu as eu [Matière] aujourd'hui — saisis ton cours pour réviser ce soir ! ».
- Si le chapitre est déjà saisi, la notification devient : « Révise tes points fragiles en [Matière] — 10 min ce soir ».
- Fréquence max : 2 notifications/soir. Priorité aux matières avec contrôle proche.

**11.3 — Révision proactive le soir (J0 soir)**
- Dès qu'un cours est capturé, le service propose une micro-session de première appropriation le soir même (5–10 min).
- Contenu : QCM flash, définitions clés, cloze texte à trous — les gabarits les plus simples (difficulté 1).
- Objectif : ancrage initial. Même sans date de contrôle. Le plan d'étude démarre immédiatement.

**11.4 — Révision veille de cours (anticipation surprise)**
- La veille au soir de chaque cours d'une matière, mini-session de rappel (5 min) sur les chapitres actifs de cette matière.
- Sélection : items FRAGILE/OK les moins récemment révisés + items UNKNOWN non encore vus.
- Justification affichée : « Tu as [Matière] demain — prépare-toi en cas d'interro surprise ».

**11.5 — Exams multi-chapitres**
- Un contrôle peut couvrir plusieurs chapitres/leçons. L'`exam_date` est attachée à un **Exam** qui référence N chapitres.
- Les contrôles blancs couvrent l'ensemble des chapitres liés à l'exam.
- Le plan d'étude répartit les items de tous les chapitres concernés dans les sessions quotidiennes.

**11.6 — Notifications parent (visibilité emploi du temps & suivi)**
- Le compte parent reçoit des notifications push et un tableau de bord récapitulatif pour 3 types d'événements :
  1. **Annulation / déplacement de cours** — notification push dès que l'élève crée une `ScheduleException`. Message : « [Prénom] a annulé son cours de [Matière] du [date] » ou « [Prénom] a déplacé son cours de [Matière] du [date] au [nouvelle date] ».
  2. **Session de révision manquée** — notification push le **lendemain matin** (même heure que le rappel élève) si une session `evening_first` ou `pre_class` n'a pas été commencée. Message : « [Prénom] n'a pas fait sa session de révision de [Matière] hier soir ». Pas de notification si l'élève a terminé la session en retard dans la nuit (avant 6h).
  3. **Inactivité prolongée** — notification push si l'élève n'a eu aucune activité (ni capture, ni session, ni review) depuis **3 jours consécutifs**. Message : « [Prénom] n'a pas utilisé ReviseMieux depuis 3 jours ». Envoyée une seule fois par période d'inactivité (pas de spam quotidien).
- **Tableau de bord parent** : vue synthétique listant les événements des 7 derniers jours (annulations, déplacements, sessions manquées, inactivité). Accessible depuis l'app parent.
- **Respect de l'autonomie** : les notifications parent sont informatives, pas bloquantes. L'élève n'est pas notifié que ses parents ont reçu une alerte (pas de mécanique de surveillance visible). Le parent ne peut pas annuler/déplacer les cours depuis son compte.
- **Opt-out parent** : le parent peut désactiver chaque catégorie de notification individuellement dans ses paramètres.

**Critères d'acceptation :**
- Notifications déclenchées correctement selon l'emploi du temps saisi.
- Micro-session soir J0 proposée dans les 4h après upload.
- Session anticipation disponible la veille de chaque cours (si chapitres actifs existent).
- Contrôle blanc multi-chapitres inclut les items de tous les chapitres de l'exam.
- Élève peut désactiver les notifications sans perdre les sessions proactives.
- Sans emploi du temps saisi, le service fonctionne normalement (mode dégradé = pas de notifications proactives, sessions standards uniquement).
- Session manquée : items dues non révisés restent priorisés sans pénalité. Un seul rappel le lendemain matin pour les sessions `evening_first` et `pre_class` non commencées. Aucune mécanique de streak.
- Exceptions emploi du temps : annulation et déplacement ponctuels gérés. Notifications et sessions `pre_class` ajustées automatiquement selon le créneau effectif.
- Notifications parent : push envoyé pour annulations/déplacements (temps réel), sessions manquées (lendemain matin), inactivité (après 3 jours). Chaque catégorie désactivable individuellement par le parent.

---

## 16. Modèle de données (MVP)

| Entité | Champs |
|---|---|
| **User** | `id` · `role (student\|parent)` · `consent_parent_at` · `linked_student_id?` · `linked_student_ids[]?` (parent multi-enfants) · `school_zone? (A\|B\|C\|reunion\|guadeloupe\|martinique\|guyane\|mayotte)` · `notification_hour?` (défaut 18h30) · `timezone?` (défaut Europe/Paris, détection auto) · `evening_window_start?` (défaut 17h) · `evening_window_end?` (défaut 22h) · `weekend_notification_hour?` (défaut 10h) |
| **ScheduleSlot** | `id` · `user_id` · `subject` · `day_of_week (1–7)` · `period (morning\|afternoon)` · `created_at` |
| **ScheduleException** | `id` · `user_id` · `subject` · `original_date` · `type (cancelled\|moved)` · `moved_to_date?` · `moved_to_period?` · `created_at` |
| **SchoolHolidayPeriod** | `id` · `zone (A\|B\|C\|reunion\|guadeloupe\|martinique\|guyane\|mayotte)` · `name (toussaint\|noel\|hiver\|printemps\|ete)` · `start_date` · `end_date` · `school_year (ex: 2025-2026)` |
| **VacationPreference** | `id` · `user_id` · `holiday_period_id` · `revision_days[] (tableau de day_of_week 1-7)` · `preferred_hour (int 8-20)` · `created_at` |
| **UnavailabilitySlot** | `id` · `user_id` · `day_of_week (1-7)` · `label? (ex: "Football")` · `recurring (boolean)` · `created_at` |
| **UnavailabilityException** | `id` · `user_id` · `unavailability_slot_id` · `date` · `type (cancel\|add)` · `created_at` |
| **Exam** | `id` · `user_id` · `name?` · `exam_date` · `chapter_ids[]` · `notion_ids[]?` (si vide = toutes les notions des chapitres) · `status (active\|past)` · `created_at` |
| **Chapter** | `id` · `subject` · `class_level` · `name` · `exam_ids[]` · `pack_id` · `current_revision_id` · `archived? (boolean, default false)` · `is_demo? (boolean, default false)` |
| **ChapterRevision** | `id` · `chapter_id` · `revision_number` · `created_at` · `pages[]` · `status` |
| **Page** | `id` · `revision_id` · `photo_url` · `order` · `ocr_status` |
| **Block** | `id` · `page_id` · `type (TEXT\|PHOTO\|SCHEMA\|MAP\|GRAPH\|TABLE\|CIRCUIT\|DECORATIVE)` · `crop_url?` · `crop_bbox {x, y, w, h}` · `confidence` · `ocr_text?` · `pedagogical_classification (pedagogical\|decorative\|null)` · `classification_confidence?` |
| **VisualBlock** | `id` · `block_id` · `chapter_id` · `type (diagram\|graph\|table\|figure\|map\|circuit\|photo)` · `image_url` · `thumbnail_url` · `width_px` · `height_px` · `labels[] { text, position {x, y} }` · `axis_labels? { x_label, y_label, x_unit?, y_unit? }` · `table_structure? { rows: int, cols: int, headers[]?, cells[][] }` · `caption?` (légende détectée sous/au-dessus du visuel) · `alt_text` (description textuelle générée par LLM pour accessibilité) · `retention_expires_at` (aligné sur RGPD J+30, cf. Z2-AC12) |
| **Document** | `id` · `chapter_id` · `type` · `tags[]` · `blocks[]` · `source_image_url?` |
| **GeneratedDocument** | `id` · `item_id` · `chapter_id` · `type (table\|graph\|schema\|map\|mindmap)` · `format (markdown\|mermaid\|svg\|ascii)` · `content` (le document rendu : Markdown table, code Mermaid, SVG inline, ou ASCII art) · `title` · `source_item_ids[]` (items dont les données servent de base) · `values_seed?` (seed pour reproduire les valeurs générées) · `created_at` · `session_id?` |
| **Notion** | `id` · `chapter_id` · `name` (libellé lisible généré par le LLM) · `concept_tags[]` · `item_ids[]` · `order (int)` |
| **Item** | `id` · `chapter_id` · `notion_id?` · `revision_id` · `type (KNOWLEDGE\|PROCEDURE\|DOCUMENT\|WRITING)` · `term?` · `keywords[]?` · `steps[]?` · `linked_doc_id?` · `visual_block_ids[]?` · `tags[]` · `confidence` · `validation_required` · `archived` · `fidelity_score?` · `fidelity_flag? (low\|medium\|null)` · `coherence_flag? (contradiction\|orphan_reference\|null)` · `anomaly_flag? (high_failure_rate\|null)` · `llm_model_version?` · `prompt_template_version?` |
| **ValidationTask** | `id` · `item_id` · `crop_url` · `suggestion` · `priority` · `status` · `resolved_by?` · `source (uncertainty_detection\|student_report\|anomaly_detection\|coherence_check\|fidelity_check)` · `student_note?` |
| **Template** | `id (template_id)` · `name` · `version` · `question_type` · `difficulty` · `eligibility{}` · `variables[]` · `prompt_template` · `grading{}` · `uses_visual? (boolean, default false)` · `visual_interaction_type? (label_completion\|describe\|matching\|read_value\|identify_zone\|null)` |
| **Question** | `id` · `template_id` · `item_id` · `visual_block_id?` · `rendered_prompt` · `rendered_visual_url?` (URL du visuel transformé : légendes masquées, zones floutées, etc.) · `expected_answer{}` · `grading_policy` · `clarification? { intent: string, starter_hint: string }` · `llm_model_version?` · `prompt_template_version?` · `times_seen? (default 0)` (incrémenté à chaque présentation en session, utilisé par Z4-AC08 pour la variété et l'anti-monotonie) |
| **Attempt** | `id` · `question_id` · `user_id` · `answer` · `score` · `feedback` · `created_at` · `source (interactive\|paper_report)` (défaut interactive) · `rapid_response? (boolean, default false)` · `response_time_ms?` (mesuré côté client, du rendu de la question au tap « Valider ») · `hint_used? (boolean)` · `clarification_used? (boolean)` · `confidence_level? (1\|2\|3\|null)` (jugement de confiance JOL, cf. Z1-AC25 ; null si non collecté — échantillonnage 1/3) |
| **Mastery** | `id` · `user_id` · `item_id` · `state (UNKNOWN\|FRAGILE\|OK\|SOLID)` · `next_due_at` · `last_review_at` · `last_success_at?` · `consecutive_successes` · `consecutive_failures? (default 0)` (pour Z1-AC19 descente difficulté) · `current_difficulty? (default null)` (override de difficulté par Z1-AC19, null = difficulté template standard) |
| **Session** | `id` · `user_id` · `chapter_ids[]` · `type (daily\|diagnostic\|mock_exam\|evening_first\|pre_class\|paper_report\|consolidation_optional)` · `trigger (manual\|scheduled\|notification)` · `questions[]` · `started_at` · `completed_at?` · `current_question_index (default 0)` · `includes_pre_class? (boolean)` |
| **EveningPlan** | `id` · `user_id` · `date` · `steps[] { type (capture\|evening_first\|daily\|pre_class), subject_label, session_id?, estimated_duration_min, status (pending\|in_progress\|completed\|skipped) }` · `total_estimated_min` · `mode (full\|express)` · `completed_at?` · `completion_rate` · `expires_at` |
| **Notification** | `id` · `user_id` · `type (capture_reminder\|review_reminder\|pre_class\|missed_session_reminder\|parent_schedule_change\|parent_missed_session\|parent_inactivity\|parent_routine_completed\|parent_digest_anticipated\|parent_exam_day\|paper_report_reminder\|pipeline_complete)` · `subject` · `scheduled_at` · `sent_at?` · `clicked_at?` · `source_session_id?` · `linked_student_id?` |
| **ParentNotificationPref** | `id` · `parent_user_id` · `schedule_change_enabled (default true)` · `missed_session_enabled (default true)` · `inactivity_enabled (default true)` · `inactivity_threshold_days (default 3)` · `routine_completed_enabled (default true)` |

> **Entités non persistées (vues calculées) :**
> - **Script 3 minutes** (Z1-AC17, §5.3) : pas d'entité en base. Les 2–3 questions orales sont calculées à la volée lors de la consultation par le parent, en filtrant les items les plus fragiles éligibles (excluant `validation_required`, `anomaly_flag`, `is_demo`). Le résultat n'est pas mis en cache (le script doit refléter l'état courant de la maîtrise).
> - **Carte de leçon** (pipeline étape 8) : vue composée à partir des Items + Notions + Blocks OCR du chapitre. Pas d'entité `LessonCard` persistée.

---

## 17. Algorithmes MVP

### 17.1 Sélection session (politique 70/20/10)

**70 %** items dus (`next_due_at ≤ aujourd'hui`) + items FRAGILE/UNKNOWN prioritaires. **20 %** items OK récemment réussis (consolidation). **10 %** items SOLID ou nouveaux (découverte / anti-oubli long terme).

**Contraintes additionnelles :** HG inclut ≥1 exercice document si items document disponibles · PC remonte les questions `unites` si erreurs récurrentes sur les 3 dernières sessions · max 1 rédaction et max 1 `long_problem` par session · les exercices documentaires utilisent de préférence des **documents générés originaux** (pas les visuels du cahier) pour éviter l'apprentissage par reconnaissance visuelle du document source.

### 17.1b Sessions proactives

> **Principe :** le service n'attend pas qu'un contrôle soit annoncé pour faire réviser. Chaque cours pris génère une boucle de révision immédiate. L'emploi du temps pilote les rappels et les sessions d'anticipation.

**Session soir J0 (`evening_first`)** — Déclenchée le soir même après un upload de cours.
- Durée cible : 5–10 min.
- Sélection : 100% items UNKNOWN du chapitre fraîchement capturé.
- Gabarits limités à difficulté 1 : `FLASH_MCQ`, `DEF_SHORT`, `CLOZE_KEYWORDS`.
- Objectif : première appropriation, ancrage du vocabulaire et des concepts clés.

**Session veille de cours (`pre_class`)** — Déclenchée la veille au soir de chaque cours d'une matière.
- Durée cible : 5 min.
- Scope : **tous les chapitres actifs** de la matière concernée (un contrôle peut couvrir plusieurs leçons).
- Sélection : items FRAGILE/OK les moins récemment révisés (priorité `last_review_at` le plus ancien), puis items UNKNOWN jamais vus.
- Gabarits : difficulté 1–2 uniquement (rappel rapide, pas d'exercice long).
- Justification affichée : « Tu as [Matière] demain — prépare-toi en cas d'interro surprise ».

**Règle de non-doublon :** si une session `daily` est déjà planifiée le même soir, la session `pre_class` fusionne ses items dans la session `daily` (les items pre_class sont ajoutés en priorité dans les 20% consolidation ou les 70% dus).

### 17.1c Session manquée (politique)

> **Principe :** le service ne culpabilise jamais. Un élève qui rate une session ne subit aucune pénalité mécanique. Le système s'adapte en douceur.

**Expiration session :** une session non complétée expire après 72h (TTL standard). Les questions non répondues sont libérées.

**Items dues non révisés — aucune pénalité :**
- Les items dont `next_due_at` est dépassé gardent leur `next_due_at` d'origine. Ils ne régressent PAS.
- Ils deviennent simplement « en retard » (`next_due_at < now`) et sont automatiquement **priorisés** dans la prochaine session composée (ils tombent dans les 70 % « items dus »).
- Le Mastery state n'est jamais modifié par l'inaction. Seule une réponse incorrecte déclenche une régression (cf. Z1-AC05 à Z1-AC07).

**1 rappel le lendemain (unique, non intrusif) :**
- Si une session `evening_first` ou `pre_class` expire sans avoir été commencée (0 questions répondues), **un seul rappel** est envoyé le lendemain matin (heure configurable, défaut 08h00).
- Type de notification : `missed_session_reminder`.
- Message : « Tu avais une révision en attente — on s'y remet ? ».
- **Aucun second rappel.** Si l'élève ignore le rappel, aucune relance supplémentaire n'est envoyée.
- Les sessions `daily` non commencées ne déclenchent PAS de rappel (pour éviter la sur-sollicitation).

**Pas de mécanique de streak :**
- Aucun compteur de jours consécutifs n'est affiché.
- Aucune mécanique de gamification liée à la régularité n'est implémentée en MVP.
- Le système ne culpabilise jamais l'élève pour une absence de révision.

### 17.2 Spaced repetition (règles simples)

**Sans contrôle posé — révision proactive continue :**

> L'absence de date de contrôle ne signifie plus « pas de révision ». Le service maintient un cycle de révision continu avec les intervalles standard ci-dessous, renforcé par les sessions `pre_class` la veille de chaque cours (si emploi du temps saisi).

| État | next_due_at |
|---|---|
| UNKNOWN | J+1 |
| FRAGILE | J+1 |
| OK | J+3 |
| SOLID | J+7 |

**Avec contrôle posé — intervalles proportionnels au temps restant T :**

Les intervalles se compriment proportionnellement au temps restant avant le contrôle (`T = exam_date − now`, en jours). Les contrôles sont typiquement annoncés à +7 jours.

| État | Intervalle si contrôle dans T jours | Minimum |
|---|---|---|
| UNKNOWN | J+1 (incompressible) | 1j |
| FRAGILE | J+1 (incompressible) | 1j |
| OK | J + max(1, ⌊T/3⌋) | 1j |
| SOLID | J + max(2, ⌊T/2⌋) | 2j |

**Cap absolu :** `next_due_at ≤ exam_date − 1 jour`.

### 17.3 Correction

- **KEYWORDS :** min N mots-clés + synonymes pack. Tolérance stemming (Fr). Score partiel si N-1.
- **NUMERIC :** tolérance ± 2 %. Unité obligatoire. Arrondi intermédiaire autorisé à 2 décimales. Faux négatif si unité absente même si valeur correcte.
- **CHECKLIST / RUBRIC :** 4 critères max, score 0/1 par critère. Score partiel affiché.
- **MCQ :** AUTO. Un seul choix correct. Distractor justifié pour les gabarits `piege`.

---

## 18. Exigences non fonctionnelles

### Performance
- Carte de leçon + items < 2 min pour ~10 pages (objectif < 90 s).
- First value (premier exercice) < 5 min après upload.
- Composition session < 1 s (lazy generation + cache).

### Sécurité
- Chiffrement transit (TLS 1.3) et repos (AES-256).
- Isolation des données par `user_id` à toutes les couches.
- Pas d'exposition des photos originales sans auth.

### RGPD / Mineurs
- Consentement parent obligatoire (mineur < 15 ans).
- Export et suppression des données sur demande.
- Minimisation : photos originales ET crops d'image supprimés par défaut 30j après extraction (opt-in conservation). Le texte OCR brut est conservé (pas d'image). Voir AC Z2-AC12 pour les détails de rétention des crops.
- Logs anonymisés pour analytics.

### Explicabilité
- Chaque exercice affiche 2 lignes max expliquant pourquoi il est proposé.
- Carte de maîtrise montre l'historique des tentatives par item.

### Mobile-first
- Upload et lecture confortables sur smartphone.
- Formulaires sans scroll excessif.
- Mode offline partiel : sessions téléchargées en avance si connexion disponible.

### Observabilité
- Métriques par `template_id` : taux réussite, temps médian, taux partial.
- Métriques par tag : zones de confusion.
- Alertes sur dégradation OCR confidence ou pipeline J0 timeout.

#### Observabilité LLM

Les appels LLM sont monitorés en continu (détails dans `docs/llm-strategy.md`) :

- **Choix de modèles** : Sonnet 4.6 pour la structuration et la cohérence (fidélité source, classification nuancée) ; Haiku 4.5 pour le fidelity check et la génération de questions (latence, coût) ; templates sans LLM pour le feedback (Z4-AC09).
- **Métriques** : latence par type d'appel, tokens consommés, coût, fidelity_score moyen, taux d'hallucinations.
- **Alertes** : dégradation fidelity (`avg < 0.6` sur 24h), hausse hallucinations (`rate(score < 0.5) > 15%`), latence p95 > seuils, changement de modèle silencieux.
- **Versioning** : chaque résultat LLM est tagué avec `llm_model_version` + `prompt_template_version` (AC Z2-AC13). Job hebdomadaire de comparaison qualité entre versions ; alerte admin si dégradation > 15%.
- **Optimisations coût** : Batch API (-50%) pour le pipeline J0, prompt caching (-90% sur system prompts répétitifs).

---

## 19. KPIs (MVP)

| Catégorie | KPI | Cible MVP | Source |
|---|---|---|---|
| Activation | Upload → diagnostic complété (J0) | > 70% | pipeline logs |
| Activation | Temps first value (premier exercice) | < 5 min | pipeline logs |
| Engagement | Sessions / semaine / élève actif | > 3 | session DB |
| Engagement | Taux complétion plan quotidien | > 60% | session DB |
| Apprentissage | Transitions UNKNOWN/FRAGILE → OK/SOLID | > 40% à J+7 | mastery DB |
| Apprentissage | Score contrôle blanc J-3 → J-1 | progression > 0 | mock exam |
| Qualité tags | Taux auto-tagging correct (chapitres pilotes) | > 85% | admin validation |
| Qualité templates | Taux réussite moyen par template | 30–70% (zone apprentissage) | attempt DB |
| Parents passif | Taux ouverture digest hebdo | > 40% | email analytics |
| Parents passif | Taux opt-out alertes | < 5% | settings DB |
| Proactivité | Taux saisie cours le soir même (après notif) | > 50% | notification DB |
| Proactivité | Taux complétion session soir J0 | > 60% | session DB |
| Proactivité | Taux complétion session pre_class | > 40% | session DB |
| Proactivité | Élèves avec emploi du temps saisi | > 70% | schedule DB |
| OCR | Confidence moyenne blocs TEXT | > 0.85 | OCR service |

---

## 20. Questions §16 tranchées

| Question | Décision | Justification |
|---|---|---|
| **Q1 — Date contrôle obligatoire ?** | **Strongly nudged, pas obligatoire.** | La rendre obligatoire crée de la friction à l'onboarding. Le nudge est répété à J+3 et J+7. En l'absence de date, la révision proactive démarre quand même dès J0 soir, avec sessions `pre_class` la veille de chaque cours. L'absence de date ne bloque plus rien. |
| **Q2 — Suppression photos après extraction ?** | **Suppression automatique à J+30 par défaut, opt-in conservation.** | Réduit l'exposition RGPD sur les données de mineurs. La conservation opt-in est utile pour re-segmentation ou debug. Communiqué clairement à l'onboarding. |
| **Q3 — Niveau collège ciblé en premier ?** | **4e en priorité.** | La 4e couvre les chapitres pilotes HG (inégalités, féodale) et PC (masse-volume) dans les programmes officiels. Les rubriques de rédaction sont calibrées au niveau 4e. Extension 3e et 5e post-MVP en ajustant les paramètres de rubrique par pack. |
| **Q4 — Gestion des synonymes ?** | **Pack + admin uniquement en MVP. Pas d'apprentissage progressif.** | L'apprentissage progressif des synonymes introduit un risque de dérive qualité non supervisée. En MVP, l'admin peut enrichir les synonymes par pack après analyse des tentatives. Une roadmap post-MVP inclura la suggestion de synonymes à valider par l'admin. |
| **Q5 — Emploi du temps : saisie manuelle ou import ENT ?** | **Saisie manuelle uniquement en MVP.** | L'import ENT/Pronote est exclu du MVP car chaque établissement utilise un système différent (Pronote, EcoleDirecte, ENT académique…). La saisie manuelle (matière + jour + matin/après-midi) prend < 3 min et suffit pour déclencher les notifications. Import automatique en roadmap post-MVP. |
| **Q6 — Un contrôle peut-il couvrir plusieurs chapitres ?** | **Oui, via l'entité Exam.** | Un Exam référence N chapitres. Le contrôle blanc couvre tous les chapitres liés. Cela reflète la réalité des contrôles de fin de séquence au collège. |

---

## 21. Risques & mitigations

| Risque | Probabilité | Impact | Mitigation |
|---|---|---|---|
| OCR faible sur manuscrit dense | Haute | Élevé | HITL + retry/fallback (§13) + tags/doc types robustes |
| Pipeline J0 > 2 min sur 30 pages | Moyenne | Élevé | Parallélisation + streaming + cache OCR permanent |
| Coût LLM élevé (génération questions) | Haute | Moyen | Lazy generation + cache pool 24h |
| UX J0 trop longue → abandon | Haute | Élevé | First value < 5 min, streaming carte, QCM flash immédiat |
| Parent incompréhension validation HITL | Moyenne | Moyen | UX simplifiée ≤3 clics, contexte visuel, max 3 validations/semaine |
| Faux négatifs correction numérique | Moyenne | Élevé | Tolérance ±2%, arrondi intermédiaire autorisé, unité feedback explicite |
| Complexité 3 matières / 4 chapitres | Faible | Moyen | 1 moteur + packs ; bibliothèque gabarits 100% générique |
| Fatigue notification / sur-sollicitation | Moyenne | Moyen | Max 2 notifs/soir, désactivation possible, fusion session pre_class+daily |
| Emploi du temps obsolète / non saisi | Moyenne | Moyen | Mode dégradé sans emploi du temps (sessions standards). Nudge si emploi du temps absent à J+3. Rappel début de trimestre pour mettre à jour. |
| Exam multi-chapitres : explosion combinatoire items | Faible | Moyen | Contrôle blanc limité à 30 min, sélection représentative par chapitre (proportionnelle au nb d'items) |
| RGPD mineurs / photos sensibles | Faible | Très élevé | Suppression J+30 par défaut, consent parental, chiffrement repos |
| Session manquée / décrochage silencieux | Moyenne | Moyen | 1 rappel unique le lendemain matin (pas de harcèlement). Items dues re-priorisés automatiquement. Aucune pénalité mastery. KPI « taux complétion » pour détecter les décrochages à l'échelle. |
| Items `fidelity_score = null` non re-vérifiés | Moyenne | Élevé | Job quotidien de re-vérification fidelity (AC Z3-AC10 définit le fidelity check ; le job de re-tentative et le timeout 3 jours sont des détails d'implémentation au-delà du scope AC). Item restreint aux templates simples après 3 jours de timeout persistant. Empêche les hallucinations LLM de rester indéfiniment en usage sans vérification. |
| ValidationTasks admin non résolues (accumulation silencieuse) | Moyenne | Élevé | SLA 7 jours avec escalade priorité CRITICAL. Restriction QCM-only à J+14 si non résolu. KPI « % résolution < 7j » trackée en dashboard admin. |
| Confusion UX "Ignorer" vs "Je ne sais pas" (perte d'items utiles) | Moyenne | Moyen | Sous-textes explicatifs permanents sur chaque action de validation (AC Z3-AC04/AC05). Mention explicite de l'impact de "Ignorer" sur le flux pédagogique. Récupération possible (AC Z3-AC16). |
| Perte silencieuse Mastery sur re-upload (variations OCR sigles) | Haute | Élevé | Normalisation étendue insensible ponctuation/points/tirets (AC Z5-AC07). Configurable par pack pour cas sémantiques. Couvre I.D.H./IDH, P.I.B./PIB, etc. |
| Exams simultanés même journée (préparation déséquilibrée) | Faible | Moyen | Avertissement non-bloquant à la création + planification alternée des matières en interleaving (AC Z6-AC32). Mock exams restent séparés. |
| **Fausse maîtrise SOLID sur items non validés (mastery inflation)** | **Haute** | **Très élevé** | Plafond maîtrise OK tant que `validation_required = true` (AC Z1-AC13). L'item ne peut pas atteindre SOLID avec uniquement des QCM simples. Badge « maîtrise partielle » visible. Empêche la pollution du signal mastery pour l'élève ET le parent. |
| Absence de feedback positif → churn élève | Haute | Élevé | Micro-célébrations sur transitions positives (AC Z1-AC14). Débrief de fin de session avec progrès et prochain objectif (AC Z1-AC15). Pas de streak mais valorisation de la progression réelle. |
| Retour après absence = session écrasante → décrochage définitif | Haute | Élevé | Mode « retour en douceur » : session courte (5 min), gabarits faciles, étalement de la dette sur 3–5 jours (AC Z6-AC22). Se désactive après 2 sessions consécutives. |
| Diagnostic initial trop dur → mauvaise première impression | Moyenne | Élevé | Rampe de difficulté progressive (AC Z6-AC24). Les 2–3 premières questions sont faciles. Redescente automatique après 3 échecs consécutifs. Message de clôture toujours positif. |
| Items post-exam saturent les sessions → fatigue | Moyenne | Moyen | Archivage automatique post-exam (AC Z6-AC23). Items SOLID passent en maintenance longue (J+14). Chapitres post-exam restent actifs mais non prioritaires. |
| Alert fatigue parent sur sessions manquées → désactivation totale | Moyenne | Moyen | Max 1 push « session manquée » / semaine (AC Z6-AC26). Sessions manquées suivantes résumées dans digest hebdo. Détail complet dans tableau de bord. |
| Digest parent mal timé par rapport aux exams | Moyenne | Moyen | Digest supplémentaire « pré-contrôle » à J-3 avant chaque exam (AC Z6-AC25). Inclut maîtrise par chapitre + items fragiles + recommandation d'action. |
| Score contrôle blanc gonflé par templates simplifiés | Haute | Élevé | Score accompagné d'un `score_confidence` + caveat explicite si items non validés (AC Z1-AC16). Parent voit la proportion de questions complètes vs simplifiées. |
| Parent non informé de capture incomplète (pages OCR échouées) | Haute | Élevé | Digest inclut signalement des pages échouées par chapitre (AC Z6-AC28). Alerte si > 30% du contenu manque. |
| Crops d'image persistent au-delà de la suppression J+30 des photos | Moyenne | Très élevé | Rétention crops alignée sur photos originales (AC Z2-AC12). Suppression crops + source_image_url à J+30. RGPD conforme. |
| Labels maîtrise incompréhensibles pour les parents | Moyenne | Moyen | Traduction UNKNOWN/FRAGILE/OK/SOLID en langage parent (AC Z6-AC29). % maîtrise calculé sur OK+SOLID uniquement. |
| Items sous investigation proposés dans script 3 minutes | Moyenne | Élevé | Script exclut items avec `validation_required` ou `anomaly_flag` (AC Z1-AC17). Évite que le parent teste l'enfant sur du contenu potentiellement faux. |
| Résolution admin sans feedback → boîte noire | Faible | Moyen | Notification in-app élève + mention dans digest parent après résolution admin. Ferme la boucle de feedback. |
| Sessions monotones (5 MCQ consécutifs) → ennui | Haute | Moyen | Variété de gabarits imposée : max 2 consécutifs du même template, alternance round-robin des question_type (AC Z4-AC08). |
| Feedback « Faux » sans explication → pas d'apprentissage | Haute | Élevé | Feedback structuré obligatoire : réponse correcte + ce qui manquait + indice 1 phrase. Templaté, pas LLM live (AC Z4-AC09). |
| Élève bloqué sans pouvoir passer → frustration → fermeture app | Moyenne | Élevé | Bouton « Passer » sans pénalité mastery, max 2/session, question remise en fin de session (AC Z4-AC10). |
| Petit chapitre (3 items) = session non viable | Moyenne | Moyen | Session minimum 4 questions, reformulation avec distractors/ordres différents, durée adaptée 3-5 min (AC Z4-AC11). |
| **Validation erronée verrouillée en base (parent confirme un item OCR faux)** | **Moyenne** | **Très élevé** | Rétractation possible via signalement élève ou détection anomalie (AC Z3-AC14). Confidence ramenée à ≤0.7, item re-restreint aux templates simples, admin notifié priorité HIGH. Non-rétroactivité mastery (pragmatisme). |
| **Clicking aveugle MCQ → progression artificielle** | **Haute** | **Élevé** | Réponse < 2s marquée `rapid_response`, progression bloquée si correcte, régression maintenue si incorrecte (AC Z3-AC15). Message non-bloquant après 3 réponses rapides/session. Seuil 2s calibré sur temps lecture minimum MCQ. |
| **Drift LLM silencieux → dégradation qualité items sans diagnostic** | **Moyenne** | **Élevé** | Versioning modèle + prompt sur chaque résultat LLM (AC Z2-AC13). Job hebdomadaire de comparaison qualité entre versions. Alerte admin si dégradation > 15%. Traçabilité complète pour rollback informé. |
| **Pas de vision globale progression → l'élève ne perçoit pas ses progrès** | **Haute** | **Élevé** | % maîtrise global cross-chapitres sur dashboard (AC Z6-AC30). Barre de progression par chapitre + message d'encouragement contextuel basé sur tendance hebdo. |
| **Vieux chapitres encombrent le dashboard et les sessions** | **Moyenne** | **Moyen** | Archivage (soft delete) de chapitre à la demande de l'élève. Données de maîtrise conservées, sessions recalculées, réactivation possible. Parent informé dans digest. |
| **Perte de connexion mid-session → réponses perdues → frustration** | **Haute** | **Très élevé** | Persistance optimiste côté client avec sync FIFO (AC Z6-AC31). Feedback correction instantané (calcul local MCQ/NUMERIC). Mastery mis à jour uniquement après sync confirmée. Retry exponentiel (5s/15s/45s). |
| **Date d'exam modifiée → intervalles de révision incohérents** | **Haute** | **Élevé** | Recalcul automatique des `next_due_at` compressés sur modification de `exam_date` (AC Z1-AC18). Message contextuel adapté (avancé vs repoussé). Parent informé dans digest. |
| **Mêmes questions vues en boucle → apprentissage de surface par reconnaissance** | **Haute** | **Élevé** | Variété de gabarits et anti-monotonie en session (AC Z4-AC08). Compteur `times_seen` pour prioriser les questions les moins vues. Renouvellement lazy non-bloquant. |
| **Cycle d'échec répété sur un item → frustration → churn** | **Haute** | **Très élevé** | Descente automatique de difficulté après 3 échecs consécutifs (AC Z1-AC19). Indice avec renvoi vers carte de leçon. Pause pédagogique J+2 après 5 échecs. |
| **LLM indisponible = rien à faire dans l'app → fermeture immédiate** | **Moyenne** | **Très élevé** | Mode « Relecture active » avec flashcards textuelles statiques (AC Z4-AC12). Données OCR en base, aucun LLM requis. Auto-évaluation sans impact mastery. |
| **Items « Ignoré » en HITL = trou permanent dans la couverture de révision** | **Haute** | **Élevé** | Section « Points non vérifiés » dans carte de leçon avec Réactiver (AC Z3-AC16). Récupération autonome sans admin. |
| **Blocage OK→SOLID sans explication → élève ne comprend pas** | **Haute** | **Élevé** | Message explicatif positif sur repos cognitif (AC Z1-AC04 enrichi). Affiché 1x/session max. |
| **Items UNKNOWN jamais présentés en session (starvation par file FRAGILE)** | **Haute** | **Élevé** | Minimum 1 item UNKNOWN/session si starvation > 7 jours (AC Z1-AC20). Alerte admin si 0 tentatives depuis > 14 jours. Indicateur « N points pas encore abordés » sur dashboard élève. |
| **Mismatch template/type d'item → échecs injustes sur items restreints** | **Moyenne** | **Élevé** | Garde-fou template adaptant la formulation au type d'item PROCEDURE (AC Z3-AC17). Log `template_type_mismatch` pour suivi admin. Template MCQ formule préféré aux templates KNOWLEDGE. |
| **Création d'exam pas dans les triggers d'invalidation cache → priorités stale** | **Haute** | **Élevé** | Invalidation `question_candidates` sur CRUD Exam (AC Z4-AC03). Recalcul next_due_at sur suppression exam (annulation compression). Log `EXAM_CACHE_INVALIDATION`. |
| **Élève diligent = rien à faire dans l'app → perte d'habitude** | **Moyenne** | **Élevé** | Message positif « à jour » + session consolidation optionnelle sans risque de régression (AC Z4-AC13). |
| **Items SOLID perdus silencieusement sur re-upload (page manquante)** | **Haute** | **Très élevé** | Alerte explicite listant les items OK/SOLID non retrouvés dans R2 (AC Z5-AC08). Option re-upload pages manquantes. Délai 14 jours avant archivage définitif. |
| **Chapter.exam_id singulier → un seul exam par chapitre, resserrement cassé** | **Haute** | **Élevé** | Modèle corrigé : `Chapter.exam_ids[]` (pluriel). Resserrement sur exam actif le plus proche, bascule automatique après exam passé (AC Z6-AC32). |

---

## 22. Machines à états formelles

Ces machines à états formalisent les transitions de statut des entités principales. Elles doivent être respectées dès le Lot 0 (§3.2).

### Session

```
COMPOSING → IN_PROGRESS → COMPLETED
                        → EXPIRED (TTL 72h, 0 réponses)
                        → ABANDONED (TTL 72h, ≥1 réponse)
```
- `COMPOSING` : les questions sont en cours de sélection/génération
- `IN_PROGRESS` : l'élève répond aux questions
- `COMPLETED` : toutes les questions répondues
- Transitions interdites : `COMPLETED → IN_PROGRESS`, `EXPIRED → *`, `ABANDONED → *`

### Page (pipeline J0)

```
UPLOADING → OCR_PENDING → OCR_PROCESSING → PROCESSED → ITEMS_GENERATING → DONE
                                         → FAILED (ocr_timeout | ocr_error)
                                                                          → NO_ITEMS
                                                                          → ITEMS_FAILED
```
- Chaque état est terminal ou a exactement 1-2 successeurs
- Un échec sur une page ne bloque pas les autres (Z2-AC02)

### ChapterRevision

```
PROCESSING → READY (≥1 item valide)
           → PARTIAL (certaines pages en échec, mais ≥1 item)
           → FAILED (0 items valides, Z2-AC07)
```

### ValidationTask

```
PENDING → CONFIRMED (Z3-AC02)
        → CORRECTED (Z3-AC03)
        → UNKNOWN_ANSWER (Z3-AC04 « je ne sais pas »)
        → IGNORED (Z3-AC05)
```

### Mastery (rappel — documenté dans les ACs Z1)

```
UNKNOWN → FRAGILE → OK → SOLID
                  ← OK ← SOLID  (régressions)
          FRAGILE ← OK          (régression)
```
- Pas de retour à UNKNOWN depuis FRAGILE (Z1-AC07)
- SOLID → OK directement, pas FRAGILE (Z1-AC05)

---

---

*Fin du document — Révise Mieux PRD*
