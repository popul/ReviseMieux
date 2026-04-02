# Parcours utilisateur — Premiere connexion

> Couvre Z8-AC01, Z8-AC02, Z8-AC04, Z8-AC08, Z6-AC04, Z6-AC05, Z4-AC09
>
> Objectif : l'eleve fait son premier quiz en moins de 3 minutes apres l'installation, sans cahier ni photo.

---

## Vue d'ensemble

```
Installation
    │
    ▼
┌─────────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Dashboard  │────▶│ Session demo │────▶│   Debrief    │────▶│  Dashboard   │
│  (vide +    │     │ (8 questions │     │  (score +    │     │  (chapitre   │
│  seed demo) │     │  ~3 min)     │     │  transitions)│     │  demo actif) │
└─────────────┘     └──────────────┘     └──────────────┘     └──────────────┘
                                                                     │
                                              ┌──────────────────────┘
                                              ▼
                                    ┌──────────────────┐
                                    │  Capture premier │
                                    │  vrai cours      │
                                    │  (quand il veut) │
                                    └──────────────────┘
```

---

## Etape 1 — Dashboard vide + chapitre demo

**Ecran :** `app/(tabs)/index.tsx`
**Declencheur :** premiere ouverture de l'app apres creation de compte
**ACs :** Z8-AC01, Z8-AC02

### Ce qui se passe

1. L'app demarre, le `_layout.tsx` recupere automatiquement un token d'authentification.
2. Le dashboard charge `GET /api/v1/chapters` → **liste vide** (aucun chapitre).
3. L'ecran affiche un **empty state** avec :
   - Icone 🧪 et titre « Aucun chapitre »
   - Texte « Essaie le chapitre demo pour decouvrir l'app ! »
   - Bouton **« Charger le chapitre demo »**

### Action utilisateur

L'eleve tape sur **« Charger le chapitre demo »**.

### Ce qui se passe cote API

1. `POST /api/v1/onboarding/seed-demo`
   - Cree le user en DB si necessaire
   - Cree le chapitre « Densite et masse volumique (demo) » avec `is_demo = true`
   - Cree 3 notions : *Masse volumique (ρ)*, *Densite et flottabilite*, *Conversions et applications*
   - Cree 8 items (6 KNOWLEDGE + 2 PROCEDURE) lies aux notions
   - Cree 8 masteries en etat UNKNOWN
   - Retourne `{ chapter_id, item_count: 8, message }`

2. `GET /api/v1/chapters` (refresh)
   - Retourne le chapitre demo avec `item_count: 8`, `mastery_breakdown: { unknown: 8 }`

### Resultat a l'ecran

Le dashboard affiche maintenant le chapitre demo :
- Badge 🧪 + « Physique-Chimie »
- Titre « Densite et masse volumique (demo) »
- Barre de mastery : 100% unknown (gris)
- « 8 items »
- Bouton **« Essayer »**

---

## Etape 2 — Detail du chapitre demo

**Ecran :** `app/chapter/[id].tsx`
**Declencheur :** tape sur le chapitre dans le dashboard
**ACs :** Z7-AC15, Z7-AC16

### Ce qui se passe

1. `GET /api/v1/chapters/{id}/lesson-card` → retourne chapitre + 8 items + 3 notions
2. `GET /api/v1/masteries` → retourne les 8 masteries (toutes UNKNOWN)
3. L'ecran affiche :
   - En-tete : « Physique-Chimie · Densite et masse volumique (demo) »
   - Barre de mastery globale (100% unknown)
   - 3 notions en accordeon, chacune avec ses items et badges mastery

### Etat des notions

| Notion | Items | Etat initial |
|--------|-------|-------------|
| Masse volumique (ρ) | 3 items (def, unite SI, formule ρ=m/V) | 3 UNKNOWN |
| Densite et flottabilite | 3 items (densite, flotte si <1, eau 1000 kg/m³) | 3 UNKNOWN |
| Conversions et applications | 2 items (g/cm³→kg/m³, identification materiau) | 2 UNKNOWN |

### Action utilisateur

L'eleve tape sur **« 🎯 Lancer une session »**.

---

## Etape 3 — Session de revision (evening_first)

