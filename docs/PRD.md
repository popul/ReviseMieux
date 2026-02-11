# One-pager : Révise mieux - Assistant de Révision Intelligent

## 1. TL;DR

Révise mieux est un service de préparation aux interrogations écrites pour les collégiens et lycéens. À partir d'un cours photographié, l'outil génère automatiquement des fiches de révision, des quiz interactifs et des cartes mentales enrichis par des ressources en ligne. Après le contrôle, l'élève peut également scanner sa copie corrigée pour obtenir une analyse personnalisée de ses erreurs. L'objectif : réviser plus efficacement, en moins de temps, et obtenir de meilleures notes.

## 2. Goals

### Business Goals
* Établir Révise mieux comme la référence en révision intelligente pour les collégiens et lycéens
* Atteindre 50 000 utilisateurs actifs mensuels dans les 12 premiers mois
* Obtenir un taux de satisfaction utilisateur supérieur à 85%
* Générer un taux de conversion freemium-premium de 8-12%
* Créer un effet de réseau via le partage de ressources entre établissements

### User Goals
* Réduire de 40% le temps de préparation aux interrogations écrites
* Améliorer la compréhension des cours grâce à des supports visuels variés
* Identifier rapidement les lacunes et les points à renforcer
* Réviser de manière autonome avec des outils adaptés à leur niveau
* Suivre leur progression et leurs améliorations dans le temps

### Non-Goals
* Rédiger les devoirs ou dissertations à la place des élèves
* Remplacer les professeurs ou les cours en classe
* Couvrir l'enseignement supérieur dans la version initiale (V1)
* Proposer du soutien scolaire en temps réel ou du tutorat
* Gérer la communication parents-professeurs

## 3. User stories

**Persona 1 : Léa, 15 ans, lycéenne en Seconde**
* "Je veux transformer mes notes de cours en fiches synthétiques pour ne pas passer 2 heures à les réécrire moi-même"
* "Je veux m'entraîner avec des quiz pour vérifier que j'ai bien compris avant l'interro"
* "Je veux comprendre mes erreurs après un contrôle raté pour ne pas les refaire"

**Persona 2 : Thomas, 13 ans, collégien en 4ème**
* "Je veux une carte mentale colorée de mon cours d'histoire pour mieux mémoriser les dates"
* "Je veux des exemples supplémentaires que le prof n'a pas donnés pour mieux comprendre les maths"
* "Je veux savoir sur quels chapitres je dois encore travailler"

**Persona 3 : Sarah, 17 ans, lycéenne en Terminale**
* "Je veux réviser efficacement malgré un emploi du temps chargé"
* "Je veux des questions type bac pour m'entraîner dans les conditions réelles"
* "Je veux analyser mes copies pour optimiser ma méthode de révision"

## 4. Functional requirements

### Priorité P0 (MVP - Must have)
* **Plan de révision (concept central)** : l'élève crée un plan pour préparer un examen ou contrôle, y regroupe ses cours avec une date cible. Architecture d'information : **Plan → Cours → Artefacts** (fiches, quiz, mindmap, résumé, concepts, lexique, examen blanc). Le plan offre une vue d'ensemble de la progression et un accès direct à tous les artefacts.
* **OCR de cours** : scan ou photo des notes manuscrites ou imprimées avec reconnaissance de texte
* **Génération de fiches de révision** : synthèse automatique structurée du cours, accessible via le plan de révision
* **Quiz adaptatifs** : QCM de 10-15 questions basées sur le contenu du cours, accessible via le plan de révision
* **Tableau de bord** : vue d'ensemble des plans de révision, cours scannés et quiz passés

### Priorité P1 (Post-MVP - Should have)
* **Génération de mindmaps** : cartes mentales visuelles et interactives du cours
* **OCR de copies corrigées** : scan de l'interro corrigée avec analyse des erreurs
* **Analyse des erreurs** : identification des types d'erreurs (compréhension, méthode, inattention)
* **Recommandations personnalisées** : suggestions de révision ciblées selon les lacunes détectées
* **Historique de progression** : suivi des notes et de l'évolution par matière

### Priorité P2 (Nice to have)
* **Mode collaboratif** : partage de fiches entre élèves d'une même classe
* **Flashcards** : système de répétition espacée pour la mémorisation long terme
* **Export multi-formats** : PDF, impression, partage sur autres apps
* **Mode hors-ligne** : accès aux contenus déjà générés sans connexion
* **Intégration agenda** : rappels de révision avant les dates de contrôle

## 5. User experience

### Parcours principal - Préparation à une interrogation
* L'élève crée un **plan de révision** (titre, matière, date de l'examen)
* Il scanne ses cours et les ajoute au plan
* L'OCR traite le document en 15-30 secondes avec indicateur de progression
* Depuis le plan, l'élève accède à chaque cours avec ses artefacts (fiches, quiz, mindmap, résumé, lexique)
* L'élève choisit : "Générer fiche", "Créer quiz" ou "Créer mindmap" pour chaque cours
* La barre de progression du plan se met à jour au fur et à mesure de la génération des artefacts
* Pour le quiz : feedback immédiat après chaque réponse avec explication
* Score final affiché avec suggestion de révision si < 70%

### Parcours secondaire - Analyse post-contrôle
* L'élève sélectionne "Analyser ma copie corrigée"
* Il photographie sa copie rendue avec les annotations du professeur
* L'OCR identifie les erreurs et les commentaires (30-45 secondes)
* Une synthèse s'affiche : types d'erreurs, points perdus par thème, note moyenne classe
* Des recommandations de révision apparaissent avec liens vers sections spécifiques du cours
* L'élève peut regénérer un quiz ciblé sur les notions mal maîtrisées

