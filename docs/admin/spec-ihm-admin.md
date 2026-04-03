# Specification fonctionnelle — IHM Admin (Backoffice)

> Date : 2026-04-04
> Issue : #26
> Statut : v1

---

## 1. Vue d'ensemble

L'IHM admin est un backoffice web destine a l'**admin produit** (PRD §4). Il ne concerne ni les eleves ni les parents. Il couvre 5 domaines fonctionnels.

### Persona

**Admin produit** — gere packs, lexiques, gabarits. Surveille la qualite de generation LLM. Analyse les metriques par template et par tag.

### Principes

- Interface **dense et fonctionnelle** — pas besoin d'etre jolie, doit etre efficace
- **Lecture-first** — 80% du temps est de la consultation, 20% de l'action
- **Pas de logique metier dans l'admin** — appelle les memes endpoints API que le mobile
- **Autonome** — tourne dans un dossier `admin/` independant du mobile

### Stack recommandee

| Choix | Justification |
|-------|---------------|
| **React + Vite** | Leger, rapide a bootstrapper, pas besoin de SSR |
| **Tailwind CSS** | Composants utilitaires, pas de design system complexe |
| **TanStack Query** | Cache serveur, refetch automatique |
| **Recharts** | Graphiques (courbes, barres, jauge) |
| **TypeScript strict** | Coherence avec le mobile |

---

## 2. Ecrans et fonctionnalites

### 2.1 Dashboard principal

**Route** : `/`
**Intention** : vue synthetique de l'etat du systeme en un coup d'oeil.

| Bloc | Donnees | Source API |
|------|---------|-----------|
| Utilisateurs actifs | Nombre eleves + parents | `GET /api/v1/admin/stats` (a creer) |
| Chapitres | Total + par matiere | `GET /api/v1/chapters` (aggrege) |
| Pipeline | En cours / succes / echec (24h) | `GET /api/v1/admin/pipeline-stats` (a creer) |
| LLM | Appels/jour, cout cumule, taux erreur | `GET /api/v1/admin/llm-stats` (a creer) |
| Validations | En attente / resolues | `GET /api/v1/validations` |
| Alertes | Dernieres alertes critiques | `GET /api/v1/admin/alerts` (a creer) |

**Etats** : loading (skeleton), empty (premier lancement), error (backend down).

---

### 2.2 Dashboard qualite LLM

**Route** : `/llm`
**Intention** : detecter les degradations de qualite et piloter les couts.
**Source** : `docs/llm-strategy.md` §6.3

#### Vue temps reel

| Metrique | Visualisation | Seuil d'alerte |
|----------|--------------|----------------|
| Appels/min | Compteur live | — |
| Latence p50/p95 | Jauge + sparkline | p95 > 30s (structuration), p95 > 2s (questions) |
| Taux d'erreur | Pourcentage + couleur | > 5% = rouge |
| Cout cumule du jour | Compteur $ | > budget_daily × 1.5 |

#### Vue qualite (courbes)

| Courbe | Axe X | Axe Y | Segmentation |
|--------|-------|-------|-------------|
| Fidelity score moyen | Jour | avg(fidelity_score) | Par `llm_model_version` |
| Taux hallucination | Jour | % items fidelity < 0.5 | Par modele |
| Items generes/page | Jour | avg(items_count) | Par matiere |

**Alerte visuelle** : banniere rouge si `avg(fidelity_score) < 0.6` sur 24h.

#### Vue couts

| Donnee | Visualisation |
|--------|--------------|
| Cout par type d'appel | Barres empilees (structuration, fidelity, questions, OCR) |
| Tendance hebdo | Courbe |
| Projection mensuelle | Nombre + tendance |

#### Vue drift

Tableau comparatif entre la version de prompt actuelle et la precedente :

| Metrique | Prompt v1.0.0 | Prompt v1.1.0 | Delta | Verdict |
|----------|---------------|---------------|-------|---------|
| avg(fidelity) | 0.82 | 0.78 | -4.8% | WARNING |
| avg(confidence) | 0.85 | 0.87 | +2.4% | OK |
| items/page | 7.2 | 6.8 | -5.6% | WARNING |

**Source API** : `GET /api/v1/admin/llm-stats?period=7d` (a creer, lit la table `llm_call_logs`).

---

### 2.3 Gestion des ValidationTasks (HITL)

**Route** : `/validations`
**Intention** : resoudre les items a faible confiance.
**Source** : `docs/ac/Z3.md`

#### Liste des tasks

| Colonne | Donnee |
|---------|--------|
| Item | Terme + type (KNOWLEDGE/PROCEDURE/...) |
| Chapitre | Nom du chapitre |
| Confiance | Score 0-1 + barre de couleur |
| Source | `uncertainty` / `fidelity_low` / `contradiction` |
| Priorite | 1-10 |
| Date | Date de creation |

**Tri** : par priorite decroissante, puis par date.
**Filtre** : par chapitre, par source, par plage de confiance.

**API** : `GET /api/v1/validations` (existe).

#### Interface de resolution

Quand l'admin clique sur une task :

