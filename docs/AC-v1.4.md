# Acceptance Criteria — Révise Mieux v1.4

> **Annexe PRD v1.4 · Zones à risque vibe coding**
>
> | | |
> |---|---|
> | **Version** | 1.4 |
> | **Date** | 7 mars 2026 |
> | **Périmètre** | 6 zones critiques identifiées — 121 AC en format Given/When/Then |
> | **Évolutions v1.4 vs v1.3** | +15 AC : confiance parent & RGPD (rétention crops Z2, score mock exam + exclusion script 3 min Z1, digest standardisé + signalement OCR parent + feedback résolution admin + labels maîtrise traduits Z6) + session experience (variété gabarits + feedback enrichi + bouton passer + petit chapitre Z4) + intégrité données (rétractation validation erronée + anti-clicking aveugle Z3, versioning LLM Z2) + UX résilience (progression globale + archivage chapitre + persistance réseau Z6) + robustesse planning (recalcul intervalles sur modif date exam Z1, anti-lassitude questions Z4) |
> | **Évolutions v1.3 vs v1.2** | +9 AC anti-désengagement : plafond maîtrise items non validés (Z1), micro-célébrations + débrief session (Z1), retour en douceur après absence + cycle post-exam + rampe diagnostic + digest pré-contrôle + anti alert-fatigue parent (Z6) |
> | **Évolutions v1.2 vs v1.1** | +8 AC couvrant les angles morts identifiés : retry élève pipeline (Z2), re-vérification fidelity timeout + UX validation + SLA admin + détection précoce (Z3), normalisation ponctuation OCR (Z5), ré-engagement inactivité + alerte exams simultanés (Z6) |
> | **Usage** | À intégrer comme contexte système avant chaque session de vibe coding, et à transformer en tests unitaires |

---

## Zones couvertes

| # | Zone | Risque | AC count |
|---|---|---|---|
| Z1 | Transitions Mastery (états + régressions + engagement + reporting) | Très élevé | 20 |
| Z2 | Pipeline J0 — Error paths, timeouts & RGPD | Très élevé | 13 |
| Z3 | Validation HITL — Skip / Ignore / Qualité items | Élevé | 23 |
| Z4 | Lazy generation — Concurrence, cache & session experience | Élevé | 15 |
| Z5 | ChapterRevision — Identité Item & héritage | Élevé | 9 |
| Z6 | Emploi du temps, Notifications, Engagement & Confiance parent | Élevé | 41 |
| | **Total** | | **121** |

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

### Z1-AC15 — Plafond maîtrise OK pour items restreints aux templates simples

| | |
|---|---|
| **GIVEN** | Un item avec `validation_required = true` qui ne reçoit que des gabarits simples (GEN.KNOW.DEF_SHORT, GEN.KNOW.FLASH_MCQ — cf. Z3-AC01). L'item est en état **OK** avec `consecutive_successes = 2`. |
| **WHEN** | L'élève répond correctement à une question liée à cet item (gabarit simple). |
| **THEN** | L'état reste **OK** (plafonné). L'item ne peut PAS passer **SOLID** tant que `validation_required = true`. `consecutive_successes` est incrémenté normalement mais la transition OK → SOLID est bloquée. Le dashboard affiche un badge « maîtrise partielle — vérification requise » sur cet item. |

> **NOTE :** C'est le AC le plus critique pour l'intégrité de la maîtrise. Sans ce plafond, un item potentiellement hallucé (fidelity_score null ou < 0.5) peut atteindre SOLID via des QCM triviaux. Le parent voit alors une maîtrise à 80%+ qui ne reflète pas la réalité. Ce AC empêche structurellement la pollution du signal mastery par des items non vérifiés. La résolution de la ValidationTask (Z3-AC02/03) lève automatiquement le plafond.

### Z1-AC16 — Micro-célébration sur transitions de maîtrise positives

| | |
|---|---|
| **GIVEN** | Un item vient de transiter vers un nouvel état positif (UNKNOWN → FRAGILE, FRAGILE → OK, OK → SOLID). |
| **WHEN** | La transition est enregistrée en base (Mastery state update). |
| **THEN** | L'interface affiche un feedback visuel de célébration adapté à la transition : — UNKNOWN → FRAGILE : message « Bien joué, tu commences à maîtriser [term] ! » (encouragement léger). — FRAGILE → OK : message « [term] est de mieux en mieux — continue comme ça ! » + animation subtile. — OK → SOLID : message « [term] est acquis — bravo ! 🎯 » + animation marquée + compteur d'items SOLID du chapitre incrémenté visiblement. Le feedback est affiché en fin de question (après le feedback de correction), pendant 2 secondes, et ne bloque pas la navigation vers la question suivante. |

> **NOTE :** L'absence de célébration est le premier facteur de désengagement identifié chez les 11-15 ans. Le service valorise la qualité (« tu maîtrises ce concept ») plutôt que la quantité (pas de streak). Les animations sont légères et non-bloquantes — le but est un micro-shot de dopamine, pas une interruption. Ce AC complète Z6-AC18 (pas de streak) : on ne célèbre pas la régularité mais la progression réelle.

### Z1-AC17 — Débrief de fin de session

| | |
|---|---|
| **GIVEN** | L'élève termine une session (toutes les questions répondues ou TTL expiré avec ≥ 1 question répondue). |
| **WHEN** | La session passe en `status = COMPLETED`. |
| **THEN** | Un écran de débrief est affiché avec : — Score global de la session (X/Y correctes). — Liste des items ayant progressé positivement (transitions vers un état supérieur) avec le nouveau badge. — 1 item prioritaire à revoir (le plus fragile encore, avec `next_due_at` le plus proche). — Message de fermeture contextuel : si session evening_first → « Super première prise de contact ! ». Si session pre_class → « Tu es prêt(e) pour demain ! ». Si session daily → « Bonne révision, à demain ! ». — Bouton unique « Terminer » (pas de partage, pas de gamification complexe). Le débrief est optionnel : l'élève peut fermer l'app sans le lire (pas de blocage). |

> **NOTE :** Le débrief est le moment le plus important pour la rétention. Un élève qui ne sait pas s'il a progressé ne reviendra pas. Ce écran doit être rapide (< 3 secondes de chargement), positif (mettre en avant les progrès, pas les échecs) et actionnable (montrer le prochain objectif). L'absence de débrief est le 2ème facteur de churn identifié après l'absence de célébration.

### Z1-AC18 — Score contrôle blanc : caveat qualité si items non validés

