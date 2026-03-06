# Acceptance Criteria — Révise Mieux v1.1

> **Annexe PRD v1.4 · Zones à risque vibe coding**
>
> | | |
> |---|---|
> | **Version** | 1.1 |
> | **Date** | 6 mars 2026 |
> | **Périmètre** | 6 zones critiques identifiées — 65 AC en format Given/When/Then |
> | **Usage** | À intégrer comme contexte système avant chaque session de vibe coding, et à transformer en tests unitaires |

---

## Zones couvertes

| # | Zone | Risque | AC count |
|---|---|---|---|
| Z1 | Transitions Mastery (états + régressions) | Très élevé | 14 |
| Z2 | Pipeline J0 — Error paths & timeouts | Très élevé | 10 |
| Z3 | Validation HITL — Skip / Ignore behavior | Élevé | 9 |
| Z4 | Lazy generation — Concurrence & cache | Élevé | 10 |
| Z5 | ChapterRevision — Identité Item & héritage | Élevé | 8 |
| Z6 | Emploi du temps, Notifications & Révision proactive | Élevé | 23 |
| | **Total** | | **74** |

---

## Z1 — Transitions Mastery

> États · Régressions · Spaced repetition · `next_due_at`
>
> La machine à états Mastery est le cœur de la valeur produit. Une régression silencieuse ou un calcul incorrect de `next_due_at` casse la répétition espacée sans que l'élève s'en aperçoive.

> **Note :** les règles Z1-AC01 à Z1-AC10 s'appliquent uniformément quel que soit le type de session (`diagnostic`, `daily`, `mock_exam`). Le type de session n'affecte pas la logique de transition de maîtrise.

### Z1-AC01 — Progression UNKNOWN → FRAGILE

| | |
|---|---|
| **GIVEN** | Un item en état **UNKNOWN** avec `consecutive_successes = 0`. |
| **WHEN** | L'élève répond correctement une fois à une question liée à cet item. |
| **THEN** | L'état passe à **FRAGILE**. `consecutive_successes = 1`. `next_due_at = now + 1 jour`. `last_review_at` est mis à jour. |

### Z1-AC02 — Progression FRAGILE → OK

| | |
|---|---|
| **GIVEN** | Un item en état **FRAGILE** avec `consecutive_successes = 1`. |
| **WHEN** | L'élève répond correctement une fois à une question liée à cet item (même session). |
| **THEN** | L'état passe à **OK**. `consecutive_successes = 2`. `next_due_at = now + 3 jours` (ajusté par Z1-AC08 si contrôle posé). |

> **NOTE :** La progression FRAGILE→OK n'exige pas un espacement de 24h. L'espacement de 24h est requis uniquement pour OK→SOLID.

> **NOTE :** « Répond correctement » signifie un score ≥ seuil de réussite du pack. Pour les questions NUMERIC avec `unit_required = true`, l'unité fait partie intégrante de la réponse : une valeur juste sans unité est un **échec** (cf. Z1-AC11). Pour les RUBRIC, le seuil est ≥ `seuil_pack` (cf. Z1-AC10). Cette définition de « réponse correcte » s'applique uniformément à tous les AC Z1.

### Z1-AC03 — Progression OK → SOLID (espacement requis)

| | |
|---|---|
| **GIVEN** | Un item en état **OK** avec `consecutive_successes ≥ 2` et `last_review_at` = hier ou avant (≥ 24h écoulées). |
| **WHEN** | L'élève répond correctement à une question liée à cet item. |
| **THEN** | L'état passe à **SOLID**. `consecutive_successes += 1`. `next_due_at = now + 7 jours` (ajusté par Z1-AC08 si contrôle posé). |

### Z1-AC04 — Blocage OK → SOLID sans espacement

| | |
|---|---|
| **GIVEN** | Un item en état **OK** avec `last_review_at < 24h`. |
| **WHEN** | L'élève répond correctement à une question liée à cet item dans la même session. |
| **THEN** | L'état reste **OK**. `consecutive_successes` n'est PAS incrémenté. `next_due_at` n'est PAS modifié. Aucun feedback trompeur n'est affiché. |

### Z1-AC05 — Régression SOLID → OK sur échec unique

| | |
|---|---|
| **GIVEN** | Un item en état **SOLID** avec `consecutive_successes ≥ 3`. |
| **WHEN** | L'élève répond incorrectement une fois à une question liée à cet item. |
| **THEN** | L'état passe à **OK** (pas FRAGILE). `consecutive_successes = 0`. `next_due_at = now + 2 jours` (ajusté par Z1-AC08 si contrôle posé). |

> **NOTE :** La régression SOLID saute FRAGILE. Tomber directement en FRAGILE serait punitif et démotivant pour un élève ayant prouvé une maîtrise solide.

### Z1-AC06 — Régression OK → FRAGILE sur échec

| | |
|---|---|
| **GIVEN** | Un item en état **OK**. |
| **WHEN** | L'élève répond incorrectement à une question liée à cet item. |
| **THEN** | L'état passe à **FRAGILE**. `consecutive_successes = 0`. `next_due_at = now + 1 jour`. |

### Z1-AC07 — Régression FRAGILE sur échec (pas de descente sous FRAGILE)

| | |
|---|---|
| **GIVEN** | Un item en état **FRAGILE**. |
| **WHEN** | L'élève répond incorrectement à une question liée à cet item. |
| **THEN** | L'état reste **FRAGILE**. `consecutive_successes = 0`. `next_due_at = now + 1 jour`. Pas de retour à UNKNOWN. |

### Z1-AC08 — Resserrement proportionnel si contrôle posé

| | |
|---|---|
| **GIVEN** | Un item dont `next_due_at` vient d'être calculé par les règles Z1-AC01 à Z1-AC07. Au moins un `Exam` actif (non expiré) référence le chapitre de cet item via `chapter_ids[]`. Le temps restant `T = min(exam.exam_date) − now` (en jours), calculé sur l'exam le plus proche parmi tous les Exams liés au chapitre. |
| **WHEN** | `next_due_at` est recalculé (après réponse ou lors de la composition de session). |
| **THEN** | L'intervalle standard est remplacé par un intervalle proportionnel au temps restant : |

| État | Intervalle standard | Intervalle si contrôle dans T jours |
|---|---|---|
| UNKNOWN | J+1 | J+1 (incompressible) |
| FRAGILE | J+1 | J+1 (incompressible) |
| OK | J+3 | J + max(1, ⌊T/3⌋) |
| SOLID | J+7 | J + max(2, ⌊T/2⌋) |
| Régression SOLID→OK | J+2 | J + max(1, ⌊T/4⌋) |

