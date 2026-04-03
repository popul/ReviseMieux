# Design System -- Revise Mieux

Le design system de Revise Mieux est concu pour resoudre un defi rare : plaire a un collegien de 13 ans (qui decide d'ouvrir l'app chaque soir) ET rassurer un parent de 40 ans (qui decide de payer et d'installer). Le positionnement vise un "outil premium pour jeunes" -- ni jouet, ni logiciel scolaire.

---

## Page interactive

Le design system complet est disponible en version interactive :

**[design-system.html](design-system.html)** -- Page HTML avec tous les composants, couleurs et typographies rendus visuellement.

---

## Documents du design system

| Document | Description |
|----------|-------------|
| [Design System](design-system.md) | Document fondateur : executive summary, analyse du defi, positionnement, recommandations globales |
| [Principes de design](design-principles.md) | 6 principes fondateurs qui guident chaque decision de design |
| [Systeme de couleurs](color-system.md) | Palette complete avec justifications, couleurs Mastery, modes clair/sombre |
| [Typographie](typography.md) | Choix typographiques, echelles, hierarchie |
| [Philosophie des composants](component-philosophy.md) | Principes de construction des composants UI |
| [Do / Don't](do-dont.md) | Exemples concrets de bonnes et mauvaises pratiques |
| [Moodboard](moodboard.md) | References visuelles et inspirations |
| [Scalabilite](scalability.md) | Strategie d'evolution du design system |

---

## Couleurs Mastery

Les couleurs Mastery sont le coeur visuel du produit. Elles communiquent la progression sans juger, encouragent sans mentir, et sont distinctes des couleurs systeme (warning, error, success).

| Etat | Couleur | Hex | Background | Label affiche |
|------|---------|-----|------------|--------------|
| **UNKNOWN** | Gris | `#9E9E9E` | `#F5F5F5` | NOUVEAU |
| **FRAGILE** | Ambre | `#F5A623` | `#FEF4E0` | EN COURS |
| **OK** | Bleu | `#4A90D9` | `#E6F0FA` | COMPRIS |
| **SOLID** | Vert | `#2ECC71` | `#E3F8ED` | ACQUIS |

L'ordre dans la MasteryBar (SOLID > OK > FRAGILE > UNKNOWN) suit une progression de saturee a desaturee, creant une lecture naturelle "les couleurs avancent" meme sans connaitre le systeme.

### Principes des couleurs Mastery

- **UNKNOWN = gris** : communique "pas encore commence" sans connotation negative
- **FRAGILE = ambre** (pas orange) : dissociee du warning systeme, ton doux qui dit "ca avance"
- **OK = bleu primaire** : coherent avec la couleur de confiance du produit, distinct du vert SOLID
- **SOLID = vert profond** : accomplissement et fierte, satisfaction calme plutot qu'euphorie
