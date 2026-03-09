# Classification MVP — Revise Mieux

| | |
|---|---|
| **Total ACs** | 155 (8 zones) · dont **53 Lot 0** (33 P1 + 20 P2) |

---

## Légende

| Classification | Description |
|---|---|
| **MVP Core** | Le produit ne fonctionne pas sans. Bloquant pour le lancement. |
| **MVP Hardening** | Non bloquant jour 1, mais churn mesurable dans les 2-3 premières semaines. Fortement recommandé au lancement. |
| **Post-MVP** | Raffinements et maturité opérationnelle. Peut attendre mois 2-3. |

| Lot 0 | Description |
|---|---|
| **P1** | La boucle fonctionne. Version locale minimale père-fils. |
| **P2** | Expérience quotidienne complète sur les 4 packs pilotes. |
| **—** | Différé au MVP complet. |

---

## Z1 — Transitions Mastery (26 ACs)

| AC | Titre | Classification | Lot 0 |
|---|---|---|---|
| Z1-AC01 | UNKNOWN → FRAGILE progression | MVP Core | P1 |
| Z1-AC02 | FRAGILE → OK progression | MVP Core | P1 |
| Z1-AC03 | OK → SOLID (espacement 24h) | MVP Core | P1 |
| Z1-AC04 | Blocage OK → SOLID sans espacement | MVP Core | P1 |
| Z1-AC05 | SOLID → OK régression sur échec | MVP Core | P1 |
| Z1-AC06 | OK → FRAGILE régression sur échec | MVP Core | P1 |
| Z1-AC07 | FRAGILE reste FRAGILE sur échec | MVP Core | P1 |
| Z1-AC07b | Récupération FRAGILE après régression (cs=0, réponse correcte) | MVP Core | P1 |
| Z1-AC07c | Récupération OK après régression (cs<2, réponse correcte) | MVP Core | P1 |
| Z1-AC08 | Resserrement proportionnel si exam posé | MVP Core | P2 |
| Z1-AC09 | Indépendance des mastery entre items | MVP Core | P1 |
| Z1-AC10 | Scoring par type de question (RUBRIC, NUMERIC, KEYWORDS) | MVP Core | P2 |
| Z1-AC11 | Échec sur item UNKNOWN (pas de sub-UNKNOWN) | MVP Core | P1 |
| Z1-AC12 | Maintien SOLID sur succès continu | MVP Core | P1 |
| Z1-AC13 | Plafond maîtrise OK pour items restreints aux templates simples | MVP Core | P2 |
| Z1-AC14 | Micro-célébrations transitions positives | MVP Hardening | — |
| Z1-AC15 | Débrief fin de session | MVP Hardening | — |
| Z1-AC16 | Caveat score mock exam items non validés | MVP Hardening | — |
| Z1-AC17 | Script 3 min : exclure items sous investigation | MVP Hardening | — |
| Z1-AC18 | Changement date exam → recalcul intervalles | MVP Hardening | — |
| Z1-AC19 | Descente de difficulté après échecs répétés | **MVP Core** | — |
| Z1-AC20 | Détection starvation items UNKNOWN & rattrapage | MVP Hardening | — |
| Z1-AC21 | Bouton « Voir ma leçon » contextuel pendant une question | **MVP Core** | — |
| Z1-AC22 | Bouton « Je ne comprends pas la question » : reformulation | **MVP Core** | — |
| Z1-AC23 | Réponse partielle encouragée : scoring graduel | **MVP Core** | — |
| Z1-AC24 | Matrice de transition maîtrise avec hint, clarification et score partiel | **MVP Core** | — |

> **Note :** Z1-AC19 reclassifié Post-MVP → MVP Core. La descente de difficulté après 3 échecs consécutifs est indispensable pour le public cible (11-15 ans) : sans elle, le cycle de frustration sur un item difficile provoque l'abandon dès la première semaine. Le renvoi vers la carte de leçon et la pause J+2 après 5 échecs transforment un moment d'échec en apprentissage.