**Ecran :** `app/session/[id].tsx`
**Declencheur :** bouton « Lancer une session » depuis le detail chapitre
**ACs :** Z6-AC04, Z6-AC05, Z4-AC05, Z4-AC09

### Ce qui se passe cote API

1. `POST /api/v1/sessions/daily` avec `{ chapter_id }`
   - Le service compose une session avec les items UNKNOWN due
   - Z6-AC05 : 100% UNKNOWN, difficulte 1, 6-10 questions
   - Z4-AC05 : pack constraints respectees (max 2 WRITING, 1 DOCUMENT si dispo)
   - Retourne `{ id, session_type: "daily", status: "COMPOSING" }`

2. `GET /api/v1/sessions/{session_id}/questions`
   - Retourne 8 questions generees depuis les items
   - Chaque question a : `question_type`, `rendered_prompt`, `choices` (pour MCQ)
   - Templates utilises en round-robin : `FLASH_MCQ`, `DEF_SHORT`, `CLOZE_KEYWORDS`

### Deroulement de la session (~3 min)

La session alterne 3 types de questions :

**Question type MCQ (Vrai/Faux)**
```
Vrai ou faux : La masse volumique est le
rapport de la masse d'un corps sur son volume.

  [A] Vrai          ← boutons de choix
  [B] Faux

  [ Valider ]
```

**Question type SHORT_ANSWER**
```
Definis : L'unite SI de la...

  ┌─────────────────────────┐
  │ Ta reponse...           │  ← champ texte
  └─────────────────────────┘

  [ Valider ]
```

**Question type CLOZE**
```
Complete avec les mots-cles : La ____ d'un
corps est le ____ de sa masse volumique...

  ┌─────────────────────────┐
  │ Ta reponse...           │
  └─────────────────────────┘

  [ Valider ]
```

### Flow par question

```
Question affichee
       │
       ▼
Eleve tape reponse + [ Valider ]
       │
       ▼
Ecran auto-evaluation :
  « Est-ce que tu connaissais la reponse ? »
  [ Je savais ✓ ]  (score = 1.0)
  [ Je ne savais pas ✗ ]  (score = 0.0)
       │
       ▼
POST /sessions/{id}/answer
  { question_id, answer, score }
       │
       ▼
Ecran feedback :
  ✅ Correct ! / ❌ Pas tout a fait...
  Reponse attendue : ...
  💡 Indice : ...          ← Z4-AC09
       │
       ▼
[ Suivant → ]  ou  [ Voir le bilan ]
```

### Transitions mastery attendues

Apres la session, selon les reponses :
- Score 1.0 (« Je savais ») → Z1-AC01 : UNKNOWN → FRAGILE, cs=1
- Score 0.0 (« Je ne savais pas ») → Z1-AC11 : reste UNKNOWN, cs=0

---

## Etape 4 — Debrief

**Ecran :** composant `DebriefView` dans `app/session/[id].tsx`
**Declencheur :** derniere question repondue
**ACs :** debrief session

### Ce qui se passe

1. `GET /api/v1/sessions/{session_id}/debrief`
   - Retourne `{ score, total, percentage, transitions[] }`

2. L'ecran affiche :
   - Emoji 🎉 + « Bravo ! »
   - Score : `X / 8` avec pourcentage
   - Duree : `Xm XXs`
   - Section « Progressions » : transitions mastery (UNKNOWN → FRAGILE pour les reussites)
   - Bouton **« Retour au chapitre »**
   - Bouton **« Encore une session »**

### Exemple de debrief

```
🎉 Bravo !

     6 / 8
     75%
     ⏱ 2 min 34s

PROGRESSIONS
📈 La masse volumique est le rapport...  UNKNOWN → FRAGILE
📈 L'unite SI de la masse...            UNKNOWN → FRAGILE
📈 ρ = m / V                            UNKNOWN → FRAGILE
📈 Un corps flotte si...                UNKNOWN → FRAGILE
📈 La masse volumique de l'eau...       UNKNOWN → FRAGILE
📈 Pour convertir g/cm³ en kg/m³...     UNKNOWN → FRAGILE

  [ Retour au chapitre ]
  [ Encore une session ]
```

---

## Etape 5 — Retour au dashboard

