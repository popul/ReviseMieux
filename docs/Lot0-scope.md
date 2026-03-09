# Lot 0 — Scope fonctionnel pré-MVP

| | |
|---|---|
| **Version** | 1.0 |
| **Date** | 9 mars 2026 |
| **Total ACs retenus** | 55 sur 185 (30%) |
| **Contexte** | Version locale pour un binôme père-fils, validant la boucle pédagogique fondamentale sur les 4 packs pilotes |

---

## Contexte

Le PRD v1.4 et les 185 ACs décrivent un MVP complet multi-utilisateur. Avant ce MVP, un **Lot 0** est nécessaire : une version locale fonctionnelle pour un binôme père-fils, qui valide la boucle pédagogique fondamentale sur les 4 packs pilotes.

Le Lot 0 tourne sur le réseau local familial, avec la base de données cible, une API backend, et l'app mobile. La réduction porte sur le **périmètre fonctionnel** (nombre d'ACs), pas sur l'architecture — chaque composant doit pouvoir évoluer vers le MVP sans réécriture.

---

## Légende

| Classification | Description |
|---|---|
| **L0-P1** | La boucle core fonctionne : photo → items → session → mastery → progression |
| **L0-P2** | L'expérience est complète pour un usage quotidien réel |
| **L0-P3** | *(réservé pour les corrections post-test avec le fils)* |

---

## Critères de sélection

### Ce qui reste (la boucle core)
- Photo du cahier → extraction d'items de connaissance → questions → scoring → mastery tracking → progression
- Les 4 packs pilotes (HG-INEG, HG-FEOD, SVT-PHOTO, PC-MVD) et les 27 templates
- La machine à états mastery complète (tous les états et transitions)
- Le pipeline J0 (upload → OCR → items → tags → notions)
- La validation HITL basique (le père corrige les items incertains)
- Les sessions (daily, diagnostic, evening_first, mock_exam)
- Les exams avec resserrement des intervalles
- L'onboarding basique (chapitre démo, empty state)

### Ce qui est coupé (reporté au MVP)
- **Multi-utilisateur** : pas de rôle parent séparé, pas de liaison par code, pas de digest parent, pas de script 3 minutes
- **Notifications push** : le fils lance ses sessions manuellement
- **Emploi du temps** : pas de ScheduleSlot, pas de ScheduleException, pas de sessions pre_class automatiques
- **Orchestration de soirée** : pas d'EveningPlan, pas de séquencement multi-matières
- **RGPD** : pas de suppression J+30, pas de consentement parental formel
- **Admin backoffice** : les packs/templates sont en données de référence, pas de CRUD admin
- **Mode vacances, indisponibilités, timezone** : le contexte local rend ces features inutiles
- **Fiches PDF, report papier** : canal papier différé

---

## Z1 — Mastery : 17 ACs retenus sur 28

| AC | Titre | Lot 0 | Simplification |
|---|---|---|---|
| Z1-AC01 | UNKNOWN → FRAGILE progression | P1 | Tel quel |
| Z1-AC02 | FRAGILE → OK progression | P1 | Tel quel |
| Z1-AC03 | OK → SOLID (espacement 24h) | P1 | Tel quel |
| Z1-AC04 | Blocage OK → SOLID sans espacement | P1 | Tel quel |
| Z1-AC05 | SOLID → OK régression sur échec | P1 | Tel quel |
| Z1-AC06 | OK → FRAGILE régression sur échec | P1 | Tel quel |
| Z1-AC07 | FRAGILE reste FRAGILE sur échec | P1 | Tel quel |
| Z1-AC07b | Récupération FRAGILE après régression (cs=0) | P1 | Tel quel |
| Z1-AC07c | Récupération OK après régression (cs<2) | P1 | Tel quel |
| Z1-AC08 | Resserrement proportionnel si exam posé | P2 | Tel quel |
| Z1-AC09 | Indépendance mastery entre items | P1 | Tel quel |
| Z1-AC10 | Score partiel RUBRIC ne déclenche pas progression | P2 | Tel quel |
| Z1-AC11 | NUMERIC sans unité = échec | P2 | Tel quel |
| Z1-AC12 | KEYWORDS scoring partiel (tolérance N-1) | P2 | Tel quel |
| Z1-AC13 | Échec sur item UNKNOWN (pas de sub-UNKNOWN) | P1 | Tel quel |
| Z1-AC14 | Maintien SOLID sur succès continu | P1 | Tel quel |
| Z1-AC15 | Plafond maîtrise OK pour items non validés | P2 | Tel quel |

