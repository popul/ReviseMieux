# Principes de design -- Revise Mieux

> 6 principes fondateurs pour guider chaque decision de design.

---

## 1. La progression, pas le jugement

**Enonce** : Chaque element visuel communique "ou tu en es" et "ou tu vas", jamais "ou tu devrais etre". L'interface montre le chemin parcouru autant que le chemin restant.

**Exemple concret** : La MasteryBar affiche les 4 segments (SOLID + OK + FRAGILE + NOUVEAU) proportionnellement. Un eleve avec 30% de maitrise voit une barre qui "avance", pas une barre "a 70% vide".

**Contre-exemple** : Afficher un gros "30%" rouge avec un message "Insuffisant". Utiliser un classement ou une comparaison avec une "moyenne attendue". Afficher un chronometre qui compte les jours sans revision.

---

## 2. Clair avant beau

**Enonce** : En cas de conflit entre esthetique et clarte, la clarte gagne toujours. Un eleve de 12 ans distrait a 18h doit comprendre en 2 secondes ce qu'il doit faire.

**Exemple concret** : Le dashboard affiche UNE action principale par contexte ("Reviser Physique-Chimie -- 15 min") avec un bouton visible. Les informations secondaires (stats, historique) sont en retrait.

**Contre-exemple** : Un dashboard avec 6 cartes de taille egale, 3 badges, 2 notifications et un graphique -- l'eleve ne sait pas par ou commencer. Mettre des informations utiles derriere un menu hamburger "pour faire propre".

---

## 3. Sobre dehors, riche dedans

**Enonce** : L'interface est epuree en surface mais revele de la profondeur quand on explore. Les informations detaillees existent, elles ne sont pas imposees.

**Exemple concret** : La carte chapitre affiche le nom + la MasteryBar + le pourcentage. Un tap revele le detail par notion, les items fragiles, l'historique des sessions.

**Contre-exemple** : Afficher toutes les statistiques sur la premiere page ("items FRAGILE: 3, items UNKNOWN: 5, derniere session: hier, prochaine due: demain, score moyen: 67%..."). A l'inverse, masquer les donnees au point que le parent ne trouve rien.

---

## 4. Respecter l'autonomie

**Enonce** : L'eleve est acteur de sa revision, pas spectateur d'un systeme. L'interface l'aide a decider, ne decide pas a sa place. Le parent est informe, pas aux commandes.

**Exemple concret** : Apres une session, le feedback dit "3 points progressent vers SOLID -- tu peux continuer ou reprendre demain". Le parent recoit un digest hebdo, pas une notification a chaque session.

**Contre-exemple** : Forcer une session au demarrage de l'app ("Tu dois reviser avant de continuer"). Envoyer au parent un recap en temps reel de chaque reponse. Bloquer l'acces au dashboard tant que la session du jour n'est pas finie.

---

## 5. Un langage visuel, deux audiences

**Enonce** : Les couleurs Mastery, la structure des composants et les patterns de navigation sont identiques pour l'eleve et le parent. Ce qui change : le niveau de detail, le ton du copy, et la densite d'information.

**Exemple concret** : L'eleve voit "12 points SOLID" avec une barre verte. Le parent voit "Maitrise : 60% (12/20 bien acquis)" avec la meme barre verte. Meme couleur, meme composant, profondeur differente.

**Contre-exemple** : Creer un theme "enfant" colore et un theme "parent" gris -- ca cree une schizophrenie visuelle et le parent ne reconnait pas l'app de son enfant. Utiliser des termes differents pour les memes concepts (le parent voit "acquis" la ou l'eleve voit "SOLID").

---

## 6. Le minimum qui fonctionne

**Enonce** : Chaque element d'interface doit justifier sa presence. Si retirer un composant ne degrade pas la comprehension ou l'experience, il n'a pas sa place. Le design system croit par besoin, pas par anticipation.

**Exemple concret** : Le Lot 0 n'a pas d'avatar, pas de themes, pas de graphiques temporels. Il a une MasteryBar, un badge d'etat, des boutons, des cards. Ca suffit pour que l'eleve revise et que le parent comprenne.

**Contre-exemple** : Implementer un systeme de themes complet "parce qu'on en aura besoin plus tard". Ajouter des illustrations a chaque ecran "pour faire plus professionnel". Creer 12 variantes de bouton pour couvrir tous les cas possibles.
