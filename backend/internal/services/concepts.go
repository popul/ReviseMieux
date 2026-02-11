// Package services contient les services métier de l'application
package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/revisemieux/backend/internal/llm"
	"github.com/revisemieux/backend/internal/store"
)

// Codes d'erreur concepts
var (
	ErrParsingConceptsEchoue = &ErreurGeneration{Code: "PARSING_CONCEPTS_ECHOUE", Message: "Erreur lors du parsing des concepts générés"}
	ErrConceptNonTrouve      = &ErreurGeneration{Code: "CONCEPT_NON_TROUVE", Message: "Concept non trouvé"}
)

// ServiceConcepts gère l'extraction et la gestion des concepts
type ServiceConcepts struct {
	gestionnaireLLM *llm.GestionnaireLLM
	coursRepo       store.CoursRepository
	conceptsRepo    store.ConceptsRepository
}

// NouveauServiceConcepts crée une nouvelle instance du service de concepts
func NouveauServiceConcepts(
	gestionnaireLLM *llm.GestionnaireLLM,
	coursRepo store.CoursRepository,
	conceptsRepo store.ConceptsRepository,
) *ServiceConcepts {
	return &ServiceConcepts{
		gestionnaireLLM: gestionnaireLLM,
		coursRepo:       coursRepo,
		conceptsRepo:    conceptsRepo,
	}
}

// ResultatExtractionConcepts contient le résultat de l'extraction
type ResultatExtractionConcepts struct {
	Concepts       []*store.Concept `json:"concepts"`
	NombreExtraits int              `json:"nombreExtraits"`
}

// ExtraireConcepts extrait les concepts clés d'un cours via le LLM
func (s *ServiceConcepts) ExtraireConcepts(ctx context.Context, coursID string) (*ResultatExtractionConcepts, error) {
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
	prompt := s.construirePromptConcepts(texte)

	// Appeler le LLM
	llmOptions := llm.OptionsGeneration{
		Temperature:   0.5,
		MaxTokens:     4000,
		FormatReponse: "json",
		SystemPrompt:  "Tu es un professeur expert en analyse de cours pour lycéens. Tu identifies les concepts clés d'un cours. Tu réponds uniquement en JSON valide.",
	}

	reponseJSON, err := s.gestionnaireLLM.GenererJSON(ctx, prompt, nil, llmOptions)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrGenerationLLMEchouee, err.Error())
	}

	// Parser la réponse
	concepts, err := s.parserReponseConcepts(reponseJSON, coursID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrParsingConceptsEchoue, err.Error())
	}

	// Sauvegarder les concepts en base
	if s.conceptsRepo != nil {
		// Supprimer les anciens concepts du cours avant d'en créer de nouveaux
		if err := s.conceptsRepo.SupprimerParCours(ctx, coursID); err != nil {
			return nil, fmt.Errorf("erreur suppression anciens concepts: %w", err)
		}

		if err := s.conceptsRepo.CreerPlusieurs(ctx, concepts); err != nil {
			return nil, fmt.Errorf("erreur sauvegarde concepts: %w", err)
		}
	}

	return &ResultatExtractionConcepts{
		Concepts:       concepts,
		NombreExtraits: len(concepts),
	}, nil
}

// GetConceptsByCours récupère les concepts existants d'un cours
func (s *ServiceConcepts) GetConceptsByCours(ctx context.Context, coursID string) ([]*store.Concept, error) {
	if s.conceptsRepo == nil {
		return nil, ErrServiceNonDisponible
	}

	return s.conceptsRepo.ListerParCours(ctx, coursID)
}

