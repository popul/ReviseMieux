# Prompt 3 — Design System intergénérationnel (collégiens + parents)

> **Objectif** : Concevoir les fondations stratégiques d'un design system qui plaît aux collégiens (11-15 ans) ET rassure les parents, sans tomber dans le piège du "trop enfantin" ou du "trop corporate".

---

You are a senior design system architect, product visual strategist, and UX researcher specialized in educational products for teenagers and family-facing digital services.

## Mission

Help me define the initial design system for an educational platform called "Révise Mieux".

## Contexte produit

Révise Mieux est une app mobile React Native 0.81 (Expo 54, iOS + Android) de révision destinée aux collégiens français (11-15 ans), avec une interface parent intégrée.

### Fonctionnalités

- Capture de photos de cahier → structuration par IA
- Sessions de révision adaptatives (questions, scoring, feedback)
- Suivi de progression : Mastery UNKNOWN → FRAGILE → OK → SOLID
- Dashboard parent (maîtrise, alertes, digest)
- Validation HITL (parent valide les items IA)

### Positionnement

- Outil **sérieux mais engageant** — pas un jeu, pas une corvée
- **Efficacité** : 10-20 min/soir suffisent
- **Confiance** : maîtrise mesurable, pas de promesse de "meilleure note"
- **Autonomie** : l'élève sait quoi faire sans qu'on lui dise

### Tension de design centrale

L'app doit être **adoptée par les deux publics** :
- L'élève (11-15 ans) décide s'il ouvre l'app chaque soir
- Le parent (35-50 ans) décide de payer et d'installer

### Ce que je ne veux PAS

- Trop enfantin (mascotte qui danse, couleurs bonbon)
- Trop austère / admin scolaire (gris, tableaux, ton froid)
- Uniquement plaisant pour les parents
- Flashy mais pas crédible
- Qui ressemble à un jouet
- Qui ressemble à un outil de punition/surveillance

### Évolution prévue

| Phase | Cible | Focus design |
|-------|-------|-------------|
| **Lot 0** | Collégiens 4e-3e (13-15 ans) | Design initial |
| **MVP** | Collégiens 6e-3e (11-15 ans) | Ajustements jeunes |
| **V2** | Lycéens (15-18 ans) | Maturation |
| **V3** | Post-bac (prépa, BTS, licence) | Professionnalisation |

Le design system doit être conçu pour le collège mais **scalable** sans refonte totale.

### Fondations design existantes

L'app mobile a déjà des tokens et composants de base. Le prompt doit les évaluer, pas repartir de zéro.

| Fichier | Contenu existant |
|---------|-----------------|
| `mobile/constants/Colors.ts` | Palette (blue, green, orange, red, grays) + `masteryColors` : UNKNOWN=gray400, FRAGILE=orange, OK=blue, SOLID=green |
| `mobile/constants/Typography.ts` | Échelle h1-h3, body, caption, small + spacing 8pt grid (xs:4, sm:8, md:16, lg:24, xl:32, xxl:48) + radius (sm:6, md:12, lg:16, xl:24, full:9999) |
| `mobile/components/MasteryBar.tsx` | Barre de progression Mastery |
| `mobile/components/MasteryBadge.tsx` | Badge état Mastery |

**Tâche 0 du prompt** : diagnostiquer ces fondations et évaluer si elles sont cohérentes avec la stratégie design proposée, ou si elles doivent évoluer.

### Documents de référence

| Document | Chemin | Contenu pertinent |
|----------|--------|-------------------|
| PRD | `docs/PRD.md` | Personas (§4), principes pédagogiques (§2.4), KPIs (§19), risques UI/UX (§21) |
| ACs Zone 5 | `docs/ac/Z5.md` | Dashboard élève, progression visuelle |
| ACs Zone 6 | `docs/ac/Z6.md` | Dashboard parent, alertes, notifications |
| Code mobile | `mobile/` | Composants et tokens existants (voir ci-dessus) |

---

## Tes tâches

### 1. Recadrer le problème de design

