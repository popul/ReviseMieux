# Edge cases -- Revise Mieux

> Catalogue exhaustif des edge cases classes par domaine. Chaque entree couvre : risque UX, traitement ideal, copy/guidance, et chemin de recuperation.

---

## 1. Capture photo

### EC-01 : Photo floue

| | |
|---|---|
| **Risque UX** | L'eleve ne se rend pas compte que la photo est mauvaise. Le pipeline echoue plus tard, frustration differee. |
| **Traitement ideal** | Detection cote client avant upload : si le blur score depasse un seuil, message d'avertissement AVANT envoi. |
| **Copy** | "Cette photo semble un peu floue. Reprends-la avec un meilleur eclairage pour de meilleurs resultats." |
| **Recovery** | L'eleve peut reprendre la photo ou continuer malgre tout. Le pipeline traite quand meme avec confidence reduite (Z2-AC03). |
| **AC** | Z2-AC03 |

### EC-02 : Photo de travers / mal cadree

| | |
|---|---|
| **Risque UX** | Texte tronque, OCR partiel. L'eleve ne voit pas le probleme. |
| **Traitement ideal** | Proposition de recadrage automatique (crop suggestion) avant validation de la page. |
| **Copy** | "La page semble coupee sur le bord droit. Veux-tu recadrer ou reprendre ?" |
| **Recovery** | Option recadrage OU reprendre la photo OU ignorer et continuer. |
| **AC** | -- (non couvert par AC existant, recommandation nouvelle) |

### EC-03 : Plusieurs pages en une seule photo

| | |
|---|---|
| **Risque UX** | Double page = texte trop petit, OCR degrade. L'eleve pense avoir fait 2 pages en 1. |
| **Traitement ideal** | Detection (ratio hauteur/largeur atypique) + avertissement. |
| **Copy** | "On dirait que tu as photographie 2 pages en meme temps. Pour de meilleurs resultats, photographie une page a la fois." |
| **Recovery** | Reprendre en 2 photos separees OU continuer (traitement best-effort). |
| **AC** | -- (recommandation nouvelle) |

### EC-04 : Pas de connexion pendant la capture

| | |
|---|---|
| **Risque UX** | L'eleve prend ses photos dans le bus sans wifi. L'upload echoue. |
| **Traitement ideal** | Stockage local des photos. Upload automatique au retour de la connexion. |
| **Copy** | "Photos enregistrees ! Elles seront analysees des que tu auras du reseau." |
| **Recovery** | Indicateur "X photos en attente d'envoi" sur le dashboard. Upload en background. |
| **AC** | Z6-AC31 (resilience reseau) |

### EC-05 : Pipeline J0 timeout ou echec complet

| | |
|---|---|
| **Risque UX** | Premier upload = moment de verite. Un echec ici = "ca marche pas" = desinstallation. |
| **Traitement ideal** | Ecran de recovery empathique (Z8-AC03) avec conseils visuels + fallback chapitre demo. |
| **Copy** | "Les photos sont un peu difficiles a lire. Pas de panique, ca arrive souvent au debut !" |
| **Recovery** | 3 chemins : reprendre les photos / essayer quand meme (si items partiels) / chapitre demo. |
| **AC** | Z8-AC03, Z2-AC07 |

### EC-06 : Photo contenant du contenu de plusieurs chapitres

| | |
|---|---|
| **Risque UX** | L'eleve a deux cours sur la meme double page. Le pipeline attribue tout a un seul chapitre. |
| **Traitement ideal** | Detection LLM (changement de sujet dans le texte) + question de clarification. |
| **Copy** | "On a detecte du contenu sur 2 sujets differents. A quel chapitre ces notes appartiennent-elles ?" |
| **Recovery** | Choix du chapitre OU creation automatique de 2 chapitres distincts. |
| **AC** | -- (non couvert, recommandation future) |

---

## 2. Session de revision

### EC-07 : L'eleve ferme l'app en pleine session

| | |
|---|---|
| **Risque UX** | Perte de progression. L'eleve ne sait pas si ses reponses ont ete enregistrees. |
| **Traitement ideal** | Persistance du current_question_index en base. Reprise automatique a la reouverture. |
| **Copy** | (a la reouverture) "Tu avais une session en cours. On reprend ou tu en etais ?" |
| **Recovery** | Reprise a la question N+1. Les masteries des questions 1-N sont deja mises a jour. TTL 72h. |
| **AC** | Z4-AC04 |

### EC-08 : Pas assez d'items pour composer une session

