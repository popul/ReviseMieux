# Acceptance Criteria — Révise Mieux

> **Annexe PRD · Zones à risque vibe coding**
>
> | | |
> |---|---|
> | **Périmètre** | 9 zones critiques · 186 AC en format Given/When/Then (Z0 = Lot -1, Z1-Z8 = Lot 0+) |
> | **Usage** | Chaque zone est un fichier séparé pour un chargement ciblé en contexte agent |

---

## Zones

| # | Zone | Risque | ACs | Fichier |
|---|---|---|---|---|
| Z0 | **Lot -1 (web-v0 one-shot generator)** | Élevé | 15 | [Z0.md](Z0.md) |
| Z1 | Transitions Mastery | Très élevé | 28 | [Z1.md](Z1.md) |
| Z2 | Pipeline J0 — Error paths & timeouts | Très élevé | 18 | [Z2.md](Z2.md) |
| Z3 | Validation HITL — Skip / Ignore / Qualité items | Élevé | 17 | [Z3.md](Z3.md) |
| Z4 | Lazy generation — Concurrence, cache & session experience | Élevé | 17 | [Z4.md](Z4.md) |
| Z5 | ChapterRevision — Identité Item & héritage | Élevé | 11 | [Z5.md](Z5.md) |
| Z6 | Emploi du temps, Notifications, Engagement & Confiance parent | Élevé | 46 | [Z6.md](Z6.md) |
| Z7 | Routine de soirée & Orchestration | Très élevé | 26 | [Z7.md](Z7.md) |
| Z8 | Onboarding & First Use Experience | Très élevé | 8 | [Z8.md](Z8.md) |
| | **Total** | | **186** | |

> **Z0 (Lot -1)** est disjoint de Z1-Z8 : pas de mastery, pas de spaced rep, pas de comptes. Il décrit la qualité d'un site web one-page qui livre une fiche HTML autonome. Il s'appuie sur les critères techniques `[AC-CONT-*]`, `[AC-HTML-*]`, `[AC-15-*]`, `[AC-20-*]` portés par la skill `study-guide` (`prompts/study-guide/system.md`). Toutes les ACs Z0 sont **bloquantes** pour le Lot -1 — pas de répartition P1/P2.

---

## Lot 0 — Périmètre pré-MVP

> **53 ACs retenus sur 171 (31%)** — Version locale pour un binôme père-fils, validant la boucle pédagogique fondamentale sur les 4 packs pilotes.

| Zone | Retenus | Différés | Ratio |
|---|---|---|---|
| Z1 Mastery | 15 | 13 | 54% |
| Z2 Pipeline | 10 | 8 | 56% |
| Z3 HITL | 8 | 9 | 47% |
| Z4 Lazy gen | 5 | 12 | 29% |
| Z5 Revision | 3 | 8 | 27% |
| Z6 Schedule | 5 | 41 | 11% |
| Z7 Routine | 2 | 24 | 8% |
| Z8 Onboarding | 5 | 3 | 63% |
| **Total** | **53** | **118** | **31%** |

**Priorités internes :** P1 (33 ACs) = la boucle fonctionne · P2 (20 ACs) = expérience quotidienne complète.

**Coupé (reporté au MVP) :** multi-utilisateur, notifications push, emploi du temps, orchestration de soirée, RGPD J+30, admin backoffice, mode vacances, fiches PDF.

---


