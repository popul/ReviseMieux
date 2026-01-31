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
