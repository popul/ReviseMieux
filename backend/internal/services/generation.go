// Package services contient les services métier de l'application
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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

// ServiceGeneration gère la génération de contenu (fiches, quiz, mindmaps, ressources)
type ServiceGeneration struct {
	gestionnaireLLM  *llm.GestionnaireLLM
	coursRepo        store.CoursRepository
	fichesRepo       store.FichesRepository
	quizRepo         store.QuizRepository
	ressourcesRepo   store.RessourcesRepository
}

// NouveauServiceGeneration crée une nouvelle instance du service de génération
func NouveauServiceGeneration(
	gestionnaireLLM *llm.GestionnaireLLM,
	coursRepo store.CoursRepository,
	fichesRepo store.FichesRepository,
	quizRepo store.QuizRepository,
	ressourcesRepo store.RessourcesRepository,
) *ServiceGeneration {
	return &ServiceGeneration{
		gestionnaireLLM:  gestionnaireLLM,
		coursRepo:        coursRepo,
		fichesRepo:       fichesRepo,
		quizRepo:         quizRepo,
		ressourcesRepo:   ressourcesRepo,
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

// --- Génération de Quiz ---

// OptionsGenerationQuiz contient les options pour la génération de quiz
type OptionsGenerationQuiz struct {
	// NombreQuestions est le nombre de questions à générer (par défaut: 10)
	NombreQuestions int
	// Difficulte définit la difficulté des questions ("facile", "moyen", "difficile", "" pour mélangé)
	Difficulte string
	// Titre optionnel pour le quiz
	Titre string
}

// ResultatGenerationQuiz contient le résultat de la génération
type ResultatGenerationQuiz struct {
	Quiz *store.Quiz `json:"quiz"`
}

// Codes d'erreur quiz
var (
	ErrParsingQuizEchoue = &ErreurGeneration{Code: "PARSING_QUIZ_ECHOUE", Message: "Erreur lors du parsing du quiz généré"}
	ErrQuizNonTrouve     = &ErreurGeneration{Code: "QUIZ_NON_TROUVE", Message: "Quiz non trouvé"}
	ErrSessionNonTrouvee = &ErreurGeneration{Code: "SESSION_NON_TROUVEE", Message: "Session de quiz non trouvée"}
)

// GenererQuiz génère un quiz pour un cours
func (s *ServiceGeneration) GenererQuiz(ctx context.Context, coursID string, options *OptionsGenerationQuiz) (*ResultatGenerationQuiz, error) {
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

	// Appliquer les valeurs par défaut
	if options == nil {
		options = &OptionsGenerationQuiz{}
	}
	if options.NombreQuestions <= 0 {
		options.NombreQuestions = 10
	}
	if options.NombreQuestions > 20 {
		options.NombreQuestions = 20
	}

	// Construire le prompt
	prompt := s.construirePromptQuiz(texte, options)

	// Appeler le LLM
	llmOptions := llm.OptionsGeneration{
		Temperature:   0.7,
		MaxTokens:     4000,
		FormatReponse: "json",
		SystemPrompt:  "Tu es un professeur expert en création de QCM pour lycéens. Tu réponds uniquement en JSON valide.",
	}

	reponseJSON, err := s.gestionnaireLLM.GenererJSON(ctx, prompt, nil, llmOptions)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrGenerationLLMEchouee, err.Error())
	}

	// Parser la réponse
	questions, err := s.parserReponseQuiz(reponseJSON)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrParsingQuizEchoue, err.Error())
	}

	// Créer le quiz
	titre := options.Titre
	if titre == "" {
		titre = fmt.Sprintf("Quiz - %s", cours.Titre)
	}

	difficulte := options.Difficulte
	if difficulte == "" {
		difficulte = "moyen"
	}

	quiz := &store.Quiz{
		CoursID:         coursID,
		Titre:           titre,
		Difficulte:      difficulte,
		NombreQuestions: len(questions),
		Questions:       questions,
	}

	// Sauvegarder le quiz en base
	if s.quizRepo != nil {
		if err := s.quizRepo.Creer(ctx, quiz); err != nil {
			return nil, fmt.Errorf("erreur sauvegarde quiz: %w", err)
		}
	}

	return &ResultatGenerationQuiz{
		Quiz: quiz,
	}, nil
}

