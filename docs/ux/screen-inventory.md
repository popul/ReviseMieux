# Inventaire des ecrans -- Revise Mieux

> Tous les ecrans de l'app eleve et parent, avec route Expo Router, intention UX, CTA principal, donnees necessaires et mapping ACs.
>
> **Legende** : (existant) = deja implemente dans mobile/ | (a creer) = manquant

---

## Tab Bar (eleve)

3 onglets : Dashboard | Capture | Parametres

---

## 1. Dashboard eleve (existant)

| Champ | Valeur |
|-------|--------|
| **Route** | `app/(tabs)/index.tsx` |
| **Intention** | Savoir quoi faire maintenant. En soiree : plan de soiree. Hors soiree : progression globale. |
| **CTA principal** | "C'est parti !" sur la premiere etape non completee du plan de soiree |
| **CTA secondaires** | "Capturer un cours", tap sur un chapitre, "Je n'ai pas beaucoup de temps" (mode express) |
| **Comprehension < 3s** | "Ce soir : N activites, ~X min" (soiree) OU "X% maitrise globale" (hors soiree) |
| **Ton emotionnel** | Encourageant, structure. "Bonsoir Hugo ! Ce soir : 3 activites, ~15 min." |
| **Donnees** | `GET /api/v1/chapters`, calcul EveningPlan cote client ou serveur, masteries par chapitre |
| **Mapping ACs** | Z6-AC30 (progression globale), Z7-AC01 (EveningPlan), Z7-AC02 (dashboard soiree), Z7-AC04 (estimation duree), Z7-AC05 (fini pour ce soir), Z7-AC08 (mode express), Z7-AC10 (rien a reviser), Z7-AC12 (arc emotionnel), Z8-AC01 (chapitre demo), Z8-AC02 (empty state) |

### Wireframe textuel -- Etat soiree

```
┌─────────────────────────────────┐
│ Bonsoir Hugo !                  │
│ Ce soir : 3 activites, ~15 min  │
│ [Je n'ai pas beaucoup de temps] │
├─────────────────────────────────┤
│ □ Saisis ton cours d'HG  ~2min  │
│   [Capturer]                    │
│                                 │
│ □ Premiere revision HG   ~5min  │
│   en attente de capture         │
│                                 │
│ □ Revision quotidienne  ~8min   │
│   Physique + Maths              │
│   [C'est parti !]               │
├─────────────────────────────────┤
│ PROGRESSION GLOBALE             │
│ ████████░░░ 62% maitrise        │
│ 5 SOLID  8 OK  4 FRAGILE       │
├─────────────────────────────────┤
│ MES CHAPITRES                   │
│ ┌─ Physique-Chimie ───────────┐ │
│ │ Densite et masse volumique  │ │
│ │ ████████░░ 72%   8 items    │ │
│ │ [Reviser]                   │ │
│ └─────────────────────────────┘ │
│ ┌─ Histoire-Geo ──────────────┐ │
│ │ Les inegalites dans le monde│ │
│ │ ████░░░░░░ 35%   12 items   │ │
│ │ [Reviser]                   │ │
│ └─────────────────────────────┘ │
│                                 │
│ [Capturer un cours]             │
└─────────────────────────────────┘
```

### Wireframe textuel -- Empty state (premier usage)

```
┌─────────────────────────────────┐
│ Bienvenue Hugo !                │
│                                 │
│      🧪                        │
│  Aucun chapitre                 │
│  Essaie le chapitre demo pour   │
│  decouvrir l'app !              │
│  [Charger le chapitre demo]     │
│                                 │
├─ OU ────────────────────────────┤
│                                 │
│  1. ✓ Compte cree               │
│  2. → Photographie ton cours    │
│     [Capturer]   30 secondes    │
│  3.   Ta premiere revision      │
│       ~5 min apres la capture   │
│                                 │
└─────────────────────────────────┘
```

---

## 2. Capture photo (existant)

