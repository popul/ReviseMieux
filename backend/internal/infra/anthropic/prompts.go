package anthropic

import (
	"crypto/sha256"
	"fmt"
)

// Prompt versions — increment when changing prompt content.
const (
	StructurationPromptVersion = "v1.0.0"
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
