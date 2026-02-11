import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { obtenirCoursRecents, listerPlansRevision } from '../services/api'
import type { CoursResume, PlanRevisionResume } from '../services/api'

// Skeleton pour la liste des cours
function CoursSkeleton() {
  return (
    <div className="bg-white rounded-lg p-6 animate-pulse">
      <div className="flex items-center gap-6">
        <div className="w-12 h-12 rounded-md bg-cream" />
        <div className="flex-1">
          <div className="h-5 bg-cream rounded w-3/4 mb-2" />
          <div className="h-4 bg-cream rounded w-1/2" />
        </div>
        <div className="text-right">
          <div className="h-4 bg-cream rounded w-16 mb-1" />
          <div className="h-3 bg-cream rounded w-12" />
        </div>
      </div>
    </div>
  )
}

// Carte de cours récent
function CoursCard({ cours }: { cours: CoursResume }) {
  const dateCreation = new Date(cours.dateCreation)
  const dateFormatee = dateCreation.toLocaleDateString('fr-FR', {
    day: 'numeric',
    month: 'short',
  })

  return (
    <Link
      to={`/cours?id=${cours.id}`}
      className="block bg-white rounded-lg p-6 hover:shadow-md transition-shadow no-underline"
    >
      <div className="flex items-center gap-6">
        <div className="w-12 h-12 rounded-md bg-cream flex items-center justify-center text-xl flex-shrink-0">
          {cours.matiere === 'mathematiques' ? '📐' :
           cours.matiere === 'francais' ? '📖' :
           cours.matiere === 'histoire' ? '🏛️' :
           cours.matiere === 'geographie' ? '🗺️' :
           cours.matiere === 'sciences' ? '🔬' :
           cours.matiere === 'anglais' ? '🇬🇧' :
           '📚'}
        </div>
        <div className="flex-1 min-w-0">
          <h3 className="font-semibold text-ink truncate">{cours.titre}</h3>
          <p className="text-sm text-ink-light capitalize">{cours.matiere || 'Non classé'}</p>
        </div>
        <div className="text-right text-sm flex-shrink-0">
          <div className="text-ink-light">
            {cours.nombreFiches} fiche{cours.nombreFiches > 1 ? 's' : ''}
          </div>
          <div className="text-ink-lighter text-xs">{dateFormatee}</div>
        </div>
      </div>
    </Link>
  )
}

