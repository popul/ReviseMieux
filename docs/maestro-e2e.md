# Tests E2E Maestro — Guide pratique

## Setup complet (de zéro à un test qui tourne)

### 1. Installer Maestro CLI

```bash
curl -Ls "https://get.maestro.mobile.dev" | bash
```

Vérifier :

```bash
maestro --version
```

### 2. Démarrer l'infra (PostgreSQL)

```bash
# Depuis la racine du monorepo
make infra-up

# Si doute sur l'état de la DB → repart de zéro (destroy volumes + recrée)
make infra-reset
```

### 3. Lancer le backend

```bash
# Terminal 1 (depuis la racine)
make backend-dev
```

Le backend applique automatiquement les migrations au démarrage (avec tracking `schema_migrations` — les migrations déjà appliquées sont skippées). Il expose `GET /dev/token` en mode dev pour l'auto-login.

### 4. Builder et lancer l'app sur le simulateur

```bash
# Terminal 2 (depuis la racine)
make mobile-sim-start
```

L'app se build, s'installe sur le simulateur iOS et se lance. Au premier démarrage, elle récupère automatiquement un dev token (`GET /dev/token`) — aucune action manuelle.

### 5. Lancer les tests

```bash
# Terminal 3 (depuis la racine, attendre que l'app soit visible sur le simulateur)
make mobile-e2e-flow FLOW=01-onboarding   # un flow avec vidéo + panneau scénario
# ou
make mobile-e2e                                   # tous les flows sans vidéo
```

### Résumé express (copier-coller)

```bash
# Terminal 1 — infra + backend
make infra-reset
make backend-dev

# Terminal 2 — mobile
make mobile-sim-start

# Terminal 3 — tests (attendre que l'app soit visible sur le simulateur)
make mobile-e2e-flow FLOW=01-onboarding
```

### Ce qui se passe automatiquement

| Étape | Mécanisme | Détail |
|-------|-----------|--------|
| Reset DB | `POST /e2e/seed/onboarding` | Appelé par `runScript` en début de flow. Truncate les tables `user_data`, préserve les tables `reference` (templates). Recrée le user dev + chapitre démo (8 items UNKNOWN). |
| Auth | `GET /dev/token` | L'app appelle ce endpoint au lancement en mode `__DEV__`. Crée un user `Hugo (dev)` avec un JWT 24h. |
| Données démo | `POST /onboarding/seed-demo` | Le flow Maestro tape sur le bouton "Charger le chapitre démo" (`dashboard-seed-demo-btn`). Crée 8 items, 3 notions, 1 chapitre "Densité et masse volumique". Idempotent. |
| Scoring texte | LLM (Gemini) | Les réponses texte libre sont scorées par le LLM. Nécessite une clé API configurée (`GEMINI_API_KEY` ou `ANTHROPIC_API_KEY`). |

### Catégorisation des tables (reference vs user_data)

Chaque table PostgreSQL est taggée via `COMMENT ON TABLE` (migration `006_table_categories.sql`) :

| Catégorie | Description | Exemples | Truncatée au reset E2E ? |
|-----------|-------------|----------|--------------------------|
| `reference` | Données seedées par les migrations, jamais modifiées au runtime | `templates`, `template_variables` | Non |
| `user_data` | Données créées par l'application au runtime (utilisateurs, sessions, etc.) | `users`, `chapters`, `items`, `masteries`, `sessions` | Oui |

