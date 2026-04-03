# Modele d'etats par ecran -- Revise Mieux

> Pour chaque ecran critique, les 11 etats d'affichage sont decrits explicitement.
>
> Les etats couverts : Default, Loading, Processing, Success, Empty, Partial, Stale, OCR Uncertainty, Error, Blocked, Retry.

---

## 1. Dashboard eleve

| Etat | Ce que l'utilisateur voit | Comportement technique |
|------|--------------------------|----------------------|
| **Default (soiree)** | Plan de soiree avec etapes, estimation duree, bouton "C'est parti" | EveningPlan calcule, sessions composees |
| **Default (hors soiree)** | Progression globale, liste des chapitres avec barres de maitrise | `GET /chapters` + masteries |
| **Loading** | Skeleton : rectangles animes pour header, plan, chapitres | `isLoading = true` sur le fetch initial |
| **Processing** | Bandeau "Traitement de tes photos en cours..." avec lien vers processing.tsx | Pipeline J0 actif pour un chapitre |
| **Success** | Plan de soiree avec coche verte sur une etape terminee, message de transition | Apres completion d'une session, refresh du plan |
| **Empty** | Checklist 3 etapes + chapitre demo. "Aucun chapitre" avec CTA "Charger le chapitre demo" | 0 chapitres en base, Z8-AC01/AC02 |
| **Partial** | Chapitres charges mais plan de soiree incomplet (schedule partiel) | Certaines matieres sans ScheduleSlots, bandeau "Ajoute ton emploi du temps" |
| **Stale** | Donnees affichees normalement, indicateur discret de synchro | Pull-to-refresh disponible, background refresh en cours |
| **OCR Uncertainty** | Badge sur le chapitre : "2 zones a verifier" | Items avec `validation_required = true` |
| **Error** | Message "Impossible de charger tes chapitres" + bouton "Reessayer" | Erreur reseau sur `GET /chapters` |
| **Blocked** | N/A -- le dashboard n'est jamais bloque | |
| **Retry** | Bouton "Reessayer" apres un echec reseau | Relance le fetch |

---

## 2. Capture photo

