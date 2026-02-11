// Package services contient les services metier de l'application
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/revisemieux/backend/internal/llm"
	"github.com/revisemieux/backend/internal/store"
)

// ServiceExamen gere la generation et correction des examens blancs
type ServiceExamen struct {
	gestionnaireLLM *llm.GestionnaireLLM
	coursRepo       store.CoursRepository
	examenRepo      store.ExamenRepository
}

// NouveauServiceExamen cree une nouvelle instance du service d'examen
func NouveauServiceExamen(
	gestionnaireLLM *llm.GestionnaireLLM,
	coursRepo store.CoursRepository,
	examenRepo store.ExamenRepository,
) *ServiceExamen {
	return &ServiceExamen{
		gestionnaireLLM: gestionnaireLLM,
		coursRepo:       coursRepo,
		examenRepo:      examenRepo,
	}
}

// Codes d'erreur examen
var (
	ErrExamenNonTrouve        = &ErreurGeneration{Code: "EXAMEN_NON_TROUVE", Message: "Examen non trouve"}
	ErrSessionExamenNonTrouvee = &ErreurGeneration{Code: "SESSION_EXAMEN_NON_TROUVEE", Message: "Session d'examen non trouvee"}
	ErrParsingExamenEchoue    = &ErreurGeneration{Code: "PARSING_EXAMEN_ECHOUE", Message: "Erreur lors du parsing de l'examen genere"}
	ErrParsingCorrectionEchoue = &ErreurGeneration{Code: "PARSING_CORRECTION_ECHOUE", Message: "Erreur lors du parsing de la correction"}
	ErrSessionExamenTerminee  = &ErreurGeneration{Code: "SESSION_EXAMEN_TERMINEE", Message: "Cette session d'examen est deja terminee"}
	ErrIndiceInvalide         = &ErreurGeneration{Code: "INDICE_INVALIDE", Message: "Index de question ou niveau d'indice invalide"}
)

// GenererExamen genere un examen blanc a partir d'un cours
func (s *ServiceExamen) GenererExamen(ctx context.Context, coursID string) (*store.ExamenBlanc, error) {
	if s.gestionnaireLLM == nil {
		return nil, ErrServiceNonDisponible
	}

	// Recuperer le cours
	cours, err := s.coursRepo.ObtenirParID(ctx, coursID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrCoursNonTrouve, err.Error())
	}

	// Verifier que le cours a du texte
	texte := cours.TexteCorrige
	if texte == "" {
		texte = cours.TexteOCR
	}
	if texte == "" {
		return nil, ErrCoursVideOCR
	}

	// Construire le prompt
	prompt := s.construirePromptExamen(texte, cours.Titre, cours.Matiere)

	// Appeler le LLM
	llmOptions := llm.OptionsGeneration{
		Temperature:   0.7,
		MaxTokens:     6000,
		FormatReponse: "json",
		SystemPrompt:  "Tu es un professeur expert en creation d'examens blancs pour collegiens et lyceens. Tu crees des evaluations completes et equilibrees. Tu reponds uniquement en JSON valide.",
	}

	reponseJSON, err := s.gestionnaireLLM.GenererJSON(ctx, prompt, nil, llmOptions)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrGenerationLLMEchouee, err.Error())
	}

	// Parser la reponse
	questions, dureeConseillee, err := s.parserReponseExamen(reponseJSON)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrParsingExamenEchoue, err.Error())
	}

	// Creer l'examen
	examen := &store.ExamenBlanc{
		CoursID:      coursID,
		Questions:    questions,
		DureeMinutes: dureeConseillee,
	}

	// Sauvegarder en base
	if s.examenRepo != nil {
		if err := s.examenRepo.CreerExamen(ctx, examen); err != nil {
			return nil, fmt.Errorf("erreur sauvegarde examen: %w", err)
		}
	}

	return examen, nil
}

// ObtenirExamen recupere un examen par son ID
func (s *ServiceExamen) ObtenirExamen(ctx context.Context, id string) (*store.ExamenBlanc, error) {
	if s.examenRepo == nil {
		return nil, ErrServiceNonDisponible
	}

	return s.examenRepo.ObtenirExamen(ctx, id)
}