**Ecran :** `app/(tabs)/index.tsx`
**Declencheur :** bouton « Retour au chapitre » puis navigation retour

### Ce qui se voit

Le dashboard affiche maintenant le chapitre demo avec la progression mise a jour :
- Barre de mastery : mix de FRAGILE (orange) et UNKNOWN (gris)
- Pourcentage affiche (ex: 0% si on compte OK+SOLID uniquement)
- « 8 items »

L'eleve peut :
1. **Relancer une session** sur le chapitre demo (les items FRAGILE seront due apres 1 jour)
2. **Capturer un vrai cours** via le bouton « 📸 Capturer un cours »
3. **Explorer les notions** du chapitre demo

---

## Etape 6 — Capture du premier vrai cours (optionnel, quand l'eleve veut)

**Ecran :** `app/capture.tsx` (modal) → `app/processing.tsx`
**ACs :** Z8-AC04, Z8-AC03, Z2-AC01

> Cette etape est non-bloquante. L'eleve peut continuer a utiliser le chapitre demo aussi longtemps qu'il veut. Le flow de capture n'est pas encore cable au pipeline (Lot 0 WIP).

### Flow prevu

1. L'eleve prend des photos de son cahier (1-30 pages)
2. `POST /api/v1/chapters/{chapter_id}/upload` avec les images
3. Ecran de progression (Z8-AC04) :
   - Phase 1 (0-5s) : « Lecture de tes pages... »
   - Phase 2 (5-15s) : « Creation des questions... »
   - Phase 3 (15s+) : « Presque fini... »
4. Si echec : ecran de recovery (Z8-AC03) avec conseils photos
5. Si succes : le chapitre reel apparait dans le dashboard
6. Z8-AC01 : le chapitre demo est automatiquement archive

---

## Sequence API complete (premiere connexion)

```
1.  GET  /dev/token                           → { token, user_id }
2.  GET  /api/v1/chapters                     → []  (vide)
3.  POST /api/v1/onboarding/seed-demo         → { chapter_id, item_count: 8 }
4.  GET  /api/v1/chapters                     → [{ chapitre demo, 8 items, 8 UNKNOWN }]
5.  GET  /api/v1/chapters/{id}/lesson-card    → { chapter, items[8], notions[3] }
6.  GET  /api/v1/masteries                    → [{ item_id, state: UNKNOWN } x8]
7.  POST /api/v1/sessions/daily               → { session_id, type: daily }
8.  GET  /api/v1/sessions/{id}/questions      → [{ question } x8]
9.  POST /api/v1/sessions/{id}/answer         → { feedback }     (x8 fois)
10. GET  /api/v1/sessions/{id}/debrief        → { score, transitions }
```

---

## Temps estime par etape

| Etape | Duree | Cumul |
|-------|-------|-------|
| Ouverture + chargement | 2s | 2s |
| Seed demo + refresh | 1s | 3s |
| Navigation chapitre | 1s | 4s |
| Chargement lesson card | 1s | 5s |
| Lancer session | 2s | 7s |
| Repondre a 8 questions | ~2.5 min | ~2 min 37s |
| Debrief | 10s | ~2 min 47s |

**Total : ~3 minutes** de l'ouverture au premier debrief.

---

## Points d'attention UX

1. **Aucune friction avant le premier quiz.** Pas de formulaire, pas d'emploi du temps, pas de configuration. Un seul tap « Charger le chapitre demo » et c'est parti.

2. **Auto-evaluation honnete.** Le flow « Valider → Je savais / Je ne savais pas » responsabilise l'eleve. Pas de correction automatique pour le Lot 0 (pas de LLM scoring en temps reel).

3. **Feedback systematique.** Chaque question montre la reponse attendue + un indice (Z4-AC09), meme en cas de reussite.

4. **Progression visible.** Apres la session, le dashboard montre concretement les transitions UNKNOWN → FRAGILE. L'eleve voit que quelque chose a change.

5. **Pas de dead end.** Si la session echoue a se composer (pas d'items due), le message est explicite et un bouton de retour est affiche.

6. **Chapitre demo archive automatiquement.** Des que l'eleve upload un vrai cours, le chapitre demo disparait du dashboard (Z8-AC01).
