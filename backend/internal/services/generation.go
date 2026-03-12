// Package services contient les services métier de l'application
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
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

// ServiceGeneration gère la génération de contenu (fiches, quiz, mindmaps)
type ServiceGeneration struct {
	gestionnaireLLM  *llm.GestionnaireLLM
	coursRepo        store.CoursRepository
	fichesRepo       store.FichesRepository
	quizRepo         store.QuizRepository
	mindmapRepo      store.MindmapRepository
	conceptsRepo     store.ConceptsRepository
}

// NouveauServiceGeneration crée une nouvelle instance du service de génération
func NouveauServiceGeneration(
	gestionnaireLLM *llm.GestionnaireLLM,
	coursRepo store.CoursRepository,
	fichesRepo store.FichesRepository,
	quizRepo store.QuizRepository,
	mindmapRepo store.MindmapRepository,
	conceptsRepo store.ConceptsRepository,
) *ServiceGeneration {
	return &ServiceGeneration{
		gestionnaireLLM:  gestionnaireLLM,
		coursRepo:        coursRepo,
		fichesRepo:       fichesRepo,
		quizRepo:         quizRepo,
		mindmapRepo:      mindmapRepo,
		conceptsRepo:     conceptsRepo,
	}
}

// obtenirTexteGeneration retourne le résumé formaté si disponible, sinon le texte complet.
// Le résumé est ~80% plus court et suffit pour générer fiches, quiz et mindmap.
func (s *ServiceGeneration) obtenirTexteGeneration(cours *store.Cours) string {
	if cours.Resume != nil && len(cours.Resume) > 0 && string(cours.Resume) != "null" {
		var resume ResultatResume
		if err := json.Unmarshal(cours.Resume, &resume); err == nil {
			return formaterResumePourGeneration(resume)
		}
	}
	texte := cours.TexteCorrige
	if texte == "" {
		texte = cours.TexteOCR
	}
	return texte
}

