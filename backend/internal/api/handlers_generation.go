// Package api contient les handlers HTTP
package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/revisemieux/backend/internal/llm"
	"github.com/revisemieux/backend/internal/services"
)

// HandlersGeneration gère les endpoints de génération de contenu
type HandlersGeneration struct {
	serviceGeneration *services.ServiceGeneration
}

// NouveauHandlersGeneration crée une nouvelle instance des handlers de génération
func NouveauHandlersGeneration(serviceGeneration *services.ServiceGeneration) *HandlersGeneration {
	return &HandlersGeneration{
		serviceGeneration: serviceGeneration,
	}
}

// RequeteGenererFiches représente la requête pour générer des fiches
type RequeteGenererFiches struct {
	CoursID      string `json:"coursId" binding:"required"`
	NombreFiches int    `json:"nombreFiches,omitempty"`
	Difficulte   string `json:"difficulte,omitempty"`
}

// ReponseFiches représente la réponse de génération de fiches
type ReponseFiches struct {
	Succes       bool                 `json:"succes"`
	Fiches       []FicheReponse       `json:"fiches,omitempty"`
	NombreGenere int                  `json:"nombreGenere,omitempty"`
	Erreur       *ErreurReponse       `json:"erreur,omitempty"`
}

// FicheReponse représente une fiche dans la réponse API
type FicheReponse struct {
	ID         string `json:"id"`
	Question   string `json:"question"`
	Reponse    string `json:"reponse"`
	Difficulte string `json:"difficulte"`
	Ordre      int    `json:"ordre"`
}

// GenererFichesHandler génère des fiches de révision pour un cours
func (h *HandlersGeneration) GenererFichesHandler(c *gin.Context) {
	// Vérifier que le service est disponible
	if h.serviceGeneration == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseFiches{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service de génération n'est pas configuré",
			},
		})
		return
	}

	// Parser la requête
	var req RequeteGenererFiches
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ReponseFiches{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "REQUETE_INVALIDE",
				Message: "Le champ coursId est requis",
			},
		})
		return
	}

	// Valider la difficulté si fournie
	if req.Difficulte != "" && req.Difficulte != "facile" && req.Difficulte != "moyen" && req.Difficulte != "difficile" {
		c.JSON(http.StatusBadRequest, ReponseFiches{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "DIFFICULTE_INVALIDE",
				Message: "La difficulté doit être: facile, moyen ou difficile",
			},
		})
		return
	}

	// Préparer les options
	options := &services.OptionsGenerationFiches{
		NombreFiches: req.NombreFiches,
		Difficulte:   req.Difficulte,
	}

	// Générer les fiches
	resultat, err := h.serviceGeneration.GenererFiches(c.Request.Context(), req.CoursID, options)
	if err != nil {
		h.gererErreurGeneration(c, err)
		return
	}

	// Convertir les fiches pour la réponse
	fichesReponse := make([]FicheReponse, 0, len(resultat.Fiches))
	for _, f := range resultat.Fiches {
		fichesReponse = append(fichesReponse, FicheReponse{
			ID:         f.ID,
			Question:   f.Question,
			Reponse:    f.Reponse,
			Difficulte: f.Difficulte,
			Ordre:      f.Ordre,
		})
	}

	c.JSON(http.StatusOK, ReponseFiches{
		Succes:       true,
		Fiches:       fichesReponse,
		NombreGenere: resultat.NombreGenere,
	})
}

// ObtenirFichesHandler récupère les fiches existantes d'un cours
func (h *HandlersGeneration) ObtenirFichesHandler(c *gin.Context) {
	coursID := c.Param("id")
	if coursID == "" {
		c.JSON(http.StatusBadRequest, ReponseFiches{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "ID_MANQUANT",
				Message: "L'ID du cours est requis",
			},
		})
		return
	}

	if h.serviceGeneration == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseFiches{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service n'est pas configuré",
			},
		})
		return
	}

	fiches, err := h.serviceGeneration.ObtenirFichesParCours(c.Request.Context(), coursID)
	if err != nil {
		h.gererErreurGeneration(c, err)
		return
	}

	// Convertir les fiches pour la réponse
	fichesReponse := make([]FicheReponse, 0, len(fiches))
	for _, f := range fiches {
		fichesReponse = append(fichesReponse, FicheReponse{
			ID:         f.ID,
			Question:   f.Question,
			Reponse:    f.Reponse,
			Difficulte: f.Difficulte,
			Ordre:      f.Ordre,
		})
	}

	c.JSON(http.StatusOK, ReponseFiches{
		Succes:       true,
		Fiches:       fichesReponse,
		NombreGenere: len(fichesReponse),
	})
}

