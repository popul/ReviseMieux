# Prompt 2 — User Flow complet pour app React Native (élève + parent)

> **Objectif** : Produire un user flow détaillé, screen-by-screen, couvrant tous les états, edge cases, et micro-interactions pour une app éducative à double audience (collégiens + parents).

---

You are a principal product designer and senior UX architect specialized in educational mobile apps, multi-audience products, and onboarding/adoption design.

## Mission

Help me design a complete user flow for a mobile application called "Révise Mieux".

## Contexte produit

Révise Mieux est une app mobile de révision destinée principalement aux collégiens français (11-15 ans), avec une interface parent intégrée.

### Stack mobile

| Composant | Technologie |
|-----------|-------------|
| Framework | React Native 0.81 + Expo 54 |
| Routing | Expo Router (file-based) |
| Langage | TypeScript strict |
| Cache serveur | État local (TanStack Query prévu mais non encore installé) |
| État global | React Context (auth uniquement), sinon état local + props |
| State management | Pas de Redux |

### Écrans Expo Router déjà implémentés (Lot 0)

| Route | Écran |
|-------|-------|
| `app/(tabs)/index.tsx` | Dashboard élève |
| `app/(tabs)/capture.tsx` | Capture photo |
| `app/(tabs)/settings.tsx` | Paramètres |
| `app/onboarding.tsx` | Onboarding initial |
| `app/processing.tsx` | Progression OCR |
| `app/chapter/[id].tsx` | Détail chapitre + leçon |
| `app/session/[id].tsx` | Session révision |

### Contrainte d'adoption critique

L'app doit être **adoptée par les deux publics** simultanément :
- **L'élève** décide s'il ouvre l'app chaque soir → le design doit donner envie
- **Le parent** décide de payer et d'installer → le design doit inspirer confiance

Ce défi de double adoption est central et non négociable.

### Fonctionnalités principales

- Capture de photos de cahier → pipeline OCR/IDP → items structurés
- Sessions de révision adaptatives (questions générées, scoring, feedback)
- Suivi de progression par chapitre (Mastery : UNKNOWN → FRAGILE → OK → SOLID)
- Dashboard parent (maîtrise, alertes, digest hebdo)
- Validation HITL (parent valide les items générés par le LLM)

### Documents de référence à consulter

| Document | Chemin | Contenu pertinent |
|----------|--------|-------------------|
| PRD | `docs/PRD.md` | Personas (§4), parcours utilisateurs (§5), fonctionnalités (§15) |
| ACs complètes | `docs/ac/Z1.md` à `Z8.md` | 171 critères d'acceptation Given/When/Then |
| Scope MVP | `docs/MVP-scope.md` | 53 ACs Lot 0 (33 P1 + 20 P2) |
| User journeys | `docs/user-journeys/first-connection.md` | Première connexion |
| API spec | `docs/openapi.yaml` | Endpoints disponibles |
| Code mobile existant | `mobile/` | Écrans déjà implémentés |

### Chapitres pilotes (Lot 0)

| Matière | Chapitre | Pack |
|---------|----------|------|
| Histoire-Géo | Les inégalités dans le monde | HG-INEG |
| Histoire-Géo | La société féodale | HG-FEOD |
| SVT | La photosynthèse | SVT-PHOTO |
| Physique-Chimie | Masse, volume et densité | PC-MVD |

---

## Tes tâches

### 1. Recadrer le défi produit

Commence par expliquer le problème UX avec précision :
- Pourquoi ce n'est PAS une app standard à audience unique ?
- Quelles tensions existent entre les attentes des élèves et celles des parents ?
- Quels risques d'adoption existent des deux côtés ?
- Qu'est-ce qui fait qu'un collégien de 13 ans revient chaque soir sur une app de révision ?
- Qu'est-ce qui fait qu'un parent de 45 ans fait confiance à l'outil ?

### 2. Définir les personas

Crée des personas réalistes pour au minimum :

**Théo (13 ans, 4e)** — élève motivé mais irrégulier
- Veut réviser rapidement (10-20 min/soir)
- A besoin de savoir "quoi faire aujourd'hui"
- Se décourage si c'est compliqué ou long
- N'aime pas les apps qui "font bébé"