---

## Z2 — Pipeline J0 — Error paths & timeouts (14 ACs)

| AC | Titre | Classification | Lot 0 |
|---|---|---|---|
| Z2-AC01 | Carte leçon partielle si OCR en cours | MVP Core | P1 |
| Z2-AC02 | Timeout OCR sur page intermédiaire | MVP Core | P2 |
| Z2-AC03 | Photo floue détectée (confidence < 0.3) | MVP Core | P2 |
| Z2-AC04 | Aucun item généré sur une page | MVP Core | P1 |
| Z2-AC05 | Échec génération items (erreur LLM) | MVP Core | P1 |
| Z2-AC06 | Idempotence pipeline au restart | MVP Core | P2 |
| Z2-AC07 | Diagnostic impossible si 0 items valides | MVP Core | P1 |
| Z2-AC08 | Bloc SCHEMA/MAP conservé comme Document image | MVP Core | P2 |
| Z2-AC09 | File validation plafonnée à 8 | MVP Core | P2 |
| Z2-AC10 | Streaming carte leçon (première page dispo) | MVP Core | P1 |
| Z2-AC11 | Retry élève pour pages en échec de génération | MVP Hardening | — |
| Z2-AC12 | Rétention crops alignée RGPD (J+30) | MVP Core | — |
| Z2-AC13 | Versioning modèle LLM pour reproductibilité | **MVP Core** | — |
| Z2-AC14 | Pipeline incrémental pour ajout de pages | **MVP Core** | — |

> **Note :** Z2-AC13 reclassifié Post-MVP → MVP Core. Pour un produit éducatif destiné à des mineurs, la traçabilité des modèles LLM utilisés est une exigence de qualité et de conformité dès le lancement. Sans versioning, un changement de modèle silencieux peut dégrader la qualité des items sans diagnostic possible.

---

## Z3 — Validation HITL — Skip / Ignore / Quality (17 ACs)

| AC | Titre | Classification | Lot 0 |
|---|---|---|---|
| Z3-AC01 | Gabarits bloqués sur item validation_required non résolu | MVP Core | P2 |
| Z3-AC02 | Action Confirmer sur ValidationTask | MVP Core | P2 |
| Z3-AC03 | Action Corriger sur ValidationTask | MVP Core | P2 |
| Z3-AC04 | Action « Je ne sais pas » sur ValidationTask | MVP Core | P2 |
| Z3-AC05 | Action Ignorer sur ValidationTask | MVP Core | P2 |
| Z3-AC06 | Chapitre utilisable avec 0 validations faites | MVP Core | P1 |
| Z3-AC07 | Régénération cache ciblée après résolution | MVP Core | — |
| Z3-AC08 | Parent actif : max 3 ValidationTasks/semaine | MVP Hardening | — |
| Z3-AC09 | Pas de ValidationTask si confidence > 0.85 | MVP Core | P2 |
| Z3-AC10 | Vérification croisée LLM (fidelity check) | MVP Core | P2 |
| Z3-AC11 | Cohérence intra-chapitre (doublons et contradictions) | MVP Hardening | — |
| Z3-AC12 | Signalement d'erreur par l'élève | MVP Hardening | — |
| Z3-AC13 | Détection anomalie par taux d'échec | MVP Hardening | — |
| Z3-AC14 | Rétractation de validation erronée | MVP Hardening | — |
| Z3-AC15 | Anti-clicking aveugle (détection réponse < 2s) | MVP Hardening | — |
| Z3-AC16 | Récupération items Ignorés par l'élève | MVP Hardening | — |
| Z3-AC17 | Garde-fou template/type pour items restreints | Post-MVP | — |

---

## Z4 — Lazy generation — Concurrence & cache (13 ACs)

