# Acceptance Criteria — Révise Mieux

> **Annexe PRD · Zones à risque vibe coding**
>
> | | |
> |---|---|
> | **Périmètre** | 8 zones critiques · 185 AC en format Given/When/Then |
> | **Usage** | Chaque zone est un fichier séparé pour un chargement ciblé en contexte agent |

---

## Zones

| # | Zone | Risque | ACs | Fichier |
|---|---|---|---|---|
| Z1 | Transitions Mastery | Très élevé | 28 | [Z1.md](Z1.md) |
| Z2 | Pipeline J0 — Error paths & timeouts | Très élevé | 14 | [Z2.md](Z2.md) |
| Z3 | Validation HITL — Skip / Ignore / Qualité items | Élevé | 25 | [Z3.md](Z3.md) |
| Z4 | Lazy generation — Concurrence, cache & session experience | Élevé | 18 | [Z4.md](Z4.md) |
| Z5 | ChapterRevision — Identité Item & héritage | Élevé | 12 | [Z5.md](Z5.md) |
| Z6 | Emploi du temps, Notifications, Engagement & Confiance parent | Élevé | 54 | [Z6.md](Z6.md) |
| Z7 | Routine de soirée & Orchestration | Très élevé | 26 | [Z7.md](Z7.md) |
| Z8 | Onboarding & First Use Experience | Très élevé | 8 | [Z8.md](Z8.md) |
| | **Total** | | **185** | |

---

## Lot 0 — Périmètre pré-MVP

> **55 ACs retenus sur 185 (30%)** — Version locale pour un binôme père-fils, validant la boucle pédagogique fondamentale sur les 4 packs pilotes.

| Zone | Retenus | Différés | Ratio |
|---|---|---|---|
| Z1 Mastery | 17 | 11 | 61% |
| Z2 Pipeline | 10 | 4 | 71% |
| Z3 HITL | 8 | 17 | 32% |
| Z4 Lazy gen | 5 | 13 | 28% |
| Z5 Revision | 3 | 9 | 25% |
| Z6 Schedule | 5 | 49 | 9% |
| Z7 Routine | 2 | 24 | 8% |
| Z8 Onboarding | 5 | 3 | 63% |
| **Total** | **55** | **130** | **30%** |

**Priorités internes :** P1 (33 ACs) = la boucle fonctionne · P2 (22 ACs) = expérience quotidienne complète.

**Coupé (reporté au MVP) :** multi-utilisateur, notifications push, emploi du temps, orchestration de soirée, RGPD J+30, admin backoffice, mode vacances, fiches PDF.

---


