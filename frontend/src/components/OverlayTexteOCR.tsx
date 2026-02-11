import { useState, useRef, useCallback } from 'react'
import type { BlocTexteOCR } from '../services/api'

interface OverlayTexteOCRProps {
  imageUrl: string
  blocs: BlocTexteOCR[]
  onBlocModifie?: (index: number, nouveauTexte: string) => void
}

function getCouleurConfiance(confiance: number): {
  border: string
  bg: string
  label: string
} {
  if (confiance >= 0.9) {
    return { border: 'border-success', bg: 'bg-success/10', label: 'Haute confiance' }
  }
  if (confiance >= 0.7) {
    return { border: 'border-gold', bg: 'bg-gold/10', label: 'Confiance moyenne' }
  }
  return { border: 'border-coral', bg: 'bg-coral/10', label: 'Confiance faible' }
}

export default function OverlayTexteOCR({
  imageUrl,
  blocs,
  onBlocModifie,
}: OverlayTexteOCRProps) {
  const [blocActif, setBlocActif] = useState<number | null>(null)
  const [zoom, setZoom] = useState(1)
  const [blocSurvole, setBlocSurvole] = useState<number | null>(null)
  const conteneurRef = useRef<HTMLDivElement>(null)
  const imageRef = useRef<HTMLImageElement>(null)
  const [imageChargee, setImageChargee] = useState(false)

  const handleZoomIn = useCallback(() => {
    setZoom((z) => Math.min(z + 0.25, 3))
  }, [])

  const handleZoomOut = useCallback(() => {
    setZoom((z) => Math.max(z - 0.25, 0.5))
  }, [])

  const handleZoomReset = useCallback(() => {
    setZoom(1)
  }, [])

  const handleTexteChange = useCallback(
    (index: number, nouveauTexte: string) => {
      onBlocModifie?.(index, nouveauTexte)
    },
    [onBlocModifie]
  )

  const handleKeyDown = useCallback(
    (index: number, e: React.KeyboardEvent) => {
      if (e.key === 'Escape') {
        setBlocActif(null)
      }
      if (e.key === 'Tab') {
        e.preventDefault()
        const suivant = e.shiftKey
          ? (index - 1 + blocs.length) % blocs.length
          : (index + 1) % blocs.length
        setBlocActif(suivant)
      }
    },
    [blocs.length]
  )

  const handleOverlayClick = useCallback(() => {
    setBlocActif(null)
  }, [])

  // Calculer la taille de police en fonction de la hauteur du bloc et de la hauteur de l'image
  const calculerTaillePolice = useCallback(
    (hauteurPct: number): string => {
      if (!imageRef.current || !imageChargee) return '12px'
      const hauteurImage = imageRef.current.clientHeight
      const hauteurBlocPx = (hauteurPct / 100) * hauteurImage
      // La taille de police est proportionnelle a la hauteur du bloc
      // On estime qu'une ligne de texte occupe ~85% de la hauteur du bloc
      const taille = Math.max(8, Math.min(32, hauteurBlocPx * 0.85))
      return `${taille}px`
    },
    [imageChargee]
  )

  return (
    <div className="space-y-4">
      {/* Controles de zoom */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <button
            onClick={handleZoomOut}
            disabled={zoom <= 0.5}
            className="w-8 h-8 flex items-center justify-center bg-cream text-ink rounded-sm text-sm font-medium hover:bg-cream-dark transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
            title="Dezoomer"
          >
            -
          </button>
          <button
            onClick={handleZoomReset}
            className="px-3 h-8 flex items-center justify-center bg-cream text-ink rounded-sm text-xs font-medium hover:bg-cream-dark transition-colors"
            title="Reinitialiser le zoom"
          >
            {Math.round(zoom * 100)}%
          </button>
          <button
            onClick={handleZoomIn}
            disabled={zoom >= 3}
            className="w-8 h-8 flex items-center justify-center bg-cream text-ink rounded-sm text-sm font-medium hover:bg-cream-dark transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
            title="Zoomer"
          >
            +
          </button>
        </div>

        {/* Legende confiance */}
        <div className="flex items-center gap-4 text-xs text-ink-muted">
          <span className="flex items-center gap-1.5">
            <span className="w-3 h-3 rounded-sm border-2 border-success bg-success/10" />
            {'\u2265'}90%
          </span>
          <span className="flex items-center gap-1.5">
            <span className="w-3 h-3 rounded-sm border-2 border-gold bg-gold/10" />
            {'\u2265'}70%
          </span>
          <span className="flex items-center gap-1.5">
            <span className="w-3 h-3 rounded-sm border-2 border-coral bg-coral/10" />
            {'<'}70%
          </span>
        </div>
      </div>

      {/* Conteneur image + overlay */}
      <div
        ref={conteneurRef}
        className="relative rounded-md overflow-auto border border-cream-dark"
        style={{ maxHeight: '70vh' }}
        onClick={handleOverlayClick}
      >
        <div
          className="relative inline-block min-w-full transition-transform duration-200"
          style={{ transform: `scale(${zoom})`, transformOrigin: 'top left' }}
        >
          <img
            ref={imageRef}
            src={imageUrl}
            alt="Page du cours"
            className="block w-full"
            draggable={false}
            onLoad={() => setImageChargee(true)}
          />

          {imageChargee &&
            blocs.map((bloc, index) => {
              const couleur = getCouleurConfiance(bloc.confiance)
              const estActif = blocActif === index
              const estSurvole = blocSurvole === index
              const taillePolice = calculerTaillePolice(bloc.position.hauteur)

              return (
                <div
                  key={index}
                  className={`absolute transition-all duration-150 ${
                    estActif
                      ? `ring-2 ring-coral shadow-lg ${couleur.border} border-2`
                      : `${couleur.border} border ${estSurvole ? 'border-2 shadow-md' : ''}`
                  }`}
                  style={{
                    left: `${bloc.position.x}%`,
                    top: `${bloc.position.y}%`,
                    width: `${bloc.position.largeur}%`,
                    height: `${bloc.position.hauteur}%`,
                  }}
                  onMouseEnter={() => setBlocSurvole(index)}
                  onMouseLeave={() => setBlocSurvole(null)}
                >
                  <div
                    className={`absolute inset-0 ${
                      estActif ? 'bg-white/90' : 'bg-white/70'
                    } ${couleur.bg} rounded-[2px]`}
                  />

                  <div className="relative w-full h-full">
                    <textarea
                      value={bloc.texte}
                      readOnly={!estActif}
                      onChange={(e) => handleTexteChange(index, e.target.value)}
                      onKeyDown={(e) => handleKeyDown(index, e)}
                      onFocus={() => setBlocActif(index)}
                      onBlur={() => setBlocActif(null)}
                      onClick={(e) => e.stopPropagation()}
                      className={`w-full h-full p-0.5 bg-transparent text-ink leading-tight resize-none focus:outline-none font-body cursor-text ${
                        estActif ? '' : 'select-none'
                      }`}
                      style={{ fontSize: taillePolice }}
                    />
                  </div>

                  {estSurvole && !estActif && (
                    <div className="absolute -top-8 left-0 z-20 px-2 py-1 bg-ink text-white text-[10px] rounded-sm whitespace-nowrap shadow-lg pointer-events-none">
                      {couleur.label} - {Math.round(bloc.confiance * 100)}%
                    </div>
                  )}
                </div>
              )
            })}
        </div>
      </div>

      {/* Instructions */}
      <div className="text-xs text-ink-muted text-center">
        <span>
          Cliquez sur un bloc pour le modifier. Tab pour naviguer, Echap pour deselectionner.
        </span>
      </div>
    </div>
  )
}