// formaterResumePourGeneration formate le résumé de manière dense pour servir d'entrée LLM
func formaterResumePourGeneration(resume ResultatResume) string {
	var sb strings.Builder

	if len(resume.PointsCles) > 0 {
		sb.WriteString("Points clés :\n")
		for _, p := range resume.PointsCles {
			sb.WriteString("- ")
			sb.WriteString(p)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	for _, section := range resume.Structure {
		sb.WriteString("## ")
		sb.WriteString(section.Titre)
		sb.WriteString("\n")
		sb.WriteString(section.Contenu)
		sb.WriteString("\n\n")
	}

	if resume.Paragraphe != "" {
		sb.WriteString("Synthèse : ")
		sb.WriteString(resume.Paragraphe)
	}

	return sb.String()
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

	// Utiliser le résumé si disponible, sinon le texte complet
	texte := s.obtenirTexteGeneration(cours)
	if texte == "" {
		return nil, ErrCoursVideOCR
	}

	// Fetch existing concepts for this course
	var concepts []*store.Concept
	if s.conceptsRepo != nil {
		concepts, _ = s.conceptsRepo.ListerParCours(ctx, coursID)
	}

	// Construire le prompt
	prompt := s.construirePromptFiches(texte, options, concepts)

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
	fiches, err := s.parserReponseFiches(reponseJSON, coursID, concepts)
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
func (s *ServiceGeneration) construirePromptFiches(texte string, options *OptionsGenerationFiches, concepts []*store.Concept) string {
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

	conceptsSection := ""
	if len(concepts) > 0 {
		conceptsSection = "\nConcepts clés identifiés dans ce cours:\n"
		for _, c := range concepts {
			conceptsSection += fmt.Sprintf("- %s (%s): %s\n", c.Nom, c.Importance, c.Definition)
		}
		conceptsSection += "\nPour chaque fiche, indique le(s) concept(s) associé(s) parmi cette liste dans le champ \"concepts\".\n"
	}

	conceptsField := ""
	if len(concepts) > 0 {
		conceptsField = `,
      "concepts": ["Nom du concept"]`
	}

	return fmt.Sprintf(`Cours :
"""%s"""
%s
Consignes :
%s%s- Questions favorisant le rappel actif (faits, concepts, cause/effet)
- Réponses concises avec exemples concrets quand pertinent (formules, règles)
- Basé UNIQUEMENT sur le contenu fourni

JSON attendu :
{"fiches": [{"question": "...", "reponse": "...", "difficulte": "facile|moyen|difficile"%s}]}`, texte, conceptsSection, nombreFiches, difficulte, conceptsField)
}

// reponseFichesJSON représente la structure de réponse du LLM
type reponseFichesJSON struct {
	Fiches []struct {
		Question   string   `json:"question"`
		Reponse    string   `json:"reponse"`
		Difficulte string   `json:"difficulte"`
		Concepts   []string `json:"concepts"`
	} `json:"fiches"`
}

// parserReponseFiches parse la réponse JSON du LLM en fiches
func (s *ServiceGeneration) parserReponseFiches(reponseJSON []byte, coursID string, concepts []*store.Concept) ([]*store.Fiche, error) {
	var reponse reponseFichesJSON
	if err := json.Unmarshal(reponseJSON, &reponse); err != nil {
		return nil, fmt.Errorf("erreur parsing JSON: %w", err)
	}

	// Build a name-to-ID map for concept matching
	conceptNameToID := make(map[string]string)
	for _, c := range concepts {
		conceptNameToID[strings.ToLower(c.Nom)] = c.ID
	}

	fiches := make([]*store.Fiche, 0, len(reponse.Fiches))
	for i, f := range reponse.Fiches {
		// Valider la difficulté
		difficulte := f.Difficulte
		if difficulte != "facile" && difficulte != "moyen" && difficulte != "difficile" {
			difficulte = "moyen" // valeur par défaut
		}

		// Map concept names to IDs
		var conceptIDs []string
		for _, nomConcept := range f.Concepts {
			if id, ok := conceptNameToID[strings.ToLower(nomConcept)]; ok {
				conceptIDs = append(conceptIDs, id)
			}
		}

		fiche := &store.Fiche{
			CoursID:    coursID,
			Question:   f.Question,
			Reponse:    f.Reponse,
			Difficulte: difficulte,
			Ordre:      i + 1,
			ConceptIDs: conceptIDs,
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

	// Utiliser le résumé si disponible, sinon le texte complet
	texte := s.obtenirTexteGeneration(cours)
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

	return fmt.Sprintf(`Cours :
"""%s"""

%d questions, difficulté : %s
- 4 choix par question, 1 correct, mauvaises réponses plausibles
- Explication pédagogique pour chaque question
- Basé UNIQUEMENT sur le contenu fourni

JSON attendu :
{"questions": [{"enonce": "...", "choix": ["A","B","C","D"], "reponse_correcte": 0, "explication": "..."}]}`, texte, options.NombreQuestions, difficulte)
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

		// Mélanger les choix et mettre à jour l'index de la bonne réponse
		choixMelanges, nouvelIndex := melangerChoix(q.Choix, reponseCorrecte)

		question := store.Question{
			ID:              fmt.Sprintf("q%d", i+1),
			Enonce:          q.Enonce,
			Choix:           choixMelanges,
			ReponseCorrecte: nouvelIndex,
			Explication:     q.Explication,
		}
		questions = append(questions, question)
	}

	return questions, nil
}

// melangerChoix mélange les choix d'une question et retourne le nouvel index de la bonne réponse
func melangerChoix(choix []string, indexCorrect int) ([]string, int) {
	n := len(choix)
	if n == 0 {
		return choix, indexCorrect
	}

	// Créer une copie des choix avec leurs indices originaux
	type choixAvecIndex struct {
		texte         string
		indexOriginal int
	}

	choixIndexes := make([]choixAvecIndex, n)
	for i, c := range choix {
		choixIndexes[i] = choixAvecIndex{texte: c, indexOriginal: i}
	}

	// Mélanger avec Fisher-Yates
	for i := n - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		choixIndexes[i], choixIndexes[j] = choixIndexes[j], choixIndexes[i]
	}

	// Reconstruire les choix et trouver le nouvel index
	resultat := make([]string, n)
	nouvelIndex := 0
	for i, c := range choixIndexes {
		resultat[i] = c.texte
		if c.indexOriginal == indexCorrect {
			nouvelIndex = i
		}
	}

	return resultat, nouvelIndex
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

// --- Génération de Mindmaps ---

// ResultatGenerationMindmap contient le résultat de la génération de mindmap
type ResultatGenerationMindmap struct {
	Mindmap *store.Mindmap `json:"mindmap"`
}

// Codes d'erreur mindmap
var (
	ErrParsingMindmapEchoue = &ErreurGeneration{Code: "PARSING_MINDMAP_ECHOUE", Message: "Erreur lors du parsing de la mindmap générée"}
	ErrMindmapNonTrouvee    = &ErreurGeneration{Code: "MINDMAP_NON_TROUVEE", Message: "Mindmap non trouvée"}
)

// GenererMindmap génère une carte mentale pour un cours
func (s *ServiceGeneration) GenererMindmap(ctx context.Context, coursID string) (*ResultatGenerationMindmap, error) {
	if s.gestionnaireLLM == nil {
		return nil, ErrServiceNonDisponible
	}

	// Récupérer le cours
	cours, err := s.coursRepo.ObtenirParID(ctx, coursID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrCoursNonTrouve, err.Error())
	}

	// Utiliser le résumé si disponible, sinon le texte complet
	texte := s.obtenirTexteGeneration(cours)
	if texte == "" {
		return nil, ErrCoursVideOCR
	}

	// Récupérer les concepts existants pour ce cours
	var concepts []*store.Concept
	if s.conceptsRepo != nil {
		concepts, _ = s.conceptsRepo.ListerParCours(ctx, coursID)
	}

	// Construire le prompt
	prompt := s.construirePromptMindmap(texte, cours.Titre, concepts)

	// Appeler le LLM
	llmOptions := llm.OptionsGeneration{
		Temperature:   0.7,
		MaxTokens:     4000,
		FormatReponse: "json",
		SystemPrompt:  "Tu es un professeur expert en création de cartes mentales pour lycéens. Tu structures les connaissances de manière hiérarchique et visuelle. Tu réponds uniquement en JSON valide.",
	}

	reponseJSON, err := s.gestionnaireLLM.GenererJSON(ctx, prompt, nil, llmOptions)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrGenerationLLMEchouee, err.Error())
	}

	// Parser la réponse
	noeuds, liens, err := s.parserReponseMindmap(reponseJSON)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrParsingMindmapEchoue, err.Error())
	}

	// Créer la mindmap
	mindmap := &store.Mindmap{
		CoursID: coursID,
		Noeuds:  noeuds,
		Liens:   liens,
	}

	// Sauvegarder la mindmap en base
	if s.mindmapRepo != nil {
		if err := s.mindmapRepo.Creer(ctx, mindmap); err != nil {
			return nil, fmt.Errorf("erreur sauvegarde mindmap: %w", err)
		}
	}

	return &ResultatGenerationMindmap{
		Mindmap: mindmap,
	}, nil
}