| | |
|---|---|
| **Risque UX** | L'eleve appuie sur "Lancer une session" et... rien. Frustration. |
| **Traitement ideal** | Session viable sur petit chapitre (min 4 questions meme avec 3 items). Reformulations. |
| **Copy** | "Petite session rapide -- [N] questions sur ce chapitre" |
| **Recovery** | Session courte 3-5 min avec reformulations (meme item, gabarits/distractors differents). |
| **AC** | Z4-AC11 |

### EC-09 : Tous les items sont SOLID

| | |
|---|---|
| **Risque UX** | L'eleve ouvre l'app, rien a faire. "Cette app ne sert plus a rien." |
| **Traitement ideal** | Message positif + session de consolidation optionnelle sans risque de regression. |
| **Copy** | "Bravo -- tu es a jour sur tous tes chapitres ! Prochaine revision prevue [date]. Session de consolidation ?" |
| **Recovery** | Bouton optionnel "Session de consolidation" (5 questions, aucune regression possible). |
| **AC** | Z4-AC13 |

### EC-10 : Reponse ambigue (scoring au seuil de 0.7)

| | |
|---|---|
| **Risque UX** | L'eleve pense avoir bien repondu, le systeme dit "echec". Sentiment d'injustice. |
| **Traitement ideal** | Scoring graduel (Z1-AC23) avec feedback nuance. Score partiel visible. |
| **Copy** | "Reponse partiellement correcte. Il te manquait [element]. Score : 0.6 -- pas loin !" |
| **Recovery** | Demi-succes ne penalise pas la maitrise autant qu'un echec total. |
| **AC** | Z1-AC23, Z1-AC24 |

### EC-11 : Quiz trop difficile (3 echecs consecutifs)

| | |
|---|---|
| **Risque UX** | Cascade d'echecs = "je suis nul" = fermeture de l'app. Critique pour Ines (anxieuse). |
| **Traitement ideal** | Descente automatique de difficulte apres 3 echecs. Retour aux gabarits simples. |
| **Copy** | "On revoit les bases ensemble -- c'est normal de ne pas tout retenir du premier coup !" |
| **Recovery** | Difficulte reduite pour les 2 questions suivantes. Renvoi vers la carte de lecon. Pause J+2 apres 5 echecs. |
| **AC** | Z1-AC19 |

### EC-12 : Quiz trop facile (> 92% de reussite)

| | |
|---|---|
| **Risque UX** | L'eleve s'ennuie. "C'est pour les bebes." Critique pour Lucas (desengage). |
| **Traitement ideal** | Montee en difficulte invisible (Z4-AC16). Gabarits plus exigeants sans message. |
| **Copy** | (aucun message -- la montee est silencieuse pour ne pas creer de pression) |
| **Recovery** | Le systeme ajuste automatiquement, l'eleve sent que les questions deviennent "plus interessantes". |
| **AC** | Z4-AC16 |

### EC-13 : Reponses trop rapides (anti-clicking aveugle)

| | |
|---|---|
| **Risque UX** | L'eleve (Lucas) clique au hasard pour finir vite. Progression fictive. |
| **Traitement ideal** | Reponses rapides correctes ne comptent pas pour la progression. Reponses incorrectes penalisent normalement. |
| **Copy** | (apres 3 reponses rapides correctes) "Prends ton temps pour bien lire les questions -- tes reponses rapides ne comptent pas pour ta progression." |
| **Recovery** | L'eleve peut continuer mais les items seront re-proposes. Pas de blocage UX. |
| **AC** | Z3-AC15 |

---

## 3. Validation HITL

### EC-14 : Le parent n'a jamais valide

| | |
|---|---|
| **Risque UX** | Items bloques en gabarits simples indefiniment. Maitrise plafonnee. |
| **Traitement ideal** | Bandeau non-bloquant sur le chapitre. Max 3 tasks/semaine pour le parent. Le chapitre reste utilisable. |
| **Copy** | "6 zones a verifier -- vos exercices seront plus precis apres verification." |
| **Recovery** | L'eleve peut tout ignorer et continuer. Admin en fallback. Le diagnostic fonctionne malgre tout. |
| **AC** | Z3-AC06, Z3-AC08 |

### EC-15 : Item rejete (correction) par le parent

