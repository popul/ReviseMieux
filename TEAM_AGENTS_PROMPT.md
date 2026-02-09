# Prompt de reprise du projet Révise Mieux

## Contexte du projet

**Révise Mieux** est une application d'aide à la révision pour collégiens et lycéens. L'objectif principal est d'accompagner les élèves dans leur préparation aux contrôles en transformant leurs cours (photos de cahiers, documents scannés) en supports de révision interactifs.

## Stack technique

| Composant | Technologie |
|-----------|-------------|
| Backend | Go (Gin framework) |
| Frontend | React 19 + TypeScript + Vite |
| CSS | Tailwind CSS v4 |
| Database | PostgreSQL 16 |
| LLM | OpenAI API (GPT-4) |
| OCR | OpenAI Vision API |
| Infra | Docker Compose (dev), Helm/Kubernetes (prod) |
| Tests E2E | Playwright |

## État actuel du projet

### Ce qui fonctionne
- OCR des images de cours via OpenAI Vision
- Sauvegarde des cours en base de données
- Génération de fiches de révision
- Génération de quiz interactifs
- Génération de mindmaps (carte mentale)
- Interface de base pour naviguer entre les pages
- Tests E2E de base

### Problèmes UX à résoudre en priorité

#### 1. Expérience d'OCRisation non intuitive
L'expérience actuelle d'OCR est frustrante pour l'utilisateur :
- **Réordonnancement des images** : Difficile de réordonner les pages scannées
- **Pas de correspondance visuelle texte/image** : L'utilisateur ne peut pas voir quel texte OCR correspond à quelle partie de l'image
- **Édition du texte déconnectée** : Le texte OCR est édité dans un champ séparé, sans contexte visuel

#### 2. UX cible pour l'OCR
L'expérience idéale serait :
- Afficher l'image en background
- Superposer le texte OCR au-dessus de l'image, positionné approximativement là où il a été détecté
- Permettre l'édition du texte directement sur l'image (comme un calque)
- Permettre le drag & drop pour réordonner les pages
- Afficher un indicateur de confiance par zone de texte

---

## Parcours utilisateur complet à implémenter

Le cas d'usage principal est : **Un élève prépare un contrôle en révisant sa leçon**.

### Étape 1 : OCRiser le cours
**Page : `/scanner`**
- Upload des photos de cours (prises au téléphone)
- Support multi-pages avec réordonnancement drag & drop
- Prévisualisation avant envoi
- Détection automatique du titre et de la matière

### Étape 2 : Revoir le résultat OCR
**Page : `/cours?id={id}` (mode édition)**
- Affichage de l'image avec le texte OCR superposé
- Navigation entre les pages (vignettes cliquables)
- Édition du texte directement sur l'image
- Zones incertaines surlignées (faible confiance OCR)
- Réordonnancement des pages par drag & drop
- Sauvegarde des modifications

### Étape 3 : Résumé et concepts clés
**Page : `/cours?id={id}` (onglet Résumé)**
- Génération automatique d'un résumé structuré du cours
- Extraction des **concepts clés** susceptibles de tomber au contrôle
- Chaque concept doit être :
  - Nommé clairement
  - Défini brièvement
  - Lié à la partie du cours source (traçabilité)
  - Taggé par importance (essentiel, important, secondaire)

### Étape 4 : Fiches de révision
**Page : `/fiches?coursId={id}`**
- Génération de fiches de révision à partir des concepts
- Chaque fiche contient :
  - Question recto
  - Réponse verso
  - Lien vers le concept source
  - Lien vers la partie du cours concernée
- Mode révision avec système de répétition espacée
- Statistiques de maîtrise par concept

### Étape 5 : Mindmap interactive
**Page : `/mindmap?coursId={id}`**
- Visualisation des concepts sous forme de carte mentale
- Nœuds cliquables menant :
  - Vers la définition du concept
  - Vers la partie du cours OCRisée correspondante
  - Vers les fiches liées
- Zoom et pan
- Export en image

### Étape 6 : Quiz par difficulté
**Page : `/quiz?coursId={id}`**
- Génération de quiz avec 3 niveaux de difficulté :
  - Facile : Questions de définition, reconnaissance
  - Moyen : Questions de compréhension, application
  - Difficile : Questions de synthèse, analyse
- Chaque question liée aux concepts testés
- Feedback immédiat avec explication
- Recommandations de révision basées sur les erreurs