| Action | Bouton | Effet |
|--------|--------|-------|
| **Confirmer** | Vert "Confirmer" | `POST /api/v1/validations/:id/resolve {action: "confirm"}` — item.confidence = max(0.85, current), item.validation_required = false |
| **Corriger** | Orange "Corriger" | Formulaire d'edition inline (term, keywords, steps) + `POST resolve {action: "correct", corrected_data: {...}}` |
| **Ignorer** | Gris "Ignorer" | `POST resolve {action: "ignore"}` — item supprime du flux de validation, conserve tel quel |
| **Je ne sais pas** | Gris "NSP" | `POST resolve {action: "skip"}` — task reste pending, reapparait plus tard |

#### Stats

| Metrique | Visualisation |
|----------|--------------|
| Tasks en attente | Compteur |
| Resolues / jour | Courbe |
| Taux par action | Camembert (confirm/correct/ignore/skip) |
| Temps moyen de resolution | Nombre |

---

### 2.4 Gestion des packs et gabarits

**Route** : `/packs`
**Intention** : CRUD packs matiere, templates de questions, lexiques.
**Source** : PRD §7-9, Epic 10

#### Packs matiere

| Champ | Type | Exemple |
|-------|------|---------|
| ID | String | `HG-INEG` |
| Matiere | Enum | `Histoire-Geographie` |
| Niveau | String | `4e` |
| Templates actives | Multi-select | Liste des template_id |
| Lexique tags | JSON | `{unites: [...], dates: [...]}` |

**CRUD** : liste paginee + formulaire creation/edition.
**API** : `GET/POST/PUT/DELETE /api/v1/admin/packs` (a creer).

#### Gabarits de questions (templates)

27 templates MVP (PRD §8). Chaque template a :

| Champ | Type | Description |
|-------|------|-------------|
| ID | String | `GEN.KNOW.FLASH_MCQ` |
| Nom | String | `Flash MCQ` |
| Type question | Enum | `MCQ`, `SHORT_ANSWER`, `CLOZE`, `NUMERIC`, `LABEL` |
| Difficulte | 1-3 | Niveau de difficulte |
| Prompt template | Text | Template Mustache/Go `{{term}}` |
| Eligibilite | JSON | Regles d'eligibilite (item_type, min_confidence) |
| Grading | JSON | Regles de scoring |

**CRUD** : tableau avec edition inline.
**API** : `GET/POST/PUT /api/v1/admin/templates` (a creer — la table `templates` existe deja).

#### Analytics par template

| Metrique | Par template_id |
|----------|----------------|
| Taux de reussite | % score >= 0.7 |
| Score moyen | avg(score) |
| Temps moyen | avg(duration_ms) |
| Partials | % 0.3 < score < 0.7 |

**API** : `GET /api/v1/admin/template-analytics` (a creer, aggrege depuis `attempts`).

---

### 2.5 Monitoring pipeline J0

**Route** : `/pipeline`
**Intention** : voir l'etat des pipelines en cours et diagnostiquer les echecs.

#### Pipelines actifs

| Colonne | Donnee |
|---------|--------|
| Chapitre | Nom |
| Statut | PROCESSING / READY / PARTIAL / FAILED |
| Pages | total / traitees / echouees |
| Items generes | Nombre |
| Debut | Timestamp |
| Duree | Temps ecoule |

**API** : `GET /api/v1/admin/pipelines` (a creer — aggrege depuis `chapter_revisions` + `pages`).

#### Metriques globales

| Metrique | Visualisation |
|----------|--------------|
| Taux de succes OCR | % pages traitees avec succes / total |
| Items/chapitre moyen | Nombre |
| Temps moyen pipeline | Duree |
| Echecs recents | Liste des 10 derniers echecs avec message d'erreur |

#### Alertes pipeline

| Alerte | Condition | Severite |
|--------|-----------|----------|
| Pipeline bloque | Duree > 10 min | WARNING |
| Echec complet | 0 items generes | CRITICAL |
| Taux echec eleve | > 20% pages en echec sur 24h | WARNING |
| Hallucination detectee | fidelity_score < 0.3 | CRITICAL |

---

### 2.6 Gestion utilisateurs

**Route** : `/users`
**Intention** : voir les utilisateurs et leurs liaisons parent-enfant.
**Scope Lot 0** : usage local pere-fils, donc tres simple.

| Colonne | Donnee |
|---------|--------|
| Nom | display_name |
| Role | student / parent |
| Chapitres | Nombre de chapitres |
| Sessions | Nombre de sessions completees |
| Derniere activite | Timestamp |
| Lie a | Parent/enfant lie (si applicable) |

**Actions** :
- Voir le detail d'un utilisateur (chapitres, masteries, sessions)
- Lier parent-enfant
- Desactiver un compte

**API** : `GET /api/v1/admin/users` (a creer).

---

## 3. Endpoints API admin a creer