| | |
|---|---|
| **GIVEN** | L'élève complète un contrôle blanc (`mock_exam`). Sur 20 questions, 6 portent sur des items avec `validation_required = true` (restreints aux templates simples, cf. Z3-AC01). |
| **WHEN** | Le score du contrôle blanc est calculé et affiché (à l'élève et dans le résumé parent). |
| **THEN** | Le score global est affiché normalement (ex. 14/20). Un sous-texte est ajouté : « Score basé sur [14] questions complètes et [6] questions simplifiées (contenu en cours de vérification) ». Le résumé envoyé au parent inclut la même mention. Le score est accompagné d'un indicateur de confiance : `score_confidence = (questions_full_templates / total_questions)` — ici 0.70. Si `score_confidence < 0.5`, un avertissement explicite est ajouté : « Plus de la moitié des questions étaient simplifiées — ce score est peu représentatif. Encouragez [Prénom] à vérifier les zones incertaines. » |

> **NOTE :** Sans ce caveat, le parent voit « 14/20 » et pense que l'enfant est prêt. Mais 6 questions étaient des QCM simples au lieu d'exercices de calcul ou rédaction — le score est structurellement gonflé. Ce AC rend l'inflation visible et actionnable. Le `score_confidence` est également exploité par le digest (Z6-AC35).

### Z1-AC19 — Script 3 minutes : exclusion des items sous investigation

| | |
|---|---|
| **GIVEN** | Le parent consulte le « script 3 minutes » pour un chapitre. 3 items sont les plus fragiles : Item A (FRAGILE, `validation_required = false`), Item B (FRAGILE, `validation_required = true`, `source = 'fidelity_check'`), Item C (UNKNOWN, `anomaly_flag = 'high_failure_rate'`). |
| **WHEN** | Le système compose les 2–3 questions orales du script. |
| **THEN** | Item B et Item C sont **exclus** du script (items sous investigation). Seul Item A est inclus. Si moins de 2 items sont éligibles après exclusion, le script est complété avec des items OK récemment révisés (consolidation orale). Le script n'inclut **jamais** un item avec `validation_required = true` ou `anomaly_flag != null`. Un message est affiché si des items ont été exclus : « [N] point(s) sont en cours de vérification et ne sont pas inclus dans le script. » |

> **NOTE :** Le script 3 minutes est le moment où le parent teste activement l'enfant à l'oral. Si le parent pose une question basée sur un item hallucé ou défectueux, et que l'enfant répond correctement selon le cours réel (pas l'item erroné), le parent conclut que l'app est défaillante. C'est un des moments de rupture de confiance les plus forts.

### Z1-AC20 — Modification de la date d'exam → recalcul des intervalles compressés

| | |
|---|---|
| **GIVEN** | Un exam « Contrôle HG » est fixé au 20 mars. 12 items du chapitre sont en états variés (3 SOLID, 5 OK, 4 FRAGILE). Les `next_due_at` ont été compressés selon Z1-AC08 (cap = exam_date − 1 jour). L'élève modifie la date d'exam au 27 mars (+7 jours). |
| **WHEN** | La mise à jour de l'exam est sauvegardée. |
| **THEN** | Tous les `next_due_at` des items liés à cet exam sont **recalculés** avec la nouvelle `exam_date`. Les intervalles reprennent les valeurs standard (1j/3j/7j) si le nouveau `T` le permet, sinon la compression est recalculée proportionnellement au nouveau `T`. Le recalcul ne touche pas les items dont le `next_due_at` est déjà passé (ils restent dus immédiatement). Si la date est avancée (ex: 20 mars → 15 mars), les intervalles se compriment davantage et un avertissement s'affiche : « Tu as peu de temps — les sessions seront plus fréquentes pour ce chapitre. » Si la date est repoussée, un message positif : « Plus de temps pour bien réviser ! » Le digest parent suivant mentionne le changement de date. |

> **NOTE :** Un contrôle reporté par le prof est un cas fréquent au collège. Si l'élève met à jour la date mais que les intervalles restent compressés sur l'ancienne date, il révisera inutilement de manière intensive pendant 7 jours de plus. Inversement, si le contrôle est avancé et que les intervalles ne se compriment pas, l'élève arrive sous-préparé. Le recalcul automatique maintient la cohérence du plan de révision.

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

### Z2-AC11 — Relance élève pour pages en échec de génération

| | |
|---|---|
| **GIVEN** | Une page du chapitre est en `status = 'items_generation_failed'` (erreur LLM après retries). Le texte OCR brut est conservé. |
| **WHEN** | L'élève consulte la carte de leçon du chapitre. |
| **THEN** | Un badge 'Exercices non générés · page X' est affiché avec un bouton 'Réessayer'. Le clic déclenche un nouveau passage par l'étape 7 du pipeline (génération items) en réutilisant le texte OCR en cache. Si le retry réussit, les items sont ajoutés à la carte et le badge disparaît. Si le retry échoue à nouveau, le badge réapparaît avec le message 'Génération toujours indisponible — réessaie plus tard'. Maximum 3 retries manuels par page. Au-delà, seul l'admin peut relancer. |

> **NOTE :** Ce mécanisme complète Z2-AC05 en offrant une action côté élève. L'admin reste le fallback ultime mais l'élève n'est plus bloqué sans recours en cas d'indisponibilité temporaire du LLM.

### Z2-AC12 — Rétention crops et OCR alignée sur politique photos (RGPD)

| | |
|---|---|
| **GIVEN** | Une page a été traitée par le pipeline. Les crops (Block.crop, Document.source_image_url) et le texte OCR brut sont stockés. La photo originale est supprimée à J+30 (politique par défaut). |
| **WHEN** | Le job de nettoyage RGPD s'exécute à J+30 pour cette page. |
| **THEN** | Si l'utilisateur n'a PAS opté pour la conservation : les crops d'image sont supprimés en même temps que la photo originale. Les `crop_url` des blocs sont remplacés par `null`. Les `source_image_url` des Documents sont remplacés par `null`. Le texte OCR brut est conservé (il ne contient pas l'image de l'écriture manuscrite). Les ValidationTasks en cours conservent un `crop_snapshot_text` (description textuelle du crop) mais pas l'image. Les gabarits de type `GEN.DOC.IMAGE.*` deviennent inéligibles pour les Documents dont le `source_image_url` est `null` — ces items sont restreints aux gabarits textuels. |

> **NOTE :** Le PRD §20 Q2 mentionne la suppression des « photos originales » à J+30, mais les crops (fragments d'image) et les `source_image_url` des Documents n'avaient pas de politique de rétention explicite. Cela créait un trou RGPD : un parent pensait les photos supprimées alors que des fragments persistaient indéfiniment. Ce AC aligne la rétention des crops sur celle des photos originales.

### Z2-AC13 — Versioning du modèle LLM pour reproductibilité et détection de drift

| | |
|---|---|
| **GIVEN** | Le pipeline J0 (segmentation, item generation, fidelity check) et la lazy generation utilisent des appels LLM. Le modèle LLM sous-jacent peut changer (mise à jour provider, bascule de modèle, changement de prompt). |
| **WHEN** | Un appel LLM est effectué à n'importe quelle étape du pipeline ou de la génération. |
| **THEN** | Chaque résultat LLM (Item, Question, fidelity_score) est taggé avec `llm_model_version` (identifiant du modèle, e.g. `gpt-4o-2024-08-06`) et `prompt_template_version` (hash ou version sémantique du prompt utilisé). Le champ `llm_model_version` est indexé. Un changement de modèle ou de prompt déclenche une alerte admin `LLM_VERSION_CHANGED`. Un job hebdomadaire compare les métriques qualité (taux fidelity_score < 0.6, taux signalements élève) entre l'ancienne et la nouvelle version. Si le taux de dégradation dépasse 15% sur l'un des indicateurs, l'admin reçoit une alerte `LLM_DRIFT_DETECTED` avec détail comparatif. |

> **NOTE :** Sans versioning LLM, un changement de modèle silencieux (ex. le provider met à jour le modèle derrière la même API) peut dégrader la qualité des items générés sans qu'on puisse identifier la cause. Le versioning permet le diagnostic (« depuis quand les fidelity_score baissent-ils ? ») et le rollback informé. C'est aussi une exigence de traçabilité pour un produit éducatif destiné à des mineurs.

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

### Z3-AC10 — Vérification croisée LLM : item fidèle au texte OCR

| | |
|---|---|
| **GIVEN** | Le pipeline J0 a généré un item KNOWLEDGE avec `term = "photosynthèse"` et `keywords = ["chloroplaste", "lumière", "CO2"]` à partir d'un bloc OCR contenant « La photosynthèse est le processus par lequel les plantes utilisent la lumière, le CO2 et l'eau pour produire de la matière organique dans les chloroplastes ». |
| **WHEN** | L'étape 7b (vérification croisée LLM) s'exécute. |
| **THEN** | Le `fidelity_score` est `≥ 0.7` (item fidèle au texte source). `fidelity_flag = null`. L'item n'est pas ajouté à la file de validation pour cette raison. |

### Z3-AC11 — Vérification croisée LLM : item déformé par le LLM

| | |
|---|---|
| **GIVEN** | Le pipeline J0 a généré un item KNOWLEDGE avec `term = "respiration cellulaire"` à partir d'un bloc OCR qui parle uniquement de photosynthèse (hallucination LLM — le terme n'apparaît pas dans le texte source). |
| **WHEN** | L'étape 7b (vérification croisée LLM) s'exécute. |
| **THEN** | Le `fidelity_score` est `< 0.5`. `fidelity_flag = 'low'`. `validation_required = true`. Une `ValidationTask` est créée avec `source = 'fidelity_check'` et `suggestion = "L'item ne correspond pas au texte source — vérifier"`. L'item est restreint aux templates simples (GEN.KNOW.DEF_SHORT, GEN.KNOW.FLASH_MCQ). |

