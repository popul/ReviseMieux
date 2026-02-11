import { type Concept } from '../services/api'

interface ConceptCardProps {
  concept: Concept
  onClick?: (concept: Concept) => void
  isSelected?: boolean
}

const couleursImportance: Record<string, { bg: string; border: string; text: string; badge: string }> = {
  essentiel: { bg: 'bg-red-50', border: 'border-[#E85D4C]', text: 'text-[#E85D4C]', badge: 'bg-[#E85D4C] text-white' },
  important: { bg: 'bg-teal-50', border: 'border-[#1A4D4D]', text: 'text-[#1A4D4D]', badge: 'bg-[#1A4D4D] text-white' },
  secondaire: { bg: 'bg-amber-50', border: 'border-[#F5C542]', text: 'text-[#8B7000]', badge: 'bg-[#F5C542] text-[#1A4D4D]' },
}

export default function ConceptCard({ concept, onClick, isSelected }: ConceptCardProps) {
  const couleurs = couleursImportance[concept.importance] || couleursImportance.secondaire

  return (
    <button
      type="button"
      onClick={() => onClick?.(concept)}
      className={`w-full text-left p-3 rounded-lg border-2 transition-all ${couleurs.bg} ${
        isSelected ? `${couleurs.border} shadow-md ring-2 ring-offset-1 ring-${couleurs.border}` : 'border-transparent hover:border-gray-200'
      } ${onClick ? 'cursor-pointer' : 'cursor-default'}`}
    >
      <div className="flex items-start justify-between gap-2">
        <h4 className={`font-semibold text-sm ${couleurs.text}`}>{concept.nom}</h4>
        <span className={`text-xs px-2 py-0.5 rounded-full font-medium whitespace-nowrap ${couleurs.badge}`}>
          {concept.importance}
        </span>
      </div>
      <p className="text-xs text-gray-600 mt-1 line-clamp-2">{concept.definition}</p>
    </button>
  )
}
