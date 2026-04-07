# Prompts LLM — source de vérité partagée

Ce dossier est la **source unique** des prompts LLM du projet. Il est consommé par :

- Le **backend Go** via `//go:embed` (`embed.go`)
- Le **skill `study-guide`** via `${CLAUDE_PROJECT_DIR}/backend/prompts/...`
- Les scripts de **benchmark** Python

## Cycle de vie d'un prompt

1. **Prototype** — itération libre dans `.claude/skills/study-guide/SKILL.md` (feedback visuel immédiat sur cas réels).
2. **Validation** — vérifié sur N cas de `backend/testdata/benchmark/cases/`.
3. **Figement** — déplacé ici, versionné, hashé SHA256, consommé par les deux côtés.

## Statut des prompts

| Prompt | Fichier | Statut | Version | Consommateurs |
|---|---|---|---|---|
| OCR | `ocr/system.md` | figé | v1.0.0 | `infra/anthropic`, `infra/openaicompat` |
| Structuration | `structuration/system.md` | figé | v1.1.0 | `infra/anthropic`, `infra/openaicompat` |
| Scoring | `scoring/system.md` | figé | v1.0.0 | `infra/openaicompat` |
| Study-guide (pédagogie, sessions, format) | — | prototype | — | `.claude/skills/study-guide/SKILL.md` |

## Règles

- **Toute modification d'un prompt figé** = bump de version dans `internal/infra/llm/prompts.go` + commit avec justification.
- **Ne jamais** dupliquer un prompt entre le skill et ce dossier : si un prompt vit ici, le skill le référence par chemin.
- **Tests** : `internal/infra/llm` expose les hashs SHA256 pour détecter toute dérive silencieuse.
