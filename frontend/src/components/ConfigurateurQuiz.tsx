interface ConfigurateurQuizProps {
  nombreQuestions: number
  difficulte: 'facile' | 'moyen' | 'difficile'
  onNombreChange: (nombre: number) => void
  onDifficulteChange: (difficulte: 'facile' | 'moyen' | 'difficile') => void
  onLancer: () => void
  chargement?: boolean
}

const optionsNombre = [5, 10, 15, 20]
const optionsDifficulte: { valeur: 'facile' | 'moyen' | 'difficile'; label: string }[] = [
  { valeur: 'facile', label: 'Facile' },
  { valeur: 'moyen', label: 'Moyen' },
  { valeur: 'difficile', label: 'Difficile' },
]

export default function ConfigurateurQuiz({
  nombreQuestions,
  difficulte,
  onNombreChange,
  onDifficulteChange,
  onLancer,
  chargement = false,
}: ConfigurateurQuizProps) {
  return (
    <div className="bg-white rounded-lg p-lg shadow-sm max-w-md mx-auto">
      <h2 className="font-display text-xl font-semibold text-ink mb-lg text-center">
        Configurer le quiz
      </h2>

      {/* Nombre de questions */}
      <div className="mb-lg">
        <label className="block text-sm font-medium text-ink mb-sm">
          Nombre de questions
        </label>
        <div className="flex gap-sm">
          {optionsNombre.map((nombre) => (
            <button
              key={nombre}
              onClick={() => onNombreChange(nombre)}
              disabled={chargement}
              className={`flex-1 py-3 rounded-lg font-medium text-sm transition-all ${
                nombreQuestions === nombre
                  ? 'bg-coral text-white'
                  : 'bg-cream text-ink hover:bg-cream-dark'
              } ${chargement ? 'opacity-50 cursor-not-allowed' : ''}`}
            >
              {nombre}
            </button>
          ))}
        </div>
      </div>

      {/* Difficulte */}
      <div className="mb-xl">
        <label className="block text-sm font-medium text-ink mb-sm">
          Difficulte
        </label>
        <div className="flex gap-sm">
          {optionsDifficulte.map((option) => (
            <button
              key={option.valeur}
              onClick={() => onDifficulteChange(option.valeur)}
              disabled={chargement}
              className={`flex-1 py-3 rounded-lg font-medium text-sm transition-all ${
                difficulte === option.valeur
                  ? option.valeur === 'facile'
                    ? 'bg-green-500 text-white'
                    : option.valeur === 'moyen'
                    ? 'bg-gold text-ink'
                    : 'bg-red-500 text-white'
                  : 'bg-cream text-ink hover:bg-cream-dark'
              } ${chargement ? 'opacity-50 cursor-not-allowed' : ''}`}
            >
              {option.label}
            </button>
          ))}
        </div>
      </div>

      {/* Bouton lancer */}
      <button
        onClick={onLancer}
        disabled={chargement}
        className={`w-full py-4 rounded-full font-semibold text-white transition-all ${
          chargement
            ? 'bg-coral/50 cursor-not-allowed'
            : 'bg-coral hover:bg-coral-dark'
        }`}
      >
        {chargement ? (
          <span className="flex items-center justify-center gap-sm">
            <svg
              className="animate-spin h-5 w-5"
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                className="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                strokeWidth="4"
              />
              <path
                className="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              />
            </svg>
            Generation en cours...
          </span>
        ) : (
          'Lancer le quiz'
        )}
      </button>
    </div>
  )
}