// DemarrerSessionExamen cree une nouvelle session pour un examen
func (s *ServiceExamen) DemarrerSessionExamen(ctx context.Context, examenID string) (*store.SessionExamen, error) {
	if s.examenRepo == nil {
		return nil, ErrServiceNonDisponible
	}

	// Verifier que l'examen existe
	examen, err := s.examenRepo.ObtenirExamen(ctx, examenID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrExamenNonTrouve, err.Error())
	}
	if examen == nil {
		return nil, ErrExamenNonTrouve
	}

	// Creer la session
	session := &store.SessionExamen{
		ExamenID: examenID,
	}

	if err := s.examenRepo.CreerSession(ctx, session); err != nil {
		return nil, fmt.Errorf("erreur creation session examen: %w", err)
	}

	return session, nil
}

// ObtenirSessionExamen recupere une session par son ID
func (s *ServiceExamen) ObtenirSessionExamen(ctx context.Context, id string) (*store.SessionExamen, error) {
	if s.examenRepo == nil {
		return nil, ErrServiceNonDisponible
	}

	return s.examenRepo.ObtenirSession(ctx, id)
}

// GetIndice retourne un indice pour une question
func (s *ServiceExamen) GetIndice(ctx context.Context, sessionID string, questionNumero int, niveauIndice int) (*store.IndiceExamen, error) {
	if s.examenRepo == nil {
		return nil, ErrServiceNonDisponible
	}

	// Recuperer la session
	session, err := s.examenRepo.ObtenirSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrSessionExamenNonTrouvee, err.Error())
	}
	if session == nil {
		return nil, ErrSessionExamenNonTrouvee
	}

	if session.Termine {
		return nil, ErrSessionExamenTerminee
	}

	// Recuperer l'examen pour acceder aux indices
	examen, err := s.examenRepo.ObtenirExamen(ctx, session.ExamenID)
	if err != nil || examen == nil {
		return nil, ErrExamenNonTrouve
	}

	// Valider le numero de question et le niveau d'indice
	if niveauIndice < 1 || niveauIndice > 3 {
		return nil, ErrIndiceInvalide
	}

	// Trouver la question
	var question *store.QuestionExamen
	for i := range examen.Questions {
		if examen.Questions[i].Numero == questionNumero {
			question = &examen.Questions[i]
			break
		}
	}

	if question == nil {
		return nil, ErrIndiceInvalide
	}

	// Trouver l'indice du niveau demande
	var indice *store.IndiceExamen
	for i := range question.Indices {
		if question.Indices[i].Niveau == niveauIndice {
			indice = &question.Indices[i]
			break
		}
	}

	if indice == nil {
		return nil, ErrIndiceInvalide
	}

	// Enregistrer l'utilisation de l'indice
	dejaUtilise := false
	for _, iu := range session.IndicesUtilises {
		if iu.QuestionNumero == questionNumero && iu.NiveauIndice == niveauIndice {
			dejaUtilise = true
			break
		}
	}

	if !dejaUtilise {
		session.IndicesUtilises = append(session.IndicesUtilises, store.IndiceUtilise{
			QuestionNumero: questionNumero,
			NiveauIndice:   niveauIndice,
		})

		if err := s.examenRepo.MettreAJourSession(ctx, session); err != nil {
			return nil, fmt.Errorf("erreur mise a jour session: %w", err)
		}
	}

	return indice, nil
}

// ResultatCorrection contient le resultat de la correction d'un examen
type ResultatCorrection struct {
	NoteEstimee   float64                   `json:"noteEstimee"`
	PointsForts   []store.PointFort         `json:"pointsForts"`
	PointsFaibles []store.PointFaible       `json:"pointsFaibles"`
	PlanRevision  []store.EtapePlanRevision `json:"planRevision"`
	Details       []DetailCorrection        `json:"details"`
}

// DetailCorrection contient le detail de correction d'une question
type DetailCorrection struct {
	QuestionNumero  int     `json:"questionNumero"`
	ReponseEleve    string  `json:"reponseEleve"`
	ReponseAttendue string  `json:"reponseAttendue"`
	NoteQuestion    float64 `json:"noteQuestion"`
	Bareme          int     `json:"bareme"`
	Commentaire     string  `json:"commentaire"`
}

