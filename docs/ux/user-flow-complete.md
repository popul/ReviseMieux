# User Flow complet -- Revise Mieux

> Document UX principal. Couvre le defi produit, les personas, les 9 parcours, le flow ecran par ecran, le modele d'etats, les edge cases, les principes UX pedagogiques, l'equilibre parent/eleve, les implications React Native, les decisions ouvertes et les recommandations finales.

---

## 1. Executive summary

Revise Mieux transforme des photos de cahier en sessions de revision adaptatives pour les collegiens (11-15 ans), avec un dashboard parent integre. Le defi UX central est la **double adoption** : l'eleve decide d'ouvrir l'app chaque soir, le parent decide de payer et d'installer. Chaque decision de design doit satisfaire les deux audiences simultanement sans compromettre ni l'engagement de l'eleve ni la confiance du parent.

Ce document detaille les 20 ecrans de l'app (8 existants, 12 a creer), les 9 parcours produit cles, les 11 etats d'affichage par ecran critique, et les 30 edge cases identifies. Il est directement exploitable par un developpeur React Native avec Expo Router.

**Chiffres cles** :
- 53 ACs couverts par le Lot 0 (33 P1 + 20 P2)
- 171 ACs totaux repartis sur 8 zones
- 4 chapitres pilotes (HG, SVT, Physique-Chimie)
- Cible : premiere session reussie en < 3 minutes apres l'install

---

## 2. Defi UX produit : la double adoption

### Pourquoi ce n'est PAS une app standard a audience unique

La plupart des apps educatives ciblent soit l'eleve (Duolingo, Quizlet) soit le parent (Klassroom, Pronote). Revise Mieux doit convaincre les deux simultanement, dans des cadres temporels differents :

- **L'eleve** ouvre l'app le soir, 10-15 min. Sa decision est repetee quotidiennement. Chaque session est un micro-choix : "est-ce que j'ouvre l'app ou je regarde TikTok ?" Le design doit donner envie, pas obliger.
- **Le parent** ouvre l'app 1-2 fois par semaine, < 2 min. Sa decision est ponctuelle (installation, renouvellement). Il a besoin d'un signal fiable sans micro-management.

### Tensions entre les deux audiences

| L'eleve veut | Le parent veut | Tension |
|-------------|---------------|---------|
| Interface attractive, moderne | Interface serieuse, credible | Esthetique intergenerationnelle |
| Pas de surveillance | Visibilite sur l'usage | Transparence vs autonomie |
| Pas de culpabilisation | Alertes si inactivite | Notification vs bien-etre |
| Recompenses implicites (progression) | Metriques explicites (scores) | Granularite du feedback |
| Sessions courtes et optionnelles | Usage regulier et mesurable | Frequence vs liberte |

### Risques d'adoption

**Cote eleve** :
- Friction dans les 30 premieres secondes -> desinstallation
- Premier OCR echoue sans recovery -> "ca marche pas"
- Sessions trop longues ou repetitives -> lassitude
- Sentiment de surveillance parentale -> rejet emotionnel

**Cote parent** :
- Aucun signal pendant 6 jours -> oubli de l'app
- Metriques incomprehensibles (UNKNOWN, FRAGILE) -> perte de confiance
- Trop de notifications -> desactivation complete
- Score gonfle sur QCM faciles -> fausse securite

### Qu'est-ce qui fait revenir un collegien de 13 ans chaque soir ?

1. **Previsibilite** : "Ce soir : 3 activites, ~15 min." L'engagement est defini, pas ouvert.
2. **Progression visible** : des items qui passent de UNKNOWN a SOLID. Un % qui monte.
3. **Permission de fermer** : l'ecran "Fini pour ce soir" dit explicitement "c'est assez".
4. **Variete** : alternance des types de questions (MCQ, CLOZE, SHORT) et des matieres.
5. **Pas de punition** : pas de streak, pas de regression par inactivite, pas de "tu as rate".

### Qu'est-ce qui fait qu'un parent de 45 ans fait confiance ?

1. **Preuve rapide** : digest anticipe J+2 avec temps de revision et matieres couvertes.
2. **Honnetete** : score_confidence visible, mention des items en verification.
3. **Engagement minimal** : notification "routine terminee" le soir + digest hebdo le dimanche.
4. **Labels lisibles** : "Bien acquis" au lieu de "SOLID", "En cours d'apprentissage" au lieu de "FRAGILE".
5. **Action concrete** : script 3 minutes avec 2-3 questions orales a poser.

---