**Cap absolu :** `next_due_at ≤ exam_date − 1 jour` (sur l'exam le plus proche). L'item reste visible dans les sessions pré-contrôle même s'il est SOLID.

> **NOTE :** Les contrôles sont typiquement annoncés à +7 jours. Exemples avec T=7 : OK → J+2, SOLID → J+3, régression → J+1. Avec T=3 : OK → J+1, SOLID → J+2, régression → J+1. Sans aucun `Exam` lié, les intervalles standard s'appliquent (cf. Z1-AC01 à Z1-AC07).

> **Multi-exam :** Si un chapitre est lié à plusieurs Exams (ex : interro chapitre 3 le 15/03 + contrôle séquence chapitres 1-3 le 20/03), c'est l'`exam_date` **le plus proche** qui pilote T. Le resserrement est donc maximal — l'élève est préparé pour l'échéance imminente.

> **Edge case T ≤ 0 :** Si tous les `exam_date` liés au chapitre sont passés (`T ≤ 0` pour chaque), le resserrement ne s'applique plus — les intervalles standard reprennent. Les Exams expirés sont ignorés (équivalent à « pas de contrôle posé »). Le système ne doit jamais produire un `next_due_at` dans le passé.

### Z1-AC09 — Indépendance des Mastery states entre items

| | |
|---|---|
| **GIVEN** | Deux items A et B dans le même chapitre, A en SOLID, B en UNKNOWN. |
| **WHEN** | L'élève échoue sur B. |
| **THEN** | Le Mastery state de A n'est pas modifié. Les states sont isolés par `(user_id, item_id)`. |

### Z1-AC10 — Score partiel ne déclenche pas de progression

| | |
|---|---|
| **GIVEN** | Un item en état **FRAGILE**. La question associée est de type RUBRIC (4 critères). |
| **WHEN** | L'élève obtient un score partiel : 2/4 critères corrects. |
| **THEN** | L'état reste **FRAGILE**. `consecutive_successes` n'est pas incrémenté. Le feedback affiche les critères manquants. Seul un score ≥ `seuil_pack` (défaut 3/4) compte comme réussite. |

> **NOTE :** Le seuil de réussite pour les RUBRIC est paramétrable par pack. Défaut MVP : 3/4 critères.

### Z1-AC11 — Réponse NUMERIC sans unité = échec mastery

| | |
|---|---|
| **GIVEN** | Un item en état **FRAGILE** lié à une question NUMERIC avec `unit_required = true`. |
| **WHEN** | L'élève donne la valeur correcte (dans la tolérance ± 2 %) mais omet l'unité. |
| **THEN** | Le score est considéré comme un **échec**. L'état reste **FRAGILE** (pas de progression). `consecutive_successes` est remis à `0`. Le feedback indique explicitement que l'unité est manquante. |

> **NOTE :** Ce AC formalise la règle PRD « Faux négatif si unité absente même si valeur correcte ». Il s'applique à tous les états de mastery, pas seulement FRAGILE — l'exemple FRAGILE est donné car c'est le cas le plus courant. La même logique vaut pour OK et SOLID (régression selon Z1-AC05 / Z1-AC06).

### Z1-AC12 — Score partiel KEYWORDS et progression mastery

| | |
|---|---|
| **GIVEN** | Un item en état **FRAGILE** lié à une question KEYWORDS exigeant N mots-clés. |
| **WHEN** | L'élève fournit N-1 mots-clés corrects sur N (score partiel). |
| **THEN** | Le score N-1 est considéré comme une **réussite**. `consecutive_successes` est incrémenté. La progression mastery s'applique normalement. Le feedback indique le mot-clé manquant à titre informatif. |

> **NOTE :** Le seuil de réussite KEYWORDS est ≥ N-1 (tolérance d'un mot-clé manquant), conformément au PRD §17.3 « Score partiel si N-1 ». En dessous de N-1 (ex. N-2 ou moins), c'est un échec. Cette tolérance compense les variations de formulation naturelles en français. Ce AC complète Z1-AC10 (RUBRIC) et Z1-AC11 (NUMERIC) pour couvrir tous les types de scoring.

### Z1-AC13 — Échec sur item UNKNOWN (pas de descente sous UNKNOWN)

| | |
|---|---|
| **GIVEN** | Un item en état **UNKNOWN** avec `consecutive_successes = 0`. |
| **WHEN** | L'élève répond incorrectement à une question liée à cet item. |
| **THEN** | L'état reste **UNKNOWN**. `consecutive_successes = 0`. `next_due_at = now + 1 jour`. Aucune régression n'est possible en dessous de UNKNOWN. |

### Z1-AC14 — Maintien SOLID sur réussite successive

| | |
|---|---|
| **GIVEN** | Un item en état **SOLID** avec `consecutive_successes ≥ 3` et `last_review_at` ≥ 24h. |
| **WHEN** | L'élève répond correctement à une question liée à cet item. |
| **THEN** | L'état reste **SOLID**. `consecutive_successes += 1`. `next_due_at = now + 7 jours` (ajusté par Z1-AC08 si contrôle posé). `last_review_at` est mis à jour. |

> **NOTE :** Un item SOLID qui continue d'être réussi reste SOLID avec un intervalle constant de J+7. L'incrémentation de `consecutive_successes` au-delà de 3 permet de distinguer un item « fraîchement SOLID » d'un item « profondément ancré » pour d'éventuelles heuristiques post-MVP.

---

## Z2 — Pipeline J0 — Error paths & timeouts

> OCR failures · Timeouts · Partial processing · État cohérent en cas d'erreur
>
> Le pipeline J0 est le point d'entrée critique. Une erreur non gérée ici peut laisser le chapitre dans un état partiellement initialisé, invisiblement corrompu pour les sessions suivantes.

### Z2-AC01 — Carte de leçon partielle si OCR en cours

| | |
|---|---|
| **GIVEN** | L'élève uploade 10 pages. Le traitement des pages 1–3 est terminé. Les pages 4–10 sont en cours de traitement (OCR worker en cours). |
| **WHEN** | L'élève ouvre la carte de leçon. |
| **THEN** | La carte affiche les items des pages 1–3. Un indicateur de progression (3/10 pages analysées) est visible. Les pages non traitées apparaissent comme placeholders. L'élève peut commencer un diagnostic sur les items disponibles. |

> **NOTE :** La carte partielle est fonctionnelle. L'absence de pages ne bloque pas l'accès.

### Z2-AC02 — Timeout OCR sur une page intermédiaire

| | |
|---|---|
| **GIVEN** | Le pipeline traite 10 pages. La page 4 dépasse le timeout (10 s, après 3 retries). |
| **WHEN** | Le worker abandonne la page 4. |
| **THEN** | La page 4 est marquée `status = FAILED` avec `reason = 'ocr_timeout'`. Les pages 5–10 continuent leur traitement normalement. L'élève voit un badge 'Zone illisible · page 4' sur la carte. Le chapitre n'est PAS bloqué. Aucune exception non gérée n'est propagée. |

### Z2-AC03 — Photo floue détectée (confidence globale < 0.3)

| | |
|---|---|
| **GIVEN** | Une page uploadée a une confidence OCR globale < 0.3 sur tous ses blocs. |
| **WHEN** | Le worker finalise l'analyse de cette page. |
| **THEN** | Un message non-bloquant est affiché : 'Photo floue — vous pouvez reprendre cette photo pour de meilleurs résultats'. La page reste dans le chapitre avec ses blocs marqués UNCERTAIN. L'élève peut continuer sans reprendre la photo. |

### Z2-AC04 — Aucun item généré sur une page (page vide ou illisible)

| | |
|---|---|
| **GIVEN** | Une page OCRisée ne produit aucun Item après la phase de génération (texte vide ou uniquement des artefacts). |
| **WHEN** | La génération d'items se termine pour cette page. |
| **THEN** | Aucun Item n'est créé pour cette page. La page est marquée `status = 'no_items_extracted'`. La carte de leçon ne montre pas de section vide pour cette page. Le chapitre est valide si au moins une autre page a produit des items. |

### Z2-AC05 — Échec génération Items (erreur LLM)

| | |
|---|---|
| **GIVEN** | L'OCR d'une page a réussi. L'appel LLM pour la génération d'items retourne une erreur (timeout, quota, ou réponse malformée). |
| **WHEN** | Le worker LLM épuise ses retries (x2 avec back-off). |
| **THEN** | Les blocs de cette page sont conservés avec leur texte OCR brut. Aucun Item n'est créé pour cette page. La page est marquée `status = 'items_generation_failed'`. Une ValidationTask admin est créée pour re-traitement manuel. Le reste du pipeline n'est pas affecté. |

### Z2-AC06 — Re-tentative pipeline après interruption (idempotence)

| | |
|---|---|
| **GIVEN** | Le pipeline J0 a été interrompu après les étapes 1–4 (upload, segmentation, OCR, plan). L'élève ou le système relance le traitement du chapitre. |
| **WHEN** | Le pipeline redémarre. |
| **THEN** | Les étapes déjà complétées ne sont pas ré-exécutées (les résultats OCR en cache permanent sont réutilisés). Le pipeline reprend à partir de l'étape non complétée. Aucun Item en double n'est créé. |

### Z2-AC07 — Diagnostic impossible si 0 items valides

| | |
|---|---|
| **GIVEN** | Le pipeline J0 se termine avec 0 items produits (toutes pages échouées ou illisibles). |
| **WHEN** | L'élève tente de lancer le diagnostic initial. |
| **THEN** | Le bouton diagnostic est désactivé. Un message explicite est affiché : 'Aucun contenu n'a pu être extrait — vérifiez la qualité des photos et ré-uploadez.' Le chapitre reste en état `PENDING_UPLOAD`. |

### Z2-AC08 — Bloc SCHEMA/MAP non OCRisé conservé comme Document image

| | |
|---|---|
| **GIVEN** | Un bloc de type MAP ou SCHEMA a une confidence OCR < 0.5. |
| **WHEN** | L'étape OCR se termine pour ce bloc. |
| **THEN** | Le bloc n'est PAS OCRisé. Un Document de type correspondant est créé avec `source_image_url = crop_url` et `tags = ['map']` ou `['schema']`. Ce Document est éligible aux gabarits `GEN.DOC.MAP.READ_ZONES` ou `GEN.DOC.IMAGE.DESCRIBE_INTERPRET`. Aucun texte OCR n'est stocké. |

### Z2-AC09 — File de validation non dépassée (max 8)

| | |
|---|---|
| **GIVEN** | Le pipeline J0 détecte 15 items avec `validation_required = true` sur un chapitre. |
| **WHEN** | La file de validation est constituée. |
| **THEN** | Seuls les 8 items avec la priorité la plus haute (score d'impact sur la note) sont inclus dans la ValidationTask. Les 7 autres items restent avec `validation_required = true` mais ne sont pas soumis à l'élève. Ils peuvent être traités par l'admin en backoffice. |

### Z2-AC10 — Streaming carte de leçon (première page disponible)

| | |
|---|---|
| **GIVEN** | Le chapitre a 8 pages en cours de traitement. La page 1 est la première traitée. |
| **WHEN** | Le traitement de la page 1 se termine (OCR + items + tags). |
| **THEN** | La carte de leçon est mise à jour en temps réel via SSE avec les items de la page 1. L'élève peut interagir avec ces items avant que les pages 2–8 soient traitées. Les pages suivantes s'ajoutent progressivement sans recharger la vue. |

---

## Z3 — Validation HITL — Skip / Ignore behavior

> États chapter · Gabarits bloqués · Comportement parent actif
>
> La validation HITL est optionnelle pour l'élève mais critique pour la qualité. Les comportements de skip/ignore doivent être explicites : un item non validé ne doit jamais générer un exercice trompeur.

### Z3-AC01 — Gabarits bloqués sur item validation_required non résolu

| | |
|---|---|
| **GIVEN** | Un item avec `validation_required = true`, `status = PENDING`, de type KNOWLEDGE avec tags `['calcul', 'unites']`. |
| **WHEN** | Le moteur de lazy generation compose une session. |
| **THEN** | Les gabarits `PC.FORMULA.APPLY`, `PC.UNITS.CONVERT`, `PC.FORMULA.ISOLATE` et tout gabarit avec `question_type = NUMERIC` sont exclus de l'éligibilité pour cet item. Seuls `GEN.KNOW.DEF_SHORT` et `GEN.KNOW.FLASH_MCQ` restent éligibles. |

> **NOTE :** Un item non validé peut apparaître en session avec des gabarits simples (rappel/reconnaissance) mais jamais avec des gabarits exigeant précision numérique.

### Z3-AC02 — Action 'Confirmer' sur ValidationTask

| | |
|---|---|
| **GIVEN** | Une ValidationTask sur un item avec `suggestion = 'ρ = m / V'`, `status PENDING`. |
| **WHEN** | L'élève (ou le parent actif) clique sur 'Confirmer'. |
| **THEN** | L'item est mis à jour : `validation_required = false`, `confidence = max(item.confidence, 0.85)`. La ValidationTask passe en `status = RESOLVED_CONFIRMED`. Les Questions liées à cet item dans le cache sont invalidées → régénération lazy au prochain accès. |

### Z3-AC03 — Action 'Corriger' sur ValidationTask

| | |
|---|---|
| **GIVEN** | Une ValidationTask sur un item. L'élève saisit une correction. |
| **WHEN** | L'élève soumet la correction. |
| **THEN** | L'item est mis à jour avec le contenu corrigé. `validation_required = false`. `confidence = 1.0` (correction humaine explicite). La ValidationTask passe en `RESOLVED_CORRECTED`. Le cache questions de cet item est invalidé. |

### Z3-AC04 — Action 'Je ne sais pas' sur ValidationTask

| | |
|---|---|
| **GIVEN** | Une ValidationTask sur un item. L'élève choisit 'Je ne sais pas'. |
| **WHEN** | L'action est soumise. |
| **THEN** | `validation_required` reste `true`. `status = DEFERRED_BY_STUDENT`. L'item est retiré de la file de validation élève. Une ValidationTask admin est créée avec priorité haute. Les gabarits bloqués restent bloqués (cf. Z3-AC01). |

### Z3-AC05 — Action 'Ignorer' sur ValidationTask

| | |
|---|---|
| **GIVEN** | Une ValidationTask sur un item non critique (impact faible). L'élève choisit 'Ignorer'. |
| **WHEN** | L'action est soumise. |
| **THEN** | `validation_required` reste `true`. `status = IGNORED_BY_STUDENT`. L'item est exclu des sessions et des contrôles blancs tant que non résolu. Il n'apparaît PAS dans la carte de maîtrise de l'élève comme UNKNOWN (il est masqué pour éviter la confusion). |

> **NOTE :** 'Ignorer' ≠ 'Je ne sais pas'. Ignorer exclut l'item du flux pédagogique. 'Je ne sais pas' le délègue à l'admin pour correction.

### Z3-AC06 — Chapitre utilisable si 0 validations effectuées

| | |
|---|---|
| **GIVEN** | Un chapitre avec 6 ValidationTasks PENDING. L'élève n'en traite aucune. |
| **WHEN** | L'élève lance le diagnostic initial. |
| **THEN** | Le diagnostic est accessible. Les items avec `validation_required = true` ne génèrent que des gabarits de rappel simple (Z3-AC01). Un bandeau informatif indique '6 zones à vérifier — vos exercices seront plus précis après vérification' sans bloquer la progression. |

### Z3-AC07 — Résolution validation → régénération ciblée uniquement

| | |
|---|---|
| **GIVEN** | Un chapitre avec 20 questions en cache. Une ValidationTask sur l'item X est résolue. |
| **WHEN** | La résolution est enregistrée. |
| **THEN** | Seules les questions dont `item_id = X` sont invalidées dans le cache. Les `20 − N` questions liées aux autres items ne sont pas invalidées. La régénération est lazy : les nouvelles questions pour X sont créées au prochain appel de composition de session. |

### Z3-AC08 — Parent actif : max 3 ValidationTasks par semaine

| | |
|---|---|
| **GIVEN** | Un parent en mode actif. Le chapitre a 8 ValidationTasks PENDING. |
| **WHEN** | Le système prépare les validations à soumettre au parent. |
| **THEN** | Seules les 3 ValidationTasks avec le score d'impact le plus élevé sont envoyées au parent cette semaine. Les 5 autres sont gardées en file pour la semaine suivante. Le parent ne reçoit jamais plus de 3 demandes de validation par semaine. |

### Z3-AC09 — Aucune ValidationTask si confiance > seuil sur tous les blocs

| | |
|---|---|
| **GIVEN** | Un chapitre dont tous les blocs ont `confidence ≥ 0.85` et aucun terme du lexique pack critique n'est détecté comme ambigu. |
| **WHEN** | Le pipeline J0 finalise la détection d'incertitudes. |
| **THEN** | Aucune ValidationTask n'est créée. La file de validation est vide. L'élève passe directement au diagnostic sans étape de validation. Tous les items ont `validation_required = false`. |

---

## Z4 — Lazy generation — Concurrence & cache

> Race conditions · Invalidation · Pool épuisé · Sessions parallèles
>
> La génération lazy avec cache partagé est le principal vecteur de bugs silencieux en contexte concurrent (plusieurs sessions ouvertes, re-upload, résolution HITL).

### Z4-AC01 — Pas de doublons si deux sessions composées simultanément

| | |
|---|---|
| **GIVEN** | Deux sessions pour le même `(user_id, chapter_id)` sont composées en parallèle (ex : onglets multiples ou race condition réseau). |
| **WHEN** | Les deux requêtes de composition de session arrivent dans un intervalle < 500 ms. |
| **THEN** | Une seule session est créée (idempotence via lock ou check `session_id` existante). La deuxième requête retourne la session existante si elle est < 30 min. Aucune question en double n'est générée pour la même session. |

### Z4-AC02 — Pool question candidates non partagé entre sessions actives

| | |
|---|---|
| **GIVEN** | L'élève a une session active S1 avec 10 questions issues du pool. Il ouvre une nouvelle session S2 (contrôle blanc) sur le même chapitre. |
| **WHEN** | S2 est composée. |
| **THEN** | S2 pioche dans le pool de candidats mais les questions déjà assignées à S1 (en `status = IN_PROGRESS`) sont exclues de S2 si le pool le permet. Si le pool est insuffisant (< 5 candidats uniques disponibles), S2 peut réutiliser des questions de S1 — ceci doit être loggué. |

> **NOTE :** En MVP, l'exhaustion de pool est acceptable si loggée. Post-MVP : régénération auto.

### Z4-AC03 — Invalidation cache après résolution HITL (atomicité)

| | |
|---|---|
| **GIVEN** | Le cache `question_candidates:{chapter_id}` contient 25 entrées. Une ValidationTask sur l'item X est résolue. |
| **WHEN** | La résolution est persistée en base. |
| **THEN** | L'invalidation du cache pour les questions liées à X est effectuée dans la même transaction (ou dans un job immédiat < 1 s). Aucune fenêtre de temps ne doit exister où la base dit 'validé' mais le cache sert encore l'ancienne version. |

### Z4-AC04 — Renouvellement TTL sur lecture (read-through)

| | |
|---|---|
| **GIVEN** | Le cache `item_pool:{chapter_id}` a un TTL de 24h. Il a été créé il y a 23h. |
| **WHEN** | Une session est composée et lit le pool. |
| **THEN** | Le TTL est **prolongé** de 24h à partir de la lecture (sliding window). Le pool n'expire pas pendant qu'il est activement utilisé. |

> **NOTE :** Le TTL sliding évite l'expiration pendant une session active longue.

### Z4-AC05 — Invalidation complète sur mise à jour de pack_version

| | |
|---|---|
| **GIVEN** | Le pack PC-MVD est mis à jour (ajout d'un concept tag). Trois chapitres actifs utilisent ce pack. |
| **WHEN** | L'admin sauvegarde la nouvelle version du pack. |
| **THEN** | Les caches `item_pool` et `question_candidates` des 3 chapitres sont invalidés. Les Mastery states ne sont **pas** invalidés. Les sessions en cours (`status = IN_PROGRESS`) continuent avec les questions déjà générées — pas d'interruption. |

### Z4-AC06 — Session interrompue reprise depuis le dernier état connu

| | |
|---|---|
| **GIVEN** | Une session S1 avec 10 questions. L'élève répond à 4 questions puis ferme l'application. |
| **WHEN** | L'élève rouvre la session (TTL session = 72h). |
| **THEN** | La session reprend à la question 5. Les Mastery states des items 1–4 sont déjà mis à jour. L'état de session (`current_question_index = 4`) est persisté en base et non uniquement en cache. |

### Z4-AC07 — Contraintes pack respectées lors de la composition lazy

| | |
|---|---|
| **GIVEN** | Chapitre HG-INEG avec 5 items document disponibles et 8 items knowledge. Pack : `max_writing_per_session = 1`, `session_must_include_doc = true`. |
| **WHEN** | Une session quotidienne est composée. |
| **THEN** | La session inclut exactement 1 exercice de type document (parmi les 5 disponibles). La session inclut au maximum 1 exercice de rédaction (`GEN.WRITE.*`). Ces contraintes sont respectées même si l'algorithme 70/20/10 sélectionnerait autrement. |

### Z4-AC08 — Pool vide : comportement dégradé gracieux

| | |
|---|---|
| **GIVEN** | Le cache du pool est expiré et le LLM est indisponible (timeout). L'élève tente de démarrer une session. |
| **WHEN** | La composition de session échoue à générer des questions. |
| **THEN** | L'interface affiche : 'Préparation de tes exercices en cours… réessaie dans quelques secondes.' Aucune session vide n'est créée en base. La page n'affiche pas d'erreur 500. |

### Z4-AC09 — OCR cache permanent (hash photo)

| | |
|---|---|
| **GIVEN** | L'élève re-uploade exactement la même photo (même contenu binaire) lors d'une révision de chapitre. |
| **WHEN** | Le pipeline J0 traite cette photo. |
| **THEN** | L'OCR n'est PAS ré-exécuté. Le résultat en cache (clé = hash SHA-256 de l'image) est réutilisé directement. Le pipeline passe à l'étape de génération d'items. |

> **NOTE :** Cache permanent sur le hash évite la re-facturation OCR sur photos identiques.

### Z4-AC10 — Contrôle blanc non affecté par une session quotidienne en cours

| | |
|---|---|
| **GIVEN** | L'élève a une session quotidienne S1 en cours (non terminée). Il lance un contrôle blanc CB1. |
| **WHEN** | CB1 est composé. |
| **THEN** | CB1 est une session indépendante avec son propre pool de questions. CB1 n'est pas bloqué par S1. Les deux sessions coexistent. Les Mastery updates de CB1 sont appliqués indépendamment de S1. |

---

## Z5 — ChapterRevision — Identité Item & héritage Mastery

> Re-upload · Clé d'identité · Héritage state · Conflits OCR
>
> Le versioning de chapitre protège l'historique de maîtrise lors d'un re-upload. La clé d'identité d'un Item détermine si on hérite ou recrée un Mastery state — une erreur ici remet à zéro silencieusement tout le travail de l'élève.

### Z5-AC01 — Clé d'identité d'un Item (définition canonique)

| | |
|---|---|
| **GIVEN** | Deux items provenant de révisions différentes du même chapitre. |
| **WHEN** | Le système évalue si ces deux items sont le même item logique. |
| **THEN** | Deux items sont considérés identiques si et seulement si : `normalized(item.term) == normalized(item2.term)` ET `item.pack_id == item2.pack_id` ET `item.type == item2.type`. La normalisation inclut : lowercase, trim, suppression des accents. Aucun autre champ (`confidence`, `tags`, `revision_id`) n'entre dans la clé. |

> **NOTE :** La clé volontairement simple évite les faux négatifs dus à l'OCR. Un terme légèrement différent (ex. 'IDH' vs 'I.D.H.') doit être géré par la normalisation, pas en créant un doublon.

### Z5-AC02 — Héritage Mastery sur re-upload (item identique retrouvé)

| | |
|---|---|
| **GIVEN** | L'item `IDH` (pack HG-INEG, type KNOWLEDGE) est en état SOLID dans la révision R1. L'élève uploade de nouvelles pages (révision R2). L'item `IDH` est re-détecté dans R2 avec une clé identique. |
| **WHEN** | La révision R2 est finalisée. |
| **THEN** | L'item `IDH` dans R2 hérite du Mastery state SOLID de R1. `next_due_at`, `consecutive_successes` et `last_review_at` sont copiés. L'item R1 est archivé (`archived = true`) mais non supprimé. |

### Z5-AC03 — Pas d'héritage si clé différente (item nouveau)

| | |
|---|---|
| **GIVEN** | L'item `PIB/habitant` est en état OK dans R1. Dans la révision R2, l'OCR produit `PIB par habitant` (terme légèrement différent, normalisation identique → même clé normalisée). |
| **WHEN** | La révision R2 est finalisée. |
| **THEN** | Les deux formes sont reconnues comme le même item (normalisation). L'héritage Mastery s'applique (cf. Z5-AC02). Aucun doublon dans la carte de maîtrise. |

### Z5-AC04 — Conflit OCR : terme ambigu entre révisions

| | |
|---|---|
| **GIVEN** | L'item `vassal` est en état FRAGILE dans R1. Dans R2, l'OCR produit `vassalle` (erreur d'OCR). Après normalisation : `vassal ≠ vassalle` → clés différentes. |
| **WHEN** | La révision R2 est finalisée. |
| **THEN** | L'item `vassalle` est créé comme NOUVEL item avec state UNKNOWN. L'item `vassal` de R1 est archivé. Une ValidationTask est créée automatiquement pour `vassalle` (clé non trouvée dans lexique pack → suspect). |

> **NOTE :** Ce cas illustre pourquoi la ValidationTask pour les termes hors-lexique est critique. Le lexique pack sert de filet de sécurité contre les erreurs OCR sur les concepts clés.

### Z5-AC05 — Révision courante unique par chapitre

| | |
|---|---|
| **GIVEN** | Un chapitre avec deux révisions R1 et R2 créées. |
| **WHEN** | L'élève ouvre la carte de leçon du chapitre. |
| **THEN** | Seuls les items de la révision `current_revision_id` (R2) sont affichés. Les items archivés de R1 ne sont pas visibles dans la carte. Les Mastery states hérités sont bien ceux affichés dans le dashboard. |

### Z5-AC06 — Sessions en cours non affectées par une nouvelle révision

| | |
|---|---|
| **GIVEN** | L'élève a une session S1 active (IN_PROGRESS) avec des questions issues de R1. Il uploade de nouvelles pages → création de R2. |
| **WHEN** | R2 est finalisée et `current_revision_id` est mis à jour. |
| **THEN** | La session S1 continue avec les questions de R1 jusqu'à sa fin. Les Mastery updates de S1 sont appliqués normalement. La prochaine session sera composée avec les items de R2. |

### Z5-AC07 — Items de R1 archivés non proposés en session

| | |
|---|---|
| **GIVEN** | Un item I1 archivé de R1 a un Mastery state FRAGILE. La révision courante est R2 (I1 non re-détecté dans R2). |
| **WHEN** | Une session est composée. |
| **THEN** | L'item I1 archivé n'est PAS inclus dans la sélection de session. `archived = true` l'exclut de toutes les requêtes de composition. Le dashboard affiche uniquement les items de la révision courante. |

### Z5-AC08 — Historique Mastery préservé entre révisions (audit trail)

| | |
|---|---|
| **GIVEN** | L'item `IDH` a eu 5 tentatives dans R1 (3 réussites, 2 échecs). Dans R2, l'item est hérité avec state SOLID. |
| **WHEN** | L'élève consulte l'historique de maîtrise de l'item `IDH`. |
| **THEN** | L'historique affiche les 5 tentatives de R1 + les nouvelles tentatives de R2. L'interface indique que les tentatives anciennes proviennent d'une révision archivée. Aucune tentative n'est supprimée lors de l'archivage. |

---

## Z6 — Emploi du temps, Notifications & Révision proactive

> ScheduleSlot · Notifications · Session evening_first · Session pre_class · Exam multi-chapitres · Mode dégradé
>
> L'emploi du temps est le socle de toute la couche proactive. Une notification mal ciblée fatigue l'élève. Une session evening_first ou pre_class mal composée dilue la valeur du rappel. Un exam multi-chapitres mal borné explose le temps de session.

### Z6-AC01 — CRUD ScheduleSlot

| | |
|---|---|
| **GIVEN** | L'élève est en onboarding ou dans ses paramètres. |
| **WHEN** | Il saisit un créneau : matière = 'Physique-Chimie', jour = mardi, période = matin. |
| **THEN** | Un `ScheduleSlot` est créé avec `user_id`, `subject`, `day_of_week = 2`, `period = 'morning'`. La modification et la suppression sont possibles à tout moment. Un doublon exact `(user_id, subject, day_of_week, period)` est rejeté (contrainte d'unicité). |

### Z6-AC02 — Notification capture_reminder déclenchée par emploi du temps

| | |
|---|---|
| **GIVEN** | L'élève a un `ScheduleSlot` mardi matin pour Physique-Chimie. Il est mardi 18h30 (heure de notification par défaut). Aucun chapitre Physique-Chimie n'a été saisi aujourd'hui. |
| **WHEN** | Le scheduler de notifications s'exécute. |
| **THEN** | Une notification `capture_reminder` est envoyée : 'Tu as eu Physique-Chimie aujourd'hui — saisis ton cours pour réviser ce soir !' L'heure d'envoi est `user.notification_hour` (défaut 18h30). La notification est loggée avec `scheduled_at`, `sent_at`. |

### Z6-AC03 — Notification review_reminder si chapitre déjà saisi

| | |
|---|---|
| **GIVEN** | L'élève a un `ScheduleSlot` mardi matin pour Physique-Chimie. Il est mardi 18h30. Un chapitre Physique-Chimie a été saisi aujourd'hui (ou un chapitre existant a des items FRAGILE/OK dues). |
| **WHEN** | Le scheduler de notifications s'exécute. |
| **THEN** | Une notification `review_reminder` est envoyée : 'Révise tes points fragiles en Physique-Chimie — 10 min ce soir'. La notification `capture_reminder` n'est PAS envoyée (le chapitre est déjà saisi). |

### Z6-AC04 — Max 2 notifications par soir

| | |
|---|---|
| **GIVEN** | L'élève a 4 matières le mardi : Physique-Chimie, Maths, SVT, Français. 3 cours n'ont pas été saisis. 1 chapitre a des items FRAGILE. |
| **WHEN** | Le scheduler prépare les notifications du mardi soir. |
| **THEN** | Seules **2 notifications** sont envoyées. La priorité est : (1) matière avec Exam le plus proche, (2) matière avec le plus d'items FRAGILE/UNKNOWN. Les 2 notifications restantes sont supprimées (pas reportées). L'élève ne reçoit jamais plus de 2 notifications par soir. |

### Z6-AC05 — Session evening_first déclenchée après upload

| | |
|---|---|
| **GIVEN** | L'élève uploade un nouveau chapitre de Physique-Chimie. Le pipeline J0 produit 12 items. |
| **WHEN** | Le pipeline J0 se termine avec ≥ 1 item valide. |
| **THEN** | Une session de type `evening_first` est automatiquement proposée (pas lancée de force). `trigger = 'scheduled'`. Durée cible : 5–10 min. La session est proposée **immédiatement**, quelle que soit l'heure. Si l'élève ne la fait pas, elle reste disponible 72h (TTL session standard). |

### Z6-AC06 — evening_first : contenu 100% UNKNOWN, gabarits difficulté 1

| | |
|---|---|
| **GIVEN** | Une session `evening_first` est composée pour un chapitre fraîchement uploadé avec 12 items UNKNOWN. |
| **WHEN** | Le moteur de composition sélectionne les questions. |
| **THEN** | 100% des items sont issus du chapitre uploadé. Tous sont en état UNKNOWN. Seuls les gabarits de difficulté 1 sont éligibles : `GEN.KNOW.FLASH_MCQ`, `GEN.KNOW.DEF_SHORT`, `GEN.KNOW.CLOZE_KEYWORDS`. Aucun gabarit NUMERIC, RUBRIC, ou de rédaction n'est inclus. Le nombre de questions est calibré pour 5–10 min (typiquement 6–10 questions). |

### Z6-AC07 — Session pre_class la veille de chaque cours

| | |
|---|---|
| **GIVEN** | L'élève a un `ScheduleSlot` mercredi matin pour Histoire-Géo. Il est mardi soir. L'élève a 3 chapitres actifs en Histoire-Géo avec des items aux états variés (UNKNOWN, FRAGILE, OK, SOLID). |
| **WHEN** | Le scheduler de sessions évalue les sessions à proposer pour mardi soir. |
| **THEN** | Une session `pre_class` est proposée. Durée cible : 5 min. La justification affichée est : 'Tu as Histoire-Géo demain — prépare-toi en cas d'interro surprise'. `trigger = 'scheduled'`. |

### Z6-AC08 — Fusion pre_class dans daily si session daily prévue le même soir

| | |
|---|---|
| **GIVEN** | L'élève a mardi soir : une session `daily` prévue (items dues de Maths + Physique) ET une session `pre_class` pour Histoire-Géo (cours mercredi matin). |
| **WHEN** | Le moteur de composition prépare les sessions du mardi soir. |
| **THEN** | Les items `pre_class` d'Histoire-Géo sont **injectés en priorité** dans la session `daily`. L'élève ne voit qu'une seule session. Les items pre_class apparaissent dans les premières questions. Le type de la session reste `daily`. L'attribut `includes_pre_class = true` est positionné pour le tracking. |

### Z6-AC09 — pre_class : scope multi-chapitres de la matière

| | |
|---|---|
| **GIVEN** | L'élève a 3 chapitres actifs en Histoire-Géo : 'Inégalités' (8 items, 3 FRAGILE), 'Mondialisation' (6 items, 1 UNKNOWN), 'Urbanisation' (10 items, 2 OK dues). |
| **WHEN** | La session `pre_class` est composée. |
| **THEN** | La sélection puise dans **tous les chapitres actifs** de la matière. Priorité : (1) items FRAGILE/OK dont `next_due_at ≤ now`, (2) items UNKNOWN jamais vus. Gabarits : difficulté 1–2 uniquement (rappel rapide, pas de problèmes longs). Durée cible : 5 min (typiquement 4–6 questions). |

### Z6-AC10 — Exam multi-chapitres : mock_exam couvre tous les chapitres liés

| | |
|---|---|
| **GIVEN** | Un `Exam` 'Contrôle séquence 1' avec `chapter_ids = [ch1, ch2, ch3]`. ch1 a 10 items, ch2 a 15 items, ch3 a 8 items. |
| **WHEN** | L'élève lance un contrôle blanc (`mock_exam`) pour cet Exam. |
| **THEN** | La session `mock_exam` inclut des items des **3 chapitres**. La sélection est proportionnelle au nombre d'items par chapitre (≈ 30% ch1, 45% ch2, 25% ch3). Tous les niveaux de difficulté sont éligibles. La durée est cappée à **30 min maximum**. Si le pool total dépasse 30 min, un échantillon représentatif est sélectionné. |

### Z6-AC11 — Exam multi-chapitres : création et liaison

| | |
|---|---|
| **GIVEN** | L'élève crée un Exam 'Interro chapitre 3' avec `exam_date = 2026-03-15` et sélectionne le chapitre 'Inégalités'. |
| **WHEN** | L'Exam est sauvegardé. |
| **THEN** | L'entité `Exam` est créée avec `chapter_ids = ['ch_inegalites']`. Le chapitre 'Inégalités' référence cet Exam. Le resserrement Z1-AC08 s'active pour tous les items du chapitre lié. Un Exam peut être modifié (ajout/retrait de chapitres, changement de date) à tout moment. |

### Z6-AC12 — Mode dégradé sans emploi du temps

| | |
|---|---|
| **GIVEN** | L'élève n'a saisi aucun `ScheduleSlot`. |
| **WHEN** | Le système évalue les sessions et notifications à planifier. |
| **THEN** | Aucune notification `capture_reminder` ou `pre_class` n'est envoyée. Aucune session `pre_class` n'est planifiée. Les sessions `daily`, `diagnostic`, `mock_exam`, et `evening_first` fonctionnent normalement. La révision espacée standard s'applique sans modification. Un nudge 'Saisis ton emploi du temps pour des révisions plus ciblées' est affiché à J+3 puis au début de chaque trimestre. |

### Z6-AC13 — Notifications désactivables sans impact sessions

| | |
|---|---|
| **GIVEN** | L'élève a un emploi du temps saisi mais désactive les notifications dans ses paramètres. |
| **WHEN** | Le scheduler de notifications s'exécute. |
| **THEN** | Aucune notification n'est envoyée. Les sessions `pre_class` et `evening_first` restent **disponibles** (composées normalement) — l'élève peut les lancer manuellement. Seul le push notification est supprimé, pas la logique de composition. |

### Z6-AC14 — Pas de session pre_class si aucun chapitre actif dans la matière

| | |
|---|---|
| **GIVEN** | L'élève a un `ScheduleSlot` mercredi matin pour SVT. Aucun chapitre SVT n'a été créé (ou tous sont archivés). |
| **WHEN** | Le scheduler évalue les sessions pre_class pour mardi soir. |
| **THEN** | Aucune session `pre_class` n'est créée pour SVT. Aucune notification `pre_class` n'est envoyée. Le système n'affiche pas d'erreur. |

### Z6-AC15 — Items dues non révisés : aucune pénalité mastery

| | |
|---|---|
| **GIVEN** | Un item en état **OK** avec `next_due_at = 2 mars`. L'élève ne révise pas du 2 au 5 mars (3 jours de retard). |
| **WHEN** | L'élève ouvre une session le 5 mars. |
| **THEN** | L'item est toujours en état **OK**. `consecutive_successes` n'a pas changé. `next_due_at` est resté au 2 mars (non modifié par l'inaction). L'item apparaît en priorité dans les 70% « items dus » de la session (car `next_due_at < now`). Aucune régression n'a eu lieu. |

> **NOTE :** Le Mastery state n'est JAMAIS modifié par l'inaction. Seule une réponse incorrecte déclenche une régression (Z1-AC05 à Z1-AC07). Un élève qui revient après une pause retrouve ses acquis intacts et reprend là où il en était. Cette règle s'applique à tous les états (FRAGILE, OK, SOLID).

### Z6-AC16 — Rappel unique le lendemain pour session evening_first manquée

| | |
|---|---|
| **GIVEN** | Une session `evening_first` a été proposée lundi soir à 18h30. L'élève n'a répondu à aucune question. Il est mardi 08h00 (heure de rappel par défaut). |
| **WHEN** | Le scheduler de notifications s'exécute mardi matin. |
| **THEN** | Une notification `missed_session_reminder` est envoyée : « Tu avais une révision en attente — on s'y remet ? ». `source_session_id` référence la session manquée. **Aucun second rappel** n'est envoyé si l'élève ignore ce rappel. La session originale reste disponible jusqu'à expiration (TTL 72h). |

> **NOTE :** Seules les sessions `evening_first` et `pre_class` non commencées (0 questions répondues) déclenchent un rappel. Les sessions `daily` non commencées ne génèrent PAS de rappel pour éviter la sur-sollicitation. Les sessions partiellement complétées (≥ 1 question répondue) ne déclenchent pas non plus de rappel.

### Z6-AC17 — Rappel unique le lendemain pour session pre_class manquée

| | |
|---|---|
| **GIVEN** | Une session `pre_class` a été proposée mardi soir. L'élève n'a répondu à aucune question. Il est mercredi 08h00. |
| **WHEN** | Le scheduler de notifications s'exécute mercredi matin. |
| **THEN** | Une notification `missed_session_reminder` est envoyée : « Tu avais une révision en attente — on s'y remet ? ». Le rappel est envoyé même si le cours a déjà eu lieu (mercredi matin). **Un seul rappel**, jamais de relance. |

### Z6-AC18 — Aucune mécanique de streak

| | |
|---|---|
| **GIVEN** | L'élève a révisé 5 jours consécutifs puis ne révise pas pendant 3 jours. |
| **WHEN** | L'élève revient le 9e jour et ouvre l'application. |
| **THEN** | Aucun compteur de « série » ou de « streak » n'est affiché. Aucun message de type « Tu as perdu ta série » n'apparaît. L'interface affiche l'état actuel de la carte de maîtrise sans référence à la régularité passée. Le message d'accueil est neutre ou positif : « Tes révisions t'attendent — on continue ? ». |

> **NOTE :** L'absence de streak est un choix produit délibéré. La gamification par streak culpabilise les élèves en cas de rupture et peut être contre-productive pour les collégiens (11–15 ans). Le service valorise la qualité de la révision, pas la quantité de jours consécutifs.

### Z6-AC19 — Annulation ponctuelle d'un cours

| | |
|---|---|
| **GIVEN** | L'élève a un `ScheduleSlot` mardi matin pour Physique-Chimie. Il crée une `ScheduleException` de type `cancelled` pour le mardi 11 mars. |
| **WHEN** | Le scheduler de notifications et de sessions s'exécute le lundi 10 mars soir et le mardi 11 mars soir. |
| **THEN** | **Lundi soir :** aucune session `pre_class` n'est proposée pour Physique-Chimie (le cours du lendemain est annulé). **Mardi soir :** aucune notification `capture_reminder` ni `review_reminder` n'est envoyée pour Physique-Chimie. Le créneau récurrent du mardi matin reste inchangé pour les semaines suivantes (le mardi 18 mars fonctionne normalement). |

> **NOTE :** L'exception est ponctuelle. Elle ne modifie pas le `ScheduleSlot` récurrent. Le scheduler résout les créneaux effectifs d'une semaine en combinant `ScheduleSlot` + `ScheduleException` du même `(user_id, subject)`.

### Z6-AC20 — Déplacement ponctuel d'un cours

| | |
|---|---|
| **GIVEN** | L'élève a un `ScheduleSlot` mardi matin pour Physique-Chimie. Il déplace le cours du mardi 11 mars vers jeudi 13 mars après-midi. |
| **WHEN** | Le système enregistre le déplacement. |
| **THEN** | Une `ScheduleException` est créée : `type = 'moved'`, `original_date = 2026-03-11`, `moved_to_date = 2026-03-13`, `moved_to_period = 'afternoon'`. **Mardi 11 :** aucune notification ni session `pre_class` pour PC (cours annulé ce jour). **Mercredi 12 soir :** une session `pre_class` est proposée pour PC (veille du cours déplacé à jeudi). **Jeudi 13 soir :** une notification `capture_reminder` ou `review_reminder` est envoyée pour PC. Le créneau récurrent mardi matin reprend normalement le mardi 18 mars. |

### Z6-AC21 — Exception sans impact sur les autres matières

| | |
|---|---|
| **GIVEN** | L'élève a un `ScheduleSlot` mardi matin pour PC et un `ScheduleSlot` mardi après-midi pour SVT. Le cours de PC du mardi 11 mars est annulé. |
| **WHEN** | Le scheduler s'exécute le mardi 11 mars soir. |
| **THEN** | La notification pour PC n'est PAS envoyée (cours annulé). La notification pour SVT EST envoyée normalement (pas d'exception sur SVT). L'exception est isolée par `(user_id, subject, original_date)`. |

### Z6-AC22 — Nettoyage automatique des exceptions passées

| | |
|---|---|
| **GIVEN** | L'élève a 5 `ScheduleException` dont 3 ont une `original_date` de plus de 30 jours. |
| **WHEN** | Le job de nettoyage s'exécute (quotidien). |
| **THEN** | Les 3 exceptions de plus de 30 jours sont supprimées. Les 2 exceptions récentes sont conservées. Aucune exception future n'est supprimée. |

### Z6-AC23 — Déplacement vers un jour déjà occupé (même matière)

| | |
|---|---|
| **GIVEN** | L'élève a un `ScheduleSlot` mardi matin pour PC et un `ScheduleSlot` jeudi matin pour PC. Il déplace le cours de mardi 11 mars vers jeudi 13 mars. |
| **WHEN** | Le scheduler évalue jeudi 13 mars. |
| **THEN** | Le cours PC du jeudi 13 mars existe déjà (créneau récurrent). Le déplacement ajoute un deuxième créneau PC le même jour. Les notifications ne sont envoyées qu'**une seule fois** pour PC ce jour-là (déduplication par `(user_id, subject, date)`). La session `pre_class` du mercredi soir couvre PC une seule fois (pas de doublon). |

> **NOTE :** Le déplacement vers un jour où la matière est déjà prévue est autorisé (l'élève peut avoir 2h de PC le même jour). Le système déduplique les notifications et sessions mais ne bloque pas la saisie.

---

> Ces 74 AC couvrent les zones à risque identifiées pour le vibe coding. Ils sont conçus pour être directement transformés en tests (Jest / Pytest / Playwright). Chaque session de génération de code doit recevoir les AC de la zone concernée comme contexte système, avec l'instruction explicite de générer les tests correspondants avant le code d'implémentation (TDD-first).

*Fin du document — Révise Mieux AC v1.1 · 6 mars 2026*
