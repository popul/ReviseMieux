interface ProgressionQuizProps {
  questionActuelle: number
  totalQuestions: number
  bonnesReponses: number
}

export default function ProgressionQuiz({
  questionActuelle,
  totalQuestions,
  bonnesReponses,
}: ProgressionQuizProps) {
  const progression = ((questionActuelle) / totalQuestions) * 100

  return (
    <div className="mb-8">
      {/* Header avec compteurs */}
      <div className="flex items-center justify-between mb-4">
        <span className="text-sm text-ink-muted">
          Question {questionActuelle + 1} sur {totalQuestions}
        </span>
        <span className="text-sm font-medium text-teal">
          {bonnesReponses} bonne{bonnesReponses > 1 ? 's' : ''} reponse{bonnesReponses > 1 ? 's' : ''}
        </span>
      </div>

      {/* Barre de progression */}
      <div className="h-2 bg-cream-dark rounded-full overflow-hidden">
        <div
          className="h-full bg-coral rounded-full transition-all duration-300"
          style={{ width: `${progression}%` }}
        />
      </div>

      {/* Indicateurs de questions */}
      <div className="flex gap-1 mt-4 justify-center">
        {Array.from({ length: totalQuestions }, (_, i) => (
          <div
            key={i}
            className={`w-2 h-2 rounded-full transition-colors ${
              i < questionActuelle
                ? 'bg-teal'
                : i === questionActuelle
                ? 'bg-coral'
                : 'bg-cream-dark'
            }`}
          />
        ))}
      </div>
    </div>
  )
}