## 3. Personas

Voir le document dedie : [`docs/ux/personas.md`](personas.md)

4 personas :
- **Theo (13 ans, 4e)** : motive mais irregulier. Veut 10-15 min/soir, pas plus.
- **Ines (11 ans, 6e)** : appliquee mais anxieuse. Besoin de reassurance et de "fini".
- **Lucas (15 ans, 3e)** : desengage. Seuil de friction zero. 30 secondes pour prouver la valeur.
- **Laurent (45 ans, pere)** : cherche un signal fiable sans micro-manager.

---

## 4. Parcours produit cles

### Parcours 1 : Onboarding -- premiere connexion

**Acteurs** : Eleve (+ parent potentiel)
**Objectif** : Premiere session reussie en < 3 minutes
**ACs** : Z8-AC01, Z8-AC02, Z8-AC04, Z8-AC08

```
Installation
    |
    v
Dashboard vide + chapitre demo (Z8-AC01, Z8-AC02)
    |
    v
Session demo evening_first (~3 min, 8 questions)
    |
    v
Debrief : 6/8 -> transitions UNKNOWN -> FRAGILE
    |
    v
Dashboard avec progression visible
    |
    v
Capture premier vrai cours (quand l'eleve veut)
    |
    v
Pipeline OCR (processing.tsx) -> chapitre cree
    |
    v
evening_first sur vrai chapitre -> debrief -> invitation parent (Z8-AC07)
```

**Gains pedagogiques** : premier contact avec le rappel actif. L'eleve decouvre que repondre a des questions est plus efficace que relire.

**Reduction d'effort** : 0 configuration requise. Pas d'emploi du temps, pas de matiere, pas de formulaire. Un seul tap "Charger le chapitre demo" et c'est parti.

---

### Parcours 2 : Capture J0

**Acteurs** : Eleve
**Objectif** : Transformer des photos de cahier en items revisables
**ACs** : Z2-AC01 a AC10, Z5-AC11, Z8-AC04

```
Tab Capture -> prendre 1-30 photos
    |
    v
(Optionnel) Saisie emploi du temps matiere (Z6-AC01)
    |
    v
Processing screen (3 phases, Z8-AC04)
    |
    v
    ├── Succes : chapitre cree -> carte de lecon avec items
    |
    └── Echec : ecran recovery (Z8-AC03)
              ├── Reprendre les photos
              ├── Essayer quand meme
              └── Chapitre demo
```

**Gains pedagogiques** : le contenu de revision est le propre cours de l'eleve, pas un contenu generique. L'ancrage est plus fort car l'eleve reconnait ses notes.

---

### Parcours 3 : Dashboard eleve -- plan d'action quotidien

**Acteurs** : Eleve
**Objectif** : Savoir quoi faire maintenant en < 5 secondes
**ACs** : Z7-AC01, Z7-AC02, Z7-AC04, Z7-AC05, Z7-AC08, Z6-AC30

```
Ouverture de l'app (17h-22h)
    |
    v
Dashboard soiree :
  "Bonsoir Hugo ! Ce soir : 3 activites, ~15 min."
    |
    v
  [C'est parti !]  ou  [Je n'ai pas beaucoup de temps]
    |                    |
    v                    v
  Plan complet          Mode express (~10 min)
    |                    |
    v                    v
  Completion progressive des etapes
    |
    v
  Ecran de cloture : "Bravo ! Soiree terminee en 14 min."
  -> notification parent "routine terminee"
```

**Reduction d'anxiete** : l'estimation de duree et l'ecran de cloture encadrent l'effort. L'eleve sait quand il commence et quand il finit.

---

### Parcours 4 : Session de revision

**Acteurs** : Eleve
**Objectif** : Repondre a des questions, recevoir du feedback, progresser
**ACs** : Z1-AC01 a AC12, Z4-AC04 a AC09, Z6-AC05

```
Lancer une session
    |
    v
Question affichee (MCQ / SHORT / CLOZE / NUMERIC)
    |
    v
Reponse + Valider
    |
    v
    ├── MCQ : scoring automatique
    └── Autres : auto-evaluation "Je savais / Je ne savais pas"
    |
    v
Feedback enrichi (reponse correcte + ce qui manquait + indice)
    |
    v
    ├── Transition positive ? -> micro-celebration
    └── Question suivante
    |
    v
Debrief : score, progressions, item prioritaire a revoir
```

**Gains pedagogiques** : rappel actif (repondre > relire). Feedback immediat avec explication. Alternance de formats pour activer differents circuits cognitifs.