### Étape 7 : Lexique et quiz vocabulaire
**Page : `/lexique?coursId={id}`**
- Extraction automatique des termes clés avec définitions
- Quiz de vocabulaire (terme → définition, définition → terme)
- Jeu de mémorisation (flashcards rapides)
- Suivi de la maîtrise du vocabulaire

### Étape 8 : Examen blanc
**Page : `/examen-blanc?coursId={id}`**
- Génération d'un examen complet couvrant tous les concepts
- Timer optionnel
- Système d'indices progressifs :
  - Indice 1 : Rappel du concept concerné
  - Indice 2 : Aide méthodologique
  - Indice 3 : Début de réponse
- Correction automatique avec :
  - Note estimée
  - Points forts identifiés
  - Concepts à retravailler
  - Plan de révision personnalisé

### Étape 9 : Upload du devoir corrigé
**Page : `/analyser`**
- Upload de la copie corrigée par le professeur
- Extraction automatique :
  - Note obtenue
  - Remarques du professeur
  - Erreurs identifiées
- Analyse des concepts non maîtrisés
- Génération d'un plan de remédiation :
  - Revoir les fiches des concepts manqués
  - Quiz ciblés sur les points faibles
  - Exercices supplémentaires

---

## Modèle de données : Concepts

Créer une entité `Concept` centrale qui relie tout :

```
Concept {
  id: UUID
  coursId: UUID
  nom: string
  definition: string
  importance: "essentiel" | "important" | "secondaire"
  positionDansCours: { pageIndex: number, debut: number, fin: number }
  fiches: Fiche[]
  questionsQuiz: Question[]
  termesLexique: Terme[]
  noeudMindmap: NoeudMindmap
}
```

---

## Fichiers clés à connaître

### Frontend
- `frontend/src/pages/Cours.tsx` - Page de détail d'un cours
- `frontend/src/pages/Scanner.tsx` - Page d'upload et OCR
- `frontend/src/pages/Fiches.tsx` - Révision par fiches
- `frontend/src/pages/Quiz.tsx` - Quiz interactif
- `frontend/src/pages/Mindmap.tsx` - Carte mentale
- `frontend/src/pages/Analyser.tsx` - Analyse de copie corrigée
- `frontend/src/services/api.ts` - Client API

### Backend
- `backend/internal/api/handlers_ocr.go` - Handler OCR
- `backend/internal/api/handlers_images.go` - Gestion des images
- `backend/internal/services/ocr.go` - Service OCR avec OpenAI Vision
- `backend/internal/services/generation.go` - Génération de contenu
- `backend/internal/store/cours.go` - Repository cours

### Configuration
- `docker-compose.yml` - Environnement de dev
- `frontend/nginx.conf` - Proxy nginx
- `docs/PRD.md` - Product Requirements Document complet

---

## Commandes utiles

```bash
# Démarrer l'environnement de dev
make dev

# Logs du backend
docker logs -f revisemieux-backend

# Tests E2E
cd frontend && npm run test:e2e

# Rebuild après modification
docker compose up -d --build
```

---

## Priorités de développement suggérées

### Phase 1 : Corriger l'UX OCR (critique)
1. Implémenter l'overlay texte sur image
2. Améliorer le réordonnancement des pages
3. Ajouter les zones de confiance visuelles

### Phase 2 : Concepts et traçabilité
1. Créer l'entité Concept en base
2. Extraction des concepts depuis le texte OCR
3. Lier les contenus existants aux concepts

### Phase 3 : Enrichir les supports de révision
1. Lexique avec quiz vocabulaire
2. Améliorer les quiz avec niveaux de difficulté
3. Traçabilité mindmap → cours

### Phase 4 : Examen blanc
1. Génération d'examen complet
2. Système d'indices progressifs
3. Correction et plan de remédiation

### Phase 5 : Boucle de feedback
1. Upload copie corrigée
2. Extraction note et remarques
3. Plan de révision personnalisé

---

## Notes importantes

- **Langue** : Toute l'interface est en français
- **Public cible** : Collégiens et lycéens (12-18 ans)
- **Design system** : Couleurs chaleureuses (coral #E85D4C, teal #1A4D4D, gold #F5C542, cream #FBF8F3)
- **Accessibilité** : Prévoir taille de texte ajustable, mode daltonien, contraste élevé
- **Mobile-first** : Les photos sont prises sur téléphone, l'interface doit être responsive
