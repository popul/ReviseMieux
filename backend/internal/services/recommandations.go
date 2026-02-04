// Package services contient les services métier de l'application
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/revisemieux/backend/internal/llm"
	"github.com/revisemieux/backend/internal/store"
)

// SourceRecommandation indique la source des données pour générer les recommandations
type SourceRecommandation struct {
	// CopieID pour générer des recommandations basées sur l'analyse d'une copie
	CopieID string
	// Matiere pour filtrer les recommandations par matière
	Matiere string
	// IncludeQuizResults pour inclure les résultats de quiz dans l'analyse
	IncludeQuizResults bool
}

// Recommandation représente une suggestion de révision personnalisée
type Recommandation struct {
	Domaine        string `json:"domaine"`        // "équations du second degré"
	Raison         string `json:"raison"`         // "3 erreurs de compréhension"
	SeveriteMax    string `json:"severiteMax"`    // "grave", "moderate", "legere"
	ActionSuggerie string `json:"actionSuggerie"` // "Revoir la section 3.2 du cours"
	TypeQuiz       string `json:"typeQuiz"`       // Suggestion de type de quiz ciblé
	Priorite       int    `json:"priorite"`       // 1-5, 1 = plus prioritaire
}

// ResultatRecommandations contient les recommandations générées
type ResultatRecommandations struct {
	Recommandations   []*Recommandation `json:"recommandations"`
	Resume            string            `json:"resume"`
	PlanAction        string            `json:"planAction"`
	ProchainQuiz      string            `json:"prochainQuiz,omitempty"`
	Motivation        string            `json:"motivation"`
	NombreRecommandations int           `json:"nombreRecommandations"`
}

// DonneesPourRecommandations agrège les données nécessaires à la génération
type DonneesPourRecommandations struct {
	// Erreurs d'analyse (si copie analysée)
	Erreurs       []*store.ErreurAnalyse
	ResumeErreurs map[string]int // type → count

	// Statistiques quiz
	QuizFaibles []*HistoriqueQuiz // quiz avec score < 60%

	// Contexte
	Matiere string
	Titre   string

	// Points identifiés dans l'analyse
	PointsAAmeliorer []string
}

// Codes d'erreur pour les recommandations
var (
	ErrDonneesInsuffisantes = &ErreurGeneration{Code: "DONNEES_INSUFFISANTES", Message: "Pas assez de données pour générer des recommandations"}
	ErrRecommandationEchouee = &ErreurGeneration{Code: "RECOMMANDATION_ECHOUEE", Message: "Échec de la génération des recommandations"}
)

// ServiceRecommandations gère la génération de recommandations personnalisées
type ServiceRecommandations struct {
	gestionnaireLLM *llm.GestionnaireLLM
	copieRepo       store.CopieExamenRepository
	erreurRepo      store.ErreurAnalyseRepository
	coursRepo       store.CoursRepository
	quizRepo        store.QuizRepository
}

// NouveauServiceRecommandations crée une nouvelle instance du service
func NouveauServiceRecommandations(
	gestionnaireLLM *llm.GestionnaireLLM,
	copieRepo store.CopieExamenRepository,
	erreurRepo store.ErreurAnalyseRepository,
	coursRepo store.CoursRepository,
	quizRepo store.QuizRepository,
) *ServiceRecommandations {
	return &ServiceRecommandations{
		gestionnaireLLM: gestionnaireLLM,
		copieRepo:       copieRepo,
		erreurRepo:      erreurRepo,
		coursRepo:       coursRepo,
		quizRepo:        quizRepo,
	}
}

// GenererRecommandations génère des recommandations personnalisées
func (s *ServiceRecommandations) GenererRecommandations(ctx context.Context, source *SourceRecommandation) (*ResultatRecommandations, error) {
	if s.gestionnaireLLM == nil {
		return nil, ErrServiceNonDisponible
	}

	if source == nil {
		source = &SourceRecommandation{IncludeQuizResults: true}
	}

	// Collecter les données
	donnees, err := s.collecterDonnees(ctx, source)
	if err != nil {
		return nil, err
	}

	// Vérifier qu'on a assez de données
	if len(donnees.Erreurs) == 0 && len(donnees.QuizFaibles) == 0 && len(donnees.PointsAAmeliorer) == 0 {
		return nil, ErrDonneesInsuffisantes
	}

	// Construire le prompt
	prompt := s.construirePromptRecommandations(donnees)

	// Appeler le LLM
	llmOptions := llm.OptionsGeneration{
		Temperature:   0.5, // Équilibré entre analytique et contextuel
		MaxTokens:     3000,
		FormatReponse: "json",
		SystemPrompt:  "Tu es un tuteur pédagogique expert et bienveillant. Tu analyses les lacunes d'un élève et proposes des recommandations ciblées, motivantes et actionnables. Tu réponds uniquement en JSON valide.",
	}

	reponseJSON, err := s.gestionnaireLLM.GenererJSON(ctx, prompt, nil, llmOptions)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrRecommandationEchouee, err.Error())
	}

	// Parser la réponse
	resultat, err := s.parserReponseRecommandations(reponseJSON)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrParsingAnalyse, err.Error())
	}

	return resultat, nil
}

