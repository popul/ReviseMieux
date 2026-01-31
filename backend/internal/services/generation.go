// Package services contient les services métier de l'application
package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/revisemieux/backend/internal/llm"
	"github.com/revisemieux/backend/internal/store"
)

// OptionsGenerationFiches contient les options pour la génération de fiches
type OptionsGenerationFiches struct {
	// NombreFiches est le nombre de fiches à générer (par défaut: auto)
	NombreFiches int
	// Difficulte filtre la difficulté des fiches ("facile", "moyen", "difficile", "" pour mélangé)
	Difficulte string
}

// ResultatGenerationFiches contient le résultat de la génération
type ResultatGenerationFiches struct {
	Fiches       []*store.Fiche `json:"fiches"`
	NombreGenere int            `json:"nombreGenere"`
}

// ErreurGeneration représente une erreur du service de génération
type ErreurGeneration struct {
	Code    string
	Message string
}

func (e *ErreurGeneration) Error() string {
	return e.Message
}

// Codes d'erreur génération
var (
	ErrCoursNonTrouve        = &ErreurGeneration{Code: "COURS_NON_TROUVE", Message: "Cours non trouvé"}
	ErrCoursVideOCR          = &ErreurGeneration{Code: "COURS_VIDE", Message: "Le cours n'a pas de texte OCR"}
	ErrGenerationLLMEchouee  = &ErreurGeneration{Code: "GENERATION_ECHOUEE", Message: "Échec de la génération par le LLM"}
	ErrParsingFichesEchoue   = &ErreurGeneration{Code: "PARSING_ECHOUE", Message: "Erreur lors du parsing des fiches générées"}
	ErrServiceNonDisponible  = &ErreurGeneration{Code: "SERVICE_NON_DISPONIBLE", Message: "Le service de génération n'est pas disponible"}
)

// ServiceGeneration gère la génération de contenu (fiches, quiz, mindmaps)
type ServiceGeneration struct {
	gestionnaireLLM *llm.GestionnaireLLM
	coursRepo       store.CoursRepository
	fichesRepo      store.FichesRepository
}

// NouveauServiceGeneration crée une nouvelle instance du service de génération
func NouveauServiceGeneration(
	gestionnaireLLM *llm.GestionnaireLLM,
	coursRepo store.CoursRepository,
	fichesRepo store.FichesRepository,
) *ServiceGeneration {
	return &ServiceGeneration{
		gestionnaireLLM: gestionnaireLLM,
		coursRepo:       coursRepo,
		fichesRepo:      fichesRepo,
	}
}

// GenererFiches génère des fiches de révision pour un cours
func (s *ServiceGeneration) GenererFiches(ctx context.Context, coursID string, options *OptionsGenerationFiches) (*ResultatGenerationFiches, error) {
	if s.gestionnaireLLM == nil {
		return nil, ErrServiceNonDisponible
	}

	// Récupérer le cours
	cours, err := s.coursRepo.ObtenirParID(ctx, coursID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrCoursNonTrouve, err.Error())
	}

	// Vérifier que le cours a du texte
	texte := cours.TexteCorrige
	if texte == "" {
		texte = cours.TexteOCR
	}
	if texte == "" {
		return nil, ErrCoursVideOCR
	}

	// Construire le prompt
	prompt := s.construirePromptFiches(texte, options)

	// Appeler le LLM
	llmOptions := llm.OptionsGeneration{
		Temperature:   0.7,
		MaxTokens:     4000,
		FormatReponse: "json",
		SystemPrompt:  "Tu es un professeur expert en création de supports de révision pour lycéens. Tu réponds uniquement en JSON valide.",
	}

	reponseJSON, err := s.gestionnaireLLM.GenererJSON(ctx, prompt, nil, llmOptions)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrGenerationLLMEchouee, err.Error())
	}

	// Parser la réponse
	fiches, err := s.parserReponseFiches(reponseJSON, coursID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrParsingFichesEchoue, err.Error())
	}

	// Sauvegarder les fiches en base
	if s.fichesRepo != nil {
		if err := s.fichesRepo.CreerPlusieurs(ctx, fiches); err != nil {
			return nil, fmt.Errorf("erreur sauvegarde fiches: %w", err)
		}
	}

	return &ResultatGenerationFiches{
		Fiches:       fiches,
		NombreGenere: len(fiches),
	}, nil
}

// construirePromptFiches construit le prompt pour la génération de fiches
func (s *ServiceGeneration) construirePromptFiches(texte string, options *OptionsGenerationFiches) string {
	nombreFiches := ""
	if options != nil && options.NombreFiches > 0 {
		nombreFiches = fmt.Sprintf("- Génère exactement %d fiches\n", options.NombreFiches)
	} else {
		nombreFiches = "- Génère entre 5 et 15 fiches selon la quantité de contenu\n"
	}

	difficulte := ""
	if options != nil && options.Difficulte != "" {
		difficulte = fmt.Sprintf("- Toutes les fiches doivent être de difficulté: %s\n", options.Difficulte)
	} else {
		difficulte = "- Varie la difficulté (facile, moyen, difficile) de manière équilibrée\n"
	}

	return fmt.Sprintf(`Tu es un professeur expert en création de supports de révision pour lycéens.

À partir du cours suivant, génère des fiches de révision efficaces.

Cours :
"""
%s
"""

Instructions :
%s%s- Crée des fiches question/réponse basées UNIQUEMENT sur le contenu fourni
- Les questions doivent favoriser le rappel actif (pas de simples définitions)
- Varie les types : faits, concepts, relations de cause à effet
- Les réponses doivent être concises mais complètes
- Attribue une difficulté à chaque fiche

Réponds UNIQUEMENT avec un JSON valide au format suivant, sans texte avant ou après :
{
  "fiches": [
    {
      "question": "...",
      "reponse": "...",
      "difficulte": "facile|moyen|difficile"
    }
  ]
}`, texte, nombreFiches, difficulte)
}

// reponseFichesJSON représente la structure de réponse du LLM
type reponseFichesJSON struct {
	Fiches []struct {
		Question   string `json:"question"`
		Reponse    string `json:"reponse"`
		Difficulte string `json:"difficulte"`
	} `json:"fiches"`
}

// parserReponseFiches parse la réponse JSON du LLM en fiches
func (s *ServiceGeneration) parserReponseFiches(reponseJSON []byte, coursID string) ([]*store.Fiche, error) {
	var reponse reponseFichesJSON
	if err := json.Unmarshal(reponseJSON, &reponse); err != nil {
		return nil, fmt.Errorf("erreur parsing JSON: %w", err)
	}

	fiches := make([]*store.Fiche, 0, len(reponse.Fiches))
	for i, f := range reponse.Fiches {
		// Valider la difficulté
		difficulte := f.Difficulte
		if difficulte != "facile" && difficulte != "moyen" && difficulte != "difficile" {
			difficulte = "moyen" // valeur par défaut
		}

		fiche := &store.Fiche{
			CoursID:    coursID,
			Question:   f.Question,
			Reponse:    f.Reponse,
			Difficulte: difficulte,
			Ordre:      i + 1,
		}
		fiches = append(fiches, fiche)
	}

	return fiches, nil
}

// ObtenirFichesParCours récupère les fiches existantes d'un cours
func (s *ServiceGeneration) ObtenirFichesParCours(ctx context.Context, coursID string) ([]*store.Fiche, error) {
	if s.fichesRepo == nil {
		return nil, ErrServiceNonDisponible
	}

	return s.fichesRepo.ListerParCours(ctx, coursID)
}
