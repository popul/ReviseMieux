import { useState, useCallback, useLayoutEffect, useEffect, useRef } from 'react'
import { useSearchParams, Link } from 'react-router-dom'
import {
  listerCours,
  genererExamen,
  demarrerSessionExamen,
  demanderIndice,
  corrigerExamen,
  type Cours,
  type ExamenBlanc as ExamenBlancType,
  type SessionExamenBlanc,
  type ReponseExamenDetail,
  type IndiceExamen,
  type ResultatCorrection,
  type DetailCorrection,
  type PointFort,
  type PointFaible,
  type EtapePlanRevision,
} from '../services/api'
import ProcessingSection from '../components/ProcessingSection'

type EtatExamen =
  | 'selection-cours'
  | 'generation'
  | 'examen'
  | 'correction'
  | 'resultats'
  | 'erreur'

// Hook pour charger les cours
function useChargementCours(coursId: string | null) {
  const [listeCours, setListeCours] = useState<Cours[]>([])
  const [chargement, setChargement] = useState(() => !coursId)
  const [erreur, setErreur] = useState<string | null>(null)

  useLayoutEffect(() => {
    if (coursId) return

    let cancelled = false

    listerCours(1, 50)
      .then((res) => {
        if (!cancelled) {
          setListeCours(res.cours || [])
          setChargement(false)
        }
      })
      .catch(() => {
        if (!cancelled) {
          setErreur('Impossible de charger la liste des cours')
          setChargement(false)
        }
      })

    return () => {
      cancelled = true
    }
  }, [coursId])

  return { listeCours, chargement, erreur }
}

// Jauge circulaire pour la note /20
function JaugeNote({ note, total = 20 }: { note: number; total?: number }) {
  const pourcentage = Math.min((note / total) * 100, 100)
  const rayon = 70
  const circonference = 2 * Math.PI * rayon

  let couleur: string
  let message: string
  if (pourcentage >= 75) {
    couleur = '#1A4D4D' // teal
    message = 'Excellent !'
  } else if (pourcentage >= 50) {
    couleur = '#F5C542' // gold
    message = 'Pas mal !'
  } else {
    couleur = '#E85D4C' // coral
    message = 'A retravailler'
  }

  return (
    <div className="text-center">
      <div className="relative w-44 h-44 mx-auto mb-4">
        <svg className="w-full h-full transform -rotate-90" viewBox="0 0 160 160">
          <circle cx="80" cy="80" r={rayon} fill="none" stroke="#F5F3EE" strokeWidth="12" />
          <circle
            cx="80"
            cy="80"
            r={rayon}
            fill="none"
            stroke={couleur}
            strokeWidth="12"
            strokeLinecap="round"
            strokeDasharray={`${(pourcentage / 100) * circonference} ${circonference}`}
            className="transition-all duration-1000"
          />
        </svg>
        <div className="absolute inset-0 flex flex-col items-center justify-center">
          <span className="font-display text-4xl font-bold text-ink">
            {note.toFixed(1)}
          </span>
          <span className="text-sm text-ink-muted">/ {total}</span>
        </div>
      </div>
      <p className="font-display text-lg font-semibold" style={{ color: couleur }}>
        {message}
      </p>
    </div>
  )
}

// Sidebar des questions
function SidebarQuestions({
  nombreQuestions,
  questionActive,
  reponses,
  onNaviguer,
}: {
  nombreQuestions: number
  questionActive: number
  reponses: Map<number, string>
  onNaviguer: (numero: number) => void
}) {
  return (
    <div className="bg-white rounded-lg p-4 shadow-sm">
      <h3 className="text-sm font-semibold text-ink-muted mb-4">Questions</h3>
      <div className="grid grid-cols-5 gap-2">
        {Array.from({ length: nombreQuestions }, (_, i) => {
          const numero = i + 1
          const aRepondu = reponses.has(numero)
          const estActive = numero === questionActive

          return (
            <button
              key={numero}
              onClick={() => onNaviguer(numero)}
              className={`w-10 h-10 rounded-lg text-sm font-medium transition-all ${
                estActive
                  ? 'bg-coral text-white shadow-md'
                  : aRepondu
                  ? 'bg-teal/20 text-teal border border-teal/30'
                  : 'bg-cream text-ink-muted hover:bg-cream-dark'
              }`}
              aria-label={`Question ${numero}${aRepondu ? ' (repondue)' : ''}`}
            >
              {numero}
            </button>
          )
        })}
      </div>
      <div className="mt-4 pt-4 border-t border-cream-dark">
        <p className="text-xs text-ink-muted">
          {reponses.size} / {nombreQuestions} repondues
        </p>
      </div>
    </div>
  )
}

