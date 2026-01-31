# Idea Honing - Requirements Clarification

Ce document capture les questions et réponses du processus de clarification des exigences.

---

## Q1: Scope des fonctionnalités P0

Le PRD définit les fonctionnalités P0 suivantes :
- OCR de cours (scan/photo)
- Génération de fiches de révision
- Quiz adaptatifs (QCM 10-15 questions)
- Enrichissement de contenu (recherche de ressources en ligne)
- Tableau de bord simple

Pour ce MVP local sans authentification, **lesquelles voulez-vous inclure ?**

Options suggérées :
- A) Toutes les P0 (OCR + fiches + quiz + enrichissement + dashboard)
- B) Core uniquement (OCR + fiches + quiz + dashboard, sans enrichissement)
- C) Minimal (OCR + fiches uniquement)
- D) Autre combinaison

**Réponse :** A) Toutes les P0 (OCR + fiches + quiz + enrichissement + dashboard)

---

## Q2: Source pour l'enrichissement de contenu

L'enrichissement de contenu consiste à trouver des ressources complémentaires en ligne (vidéos, articles, schémas) pour compléter le cours scanné.

**Quelle approche préférez-vous pour le MVP ?**

Options :
- A) Recherche web automatique (via API de recherche type Brave/Google)
- B) Génération par le LLM uniquement (suggère des ressources connues sans vérifier leur existence)
- C) Liens statiques par matière (base de données de ressources éducatives pré-sélectionnées)
- D) Reporter cette fonctionnalité au post-MVP

**Réponse :** B) Génération par le LLM uniquement (suggère des ressources connues sans vérifier leur existence)

Note : Le LLM suggérera des ressources éducatives (vidéos YouTube éducatives, sites comme Khan Academy, etc.) basées sur le sujet du cours, avec un avertissement que les liens doivent être vérifiés.

---

## Q3: Fournisseur LLM/OCR

Le backend utilise un pattern d'adaptateurs pour permettre de changer de fournisseur. Pour le MVP :

**Quel fournisseur LLM/OCR voulez-vous implémenter en premier ?**

Options :
- A) OpenAI uniquement (GPT-4o pour OCR + génération) - le plus mature, coût modéré
- B) OpenAI + Mistral (fallback) - plus résilient, Mistral moins cher pour la génération
- C) Mistral uniquement (Pixtral pour OCR + Mistral Large pour génération) - moins cher, européen
- D) Autre (préciser)

**Réponse :** B) OpenAI + Mistral (fallback)

Configuration prévue :
- OCR : OpenAI GPT-4o (primary), Mistral Pixtral (fallback)
- Génération de contenu : OpenAI GPT-4o (primary), Mistral Large (fallback)
- Le fallback s'active en cas d'erreur API ou de rate limiting

---

## Q4: Stockage des données

Pour le MVP local sans authentification, comment voulez-vous gérer le stockage ?

**Options :**
- A) PostgreSQL complet (courses, cards, quizzes, sessions) - prêt pour la prod
- B) PostgreSQL minimal (courses + contenu généré uniquement) - pas de tracking de progression
- C) SQLite (fichier local) - plus simple, pas de container DB séparé
- D) En mémoire uniquement (pas de persistance entre redémarrages)

**Réponse :** A) PostgreSQL complet (courses, cards, quizzes, sessions)

Schéma prévu :
- `courses` : cours scannés avec texte OCR
- `cards` : fiches de révision générées
- `quizzes` : quiz avec questions/réponses
- `quiz_sessions` : historique des tentatives
- `resources` : ressources complémentaires suggérées

---

## Q5: Intégration Frontend-Backend

Le frontend actuel est en HTML/CSS/JS statique avec des données mockées. Pour le MVP :

**Comment voulez-vous connecter le frontend au backend ?**

Options :
- A) Garder HTML statique + appels fetch() vers l'API REST
- B) Migrer vers un framework JS (React/Vue/Svelte) pour une meilleure gestion d'état
- C) Server-side rendering avec Go templates (HTML généré côté serveur)
- D) Autre approche

**Réponse :** D) Migrer vers React

Plan de migration :
- Convertir les pages HTML existantes en composants React
- Conserver le design system actuel (CSS variables, Fraunces/DM Sans)
- Utiliser React Router pour la navigation
- État global pour les données (React Context ou Zustand)
- Appels API via fetch ou React Query

---

## Q6: Setup React

Pour le projet React, quelle configuration préférez-vous ?

**Options :**
- A) Vite + React (rapide, léger, moderne)
- B) Next.js (SSR/SSG, routing intégré, plus lourd)
- C) Create React App (classique mais moins maintenu)
- D) Autre (préciser)

