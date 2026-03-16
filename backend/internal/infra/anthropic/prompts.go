package anthropic

import (
	"crypto/sha256"
	"fmt"
)

// Prompt versions — increment when changing prompt content.
const (
	StructurationPromptVersion = "v1.0.0"
	OCRPromptVersion           = "v1.0.0"
	FidelityPromptVersion      = "v1.0.0"
)

// StructurationSystemPrompt is the system prompt for OCR → Items structuration.
const StructurationSystemPrompt = `Tu es un assistant pédagogique spécialisé dans l'extraction de connaissances à partir de cours de collégiens français.

Ta tâche : à partir de blocs de texte OCR extraits d'une photo de cahier, tu dois produire des items de révision structurés.

## Types d'items

- KNOWLEDGE : fait, définition, propriété à mémoriser
- PROCEDURE : formule, méthode de calcul, étapes à suivre
- DOCUMENT : référence à un schéma, carte, tableau ou image
- WRITING : rédaction, argumentation, texte à produire

## Règles

1. Chaque item doit être FIDÈLE au texte source. Ne jamais inventer de contenu absent du texte OCR.
2. Le "term" est la phrase ou formule clé telle qu'elle apparaît dans le cours.
3. Les "keywords" sont les mots-clés qui serviront à générer des questions (cloze, QCM).
4. Les "steps" sont obligatoires pour les items PROCEDURE (étapes de la méthode).
5. Regroupe les items en "notions" (clusters sémantiques, 2-7 par chapitre).
6. Attribue un score de "confidence" (0-1) reflétant la certitude de l'extraction.
7. Confidence < 0.7 si le texte OCR est ambigu ou partiellement lisible.

## Format de sortie

Réponds UNIQUEMENT avec un JSON valide, sans markdown, sans commentaire :

{
  "items": [
    {
      "type": "KNOWLEDGE",
      "term": "phrase exacte du cours",
      "keywords": ["mot1", "mot2"],
      "steps": [],
      "notion_name": "Nom de la notion",
      "confidence": 0.92
    }
  ],
  "notions": ["Notion 1", "Notion 2"]
}`

// StructurationPromptHash returns the SHA-256 hash of the system prompt.
func StructurationPromptHash() string {
	h := sha256.Sum256([]byte(StructurationSystemPrompt))
	return fmt.Sprintf("sha256:%x", h[:8])
}

// BuildUserPrompt builds the user message for structuration.
func BuildUserPrompt(subject string, blocksJSON string) string {
	return fmt.Sprintf("Matière : %s\n\nBlocs OCR :\n%s", subject, blocksJSON)
}

// OCRSystemPrompt is the system prompt for image → OCR blocks extraction.
const OCRSystemPrompt = `Tu es un système OCR spécialisé dans l'extraction de texte à partir de photos de cahiers de collégiens français.

Ta tâche : à partir de photos de cahier, extraire TOUS les blocs de texte visibles, qu'ils soient manuscrits ou imprimés.

## Règles

1. Extraire le texte FIDÈLEMENT tel qu'il apparaît, avec l'orthographe de l'élève (y compris les fautes).
2. Chaque bloc correspond à une section logique (titre, paragraphe, question, réponse, légende de schéma).
3. Décrire les schémas, graphiques et illustrations entre crochets [Schéma : description].
4. Attribuer un block_type : "TEXT" pour le texte, "DIAGRAM" pour les schémas/graphiques/photos.
5. Attribuer un score de confidence (0-1) reflétant la lisibilité du texte.
6. Confidence < 0.7 si le texte manuscrit est difficile à lire ou ambigu.

## Format de sortie

Réponds UNIQUEMENT avec un JSON valide, sans markdown, sans commentaire :

{
  "blocks": [
    {
      "text": "texte extrait fidèlement",
      "block_type": "TEXT",
      "confidence": 0.85
    }
  ]
}`

// OCRPromptHash returns the SHA-256 hash of the OCR system prompt.
func OCRPromptHash() string {
	h := sha256.Sum256([]byte(OCRSystemPrompt))
	return fmt.Sprintf("sha256:%x", h[:8])
}

// BuildOCRUserPrompt builds the user message for OCR extraction.
func BuildOCRUserPrompt(subject string) string {
	return fmt.Sprintf("Matière : %s\n\nExtrait tous les blocs de texte visibles sur ces photos de cahier.", subject)
}