// CorrigerExamen corrige un examen et genere un plan de revision
func (s *ServiceExamen) CorrigerExamen(ctx context.Context, sessionID string, reponses []store.ReponseExamen) (*ResultatCorrection, error) {
	if s.gestionnaireLLM == nil {
		return nil, ErrServiceNonDisponible
	}
	if s.examenRepo == nil {
		return nil, ErrServiceNonDisponible
	}

	// Recuperer la session
	session, err := s.examenRepo.ObtenirSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrSessionExamenNonTrouvee, err.Error())
	}
	if session == nil {
		return nil, ErrSessionExamenNonTrouvee
	}

	if session.Termine {
		return nil, ErrSessionExamenTerminee
	}

	// Recuperer l'examen
	examen, err := s.examenRepo.ObtenirExamen(ctx, session.ExamenID)
	if err != nil || examen == nil {
		return nil, ErrExamenNonTrouve
	}

	// Construire le prompt de correction
	prompt := s.construirePromptCorrection(examen.Questions, reponses, session.IndicesUtilises)

	// Appeler le LLM
	llmOptions := llm.OptionsGeneration{
		Temperature:   0.3,
		MaxTokens:     6000,
		FormatReponse: "json",
		SystemPrompt:  "Tu es un professeur bienveillant qui corrige des examens blancs. Tu donnes des notes justes et des conseils constructifs pour aider l'eleve a progresser. Tu reponds uniquement en JSON valide.",
	}

	reponseJSON, err := s.gestionnaireLLM.GenererJSON(ctx, prompt, nil, llmOptions)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrGenerationLLMEchouee, err.Error())
	}

	// Parser la correction
	resultat, err := s.parserReponseCorrection(reponseJSON)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrParsingCorrectionEchoue, err.Error())
	}

	// Enrichir les details avec les reponses de l'eleve, les reponses attendues et le bareme
	for i := range resultat.Details {
		d := &resultat.Details[i]
		// Trouver la question correspondante
		for _, q := range examen.Questions {
			if q.Numero == d.QuestionNumero {
				d.ReponseAttendue = q.ReponseAttendue
				d.Bareme = q.Bareme
				break
			}
		}
		// Trouver la reponse de l'eleve
		for _, r := range reponses {
			if r.QuestionNumero == d.QuestionNumero {
				d.ReponseEleve = r.Texte
				break
			}
		}
	}

	// Mettre a jour la session
	session.Reponses = reponses
	session.NoteEstimee = &resultat.NoteEstimee
	session.PointsForts = resultat.PointsForts
	session.PointsFaibles = resultat.PointsFaibles
	session.PlanRevision = resultat.PlanRevision
	session.Termine = true
	nowTime := time.Now()
	session.DateFin = &nowTime

	if err := s.examenRepo.MettreAJourSession(ctx, session); err != nil {
		return nil, fmt.Errorf("erreur mise a jour session: %w", err)
	}

	return resultat, nil
}

// construirePromptExamen construit le prompt pour la generation d'examen
func (s *ServiceExamen) construirePromptExamen(texte, titre, matiere string) string {
	titreInfo := ""
	if titre != "" {
		titreInfo = fmt.Sprintf("Titre du cours : %s\n", titre)
	}
	matiereInfo := ""
	if matiere != "" {
		matiereInfo = fmt.Sprintf("Matiere : %s\n", matiere)
	}

	return fmt.Sprintf(`Tu es un professeur qui cree un examen blanc complet pour tester la maitrise d'un cours.

%s%sCours :
"""
%s
"""

Instructions :
1. Cree un examen blanc realiste avec 6 a 10 questions
2. Varie les types de questions :
   - "definition" : questions de connaissances (2-3 points)
   - "comprehension" : questions de comprehension (3-4 points)
   - "application" : exercices pratiques, calculs, analyses (4-5 points)
   - "synthese" : questions de reflexion, argumentation (5-6 points)
3. Le total des points doit etre 20 (systeme francais)
4. Varie les difficultes (facile, moyen, difficile)
5. Pour chaque question, fournis la reponse attendue detaillee
6. Pour chaque question, fournis 3 niveaux d'indices progressifs :
   - Niveau 1 : Rappel du concept concerne (aide legere)
   - Niveau 2 : Indication methodologique (aide moyenne)
   - Niveau 3 : Debut de la reponse (aide forte)
7. La duree conseillee doit etre realiste (20-45 minutes)
8. Base-toi UNIQUEMENT sur le contenu du cours fourni
9. Les questions doivent etre bienveillantes et encourageantes

Reponds UNIQUEMENT avec un JSON valide au format suivant, sans texte avant ou apres :
{
  "questions": [
    {
      "numero": 1,
      "type": "definition",
      "difficulte": "facile",
      "enonce": "...",
      "bareme": 2,
      "reponse_attendue": "...",
      "indices": [
        { "niveau": 1, "texte": "Rappel: ce concept concerne..." },
        { "niveau": 2, "texte": "Methodologie: commencez par..." },
        { "niveau": 3, "texte": "La reponse commence par..." }
      ]
    }
  ],
  "duree_conseillee": 30
}`, titreInfo, matiereInfo, texte)
}

