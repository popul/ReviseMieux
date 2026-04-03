# Do / Don't -- Revise Mieux

> Exemples concrets avec descriptions visuelles pour calibrer le curseur design.

---

## Trop enfantin

| Exemple | Description visuelle | Pourquoi c'est un probleme |
|---------|---------------------|---------------------------|
| Mascotte animee qui reagit aux reponses | Un hibou/robot qui saute de joie quand c'est correct, qui pleure quand c'est faux, avec des bulles de dialogue | L'eleve de 14 ans se sent infantilise. Il ne veut pas qu'un personnage de dessin anime juge ses reponses. Il ferme l'app et ne revient pas. |
| Palette bonbon | Fond rose pastel, boutons vert menthe, titres violet bubblegum, emojis partout | Le parent ne prend pas l'outil au serieux ("ca ressemble a une app pour les 5 ans"). L'eleve de 3e a honte si un camarade voit son ecran. |
| Recompenses visuelles excessives | Confettis, etoiles dorees qui tombent, animation de "level up" avec fanfare, badge "Champion du jour" | La gamification detourne l'attention de l'apprentissage. L'eleve cherche la recompense, pas la comprehension. Le parent se demande si l'enfant apprend vraiment. |
| Langage surjoue | "WOOOW GENIAL ! Tu es un SUPER HERO de la physique !" | Faux positif -- l'eleve sait qu'il a juste repondu a un QCM facile. Il perd confiance dans le feedback de l'app. |

---

## Trop froid

| Exemple | Description visuelle | Pourquoi c'est un probleme |
|---------|---------------------|---------------------------|
| Interface grise monochrome | Fond blanc, texte gris fonce, pas de couleur sauf les boutons bleu generique, tableaux de donnees | L'eleve percoit un outil de surveillance impose par ses parents. Zero envie d'ouvrir l'app le soir. |
| Donnees brutes sans interpretation | "Items FRAGILE: 7, Items UNKNOWN: 12, Score moyen: 0.43, Ecart-type: 0.21" | Le parent data-scientist comprend, les 95% autres ne savent pas si c'est bien ou pas. L'eleve ne comprend rien. |
| Feedback minimal | Apres une reponse incorrecte : "Incorrect." Point. Pas d'explication, pas d'encouragement. | L'eleve ne sait pas pourquoi il s'est trompe ni comment progresser. L'app n'enseigne pas, elle juge. |
| Ton distant | "L'utilisateur n'a pas complete la session planifiee." (dans une notification) | L'eleve se sent comme un numero dans un systeme. Le parent a l'impression d'utiliser un logiciel RH. |

---

## Trop gamifie

| Exemple | Description visuelle | Pourquoi c'est un probleme |
|---------|---------------------|---------------------------|
| Systeme de points et boutique | XP gagnes par question, catalogue de themes/avatars a debloquer, monnaie virtuelle | L'eleve optimise le systeme de points (reponses rapides aux QCM faciles) au lieu d'apprendre. Le parent voit un jeu, pas un outil educatif. |
| Streaks avec penalite visuelle | "Serie de 12 jours !" en vert fluo, puis "SERIE PERDUE" en rouge avec un compteur remis a zero | Anxiete, culpabilisation. L'eleve revise par peur de perdre sa serie, pas par envie d'apprendre. Quand il la perd (vacances, maladie), il ne revient pas. |
| Classement / competition | "Tu es 3eme de ta classe ! Depasse Clara pour atteindre la 2eme place !" | Stress social, comparaison toxique. Les eleves en difficulte sont humilies. Les bons eleves ne veulent pas etre identifies. |
| Defi quotidien force | "Defi du jour : reponds a 20 questions en 5 minutes !" avec un timer clignotant | Transforme la revision en course contre la montre. L'eleve bacle pour finir le defi. La qualite de l'apprentissage s'effondre. |

---

## Trop institutionnel

| Exemple | Description visuelle | Pourquoi c'est un probleme |
|---------|---------------------|---------------------------|
| Interface type ENT/Pronote | Navigation par menus textuels, tableau "matieres x competences", boutons rectangulaires gris, police Times New Roman | L'eleve a deja cette interface a l'ecole. Revise Mieux est percue comme "encore un truc scolaire", pas comme un outil qui l'aide. |
| Jargon educatif | "Competences transversales", "Evaluation formative", "Progression curricullaire" | L'eleve ne comprend pas. Le parent se demande si c'est une app ou un bulletin scolaire. |
| Structure rigide par matiere/competence | Arborescence Matiere > Theme > Competence > Exercice, avec des coches vertes/rouges | L'eleve se sent enferme dans un programme. Il ne voit pas SES progres, il voit un programme a remplir. |
| Notifications type administration | "Votre enfant doit completer l'evaluation de la sequence 3 avant le 15 mars." | Le parent se sent dans une relation ecole-famille, pas dans un outil d'aide. L'eleve recoit une injonction, pas une invitation. |

---

## Bien equilibre (ce qu'on vise)

| Exemple | Description visuelle | Pourquoi ca marche |
|---------|---------------------|-------------------|
| Dashboard avec MasteryBar | Titre "Bonjour Hugo", une barre coloree SOLID/OK/FRAGILE/NOUVEAU, un gros "72%", un bouton "Reviser". Fond blanc, touches de bleu et vert. | L'eleve voit sa progression en 1 seconde. Le parent voit un chiffre concret. Le design est propre sans etre austere, colore sans etre criard. |
| Feedback de quiz | Question en haut, choix MCQ en cards blanches. Apres reponse : la bonne reponse passe en vert clair, une ligne d'explication apparait en dessous, le bouton "Suivant" est au centre. | Pas de jugement visuel ("FAUX !"), pas d'exclamation ("GENIAL !"). Juste l'information utile : voici la bonne reponse, voici pourquoi, on continue. |
| Transition FRAGILE vers OK | La section bleue de la MasteryBar grandit avec une animation fluide (300ms). Un message discret apparait : "Progres : 1 point de plus vers OK". | L'eleve voit sa progression en temps reel. C'est satisfaisant sans etre excessif. Le parent voit les memes couleurs dans son digest. |
| Carte chapitre | Nom du chapitre en gras, matiere en caption au-dessus, MasteryBar en dessous, nombre d'items en small. Un bouton "Reviser" discret a droite. | Hierarchie claire, information essentielle visible, action possible. Pas surcharge, pas vide. |
| Notification eleve | "Tes revisions t'attendent -- 8 points a revoir en Physique-Chimie." | Direct, informatif, pas culpabilisant. L'eleve sait ce qui l'attend sans se sentir juge. |
| Digest parent | "Hugo a revise 4 fois cette semaine. Physique-Chimie : 65% de maitrise (+8%). 2 points restent fragiles : densite, unites. Conseil : encouragez-le a refaire une session avant le controle de vendredi." | Le parent a toutes les infos en 30 secondes. Les couleurs Mastery sont les memes que dans l'app. Le conseil est actionnable. |
| Empty state (aucun chapitre) | Illustration geometrique minimaliste (lignes, cercles, une camera stylisee). Texte : "Commence par prendre en photo une page de cours". Bouton primaire "Capturer un cours". | Pas de mascotte triste. Pas de texte qui culpabilise. Juste un guide clair vers l'action suivante. |
