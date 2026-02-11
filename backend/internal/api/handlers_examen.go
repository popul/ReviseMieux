// Package api contient les handlers HTTP
package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/revisemieux/backend/internal/llm"
	"github.com/revisemieux/backend/internal/services"
	"github.com/revisemieux/backend/internal/store"
)

// HandlersExamen gere les endpoints d'examen blanc
type HandlersExamen struct {
	serviceExamen *services.ServiceExamen
}

// NouveauHandlersExamen cree une nouvelle instance des handlers d'examen
func NouveauHandlersExamen(serviceExamen *services.ServiceExamen) *HandlersExamen {
	return &HandlersExamen{
		serviceExamen: serviceExamen,
	}
}

// --- Types de reponse API ---

// ReponseExamenBlanc represente la reponse de generation d'examen
type ReponseExamenBlanc struct {
	Succes bool                `json:"succes"`
	Examen *ExamenBlancReponse `json:"examen,omitempty"`
	Erreur *ErreurReponse      `json:"erreur,omitempty"`
}

// ExamenBlancReponse represente un examen dans la reponse API
type ExamenBlancReponse struct {
	ID           string                  `json:"id"`
	CoursID      string                  `json:"coursId"`
	Questions    []QuestionExamenReponse `json:"questions"`
	DureeMinutes int                     `json:"dureeMinutes"`
	DateCreation string                  `json:"dateCreation"`
}

// QuestionExamenReponse represente une question d'examen dans la reponse API
type QuestionExamenReponse struct {
	Numero          int                   `json:"numero"`
	Type            string                `json:"type"`
	Difficulte      string                `json:"difficulte"`
	Enonce          string                `json:"enonce"`
	Bareme          int                   `json:"bareme"`
	ReponseAttendue string                `json:"reponseAttendue"`
	Indices         []IndiceExamenReponse `json:"indices"`
}

// IndiceExamenReponse represente un indice dans la reponse API
type IndiceExamenReponse struct {
	Niveau int    `json:"niveau"`
	Texte  string `json:"texte"`
}

// ReponseSessionExamen represente la reponse API d'une session d'examen
type ReponseSessionExamen struct {
	Succes  bool                  `json:"succes"`
	Session *SessionExamenReponse `json:"session,omitempty"`
	Erreur  *ErreurReponse        `json:"erreur,omitempty"`
}

// SessionExamenReponse represente une session d'examen dans la reponse API
type SessionExamenReponse struct {
	ID              string                     `json:"id"`
	ExamenID        string                     `json:"examenId"`
	Reponses        []ReponseExamenDetail      `json:"reponses"`
	IndicesUtilises []IndiceUtiliseReponse     `json:"indicesUtilises"`
	NoteEstimee     *float64                   `json:"noteEstimee,omitempty"`
	PointsForts     []PointFortReponse         `json:"pointsForts,omitempty"`
	PointsFaibles   []PointFaibleReponse       `json:"pointsFaibles,omitempty"`
	PlanRevision    []EtapePlanRevisionReponse `json:"planRevision,omitempty"`
	Termine         bool                       `json:"termine"`
	DateDebut       string                     `json:"dateDebut"`
	DateFin         *string                    `json:"dateFin,omitempty"`
}

// ReponseExamenDetail represente une reponse dans la session
type ReponseExamenDetail struct {
	QuestionNumero int    `json:"questionNumero"`
	Texte          string `json:"texte"`
}

// IndiceUtiliseReponse represente un indice utilise dans la reponse API
type IndiceUtiliseReponse struct {
	QuestionNumero int `json:"questionNumero"`
	NiveauIndice   int `json:"niveauIndice"`
}

// PointFortReponse represente un point fort dans la reponse API
type PointFortReponse struct {
	Concept     string `json:"concept"`
	Commentaire string `json:"commentaire"`
}

// PointFaibleReponse represente un point faible dans la reponse API
type PointFaibleReponse struct {
	Concept     string `json:"concept"`
	Commentaire string `json:"commentaire"`
}

// EtapePlanRevisionReponse represente une etape du plan de revision dans la reponse API
type EtapePlanRevisionReponse struct {
	Priorite  int    `json:"priorite"`
	Action    string `json:"action"`
	Concept   string `json:"concept"`
	Ressource string `json:"ressource,omitempty"`
}