### Z3-AC12 — Vérification croisée LLM : timeout du service

| | |
|---|---|
| **GIVEN** | L'étape 7b est lancée mais le service LLM de vérification ne répond pas dans le délai imparti (timeout). |
| **WHEN** | Le timeout expire. |
| **THEN** | L'item est conservé avec `fidelity_score = null` et `fidelity_flag = null`. Le pipeline continue normalement (comportement dégradé = confiance OCR seule, pas de blocage). L'incident est loggé pour monitoring. |

### Z3-AC13 — Cohérence intra-chapitre : détection de doublons

| | |
|---|---|
| **GIVEN** | Le pipeline a généré deux items dans le même chapitre : Item A (`term = "chloroplaste"`, `confidence = 0.9`) et Item B (`term = "chloroplaste"`, `confidence = 0.7`). |
| **WHEN** | L'étape 7c (cohérence intra-chapitre) s'exécute. |
| **THEN** | Les deux items sont identifiés comme doublons (même `term`). L'item de plus faible confidence (Item B, 0.7) est archivé automatiquement (`archived = true`). L'item A est conservé. Aucune `ValidationTask` n'est créée (résolution automatique). Le Mastery associé à Item B, s'il existe, est transféré à Item A. |

### Z3-AC14 — Cohérence intra-chapitre : détection de contradictions

| | |
|---|---|
| **GIVEN** | Le pipeline a généré deux items dans le même chapitre : Item A (`term = "densité"`, définition = « masse divisée par le volume ») et Item B (`term = "masse volumique"`, définition = « volume divisé par la masse »). La détection LLM identifie une contradiction. |
| **WHEN** | L'étape 7c s'exécute. |
| **THEN** | Les deux items sont flaggés `coherence_flag = 'contradiction'` et `validation_required = true`. Une `ValidationTask` est créée pour chacun avec `source = 'coherence_check'` et `suggestion = "Contradiction détectée avec l'item [autre_item_id] — vérifier les définitions"`. Les deux items sont restreints aux templates simples jusqu'à résolution. |

### Z3-AC15 — Feedback élève : signalement d'erreur sur un item

| | |
|---|---|
| **GIVEN** | L'élève est en session de révision. Une question affiche « La photosynthèse produit du méthane » (item mal extrait). |
| **WHEN** | L'élève appuie sur « Signaler une erreur » et saisit optionnellement « C'est de l'O2, pas du méthane ». |
| **THEN** | Une `ValidationTask` est créée avec `source = 'student_report'`, `student_note = "C'est de l'O2, pas du méthane"`, `priority = HIGH` (signalement élève toujours prioritaire). L'item reste utilisable en mode dégradé (templates simples uniquement). Si une `ValidationTask` existe déjà pour cet item, le signalement est ajouté comme note complémentaire sur la tâche existante (pas de doublon). La réponse de l'élève à cette question n'est **pas** comptée dans le score Mastery (item sous investigation). |

### Z3-AC16 — Détection par taux d'échec anormal