Analyse le problème visuel et émotionnel :
- Pourquoi cette cible de design est-elle difficile ?
- Quels pièges visuels éviter absolument ?
- Quelle réponse émotionnelle l'interface doit créer chez l'élève ?
- Quelle réponse émotionnelle chez le parent ?
- Qu'est-ce qui différencie un design "engageant pour ado" d'un design "enfantin" ?

### 2. Définir les personas orientées design

Crée des personas centrées sur les **attentes esthétiques** :

| Persona | Âge | Ce qui attire | Ce qui repousse | Apps de référence aimées |
|---------|-----|---------------|-----------------|-------------------------|
| Collégien 6e (11-12 ans) | | | | |
| Collégien 3e (14-15 ans) | | | | |
| Parent "impliqué" (35-50) | | | | |
| Parent "délégateur" (35-50) | | | | |

Pour chaque persona :
- 3 apps qu'il/elle utilise et aime visuellement (et pourquoi)
- 3 patterns visuels qui le/la rebuteraient
- Le curseur acceptable entre "fun" et "sérieux"

Identifie la **zone de confort partagée** : quels codes visuels fonctionnent pour les 4 personas simultanément ?

### 3. Définir le positionnement de design

Articule clairement où "Révise Mieux" se situe sur ces axes :

```
playful ◄──────────────► serious
youthful ◄──────────────► childish
modern ◄──────────────► gimmicky
educational ◄──────────────► institutional
reassuring ◄──────────────► controlling
```

Justifie chaque positionnement.

### 4. Définir les principes émotionnels et tonals

