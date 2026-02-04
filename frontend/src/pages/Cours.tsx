import { useEffect, useState, useCallback } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { listerCours, obtenirCours, supprimerCours, type Cours } from '../services/api'
import ProcessingSection from '../components/ProcessingSection'

// Icônes des matières
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

// Format de date
function formatDate(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleDateString('fr-FR', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  })
}

// Skeleton pour les cours
function CoursSkeleton() {
  return (
    <div className="bg-white rounded-lg p-lg animate-pulse">
      <div className="flex items-start gap-md">
        <div className="w-14 h-14 rounded-md bg-cream" />
        <div className="flex-1">
          <div className="h-6 bg-cream rounded w-3/4 mb-2" />
          <div className="h-4 bg-cream rounded w-1/2 mb-3" />
          <div className="flex gap-md">
            <div className="h-4 bg-cream rounded w-20" />
            <div className="h-4 bg-cream rounded w-20" />
          </div>
        </div>
      </div>
    </div>
  )
}

// Carte de cours dans la liste
function CoursCard({
  cours,
  onSupprimer,
}: {
  cours: Cours
  onSupprimer: (id: string) => void
}) {
  const [confirmationSuppression, setConfirmationSuppression] = useState(false)

  const handleSupprimer = () => {
    if (confirmationSuppression) {
      onSupprimer(cours.id)
      setConfirmationSuppression(false)
    } else {
      setConfirmationSuppression(true)
    }
  }

  return (
    <div className="bg-white rounded-lg p-lg shadow-sm hover:shadow-md transition-shadow border border-cream-dark">
      <div className="flex items-start gap-md">
        <div className="w-14 h-14 rounded-md bg-cream flex items-center justify-center text-2xl flex-shrink-0">
          {getIconeMatiere(cours.matiere)}
        </div>
        <div className="flex-1 min-w-0">
          <h2 className="font-display text-lg font-semibold text-ink mb-xs truncate">
            {cours.titre || 'Cours sans titre'}
          </h2>
          <p className="text-sm text-ink-muted mb-sm capitalize">
            {cours.matiere || 'Matière non définie'}
          </p>
          <div className="flex items-center gap-lg text-sm text-ink-light">
            <span>Créé le {formatDate(cours.dateCreation)}</span>
            {cours.confianceOcr > 0 && (
              <span className="flex items-center gap-1">
                OCR: {Math.round(cours.confianceOcr * 100)}%
              </span>
            )}
          </div>
        </div>
        <div className="flex flex-col gap-xs">
          <Link
            to={`/fiches?cours=${cours.id}`}
            className="px-md py-2 bg-coral text-white rounded-full text-sm font-medium hover:bg-coral-dark transition-colors text-center"
          >
            Fiches
          </Link>
          <Link
            to={`/quiz?cours=${cours.id}`}
            className="px-md py-2 bg-teal text-white rounded-full text-sm font-medium hover:bg-teal-light transition-colors text-center"
          >
            Quiz
          </Link>
          <Link
            to={`/mindmap?cours=${cours.id}`}
            className="px-md py-2 bg-white text-ink border border-ink rounded-full text-sm font-medium hover:bg-cream transition-colors text-center"
          >
            Mindmap
          </Link>
        </div>
      </div>

      {/* Aperçu du texte OCR */}
      {cours.texteOcr && (
        <div className="mt-md pt-md border-t border-cream-dark">
          <p className="text-sm text-ink-light line-clamp-2">
            {cours.texteOcr.slice(0, 200)}
            {cours.texteOcr.length > 200 ? '...' : ''}
          </p>
        </div>
      )}

      {/* Actions */}
      <div className="mt-md pt-md border-t border-cream-dark flex justify-end">
        {confirmationSuppression ? (
          <div className="flex items-center gap-sm">
            <span className="text-sm text-ink-muted">Confirmer la suppression ?</span>
            <button
              onClick={handleSupprimer}
              className="px-md py-1.5 bg-red-500 text-white rounded-full text-sm font-medium hover:bg-red-600 transition-colors"
            >
              Supprimer
            </button>
            <button
              onClick={() => setConfirmationSuppression(false)}
              className="px-md py-1.5 bg-cream text-ink rounded-full text-sm font-medium hover:bg-cream-dark transition-colors"
            >
              Annuler
            </button>
          </div>
        ) : (
          <button
            onClick={handleSupprimer}
            className="text-sm text-ink-muted hover:text-red-500 transition-colors"
          >
            Supprimer
          </button>
        )}
      </div>
    </div>
  )
}