// MettreAJourConcept met à jour un concept existant
func (s *ServiceConcepts) MettreAJourConcept(ctx context.Context, conceptID string, nom string, definition string, importance string, position interface{}) (*store.Concept, error) {
	if s.conceptsRepo == nil {
		return nil, ErrServiceNonDisponible
	}

	// Récupérer le concept existant
	concept, err := s.conceptsRepo.ObtenirParID(ctx, conceptID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrConceptNonTrouve, err.Error())
	}

	// Mettre à jour les champs si fournis
	if nom != "" {
		concept.Nom = nom
	}
	if definition != "" {
		concept.Definition = definition
	}
	if importance != "" {
		concept.Importance = importance
	}

	// Gérer la position (interface{} pour accepter le type API)
	if position != nil {
		// Sérialiser et désérialiser pour convertir le type
		posJSON, err := json.Marshal(position)
		if err == nil {
			var pos store.PositionDansCours
			if err := json.Unmarshal(posJSON, &pos); err == nil {
				concept.PositionDansCours = &pos
			}
		}
	}

	if err := s.conceptsRepo.MettreAJour(ctx, concept); err != nil {
		return nil, fmt.Errorf("erreur mise à jour concept: %w", err)
	}

	return concept, nil
}

// SupprimerConcept supprime un concept par son ID
func (s *ServiceConcepts) SupprimerConcept(ctx context.Context, conceptID string) error {
	if s.conceptsRepo == nil {
		return ErrServiceNonDisponible
	}

	if err := s.conceptsRepo.Supprimer(ctx, conceptID); err != nil {
		return fmt.Errorf("%w: %s", ErrConceptNonTrouve, err.Error())
	}

	return nil
}

// construirePromptConcepts construit le prompt pour l'extraction de concepts
func (s *ServiceConcepts) construirePromptConcepts(texte string) string {
	return fmt.Sprintf(`Tu es un professeur expert en analyse de contenus pédagogiques pour lycéens.

À partir du cours suivant, identifie les concepts clés (notions, termes importants, formules, théorèmes, etc.).

Cours :
"""
%s
"""

Instructions :
- Identifie entre 5 et 20 concepts selon la richesse du contenu
- Pour chaque concept, fournis :
  - Le nom du concept (court et précis)
  - Une définition claire et pédagogique (2-3 phrases maximum)
  - Le niveau d'importance : "essentiel" (incontournable pour comprendre le cours), "important" (nécessaire pour bien maîtriser le sujet), ou "secondaire" (complément utile)
  - La position approximative dans le texte (index de début et de fin du passage concerné, si identifiable)
- Base-toi UNIQUEMENT sur le contenu du cours fourni
- Classe les concepts du plus essentiel au plus secondaire

Réponds UNIQUEMENT avec un JSON valide au format suivant, sans texte avant ou après :
{
  "concepts": [
    {
      "nom": "Nom du concept",
      "definition": "Définition claire et pédagogique du concept.",
      "importance": "essentiel|important|secondaire",
      "position": {"debut": 0, "fin": 100}
    }
  ]
}`, texte)
}

// reponseConceptsJSON représente la structure de réponse du LLM pour les concepts
type reponseConceptsJSON struct {
	Concepts []struct {
		Nom        string `json:"nom"`
		Definition string `json:"definition"`
		Importance string `json:"importance"`
		Position   *struct {
			Debut int `json:"debut"`
			Fin   int `json:"fin"`
		} `json:"position,omitempty"`
	} `json:"concepts"`
}

// parserReponseConcepts parse la réponse JSON du LLM en concepts
func (s *ServiceConcepts) parserReponseConcepts(reponseJSON []byte, coursID string) ([]*store.Concept, error) {
	var reponse reponseConceptsJSON
	if err := json.Unmarshal(reponseJSON, &reponse); err != nil {
		return nil, fmt.Errorf("erreur parsing JSON: %w", err)
	}

	concepts := make([]*store.Concept, 0, len(reponse.Concepts))
	for _, c := range reponse.Concepts {
		// Valider l'importance
		importance := c.Importance
		if importance != "essentiel" && importance != "important" && importance != "secondaire" {
			importance = "important" // valeur par défaut
		}

		concept := &store.Concept{
			CoursID:    coursID,
			Nom:        c.Nom,
			Definition: c.Definition,
			Importance: importance,
		}

		// Ajouter la position si disponible
		if c.Position != nil {
			concept.PositionDansCours = &store.PositionDansCours{
				Debut: c.Position.Debut,
				Fin:   c.Position.Fin,
			}
		}

		concepts = append(concepts, concept)
	}

	return concepts, nil
}
