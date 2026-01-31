import { useState, useMemo, useCallback, useLayoutEffect } from 'react'
import { useSearchParams, Link } from 'react-router-dom'
import { obtenirFichesCours, listerCours, type Fiche, type Cours } from '../services/api'
import CarteFiche from '../components/CarteFiche'
import ControlesFiches from '../components/ControlesFiches'
import ListeFichesSidebar from '../components/ListeFichesSidebar'
import FiltreDifficulte, { type Difficulte } from '../components/FiltreDifficulte'
import ProcessingSection from '../components/ProcessingSection'

type ModeAffichage = 'reviser' | 'lire'

// Hook pour charger les cours
function useChargementCours(coursId: string | null) {
  const [listeCours, setListeCours] = useState<Cours[]>([])
  // État initial: en chargement si pas de coursId
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

// Hook pour charger les fiches d'un cours
function useChargementFiches(coursId: string | null) {
  const [fiches, setFiches] = useState<Fiche[]>([])
  // État initial: en chargement si coursId est présent
  const [chargement, setChargement] = useState(() => !!coursId)
  const [erreur, setErreur] = useState<string | null>(null)
  // Track coursId changes to trigger reload
  const [prevCoursId, setPrevCoursId] = useState(coursId)

  // Reset state when coursId changes (derived state pattern)
  if (coursId !== prevCoursId) {
    setPrevCoursId(coursId)
    if (coursId) {
      setChargement(true)
      setErreur(null)
      setFiches([])
    }
  }

  useLayoutEffect(() => {
    if (!coursId) return

    let cancelled = false

    obtenirFichesCours(coursId)
      .then((res) => {
        if (!cancelled) {
          if (res.succes) {
            setFiches(res.fiches)
          } else {
            setErreur(res.erreur?.message || 'Erreur lors du chargement')
          }
          setChargement(false)
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setErreur(err.message)
          setChargement(false)
        }
      })

    return () => {
      cancelled = true
    }
  }, [coursId])

  return { fiches, chargement, erreur }
}

export default function Fiches() {
  const [searchParams] = useSearchParams()
  const coursId = searchParams.get('cours')

  const { listeCours, chargement: chargementCours, erreur: erreurCours } = useChargementCours(coursId)
  const { fiches, chargement: chargementFiches, erreur: erreurFiches } = useChargementFiches(coursId)

  const [indexActuel, setIndexActuel] = useState(0)
  const [modeAffichage, setModeAffichage] = useState<ModeAffichage>('reviser')
  const [filtreDifficulte, setFiltreDifficulte] = useState<Difficulte>('toutes')
  const [prevFiltre, setPrevFiltre] = useState<Difficulte>('toutes')

  // Reset index quand le filtre change (pattern derived state)
  if (filtreDifficulte !== prevFiltre) {
    setPrevFiltre(filtreDifficulte)
    setIndexActuel(0)
  }

  // Fiches filtrées par difficulté
  const fichesFiltrees = useMemo(() => {
    if (filtreDifficulte === 'toutes') return fiches
    return fiches.filter((f) => f.difficulte === filtreDifficulte)
  }, [fiches, filtreDifficulte])

  const ficheCourante = fichesFiltrees[indexActuel]

  const allerPrecedent = useCallback(() => {
    setIndexActuel((prev) => (prev > 0 ? prev - 1 : prev))
  }, [])

  const allerSuivant = useCallback(() => {
    setIndexActuel((prev) => (prev < fichesFiltrees.length - 1 ? prev + 1 : prev))
  }, [fichesFiltrees.length])

  // Gestion clavier pour navigation
  useLayoutEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'ArrowLeft') allerPrecedent()
      if (e.key === 'ArrowRight') allerSuivant()
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [allerPrecedent, allerSuivant])

  // Page de sélection de cours (si pas de coursId)
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
        <div className="text-center py-xl">
          <div className="text-5xl mb-md">📚</div>
          <h1 className="font-display text-2xl font-semibold text-ink mb-sm">
            {erreurCours || 'Aucun cours disponible'}
          </h1>
          <p className="text-ink-light mb-lg">
            Scannez d'abord un cours pour générer des fiches de révision.
          </p>
          <Link
            to="/scanner"
            className="inline-flex items-center gap-sm px-lg py-3 bg-coral text-white rounded-full font-medium hover:bg-coral-dark transition-colors"
          >
            Scanner un cours
          </Link>
        </div>
      )
    }

    return (
      <div>
        <h1 className="font-display text-3xl font-semibold text-ink mb-lg">
          Fiches de révision
        </h1>
        <p className="text-ink-light mb-lg">
          Sélectionnez un cours pour réviser ses fiches.
        </p>

        <div className="grid gap-md">
          {listeCours.map((c) => (
            <Link
              key={c.id}
              to={`/fiches?cours=${c.id}`}
              className="bg-white rounded-lg p-lg shadow-sm hover:shadow-md transition-shadow border border-cream-dark"
            >
              <div className="flex items-start justify-between">
                <div>
                  <h2 className="font-display text-lg font-semibold text-ink mb-xs">
                    {c.titre || 'Cours sans titre'}
                  </h2>
                  <p className="text-sm text-ink-muted">{c.matiere || 'Matière non définie'}</p>
                </div>
                <span className="text-coral font-medium">Réviser →</span>
              </div>
            </Link>
          ))}
        </div>
      </div>
    )
  }

  // Page de révision des fiches
  if (chargementFiches) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <ProcessingSection message="Chargement des fiches..." />
      </div>
    )
  }

  if (erreurFiches) {
    return (
      <div className="text-center py-xl">
        <div className="text-5xl mb-md">😕</div>
        <h1 className="font-display text-2xl font-semibold text-ink mb-sm">
          Erreur
        </h1>
        <p className="text-ink-light mb-lg">{erreurFiches}</p>
        <Link
          to="/fiches"
          className="inline-flex items-center gap-sm px-lg py-3 bg-coral text-white rounded-full font-medium hover:bg-coral-dark transition-colors"
        >
          ← Retour aux cours
        </Link>
      </div>
    )
  }

  if (fiches.length === 0) {
    return (
      <div className="text-center py-xl">
        <div className="text-5xl mb-md">📄</div>
        <h1 className="font-display text-2xl font-semibold text-ink mb-sm">
          Aucune fiche disponible
        </h1>
        <p className="text-ink-light mb-lg">
          Ce cours n'a pas encore de fiches de révision.
        </p>
        <Link
          to="/fiches"
          className="inline-flex items-center gap-sm px-lg py-3 bg-coral text-white rounded-full font-medium hover:bg-coral-dark transition-colors"
        >
          ← Retour aux cours
        </Link>
      </div>
    )
  }

  return (
    <div className="flex min-h-[calc(100vh-6rem)] -mx-xl -mt-xl">
      {/* Sidebar avec liste des fiches */}
      <ListeFichesSidebar
        fiches={fichesFiltrees}
        indexActuel={indexActuel}
        onSelectFiche={setIndexActuel}
      />

      {/* Contenu principal */}
      <main className="flex-1 p-xl flex flex-col items-center">
        {/* Header avec actions */}
        <div className="w-full max-w-[700px] flex items-center justify-between mb-lg">
          <Link
            to="/fiches"
            className="flex items-center gap-xs text-ink-light hover:text-ink transition-colors"
          >
            ← Retour
          </Link>
          <div className="flex items-center gap-sm">
            <Link
              to={`/quiz?cours=${coursId}`}
              className="px-md py-2.5 bg-coral text-white rounded-full text-sm font-medium hover:bg-coral-dark transition-colors"
            >
              Lancer un quiz
            </Link>
          </div>
        </div>

        {/* Toggle mode affichage */}
        <div className="flex gap-xs bg-white p-1 rounded-full shadow-sm mb-lg">
          <button
            onClick={() => setModeAffichage('reviser')}
            className={`px-md py-2.5 rounded-full text-sm font-medium transition-all ${
              modeAffichage === 'reviser'
                ? 'bg-ink text-white'
                : 'text-ink-light hover:bg-cream'
            }`}
          >
            Réviser
          </button>
          <button
            onClick={() => setModeAffichage('lire')}
            className={`px-md py-2.5 rounded-full text-sm font-medium transition-all ${
              modeAffichage === 'lire'
                ? 'bg-ink text-white'
                : 'text-ink-light hover:bg-cream'
            }`}
          >
            Lire tout
          </button>
        </div>

        {/* Filtre difficulté */}
        <div className="mb-lg">
          <FiltreDifficulte
            valeur={filtreDifficulte}
            onChange={setFiltreDifficulte}
          />
        </div>

        {/* Progression */}
        <div className="w-full max-w-[600px] mb-lg">
          <div className="flex justify-between text-sm mb-xs">
            <span className="text-ink-muted">Progression</span>
            <span className="font-semibold text-ink">
              {indexActuel + 1} / {fichesFiltrees.length}
            </span>
          </div>
          <div className="h-1.5 bg-cream-dark rounded-full overflow-hidden">
            <div
              className="h-full bg-coral rounded-full transition-all duration-300"
              style={{ width: `${((indexActuel + 1) / fichesFiltrees.length) * 100}%` }}
            />
          </div>
        </div>

        {fichesFiltrees.length === 0 ? (
          <div className="text-center py-lg">
            <p className="text-ink-muted">
              Aucune fiche ne correspond à ce filtre.
            </p>
          </div>
        ) : modeAffichage === 'reviser' ? (
          /* Mode révision - Flashcard */
          <>
            <CarteFiche fiche={ficheCourante} />
            <div className="mt-lg">
              <ControlesFiches
                indexActuel={indexActuel}
                total={fichesFiltrees.length}
                onPrecedent={allerPrecedent}
                onSuivant={allerSuivant}
              />
            </div>
          </>
        ) : (
          /* Mode lecture - Liste de toutes les fiches */
          <div className="w-full max-w-[700px] space-y-md">
            {fichesFiltrees.map((fiche, index) => (
              <div
                key={fiche.id}
                className="bg-white rounded-lg p-lg shadow-sm hover:shadow-md transition-shadow"
              >
                <div className="flex items-start justify-between mb-sm">
                  <span className={`text-xs font-semibold uppercase tracking-wide px-2 py-1 rounded-full ${
                    fiche.difficulte === 'facile' ? 'bg-green-100 text-green-700' :
                    fiche.difficulte === 'moyen' ? 'bg-gold/20 text-amber-700' :
                    'bg-red-100 text-red-700'
                  }`}>
                    {fiche.difficulte}
                  </span>
                  <span className="text-sm text-ink-muted">#{index + 1}</span>
                </div>
                <h3 className="font-display text-lg font-semibold text-ink mb-sm">
                  {fiche.question}
                </h3>
                <p className="text-ink-light leading-relaxed">
                  {fiche.reponse}
                </p>
              </div>
            ))}
          </div>
        )}
      </main>
    </div>
  )
}