| Champ | Valeur |
|-------|--------|
| **Route** | `app/(tabs)/capture.tsx` |
| **Intention** | Photographier les pages du cahier de l'eleve, rapidement et sans friction |
| **CTA principal** | "Prendre une photo" / "Terminer" |
| **CTA secondaires** | "Galerie" (importer depuis album), miniatures des pages prises, bouton "+" |
| **Comprehension < 3s** | "Cadre ton cahier dans le rectangle" |
| **Ton emotionnel** | Fonctionnel, rapide. Pas de decoration inutile. |
| **Donnees** | Acces camera (expo-camera), stockage local temporaire des images, compteur pages X/30 |
| **Mapping ACs** | Z2-AC03 (photo floue), Z5-AC11 (ajout incremental), Z8-AC03 (recovery premier OCR) |

### Wireframe textuel

```
┌─────────────────────────────────┐
│ Capturer un cours     3/30      │
├─────────────────────────────────┤
│ ┌ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┐│
│                                 │
│       📷                        │
│   Cadre ton cahier              │
│   dans le rectangle             │
│                                 │
│   [Prendre une photo]           │
│                                 │
│ └ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┘│
│                                 │
│ Pages capturees                 │
│ [1] [2] [3] [+]                │
│                                 │
│ [Galerie]      [Terminer ->]    │
└─────────────────────────────────┘
```

---

## 3. Parametres (existant)

| Champ | Valeur |
|-------|--------|
| **Route** | `app/(tabs)/settings.tsx` |
| **Intention** | Configurer le profil, l'emploi du temps, inviter un parent, gerer les notifications |
| **CTA principal** | Navigation vers les sous-ecrans de config |
| **CTA secondaires** | Inviter un parent (badge si non lie), emploi du temps, notifications |
| **Comprehension < 3s** | Liste de parametres organisee par categorie |
| **Ton emotionnel** | Neutre, professionnel |
| **Donnees** | Profil utilisateur, ScheduleSlots, ParentLink status, NotificationPrefs |
| **Mapping ACs** | Z6-AC01 (emploi du temps), Z6-AC20 (opt-out parent), Z6-AC44 (liaison parent), Z8-AC07 (invitation parent) |

---

## 4. Onboarding (existant)

| Champ | Valeur |
|-------|--------|
| **Route** | `app/onboarding.tsx` |
| **Intention** | Guider l'eleve vers sa premiere action (demo ou capture) en < 60 secondes |
| **CTA principal** | "Capturer" (premier cours) OU "Charger le chapitre demo" |
| **CTA secondaires** | Aucun -- parcours lineaire |
| **Comprehension < 3s** | "Pret en 3 etapes" avec checklist visuelle |
| **Ton emotionnel** | Accueillant, simple. Aucun formulaire. |
| **Donnees** | Statut onboarding (etapes completees), user profile minimal |
| **Mapping ACs** | Z8-AC01 (demo), Z8-AC02 (empty state), Z8-AC08 (sequence deterministe) |

---

## 5. Processing / Pipeline OCR (existant)

| Champ | Valeur |
|-------|--------|
| **Route** | `app/processing.tsx` |
| **Intention** | Informer l'eleve que le traitement est en cours, reduire l'anxiete d'attente |
| **CTA principal** | Aucun pendant le traitement. "Tu peux fermer l'app" apres 30s. |
| **CTA secondaires** | Bouton retour, notification push a la fin |
| **Comprehension < 3s** | Animation + message contextuel ("Lecture de tes pages...", "Creation des questions...", "Presque fini...") |
| **Ton emotionnel** | Rassurant, progressif. 3 phases visuelles distinctes. |
| **Donnees** | Pipeline status (polling ou SSE), progression par page |
| **Mapping ACs** | Z8-AC04 (ecran progression), Z2-AC01 (carte partielle), Z2-AC10 (streaming premiere page) |

---

## 6. Detail chapitre / Carte de lecon (existant)