| Etat | Ce que l'utilisateur voit | Comportement technique |
|------|--------------------------|----------------------|
| **Default** | Zone camera avec guidage "Cadre ton cahier", compteur pages, vignettes | Camera active, aucune photo prise |
| **Loading** | Spinner bref pendant l'initialisation camera | Chargement expo-camera |
| **Processing** | N/A -- le processing est sur un ecran dedie | |
| **Success** | Flash visuel + vignette ajoutee a la bande de pages | Photo capturee, ajoutee a la liste locale |
| **Empty** | Zone camera vide, bouton "Terminer" inactif (opacite reduite) | 0 pages capturees |
| **Partial** | Vignettes des pages capturees + bouton "+" pour en ajouter | Entre 1 et 29 pages |
| **Stale** | N/A -- tout est local | |
| **OCR Uncertainty** | N/A -- pre-processing (l'analyse vient apres) | |
| **Error** | "Camera indisponible" + option "Importer depuis la galerie" | Permission camera refusee ou hardware |
| **Blocked** | "30 pages maximum atteintes" -- bouton capture desactive | pages.length >= 30 |
| **Retry** | "Permission camera requise" + bouton vers les reglages | Apres refus de permission |

---

## 3. Processing (pipeline OCR)

| Etat | Ce que l'utilisateur voit | Comportement technique |
|------|--------------------------|----------------------|
| **Default** | N/A -- cet ecran n'a pas d'etat au repos | |
| **Loading** | Phase 1 : "Lecture de tes pages..." avec animation de scan | Progression 0-30% |
| **Processing** | Phase 2 : "Creation des questions..." puis Phase 3 : "Presque fini..." | Progression 30-70% puis 70-100% |
| **Success** | Redirection automatique vers le chapitre cree | Pipeline termine avec >= 1 item valide |
| **Empty** | N/A -- redirige vers Recovery si 0 items | |
| **Partial** | "3/10 pages analysees" avec items deja disponibles | Traitement en cours, pages traitees progressivement (Z2-AC01) |
| **Stale** | "Tu peux fermer l'app, on te previent quand c'est pret" (apres 30s) | Notification push programmee a la fin du pipeline |
| **OCR Uncertainty** | N/A -- l'incertitude est geree apres, sur le chapitre | |
| **Error** | Redirection vers l'ecran Recovery (Z8-AC03) | Pipeline echoue ou timeout > 60s |
| **Blocked** | N/A | |
| **Retry** | Bouton "Reessayer" sur l'ecran Recovery | Relance le pipeline J0 |

---

## 4. Detail chapitre / Carte de lecon

| Etat | Ce que l'utilisateur voit | Comportement technique |
|------|--------------------------|----------------------|
| **Default** | Barre maitrise globale + notions en accordeon + bouton "Lancer une session" | Lesson card + masteries charges |
| **Loading** | Skeleton : barre de maitrise grise, accordeons placeholder | `GET /chapters/{id}/lesson-card` + `GET /masteries` |
| **Processing** | Bandeau "Analyse des pages 4-6 en cours..." avec progression | Pipeline incremental actif (Z2-AC14) |
| **Success** | Animation transition maitrise apres une session completee | Refresh des masteries apres session |
| **Empty** | "Aucun contenu extrait. Verifie la qualite des photos et re-uploade." | 0 items valides (Z2-AC07) |
| **Partial** | Items des pages traitees affiches + placeholders pages en cours | Pipeline partiel (Z2-AC01), indicateur "3/10 pages analysees" |
| **Stale** | Donnees affichees, fond refresh en cours | Cache masteries potentiellement obsolete |
| **OCR Uncertainty** | Badges "Zone illisible · page 4" + bandeau "N zones a verifier" | Pages en FAILED, items avec validation_required |
| **Error** | "Impossible de charger le chapitre" + bouton "Reessayer" | Erreur reseau |
| **Blocked** | Bouton diagnostic desactive + message "Aucun contenu n'a pu etre extrait" | 0 items valides (Z2-AC07) |
| **Retry** | Badge "Exercices non generes · page X" avec bouton "Reessayer" | Z2-AC11, max 3 retries manuels |

---

## 5. Session de revision

| Etat | Ce que l'utilisateur voit | Comportement technique |
|------|--------------------------|----------------------|
| **Default** | Question affichee + barre progression X/Y + zone de reponse | Session IN_PROGRESS, question courante |
| **Loading** | Skeleton question + spinner pendant la composition | `POST /sessions/daily` + `GET /sessions/{id}/questions` |
| **Processing** | N/A -- le LLM a deja genere les questions (lazy cache) | |
| **Success** | Feedback vert "Bonne reponse !" + micro-celebration si transition positive | Score >= 0.7, transition maitrise positive |
| **Empty** | "Tous tes items sont a jour ! Prochaine revision : [date]" + option consolidation | Aucun item eligible (Z4-AC13) |
| **Partial** | "Petite session rapide -- N questions sur ce chapitre" | < 5 items dans le chapitre (Z4-AC11) |
| **Stale** | N/A -- les questions sont fraiches a la composition | |
| **OCR Uncertainty** | Question avec badge "Contenu en cours de verification" | Item avec validation_required, gabarit simple uniquement |
| **Error** | "Preparation de tes exercices en cours... reessaie dans quelques secondes" | LLM timeout (Z4-AC06) |
| **Blocked** | N/A -- la session ne bloque jamais (gabarits simples si validation en cours) | |
| **Retry** | Apres 10s LLM timeout : mode "Relecture active" avec flashcards statiques | Fallback Z4-AC12 |

---

## 6. Validation HITL

| Etat | Ce que l'utilisateur voit | Comportement technique |
|------|--------------------------|----------------------|
| **Default** | Item avec suggestion + 4 boutons d'action (Confirmer/Corriger/JNSP/Ignorer) | ValidationTask PENDING |
| **Loading** | Spinner pendant le chargement des validations | `GET /validations` |
| **Processing** | N/A | |
| **Success** | Confirmation visuelle "Mis a jour !" + navigation vers la validation suivante | Task resolue (CONFIRMED/CORRECTED) |
| **Empty** | "Aucune verification en attente. Tout est bon !" | 0 ValidationTasks PENDING (Z3-AC09) |
| **Partial** | N/A | |
| **Stale** | N/A | |
| **OCR Uncertainty** | Crop de la page source affiche a cote de la suggestion pour comparaison | Item a faible confidence |
| **Error** | "Impossible de sauvegarder ta reponse" + bouton "Reessayer" | Erreur reseau sur POST |
| **Blocked** | N/A | |
| **Retry** | Bouton "Reessayer" apres echec de sauvegarde | Relance le POST |

---

## 7. Dashboard parent

| Etat | Ce que l'utilisateur voit | Comportement technique |
|------|--------------------------|----------------------|
| **Default** | Maitrise globale de l'enfant + chapitres + activite semaine | Donnees fraiches |
| **Loading** | Skeleton dashboard | Chargement initial |
| **Processing** | N/A | |
| **Success** | N/A (pas d'action mutative depuis le dashboard parent) | |
| **Empty (enfant n'a pas commence)** | "[Prenom] n'a pas encore commence -- vous serez notifie des sa premiere session" | 0 sessions, 0 chapitres |
| **Partial** | Chapitres avec items en verification mentionnes | Items validation_required |
| **Stale** | Donnees du dernier digest, indicateur "Mis a jour il y a 2j" | Parent ne consulte pas frequemment |
| **OCR Uncertainty** | "2 points en verification -- exercices simplifies en attendant" | Items validation_required dans le chapitre |
| **Error** | "Connexion perdue -- donnees du dernier chargement affichees" | Erreur reseau |
| **Blocked** | N/A | |
| **Retry** | Pull-to-refresh | |

---

## 8. Onboarding eleve

| Etat | Ce que l'utilisateur voit | Comportement technique |
|------|--------------------------|----------------------|
| **Default** | Checklist 3 etapes + CTAs | Premier acces |
| **Loading** | Spinner sur le bouton "Charger le chapitre demo" | POST /onboarding/seed-demo |
| **Processing** | N/A | |
| **Success** | Redirection vers le dashboard avec le chapitre demo | Seed reussi |
| **Empty** | N/A (c'est l'etat initial par definition) | |
| **Partial** | N/A | |
| **Stale** | N/A | |
| **OCR Uncertainty** | N/A | |
| **Error** | "Impossible de charger le chapitre demo" + fallback vers le dashboard vide | Erreur seed |
| **Blocked** | N/A | |
| **Retry** | Bouton "Reessayer" ou navigation directe vers capture | |

---

## Matrice recapitulative : etats critiques par ecran

| Ecran | Etats les plus risques (UX) | AC de reference |
|-------|---------------------------|----------------|
| Dashboard | Empty (premier usage), Error (perte de confiance) | Z8-AC01, Z8-AC02 |
| Capture | Error (permission camera), Blocked (30 pages) | Z2-AC03 |
| Processing | Error (pipeline echoue, premier upload critique) | Z8-AC03, Z8-AC04 |
| Detail chapitre | Partial (pages en cours), OCR Uncertainty (zones illisibles) | Z2-AC01, Z2-AC07 |
| Session | Empty (rien a reviser), Error (LLM timeout), Retry (fallback flashcards) | Z4-AC06, Z4-AC12, Z4-AC13 |
| Validation HITL | Default (comprehension des actions) | Z3-AC02 a Z3-AC05 |
| Dashboard parent | Empty (enfant pas commence), Stale (pas consulte depuis longtemps) | Z8-AC05, Z6-AC27 |
