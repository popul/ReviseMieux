interface ProcessingSectionProps {
  message?: string
  progression?: { page: number; total: number } | null
}

export default function ProcessingSection({
  message = 'Traitement en cours...',
  progression,
}: ProcessingSectionProps) {
  const pourcentage = progression
    ? Math.round((progression.page / progression.total) * 100)
    : null

  return (
    <div className="bg-white rounded-lg p-12 text-center">
      {/* Spinner animé */}
      <div className="relative w-20 h-20 mx-auto mb-8">
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
      <p className="text-lg font-medium text-ink mb-4">
        {progression
          ? `Page ${progression.page} sur ${progression.total} traitée`
          : message}
      </p>
      <p className="text-ink-light text-sm">
        {progression
          ? `${pourcentage}% terminé`
          : "L'extraction du texte peut prendre quelques secondes..."}
      </p>

      {/* Barre de progression */}
      <div className="mt-8 mx-auto max-w-xs h-1.5 bg-cream rounded-full overflow-hidden">
        {pourcentage !== null ? (
          <div
            className="h-full bg-gradient-to-r from-coral to-gold rounded-full transition-all duration-500 ease-out"
            style={{ width: `${pourcentage}%` }}
          />
        ) : (
          <div className="h-full bg-gradient-to-r from-coral to-gold rounded-full animate-progress" />
        )}
      </div>
    </div>
  )
}
