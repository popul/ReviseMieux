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
	lexiqueRepo     store.LexiqueRepository
}

// NouveauServiceConcepts crée une nouvelle instance du service de concepts
func NouveauServiceConcepts(
	gestionnaireLLM *llm.GestionnaireLLM,
	coursRepo store.CoursRepository,
	conceptsRepo store.ConceptsRepository,
	lexiqueRepo store.LexiqueRepository,
) *ServiceConcepts {
	return &ServiceConcepts{
		gestionnaireLLM: gestionnaireLLM,
		coursRepo:       coursRepo,
		conceptsRepo:    conceptsRepo,
		lexiqueRepo:     lexiqueRepo,
	}
}

// ResultatExtractionConcepts contient le résultat de l'extraction
type ResultatExtractionConcepts struct {
	Concepts       []*store.Concept `json:"concepts"`
	NombreExtraits int              `json:"nombreExtraits"`
}

// ExtraireConcepts extrait les concepts clés d'un cours via le LLM.
// Cette méthode extrait aussi les termes du lexique en un seul appel (stratégie d'optimisation tokens).
func (s *ServiceConcepts) ExtraireConcepts(ctx context.Context, coursID string) (*ResultatExtractionConcepts, error) {
	concepts, _, err := s.ExtraireConceptsEtTermes(ctx, coursID)
	if err != nil {
		return nil, err
	}
	return concepts, nil
}

// ExtraireConceptsEtTermes extrait les concepts ET les termes du lexique en un seul appel LLM.
func (s *ServiceConcepts) ExtraireConceptsEtTermes(ctx context.Context, coursID string) (*ResultatExtractionConcepts, *ResultatExtractionLexique, error) {
	if s.gestionnaireLLM == nil {
		return nil, nil, ErrServiceNonDisponible
	}

	// Récupérer le cours
	cours, err := s.coursRepo.ObtenirParID(ctx, coursID)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %s", ErrCoursNonTrouve, err.Error())
	}

	// Vérifier que le cours a du texte
	texte := cours.TexteCorrige
	if texte == "" {
		texte = cours.TexteOCR
	}
	if texte == "" {
		return nil, nil, ErrCoursVideOCR
	}

	// Construire le prompt combiné
	prompt := s.construirePromptConceptsEtTermes(texte)

	// Appeler le LLM (gpt-4o-mini suffit pour l'extraction)
	llmOptions := llm.OptionsGeneration{
		Modele:        "gpt-4o-mini",
		Temperature:   0.5,
		MaxTokens:     6000,
		FormatReponse: "json",
		SystemPrompt:  "Tu es un professeur expert en analyse de cours pour lycéens. Tu identifies les concepts clés et les termes importants. Tu réponds uniquement en JSON valide.",
	}

	reponseJSON, err := s.gestionnaireLLM.GenererJSON(ctx, prompt, nil, llmOptions)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %s", ErrGenerationLLMEchouee, err.Error())
	}

	// Parser la réponse combinée
	concepts, termes, err := s.parserReponseConceptsEtTermes(reponseJSON, coursID)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %s", ErrParsingConceptsEchoue, err.Error())
	}

	// Sauvegarder les concepts en base
	if s.conceptsRepo != nil {
		if err := s.conceptsRepo.SupprimerParCours(ctx, coursID); err != nil {
			return nil, nil, fmt.Errorf("erreur suppression anciens concepts: %w", err)
		}
		if err := s.conceptsRepo.CreerPlusieurs(ctx, concepts); err != nil {
			return nil, nil, fmt.Errorf("erreur sauvegarde concepts: %w", err)
		}
	}

	// Sauvegarder les termes en base
	if s.lexiqueRepo != nil {
		if err := s.lexiqueRepo.SupprimerParCours(ctx, coursID); err != nil {
			return nil, nil, fmt.Errorf("erreur suppression anciens termes: %w", err)
		}
		if err := s.lexiqueRepo.CreerPlusieurs(ctx, termes); err != nil {
			return nil, nil, fmt.Errorf("erreur sauvegarde termes: %w", err)
		}
	}

	return &ResultatExtractionConcepts{
			Concepts:       concepts,
			NombreExtraits: len(concepts),
		}, &ResultatExtractionLexique{
			Termes:         termes,
			NombreExtraits: len(termes),
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

// construirePromptConceptsEtTermes construit le prompt combiné pour extraire concepts et termes
func (s *ServiceConcepts) construirePromptConceptsEtTermes(texte string) string {
	return fmt.Sprintf(`Cours :
"""%s"""

Extrais en une seule passe :
1. 5-20 concepts clés (notions, formules, théorèmes), classés par importance
2. 10-30 termes de vocabulaire par ordre alphabétique

Pour chaque concept : nom court, définition (2-3 phrases), importance (essentiel/important/secondaire), position dans le texte.
Pour chaque terme : terme, définition (2-3 phrases), contexte, exemple, catégorie.
Basé UNIQUEMENT sur le contenu fourni.

JSON attendu :
{"concepts": [{"nom": "...", "definition": "...", "importance": "essentiel|important|secondaire", "position": {"debut": 0, "fin": 100}}], "termes": [{"terme": "...", "definition": "...", "contexte": "...", "exemple": "...", "categorie": "..."}]}`, texte)
}

// reponseConceptsEtTermesJSON représente la structure combinée concepts + termes
type reponseConceptsEtTermesJSON struct {
	Concepts []struct {
		Nom        string `json:"nom"`
		Definition string `json:"definition"`
		Importance string `json:"importance"`
		Position   *struct {
			Debut int `json:"debut"`
			Fin   int `json:"fin"`
		} `json:"position,omitempty"`
	} `json:"concepts"`
	Termes []struct {
		Terme      string `json:"terme"`
		Definition string `json:"definition"`
		Contexte   string `json:"contexte,omitempty"`
		Exemple    string `json:"exemple,omitempty"`
		Categorie  string `json:"categorie,omitempty"`
	} `json:"termes"`
}

// parserReponseConceptsEtTermes parse la réponse JSON combinée en concepts et termes
func (s *ServiceConcepts) parserReponseConceptsEtTermes(reponseJSON []byte, coursID string) ([]*store.Concept, []*store.TermeLexique, error) {
	var reponse reponseConceptsEtTermesJSON
	if err := json.Unmarshal(reponseJSON, &reponse); err != nil {
		return nil, nil, fmt.Errorf("erreur parsing JSON: %w", err)
	}

	// Parser les concepts
	concepts := make([]*store.Concept, 0, len(reponse.Concepts))
	for _, c := range reponse.Concepts {
		importance := c.Importance
		if importance != "essentiel" && importance != "important" && importance != "secondaire" {
			importance = "important"
		}

		concept := &store.Concept{
			CoursID:    coursID,
			Nom:        c.Nom,
			Definition: c.Definition,
			Importance: importance,
		}

		if c.Position != nil {
			concept.PositionDansCours = &store.PositionDansCours{
				Debut: c.Position.Debut,
				Fin:   c.Position.Fin,
			}
		}

		concepts = append(concepts, concept)
	}

	// Parser les termes
	termes := make([]*store.TermeLexique, 0, len(reponse.Termes))
	for _, t := range reponse.Termes {
		terme := &store.TermeLexique{
			CoursID:    coursID,
			Terme:      t.Terme,
			Definition: t.Definition,
			Contexte:   t.Contexte,
			Exemple:    t.Exemple,
			Categorie:  t.Categorie,
			Maitrise:   0,
		}
		termes = append(termes, terme)
	}

	return concepts, termes, nil
}