---

### Parcours 5 : Revue des erreurs

**Acteurs** : Eleve
**Objectif** : Comprendre ce qui a ete rate, pas juste constater
**ACs** : Z4-AC09, Z1-AC19, Z1-AC21

Le feedback enrichi (Z4-AC09) apres chaque reponse incorrecte contient 3 elements :
1. La reponse correcte
2. Ce qui manquait (specifique au type de question)
3. Un indice pour la prochaine fois (1 phrase max)

L'eleve peut aussi :
- "Voir ma lecon" (Z1-AC21) pour le contexte complet
- "Je ne comprends pas la question" (Z1-AC22) pour une reformulation

Apres 3 echecs consecutifs sur le meme item, la difficulte descend automatiquement (Z1-AC19).

---

### Parcours 6 : Retour apres absence (3-7 jours)

**Acteurs** : Eleve
**Objectif** : Reprendre sans etre submerge
**ACs** : Z6-AC13, Z6-AC22, Z6-AC21

```
Ouverture apres 5 jours d'absence
    |
    v
Message : "Content de te revoir ! On reprend doucement."
    |
    v
Session "retour en douceur" :
  - 5 min max
  - 4-6 items les plus anciens en retard
  - Gabarits difficulte 1-2
    |
    v
Debrief : "Bon retour ! Les points restants arrivent demain."
    |
    v
Items restants etales sur 3-5 jours
```

**Principes critiques** : aucune regression par inactivite (Z6-AC13). Pas de streak. Pas de culpabilisation. Le message d'accueil est neutre ou positif.

---

### Parcours 7 : Dashboard parent

**Acteurs** : Parent
**Objectif** : Comprendre ce que fait l'app, verifier l'utilite
**ACs** : Z6-AC27, Z6-AC29, Z6-AC30, Z8-AC05

```
Premier acces apres liaison (code 6 caracteres)
    |
    v
Onboarding 3 ecrans (30s) :
  1. "Voici ce que fait [Prenom]"
  2. "Votre tableau de bord"
  3. "Restez informe sans effort"
    |
    v
Dashboard parent :
  - Maitrise globale (% items OK+SOLID)
  - Chapitres par matiere avec barres
  - Activite de la semaine
  - Script 3 minutes
  - Validations en attente (si mode actif)
```

**Labels traduits** (Z6-AC29) :
- UNKNOWN -> "Pas encore vu"
- FRAGILE -> "En cours d'apprentissage"
- OK -> "Compris, a consolider"
- SOLID -> "Bien acquis"
- validation_required -> "En verification"

---

### Parcours 8 : Validation HITL

**Acteurs** : Eleve ou parent
**Objectif** : Valider les items incertains generes par le LLM
**ACs** : Z3-AC01 a AC06

```
Bandeau sur le chapitre : "6 zones a verifier"
    |
    v
Ecran validation : item + suggestion IA
    |
    v
    ├── Confirmer -> confidence boostee, cache invalidee
    ├── Corriger -> contenu mis a jour, confidence = 1.0
    ├── Je ne sais pas -> delegue a l'admin, item en mode simple
    └── Ignorer -> item exclu des sessions, recuperable plus tard
```

**Point critique** : le chapitre est TOUJOURS utilisable meme si 0 validations faites (Z3-AC06). La validation est encouragee mais jamais bloquante.

---

### Parcours 9 : Gestion de contenu OCR incomplet

**Acteurs** : Eleve, systeme
**Objectif** : Continuer a travailler malgre des pages problematiques
**ACs** : Z2-AC01, Z2-AC02, Z2-AC04, Z2-AC05, Z2-AC11

```
Pipeline OCR avec 10 pages
    |
    v
Pages 1-3 traitees, 4-10 en cours
    |
    v
Carte de lecon affiche les items des pages 1-3 (Z2-AC01)
    + indicateur "3/10 pages analysees"
    |
    v
Page 4 timeout -> badge "Zone illisible page 4" (Z2-AC02)
Page 5 = 0 items -> marquee no_items_extracted (Z2-AC04)
Page 6 = erreur LLM -> badge "Exercices non generes page 6" + bouton Reessayer (Z2-AC11)
    |
    v
L'eleve peut lancer le diagnostic sur les items disponibles
Le chapitre N'EST PAS bloque
```

---

## 5. Flow detaille ecran par ecran

Voir le document dedie : [`docs/ux/screen-inventory.md`](screen-inventory.md)

