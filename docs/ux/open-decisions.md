# Decisions produit -- Revise Mieux

> Decisions produit et UX pour le Lot 0. Chaque decision est documentee avec son statut, sa justification et sa date.

---

## Haute priorite (bloquant pour le Lot 0 P2)

### OD-01 : App unique ou app parent separee ?

**Statut** : Decide -- 2026-04-03

**Question** : Le parent utilise-t-il la meme app que l'eleve (avec un switch de profil) ou une app/PWA distincte ?

**Decision** : App unique, profil parent avec switch.

**Justification** : Simplifie le developpement, une seule codebase. Un seul store listing reduit la friction a l'installation. L'eleve ne voit jamais le dashboard parent grace au switch de profil. Migration possible vers dashboard web (option C) en post-MVP si le parent utilise principalement le digest.

**Impact** : Architecture des routes Expo Router, strategie d'authentification, structure de navigation.

---

### OD-02 : Auto-evaluation vs scoring automatique en Lot 0

**Statut** : Decide -- 2026-04-03

**Question** : Le Lot 0 utilise l'auto-evaluation ("Je savais / Je ne savais pas") car le scoring LLM en temps reel n'est pas implemente. Quand basculer vers le scoring automatique ?

**Decision** : Scoring automatique sur tout (pas seulement MCQ), via LLM backend (Gemini Flash).

**Justification** : Plus fiable que l'auto-evaluation. Le scoring LLM est disponible via Gemini Flash cote backend, ce qui elimine le risque de triche (surtout pour les eleves comme Lucas). Couvre 100% des types de questions (MCQ, SHORT, KEYWORDS, RUBRIC).

**Impact** : Composant de session (affichage feedback), API submit answer, logique de scoring.

---

### OD-03 : Position du bouton "Ajouter des pages" dans la navigation

**Statut** : Decide -- 2026-04-03

**Question** : Le bouton "Ajouter des pages" (Z5-AC11) est-il sur la carte du chapitre, dans le tab capture, ou les deux ?

**Decision** : Deux boutons "Ajouter des pages" (carte chapitre + tab capture).

**Justification** : Accessibilite maximale, les deux contextes sont naturels. Le bouton sur la carte du chapitre couvre le cas "je revise et je vois qu'il manque des pages". Le tab capture couvre le cas "j'ai pris des notes aujourd'hui, je veux les ajouter".

**Impact** : Navigation capture, flow de selection de chapitre dans le tab capture.

---

## Priorite moyenne (impacte l'experience mais pas bloquant)

### OD-04 : Moment d'affichage de la saisie d'emploi du temps

**Statut** : Reporte -- 2026-04-03

**Question** : Z6-AC01 specifie que la saisie d'emploi du temps apparait au premier upload d'une matiere. Est-ce le bon moment, ou est-ce trop tot (friction dans le flow de capture) ?

**Decision** : Reporter -- pas d'emploi du temps en Lot 0.

**Justification** : PRD S11 est post-MVP. La saisie d'emploi du temps ajoute de la friction dans le flow de capture sans apporter de valeur immediate en Lot 0 (1 pere + 1 fils, usage local). Le mode degrade (Z6-AC11) couvre le cas ou l'emploi du temps n'est pas rempli.

**Impact** : Flow de creation de chapitre, timing des ecrans intermediaires.

---

### OD-05 : Presentation des regressions dans le debrief

**Statut** : Decide -- 2026-04-03

**Question** : Quand un item regresse (SOLID -> OK, OK -> FRAGILE), faut-il le montrer dans le debrief ou le masquer ?

**Decision** : Ton neutre pour les regressions ("A revoir", pas de rouge).

**Justification** : Growth mindset, ne pas decourager l'eleve. Le debrief reste positif par defaut (progressions uniquement). Le framing growth mindset s'applique quand le detail est consulte. Pas de signaux visuels anxiogenes (rouge, croix, alertes).

**Impact** : Composant debrief, structure de la reponse API debrief.

---

### OD-06 : Nombre de notifications parent par defaut

**Statut** : Decide -- 2026-04-04

**Question** : Quelles notifications parent sont activees par defaut a la liaison ?

**Decision** : 1 notification parent/semaine (digest).

**Justification** : Pas de spam en Lot 0, le parent consulte quand il veut. Le signal positif (routine terminee) et le digest hebdo couvrent 80% du besoin. Les alertes negatives (session manquee, inactivite) sont opt-in pour eviter l'alert fatigue.

