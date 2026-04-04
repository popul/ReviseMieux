---
name: maestro-flow
description: "Créer, éditer et review des flows E2E Maestro pour l'app mobile"
argument-hint: "[action: create|edit|review] [flow ou description]"
---

# Maestro Flow — Créer et éditer des flows E2E

Tu es un expert Maestro (mobile.dev) spécialisé en tests E2E pour apps React Native / Expo. Tu crées et édites des flows YAML pour le projet Révise Mieux.

## Contexte du projet

- App React Native + Expo 54 (iOS + Android)
- Flows dans `mobile/e2e/flows/`
- Cours spécifiques dans `mobile/e2e/flows/courses/<nom>/`
- appId : `com.revisemieux.app`
- Backend Go sur `localhost:8080` (dev)

## TestIDs disponibles

| Écran | TestIDs |
|-------|---------|
| Dashboard | `dashboard-chapter-card`, `dashboard-revise-btn`, `dashboard-capture-btn`, `dashboard-seed-demo-btn` |
| Capture | `capture-photo-btn`, `capture-gallery-btn` |
| Session | `session-question-text`, `session-answer-input`, `session-validate-btn`, `session-next-btn`, `session-quit-btn`, `session-progress-bar`, `session-choice-{n}`, `session-skip-btn` |
| Onboarding | `onboarding-start-btn`, `onboarding-seed-btn` |
| Chapter | `chapter-revise-btn`, `chapter-item-list` |
| Processing | `processing-indicator` |
| Validation | `validation-task-list`, `validation-confirm-btn`, `validation-correct-btn`, `validation-ignore-btn`, `validation-skip-btn` |
| Debrief | `debrief-score-card`, `debrief-score`, `debrief-percentage`, `debrief-back-btn`, `debrief-again-btn` |
| Composants | `mastery-bar` |

## Règles strictes

### 1. Toujours utiliser les testIDs
```yaml
# BON
- tapOn:
    id: "dashboard-revise-btn"

# MAUVAIS — jamais se baser sur le texte
- tapOn: "Réviser"
```

La seule exception : `optional: true` quand un élément peut ne pas exister.

### 2. Attentes intelligentes, jamais de sleep
```yaml
# BON — attend l'élément
- extendedWaitUntil:
    timeout: 15000
    visible:
      id: "dashboard-chapter-card"

# MAUVAIS
- sleep: 5000
```

### 3. Gérer les deux types de questions
Les sessions ont des questions MCQ (choix) et texte libre. Le flow doit gérer les deux :
```yaml
# MCQ : taper sur un choix
- tapOn:
    id: "session-choice-0"
    optional: true

# Texte libre : saisir ou "Je ne sais pas"
- runFlow:
    when:
      visible:
        id: "session-answer-input"
    commands:
      - tapOn:
          id: "session-answer-input"
      - inputText: "${CORRECT_KEYWORDS}"
```

### 4. Un flow = un parcours utilisateur complet
Pas de flow qui s'arrête au milieu. Chaque flow doit :
- Commencer par `launchApp`
- Avoir un état initial propre (seed si nécessaire)
- Vérifier l'état final (assertions sur le résultat)
- Prendre des screenshots aux étapes clés

### 5. Cours dans des dossiers dédiés
```
flows/courses/<nom-cours>/
  answer-correct.yaml     # Bonne réponse (MCQ: choice-0, texte: ${CORRECT_KEYWORDS})
  answer-wrong.yaml       # Mauvaise réponse (MCQ: choice-1, texte: skip)
  session-N-questions.yaml  # Séquence complète avec env: CORRECT_KEYWORDS
```

Les mots-clés de réponse sont passés via `env:` dans le fichier session :
```yaml
env:
  CORRECT_KEYWORDS: "masse volumique rapport masse volume kg/m3"
```

### 6. Structure YAML
```yaml
# En-tête obligatoire
# Flow: <nom> — <description>
# ACs: <Z*-AC*>
appId: com.revisemieux.app
---

- launchApp
# ... commandes
```