// --- Quiz Handlers ---

// RequeteGenererQuiz représente la requête pour générer un quiz
type RequeteGenererQuiz struct {
	CoursID         string `json:"coursId" binding:"required"`
	NombreQuestions int    `json:"nombreQuestions,omitempty"`
	Difficulte      string `json:"difficulte,omitempty"`
	Titre           string `json:"titre,omitempty"`
}

// ReponseQuiz représente la réponse de génération de quiz
type ReponseQuiz struct {
	Succes bool           `json:"succes"`
	Quiz   *QuizReponse   `json:"quiz,omitempty"`
	Erreur *ErreurReponse `json:"erreur,omitempty"`
}

// QuizReponse représente un quiz dans la réponse API
type QuizReponse struct {
	ID              string            `json:"id"`
	CoursID         string            `json:"coursId"`
	Titre           string            `json:"titre"`
	Difficulte      string            `json:"difficulte"`
	NombreQuestions int               `json:"nombreQuestions"`
	Questions       []QuestionReponse `json:"questions"`
}

// QuestionReponse représente une question dans la réponse API
type QuestionReponse struct {
	ID              string   `json:"id"`
	Enonce          string   `json:"enonce"`
	Choix           []string `json:"choix"`
	ReponseCorrecte int      `json:"reponseCorrecte"`
	Explication     string   `json:"explication"`
}

// ReponseSessionQuiz représente la réponse API d'une session
type ReponseSessionQuiz struct {
	Succes  bool                `json:"succes"`
	Session *SessionQuizReponse `json:"session,omitempty"`
	Erreur  *ErreurReponse      `json:"erreur,omitempty"`
}

// SessionQuizReponse représente une session dans la réponse API
type SessionQuizReponse struct {
	ID        string                 `json:"id"`
	QuizID    string                 `json:"quizId"`
	Reponses  []ReponseSessionDetail `json:"reponses"`
	Score     *float64               `json:"score,omitempty"`
	Termine   bool                   `json:"termine"`
	DateDebut string                 `json:"dateDebut"`
	DateFin   *string                `json:"dateFin,omitempty"`
}

// ReponseSessionDetail représente une réponse dans une session
type ReponseSessionDetail struct {
	QuestionID  string `json:"questionId"`
	ChoixIndex  int    `json:"choixIndex"`
	EstCorrecte bool   `json:"estCorrecte"`
}

// RequeteRepondre représente la requête pour répondre à une question
type RequeteRepondre struct {
	QuestionID string `json:"questionId" binding:"required"`
	ChoixIndex int    `json:"choixIndex"`
}

// ReponseRepondre représente la réponse après avoir soumis une réponse
type ReponseRepondre struct {
	Succes      bool           `json:"succes"`
	EstCorrecte bool           `json:"estCorrecte"`
	Explication string         `json:"explication,omitempty"`
	Erreur      *ErreurReponse `json:"erreur,omitempty"`
}

