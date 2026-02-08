import { useState, useLayoutEffect } from 'react'
import { useSearchParams, Link } from 'react-router-dom'
import {
  listerCours,
  obtenirMindmap,
  genererMindmap,
  type Cours,
  type Mindmap as MindmapType,
} from '../services/api'
import MindmapView from '../components/MindmapView'
import ProcessingSection from '../components/ProcessingSection'

// Hook pour charger les cours (si pas de coursId)
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
          setListeCours(res.cours)
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

// Hook pour charger/generer la mindmap d'un cours
function useMindmap(coursId: string | null) {
  const [mindmap, setMindmap] = useState<MindmapType | null>(null)
  const [chargement, setChargement] = useState(() => !!coursId)
  const [generation, setGeneration] = useState(false)
  const [erreur, setErreur] = useState<string | null>(null)
  const [prevCoursId, setPrevCoursId] = useState(coursId)

  // Reset state when coursId changes
  if (coursId !== prevCoursId) {
    setPrevCoursId(coursId)
    if (coursId) {
      setChargement(true)
      setErreur(null)
      setMindmap(null)
    }
  }

  // Charger mindmap existante
  useLayoutEffect(() => {
    if (!coursId) return

    let cancelled = false

    obtenirMindmap(coursId)
      .then((res) => {
        if (!cancelled) {
          if (res.succes && res.mindmap) {
            setMindmap(res.mindmap)
          }
          // Pas de mindmap existante n'est pas une erreur
          setChargement(false)
        }
      })
      .catch(() => {
        if (!cancelled) {
          // 404 signifie juste pas de mindmap existante
          setChargement(false)
        }
      })

    return () => {
      cancelled = true
    }
  }, [coursId])

  // Fonction pour generer une nouvelle mindmap
  const generer = async () => {
    if (!coursId) return

    setGeneration(true)
    setErreur(null)

    try {
      const res = await genererMindmap(coursId)
      if (res.succes && res.mindmap) {
        setMindmap(res.mindmap)
      } else {
        setErreur(res.erreur?.message || 'Erreur lors de la generation')
      }
    } catch (err) {
      setErreur(err instanceof Error ? err.message : 'Erreur inconnue')
    } finally {
      setGeneration(false)
    }
  }

  return { mindmap, chargement, generation, erreur, generer }
}

export default function Mindmap() {
  const [searchParams] = useSearchParams()
  const coursId = searchParams.get('cours')

  const { listeCours, chargement: chargementCours, erreur: erreurCours } = useChargementCours(coursId)
  const { mindmap, chargement: chargementMindmap, generation, erreur: erreurMindmap, generer } = useMindmap(coursId)

  // Page de selection de cours (si pas de coursId)
  if (!coursId) {
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
          <div className="text-5xl mb-6">🧠</div>
          <h1 className="font-display text-2xl font-semibold text-ink mb-4">
            {erreurCours || 'Aucun cours disponible'}
          </h1>
          <p className="text-ink-light mb-8">
            Scannez d'abord un cours pour generer une carte mentale.
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
        <h1 className="font-display text-3xl font-semibold text-ink mb-8">
          Cartes mentales
        </h1>
        <p className="text-ink-light mb-8">
          Selectionnez un cours pour visualiser ou generer sa carte mentale.
        </p>

        <div className="grid gap-6">
          {listeCours.map((c) => (
            <Link
              key={c.id}
              to={`/mindmap?cours=${c.id}`}
              className="bg-white rounded-lg p-8 shadow-sm hover:shadow-md transition-shadow border border-cream-dark"
            >
              <div className="flex items-start justify-between">
                <div>
                  <h2 className="font-display text-lg font-semibold text-ink mb-2">
                    {c.titre || 'Cours sans titre'}
                  </h2>
                  <p className="text-sm text-ink-muted">{c.matiere || 'Matiere non definie'}</p>
                </div>
                <span className="text-coral font-medium">Voir la carte →</span>
              </div>
            </Link>
          ))}
        </div>
      </div>
    )
  }

  // Page de la mindmap
  if (chargementMindmap) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <ProcessingSection message="Chargement de la carte mentale..." />
      </div>
    )
  }

  // Pas de mindmap - proposer de generer
  if (!mindmap && !generation) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[500px]">
        <div className="text-center max-w-md">
          <div className="w-24 h-24 mx-auto mb-8 bg-cream rounded-full flex items-center justify-center">
            <svg width="48" height="48" viewBox="0 0 48 48" fill="none" className="text-teal">
              <circle cx="24" cy="24" r="8" stroke="currentColor" strokeWidth="3" />
              <line x1="24" y1="6" x2="24" y2="14" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
              <line x1="24" y1="34" x2="24" y2="42" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
              <line x1="6" y1="24" x2="14" y2="24" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
              <line x1="34" y1="24" x2="42" y2="24" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
              <circle cx="24" cy="6" r="4" fill="currentColor" />
              <circle cx="24" cy="42" r="4" fill="currentColor" />
              <circle cx="6" cy="24" r="4" fill="currentColor" />
              <circle cx="42" cy="24" r="4" fill="currentColor" />
            </svg>
          </div>

          <h1 className="font-display text-2xl font-semibold text-ink mb-4">
            Pas encore de carte mentale
          </h1>
          <p className="text-ink-light mb-8">
            Generez une carte mentale pour visualiser les concepts cles de ce cours et leurs relations.
          </p>

          {erreurMindmap && (
            <div className="mb-6 p-6 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
              {erreurMindmap}
            </div>
          )}

          <div className="flex flex-col gap-4 items-center">
            <button
              onClick={generer}
              className="px-8 py-3 bg-coral text-white rounded-full font-medium hover:bg-coral-dark transition-colors"
            >
              Generer la carte mentale
            </button>
            <Link
              to="/mindmap"
              className="text-sm text-ink-light hover:text-ink transition-colors"
            >
              ← Choisir un autre cours
            </Link>
          </div>
        </div>
      </div>
    )
  }

  // Generation en cours
  if (generation) {
    return (
      <div className="flex items-center justify-center min-h-[500px]">
        <ProcessingSection message="Generation de la carte mentale..." />
      </div>
    )
  }

  // Affichage de la mindmap
  return (
    <div className="flex flex-col h-[calc(100vh-8rem)]">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-6">
          <Link
            to="/mindmap"
            className="flex items-center gap-2 text-ink-light hover:text-ink transition-colors"
          >
            ← Retour
          </Link>
          <h1 className="font-display text-xl font-semibold text-ink">
            Carte mentale
          </h1>
        </div>

        <div className="flex items-center gap-4">
          <span className="text-sm text-ink-muted">
            {mindmap!.noeuds.length} concepts
          </span>
          <button
            onClick={generer}
            disabled={generation}
            className="px-6 py-2 text-sm text-coral border border-coral rounded-full hover:bg-coral hover:text-white transition-colors disabled:opacity-50"
          >
            Regenerer
          </button>
          <Link
            to={`/fiches?cours=${coursId}`}
            className="px-6 py-2 text-sm bg-coral text-white rounded-full hover:bg-coral-dark transition-colors"
          >
            Voir les fiches
          </Link>
        </div>
      </div>

      {/* Mindmap View */}
      <div className="flex-1 bg-white rounded-xl shadow-sm overflow-hidden">
        <MindmapView mindmap={mindmap!} />
      </div>
    </div>
  )
}