| AC | Titre | Classification | Lot 0 |
|---|---|---|---|
| Z4-AC01 | Pas de doublons sur composition simultanée | MVP Core | — |
| Z4-AC02 | Pool de questions non partagé entre sessions actives | MVP Hardening | — |
| Z4-AC03 | Invalidation cache (HITL, pack, exam) | MVP Core | — |
| Z4-AC04 | Session interrompue reprise depuis dernier état | MVP Core | P1 |
| Z4-AC05 | Contraintes pack respectées dans composition lazy | MVP Core | P1 |
| Z4-AC06 | Pool vide : dégradation gracieuse | MVP Core | P1 |
| Z4-AC07 | Mock exam non bloqué par session daily active | MVP Core | P2 |
| Z4-AC08 | Variété gabarits en session (anti-monotonie) | MVP Hardening | — |
| Z4-AC09 | Feedback enrichi après réponse incorrecte | MVP Core | P1 |
| Z4-AC10 | Bouton Passer sans pénalité maîtrise | MVP Hardening | — |
| Z4-AC11 | Session viable sur petit chapitre (< 5 items) | MVP Hardening | — |
| Z4-AC12 | Contenu fallback si LLM indisponible | MVP Hardening | — |
| Z4-AC13 | État « tout à jour » : session consolidation optionnelle | Post-MVP | — |

---

## Z5 — ChapterRevision — Item identity & Mastery inheritance (10 ACs)

| AC | Titre | Classification | Lot 0 |
|---|---|---|---|
| Z5-AC01 | Clé d'identité canonique de l'item | MVP Core | P1 |
| Z5-AC02 | Héritage mastery sur re-upload (même item) | MVP Core | — |
| Z5-AC03 | Conflit OCR : terme ambigu entre révisions | MVP Core | — |
| Z5-AC04 | Révision courante unique par chapitre | MVP Core | P1 |
| Z5-AC05 | Sessions actives non affectées par nouvelle révision | MVP Core | — |
| Z5-AC06 | Items archivés non proposés en session | MVP Core | P1 |
| Z5-AC07 | Normalisation insensible ponctuation (I.D.H. = IDH) | MVP Core | — |
| Z5-AC08 | Alerte items haute maîtrise absents de la nouvelle révision | MVP Hardening | — |
| Z5-AC09 | Ajout incrémental de pages sans nouvelle révision | **MVP Core** | — |
| Z5-AC10 | Pas de re-OCR des pages existantes lors d'un ajout | **MVP Core** | — |

---

## Z6 — Schedule, Notifications, Engagement & Parent trust (43 ACs)