// GenererQuizHandler génère un quiz pour un cours
func (h *HandlersGeneration) GenererQuizHandler(c *gin.Context) {
	if h.serviceGeneration == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseQuiz{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service de génération n'est pas configuré",
			},
		})
		return
	}

	var req RequeteGenererQuiz
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ReponseQuiz{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "REQUETE_INVALIDE",
				Message: "Le champ coursId est requis",
			},
		})
		return
	}

	// Valider la difficulté si fournie
	if req.Difficulte != "" && req.Difficulte != "facile" && req.Difficulte != "moyen" && req.Difficulte != "difficile" {
		c.JSON(http.StatusBadRequest, ReponseQuiz{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "DIFFICULTE_INVALIDE",
				Message: "La difficulté doit être: facile, moyen ou difficile",
			},
		})
		return
	}

	options := &services.OptionsGenerationQuiz{
		NombreQuestions: req.NombreQuestions,
		Difficulte:      req.Difficulte,
		Titre:           req.Titre,
	}

	resultat, err := h.serviceGeneration.GenererQuiz(c.Request.Context(), req.CoursID, options)
	if err != nil {
		h.gererErreurGenerationQuiz(c, err)
		return
	}

	// Convertir le quiz pour la réponse
	questions := make([]QuestionReponse, 0, len(resultat.Quiz.Questions))
	for _, q := range resultat.Quiz.Questions {
		questions = append(questions, QuestionReponse{
			ID:              q.ID,
			Enonce:          q.Enonce,
			Choix:           q.Choix,
			ReponseCorrecte: q.ReponseCorrecte,
			Explication:     q.Explication,
		})
	}

	c.JSON(http.StatusOK, ReponseQuiz{
		Succes: true,
		Quiz: &QuizReponse{
			ID:              resultat.Quiz.ID,
			CoursID:         resultat.Quiz.CoursID,
			Titre:           resultat.Quiz.Titre,
			Difficulte:      resultat.Quiz.Difficulte,
			NombreQuestions: resultat.Quiz.NombreQuestions,
			Questions:       questions,
		},
	})
}

// ObtenirQuizHandler récupère un quiz par son ID
func (h *HandlersGeneration) ObtenirQuizHandler(c *gin.Context) {
	quizID := c.Param("id")
	if quizID == "" {
		c.JSON(http.StatusBadRequest, ReponseQuiz{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "ID_MANQUANT",
				Message: "L'ID du quiz est requis",
			},
		})
		return
	}

	if h.serviceGeneration == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseQuiz{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service n'est pas configuré",
			},
		})
		return
	}

	quiz, err := h.serviceGeneration.ObtenirQuizParID(c.Request.Context(), quizID)
	if err != nil {
		h.gererErreurGenerationQuiz(c, err)
		return
	}

	if quiz == nil {
		c.JSON(http.StatusNotFound, ReponseQuiz{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "QUIZ_NON_TROUVE",
				Message: "Quiz non trouvé",
			},
		})
		return
	}

	// Convertir le quiz pour la réponse
	questions := make([]QuestionReponse, 0, len(quiz.Questions))
	for _, q := range quiz.Questions {
		questions = append(questions, QuestionReponse{
			ID:              q.ID,
			Enonce:          q.Enonce,
			Choix:           q.Choix,
			ReponseCorrecte: q.ReponseCorrecte,
			Explication:     q.Explication,
		})
	}

	c.JSON(http.StatusOK, ReponseQuiz{
		Succes: true,
		Quiz: &QuizReponse{
			ID:              quiz.ID,
			CoursID:         quiz.CoursID,
			Titre:           quiz.Titre,
			Difficulte:      quiz.Difficulte,
			NombreQuestions: quiz.NombreQuestions,
			Questions:       questions,
		},
	})
}

// DemarrerSessionHandler démarre une nouvelle session de quiz
func (h *HandlersGeneration) DemarrerSessionHandler(c *gin.Context) {
	quizID := c.Param("id")
	if quizID == "" {
		c.JSON(http.StatusBadRequest, ReponseSessionQuiz{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "ID_MANQUANT",
				Message: "L'ID du quiz est requis",
			},
		})
		return
	}

	if h.serviceGeneration == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseSessionQuiz{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service n'est pas configuré",
			},
		})
		return
	}

	session, err := h.serviceGeneration.DemarrerSession(c.Request.Context(), quizID)
	if err != nil {
		h.gererErreurGenerationSession(c, err)
		return
	}

	c.JSON(http.StatusCreated, ReponseSessionQuiz{
		Succes: true,
		Session: &SessionQuizReponse{
			ID:        session.ID,
			QuizID:    session.QuizID,
			Reponses:  []ReponseSessionDetail{},
			Termine:   false,
			DateDebut: session.DateDebut.Format("2006-01-02T15:04:05Z07:00"),
		},
	})
}