**Différés au MVP :** Z1-AC16 (micro-célébrations), Z1-AC17 (débrief fin de session), Z1-AC18 (caveat score mock exam), Z1-AC19 (script 3 min), Z1-AC20 (recalcul date exam), Z1-AC21 (descente difficulté après échecs répétés), Z1-AC22 (anti-starvation UNKNOWN), Z1-AC23 (bouton « Voir ma leçon »), Z1-AC24 (reformulation question), Z1-AC25 (réponse partielle encouragée), Z1-AC26 (matrice complète hint/clarification/partiel).

> **Justification :** Les 17 ACs retenus couvrent la machine à états complète et tous les cas de scoring. Les ACs différés sont de la richesse UX (célébrations, débrief, aide contextuelle) et des edge cases avancés (descente de difficulté, anti-starvation). Importants pour l'adoption à l'échelle mais pas pour valider la boucle pédagogique avec un seul utilisateur motivé.

---

## Z2 — Pipeline J0 : 10 ACs retenus sur 14

| AC | Titre | Lot 0 | Simplification |
|---|---|---|---|
| Z2-AC01 | Carte leçon partielle si OCR en cours | P1 | Tel quel |
| Z2-AC02 | Timeout OCR sur page intermédiaire | P2 | Tel quel |
| Z2-AC03 | Photo floue détectée (confidence < 0.3) | P2 | Tel quel |
| Z2-AC04 | Aucun item généré sur une page | P1 | Tel quel |
| Z2-AC05 | Échec génération items (erreur LLM) | P1 | Tel quel |
| Z2-AC06 | Idempotence pipeline au restart | P2 | Tel quel |
| Z2-AC07 | Diagnostic impossible si 0 items valides | P1 | Tel quel |
| Z2-AC08 | Bloc SCHEMA/MAP conservé comme Document image | P2 | Tel quel |
| Z2-AC09 | File validation plafonnée à 8 | P2 | Tel quel |
| Z2-AC10 | Streaming carte leçon (première page dispo) | P1 | Tel quel |

**Différés au MVP :** Z2-AC11 (retry élève pages en échec), Z2-AC12 (rétention crops RGPD J+30), Z2-AC13 (versioning modèle LLM), Z2-AC14 (pipeline incrémental ajout de pages).

> **Justification :** Le pipeline doit être robuste dès le Lot 0 car c'est le point d'entrée du produit. Les différés sont : RGPD (inutile en local), versioning LLM (traçabilité opérationnelle), pipeline incrémental (un re-upload complet suffit au Lot 0).

---

## Z3 — Validation HITL : 8 ACs retenus sur 25

| AC | Titre | Lot 0 | Simplification |
|---|---|---|---|
| Z3-AC01 | Gabarits bloqués sur item validation_required non résolu | P2 | Tel quel |
| Z3-AC02 | Action Confirmer sur ValidationTask | P2 | Tel quel |
| Z3-AC03 | Action Corriger sur ValidationTask | P2 | Le père corrige directement |
| Z3-AC04 | Action « Je ne sais pas » sur ValidationTask | P2 | Tel quel |
| Z3-AC05 | Action Ignorer sur ValidationTask | P2 | Tel quel |
| Z3-AC06 | Chapitre utilisable avec 0 validations faites | P1 | Tel quel |
| Z3-AC09 | Pas de ValidationTask si confidence > 0.85 | P2 | Tel quel |
| Z3-AC10 | Cross-check LLM : item fidèle | P2 | Tel quel |

