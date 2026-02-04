import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { obtenirProgression, obtenirStatistiques } from '../services/api'
import type { HistoriqueQuiz, StatsParMatiere, Statistiques } from '../services/api'

// Icônes par matière
function getIconeMatiere(matiere: string): string {
  const icones: Record<string, string> = {
    mathematiques: '📐',
    francais: '📖',
    histoire: '🏛️',
    geographie: '🗺️',
    sciences: '🔬',
    anglais: '🇬🇧',
    physique: '⚛️',
    chimie: '🧪',
    svt: '🌿',
    'Non classé': '📚',
  }
  return icones[matiere] || '📚'
}

// Couleur selon le score
function getCouleurScore(score: number): string {
  if (score >= 80) return 'text-success'
  if (score >= 60) return 'text-gold'
  if (score >= 40) return 'text-warning'
  return 'text-error'
}

function getBgCouleurScore(score: number): string {
  if (score >= 80) return 'bg-success/10'
  if (score >= 60) return 'bg-gold/10'
  if (score >= 40) return 'bg-warning/10'
  return 'bg-error/10'
}

// Skeleton pour les cartes
function CardSkeleton() {
  return (
    <div className="bg-white rounded-lg p-md animate-pulse">
      <div className="flex items-center gap-md">
        <div className="w-12 h-12 rounded-md bg-cream" />
        <div className="flex-1">
          <div className="h-5 bg-cream rounded w-3/4 mb-2" />
          <div className="h-4 bg-cream rounded w-1/2" />
        </div>
        <div className="h-8 w-16 bg-cream rounded" />
      </div>
    </div>
  )
}

// Carte de statistique par matière
function MatiereCard({ stats }: { stats: StatsParMatiere }) {
  return (
    <div className="bg-white rounded-lg p-md">
      <div className="flex items-center gap-md mb-sm">
        <div className="w-10 h-10 rounded-md bg-cream flex items-center justify-center text-xl">
          {getIconeMatiere(stats.matiere)}
        </div>
        <h3 className="font-semibold text-ink capitalize flex-1">{stats.matiere}</h3>
        <span className="text-sm text-ink-light">{stats.nombreQuiz} quiz</span>
      </div>
      <div className="flex items-center justify-between">
        <div className="text-sm">
          <span className="text-ink-light">Moyenne: </span>
          <span className={`font-semibold ${getCouleurScore(stats.scoreMoyen)}`}>
            {Math.round(stats.scoreMoyen)}%
          </span>
        </div>
        <div className="text-sm">
          <span className="text-ink-light">Meilleur: </span>
          <span className="font-semibold text-success">{Math.round(stats.meilleurScore)}%</span>
        </div>
      </div>
    </div>
  )
}

// Carte d'historique
function HistoriqueCard({ entry }: { entry: HistoriqueQuiz }) {
  const dateFin = new Date(entry.dateFin)
  const dateFormatee = dateFin.toLocaleDateString('fr-FR', {
    day: 'numeric',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  })

  return (
    <Link
      to={`/cours?id=${entry.coursId}`}
      className="block bg-white rounded-lg p-md hover:shadow-md transition-shadow no-underline"
    >
      <div className="flex items-center gap-md">
        <div className="w-12 h-12 rounded-md bg-cream flex items-center justify-center text-xl flex-shrink-0">
          {getIconeMatiere(entry.matiere)}
        </div>
        <div className="flex-1 min-w-0">
          <h3 className="font-semibold text-ink truncate">{entry.quizTitre}</h3>
          <p className="text-sm text-ink-light truncate">{entry.coursTitre}</p>
        </div>
        <div className="text-right flex-shrink-0">
          <div
            className={`text-lg font-bold px-3 py-1 rounded-full ${getCouleurScore(entry.score)} ${getBgCouleurScore(entry.score)}`}
          >
            {Math.round(entry.score)}%
          </div>
          <div className="text-xs text-ink-lighter mt-1">{dateFormatee}</div>
        </div>
      </div>
    </Link>
  )
}

