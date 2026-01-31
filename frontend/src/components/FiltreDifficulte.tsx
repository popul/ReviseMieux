type Difficulte = 'toutes' | 'facile' | 'moyen' | 'difficile'

interface FiltreDifficulteProps {
  valeur: Difficulte
  onChange: (difficulte: Difficulte) => void
}

const OPTIONS: { valeur: Difficulte; libelle: string }[] = [
  { valeur: 'toutes', libelle: 'Toutes' },
  { valeur: 'facile', libelle: 'Facile' },
  { valeur: 'moyen', libelle: 'Moyen' },
  { valeur: 'difficile', libelle: 'Difficile' },
]

export default function FiltreDifficulte({ valeur, onChange }: FiltreDifficulteProps) {
  return (
    <div className="flex gap-xs bg-white p-1 rounded-full shadow-sm">
      {OPTIONS.map((option) => (
        <button
          key={option.valeur}
          onClick={() => onChange(option.valeur)}
          className={`px-md py-2.5 rounded-full text-sm font-medium transition-all ${
            valeur === option.valeur
              ? 'bg-ink text-white'
              : 'text-ink-light hover:bg-cream'
          }`}
        >
          {option.libelle}
        </button>
      ))}
    </div>
  )
}

export type { Difficulte }