// Vue détaillée d'un cours
function CoursDetail({ coursId, onRetour }: { coursId: string; onRetour: () => void }) {
  const [cours, setCours] = useState<Cours | null>(null)
  const [chargement, setChargement] = useState(true)
  const [erreur, setErreur] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    obtenirCours(coursId)
      .then((c) => {
        if (!cancelled) {
          setCours(c)
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

  if (chargement) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <ProcessingSection message="Chargement du cours..." />
      </div>
    )
  }

  if (erreur || !cours) {
    return (
      <div className="text-center py-xl">
        <div className="text-5xl mb-md">😕</div>
        <h1 className="font-display text-2xl font-semibold text-ink mb-sm">
          Cours introuvable
        </h1>
        <p className="text-ink-light mb-lg">{erreur || 'Ce cours n\'existe pas.'}</p>
        <button
          onClick={onRetour}
          className="inline-flex items-center gap-sm px-lg py-3 bg-coral text-white rounded-full font-medium hover:bg-coral-dark transition-colors"
        >
          ← Retour aux cours
        </button>
      </div>
    )
  }

  return (
    <div>
      {/* Header */}
      <div className="flex items-center justify-between mb-lg">
        <button
          onClick={onRetour}
          className="flex items-center gap-xs text-ink-light hover:text-ink transition-colors"
        >
          ← Retour aux cours
        </button>
        <div className="flex gap-sm">
          <Link
            to={`/fiches?cours=${cours.id}`}
            className="px-md py-2 bg-coral text-white rounded-full text-sm font-medium hover:bg-coral-dark transition-colors"
          >
            Réviser les fiches
          </Link>
          <Link
            to={`/quiz?cours=${cours.id}`}
            className="px-md py-2 bg-teal text-white rounded-full text-sm font-medium hover:bg-teal-light transition-colors"
          >
            Lancer un quiz
          </Link>
        </div>
      </div>

      {/* Contenu */}
      <div className="bg-white rounded-lg p-xl shadow-sm">
        <div className="flex items-start gap-lg mb-lg">
          <div className="w-20 h-20 rounded-lg bg-cream flex items-center justify-center text-4xl flex-shrink-0">
            {getIconeMatiere(cours.matiere)}
          </div>
          <div>
            <h1 className="font-display text-2xl font-bold text-ink mb-xs">
              {cours.titre || 'Cours sans titre'}
            </h1>
            <p className="text-ink-muted capitalize">{cours.matiere || 'Matière non définie'}</p>
            <p className="text-sm text-ink-light mt-sm">
              Créé le {formatDate(cours.dateCreation)}
            </p>
          </div>
        </div>

        {/* Indicateur OCR */}
        {cours.confianceOcr > 0 && (
          <div className="mb-lg p-md bg-cream rounded-md">
            <div className="flex items-center justify-between text-sm mb-xs">
              <span className="text-ink-muted">Confiance OCR</span>
              <span className="font-semibold text-ink">
                {Math.round(cours.confianceOcr * 100)}%
              </span>
            </div>
            <div className="h-2 bg-cream-dark rounded-full overflow-hidden">
              <div
                className={`h-full rounded-full transition-all ${
                  cours.confianceOcr >= 0.9
                    ? 'bg-success'
                    : cours.confianceOcr >= 0.7
                    ? 'bg-gold'
                    : 'bg-coral'
                }`}
                style={{ width: `${cours.confianceOcr * 100}%` }}
              />
            </div>
          </div>
        )}

        {/* Zones incertaines */}
        {cours.zonesIncertaines && cours.zonesIncertaines.length > 0 && (
          <div className="mb-lg p-md bg-gold/10 border border-gold rounded-md">
            <h3 className="font-semibold text-ink mb-sm">
              ⚠️ Zones incertaines ({cours.zonesIncertaines.length})
            </h3>
            <ul className="space-y-xs text-sm text-ink-light">
              {cours.zonesIncertaines.slice(0, 3).map((zone, i) => (
                <li key={i} className="flex items-start gap-2">
                  <span className="text-gold">•</span>
                  <span>"{zone.texte}" - {zone.raison}</span>
                </li>
              ))}
              {cours.zonesIncertaines.length > 3 && (
                <li className="text-ink-muted">
                  Et {cours.zonesIncertaines.length - 3} autre(s)...
                </li>
              )}
            </ul>
          </div>
        )}

        {/* Texte OCR */}
        <div>
          <h3 className="font-semibold text-ink mb-sm">Contenu du cours</h3>
          <div className="p-md bg-cream rounded-md max-h-[400px] overflow-y-auto">
            <pre className="whitespace-pre-wrap text-sm text-ink-light font-body">
              {cours.texteOcr || 'Aucun contenu textuel disponible.'}
            </pre>
          </div>
        </div>
      </div>
    </div>
  )
}

// Page principale des cours
export default function CoursPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const coursId = searchParams.get('id')

  const [cours, setCours] = useState<Cours[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [chargement, setChargement] = useState(true)
  const [erreur, setErreur] = useState<string | null>(null)

  const limite = 10

  const chargerCours = useCallback(async () => {
    try {
      setChargement(true)
      setErreur(null)
      const res = await listerCours(page, limite)
      setCours(res.cours)
      setTotal(res.total)
    } catch (err) {
      setErreur(err instanceof Error ? err.message : 'Erreur inconnue')
    } finally {
      setChargement(false)
    }
  }, [page])

  useEffect(() => {
    if (!coursId) {
      chargerCours()
    }
  }, [chargerCours, coursId])

  const handleSupprimer = async (id: string) => {
    try {
      await supprimerCours(id)
      setCours((prev) => prev.filter((c) => c.id !== id))
      setTotal((prev) => prev - 1)
    } catch (err) {
      console.error('Erreur lors de la suppression:', err)
    }
  }

  const handleRetour = () => {
    setSearchParams({})
  }

  // Vue détail
  if (coursId) {
    return <CoursDetail coursId={coursId} onRetour={handleRetour} />
  }

  // Vue liste
  const totalPages = Math.ceil(total / limite)

  return (
    <div>
      {/* Header */}
      <div className="flex items-center justify-between mb-lg">
        <div>
          <h1 className="font-display text-3xl font-bold text-ink mb-xs">
            Mes cours
          </h1>
          <p className="text-ink-light">
            {total} cours enregistré{total > 1 ? 's' : ''}
          </p>
        </div>
        <Link
          to="/scanner"
          className="inline-flex items-center gap-sm px-lg py-3 bg-coral text-white rounded-full font-semibold hover:bg-coral-dark transition-colors"
        >
          <span role="img" aria-label="Scanner">📸</span>
          Scanner un cours
        </Link>
      </div>

      {/* Erreur */}
      {erreur && (
        <div className="bg-red-50 text-red-600 rounded-lg p-md mb-lg">
          <p className="font-medium">Erreur de chargement</p>
          <p className="text-sm">{erreur}</p>
        </div>
      )}

      {/* Liste des cours */}
      {chargement ? (
        <div className="space-y-md">
          <CoursSkeleton />
          <CoursSkeleton />
          <CoursSkeleton />
        </div>
      ) : cours.length === 0 ? (
        <div className="bg-white rounded-lg p-xl text-center">
          <div className="text-5xl mb-md" role="img" aria-label="Livres">📚</div>
          <h2 className="font-display text-xl font-semibold text-ink mb-sm">
            Aucun cours enregistré
          </h2>
          <p className="text-ink-light mb-lg max-w-md mx-auto">
            Commencez par scanner un cours pour générer des fiches de révision et des quiz interactifs.
          </p>
          <Link
            to="/scanner"
            className="inline-flex items-center gap-2 bg-coral text-white px-lg py-3 rounded-full font-semibold no-underline transition-all hover:bg-coral-dark hover:-translate-y-0.5"
          >
            <span role="img" aria-label="Scanner">📸</span>
            Scanner mon premier cours
          </Link>
        </div>
      ) : (
        <>
          <div className="space-y-md">
            {cours.map((c) => (
              <CoursCard key={c.id} cours={c} onSupprimer={handleSupprimer} />
            ))}
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex items-center justify-center gap-sm mt-lg">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                className="px-md py-2 rounded-full text-sm font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed bg-white border border-cream-dark hover:bg-cream"
              >
                ← Précédent
              </button>
              <span className="px-md py-2 text-sm text-ink-muted">
                Page {page} sur {totalPages}
              </span>
              <button
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
                className="px-md py-2 rounded-full text-sm font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed bg-white border border-cream-dark hover:bg-cream"
              >
                Suivant →
              </button>
            </div>
          )}
        </>
      )}
    </div>
  )
}