**Resume** : 20 ecrans au total.
- 8 existants : Dashboard, Capture, Settings, Onboarding, Processing, Chapter detail, Session, Debrief
- 12 a creer : Validation HITL, Recovery OCR, Creation exam, Emploi du temps, Invitation parent, Dashboard parent, Onboarding parent, Angles morts, Fiche PDF, Report papier, Cloture soiree, Controle blanc

---

## 6. Modele d'etats par ecran critique

Voir le document dedie : [`docs/ux/state-model.md`](state-model.md)

11 etats couverts pour 8 ecrans critiques : Default, Loading, Processing, Success, Empty, Partial, Stale, OCR Uncertainty, Error, Blocked, Retry.

---

## 7. Edge cases et chemins de recuperation

Voir le document dedie : [`docs/ux/edge-cases.md`](edge-cases.md)

30 edge cases repartis en 6 categories : Capture photo (6), Session de revision (7), Validation HITL (4), Adoption et engagement (6), Connectivite (3), Cas limites techniques (4).

---

## 8. Principes UX pedagogiques

### Principe 1 : Rappel actif > relecture

Toute session genere des questions. L'eleve ne relit jamais passivement. Meme le mode fallback (LLM indisponible, Z4-AC12) utilise des flashcards avec "Je savais / Je ne savais pas".

**Implication design** : pas d'ecran de "lecture de la lecon" comme activite principale. Le bouton "Voir ma lecon" (Z1-AC21) est contextuel (pendant une question), pas un mode de revision.

### Principe 2 : Repetition espacee invisible

L'algorithme SRS (UNKNOWN -> FRAGILE -> OK -> SOLID avec next_due_at) est entierement transparent pour l'eleve. Il ne voit pas les intervalles, les dates, les algorithmes. Il voit : "Ce soir : 3 activites, ~15 min."

**Implication design** : jamais montrer next_due_at a l'eleve. Le dashboard soiree presente un plan, pas un calendrier de revision.

### Principe 3 : Espacement obligatoire pour SOLID

La transition OK -> SOLID exige 2 reussites espacees de 24h minimum (Z1-AC03/AC04). L'eleve ne peut pas "grinder" un item en SOLID en une seule session.

**Implication design** : message explicatif quand l'espacement bloque la progression : "Bonne reponse ! Reviens demain pour verrouiller ce point -- ton cerveau a besoin d'une nuit pour bien memoriser."

### Principe 4 : Double codage (Paivio)

Les blocs visuels du cahier (schemas, graphiques, tableaux) sont preserves et reutilises dans les questions (Z2-AC15, Z4-AC17). L'eleve voit le schema de son cahier dans le quiz.

**Implication design** : les questions visuelles affichent l'image au-dessus du texte, avec zoom plein ecran (Z2-AC18). Placeholder aux dimensions exactes pour eviter le reflow.

### Principe 5 : Growth mindset dans tous les messages

Aucun message ne dit "echec", "regression", "tu as perdu". Les regressions sont framed comme "refresh", "a revoir", "pas encore acquis" (Z1-AC05, Z1-AC06, Z1-AC26).

**Implication design** : fichier de strings centralise avec review systematique du framing. Chaque message negatif doit etre reformule en action future positive.

### Principe 6 : Pas de streak, pas de gamification punitive

L'app ne compte pas les jours consecutifs (Z6-AC15). Pas de mecanique de type "tu as perdu ta serie de 7 jours". La seule recompense est la progression de maitrise (UNKNOWN -> SOLID).

**Implication design** : les micro-celebrations (Z1-AC14) portent sur les transitions de maitrise, pas sur la regularite. Le debrief valorise ce qui a progresse, pas combien de jours d'affilee.

---

## 9. Principes d'equilibre parent/eleve

### Ce que le parent DOIT voir

- Maitrise globale par chapitre (% items OK+SOLID)
- Frequence de revision (sessions/semaine, temps total)
- Alertes risque reel : inactivite 3+ jours, controle dans 3 jours avec items FRAGILE
- Qualite du signal : score_confidence, items en verification

### Ce que le parent NE DOIT PAS dominer

- Le detail question par question (l'eleve a droit a l'intimite de ses erreurs)
- Le temps exact de chaque session (le parent n'est pas un chrono)
- Les regressions individuelles (le parent voit les tendances, pas les micro-echecs)
- L'emploi du temps modifie (l'eleve gere, le parent est notifie si opt-in)

### Comment l'eleve percoit la presence du parent