// ReponseIndice represente la reponse d'un indice
type ReponseIndice struct {
	Succes bool                 `json:"succes"`
	Indice *IndiceExamenReponse `json:"indice,omitempty"`
	Erreur *ErreurReponse       `json:"erreur,omitempty"`
}

// ReponseCorrectionExamen represente la reponse de correction
type ReponseCorrectionExamen struct {
	Succes   bool                       `json:"succes"`
	Resultat *ResultatCorrectionReponse `json:"resultat,omitempty"`
	Erreur   *ErreurReponse             `json:"erreur,omitempty"`
}

// ResultatCorrectionReponse represente le resultat de correction dans la reponse API
type ResultatCorrectionReponse struct {
	NoteEstimee   float64                    `json:"noteEstimee"`
	PointsForts   []PointFortReponse         `json:"pointsForts"`
	PointsFaibles []PointFaibleReponse       `json:"pointsFaibles"`
	PlanRevision  []EtapePlanRevisionReponse `json:"planRevision"`
	Details       []DetailCorrectionReponse  `json:"details"`
}

// DetailCorrectionReponse represente le detail de correction d'une question
type DetailCorrectionReponse struct {
	QuestionNumero  int     `json:"questionNumero"`
	ReponseEleve    string  `json:"reponseEleve"`
	ReponseAttendue string  `json:"reponseAttendue"`
	NoteQuestion    float64 `json:"noteQuestion"`
	Bareme          int     `json:"bareme"`
	Commentaire     string  `json:"commentaire"`
}

// --- Handlers ---

// GenererExamenHandler genere un examen blanc pour un cours
func (h *HandlersExamen) GenererExamenHandler(c *gin.Context) {
	if h.serviceExamen == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseExamenBlanc{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service d'examen n'est pas configure",
			},
		})
		return
	}

	coursID := c.Param("id")
	if coursID == "" {
		c.JSON(http.StatusBadRequest, ReponseExamenBlanc{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "ID_MANQUANT",
				Message: "L'identifiant du cours est requis",
			},
		})
		return
	}

	examen, err := h.serviceExamen.GenererExamen(c.Request.Context(), coursID)
	if err != nil {
		h.gererErreurExamen(c, err)
		return
	}

	c.JSON(http.StatusOK, ReponseExamenBlanc{
		Succes: true,
		Examen: convertirExamenEnReponse(examen),
	})
}

// ObtenirExamenHandler recupere un examen par son ID
func (h *HandlersExamen) ObtenirExamenHandler(c *gin.Context) {
	if h.serviceExamen == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseExamenBlanc{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service d'examen n'est pas configure",
			},
		})
		return
	}

	examenID := c.Param("id")
	if examenID == "" {
		c.JSON(http.StatusBadRequest, ReponseExamenBlanc{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "ID_MANQUANT",
				Message: "L'identifiant de l'examen est requis",
			},
		})
		return
	}

	examen, err := h.serviceExamen.ObtenirExamen(c.Request.Context(), examenID)
	if err != nil {
		h.gererErreurExamen(c, err)
		return
	}

	if examen == nil {
		c.JSON(http.StatusNotFound, ReponseExamenBlanc{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "EXAMEN_NON_TROUVE",
				Message: "Examen non trouve",
			},
		})
		return
	}

	c.JSON(http.StatusOK, ReponseExamenBlanc{
		Succes: true,
		Examen: convertirExamenEnReponse(examen),
	})
}

// DemarrerSessionExamenHandler demarre une nouvelle session d'examen
func (h *HandlersExamen) DemarrerSessionExamenHandler(c *gin.Context) {
	if h.serviceExamen == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseSessionExamen{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service d'examen n'est pas configure",
			},
		})
		return
	}

	examenID := c.Param("id")
	if examenID == "" {
		c.JSON(http.StatusBadRequest, ReponseSessionExamen{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "ID_MANQUANT",
				Message: "L'identifiant de l'examen est requis",
			},
		})
		return
	}

	session, err := h.serviceExamen.DemarrerSessionExamen(c.Request.Context(), examenID)
	if err != nil {
		h.gererErreurSessionExamen(c, err)
		return
	}

	c.JSON(http.StatusCreated, ReponseSessionExamen{
		Succes:  true,
		Session: convertirSessionEnReponse(session),
	})
}