// construirePromptQuiz construit le prompt pour la génération de quiz
func (s *ServiceGeneration) construirePromptQuiz(texte string, options *OptionsGenerationQuiz) string {
	difficulte := "mélangée (facile, moyen, difficile)"
	if options.Difficulte != "" {
		difficulte = options.Difficulte
	}

	return fmt.Sprintf(`Tu es un professeur créant un QCM pour tester la compréhension d'un cours.

Cours :
"""
%s
"""

Paramètres :
- Nombre de questions : %d
- Difficulté : %s

Instructions :
1. Crée des questions de compréhension (pas de piège)
2. 4 choix par question, 1 seul correct
3. Les mauvaises réponses doivent être plausibles mais clairement fausses
4. Ajoute une explication pédagogique pour chaque question
5. Base-toi UNIQUEMENT sur le contenu du cours fourni

Réponds UNIQUEMENT avec un JSON valide au format suivant, sans texte avant ou après :
{
  "questions": [
    {
      "enonce": "...",
      "choix": ["A", "B", "C", "D"],
      "reponse_correcte": 0,
      "explication": "..."
    }
  ]
}`, texte, options.NombreQuestions, difficulte)
}

// reponseQuizJSON représente la structure de réponse du LLM pour les quiz
type reponseQuizJSON struct {
	Questions []struct {
		Enonce          string   `json:"enonce"`
		Choix           []string `json:"choix"`
		ReponseCorrecte int      `json:"reponse_correcte"`
		Explication     string   `json:"explication"`
	} `json:"questions"`
}

// parserReponseQuiz parse la réponse JSON du LLM en questions
func (s *ServiceGeneration) parserReponseQuiz(reponseJSON []byte) ([]store.Question, error) {
	var reponse reponseQuizJSON
	if err := json.Unmarshal(reponseJSON, &reponse); err != nil {
		return nil, fmt.Errorf("erreur parsing JSON: %w", err)
	}

	questions := make([]store.Question, 0, len(reponse.Questions))
	for i, q := range reponse.Questions {
		// Valider le nombre de choix
		if len(q.Choix) != 4 {
			continue // Ignorer les questions mal formées
		}

		// Valider l'index de la réponse correcte
		reponseCorrecte := q.ReponseCorrecte
		if reponseCorrecte < 0 || reponseCorrecte > 3 {
			reponseCorrecte = 0
		}

		question := store.Question{
			ID:              fmt.Sprintf("q%d", i+1),
			Enonce:          q.Enonce,
			Choix:           q.Choix,
			ReponseCorrecte: reponseCorrecte,
			Explication:     q.Explication,
		}
		questions = append(questions, question)
	}

	return questions, nil
}

// --- Sessions de Quiz ---

// DemarrerSession crée une nouvelle session pour un quiz
func (s *ServiceGeneration) DemarrerSession(ctx context.Context, quizID string) (*store.QuizSession, error) {
	if s.quizRepo == nil {
		return nil, ErrServiceNonDisponible
	}

	// Vérifier que le quiz existe
	quiz, err := s.quizRepo.ObtenirParID(ctx, quizID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrQuizNonTrouve, err.Error())
	}
	if quiz == nil {
		return nil, ErrQuizNonTrouve
	}

	// Créer la session
	session := &store.QuizSession{
		QuizID: quizID,
	}

	if err := s.quizRepo.CreerSession(ctx, session); err != nil {
		return nil, fmt.Errorf("erreur création session: %w", err)
	}

	return session, nil
}

