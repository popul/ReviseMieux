// Package services contient les services métier de l'application
package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/revisemieux/backend/internal/llm"
	"github.com/revisemieux/backend/internal/store"
)

// OptionsAnalyseErreurs contient les options pour l'analyse des erreurs
type OptionsAnalyseErreurs struct {
	// InclusAnnotations inclut les annotations du professeur dans l'analyse
	InclusAnnotations bool
	// CoursID optionnel pour contextualiser l'analyse avec le contenu du cours
	CoursID string
}

// ResultatAnalyseErreurs contient le résultat de l'analyse des erreurs
type ResultatAnalyseErreurs struct {
	Erreurs          []*store.ErreurAnalyse `json:"erreurs"`
	NombreErreurs    int                    `json:"nombreErreurs"`
	ResumeParType    map[string]int         `json:"resumeParType"`
	ConseilGlobal    string                 `json:"conseilGlobal"`
	PointsForts      []string               `json:"pointsForts,omitempty"`
	PointsAAmeliorer []string               `json:"pointsAAmeliorer,omitempty"`
}

// Codes d'erreur pour l'analyse
var (
	ErrCopieNonTrouvee    = &ErreurGeneration{Code: "COPIE_NON_TROUVEE", Message: "Copie d'examen non trouvée"}
	ErrCopieVideOCR       = &ErreurGeneration{Code: "COPIE_VIDE", Message: "La copie n'a pas de texte OCR"}
	ErrAnalyseEchouee     = &ErreurGeneration{Code: "ANALYSE_ECHOUEE", Message: "Échec de l'analyse par le LLM"}
	ErrParsingAnalyse     = &ErreurGeneration{Code: "PARSING_ANALYSE_ECHOUE", Message: "Erreur lors du parsing de l'analyse"}
)

// ServiceAnalyseErreurs gère l'analyse des erreurs dans les copies d'examens
type ServiceAnalyseErreurs struct {
	gestionnaireLLM *llm.GestionnaireLLM
	copieRepo       store.CopieExamenRepository
	erreurRepo      store.ErreurAnalyseRepository
	coursRepo       store.CoursRepository
}

// NouveauServiceAnalyseErreurs crée une nouvelle instance du service d'analyse
func NouveauServiceAnalyseErreurs(
	gestionnaireLLM *llm.GestionnaireLLM,
	copieRepo store.CopieExamenRepository,
	erreurRepo store.ErreurAnalyseRepository,
	coursRepo store.CoursRepository,
) *ServiceAnalyseErreurs {
	return &ServiceAnalyseErreurs{
		gestionnaireLLM: gestionnaireLLM,
		copieRepo:       copieRepo,
		erreurRepo:      erreurRepo,
		coursRepo:       coursRepo,
	}
}