### Edge cases et notes UI
* **Qualité photo médiocre** : message invitant à reprendre avec meilleurs éclairage/angle + tips
* **Contenu non reconnu** : option de saisie manuelle partielle ou assistance support
* **Cours trop long** : suggestion de diviser en plusieurs sections thématiques
* **Pas de connexion internet** : mode dégradé sans enrichissement externe, avec notification
* **Matière non identifiée** : menu déroulant pour spécifier manuellement (maths, français, histoire, etc.)
* **Interface accessible** : taille de texte ajustable, mode daltonien, contraste élevé

## 6. Narrative

**Lundi 16h30. Mathilde, élève de Première, sort de son cours d'histoire.**

Demain, interrogation surprise sur la Révolution française annoncée il y a une semaine. Elle a trois heures de sport ce soir et une dissertation de philo à terminer. Pas le temps de réécrire des fiches pendant deux heures comme d'habitude.

Dans le bus, elle ouvre Révise mieux et photographie ses sept pages de notes. En trente secondes, l'application a tout lu, même son écriture de médecin. Elle appuie sur "Générer fiche" et reçoit une synthèse parfaite : chronologie, acteurs clés, enjeux, citations importantes. L'app a même ajouté deux vidéos courtes et un schéma qu'elle n'avait pas dans son cours.

Elle lance ensuite un quiz de quinze questions. Première question ratée sur les Girondins. L'application lui explique immédiatement la différence avec les Montagnards, avec un petit tableau comparatif. Mathilde continue : 12/15, pas mal. Le système lui suggère de revoir la section "Terreur et Comité de Salut Public" avant demain.

Le soir, après le sport, vingt minutes de mindmap interactive sur son téléphone. Les couleurs et les liens visuels l'aident à tout mémoriser sans effort.

**Mercredi, copie rendue : 15/20.**

Mathilde scanne sa copie corrigée. Révise mieux analyse : "Bonne maîtrise chronologique, mais définitions imprécises sur le Tiers-État et confusion sur la Constitution de 1791". L'appli génère un mini-quiz de rattrapage sur ces deux points précis. Cinq minutes plus tard, c'est clair dans sa tête.

Pour la prochaine interro, elle sait exactement quoi réviser.

## 7. Success metrics

### Métriques d'engagement
* Nombre de plans créés par utilisateur (objectif : 2-3 par trimestre)
* Nombre de cours par plan moyen (objectif : 3-5)
* Taux de complétion des plans (% d'artefacts générés) (objectif : > 60%)
* Nombre de cours scannés par utilisateur actif par semaine (objectif : 2-3)
* Taux de complétion des quiz générés (objectif : > 70%)
* Temps moyen passé sur l'application par session de révision (objectif : 15-25 min)
* Taux de retour sur l'app après premier usage (D1, D7, D30) (objectif : 60%, 40%, 25%)

### Métriques de qualité
* Score moyen de satisfaction sur la qualité des fiches générées (objectif : 4,2/5)
* Taux de précision OCR (objectif : > 92% pour texte imprimé, > 85% pour manuscrit)
* Taux de pertinence des ressources complémentaires trouvées (objectif : 80% jugées utiles)
* NPS (Net Promoter Score) global (objectif : > 40)

### Métriques d'impact
* Amélioration moyenne des notes après utilisation régulière (étude sur 3 mois)
* Temps de révision économisé déclaré par les utilisateurs (objectif : -35 à -45%)
* Taux d'utilisation de la fonctionnalité "Analyse de copie" (objectif : 30% des users actifs)
* Taux de conversion vers abonnement premium (objectif : 10%)

### Métriques techniques
* Temps de traitement OCR moyen (objectif : < 3 secondes/page)
* Temps de génération de contenu (fiche/quiz) (objectif : < 20 secondes)
* Taux de disponibilité du service (objectif : > 99,5%)

## 8. Milestones & sequencing

### Phase 1 : MVP (M0-M3) - Validation du concept
* **M0-M1** : Développement OCR cours + génération fiches basiques
* **M1-M2** : Génération de quiz + enrichissement contenu externe
* **M2** : Bêta fermée avec 100 élèves testeurs volontaires (2 établissements pilotes)
* **M3** : Ajustements UX/UI + lancement bêta publique (1000 early adopters)
* **Équipe** : 1 PM, 2 devs full-stack, 1 designer, 1 data scientist (OCR/NLP)

### Phase 2 : Enrichissement (M4-M6) - Différenciation
* **M4** : Génération de mindmaps interactives
* **M5** : OCR de copies corrigées + analyse d'erreurs basique
* **M6** : Tableau de bord progression + recommandations personnalisées
* **M6** : Lancement version 1.0 grand public (objectif : 10K utilisateurs)
* **Équipe** : +1 dev backend, +1 content manager (QA des ressources)

### Phase 3 : Monétisation & Scale (M7-M12) - Croissance
* **M7-M8** : Lancement modèle freemium (fiches illimitées, quiz/mindmaps premium)
* **M9** : Fonctionnalités collaboratives (partage de fiches entre élèves)
* **M10** : Partenariats avec établissements scolaires (licences B2B)
* **M11** : Optimisation IA pour analyse d'erreurs avancée
* **M12** : Bilan et planification V2 (expansion collège/lycée pro, nouvelles matières)
* **Équipe** : +1 growth marketer, +1 customer success, +1 dev mobile

### Risques & dépendances
* **Qualité OCR manuscrit** : forte variabilité selon écritures → dataset d'entraînement diversifié nécessaire
* **Pertinence des ressources externes** : nécessite curation et algorithme de filtrage robuste
* **RGPD & données mineures** : conformité stricte requise dès la conception
* **Adoption en milieu scolaire** : résistances possibles → stratégie d'évangélisation éducative 