// construirePromptMindmap construit le prompt pour la génération de mindmap
func (s *ServiceGeneration) construirePromptMindmap(texte, titre string, concepts []*store.Concept) string {
	titreInfo := ""
	if titre != "" {
		titreInfo = fmt.Sprintf("Titre du cours : %s\n\n", titre)
	}

	conceptsSection := ""
	if len(concepts) > 0 {
		conceptsSection = "\nConcepts clés identifiés dans ce cours (avec leurs IDs):\n"
		for _, c := range concepts {
			conceptsSection += fmt.Sprintf("- %s (ID: %s, %s)\n", c.Nom, c.ID, c.Importance)
		}
		conceptsSection += "\nPour chaque noeud de type \"branche\" ou \"feuille\", associe un \"concept_id\" correspondant si le noeud correspond à un des concepts ci-dessus. Utilise l'ID exact du concept. Si aucun concept ne correspond, laisse concept_id vide.\n"
	}

	return fmt.Sprintf(`%sCours :
"""%s"""
%s
Consignes :
- 1 nœud central, 3-6 branches, 2-4 feuilles par branche (20-30 nœuds max)
- Labels courts (max 4-5 mots), basé UNIQUEMENT sur le contenu fourni
- Types : "central" (position 400,300), "branche", "feuille"

JSON attendu :
{"noeuds": [{"id": "central", "label": "Thème", "type": "central", "position": {"x": 400, "y": 300}, "concept_id": ""}], "liens": [{"source": "central", "target": "branche1"}]}`, titreInfo, texte, conceptsSection)
}

// reponseMindmapJSON représente la structure de réponse du LLM pour les mindmaps
type reponseMindmapJSON struct {
	Noeuds []struct {
		ID       string `json:"id"`
		Label    string `json:"label"`
		Type     string `json:"type"`
		Position struct {
			X float64 `json:"x"`
			Y float64 `json:"y"`
		} `json:"position"`
		ConceptID string `json:"concept_id"`
	} `json:"noeuds"`
	Liens []struct {
		Source string `json:"source"`
		Target string `json:"target"`
	} `json:"liens"`
}

