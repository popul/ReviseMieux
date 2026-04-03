import { useState, useEffect, useCallback } from 'react'
import {
  listerValidations,
  resoudreValidation,
  type TacheValidation,
} from '../services/api'

type EtatPage = 'chargement' | 'liste' | 'detail' | 'erreur'

function BadgeSource({ source }: { source: string }) {
  const couleurs: Record<string, string> = {
    ocr: 'bg-teal/10 text-teal',
    llm: 'bg-coral/10 text-coral',
    manual: 'bg-gold/20 text-ink',
  }
  const classe = couleurs[source] || 'bg-cream-dark text-ink-muted'

  return (
    <span className={`text-xs font-medium px-2 py-0.5 rounded-full ${classe}`}>
      {source.toUpperCase()}
    </span>
  )
}

function BarreConfiance({ priorite }: { priorite: number }) {
  // priorite: 1 = haute confiance (vert), 5 = basse confiance (rouge)
  const pourcent = Math.max(0, Math.min(100, ((5 - priorite) / 4) * 100))
  const couleur =
    pourcent >= 66 ? 'bg-[#2ECC71]' : pourcent >= 33 ? 'bg-[#F5A623]' : 'bg-[#E74C3C]'

  return (
    <div className="flex items-center gap-2">
      <div className="flex-1 h-1.5 bg-cream-dark rounded-full overflow-hidden">
        <div
          className={`h-full rounded-full transition-all ${couleur}`}
          style={{ width: `${pourcent}%` }}
        />
      </div>
      <span className="text-xs text-ink-muted whitespace-nowrap">P{priorite}</span>
    </div>
  )
}

function CarteTache({
  tache,
  onSelect,
}: {
  tache: TacheValidation
  onSelect: (t: TacheValidation) => void
}) {
  return (
    <button
      onClick={() => onSelect(tache)}
      className="w-full bg-white rounded-md p-5 shadow-sm hover:shadow-md transition-shadow border border-cream-dark text-left"
      data-testid="validation-task-item"
    >
      <div className="flex items-start justify-between gap-4 mb-3">
        <div className="flex-1 min-w-0">
          <h3 className="font-display text-base font-semibold text-ink truncate">
            {tache.item_term || tache.suggestion || 'Item sans titre'}
          </h3>
        </div>
        <BadgeSource source={tache.source} />
      </div>
      {tache.suggestion && tache.suggestion !== tache.item_term && (
        <p className="text-sm text-ink-light mb-3 line-clamp-2">{tache.suggestion}</p>
      )}
      <BarreConfiance priorite={tache.priority} />
    </button>
  )
}

