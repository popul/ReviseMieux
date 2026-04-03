# Decisions produit ouvertes -- Revise Mieux

> Decisions qui necessitent un arbitrage humain avant l'implementation. Classees par impact et urgence.

---

## Haute priorite (bloquant pour le Lot 0 P2)

### OD-01 : App unique ou app parent separee ?

**Question** : Le parent utilise-t-il la meme app que l'eleve (avec un switch de profil) ou une app/PWA distincte ?

**Options** :
- **(A) Meme app, profil parent** : un compte parent dans la meme app Expo. Toggle entre les vues. Moins de friction a l'installation (un seul store listing). Risque : l'eleve percoit l'app comme "celle de mes parents".
- **(B) App parent separee** : PWA ou app legere. Separation nette des audiences. L'eleve ne voit jamais l'interface parent. Cout : deux codebases a maintenir.
- **(C) Meme app, acces parent via web** : le parent accede a un dashboard web. Zero installation supplementaire. Le lien est dans le digest email. Cout : developper un dashboard web.

**Recommandation** : Option (A) pour le Lot 0 avec un switch de profil clair. L'eleve ne doit jamais voir le dashboard parent par accident. Migrer vers (C) en post-MVP si le parent utilise principalement le digest.

**Impact** : Architecture des routes Expo Router, strategie d'authentification, structure de navigation.

---

### OD-02 : Auto-evaluation vs scoring automatique en Lot 0

**Question** : Le Lot 0 utilise l'auto-evaluation ("Je savais / Je ne savais pas") car le scoring LLM en temps reel n'est pas implemente. Quand basculer vers le scoring automatique ?

**Options** :
- **(A) Rester en auto-evaluation pour le Lot 0 entier** : simple, pas de dependance LLM en temps reel. Risque : l'eleve triche (surtout Lucas).
- **(B) Scoring automatique MCQ des le Lot 0** : les MCQ ont une reponse deterministe, pas besoin de LLM. L'auto-evaluation reste pour les SHORT/KEYWORDS/RUBRIC.
- **(C) Scoring LLM pour tout des le Lot 0** : experience complete mais dependance forte au LLM pendant les sessions.

**Recommandation** : Option (B). Les MCQ representent ~40% des questions en difficulte 1. Le scoring automatique des MCQ est trivial (comparaison de choix). L'auto-evaluation reste pour les questions ouvertes jusqu'au scoring LLM.

**Impact** : Composant de session (affichage feedback), API submit answer, logique de scoring.

---

### OD-03 : Position du bouton "Ajouter des pages" dans la navigation

**Question** : Le bouton "Ajouter des pages" (Z5-AC11) est-il sur la carte du chapitre, dans le tab capture, ou les deux ?

**Options** :
- **(A) Sur la carte du chapitre uniquement** : l'eleve qui veut ajouter va naturellement sur le chapitre. Coherent avec le contexte.
- **(B) Dans le tab capture avec selection du chapitre** : l'eleve qui ouvre l'app pour capturer commence par le tab capture. Il choisit ensuite "Nouveau chapitre" ou "Ajouter a [chapitre existant]".
- **(C) Les deux** : duplication de l'entree mais couverture maximale des comportements.

**Recommandation** : Option (C). Le bouton sur la carte du chapitre couvre le cas "je revise et je vois qu'il manque des pages". Le tab capture couvre le cas "j'ai pris des notes aujourd'hui, je veux les ajouter". Les deux sont des scenarios reels.

**Impact** : Navigation capture, flow de selection de chapitre dans le tab capture.

---

## Priorite moyenne (impacte l'experience mais pas bloquant)

### OD-04 : Moment d'affichage de la saisie d'emploi du temps

**Question** : Z6-AC01 specifie que la saisie d'emploi du temps apparait au premier upload d'une matiere. Est-ce le bon moment, ou est-ce trop tot (friction dans le flow de capture) ?

**Options** :
- **(A) Au premier upload (spec actuelle)** : contextuel, l'eleve sait quand il a cette matiere. Friction : interrompt le flow capture -> pipeline.
- **(B) Apres la premiere session completee** : l'eleve a deja vecu la valeur. Moins de friction initiale.
- **(C) Dans les parametres uniquement** : aucune interruption. Risque : personne ne le remplit jamais.

**Recommandation** : Option (A) avec un bouton "Plus tard" tres visible. Le moment est naturel ("tu viens de capturer ton cours de physique, quand as-tu physique ?"). Le mode degrade (Z6-AC11) protege contre le skip.

**Impact** : Flow de creation de chapitre, timing des ecrans intermediaires.

---

### OD-05 : Presentation des regressions dans le debrief

**Question** : Quand un item regresse (SOLID -> OK, OK -> FRAGILE), faut-il le montrer dans le debrief ou le masquer ?

