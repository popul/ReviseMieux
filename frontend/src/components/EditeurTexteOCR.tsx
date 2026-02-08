import { useState, useMemo, useCallback } from 'react'
import type { ZoneIncertaine } from '../services/api'

interface EditeurTexteOCRProps {
  texte: string
  zonesIncertaines: ZoneIncertaine[]
  confiance: number
  onTexteChange: (nouveauTexte: string) => void
  onValider: () => void
  enChargement?: boolean
}

interface SegmentTexte {
  texte: string
  estIncertain: boolean
  zone?: ZoneIncertaine
}

export default function EditeurTexteOCR({
  texte,
  zonesIncertaines,
  confiance,
  onTexteChange,
  onValider,
  enChargement = false,
}: EditeurTexteOCRProps) {
  const [modeEdition, setModeEdition] = useState(false)
  const [texteEdite, setTexteEdite] = useState(texte)

  // Segmenter le texte avec les zones incertaines
  const segments = useMemo((): SegmentTexte[] => {
    if (zonesIncertaines.length === 0) {
      return [{ texte, estIncertain: false }]
    }

    const result: SegmentTexte[] = []
    let curseur = 0

    // Trier les zones par position de début
    const zonesSorted = [...zonesIncertaines].sort((a, b) => a.debut - b.debut)

    for (const zone of zonesSorted) {
      // Ajouter le texte avant la zone
      if (zone.debut > curseur) {
        result.push({
          texte: texte.slice(curseur, zone.debut),
          estIncertain: false,
        })
      }

      // Ajouter la zone incertaine
      result.push({
        texte: texte.slice(zone.debut, zone.fin),
        estIncertain: true,
        zone,
      })

      curseur = zone.fin
    }

    // Ajouter le reste du texte
    if (curseur < texte.length) {
      result.push({
        texte: texte.slice(curseur),
        estIncertain: false,
      })
    }

    return result
  }, [texte, zonesIncertaines])

  const handleSauvegarder = useCallback(() => {
    onTexteChange(texteEdite)
    setModeEdition(false)
  }, [texteEdite, onTexteChange])

  const handleAnnuler = useCallback(() => {
    setTexteEdite(texte)
    setModeEdition(false)
  }, [texte])

  const confiancePourcent = Math.round(confiance * 100)
  const couleurConfiance =
    confiancePourcent >= 90 ? 'text-teal' : confiancePourcent >= 70 ? 'text-gold' : 'text-coral'

  return (
    <div className="bg-white rounded-lg p-8">
      {/* Header avec confiance */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h3 className="font-display text-xl font-semibold">Texte extrait</h3>
          <p className="text-sm text-ink-light">
            Confiance:{' '}
            <span className={`font-medium ${couleurConfiance}`}>{confiancePourcent}%</span>
            {zonesIncertaines.length > 0 && (
              <span className="ml-4">
                ({zonesIncertaines.length} zone{zonesIncertaines.length > 1 ? 's' : ''} incertaine
                {zonesIncertaines.length > 1 ? 's' : ''})
              </span>
            )}
          </p>
        </div>

        {!modeEdition && (
          <button
            type="button"
            className="px-6 py-4 bg-cream text-ink rounded-md font-medium hover:bg-cream/80 transition-colors"
            onClick={() => setModeEdition(true)}
          >
            Modifier
          </button>
        )}
      </div>

      {/* Legende zones incertaines */}
      {zonesIncertaines.length > 0 && !modeEdition && (
        <div className="mb-6 px-4 py-2 bg-gold/10 rounded-md text-sm text-ink-light flex items-center gap-4">
          <span className="inline-block w-4 h-4 bg-gold/30 rounded" />
          <span>
            Les zones en jaune sont incertaines. Vérifiez et corrigez si nécessaire.
          </span>
        </div>
      )}

      {/* Zone de texte */}
      {modeEdition ? (
        <div className="space-y-6">
          <textarea
            className="w-full h-64 p-6 border border-ink-muted rounded-md font-mono text-sm resize-y focus:border-coral focus:ring-1 focus:ring-coral"
            value={texteEdite}
            onChange={(e) => setTexteEdite(e.target.value)}
            autoFocus
          />
          <div className="flex gap-4 justify-end">
            <button
              type="button"
              className="px-6 py-4 bg-cream text-ink rounded-md font-medium hover:bg-cream/80 transition-colors"
              onClick={handleAnnuler}
            >
              Annuler
            </button>
            <button
              type="button"
              className="px-6 py-4 bg-teal text-white rounded-md font-medium hover:bg-teal-light transition-colors"
              onClick={handleSauvegarder}
            >
              Sauvegarder
            </button>
          </div>
        </div>
      ) : (
        <div className="p-6 bg-cream/30 rounded-md font-mono text-sm leading-relaxed whitespace-pre-wrap">
          {segments.map((segment, index) => (
            <span
              key={index}
              className={segment.estIncertain ? 'bg-gold/30 rounded px-0.5' : ''}
              title={segment.zone?.raison}
            >
              {segment.texte}
            </span>
          ))}
        </div>
      )}

      {/* Bouton de validation */}
      {!modeEdition && (
        <div className="mt-8 pt-6 border-t border-cream">
          <button
            type="button"
            className={`
              w-full py-6 px-8 rounded-full font-semibold text-lg transition-all
              ${
                enChargement
                  ? 'bg-ink-muted text-white cursor-not-allowed'
                  : 'bg-teal text-white hover:bg-teal-light hover:-translate-y-0.5 hover:shadow-xl'
              }
            `}
            onClick={onValider}
            disabled={enChargement}
          >
            {enChargement ? 'Validation...' : 'Valider et continuer'}
          </button>
        </div>
      )}
    </div>
  )
}
