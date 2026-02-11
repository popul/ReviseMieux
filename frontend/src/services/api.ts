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
  blocsTexte?: BlocTexteParPage[]
}

export interface ZoneIncertaine {
  debut: number
  fin: number
  texte: string
  raison: string
}

export interface PositionBlocOCR {
  x: number
  y: number
  largeur: number
  hauteur: number
}

export interface BlocTexteOCR {
  texte: string
  position: PositionBlocOCR
  confiance: number
}

export interface BlocTexteParPage {
  page: number
  blocs_texte: BlocTexteOCR[]
}

// Types Résumé
export interface SectionResume {
  titre: string
  contenu: string
}

export interface ResumeCours {
  pointsCles: string[]
  structure: SectionResume[]
  paragraphe: string
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
  blocsTexte?: BlocTexteParPage[]
  resume?: ResumeCours
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

// Types Config
export interface ConfigFrontend {
  nombreMaxPages: number
}

// API Config
export async function obtenirConfig(): Promise<ConfigFrontend> {
  try {
    const response = await fetch(`${API_BASE}/config`)
    const data = await gererReponse<{ succes: boolean; config: ConfigFrontend }>(response)
    return data.config
  } catch {
    return { nombreMaxPages: 30 }
  }
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

// Dimension max pour les images envoyées au LLM (OpenAI resize en interne à 2048px)
const DIMENSION_MAX_IMAGE = 2048
const QUALITE_JPEG = 0.85

// Redimensionne une image côté client via Canvas si elle dépasse DIMENSION_MAX_IMAGE
async function redimensionnerImage(fichier: File): Promise<File> {
  // Ne pas toucher aux PDF
  if (fichier.type === 'application/pdf') return fichier

  // Charger l'image pour obtenir ses dimensions
  const url = URL.createObjectURL(fichier)
  const img = new Image()
  await new Promise<void>((resolve, reject) => {
    img.onload = () => resolve()
    img.onerror = () => reject(new Error('Impossible de charger l\'image'))
    img.src = url
  })
  URL.revokeObjectURL(url)

  // Pas besoin de redimensionner si déjà assez petit
  if (img.width <= DIMENSION_MAX_IMAGE && img.height <= DIMENSION_MAX_IMAGE) {
    return fichier
  }

  // Calculer les nouvelles dimensions en gardant le ratio
  let newW = img.width
  let newH = img.height
  if (newW > newH) {
    newH = Math.round(newH * DIMENSION_MAX_IMAGE / newW)
    newW = DIMENSION_MAX_IMAGE
  } else {
    newW = Math.round(newW * DIMENSION_MAX_IMAGE / newH)
    newH = DIMENSION_MAX_IMAGE
  }

  // Dessiner sur un canvas redimensionné
  const canvas = document.createElement('canvas')
  canvas.width = newW
  canvas.height = newH
  const ctx = canvas.getContext('2d')!
  ctx.drawImage(img, 0, 0, newW, newH)

  // Exporter en JPEG (meilleur ratio taille/qualité)
  const blob = await new Promise<Blob>((resolve) => {
    canvas.toBlob((b) => resolve(b!), 'image/jpeg', QUALITE_JPEG)
  })

  // Garder le même nom mais avec extension .jpg
  const nom = fichier.name.replace(/\.[^.]+$/, '.jpg')
  return new File([blob], nom, { type: 'image/jpeg' })
}

// Redimensionne toutes les images en parallèle
async function redimensionnerImages(fichiers: File[]): Promise<File[]> {
  return Promise.all(fichiers.map(redimensionnerImage))
}

// API OCR avec streaming SSE (progression page par page)
export async function envoyerOCRStream(
  fichiers: File[],
  onProgression: (page: number, total: number) => void,
  options?: { titre?: string; matiere?: string; sauvegarder?: boolean },
  signal?: AbortSignal
): Promise<ReponseOCR> {
  // Redimensionner les images côté client avant upload
  const fichiersRedim = await redimensionnerImages(fichiers)

  const formData = new FormData()
  fichiersRedim.forEach((fichier) => {
    formData.append('fichiers[]', fichier)
  })

  if (options?.titre) formData.append('titre', options.titre)
  if (options?.matiere) formData.append('matiere', options.matiere)
  if (options?.sauvegarder) formData.append('sauvegarder', 'true')

  const response = await fetch(`${API_BASE}/ocr`, {
    method: 'POST',
    body: formData,
    signal,
  })

  // Erreurs de validation (avant le SSE) : JSON classique
  if (!response.ok) {
    const erreur = await response.json().catch(() => ({
      erreur: { code: 'ERREUR_INCONNUE', message: `Erreur HTTP ${response.status}` },
    }))
    const err = new Error(erreur.erreur?.message || `Erreur HTTP ${response.status}`) as Error & { status?: number; code?: string }
    err.status = response.status
    err.code = erreur.erreur?.code
    throw err
  }

  // Lire le flux SSE
  const reader = response.body!.getReader()
  const decoder = new TextDecoder()
  let buffer = ''

  while (true) {
    const { done, value } = await reader.read()
    if (done) break

    buffer += decoder.decode(value, { stream: true })

    // Séparer les événements SSE (délimiteur : double saut de ligne)
    const parties = buffer.split('\n\n')
    buffer = parties.pop()! // garder le fragment incomplet

    for (const partie of parties) {
      if (!partie.trim()) continue

      let eventType = ''
      let eventData = ''

      for (const ligne of partie.split('\n')) {
        if (ligne.startsWith('event: ')) eventType = ligne.slice(7)
        else if (ligne.startsWith('data: ')) eventData = ligne.slice(6)
      }

      if (!eventData) continue

      try {
        const parsed = JSON.parse(eventData)
        switch (eventType) {
          case 'progress':
            onProgression(parsed.page, parsed.total)
            break
          case 'complete':
            return parsed as ReponseOCR
          case 'error':
            throw new Error(parsed.message || 'Erreur OCR')
        }
      } catch (e) {
        if (e instanceof Error && e.message !== 'Erreur OCR' && !e.message.startsWith('Erreur')) continue
        throw e
      }
    }
  }

  throw new Error('La connexion avec le serveur a été interrompue')
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
  donnees: { titre?: string; matiere?: string; texteOCR?: string; zonesIncertaines?: ZoneIncertaine[]; images?: string[]; blocsTexte?: BlocTexteParPage[] }
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
  conceptIds?: string[]
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
  conceptId?: string
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


// Types Concepts
export interface PositionDansCours {
  debut: number
  fin: number
}

export interface Concept {
  id: string
  coursId: string
  nom: string
  definition: string
  importance: 'essentiel' | 'important' | 'secondaire'
  positionDansCours?: PositionDansCours
  createdAt: string
  updatedAt: string
}

export interface ReponseConcepts {
  succes: boolean
  concepts: Concept[]
  nombreExtraits: number
  erreur?: ErreurAPI
}

export interface ReponseConceptUnique {
  succes: boolean
  concept?: Concept
  erreur?: ErreurAPI
}

// API Concepts
export async function extraireConcepts(coursId: string): Promise<ReponseConcepts> {
  const response = await fetch(`${API_BASE}/cours/${coursId}/concepts/extraire`, {
    method: 'POST',
  })
  return gererReponse<ReponseConcepts>(response)
}

export async function getConceptsByCours(coursId: string): Promise<ReponseConcepts> {
  const response = await fetch(`${API_BASE}/cours/${coursId}/concepts`)
  return gererReponse<ReponseConcepts>(response)
}

export async function updateConcept(
  id: string,
  data: Partial<Pick<Concept, 'nom' | 'definition' | 'importance' | 'positionDansCours'>>
): Promise<ReponseConceptUnique> {
  const response = await fetch(`${API_BASE}/concepts/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  return gererReponse<ReponseConceptUnique>(response)
}

export async function deleteConcept(id: string): Promise<void> {
  const response = await fetch(`${API_BASE}/concepts/${id}`, { method: 'DELETE' })
  if (!response.ok) {
    throw new Error(`Erreur lors de la suppression: ${response.status}`)
  }
}
// Types Examen Blanc
export interface IndiceExamen {
  niveau: number
  texte: string
}

export interface QuestionExamen {
  numero: number
  type: 'definition' | 'comprehension' | 'application' | 'synthese'
  difficulte: 'facile' | 'moyen' | 'difficile'
  enonce: string
  bareme: number
  reponseAttendue: string
  indices: IndiceExamen[]
}

export interface ExamenBlanc {
  id: string
  coursId: string
  questions: QuestionExamen[]
  dureeMinutes: number
  dateCreation: string
}

export interface ReponseExamenBlanc {
  succes: boolean
  examen?: ExamenBlanc
  erreur?: ErreurAPI
}

export interface ReponseExamenDetail {
  questionNumero: number
  texte: string
}

export interface IndiceUtilise {
  questionNumero: number
  niveauIndice: number
}

export interface PointFort {
  concept: string
  commentaire: string
}

export interface PointFaible {
  concept: string
  commentaire: string
}

export interface EtapePlanRevision {
  priorite: number
  action: string
  concept: string
  ressource?: string
}

export interface SessionExamenBlanc {
  id: string
  examenId: string
  reponses: ReponseExamenDetail[]
  indicesUtilises: IndiceUtilise[]
  noteEstimee?: number
  pointsForts?: PointFort[]
  pointsFaibles?: PointFaible[]
  planRevision?: EtapePlanRevision[]
  termine: boolean
  dateDebut: string
  dateFin?: string
}

export interface ReponseSessionExamen {
  succes: boolean
  session?: SessionExamenBlanc
  erreur?: ErreurAPI
}

export interface ReponseIndice {
  succes: boolean
  indice?: IndiceExamen
  erreur?: ErreurAPI
}

export interface DetailCorrection {
  questionNumero: number
  reponseEleve: string
  reponseAttendue: string
  noteQuestion: number
  bareme: number
  commentaire: string
}

export interface ResultatCorrection {
  noteEstimee: number
  pointsForts: PointFort[]
  pointsFaibles: PointFaible[]
  planRevision: EtapePlanRevision[]
  details: DetailCorrection[]
}

export interface ReponseCorrectionExamen {
  succes: boolean
  resultat?: ResultatCorrection
  erreur?: ErreurAPI
}

// API Examen Blanc
export async function genererExamen(coursId: string): Promise<ReponseExamenBlanc> {
  const response = await fetch(API_BASE + '/cours/' + coursId + '/examen/generer', {
    method: 'POST',
  })
  return gererReponse<ReponseExamenBlanc>(response)
}

export async function obtenirExamen(examenId: string): Promise<ReponseExamenBlanc> {
  const response = await fetch(API_BASE + '/examens/' + examenId)
  return gererReponse<ReponseExamenBlanc>(response)
}

export async function demarrerSessionExamen(examenId: string): Promise<ReponseSessionExamen> {
  const response = await fetch(API_BASE + '/examens/' + examenId + '/sessions', {
    method: 'POST',
  })
  return gererReponse<ReponseSessionExamen>(response)
}

export async function obtenirSessionExamen(sessionId: string): Promise<ReponseSessionExamen> {
  const response = await fetch(API_BASE + '/sessions-examen/' + sessionId)
  return gererReponse<ReponseSessionExamen>(response)
}

export async function demanderIndice(
  sessionId: string,
  questionNumero: number,
  niveauIndice: number
): Promise<ReponseIndice> {
  const response = await fetch(API_BASE + '/sessions-examen/' + sessionId + '/indice', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ questionNumero, niveauIndice }),
  })
  return gererReponse<ReponseIndice>(response)
}

export async function corrigerExamen(
  sessionId: string,
  reponses: ReponseExamenDetail[]
): Promise<ReponseCorrectionExamen> {
  const response = await fetch(API_BASE + '/sessions-examen/' + sessionId + '/corriger', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ reponses }),
  })
  return gererReponse<ReponseCorrectionExamen>(response)
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


// Types Lexique
export interface TermeLexique {
  id: string
  coursId: string
  terme: string
  definition: string
  contexte?: string
  exemple?: string
  categorie?: string
  maitrise: number
  createdAt: string
  updatedAt: string
}

export interface QuestionVocabulaire {
  id: string
  question: string
  reponseAttendue: string
  indices: string[]
  terme: string
}

export interface QuizVocabulaire {
  mode: string
  questions: QuestionVocabulaire[]
}

export interface ReponseLexique {
  succes: boolean
  termes: TermeLexique[]
  nombreExtraits: number
  erreur?: ErreurAPI
}

export interface ReponseQuizVocabulaire {
  succes: boolean
  quiz?: QuizVocabulaire
  erreur?: ErreurAPI
}

// API Lexique
export async function extraireTermesLexique(coursId: string): Promise<ReponseLexique> {
  const response = await fetch(`${API_BASE}/cours/${coursId}/lexique/extraire`, {
    method: 'POST',
  })
  return gererReponse<ReponseLexique>(response)
}

export async function getTermesLexique(coursId: string): Promise<ReponseLexique> {
  const response = await fetch(`${API_BASE}/cours/${coursId}/lexique`)
  return gererReponse<ReponseLexique>(response)
}

export async function mettreAJourMaitrise(termeId: string, maitrise: number): Promise<{ succes: boolean; maitrise: number }> {
  const response = await fetch(`${API_BASE}/lexique/${termeId}/maitrise`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ maitrise }),
  })
  return gererReponse(response)
}

export async function genererQuizVocabulaire(
  coursId: string,
  mode: 'terme_vers_definition' | 'definition_vers_terme'
): Promise<ReponseQuizVocabulaire> {
  const response = await fetch(`${API_BASE}/cours/${coursId}/lexique/quiz`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ mode }),
  })
  return gererReponse<ReponseQuizVocabulaire>(response)
}

// API Résumé
export async function genererResume(coursId: string): Promise<{ succes: boolean; resume: ResumeCours }> {
  const response = await fetch(`${API_BASE}/cours/${coursId}/resume/generer`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
  })
  return gererReponse(response)
}

// API Re-OCR
export async function retraiterOCRCours(coursId: string): Promise<ReponseOCR> {
  const response = await fetch(`${API_BASE}/cours/${coursId}/reocr`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
  })
  return gererReponse(response)
}

// ============================================================
// Plans de révision
// ============================================================

export interface PlanRevisionResume {
  id: string
  titre: string
  iconeMatiere: string
  nombreCours: number
  progression: number
  dateEcheance?: string
}

export interface PlanRevision {
  id: string
  titre: string
  description?: string
  matiere?: string
  iconeMatiere: string
  dateEcheance?: string
  dateCreation: string
  dateModification: string
}

export interface CoursAvecArtifacts {
  cours: Cours
  nombreFiches: number
  nombreQuiz: number
  aMindmap: boolean
  aResume: boolean
}

export interface ReponsePlanComplet {
  succes: boolean
  plan?: PlanRevision
  cours?: CoursAvecArtifacts[]
  erreur?: ErreurAPI
}

export async function listerPlansRevision(): Promise<PlanRevisionResume[]> {
  try {
    const response = await fetch(`${API_BASE}/plans?resume=true`)
    if (response.status === 404) return []
    const data = await gererReponse<{ plans: PlanRevisionResume[] }>(response)
    return data.plans || []
  } catch {
    return []
  }
}

export async function creerPlanRevision(donnees: {
  titre: string
  description?: string
  matiere?: string
  iconeMatiere?: string
  dateEcheance?: string
  coursIds?: string[]
}): Promise<{ succes: boolean; plan?: PlanRevision; erreur?: ErreurAPI }> {
  const response = await fetch(`${API_BASE}/plans`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(donnees),
  })
  return gererReponse(response)
}

export async function obtenirPlanComplet(id: string): Promise<ReponsePlanComplet> {
  const response = await fetch(`${API_BASE}/plans/${id}/complet`)
  return gererReponse<ReponsePlanComplet>(response)
}

export async function mettreAJourPlanRevision(
  id: string,
  donnees: {
    titre?: string
    description?: string
    matiere?: string
    iconeMatiere?: string
    dateEcheance?: string
  }
): Promise<{ succes: boolean; plan?: PlanRevision; erreur?: ErreurAPI }> {
  const response = await fetch(`${API_BASE}/plans/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(donnees),
  })
  return gererReponse(response)
}

export async function supprimerPlanRevision(id: string): Promise<void> {
  const response = await fetch(`${API_BASE}/plans/${id}`, { method: 'DELETE' })
  if (!response.ok) {
    throw new Error(`Erreur lors de la suppression: ${response.status}`)
  }
}

export async function ajouterCoursAuPlan(planId: string, coursId: string): Promise<{ succes: boolean }> {
  const response = await fetch(`${API_BASE}/plans/${planId}/cours`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ coursId }),
  })
  return gererReponse(response)
}

export async function retirerCoursDuPlan(planId: string, coursId: string): Promise<{ succes: boolean }> {
  const response = await fetch(`${API_BASE}/plans/${planId}/cours/${coursId}`, {
    method: 'DELETE',
  })
  return gererReponse(response)
}

export async function listerPlansParCours(coursId: string): Promise<PlanRevisionResume[]> {
  try {
    const response = await fetch(`${API_BASE}/cours/${coursId}/plans`)
    const data = await gererReponse<{ plans: PlanRevisionResume[] }>(response)
    return data.plans || []
  } catch {
    return []
  }
}
