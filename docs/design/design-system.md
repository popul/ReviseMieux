# Design System -- Revise Mieux

> Fondations strategiques d'un design system intergenerationnel pour collegiens (11-15 ans) et parents.

---

## 1. Executive summary

Revise Mieux est une app de revision pour collegiens ou le design doit resoudre une equation rare : plaire a un ado de 13 ans (qui decide d'ouvrir l'app chaque soir) ET rassurer un parent de 40 ans (qui decide de payer et d'installer). Le design system propose un positionnement "outil premium pour jeunes" -- ni jouet, ni logiciel scolaire -- avec une identite visuelle moderne, des codes empruntes aux apps que les ados utilisent deja (Notion, Spotify, Instagram), et une structure qui inspire confiance aux parents.

Les couleurs Mastery (UNKNOWN, FRAGILE, OK, SOLID) sont le coeur visuel du produit. Elles doivent communiquer la progression sans juger, encourager sans mentir, et etre lisibles en un coup d'oeil par les deux audiences.

---

## 2. Analyse du defi de design

### Pourquoi c'est difficile

L'app doit etre adoptee par deux publics aux codes visuels opposes :
- Un collegien de 14 ans associe "serieux" a "ennuyeux" et "scolaire" a "punition"
- Un parent de 40 ans associe "fun" a "pas serieux" et "gamifie" a "distraction"

La zone de chevauchement est etroite : un design qui communique **competence**, **modernite** et **progression** sans tomber dans les extremes.

### Pieges visuels a eviter

| Piege | Symptome | Pourquoi c'est fatal |
|-------|----------|---------------------|
| Infantilisation | Mascotte animee, couleurs pastel bonbon, emojis partout | L'ado de 14 ans quitte immediatement -- il ne veut pas d'une app "pour les petits" |
| Froideur institutionnelle | Gris dominant, tableaux de donnees, ton formel | L'eleve percoit un outil de surveillance impose par le parent |
| Sur-gamification | Streaks, classements, coffres, XP | Le parent doute du serieux ; l'eleve peut devenir dependant au systeme plutot qu'a l'apprentissage |
| Copie scolaire | Interface qui ressemble a Pronote ou un ENT | Association negative immediate -- l'app "ressemble a l'ecole" |

### Reponse emotionnelle visee

| Public | Emotion a l'ouverture | Emotion apres 5 min |
|--------|----------------------|---------------------|
| Eleve | "C'est propre, ca a l'air bien fait" | "Je vois ou j'en suis, je sais quoi faire" |
| Parent | "C'est serieux, bien structure" | "Je comprends la progression de mon enfant" |

### Ce qui differencie "engageant pour ado" de "enfantin"

- **Enfantin** : couleurs saturees, formes rondes, animations excessives, langage simplifie a l'extreme
- **Engageant pour ado** : palette equilibree, micro-animations subtiles, mise en page aeree, ton direct sans etre familier, codes visuels empruntes aux apps "adultes" que les ados admirent

---

## 3. Personas et attentes audiences

### Personas design

| Persona | Age | Ce qui attire | Ce qui repousse | Apps aimees |
|---------|-----|---------------|-----------------|-------------|
| **Collegien 6e** (11-12 ans) | 11-12 | Couleurs vives (pas criardes), animations, retour positif immediat, sentiment de progression | Texte dense, interfaces austeres, impression de "devoirs en plus" | Duolingo (fun sans etre bebe), YouTube (feed personnalise), Brawl Stars (progression visible) |
| **Collegien 3e** (14-15 ans) | 14-15 | Design epure et "mature", typographie moderne, dark mode, interface qui ressemble aux apps qu'il utilise | Tout ce qui fait "bebe", mascotte, couleurs pastel, emojis excessifs, comparaison avec les autres | Spotify (dark mode, epure), Instagram (visuels soignes), Notion (outil serieux qui a du style) |
| **Parent implique** (35-50 ans) | 35-50 | Donnees claires, structure lisible, credibilite visuelle, sentiment de controle | Design "jouet", absence de donnees concretes, flou sur ce que fait l'enfant | Bankin (dashboard clair), Apple Health (progression data), MyFitnessPal (suivi structure) |
| **Parent delegateur** (35-50 ans) | 35-50 | Simplicite, pas de surcharge, alertes claires, confiance que "ca tourne" | Complexite, trop de parametres, sollicitation frequente | Doctolib (simple, efficace), Meteo France (info rapide), WhatsApp (notifications claires) |

### Pour chaque persona : curseur fun/serieux

- **6e** : 60% fun / 40% serieux -- besoin de recompense visible, tolerance pour les animations
- **3e** : 35% fun / 65% serieux -- veut se sentir traite comme un "grand", prefere l'efficacite
- **Parent implique** : 20% fun / 80% serieux -- veut des donnees, accepte une touche de couleur
- **Parent delegateur** : 10% fun / 90% serieux -- veut juste savoir si ca va ou pas

### Zone de confort partagee

Les 4 personas convergent sur :
1. **Proprete visuelle** : espaces blancs, hierarchie claire, pas de surcharge
2. **Progression visible** : barres, pourcentages, indicateurs colores lisibles
3. **Modernite** : coins arrondis, ombres subtiles, typographie contemporaine
4. **Efficacite** : pas de clics inutiles, information accessible rapidement
5. **Couleurs a signification** : vert = bien, orange = attention, pas d'ambiguite

---

## 4. Positionnement de design

```
playful     ●○○○○○○●○○  serious
             ^--- ICI (position 3/10)
             Engageant sans etre ludique. Des micro-animations et couleurs
             expressives, mais pas de jeu. L'emotion vient de la progression, pas du fun.

childish    ○○○○○○○○●○  youthful
                         ^--- ICI (position 8/10)
             Resolument jeune et moderne, jamais enfantin. Les codes visuels
             sont ceux des apps que les ados admirent, pas des apps pour enfants.

gimmicky    ○○○○○○○○●○  modern
                         ^--- ICI (position 8/10)
             Design system structure et perenne, pas d'effets de mode.
             Sobriete graphique inspiree de Notion/Linear, pas de gradients flashy.

institutional ○○○●○○○○○○  educational
               ^--- ICI (position 3/10)
               Educatif par le contenu et les feedbacks, pas par l'esthetique.
               L'interface ne ressemble pas a un ENT ou un manuel scolaire.

controlling ○○○○○○○●○○  reassuring
                     ^--- ICI (position 7/10)
             Le parent est informe, pas aux commandes. L'eleve garde
             l'autonomie. Le design communique "tu es capable" pas "on te surveille".
```

---

## 5. Principes emotionnels et tonals

### Ton emotionnel

L'interface communique : **"Tu es capable, et je suis la pour t'aider a le voir."**

### Feeling de marque (3 mots)

**Clair. Capable. Progressif.**

### Personnalite visuelle

Si l'app etait une personne : un grand frere/une grande soeur etudiant(e) en prepa -- organise(e), bienveillant(e), un peu cool, qui explique sans juger et celebre les progres sans en faire trop.

### Ton des interactions

- **Micro-copy** : direct, deuxieme personne du singulier (tutoiement pour l'eleve), phrases courtes
- **Notifications** : informatives, jamais culpabilisantes ("Tes revisions t'attendent" plutot que "Tu n'as pas revise")
- **Erreurs** : expliquees simplement, avec un chemin de resolution ("L'envoi a echoue. Reessaie ou verifie ta connexion.")
- **Feedback positif** : sobre mais satisfaisant ("Bien joue" plutot que "INCROYABLE !!!!")

### Copywriting UI

- **Eleve** : tutoiement, ton direct et encourageant
- **Parent** : vouvoiement, ton informatif et rassurant
- **Transitions entre espaces** : le changement de ton est naturel, pas abrupt

---

## 6. Design tokens fondamentaux

### 6.1 Couleurs

Voir le document dedie `color-system.md` pour la palette complete. Resume :

- **Primaire** : Bleu `#4A90D9` (confiance, focus, credibilite) -- conserve de l'existant
- **Secondaire** : Violet doux `#7C5CFC` (creativite, modernite, differenciation)
- **Mastery** : palette de 4 couleurs semantiques dediees (voir section 9)
- **Mode sombre** : oui, supporte des la V1 -- les 3e le demandent, l'existant le supporte deja

### 6.2 Typographie

Voir le document dedie `typography.md`. Resume :

- **Font principale** : Inter (open source, excellente lisibilite mobile, support FR complet)
- **Echelle** : 7 niveaux (h1 a small), coherente avec l'existant
- **Contenu pedagogique** : traitements specifiques pour definitions, formules, termes-cles

### 6.3 Iconographie

- **Style** : outline avec poids moyen (1.5-2px), coins arrondis
- **Librairie** : Phosphor Icons (open source, coherent, outline + filled + duotone disponibles)
- **Illustrations** : style flat minimaliste pour les empty states et l'onboarding. Pas de personnages realistes ni de style cartoon
- **Mascotte** : NON. Pas de mascotte. Une mascotte est impossible a calibrer pour les 11-15 ans (trop bebe pour les 3e, trop "serieux" pour les 6e). A la place, utiliser des illustrations geometriques/abstraites pour les moments clefs (onboarding, celebrations, empty states)

### 6.4 Tokens de base

- **Border radius** : conserve de l'existant (sm:6, md:12, lg:16, xl:24, full:9999). Le `lg:16` est le radius par defaut des cards -- assez arrondi pour etre moderne sans faire "bulle"
- **Spacing** : 8pt grid conserve (xs:4, sm:8, md:16, lg:24, xl:32, xxl:48) -- standard et fonctionnel
- **Shadows** : systeme a 3 niveaux
  - `shadow.sm` : cards au repos (offset 0/1, blur 3, opacity 0.08)
  - `shadow.md` : cards survolees/pressees (offset 0/2, blur 8, opacity 0.12)
  - `shadow.lg` : modals, popovers (offset 0/4, blur 16, opacity 0.16)
- **Animations** :
  - Duree standard : 200ms (interactions), 300ms (transitions ecrans), 600ms (celebrations)
  - Easing : `ease-out` pour les entrees, `ease-in` pour les sorties
  - Quand animer : transitions mastery, feedback reponse, apparition de cartes
  - Quand NE PAS animer : navigation standard, scroll, saisie texte

---

## 7. Philosophie des composants

Voir le document dedie `component-philosophy.md`. Les composants sont conçus autour de 3 principes :
1. **Un seul systeme, deux contextes** : memes composants pour eleve et parent, layout adapte
2. **L'etat Mastery est omnipresent** : la progression est visible partout sans etre envahissante
3. **Le feedback est immediat** : chaque action a une consequence visuelle claire

---

## 8. Principes d'utilisabilite educative

### Comprehension
- Hierarchie visuelle stricte : 1 titre, 1 action principale, informations secondaires en retrait
- Pas plus de 3 niveaux d'information par ecran
- Les couleurs Mastery sont le seul systeme de couleur semantique -- pas de couleurs arbitraires

### Attention
- Sessions de 10-20 min : interface epuree pendant le quiz (pas de barre de nav, pas de notifications)
- Progression visible en permanence (question 3/10, barre de progression en haut)
- Micro-feedback apres chaque reponse pour maintenir l'engagement

### Progression
- Le pourcentage de maitrise est TOUJOURS visible sur le dashboard
- Les transitions Mastery sont celebrees (animation subtile + message)
- L'historique est accessible mais pas impose (eviter la surcharge cognitive)

### Confiance
- Les erreurs ne sont jamais "rouges" dans le feedback de quiz -- le rouge est reserve aux erreurs systeme
- Le feedback incorrect utilise l'orange doux avec une explication
- Le label UNKNOWN est affiche "NOUVEAU" (pas "inconnu" qui sous-entend un manque)

### Clarte dans l'incertitude
- Les items en attente de validation sont visuellement distincts (badge, bordure pointillee)
- Le score OCR bas est communique via un indicateur de confiance, pas une alarme

---

## 9. Strategie d'equilibre eleve/parent

### Un systeme, deux vues

Pas deux themes separes -- un seul design system avec des **variations de layout** :
- L'espace eleve est la vue par defaut : focus sur l'action (reviser, progresser)
- L'espace parent est accessible via un switch de profil : focus sur l'observation (comprendre, verifier)

### Comment le parent sait qu'il est dans "son" espace

- Header different : nom de l'enfant + avatar en en-tete ("Progression de Hugo")
- Ton de voix : vouvoiement automatique
- Layout data-first : plus de chiffres, de graphiques, moins d'animations
- Palette identique mais predominance de gris/blanc (moins de couleur vive)

### Personnalisation eleve

Niveau minimal au Lot 0, expansible ensuite :
- **Lot 0** : choix dark/light mode
- **MVP** : avatar simple (selection dans une grille), couleur d'accent personnalisable
- **V2** : themes de couleur complets

### Coherence

Le lien entre les deux espaces passe par les **couleurs Mastery** : le parent voit les memes couleurs (vert, bleu, orange, gris) que l'eleve. C'est le langage commun.

---

## 10. Patterns Mastery et gamification

### Representation du Mastery

Voir la section dediee dans `color-system.md` pour les couleurs. La representation visuelle :

#### Barre segmentee (composant principal)
La `MasteryBar` existante est le bon pattern. Ameliorations proposees :
- Ajouter une micro-animation lors des transitions (la section "solid" qui grandit)
- Arrondir les segments interieurs (pas juste les extremites)

#### Badge d'etat
Le `MasteryBadge` existant fonctionne bien (dot + label). Ajustements :
- Renommer "UNKNOWN" en "NOUVEAU" dans l'affichage (meme terme interne)
- Renommer "FRAGILE" en "EN COURS" pour l'affichage eleve (terme moins negatif)
- Garder les termes techniques dans l'espace parent ("Pas encore vu / En cours / Compris / Bien acquis" comme prevu dans Z6-AC29)

#### Progression temporelle
- Mini-graphique sparkline sur la page chapitre : evolution du % maitrise sur 7 jours
- Pas de graphique complexe -- un simple trait qui monte ou descend

#### Eviter que FRAGILE soit percu comme un echec
- La couleur est un ambre chaud (`#F5A623`), pas un rouge-orange
- Le label est "EN COURS" (pas "FRAGILE" pour l'eleve)
- Le feedback associe est "Continue, tu progresses"

#### Rendre SOLID gratifiant sans exces
- Couleur verte franche mais pas neon (`#2ECC71`)
- Animation unique lors de la transition : la barre pulse une fois, un check apparait
- Pas de confettis, pas de son, pas de popup modal

### Gamification dosee

#### Ce qui est acceptable
- **Micro-celebrations** sur les transitions positives (animation + message court)
- **Decompte d'items maitrises** ("12 points bien acquis sur 20")
- **Message d'encouragement contextuel** en fin de session ("3 points de plus vers SOLID")

#### Ce qui est interdit (cf Z6-AC15)
- **Streaks** : explicitement exclus par le produit
- **Classements** : pas de comparaison sociale
- **Points/XP** : pas de monnaie virtuelle
- **Recompenses aleatoires** : pas de "coffres" ou "loot boxes"
- **Notifications de culpabilisation** : jamais "Tu as perdu ta serie"

---

## 11. Strategie de scalabilite

Voir le document dedie `scalability.md`.

---

## 12. Do / Don't

Voir le document dedie `do-dont.md`.

---

## 13. Plan de premiere iteration

### Definir en premier (Lot 0)
1. Palette Mastery finalisee (4 couleurs + light variants)
2. Typographie : integrer Inter, ajuster l'echelle
3. Composants core : MasteryBar (animer), MasteryBadge (renommer labels), Button, Card
4. Ecran de session quiz : c'est la ou l'eleve passe le plus de temps

### Ce qui peut attendre (MVP)
- Illustrations empty states
- Mode avatar/personnalisation
- Graphiques de progression temporelle
- Composants dashboard parent avances

### Prototyper en premier
- L'ecran de session quiz (question + feedback + progression)
- Le dashboard eleve avec la MasteryBar globale
- La transition FRAGILE -> OK (moment emotionnel cle)

### Tester en premier
- Avec un eleve de 4e : "Est-ce que ca fait bebe ?" + "Est-ce que tu comprends ou tu en es ?"
- Avec un parent : "Est-ce que tu fais confiance a cet outil ?" + "Comprends-tu la progression ?"

---

## 14. Moodboard textuel

Voir le document dedie `moodboard.md`.

---

## 15. Recommandation finale

Le design system de Revise Mieux doit etre **sobre, moderne et expressif par les donnees**. La progression Mastery EST l'identite visuelle -- pas besoin d'enjoliver avec des mascotte ou de la gamification lourde.

Les fondations existantes (Colors.ts, Typography.ts, MasteryBar) sont solides et coherentes. Les ajustements proposes sont evolutifs, pas revolutionnaires :
- Affiner la couleur FRAGILE (orange -> ambre chaud)
- Ajouter une couleur secondaire (violet)
- Renommer les labels Mastery cote eleve
- Integrer Inter comme font
- Ajouter des micro-animations sur les transitions Mastery

Le plus grand risque n'est pas d'etre "pas assez fun" -- c'est d'etre "trop scolaire". Le curseur doit pencher vers Spotify/Notion plutot que vers Pronote/Duolingo.
