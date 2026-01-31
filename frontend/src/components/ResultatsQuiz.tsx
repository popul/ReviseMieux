import { Link } from 'react-router-dom'

interface ResultatsQuizProps {
  score: number
  totalQuestions: number
  bonnesReponses: number
  coursId: string
  onRefaire: () => void
  onNouveauQuiz: () => void
}

export default function ResultatsQuiz({
  score,
  totalQuestions,
  bonnesReponses,
  coursId,
  onRefaire,
  onNouveauQuiz,
}: ResultatsQuizProps) {
  // Determiner le message et la couleur selon le score
  let message: string
  let emoji: string
  let couleur: string

  if (score >= 80) {
    message = 'Excellent travail !'
    emoji = '1f389' // Tada
    couleur = 'text-green-600'
  } else if (score >= 60) {
    message = 'Bien joue !'
    emoji = '1f44d' // Thumbs up
    couleur = 'text-teal'
  } else if (score >= 40) {
    message = 'Pas mal, continue !'
    emoji = '1f4aa' // Muscle
    couleur = 'text-gold'
  } else {
    message = 'Continue a reviser !'
    emoji = '1f4d6' // Book
    couleur = 'text-coral'
  }

  return (
    <div className="bg-white rounded-lg p-xl shadow-sm max-w-md mx-auto text-center">
      {/* Emoji et message */}
      <div className="mb-lg">
        <span className="text-6xl" role="img" aria-label="resultat">
          {String.fromCodePoint(parseInt(emoji, 16))}
        </span>
        <h2 className={`font-display text-2xl font-bold mt-md ${couleur}`}>
          {message}
        </h2>
      </div>

      {/* Score circular */}
      <div className="relative w-40 h-40 mx-auto mb-lg">
        <svg className="w-full h-full transform -rotate-90">
          {/* Cercle de fond */}
          <circle
            cx="80"
            cy="80"
            r="70"
            fill="none"
            stroke="#F5F3EE"
            strokeWidth="12"
          />
          {/* Cercle de progression */}
          <circle
            cx="80"
            cy="80"
            r="70"
            fill="none"
            stroke={score >= 60 ? '#1A4D4D' : '#E85D4C'}
            strokeWidth="12"
            strokeLinecap="round"
            strokeDasharray={`${(score / 100) * 440} 440`}
            className="transition-all duration-1000"
          />
        </svg>
        {/* Score au centre */}
        <div className="absolute inset-0 flex flex-col items-center justify-center">
          <span className="font-display text-4xl font-bold text-ink">
            {Math.round(score)}%
          </span>
          <span className="text-sm text-ink-muted">Score</span>
        </div>
      </div>

      {/* Details */}
      <div className="bg-cream rounded-lg p-md mb-lg">
        <div className="flex justify-center gap-xl">
          <div>
            <p className="text-2xl font-bold text-teal">{bonnesReponses}</p>
            <p className="text-sm text-ink-muted">Bonnes reponses</p>
          </div>
          <div className="w-px bg-cream-dark" />
          <div>
            <p className="text-2xl font-bold text-ink">{totalQuestions}</p>
            <p className="text-sm text-ink-muted">Questions</p>
          </div>
        </div>
      </div>

      {/* Actions */}
      <div className="space-y-sm">
        <button
          onClick={onRefaire}
          className="w-full py-3 rounded-full font-semibold text-white bg-coral hover:bg-coral-dark transition-colors"
        >
          Refaire ce quiz
        </button>
        <button
          onClick={onNouveauQuiz}
          className="w-full py-3 rounded-full font-semibold text-coral border-2 border-coral hover:bg-coral/5 transition-colors"
        >
          Nouveau quiz
        </button>
        <Link
          to={`/fiches?cours=${coursId}`}
          className="block w-full py-3 rounded-full font-semibold text-ink-light border-2 border-cream-dark hover:bg-cream transition-colors"
        >
          Revoir les fiches
        </Link>
      </div>
    </div>
  )
}
