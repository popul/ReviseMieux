# Scratchpad

## Objectif: Réviser les tests E2E des étapes déjà faites

### Analyse (2026-02-02)

J'ai passé en revue les tests E2E existants. Voici ce que j'ai trouvé:

**Résultat des tests:**
- 31 tests passent ✅
- 15 tests sont marqués `.skip` (tests avec données mock non implémentés)
- 4 tests nécessitent l'API réelle (OCR avec OpenAI)

**Fichiers de tests:**
1. `scanner.spec.ts` - Tests du flux Scanner/OCR
2. `dashboard.spec.ts` - Tests du Dashboard et navigation
3. `accessibilite.spec.ts` - Tests WCAG 2.1 AA
4. `fiches.spec.ts` - Tests des fiches de révision
5. `quiz.spec.ts` - Tests du quiz interactif
6. `mindmap.spec.ts` - Tests des cartes mentales

**Tests qui passent:**
- Interface Scanner (charge, affiche zone upload, keyboard nav)
- Upload de fichiers (sélection, preview, suppression)
- Options de génération (titre, matière, toggle options)
- Dashboard (navigation, titre, skip-to-content)
- Accessibilité (skip link, ARIA, focus visible, taille texte, contraste)
- Pages Fiches/Quiz/Mindmap (chargement basique)

**Tests skippés (TODO pour plus tard):**
- Fiches avec données mock (flip, navigation)
- Quiz avec données mock (session, sélection réponses, feedback)
- Mindmap avec données mock (SVG, zoom, pan)
- Mode daltonien

**Problèmes trouvés et corrigés:**
1. ✅ La dépendance `@playwright/test` était absente du `frontend/package.json` → Ajoutée
2. ⚠️ Node.js 18.3.0 est trop ancien (Playwright 1.58.1 nécessite 18.19+) → Utiliser `nvm use 20`

### Tâches terminées

- ✅ Commit `9f6c42d`: Ajout de `@playwright/test` à `frontend/package.json`

### Résultat final

```
31 passed
15 skipped (tests avec données mock)
4 tests API réelle (non exécutés - nécessitent OpenAI)
```

Les tests E2E couvrent correctement les fonctionnalités implémentées:
- Scanner (interface, upload, options de génération)
- Dashboard et navigation
- Accessibilité WCAG 2.1 AA
- Pages Fiches/Quiz/Mindmap (chargement basique)
