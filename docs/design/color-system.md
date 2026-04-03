# Systeme de couleurs -- Revise Mieux

> Palette complete avec justifications, evaluation de l'existant, et systeme Mastery.

---

## Diagnostic de l'existant (Colors.ts)

### Ce qui fonctionne bien
- **Structure** : palette de base + themes light/dark + masteryColors -- architecture saine et extensible
- **Grille de gris** : 9 niveaux (50 a 900) -- suffisant, bien calibre
- **Couleurs semantiques** : success/warning/error + variantes light -- correct
- **masteryColors** : separation explicite des couleurs Mastery -- excellent pattern

### Ce qui doit evoluer
- **Couleur FRAGILE (orange #FF9500)** : trop proche du "warning" systeme. Risque de confusion avec les erreurs. Passer a un ambre plus chaud et moins alarmiste.
- **Couleur SOLID (green #34C759)** : c'est le vert Apple iOS. Fonctionnel mais generique. Un vert legerement plus profond apporterait plus de "satisfaction".
- **Pas de couleur secondaire** : le bleu primaire est seul. Ajouter une secondaire pour les accents, CTA secondaires, elements differenciants.
- **Pas de couleurs "confiance/incertitude"** : manquantes pour les scores OCR et items en validation.
- **Mode sombre** : les variantes dark sont fonctionnelles mais les couleurs light (tintLight: '#1A3A5C') semblent calcules mecaniquement. A affiner.

---

## Palette proposee

### Couleur primaire

| Token | Hex | Usage | Justification |
|-------|-----|-------|---------------|
| `primary.500` | `#4A90D9` | Boutons principaux, liens, accents, barre de nav active | **Conserve.** Bleu moyen qui evoque la confiance et la concentration. Ni froid (corporate) ni vif (enfantin). Bien lisible sur fond blanc et fonce. |
| `primary.400` | `#6BA3E3` | Hover/pressed sur fond clair | |
| `primary.600` | `#3A7BC8` | Hover/pressed sur fond fonce | Conserve de l'existant (blueDark) |
| `primary.100` | `#E6F0FA` | Backgrounds, badges, boutons secondary | Conserve de l'existant (blueLight) |
| `primary.50` | `#F0F6FC` | Fond de section highlight | |

### Couleur secondaire

| Token | Hex | Usage | Justification |
|-------|-----|-------|---------------|
| `secondary.500` | `#7C5CFC` | Accents differenciants, elements creatifs, onboarding | Violet moyen -- apporte de la modernite et de la personnalite. Les ados associent le violet aux apps "creatives" (Figma, Discord). Bon contraste avec le bleu primaire. |
| `secondary.400` | `#9478FD` | Variante claire | |
| `secondary.600` | `#6344E5` | Variante foncee | |
| `secondary.100` | `#F0ECFE` | Background accent | |

### Couleurs semantiques systeme

| Token | Hex | Usage | Note |
|-------|-----|-------|------|
| `success` | `#34C759` | Confirmations systeme, validation OK | Conserve |
| `successLight` | `#E8F9ED` | Background success | Conserve |
| `warning` | `#FF9500` | Alertes systeme, avertissements | Conserve -- DISTINCT de la couleur Mastery FRAGILE |
| `warningLight` | `#FFF3E0` | Background warning | Conserve |
| `error` | `#FF3B30` | Erreurs systeme, echecs techniques | Conserve |
| `errorLight` | `#FFEBEE` | Background error | Conserve |
| `info` | `#4A90D9` | Informations, tips | = primaire |

---

## Couleurs Mastery -- Coeur du produit

Les couleurs Mastery sont le systeme de couleur le plus critique du produit. Elles doivent :
1. Former une **progression visuelle lisible** (du neutre au positif)
2. **Ne pas juger** : UNKNOWN n'est pas "mauvais", FRAGILE n'est pas "en echec"
3. Etre **distinctes des couleurs systeme** (warning, error, success) pour eviter la confusion
4. Fonctionner en **mode clair ET sombre**
5. Etre **lisibles pour les daltoniens** (pas reposer uniquement sur rouge/vert)

### Palette Mastery

| Etat | Hex principal | Hex light (bg) | Label eleve | Label parent | Intention emotionnelle |
|------|--------------|----------------|-------------|--------------|----------------------|
| **UNKNOWN** | `#9E9E9E` | `#F5F5F5` | NOUVEAU | Pas encore vu | Neutre, pas decourageant. C'est un point de depart, pas un manque. |
| **FRAGILE** | `#F5A623` | `#FEF4E0` | EN COURS | En cours | Attention douce. Un ambre chaud (pas orange vif). Communique "ca avance" pas "attention danger". |
| **OK** | `#4A90D9` | `#E6F0FA` | COMPRIS | Compris | Encourageant, progression visible. Le bleu primaire = c'est sur la bonne voie. |
| **SOLID** | `#2ECC71` | `#E3F8ED` | ACQUIS | Bien acquis | Accomplissement, fierte. Un vert franc mais pas neon. Satisfaction calme, pas euphorie. |

### Pourquoi ces choix specifiques

**UNKNOWN = Gris `#9E9E9E`** (conserve, = gray500)
- Le gris est la seule couleur qui communique "pas encore commence" sans connotation negative
- IMPORTANT : ne pas utiliser un bleu pale ou un violet -- ca donnerait une fausse impression de progression
- L'existant utilise gray400 (#BDBDBD) -- passer a gray500 pour un meilleur contraste

**FRAGILE = Ambre `#F5A623`** (modifie, etait orange #FF9500)
- L'orange #FF9500 existant est le meme que la couleur "warning" systeme -- confusion semantique
- Le nouvel ambre est plus chaud, moins "alerte"
- Il evoque le "en progression" plutot que "attention probleme"
- Bonus : il reste distinct du rouge error et du jaune d'avertissement standard

**OK = Bleu `#4A90D9`** (conserve, = primaire)
- Choix excellent dans l'existant : utiliser la couleur primaire pour OK communique "tu es dans la bonne direction"
- Le bleu est aussi la couleur de la confiance -- coherent avec le sentiment vise
- Distinction claire avec le vert SOLID (pas de confusion "est-ce bon ou tres bon ?")

**SOLID = Vert `#2ECC71`** (modifie, etait #34C759)
- Le vert iOS (#34C759) est fonctionnel mais generique
- #2ECC71 est un vert legerement plus profond, plus "riche" visuellement
- Il communique mieux la satisfaction et l'accomplissement
- Il reste bien distinct du vert success systeme tout en etant dans la meme famille

### Ordre visuel et lisibilite

L'ordre des couleurs dans la MasteryBar (de gauche a droite : SOLID > OK > FRAGILE > UNKNOWN) suit une progression de **saturee a desaturee**, ce qui cree une lecture naturelle "les couleurs avancent" meme sans connaitre le systeme.

Pour les daltoniens : le systeme ne repose pas uniquement sur la teinte. La luminosite progresse aussi (gris fonce -> ambre moyen -> bleu moyen -> vert vif), et chaque etat a un label textuel associe.

### Mode sombre

| Etat | Hex dark mode | Hex dark bg |
|------|--------------|-------------|
| UNKNOWN | `#757575` | `#2A2A2A` |
| FRAGILE | `#D4901F` | `#3A2E1A` |
| OK | `#5B9FE0` | `#1A3050` |
| SOLID | `#27AE60` | `#1A3A2A` |

Les couleurs sont legerement desaturees en dark mode pour eviter la fatigue visuelle tout en restant distinctes.

---

## Couleurs de confiance / incertitude

Pour les scores OCR et les items en attente de validation HITL :

| Token | Hex | Usage |
|-------|-----|-------|
| `confidence.high` | `#2ECC71` | Score OCR > 0.9 (pas affiche explicitement, juste pour l'indication interne) |
| `confidence.medium` | `#F5A623` | Score OCR 0.7-0.9 : bordure ambre sur les blocs concernes |
| `confidence.low` | `#E74C3C` | Score OCR < 0.7 : alerte visuelle, bloc marque comme "a verifier" |
| `validation.pending` | `#7C5CFC` | Items en attente de validation HITL : bordure violette pointillee |
| `validation.resolved` | transparent | Items valides : pas de traitement special |

### Pattern visuel "incertitude"

Les items en attente de validation utilisent :
- Bordure pointillee en violet secondaire
- Opacite 0.85 sur le contenu
- Badge "A verifier" en violet
- Ce pattern est commun eleve/parent (meme visuel, ton adapte)

---

## Resume des modifications par rapport a l'existant

| Element | Avant (Colors.ts) | Apres | Raison |
|---------|-------------------|-------|--------|
| UNKNOWN | gray400 `#BDBDBD` | gray500 `#9E9E9E` | Meilleur contraste |
| FRAGILE | orange `#FF9500` | ambre `#F5A623` | Dissocier du warning systeme, ton plus doux |
| OK | blue `#4A90D9` | blue `#4A90D9` | Inchange -- bon choix |
| SOLID | green `#34C759` | green `#2ECC71` | Plus profond, plus satisfaisant |
| Secondaire | (inexistant) | violet `#7C5CFC` | Nouvelle couleur pour differenciation |
| Confiance | (inexistant) | voir tableau | Nouveau systeme pour OCR/HITL |
