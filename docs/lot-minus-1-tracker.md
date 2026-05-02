# Suivi Lot -1 — Révise Mieux (web-v0)

| | |
|---|---|
| **Périmètre** | 15 ACs (Z0) · site web one-page autonome · pas de DB, pas de comptes |
| **Démarrage** | 2026-04-28 (premier déploiement homelab) |
| **Code** | [`web-v0/`](../web-v0/) · [`prompts/study-guide/`](../prompts/study-guide/) |
| **Production** | `https://revise-lab.musso.io/` (homelab K3s + Cloudflare) |
| **Image** | `ghcr.io/popul/revisemieux-web-v0` (publique) |

---

## Légende statut

| Icône | Statut |
|---|---|
| `[ ]` | À faire |
| `[~]` | En cours |
| `[x]` | Terminé |
| `[—]` | Bloqué / en attente |

---

## Phase 0 — Shell & schéma

Précondition de toute le reste : un chrome HTML stable et déterministe.

| # | Tâche | Statut | ACs couvertes | Livrables |
|---|---|---|---|---|
| 0.1 | Shell Jinja2 mobile-first avec drawer hamburger fonctionnel | `[x]` | Z0-AC04, Z0-AC07 | `web-v0/shell.html.j2` |
| 0.2 | Deux boutons d'impression (élève / parent) avec safe-area iOS | `[x]` | Z0-AC05, Z0-AC06 | `web-v0/shell.html.j2` |
| 0.3 | Schéma Pydantic `Fiche` couvrant Sections 1-6 + 8 + 9 | `[x]` | Z0-AC11 | `web-v0/schema.py` |
| 0.4 | Endpoint `/preview` qui rend une fixture sans appel LLM | `[x]` | Z0-AC12 | `web-v0/main.py`, `web-v0/fixture.json` |
| 0.5 | `@media print` complet : élève masque corrigés, parent imprime tout | `[x]` | Z0-AC06 | `web-v0/shell.html.j2` |
| 0.6 | Theme + breakpoints responsive (768/480/landscape) | `[x]` | Z0-AC07 | `web-v0/shell.html.j2` |

> **Critère de validation Phase 0 :** `GET /preview` renvoie 200 + ~46 kB de HTML autonome qui exerce toutes les sections du template ; chrome (drawer + boutons impression) testé manuellement sur Safari iOS et Chrome desktop.

---

## Phase 1 — Pipeline 1-stage (legacy, conservé)

Première itération : un seul appel LLM, sortie HTML directe. Permet de livrer immédiatement avant de basculer en 2-stage.