function DetailTache({
  tache,
  onRetour,
  onAction,
  enCours,
}: {
  tache: TacheValidation
  onRetour: () => void
  onAction: (action: string) => void
  enCours: boolean
}) {
  const [correction, setCorrection] = useState('')
  const [modeCorrection, setModeCorrection] = useState(false)

  return (
    <div>
      <button
        onClick={onRetour}
        className="flex items-center gap-2 text-ink-light hover:text-ink transition-colors mb-6"
      >
        ← Retour a la liste
      </button>

      <div className="bg-white rounded-md p-6 shadow-sm border border-cream-dark mb-6">
        <div className="flex items-start justify-between gap-4 mb-4">
          <h2 className="font-display text-xl font-semibold text-ink">
            {tache.item_term || 'Item sans titre'}
          </h2>
          <BadgeSource source={tache.source} />
        </div>

        {tache.suggestion && (
          <div className="mb-4">
            <div className="text-xs font-semibold text-ink-muted uppercase tracking-wider mb-1">
              Suggestion
            </div>
            <p className="text-ink-light">{tache.suggestion}</p>
          </div>
        )}

        <div className="mb-4">
          <div className="text-xs font-semibold text-ink-muted uppercase tracking-wider mb-1">
            Confiance
          </div>
          <BarreConfiance priorite={tache.priority} />
        </div>

        {tache.crop_url && (
          <div className="mb-4">
            <div className="text-xs font-semibold text-ink-muted uppercase tracking-wider mb-2">
              Source OCR
            </div>
            <img
              src={tache.crop_url}
              alt="Extrait du document source"
              className="max-w-full rounded border border-cream-dark"
            />
          </div>
        )}

        <div className="text-xs text-ink-muted">
          Cree le {new Date(tache.created_at).toLocaleDateString('fr-FR', {
            day: 'numeric',
            month: 'long',
            year: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
          })}
        </div>
      </div>

      {modeCorrection && (
        <div className="bg-white rounded-md p-6 shadow-sm border border-cream-dark mb-6">
          <div className="text-xs font-semibold text-ink-muted uppercase tracking-wider mb-2">
            Correction
          </div>
          <textarea
            value={correction}
            onChange={(e) => setCorrection(e.target.value)}
            placeholder="Saisissez la correction..."
            className="w-full border border-cream-dark rounded-sm p-3 text-sm min-h-[100px] focus:outline-none focus:ring-2 focus:ring-coral/30 focus:border-coral"
            data-testid="validation-correction-input"
          />
          <div className="flex gap-3 mt-3">
            <button
              onClick={() => onAction('correct')}
              disabled={enCours || !correction.trim()}
              className="px-5 py-2 bg-[#F5A623] text-white rounded-full text-sm font-medium hover:bg-[#E09910] transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              data-testid="validation-submit-correction-btn"
            >
              {enCours ? 'En cours...' : 'Envoyer la correction'}
            </button>
            <button
              onClick={() => setModeCorrection(false)}
              className="px-5 py-2 border border-cream-dark text-ink-muted rounded-full text-sm font-medium hover:bg-cream transition-colors"
            >
              Annuler
            </button>
          </div>
        </div>
      )}

      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
        <button
          onClick={() => onAction('approve')}
          disabled={enCours}
          className="px-4 py-3 bg-[#2ECC71] text-white rounded-md text-sm font-medium hover:bg-[#27AE60] transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          data-testid="validation-confirm-btn"
        >
          Confirmer
        </button>
        <button
          onClick={() => setModeCorrection(true)}
          disabled={enCours}
          className="px-4 py-3 bg-[#F5A623] text-white rounded-md text-sm font-medium hover:bg-[#E09910] transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          data-testid="validation-correct-btn"
        >
          Corriger
        </button>
        <button
          onClick={() => onAction('reject')}
          disabled={enCours}
          className="px-4 py-3 bg-ink-muted/20 text-ink-light rounded-md text-sm font-medium hover:bg-ink-muted/30 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          data-testid="validation-ignore-btn"
        >
          Ignorer
        </button>
        <button
          onClick={() => onAction('skip')}
          disabled={enCours}
          className="px-4 py-3 bg-cream-dark text-ink-muted rounded-md text-sm font-medium hover:bg-cream transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          data-testid="validation-skip-btn"
        >
          NSP
        </button>
      </div>
    </div>
  )
}