// RepondreHandler enregistre la réponse à une question
func (h *HandlersGeneration) RepondreHandler(c *gin.Context) {
	quizID := c.Param("id")
	sessionID := c.Param("sessionId")

	if quizID == "" || sessionID == "" {
		c.JSON(http.StatusBadRequest, ReponseRepondre{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "ID_MANQUANT",
				Message: "L'ID du quiz et de la session sont requis",
			},
		})
		return
	}

	if h.serviceGeneration == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseRepondre{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service n'est pas configuré",
			},
		})
		return
	}

	var req RequeteRepondre
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ReponseRepondre{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "REQUETE_INVALIDE",
				Message: "Les champs questionId et choixIndex sont requis",
			},
		})
		return
	}

	// Valider l'index du choix
	if req.ChoixIndex < 0 || req.ChoixIndex > 3 {
		c.JSON(http.StatusBadRequest, ReponseRepondre{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "CHOIX_INVALIDE",
				Message: "L'index du choix doit être entre 0 et 3",
			},
		})
		return
	}

	_, estCorrecte, err := h.serviceGeneration.RepondreQuestion(c.Request.Context(), sessionID, req.QuestionID, req.ChoixIndex)
	if err != nil {
		h.gererErreurGenerationSession(c, err)
		return
	}

	// Récupérer l'explication de la question
	var explication string
	quiz, _ := h.serviceGeneration.ObtenirQuizParID(c.Request.Context(), quizID)
	if quiz != nil {
		for _, q := range quiz.Questions {
			if q.ID == req.QuestionID {
				explication = q.Explication
				break
			}
		}
	}

	c.JSON(http.StatusOK, ReponseRepondre{
		Succes:      true,
		EstCorrecte: estCorrecte,
		Explication: explication,
	})
}

// TerminerSessionHandler termine une session de quiz
func (h *HandlersGeneration) TerminerSessionHandler(c *gin.Context) {
	sessionID := c.Param("sessionId")

	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ReponseSessionQuiz{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "ID_MANQUANT",
				Message: "L'ID de la session est requis",
			},
		})
		return
	}

	if h.serviceGeneration == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseSessionQuiz{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service n'est pas configuré",
			},
		})
		return
	}

	session, err := h.serviceGeneration.TerminerSession(c.Request.Context(), sessionID)
	if err != nil {
		h.gererErreurGenerationSession(c, err)
		return
	}

	// Convertir les réponses
	reponses := make([]ReponseSessionDetail, 0, len(session.Reponses))
	for _, r := range session.Reponses {
		reponses = append(reponses, ReponseSessionDetail{
			QuestionID:  r.QuestionID,
			ChoixIndex:  r.ChoixIndex,
			EstCorrecte: r.EstCorrecte,
		})
	}

	sessionReponse := &SessionQuizReponse{
		ID:        session.ID,
		QuizID:    session.QuizID,
		Reponses:  reponses,
		Score:     session.Score,
		Termine:   session.Termine,
		DateDebut: session.DateDebut.Format("2006-01-02T15:04:05Z07:00"),
	}

	if session.DateFin != nil {
		dateFin := session.DateFin.Format("2006-01-02T15:04:05Z07:00")
		sessionReponse.DateFin = &dateFin
	}

	c.JSON(http.StatusOK, ReponseSessionQuiz{
		Succes:  true,
		Session: sessionReponse,
	})
}

