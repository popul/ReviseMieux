import type { Question } from '../services/api'

interface FeedbackReponseProps {
  question: Question
  choixSelectionne: number
  estCorrecte: boolean
  explication: string
  onSuivant: () => void
  estDerniere?: boolean
}

export default function FeedbackReponse({
  question,
  choixSelectionne,
  estCorrecte,
  explication,
  onSuivant,
  estDerniere = false,
}: FeedbackReponseProps) {
  return (
    <div className="bg-white rounded-lg p-lg shadow-sm max-w-2xl mx-auto">
      {/* Resultat */}
      <div
        className={`flex items-center gap-md p-md rounded-lg mb-lg ${
          estCorrecte ? 'bg-green-50' : 'bg-red-50'
        }`}
      >
        <div
          className={`flex-shrink-0 w-12 h-12 rounded-full flex items-center justify-center ${
            estCorrecte ? 'bg-green-500' : 'bg-red-500'
          }`}
        >
          {estCorrecte ? (
            <svg
              className="w-6 h-6 text-white"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M5 13l4 4L19 7"
              />
            </svg>
          ) : (
            <svg
              className="w-6 h-6 text-white"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          )}
        </div>
        <div>
          <p
            className={`font-semibold text-lg ${
              estCorrecte ? 'text-green-700' : 'text-red-700'
            }`}
          >
            {estCorrecte ? 'Bonne reponse !' : 'Mauvaise reponse'}
          </p>
          {!estCorrecte && (
            <p className="text-red-600 text-sm">
              La bonne reponse etait : {question.choix[question.reponseCorrecte]}
            </p>
          )}
        </div>
      </div>

      {/* Question et choix avec highlight */}
      <div className="mb-lg">
        <h3 className="font-display text-lg font-semibold text-ink mb-md">
          {question.enonce}
        </h3>
        <div className="space-y-xs">
          {question.choix.map((choix, index) => {
            const estBonneReponse = index === question.reponseCorrecte
            const estChoixUtilisateur = index === choixSelectionne
            const lettre = String.fromCharCode(65 + index)

            let bgClass = 'bg-cream'
            let borderClass = 'border-transparent'
            let textClass = 'text-ink-light'

            if (estBonneReponse) {
              bgClass = 'bg-green-100'
              borderClass = 'border-green-500'
              textClass = 'text-green-700'
            } else if (estChoixUtilisateur && !estCorrecte) {
              bgClass = 'bg-red-100'
              borderClass = 'border-red-500'
              textClass = 'text-red-700'
            }

            return (
              <div
                key={index}
                className={`flex items-start gap-sm p-sm rounded-lg border-2 ${bgClass} ${borderClass}`}
              >
                <span
                  className={`flex-shrink-0 w-6 h-6 rounded-full flex items-center justify-center text-xs font-semibold ${
                    estBonneReponse
                      ? 'bg-green-500 text-white'
                      : estChoixUtilisateur && !estCorrecte
                      ? 'bg-red-500 text-white'
                      : 'bg-cream-dark text-ink-muted'
                  }`}
                >
                  {lettre}
                </span>
                <span className={`text-sm ${textClass}`}>{choix}</span>
              </div>
            )
          })}
        </div>
      </div>

      {/* Explication */}
      {explication && (
        <div className="bg-gold/10 border-l-4 border-gold p-md rounded-r-lg mb-lg">
          <p className="text-sm font-medium text-amber-800 mb-xs">Explication</p>
          <p className="text-sm text-ink leading-relaxed">{explication}</p>
        </div>
      )}

      {/* Bouton suivant */}
      <button
        onClick={onSuivant}
        className="w-full py-4 rounded-full font-semibold text-white bg-coral hover:bg-coral-dark transition-all"
      >
        {estDerniere ? 'Voir les resultats' : 'Question suivante'}
      </button>
    </div>
  )
}
