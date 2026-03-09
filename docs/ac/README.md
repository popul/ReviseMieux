# Acceptance Criteria — Révise Mieux v1.4

> **Annexe PRD v1.4 · Zones à risque vibe coding**
>
> | | |
> |---|---|
> | **Version** | 1.4 |
> | **Date** | 7 mars 2026 |
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

## Historique des évolutions

> | **Évolutions v1.5.5 vs v1.5.4** | 6 corrections cross-zones : Z1-AC26 matrice inclut exception consolidation_optional (Z4-AC18) y compris protection cs, Z1-AC19 script 3 min exclut chapitre démo (Z8-AC01), Z6-AC50 digest anticipé fusionné multi-enfants (Z8-AC06), Z6-AC33 digest pré-contrôle respecte notion_ids (Z7-AC19), Z1-AC07b/c ajout mention Z1-AC08 resserrement, Z4-AC18 protège cs en consolidation_optional |
| **Évolutions v1.5.4 vs v1.5.3** | Audit de cohérence fonctionnelle : +2 AC zone Z1 (AC07b/AC07c dead end recovery après régression), +1 AC zone Z8 (AC08 séquence onboarding déterministe). Corrections : Z1-AC01/02/03/04 alignés sur `last_success_at` au lieu de `last_review_at` pour le gating 24h, convention timestamps globale en header Z1, exception `consolidation_optional` (Z4-AC18) dans header Z1, Z1-AC08 interaction avec indisponibilités (Z6-AC54) + notion_ids (Z7-AC19), Z7-AC16 GIVEN corrigé pour chapitre démo (Z8-AC01) + hors fenêtre soirée, Z7-AC15 champ standardisé `routine_completed_enabled`, Z8-AC01 items démo exclus du pool daily/pre_class/EveningPlan + archivage différé post-pipeline, Z8-AC05 fallback si parent link avant session, Z8-AC06 GIVEN aligné + dedup avec digest pré-contrôle (Z6-AC33), Z8-AC07 restreint aux vrais chapitres |
| **Évolutions v1.5.3 vs v1.5.2** | +7 AC zone Z8 « Onboarding & First Use Experience » : chapitre démo cold start (AC01), empty state guidé (AC02), UX recovery premier OCR (AC03), écran progression J0 (AC04), onboarding parent 3 écrans (AC05), digest parent anticipé J+2 (AC06), timing invitation parent post-1re session (AC07) |
| **Évolutions v1.5 vs v1.4.1** | +4 AC zone Z1 : accès leçon contextuel pendant question (AC23), reformulation « je ne comprends pas » (AC24), scoring de réponse partielle (AC25), matrice transition maîtrise avec hint/clarification/partiel (AC26). +9 AC zone Z6 : onboarding parent et liaison élève (AC45), comportement jour d'examen (AC46), mode dégradé par matière (AC47 réécrit), multi-exam overlapping reset+recompression (AC48), mock exam J-3 auto + à la demande (AC49), parent multi-enfants notif par enfant (AC50), timezone locale auto (AC51), zone scolaire et calendrier vacances intégré (AC52), mode vacances jours + créneau (AC53). AC01 réécrit : saisie emploi du temps contextuelle au premier upload d'une matière (grille jour/période inline). Indisponibilité récurrente sport/activités (AC54). +26 AC zone Z7 : +3 nouveaux — bouton « S'avancer » weekend (AC24), fiches PDF imprimables pour heures d'étude (AC25), report résultats papier par checklist (AC26) « Routine de soirée & Orchestration ». Orchestration : EveningPlan (AC01), dashboard soirée (AC02), séquencement multi-matières (AC03), estimation durée (AC04), état « fini pour ce soir » (AC05), guidage capture in-app (AC06), séquencement sessions (AC07), mode express (AC08), complétion partielle (AC09), rien à faire (AC10), week-end (AC11), capture cours demain (AC12), devoirs (AC13), arc émotionnel (AC14), notif parent routine (AC15), onboarding 1re soirée (AC16). Hiérarchie contenu & exam : Notions par concept_tag (AC17), vue chapitre par notion (AC18), périmètre exam par notion (AC19), auto-suggestion exam (AC20), vue angles morts (AC21), prédiction interro surprise (AC22), alerte fragile × non testée (AC23) |
> | **Évolutions v1.4.1 vs v1.4** | +5 AC upload incrémental : ajout de pages sans nouvelle révision (Z5-AC11), pas de re-OCR des pages existantes (Z5-AC12), pipeline incrémental (Z2-AC14), session evening_first incrémentale (Z6-AC43), explication dilution maîtrise dashboard (Z6-AC44) |
> | **Évolutions v1.4 vs v1.3** | +15 AC : confiance parent & RGPD (rétention crops Z2, score mock exam + exclusion script 3 min Z1, digest standardisé + signalement OCR parent + feedback résolution admin + labels maîtrise traduits Z6) + session experience (variété gabarits + feedback enrichi + bouton passer + petit chapitre Z4) + intégrité données (rétractation validation erronée + anti-clicking aveugle Z3, versioning LLM Z2) + UX résilience (progression globale + archivage chapitre + persistance réseau Z6) + robustesse planning (recalcul intervalles sur modif date exam Z1, anti-lassitude questions Z4) + anti-frustration élève (feedback explicatif blocage 24h Z1, descente difficulté échecs répétés Z1, fallback LLM indisponible Z4, récupération items ignorés Z3) + anti-silent-failures (anti-starvation items UNKNOWN Z1, garde-fou template/type Z3, invalidation cache exam Z4) + correctifs modèle (session all-SOLID Z4, alerte items perdus re-upload Z5, multi-exam par chapitre Z6, fix Chapter.exam_ids[] pluriel) |
> | **Évolutions v1.3 vs v1.2** | +9 AC anti-désengagement : plafond maîtrise items non validés (Z1), micro-célébrations + débrief session (Z1), retour en douceur après absence + cycle post-exam + rampe diagnostic + digest pré-contrôle + anti alert-fatigue parent (Z6) |
> | **Évolutions v1.2 vs v1.1** | +8 AC couvrant les angles morts identifiés : retry élève pipeline (Z2), re-vérification fidelity timeout + UX validation + SLA admin + détection précoce (Z3), normalisation ponctuation OCR (Z5), ré-engagement inactivité + alerte exams simultanés (Z6) |