| | |
|---|---|
| **GIVEN** | Un item en état FRAGILE a reçu 6 tentatives sur les 7 derniers jours, dont 5 échecs (taux d'échec = 83%). |
| **WHEN** | Le job quotidien de détection d'anomalies s'exécute. |
| **THEN** | `anomaly_flag = 'high_failure_rate'` est positionné. `validation_required = true`. Une `ValidationTask` est créée avec `source = 'anomaly_detection'` et `suggestion = "Taux d'échec anormal (83%) — vérifier l'item"`. L'item est restreint aux templates simples jusqu'à vérification. |

### Z3-AC17 — Détection par taux d'échec : exclusion des items UNKNOWN

| | |
|---|---|
| **GIVEN** | Un item en état UNKNOWN a reçu 5 tentatives, toutes en échec (taux = 100%). |
| **WHEN** | Le job quotidien de détection d'anomalies s'exécute. |
| **THEN** | L'item n'est **pas** flaggé `anomaly_flag` car il est en état UNKNOWN (taux d'échec élevé attendu à la première exposition). Aucune `ValidationTask` créée. L'item continue à être proposé normalement pour permettre l'apprentissage. Le job ne considère que les items en état FRAGILE, OK ou SOLID. |

> **NOTE :** Ces 4 mécanismes (vérification croisée, cohérence, feedback élève, détection anomalie) forment une boucle de qualité continue : la vérification croisée et la cohérence agissent en amont (pipeline J0), le feedback élève en temps réel, et la détection par taux d'échec en aval (post-usage). Un item peut cumuler plusieurs flags simultanément.

### Z3-AC18 — Re-vérification fidelity différée (rattrapage timeout)

| | |
|---|---|
| **GIVEN** | Un item a `fidelity_score = null` et `fidelity_flag = null` suite à un timeout de l'étape 7b (cf. Z3-AC12). |
| **WHEN** | Le job quotidien de maintenance qualité s'exécute. |
| **THEN** | L'item est automatiquement soumis à une nouvelle vérification fidelity (étape 7b). Si le service LLM répond : le `fidelity_score` et `fidelity_flag` sont mis à jour normalement. Si `fidelity_score < 0.5`, l'item passe en `validation_required = true` avec création de ValidationTask (cf. Z3-AC11). Si le timeout se reproduit 3 jours consécutifs, l'item est flaggé `validation_required = true` avec `source = 'fidelity_timeout_persistent'` et restreint aux templates simples. |

> **NOTE :** Ce mécanisme empêche les hallucinations LLM de rester indéfiniment non vérifiées. L'item ne peut pas rester en `fidelity_score = null` plus de 3 jours sans action corrective.

### Z3-AC19 — UX clarification "Ignorer" vs "Je ne sais pas"

| | |
|---|---|
| **GIVEN** | L'élève consulte une ValidationTask avec les actions disponibles : 'Confirmer', 'Corriger', 'Je ne sais pas', 'Ignorer'. |
| **WHEN** | L'écran de validation s'affiche. |
| **THEN** | Chaque action affiche un sous-texte explicatif permanent (pas un tooltip) : — 'Confirmer' → « C'est correct, je valide » — 'Corriger' → « Je corrige moi-même » — 'Je ne sais pas' → « Mon prof ou un adulte vérifiera » — 'Ignorer' → « Retirer de mes révisions ». Le sous-texte de 'Ignorer' précise en rouge atténué : « Cet élément ne sera plus proposé en exercice tant qu'il n'est pas résolu ». |

> **NOTE :** Ce AC adresse le risque de confusion sémantique entre "Ignorer" et "Je ne sais pas" identifié comme source de perte de confiance. La clarification permanente (pas hover/tooltip) est essentielle pour un public collégien.

### Z3-AC20 — Escalade admin : SLA 7 jours sur ValidationTasks non résolues

| | |
|---|---|
| **GIVEN** | Une ValidationTask est en statut `PENDING` ou `DEFERRED_BY_STUDENT` depuis 7 jours. L'item associé a `validation_required = true`. |
| **WHEN** | Le job quotidien de suivi qualité s'exécute. |
| **THEN** | La ValidationTask est promue en priorité `CRITICAL`. Une alerte admin est créée dans le backoffice avec le tag `sla_breach`. Si la tâche reste non résolue à J+14, l'item est automatiquement restreint aux templates de type QCM uniquement (`GEN.KNOW.FLASH_MCQ`) et un compteur `unresolved_validation_days` est incrémenté dans le dashboard admin. Le KPI « % ValidationTasks résolues < 7j » est tracké. |

> **NOTE :** Sans SLA, les ValidationTasks admin s'accumulent silencieusement. Ce mécanisme garantit une dégradation progressive plutôt qu'un oubli. Le passage en QCM-only à J+14 protège l'élève d'exercices trompeurs sur des items non vérifiés.

### Z3-AC21 — Détection anomalie précoce sur items validation_required

| | |
|---|---|
| **GIVEN** | Un item avec `validation_required = true` (templates restreints) a reçu 3 tentatives, dont 3 échecs (taux = 100%). |
| **WHEN** | Le job quotidien de détection d'anomalies s'exécute. |
| **THEN** | L'item est flaggé `anomaly_flag = 'high_failure_rate'` malgré le seuil normal de 5 tentatives (cf. Z3-AC16). Le seuil est abaissé à **3 tentatives** pour les items `validation_required = true` car la probabilité que l'item soit défectueux est plus élevée. La ValidationTask existante est promue en priorité `HIGH` si elle ne l'est pas déjà. |

> **NOTE :** Les items non validés sont plus susceptibles d'être défectueux. Attendre 5 échecs sur un item déjà suspect fait subir à l'élève des échecs évitables qui érodent sa confiance. Ce seuil abaissé ne s'applique qu'aux items `validation_required = true`.

### Z3-AC22 — Rétractation d'une validation erronée (parent ou élève)

| | |
|---|---|
| **GIVEN** | Un parent (ou un élève) a confirmé un item via Z3-AC02 (confidence boostée à 0.85). Trois jours plus tard, l'élève signale une erreur sur ce même item (Z3-AC15) ou le taux d'échec dépasse le seuil (Z3-AC16/AC21). |
| **WHEN** | Une nouvelle ValidationTask est créée pour un item déjà `RESOLVED_CONFIRMED`. |
| **THEN** | L'item repasse en `validation_required = true`. La confidence est ramenée à `min(item.confidence, 0.7)` (annulation du boost). Les gabarits sont re-restreints aux templates simples. La ValidationTask originale est marquée `REOPENED` avec une note « Rouvert suite à [source : signalement élève / anomalie détectée] ». L'admin est notifié avec priorité `HIGH`. Les Mastery updates effectués entre la confirmation et la réouverture ne sont **pas** annulés (pas de rétroactivité). |

> **NOTE :** C'est le filet de sécurité pour les validations erronées. Un parent qui confirme « ρ = m/V » alors que l'OCR a mal lu « ρ = m × V » verrouille une erreur en base. Sans rétractation, l'enfant étudie du contenu faux avec une confidence de 0.85. Le signalement élève (Z3-AC15) ou la détection d'anomalie (Z3-AC16) servent de second regard. La non-rétroactivité des Mastery est un compromis pragmatique : corriger le contenu suffit, recalculer le passé serait trop complexe et déstabilisant.

### Z3-AC23 — Détection de réponses trop rapides (anti-clicking aveugle)

| | |
|---|---|
| **GIVEN** | L'élève répond à une question MCQ en moins de **2 secondes** (temps entre affichage de la question et soumission de la réponse). |
| **WHEN** | La réponse est soumise. |
| **THEN** | La réponse est **acceptée et corrigée normalement** (pas de blocage UX). Mais elle est marquée `rapid_response = true` dans l'Attempt. Le Mastery state est mis à jour **uniquement si la réponse est incorrecte** (régression appliquée normalement). Si la réponse est correcte : `consecutive_successes` n'est **pas** incrémenté et la transition mastery n'est **pas** appliquée. L'item reste dans l'état actuel et sera re-proposé à la prochaine session. Un compteur `rapid_correct_count` est maintenu par session. Si `rapid_correct_count ≥ 3` dans une même session, un message non-bloquant apparaît : « Prends ton temps pour bien lire les questions — tes réponses rapides ne comptent pas pour ta progression. » |

> **NOTE :** Un élève qui clique random sur des MCQ a 25% de chance de répondre juste (4 options). Sans ce AC, 2 MCQ correctes par chance suffisent pour passer de UNKNOWN à OK (Z1-AC01/AC02). Ce AC neutralise les réponses trop rapides côté progression sans bloquer l'UX (l'élève peut toujours cliquer, mais ça ne "compte" pas positivement). La régression sur réponse incorrecte est maintenue car elle incite à réfléchir plutôt qu'à cliquer au hasard. Le seuil de 2 secondes est calibré sur le temps minimum de lecture d'une question MCQ (titre + 4 options).

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

### Z4-AC11 — Variété de gabarits dans une session (anti-monotonie)

| | |
|---|---|
| **GIVEN** | Une session `daily` est composée avec 10 questions. Le pool contient des questions de types FLASH_MCQ, DEF_SHORT, CLOZE_KEYWORDS, ASSOC_TERM_DEF, et DOC.EXTRACT_EVIDENCE. |
| **WHEN** | Le moteur de composition ordonne les questions dans la session. |
| **THEN** | Aucun gabarit identique (`template_id`) ne peut apparaître plus de **2 fois consécutives**. Si la contrainte ne peut pas être respectée (pool trop petit), elle est relaxée mais loggée. La session alterne les types de questions (`question_type`) autant que possible : pas 3 MCQ d'affilée si des SHORT_ANSWER sont disponibles. L'ordonnancement est déterministe (pas random) : il alterne les types par round-robin sur les types disponibles. |

> **NOTE :** 5 flashcards consécutives rendent la session monotone et mécanique. L'alternance des types de questions maintient l'attention et active différents circuits cognitifs (reconnaissance ≠ rappel ≠ production). Cette contrainte n'affecte pas la sélection des items (qui reste 70/20/10), seulement l'ordonnancement des questions dans la session.

### Z4-AC12 — Feedback enrichi après réponse incorrecte

| | |
|---|---|
| **GIVEN** | L'élève répond incorrectement à une question de type KEYWORDS sur l'item « photosynthèse » (keywords attendus : chloroplaste, lumière, CO2). L'élève a répondu « les plantes font de la nourriture ». |
| **WHEN** | La correction est affichée. |
| **THEN** | Le feedback contient **3 éléments obligatoires** : — **1. La réponse correcte** : « Les mots-clés attendus étaient : chloroplaste, lumière, CO2 ». — **2. Ce qui manquait** (spécifique au type) : KEYWORDS → mots-clés manquants listés. NUMERIC → valeur attendue + unité + formule utilisée. MCQ → explication du distractor choisi (« Tu as choisi X — en réalité, X est faux car... »). — **3. Un indice pour la prochaine fois** (1 phrase max) : « Retiens que la photosynthèse se passe dans les chloroplastes grâce à la lumière et au CO2. ». Le feedback est généré par template (pas par LLM en temps réel) pour garantir la cohérence et la vitesse (< 200 ms). Le feedback est toujours factuel, jamais culpabilisant. |

> **NOTE :** Un feedback qui dit juste « Faux — la bonne réponse est X » n'enseigne rien. L'indice en 1 phrase est le micro-moment d'apprentissage le plus puissant de la session. Le contenu est templaté (pas LLM live) pour garantir la rapidité et éviter les hallucinations dans le feedback lui-même.

### Z4-AC13 — Bouton « Passer » sur une question (sans pénalité mastery)

| | |
|---|---|
| **GIVEN** | L'élève est bloqué sur une question de type NUMERIC. Il ne connaît pas la formule. |
| **WHEN** | L'élève appuie sur « Passer cette question ». |
| **THEN** | La question est marquée `status = SKIPPED`. Le Mastery state de l'item n'est **PAS modifié** (ni progression, ni régression). `consecutive_successes` n'est pas remis à 0. La question est placée en fin de session (si la session a encore ≥ 3 questions restantes) pour une seconde chance. Si l'élève la passe à nouveau, elle est comptée comme non répondue. Le feedback de la réponse correcte est affiché après le 2ème passage (l'élève voit la solution même s'il n'a pas répondu). Maximum 2 « Passer » par session (au-delà, le bouton est grisé). |

> **NOTE :** Sans mécanisme « Passer », un élève bloqué sur une question de calcul reste immobile → frustration → fermeture de l'app. Le « Passer » sans pénalité mastery est cohérent avec le principe « seule une réponse incorrecte déclenche une régression » (Z1-AC05 à Z1-AC07). La limite de 2 passes par session empêche l'abus (tout passer sans réfléchir). La solution affichée après le 2ème passage transforme un moment de blocage en moment d'apprentissage.

### Z4-AC14 — Session viable sur petit chapitre (< 5 items)

| | |
|---|---|
| **GIVEN** | Un chapitre n'a que 3 items valides. Le moteur de composition doit créer une session `daily`. |
| **WHEN** | La composition est lancée. |
| **THEN** | La session est composée avec **au minimum 4 questions** : chaque item génère au moins 1 question, et l'item le plus fragile en génère 2 (gabarits différents sur le même item). La durée cible est réduite à 3–5 min (au lieu de 10–20 min). Si le pool de gabarits éligibles est épuisé (tous les gabarits déjà utilisés pour ces 3 items), la session utilise des reformulations : même item + même gabarit mais avec des distractors différents (MCQ) ou un ordre de keywords différent (CLOZE). Le message d'introduction adapte les attentes : « Petite session rapide — [N] questions sur ce chapitre ». |

> **NOTE :** Un chapitre avec 3 items est courant (élève qui n'a photographié qu'une seule page, ou cours très court). Sans ce AC, la session serait de 2 questions identiques à la veille — l'élève sent qu'il tourne en rond. La reformulation (distractors différents, ordre différent) crée une illusion de nouveauté tout en testant les mêmes connaissances sous des angles différents.

### Z4-AC15 — Anti-lassitude : renouvellement des questions vues fréquemment

| | |
|---|---|
| **GIVEN** | Un item FRAGILE a été présenté 4 fois à l'élève dans les 7 derniers jours. Le pool contient 3 questions pour cet item, toutes déjà vues (identifiées via `question_id` dans les Attempts récents). |
| **WHEN** | Le moteur de composition sélectionne cet item pour la session suivante. |
| **THEN** | Le pool vérifie si l'élève a déjà vu toutes les questions disponibles pour cet item au cours des 5 dernières sessions. Si oui, une **régénération ciblée** est déclenchée pour cet item uniquement : le LLM génère 1–2 nouvelles questions avec des gabarits ou des angles différents (distractors variés, reformulation de la consigne). Les anciennes questions restent dans le pool (elles redeviennent éligibles après 14 jours sans vue). La régénération est **lazy** et non bloquante : si le LLM est indisponible, une question déjà vue est réutilisée plutôt que de bloquer la session. Un compteur `times_seen` est maintenu par `(user_id, question_id)` pour informer l'algorithme de sélection (préférence aux questions les moins vues). |

> **NOTE :** Un item FRAGILE en spaced repetition est revu toutes les 24h. Avec un pool de 3 questions, l'élève voit la même MCQ au bout de 3 jours. Au 7ème jour, il reconnaît la question et la réponse par mémoire photographique — il ne révise plus le concept, il reconnaît le pattern visuel. C'est une forme de mastery inflation silencieuse. Le renouvellement ciblé force le cerveau à réengager avec le concept sous un angle neuf.

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

### Z5-AC09 — Normalisation insensible à la ponctuation et aux abréviations

| | |
|---|---|
| **GIVEN** | L'item `I.D.H.` existe dans R1 (état OK). Dans R2, l'OCR produit `IDH` (sans points). |
| **WHEN** | La normalisation de clé d'identité est appliquée (cf. Z5-AC01). |
| **THEN** | La normalisation supprime les points (`.`), tirets (`-`), barres obliques (`/`), espaces multiples et apostrophes typographiques avant comparaison. `normalized("I.D.H.") == normalized("IDH") == "idh"`. L'héritage Mastery s'applique (cf. Z5-AC02). La liste des caractères supprimés est configurable par pack (pour les cas où le tiret est sémantique, ex. `demi-vie`). |

> **NOTE :** Ce AC complète Z5-AC01 et Z5-AC03 pour couvrir les variations OCR fréquentes sur les sigles et abréviations (IDH/I.D.H., PIB/P.I.B., pH/p.H.). Sans cette normalisation étendue, chaque variation OCR crée un doublon et orpheline le Mastery existant — c'est la source principale de régression silencieuse sur les re-uploads. La configurabilité par pack permet de préserver les cas où la ponctuation est sémantique.

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

### Z6-AC24 — Notification parent : annulation / déplacement de cours (temps réel)

| | |
|---|---|
| **GIVEN** | L'élève a un parent lié (`linked_student_id`). Le parent a `schedule_change_enabled = true` dans ses `ParentNotificationPref`. L'élève annule son cours de PC du mardi 11 mars. |
| **WHEN** | L'élève confirme l'annulation. |
| **THEN** | Une notification push est envoyée **immédiatement** au parent : « [Prénom] a annulé son cours de Physique-Chimie du mardi 11 mars ». L'événement apparaît dans le tableau de bord parent (section « 7 derniers jours »). L'élève n'est **pas** notifié que son parent a reçu l'alerte. Si le parent a `schedule_change_enabled = false`, aucune notification n'est envoyée mais l'événement reste visible dans le tableau de bord. |

### Z6-AC25 — Notification parent : session de révision manquée (lendemain matin)

| | |
|---|---|
| **GIVEN** | L'élève avait une session `pre_class` planifiée le lundi 10 mars soir pour PC. La session n'a pas été commencée à 23h59. Le parent a `missed_session_enabled = true`. |
| **WHEN** | Le scheduler parent s'exécute le mardi 11 mars matin (même heure que le rappel élève). |
| **THEN** | Notification push parent : « [Prénom] n'a pas fait sa session de révision de Physique-Chimie hier soir ». L'événement apparaît dans le tableau de bord parent. **Exception** : si l'élève a terminé la session entre 00h00 et 06h00 le mardi, la notification parent n'est **pas** envoyée (session comptée comme faite en retard). |

### Z6-AC26 — Notification parent : inactivité prolongée (3 jours)

| | |
|---|---|
| **GIVEN** | L'élève n'a eu aucune activité (capture, session, review) depuis 3 jours consécutifs. Le parent a `inactivity_enabled = true` et `inactivity_threshold_days = 3`. |
| **WHEN** | Le job d'inactivité s'exécute le matin du 4ème jour sans activité. |
| **THEN** | Notification push parent : « [Prénom] n'a pas utilisé ReviseMieux depuis 3 jours ». L'alerte est envoyée **une seule fois** par période d'inactivité. Aucune nouvelle notification tant que l'élève n'a pas repris une activité puis recommencé une nouvelle période d'inactivité. Le seuil est configurable par le parent (`inactivity_threshold_days`). |

### Z6-AC27 — Opt-out parent par catégorie

| | |
|---|---|
| **GIVEN** | Le parent désactive `missed_session_enabled` dans ses paramètres mais laisse `schedule_change_enabled` et `inactivity_enabled` activés. |
| **WHEN** | L'élève manque une session ET annule un cours le même jour. |
| **THEN** | Le parent reçoit **uniquement** la notification d'annulation de cours (schedule_change). Aucune notification pour la session manquée. Les deux événements restent visibles dans le tableau de bord parent (le tableau de bord n'est pas filtré par les préférences de notification). |

> **NOTE :** Les notifications parent respectent l'autonomie de l'élève. L'objectif est d'informer les parents sans créer une dynamique de surveillance. L'élève ne voit jamais « ton parent a été prévenu ». Le parent ne peut pas agir sur l'emploi du temps de l'élève depuis son compte.

### Z6-AC28 — Ré-engagement progressif après inactivité prolongée (7+ jours)

| | |
|---|---|
| **GIVEN** | L'élève n'a eu aucune activité depuis 7 jours consécutifs. Le rappel d'inactivité parent (Z6-AC26) a déjà été envoyé à J+3. |
| **WHEN** | Le job de ré-engagement s'exécute le matin du 8ème jour. |
| **THEN** | Une notification unique est envoyée à l'élève : « Tes révisions t'attendent — [N] points à consolider en [matière]. On reprend doucement ? ». La notification inclut un deep link vers une session courte (3–5 min, gabarits faciles, priorité items FRAGILE). Si l'élève ne réagit pas, **aucune relance** supplémentaire n'est envoyée (respect du choix). Le système ne relance qu'au prochain changement de contexte : nouvel Exam créé, nouveau chapitre uploadé, ou début de trimestre. |

> **NOTE :** Ce AC complète Z6-AC26 (alerte parent) et Z6-AC18 (pas de streak) en ajoutant un seul point de contact côté élève. L'approche "1 notification + deep link facile" respecte le principe anti-culpabilisation tout en offrant un chemin de retour à faible friction. Le déclencheur contextuel (exam, upload, trimestre) évite le harcèlement tout en maintenant des occasions naturelles de reprise.

### Z6-AC29 — Alerte dates d'exams simultanées sur même journée

| | |
|---|---|
| **GIVEN** | L'élève crée un Exam pour Physique-Chimie le 15 mars. Un Exam pour Maths existe déjà le 15 mars. |
| **WHEN** | L'élève sauvegarde le nouvel Exam. |
| **THEN** | Un avertissement non-bloquant est affiché : « Tu as déjà un contrôle de Maths le 15 mars — les deux révisions seront planifiées en parallèle ». L'Exam est créé normalement. Le moteur de planification répartit les sessions de révision en alternant les matières les jours précédant le 15 mars (pas de soirée 100% PC + soirée 100% Maths, mais un mix). Le contrôle blanc multi-exam n'est **pas** fusionné (chaque Exam garde son propre mock_exam). |

> **NOTE :** Sans cette détection, l'élève peut se retrouver avec deux contrôles blancs le même jour sans préparation équilibrée. L'alternance des matières dans les sessions pré-exam est plus efficace pour la mémorisation (interleaving effect) et évite la saturation cognitive sur une seule matière.

### Z6-AC30 — Session « retour en douceur » après absence prolongée

| | |
|---|---|
| **GIVEN** | L'élève n'a eu aucune activité depuis ≥ 5 jours. Il a 15 items en retard (`next_due_at < now`), dont 8 FRAGILE et 7 OK. Il ouvre l'app et lance une session. |
| **WHEN** | Le moteur de composition prépare la session. |
| **THEN** | La session est composée en mode « retour en douceur » : — Durée réduite : 5 min max (au lieu de 10–20 min). — Sélection : uniquement les 4–6 items les plus anciens en retard (pas les 15). — Gabarits : difficulté 1–2 uniquement (rappel, pas d'exercice long). — Message d'accueil : « Content de te revoir ! On reprend doucement avec quelques rappels. ». — Les items non sélectionnés restent en retard et seront proposés dans les sessions suivantes (étalement sur 3–5 jours). Le mode « retour en douceur » se désactive automatiquement après 2 sessions complétées consécutivement. |

> **NOTE :** Sans ce mécanisme, un élève qui revient après une semaine voit une session de 20 min bourrée d'items qu'il a oubliés → cascade d'échecs → sentiment d'incompétence → décrochage définitif. Le « retour en douceur » étale la dette sur plusieurs jours et utilise des gabarits faciles pour reconstruire la confiance avant de monter en difficulté. C'est le premier anti-pattern de churn identifié dans les apps de spaced repetition (cf. problème connu d'Anki).

### Z6-AC31 — Cycle de vie post-exam : archivage automatique

| | |
|---|---|
| **GIVEN** | Un Exam avec `exam_date = 10 mars` couvre les chapitres ch1 et ch2. La date est dépassée (`now > exam_date + 1 jour`). |
| **WHEN** | Le job quotidien de maintenance s'exécute le 12 mars. |
| **THEN** | L'Exam passe en `status = 'past'`. Les chapitres ch1 et ch2 ne sont **plus** soumis au resserrement de planning Z1-AC08 (les intervalles reviennent aux valeurs standard sans exam). Les items SOLID de ch1/ch2 passent en intervalle J+14 (repos long terme au lieu de J+7). Les items FRAGILE/OK gardent leurs intervalles standard (J+1, J+3). Les chapitres restent actifs et révisables mais ne sont plus prioritaires dans la session quotidienne. Un message « Contrôle passé — tes acquis sont en maintenance longue » est affiché sur la carte du chapitre. L'élève peut relancer un contrôle blanc à tout moment (utile pour un futur brevet ou examen global). |

> **NOTE :** Sans ce AC, les items post-exam continuent de saturer les sessions quotidiennes avec le même rythme qu'avant l'exam. L'élève a mentalement tourné la page mais l'app insiste. C'est la 3ème cause de désinstallation identifiée. Le passage en « maintenance longue » (J+14 pour SOLID) maintient l'ancrage sans fatiguer.

### Z6-AC32 — Diagnostic initial : rampe de difficulté progressive

| | |
|---|---|
| **GIVEN** | L'élève lance son premier diagnostic sur un chapitre avec 12 items UNKNOWN. |
| **WHEN** | Le moteur de composition prépare le diagnostic initial. |
| **THEN** | Les 2–3 premières questions sont des gabarits de difficulté 1 (FLASH_MCQ, DEF_SHORT) sur les items les plus simples (confidence la plus haute). Les questions suivantes montent progressivement en difficulté (2, puis 3 si disponible). Si l'élève enchaîne 3 échecs consécutifs, le moteur redescend en difficulté 1 pour les 2 questions suivantes avant de remonter. Le diagnostic ne commence jamais par un gabarit NUMERIC, RUBRIC ou ORDERING. En fin de diagnostic, le message de clôture est toujours positif : « Bonne première exploration ! Tu as [X] points acquis et [Y] à travailler — on s'y met dès ce soir ! ». |

> **NOTE :** La première impression détermine la rétention. Un diagnostic qui commence par un exercice de calcul complexe sur un chapitre jamais vu → échec → l'élève pense « cette app est trop dure ». La rampe progressive garantit 2–3 réussites rapides en début de session (effet psychologique de compétence perçue) avant d'augmenter le challenge.

### Z6-AC33 — Notification parent : digest hebdo calé sur calendrier d'exams

| | |
|---|---|
| **GIVEN** | Un Exam est prévu le mercredi 12 mars. Le digest hebdo parent est normalement envoyé le dimanche. |
| **WHEN** | Le scheduler prépare le digest de la semaine contenant un exam à J-5 ou moins. |
| **THEN** | Un digest supplémentaire « pré-contrôle » est envoyé **3 jours avant l'exam** (samedi 9 mars) en plus du digest hebdo standard. Ce digest inclut : — Titre : « Contrôle [Matière] dans 3 jours ». — Maîtrise par chapitre concerné (% items OK+SOLID). — Items encore FRAGILE/UNKNOWN (liste courte, max 5). — Recommandation : « Encouragez [Prénom] à faire un dernier contrôle blanc ce week-end ». Le digest standard du dimanche inclut une section « Contrôle dans 3 jours » en haut si non envoyé samedi. Aucun digest supplémentaire si le parent a opt-out de la catégorie. |

> **NOTE :** Le digest hebdo à date fixe ne suffit pas : un parent qui reçoit le récap dimanche pour un contrôle lundi n'a plus le temps d'agir. Le digest pré-contrôle à J-3 donne une fenêtre d'action (week-end). C'est le moment où le parent a le plus besoin du signal et où la valeur perçue du service est la plus haute.

### Z6-AC34 — Notification parent : résumé hebdo sessions manquées (anti alert-fatigue)

| | |
|---|---|
| **GIVEN** | L'élève a manqué 3 sessions cette semaine (lundi pre_class, mercredi evening_first, vendredi daily). Le parent a `missed_session_enabled = true`. |
| **WHEN** | Le scheduler parent prépare les notifications de session manquée. |
| **THEN** | Le parent ne reçoit **PAS** 3 notifications individuelles. À la place : — **Première session manquée de la semaine** : notification push individuelle (cf. Z6-AC25). — **Sessions manquées suivantes (2ème et 3ème)** : regroupées dans le digest hebdo sous la section « Sessions manquées cette semaine : 3 ». Pas de notification push supplémentaire. Maximum **1 notification push « session manquée » par semaine** pour le parent. Le détail complet reste visible dans le tableau de bord parent (Z6-AC27). |

> **NOTE :** Un parent qui reçoit 3+ notifications « session manquée » par semaine désactive les alertes. L'alert fatigue est la première cause de désactivation des notifications parent dans les apps éducatives. La règle « 1 push/semaine + résumé dans le digest » maintient le signal sans créer de bruit. Le tableau de bord reste exhaustif pour les parents qui veulent le détail.

### Z6-AC35 — Digest parent hebdo : contenu standardisé avec indicateur qualité

| | |
|---|---|
| **GIVEN** | Le parent est en mode passif. Le scheduler prépare le digest hebdomadaire. L'élève a 2 chapitres actifs : 'Photosynthèse' (15 items, 3 validation_required, 80% OK+SOLID) et 'Densité' (10 items, 0 validation_required, 60% OK+SOLID). |
| **WHEN** | Le digest est généré. |
| **THEN** | Le digest contient **obligatoirement** les sections suivantes, dans cet ordre : — **1. Résumé activité** : « [Prénom] a révisé [X] fois cette semaine, [Y] min au total ». Si aucune activité : « [Prénom] n'a pas révisé cette semaine » (pas de données masquées). — **2. Maîtrise par chapitre** : pour chaque chapitre actif, le % d'items OK+SOLID et le nombre d'items restants (FRAGILE+UNKNOWN). Si des items ont `validation_required = true`, mention « [N] point(s) en vérification — exercices simplifiés en attendant ». — **3. Alertes** (si applicable) : items à risque (FRAGILE + prochain exam < 5j), sessions manquées (résumé, cf. Z6-AC34), inactivité. — **4. Prochaine action** : « Encouragez [Prénom] à [action concrète] ». Ex. « faire le contrôle blanc de Physique-Chimie ce week-end ». — **5. Score dernier contrôle blanc** (si complété cette semaine) : score + `score_confidence` (cf. Z1-AC18). Le digest est lisible en < 30 secondes (max 150 mots hors titres). |

> **NOTE :** Le digest est le touchpoint principal des parents passifs (mode par défaut). Son contenu était sous-spécifié — un vague « couverture, maîtrise, risques » sans format. Ce AC standardise les 5 sections obligatoires et surtout rend visible le statut de qualité des items (section 2). Un parent qui voit « 3 points en vérification » comprend que la maîtrise affichée est provisoire, ce qui évite la fausse confiance.

### Z6-AC36 — Digest parent : signalement capture incomplète (pages OCR échouées)

| | |
|---|---|
| **GIVEN** | L'élève a uploadé 10 pages pour le chapitre 'Photosynthèse'. 2 pages sont en `status = FAILED` ou `items_generation_failed`. Le parent reçoit le digest hebdomadaire. |
| **WHEN** | Le digest est généré pour ce chapitre. |
| **THEN** | La section maîtrise du chapitre inclut une mention : « ⚠ 2 pages sur 10 n'ont pas pu être analysées — le cours est partiellement couvert. [Prénom] peut reprendre les photos pour compléter. ». Si le nombre de pages échouées représente > 30% du total, la mention est promue en alerte (section 3) : « Attention : plus de 30% du cours de [Matière] n'a pas été analysé. Les exercices et le contrôle blanc ne couvrent pas tout le programme. ». L'alerte est envoyée **une seule fois** (pas répétée chaque semaine si l'élève ne corrige pas). |

