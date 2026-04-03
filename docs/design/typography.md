# Typographie -- Revise Mieux

> Echelle typographique, fonts, traitement du contenu pedagogique.

---

## Diagnostic de l'existant (Typography.ts)

### Ce qui fonctionne
- **Echelle a 7 niveaux** (h1, h2, h3, body, bodyBold, caption, captionBold, small) : complete et suffisante
- **Grid spacing 8pt** (xs:4 a xxl:48) : standard et bien dimensionne
- **Radius** (sm:6 a full:9999) : coherent avec un design moderne arrondi
- **Line-heights** : ratios corrects (1.2x pour titres, 1.375x pour body)
- **Poids** : 400 (regular) et 600/700 (semibold/bold) -- suffisant

### Ce qui doit evoluer
- **Pas de font-family specifie** : l'app utilise la font systeme par defaut (SF Pro sur iOS, Roboto sur Android). C'est fonctionnel mais ne cree pas d'identite.
- **Pas de style pour le contenu pedagogique** : definitions, formules, termes-cles n'ont pas de traitement typographique specifique
- **Pas de style overline/label** : utile pour les categories ("PHYSIQUE-CHIMIE", "SESSION 3/10")
- **Taille h1 a 28** : un peu petit pour les titres d'ecran sur les grands telephones recents. 32 serait plus impactant.

---

## Font recommandee

### Font principale : Inter

| Critere | Evaluation |
|---------|-----------|
| Lisibilite mobile | Excellente -- concue pour les ecrans, x-height elevee |
| Caractere distinctif | Moderne, neutre-positif, ni trop froid ni trop personnel |
| Support FR | Complet (accents, ligatures, oe, ae) |
| Licence | SIL Open Font License -- gratuit, sans restriction |
| Poids disponibles | 100 a 900 + italique |
| Performance | Variable font disponible (un seul fichier) |
| Expo/RN | Integrateur via `expo-font` ou `@expo-google-fonts/inter` |

### Pourquoi Inter

- C'est la font de Notion, Linear, Vercel -- des apps que les ados associent au "bien fait" et au "professionnel moderne"
- Elle fonctionne aussi bien en petit (caption 12px) qu'en grand (h1 32px)
- Elle est neutre-positive : ni "fun" (comme Quicksand) ni "corporate" (comme Arial)
- Elle a un support typographique complet pour les caracteres scientifiques courants

### Alternative : sans font custom

Si le budget de chargement est une contrainte au Lot 0, garder la font systeme (SF Pro / Roboto) est acceptable. Ces fonts sont excellentes et n'ajoutent pas de temps de chargement. L'integration d'Inter peut etre faite au MVP.

### Font secondaire : non

Pas besoin d'une font secondaire. Inter en differents poids (400, 600, 700) suffit a creer la hierarchie necessaire. Ajouter une seconde font complexifierait le systeme sans benefice clair pour un MVP.

---

## Echelle typographique

### Proposition ajustee

| Token | Taille | Poids | Line-height | Usage | Modif vs existant |
|-------|--------|-------|-------------|-------|-------------------|
| `display` | 36 | 700 | 42 | NOUVEAU -- Chiffres hero (pourcentage global, score) | Nouveau |
| `h1` | 32 | 700 | 38 | Titres d'ecran ("Bonjour Hugo") | +4px |
| `h2` | 22 | 700 | 28 | Sous-titres, noms de section | Inchange |
| `h3` | 18 | 600 | 24 | Titres de card, noms de chapitre | Inchange |
| `body` | 16 | 400 | 22 | Texte courant, descriptions | Inchange |
| `bodyBold` | 16 | 600 | 22 | Texte courant accentue | Inchange |
| `label` | 14 | 700 | 18 | NOUVEAU -- Categories, overlines ("PHYSIQUE-CHIMIE") | Nouveau |
| `caption` | 14 | 400 | 18 | Texte secondaire, metadata | Inchange |
| `captionBold` | 14 | 600 | 18 | Texte secondaire accentue | Inchange |
| `small` | 12 | 400 | 16 | Sous-textes, horodatages | Inchange |

### Changements proposes

1. **Ajout de `display` (36px)** : pour les chiffres hero type "72% de maitrise" sur le dashboard. Plus impactant visuellement, cree un point focal.
2. **h1 passe de 28 a 32** : les ecrans mobiles sont plus grands qu'avant, 32px est standard pour les titres principaux.
3. **Ajout de `label` (14px, 700, uppercase)** : pour les categories et metadata structurantes. Distinct de `captionBold` par le poids (700 vs 600) et l'usage systematique en uppercase + letter-spacing.

### Letter-spacing

| Token | Tracking |
|-------|----------|
| `display` | -0.5 |
| `h1` | -0.3 |
| `h2` | -0.2 |
| `label` | +1.0 (uppercase) |
| Autres | 0 (defaut) |

---

## Traitement du contenu pedagogique

### Definitions

```
[Terme en bodyBold, couleur primaire]
[Definition en body, couleur text]
```

Le terme est mis en evidence par le poids et la couleur, pas par un encadre lourd. Pattern simple, repetable.

### Formules mathematiques

Pour le Lot 0 (physique-chimie : masse, volume, densite) :

```
[Fond primary.100, padding md, radius md]
[Formule en h3, font monospace (pour l'alignement)]
rho = m / V
```

- Utiliser une font monospace pour les formules (Menlo systeme ou JetBrains Mono)
- Fond colore leger pour delimiter visuellement la formule du texte courant
- Pas de rendu LaTeX au Lot 0 -- texte simple avec symboles Unicode (rho, delta, etc.)

### Termes-cles a memoriser

```
[Badge : fond primary.100, texte primary.500, radius full]
IDH    vassal    chloroplaste
```

Les termes-cles apparaissent comme des badges inline -- visuellement distincts du texte courant, cliquables pour voir la definition.

### Contenu OCR (possiblement bruite)

```
[Si confidence >= 0.85 : rendu normal]
[Si confidence < 0.85 : texte avec opacite 0.7 + icone "?" + bordure pointillee]
```

Le contenu OCR incertain est visuellement "attenue" pour communiquer le doute sans alarmer. Le parent dans l'interface HITL voit le meme traitement avec un bouton "Confirmer/Corriger".

---

## Spacing et radius (confirmation)

### Spacing (inchange)

L'echelle 8pt existante est standard et fonctionnelle :

| Token | Valeur | Usage |
|-------|--------|-------|
| `xs` | 4 | Espaces internes denses (entre dot et label dans MasteryBadge) |
| `sm` | 8 | Padding interne composants compacts, gap entre elements proches |
| `md` | 16 | Padding standard des cards, sections |
| `lg` | 24 | Espacement entre sections, margin des titres |
| `xl` | 32 | Separation majeure entre blocs |
| `xxl` | 48 | Hero spacing (haut d'ecran, avant footer) |

### Radius (inchange)

| Token | Valeur | Usage |
|-------|--------|-------|
| `sm` | 6 | Petits elements (input, badges compacts) |
| `md` | 12 | Boutons, champs de saisie |
| `lg` | 16 | Cards, popovers |
| `xl` | 24 | Grandes cards, bottom sheets |
| `full` | 9999 | Pills, avatars, badges arrondis |

Le radius `lg: 16` est bien calibre : assez arrondi pour etre "moderne" et "doux" (pas angular/corporate) sans etre circulaire/enfantin.
