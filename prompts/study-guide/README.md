# Skill `study-guide` — source unique du prompt

> **Source de vérité unique** pour le prompt système de la skill `study-guide` /
> du Lot -1 web-v0. Modifie `system.md` ici, puis régénère les artefacts
> dérivés avec `make sync-skill-prompt` à la racine du dépôt.

## Fichiers

| Fichier | Rôle |
|---|---|
| `system.md` | **Source unique** du prompt système (sans frontmatter Claude). Décrit le pipeline OCR → structuration → fiche, les types d'items transverses, le format de sortie 1→6 + 8 + 9, les règles HTML, et la section formelle « Critères d'acceptance » avec IDs `[AC-CONT-*]`, `[AC-HTML-*]`, `[AC-15-*]`, `[AC-20-*]`. |
| `skill-frontmatter.yaml` | Frontmatter Claude Code (name, description, hooks Stop) requis par la skill. Concaténé avec `system.md` pour produire `.claude/skills/study-guide/SKILL.md`. |
| `README.md` | Ce fichier. |

## Artefacts dérivés (générés, NE PAS éditer à la main)

| Cible | Source | Comment c'est généré |
|---|---|---|
| `.claude/skills/study-guide/SKILL.md` | `skill-frontmatter.yaml` + `system.md` | `make sync-skill-prompt` |
| `web-v0` runtime (`/app/system.md` dans l'image Docker) | `prompts/study-guide/system.md` | `Dockerfile` du Lot -1, `COPY prompts/study-guide/system.md /app/` |

## Workflow d'édition

1. Modifier `prompts/study-guide/system.md`.
2. Lancer `make sync-skill-prompt` — régénère `.claude/skills/study-guide/SKILL.md`.
3. Commiter les deux fichiers ensemble.
4. CI vérifie que `SKILL.md` est bien à jour vis-à-vis de `system.md` (cible `make check-skill-prompt`).

Pour le Lot -1 web-v0, aucun rebuild n'est nécessaire à la modification — le `Dockerfile` consomme `prompts/study-guide/system.md` directement au prochain build d'image.