- **Invisible au quotidien** : l'eleve ne voit JAMAIS "ton parent a ete prevenu" (Z6-AC17 note). La notification parent est unidirectionnelle.
- **Presente a 2 moments** : (1) l'invitation initiale apres la premiere session, (2) les ValidationTasks si le parent est en mode actif (max 3/semaine).
- **Percue comme un allie, pas un flic** : le script 3 minutes est un moment de discussion parent-enfant autour du cours, pas un interrogatoire.

### Design principles pour l'intergenerationnel

1. **Pas enfantin** : pas de mascottes, pas d'etoiles, pas de confettis excessifs. Les micro-celebrations sont des messages textuels + animations subtiles.
2. **Pas froid** : le tutoiement (eleve), les messages personalises avec prenom, le ton encourageant.
3. **Pas bureaucratique** : le parent voit des barres de couleur et des % lisibles, pas des tableaux Excel.
4. **Deux niveaux de lecture** : le meme score "72% maitrise" est lisible par l'eleve (barre de couleur) et par le parent (% numerique + label "Compris, a consolider").

---

## 10. Implications pour l'implementation React Native

### Liste des ecrans et routes Expo Router

Voir [`docs/ux/screen-inventory.md`](screen-inventory.md) pour la liste complete.

### Logique de navigation

```
Root (_layout.tsx)
├── (tabs)/
│   ├── index.tsx          -- Dashboard eleve
│   ├── capture.tsx        -- Capture photo
│   └── settings.tsx       -- Parametres
├── onboarding.tsx         -- Onboarding eleve
├── processing.tsx         -- Pipeline OCR
├── recovery.tsx           -- Recovery OCR       (a creer)
├── chapter/[id].tsx       -- Detail chapitre
├── session/[id].tsx       -- Session revision
├── validation/[id].tsx    -- Validation HITL    (a creer)
├── exam/create.tsx        -- Creation exam      (a creer)
├── schedule.tsx           -- Emploi du temps    (a creer)
├── invite-parent.tsx      -- Invitation parent  (a creer)
├── blind-spots.tsx        -- Angles morts       (a creer)
├── paper-report/[id].tsx  -- Report papier      (a creer)
└── parent/                                      (a creer)
    ├── dashboard.tsx      -- Dashboard parent
    └── onboarding.tsx     -- Onboarding parent
```

### Etat partage : Context vs local vs TanStack Query (prevu)

| Donnee | Ou vit l'etat | Justification |
|--------|---------------|---------------|
| Token JWT / auth | React Context (global) | Necessaire partout |
| User profile (prenom, zone) | React Context (global) | Utilise dans les messages personnalises |
| Liste des chapitres | TanStack Query (cache serveur) | Rafraichi par pull-to-refresh, invalide apres upload |
| Masteries | TanStack Query (cache serveur) | Invalide apres chaque session |
| EveningPlan | Etat local du dashboard | Calcule a chaque ouverture, ephemere |
| Session en cours | Etat local de session/[id].tsx | Specifique a un ecran |
| Photos capturees | Etat local de capture.tsx | Pas de persistence entre navigations |
| Reponses pending_sync | AsyncStorage (persistance locale) | Survit au kill de l'app |

### Patterns reutilisables

| Composant | Usage | Props cles |
|-----------|-------|-----------|
| `EmptyState` | Dashboard vide, 0 validations, 0 items | `icon`, `title`, `subtitle`, `ctaLabel`, `onCta` |
| `ErrorState` | Erreur reseau sur tout ecran | `message`, `onRetry` |
| `LoadingSkeleton` | Placeholder pendant chargement | `variant` (card, list, bar) |
| `MasteryBar` | Barre de maitrise (existant) | `breakdown`, `height`, `showLabel` |
| `MasteryBadge` | Badge etat d'un item (existant) | `state` |
| `FeedbackCard` | Feedback post-reponse | `isCorrect`, `expected`, `missing`, `hint` |
| `MicroCelebration` | Animation transition positive | `fromState`, `toState`, `term` |
| `ProgressSteps` | Checklist etapes (onboarding, plan soiree) | `steps[]`, `currentIndex` |
| `ProcessingPhase` | Phase 1/2/3 du pipeline | `phase`, `progress` |

### Types TypeScript (props majeurs)