Spécifie :
- **Ton émotionnel** de l'interface
- **Feeling de marque** (en 3 mots)
- **Personnalité visuelle** (si l'app était une personne)
- **Ton des interactions** (micro-copy, notifications, erreurs)
- **Ton du copywriting UI** (tutoiement ? vouvoiement ? neutre ?)

Le ton doit fonctionner pour élèves ET parents.

### 5. Définir les design tokens fondamentaux

#### Couleurs

- **Primaire** + justification (énergie ? confiance ? focus ?)
- **Secondaire** + usage
- **Couleurs sémantiques** : succès, erreur, warning, info
- **Couleurs de Mastery** — critiques car au coeur du produit :
  - UNKNOWN → quelle couleur ? (neutre, pas décourageant)
  - FRAGILE → quelle couleur ? (attention douce, pas alarme)
  - OK → quelle couleur ? (progression, encourageant)
  - SOLID → quelle couleur ? (accomplissement, fierté)
- **Couleurs confiance/incertitude** (pour les scores OCR, items à valider)
- **Mode sombre** : oui/non ? Justification. Si oui, stratégie.

#### Typographie

- Font principale (body) — critères : lisibilité mobile, caractère, licence, support caractères FR
- Font secondaire (headings) si nécessaire
- Échelle typographique (tailles, line-heights, poids)
- Traitement spécifique pour le contenu pédagogique :
  - Formules mathématiques
  - Définitions (terme + explication)
  - Termes-clés à mémoriser
  - Contenu OCR (possiblement bruité)

#### Iconographie et illustrations

- Style d'icônes (outline, filled, duotone ?) + justification
- Librairie recommandée ou custom ?
- Illustrations : style, fréquence, quand les utiliser (empty states, onboarding, récompenses)
- **Mascotte : oui ou non ?** Si oui, quel registre (compagnon discret vs personnage central) ? Si non, pourquoi ?

#### Tokens de base

- Border radius (sharp, rounded, pill ?) — impact sur la perception d'âge
- Spacing scale (8pt grid ?)
- Shadow/elevation system
- Principes d'animation (durées, easing, quand animer, quand ne pas animer)

### 6. Définir la philosophie des composants

Décris comment les composants principaux doivent se comporter et être perçus :

| Composant | Feeling visé | Pièges à éviter |
|-----------|-------------|-----------------|
| Boutons | | |
| Cards (chapitres, items) | | |
| Indicateurs de progression | | |
| Éléments de quiz | | |
| Alertes / banners | | |
| Champs de saisie | | |
| Composants de capture (photo) | | |
| Composants de review/correction | | |
| Blocs de contenu pédagogique | | |
| Composants de réassurance parent | | |

Pour chaque composant, donne des **principes**, pas juste des noms.

### 7. Adresser la dimension éducative

Ce n'est pas une app consumer générique. Le design system doit supporter :
- **Compréhension** : hiérarchie visuelle claire, pas de surcharge cognitive
- **Attention** : maintien du focus sur 15-20 min sans distraction
- **Progression** : sentiment d'avancement visible et motivant
- **Confiance** : l'élève se sent capable, pas jugé
- **Clarté dans l'incertitude** : quand l'OCR doute, quand le score est limite

### 8. Adresser le problème dual audience explicitement

Décris comment un même design system peut :
- attirer les élèves (couleurs, animations, ton)
- rassurer les parents (structure, data, crédibilité)
- rester cohérent (pas de "schizophrénie visuelle")

Questions à trancher :
- Mêmes composants avec variations de layout ? Ou deux thèmes ?
- Comment le parent sait qu'il est dans "son" espace ?
- Quel niveau de personnalisation pour l'élève ? (thème, avatar, couleur ?)

### 9. Patterns spécifiques Révise Mieux

#### Représentation du Mastery

- Comment visualiser UNKNOWN → FRAGILE → OK → SOLID de manière motivante ?
- Barres, jauges, étoiles, autre chose ?
- Comment éviter que FRAGILE soit perçu comme un échec ?
- Comment rendre SOLID gratifiant sans être over-the-top ?
- Comment montrer la progression temporelle (pas juste l'état actuel) ?

#### Gamification dosée

- Quel niveau est acceptable ? (streaks, badges, points ?)
- Qu'est-ce qui motive un collégien SANS décrédibiliser l'outil aux yeux du parent ?
- Quels patterns sont à ÉVITER absolument ? (classements, comparaisons sociales, récompenses aléatoires)

### 10. Planifier la scalabilité

| Élément | Invariant (tous publics) | Curseur ajustable par âge | Comment ça évolue |
|---------|------------------------|---------------------------|-------------------|
| Palette | | | |
| Typographie | | | |
| Animations | | | |
| Ton du copy | | | |
| Gamification | | | |
| Layout | | | |

Comment éviter un redesign complet en passant au lycée ?

### 11. Do / Don't concrets

Donne des exemples explicites avec descriptions visuelles :

| Catégorie | Exemple | Pourquoi c'est un problème |
|-----------|---------|---------------------------|
| Trop enfantin | | |
| Trop froid | | |
| Trop gamifié | | |
| Trop institutionnel | | |
| **Bien équilibré** | | |

### 12. Recommander un point de départ pratique

- Quoi définir en premier ?
- Quoi peut attendre ?
- Quoi prototyper en premier ?
- Quoi tester en premier avec des élèves et des parents ?

---

## Format de sortie attendu

1. **Executive summary**
2. **Analyse du défi de design**
3. **Personas et attentes audiences**
4. **Positionnement de design** (axes)
5. **Principes émotionnels et tonals**
6. **Design tokens** (couleurs, typo, spacing, radius, animations)
7. **Philosophie des composants** (tableau)
8. **Principes d'utilisabilité éducative**
9. **Stratégie d'équilibre élève/parent**
10. **Patterns Mastery et gamification**
11. **Stratégie de scalabilité** (tableau)
12. **Do / Don't** (tableau avec exemples)
13. **Plan de première itération**
14. **Moodboard textuel** : 5 apps/sites de référence avec ce qu'on emprunte et ce qu'on évite
15. **Recommandation finale**

## Barre de qualité

Ta réponse doit être :
- **stratégique** — pas juste une liste de couleurs
- **pratique** — exploitable pour initialiser un vrai design system
- **ciblée 11-15 ans** — pas du design "jeune" générique
- **profondément consciente de la tension parent/élève**
- **utilisable pour démarrer** — pas un traité théorique

Ne réponds pas comme un consultant branding générique. Pense comme la personne qui définit la première version sérieuse du design system d'une app éducative familiale.
