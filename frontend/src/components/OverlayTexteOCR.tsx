import { useState, useRef, useCallback, useEffect } from 'react'
import type { BlocTexteOCR, PositionBlocOCR } from '../services/api'

interface OverlayTexteOCRProps {
  imageUrl: string
  blocs: BlocTexteOCR[]
  onBlocModifie?: (index: number, nouveauTexte: string) => void
  onBlocSupprime?: (index: number) => void
  onBlocDeplace?: (index: number, nouvellePosition: PositionBlocOCR) => void
  actionSlot?: React.ReactNode
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
  onBlocSupprime,
  onBlocDeplace,
  actionSlot,
}: OverlayTexteOCRProps) {
  const [blocActif, setBlocActif] = useState<number | null>(null)
  const [zoom, setZoom] = useState(1)
  const [blocSurvole, setBlocSurvole] = useState<number | null>(null)
  const conteneurRef = useRef<HTMLDivElement>(null)
  const imageRef = useRef<HTMLImageElement>(null)
  const [imageChargee, setImageChargee] = useState(false)
  const [minimisesParPage, setMinimisesParPage] = useState<Map<string, Set<number>>>(new Map())

  // Drag state
  const [blocGlisse, setBlocGlisse] = useState<number | null>(null)
  const [positionGlissement, setPositionGlissement] = useState<{ x: number; y: number } | null>(null)
  const pointDepart = useRef<{ clientX: number; clientY: number; origX: number; origY: number } | null>(null)

  const blocsMinimises = minimisesParPage.get(imageUrl) ?? new Set<number>()
  const tousMinimises = blocs.length > 0 && blocsMinimises.size === blocs.length

  const toggleMinimiser = useCallback((index: number) => {
    setMinimisesParPage(prev => {
      const next = new Map(prev)
      const pageSet = new Set(next.get(imageUrl) ?? [])
      if (pageSet.has(index)) pageSet.delete(index)
      else pageSet.add(index)
      next.set(imageUrl, pageSet)
      return next
    })
  }, [imageUrl])

  const toggleTousMinimises = useCallback(() => {
    setMinimisesParPage(prev => {
      const next = new Map(prev)
      if (tousMinimises) {
        next.set(imageUrl, new Set())
      } else {
        next.set(imageUrl, new Set(blocs.map((_, i) => i)))
      }
      return next
    })
  }, [imageUrl, tousMinimises, blocs])

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

  // Drag handlers
  const handleDebutGlissement = useCallback(
    (e: React.MouseEvent, index: number) => {
      e.preventDefault()
      e.stopPropagation()
      const bloc = blocs[index]
      if (!bloc) return
      pointDepart.current = {
        clientX: e.clientX,
        clientY: e.clientY,
        origX: bloc.position.x,
        origY: bloc.position.y,
      }
      setBlocGlisse(index)
      setPositionGlissement({ x: bloc.position.x, y: bloc.position.y })
    },
    [blocs]
  )

  const handleDebutGlissementTactile = useCallback(
    (e: React.TouchEvent, index: number) => {
      e.stopPropagation()
      const touch = e.touches[0]
      if (!touch) return
      const bloc = blocs[index]
      if (!bloc) return
      pointDepart.current = {
        clientX: touch.clientX,
        clientY: touch.clientY,
        origX: bloc.position.x,
        origY: bloc.position.y,
      }
      setBlocGlisse(index)
      setPositionGlissement({ x: bloc.position.x, y: bloc.position.y })
    },
    [blocs]
  )

  const calculerNouvellePosition = useCallback(
    (clientX: number, clientY: number): { x: number; y: number } | null => {
      if (!pointDepart.current || !imageRef.current || blocGlisse === null) return null
      const imageRect = imageRef.current.getBoundingClientRect()
      const bloc = blocs[blocGlisse]
      if (!bloc) return null

      const deltaX = ((clientX - pointDepart.current.clientX) / imageRect.width) * 100
      const deltaY = ((clientY - pointDepart.current.clientY) / imageRect.height) * 100

      const newX = Math.max(0, Math.min(100 - bloc.position.largeur, pointDepart.current.origX + deltaX))
      const newY = Math.max(0, Math.min(100 - bloc.position.hauteur, pointDepart.current.origY + deltaY))

      return { x: newX, y: newY }
    },
    [blocGlisse, blocs]
  )

  const handleMouvementGlissement = useCallback(
    (e: MouseEvent) => {
      const pos = calculerNouvellePosition(e.clientX, e.clientY)
      if (pos) setPositionGlissement(pos)
    },
    [calculerNouvellePosition]
  )

  const handleMouvementTactile = useCallback(
    (e: TouchEvent) => {
      e.preventDefault()
      const touch = e.touches[0]
      if (!touch) return
      const pos = calculerNouvellePosition(touch.clientX, touch.clientY)
      if (pos) setPositionGlissement(pos)
    },
    [calculerNouvellePosition]
  )

  const handleFinGlissement = useCallback(() => {
    if (blocGlisse !== null && positionGlissement && onBlocDeplace) {
      const bloc = blocs[blocGlisse]
      if (bloc) {
        onBlocDeplace(blocGlisse, {
          ...bloc.position,
          x: positionGlissement.x,
          y: positionGlissement.y,
        })
      }
    }
    setBlocGlisse(null)
    setPositionGlissement(null)
    pointDepart.current = null
  }, [blocGlisse, positionGlissement, onBlocDeplace, blocs])

  // Window event listeners for drag
  useEffect(() => {
    if (blocGlisse === null) return

    window.addEventListener('mousemove', handleMouvementGlissement)
    window.addEventListener('mouseup', handleFinGlissement)
    window.addEventListener('touchmove', handleMouvementTactile, { passive: false })
    window.addEventListener('touchend', handleFinGlissement)

    return () => {
      window.removeEventListener('mousemove', handleMouvementGlissement)
      window.removeEventListener('mouseup', handleFinGlissement)
      window.removeEventListener('touchmove', handleMouvementTactile)
      window.removeEventListener('touchend', handleFinGlissement)
    }
  }, [blocGlisse, handleMouvementGlissement, handleFinGlissement, handleMouvementTactile])

  const calculerTaillePolice = useCallback(
    (hauteurPct: number): string => {
      if (!imageRef.current || !imageChargee) return '12px'
      const hauteurImage = imageRef.current.clientHeight
      const hauteurBlocPx = (hauteurPct / 100) * hauteurImage
      const taille = Math.max(8, Math.min(32, hauteurBlocPx * 0.85))
      return `${taille}px`
    },
    [imageChargee]
  )

  return (
    <div className="space-y-4">
      {/* Controles */}
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <button
            data-testid="toggle-minimiser-bbox"
            onClick={toggleTousMinimises}
            className={`px-4 py-2 rounded-full text-sm font-medium transition-colors ${
              !tousMinimises
                ? 'bg-teal text-white hover:bg-teal-light'
                : 'bg-cream text-ink hover:bg-cream-dark'
            }`}
          >
            {tousMinimises ? 'Voir le texte OCR' : 'Masquer le texte OCR'}
          </button>
          {/* Zoom */}
          <div className="flex items-center gap-1">
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
              className="px-2 h-8 flex items-center justify-center bg-cream text-ink rounded-sm text-xs font-medium hover:bg-cream-dark transition-colors"
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
        </div>

        {/* Legende confiance */}
        {!tousMinimises && (
          <div className="hidden sm:flex items-center gap-4 text-xs text-ink-muted">
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
        )}
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
          {actionSlot && (
            <div className="absolute top-2 right-2 z-30">
              {actionSlot}
            </div>
          )}

          {imageChargee &&
            blocs.map((bloc, index) => {
              const couleur = getCouleurConfiance(bloc.confiance)
              const estActif = blocActif === index
              const estSurvole = blocSurvole === index
              const estMinimise = blocsMinimises.has(index)
              const estGlisse = blocGlisse === index
              const taillePolice = calculerTaillePolice(bloc.position.hauteur)

              const posX = estGlisse && positionGlissement ? positionGlissement.x : bloc.position.x
              const posY = estGlisse && positionGlissement ? positionGlissement.y : bloc.position.y

              return (
                <div
                  key={index}
                  data-testid="bbox-bloc"
                  className={`absolute overflow-visible ${
                    estGlisse
                      ? 'shadow-xl z-50'
                      : 'transition-all duration-150'
                  } ${
                    estActif
                      ? `ring-2 ring-coral shadow-lg ${couleur.border} border-2`
                      : `${couleur.border} border ${estSurvole ? 'border-2 shadow-md' : ''}`
                  }`}
                  style={{
                    left: `${posX}%`,
                    top: `${posY}%`,
                    width: `${bloc.position.largeur}%`,
                    height: estMinimise ? '0.6%' : `${bloc.position.hauteur}%`,
                    minHeight: estMinimise ? '6px' : undefined,
                    opacity: estGlisse ? 0.85 : undefined,
                    cursor: estGlisse ? 'grabbing' : undefined,
                  }}
                  onMouseEnter={() => { if (blocGlisse === null) setBlocSurvole(index) }}
                  onMouseLeave={() => { if (blocGlisse === null) setBlocSurvole(null) }}
                >
                  <div
                    className={`absolute inset-0 ${
                      estActif ? 'bg-white/80' : estSurvole ? 'bg-white/60' : 'bg-white/40'
                    } ${couleur.bg} rounded-[2px]`}
                  />

                  {!estMinimise && (
                    <div className="relative w-full h-full">
                      <textarea
                        value={bloc.texte}
                        readOnly={!estActif}
                        onChange={(e) => handleTexteChange(index, e.target.value)}
                        onKeyDown={(e) => handleKeyDown(index, e)}
                        onFocus={() => { if (blocGlisse === null) setBlocActif(index) }}
                        onBlur={() => setBlocActif(null)}
                        onClick={(e) => e.stopPropagation()}
                        className={`w-full h-full p-0.5 bg-transparent text-ink leading-tight resize-none focus:outline-none font-body cursor-text ${
                          estActif ? '' : 'select-none'
                        }`}
                        style={{ fontSize: taillePolice }}
                      />
                    </div>
                  )}

                  {/* Boutons d'action — coin haut droit, dans la bbox (après textarea pour z-order) */}
                  {(estSurvole || estActif) && blocGlisse === null && (
                    <div className="absolute top-0 right-0 z-20 flex gap-0.5 p-0.5" onClick={(e) => e.stopPropagation()}>
                      {onBlocDeplace && (
                        <button
                          data-testid="btn-deplacer-bloc"
                          onMouseDown={(e) => handleDebutGlissement(e, index)}
                          onTouchStart={(e) => handleDebutGlissementTactile(e, index)}
                          className="w-5 h-5 bg-teal/80 text-white rounded-sm text-[10px] flex items-center justify-center hover:bg-teal transition-colors cursor-grab active:cursor-grabbing"
                          title="Déplacer ce bloc"
                        >
                          {'\u2807'}
                        </button>
                      )}
                      <button
                        data-testid="btn-minimiser-bloc"
                        onClick={() => toggleMinimiser(index)}
                        className="w-5 h-5 bg-ink/70 text-white rounded-sm text-[10px] flex items-center justify-center hover:bg-ink transition-colors"
                        title={estMinimise ? 'Agrandir' : 'Réduire'}
                      >
                        {estMinimise ? '\u2195' : '\u2014'}
                      </button>
                      {onBlocSupprime && (
                        <button
                          data-testid="btn-supprimer-bloc"
                          onClick={() => onBlocSupprime(index)}
                          className="w-5 h-5 bg-coral/80 text-white rounded-sm text-[10px] flex items-center justify-center hover:bg-coral transition-colors"
                          title="Supprimer ce bloc"
                        >
                          {'×'}
                        </button>
                      )}
                    </div>
                  )}

                  {/* Tooltip confiance — au-dessus à gauche */}
                  {estSurvole && !estActif && !estMinimise && (
                    <div className="absolute -top-6 left-0 z-20 px-2 py-0.5 bg-ink text-white text-[10px] rounded-sm whitespace-nowrap shadow-lg pointer-events-none">
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
          Cliquez sur un bloc pour le modifier. Tab pour naviguer, Echap pour deselectionner.{onBlocDeplace ? ' Glissez ⁞ pour repositionner.' : ''}
        </span>
      </div>
    </div>
  )
}