// reponseExamenJSON represente la structure de reponse du LLM pour les examens
type reponseExamenJSON struct {
	Questions []struct {
		Numero          int    `json:"numero"`
		Type            string `json:"type"`
		Difficulte      string `json:"difficulte"`
		Enonce          string `json:"enonce"`
		Bareme          int    `json:"bareme"`
		ReponseAttendue string `json:"reponse_attendue"`
		Indices         []struct {
			Niveau int    `json:"niveau"`
			Texte  string `json:"texte"`
		} `json:"indices"`
	} `json:"questions"`
	DureeConseillee int `json:"duree_conseillee"`
}

// parserReponseExamen parse la reponse JSON du LLM en questions d'examen
func (s *ServiceExamen) parserReponseExamen(reponseJSON []byte) ([]store.QuestionExamen, int, error) {
	var reponse reponseExamenJSON
	if err := json.Unmarshal(reponseJSON, &reponse); err != nil {
		return nil, 0, fmt.Errorf("erreur parsing JSON: %w", err)
	}

	questions := make([]store.QuestionExamen, 0, len(reponse.Questions))
	for _, q := range reponse.Questions {
		// Valider le type
		typeQuestion := q.Type
		if typeQuestion != "definition" && typeQuestion != "comprehension" &&
			typeQuestion != "application" && typeQuestion != "synthese" {
			typeQuestion = "comprehension"
		}

		// Valider la difficulte
		difficulte := q.Difficulte
		if difficulte != "facile" && difficulte != "moyen" && difficulte != "difficile" {
			difficulte = "moyen"
		}

		// Convertir les indices
		indices := make([]store.IndiceExamen, 0, len(q.Indices))
		for _, idx := range q.Indices {
			if idx.Niveau >= 1 && idx.Niveau <= 3 {
				indices = append(indices, store.IndiceExamen{
					Niveau: idx.Niveau,
					Texte:  idx.Texte,
				})
			}
		}

		question := store.QuestionExamen{
			Numero:          q.Numero,
			Type:            typeQuestion,
			Difficulte:      difficulte,
			Enonce:          q.Enonce,
			Bareme:          q.Bareme,
			ReponseAttendue: q.ReponseAttendue,
			Indices:         indices,
		}
		questions = append(questions, question)
	}

	duree := reponse.DureeConseillee
	if duree <= 0 {
		duree = 30
	}

	return questions, duree, nil
}

// construirePromptCorrection construit le prompt pour la correction
func (s *ServiceExamen) construirePromptCorrection(questions []store.QuestionExamen, reponses []store.ReponseExamen, indicesUtilises []store.IndiceUtilise) string {
	// Construire la description des questions et reponses
	questionsStr := ""
	for _, q := range questions {
		// Trouver la reponse de l'eleve
		reponseEleve := "(pas de reponse)"
		for _, r := range reponses {
			if r.QuestionNumero == q.Numero {
				reponseEleve = r.Texte
				if reponseEleve == "" {
					reponseEleve = "(pas de reponse)"
				}
				break
			}
		}

		// Compter les indices utilises pour cette question
		nbIndices := 0
		for _, iu := range indicesUtilises {
			if iu.QuestionNumero == q.Numero {
				nbIndices++
			}
		}

		questionsStr += fmt.Sprintf(`
Question %d (%s, %s, %d points) :
Enonce : %s
Reponse attendue : %s
Reponse de l'eleve : %s
Indices utilises : %d/3
`, q.Numero, q.Type, q.Difficulte, q.Bareme, q.Enonce, q.ReponseAttendue, reponseEleve, nbIndices)
	}

	return fmt.Sprintf(`Tu es un professeur bienveillant qui corrige un examen blanc. Tu dois evaluer les reponses de l'eleve, attribuer une note sur 20 et donner des conseils constructifs.

Examen et reponses de l'eleve :
%s

Instructions :
1. Evalue chaque reponse par rapport a la reponse attendue
2. Attribue une note sur le bareme pour chaque question (peut etre partielle : 1.5/3 par exemple)
3. Si l'eleve a utilise des indices, tu peux reduire legerement la note (mais reste bienveillant)
4. Calcule la note totale sur 20
5. Identifie 2-3 points forts de l'eleve
6. Identifie 2-3 concepts a revoir
7. Propose un plan de revision personnalise avec 3-5 actions concretes
8. Sois encourageant et constructif dans tes commentaires

Reponds UNIQUEMENT avec un JSON valide au format suivant, sans texte avant ou apres :
{
  "note_estimee": 14.5,
  "details": [
    {
      "question_numero": 1,
      "note_question": 1.5,
      "commentaire": "Bonne comprehension du concept mais la formulation peut etre amelioree..."
    }
  ],
  "points_forts": [
    {
      "concept": "Les definitions de base",
      "commentaire": "Tu maitrises bien les concepts fondamentaux"
    }
  ],
  "points_faibles": [
    {
      "concept": "L'application pratique",
      "commentaire": "Revois les exercices d'application pour mieux maitriser la methode"
    }
  ],
  "plan_revision": [
    {
      "priorite": 1,
      "action": "Revoir les exercices d'application",
      "concept": "Application pratique",
      "ressource": "Refaire les exercices du cours en se concentrant sur la methode"
    }
  ]
}`, questionsStr)
}