**Différés au MVP :** Z3-AC07 (régénération cache ciblée), Z3-AC08 (parent max 3/sem), Z3-AC11 (item halluciné détecté), Z3-AC12 (cross-check timeout), Z3-AC13 (doublons intra-chapitre), Z3-AC14 (contradictions), Z3-AC15 (signalement élève), Z3-AC16/17 (détection anomalie taux d'échec), Z3-AC18 (re-check fidelity différé), Z3-AC19 (UX clarification Ignorer), Z3-AC20 (SLA admin), Z3-AC21 (détection anomalie précoce), Z3-AC22 (rétractation), Z3-AC23 (anti-clicking aveugle), Z3-AC24 (récupération items Ignorés), Z3-AC25 (garde-fou template/type).

> **Justification :** La HITL basique est nécessaire car le LLM génère des items de qualité variable. Le père doit pouvoir vérifier et corriger. Les mécanismes avancés de détection (doublons, taux d'échec anormal, rétractation) sont de la robustesse opérationnelle pour un usage à l'échelle.

---

## Z4 — Lazy generation & cache : 5 ACs retenus sur 18

| AC | Titre | Lot 0 | Simplification |
|---|---|---|---|
| Z4-AC06 | Session interrompue reprise depuis dernier état | P1 | Tel quel |
| Z4-AC07 | Contraintes pack respectées dans composition | P1 | Tel quel |
| Z4-AC08 | Pool vide : dégradation gracieuse | P1 | Message simple |
| Z4-AC10 | Mock exam non bloqué par session daily active | P2 | Tel quel |
| Z4-AC12 | Feedback enrichi après réponse incorrecte | P1 | Bonne réponse affichée + explication courte |

**Différés au MVP :** Z4-AC01/02 (concurrence sessions — 1 seul user), Z4-AC03/04/05 (invalidation cache avancée), Z4-AC09 (cache OCR permanent), Z4-AC11 (variété gabarits anti-monotonie), Z4-AC13 (bouton Passer), Z4-AC14 (petit chapitre < 5 items), Z4-AC15 (anti-lassitude), Z4-AC16 (fallback LLM indisponible), Z4-AC17 (invalidation cache CRUD exam), Z4-AC18 (session consolidation optionnelle).

---

## Z5 — ChapterRevision : 3 ACs retenus sur 12

| AC | Titre | Lot 0 | Simplification |
|---|---|---|---|
| Z5-AC01 | Clé d'identité canonique de l'item | P1 | Tel quel |
| Z5-AC05 | Révision courante unique par chapitre | P1 | Tel quel |
| Z5-AC07 | Items archivés non proposés en session | P1 | Tel quel |

**Différés au MVP :** Z5-AC02 (héritage mastery re-upload), Z5-AC03 (normalisation PIB/habitant), Z5-AC04 (conflit OCR), Z5-AC06 (sessions actives non affectées), Z5-AC08 (historique mastery), Z5-AC09 (normalisation ponctuation), Z5-AC10 (alerte items perdus), Z5-AC11 (ajout incrémental pages), Z5-AC12 (pas de re-OCR existant).

> **Justification :** Au Lot 0, un re-upload complet crée une nouvelle révision. L'héritage de mastery et l'upload incrémental sont des optimisations différées.

---

## Z6 — Schedule & Engagement : 5 ACs retenus sur 54

| AC | Titre | Lot 0 | Simplification |
|---|---|---|---|
| Z6-AC05 | Session evening_first déclenchée après upload | P1 | Proposée manuellement, pas planifiée |
| Z6-AC06 | evening_first : 100% UNKNOWN, difficulté 1 uniquement | P1 | Tel quel |
| Z6-AC11 | Exam multi-chapitre : création et liaison | P2 | Tel quel |
| Z6-AC12 | Mode dégradé sans emploi du temps | P1 | = le mode par défaut du Lot 0 |
| Z6-AC15 | Pas de pénalité maîtrise pour items en retard | P1 | Tel quel |

**Différés au MVP :** 49 ACs — tout l'emploi du temps (Z6-AC01→AC04, AC07→AC10, AC19→AC23, AC47, AC51→AC54), toutes les notifications (AC02/03/04, AC13, AC16/17, AC24→AC27), toute la partie parent (AC33→AC39, AC45, AC50), les exams avancés (AC42, AC46, AC48→AC49), le mode vacances et indisponibilités.

---

## Z7 — Routine de soirée & Orchestration : 2 ACs retenus sur 26

| AC | Titre | Lot 0 | Simplification |
|---|---|---|---|
| Z7-AC17 | Notions : regroupement des items par concept_tag | P1 | Le LLM génère les notions à l'extraction |
| Z7-AC18 | Vue chapitre par notion (accordéon + maîtrise) | P2 | Affichage basique |

**Différés au MVP :** 24 ACs — EveningPlan (AC01→AC16), périmètre exam par notion (AC19→AC20), angles morts (AC21), prédiction interro surprise (AC22→AC23), bouton « S'avancer » (AC24), fiches PDF (AC25), report papier (AC26).

---

## Z8 — Onboarding : 5 ACs retenus sur 8

| AC | Titre | Lot 0 | Simplification |
|---|---|---|---|
| Z8-AC01 | Chapitre démo pré-chargé (cold start) | P1 | Données statiques embarquées |
| Z8-AC02 | Empty state guidé avant premier upload | P1 | Tel quel |
| Z8-AC03 | UX de recovery si premier OCR échoue | P2 | Message clair + retry |
| Z8-AC04 | Écran de progression pendant traitement J0 | P1 | Indicateur de progression basique |
| Z8-AC08 | Séquence d'onboarding déterministe | P2 | Flow linéaire simplifié |

**Différés au MVP :** Z8-AC05 (onboarding parent 3 écrans), Z8-AC06 (digest parent anticipé J+2), Z8-AC07 (invitation parent timing optimal).

---

## Résumé

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

### Par priorité interne

| Priorité | Description | ACs |
|---|---|---|
| **P1** | La boucle fonctionne : photo → items → session → mastery | 30 |
| **P2** | L'expérience est complète pour un usage quotidien | 25 |

---

## Machines à états formelles

Ces machines à états complètent les ACs en formalisant les transitions de statut des entités principales. Elles doivent être respectées dès le Lot 0.

### Session

```
COMPOSING → IN_PROGRESS → COMPLETED
                        → EXPIRED (TTL 72h, 0 réponses)
                        → ABANDONED (TTL 72h, ≥1 réponse)
```
- `COMPOSING` : les questions sont en cours de sélection/génération
- `IN_PROGRESS` : l'élève répond aux questions
- `COMPLETED` : toutes les questions répondues
- Transitions interdites : `COMPLETED → IN_PROGRESS`, `EXPIRED → *`, `ABANDONED → *`

### Page (pipeline J0)

```
UPLOADING → OCR_PENDING → OCR_PROCESSING → PROCESSED → ITEMS_GENERATING → DONE
                                         → FAILED (ocr_timeout | ocr_error)
                                                                          → NO_ITEMS
                                                                          → ITEMS_FAILED
```
- Chaque état est terminal ou a exactement 1-2 successeurs
- Un échec sur une page ne bloque pas les autres (Z2-AC02)

### ChapterRevision

```
PROCESSING → READY (≥1 item valide)
           → PARTIAL (certaines pages en échec, mais ≥1 item)
           → FAILED (0 items valides, Z2-AC07)
```

### ValidationTask

```
PENDING → CONFIRMED (Z3-AC02)
        → CORRECTED (Z3-AC03)
        → UNKNOWN_ANSWER (Z3-AC04 « je ne sais pas »)
        → IGNORED (Z3-AC05)
```

### Mastery (rappel — documenté dans les ACs Z1)

```
UNKNOWN → FRAGILE → OK → SOLID
                  ← OK ← SOLID  (régressions)
          FRAGILE ← OK          (régression)
```
- Pas de retour à UNKNOWN depuis FRAGILE (Z1-AC07)
- SOLID → OK directement, pas FRAGILE (Z1-AC05)

---

## Phases d'implémentation

### Phase 0 — Fondations

Créer les artefacts d'ancrage **avant toute fonctionnalité** :
1. **Schéma de base de données** — traduire le §16 du PRD en schéma exécutable, avec les corrections INC-2 à INC-10 déjà appliquées. Décisions sur les structures de données (arrays vs tables de jointure, JSON vs tables enfants).
2. **Contrat d'interface** — les routes d'échange entre l'app mobile et le serveur, avec les formats de requête/réponse.
3. **Données de référence** — les 27 templates, les 4 packs (HG-INEG, HG-FEOD, SVT-PHOTO, PC-MVD) avec leurs lexiques et concept_tags.
4. **Prompts LLM versionnés** — les prompts du pipeline (extraction items, vérification fidélité, regroupement notions, instanciation questions).
5. **Données de test** — chapitre démo (8 items statiques), fixtures pour les edge cases mastery.

### Phase 1 — Le moteur (Z1 P1 + Z4 core)

Le cœur invisible du produit :
1. Machine à états mastery (Z1-AC01→AC07c, AC09, AC13, AC14) — toutes les transitions et régressions
2. Calcul de `next_due_at` (spaced repetition basique §17.2)
3. Scoring (KEYWORDS, MCQ, SHORT_ANSWER)
4. Composition de session (algorithme 70/20/10, §17.1)
5. Reprise de session interrompue (Z4-AC06)
6. Feedback après réponse (Z4-AC12 simplifié)

> **Critère de validation :** le fils peut jouer une session sur le chapitre démo et voir sa maîtrise évoluer correctement.

### Phase 2 — Le pipeline (Z2 + Z5 + Z7-AC17)

Le point d'entrée du produit :
1. Upload de photos → stockage
2. OCR par page (service externe) → texte brut + confidence
3. Génération d'items par le LLM → Items structurés avec tags
4. Regroupement en Notions (Z7-AC17)
5. Streaming de la carte de leçon (Z2-AC10)
6. Gestion des erreurs : page en échec (Z2-AC02/04/05), 0 items (Z2-AC07)
7. Identité canonique des items (Z5-AC01)

> **Critère de validation :** le fils prend en photo une page de son cahier de Physique, la carte de leçon apparaît avec les items regroupés par notion.

### Phase 3 — L'expérience complète (P2)

Ce qui rend le produit utilisable au quotidien :
1. Validation HITL (Z3-AC01→AC05, AC09, AC10) — le père vérifie les items incertains
2. Plafond mastery items non validés (Z1-AC15)
3. Scoring avancé : RUBRIC (Z1-AC10), NUMERIC avec unité (Z1-AC11), KEYWORDS N-1 (Z1-AC12)
4. Exam + resserrement (Z1-AC08, Z6-AC11)
5. Mock exam (Z4-AC10)
6. Session evening_first (Z6-AC05/AC06)
7. Vue par notion (Z7-AC18)
8. Onboarding complet (Z8-AC01→AC04, AC08)

> **Critère de validation :** le fils utilise l'app quotidiennement pendant 1 semaine sur un vrai chapitre, avec un exam posé. Le père a validé les items incertains. La progression mastery est visible et cohérente.

---

## Vérification finale

Pour valider que le Lot 0 est fonctionnellement complet :

| Scénario | Parcours | ACs vérifiés |
|---|---|---|
| **Cold start** | Ouvrir l'app → chapitre démo visible → session démo → mastery évolue | Z8-AC01/02, Z1-AC01/02/05/06, Z4-AC12 |
| **Upload réel** | Photographier un cahier → pipeline → carte de leçon avec notions | Z2-AC01/04/05/07/10, Z5-AC01, Z7-AC17 |
| **Session quotidienne** | Lancer une session daily → répondre → feedback → reprise si interrompu | Z4-AC06/07/08, Z1-AC01→14 |
| **Validation père** | Voir les items incertains → confirmer/corriger → templates avancés débloqués | Z3-AC01→06/09/10, Z1-AC15 |
| **Exam** | Créer un exam → resserrement intervalles → mock exam | Z6-AC11, Z1-AC08, Z4-AC10 |
| **Edge case mastery** | Vérifier récupération après régression (dead-end fix) | Z1-AC07b/07c |