// Carte d'un plan de revision
function PlanCard({ plan }: { plan: PlanRevisionResume }) {
  const joursRestants = plan.dateEcheance
    ? Math.ceil((new Date(plan.dateEcheance).getTime() - Date.now()) / (1000 * 60 * 60 * 24))
    : null

  return (
    <Link
      to={`/plans/${plan.id}`}
      className="block bg-white rounded-lg p-6 hover:shadow-md transition-shadow no-underline border border-cream-dark"
    >
      <div className="flex items-start gap-4">
        <span className="text-2xl">{plan.iconeMatiere || '📋'}</span>
        <div className="flex-1 min-w-0">
          <h3 className="font-semibold text-ink truncate">{plan.titre}</h3>
          <div className="flex items-center gap-3 mt-2">
            <span className="text-xs text-ink-muted">{plan.nombreCours} cours</span>
            {joursRestants !== null && joursRestants >= 0 && (
              <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${
                joursRestants <= 3 ? 'bg-coral/20 text-coral' :
                joursRestants <= 7 ? 'bg-gold/20 text-gold' :
                'bg-teal/20 text-teal'
              }`}>
                J-{joursRestants}
              </span>
            )}
          </div>
          {/* Barre de progression */}
          <div className="flex items-center gap-2 mt-3">
            <div className="flex-1 h-1.5 bg-cream rounded-full overflow-hidden">
              <div
                className="h-full bg-teal rounded-full transition-all"
                style={{ width: `${plan.progression}%` }}
              />
            </div>
            <span className="text-xs text-ink-muted">{plan.progression}%</span>
          </div>
        </div>
      </div>
    </Link>
  )
}

export default function Dashboard() {
  const [coursRecents, setCoursRecents] = useState<CoursResume[]>([])
  const [plans, setPlans] = useState<PlanRevisionResume[]>([])
  const [chargement, setChargement] = useState(true)
  const [erreur, setErreur] = useState<string | null>(null)

  useEffect(() => {
    async function chargerDonnees() {
      try {
        setChargement(true)
        setErreur(null)

        const [coursReponse, plansReponse] = await Promise.all([
          obtenirCoursRecents(),
          listerPlansRevision(),
        ])

        if (coursReponse.succes) {
          setCoursRecents(coursReponse.cours || [])
        }

        setPlans(plansReponse || [])
      } catch (err) {
        setErreur(err instanceof Error ? err.message : 'Erreur inconnue')
      } finally {
        setChargement(false)
      }
    }

    chargerDonnees()
  }, [])

  return (
    <>
      {/* Header */}
      <header className="mb-8 md:mb-12">
        <h1 className="font-display text-2xl md:text-4xl font-bold text-ink mb-2">
          Bienvenue sur Révise mieux
        </h1>
        <p className="text-ink-light text-lg">
          Transforme tes cours en fiches de révision et quiz interactifs
        </p>
      </header>

      {/* Quick Actions */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 md:gap-6 mb-8 md:mb-12">
        <Link
          to="/scanner"
          className="bg-coral text-white rounded-lg p-5 md:p-8 flex items-center gap-4 md:gap-6 no-underline transition-all hover:-translate-y-1 hover:shadow-xl hover:bg-coral-dark"
        >
          <div className="w-14 h-14 rounded-md bg-white/20 flex items-center justify-center text-2xl flex-shrink-0">
            <span role="img" aria-label="Scanner">📸</span>
          </div>
          <div>
            <h3 className="font-display text-lg font-semibold mb-1">Scanner un cours</h3>
            <p className="text-sm opacity-80">Transforme tes notes en fiches et quiz</p>
          </div>
        </Link>
        <Link
          to="/plans/nouveau"
          className="bg-teal text-white rounded-lg p-5 md:p-8 flex items-center gap-4 md:gap-6 no-underline transition-all hover:-translate-y-1 hover:shadow-xl hover:bg-teal-light"
        >
          <div className="w-14 h-14 rounded-md bg-white/20 flex items-center justify-center text-2xl flex-shrink-0">
            <span role="img" aria-label="Plan">📋</span>
          </div>
          <div>
            <h3 className="font-display text-lg font-semibold mb-1">Creer un plan</h3>
            <p className="text-sm opacity-80">Organise tes revisions</p>
          </div>
        </Link>
      </div>

      {/* Plans de revision */}
      {!chargement && plans.length > 0 && (
        <section className="mb-12">
          <div className="flex items-center justify-between mb-6">
            <h2 className="font-display text-xl font-semibold text-ink">Tes plans de revision</h2>
            <Link to="/plans/nouveau" className="text-teal text-sm font-medium hover:underline">
              + Nouveau plan
            </Link>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {plans.map((plan) => (
              <PlanCard key={plan.id} plan={plan} />
            ))}
          </div>
        </section>
      )}

      {/* Erreur */}
      {erreur && (
        <div className="bg-error/10 text-error rounded-lg p-6 mb-12">
          <p className="font-medium">Erreur de chargement</p>
          <p className="text-sm">{erreur}</p>
        </div>
      )}

      {/* Cours récents ou état vide */}
      {chargement ? (
        <section className="space-y-4">
          <h2 className="font-display text-xl font-semibold text-ink mb-6">Tes cours récents</h2>
          <CoursSkeleton />
          <CoursSkeleton />
          <CoursSkeleton />
        </section>
      ) : coursRecents.length > 0 ? (
        <section>
          <div className="flex items-center justify-between mb-6">
            <h2 className="font-display text-xl font-semibold text-ink">Tes cours récents</h2>
            <Link to="/cours" className="text-coral text-sm font-medium hover:underline">
              Voir tous
            </Link>
          </div>
          <div className="space-y-4">
            {coursRecents.slice(0, 5).map((cours) => (
              <CoursCard key={cours.id} cours={cours} />
            ))}
          </div>
        </section>
      ) : (
        <section className="bg-white rounded-lg p-8 md:p-12 text-center">
          <div className="text-5xl mb-6" role="img" aria-label="Livres">📚</div>
          <h2 className="font-display text-xl font-semibold text-ink mb-4">
            Aucun cours pour le moment
          </h2>
          <p className="text-ink-light mb-8 max-w-md mx-auto">
            Commence par scanner un cours pour générer des fiches de révision et des quiz interactifs.
          </p>
          <Link
            to="/scanner"
            className="inline-flex items-center gap-2 bg-coral text-white px-8 py-3 rounded-full font-semibold no-underline transition-all hover:bg-coral-dark hover:-translate-y-0.5"
          >
            <span role="img" aria-label="Scanner">📸</span>
            Scanner mon premier cours
          </Link>
        </section>
      )}
    </>
  )
}
