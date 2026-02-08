interface Etape {
  numero: number
  libelle: string
}

interface IndicateurEtapesProps {
  etapes: Etape[]
  etapeActive: number
}

export default function IndicateurEtapes({ etapes, etapeActive }: IndicateurEtapesProps) {
  return (
    <div className="flex items-center gap-4">
      {etapes.map((etape, index) => {
        const estCompletee = etape.numero < etapeActive
        const estActive = etape.numero === etapeActive

        return (
          <div key={etape.numero} className="flex items-center">
            {/* Dot avec numéro */}
            <div className="flex items-center gap-2">
              <div
                className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-semibold transition-all ${
                  estCompletee
                    ? 'bg-success text-white'
                    : estActive
                    ? 'bg-coral text-white'
                    : 'bg-cream text-ink-muted'
                }`}
              >
                {estCompletee ? '✓' : etape.numero}
              </div>
              <span
                className={`text-sm transition-colors ${
                  estActive ? 'text-ink font-medium' : 'text-ink-muted hidden'
                }`}
              >
                {etape.libelle}
              </span>
            </div>

            {/* Ligne de connexion */}
            {index < etapes.length - 1 && (
              <div
                className={`w-10 h-0.5 mx-4 transition-colors ${
                  estCompletee ? 'bg-success' : 'bg-cream-dark'
                }`}
              />
            )}
          </div>
        )
      })}
    </div>
  )
}