| | |
|---|---|
| **Risque UX** | L'item corrige invalide le cache des questions. L'eleve voit un changement inexplique. |
| **Traitement ideal** | Cache invalidation ciblee (uniquement les questions de l'item corrige). Regeneration lazy. |
| **Copy** | (pas de message visible cote eleve -- la correction est transparente) |
| **Recovery** | Les prochaines questions sur cet item utiliseront le contenu corrige. Transition fluide. |
| **AC** | Z3-AC03, Z3-AC07 |

### EC-16 : Parent ne comprend pas un item

| | |
|---|---|
| **Risque UX** | Le parent confirme un item faux par defaut. Pire cas : l'eleve etudie du contenu incorrect. |
| **Traitement ideal** | Option "Je ne sais pas" qui delegue a l'admin. Retractation possible si signalement ulterieur. |
| **Copy** | Bouton "Je ne sais pas" + "L'item sera verifie par notre equipe." |
| **Recovery** | Admin review. Si l'item a ete confirme par erreur, le signalement eleve (Z3-AC12) ou l'anomalie (Z3-AC13) le detectent. Retractation Z3-AC14. |
| **AC** | Z3-AC04, Z3-AC14 |

### EC-17 : Ado qui clique "Ignorer" sur tout

| | |
|---|---|
| **Risque UX** | Tous les items incertains disparaissent du circuit pedagogique. Lacunes invisibles. |
| **Traitement ideal** | Section "Points non verifies" visible dans la carte lecon. Nudge discret. Bouton "Reactiver". |
| **Copy** | "4 points n'ont pas ete verifies -- ils ne seront pas dans tes exercices." + "Tout laisser ignore" si choix delibere. |
| **Recovery** | L'eleve peut reactiver a tout moment. L'admin traite en backoffice. |
| **AC** | Z3-AC05, Z3-AC16 |

---

## 4. Adoption et engagement

### EC-18 : Eleve abandonne en plein flow

| | |
|---|---|
| **Risque UX** | L'eleve quitte pendant le pipeline, pendant une session, pendant la validation. Etat incoherent. |
| **Traitement ideal** | Chaque etape est interruptible et reprenante. Aucun etat incoherent possible. |
| **Copy** | (a la reouverture) Message contextuel adapte : "Ton cours est en cours d'analyse" / "Tu avais une session en cours" |
| **Recovery** | Pipeline continue en background. Session reprend au dernier point (Z4-AC04). Validation reste en file. |
| **AC** | Z4-AC04, Z8-AC04 |

### EC-19 : Parent ne comprend pas la proposition de valeur

| | |
|---|---|
| **Risque UX** | Le parent voit des metriques incomprehensibles. Desinstalle ou ignore. |
| **Traitement ideal** | Onboarding parent en 3 ecrans (30s). Labels traduits. Digest anticipe a J+2. |
| **Copy** | Ecran 1 : "Voici ce que fait [Prenom]" (schema visuel simple). Labels : "Pas encore vu", "En cours d'apprentissage", "Compris, a consolider", "Bien acquis". |
| **Recovery** | Digest anticipe J+2 avec preuves tangibles. Script 3 min pour engagement minimal. |
| **AC** | Z8-AC05, Z8-AC06, Z6-AC29 |

### EC-20 : L'eleve se sent juge par le scoring

| | |
|---|---|
| **Risque UX** | L'eleve (Ines) interprete les regressions comme un jugement. Anxiete et decrochage. |
| **Traitement ideal** | Growth mindset systematique dans tous les messages. Jamais "echec", toujours "refresh" ou "a revoir". |
| **Copy** | Regression SOLID->OK : "Ce point a besoin d'un petit refresh -- on le revoit bientot !" (jamais "tu as perdu ta maitrise") |
| **Recovery** | Le framing positif est applique partout : feedback, debrief, dashboard. Pas de score negatif visible. |
| **AC** | Z1-AC05, Z1-AC06, Z1-AC26 |

### EC-21 : Aucun contenu exploitable detecte

| | |
|---|---|
| **Risque UX** | L'eleve uploade un dessin, une page blanche, ou du texte dans une langue etrangere. 0 items. |
| **Traitement ideal** | Message empathique + conseils + fallback chapitre demo. |
| **Copy** | "Aucun contenu n'a pu etre extrait -- verifie la qualite des photos et re-uploade." |
| **Recovery** | Bouton "Re-uploader", chapitre demo disponible, chapitre cree mais en etat PENDING_UPLOAD. |
| **AC** | Z2-AC07 |

### EC-22 : Retour apres 5+ jours d'absence

| | |
|---|---|
| **Risque UX** | 15 items en retard. Session de 20 min bourrée d'echecs. Decrochage definitif. |
| **Traitement ideal** | Mode "retour en douceur" : session courte (5 min), gabarits faciles, etalement sur 3-5 jours. |
| **Copy** | "Content de te revoir ! On reprend doucement avec quelques rappels." |
| **Recovery** | 4-6 items les plus anciens seulement. Difficulte 1-2. Desactivation auto apres 2 sessions consecutives. |
| **AC** | Z6-AC22 |

### EC-23 : Retour apres 7+ jours (relance)

| | |
|---|---|
| **Risque UX** | L'eleve a completement decroche. La notification de relance est le dernier point de contact. |
| **Traitement ideal** | 1 seule notification avec deep link vers session courte et facile. Aucune relance supplementaire. |
| **Copy** | "Tes revisions t'attendent -- [N] points a consolider en [matiere]. On reprend doucement ?" |
| **Recovery** | Deep link vers session 3-5 min, gabarits faciles. Si ignore, pas de relance jusqu'au prochain evenement contextuel (exam, upload, trimestre). |
| **AC** | Z6-AC21 |

---

## 5. Connectivite

### EC-24 : Perte de connexion pendant une session

| | |
|---|---|
| **Risque UX** | L'eleve repond a une question, la reponse ne part pas. Perte de progression. |
| **Traitement ideal** | Persistance locale optimiste. Feedback calcule cote client. Sync au retour reseau. |
| **Copy** | Indicateur discret "synchronisation en cours..." + "Certaines reponses n'ont pas pu etre enregistrees. Elles seront synchronisees a la prochaine connexion." |
| **Recovery** | Reponses stockees en pending_sync local. FIFO au retour. Mastery mis a jour uniquement a la synchro serveur. |
| **AC** | Z6-AC31 |

### EC-25 : Perte de connexion pendant un upload

| | |
|---|---|
| **Risque UX** | L'eleve a pris 8 photos et l'upload echoue. Il pense que ses photos sont perdues. |
| **Traitement ideal** | Photos stockees localement. Retry automatique avec indicateur. |
| **Copy** | "Photos enregistrees ! Elles seront envoyees des que tu auras du reseau." |
| **Recovery** | Retry automatique (3 tentatives espacees). Indicateur "X photos en attente". L'eleve peut forcer le retry. |
| **AC** | Z6-AC31 |

### EC-26 : Mode avion prolonge

| | |
|---|---|
| **Risque UX** | L'eleve est dans le train ou en zone blanche. Rien ne fonctionne. |
| **Traitement ideal** | Mode offline partiel : consultation de la carte de lecon (en cache), flashcards statiques. |
| **Copy** | "Pas de connexion -- tu peux relire ta lecon en attendant." |
| **Recovery** | Au retour de la connexion, sync automatique de tout ce qui est pending. |
| **AC** | Z4-AC12 (fallback), Z6-AC31 (resilience) |

---

## 6. Cas limites techniques

### EC-27 : Chapitre avec plus de 30 items

| | |
|---|---|
| **Risque UX** | Sessions trop longues. Notions trop nombreuses. Interface surchargee. |
| **Traitement ideal** | Le pipeline regroupe en 3-7 notions. La session quotidienne reste 10-20 min (selection 70/20/10). |
| **Copy** | Aucun message special -- l'abstraction par notion masque la complexite. |
| **Recovery** | L'algorithme de composition gere naturellement les grands chapitres. |
| **AC** | Z7-AC15 (notions 3-7) |

### EC-28 : Double page identique uploadee

| | |
|---|---|
| **Risque UX** | Items en doublon. Confusion dans la carte de lecon. |
| **Traitement ideal** | Detection de similarite visuelle (> 80%) lors de l'ajout incremental. Avertissement. |
| **Copy** | "Cette page ressemble a la page 2 deja dans ton chapitre. Ajouter quand meme ?" |
| **Recovery** | L'eleve confirme ou annule. Si ajoutee, Z3-AC11 (coherence intra-chapitre) detecte les doublons et archive automatiquement l'item de plus faible confidence. |
| **AC** | Z5-AC11, Z3-AC11 |

### EC-29 : LLM indisponible prolonge (> 5 min)

| | |
|---|---|
| **Risque UX** | Aucune session possible. L'eleve ne peut pas reviser. |
| **Traitement ideal** | Mode "Relecture active" avec flashcards statiques generees a partir des items existants (pas de LLM necessaire). |
| **Copy** | "Les exercices ne sont pas disponibles pour le moment -- en attendant, revise ta lecon avec ces flashcards !" |
| **Recovery** | Auto-check toutes les 30s. Retour au mode normal des que le LLM repond. |
| **AC** | Z4-AC12 |

### EC-30 : Session tardive (apres 21h)

| | |
|---|---|
| **Risque UX** | L'eleve revise trop tard, mauvais sommeil, contre-productif. |
| **Traitement ideal** | Nudge bienveillant + mode express pre-selectionne + reduction progressive. |
| **Copy** | 21h : "Il est un peu tard -- une session express pour l'essentiel, et au dodo !" / 22h : "Il est tard ! Ton cerveau retient mieux quand tu dors bien. On reprend demain ?" |
| **Recovery** | L'eleve peut forcer la session complete si il veut. Aucun blocage. |
| **AC** | Z7-AC25 |
