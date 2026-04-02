# Révise Mieux — Mobile

Application mobile React Native + Expo pour collégiens. Transforme des photos de cahier en sessions de révision interactives avec répétition espacée.

## Prérequis

- **Node.js** 20+
- **Expo CLI** : `npm install -g expo-cli` (ou via `npx`)
- **EAS CLI** : `npm install -g eas-cli` (pour les builds natifs)
- **Compte Expo** : [expo.dev](https://expo.dev) (gratuit)
- **Apple Developer Account** : pour installer sur iPhone physique

## Démarrage rapide

```bash
# 1. Installer les dépendances
make deps

# 2. Builder le dev client pour ton iPhone (première fois uniquement)
make eas-login          # connexion Expo
make device-register    # enregistre ton iPhone (QR code)
make build-dev          # build EAS → lien d'installation

# 3. Lancer le serveur Metro
make dev
```

Ouvre l'app **Révise Mieux** sur ton iPhone — elle se connecte automatiquement au serveur Metro (même réseau Wi-Fi requis).

## Commandes

| Commande | Description |
|----------|-------------|
| `make dev` | Lance le serveur Metro (dev client) |
| `make ios` | Lance sur simulateur iOS |
| `make web` | Lance sur navigateur web |
| `make test` | Tests Jest |
| `make test-watch` | Tests en mode watch |
| `make typecheck` | Vérification TypeScript |
| `make lint` | Lint complet (TypeScript) |

### Builds EAS (natifs)

| Commande | Description |
|----------|-------------|
| `make build-dev` | Build dev client iPhone physique |
| `make build-dev-sim` | Build dev client simulateur iOS |
| `make build-preview` | Build preview (TestFlight-ready) |
| `make build-android` | Build dev client Android |
| `make device-register` | Enregistrer un device pour les builds dev |

### Utilitaires

| Commande | Description |
|----------|-------------|
| `make deps` | Installe les dépendances npm |
| `make prebuild` | Génère les projets natifs `ios/` et `android/` |
| `make clean` | Supprime les caches Expo |
| `make nuke` | Supprime tout (node_modules, natifs, caches) |

## Structure

```
mobile/
├── app/                  # Écrans (Expo Router, file-based routing)
│   ├── (tabs)/           # Navigation par onglets
│   │   ├── index.tsx     #   Dashboard (chapitres, progression)
│   │   ├── capture.tsx   #   Capture photo
│   │   └── settings.tsx  #   Paramètres
│   ├── chapter/[id].tsx  # Détail chapitre (notions, items, mastery)
│   ├── session/[id].tsx  # Session de quiz (6 types de questions)
│   ├── onboarding.tsx    # Onboarding premier lancement
│   ├── capture.tsx       # Capture plein écran (modal)
│   └── processing.tsx    # Animation traitement OCR
├── components/           # Composants réutilisables
├── constants/            # Design tokens (couleurs, typo, spacing)
├── services/             # Client API typé
└── __tests__/            # Tests Jest
```

## Stack

| Technologie | Version |
|-------------|---------|
| Expo | 55 |
| React Native | 0.83 |
| React | 19 |
| TypeScript | 5.9 (strict) |
| Expo Router | 55 (file-based) |
| Jest | 30 |

## Tester sur iPhone physique

1. **Compte Expo** : `make eas-login`
2. **Enregistrer ton iPhone** : `make device-register` → scanne le QR code depuis Safari sur ton iPhone
3. **Lancer le build** : `make build-dev` → le build se fait dans le cloud EAS (~10-15 min la première fois)
4. **Installer** : ouvre le lien reçu par mail/dans le terminal depuis ton iPhone
5. **Développer** : `make dev` → l'app se connecte au serveur Metro

Les builds suivants sont plus rapides grâce au cache EAS. Tu n'as besoin de rebuilder que si tu modifies les dépendances natives ou `app.json`.

## API Backend

L'app communique avec le backend Go sur `http://localhost:8080/api/v1` (configurable via Expo Constants). Pour que l'iPhone physique atteigne le backend :

- Le Mac et l'iPhone doivent être sur le **même réseau Wi-Fi**
- Remplace `localhost` par l'IP locale du Mac dans la config (ex: `http://192.168.1.42:8080/api/v1`)
- Ou lance le backend depuis la racine : `make dev` (lance backend + mobile)