**Options** :
- **(A) Montrer avec framing positif** : "Ce point a besoin d'un refresh -- on le revoit bientot." Transparent mais potentiellement anxiogene (Ines).
- **(B) Masquer les regressions** : le debrief ne montre que les progressions positives. L'eleve decouvre la regression dans le dashboard. Risque : sentiment de tromperie.
- **(C) Montrer uniquement si l'eleve est en mode "detail"** : debrief minimal (score + progressions), bouton "Voir le detail" pour les regressions.

**Recommandation** : Option (C). Le debrief reste positif par defaut (progressions uniquement). L'eleve curieux peut voir le detail. Le framing growth mindset s'applique dans le detail.

**Impact** : Composant debrief, structure de la reponse API debrief.

---

### OD-06 : Nombre de notifications parent par defaut

**Question** : Quelles notifications parent sont activees par defaut a la liaison ?

**Options** :
- **(A) Tout active** : routine terminee, session manquee, inactivite, changement EDT. Risque : trop de bruit, le parent desactive tout.
- **(B) Minimal** : uniquement routine terminee + digest hebdo. Le parent active le reste si il veut.
- **(C) Progressive** : semaine 1 = routine terminee uniquement. Semaine 2 = ajout digest. Semaine 3 = ajout alertes inactivite.

**Recommandation** : Option (B). Le signal positif (routine terminee) et le digest hebdo couvrent 80% du besoin. Les alertes negatives (session manquee, inactivite) sont opt-in pour eviter l'alert fatigue.

**Impact** : ParentNotificationPref defaults, ecran de preferences parent.

---

### OD-07 : Traitement du chapitre demo apres le premier upload

**Question** : Z8-AC01 specifie que le chapitre demo est archive automatiquement quand le premier vrai chapitre produit >= 1 item. Faut-il le supprimer ou l'archiver ?

**Options** :
- **(A) Archiver (masquer du dashboard)** : le chapitre demo reste en base, les masteries demo sont conservees. L'eleve peut le retrouver dans les parametres.
- **(B) Supprimer** : nettoyage complet. Pas de pollution en base.

**Recommandation** : Option (A). L'archivage est non-destructif. Si l'eleve veut revenir au chapitre demo (rare mais possible), il peut.

**Impact** : Logique d'archivage, filtre dashboard.

---

## Priorite basse (post-MVP mais a documenter)

### OD-08 : Gestion multi-enfants pour le parent

**Question** : Un parent avec 2 enfants dans l'app (ex: Theo en 4e et sa soeur en 6e) voit-il un dashboard unifie ou deux dashboards separes ?

**Recommandation** : Dashboard avec un selecteur d'enfant en haut. Notifications distinctes par enfant (Z6-AC39). A specifier en detail post-MVP.

---

### OD-09 : Partage de contenu entre eleves

**Question** : Deux eleves de la meme classe peuvent-ils partager un chapitre (eviter le double OCR) ?

**Recommandation** : Hors scope Lot 0. Le partage introduit des problemes de propriete des masteries, de RGPD, et d'integrite pedagogique. A evaluer en phase 2.

---

### OD-10 : Localisation (langues autres que le francais)

**Question** : L'app est entierement en francais. Faut-il prevoir l'internationalisation des le depart ?

**Recommandation** : Non pour le Lot 0. Les messages de feedback (growth mindset, debrief) sont trop culturellement specifiques pour etre traduits mecaniquement. Structurer le code pour i18n (fichiers de strings, pas de texte en dur dans les composants) mais ne pas investir dans les traductions.

---

### OD-11 : Monetisation et impact sur l'UX

**Question** : Le modele freemium (prevu en phase 3 du PRD) impactera l'experience. Quelles fonctionnalites sont candidates au paywall ?

**Recommandation** : A definir post-MVP. Ne pas designer de fonctionnalites avec un paywall en tete pour le Lot 0. Le risque est de degrader l'experience gratuite au point de perdre l'adoption.

---

## Resume par priorite

| ID | Decision | Priorite | Bloquant pour |
|----|----------|----------|---------------|
| OD-01 | App unique vs separee (parent) | Haute | Architecture routes |
| OD-02 | Auto-eval vs scoring auto MCQ | Haute | Composant session |
| OD-03 | Position bouton "Ajouter des pages" | Haute | Navigation capture |
| OD-04 | Moment saisie emploi du temps | Moyenne | Flow creation chapitre |
| OD-05 | Regressions dans le debrief | Moyenne | Composant debrief |
| OD-06 | Notifications parent par defaut | Moyenne | Preferences parent |
| OD-07 | Demo archive vs supprime | Moyenne | Logique archivage |
| OD-08 | Multi-enfants parent | Basse | Post-MVP |
| OD-09 | Partage entre eleves | Basse | Post-MVP |
| OD-10 | Internationalisation | Basse | Post-MVP |
| OD-11 | Monetisation et paywall | Basse | Phase 3 |
