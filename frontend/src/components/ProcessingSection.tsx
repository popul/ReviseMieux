interface ProcessingSectionProps {
  message?: string
}

export default function ProcessingSection({ message = 'Traitement en cours...' }: ProcessingSectionProps) {
  return (
    <div className="bg-white rounded-lg p-xl text-center">
      {/* Spinner animé */}
      <div className="relative w-20 h-20 mx-auto mb-lg">
        {/* Cercle externe */}
        <div className="absolute inset-0 border-4 border-cream rounded-full" />
        {/* Arc animé */}
        <div className="absolute inset-0 border-4 border-transparent border-t-coral rounded-full animate-spin" />
        {/* Icône centrale */}
        <div className="absolute inset-0 flex items-center justify-center">
          <span className="text-2xl">📄</span>
        </div>
      </div>

      {/* Message */}
      <p className="text-lg font-medium text-ink mb-sm">{message}</p>
      <p className="text-ink-light text-sm">
        L'extraction du texte peut prendre quelques secondes...
      </p>

      {/* Barre de progression indéterminée */}
      <div className="mt-lg mx-auto max-w-xs h-1.5 bg-cream rounded-full overflow-hidden">
        <div className="h-full bg-gradient-to-r from-coral to-gold rounded-full animate-progress" />
      </div>
    </div>
  )
}