// AnalyserCopie analyse les erreurs dans une copie d'examen
func (s *ServiceAnalyseErreurs) AnalyserCopie(ctx context.Context, copieID string, options *OptionsAnalyseErreurs) (*ResultatAnalyseErreurs, error) {
	if s.gestionnaireLLM == nil {
		return nil, ErrServiceNonDisponible
	}

	// Récupérer la copie
	copie, err := s.copieRepo.ObtenirParID(ctx, copieID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrCopieNonTrouvee, err.Error())
	}

	// Vérifier que la copie a du texte
	if copie.TexteOCR == "" {
		return nil, ErrCopieVideOCR
	}

	// Appliquer les valeurs par défaut
	if options == nil {
		options = &OptionsAnalyseErreurs{
			InclusAnnotations: true,
		}
	}

	// Récupérer le cours si spécifié
	var contexteCours string
	if options.CoursID != "" || copie.CoursID != "" {
		coursID := options.CoursID
		if coursID == "" {
			coursID = copie.CoursID
		}
		if coursID != "" {
			cours, err := s.coursRepo.ObtenirParID(ctx, coursID)
			if err == nil && cours != nil {
				texte := cours.TexteCorrige
				if texte == "" {
					texte = cours.TexteOCR
				}
				if texte != "" {
					contexteCours = texte
				}
			}
		}
	}

	// Construire le prompt
	prompt := s.construirePromptAnalyse(copie, contexteCours, options)

	// Appeler le LLM
	llmOptions := llm.OptionsGeneration{
		Temperature:   0.3, // Plus déterministe pour l'analyse
		MaxTokens:     4000,
		FormatReponse: "json",
		SystemPrompt:  "Tu es un correcteur expert qui analyse les copies d'examens pour identifier les erreurs et aider les élèves à progresser. Tu classes les erreurs par type et proposes des conseils personnalisés. Tu réponds uniquement en JSON valide.",
	}

	reponseJSON, err := s.gestionnaireLLM.GenererJSON(ctx, prompt, nil, llmOptions)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrAnalyseEchouee, err.Error())
	}

	// Parser la réponse
	resultat, err := s.parserReponseAnalyse(reponseJSON, copieID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrParsingAnalyse, err.Error())
	}

	// Supprimer les anciennes erreurs et sauvegarder les nouvelles
	if s.erreurRepo != nil {
		if err := s.erreurRepo.SupprimerParCopie(ctx, copieID); err != nil {
			return nil, fmt.Errorf("erreur suppression anciennes erreurs: %w", err)
		}
		if len(resultat.Erreurs) > 0 {
			if err := s.erreurRepo.CreerPlusieurs(ctx, resultat.Erreurs); err != nil {
				return nil, fmt.Errorf("erreur sauvegarde erreurs: %w", err)
			}
		}
	}

	return resultat, nil
}

// construirePromptAnalyse construit le prompt pour l'analyse des erreurs
func (s *ServiceAnalyseErreurs) construirePromptAnalyse(copie *store.CopieExamen, contexteCours string, options *OptionsAnalyseErreurs) string {
	// Informations sur la copie
	matiereInfo := ""
	if copie.Matiere != "" {
		matiereInfo = fmt.Sprintf("Matière : %s\n", copie.Matiere)
	}

	noteInfo := ""
	if copie.NoteObtenue != nil && copie.NoteTotale != nil {
		noteInfo = fmt.Sprintf("Note obtenue : %.1f / %.1f\n", *copie.NoteObtenue, *copie.NoteTotale)
	}

	// Annotations du professeur
	annotationsSection := ""
	if options.InclusAnnotations && copie.AnnotationsProfesseur != "" {
		annotationsSection = fmt.Sprintf(`
Annotations du professeur (corrections et commentaires) :
"""
%s
"""

`, copie.AnnotationsProfesseur)
	}

	// Contexte du cours
	coursSection := ""
	if contexteCours != "" {
		// Limiter le cours pour ne pas dépasser le contexte
		if len(contexteCours) > 3000 {
			contexteCours = contexteCours[:3000] + "..."
		}
		coursSection = fmt.Sprintf(`
Contenu du cours de référence :
"""
%s
"""

`, contexteCours)
	}

	return fmt.Sprintf(`Tu es un correcteur expert qui analyse une copie d'examen pour identifier les erreurs et aider l'élève à progresser.

%s%sTitre de l'évaluation : %s

Copie de l'élève (texte extrait par OCR) :
"""
%s
"""
%s%s
Instructions :
1. Analyse attentivement la copie et identifie TOUTES les erreurs
2. Classe chaque erreur dans l'un des 3 types :
   - "comprehension" : l'élève n'a pas compris le concept (mauvaise définition, contresens, confusion de notions)
   - "methode" : l'élève a compris mais applique mal (erreur de calcul, mauvais raisonnement, étapes manquantes)
   - "inattention" : faute d'étourderie (orthographe, oubli d'unité, erreur de recopie)
3. Attribue une sévérité à chaque erreur :
   - "legere" : erreur mineure qui n'affecte pas beaucoup la note
   - "moderate" : erreur qui impacte la note
   - "grave" : erreur fondamentale qui montre une lacune importante
4. Pour chaque erreur, fournis :
   - Le texte original erroné (si identifiable)
   - La correction
   - Une explication pédagogique claire
   - Un conseil pour ne plus faire cette erreur
5. Identifie les points forts de la copie
6. Propose un conseil global pour progresser

Réponds UNIQUEMENT avec un JSON valide au format suivant, sans texte avant ou après :
{
  "erreurs": [
    {
      "type_erreur": "comprehension|methode|inattention",
      "texte_original": "texte de l'élève (si identifiable)",
      "correction": "ce qu'il aurait fallu écrire",
      "explication": "pourquoi c'est une erreur et ce qu'il faut comprendre",
      "conseil": "comment éviter cette erreur à l'avenir",
      "severite": "legere|moderate|grave"
    }
  ],
  "points_forts": ["point fort 1", "point fort 2"],
  "points_a_ameliorer": ["point à améliorer 1", "point à améliorer 2"],
  "conseil_global": "Conseil général pour progresser basé sur l'analyse globale de la copie"
}`, matiereInfo, noteInfo, copie.Titre, copie.TexteOCR, annotationsSection, coursSection)
}