| Champ | Valeur |
|-------|--------|
| **Route** | `app/chapter/[id].tsx` |
| **Intention** | Voir le contenu d'un chapitre, comprendre ou on en est par notion, lancer une session |
| **CTA principal** | "Lancer une session" |
| **CTA secondaires** | "Ajouter des pages", "Voir les items ignores", deployer/replier les notions, creer un exam |
| **Comprehension < 3s** | Barre de maitrise globale + notions en accordeon avec leur propre barre |
| **Ton emotionnel** | Informatif, organise. L'eleve identifie visuellement ses points faibles. |
| **Donnees** | `GET /api/v1/chapters/{id}/lesson-card`, `GET /api/v1/masteries`, notions avec items |
| **Mapping ACs** | Z7-AC15 (notions par concept_tag), Z7-AC16 (accordeon + maitrise par notion), Z5-AC04 (revision courante unique), Z5-AC11 (bouton ajouter des pages), Z2-AC01 (carte partielle), Z3-AC06 (utilisable avec 0 validations), Z3-AC16 (items ignores recuperables) |

### Wireframe textuel

```
┌─────────────────────────────────┐
│ <- Physique-Chimie              │
│ Densite et masse volumique      │
│ ████████░░ 72% maitrise         │
│ 8 items · revise il y a 1 jour  │
├─────────────────────────────────┤
│ [Lancer une session]            │
│ [Ajouter des pages]             │
├─────────────────────────────────┤
│ NOTIONS                         │
│                                 │
│ ▼ Masse volumique (rho)   44%  │
│   ███░░░░░                      │
│   · rho = m/V          FRAGILE  │
│   · unite SI            OK      │
│   · definition          SOLID   │
│                                 │
│ ▶ Deplacement d'eau     100%   │
│   ████████ (3 items SOLID)      │
│                                 │
│ ▶ Materiaux et densite   0%    │
│   ░░░░░░░░ (2 items UNKNOWN)   │
├─────────────────────────────────┤
│ ⚠ 2 zones a verifier           │
│ [Voir les validations]          │
├─────────────────────────────────┤
│ Points non verifies (1)         │
│ · [Reactiver]                   │
└─────────────────────────────────┘
```

---

## 7. Session de revision (existant)

| Champ | Valeur |
|-------|--------|
| **Route** | `app/session/[id].tsx` |
| **Intention** | Repondre a des questions adaptees, recevoir du feedback immediat, progresser |
| **CTA principal** | "Valider" (soumettre la reponse) |
| **CTA secondaires** | "Passer", "Voir ma lecon", "Je ne comprends pas", "Signaler une erreur", choices MCQ |
| **Comprehension < 3s** | Question en cours + progression (X/Y) + barre de progression visuelle |
| **Ton emotionnel** | Focus, encourageant. Le feedback est toujours factuel et positif. |
| **Donnees** | Session ID, questions[], current_question_index, attempt responses |
| **Mapping ACs** | Z1-AC01 a AC12 (transitions maitrise), Z1-AC14 (micro-celebrations), Z4-AC04 (reprise session), Z4-AC09 (feedback enrichi), Z4-AC10 (bouton passer), Z6-AC05 (evening_first 100% UNKNOWN), Z1-AC21 (voir lecon), Z1-AC22 (reformulation), Z1-AC23 (scoring graduel), Z3-AC12 (signalement erreur), Z3-AC15 (anti-clicking rapide) |

### Wireframe textuel -- Question MCQ

```
┌─────────────────────────────────┐
│ Question 3/8        ████░░░░    │
├─────────────────────────────────┤
│                                 │
│ Vrai ou faux :                  │
│ La masse volumique est le       │
│ rapport de la masse d'un corps  │
│ sur son volume.                 │
│                                 │
│ ┌───────────────────────────┐   │
│ │ [A] Vrai                  │   │
│ └───────────────────────────┘   │
│ ┌───────────────────────────┐   │
│ │ [B] Faux                  │   │
│ └───────────────────────────┘   │
│                                 │
│ [Passer]  [Voir ma lecon]       │
│                                 │
│        [Valider]                │
└─────────────────────────────────┘
```