> **NOTE :** C'est un gap critique identifié : le parent ne savait pas que la capture était incomplète. Il voyait « 80% de maîtrise sur Photosynthèse » sans savoir que 20% du contenu manquait. Le contrôle blanc sur contenu incomplet donne un score trompeur. Ce AC ferme la boucle entre « problème pipeline » et « parent informé ».

### Z6-AC37 — Feedback après résolution d'une ValidationTask admin

| | |
|---|---|
| **GIVEN** | Un item avait `validation_required = true`. L'élève a cliqué « Je ne sais pas » (Z3-AC04). La ValidationTask a été résolue par l'admin (corrigée ou confirmée). |
| **WHEN** | L'admin résout la tâche. |
| **THEN** | L'élève reçoit une notification in-app (pas push) la prochaine fois qu'il ouvre l'app : « Un point que tu avais signalé a été vérifié : [term] — [action : confirmé / corrigé]. Tes exercices sont mis à jour. ». Si le parent est en mode actif et avait vu l'item dans sa file, le prochain digest mentionne « [N] vérification(s) résolue(s) cette semaine ». Les templates complets sont débloqués pour cet item (cf. Z3-AC01 levé). |

> **NOTE :** Sans ce feedback, les items disparaissent dans une boîte noire. L'élève signale un problème et n'a jamais de retour. Le parent valide des items et ne sait pas si ça a servi. Ce AC ferme la boucle de feedback et renforce la confiance dans le système de qualité.