// reponseAnalyseJSON représente la structure de réponse du LLM
type reponseAnalyseJSON struct {
	Erreurs []struct {
		TypeErreur    string `json:"type_erreur"`
		TexteOriginal string `json:"texte_original"`
		Correction    string `json:"correction"`
		Explication   string `json:"explication"`
		Conseil       string `json:"conseil"`
		Severite      string `json:"severite"`
	} `json:"erreurs"`
	PointsForts      []string `json:"points_forts"`
	PointsAAmeliorer []string `json:"points_a_ameliorer"`
	ConseilGlobal    string   `json:"conseil_global"`
}

// parserReponseAnalyse parse la réponse JSON du LLM
func (s *ServiceAnalyseErreurs) parserReponseAnalyse(reponseJSON []byte, copieID string) (*ResultatAnalyseErreurs, error) {
	var reponse reponseAnalyseJSON
	if err := json.Unmarshal(reponseJSON, &reponse); err != nil {
		return nil, fmt.Errorf("erreur parsing JSON: %w", err)
	}

	// Convertir les erreurs
	erreurs := make([]*store.ErreurAnalyse, 0, len(reponse.Erreurs))
	resumeParType := map[string]int{
		"comprehension": 0,
		"methode":       0,
		"inattention":   0,
	}

	for _, e := range reponse.Erreurs {
		// Valider le type d'erreur
		typeErreur := e.TypeErreur
		if typeErreur != "comprehension" && typeErreur != "methode" && typeErreur != "inattention" {
			typeErreur = "methode" // valeur par défaut
		}

		// Valider la sévérité
		severite := e.Severite
		if severite != "legere" && severite != "moderate" && severite != "grave" {
			severite = "moderate" // valeur par défaut
		}

		erreur := &store.ErreurAnalyse{
			CopieID:       copieID,
			TypeErreur:    typeErreur,
			TexteOriginal: e.TexteOriginal,
			Correction:    e.Correction,
			Explication:   e.Explication,
			Conseil:       e.Conseil,
			Severite:      severite,
		}
		erreurs = append(erreurs, erreur)
		resumeParType[typeErreur]++
	}

	return &ResultatAnalyseErreurs{
		Erreurs:          erreurs,
		NombreErreurs:    len(erreurs),
		ResumeParType:    resumeParType,
		ConseilGlobal:    reponse.ConseilGlobal,
		PointsForts:      reponse.PointsForts,
		PointsAAmeliorer: reponse.PointsAAmeliorer,
	}, nil
}

// ObtenirErreursParCopie récupère les erreurs d'analyse existantes
func (s *ServiceAnalyseErreurs) ObtenirErreursParCopie(ctx context.Context, copieID string) ([]*store.ErreurAnalyse, error) {
	if s.erreurRepo == nil {
		return nil, ErrServiceNonDisponible
	}

	return s.erreurRepo.ListerParCopie(ctx, copieID)
}

// CompterErreursParType retourne les statistiques d'erreurs par type
func (s *ServiceAnalyseErreurs) CompterErreursParType(ctx context.Context, copieID string) (map[string]int, error) {
	if s.erreurRepo == nil {
		return nil, ErrServiceNonDisponible
	}

	return s.erreurRepo.CompterParType(ctx, copieID)
}
