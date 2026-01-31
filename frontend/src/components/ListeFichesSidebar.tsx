import type { Fiche } from '../services/api'

interface ListeFichesSidebarProps {
  fiches: Fiche[]
  indexActuel: number
  onSelectFiche: (index: number) => void
}

function BadgeDifficulte({ difficulte }: { difficulte: Fiche['difficulte'] }) {
  const styles = {
    facile: 'bg-green-100 text-green-700',
    moyen: 'bg-gold/20 text-amber-700',
    difficile: 'bg-red-100 text-red-700',
  }

  const libelles = {
    facile: 'Facile',
    moyen: 'Moyen',
    difficile: 'Difficile',
  }

  return (
    <span className={`text-xs font-medium px-2 py-0.5 rounded-full ${styles[difficulte]}`}>
      {libelles[difficulte]}
    </span>
  )
}

export default function ListeFichesSidebar({
  fiches,
  indexActuel,
  onSelectFiche,
}: ListeFichesSidebarProps) {
  return (
    <aside className="w-[300px] bg-white border-r border-cream-dark p-lg overflow-y-auto flex-shrink-0">
      <h2 className="font-display text-sm font-semibold text-ink-muted uppercase tracking-wider mb-md">
        Fiches ({fiches.length})
      </h2>

      <ul className="space-y-xs">
        {fiches.map((fiche, index) => (
          <li key={fiche.id}>
            <button
              onClick={() => onSelectFiche(index)}
              className={`w-full text-left p-sm rounded-md transition-all border-l-[3px] ${
                index === indexActuel
                  ? 'bg-coral/10 border-coral'
                  : 'border-transparent hover:bg-cream'
              }`}
            >
              <div className="flex items-start justify-between gap-sm mb-1">
                <span className="font-semibold text-sm text-ink line-clamp-1">
                  Fiche {index + 1}
                </span>
                <BadgeDifficulte difficulte={fiche.difficulte} />
              </div>
              <p className="text-sm text-ink-muted line-clamp-2">
                {fiche.question}
              </p>
            </button>
          </li>
        ))}
      </ul>
    </aside>
  )
}
