# Résumé du projet - Révise mieux MVP

## Vue d'ensemble

Ce document résume le processus de planification pour la première version MVP de Révise mieux, un assistant d'étude alimenté par l'IA pour les collégiens et lycéens.

## Artifacts créés

| Fichier | Description |
|---------|-------------|
| `rough-idea.md` | Concept initial du MVP |
| `idea-honing.md` | 18 questions/réponses de clarification des exigences |
| `research/codebase-analysis.md` | Analyse de l'état actuel du code |
| `research/external-best-practices.md` | Recherche sur OCR et IA éducative |
| `design/detailed-design.md` | Design complet (architecture, API, modèles, tests) |
| `implementation/plan.md` | Plan en 20 étapes avec checklist |

## Décisions clés

### Stack technique
- **Backend** : Go 1.21+ avec Gin
- **Frontend** : React 18 + Vite + TypeScript + Tailwind CSS
- **Database** : PostgreSQL 16
- **LLM/OCR** : OpenAI GPT-4o (primary) + Mistral (fallback)
- **Mindmaps** : react-flow
- **Tests E2E** : Playwright

### Fonctionnalités MVP (P0)
1. OCR de cours (images + PDF, max 10 pages)
2. Génération de fiches de révision
3. Quiz interactifs configurables
4. Mindmaps interactives
5. Enrichissement de contenu (suggestions LLM)
6. Dashboard avec statistiques globales

### Caractéristiques spéciales
- **Sans authentification** : Mode anonyme, données locales
- **Accessibilité complète** : WCAG 2.1 AA (taille texte, daltonien, contraste, clavier)
- **Quotas configurables** : Limites journalières via env vars
- **Fallback LLM** : Bascule automatique OpenAI → Mistral
- **Code en français** : Variables, fonctions, commentaires, docs

### Ordre de priorité
1. OCR + affichage du texte
2. Génération de fiches
3. Génération de quiz
4. Enrichissement de contenu
5. Mindmaps

## Plan d'implémentation

20 étapes organisées en phases :

| Phase | Étapes | Description |
|-------|--------|-------------|
| Infrastructure | 1-3 | Docker, DB, backend base |
| LLM | 4 | Adaptateurs OpenAI + Mistral |
| OCR | 5, 7-8 | Service OCR + intégration frontend |
| Fiches | 9-10 | Service + page fiches |
| Quiz | 11-12 | Service + page quiz |
| Mindmap | 13-14 | Service + page mindmap |
| Enrichissement | 15 | Ressources complémentaires |
| Dashboard | 16 | Statistiques |
| Polish | 17-20 | Quotas, accessibilité, tests, docs |

## Prochaines étapes

1. **Revoir le plan** : Lire `implementation/plan.md` en détail
2. **Commencer l'étape 1** : Configuration Docker Compose
3. **Itérer** : Chaque étape se termine par une démo fonctionnelle

## Risques identifiés

| Risque | Mitigation |
|--------|------------|
| Qualité OCR manuscrit | Score de confiance + correction manuelle |
| Coûts API | Quotas configurables + fallback Mistral |
| Complexité mindmap | Utiliser react-flow (lib mature) |
| Tests flaky | Playwright + CI robuste |

## Métriques de succès

- OCR : < 3s/page, > 85% précision manuscrit
- Génération : < 20s pour fiches/quiz
- Accessibilité : Lighthouse > 90
- Tests : Couverture > 80% backend, > 70% frontend
