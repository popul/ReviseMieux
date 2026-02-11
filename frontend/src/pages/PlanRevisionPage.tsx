import { useEffect, useState } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import {
  obtenirPlanComplet,
  supprimerPlanRevision,
  mettreAJourPlanRevision,
  type PlanRevision,
  type CoursAvecArtifacts,
} from '../services/api'

const iconesMatiere: Record<string, string> = {
  mathematiques: '📐',
  francais: '📖',
  histoire: '🏛️',
  geographie: '🗺️',
  sciences: '🔬',
  anglais: '🇬🇧',
  physique: '⚛️',
  chimie: '🧪',
  svt: '🌿',
  ses: '📊',
  philosophie: '🤔',
}

function getIconeMatiere(matiere: string | undefined): string {
  if (!matiere) return '📚'
  return iconesMatiere[matiere.toLowerCase()] || '📚'
}

function CoursAvecArtifactsCard({ ca }: { ca: CoursAvecArtifacts }) {
  const cours = ca.cours

  return (
    <div className="bg-white rounded-lg p-6 shadow-sm border border-cream-dark">
      <div className="flex items-start gap-4">
        <div className="w-12 h-12 rounded-md bg-cream flex items-center justify-center text-xl flex-shrink-0">
          {getIconeMatiere(cours.matiere)}
        </div>
        <div className="flex-1 min-w-0">
          <h3 className="font-semibold text-ink truncate">{cours.titre}</h3>
          <p className="text-xs text-ink-muted capitalize mt-1">{cours.matiere || 'Non classe'}</p>

          {/* Badges artefacts */}
          <div className="flex flex-wrap gap-2 mt-3">
            <span className={`text-xs px-2 py-1 rounded-full ${ca.aResume ? 'bg-success/20 text-success' : 'bg-cream text-ink-muted'}`}>
              Resume {ca.aResume ? '✓' : '—'}
            </span>
            <span className={`text-xs px-2 py-1 rounded-full ${ca.nombreFiches > 0 ? 'bg-coral/20 text-coral' : 'bg-cream text-ink-muted'}`}>
              {ca.nombreFiches} fiche{ca.nombreFiches !== 1 ? 's' : ''}
            </span>
            <span className={`text-xs px-2 py-1 rounded-full ${ca.nombreQuiz > 0 ? 'bg-teal/20 text-teal' : 'bg-cream text-ink-muted'}`}>
              {ca.nombreQuiz} quiz
            </span>
            <span className={`text-xs px-2 py-1 rounded-full ${ca.aMindmap ? 'bg-gold/20 text-gold' : 'bg-cream text-ink-muted'}`}>
              Mindmap {ca.aMindmap ? '✓' : '—'}
            </span>
          </div>
        </div>
      </div>

      {/* Actions */}
      <div className="flex flex-wrap gap-2 mt-4 pt-4 border-t border-cream">
        <Link
          to={`/fiches?cours=${cours.id}`}
          className="px-4 py-1.5 bg-coral text-white rounded-full text-xs font-medium hover:bg-coral-dark transition-colors"
        >
          Fiches
        </Link>
        <Link
          to={`/quiz?cours=${cours.id}`}
          className="px-4 py-1.5 bg-teal text-white rounded-full text-xs font-medium hover:bg-teal-light transition-colors"
        >
          Quiz
        </Link>
        <Link
          to={`/mindmap?cours=${cours.id}`}
          className="px-4 py-1.5 bg-white text-ink border border-ink rounded-full text-xs font-medium hover:bg-cream transition-colors"
        >
          Mindmap
        </Link>
        <Link
          to={`/lexique?cours=${cours.id}`}
          className="px-4 py-1.5 bg-white text-ink border border-ink rounded-full text-xs font-medium hover:bg-cream transition-colors"
        >
          Lexique
        </Link>
        <Link
          to={`/examen-blanc?cours=${cours.id}`}
          className="px-4 py-1.5 bg-ink text-white rounded-full text-xs font-medium hover:bg-ink/80 transition-colors"
        >
          Examen
        </Link>
        <Link
          to={`/cours?id=${cours.id}`}
          className="px-4 py-1.5 bg-cream text-ink rounded-full text-xs font-medium hover:bg-cream-dark transition-colors"
        >
          Voir le cours
        </Link>
      </div>
    </div>
  )
}