### Z6-AC38 — Labels de maîtrise traduits pour le parent

| | |
|---|---|
| **GIVEN** | Le digest ou le tableau de bord parent affiche les états de maîtrise des items. |
| **WHEN** | Le parent consulte les données de maîtrise. |
| **THEN** | Les labels techniques sont traduits en langage parent : — UNKNOWN → « Pas encore vu ». — FRAGILE → « En cours d'apprentissage ». — OK → « Compris, à consolider ». — SOLID → « Bien acquis ». — Un item avec `validation_required = true` affiche « En vérification » à la place de son état mastery. Le tooltip (ou sous-texte au premier affichage) explique brièvement ce que signifie chaque niveau. Le % de maîtrise du digest est calculé sur les items OK + SOLID uniquement (FRAGILE et UNKNOWN ne comptent pas comme « maîtrisés »). |

> **NOTE :** Les labels internes (UNKNOWN, FRAGILE, OK, SOLID) sont du jargon développeur. Un parent qui voit « 3 items FRAGILE » peut paniquer (« fragile = mauvais ») alors que ça signifie « en cours d'apprentissage, normal après 1 session ». La traduction en langage naturel et l'explication au premier affichage éliminent cette source de confusion.

### Z6-AC39 — Vue progression globale cross-chapitres