// collecterDonnees rassemble les données pour la génération
func (s *ServiceRecommandations) collecterDonnees(ctx context.Context, source *SourceRecommandation) (*DonneesPourRecommandations, error) {
	donnees := &DonneesPourRecommandations{
		ResumeErreurs:    make(map[string]int),
		QuizFaibles:      []*HistoriqueQuiz{},
		PointsAAmeliorer: []string{},
	}

	// Si une copie est spécifiée, récupérer ses erreurs
	if source.CopieID != "" && s.erreurRepo != nil && s.copieRepo != nil {
		// Récupérer la copie pour le contexte
		copie, err := s.copieRepo.ObtenirParID(ctx, source.CopieID)
		if err == nil && copie != nil {
			donnees.Matiere = copie.Matiere
			donnees.Titre = copie.Titre
		}

		// Récupérer les erreurs
		erreurs, err := s.erreurRepo.ListerParCopie(ctx, source.CopieID)
		if err == nil {
			donnees.Erreurs = erreurs

			// Compter par type
			for _, e := range erreurs {
				donnees.ResumeErreurs[e.TypeErreur]++
			}

			// Extraire les points à améliorer (erreurs graves ou modérées)
			for _, e := range erreurs {
				if e.Severite == "grave" || e.Severite == "moderate" {
					if e.Explication != "" {
						donnees.PointsAAmeliorer = append(donnees.PointsAAmeliorer, e.Explication)
					}
				}
			}
		}
	}

	// Récupérer les quiz faibles si demandé
	if source.IncludeQuizResults && s.quizRepo != nil {
		sessions, err := s.quizRepo.ListerSessionsCompletes(ctx, 20)
		if err == nil {
			for _, session := range sessions {
				// Filtrer par matière si spécifiée
				if source.Matiere != "" && session.Matiere != source.Matiere {
					continue
				}

				// Garder les quiz avec score < 60%
				if session.Score < 60 {
					donnees.QuizFaibles = append(donnees.QuizFaibles, &HistoriqueQuiz{
						SessionID:  session.SessionID,
						QuizID:     session.QuizID,
						QuizTitre:  session.QuizTitre,
						CoursID:    session.CoursID,
						CoursTitre: session.CoursTitre,
						Matiere:    session.Matiere,
						Score:      session.Score,
						DateFin:    session.DateFin,
					})
				}
			}

			// Trier par score croissant (pires d'abord)
			sort.Slice(donnees.QuizFaibles, func(i, j int) bool {
				return donnees.QuizFaibles[i].Score < donnees.QuizFaibles[j].Score
			})

			// Limiter à 5 quiz faibles
			if len(donnees.QuizFaibles) > 5 {
				donnees.QuizFaibles = donnees.QuizFaibles[:5]
			}
		}
	}

	return donnees, nil
}