**Réponse :** A) Vite + React

Stack frontend :
- Vite comme bundler (dev rapide, HMR)
- React 18+ avec TypeScript
- React Router pour la navigation
- CSS Modules ou styled-components pour le styling (à confirmer)

---

## Q7: Gestion du styling React

Le design actuel utilise des CSS variables et du CSS inline dans chaque fichier HTML. Pour React :

**Comment voulez-vous gérer le styling ?**

Options :
- A) CSS Modules (fichiers .module.css par composant, isolation native)
- B) Tailwind CSS (utility-first, classes dans le JSX)
- C) Styled-components / Emotion (CSS-in-JS)
- D) CSS global + variables (garder l'approche actuelle)

**Réponse :** B) Tailwind CSS

Configuration Tailwind :
- Configurer les couleurs du design system (coral, teal, gold, cream)
- Configurer les fonts (Fraunces, DM Sans)
- Utiliser les espacements et rayons existants
- Plugin @tailwindcss/forms pour les inputs

---

## Q8: Formats d'upload acceptés

Pour l'OCR, le PRD mentionne "scan ou photo". Quels formats voulez-vous supporter dans le MVP ?

**Options :**
- A) Images uniquement (JPG, PNG, WebP)
- B) Images + PDF (avec extraction page par page)
- C) Images + PDF + capture caméra directe (via webcam/mobile)
- D) Autre

**Réponse :** B) Images + PDF (avec extraction page par page)

Implémentation :
- Images : JPG, PNG, WebP (envoi direct à l'API Vision)
- PDF : extraction des pages en images côté backend (lib Go type pdfcpu ou appel à un service)
- Limite : 10 pages max par upload (comme indiqué dans l'UI actuelle)

---

## Q9: Gestion des erreurs OCR

Le PRD mentionne une "possibilité de correction manuelle" si l'OCR échoue. Pour le MVP :

**Quel niveau de gestion d'erreur voulez-vous ?**

Options :
- A) Basique : afficher le texte OCR, permettre l'édition avant génération
- B) Avancé : score de confiance par zone, highlighting des parties incertaines
- C) Minimal : pas de correction, régénérer si problème
- D) Autre

**Réponse :** B) Avancé : score de confiance par zone, highlighting des parties incertaines

Implémentation :
- Demander au LLM d'indiquer les zones de faible confiance dans sa réponse
- Interface d'édition avec highlighting des parties incertaines (fond jaune/orange)
- Permettre la correction manuelle avant de lancer la génération
- Stocker le texte corrigé comme référence

---

## Q10: Personnalisation des quiz

Le PRD mentionne "QCM de 10-15 questions". Pour le MVP :

**Quel niveau de personnalisation pour les quiz ?**

Options :
- A) Fixe : toujours 10 questions, difficulté moyenne
- B) Configurable : nombre de questions (5-20), niveau de difficulté
- C) Adaptatif : ajuste la difficulté en temps réel selon les réponses
- D) Autre

**Réponse :** B) Configurable : nombre de questions (5-20), niveau de difficulté

Options de configuration :
- Nombre de questions : 5, 10, 15, 20
- Difficulté : Facile, Moyen, Difficile
- Type : QCM uniquement pour le MVP (vrai/faux et réponse courte en post-MVP)

---

## Q11: Dashboard - Données affichées

Le dashboard actuel montre des statistiques mockées (cours scannés, quiz complétés, fiches créées, score moyen).

**Sans authentification, quelles données voulez-vous afficher ?**

Options :
- A) Statistiques globales de l'instance locale (tous les cours/quiz de cette installation)
- B) Session-based : stats de la session navigateur actuelle uniquement
- C) Simplifié : juste la liste des cours récents, pas de stats agrégées
- D) Autre

**Réponse :** A) Statistiques globales de l'instance locale

Le dashboard affichera :
- Nombre total de cours scannés
- Nombre total de quiz complétés
- Nombre total de fiches créées
- Score moyen sur tous les quiz
- Liste des cours récents avec progression
- Activité récente (dernières actions)

---

## Q12: Mindmaps (Cartes mentales)

Le PRD classe les mindmaps en P1 (post-MVP), mais l'UI scan.html propose déjà l'option "Carte mentale".

**Voulez-vous inclure les mindmaps dans ce MVP ?**

Options :
- A) Oui, génération basique (structure JSON convertie en visualisation simple)
- B) Oui, avec une lib de visualisation interactive (type D3.js ou react-flow)
- C) Non, désactiver l'option dans l'UI pour le MVP
- D) Autre