| | |
|---|---|
| **GIVEN** | L'élève a 3 chapitres actifs (HG-INEG avec 12 items, HG-FEOD avec 8 items, PC-TRANSF avec 15 items). Les états de maîtrise sont variés (mix UNKNOWN/FRAGILE/OK/SOLID). |
| **WHEN** | L'élève accède à son tableau de bord principal. |
| **THEN** | Un indicateur de progression globale est affiché : **% maîtrise global** = nombre d'items OK+SOLID / nombre total d'items actifs (non archivés). Chaque chapitre est listé avec sa propre barre de progression visuelle (proportionnelle au nombre d'items). Les chapitres avec un exam à venir dans les 7 jours sont marqués visuellement (badge ou couleur). Un message d'encouragement contextuel est affiché basé sur la tendance (ex: « +12% cette semaine, continue ! » ou « Tu reprends bien après ta pause »). Les items avec `validation_required = true` sont comptés dans le total mais marqués visuellement comme « en vérification ». |

> **NOTE :** Un élève de 13 ans a besoin de voir sa progression globale, pas juste chapitre par chapitre. Sans cet indicateur, l'élève qui a 3 chapitres en cours ne perçoit pas qu'il progresse (« j'ai SOLID sur 2 trucs en HG mais je sais pas où j'en suis au total »). La barre de progression visuelle + le message contextuel exploitent le biais d'engagement de la progression : un % qui monte motive à continuer.

### Z6-AC40 — Suppression (archivage) d'un chapitre par l'élève

| | |
|---|---|
| **GIVEN** | L'élève a un chapitre 'La société féodale' avec 8 items (3 SOLID, 3 OK, 2 FRAGILE) et un exam passé le 3 mars (status = `past`). |
| **WHEN** | L'élève demande à « supprimer » ce chapitre. |
| **THEN** | Le chapitre passe en `archived = true` (soft delete, jamais de suppression physique). Tous les items associés passent en `archived = true`. Les sessions en cours incluant ce chapitre sont recalculées sans ses items (les questions déjà répondues sont conservées). Le chapitre n'apparaît plus dans le tableau de bord ni dans la composition de sessions. L'historique de maîtrise et les Attempts sont conservés (consultables dans un onglet « Archives »). Les exams liés au chapitre ne déclenchent plus de notifications. Une confirmation est demandée avant l'archivage : « Tu veux archiver ce chapitre ? Tes progrès seront conservés et tu pourras le réactiver plus tard. » Le parent est informé dans le prochain digest (« Chapitre archivé : La société féodale »). |

> **NOTE :** Après un contrôle, l'élève veut « faire le ménage ». Sans mécanisme de suppression, les vieux chapitres encombrent le dashboard et continuent d'injecter des items dans les sessions (même post-exam via Z6-AC31, les items ne disparaissent pas tous). Le soft delete (archivage) est préférable à une suppression physique : l'élève peut réactiver en cas d'erreur, et les données de maîtrise sont préservées pour les analytics parent.

### Z6-AC41 — Résilience réseau : persistance optimiste des réponses en session

| | |
|---|---|
| **GIVEN** | L'élève est en session et répond à une question. La connexion réseau est instable ou momentanément perdue. |
| **WHEN** | La réponse est soumise par le client. |
| **THEN** | La réponse est immédiatement persistée localement (storage client) avec un statut `pending_sync`. Le feedback de correction est affiché instantanément (calcul client pour MCQ/NUMERIC, grading local basé sur `expected_answer`). La session continue sans attendre la confirmation serveur. Un indicateur discret « synchronisation en cours… » est visible si la connexion est perdue. Dès que la connexion revient, les Attempts `pending_sync` sont envoyés au serveur en FIFO. En cas de conflit (Attempt déjà existant côté serveur pour la même question), le client gagne (last-write-wins sur le même `question_id + user_id`). Si la synchronisation échoue après 3 retries espacés (5s, 15s, 45s), l'Attempt reste `pending_sync` et un message s'affiche : « Certaines réponses n'ont pas pu être enregistrées. Elles seront synchronisées à la prochaine connexion. » Les Mastery updates côté serveur sont appliqués **uniquement** à la réception des Attempts synchronisés (pas de mise à jour optimiste du mastery, seul le feedback est optimiste). |

> **NOTE :** Un collégien utilise l'app en transport en commun, dans sa chambre avec du wifi instable, ou en zone blanche. Sans persistance optimiste, une déconnexion de 10 secondes = réponse perdue + l'élève doit recommencer = frustration maximale → fermeture de l'app. Le feedback optimiste local permet une UX fluide. Le mastery serveur reste cohérent car il n'est mis à jour qu'à la synchro confirmée.

---

> Ces 121 AC couvrent les zones à risque identifiées pour le vibe coding. Ils sont conçus pour être directement transformés en tests (Jest / Pytest / Playwright). Chaque session de génération de code doit recevoir les AC de la zone concernée comme contexte système, avec l'instruction explicite de générer les tests correspondants avant le code d'implémentation (TDD-first).

*Fin du document — Révise Mieux AC v1.4 · 7 mars 2026*