### Wireframe textuel -- Feedback (reponse incorrecte)

```
┌─────────────────────────────────┐
│ Question 3/8        ████░░░░    │
├─────────────────────────────────┤
│                                 │
│ ❌ Pas tout a fait...           │
│                                 │
│ Reponse attendue :              │
│ "Vrai -- rho = m / V"          │
│                                 │
│ Ce qui manquait :               │
│ La formule associee             │
│                                 │
│ 💡 Retiens que la masse         │
│ volumique est toujours le       │
│ rapport masse / volume.         │
│                                 │
│ [Signaler une erreur]           │
│                                 │
│        [Suivant ->]             │
└─────────────────────────────────┘
```

---

## 8. Debrief session (sous-ecran de session)

| Champ | Valeur |
|-------|--------|
| **Route** | Composant `DebriefView` dans `app/session/[id].tsx` |
| **Intention** | Valoriser l'effort, montrer les progressions, donner le prochain objectif |
| **CTA principal** | "Terminer" |
| **CTA secondaires** | "Encore une session" (si items non couverts), "Retour au chapitre" |
| **Comprehension < 3s** | Score X/Y + liste des progressions positives |
| **Ton emotionnel** | Celebratoire et positif. Mettre en avant les progres, pas les echecs. |
| **Donnees** | `GET /api/v1/sessions/{id}/debrief` |
| **Mapping ACs** | Z1-AC15 (debrief fin session), Z1-AC14 (micro-celebrations), Z7-AC05 (ecran cloture soiree) |

---

## 9. Validation HITL (a creer)

| Champ | Valeur |
|-------|--------|
| **Route** | `app/validation/[id].tsx` |
| **Intention** | Permettre a l'eleve ou au parent de confirmer, corriger ou deleguer un item incertain |
| **CTA principal** | "Confirmer" / "Corriger" / "Je ne sais pas" |
| **CTA secondaires** | "Ignorer", navigation entre les validations |
| **Comprehension < 3s** | "L'IA a detecte : [suggestion]. C'est correct ?" |
| **Ton emotionnel** | Neutre, collaboratif. L'eleve aide l'app, pas l'inverse. |
| **Donnees** | `GET /api/v1/validations`, `POST /api/v1/validations/{id}/resolve` |
| **Mapping ACs** | Z3-AC01 (gabarits bloques), Z3-AC02 (confirmer), Z3-AC03 (corriger), Z3-AC04 (je ne sais pas), Z3-AC05 (ignorer), Z3-AC06 (chapitre utilisable avec 0 validations), Z3-AC09 (pas de validation si confidence haute) |

### Wireframe textuel

```
┌─────────────────────────────────┐
│ Verification 1/3                │
├─────────────────────────────────┤
│ L'IA a detecte :                │
│                                 │
│ ┌───────────────────────────┐   │
│ │ "rho = m / V"             │   │
│ │ (formule de la masse      │   │
│ │  volumique)               │   │
│ └───────────────────────────┘   │
│                                 │
│ C'est correct ?                 │
│                                 │
│ [✓ Confirmer]                   │
│ [✏ Corriger]                    │
│ [? Je ne sais pas]              │
│ [Ignorer]                       │
│                                 │
│ 6 zones a verifier              │
│ Tes exercices seront plus       │
│ precis apres verification.      │
└─────────────────────────────────┘
```

---

## 10. Ecran de recovery OCR (a creer)

| Champ | Valeur |
|-------|--------|
| **Route** | `app/recovery.tsx` (ou composant dans `processing.tsx`) |
| **Intention** | Guider l'eleve apres un echec OCR sans le bloquer |
| **CTA principal** | "Reprendre les photos" |
| **CTA secondaires** | "Essayer quand meme" (si items partiels), "Chapitre demo" (si premier upload) |
| **Comprehension < 3s** | "Les photos sont un peu difficiles a lire. Pas de panique !" |
| **Ton emotionnel** | Empathique, deculpabilisant. Conseils visuels concrets. |
| **Donnees** | Pipeline status, nombre de pages echouees, items partiels generes |
| **Mapping ACs** | Z8-AC03 (UX recovery premier OCR), Z2-AC03 (photo floue), Z2-AC07 (diagnostic impossible si 0 items) |