// construirePromptRecommandations construit le prompt pour le LLM
func (s *ServiceRecommandations) construirePromptRecommandations(donnees *DonneesPourRecommandations) string {
	// Section contexte
	contexteSection := ""
	if donnees.Matiere != "" {
		contexteSection = fmt.Sprintf("Matière : %s\n", donnees.Matiere)
	}
	if donnees.Titre != "" {
		contexteSection += fmt.Sprintf("Évaluation analysée : %s\n", donnees.Titre)
	}

	// Section erreurs
	erreursSection := ""
	if len(donnees.Erreurs) > 0 {
		erreursSection = fmt.Sprintf(`
Erreurs identifiées dans la copie (%d au total) :
- Erreurs de compréhension : %d
- Erreurs de méthode : %d
- Erreurs d'inattention : %d

Détail des erreurs principales :
`, len(donnees.Erreurs),
			donnees.ResumeErreurs["comprehension"],
			donnees.ResumeErreurs["methode"],
			donnees.ResumeErreurs["inattention"])

		// Ajouter les 5 premières erreurs graves/modérées
		count := 0
		for _, e := range donnees.Erreurs {
			if count >= 5 {
				break
			}
			if e.Severite == "grave" || e.Severite == "moderate" {
				erreursSection += fmt.Sprintf("- [%s/%s] %s\n", e.TypeErreur, e.Severite, e.Explication)
				count++
			}
		}
	}

	// Section quiz
	quizSection := ""
	if len(donnees.QuizFaibles) > 0 {
		quizSection = "\nQuiz avec score insuffisant (< 60%) :\n"
		for _, q := range donnees.QuizFaibles {
			quizSection += fmt.Sprintf("- %s (%s) : %.0f%%\n", q.QuizTitre, q.Matiere, q.Score)
		}
	}

	// Section points à améliorer
	pointsSection := ""
	if len(donnees.PointsAAmeliorer) > 0 {
		pointsSection = "\nPoints à améliorer identifiés :\n"
		for i, p := range donnees.PointsAAmeliorer {
			if i >= 5 {
				break
			}
			pointsSection += fmt.Sprintf("- %s\n", p)
		}
	}

	return fmt.Sprintf(`Tu es un tuteur pédagogique expert. Basé sur l'analyse des erreurs et résultats d'un élève, propose des recommandations ciblées et motivantes.

%s%s%s%s
Instructions :
1. Identifie les 3 à 5 lacunes PRINCIPALES (pas plus de 5)
2. Pour chaque lacune, propose :
   - Le domaine exact à réviser
   - La raison (basée sur les erreurs ou quiz)
   - Une action concrète à réaliser
   - Une priorité de 1 (urgent) à 5 (secondaire)
3. Écris un résumé concis de la situation (1-2 phrases)
4. Propose un plan d'action global (ce que l'élève devrait faire cette semaine)
5. Suggère quel type de quiz faire en premier
6. Ajoute une phrase de motivation personnalisée

Réponds UNIQUEMENT avec un JSON valide au format suivant :
{
  "recommandations": [
    {
      "domaine": "Nom du concept/domaine à réviser",
      "raison": "Pourquoi c'est une lacune (ex: 3 erreurs de compréhension)",
      "severiteMax": "grave|moderate|legere",
      "actionSuggerie": "Action concrète (ex: Revoir la leçon X, faire 5 exercices sur Y)",
      "typeQuiz": "Type de quiz suggéré pour s'entraîner",
      "priorite": 1
    }
  ],
  "resume": "Résumé concis de l'analyse des lacunes",
  "planAction": "Plan d'action détaillé pour la semaine (2-3 phrases)",
  "prochainQuiz": "Description du prochain quiz à passer",
  "motivation": "Message d'encouragement personnalisé"
}`, contexteSection, erreursSection, quizSection, pointsSection)
}

// reponseRecommandationsJSON représente la structure de réponse du LLM
type reponseRecommandationsJSON struct {
	Recommandations []struct {
		Domaine        string `json:"domaine"`
		Raison         string `json:"raison"`
		SeveriteMax    string `json:"severiteMax"`
		ActionSuggerie string `json:"actionSuggerie"`
		TypeQuiz       string `json:"typeQuiz"`
		Priorite       int    `json:"priorite"`
	} `json:"recommandations"`
	Resume       string `json:"resume"`
	PlanAction   string `json:"planAction"`
	ProchainQuiz string `json:"prochainQuiz"`
	Motivation   string `json:"motivation"`
}

// parserReponseRecommandations parse la réponse JSON du LLM
func (s *ServiceRecommandations) parserReponseRecommandations(reponseJSON []byte) (*ResultatRecommandations, error) {
	var reponse reponseRecommandationsJSON
	if err := json.Unmarshal(reponseJSON, &reponse); err != nil {
		return nil, fmt.Errorf("erreur parsing JSON: %w", err)
	}

	// Convertir les recommandations
	recommandations := make([]*Recommandation, 0, len(reponse.Recommandations))
	for _, r := range reponse.Recommandations {
		// Valider la sévérité
		severite := r.SeveriteMax
		if severite != "grave" && severite != "moderate" && severite != "legere" {
			severite = "moderate"
		}

		// Valider la priorité (1-5)
		priorite := r.Priorite
		if priorite < 1 {
			priorite = 1
		} else if priorite > 5 {
			priorite = 5
		}

		recommandations = append(recommandations, &Recommandation{
			Domaine:        r.Domaine,
			Raison:         r.Raison,
			SeveriteMax:    severite,
			ActionSuggerie: r.ActionSuggerie,
			TypeQuiz:       r.TypeQuiz,
			Priorite:       priorite,
		})
	}

	// Trier par priorité
	sort.Slice(recommandations, func(i, j int) bool {
		return recommandations[i].Priorite < recommandations[j].Priorite
	})

	return &ResultatRecommandations{
		Recommandations:       recommandations,
		Resume:                reponse.Resume,
		PlanAction:            reponse.PlanAction,
		ProchainQuiz:          reponse.ProchainQuiz,
		Motivation:            reponse.Motivation,
		NombreRecommandations: len(recommandations),
	}, nil
}

// GenererRecommandationsPrioritaires génère les recommandations les plus urgentes (top 3)
func (s *ServiceRecommandations) GenererRecommandationsPrioritaires(ctx context.Context, matiere string) (*ResultatRecommandations, error) {
	resultat, err := s.GenererRecommandations(ctx, &SourceRecommandation{
		Matiere:            matiere,
		IncludeQuizResults: true,
	})
	if err != nil {
		return nil, err
	}

	// Limiter à 3 recommandations
	if len(resultat.Recommandations) > 3 {
		resultat.Recommandations = resultat.Recommandations[:3]
		resultat.NombreRecommandations = 3
	}

	return resultat, nil
}