// parserReponseMindmap parse la réponse JSON du LLM en noeuds et liens
func (s *ServiceGeneration) parserReponseMindmap(reponseJSON []byte) ([]store.NoeudMindmap, []store.LienMindmap, error) {
	var reponse reponseMindmapJSON
	if err := json.Unmarshal(reponseJSON, &reponse); err != nil {
		return nil, nil, fmt.Errorf("erreur parsing JSON: %w", err)
	}

	// Convertir les nœuds
	noeuds := make([]store.NoeudMindmap, 0, len(reponse.Noeuds))
	for _, n := range reponse.Noeuds {
		// Valider le type de nœud
		typeNoeud := store.TypeNoeud(n.Type)
		if typeNoeud != store.TypeNoeudCentral &&
			typeNoeud != store.TypeNoeudBranche &&
			typeNoeud != store.TypeNoeudFeuille {
			typeNoeud = store.TypeNoeudFeuille // valeur par défaut
		}

		noeud := store.NoeudMindmap{
			ID:    n.ID,
			Label: n.Label,
			Type:  typeNoeud,
			Position: store.Position{
				X: n.Position.X,
				Y: n.Position.Y,
			},
			ConceptID: n.ConceptID,
		}
		noeuds = append(noeuds, noeud)
	}

	// Convertir les liens (ajouter un ID)
	liens := make([]store.LienMindmap, 0, len(reponse.Liens))
	for i, l := range reponse.Liens {
		lien := store.LienMindmap{
			ID:     fmt.Sprintf("lien%d", i+1),
			Source: l.Source,
			Target: l.Target,
		}
		liens = append(liens, lien)
	}

	return noeuds, liens, nil
}

// ObtenirMindmapParCours récupère la mindmap existante d'un cours
func (s *ServiceGeneration) ObtenirMindmapParCours(ctx context.Context, coursID string) (*store.Mindmap, error) {
	if s.mindmapRepo == nil {
		return nil, ErrServiceNonDisponible
	}

	return s.mindmapRepo.ObtenirParCours(ctx, coursID)
}

// --- Génération de Résumé ---

// SectionResume représente une section du résumé structuré
type SectionResume struct {
	Titre   string `json:"titre"`
	Contenu string `json:"contenu"`
}

// ResultatResume contient le résultat de la génération du résumé
type ResultatResume struct {
	PointsCles []string        `json:"pointsCles"`
	Structure  []SectionResume `json:"structure"`
	Paragraphe string          `json:"paragraphe"`
}

// Codes d'erreur résumé
var (
	ErrParsingResumeEchoue = &ErreurGeneration{Code: "PARSING_RESUME_ECHOUE", Message: "Erreur lors du parsing du résumé généré"}
	ErrResumeNonTrouve     = &ErreurGeneration{Code: "RESUME_NON_TROUVE", Message: "Aucun résumé trouvé pour ce cours"}
)

// GenererResume génère un résumé pour un cours
func (s *ServiceGeneration) GenererResume(ctx context.Context, coursID string) (*ResultatResume, error) {
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
	prompt := s.construirePromptResume(texte)

	// Appeler le LLM (gpt-4o-mini suffit pour résumer)
	llmOptions := llm.OptionsGeneration{
		Modele:        "gpt-4o-mini",
		Temperature:   0.7,
		MaxTokens:     4000,
		FormatReponse: "json",
		SystemPrompt:  "Tu es un professeur expert en synthèse de cours pour lycéens et collégiens. Tu réponds uniquement en JSON valide.",
	}

	reponseJSON, err := s.gestionnaireLLM.GenererJSON(ctx, prompt, nil, llmOptions)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrGenerationLLMEchouee, err.Error())
	}

	// Parser la réponse
	var resultat ResultatResume
	if err := json.Unmarshal(reponseJSON, &resultat); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrParsingResumeEchoue, err.Error())
	}

	// Sauvegarder en base
	resumeJSON, err := json.Marshal(resultat)
	if err != nil {
		return nil, fmt.Errorf("erreur sérialisation résumé: %w", err)
	}
	cours.Resume = json.RawMessage(resumeJSON)
	if err := s.coursRepo.MettreAJour(ctx, cours); err != nil {
		return nil, fmt.Errorf("erreur sauvegarde résumé: %w", err)
	}

	return &resultat, nil
}

// construirePromptResume construit le prompt pour la génération de résumé
func (s *ServiceGeneration) construirePromptResume(texte string) string {
	return fmt.Sprintf(`Cours :
"""%s"""

Génère un résumé structuré :
- 5-10 points clés
- Sections avec titre et contenu détaillé
- Paragraphe de synthèse (3-5 phrases)
- Langage clair, basé UNIQUEMENT sur le contenu fourni

JSON attendu :
{"pointsCles": ["..."], "structure": [{"titre": "...", "contenu": "..."}], "paragraphe": "..."}`, texte)
}

// ObtenirResume récupère le résumé existant d'un cours
func (s *ServiceGeneration) ObtenirResume(ctx context.Context, coursID string) (*ResultatResume, error) {
	// Récupérer le cours
	cours, err := s.coursRepo.ObtenirParID(ctx, coursID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrCoursNonTrouve, err.Error())
	}

	if cours.Resume == nil || len(cours.Resume) == 0 || string(cours.Resume) == "null" {
		return nil, ErrResumeNonTrouve
	}

	var resultat ResultatResume
	if err := json.Unmarshal(cours.Resume, &resultat); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrParsingResumeEchoue, err.Error())
	}

	return &resultat, nil
}