| # | Tâche | Statut | ACs couvertes | Livrables |
|---|---|---|---|---|
| 1.1 | Job pattern async (POST /jobs immédiat + polling) | `[x]` | Z0-AC01, Z0-AC09 | `web-v0/main.py` |
| 1.2 | Appel LLM Qwen 3.6 35B-A3B (vision + thinking) via LM Studio | `[x]` | Z0-AC01 | `web-v0/main.py` `_run_llm` |
| 1.3 | Logs structurés par job (created/llm_request/llm_response/html) | `[x]` | Z0-AC10 | `web-v0/main.py` |
| 1.4 | Strip `<think>...</think>` et bloc ```html en sortie | `[x]` | Z0-AC02 | `web-v0/main.py` |
| 1.5 | Déploiement K3s via image GHCR publique | `[x]` | Z0-AC15 | `homelab` repo (`services/revise/`) |
| 1.6 | Source unique du prompt (`prompts/study-guide/system.md`) | `[x]` | — | PR factorisation |

> **Critère de validation Phase 1 :** une fiche cible 18-20 sur 12 photos est générée en < 5 min, présente Sections 1-6 + 8 + 9, finit par `</html>`, taille HTML > 30 kB. **Atteint** sur le run `cff6af4d-...` du 2026-04-29 (84 kB, finish=stop).

---

## Phase 2 — Pipeline 2-stage JSON (en cours)

Bascule du LLM en producteur de JSON validé Pydantic, rendu HTML déterministe via Jinja. Élimine les régressions chrome qu'on observait avec le 1-stage.

| # | Tâche | Statut | ACs couvertes | Livrables |
|---|---|---|---|---|
| 2.1 | Schéma Fiche complet (Pydantic v2, alias `def`/keys, énums Mastery) | `[x]` | Z0-AC11 | `web-v0/schema.py` |
| 2.2 | Réécriture du prompt système en mode extraction JSON | `[ ]` | Z0-AC02 | `prompts/study-guide/system.md` (refonte) |
| 2.3 | Activation `response_format: {"type":"json_object"}` côté LM Studio | `[ ]` | Z0-AC11 | `web-v0/main.py` |
| 2.4 | Validation Pydantic stricte ; échec → job=error + message | `[ ]` | Z0-AC11 | `web-v0/main.py` |
| 2.5 | Render Jinja `shell.html.j2` à partir du `Fiche` validé | `[~]` | Z0-AC02, Z0-AC04, Z0-AC05 | `web-v0/main.py` |
| 2.6 | Bench comparatif latence + qualité 1-stage vs 2-stage | `[ ]` | Z0-AC08 | rapport `docs/lot-minus-1-bench.md` |
| 2.7 | Bascule par défaut sur 2-stage, suppression du chemin 1-stage | `[ ]` | — | `web-v0/main.py` |

> **Critère de validation Phase 2 :** sur les mêmes 12 photos qu'en Phase 1, la fiche 2-stage couvre tous les Z0-AC02/03 sans avoir à corriger le HTML, et le shell ne dépend plus du tout de la sortie LLM (qui ne touche plus au HTML).

---

## Phase 3 — Itération AC + ouverture publique

Une fois le pipeline 2-stage stable, durcir les Z0-AC restants et ouvrir l'accès au-delà du dev.

| # | Tâche | Statut | ACs couvertes | Livrables |
|---|---|---|---|---|
| 3.1 | Suite de tests d'intégration sur 4 chapitres pilotes (HG/SVT/PC/français) | `[ ]` | Z0-AC02, Z0-AC03 | `web-v0/tests/` |
| 3.2 | Audit privacy : aucun log/disque ne contient les bytes des images | `[ ]` | Z0-AC13 | revue manuelle + checklist |
| 3.3 | Cloudflare Access (zero-trust) pour POST /jobs (whitelist e-mail) | `[ ]` | Z0-AC14 | config Cloudflare |
| 3.4 | Rate-limit basique côté FastAPI (slowapi ou middleware) | `[ ]` | Z0-AC14 | `web-v0/main.py` |
| 3.5 | Auto-continuation si `finish_reason=length` (rare avec 2-stage) | `[ ]` | Z0-AC08 | `web-v0/main.py` |
| 3.6 | Bench fréquence d'usage Louis (≥ 1 contrôle / semaine) | `[ ]` | KPI adoption | log analysis |

> **Critère de validation Phase 3 :** Louis génère ≥ 5 fiches sur 4 chapitres distincts, taux de fiche utilisable au premier shot ≥ 90%, latence P95 ≤ 5 min, aucun incident privacy.

---

## Roadmap synthétique

```
2026-04-28  ─────  Phase 0  ✓  shell + /preview en ligne
2026-04-29  ─────  Phase 1  ✓  pipeline 1-stage HTML
2026-05-02  ─────  factor : prompt en source unique, image GHCR
2026-05-??  ─────  Phase 2     pipeline 2-stage JSON
2026-05-??  ─────  Phase 3     ouverture famille + bench
```

> Les jalons Phase 2/3 ne sont pas datés : Lot -1 reste un side track de Lot 0 (qui prime). On itère sur Lot -1 quand un usage réel le demande (un contrôle de Louis, un bug remonté, une amélioration triviale).
