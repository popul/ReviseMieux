const API_BASE = '/api'

// Types de base
export interface ErreurAPI {
  code: string
  message: string
}

export interface ReponseOCR {
  succes: boolean
  texte: string
  confiance: number
  zonesIncertaines: ZoneIncertaine[]
  nombrePages: number
  coursId?: string
}

export interface ZoneIncertaine {
  debut: number
  fin: number
  texte: string
  raison: string
}

export interface Cours {
  id: string
  titre: string
  matiere: string
  texteOcr: string
  confianceOcr: number
  zonesIncertaines: ZoneIncertaine[]
  dateCreation: string
  dateMiseAJour: string
}

export interface StatutAPI {
  version: string
  baseDeDonnees: string
  llm: string
}

// Helpers
async function gererReponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const erreur = await response.json().catch(() => ({
      code: 'ERREUR_INCONNUE',
      message: `Erreur HTTP ${response.status}`,
    }))
    throw new Error(erreur.message || `Erreur HTTP ${response.status}`)
  }
  return response.json()
}

// API Statut
export async function obtenirStatut(): Promise<StatutAPI> {
  const response = await fetch(`${API_BASE}/statut`)
  return gererReponse<StatutAPI>(response)
}

// API OCR
export async function envoyerOCR(
  fichiers: File[],
  options?: { titre?: string; matiere?: string; sauvegarder?: boolean }
): Promise<ReponseOCR> {
  const formData = new FormData()
  fichiers.forEach((fichier) => {
    formData.append('fichiers[]', fichier)
  })

  if (options?.titre) formData.append('titre', options.titre)
  if (options?.matiere) formData.append('matiere', options.matiere)
  if (options?.sauvegarder) formData.append('sauvegarder', 'true')

  const response = await fetch(`${API_BASE}/ocr`, {
    method: 'POST',
    body: formData,
  })

  return gererReponse<ReponseOCR>(response)
}

// API Cours
export async function listerCours(page = 1, limite = 10): Promise<{ cours: Cours[]; total: number }> {
  const response = await fetch(`${API_BASE}/cours?page=${page}&limite=${limite}`)
  return gererReponse(response)
}

export async function obtenirCours(id: string): Promise<Cours> {
  const response = await fetch(`${API_BASE}/cours/${id}`)
  return gererReponse<Cours>(response)
}

export async function supprimerCours(id: string): Promise<void> {
  const response = await fetch(`${API_BASE}/cours/${id}`, { method: 'DELETE' })
  if (!response.ok) {
    throw new Error(`Erreur lors de la suppression: ${response.status}`)
  }
}

// Types Fiches
export interface Fiche {
  id: string
  question: string
  reponse: string
  difficulte: 'facile' | 'moyen' | 'difficile'
  ordre: number
}

export interface ReponseFiches {
  succes: boolean
  fiches: Fiche[]
  nombreGenere: number
  erreur?: ErreurAPI
}

// API Fiches
export async function obtenirFichesCours(coursId: string): Promise<ReponseFiches> {
  const response = await fetch(`${API_BASE}/cours/${coursId}/fiches`)
  return gererReponse<ReponseFiches>(response)
}

export async function genererFiches(
  coursId: string,
  options?: { nombreFiches?: number; difficulte?: string }
): Promise<ReponseFiches> {
  const response = await fetch(`${API_BASE}/generer/fiches`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      coursId,
      nombreFiches: options?.nombreFiches,
      difficulte: options?.difficulte,
    }),
  })
  return gererReponse<ReponseFiches>(response)
}

// Types Quiz
export interface Question {
  id: string
  enonce: string
  choix: string[]
  reponseCorrecte: number
  explication: string
}

export interface Quiz {
  id: string
  coursId: string
  titre: string
  difficulte: 'facile' | 'moyen' | 'difficile'
  nombreQuestions: number
  questions: Question[]
}

export interface ReponseQuiz {
  succes: boolean
  quiz?: Quiz
  erreur?: ErreurAPI
}

export interface ReponseSession {
  questionId: string
  choixIndex: number
  estCorrecte: boolean
}

