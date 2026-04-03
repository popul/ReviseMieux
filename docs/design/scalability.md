# Scalabilite -- Revise Mieux

> Ce qui est invariant, ce qui s'ajuste par tranche d'age, strategie d'evolution.

---

## Tableau invariant vs ajustable

| Element | Invariant (tous publics) | Curseur ajustable par age | Evolution college -> lycee -> post-bac |
|---------|------------------------|---------------------------|---------------------------------------|
| **Palette Mastery** | Les 4 couleurs (gris, ambre, bleu, vert) et leur signification. C'est le langage universel du produit. | Aucun ajustement -- les couleurs sont fixes. | Identique. Les couleurs Mastery ne changent jamais. |
| **Palette primaire/secondaire** | Bleu primaire, violet secondaire. | La saturation peut baisser legerement pour un public post-bac (plus pro, moins vif). | College : saturation standard. Lycee : idem. Post-bac : saturation -10%, plus de gris dans la palette. |
| **Typographie (echelle)** | L'echelle 7 niveaux et la grille 8pt. | Les tailles peuvent augmenter/diminuer de 1-2px. La font peut changer. | College : Inter, tailles standard. Lycee : Inter, memes tailles. Post-bac : possibilite de passer a une font serif pour les headings (credibilite). |
| **Animations** | Principes d'animation (durees, easing). | La frequence et l'intensite des celebrations. | College : micro-celebrations a chaque transition Mastery. Lycee : celebrations uniquement pour SOLID. Post-bac : animations minimales, feedback data uniquement. |
| **Ton du copy** | Informatif, jamais culpabilisant, jamais condescendant. | Registre de langue, niveau de tutoiement. | College : tutoiement, ton encourageant, phrases courtes. Lycee : tutoiement, ton plus direct, moins de "bravo". Post-bac : vouvoiement optionnel, ton neutre-professionnel. |
| **Gamification** | Pas de streaks, pas de classement, pas de XP (jamais, pour aucune tranche d'age). | Niveau de feedback positif. | College : decompte d'items maitrises, messages d'encouragement, animation transitions. Lycee : pourcentages et tendances, messages plus sobres. Post-bac : donnees pures, graphiques de progression, pas de messages "motivationnels". |
| **Layout** | Navigation bottom tabs, dashboard en home, cards comme unite d'information. | Densite d'information par ecran. | College : 1-3 cards par ecran, beaucoup d'espace. Lycee : 3-5 cards, plus dense. Post-bac : listes, tableaux, densite desktop-like. |
| **Composants Mastery** | MasteryBar, MasteryBadge, MasteryDot -- meme structure, memes couleurs. | Labels d'affichage. | College : NOUVEAU/EN COURS/COMPRIS/ACQUIS. Lycee : memes labels ou version technique. Post-bac : UNKNOWN/FRAGILE/OK/SOLID (termes techniques). |
| **Spacing/Radius** | Grille 8pt. | Radius default. | College : radius lg (16) -- arrondi, doux. Lycee : idem. Post-bac : radius md (12) ou sm (6) -- plus angulaire, plus "pro". |
| **Iconographie** | Style outline coherent. | Densite d'icones. | College : icones + labels. Lycee : icones seules pour les actions frequentes. Post-bac : texte-first, icones en complement. |

---

## Strategie d'evolution : pas de redesign complet

### Principe fondamental

Le design system est concu comme un ensemble de **curseurs**, pas de **versions**. Passer du college au lycee ne necessite pas de redesign -- c'est un ajustement de parametres.

### Les 5 curseurs

```
1. Densite         leger ◄────────► dense
2. Celebration     expressif ◄────────► sobre
3. Ton             encourageant ◄────────► neutre
4. Rondeur         arrondi ◄────────► angulaire
5. Couleur         colore ◄────────► monochrome
```

### Positions par phase

| Curseur | College (Lot 0 - MVP) | Lycee (V2) | Post-bac (V3) |
|---------|----------------------|------------|----------------|
| Densite | 2/10 | 4/10 | 7/10 |
| Celebration | 6/10 | 4/10 | 2/10 |
| Ton | 7/10 | 5/10 | 3/10 |
| Rondeur | 7/10 | 6/10 | 4/10 |
| Couleur | 6/10 | 5/10 | 3/10 |

### Ce qui ne change JAMAIS (invariants de marque)

1. Les 4 couleurs Mastery et leur signification
2. Le bleu primaire comme couleur de confiance
3. L'absence de gamification lourde (pas de streaks, pas de XP, pas de classements)
4. Le ton non-culpabilisant
5. La separation eleve/parent dans les espaces
6. La MasteryBar comme composant central de progression

### Comment implementer les curseurs techniquement

Le fichier de tokens (`Colors.ts`, `Typography.ts`) peut etre parametre par une variable `audience` :

```typescript
// Concept -- pas a implementer au Lot 0
const config = {
  audience: 'college' | 'lycee' | 'postbac',
  celebrations: audience === 'college' ? 'full' : audience === 'lycee' ? 'minimal' : 'none',
  defaultRadius: audience === 'postbac' ? radius.sm : radius.lg,
  masteryLabels: audience === 'postbac'
    ? { unknown: 'UNKNOWN', fragile: 'FRAGILE', ok: 'OK', solid: 'SOLID' }
    : { unknown: 'NOUVEAU', fragile: 'EN COURS', ok: 'COMPRIS', solid: 'ACQUIS' },
};
```

Au Lot 0, c'est en dur sur "college". Le parametrage viendra plus tard.

---

## Risques de scalabilite et mitigations

| Risque | Mitigation |
|--------|-----------|
| Le design "college" est trop enfantin pour etre reajuste | C'est exactement pourquoi on vise Spotify/Notion comme references, pas Duolingo/Brawl Stars. Un design "jeune-moderne" vieillit bien ; un design "enfantin" ne peut que rajeunir. |
| Les curseurs creent des variations non testees | Au Lot 0, on ne teste qu'une seule configuration. Les curseurs sont documentes mais pas implementes. Chaque nouvelle tranche d'age sera testee avec de vrais utilisateurs. |
| Le parent voit un design "college" et le juge trop simple | L'espace parent est deja plus "sobre" et "data-first". L'ecart entre les espaces est un curseur en soi. |
| Le passage lycee casse l'habitude des utilisateurs existants | Le changement est progressif (parametrable par l'utilisateur, pas impose). L'eleve qui passe en 2nde peut garder le mode "college" s'il prefere. |