// Zone de question avec indices
function ZoneQuestion({
  numero,
  enonce,
  type,
  difficulte,
  bareme,
  reponse,
  indices,
  onReponseChange,
  onDemanderIndice,
  chargementIndice,
}: {
  numero: number
  enonce: string
  type: string
  difficulte: string
  bareme: number
  reponse: string
  indices: IndiceExamen[]
  onReponseChange: (texte: string) => void
  onDemanderIndice: () => void
  chargementIndice: boolean
}) {
  const typesLibelles: Record<string, string> = {
    definition: 'Definition',
    comprehension: 'Comprehension',
    application: 'Application',
    synthese: 'Synthese',
  }

  const difficulteStyles: Record<string, string> = {
    facile: 'bg-green-100 text-green-700',
    moyen: 'bg-gold/20 text-gold',
    difficile: 'bg-coral/20 text-coral',
  }

  return (
    <div className="bg-white rounded-lg p-4 md:p-8 shadow-sm">
      {/* En-tete question */}
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-3">
          <span className="bg-coral text-white text-sm font-bold px-3 py-1.5 rounded-full">
            Q{numero}
          </span>
          <span className={`text-xs px-2 py-1 rounded-full ${difficulteStyles[difficulte] || ''}`}>
            {difficulte}
          </span>
          <span className="text-xs px-2 py-1 rounded-full bg-cream text-ink-muted">
            {typesLibelles[type] || type}
          </span>
        </div>
        <span className="text-sm text-ink-muted font-medium">
          {bareme} point{bareme > 1 ? 's' : ''}
        </span>
      </div>

      {/* Enonce */}
      <div className="mb-6 p-6 bg-cream rounded-lg">
        <p className="text-ink leading-relaxed">{enonce}</p>
      </div>

      {/* Zone de reponse */}
      <div className="mb-6">
        <label className="block text-sm font-medium text-ink mb-2">
          Votre reponse
        </label>
        <textarea
          value={reponse}
          onChange={(e) => onReponseChange(e.target.value)}
          placeholder="Ecrivez votre reponse ici..."
          className="w-full h-40 p-4 bg-cream rounded-lg text-sm text-ink font-body border border-cream-dark focus:border-coral focus:outline-none resize-none"
        />
      </div>

      {/* Indices */}
      <div>
        <div className="flex items-center justify-between mb-3">
          <h4 className="text-sm font-medium text-ink-muted">
            Indices ({indices.length}/3)
          </h4>
          {indices.length < 3 && (
            <button
              onClick={onDemanderIndice}
              disabled={chargementIndice}
              className="text-sm text-teal hover:text-teal-light transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {chargementIndice ? 'Chargement...' : 'Demander un indice'}
            </button>
          )}
        </div>
        {indices.length > 0 && (
          <div className="space-y-2">
            {indices.map((indice, i) => (
              <div
                key={i}
                className="p-3 bg-gold/10 border border-gold/30 rounded-md text-sm text-ink"
              >
                <span className="font-medium text-gold mr-2">Niveau {indice.niveau} :</span>
                {indice.texte}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

// Vue resultats
function VueResultats({
  resultat,
  coursId,
  onRecommencer,
}: {
  resultat: ResultatCorrection
  coursId: string
  onRecommencer: () => void
}) {
  const [ongletActif, setOngletActif] = useState<'resume' | 'details' | 'plan'>('resume')

  return (
    <div>
      {/* Jauge de note */}
      <div className="bg-white rounded-lg p-6 md:p-12 shadow-sm mb-8">
        <h2 className="font-display text-2xl font-bold text-ink text-center mb-8">
          Resultats de l'examen
        </h2>
        <JaugeNote note={resultat.noteEstimee} />
      </div>

      {/* Onglets */}
      <div className="bg-white rounded-lg shadow-sm mb-8">
        <div className="flex border-b border-cream-dark">
          <button
            onClick={() => setOngletActif('resume')}
            className={`flex-1 py-4 text-sm font-medium transition-colors ${
              ongletActif === 'resume'
                ? 'text-coral border-b-2 border-coral'
                : 'text-ink-muted hover:text-ink'
            }`}
          >
            Resume
          </button>
          <button
            onClick={() => setOngletActif('details')}
            className={`flex-1 py-4 text-sm font-medium transition-colors ${
              ongletActif === 'details'
                ? 'text-coral border-b-2 border-coral'
                : 'text-ink-muted hover:text-ink'
            }`}
          >
            Detail par question
          </button>
          <button
            onClick={() => setOngletActif('plan')}
            className={`flex-1 py-4 text-sm font-medium transition-colors ${
              ongletActif === 'plan'
                ? 'text-coral border-b-2 border-coral'
                : 'text-ink-muted hover:text-ink'
            }`}
          >
            Plan de revision
          </button>
        </div>

        <div className="p-8">
          {/* Onglet Resume */}
          {ongletActif === 'resume' && (
            <div className="space-y-8">
              {/* Points forts */}
              {resultat.pointsForts.length > 0 && (
                <div>
                  <h3 className="font-semibold text-ink mb-4 flex items-center gap-2">
                    <span className="text-green-600">+</span> Points forts
                  </h3>
                  <div className="space-y-3">
                    {resultat.pointsForts.map((pf: PointFort, i: number) => (
                      <div key={i} className="p-4 bg-green-50 border border-green-200 rounded-lg">
                        <p className="font-medium text-green-800 text-sm">{pf.concept}</p>
                        <p className="text-sm text-green-700 mt-1">{pf.commentaire}</p>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Points faibles */}
              {resultat.pointsFaibles.length > 0 && (
                <div>
                  <h3 className="font-semibold text-ink mb-4 flex items-center gap-2">
                    <span className="text-coral">-</span> Points a ameliorer
                  </h3>
                  <div className="space-y-3">
                    {resultat.pointsFaibles.map((pf: PointFaible, i: number) => (
                      <div key={i} className="p-4 bg-coral/5 border border-coral/20 rounded-lg">
                        <p className="font-medium text-coral text-sm">{pf.concept}</p>
                        <p className="text-sm text-ink-light mt-1">{pf.commentaire}</p>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}

          {/* Onglet Details */}
          {ongletActif === 'details' && (
            <div className="space-y-6">
              {resultat.details.map((detail: DetailCorrection) => (
                <div
                  key={detail.questionNumero}
                  className="p-6 border border-cream-dark rounded-lg"
                >
                  <div className="flex items-center justify-between mb-4">
                    <span className="bg-cream text-ink font-bold text-sm px-3 py-1.5 rounded-full">
                      Q{detail.questionNumero}
                    </span>
                    <span
                      className={`text-sm font-semibold ${
                        detail.noteQuestion >= detail.bareme * 0.75
                          ? 'text-green-600'
                          : detail.noteQuestion >= detail.bareme * 0.5
                          ? 'text-gold'
                          : 'text-coral'
                      }`}
                    >
                      {detail.noteQuestion} / {detail.bareme}
                    </span>
                  </div>

                  {/* Reponse de l'eleve */}
                  <div className="mb-4">
                    <p className="text-xs font-medium text-ink-muted mb-1">Votre reponse</p>
                    <div className="p-3 bg-cream rounded-md text-sm text-ink">
                      {detail.reponseEleve || <em className="text-ink-muted">Pas de reponse</em>}
                    </div>
                  </div>

                  {/* Reponse attendue */}
                  <div className="mb-4">
                    <p className="text-xs font-medium text-ink-muted mb-1">Reponse attendue</p>
                    <div className="p-3 bg-teal/5 border border-teal/20 rounded-md text-sm text-ink">
                      {detail.reponseAttendue}
                    </div>
                  </div>

                  {/* Commentaire */}
                  {detail.commentaire && (
                    <div className="p-3 bg-gold/10 border border-gold/30 rounded-md text-sm text-ink">
                      <span className="font-medium">Commentaire : </span>
                      {detail.commentaire}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}

          {/* Onglet Plan de revision */}
          {ongletActif === 'plan' && (
            <div>
              {resultat.planRevision.length > 0 ? (
                <div className="space-y-4">
                  {resultat.planRevision.map((etape: EtapePlanRevision, i: number) => (
                    <div
                      key={i}
                      className="flex items-start gap-4 p-4 bg-cream rounded-lg"
                    >
                      <span className="bg-coral text-white text-xs font-bold w-7 h-7 rounded-full flex items-center justify-center flex-shrink-0">
                        {etape.priorite}
                      </span>
                      <div className="flex-1">
                        <p className="font-medium text-ink text-sm">{etape.action}</p>
                        <p className="text-xs text-ink-muted mt-1">
                          Concept : {etape.concept}
                        </p>
                        {etape.ressource && (
                          <p className="text-xs text-teal mt-1">
                            Ressource : {etape.ressource}
                          </p>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-center text-ink-muted py-8">
                  Aucun plan de revision disponible.
                </p>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Actions */}
      <div className="flex flex-col sm:flex-row gap-3 justify-center">
        <button
          onClick={onRecommencer}
          className="px-6 py-2.5 md:px-8 md:py-3 bg-coral text-white rounded-full font-semibold hover:bg-coral-dark transition-colors"
        >
          Nouvel examen
        </button>
        <Link
          to={'/fiches?cours=' + coursId}
          className="px-6 py-2.5 md:px-8 md:py-3 border-2 border-coral text-coral rounded-full font-semibold hover:bg-coral/5 transition-colors text-center"
        >
          Revoir les fiches
        </Link>
        <Link
          to={'/quiz?cours=' + coursId}
          className="px-6 py-2.5 md:px-8 md:py-3 border-2 border-teal text-teal rounded-full font-semibold hover:bg-teal/5 transition-colors text-center"
        >
          Lancer un quiz
        </Link>
      </div>
    </div>
  )
}

// Composant timer
function Timer({ dureeMinutes, onTempsEcoule }: { dureeMinutes: number; onTempsEcoule: () => void }) {
  const [secondesRestantes, setSecondesRestantes] = useState(dureeMinutes * 60)
  const onTempsEcouleRef = useRef(onTempsEcoule)
  onTempsEcouleRef.current = onTempsEcoule

  useEffect(() => {
    const interval = setInterval(() => {
      setSecondesRestantes((prev) => {
        if (prev <= 1) {
          clearInterval(interval)
          onTempsEcouleRef.current()
          return 0
        }
        return prev - 1
      })
    }, 1000)

    return () => clearInterval(interval)
  }, [])

  const minutes = Math.floor(secondesRestantes / 60)
  const secondes = secondesRestantes % 60
  const pourcentage = (secondesRestantes / (dureeMinutes * 60)) * 100

  let couleurTimer = 'text-ink'
  if (pourcentage < 10) couleurTimer = 'text-coral animate-pulse'
  else if (pourcentage < 25) couleurTimer = 'text-coral'
  else if (pourcentage < 50) couleurTimer = 'text-gold'

  return (
    <div className="flex items-center gap-3">
      <div className={`font-mono text-lg font-bold ${couleurTimer}`}>
        {String(minutes).padStart(2, '0')}:{String(secondes).padStart(2, '0')}
      </div>
      <div className="w-24 h-2 bg-cream rounded-full overflow-hidden">
        <div
          className={`h-full rounded-full transition-all duration-1000 ${
            pourcentage < 25 ? 'bg-coral' : pourcentage < 50 ? 'bg-gold' : 'bg-teal'
          }`}
          style={{ width: pourcentage + '%' }}
        />
      </div>
    </div>
  )
}

export default function ExamenBlanc() {
  const [searchParams, setSearchParams] = useSearchParams()
  const coursId = searchParams.get('cours')

  const { listeCours, chargement: chargementCours, erreur: erreurCours } = useChargementCours(coursId)

  // Etat general
  const [etat, setEtat] = useState<EtatExamen>(coursId ? 'generation' : 'selection-cours')
  const [examen, setExamen] = useState<ExamenBlancType | null>(null)
  const [session, setSession] = useState<SessionExamenBlanc | null>(null)
  const [erreur, setErreur] = useState<string | null>(null)

  // Etat de l'examen en cours
  const [questionActive, setQuestionActive] = useState(1)
  const [reponses, setReponses] = useState<Map<number, string>>(new Map())
  const [indicesParQuestion, setIndicesParQuestion] = useState<Map<number, IndiceExamen[]>>(new Map())
  const [chargementIndice, setChargementIndice] = useState(false)
  const [resultat, setResultat] = useState<ResultatCorrection | null>(null)

  // Reinitialiser quand coursId change
  const [prevCoursId, setPrevCoursId] = useState(coursId)
  if (coursId !== prevCoursId) {
    setPrevCoursId(coursId)
    if (coursId) {
      setEtat('generation')
      setExamen(null)
      setSession(null)
      setQuestionActive(1)
      setReponses(new Map())
      setIndicesParQuestion(new Map())
      setResultat(null)
      setErreur(null)
    } else {
      setEtat('selection-cours')
    }
  }

  // Lancer la generation automatiquement quand on arrive avec un coursId
  const hasLaunched = useRef(false)
  useEffect(() => {
    if (coursId && etat === 'generation' && !hasLaunched.current) {
      hasLaunched.current = true
      lancerExamen()
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [coursId, etat])

  // Generer et demarrer l'examen
  const lancerExamen = useCallback(async () => {
    if (!coursId) return

    setEtat('generation')
    setErreur(null)

    try {
      // Generer l'examen
      const resExamen = await genererExamen(coursId)
      if (!resExamen.succes || !resExamen.examen) {
        throw new Error(resExamen.erreur?.message || 'Erreur lors de la generation de l\'examen')
      }
      setExamen(resExamen.examen)

      // Demarrer la session
      const resSession = await demarrerSessionExamen(resExamen.examen.id)
      if (!resSession.succes || !resSession.session) {
        throw new Error(resSession.erreur?.message || 'Erreur lors du demarrage de la session')
      }
      setSession(resSession.session)

      // Reinitialiser
      setQuestionActive(1)
      setReponses(new Map())
      setIndicesParQuestion(new Map())
      setResultat(null)
      setEtat('examen')
    } catch (err) {
      setErreur(err instanceof Error ? err.message : 'Une erreur est survenue')
      setEtat('erreur')
    }
  }, [coursId])

  // Mettre a jour une reponse
  const mettreAJourReponse = useCallback((numero: number, texte: string) => {
    setReponses((prev) => {
      const next = new Map(prev)
      if (texte) {
        next.set(numero, texte)
      } else {
        next.delete(numero)
      }
      return next
    })
  }, [])

  // Demander un indice
  const handleDemanderIndice = useCallback(async () => {
    if (!session) return

    const indicesActuels = indicesParQuestion.get(questionActive) || []
    const niveauSuivant = indicesActuels.length + 1
    if (niveauSuivant > 3) return

    setChargementIndice(true)
    try {
      const res = await demanderIndice(session.id, questionActive, niveauSuivant)
      if (res.succes && res.indice) {
        setIndicesParQuestion((prev) => {
          const next = new Map(prev)
          next.set(questionActive, [...indicesActuels, res.indice!])
          return next
        })
      }
    } catch (err) {
      console.error('Erreur lors de la demande d\'indice:', err)
    } finally {
      setChargementIndice(false)
    }
  }, [session, questionActive, indicesParQuestion])

  // Soumettre l'examen pour correction
  const soumettreExamen = useCallback(async () => {
    if (!session || !examen) return

    setEtat('correction')
    setErreur(null)

    try {
      // Construire les reponses
      const reponsesArray: ReponseExamenDetail[] = examen.questions.map((q) => ({
        questionNumero: q.numero,
        texte: reponses.get(q.numero) || '',
      }))

      const res = await corrigerExamen(session.id, reponsesArray)
      if (!res.succes || !res.resultat) {
        throw new Error(res.erreur?.message || 'Erreur lors de la correction')
      }

      setResultat(res.resultat)
      setEtat('resultats')
    } catch (err) {
      setErreur(err instanceof Error ? err.message : 'Une erreur est survenue')
      setEtat('erreur')
    }
  }, [session, examen, reponses])

  // Gestion du temps ecoule
  const handleTempsEcoule = useCallback(() => {
    soumettreExamen()
  }, [soumettreExamen])

  // Recommencer avec un nouvel examen
  const recommencer = useCallback(() => {
    hasLaunched.current = false
    setExamen(null)
    setSession(null)
    setQuestionActive(1)
    setReponses(new Map())
    setIndicesParQuestion(new Map())
    setResultat(null)
    setErreur(null)
    setEtat('selection-cours')
    setSearchParams({})
  }, [setSearchParams])

  // Selection d'un cours
  const selectionnerCours = useCallback(
    (id: string) => {
      hasLaunched.current = false
      setSearchParams({ cours: id })
    },
    [setSearchParams]
  )

  // -- RENDUS --

  // Page de selection de cours
  if (etat === 'selection-cours') {
    if (chargementCours) {
      return (
        <div className="flex items-center justify-center min-h-[400px]">
          <ProcessingSection message="Chargement des cours..." />
        </div>
      )
    }

    if (erreurCours || listeCours.length === 0) {
      return (
        <div className="text-center py-12">
          <div className="text-5xl mb-6">
            {String.fromCodePoint(0x1f4da)}
          </div>
          <h1 className="font-display text-2xl font-semibold text-ink mb-4">
            {erreurCours || 'Aucun cours disponible'}
          </h1>
          <p className="text-ink-light mb-8">
            Scannez d'abord un cours pour generer un examen blanc.
          </p>
          <Link
            to="/scanner"
            className="inline-flex items-center gap-4 px-8 py-3 bg-coral text-white rounded-full font-medium hover:bg-coral-dark transition-colors"
          >
            Scanner un cours
          </Link>
        </div>
      )
    }

    return (
      <div>
        <h1 className="font-display text-xl md:text-3xl font-semibold text-ink mb-4">
          Examen blanc
        </h1>
        <p className="text-ink-light mb-8">
          Selectionnez un cours pour generer un examen blanc avec correction detaillee.
        </p>

        <div className="grid gap-6">
          {listeCours.map((c) => (
            <button
              key={c.id}
              onClick={() => selectionnerCours(c.id)}
              className="bg-white rounded-lg p-8 shadow-sm hover:shadow-md transition-shadow border border-cream-dark text-left"
            >
              <div className="flex items-start justify-between">
                <div>
                  <h2 className="font-display text-lg font-semibold text-ink mb-2">
                    {c.titre || 'Cours sans titre'}
                  </h2>
                  <p className="text-sm text-ink-muted">{c.matiere || 'Matiere non definie'}</p>
                </div>
                <span className="text-coral font-medium">Passer un examen blanc</span>
              </div>
            </button>
          ))}
        </div>
      </div>
    )
  }

  // Page d'erreur
  if (etat === 'erreur') {
    return (
      <div className="text-center py-12">
        <div className="text-5xl mb-6">
          {String.fromCodePoint(0x1f615)}
        </div>
        <h1 className="font-display text-2xl font-semibold text-ink mb-4">
          Une erreur est survenue
        </h1>
        <p className="text-ink-light mb-8">{erreur}</p>
        <div className="flex flex-col sm:flex-row gap-4 justify-center">
          <button
            onClick={recommencer}
            className="px-6 py-2.5 md:px-8 md:py-3 bg-coral text-white rounded-full font-medium hover:bg-coral-dark transition-colors"
          >
            Reessayer
          </button>
          <Link
            to="/examen-blanc"
            className="px-6 py-2.5 md:px-8 md:py-3 border-2 border-coral text-coral rounded-full font-medium hover:bg-coral/5 transition-colors text-center"
          >
            Retour aux cours
          </Link>
        </div>
      </div>
    )
  }

  // Page de generation
  if (etat === 'generation') {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <ProcessingSection message="Generation de l'examen blanc en cours..." />
      </div>
    )
  }

  // Page de correction
  if (etat === 'correction') {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <ProcessingSection message="Correction de votre examen en cours..." />
      </div>
    )
  }

  // Page de l'examen en cours
  if (etat === 'examen' && examen && session) {
    const questionCourante = examen.questions.find((q) => q.numero === questionActive)

    return (
      <div>
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6">
          <div>
            <h1 className="font-display text-xl md:text-2xl font-bold text-ink">
              Examen blanc
            </h1>
            <p className="text-sm text-ink-muted mt-1">
              {examen.questions.length} questions - {examen.dureeMinutes} minutes
            </p>
          </div>
          <div className="flex items-center gap-4 md:gap-6">
            <Timer dureeMinutes={examen.dureeMinutes} onTempsEcoule={handleTempsEcoule} />
            <button
              onClick={soumettreExamen}
              className="px-4 py-2 md:px-6 md:py-2.5 bg-coral text-white rounded-full text-sm font-semibold hover:bg-coral-dark transition-colors whitespace-nowrap"
            >
              Rendre ma copie
            </button>
          </div>
        </div>

        {/* Contenu */}
        <div className="grid grid-cols-1 md:grid-cols-[1fr_200px] gap-6">
          {/* Sidebar sur mobile - affichee en premier */}
          <div className="md:hidden">
            <SidebarQuestions
              nombreQuestions={examen.questions.length}
              questionActive={questionActive}
              reponses={reponses}
              onNaviguer={setQuestionActive}
            />
          </div>

          {/* Zone de question */}
          <div>
            {questionCourante && (
              <ZoneQuestion
                numero={questionCourante.numero}
                enonce={questionCourante.enonce}
                type={questionCourante.type}
                difficulte={questionCourante.difficulte}
                bareme={questionCourante.bareme}
                reponse={reponses.get(questionCourante.numero) || ''}
                indices={indicesParQuestion.get(questionCourante.numero) || []}
                onReponseChange={(texte) => mettreAJourReponse(questionCourante.numero, texte)}
                onDemanderIndice={handleDemanderIndice}
                chargementIndice={chargementIndice}
              />
            )}

            {/* Navigation entre questions */}
            <div className="flex items-center justify-between mt-6">
              <button
                onClick={() => setQuestionActive((prev) => Math.max(1, prev - 1))}
                disabled={questionActive <= 1}
                className="px-4 py-2 md:px-6 rounded-full text-sm font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed bg-white border border-cream-dark hover:bg-cream"
              >
                <span className="hidden sm:inline">Question </span>precedente
              </button>
              <span className="text-sm text-ink-muted">
                {questionActive} / {examen.questions.length}
              </span>
              <button
                onClick={() => setQuestionActive((prev) => Math.min(examen.questions.length, prev + 1))}
                disabled={questionActive >= examen.questions.length}
                className="px-4 py-2 md:px-6 rounded-full text-sm font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed bg-white border border-cream-dark hover:bg-cream"
              >
                <span className="hidden sm:inline">Question </span>suivante
              </button>
            </div>
          </div>

          {/* Sidebar desktop */}
          <div className="hidden md:block space-y-4">
            <SidebarQuestions
              nombreQuestions={examen.questions.length}
              questionActive={questionActive}
              reponses={reponses}
              onNaviguer={setQuestionActive}
            />
          </div>
        </div>
      </div>
    )
  }

  // Page de resultats
  if (etat === 'resultats' && resultat && coursId) {
    return (
      <VueResultats
        resultat={resultat}
        coursId={coursId}
        onRecommencer={recommencer}
      />
    )
  }

  // Fallback
  return null
}