### Wireframe textuel

```
┌─────────────────────────────────┐
│ Les photos sont un peu          │
│ difficiles a lire.              │
│ Pas de panique !                │
│                                 │
│ ┌────┐ ┌────┐ ┌────┐           │
│ │ ☀️ │ │ 📐 │ │ ✍️ │           │
│ │Bon │ │Bien│ │Lis-│           │
│ │ecl.│ │cad.│ │ible│           │
│ └────┘ └────┘ └────┘           │
│                                 │
│ [Reprendre les photos]          │
│                                 │
│ [Essayer quand meme]            │
│ 2 items quand meme detectes     │
│                                 │
│ [S'entrainer sur le chapitre    │
│  demo en attendant]             │
└─────────────────────────────────┘
```

---

## 11. Creation d'exam (a creer)

| Champ | Valeur |
|-------|--------|
| **Route** | `app/exam/create.tsx` |
| **Intention** | Enregistrer un controle a venir pour activer le resserrement des intervalles |
| **CTA principal** | "Creer le controle" |
| **CTA secondaires** | Selection matiere, date, chapitres (pre-coches heuristique 6 semaines), notions (toggles) |
| **Comprehension < 3s** | "Quand est ton prochain controle ?" |
| **Ton emotionnel** | Rapide, structure. Parcours complet < 90 secondes. |
| **Donnees** | Chapitres par matiere, notions par chapitre, exams existants |
| **Mapping ACs** | Z6-AC10 (creation exam), Z7-AC17 (selection par notion), Z7-AC18 (auto-suggestion chapitres), Z1-AC08 (resserrement) |

---

## 12. Controle blanc (a creer)

| Champ | Valeur |
|-------|--------|
| **Route** | `app/session/[id].tsx` (meme composant, type `mock_exam`) |
| **Intention** | Simuler un controle dans les conditions reelles, evaluer la preparation |
| **CTA principal** | "Lancer le controle blanc" |
| **Comprehension < 3s** | "Controle blanc [matiere] -- X questions, ~Y min" |
| **Ton emotionnel** | Serieux mais pas stressant. "Simuler pour se preparer, pas pour se juger." |
| **Donnees** | Exam ID, questions multi-chapitres, score_confidence |
| **Mapping ACs** | Z6-AC09 (mock exam multi-chapitres), Z4-AC07 (non bloque par session daily), Z1-AC16 (caveat score si items non valides) |

---

## 13. Emploi du temps (a creer)

| Champ | Valeur |
|-------|--------|
| **Route** | `app/schedule.tsx` ou sous-ecran de `settings.tsx` |
| **Intention** | Renseigner/modifier les creneaux de cours pour activer les notifications et pre_class |
| **CTA principal** | Cocher les creneaux dans la grille Lu-Ve x Matin/AM |
| **CTA secondaires** | "Plus tard" (mode degrade accepte) |
| **Comprehension < 3s** | "Quand as-tu [Matiere] dans la semaine ?" |
| **Ton emotionnel** | Rapide, low-friction. Grille visuelle, pas de formulaire. |
| **Donnees** | ScheduleSlots par matiere, ScheduleExceptions |
| **Mapping ACs** | Z6-AC01 (saisie contextuelle), Z6-AC11 (mode degrade), Z6-AC16 (exceptions) |

---

## 14. Invitation parent (a creer)