// gererErreurGenerationQuiz gère les erreurs spécifiques aux quiz
func (h *HandlersGeneration) gererErreurGenerationQuiz(c *gin.Context, err error) {
	var errGen *services.ErreurGeneration
	if errors.As(err, &errGen) {
		statusCode := http.StatusInternalServerError
		switch errGen.Code {
		case "COURS_NON_TROUVE", "QUIZ_NON_TROUVE":
			statusCode = http.StatusNotFound
		case "COURS_VIDE":
			statusCode = http.StatusBadRequest
		case "SERVICE_NON_DISPONIBLE":
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, ReponseQuiz{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    errGen.Code,
				Message: errGen.Message,
			},
		})
		return
	}

	var errLLM *llm.ErreurLLM
	if errors.As(err, &errLLM) {
		if errLLM.RateLimited {
			c.JSON(http.StatusTooManyRequests, ReponseQuiz{
				Succes: false,
				Erreur: &ErreurReponse{
					Code:    "QUOTA_DEPASSE",
					Message: "Limite d'appels API atteinte, réessayez plus tard",
				},
			})
			return
		}
	}

	c.JSON(http.StatusInternalServerError, ReponseQuiz{
		Succes: false,
		Erreur: &ErreurReponse{
			Code:    "ERREUR_INTERNE",
			Message: "Une erreur est survenue lors de la génération",
		},
	})
}

// gererErreurGenerationSession gère les erreurs spécifiques aux sessions
func (h *HandlersGeneration) gererErreurGenerationSession(c *gin.Context, err error) {
	var errGen *services.ErreurGeneration
	if errors.As(err, &errGen) {
		statusCode := http.StatusInternalServerError
		switch errGen.Code {
		case "QUIZ_NON_TROUVE", "SESSION_NON_TROUVEE":
			statusCode = http.StatusNotFound
		case "SESSION_TERMINEE":
			statusCode = http.StatusBadRequest
		case "SERVICE_NON_DISPONIBLE":
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, ReponseSessionQuiz{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    errGen.Code,
				Message: errGen.Message,
			},
		})
		return
	}

	c.JSON(http.StatusInternalServerError, ReponseSessionQuiz{
		Succes: false,
		Erreur: &ErreurReponse{
			Code:    "ERREUR_INTERNE",
			Message: "Une erreur est survenue",
		},
	})
}

// --- Ressources Handlers ---

// RequeteGenererRessources représente la requête pour générer des ressources
type RequeteGenererRessources struct {
	CoursID string `json:"coursId" binding:"required"`
}

// ReponseRessources représente la réponse de génération de ressources
type ReponseRessources struct {
	Succes        bool               `json:"succes"`
	Ressources    []RessourceReponse `json:"ressources,omitempty"`
	NombreGenere  int                `json:"nombreGenere,omitempty"`
	Avertissement string             `json:"avertissement,omitempty"`
	Erreur        *ErreurReponse     `json:"erreur,omitempty"`
}

