import type { Question } from '../services/api'

interface QuestionQuizProps {
  question: Question
  choixSelectionne: number | null
  onSelectChoix: (index: number) => void
  onValider: () => void
  desactive?: boolean
}

export default function QuestionQuiz({
  question,
  choixSelectionne,
  onSelectChoix,
  onValider,
  desactive = false,
}: QuestionQuizProps) {
  const handleKeyDown = (e: React.KeyboardEvent, index: number) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault()
      if (!desactive) {
        onSelectChoix(index)
      }
    }
  }

  return (
    <div className="bg-white rounded-lg p-lg shadow-sm max-w-2xl mx-auto">
      {/* Enonce */}
      <h2 className="font-display text-xl font-semibold text-ink mb-lg leading-relaxed">
        {question.enonce}
      </h2>

      {/* Choix */}
      <div className="space-y-sm mb-lg">
        {question.choix.map((choix, index) => {
          const estSelectionne = choixSelectionne === index
          const lettre = String.fromCharCode(65 + index) // A, B, C, D

          return (
            <button
              key={index}
              onClick={() => !desactive && onSelectChoix(index)}
              onKeyDown={(e) => handleKeyDown(e, index)}
              disabled={desactive}
              className={`w-full flex items-start gap-md p-md rounded-lg border-2 text-left transition-all ${
                desactive
                  ? 'opacity-70 cursor-not-allowed'
                  : 'cursor-pointer hover:border-coral/50'
              } ${
                estSelectionne
                  ? 'border-coral bg-coral/5'
                  : 'border-cream-dark bg-white'
              }`}
              aria-pressed={estSelectionne}
            >
              {/* Lettre indicateur */}
              <span
                className={`flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center font-semibold text-sm ${
                  estSelectionne
                    ? 'bg-coral text-white'
                    : 'bg-cream text-ink'
                }`}
              >
                {lettre}
              </span>

              {/* Texte du choix */}
              <span className="text-ink leading-relaxed pt-1">{choix}</span>
            </button>
          )
        })}
      </div>

      {/* Bouton valider */}
      <button
        onClick={onValider}
        disabled={desactive || choixSelectionne === null}
        className={`w-full py-4 rounded-full font-semibold text-white transition-all ${
          desactive || choixSelectionne === null
            ? 'bg-gray-300 cursor-not-allowed'
            : 'bg-teal hover:bg-teal-dark'
        }`}
      >
        Valider ma reponse
      </button>
    </div>
  )
}