| Endpoint | Methode | Source de donnees | Priorite |
|----------|---------|-------------------|----------|
| `/api/v1/admin/stats` | GET | Aggrege users, chapters, sessions | P1 |
| `/api/v1/admin/llm-stats` | GET | Table `llm_call_logs` | P1 |
| `/api/v1/admin/pipeline-stats` | GET | Tables `chapter_revisions`, `pages` | P1 |
| `/api/v1/admin/alerts` | GET | Alertes recentes (metriques LLM + pipeline) | P2 |
| `/api/v1/admin/packs` | CRUD | Table `packs` (a creer) | P2 |
| `/api/v1/admin/templates` | GET/PUT | Table `templates` (existe) | P1 |
| `/api/v1/admin/template-analytics` | GET | Aggrege `attempts` par template_id | P2 |
| `/api/v1/admin/pipelines` | GET | Tables `chapter_revisions`, `pages`, `items` | P1 |
| `/api/v1/admin/users` | GET | Table `users` | P2 |

**Securite** : tous les endpoints admin requierent un JWT avec `role=admin`. Middleware `adminOnly` dans le router.

---

## 4. Navigation

```
┌─────────────────────────────────────┐
│  Dashboard  │ LLM │ HITL │ Packs │ Pipeline │ Users  │
└─────────────────────────────────────┘
                    │
                    ▼
           Contenu de l'ecran selectionne
```

Sidebar ou tabs horizontaux. Navigation simple, pas de nesting profond.

---

## 5. Priorites d'implementation

### Phase 1 — MVP admin (1 semaine)

- [ ] Setup React + Vite + Tailwind + TanStack Query dans `admin/`
- [ ] Dashboard principal (stats basiques)
- [ ] Interface HITL (liste + resolution) — reutilise les endpoints existants
- [ ] Monitoring pipeline (liste des revisions)

### Phase 2 — Qualite LLM (1 semaine)

- [ ] Dashboard LLM (metriques, courbes, couts)
- [ ] Endpoints `admin/llm-stats`
- [ ] Alertes visuelles

### Phase 3 — Gestion (1 semaine)

- [ ] CRUD packs et templates
- [ ] Analytics par template
- [ ] Gestion utilisateurs
- [ ] Endpoints admin restants

---

## 6. Ecrans — Wireframes textuels

### Dashboard principal

```
┌──────────────────────────────────────────────────────┐
│  Revise Mieux — Admin                    [Deconnexion]│
├──────────────────────────────────────────────────────┤
│                                                       │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐   │
│  │ 12      │ │ 47      │ │ 3/0/0   │ │ 5       │   │
│  │ Eleves  │ │ Chapitr.│ │ Pipeline│ │ A valid.│   │
│  └─────────┘ └─────────┘ └─────────┘ └─────────┘   │
│                                                       │
│  LLM aujourd'hui          │  Alertes recentes        │
│  ├─ 234 appels             │  ├─ [WARN] p95 lat. 4.2s│
│  ├─ $1.23 cout             │  ├─ [INFO] new model v2 │
│  ├─ 0.3% erreurs           │  └─ (aucune critique)   │
│  └─ fidelity moy: 0.84    │                          │
│                                                       │
│  Derniers pipelines                                   │
│  ├─ SVT Photosynthese    READY   8 items   42s       │
│  ├─ PC Densite           READY   12 items  38s       │
│  └─ HG Inegalites        PROCESS 3/5 pages           │
└──────────────────────────────────────────────────────┘
```

### Interface HITL

```
┌──────────────────────────────────────────────────────┐
│  Validations en attente (5)         [Filtre: Tous ▼] │
├──────────────────────────────────────────────────────┤
│  │ Terme              │ Type    │ Conf. │ Source   │  │
│  ├────────────────────┼─────────┼───────┼──────────┤  │
│  │ Masse volumique    │ KNOW    │ 0.42  │ fidelity │  │
│  │ rho = m/V          │ PROC    │ 0.38  │ uncertai │  │
│  │ Photosynthese def  │ KNOW    │ 0.51  │ fidelity │  │
│  │ Schema cellule     │ DOC     │ 0.29  │ uncertai │  │
│  │ Societe feodale    │ KNOW    │ 0.55  │ fidelity │  │
│  └────────────────────┴─────────┴───────┴──────────┘  │
│                                                       │
│  Detail: "Masse volumique"                            │
│  ┌─────────────────────────────────────────────┐      │
│  │ Terme: Masse volumique                       │      │
│  │ Type: KNOWLEDGE                              │      │
│  │ Confidence: 0.42                             │      │
│  │ Keywords: masse, volume, densite, rho        │      │
│  │ Source OCR: "La masse volumique est le..."   │      │
│  │                                              │      │
│  │ [Confirmer]  [Corriger]  [Ignorer]  [NSP]   │      │
│  └─────────────────────────────────────────────┘      │
└──────────────────────────────────────────────────────┘
```

---

## 7. References

- PRD §4 (Admin produit), §7-9 (Packs, gabarits), Epic 10
- `docs/llm-strategy.md` §6.3 (monitoring, alertes, dashboard)
- `docs/ac/Z3.md` (validation HITL)
- `backend/internal/http/router.go` (routes existantes)
- `backend/internal/infra/postgres/` (repositories existants)
- Issue #19 (frontend supprime), Issue #26