| AC | Titre | Classification | Lot 0 |
|---|---|---|---|
| Z6-AC01 | Saisie emploi du temps contextuelle au premier upload d'une matière | **MVP Core** | — |
| Z6-AC02 | Notifications schedule-based (capture ou review) | MVP Core | — |
| Z6-AC03 | Max 2 notifications par soirée | MVP Core | — |
| Z6-AC04 | Session evening_first déclenchée après upload | MVP Core | P1 |
| Z6-AC05 | evening_first : 100% UNKNOWN, difficulté 1 uniquement | MVP Core | P1 |
| Z6-AC06 | Session pre_class la veille de chaque cours | MVP Core | — |
| Z6-AC07 | Fusion pre_class dans daily si même soirée | MVP Core | — |
| Z6-AC08 | pre_class : scope multi-chapitres par matière | MVP Core | — |
| Z6-AC09 | Mock exam multi-chapitres couvre tous les chapitres liés | MVP Core | — |
| Z6-AC10 | Exam multi-chapitre : création et liaison | MVP Core | P2 |
| Z6-AC11 | Mode dégradé sans emploi du temps (global et par matière) | MVP Core | P1 |
| Z6-AC12 | Pas de pre_class si aucun chapitre actif dans la matière | MVP Core | — |
| Z6-AC13 | Pas de pénalité maîtrise pour items en retard | MVP Core | P1 |
| Z6-AC14 | Rappel unique le lendemain pour session manquée | MVP Hardening | — |
| Z6-AC15 | Pas de mécanique de streak | MVP Core | — |
| Z6-AC16 | Exceptions d'emploi du temps (annulation, déplacement) | MVP Hardening | — |
| Z6-AC17 | Notification parent : changement emploi du temps (temps réel) | MVP Hardening | — |
| Z6-AC18 | Notification parent : session manquée (lendemain matin) | MVP Hardening | — |
| Z6-AC19 | Notification parent : inactivité prolongée (3 jours) | MVP Hardening | — |
| Z6-AC20 | Opt-out parent par catégorie de notification | MVP Core | — |
| Z6-AC21 | Ré-engagement progressif après 7+ jours d'inactivité | Post-MVP | — |
| Z6-AC22 | Session « retour en douceur » après absence prolongée | MVP Hardening | — |
| Z6-AC23 | Cycle post-exam : archivage progressif des intervalles | MVP Hardening | — |
| Z6-AC24 | Diagnostic initial : rampe de difficulté progressive | MVP Hardening | — |
| Z6-AC25 | Digest parent : supplément pré-contrôle à J-3 | MVP Hardening | — |
| Z6-AC26 | Notification parent : résumé hebdo sessions manquées (anti-fatigue) | MVP Hardening | — |
| Z6-AC27 | Digest parent hebdo : contenu 5 sections standardisées | MVP Core | — |
| Z6-AC28 | Digest parent : signalement OCR échoué | MVP Hardening | — |
| Z6-AC29 | Labels maîtrise traduits pour les parents | MVP Core | — |
| Z6-AC30 | Vue progression globale cross-chapitres | MVP Hardening | — |
| Z6-AC31 | Résilience réseau : persistance optimiste des réponses | MVP Hardening | — |
| Z6-AC32 | Multi-exam par chapitre + resserrement exam le plus proche | MVP Core | — |
| Z6-AC33 | Session evening_first incrémentale après ajout de pages | **MVP Core** | — |
| Z6-AC34 | Dashboard : explication dilution maîtrise après ajout pages | **MVP Core** | — |
| Z6-AC35 | Onboarding parent : création de compte et liaison à l'élève | **MVP Core** | — |
| Z6-AC36 | Comportement le jour de l'examen | **MVP Core** | — |
| Z6-AC37 | Multi-exam overlapping : reset + recompression entre exams | **MVP Core** | — |
| Z6-AC38 | Proposition contrôle blanc : J-3 auto + à la demande | **MVP Core** | — |
| Z6-AC39 | Parent multi-enfants : une notification par enfant | **MVP Hardening** | — |
| Z6-AC40 | Timezone locale : détection automatique et configuration | **MVP Core** | — |
| Z6-AC41 | Zone scolaire et calendrier de vacances intégré | **MVP Core** | — |
| Z6-AC42 | Mode vacances : jours et créneau de révision | **MVP Core** | — |
| Z6-AC43 | Créneaux d'indisponibilité récurrents (sport, activités) | **MVP Core** | — |

---

## Z7 — Routine de soirée & Orchestration (24 ACs)

