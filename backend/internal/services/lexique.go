// Package services contient les services metier de l application
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/revisemieux/backend/internal/llm"
	"github.com/revisemieux/backend/internal/store"
)

// Codes d erreur lexique
var (
	ErrParsingLexiqueEchoue = &ErreurGeneration{Code: "PARSING_LEXIQUE_ECHOUE", Message: "Erreur lors du parsing des termes generes"}
	ErrTermeNonTrouve       = &ErreurGeneration{Code: "TERME_NON_TROUVE", Message: "Terme non trouve"}
	ErrMaitriseInvalide     = &ErreurGeneration{Code: "MAITRISE_INVALIDE", Message: "Niveau de maitrise invalide (0-5)"}
	ErrModeQuizInvalide     = &ErreurGeneration{Code: "MODE_QUIZ_INVALIDE", Message: "Mode de quiz invalide"}
)

// ServiceLexique gere l extraction et la gestion des termes du lexique
type ServiceLexique struct {
	gestionnaireLLM *llm.GestionnaireLLM
	coursRepo       store.CoursRepository
	lexiqueRepo     store.LexiqueRepository
}

// NouveauServiceLexique cree une nouvelle instance du service de lexique
func NouveauServiceLexique(
	gestionnaireLLM *llm.GestionnaireLLM,
	coursRepo store.CoursRepository,
	lexiqueRepo store.LexiqueRepository,
) *ServiceLexique {
	return &ServiceLexique{
		gestionnaireLLM: gestionnaireLLM,
		coursRepo:       coursRepo,
		lexiqueRepo:     lexiqueRepo,
	}
}

// ResultatExtractionLexique contient le resultat de l extraction
type ResultatExtractionLexique struct {
	Termes         []*store.TermeLexique `json:"termes"`
	NombreExtraits int                  `json:"nombreExtraits"`
}

// QuestionVocabulaire represente une question de quiz vocabulaire
type QuestionVocabulaire struct {
	ID              string   `json:"id"`
	Question        string   `json:"question"`
	ReponseAttendue string   `json:"reponseAttendue"`
	Indices         []string `json:"indices"`
	Terme           string   `json:"terme"`
}

// QuizVocabulaire represente un quiz de vocabulaire
type QuizVocabulaire struct {
	Mode      string                `json:"mode"`
	Questions []QuestionVocabulaire `json:"questions"`
}

// ExtraireTermes extrait les termes clés d'un cours.
// Vérifie d'abord si des termes existent déjà en base (sauvegardés par l'extraction combinée concepts+lexique).
// Si oui, les retourne directement sans appel LLM.
// Sinon, fait un appel LLM dédié.
func (s *ServiceLexique) ExtraireTermes(ctx context.Context, coursID string) (*ResultatExtractionLexique, error) {
	// Vérifier si des termes existent déjà (sauvegardés par ExtraireConceptsEtTermes)
	if s.lexiqueRepo != nil {
		termesExistants, err := s.lexiqueRepo.ListerParCours(ctx, coursID)
		if err == nil && len(termesExistants) > 0 {
			return &ResultatExtractionLexique{
				Termes:         termesExistants,
				NombreExtraits: len(termesExistants),
			}, nil
		}
	}

	// Aucun terme existant, faire un appel LLM
	if s.gestionnaireLLM == nil {
		return nil, ErrServiceNonDisponible
	}

	cours, err := s.coursRepo.ObtenirParID(ctx, coursID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrCoursNonTrouve, err.Error())
	}

	texte := cours.TexteCorrige
	if texte == "" {
		texte = cours.TexteOCR
	}
	if texte == "" {
		return nil, ErrCoursVideOCR
	}

	prompt := s.construirePromptLexique(texte)

	llmOptions := llm.OptionsGeneration{
		Modele:        "gpt-4o-mini",
		Temperature:   0.5,
		MaxTokens:     4000,
		FormatReponse: "json",
		SystemPrompt:  "Tu es un professeur expert en analyse de cours pour lyceens. Tu identifies les termes cles et leur definition. Tu reponds uniquement en JSON valide.",
	}

	reponseJSON, err := s.gestionnaireLLM.GenererJSON(ctx, prompt, nil, llmOptions)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrGenerationLLMEchouee, err.Error())
	}

	termes, err := s.parserReponseLexique(reponseJSON, coursID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrParsingLexiqueEchoue, err.Error())
	}

	if s.lexiqueRepo != nil {
		if err := s.lexiqueRepo.SupprimerParCours(ctx, coursID); err != nil {
			return nil, fmt.Errorf("erreur suppression anciens termes: %w", err)
		}
		if err := s.lexiqueRepo.CreerPlusieurs(ctx, termes); err != nil {
			return nil, fmt.Errorf("erreur sauvegarde termes: %w", err)
		}
	}

	return &ResultatExtractionLexique{
		Termes:         termes,
		NombreExtraits: len(termes),
	}, nil
}

// GetTermesByCours recupere les termes existants d un cours
func (s *ServiceLexique) GetTermesByCours(ctx context.Context, coursID string) ([]*store.TermeLexique, error) {
	if s.lexiqueRepo == nil {
		return nil, ErrServiceNonDisponible
	}

	return s.lexiqueRepo.ListerParCours(ctx, coursID)
}