// RessourceReponse représente une ressource dans la réponse API
type RessourceReponse struct {
	ID          string `json:"id"`
	Titre       string `json:"titre"`
	URL         string `json:"url,omitempty"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// GenererRessourcesHandler génère des ressources complémentaires pour un cours
func (h *HandlersGeneration) GenererRessourcesHandler(c *gin.Context) {
	if h.serviceGeneration == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseRessources{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service de génération n'est pas configuré",
			},
		})
		return
	}

	var req RequeteGenererRessources
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ReponseRessources{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "REQUETE_INVALIDE",
				Message: "Le champ coursId est requis",
			},
		})
		return
	}

	resultat, err := h.serviceGeneration.GenererRessources(c.Request.Context(), req.CoursID)
	if err != nil {
		h.gererErreurGenerationRessources(c, err)
		return
	}

	// Convertir les ressources pour la réponse
	ressourcesReponse := make([]RessourceReponse, 0, len(resultat.Ressources))
	for _, r := range resultat.Ressources {
		ressourcesReponse = append(ressourcesReponse, RessourceReponse{
			ID:          r.ID,
			Titre:       r.Titre,
			URL:         r.URL,
			Type:        string(r.Type),
			Description: r.Description,
		})
	}

	c.JSON(http.StatusOK, ReponseRessources{
		Succes:        true,
		Ressources:    ressourcesReponse,
		NombreGenere:  resultat.NombreGenere,
		Avertissement: resultat.Avertissement,
	})
}

// ObtenirRessourcesHandler récupère les ressources existantes d'un cours
func (h *HandlersGeneration) ObtenirRessourcesHandler(c *gin.Context) {
	coursID := c.Param("id")
	if coursID == "" {
		c.JSON(http.StatusBadRequest, ReponseRessources{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "ID_MANQUANT",
				Message: "L'ID du cours est requis",
			},
		})
		return
	}

	if h.serviceGeneration == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseRessources{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service n'est pas configuré",
			},
		})
		return
	}

	ressources, err := h.serviceGeneration.ObtenirRessourcesParCours(c.Request.Context(), coursID)
	if err != nil {
		h.gererErreurGenerationRessources(c, err)
		return
	}

	// Convertir les ressources pour la réponse
	ressourcesReponse := make([]RessourceReponse, 0, len(ressources))
	for _, r := range ressources {
		ressourcesReponse = append(ressourcesReponse, RessourceReponse{
			ID:          r.ID,
			Titre:       r.Titre,
			URL:         r.URL,
			Type:        string(r.Type),
			Description: r.Description,
		})
	}

	c.JSON(http.StatusOK, ReponseRessources{
		Succes:        true,
		Ressources:    ressourcesReponse,
		NombreGenere:  len(ressourcesReponse),
		Avertissement: "Les liens suggérés sont générés par IA et doivent être vérifiés avant utilisation.",
	})
}

// gererErreurGenerationRessources gère les erreurs spécifiques aux ressources
func (h *HandlersGeneration) gererErreurGenerationRessources(c *gin.Context, err error) {
	var errGen *services.ErreurGeneration
	if errors.As(err, &errGen) {
		statusCode := http.StatusInternalServerError
		switch errGen.Code {
		case "COURS_NON_TROUVE":
			statusCode = http.StatusNotFound
		case "COURS_VIDE":
			statusCode = http.StatusBadRequest
		case "SERVICE_NON_DISPONIBLE":
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, ReponseRessources{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    errGen.Code,
				Message: errGen.Message,
			},
		})
		return
	}

	var errLLM *llm.ErreurLLM
	if errors.As(err, &errLLM) {
		if errLLM.RateLimited {
			c.JSON(http.StatusTooManyRequests, ReponseRessources{
				Succes: false,
				Erreur: &ErreurReponse{
					Code:    "QUOTA_DEPASSE",
					Message: "Limite d'appels API atteinte, réessayez plus tard",
				},
			})
			return
		}
	}

	c.JSON(http.StatusInternalServerError, ReponseRessources{
		Succes: false,
		Erreur: &ErreurReponse{
			Code:    "ERREUR_INTERNE",
			Message: "Une erreur est survenue lors de la génération des ressources",
		},
	})
}

// gererErreurGeneration gère les erreurs du service de génération
func (h *HandlersGeneration) gererErreurGeneration(c *gin.Context, err error) {
	// Vérifier si c'est une erreur LLM
	var errLLM *llm.ErreurLLM
	if errors.As(err, &errLLM) {
		if errLLM.RateLimited {
			c.JSON(http.StatusTooManyRequests, ReponseFiches{
				Succes: false,
				Erreur: &ErreurReponse{
					Code:    "QUOTA_DEPASSE",
					Message: "Limite d'appels API atteinte, réessayez plus tard",
				},
			})
			return
		}
	}

	// Vérifier les erreurs spécifiques du service
	var errGen *services.ErreurGeneration
	if errors.As(err, &errGen) {
		statusCode := http.StatusInternalServerError
		switch errGen.Code {
		case "COURS_NON_TROUVE":
			statusCode = http.StatusNotFound
		case "COURS_VIDE":
			statusCode = http.StatusBadRequest
		case "SERVICE_NON_DISPONIBLE":
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, ReponseFiches{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    errGen.Code,
				Message: errGen.Message,
			},
		})
		return
	}

	// Erreur générique
	c.JSON(http.StatusInternalServerError, ReponseFiches{
		Succes: false,
		Erreur: &ErreurReponse{
			Code:    "ERREUR_INTERNE",
			Message: "Une erreur est survenue lors de la génération",
		},
	})
}