| Champ | Valeur |
|-------|--------|
| **Route** | `app/invite-parent.tsx` (ou modal depuis settings) |
| **Intention** | Generer un code de liaison et le partager au parent |
| **CTA principal** | "Partager le code" (share sheet OS) |
| **CTA secondaires** | "Copier le code", "Plus tard" |
| **Comprehension < 3s** | Code 6 caracteres visible en grand + instruction "Envoie ce code a ton parent" |
| **Ton emotionnel** | Simple, rapide. Pas de friction. |
| **Donnees** | Code genere via `POST /api/v1/parent/link-code` |
| **Mapping ACs** | Z6-AC44 (liaison parent-eleve), Z8-AC07 (timing optimal apres 1ere session) |

---

## 15. Dashboard parent (a creer)

| Champ | Valeur |
|-------|--------|
| **Route** | `app/parent/dashboard.tsx` (ou app parent separee) |
| **Intention** | Vue d'ensemble rapide de la progression de l'enfant, sans micro-management |
| **CTA principal** | "Script 3 min" (questions orales a poser) |
| **CTA secondaires** | Voir le detail par chapitre, gerer notifications, ValidationTasks |
| **Comprehension < 3s** | "[Prenom] : X% maitrise globale, Y sessions cette semaine" |
| **Ton emotionnel** | Rassurant, factuel. Labels traduits (pas UNKNOWN mais "Pas encore vu"). |
| **Donnees** | Mastery breakdown par chapitre, sessions completees, digest data |
| **Mapping ACs** | Z6-AC27 (digest hebdo), Z6-AC29 (labels traduits), Z6-AC30 (progression globale), Z1-AC17 (script 3 min), Z8-AC05 (onboarding parent) |

### Wireframe textuel

```
┌─────────────────────────────────┐
│ Theo · 4e                       │
│ 68% maitrise globale            │
│ 4 sessions cette semaine, 42min │
├─────────────────────────────────┤
│ CHAPITRES                       │
│                                 │
│ Physique · Densite         72%  │
│ ████████░░  Bien acquis         │
│                                 │
│ Histoire · Inegalites      35%  │
│ ████░░░░░░  En cours            │
│ ⚠ 2 points en verification     │
│                                 │
│ [Script 3 min]                  │
├─────────────────────────────────┤
│ CETTE SEMAINE                   │
│ · 4 routines completees / 5     │
│ · 1 session manquee (mercredi)  │
│                                 │
│ PROCHAINS CONTROLES             │
│ · Physique : dans 5 jours       │
│   [Voir le detail]              │
├─────────────────────────────────┤
│ Validations en attente (2)      │
│ [Voir]                          │
└─────────────────────────────────┘
```

---

## 16. Onboarding parent (a creer)

| Champ | Valeur |
|-------|--------|
| **Route** | `app/parent/onboarding.tsx` (3 ecrans swipeable) |
| **Intention** | Expliquer la proposition de valeur en 30 secondes |
| **CTA principal** | Swipe ou "Suivant" pour progresser, "Passer" |
| **Comprehension < 3s** | Ecran 1 : "Voici ce que fait [Prenom]" avec schema photo->quiz->maitrise |
| **Ton emotionnel** | Professionnel, rassurant. Pas enfantin. |
| **Donnees** | Prenom enfant, statut sessions (si deja faites) |
| **Mapping ACs** | Z8-AC05 (onboarding parent 3 ecrans), Z8-AC06 (digest anticipe) |

---

## 17. Angles morts (a creer)

| Champ | Valeur |
|-------|--------|
| **Route** | `app/blind-spots.tsx` (ou onglet dans dashboard) |
| **Intention** | Identifier les chapitres sans controle a venir -- zones de risque |
| **CTA principal** | "Creer un controle" (pre-rempli avec le chapitre) |
| **Comprehension < 3s** | "N chapitres sans controle prevu. Verifie avec ton prof !" |
| **Ton emotionnel** | Proactif, non anxiogene. Information actionnable. |
| **Donnees** | Chapitres sans exam actif, historique controles par matiere |
| **Mapping ACs** | Z7-AC19 (angles morts), Z7-AC20 (prediction interro surprise), Z7-AC21 (alerte notion fragile) |

---

## 18. Fiche PDF (a creer)