export default function Progression() {
  const [historique, setHistorique] = useState<HistoriqueQuiz[]>([])
  const [parMatiere, setParMatiere] = useState<StatsParMatiere[]>([])
  const [statistiques, setStatistiques] = useState<Statistiques | null>(null)
  const [chargement, setChargement] = useState(true)
  const [erreur, setErreur] = useState<string | null>(null)

  useEffect(() => {
    async function chargerDonnees() {
      try {
        setChargement(true)
        setErreur(null)

        const [progressionReponse, statsReponse] = await Promise.all([
          obtenirProgression(),
          obtenirStatistiques(),
        ])

        if (progressionReponse.succes) {
          setHistorique(progressionReponse.historique || [])
          setParMatiere(progressionReponse.parMatiere || [])
        }

        if (statsReponse.succes && statsReponse.statistiques) {
          setStatistiques(statsReponse.statistiques)
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
        <h1 className="font-display text-4xl font-bold text-ink mb-xs">Ma progression</h1>
        <p className="text-ink-light text-lg">Suis ton évolution et identifie tes points forts</p>
      </header>

      {/* Erreur */}
      {erreur && (
        <div className="bg-error/10 text-error rounded-lg p-md mb-xl">
          <p className="font-medium">Erreur de chargement</p>
          <p className="text-sm">{erreur}</p>
        </div>
      )}

      {/* Stats globales */}
      <section className="mb-xl">
        <div className="grid grid-cols-3 gap-md">
          <div className="bg-white rounded-lg p-md text-center">
            <div className="font-display text-4xl font-bold text-teal mb-1">
              {chargement ? '—' : (statistiques?.quizCompletes ?? 0)}
            </div>
            <div className="text-sm text-ink-light">Quiz complétés</div>
          </div>
          <div className="bg-white rounded-lg p-md text-center">
            <div className={`font-display text-4xl font-bold mb-1 ${statistiques?.scoreMoyen ? getCouleurScore(statistiques.scoreMoyen) : 'text-ink'}`}>
              {chargement ? '—' : scoreMoyenFormate}
            </div>
            <div className="text-sm text-ink-light">Score moyen</div>
          </div>
          <div className="bg-white rounded-lg p-md text-center">
            <div className="font-display text-4xl font-bold text-coral mb-1">
              {chargement ? '—' : parMatiere.length}
            </div>
            <div className="text-sm text-ink-light">Matières</div>
          </div>
        </div>
      </section>

      {chargement ? (
        <>
          {/* Skeleton stats par matière */}
          <section className="mb-xl">
            <h2 className="font-display text-xl font-semibold text-ink mb-md">
              Statistiques par matière
            </h2>
            <div className="grid grid-cols-2 gap-md">
              <CardSkeleton />
              <CardSkeleton />
            </div>
          </section>

          {/* Skeleton historique */}
          <section>
            <h2 className="font-display text-xl font-semibold text-ink mb-md">
              Historique des quiz
            </h2>
            <div className="space-y-sm">
              <CardSkeleton />
              <CardSkeleton />
              <CardSkeleton />
            </div>
          </section>
        </>
      ) : historique.length === 0 ? (
        /* État vide */
        <section className="bg-white rounded-lg p-xl text-center">
          <div className="text-5xl mb-md" role="img" aria-label="Graphique">
            📊
          </div>
          <h2 className="font-display text-xl font-semibold text-ink mb-sm">
            Aucun quiz complété
          </h2>
          <p className="text-ink-light mb-lg max-w-md mx-auto">
            Complete des quiz pour voir ta progression ici. Tes scores et statistiques
            apparaîtront automatiquement.
          </p>
          <Link
            to="/quiz"
            className="inline-flex items-center gap-2 bg-coral text-white px-lg py-3 rounded-full font-semibold no-underline transition-all hover:bg-coral-dark hover:-translate-y-0.5"
          >
            <span role="img" aria-label="Quiz">❓</span>
            Passer un quiz
          </Link>
        </section>
      ) : (
        <>
          {/* Stats par matière */}
          {parMatiere.length > 0 && (
            <section className="mb-xl">
              <h2 className="font-display text-xl font-semibold text-ink mb-md">
                Statistiques par matière
              </h2>
              <div className="grid grid-cols-2 gap-md">
                {parMatiere.map((stats) => (
                  <MatiereCard key={stats.matiere} stats={stats} />
                ))}
              </div>
            </section>
          )}

          {/* Historique */}
          <section>
            <h2 className="font-display text-xl font-semibold text-ink mb-md">
              Historique des quiz
            </h2>
            <div className="space-y-sm">
              {historique.map((entry) => (
                <HistoriqueCard key={entry.sessionId} entry={entry} />
              ))}
            </div>
          </section>
        </>
      )}
    </>
  )
}