// reponseCorrectionJSON represente la structure de reponse du LLM pour la correction
type reponseCorrectionJSON struct {
	NoteEstimee float64 `json:"note_estimee"`
	Details     []struct {
		QuestionNumero int     `json:"question_numero"`
		NoteQuestion   float64 `json:"note_question"`
		Commentaire    string  `json:"commentaire"`
	} `json:"details"`
	PointsForts []struct {
		Concept     string `json:"concept"`
		Commentaire string `json:"commentaire"`
	} `json:"points_forts"`
	PointsFaibles []struct {
		Concept     string `json:"concept"`
		Commentaire string `json:"commentaire"`
	} `json:"points_faibles"`
	PlanRevision []struct {
		Priorite  int    `json:"priorite"`
		Action    string `json:"action"`
		Concept   string `json:"concept"`
		Ressource string `json:"ressource"`
	} `json:"plan_revision"`
}

// parserReponseCorrection parse la reponse JSON du LLM pour la correction
func (s *ServiceExamen) parserReponseCorrection(reponseJSON []byte) (*ResultatCorrection, error) {
	var reponse reponseCorrectionJSON
	if err := json.Unmarshal(reponseJSON, &reponse); err != nil {
		return nil, fmt.Errorf("erreur parsing JSON: %w", err)
	}

	// Convertir les details
	details := make([]DetailCorrection, 0, len(reponse.Details))
	for _, d := range reponse.Details {
		details = append(details, DetailCorrection{
			QuestionNumero: d.QuestionNumero,
			NoteQuestion:   d.NoteQuestion,
			Bareme:         0, // sera rempli par l'appelant
			Commentaire:    d.Commentaire,
		})
	}

	// Convertir les points forts
	pointsForts := make([]store.PointFort, 0, len(reponse.PointsForts))
	for _, pf := range reponse.PointsForts {
		pointsForts = append(pointsForts, store.PointFort{
			Concept:     pf.Concept,
			Commentaire: pf.Commentaire,
		})
	}

	// Convertir les points faibles
	pointsFaibles := make([]store.PointFaible, 0, len(reponse.PointsFaibles))
	for _, pf := range reponse.PointsFaibles {
		pointsFaibles = append(pointsFaibles, store.PointFaible{
			Concept:     pf.Concept,
			Commentaire: pf.Commentaire,
		})
	}

	// Convertir le plan de revision
	planRevision := make([]store.EtapePlanRevision, 0, len(reponse.PlanRevision))
	for _, pr := range reponse.PlanRevision {
		planRevision = append(planRevision, store.EtapePlanRevision{
			Priorite:  pr.Priorite,
			Action:    pr.Action,
			Concept:   pr.Concept,
			Ressource: pr.Ressource,
		})
	}

	return &ResultatCorrection{
		NoteEstimee:   reponse.NoteEstimee,
		PointsForts:   pointsForts,
		PointsFaibles: pointsFaibles,
		PlanRevision:  planRevision,
		Details:       details,
	}, nil
}