**Impact** : ParentNotificationPref defaults, ecran de preferences parent.

---

### OD-07 : Traitement du chapitre demo apres le premier upload

**Statut** : Decide -- 2026-04-03

**Question** : Z8-AC01 specifie que le chapitre demo est archive automatiquement quand le premier vrai chapitre produit >= 1 item. Faut-il le supprimer ou l'archiver ?

**Decision** : Garder le chapitre demo visible, marque "Demo", suppressible.

**Justification** : L'eleve peut le supprimer, mais il reste comme reference. L'archivage est non-destructif. Si l'eleve veut revenir au chapitre demo (rare mais possible), il peut. Le marquage "Demo" evite la confusion avec les vrais chapitres.

**Impact** : Logique d'archivage, filtre dashboard.

---

## Priorite basse (post-MVP mais documente)

### OD-08 : Gestion multi-enfants pour le parent

**Statut** : Reporte -- 2026-04-03

**Question** : Un parent avec 2 enfants dans l'app (ex: Theo en 4e et sa soeur en 6e) voit-il un dashboard unifie ou deux dashboards separes ?

**Decision** : Reporter multi-enfants au MVP.

**Justification** : Lot 0 = 1 pere + 1 fils. Le multi-enfants introduit de la complexite (selecteur d'enfant, notifications distinctes par enfant Z6-AC39) sans valeur pour le cas d'usage Lot 0. A specifier en detail post-MVP.

---

### OD-09 : Partage de contenu entre eleves

**Statut** : Reporte -- 2026-04-03

**Question** : Deux eleves de la meme classe peuvent-ils partager un chapitre (eviter le double OCR) ?

**Decision** : Reporter partage entre eleves.

**Justification** : Hors scope Lot 0 et MVP. Le partage introduit des problemes de propriete des masteries, de RGPD, et d'integrite pedagogique. A evaluer en phase 2.

---

### OD-10 : Localisation (langues autres que le francais)

**Statut** : Reporte -- 2026-04-03

**Question** : L'app est entierement en francais. Faut-il prevoir l'internationalisation des le depart ?

**Decision** : Reporter localisation (francais uniquement).

**Justification** : MVP = marche francais. Les messages de feedback (growth mindset, debrief) sont trop culturellement specifiques pour etre traduits mecaniquement. Structurer le code pour i18n (fichiers de strings, pas de texte en dur dans les composants) mais ne pas investir dans les traductions.

---

### OD-11 : Monetisation et impact sur l'UX

**Statut** : Reporte -- 2026-04-03

**Question** : Le modele freemium (prevu en phase 3 du PRD) impactera l'experience. Quelles fonctionnalites sont candidates au paywall ?

**Decision** : Reporter monetisation.

**Justification** : Gratuit en Lot 0, monetisation post-MVP. Ne pas designer de fonctionnalites avec un paywall en tete pour le Lot 0. Le risque est de degrader l'experience gratuite au point de perdre l'adoption.

---

## Resume

| ID | Decision | Statut | Date | Justification |
|----|----------|--------|------|---------------|
| OD-01 | App unique, profil parent avec switch | Decide | 2026-04-03 | Simplifie le developpement, une seule codebase |
| OD-02 | Scoring automatique sur tout via LLM | Decide | 2026-04-03 | Plus fiable que l'auto-evaluation |
| OD-03 | Deux boutons "Ajouter des pages" | Decide | 2026-04-03 | Accessibilite maximale, deux contextes naturels |
| OD-04 | Reporter emploi du temps | Reporte | 2026-04-03 | PRD S11 est post-MVP |
| OD-05 | Ton neutre pour regressions | Decide | 2026-04-03 | Growth mindset |
| OD-06 | 1 notification parent/semaine | Decide | 2026-04-04 | Pas de spam en Lot 0 |
| OD-07 | Demo visible, marquee, suppressible | Decide | 2026-04-03 | Reference pour l'eleve |
| OD-08 | Reporter multi-enfants | Reporte | 2026-04-03 | Lot 0 = 1 pere + 1 fils |
| OD-09 | Reporter partage entre eleves | Reporte | 2026-04-03 | Hors scope Lot 0 et MVP |
| OD-10 | Reporter localisation | Reporte | 2026-04-03 | MVP = marche francais |
| OD-11 | Reporter monetisation | Reporte | 2026-04-03 | Gratuit en Lot 0 |
