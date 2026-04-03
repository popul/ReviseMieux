# Etude de faisabilite — Tests E2E mobile avec video

> Date : 2026-04-03
> Issue : #35
> Statut : Recommandation finale

---

## 1. Objectif

Transformer les user flows documentes (`docs/ux/`) en tests E2E automatises sur l'app mobile React Native (Expo 54), avec **generation de video** a vitesse humaine, executes dans la CI GitHub Actions.

### Sources de scenarios

| Document | Scenarios |
|----------|-----------|
| `docs/ux/user-flow-complete.md` | 9 parcours happy path |
| `docs/ux/screen-inventory.md` | 20 ecrans avec CTA |
| `docs/ux/edge-cases.md` | 30 edge cases |
| `docs/ux/state-model.md` | 11 etats par ecran |

---

## 2. Comparatif des frameworks

### 2.1 Criteres d'evaluation

| Critere | Poids | Description |
|---------|-------|-------------|
| Compatibilite Expo | 30% | Fonctionne avec Expo dev builds (pas d'eject) |
| Facilite d'ecriture | 20% | Syntaxe proche des user flows |
| Generation de video | 20% | Enregistrement natif ou facilement configurable |
| CI GitHub Actions | 15% | Runner compatible, temps raisonnable |
| Maturite / communaute | 15% | Docs, issues, maintenance active |

### 2.2 Evaluation

| Critere | Maestro | Detox | Appium |
|---------|---------|-------|--------|
| **Compatibilite Expo** | Excellent — teste sur le .app/.apk built, pas besoin de config speciale | Bon — support officiel Expo dev builds, mais config initiale complexe | Correct — generique, fonctionne avec tout mais setup lourd |
| **Facilite d'ecriture** | **Excellent** — YAML declaratif, 1 flow = 1 fichier lisible | Moyen — JavaScript/Jest, verbose, APIs d'attente complexes | Faible — WebDriver protocol, tres verbeux |
| **Generation de video** | **Natif** — `maestro record flow.yaml` produit un MP4 | Manuel — `xcrun simctl recordVideo` + `adb screenrecord`, scripts custom | Manuel — plugins, config supplementaire |
| **CI GitHub Actions** | Bon — `maestro cloud` ou local avec emulateur | Bon — macOS runner requis pour iOS, Android sur Linux | Lourd — Appium server + emulateur, lent |
| **Maturite** | En croissance rapide (mobile.dev, acquis par Atlassian) | Mature (Wix, 5+ ans) | Tres mature mais decline pour mobile |
| **Vitesse humaine** | `maestro record --speed 0.5` ou delais YAML (`extendedWaitUntil`) | `waitFor` manuels, pas de speed control natif | Delais manuels |
| **Mapping user flows** | **1:1** — un fichier YAML par flow documente | Indirect — code JS a ecrire pour chaque flow | Indirect |

### 2.3 Exemple comparatif — Parcours Onboarding

**Maestro (YAML)** :
```yaml
# e2e/flows/01-onboarding.yaml
appId: com.revisemieux.app
---
- assertVisible: "Bienvenue"
- tapOn: "Commencer"
- assertVisible: "Chapitre demo"
- tapOn: "Reviser"
- assertVisible: "Q 1/"
- tapOn:
    id: "answer-input"
- inputText: "rho = m/V"
- tapOn: "Valider"
- assertVisible: "Correct"
```

**Detox (JavaScript)** :
```javascript
describe('Onboarding', () => {
  it('should complete first session', async () => {
    await expect(element(by.text('Bienvenue'))).toBeVisible();
    await element(by.text('Commencer')).tap();
    await expect(element(by.text('Chapitre demo'))).toBeVisible();
    await element(by.text('Reviser')).tap();
    await expect(element(by.text('Q 1/'))).toBeVisible();
    await element(by.id('answer-input')).typeText('rho = m/V');
    await element(by.text('Valider')).tap();
    await expect(element(by.text('Correct'))).toBeVisible();
  });
});
```

**Verdict** : Maestro est 3x plus concis et directement lisible par un non-dev (PO, designer).

---

## 3. Recommandation : Maestro

### Pourquoi Maestro

1. **Syntaxe YAML = mapping 1:1 avec les user flows** — chaque fichier `docs/ux/user-flow-complete.md` se traduit en un `.yaml` Maestro
2. **Video native** — `maestro record` produit un MP4 sans config
3. **Pas de code** — le PO peut relire et valider les flows
4. **Rapide a setup** — `npm install -g maestro` + `maestro test`
5. **CI compatible** — fonctionne avec l'emulateur Android sur GitHub Actions (Ubuntu + KVM)

### Limites

- Moins flexible que Detox pour les assertions complexes
- Pas de support direct des mocks reseau (backend doit tourner ou etre stubbe)
- Plus jeune que Detox (mais Atlassian le maintient maintenant)

---

## 4. Architecture proposee

### Structure des fichiers

```
mobile/e2e/
├── flows/                          # 1 fichier = 1 parcours user-flow
│   ├── 01-onboarding.yaml          # docs/ux §Onboarding
│   ├── 02-capture-j0.yaml          # §Capture J0
│   ├── 03-dashboard.yaml           # §Dashboard eleve
│   ├── 04-session-revision.yaml    # §Session de revision
│   ├── 05-revue-erreurs.yaml       # §Revue des erreurs
│   ├── 06-retour-absence.yaml      # §Retour apres absence
│   ├── 07-dashboard-parent.yaml    # §Dashboard parent
│   ├── 08-validation-hitl.yaml     # §Validation HITL
│   └── 09-recovery-ocr.yaml        # §Gestion OCR incomplet
├── edge-cases/                     # 1 fichier = 1 edge case
│   ├── EC-01-photo-floue.yaml
│   ├── EC-07-session-fermee.yaml
│   ├── EC-09-tous-solid.yaml
│   └── EC-22-retour-5-jours.yaml
├── config/
│   └── maestro-config.yaml         # Config globale (timeouts, device)
└── videos/                         # Gitignored — videos generees
```

### Convention de nommage

Chaque flow reference les ACs couvertes en commentaire :
```yaml
# Flow: 01-onboarding
# ACs: Z8-AC01, Z8-AC02, Z8-AC08
# Source: docs/ux/user-flow-complete.md §Onboarding
appId: com.revisemieux.app
---
```

---

## 5. Generation de video a vitesse humaine

### Strategie

Maestro supporte des delais entre les actions via `extendedWaitUntil` et `waitForAnimationToEnd`. Pour une video lisible :

```yaml
# Delai global entre chaque action (ms)
- extendedWaitUntil:
    timeout: 1000
    visible: ".*"

# Ou delai explicite apres une action cle
- tapOn: "Valider"
- scroll  # pause visuelle
- extendedWaitUntil:
    timeout: 800
    visible: ".*"
```

### Enregistrement

```bash
# Enregistre le flow en video MP4
maestro record mobile/e2e/flows/01-onboarding.yaml --output mobile/e2e/videos/01-onboarding.mp4

# Avec vitesse reduite (0.5x)
maestro record --speed 0.5 mobile/e2e/flows/01-onboarding.yaml
```

### Post-processing (optionnel)

Si la vitesse native n'est pas assez lente :
```bash
# ffmpeg ralentit a 0.5x
ffmpeg -i input.mp4 -filter:v "setpts=2.0*PTS" -filter:a "atempo=0.5" output-slow.mp4
```

---

## 6. Execution dans GitHub Actions

### Workflow propose

```yaml
# .github/workflows/e2e-mobile.yml
name: E2E Mobile

on:
  pull_request:
    branches: [reboot]
  schedule:
    - cron: '0 6 * * 1-5'  # Lundi-vendredi 6h UTC

jobs:
  e2e-android:
    runs-on: ubuntu-latest
    timeout-minutes: 30

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with: { node-version: '20' }

      - uses: actions/setup-java@v4
        with:
          distribution: 'temurin'
          java-version: '17'

      - name: Install Maestro
        run: |
          curl -Ls "https://get.maestro.mobile.dev" | bash
          echo "$HOME/.maestro/bin" >> $GITHUB_PATH

      - name: Build Expo dev client (Android)
        working-directory: mobile
        run: |
          npm ci --legacy-peer-deps
          npx expo prebuild --platform android
          cd android && ./gradlew assembleDebug

      - name: Enable KVM
        run: |
          echo 'KERNEL=="kvm", GROUP="kvm", MODE="0666", OPTIONS+="static_node=kvm"' | sudo tee /etc/udev/rules.d/99-kvm4all.rules
          sudo udevadm control --reload-rules
          sudo udevadm trigger --name-match=kvm

      - name: Start Android Emulator
        uses: reactivecircus/android-emulator-runner@v2
        with:
          api-level: 34
          arch: x86_64
          script: |
            # Install app
            adb install mobile/android/app/build/outputs/apk/debug/app-debug.apk

            # Run all flows with video recording
            maestro record mobile/e2e/flows/ --output mobile/e2e/videos/ --format mp4

      - name: Upload videos
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: e2e-videos
          path: mobile/e2e/videos/*.mp4
          retention-days: 7
```

### Estimation couts

| Runner | Cout/min | Temps estime | Cout/run | Cout/mois (1x/jour) |
|--------|----------|-------------|----------|---------------------|
| ubuntu-latest (KVM) | $0.008 | ~15 min | ~$0.12 | ~$2.50 |
| macos-latest (iOS) | $0.08 | ~10 min | ~$0.80 | ~$16.80 |

**Recommandation** : Android seulement en CI (~$2.50/mois). iOS teste manuellement en local.

---

## 7. Prerequis techniques

| Prerequis | Statut | Action |
|-----------|--------|--------|
| Build Expo dev Android | Non fait | `npx expo prebuild --platform android` |
| Maestro installe | Non fait | `curl -Ls "https://get.maestro.mobile.dev" \| bash` |
| Backend mock ou stub | Non fait | Options : mock server, ou backend reel Docker |
| Fixtures (chapitre demo) | Fait | `POST /api/v1/onboarding/seed-demo` |
| testIDs sur les composants | Partiel | Ajouter `testID` sur les elements interactifs |

### Backend pour les tests E2E

Deux approches :

**Option A — Backend reel (Docker)** :
```yaml
services:
  backend:
    image: revisemieux-backend
    depends_on: [postgres]
  postgres:
    image: postgres:16
```
Plus realiste mais plus lent et complexe en CI.

**Option B — Mock server** :
```bash
# Maestro supporte les mock responses
maestro mock --server http://localhost:3000 mobile/e2e/mocks/
```
Plus simple mais moins fidele.

**Recommandation** : Option A pour les flows critiques (onboarding, session), Option B pour les edge cases.

---

## 8. Plan de deploiement

### Phase 1 — Proof of Concept (1-2 jours)

- [ ] Installer Maestro en local
- [ ] Ecrire le flow `01-onboarding.yaml`
- [ ] Enregistrer la video
- [ ] Valider que le PO peut relire le YAML

### Phase 2 — 9 flows principaux (3-5 jours)

- [ ] Ecrire les 9 flows principaux
- [ ] Ajouter les `testID` manquants dans les composants mobile
- [ ] Configurer le backend Docker pour les tests
- [ ] Enregistrer les 9 videos

### Phase 3 — CI (1-2 jours)

- [ ] Configurer `.github/workflows/e2e-mobile.yml`
- [ ] Build Expo Android en CI
- [ ] Executer les flows + upload videos comme artifacts

### Phase 4 — Edge cases (ongoing)

- [ ] Ajouter les edge cases progressivement
- [ ] 1 edge case = 1 fichier YAML + 1 video

---

## 9. Go/No-go

| Critere | Evaluation | Verdict |
|---------|-----------|---------|
| Framework existe et est maintenu | Maestro par Atlassian, releases frequentes | GO |
| Compatible Expo 54 | Oui (teste sur le .apk built) | GO |
| Video native | `maestro record` produit MP4 | GO |
| CI faisable | Ubuntu + KVM + emulateur Android | GO |
| Cout acceptable | ~$2.50/mois (Android, 1x/jour) | GO |
| Effort initial | ~1-2 jours pour le PoC | GO |
| Mapping user flows | YAML 1:1, lisible par PO | GO |

**Recommandation finale : GO** — commencer par le PoC sur le parcours onboarding.

---

## 10. Risques et mitigations

| Risque | Probabilite | Impact | Mitigation |
|--------|------------|--------|------------|
| Build Expo Android trop long en CI | Moyenne | CI lente (~20 min) | Cache du build Gradle + prebuild |
| Emulateur instable en CI | Faible | Tests flaky | Retry x2 + timeouts genereux |
| testIDs manquants | Haute | Flows ne trouvent pas les elements | Sprint dedie pour ajouter les testIDs |
| Backend non disponible en CI | Moyenne | Flows echouent | Docker Compose dans le workflow |
| Maestro ne supporte pas un geste specifique | Faible | Flow incomplet | Fallback sur `evalScript` (JS inline) |