| AC | Titre | Classification | Lot 0 |
|---|---|---|---|
| Z7-AC01 | Calcul automatique du plan de soirée (EveningPlan) | **MVP Core** | — |
| Z7-AC02 | Dashboard soirée contextuel (écran d'accueil du soir) | **MVP Core** | — |
| Z7-AC03 | Séquencement multi-matières dans le plan | **MVP Core** | — |
| Z7-AC04 | Estimation de durée visible avant le début | **MVP Core** | — |
| Z7-AC05 | État « fini pour ce soir » et écran de clôture | **MVP Core** | — |
| Z7-AC06 | Guidage capture in-app (matières du jour non capturées) | **MVP Core** | — |
| Z7-AC07 | Règles de séquencement des types de session | **MVP Core** | — |
| Z7-AC08 | Mode express pour soirée courte | **MVP Core** | — |
| Z7-AC09 | Complétion partielle et reprise le lendemain | **MVP Hardening** | — |
| Z7-AC10 | Rien à réviser ce soir / week-end | **MVP Hardening** | — |
| Z7-AC11 | Guidage de capture lié au cours de demain | **MVP Hardening** | — |
| Z7-AC12 | Arc émotionnel de la soirée (accueil, transitions, clôture) | **MVP Hardening** | — |
| Z7-AC13 | Notification parent « routine terminée » (signal positif) | **MVP Core** | — |
| Z7-AC14 | Onboarding de la première soirée | **MVP Core** | — |
| Z7-AC15 | Notions : regroupement des items par concept_tag | **MVP Core** | P1 |
| Z7-AC16 | Vue chapitre par notion (accordéon + maîtrise) | **MVP Core** | P2 |
| Z7-AC17 | Périmètre exam : sélection par notion (auto-scope + ajustement) | **MVP Core** | — |
| Z7-AC18 | Auto-suggestion chapitres et notions à la création d'exam | **MVP Core** | — |
| Z7-AC19 | Vue « Angles morts » : contenu sans contrôle à venir | **MVP Core** | — |
| Z7-AC20 | Prédiction interro surprise : score de probabilité par matière | **MVP Hardening** | — |
| Z7-AC21 | Alerte croisée « notion fragile × jamais testée en contrôle » | **MVP Hardening** | — |
| Z7-AC22 | Bouton « S'avancer » : réviser les items des jours suivants | **MVP Core** | — |
| Z7-AC23 | Fiches de révision PDF pour heures d'étude (hors téléphone) | **MVP Core** | — |
| Z7-AC24 | Report des résultats papier (checklist post-étude) | **MVP Core** | — |

---

## Z8 — Onboarding & First Use Experience (8 ACs)

| AC | Titre | Classification | Lot 0 |
|---|---|---|---|
| Z8-AC01 | Chapitre démo pré-chargé (cold start) | MVP Core | P1 |
| Z8-AC02 | Empty state guidé avant premier upload | MVP Core | P1 |
| Z8-AC03 | UX de recovery si le premier OCR échoue | MVP Core | P2 |
| Z8-AC04 | Écran de progression pendant le traitement J0 | MVP Core | P1 |
| Z8-AC05 | Onboarding parent : premiers écrans + digest anticipé | MVP Core | — |
| Z8-AC06 | Digest parent anticipé (J+2 après liaison) | MVP Core | — |
| Z8-AC07 | Invitation parent : timing optimal dans l'onboarding | MVP Core | — |
| Z8-AC08 | Séquence d'onboarding déterministe (orchestration Day 0) | MVP Core | P2 |

---

## Résumé

| Classification | Z1 | Z2 | Z3 | Z4 | Z5 | Z6 | Z7 | Z8 | **Total** |
|---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| **MVP Core** | 20 | 13 | 8 | 6 | 8 | 26 | 16 | 8 | **105** |
| **MVP Hardening** | 8 | 2 | 6 | 8 | 2 | 12 | 9 | 0 | **47** |
| **Post-MVP** | 0 | 0 | 3 | 2 | 0 | 5 | 0 | 0 | **10** |
| **Total** | 28 | 15 | 17 | 16 | 10 | 43 | 25 | 8 | **162** |
| **Lot 0 P1** | 12 | 5 | 1 | 4 | 3 | 4 | 1 | 3 | **33** |
| **Lot 0 P2** | 3 | 5 | 7 | 1 | 0 | 1 | 1 | 2 | **20** |
| **Lot 0 Total** | 15 | 10 | 8 | 5 | 3 | 5 | 2 | 5 | **53** |

