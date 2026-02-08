interface ControlesFichesProps {
  indexActuel: number
  total: number
  onPrecedent: () => void
  onSuivant: () => void
  onMarquerCorrect?: () => void
  onMarquerIncorrect?: () => void
}

export default function ControlesFiches({
  indexActuel,
  total,
  onPrecedent,
  onSuivant,
  onMarquerCorrect,
  onMarquerIncorrect,
}: ControlesFichesProps) {
  const estPremier = indexActuel === 0
  const estDernier = indexActuel === total - 1

  return (
    <div className="flex items-center gap-6">
      {/* Bouton précédent */}
      <button
        onClick={onPrecedent}
        disabled={estPremier}
        className="w-14 h-14 rounded-full bg-white shadow-md flex items-center justify-center text-2xl text-ink transition-all hover:scale-110 hover:shadow-lg disabled:opacity-30 disabled:cursor-not-allowed disabled:hover:scale-100"
        aria-label="Fiche précédente"
      >
        ←
      </button>

      {/* Bouton incorrect (optionnel) */}
      {onMarquerIncorrect && (
        <button
          onClick={onMarquerIncorrect}
          className="w-14 h-14 rounded-full bg-red-50 text-red-700 flex items-center justify-center text-2xl transition-all hover:scale-110 hover:bg-red-100"
          aria-label="À revoir"
          title="À revoir"
        >
          ✗
        </button>
      )}

      {/* Compteur */}
      <span className="min-w-[80px] text-center text-ink-muted">
        {indexActuel + 1} / {total}
      </span>

      {/* Bouton correct (optionnel) */}
      {onMarquerCorrect && (
        <button
          onClick={onMarquerCorrect}
          className="w-14 h-14 rounded-full bg-green-50 text-green-700 flex items-center justify-center text-2xl transition-all hover:scale-110 hover:bg-green-100"
          aria-label="Maîtrisée"
          title="Maîtrisée"
        >
          ✓
        </button>
      )}

      {/* Bouton suivant */}
      <button
        onClick={onSuivant}
        disabled={estDernier}
        className="w-14 h-14 rounded-full bg-white shadow-md flex items-center justify-center text-2xl text-ink transition-all hover:scale-110 hover:shadow-lg disabled:opacity-30 disabled:cursor-not-allowed disabled:hover:scale-100"
        aria-label="Fiche suivante"
      >
        →
      </button>
    </div>
  )
}