| Champ | Valeur |
|-------|--------|
| **Route** | `app/chapter/[id]/pdf.tsx` (ou modal) |
| **Intention** | Generer une fiche de revision papier pour l'etude sans telephone |
| **CTA principal** | "Telecharger la fiche" |
| **CTA secondaires** | Choix du nombre de questions, selection par chapitre ou matiere |
| **Comprehension < 3s** | "Fiche revision -- 15 questions, Physique-Chimie" |
| **Ton emotionnel** | Utilitaire, efficace |
| **Donnees** | Items urgents, questions en cache, generation PDF serveur |
| **Mapping ACs** | Z7-AC23 (fiche PDF), Z7-AC24 (report resultats papier) |

---

## 19. Report papier (a creer)

| Champ | Valeur |
|-------|--------|
| **Route** | `app/paper-report/[id].tsx` (ou via QR code) |
| **Intention** | Reporter rapidement les resultats d'une fiche papier |
| **CTA principal** | "Valider" le report |
| **Comprehension < 3s** | Checklist rapide : ✓/✗ pour chaque question |
| **Ton emotionnel** | Rapide, valorisant. "Tu as revise 12 points en etude !" |
| **Donnees** | PDF log avec item_ids, resultats tap ✓/✗ |
| **Mapping ACs** | Z7-AC24 (report resultats), Z7-AC26 (impact maitrise) |

---

## 20. Ecran de cloture soiree (a creer)

| Champ | Valeur |
|-------|--------|
| **Route** | Composant modal apres la derniere session du plan |
| **Intention** | Donner la permission explicite de fermer l'app |
| **CTA principal** | "Fermer" |
| **Comprehension < 3s** | "Bravo ! Soiree terminee en X min." |
| **Ton emotionnel** | Celebratoire, conclusif. L'eleve peut aller faire autre chose. |
| **Donnees** | EveningPlan.completed_at, resume matrieres, progressions |
| **Mapping ACs** | Z7-AC05 (etat fini pour ce soir), Z7-AC13 (notification parent routine terminee) |

---

## Resume : ecrans existants vs manquants

| # | Ecran | Route | Statut |
|---|-------|-------|--------|
| 1 | Dashboard eleve | `app/(tabs)/index.tsx` | Existant (a enrichir avec plan soiree) |
| 2 | Capture photo | `app/(tabs)/capture.tsx` | Existant |
| 3 | Parametres | `app/(tabs)/settings.tsx` | Existant (a enrichir) |
| 4 | Onboarding | `app/onboarding.tsx` | Existant |
| 5 | Processing OCR | `app/processing.tsx` | Existant |
| 6 | Detail chapitre | `app/chapter/[id].tsx` | Existant |
| 7 | Session revision | `app/session/[id].tsx` | Existant |
| 8 | Debrief session | composant dans session | Existant |
| 9 | Validation HITL | `app/validation/[id].tsx` | **A creer** |
| 10 | Recovery OCR | `app/recovery.tsx` | **A creer** |
| 11 | Creation exam | `app/exam/create.tsx` | **A creer** |
| 12 | Controle blanc | reutilise session | Existant (a parametrer) |
| 13 | Emploi du temps | `app/schedule.tsx` | **A creer** |
| 14 | Invitation parent | `app/invite-parent.tsx` | **A creer** |
| 15 | Dashboard parent | `app/parent/dashboard.tsx` | **A creer** |
| 16 | Onboarding parent | `app/parent/onboarding.tsx` | **A creer** |
| 17 | Angles morts | `app/blind-spots.tsx` | **A creer** |
| 18 | Fiche PDF | `app/chapter/[id]/pdf.tsx` | **A creer** |
| 19 | Report papier | `app/paper-report/[id].tsx` | **A creer** |
| 20 | Cloture soiree | composant modal | **A creer** |

**Bilan** : 8 ecrans existants, 12 ecrans a creer. Les ecrans existants couvrent le Lot 0 P1 (boucle minimale). Les ecrans manquants couvrent le Lot 0 P2 et le MVP complet.
