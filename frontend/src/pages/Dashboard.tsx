import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { obtenirStatistiques, obtenirCoursRecents } from '../services/api'
import type { Statistiques, CoursResume } from '../services/api'

// Skeleton pour les statistiques pendant le chargement
function StatSkeleton() {
  return (
    <div className="bg-white rounded-lg p-md text-center animate-pulse">
      <div className="h-10 bg-cream rounded w-12 mx-auto mb-1" />
      <div className="h-4 bg-cream rounded w-20 mx-auto" />
    </div>
  )
}

// Skeleton pour la liste des cours
function CoursSkeleton() {
  return (
    <div className="bg-white rounded-lg p-md animate-pulse">
      <div className="flex items-center gap-md">
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

// Carte de statistique
function StatCard({ valeur, label, couleur }: { valeur: string | number; label: string; couleur: string }) {
  return (
    <div className="bg-white rounded-lg p-md text-center">
      <div className={`font-display text-4xl font-bold ${couleur} mb-1`}>{valeur}</div>
      <div className="text-sm text-ink-light">{label}</div>
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
      to={`/fiches?coursId=${cours.id}`}
      className="block bg-white rounded-lg p-md hover:shadow-md transition-shadow no-underline"
    >
      <div className="flex items-center gap-md">
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

export default function Dashboard() {
  const [statistiques, setStatistiques] = useState<Statistiques | null>(null)
  const [coursRecents, setCoursRecents] = useState<CoursResume[]>([])
  const [chargement, setChargement] = useState(true)
  const [erreur, setErreur] = useState<string | null>(null)

  useEffect(() => {
    async function chargerDonnees() {
      try {
        setChargement(true)
        setErreur(null)

        // Charger les statistiques et les cours en parallèle
        const [statsReponse, coursReponse] = await Promise.all([
          obtenirStatistiques(),
          obtenirCoursRecents(),
        ])

        if (statsReponse.succes && statsReponse.statistiques) {
          setStatistiques(statsReponse.statistiques)
        }

        if (coursReponse.succes) {
          setCoursRecents(coursReponse.cours)
        }
      } catch (err) {
        setErreur(err instanceof Error ? err.message : 'Erreur inconnue')
      } finally {
        setChargement(false)
      }
    }

    chargerDonnees()
  }, [])

  const scoreMoyenFormate = statistiques?.scoreMoyen
    ? `${Math.round(statistiques.scoreMoyen)}%`
    : '—'

  return (
    <>
      {/* Header */}
      <header className="mb-xl">
        <h1 className="font-display text-4xl font-bold text-ink mb-xs">
          Bienvenue sur Révise mieux
        </h1>
        <p className="text-ink-light text-lg">
          Transforme tes cours en fiches de révision et quiz interactifs
        </p>
      </header>

      {/* Quick Actions */}
      <div className="grid grid-cols-2 gap-md mb-xl">
        <Link
          to="/scanner"
          className="bg-coral text-white rounded-lg p-lg flex items-center gap-md no-underline transition-all hover:-translate-y-1 hover:shadow-xl hover:bg-coral-dark"
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
          to="/quiz"
          className="bg-white text-ink rounded-lg p-lg flex items-center gap-md no-underline transition-all border-2 border-transparent hover:-translate-y-1 hover:shadow-xl"
        >
          <div className="w-14 h-14 rounded-md bg-cream flex items-center justify-center text-2xl flex-shrink-0">
            <span role="img" aria-label="Quiz">❓</span>
          </div>
          <div>
            <h3 className="font-display text-lg font-semibold mb-1">Passer un quiz</h3>
            <p className="text-sm text-ink-light">Teste tes connaissances</p>
          </div>
        </Link>
      </div>

      {/* Erreur */}
      {erreur && (
        <div className="bg-error/10 text-error rounded-lg p-md mb-xl">
          <p className="font-medium">Erreur de chargement</p>
          <p className="text-sm">{erreur}</p>
        </div>
      )}

      {/* Stats */}
      <section className="mb-xl">
        <h2 className="font-display text-xl font-semibold text-ink mb-md">Tes statistiques</h2>
        <div className="grid grid-cols-4 gap-md">
          {chargement ? (
            <>
              <StatSkeleton />
              <StatSkeleton />
              <StatSkeleton />
              <StatSkeleton />
            </>
          ) : (
            <>
              <StatCard valeur={statistiques?.nombreCours ?? 0} label="Cours scannés" couleur="text-coral" />
              <StatCard valeur={statistiques?.quizCompletes ?? 0} label="Quiz complétés" couleur="text-teal" />
              <StatCard valeur={statistiques?.nombreFiches ?? 0} label="Fiches créées" couleur="text-gold" />
              <StatCard valeur={scoreMoyenFormate} label="Score moyen" couleur="text-success" />
            </>
          )}
        </div>
      </section>

      {/* Cours récents ou état vide */}
      {chargement ? (
        <section className="space-y-sm">
          <h2 className="font-display text-xl font-semibold text-ink mb-md">Tes cours récents</h2>
          <CoursSkeleton />
          <CoursSkeleton />
          <CoursSkeleton />
        </section>
      ) : coursRecents.length > 0 ? (
        <section>
          <div className="flex items-center justify-between mb-md">
            <h2 className="font-display text-xl font-semibold text-ink">Tes cours récents</h2>
            <Link to="/fiches" className="text-coral text-sm font-medium hover:underline">
              Voir tous
            </Link>
          </div>
          <div className="space-y-sm">
            {coursRecents.slice(0, 5).map((cours) => (
              <CoursCard key={cours.id} cours={cours} />
            ))}
          </div>
        </section>
      ) : (
        <section className="bg-white rounded-lg p-xl text-center">
          <div className="text-5xl mb-md" role="img" aria-label="Livres">📚</div>
          <h2 className="font-display text-xl font-semibold text-ink mb-sm">
            Aucun cours pour le moment
          </h2>
          <p className="text-ink-light mb-lg max-w-md mx-auto">
            Commence par scanner un cours pour générer des fiches de révision et des quiz interactifs.
          </p>
          <Link
            to="/scanner"
            className="inline-flex items-center gap-2 bg-coral text-white px-lg py-3 rounded-full font-semibold no-underline transition-all hover:bg-coral-dark hover:-translate-y-0.5"
          >
            <span role="img" aria-label="Scanner">📸</span>
            Scanner mon premier cours
          </Link>
        </section>
      )}
    </>
  )
}