**Réponse :** B) Oui, avec une lib de visualisation interactive (type react-flow)

Implémentation :
- LLM génère une structure de mindmap en JSON (nœuds + liens)
- Frontend utilise react-flow ou similar pour le rendu interactif
- Fonctionnalités : zoom, pan, collapse/expand des branches
- Export image possible (PNG/SVG)

---

## Q13: Gestion des coûts API

Les appels OpenAI/Mistral ont un coût. Pour un MVP local :

**Comment voulez-vous gérer les coûts API ?**

Options :
- A) Pas de limite (l'utilisateur gère sa clé API)
- B) Afficher une estimation du coût avant chaque opération
- C) Limites configurables (max pages/jour, max générations/jour)
- D) Cache agressif (réutiliser les générations identiques)

**Réponse :** C) Limites configurables (max pages/jour, max générations/jour)

Configuration via variables d'environnement :
- `MAX_OCR_PAGES_PER_DAY` : limite de pages OCR par jour
- `MAX_GENERATIONS_PER_DAY` : limite de générations (fiches/quiz/mindmaps)
- Afficher le quota restant dans l'UI
- Message d'erreur clair quand la limite est atteinte

---

## Q14: Accessibilité

Le PRD mentionne : "taille de texte ajustable, mode daltonien, contraste élevé".

**Quel niveau d'accessibilité pour le MVP ?**

Options :
- A) Complet : taille texte, mode daltonien, contraste élevé, navigation clavier
- B) Partiel : taille texte ajustable + bon contraste de base
- C) Minimal : respect des standards HTML sémantique, pas de features spécifiques
- D) Autre

**Réponse :** A) Complet : taille texte, mode daltonien, contraste élevé, navigation clavier

Fonctionnalités d'accessibilité :
- Bouton pour ajuster la taille du texte (3 niveaux)
- Mode daltonien (palette alternative)
- Mode contraste élevé
- Navigation clavier complète (focus visible, tab order)
- Attributs ARIA appropriés
- Respect WCAG 2.1 niveau AA

---

## Q15: Tests et qualité

Pour garantir la qualité du MVP :

**Quelle stratégie de tests voulez-vous adopter ?**

Options :
- A) Tests complets : unit tests (Go + React), integration tests API, E2E (Playwright/Cypress)
- B) Tests backend focus : unit tests Go + integration tests API
- C) Tests critiques : E2E sur les parcours principaux uniquement
- D) Minimal : tests manuels, pas de tests automatisés pour le MVP

**Réponse :** A) Tests complets : unit tests (Go + React), integration tests API, E2E (Playwright/Cypress)

Stack de tests :
- **Backend** : Go testing package + testify pour les assertions
- **Frontend** : Vitest + React Testing Library pour les composants
- **E2E** : Playwright pour les parcours utilisateur complets
- **CI** : GitHub Actions pour exécuter les tests automatiquement

---

## Q16: Langue de l'interface

Le PRD indique "All user-facing content is in French". Pour le code et la documentation technique :

**Quelle langue pour le code et les messages techniques ?**

Options :
- A) Tout en français (code, commentaires, commits, docs)
- B) Code en anglais, UI en français (standard industrie)
- C) Mixte : UI français, code anglais, docs français
- D) Autre

**Réponse :** A) Tout en français (code, commentaires, commits, docs)

Convention :
- Noms de fonctions/variables en français : `creerCours()`, `genererFiches()`
- Commentaires en français
- Messages de commit en français
- Documentation technique en français
- UI entièrement en français

---

## Q17: Priorité de développement

Si le temps est limité, quel ordre de priorité pour les fonctionnalités ?

**Classez de 1 (plus prioritaire) à 5 (moins prioritaire) :**

Options à classer :
- OCR + affichage du texte extrait
- Génération de fiches de révision
- Génération de quiz
- Génération de mindmaps
- Enrichissement de contenu (ressources)

**Réponse (ordre) :** 1, 2, 3, 5, 4

Ordre de priorité :
1. **OCR + affichage du texte extrait** - Fondation de tout le système
2. **Génération de fiches de révision** - Fonctionnalité principale
3. **Génération de quiz** - Engagement utilisateur
4. **Enrichissement de contenu** - Valeur ajoutée
5. **Génération de mindmaps** - Fonctionnalité bonus

---

## Q18: Router Go

**Question :** Quel router Go préférez-vous ?

Options considérées :
- chi (léger, idiomatique)
- Gin (rapide, binding/validation intégrés)
- Echo, Fiber, gorilla/mux

**Réponse :** Gin - pour le binding JSON automatique, la validation intégrée, et la large communauté.

---