Cette metadata est la source de vérité unique — le code Go (endpoint E2E et tests d'intégration) la lit dynamiquement, sans liste hardcodée. Pour ajouter une nouvelle table, il suffit d'ajouter un `COMMENT ON TABLE xxx IS 'user_data'` (ou `'reference'`) dans la migration qui la crée.

### Variables d'environnement du backend

Le backend a besoin au minimum de :

```bash
DATABASE_URL=postgres://revisemieux:revisemieux@localhost:5432/revisemieux?sslmode=disable
JWT_SECRET=dev-secret           # n'importe quelle valeur en dev
GIN_MODE=debug                  # active /dev/token
```

Ces valeurs sont les défauts de `docker-compose.yml` — rien à configurer si on utilise `make setup`.

---

## Commandes rapides

Depuis la racine du monorepo :

```bash
# Lancer UN flow (sans vidéo)
make mobile-e2e-flow FLOW=01-onboarding

# Lancer tous les flows
make mobile-e2e

# Lancer UN flow avec vidéo + panneau scénario
make mobile-e2e-record-flow FLOW=01-onboarding

# Enregistrer tous les flows avec vidéo
make mobile-e2e-record
```

Les rapports (screenshots, logs) sont dans `mobile/e2e/reports/`. Les vidéos dans `mobile/e2e/videos/`.

---

## Enregistrement vidéo

La vidéo n'est générée que sur demande explicite via les commandes `record`.

### `maestro record` — vidéo avec panneau scénario

Génère une vidéo avec le **panneau scénario à gauche** (étapes Maestro) et l'écran du simulateur à droite. **Limitée à 2 minutes** — les timeouts des flows sont calibrés pour rester sous cette limite.

```bash
make mobile-e2e-record-flow FLOW=01-onboarding
# -> e2e/videos/01-onboarding.mp4
```

---

## Structure des fichiers E2E

```
mobile/e2e/
├── config/
│   └── maestro-config.yaml     # Config globale (appId, délais)
├── flows/
│   ├── 01-onboarding.yaml      # Parcours complet premier lancement
│   ├── 02-capture.yaml         # Upload photo + pipeline
│   ├── 03-dashboard.yaml       # Navigation dashboard
│   ├── 04-session-complete.yaml # Session de révision
│   ├── 05-validation-hitl.yaml # Validation humaine
│   ├── lib/                    # Sous-flows réutilisables
│   │   ├── answer-choice.yaml  # Répondre à un QCM
│   │   ├── answer-text.yaml    # Répondre en texte libre
│   │   └── answer-skip.yaml    # "Je ne sais pas"
│   └── courses/
│       └── demo-densite/
│           └── session-8-all-correct.yaml  # Réponses du cours démo
├── edge-cases/
│   └── EC-09-tous-solid.yaml   # Tous items SOLID
├── record-flow.sh              # Script d'enregistrement standalone
└── videos/                     # Vidéos générées (gitignored)
```

---

## Anatomie d'un flow

Chaque flow est un fichier YAML. Exemple simplifié :

```yaml
appId: com.revisemieux.app
---

- launchApp

# Attendre qu'un élément soit visible
- extendedWaitUntil:
    timeout: 10000
    visible:
      id: "dashboard-capture-btn"

# Tapper sur un élément
- tapOn:
    id: "dashboard-chapter-card"

# Vérifier la présence d'un élément
- assertVisible:
    id: "mastery-bar"

# Exécuter un sous-flow avec paramètres
- runFlow:
    file: lib/answer-choice.yaml
    env:
      CHOICE_INDEX: "0"
```

### Conventions de nommage des testID

Les flows utilisent la prop `testID` des composants React Native. Convention : `{écran}-{élément}[-{variante}]`.

Exemples : `dashboard-chapter-card`, `session-question-text`, `session-choice-0`, `debrief-score-card`.

---

## Sous-flows réutilisables (`lib/`)

### `answer-choice.yaml` — Répondre à un QCM

```yaml
- runFlow:
    file: lib/answer-choice.yaml
    env:
      CHOICE_INDEX: "0"    # 0 = premier choix
```

Séquence : attend la question → tape le choix → valide → attend le feedback → tape "suivant".

### `answer-text.yaml` — Répondre en texte libre

```yaml
- runFlow:
    file: lib/answer-text.yaml
    env:
      ANSWER_TEXT: "masse volumique eau 1000 kg/m3"
```

Séquence : attend la question → tape dans l'input → saisit le texte → valide → attend le feedback → tape "suivant".

### `answer-skip.yaml` — Passer une question

```yaml
- runFlow:
    file: lib/answer-skip.yaml
```

Séquence : attend la question → tape "je ne sais pas" → attend le feedback → tape "suivant".

---

## Flow onboarding en détail

Le flow `01-onboarding.yaml` teste le parcours complet d'un premier lancement :

| Étape | Actions | Vérifications |
|-------|---------|---------------|
| Dashboard | Seed cours démo | Carte chapitre visible, barre mastery |
| Carte leçon | Tap sur la carte | Liste des items, bouton "Réviser" |
| Session | 8 questions (via sous-flow) | Barre de progression |
| Débrief | Lecture du score | Score >= 87%, boutons retour/rejouer |
| Retour | Navigation arrière | Dashboard avec mastery mis à jour |

Le sous-flow `courses/demo-densite/session-8-all-correct.yaml` enchaîne 8 réponses correctes (3 QCM + 5 texte libre) sur le cours "Densité et masse volumique".

---

## Créer un nouveau flow

1. Créer `mobile/e2e/flows/XX-nom-du-flow.yaml`
2. Commencer par `appId: com.revisemieux.app` + séparateur `---`
3. Utiliser les sous-flows `lib/` pour les interactions récurrentes
4. Documenter les ACs couvertes en commentaire d'en-tête
5. Tester : `make e2e-flow FLOW=XX-nom-du-flow`

---

## Troubleshooting

| Problème | Solution |
|----------|----------|
| `maestro: command not found` | Vérifier que `~/.maestro/bin` est dans le PATH, ou utiliser le chemin complet |
| `No booted device` | Lancer le simulateur : `open -a Simulator` puis `make sim-start` |
| Vidéo vide ou non créée | Vérifier qu'un seul simulateur est "booted" : `xcrun simctl list devices booted` |
| Timeout sur `extendedWaitUntil` | L'app est peut-être lente au démarrage — augmenter le timeout ou vérifier que l'app tourne |
| `maestro record` coupe à 2 min | Utiliser `make e2e-flow` (simctl) à la place, pas de limite de durée |
| Flow échoue sur `tapOn` | Vérifier que le `testID` existe dans le composant React Native |