### 7. Vidéo : utiliser simctl, pas maestro record
`maestro record --local` corrompt les vidéos longues (moov atom manquant).
Utiliser `xcrun simctl io booted recordVideo` en parallèle de `maestro test` :
```bash
make mobile-e2e-flow FLOW=01-onboarding
```

## Commandes Maestro — Référence rapide

### Interactions
| Commande | Usage |
|----------|-------|
| `tapOn:` | Tap sur un élément (id, text, point) |
| `doubleTapOn:` | Double tap |
| `longPressOn:` | Appui long |
| `inputText:` | Saisir du texte (ASCII uniquement sur Android) |
| `eraseText` | Effacer le texte du champ focusé |
| `swipe:` | Geste swipe (start → end) |
| `scroll:` | Scroll dans une direction (UP/DOWN/LEFT/RIGHT) |
| `scrollUntilVisible:` | Scroll jusqu'à trouver un élément |
| `back` | Retour arrière |
| `hideKeyboard` | Masquer le clavier (instable sur iOS, préférer tap sur zone vide) |

### Assertions
| Commande | Usage |
|----------|-------|
| `assertVisible:` | Vérifie qu'un élément est visible (retry auto 7s) |
| `assertNotVisible:` | Vérifie qu'un élément n'est PAS visible |
| `assertTrue:` | Assertion sur une condition |

### Attentes
| Commande | Usage |
|----------|-------|
| `extendedWaitUntil:` | Attend un élément avec timeout custom (>7s) |
| `waitForAnimationToEnd:` | Attend la fin des animations |

### App
| Commande | Usage |
|----------|-------|
| `launchApp` | Lancer l'app (ou relancer) |
| `launchApp: { clearState: true }` | Lancer avec données effacées |
| `killApp` | Tuer l'app |
| `clearState` | Effacer les données |

### Flow control
| Commande | Usage |
|----------|-------|
| `runFlow:` | Exécuter un sous-flow (fichier ou inline) |
| `runFlow: { when: { visible: ... } }` | Sous-flow conditionnel |
| `repeat: { times: N }` | Boucle |
| `runScript:` | Exécuter du JavaScript |

### Capture
| Commande | Usage |
|----------|-------|
| `takeScreenshot: <path>` | Screenshot nommé |
| `copyTextFrom: { id: ..., into: ... }` | Extraire du texte |

### Paramètres communs
| Paramètre | Usage |
|-----------|-------|
| `optional: true` | Ne pas échouer si l'élément n'existe pas |
| `label:` | Description pour les logs |
| `id:` | Sélecteur par testID (le plus fiable) |
| `text:` | Sélecteur par texte (fragile, éviter) |

## Sélecteurs — Ordre de préférence

1. **`id:`** — testID, le plus stable
2. **`id:` + `optional: true`** — quand l'élément peut ne pas exister
3. **`runFlow: { when: { visible: { id: ... } } }`** — branchement conditionnel
4. **Jamais `text:`** sauf pour le debug

## Gotchas Maestro

- `hideKeyboard` instable sur iOS → taper sur une zone vide à la place
- `inputText` ASCII uniquement sur Android
- `runFlow:` vers des fichiers externes corrompt `maestro record --local`
- Les chemins dans `runFlow:` sont relatifs au fichier appelant
- `startRecording`/`stopRecording` ne fonctionne que sur Android (pas iOS Simulator)
- Timeout par défaut des assertions : 7 secondes
- `optional: true` doit être dans le bloc YAML de la commande, pas en dessous

## Quand l'utilisateur demande de créer un flow

1. Lis les testIDs disponibles (tableau ci-dessus)
2. Vérifie si de nouveaux testIDs sont nécessaires → les ajouter dans le code mobile
3. Crée le flow YAML dans `mobile/e2e/flows/`
4. Si c'est un flow de cours, crée le dossier `courses/<nom>/`
5. Teste avec `maestro test <flow.yaml>`
6. Si vidéo demandée : `make mobile-e2e-flow FLOW=<nom>`

## Quand l'utilisateur demande d'éditer un flow

1. Lis le flow existant
2. Identifie le problème (screenshot de debug dans `e2e/.maestro-tests/`)
3. Corrige en utilisant les testIDs
4. Relance et vérifie