// ObtenirSessionExamenHandler recupere une session d'examen par son ID
func (h *HandlersExamen) ObtenirSessionExamenHandler(c *gin.Context) {
	if h.serviceExamen == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseSessionExamen{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service d'examen n'est pas configure",
			},
		})
		return
	}

	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ReponseSessionExamen{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "ID_MANQUANT",
				Message: "L'identifiant de la session est requis",
			},
		})
		return
	}

	session, err := h.serviceExamen.ObtenirSessionExamen(c.Request.Context(), sessionID)
	if err != nil {
		h.gererErreurSessionExamen(c, err)
		return
	}

	if session == nil {
		c.JSON(http.StatusNotFound, ReponseSessionExamen{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SESSION_EXAMEN_NON_TROUVEE",
				Message: "Session d'examen non trouvee",
			},
		})
		return
	}

	c.JSON(http.StatusOK, ReponseSessionExamen{
		Succes:  true,
		Session: convertirSessionEnReponse(session),
	})
}

// RequeteIndice represente la requete pour obtenir un indice
type RequeteIndice struct {
	QuestionNumero int `json:"questionNumero" binding:"required"`
	NiveauIndice   int `json:"niveauIndice" binding:"required"`
}

// DemanderIndiceHandler retourne un indice pour une question
func (h *HandlersExamen) DemanderIndiceHandler(c *gin.Context) {
	if h.serviceExamen == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseIndice{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service d'examen n'est pas configure",
			},
		})
		return
	}

	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ReponseIndice{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "ID_MANQUANT",
				Message: "L'identifiant de la session est requis",
			},
		})
		return
	}

	var req RequeteIndice
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ReponseIndice{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "REQUETE_INVALIDE",
				Message: "Les champs questionNumero et niveauIndice sont requis",
			},
		})
		return
	}

	indice, err := h.serviceExamen.GetIndice(c.Request.Context(), sessionID, req.QuestionNumero, req.NiveauIndice)
	if err != nil {
		h.gererErreurSessionExamen(c, err)
		return
	}

	c.JSON(http.StatusOK, ReponseIndice{
		Succes: true,
		Indice: &IndiceExamenReponse{
			Niveau: indice.Niveau,
			Texte:  indice.Texte,
		},
	})
}

// RequeteCorrigerExamen represente la requete pour corriger un examen
type RequeteCorrigerExamen struct {
	Reponses []ReponseExamenDetail `json:"reponses" binding:"required"`
}

// CorrigerExamenHandler corrige un examen et retourne les resultats
func (h *HandlersExamen) CorrigerExamenHandler(c *gin.Context) {
	if h.serviceExamen == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseCorrectionExamen{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service d'examen n'est pas configure",
			},
		})
		return
	}

	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ReponseCorrectionExamen{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "ID_MANQUANT",
				Message: "L'identifiant de la session est requis",
			},
		})
		return
	}

	var req RequeteCorrigerExamen
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ReponseCorrectionExamen{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "REQUETE_INVALIDE",
				Message: "Le champ reponses est requis",
			},
		})
		return
	}

	// Convertir les reponses
	reponses := make([]store.ReponseExamen, 0, len(req.Reponses))
	for _, r := range req.Reponses {
		reponses = append(reponses, store.ReponseExamen{
			QuestionNumero: r.QuestionNumero,
			Texte:          r.Texte,
		})
	}

	resultat, err := h.serviceExamen.CorrigerExamen(c.Request.Context(), sessionID, reponses)
	if err != nil {
		h.gererErreurCorrectionExamen(c, err)
		return
	}

	details := make([]DetailCorrectionReponse, 0, len(resultat.Details))
	for _, d := range resultat.Details {
		details = append(details, DetailCorrectionReponse{
			QuestionNumero:  d.QuestionNumero,
			ReponseEleve:    d.ReponseEleve,
			ReponseAttendue: d.ReponseAttendue,
			NoteQuestion:    d.NoteQuestion,
			Bareme:          d.Bareme,
			Commentaire:     d.Commentaire,
		})
	}

	pointsForts := make([]PointFortReponse, 0, len(resultat.PointsForts))
	for _, pf := range resultat.PointsForts {
		pointsForts = append(pointsForts, PointFortReponse{
			Concept:     pf.Concept,
			Commentaire: pf.Commentaire,
		})
	}

	pointsFaibles := make([]PointFaibleReponse, 0, len(resultat.PointsFaibles))
	for _, pf := range resultat.PointsFaibles {
		pointsFaibles = append(pointsFaibles, PointFaibleReponse{
			Concept:     pf.Concept,
			Commentaire: pf.Commentaire,
		})
	}

	planRevision := make([]EtapePlanRevisionReponse, 0, len(resultat.PlanRevision))
	for _, pr := range resultat.PlanRevision {
		planRevision = append(planRevision, EtapePlanRevisionReponse{
			Priorite:  pr.Priorite,
			Action:    pr.Action,
			Concept:   pr.Concept,
			Ressource: pr.Ressource,
		})
	}

	c.JSON(http.StatusOK, ReponseCorrectionExamen{
		Succes: true,
		Resultat: &ResultatCorrectionReponse{
			NoteEstimee:   resultat.NoteEstimee,
			PointsForts:   pointsForts,
			PointsFaibles: pointsFaibles,
			PlanRevision:  planRevision,
			Details:       details,
		},
	})
}