```typescript
// EveningPlan
type EveningStep = {
  type: 'capture' | 'evening_first' | 'daily' | 'pre_class';
  subject_label: string;
  estimated_duration_min: number;
  status: 'pending' | 'in_progress' | 'completed' | 'skipped';
  session_id?: string;
};

type EveningPlan = {
  steps: EveningStep[];
  total_duration_min: number;
  mode: 'full' | 'express';
  completed_at?: string;
};

// Feedback
type FeedbackData = {
  is_correct: boolean;
  expected_answer: string;
  missing_elements?: string[];
  hint: string;
  transition?: {
    from_state: MasteryState;
    to_state: MasteryState;
    term: string;
  };
};

// Validation
type ValidationAction = 'confirm' | 'correct' | 'skip' | 'ignore';

type ValidationTask = {
  id: string;
  item_id: string;
  suggestion: string;
  source: 'fidelity_check' | 'coherence_check' | 'student_report' | 'anomaly_detection';
  status: 'PENDING' | 'RESOLVED_CONFIRMED' | 'RESOLVED_CORRECTED' | 'DEFERRED_BY_STUDENT' | 'IGNORED_BY_STUDENT';
};

// Parent dashboard
type ParentChapterSummary = {
  id: string;
  name: string;
  subject: string;
  mastery_pct: number;           // % items OK+SOLID
  validation_pending_count: number;
  mastery_label: string;         // "Bien acquis", "En cours d'apprentissage"
};

type ParentDigest = {
  sessions_count: number;
  total_revision_min: number;
  chapters: ParentChapterSummary[];
  alerts: string[];
  next_action: string;
};
```

---

## 11. Decisions produit ouvertes

Voir le document dedie : [`docs/ux/open-decisions.md`](open-decisions.md)

11 decisions identifiees, dont 3 bloquantes pour le Lot 0 P2 :
1. App unique vs separee pour le parent
2. Auto-evaluation vs scoring automatique MCQ
3. Position du bouton "Ajouter des pages"

---

## 12. Recommandations UX finales

### R1 : Prioriser le "Time to First Quiz" (TTFQ)

Le chapitre demo (Z8-AC01) est le meilleur investissement UX du Lot 0. Un eleve qui fait son premier quiz en 3 minutes a 5x plus de chances de revenir le lendemain. Le TTFQ doit etre mesure et optimise comme metrique produit primaire.

### R2 : Le plan de soiree est le produit

Le dashboard soiree (Z7-AC02) n'est pas une feature -- c'est l'interface principale. L'eleve ne devrait jamais avoir a decider quoi faire. Le plan decide pour lui. L'estimation de duree et l'ecran de cloture sont les deux extremites du "contrat" avec l'eleve.

### R3 : Investir dans le feedback enrichi (Z4-AC09) avant tout le reste

Un feedback qui dit "Faux -- la bonne reponse est X" n'enseigne rien. Le feedback enrichi (reponse correcte + ce qui manquait + indice 1 phrase) est le micro-moment d'apprentissage le plus puissant de chaque session. Il doit etre templatise (pas LLM en temps reel) pour garantir la rapidite (< 200 ms).

### R4 : Tester le mode express (Z7-AC08) des la premiere semaine

Le mode "Je n'ai pas beaucoup de temps" est un anti-churn majeur. 5 minutes de revision ciblee valent infiniment mieux que 0 minute. Si l'eleve ne voit pas ce bouton, il saute la revision entierement les soirs charges.

### R5 : La notification "routine terminee" est le killer feature parent

Le parent qui recoit "Theo a termine sa revision du soir : 3 matieres, 14 min" chaque soir n'a pas besoin d'ouvrir l'app. C'est le signal de confiance maximal avec le cout minimal. Implementer Z7-AC13 le plus tot possible.

### R6 : Ne jamais montrer un ecran vide

Chaque ecran a un empty state explicite avec un CTA. Le dashboard vide a le chapitre demo. La session sans items a le message "Bravo, tu es a jour". Le dashboard parent sans activite a "Vous serez notifie des la premiere session". Aucun ecran ne doit laisser l'utilisateur face a du vide sans action possible.

### R7 : Mesurer le taux d'echec du premier OCR

L'echec du premier OCR (Z8-AC03) est le point de churn #1 du funnel. Logger chaque `j0_first_upload_recovery` et monitorer le taux. Si > 20% des premiers uploads echouent, c'est un probleme systeme, pas un probleme utilisateur.

### R8 : Implementer la persistance locale des reponses (Z6-AC31) tot

Un collegien utilise l'app dans le bus, en zone blanche, avec du wifi instable. Une deconnexion de 10 secondes ne doit jamais perdre une reponse. La persistance optimiste cote client est un prerequis pour toute session fiable en mobilite.