**Inès (11 ans, 6e)** — élève appliquée mais anxieuse
- Veut bien faire mais doute d'elle
- A besoin de validation et encouragement
- Stresse avant les contrôles

**Lucas (15 ans, 3e)** — élève désengagé
- Ne voit pas l'intérêt de réviser
- Utilise l'app parce que ses parents l'ont installée
- Décrochera au moindre friction

**Laurent (45 ans, père de Théo)** — parent cherchant de la réassurance
- Veut un signal fiable sur la maîtrise sans micro-manager
- Veut être alerté uniquement si risque réel
- Setup initial avec Théo, puis consultation ponctuelle

Pour chaque persona, décris :
- objectifs
- frustrations
- déclencheurs de confiance
- risques de décrochage
- état émotionnel en ouvrant l'app

### 3. Définir les parcours produit clés

Mappe les journeys end-to-end :
1. **Onboarding** : première connexion, création compte parent + enfant
2. **Capture J0** : photographier un cours → pipeline → items générés
3. **Dashboard élève** : plan d'action quotidien, progression par chapitre
4. **Session de révision** : questions adaptatives, feedback immédiat, scoring
5. **Revue des erreurs** : comprendre ce qui a été raté, refaire
6. **Retour après absence** : reprendre après 3 jours sans ouvrir l'app
7. **Dashboard parent** : comprendre ce que fait l'app, vérifier l'utilité
8. **Validation HITL** : parent valide les items générés par le LLM
9. **Gestion de contenu OCR incomplet** : pages floues, extraction partielle

### 4. Détailler chaque parcours écran par écran

Pour chaque écran de chaque parcours, spécifie :

| Champ | Description |
|-------|-------------|
| **Nom et route** | Expo Router file-based (ex: `/(tabs)/dashboard`) |
| **Intention utilisateur** | Le "job to be done" sur cet écran |
| **Blocs UI requis** | Header, contenu, actions, navigation |
| **CTA principal** | L'action principale attendue |
| **Actions secondaires** | Autres interactions possibles |
| **Compréhension immédiate** | Ce que l'utilisateur doit comprendre en < 3 secondes |
| **Ton émotionnel** | L'ambiance que l'écran doit créer |
| **Données nécessaires** | API calls, états TanStack Query, paramètres de route |
| **Mapping ACs** | Quels Z*-AC* cet écran couvre |

### 5. Définir les états d'affichage exhaustivement

Pour chaque écran critique, décrire EXPLICITEMENT ces états (pas une mention générique) :

| État | Ce que l'utilisateur voit | Comportement technique |
|------|--------------------------|----------------------|
| **Default** | Écran au repos, données fraîches | Render normal |
| **Loading** | Skeleton, spinner, ou placeholder | TanStack Query `isLoading` |
| **Processing** | Pipeline OCR en cours | Polling ou WebSocket |
| **Success** | Confirmation d'action | Feedback visuel + haptic |
| **Empty** | Aucune donnée (premier usage) | Message d'encouragement + CTA |
| **Partial** | Données partielles (offline, timeout) | Ce qui est disponible + indicateur |
| **Stale** | Données en cache potentiellement obsolètes | Indicateur discret + background refresh |
| **OCR uncertainty** | Extraction incertaine | Badges confiance + option review |
| **Error** | Erreur réseau ou serveur | Message clair + retry |
| **Blocked** | Nécessite validation (HITL) | Explication + redirection |
| **Retry** | Après échec, nouvelle tentative | Bouton retry + contexte |

### 6. Intégrer l'intention pédagogique

Pour chaque étape majeure, explique :
- ce que l'élève est censé **gagner** (pas juste "voir un écran")
- comment l'app **réduit l'effort** ou l'anxiété
- comment elle **augmente** la clarté, la confiance ou la motivation
- comment l'expérience reste **utile** sans devenir une surcharge de devoirs

Principes pédagogiques non négociables (du PRD) :
- Rappel actif > relecture — toute session génère des questions
- Répétition espacée — SOLID uniquement après 2 réussites espacées de 24h
- Révision proactive dès J0 — pas besoin d'attendre un contrôle
- Double codage (Paivio) — verbal + visuel
- Human-in-the-loop — validation des zones incertaines