export interface SessionQuiz {
  id: string
  quizId: string
  reponses: ReponseSession[]
  score?: number
  termine: boolean
  dateDebut: string
  dateFin?: string
}

export interface ReponseSessionQuiz {
  succes: boolean
  session?: SessionQuiz
  erreur?: ErreurAPI
}

export interface ReponseRepondre {
  succes: boolean
  estCorrecte: boolean
  explication?: string
  erreur?: ErreurAPI
}

// API Quiz
export async function genererQuiz(
  coursId: string,
  options?: { nombreQuestions?: number; difficulte?: string; titre?: string }
): Promise<ReponseQuiz> {
  const response = await fetch(`${API_BASE}/generer/quiz`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      coursId,
      nombreQuestions: options?.nombreQuestions,
      difficulte: options?.difficulte,
      titre: options?.titre,
    }),
  })
  return gererReponse<ReponseQuiz>(response)
}

export async function obtenirQuiz(quizId: string): Promise<ReponseQuiz> {
  const response = await fetch(`${API_BASE}/quiz/${quizId}`)
  return gererReponse<ReponseQuiz>(response)
}

export async function demarrerSession(quizId: string): Promise<ReponseSessionQuiz> {
  const response = await fetch(`${API_BASE}/quiz/${quizId}/demarrer`, {
    method: 'POST',
  })
  return gererReponse<ReponseSessionQuiz>(response)
}

export async function repondreQuestion(
  quizId: string,
  sessionId: string,
  questionId: string,
  choixIndex: number
): Promise<ReponseRepondre> {
  const response = await fetch(`${API_BASE}/quiz/${quizId}/session/${sessionId}/repondre`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      questionId,
      choixIndex,
    }),
  })
  return gererReponse<ReponseRepondre>(response)
}

export async function terminerSession(
  quizId: string,
  sessionId: string
): Promise<ReponseSessionQuiz> {
  const response = await fetch(`${API_BASE}/quiz/${quizId}/session/${sessionId}/terminer`, {
    method: 'POST',
  })
  return gererReponse<ReponseSessionQuiz>(response)
}

// Types Statistiques
export interface Statistiques {
  nombreCours: number
  nombreFiches: number
  nombreQuiz: number
  quizCompletes: number
  scoreMoyen?: number
}

export interface ReponseStatistiques {
  succes: boolean
  statistiques?: Statistiques
  erreur?: ErreurAPI
}

export interface CoursResume {
  id: string
  titre: string
  matiere?: string
  nombreFiches: number
  nombreQuiz: number
  dateCreation: string
  dateModification: string
}

export interface ReponseCoursRecents {
  succes: boolean
  cours: CoursResume[]
  erreur?: ErreurAPI
}

// API Statistiques
export async function obtenirStatistiques(): Promise<ReponseStatistiques> {
  const response = await fetch(`${API_BASE}/statistiques`)
  return gererReponse<ReponseStatistiques>(response)
}

export async function obtenirCoursRecents(): Promise<ReponseCoursRecents> {
  const response = await fetch(`${API_BASE}/cours/recents`)
  return gererReponse<ReponseCoursRecents>(response)
}

// Types Mindmap
export interface Position {
  x: number
  y: number
}

export interface NoeudMindmap {
  id: string
  label: string
  type: 'central' | 'branche' | 'feuille'
  position: Position
}

export interface LienMindmap {
  id: string
  source: string
  target: string
}

export interface Mindmap {
  id: string
  coursId: string
  noeuds: NoeudMindmap[]
  liens: LienMindmap[]
  dateCreation: string
}

export interface ReponseMindmap {
  succes: boolean
  mindmap?: Mindmap
  erreur?: ErreurAPI
}

// API Mindmap
export async function genererMindmap(coursId: string): Promise<ReponseMindmap> {
  const response = await fetch(`${API_BASE}/generer/mindmap`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ coursId }),
  })
  return gererReponse<ReponseMindmap>(response)
}

export async function obtenirMindmap(coursId: string): Promise<ReponseMindmap> {
  const response = await fetch(`${API_BASE}/cours/${coursId}/mindmap`)
  return gererReponse<ReponseMindmap>(response)
}
