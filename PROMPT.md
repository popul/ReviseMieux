# PROMPT - Révise mieux MVP

## Objectif

Implémenter la première version MVP de Révise mieux, un assistant d'étude IA pour collégiens/lycéens. L'application transforme des notes de cours (via OCR) en fiches de révision, quiz interactifs et mindmaps.

## Contexte

- **Mode** : Local uniquement (Docker Compose), sans authentification
- **Langue** : Code, variables, fonctions et commentaires en français
- **Design existant** : Voir `.sop/planning/design/detailed-design.md`
- **Plan détaillé** : Voir `.sop/planning/implementation/plan.md`

## Exigences principales

### Backend (Go + Gin)
- [ ] API REST avec Gin framework
- [ ] Adaptateurs LLM : OpenAI (primary) + Mistral (fallback)
- [ ] Service OCR : images (JPG, PNG, WebP) + PDF (extraction pages)
- [ ] Score de confiance OCR avec zones incertaines
- [ ] Génération : fiches, quiz (configurable), mindmaps, ressources
- [ ] PostgreSQL : cours, fiches, quiz, sessions, mindmaps, ressources, activités
- [ ] Quotas journaliers configurables (env vars)

### Frontend (React + Vite + Tailwind)
- [ ] Migration des prototypes HTML vers React/TypeScript
- [ ] Design system : coral/teal/gold/cream, Fraunces + DM Sans
- [ ] Pages : Dashboard, Scanner, Fiches, Quiz, Mindmap
- [ ] Upload drag-drop avec preview
- [ ] Éditeur OCR avec highlighting des zones incertaines
- [ ] Quiz interactif avec feedback par question
- [ ] Mindmap interactive (react-flow) avec export PNG/SVG
- [ ] Accessibilité WCAG 2.1 AA (taille texte, daltonien, contraste, clavier)

### Infrastructure
- [ ] Docker Compose : PostgreSQL + Backend + Frontend
- [ ] Migrations SQL
- [ ] Makefile avec commandes dev/test/build

## Critères d'acceptation

1. `make dev` démarre l'environnement complet sans erreur
2. Upload d'une image → OCR → texte affiché avec zones incertaines
3. Correction du texte → génération fiches → affichage en mode révision
4. Génération quiz → parcours complet → score affiché
5. Génération mindmap → visualisation interactive → export PNG
6. Dashboard affiche les statistiques globales
7. Quotas bloquent après limite atteinte
8. Navigation clavier complète
9. Tests : unit (Go + React), integration API, E2E (Playwright)
10. Couverture : > 80% backend, > 70% frontend

## Ordre de priorité

1. **OCR** (étapes 1-8) : Infrastructure → DB → Backend → LLM → OCR → Frontend Scanner → Intégration
2. **Fiches** (étapes 9-10)
3. **Quiz** (étapes 11-12)
4. **Enrichissement + Dashboard** (étapes 15-16)
5. **Mindmaps** (étapes 13-14)
6. **Polish** (étapes 17-20) : Quotas, Accessibilité, Tests E2E, Docs

## Fichiers de référence

| Document | Chemin |
|----------|--------|
| Design détaillé | `.sop/planning/design/detailed-design.md` |
| Plan d'implémentation | `.sop/planning/implementation/plan.md` |
| Clarification exigences | `.sop/planning/idea-honing.md` |
| Recherche | `.sop/planning/research/` |
| PRD original | `docs/PRD.md` |
| Prototypes HTML | `frontend/*.html` |

## Notes importantes

- **Fallback LLM** : Si OpenAI échoue (erreur/rate limit), basculer automatiquement sur Mistral
- **Source-grounded** : Les fiches/quiz doivent être générés uniquement à partir du contenu OCR
- **Structured outputs** : Utiliser les JSON schemas OpenAI pour garantir le format des réponses
- **Pas d'auth** : Mode anonyme, toutes les données sont globales à l'instance locale