// MettreAJourMaitrise met a jour le niveau de maitrise d un terme
func (s *ServiceLexique) MettreAJourMaitrise(ctx context.Context, id string, maitrise int) error {
	if s.lexiqueRepo == nil {
		return ErrServiceNonDisponible
	}

	if maitrise < 0 || maitrise > 5 {
		return ErrMaitriseInvalide
	}

	if err := s.lexiqueRepo.MettreAJourMaitrise(ctx, id, maitrise); err != nil {
		return fmt.Errorf("%w: %s", ErrTermeNonTrouve, err.Error())
	}

	return nil
}

// GenererQuizVocabulaire genere un quiz de vocabulaire a partir des termes d un cours
func (s *ServiceLexique) GenererQuizVocabulaire(ctx context.Context, coursID string, mode string) (*QuizVocabulaire, error) {
	if mode != "terme_vers_definition" && mode != "definition_vers_terme" {
		return nil, ErrModeQuizInvalide
	}

	if s.lexiqueRepo == nil {
		return nil, ErrServiceNonDisponible
	}

	termes, err := s.lexiqueRepo.ListerParCours(ctx, coursID)
	if err != nil {
		return nil, fmt.Errorf("erreur recuperation termes: %w", err)
	}

	if len(termes) == 0 {
		return nil, fmt.Errorf("%w: aucun terme disponible pour ce cours", ErrCoursVideOCR)
	}

	// Melanger les termes
	shuffled := make([]*store.TermeLexique, len(termes))
	copy(shuffled, termes)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	// Limiter a 10 questions maximum
	maxQuestions := 10
	if len(shuffled) < maxQuestions {
		maxQuestions = len(shuffled)
	}
	shuffled = shuffled[:maxQuestions]

	questions := make([]QuestionVocabulaire, 0, len(shuffled))
	for i, terme := range shuffled {
		q := QuestionVocabulaire{
			ID:    fmt.Sprintf("q%d", i+1),
			Terme: terme.Terme,
		}

		if mode == "terme_vers_definition" {
			q.Question = terme.Terme
			q.ReponseAttendue = terme.Definition
			q.Indices = genererIndicesDefinition(terme)
		} else {
			q.Question = terme.Definition
			q.ReponseAttendue = terme.Terme
			q.Indices = genererIndicesTerme(terme)
		}

		questions = append(questions, q)
	}

	return &QuizVocabulaire{
		Mode:      mode,
		Questions: questions,
	}, nil
}

// construirePromptLexique construit le prompt pour l extraction de termes
func (s *ServiceLexique) construirePromptLexique(texte string) string {
	return fmt.Sprintf(`Cours :
"""%s"""

Identifie 10-30 termes cles par ordre alphabetique.
Pour chaque terme : terme, definition (2-3 phrases), contexte, exemple, categorie.
Base UNIQUEMENT sur le contenu fourni.

JSON attendu :
{"termes": [{"terme": "...", "definition": "...", "contexte": "...", "exemple": "...", "categorie": "..."}]}`, texte)
}

// reponseLexiqueJSON represente la structure de reponse du LLM pour le lexique
type reponseLexiqueJSON struct {
	Termes []struct {
		Terme      string `json:"terme"`
		Definition string `json:"definition"`
		Contexte   string `json:"contexte,omitempty"`
		Exemple    string `json:"exemple,omitempty"`
		Categorie  string `json:"categorie,omitempty"`
	} `json:"termes"`
}

// parserReponseLexique parse la reponse JSON du LLM en termes
func (s *ServiceLexique) parserReponseLexique(reponseJSON []byte, coursID string) ([]*store.TermeLexique, error) {
	var reponse reponseLexiqueJSON
	if err := json.Unmarshal(reponseJSON, &reponse); err != nil {
		return nil, fmt.Errorf("erreur parsing JSON: %w", err)
	}

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

	return termes, nil
}

// genererIndicesDefinition genere des indices progressifs pour deviner la definition
func genererIndicesDefinition(terme *store.TermeLexique) []string {
	indices := []string{}

	if terme.Categorie != "" {
		indices = append(indices, fmt.Sprintf("Categorie : %s", terme.Categorie))
	}

	if terme.Contexte != "" {
		indices = append(indices, fmt.Sprintf("Contexte : %s", terme.Contexte))
	}

	if len(terme.Definition) > 20 {
		premiersMots := terme.Definition
		if len(premiersMots) > 30 {
			premiersMots = premiersMots[:30] + "..."
		}
		indices = append(indices, fmt.Sprintf("Debut de la definition : %s", premiersMots))
	}

	if terme.Exemple != "" {
		indices = append(indices, fmt.Sprintf("Exemple : %s", terme.Exemple))
	}

	return indices
}

// genererIndicesTerme genere des indices progressifs pour deviner le terme
func genererIndicesTerme(terme *store.TermeLexique) []string {
	indices := []string{}

	if terme.Categorie != "" {
		indices = append(indices, fmt.Sprintf("Categorie : %s", terme.Categorie))
	}

	indices = append(indices, fmt.Sprintf("Le terme contient %d lettres", len([]rune(terme.Terme))))

	runes := []rune(terme.Terme)
	if len(runes) > 0 {
		indices = append(indices, fmt.Sprintf("Le terme commence par la lettre : %c", runes[0]))
	}

	if terme.Exemple != "" {
		indices = append(indices, fmt.Sprintf("Exemple : %s", terme.Exemple))
	}

	return indices
}