export default function Validation() {
  const [etat, setEtat] = useState<EtatPage>('chargement')
  const [taches, setTaches] = useState<TacheValidation[]>([])
  const [tacheSelectionnee, setTacheSelectionnee] = useState<TacheValidation | null>(null)
  const [erreur, setErreur] = useState<string | null>(null)
  const [enCours, setEnCours] = useState(false)
  const [toast, setToast] = useState<{ message: string; type: 'succes' | 'erreur' } | null>(null)

  const chargerTaches = useCallback(async () => {
    setEtat('chargement')
    setErreur(null)
    try {
      const res = await listerValidations()
      const pending = (res.validations || []).filter((t) => t.status === 'pending')
      setTaches(pending)
      setEtat('liste')
    } catch (err) {
      setErreur(err instanceof Error ? err.message : 'Impossible de charger les validations')
      setEtat('erreur')
    }
  }, [])

  useEffect(() => {
    chargerTaches()
  }, [chargerTaches])

  const afficherToast = useCallback((message: string, type: 'succes' | 'erreur') => {
    setToast({ message, type })
    setTimeout(() => setToast(null), 3000)
  }, [])

  const gererAction = useCallback(
    async (action: string) => {
      if (!tacheSelectionnee) return
      setEnCours(true)
      try {
        const correction =
          action === 'correct'
            ? { suggestion: (document.querySelector('[data-testid="validation-correction-input"]') as HTMLTextAreaElement)?.value }
            : undefined
        await resoudreValidation(tacheSelectionnee.id, action, correction)
        setTaches((prev) => prev.filter((t) => t.id !== tacheSelectionnee.id))
        setTacheSelectionnee(null)
        setEtat('liste')

        const labels: Record<string, string> = {
          approve: 'confirmee',
          correct: 'corrigee',
          reject: 'ignoree',
          skip: 'passee',
        }
        afficherToast(`Validation ${labels[action] || 'traitee'} avec succes`, 'succes')
      } catch (err) {
        afficherToast(
          err instanceof Error ? err.message : 'Erreur lors de la resolution',
          'erreur'
        )
      } finally {
        setEnCours(false)
      }
    },
    [tacheSelectionnee, afficherToast]
  )

  const selectionnerTache = useCallback((tache: TacheValidation) => {
    setTacheSelectionnee(tache)
    setEtat('detail')
  }, [])

  const retourListe = useCallback(() => {
    setTacheSelectionnee(null)
    setEtat('liste')
  }, [])

  // Toast notification
  const toastElement = toast && (
    <div
      className={`fixed bottom-6 right-6 z-50 px-6 py-3 rounded-md shadow-lg text-white text-sm font-medium transition-all ${
        toast.type === 'succes' ? 'bg-[#2ECC71]' : 'bg-[#E74C3C]'
      }`}
      role="alert"
      data-testid="validation-toast"
    >
      {toast.message}
    </div>
  )

  // Chargement
  if (etat === 'chargement') {
    return (
      <div className="flex items-center justify-center min-h-[400px]" data-testid="validation-loading">
        <div className="text-center">
          <div className="w-8 h-8 border-3 border-coral border-t-transparent rounded-full animate-spin mx-auto mb-4" />
          <p className="text-ink-muted">Chargement des validations...</p>
        </div>
      </div>
    )
  }

  // Erreur
  if (etat === 'erreur') {
    return (
      <div className="text-center py-12" data-testid="validation-error">
        <div className="text-5xl mb-6">{String.fromCodePoint(0x1f615)}</div>
        <h1 className="font-display text-2xl font-semibold text-ink mb-4">
          Erreur de chargement
        </h1>
        <p className="text-ink-light mb-8">{erreur}</p>
        <button
          onClick={chargerTaches}
          className="px-8 py-3 bg-coral text-white rounded-full font-medium hover:bg-coral-dark transition-colors"
        >
          Reessayer
        </button>
        {toastElement}
      </div>
    )
  }

  // Detail
  if (etat === 'detail' && tacheSelectionnee) {
    return (
      <div>
        <div className="flex items-center justify-between mb-8">
          <h1 className="font-display text-xl md:text-3xl font-semibold text-ink">
            Validation
          </h1>
          <span className="text-sm text-ink-muted">
            {taches.length} en attente
          </span>
        </div>
        <DetailTache
          tache={tacheSelectionnee}
          onRetour={retourListe}
          onAction={gererAction}
          enCours={enCours}
        />
        {toastElement}
      </div>
    )
  }

  // Liste (vide)
  if (taches.length === 0) {
    return (
      <div className="text-center py-12" data-testid="validation-empty">
        <div className="text-5xl mb-6">{String.fromCodePoint(0x2705)}</div>
        <h1 className="font-display text-2xl font-semibold text-ink mb-4">
          Aucune validation en attente
        </h1>
        <p className="text-ink-light">
          Toutes les suggestions ont ete traitees. Revenez plus tard.
        </p>
        {toastElement}
      </div>
    )
  }

  // Liste
  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <h1 className="font-display text-xl md:text-3xl font-semibold text-ink">
          Validation
        </h1>
        <span className="bg-coral/10 text-coral text-sm font-medium px-3 py-1 rounded-full">
          {taches.length} en attente
        </span>
      </div>

      <div className="grid gap-4" data-testid="validation-task-list">
        {taches.map((tache) => (
          <CarteTache key={tache.id} tache={tache} onSelect={selectionnerTache} />
        ))}
      </div>
      {toastElement}
    </div>
  )
}