// RepondreQuestion enregistre la réponse à une question
func (s *ServiceGeneration) RepondreQuestion(ctx context.Context, sessionID string, questionID string, choixIndex int) (*store.QuizSession, bool, error) {
	if s.quizRepo == nil {
		return nil, false, ErrServiceNonDisponible
	}

	// Récupérer la session
	session, err := s.quizRepo.ObtenirSession(ctx, sessionID)
	if err != nil {
		return nil, false, fmt.Errorf("%w: %s", ErrSessionNonTrouvee, err.Error())
	}
	if session == nil {
		return nil, false, ErrSessionNonTrouvee
	}

	if session.Termine {
		return nil, false, &ErreurGeneration{Code: "SESSION_TERMINEE", Message: "Cette session est déjà terminée"}
	}

	// Récupérer le quiz pour vérifier la réponse
	quiz, err := s.quizRepo.ObtenirParID(ctx, session.QuizID)
	if err != nil || quiz == nil {
		return nil, false, ErrQuizNonTrouve
	}

	// Trouver la question et vérifier la réponse
	var estCorrecte bool
	for _, q := range quiz.Questions {
		if q.ID == questionID {
			estCorrecte = (choixIndex == q.ReponseCorrecte)
			break
		}
	}

	// Ajouter la réponse
	session.Reponses = append(session.Reponses, store.ReponseSession{
		QuestionID:  questionID,
		ChoixIndex:  choixIndex,
		EstCorrecte: estCorrecte,
	})

	// Mettre à jour la session
	if err := s.quizRepo.MettreAJourSession(ctx, session); err != nil {
		return nil, false, fmt.Errorf("erreur mise à jour session: %w", err)
	}

	return session, estCorrecte, nil
}

// TerminerSession termine une session et calcule le score
func (s *ServiceGeneration) TerminerSession(ctx context.Context, sessionID string) (*store.QuizSession, error) {
	if s.quizRepo == nil {
		return nil, ErrServiceNonDisponible
	}

	// Récupérer la session
	session, err := s.quizRepo.ObtenirSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrSessionNonTrouvee, err.Error())
	}
	if session == nil {
		return nil, ErrSessionNonTrouvee
	}

	if session.Termine {
		return session, nil // Déjà terminée
	}

	// Calculer le score
	var correctes int
	for _, r := range session.Reponses {
		if r.EstCorrecte {
			correctes++
		}
	}

	var score float64
	if len(session.Reponses) > 0 {
		score = float64(correctes) / float64(len(session.Reponses)) * 100
	}

	now := time.Now()
	session.Score = &score
	session.Termine = true
	session.DateFin = &now

	// Mettre à jour la session
	if err := s.quizRepo.MettreAJourSession(ctx, session); err != nil {
		return nil, fmt.Errorf("erreur mise à jour session: %w", err)
	}

	return session, nil
}

// ObtenirQuizParID récupère un quiz par son ID
func (s *ServiceGeneration) ObtenirQuizParID(ctx context.Context, quizID string) (*store.Quiz, error) {
	if s.quizRepo == nil {
		return nil, ErrServiceNonDisponible
	}

	return s.quizRepo.ObtenirParID(ctx, quizID)
}

// ObtenirSession récupère une session par son ID
func (s *ServiceGeneration) ObtenirSession(ctx context.Context, sessionID string) (*store.QuizSession, error) {
	if s.quizRepo == nil {
		return nil, ErrServiceNonDisponible
	}

	return s.quizRepo.ObtenirSession(ctx, sessionID)
}

// ListerQuizParCours récupère tous les quiz d'un cours
func (s *ServiceGeneration) ListerQuizParCours(ctx context.Context, coursID string) ([]*store.Quiz, error) {
	if s.quizRepo == nil {
		return nil, ErrServiceNonDisponible
	}

	return s.quizRepo.ListerParCours(ctx, coursID)
}

// --- Génération de Ressources ---

// ResultatGenerationRessources contient le résultat de la génération de ressources
type ResultatGenerationRessources struct {
	Ressources    []*store.Ressource `json:"ressources"`
	NombreGenere  int                `json:"nombreGenere"`
	Avertissement string             `json:"avertissement"`
}

// Codes d'erreur ressources
var (
	ErrParsingRessourcesEchoue = &ErreurGeneration{Code: "PARSING_RESSOURCES_ECHOUE", Message: "Erreur lors du parsing des ressources générées"}
)

// GenererRessources génère des suggestions de ressources pour un cours
func (s *ServiceGeneration) GenererRessources(ctx context.Context, coursID string) (*ResultatGenerationRessources, error) {
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
	prompt := s.construirePromptRessources(texte, cours.Titre, cours.Matiere)

	// Appeler le LLM
	llmOptions := llm.OptionsGeneration{
		Temperature:   0.7,
		MaxTokens:     2000,
		FormatReponse: "json",
		SystemPrompt:  "Tu es un assistant pédagogique expert en recherche de ressources éducatives pour lycéens. Tu réponds uniquement en JSON valide.",
	}

	reponseJSON, err := s.gestionnaireLLM.GenererJSON(ctx, prompt, nil, llmOptions)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrGenerationLLMEchouee, err.Error())
	}

	// Parser la réponse
	ressources, err := s.parserReponseRessources(reponseJSON, coursID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrParsingRessourcesEchoue, err.Error())
	}

	// Sauvegarder les ressources en base
	if s.ressourcesRepo != nil {
		if err := s.ressourcesRepo.CreerPlusieurs(ctx, ressources); err != nil {
			return nil, fmt.Errorf("erreur sauvegarde ressources: %w", err)
		}
	}

	return &ResultatGenerationRessources{
		Ressources:    ressources,
		NombreGenere:  len(ressources),
		Avertissement: "Les liens suggérés sont générés par IA et doivent être vérifiés avant utilisation.",
	}, nil
}

