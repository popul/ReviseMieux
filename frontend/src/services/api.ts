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
  titreSuggere?: string
  matiereSuggeree?: string
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
  matiere?: string
  texteOCR: string
  confiance: number
  zonesIncertaines: ZoneIncertaine[]
  fichiersOriginaux?: string[]
  images: string[]
  dateCreation: string
  dateModification: string
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
  const data = await gererReponse<{ succes: boolean; cours: Cours }>(response)
  return data.cours
}

export async function supprimerCours(id: string): Promise<void> {
  const response = await fetch(`${API_BASE}/cours/${id}`, { method: 'DELETE' })
  if (!response.ok) {
    throw new Error(`Erreur lors de la suppression: ${response.status}`)
  }
}

export async function mettreAJourCours(
  id: string,
  donnees: { titre?: string; matiere?: string; texteOCR?: string; zonesIncertaines?: ZoneIncertaine[]; images?: string[] }
): Promise<Cours> {
  const response = await fetch(`${API_BASE}/cours/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(donnees),
  })
  const data = await gererReponse<{ succes: boolean; cours: Cours }>(response)
  return data.cours
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

// Types Quotas
export interface StatutQuota {
  pagesOcrUtilisees: number
  pagesOcrMax: number
  generationsUtilisees: number
  generationsMax: number
  pagesOcrRestantes: number
  generationsRestantes: number
}

export interface ReponseQuotas {
  succes: boolean
  quotas?: StatutQuota
  erreur?: ErreurAPI
}

// API Quotas
export async function obtenirQuotas(): Promise<ReponseQuotas> {
  const response = await fetch(`${API_BASE}/quotas`)
  return gererReponse<ReponseQuotas>(response)
}

// Types Ressources
export interface Ressource {
  id: string
  titre: string
  url?: string
  type: 'video' | 'article' | 'exercice' | 'cours' | 'autre'
  description?: string
}

export interface ReponseRessources {
  succes: boolean
  ressources: Ressource[]
  nombreGenere: number
  avertissement?: string
  erreur?: ErreurAPI
}

// API Ressources
export async function obtenirRessourcesCours(coursId: string): Promise<ReponseRessources> {
  const response = await fetch(`${API_BASE}/cours/${coursId}/ressources`)
  return gererReponse<ReponseRessources>(response)
}

export async function genererRessources(coursId: string): Promise<ReponseRessources> {
  const response = await fetch(`${API_BASE}/generer/ressources`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ coursId }),
  })
  return gererReponse<ReponseRessources>(response)
}

// Types Progression
export interface HistoriqueQuiz {
  sessionId: string
  quizId: string
  quizTitre: string
  coursId: string
  coursTitre: string
  matiere: string
  score: number
  dateFin: string
}

export interface StatsParMatiere {
  matiere: string
  nombreQuiz: number
  scoreMoyen: number
  meilleurScore: number
}

export interface ReponseProgression {
  succes: boolean
  historique: HistoriqueQuiz[]
  parMatiere: StatsParMatiere[]
  erreur?: ErreurAPI
}

// API Progression
export async function obtenirProgression(): Promise<ReponseProgression> {
  const response = await fetch(`${API_BASE}/progression`)
  return gererReponse<ReponseProgression>(response)
}

// Types Copies d'examens
export interface CopieExamen {
  id: string
  coursId?: string
  titre: string
  matiere?: string
  noteObtenue?: number
  noteTotale?: number
  texteOCR: string
  annotationsProfesseur?: string
  confiance: number
  zonesIncertaines: ZoneIncertaine[]
  fichiersOriginaux: string[]
  dateExamen?: string
  dateCreation: string
  dateModification: string
}

export interface ErreurAnalyse {
  id: string
  copieId: string
  typeErreur: 'comprehension' | 'methode' | 'inattention'
  texteOriginal?: string
  correction?: string
  explication: string
  conseil?: string
  severite: 'legere' | 'moderate' | 'grave'
  positionDebut?: number
  positionFin?: number
  dateCreation: string
}

export interface ResultatAnalyse {
  erreurs: ErreurAnalyse[]
  nombreErreurs: number
  resumeParType: Record<string, number>
  conseilGlobal: string
  pointsForts?: string[]
  pointsAAmeliorer?: string[]
}

export interface ReponseCopie {
  succes: boolean
  copie?: CopieExamen
  erreur?: ErreurAPI
}

export interface ReponseCopies {
  succes: boolean
  copies?: CopieExamen[]
  total?: number
  page?: number
  limite?: number
  erreur?: ErreurAPI
}

export interface ReponseAnalyse {
  succes: boolean
  resultat?: ResultatAnalyse
  erreur?: ErreurAPI
}

export interface ReponseErreurs {
  succes: boolean
  erreurs?: ErreurAnalyse[]
  nombreErreurs?: number
  comptesParType?: Record<string, number>
  erreur?: ErreurAPI
}

// API Copies d'examens
export async function envoyerOCRCopie(
  fichiers: File[],
  options?: {
    titre?: string
    matiere?: string
    coursId?: string
    noteObtenue?: number
    noteTotale?: number
    annotationsProfesseur?: string
  }
): Promise<ReponseCopie> {
  const formData = new FormData()
  fichiers.forEach((fichier) => {
    formData.append('fichiers[]', fichier)
  })

  if (options?.titre) formData.append('titre', options.titre)
  if (options?.matiere) formData.append('matiere', options.matiere)
  if (options?.coursId) formData.append('coursId', options.coursId)
  if (options?.noteObtenue !== undefined) formData.append('noteObtenue', options.noteObtenue.toString())
  if (options?.noteTotale !== undefined) formData.append('noteTotale', options.noteTotale.toString())
  if (options?.annotationsProfesseur) formData.append('annotationsProfesseur', options.annotationsProfesseur)

  const response = await fetch(`${API_BASE}/copies/ocr`, {
    method: 'POST',
    body: formData,
  })

  return gererReponse<ReponseCopie>(response)
}

export async function listerCopies(page = 1, limite = 20): Promise<ReponseCopies> {
  const response = await fetch(`${API_BASE}/copies?page=${page}&limite=${limite}`)
  return gererReponse<ReponseCopies>(response)
}

export async function obtenirCopie(id: string): Promise<ReponseCopie> {
  const response = await fetch(`${API_BASE}/copies/${id}`)
  return gererReponse<ReponseCopie>(response)
}

export async function supprimerCopie(id: string): Promise<void> {
  const response = await fetch(`${API_BASE}/copies/${id}`, { method: 'DELETE' })
  if (!response.ok) {
    throw new Error(`Erreur lors de la suppression: ${response.status}`)
  }
}

export async function analyserCopie(
  copieId: string,
  options?: { inclusAnnotations?: boolean; coursId?: string }
): Promise<ReponseAnalyse> {
  const response = await fetch(`${API_BASE}/copies/${copieId}/analyser`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      inclusAnnotations: options?.inclusAnnotations ?? true,
      coursId: options?.coursId,
    }),
  })
  return gererReponse<ReponseAnalyse>(response)
}

export async function obtenirErreursCopie(copieId: string): Promise<ReponseErreurs> {
  const response = await fetch(`${API_BASE}/copies/${copieId}/erreurs`)
  return gererReponse<ReponseErreurs>(response)
}

// Types Recommandations
export interface Recommandation {
  domaine: string
  raison: string
  severiteMax: 'grave' | 'moderate' | 'legere'
  actionSuggerie: string
  typeQuiz?: string
  priorite: number
}

export interface ResultatRecommandations {
  recommandations: Recommandation[]
  resume: string
  planAction: string
  prochainQuiz?: string
  motivation: string
  nombreRecommandations: number
}

export interface ReponseRecommandations {
  succes: boolean
  recommandations?: ResultatRecommandations
  erreur?: ErreurAPI
}

// API Recommandations
export async function genererRecommandationsCopie(
  copieId: string,
  options?: { inclureQuiz?: boolean }
): Promise<ReponseRecommandations> {
  const params = new URLSearchParams()
  if (options?.inclureQuiz !== undefined) {
    params.append('inclure_quiz', String(options.inclureQuiz))
  }

  const url = `${API_BASE}/copies/${copieId}/recommandations${params.toString() ? '?' + params.toString() : ''}`
  const response = await fetch(url, {
    method: 'POST',
  })
  return gererReponse<ReponseRecommandations>(response)
}
