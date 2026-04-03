# Philosophie des composants -- Revise Mieux

> Pour chaque famille de composants : feeling vise, principes, pieges a eviter.

---

## Principes transversaux

1. **Un composant, un job** : chaque composant fait une seule chose clairement
2. **L'etat Mastery est un citoyen de premiere classe** : les couleurs UNKNOWN/FRAGILE/OK/SOLID se retrouvent dans la majorite des composants
3. **Le feedback est structurel** : les etats de chargement, d'erreur et de succes sont integres dans chaque composant, pas ajoutes apres coup

---

## Boutons

| | |
|---|---|
| **Feeling vise** | Clair, invitant, accessible. Un bouton dit "appuie ici, ca va bien se passer". Ni agressif ni timide. |
| **Variants** | `primary` (fond bleu, texte blanc), `secondary` (fond bleu clair, texte bleu), `outline` (bordure), `ghost` (texte seul) |
| **Principes** | Un seul bouton primary par ecran. Le bouton primary est TOUJOURS l'action principale. Les boutons secondary/ghost sont pour les actions secondaires. La taille tactile minimum est 44px. |
| **Pieges a eviter** | Trop de boutons sur un ecran (l'eleve ne sait pas ou appuyer). Boutons avec des couleurs Mastery (confondre action et information). Bouton "Commencer la session" trop petit ou enfoui. |

**Evaluation de l'existant** : Le composant `Button` dans Themed.tsx est bien structure (4 variants, pressed opacity, icone optionnelle). Le radius md:12 et le padding sont corrects. Pas de changement structurel necessaire.

---

## Cards (chapitres, items, notions)

| | |
|---|---|
| **Feeling vise** | Conteneur propre et structure qui organise l'information en blocs digestibles. La card invite a explorer, pas a tout lire d'un coup. |
| **Principes** | Chaque card a une hierarchie claire : titre > indicateur principal (MasteryBar) > metadata secondaire. La card est pressable si elle mene quelque part (feedback visuel au press). Le fond est blanc (light) ou gray800 (dark), la bordure est subtile (1px gray200). |
| **Pieges a eviter** | Cards trop chargees (plus de 4 lignes d'info). Cards identiques sans differenciation visuelle (l'eleve ne distingue pas "chapitre" de "notion" de "item"). Cards avec ombres lourdes qui font "application enterprise". |

**Evaluation de l'existant** : Le composant `Card` dans Themed.tsx est minimaliste et correct (fond, bordure, radius lg, padding md). La `ChapterCard` dans index.tsx est bien structuree. Ameliorations possibles : ajouter un etat pressed (scale 0.98 ou background shift) pour le feedback tactile.

---

## Indicateurs de progression

| | |
|---|---|
| **Feeling vise** | Motivant et honnete. L'indicateur communique "tu avances" meme quand il reste du chemin. Il ne ment pas sur la progression mais il la presente positivement. |
| **Principes** | La MasteryBar est le composant central. Les pourcentages sont affiches en grand (style `display`). Les transitions entre etats sont animees (200ms). L'accent est sur les acquis (SOLID + OK) pas sur les manques (UNKNOWN). |
| **Pieges a eviter** | Barres de progression "vides" (fond gris avec 10% de couleur -- decourageant). Chiffres sans contexte ("4/20" sans dire ce que ca represente). Animations trop longues sur les progressions (l'eleve attend). |

**Evaluation de l'existant** : `MasteryBar` est le meilleur composant existant. La barre segmentee avec l'ordre SOLID > OK > FRAGILE > UNKNOWN est le bon pattern. `MasteryDot` et `MasteryBadge` sont corrects. Amelioration : animer les transitions de largeur des segments, renommer les labels (UNKNOWN -> NOUVEAU pour l'eleve).

---

## Elements de quiz

| | |
|---|---|
| **Feeling vise** | Focus et clarte. Pendant le quiz, l'eleve est "dans le tunnel" : pas de distraction, une question a la fois, feedback immediat. Le design communique "concentre-toi sur ca, rien d'autre". |
| **Principes** | Plein ecran (pas de tab bar). Une question par ecran. Barre de progression en haut (question 3/10). Les choix MCQ sont des cards pressables avec un etat selectionne franc. Le feedback post-reponse utilise les couleurs Mastery (pas rouge/vert binaire) : vert pour correct, ambre pour incorrect avec explication. |
| **Pieges a eviter** | Trop d'information par ecran (question + aide + timer + score + progression). Timer visible qui cree du stress. Feedback "FAUX" en gros rouge. Pas de bouton "Passer" (l'eleve se bloque et ferme l'app). |

---

## Alertes / banners

| | |
|---|---|
| **Feeling vise** | Informatif, jamais alarmiste. Les alertes attirent l'attention sans creer de panique. Le ton est "voici quelque chose a savoir" pas "ATTENTION DANGER". |
| **Principes** | 3 niveaux : info (bleu), attention (ambre), critique (rouge). Les alertes eleve sont courtes (1 ligne + action). Les alertes parent sont plus detaillees (contexte + recommandation). Les banners non-bloquants disparaissent apres action ou timeout. |
| **Pieges a eviter** | Alertes rouges pour des situations non critiques (cree une fatigue d'alerte). Pop-ups modales bloquantes pour des informations non urgentes. Alertes sans action possible ("Votre enfant n'a pas revise" -- et ensuite ?). |

---

## Champs de saisie

| | |
|---|---|
| **Feeling vise** | Simple et guide. La saisie doit etre minimale (l'app est un outil de revision, pas un formulaire). Quand elle est necessaire, elle est assistee. |
| **Principes** | Labels au-dessus du champ (pas de placeholder-only). Validation en temps reel pour les formats (date, nombre). Keyboard adapte au type de saisie (numerique pour les scores, texte pour les noms). |
| **Pieges a eviter** | Formulaires longs (un collegien de 12 ans abandonne apres 3 champs). Saisie libre quand un selecteur suffit (pour la matiere, proposer une liste). Messages d'erreur techniques ("Format invalide" au lieu de "Entre un nombre entre 0 et 20"). |

---

## Composants de capture (photo)

| | |
|---|---|
| **Feeling vise** | Rapide et rassurant. La capture doit etre aussi simple que prendre une photo sur Instagram. L'eleve ne doit pas penser "est-ce que c'est assez bien cadre ?" |
| **Principes** | Interface camera plein ecran, guides visuels legers (cadre semi-transparent). Feedback immediat apres capture ("Photo ajoutee"). Preview des photos prises avec possibilite de supprimer. Compteur de pages visible. |
| **Pieges a eviter** | Forcer un recadrage manuel (l'eleve ne le fera pas bien). Trop d'options (flash, HDR, grille). Message d'erreur "Photo trop floue" sans proposition (mieux : "Cette photo semble floue -- reprendre ?"). |

---

## Composants de review / correction

| | |
|---|---|
| **Feeling vise** | Constructif. La correction est un moment d'apprentissage, pas de sanction. Le design montre "voici ce qui etait attendu et pourquoi" plutot que "tu as eu tort". |
| **Principes** | Reponse de l'eleve affichee a cote de la reponse correcte. La bonne reponse est mise en evidence (background vert clair). L'explication courte apparait en dessous. Le bouton "Suivant" est le focus principal apres correction. |
| **Pieges a eviter** | Barrer la reponse incorrecte en rouge (associe a "faute" scolaire). Pas d'explication (juste "Faux"). Explication trop longue (l'eleve ne lit pas 5 lignes pendant un quiz). |

---

## Blocs de contenu pedagogique

| | |
|---|---|
| **Feeling vise** | Clair, structure, digne de confiance. Le contenu pedagogique doit avoir une "autorite visuelle" -- l'eleve percoit qu'il peut s'appuyer dessus pour reviser. |
| **Principes** | Fond legerement different du reste de la page (gray50). Hierachie forte : notion > definition > detail. Les termes-cles sont des badges inline cliquables. Les formules ont un traitement visuel distinct (fond primary.100, font mono). Les schemas/images ont un cadre et une legende. |
| **Pieges a eviter** | Mur de texte sans structure. Contenu qui ressemble a une page Wikipedia (trop dense, pas de hierarchie). Typo trop petite pour "tout faire tenir" (mieux : paginer ou utiliser un accordeon). |

---

## Composants de reassurance parent

| | |
|---|---|
| **Feeling vise** | Professionnel et lisible. Le parent doit sentir qu'il a affaire a un outil serieux qui mesure vraiment la progression, pas a une app qui dit "tout va bien" pour garder l'abonnement. |
| **Principes** | Donnees chiffrees au premier plan (pourcentages, nombre d'items). Les couleurs Mastery sont les memes que dans l'espace eleve (langage commun). Le digest est structure en sections fixes (cf Z6-AC27 : activite, maitrise, alertes, prochaine action, score). Le ton est informatif et mesure. |
| **Pieges a eviter** | Dashboard "tableau de bord d'avion" (trop de metrics). Seuls les chiffres sans contexte ("Score: 67%" -- c'est bien ou pas ?). Jargon technique (Mastery, HITL, OCR). Graphiques complexes sans conclusion ("La courbe monte, et donc ?"). |