// --- Helpers ---

func convertirExamenEnReponse(examen *store.ExamenBlanc) *ExamenBlancReponse {
	questions := make([]QuestionExamenReponse, 0, len(examen.Questions))
	for _, q := range examen.Questions {
		indices := make([]IndiceExamenReponse, 0, len(q.Indices))
		for _, idx := range q.Indices {
			indices = append(indices, IndiceExamenReponse{
				Niveau: idx.Niveau,
				Texte:  idx.Texte,
			})
		}
		questions = append(questions, QuestionExamenReponse{
			Numero:          q.Numero,
			Type:            q.Type,
			Difficulte:      q.Difficulte,
			Enonce:          q.Enonce,
			Bareme:          q.Bareme,
			ReponseAttendue: q.ReponseAttendue,
			Indices:         indices,
		})
	}

	return &ExamenBlancReponse{
		ID:           examen.ID,
		CoursID:      examen.CoursID,
		Questions:    questions,
		DureeMinutes: examen.DureeMinutes,
		DateCreation: examen.DateCreation.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func convertirSessionEnReponse(session *store.SessionExamen) *SessionExamenReponse {
	reponses := make([]ReponseExamenDetail, 0, len(session.Reponses))
	for _, r := range session.Reponses {
		reponses = append(reponses, ReponseExamenDetail{
			QuestionNumero: r.QuestionNumero,
			Texte:          r.Texte,
		})
	}

	indices := make([]IndiceUtiliseReponse, 0, len(session.IndicesUtilises))
	for _, iu := range session.IndicesUtilises {
		indices = append(indices, IndiceUtiliseReponse{
			QuestionNumero: iu.QuestionNumero,
			NiveauIndice:   iu.NiveauIndice,
		})
	}

	pointsForts := make([]PointFortReponse, 0)
	for _, pf := range session.PointsForts {
		pointsForts = append(pointsForts, PointFortReponse{Concept: pf.Concept, Commentaire: pf.Commentaire})
	}

	pointsFaibles := make([]PointFaibleReponse, 0)
	for _, pf := range session.PointsFaibles {
		pointsFaibles = append(pointsFaibles, PointFaibleReponse{Concept: pf.Concept, Commentaire: pf.Commentaire})
	}

	planRevision := make([]EtapePlanRevisionReponse, 0)
	for _, pr := range session.PlanRevision {
		planRevision = append(planRevision, EtapePlanRevisionReponse{
			Priorite: pr.Priorite, Action: pr.Action, Concept: pr.Concept, Ressource: pr.Ressource,
		})
	}

	result := &SessionExamenReponse{
		ID:              session.ID,
		ExamenID:        session.ExamenID,
		Reponses:        reponses,
		IndicesUtilises: indices,
		NoteEstimee:     session.NoteEstimee,
		Termine:         session.Termine,
		DateDebut:       session.DateDebut.Format("2006-01-02T15:04:05Z07:00"),
	}

	if len(pointsForts) > 0 {
		result.PointsForts = pointsForts
	}
	if len(pointsFaibles) > 0 {
		result.PointsFaibles = pointsFaibles
	}
	if len(planRevision) > 0 {
		result.PlanRevision = planRevision
	}
	if session.DateFin != nil {
		dateFin := session.DateFin.Format("2006-01-02T15:04:05Z07:00")
		result.DateFin = &dateFin
	}

	return result
}

func (h *HandlersExamen) gererErreurExamen(c *gin.Context, err error) {
	var errGen *services.ErreurGeneration
	if errors.As(err, &errGen) {
		statusCode := http.StatusInternalServerError
		switch errGen.Code {
		case "COURS_NON_TROUVE", "EXAMEN_NON_TROUVE":
			statusCode = http.StatusNotFound
		case "COURS_VIDE":
			statusCode = http.StatusBadRequest
		case "SERVICE_NON_DISPONIBLE":
			statusCode = http.StatusServiceUnavailable
		}
		c.JSON(statusCode, ReponseExamenBlanc{Succes: false, Erreur: &ErreurReponse{Code: errGen.Code, Message: errGen.Message}})
		return
	}
	var errLLM *llm.ErreurLLM
	if errors.As(err, &errLLM) && errLLM.RateLimited {
		c.JSON(http.StatusTooManyRequests, ReponseExamenBlanc{Succes: false, Erreur: &ErreurReponse{Code: "QUOTA_DEPASSE", Message: "Limite d'appels API atteinte, reessayez plus tard"}})
		return
	}
	c.JSON(http.StatusInternalServerError, ReponseExamenBlanc{Succes: false, Erreur: &ErreurReponse{Code: "ERREUR_INTERNE", Message: "Une erreur est survenue lors de la generation de l'examen"}})
}

func (h *HandlersExamen) gererErreurSessionExamen(c *gin.Context, err error) {
	var errGen *services.ErreurGeneration
	if errors.As(err, &errGen) {
		statusCode := http.StatusInternalServerError
		switch errGen.Code {
		case "EXAMEN_NON_TROUVE", "SESSION_EXAMEN_NON_TROUVEE":
			statusCode = http.StatusNotFound
		case "SESSION_EXAMEN_TERMINEE", "INDICE_INVALIDE":
			statusCode = http.StatusBadRequest
		case "SERVICE_NON_DISPONIBLE":
			statusCode = http.StatusServiceUnavailable
		}
		c.JSON(statusCode, ReponseSessionExamen{Succes: false, Erreur: &ErreurReponse{Code: errGen.Code, Message: errGen.Message}})
		return
	}
	c.JSON(http.StatusInternalServerError, ReponseSessionExamen{Succes: false, Erreur: &ErreurReponse{Code: "ERREUR_INTERNE", Message: "Une erreur est survenue"}})
}

func (h *HandlersExamen) gererErreurCorrectionExamen(c *gin.Context, err error) {
	var errGen *services.ErreurGeneration
	if errors.As(err, &errGen) {
		statusCode := http.StatusInternalServerError
		switch errGen.Code {
		case "EXAMEN_NON_TROUVE", "SESSION_EXAMEN_NON_TROUVEE":
			statusCode = http.StatusNotFound
		case "SESSION_EXAMEN_TERMINEE":
			statusCode = http.StatusBadRequest
		case "SERVICE_NON_DISPONIBLE":
			statusCode = http.StatusServiceUnavailable
		}
		c.JSON(statusCode, ReponseCorrectionExamen{Succes: false, Erreur: &ErreurReponse{Code: errGen.Code, Message: errGen.Message}})
		return
	}
	var errLLM *llm.ErreurLLM
	if errors.As(err, &errLLM) && errLLM.RateLimited {
		c.JSON(http.StatusTooManyRequests, ReponseCorrectionExamen{Succes: false, Erreur: &ErreurReponse{Code: "QUOTA_DEPASSE", Message: "Limite d'appels API atteinte, reessayez plus tard"}})
		return
	}
	c.JSON(http.StatusInternalServerError, ReponseCorrectionExamen{Succes: false, Erreur: &ErreurReponse{Code: "ERREUR_INTERNE", Message: "Une erreur est survenue lors de la correction"}})
}
