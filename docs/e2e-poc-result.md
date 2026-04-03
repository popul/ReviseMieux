# PoC E2E Maestro — Resultats

> Date : 2026-04-04
> Issue : #47
> Statut : Structure prete, execution en attente de build Expo

---

## 1. Installation

Maestro 2.4.0 installe avec succes via `curl -Ls "https://get.maestro.mobile.dev" | bash`.

## 2. Structure creee

```
mobile/e2e/
├── .gitignore              # Ignore videos/
├── config/
│   └── maestro-config.yaml # Config globale (appId, delais)
├── flows/
│   ├── 01-onboarding.yaml        # Onboarding → seed demo → premiere session
│   ├── 02-capture.yaml           # Dashboard → capture photo
│   ├── 03-dashboard.yaml         # Verification dashboard post-onboarding
│   ├── 04-session-complete.yaml  # Session complete avec reponses
│   └── 05-validation-hitl.yaml   # Validation parent — lister + confirmer
└── edge-cases/
    └── EC-09-tous-solid.yaml     # Tous items SOLID → message felicitation
```

## 3. TestIDs disponibles (22)

Tous les testIDs necessaires ont ete ajoutes (issue #42) :

| Ecran | TestIDs |
|-------|---------|
| Dashboard | dashboard-chapter-card, dashboard-revise-btn, dashboard-capture-btn |
| Capture | capture-photo-btn, capture-gallery-btn |
| Session | session-question-text, session-answer-input, session-validate-btn, session-next-btn, session-quit-btn, session-progress-bar |
| Onboarding | onboarding-start-btn, onboarding-seed-btn |
| Chapter | chapter-revise-btn, chapter-item-list |
| Processing | processing-indicator |
| Validation | validation-task-list, validation-confirm-btn, validation-correct-btn, validation-ignore-btn, validation-skip-btn |
| Composants | mastery-bar |

## 4. Prochaine etape : build et execution

Pour executer les flows, il faut :

### iOS (simulateur local)
```bash
cd mobile
npx expo prebuild --platform ios
cd ios && xcodebuild -workspace revisemieux.xcworkspace -scheme revisemieux -configuration Debug -destination 'platform=iOS Simulator,name=iPhone 16' build
# Puis
maestro test e2e/flows/01-onboarding.yaml
maestro record e2e/flows/01-onboarding.yaml --output e2e/videos/01-onboarding.mp4
```

### Android (emulateur local)
```bash
cd mobile
npx expo prebuild --platform android
cd android && ./gradlew assembleDebug
# Installer sur emulateur
adb install app/build/outputs/apk/debug/app-debug.apk
# Puis
maestro test e2e/flows/01-onboarding.yaml
maestro record e2e/flows/01-onboarding.yaml --output e2e/videos/01-onboarding.mp4
```

### Prerequis
- [ ] Backend running (`make dev` ou backend Docker)
- [ ] Build Expo dev (`npx expo prebuild`)
- [ ] Simulateur/emulateur demarre

## 5. Cibles Makefile

```bash
make mobile-e2e           # Executer tous les flows
make mobile-e2e-record    # Enregistrer les videos
```

## 6. Verdict

| Critere | Resultat |
|---------|---------|
| Maestro installe | OK (v2.4.0) |
| Flows YAML ecrits | 5 flows + 1 edge case |
| testIDs disponibles | 22 testIDs sur 8 ecrans |
| Execution | En attente de build Expo |
| CI workflow | A creer apres validation du PoC local |

Le PoC est **structurellement pret**. L'execution depend du build Expo (prebuild iOS ou Android) et du backend running.