export default function PlanRevisionPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()

  const [plan, setPlan] = useState<PlanRevision | null>(null)
  const [coursAvecArtifacts, setCoursAvecArtifacts] = useState<CoursAvecArtifacts[]>([])
  const [chargement, setChargement] = useState(true)
  const [erreur, setErreur] = useState<string | null>(null)

  // Edition inline du titre
  const [editionTitre, setEditionTitre] = useState(false)
  const [titreEdite, setTitreEdite] = useState('')

  // Suppression
  const [confirmationSuppression, setConfirmationSuppression] = useState(false)

  const chargerPlan = async () => {
    if (!id) return
    try {
      setChargement(true)
      const res = await obtenirPlanComplet(id)
      if (res.succes && res.plan) {
        setPlan(res.plan)
        setTitreEdite(res.plan.titre)
        setCoursAvecArtifacts(res.cours || [])
      } else {
        setErreur('Plan non trouve')
      }
    } catch (err) {
      setErreur(err instanceof Error ? err.message : 'Erreur inconnue')
    } finally {
      setChargement(false)
    }
  }

  useEffect(() => {
    chargerPlan()
  }, [id]) // eslint-disable-line react-hooks/exhaustive-deps

  const handleSauvegarderTitre = async () => {
    if (!id || !titreEdite.trim()) return
    try {
      await mettreAJourPlanRevision(id, { titre: titreEdite.trim() })
      setPlan(prev => prev ? { ...prev, titre: titreEdite.trim() } : prev)
      setEditionTitre(false)
    } catch {
      // Ignorer
    }
  }

  const handleSupprimer = async () => {
    if (!id) return
    try {
      await supprimerPlanRevision(id)
      navigate('/')
    } catch {
      // Ignorer
    }
  }

  // Calcul du countdown
  const joursRestants = plan?.dateEcheance
    ? Math.ceil((new Date(plan.dateEcheance).getTime() - Date.now()) / (1000 * 60 * 60 * 24))
    : null

  // Calcul de la progression
  const progression = coursAvecArtifacts.length === 0
    ? 0
    : Math.round(
        (coursAvecArtifacts.filter(ca => ca.nombreFiches > 0).length / coursAvecArtifacts.length) * 100
      )

  if (chargement) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-center">
          <div className="animate-spin h-8 w-8 border-2 border-coral border-t-transparent rounded-full mx-auto mb-4" />
          <p className="text-ink-muted">Chargement du plan...</p>
        </div>
      </div>
    )
  }

  if (erreur || !plan) {
    return (
      <div className="text-center py-12">
        <div className="text-5xl mb-6">😕</div>
        <h1 className="font-display text-2xl font-semibold text-ink mb-4">Plan introuvable</h1>
        <p className="text-ink-light mb-8">{erreur || 'Ce plan n\'existe pas.'}</p>
        <button
          onClick={() => navigate('/')}
          className="px-8 py-3 bg-coral text-white rounded-full font-medium hover:bg-coral-dark transition-colors"
        >
          Retour au tableau de bord
        </button>
      </div>
    )
  }

  return (
    <div>
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center gap-4 mb-4">
          <button
            onClick={() => navigate('/')}
            className="text-ink-light hover:text-ink transition-colors text-sm"
          >
            ← Retour
          </button>
        </div>

        <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-4">
          <div className="flex items-start gap-4">
            <span className="text-4xl">{plan.iconeMatiere || '📋'}</span>
            <div>
              {editionTitre ? (
                <div className="flex items-center gap-2">
                  <input
                    type="text"
                    value={titreEdite}
                    onChange={(e) => setTitreEdite(e.target.value)}
                    onKeyDown={(e) => e.key === 'Enter' && handleSauvegarderTitre()}
                    className="font-display text-2xl font-bold text-ink border-b-2 border-coral focus:outline-none bg-transparent"
                    autoFocus
                  />
                  <button
                    onClick={handleSauvegarderTitre}
                    className="text-sm text-success hover:underline"
                  >
                    Sauvegarder
                  </button>
                  <button
                    onClick={() => { setEditionTitre(false); setTitreEdite(plan.titre) }}
                    className="text-sm text-ink-muted hover:underline"
                  >
                    Annuler
                  </button>
                </div>
              ) : (
                <h1
                  className="font-display text-2xl font-bold text-ink cursor-pointer hover:text-coral transition-colors"
                  onClick={() => setEditionTitre(true)}
                  title="Cliquer pour modifier"
                >
                  {plan.titre}
                </h1>
              )}

              <div className="flex items-center gap-4 mt-2">
                {plan.matiere && (
                  <span className="text-sm text-ink-muted capitalize">{plan.matiere}</span>
                )}
                {joursRestants !== null && (
                  <span className={`text-sm font-medium px-3 py-1 rounded-full ${
                    joursRestants <= 3 ? 'bg-coral/20 text-coral' :
                    joursRestants <= 7 ? 'bg-gold/20 text-gold' :
                    'bg-teal/20 text-teal'
                  }`}>
                    {joursRestants >= 0 ? `J-${joursRestants}` : 'Date passee'}
                  </span>
                )}
                {coursAvecArtifacts.length > 0 && (
                  <span className="text-sm text-ink-muted">
                    {coursAvecArtifacts[0].cours.titre}
                  </span>
                )}
              </div>

              {/* Barre de progression */}
              <div className="flex items-center gap-3 mt-3">
                <div className="w-48 h-2 bg-cream rounded-full overflow-hidden">
                  <div
                    className="h-full bg-teal rounded-full transition-all"
                    style={{ width: `${progression}%` }}
                  />
                </div>
                <span className="text-xs text-ink-muted">{progression}%</span>
              </div>
            </div>
          </div>

          <div className="flex gap-2">
            {confirmationSuppression ? (
              <div className="flex items-center gap-2">
                <span className="text-sm text-ink-muted">Supprimer ce plan ?</span>
                <button
                  onClick={handleSupprimer}
                  className="px-4 py-2 bg-red-500 text-white rounded-full text-sm font-medium hover:bg-red-600 transition-colors"
                >
                  Confirmer
                </button>
                <button
                  onClick={() => setConfirmationSuppression(false)}
                  className="px-4 py-2 bg-cream text-ink rounded-full text-sm font-medium hover:bg-cream-dark transition-colors"
                >
                  Annuler
                </button>
              </div>
            ) : (
              <button
                onClick={() => setConfirmationSuppression(true)}
                className="text-sm text-ink-muted hover:text-red-500 transition-colors"
              >
                Supprimer
              </button>
            )}
          </div>
        </div>

        {plan.description && (
          <p className="text-ink-light mt-4 ml-16">{plan.description}</p>
        )}
      </div>

      {/* Cours du plan */}
      {coursAvecArtifacts.length === 0 ? (
        <div className="bg-white rounded-lg p-8 md:p-12 text-center shadow-sm">
          <div className="text-5xl mb-6">📚</div>
          <h2 className="font-display text-xl font-semibold text-ink mb-4">
            Aucun cours associe
          </h2>
          <p className="text-ink-light mb-8 max-w-md mx-auto">
            Ce plan n'a pas de cours. Cree un nouveau plan en associant un cours.
          </p>
          <Link
            to="/scanner"
            className="px-8 py-3 bg-coral text-white rounded-full font-semibold hover:bg-coral-dark transition-colors"
          >
            Scanner un cours
          </Link>
        </div>
      ) : (
        <div>
          <h2 className="font-display text-xl font-semibold text-ink mb-6">
            Supports de revision
          </h2>
          <div className="space-y-4">
            {coursAvecArtifacts.map((ca) => (
              <CoursAvecArtifactsCard key={ca.cours.id} ca={ca} />
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