### 7. Intégrer la réassurance parent SANS nuire à l'adoption élève

C'est un point critique. Explique comment l'app peut :
- paraître **assez sérieuse** pour les parents
- rester **assez attractive** pour les élèves
- éviter de paraître enfantine
- éviter de paraître froide ou bureaucratique
- créer de la confiance sans devenir un outil de surveillance

Sois explicite sur :
- ce que les parents **doivent** voir
- ce qu'ils **ne doivent pas** dominer
- comment l'élève perçoit la présence du parent dans l'app

### 8. Explorer les edge cases en profondeur

Pour chaque edge case, décris : risque UX, traitement idéal, stratégie de copy/guidance, chemin de récupération.

**Capture photo** :
- Photo floue → détection et message avant upload
- Photo de travers → proposition de recadrage
- Plusieurs pages en une photo → avertissement
- Pas de connexion → mise en file d'attente locale
- Pipeline J0 échoue ou timeout → état "en cours de traitement" avec retry

**Session de révision** :
- L'élève ferme l'app en pleine session → reprise automatique
- Pas assez d'items pour composer une session → message adapté
- Tous les items sont SOLID → félicitation + suggestion (réviser un autre chapitre)
- Réponse ambiguë (scoring ≈ seuil 0.7) → feedback nuancé
- Quiz trop dur → indice ou simplification
- Quiz trop facile → escalade de difficulté

**Validation HITL** :
- Le parent n'a jamais validé → rappel non-intrusif
- Item rejeté par le parent → impact visible sur le chapitre
- Parent ne comprend pas un item → option "signaler comme incompréhensible"

**Adoption** :
- Élève abandonne en plein flow → quels points de friction ?
- Parent ne comprend pas la proposition de valeur → onboarding parent dédié ?
- Élève se sent jugé par le scoring → comment formuler l'échec ?
- Aucun contenu exploitable détecté → que montrer ?
- Contenu appartient à plusieurs cours → demander clarification

**Connectivité** :
- Perte de connexion pendant une session → continuer en local, sync au retour
- Perte de connexion pendant un upload → retry automatique avec indicateur
- Mode avion prolongé → ce qui reste fonctionnel

### 9. Définir les micro-interactions et principes de feedback

Décris les micro-feedbacks appropriés pour des 11-15 ans tout en restant acceptable pour les parents :
- Confirmations (haptic, animation)
- Indicateurs de progression
- Encouragements (sans infantiliser)
- Avertissements d'incertitude
- Guidance pédagogique
- Récompenses de complétion (dosées)

### 10. Préparer le flow pour l'implémentation React Native

Termine avec une structure directement exploitable :

| Livrable | Description |
|----------|-------------|
| **Liste des écrans** | Nom, route Expo Router, composant |
| **Logique de navigation** | Stack, tabs, modals, deep links |
| **État partagé** | Ce qui vit dans Context vs local vs TanStack Query |
| **Patterns réutilisables** | Composants d'état (Empty, Error, Loading), feedback, etc. |
| **Types TypeScript** | Props et état local de chaque composant majeur |
| **Points d'arbitrage produit** | Décisions à prendre avant de coder |

---

## Format de sortie attendu

1. **Executive summary**
2. **Défi UX produit** (double audience)
3. **Personas** (4 minimum, avec états émotionnels)
4. **Parcours majeurs** (9 journeys)
5. **Flow détaillé écran par écran** (wireframe textuel + tableau par écran)
6. **Modèle d'états par écran critique** (tableau 11 états)
7. **Edge cases et chemins de récupération** (tableau)
8. **Principes UX pédagogiques**
9. **Principes d'équilibre parent/élève**
10. **Implications pour l'implémentation React Native**
11. **Décisions produit ouvertes à résoudre**
12. **Recommandations UX finales**

## Barre de qualité

Ta réponse doit être :
- **concrète** — au niveau de l'écran, pas du parcours abstrait
- **multi-audience** — chaque décision pesée pour élève ET parent
- **pédagogiquement informée** — pas juste du UX générique
- **implémentable** — exploitable par un dev React Native

Ne reste pas générique. Ne décris pas juste un user journey en prose. Produis quelque chose qui peut réellement guider les décisions produit et engineering.
