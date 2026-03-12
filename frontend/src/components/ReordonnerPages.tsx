import { useState, useRef, useCallback, useEffect } from 'react'

export type EtatPage = 'upload' | 'attente' | 'en_cours' | 'termine'

interface ReordonnerPagesProps {
  images: string[]
  getImageUrl: (nomFichier: string) => string
  imageSelectionnee: number
  onSelectionner: (index: number) => void
  onReordonner: (nouvelOrdre: string[]) => void
  onSupprimer?: (nomFichier: string) => void
  desactive?: boolean
  etatPages?: Map<number, EtatPage>
}

export default function ReordonnerPages({
  images,
  getImageUrl,
  imageSelectionnee,
  onSelectionner,
  onReordonner,
  onSupprimer,
  desactive = false,
  etatPages,
}: ReordonnerPagesProps) {
  const [indexGlisse, setIndexGlisse] = useState<number | null>(null)
  const [indexCible, setIndexCible] = useState<number | null>(null)
  const [touchActive, setTouchActive] = useState(false)
  const conteneurRef = useRef<HTMLDivElement>(null)
  const longPressTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const touchStartPos = useRef({ x: 0, y: 0 })

  const peutReordonner = !desactive && images.length > 1

  const appliquerReordonnancement = useCallback(
    (fromIdx: number, toIdx: number) => {
      if (fromIdx === toIdx) return
      const arr = [...images]
      const [item] = arr.splice(fromIdx, 1)
      arr.splice(toIdx, 0, item)
      onReordonner(arr)
      // Suivre la selection
      if (imageSelectionnee === fromIdx) {
        onSelectionner(toIdx)
      } else if (fromIdx < imageSelectionnee && toIdx >= imageSelectionnee) {
        onSelectionner(imageSelectionnee - 1)
      } else if (fromIdx > imageSelectionnee && toIdx <= imageSelectionnee) {
        onSelectionner(imageSelectionnee + 1)
      }
    },
    [images, imageSelectionnee, onReordonner, onSelectionner]
  )

  // === Desktop : HTML5 drag & drop ===
  const handleDragStart = useCallback(
    (e: React.DragEvent, idx: number) => {
      if (!peutReordonner) {
        e.preventDefault()
        return
      }
      setIndexGlisse(idx)
      e.dataTransfer.effectAllowed = 'move'
    },
    [peutReordonner]
  )

  const handleDragOver = useCallback(
    (e: React.DragEvent, idx: number) => {
      e.preventDefault()
      e.dataTransfer.dropEffect = 'move'
      if (indexGlisse !== null && idx !== indexGlisse) {
        setIndexCible(idx)
      }
    },
    [indexGlisse]
  )

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault()
      if (indexGlisse !== null && indexCible !== null) {
        appliquerReordonnancement(indexGlisse, indexCible)
      }
      setIndexGlisse(null)
      setIndexCible(null)
    },
    [indexGlisse, indexCible, appliquerReordonnancement]
  )

  const handleDragEnd = useCallback(() => {
    setIndexGlisse(null)
    setIndexCible(null)
  }, [])

  // === Mobile : long press + drag tactile ===
  const calculerIndexCible = useCallback(
    (touchX: number): number | null => {
      if (!conteneurRef.current) return null
      const items = conteneurRef.current.querySelectorAll('[data-page-idx]')
      for (let i = 0; i < items.length; i++) {
        const rect = items[i].getBoundingClientRect()
        const center = rect.left + rect.width / 2
        if (touchX < center) return i
      }
      return items.length - 1
    },
    []
  )

  const handleTouchStart = useCallback(
    (e: React.TouchEvent, idx: number) => {
      if (!peutReordonner) return
      const touch = e.touches[0]
      touchStartPos.current = { x: touch.clientX, y: touch.clientY }
      longPressTimer.current = setTimeout(() => {
        setIndexGlisse(idx)
        setTouchActive(true)
        if (navigator.vibrate) navigator.vibrate(30)
      }, 400)
    },
    [peutReordonner]
  )

  const handleTouchMove = useCallback(
    (e: React.TouchEvent) => {
      const touch = e.touches[0]
      // Annuler le long press si le doigt bouge trop avant l'activation
      if (longPressTimer.current && !touchActive) {
        const dx = Math.abs(touch.clientX - touchStartPos.current.x)
        const dy = Math.abs(touch.clientY - touchStartPos.current.y)
        if (dx > 10 || dy > 10) {
          clearTimeout(longPressTimer.current)
          longPressTimer.current = null
        }
        return
      }
      if (!touchActive) return
      e.preventDefault()
      const cible = calculerIndexCible(touch.clientX)
      if (cible !== null && cible !== indexGlisse) {
        setIndexCible(cible)
      }
    },
    [touchActive, indexGlisse, calculerIndexCible]
  )

  const handleTouchEnd = useCallback(() => {
    if (longPressTimer.current) {
      clearTimeout(longPressTimer.current)
      longPressTimer.current = null
    }
    if (touchActive && indexGlisse !== null && indexCible !== null) {
      appliquerReordonnancement(indexGlisse, indexCible)
    }
    setTouchActive(false)
    setIndexGlisse(null)
    setIndexCible(null)
  }, [touchActive, indexGlisse, indexCible, appliquerReordonnancement])

  useEffect(() => {
    return () => {
      if (longPressTimer.current) clearTimeout(longPressTimer.current)
    }
  }, [])

  if (images.length === 0) return null

  return (
    <div className="space-y-3">
      {/* Vignettes avec drag & drop */}
      <div
        ref={conteneurRef}
        className="flex flex-wrap gap-4 py-2 px-1"
        style={touchActive ? { touchAction: 'none' } : undefined}
        onDrop={handleDrop}
        onDragOver={(e) => e.preventDefault()}
      >
        {images.map((img, idx) => {
          const estGlisse = indexGlisse === idx
          const estCible =
            indexCible === idx && indexGlisse !== null && indexGlisse !== idx

          return (
            <div
              key={img}
              data-page-idx={idx}
              className={`relative flex-shrink-0 transition-all duration-150 ${
                estGlisse ? 'opacity-40 scale-90' : ''
              }`}
              draggable={peutReordonner}
              onDragStart={(e) => handleDragStart(e, idx)}
              onDragOver={(e) => handleDragOver(e, idx)}
              onDragEnd={handleDragEnd}
              onTouchStart={(e) => handleTouchStart(e, idx)}
              onTouchMove={handleTouchMove}
              onTouchEnd={handleTouchEnd}
            >
              {/* Indicateur de drop */}
              {estCible && (
                <div className="absolute -left-2.5 top-0 bottom-6 w-1 bg-coral rounded-full z-10" />
              )}

              <button
                data-testid="image-vignette"
                onClick={() => {
                  if (!touchActive) onSelectionner(idx)
                }}
                className={`w-24 h-24 rounded-lg overflow-hidden border-3 transition-all ${
                  imageSelectionnee === idx
                    ? 'border-coral ring-2 ring-coral/30 shadow-lg'
                    : 'border-cream-dark hover:border-coral/50'
                } ${peutReordonner ? 'cursor-grab active:cursor-grabbing' : 'cursor-pointer'}`}
              >
                <img
                  src={getImageUrl(img)}
                  alt={`Page ${idx + 1}`}
                  className="w-full h-full object-cover"
                  draggable={false}
                />
                {/* Indicateur de statut OCR */}
                {etatPages?.get(idx) === 'upload' && (
                  <div data-testid="ocr-page-upload" className="absolute top-1 right-1 w-5 h-5 bg-blue-500 rounded-full flex items-center justify-center shadow-sm">
                    <svg className="w-3 h-3 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={3}>
                      <path strokeLinecap="round" strokeLinejoin="round" d="M5 10l7-7m0 0l7 7m-7-7v18" />
                    </svg>
                  </div>
                )}
                {etatPages?.get(idx) === 'attente' && (
                  <div data-testid="ocr-page-waiting" className="absolute top-1 right-1 w-5 h-5 bg-gray-300 rounded-full flex items-center justify-center shadow-sm">
                    <div className="w-2 h-2 bg-white rounded-full opacity-60" />
                  </div>
                )}
                {etatPages?.get(idx) === 'en_cours' && (
                  <div data-testid="ocr-page-in-progress" className="absolute top-1 right-1 w-5 h-5 bg-coral rounded-full flex items-center justify-center shadow-sm">
                    <div className="w-3 h-3 border-2 border-white border-t-transparent rounded-full animate-spin" />
                  </div>
                )}
                {etatPages?.get(idx) === 'termine' && (
                  <div data-testid="ocr-page-done" className="absolute top-1 right-1 w-5 h-5 bg-emerald-500 rounded-full flex items-center justify-center shadow-sm">
                    <svg className="w-3 h-3 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={3}>
                      <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                    </svg>
                  </div>
                )}
              </button>
              {/* Bouton supprimer sur la vignette */}
              {onSupprimer && !desactive && !etatPages?.has(idx) && (
                <button
                  data-testid="supprimer-image"
                  onClick={(e) => { e.stopPropagation(); onSupprimer(img) }}
                  className="absolute -top-1.5 -right-1.5 w-5 h-5 bg-coral/80 text-white rounded-full text-[10px] flex items-center justify-center hover:bg-coral transition-colors shadow-sm z-10"
                  title="Supprimer cette page"
                >
                  {'×'}
                </button>
              )}
              <div className="absolute -bottom-2 left-1/2 -translate-x-1/2">
                <span
                  className={`text-xs font-semibold px-2.5 py-1 rounded-full shadow-sm ${
                    imageSelectionnee === idx
                      ? 'bg-coral text-white'
                      : 'bg-ink text-white'
                  }`}
                >
                  {idx + 1}
                </span>
              </div>
            </div>
          )
        })}
      </div>

      {/* Indicateur de page */}
      {images.length > 1 && (
        <div className="flex items-center justify-center">
          <span className="text-sm text-ink-muted">
            Page {imageSelectionnee + 1} sur {images.length}
          </span>
        </div>
      )}

      {desactive && (
        <p className="text-xs text-ink-muted text-center">
          Reorganisation indisponible pendant le traitement OCR
        </p>
      )}
    </div>
  )
}