// construirePromptRessources construit le prompt pour la génération de ressources
func (s *ServiceGeneration) construirePromptRessources(texte, titre, matiere string) string {
	matiereInfo := ""
	if matiere != "" {
		matiereInfo = fmt.Sprintf("Matière : %s\n", matiere)
	}
	titreInfo := ""
	if titre != "" {
		titreInfo = fmt.Sprintf("Titre du cours : %s\n", titre)
	}

	return fmt.Sprintf(`Tu es un assistant pédagogique qui suggère des ressources complémentaires pour aider un lycéen à approfondir un cours.

%s%sCours :
"""
%s
"""

Instructions :
1. Suggère entre 5 et 10 ressources pertinentes et éducatives
2. Varie les types : vidéos YouTube éducatives, articles de référence, sites pédagogiques
3. Privilégie les ressources en français quand possible
4. Choisis des sources fiables : Khan Academy, Les Bons Profs, Lumni, Wikipedia, sites .edu/.gouv
5. Décris brièvement chaque ressource (1-2 phrases)
6. Pour les vidéos YouTube, suggère des chaînes éducatives reconnues
7. Les URLs doivent être plausibles mais n'ont pas besoin d'être vérifiées (l'utilisateur sera averti)

Types de ressources à inclure :
- video : vidéos YouTube ou cours en ligne
- article : articles Wikipedia, encyclopédies, blogs éducatifs
- site : sites web éducatifs, plateformes d'apprentissage

Réponds UNIQUEMENT avec un JSON valide au format suivant, sans texte avant ou après :
{
  "ressources": [
    {
      "titre": "Nom de la ressource",
      "url": "https://...",
      "type": "video|article|site",
      "description": "Brève description de la ressource"
    }
  ]
}`, titreInfo, matiereInfo, texte)
}

// reponseRessourcesJSON représente la structure de réponse du LLM pour les ressources
type reponseRessourcesJSON struct {
	Ressources []struct {
		Titre       string `json:"titre"`
		URL         string `json:"url"`
		Type        string `json:"type"`
		Description string `json:"description"`
	} `json:"ressources"`
}

// parserReponseRessources parse la réponse JSON du LLM en ressources
func (s *ServiceGeneration) parserReponseRessources(reponseJSON []byte, coursID string) ([]*store.Ressource, error) {
	var reponse reponseRessourcesJSON
	if err := json.Unmarshal(reponseJSON, &reponse); err != nil {
		return nil, fmt.Errorf("erreur parsing JSON: %w", err)
	}

	ressources := make([]*store.Ressource, 0, len(reponse.Ressources))
	for _, r := range reponse.Ressources {
		// Valider le type de ressource
		typeRessource := store.TypeRessource(r.Type)
		if typeRessource != store.TypeRessourceVideo &&
			typeRessource != store.TypeRessourceArticle &&
			typeRessource != store.TypeRessourceSite {
			typeRessource = store.TypeRessourceSite // valeur par défaut
		}

		ressource := &store.Ressource{
			CoursID:     coursID,
			Titre:       r.Titre,
			URL:         r.URL,
			Type:        typeRessource,
			Description: r.Description,
		}
		ressources = append(ressources, ressource)
	}

	return ressources, nil
}

// ObtenirRessourcesParCours récupère les ressources existantes d'un cours
func (s *ServiceGeneration) ObtenirRessourcesParCours(ctx context.Context, coursID string) ([]*store.Ressource, error) {
	if s.ressourcesRepo == nil {
		return nil, ErrServiceNonDisponible
	}

	return s.ressourcesRepo.ListerParCours(ctx, coursID)
}
