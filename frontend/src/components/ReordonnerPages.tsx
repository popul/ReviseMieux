import { useState, useCallback } from 'react'

interface ReordonnerPagesProps {
  images: string[]
  getImageUrl: (nomFichier: string) => string
  imageSelectionnee: number
  onSelectionner: (index: number) => void
  onReordonner: (nouvelOrdre: string[]) => void
  modeEdition: boolean
  onSupprimer?: (nomFichier: string) => void
}

export default function ReordonnerPages({
  images,
  getImageUrl,
  imageSelectionnee,
  onSelectionner,
  onReordonner,
  modeEdition,
  onSupprimer,
}: ReordonnerPagesProps) {
  const [dragIndex, setDragIndex] = useState<number | null>(null)
  const [dragOverIndex, setDragOverIndex] = useState<number | null>(null)

  const handleDragStart = useCallback((index: number, e: React.DragEvent) => {
    setDragIndex(index)
    e.dataTransfer.effectAllowed = 'move'
    const target = e.currentTarget as HTMLElement
    target.style.opacity = '0.5'
  }, [])

  const handleDragEnd = useCallback((e: React.DragEvent) => {
    const target = e.currentTarget as HTMLElement
    target.style.opacity = '1'
    setDragIndex(null)
    setDragOverIndex(null)
  }, [])

  const handleDragOver = useCallback((index: number, e: React.DragEvent) => {
    e.preventDefault()
    e.dataTransfer.dropEffect = 'move'
    setDragOverIndex(index)
  }, [])

  const handleDragLeave = useCallback(() => {
    setDragOverIndex(null)
  }, [])

  const handleDrop = useCallback((targetIndex: number, e: React.DragEvent) => {
    e.preventDefault()
    if (dragIndex === null || dragIndex === targetIndex) {
      setDragIndex(null)
      setDragOverIndex(null)
      return
    }
    const nouvellesImages = [...images]
    const [element] = nouvellesImages.splice(dragIndex, 1)
    nouvellesImages.splice(targetIndex, 0, element)
    onReordonner(nouvellesImages)
    if (imageSelectionnee === dragIndex) {
      onSelectionner(targetIndex)
    } else if (dragIndex < imageSelectionnee && targetIndex >= imageSelectionnee) {
      onSelectionner(imageSelectionnee - 1)
    } else if (dragIndex > imageSelectionnee && targetIndex <= imageSelectionnee) {
      onSelectionner(imageSelectionnee + 1)
    }
    setDragIndex(null)
    setDragOverIndex(null)
  }, [dragIndex, images, imageSelectionnee, onReordonner, onSelectionner])

  if (images.length === 0) return null

  return (
    <div className="space-y-3">
      <div className="flex gap-4 overflow-x-auto py-2 px-1">
        {images.map((img, idx) => (
          <div
            key={img}
            className={`relative flex-shrink-0 transition-all duration-200 ${dragOverIndex === idx && dragIndex !== idx ? 'translate-x-4' : ''}`}
            draggable={modeEdition}
            onDragStart={(e) => handleDragStart(idx, e)}
            onDragEnd={handleDragEnd}
            onDragOver={(e) => handleDragOver(idx, e)}
            onDragLeave={handleDragLeave}
            onDrop={(e) => handleDrop(idx, e)}
          >
            <button
              data-testid="image-vignette"
              onClick={() => onSelectionner(idx)}
              className={`w-24 h-24 rounded-lg overflow-hidden border-3 transition-all ${imageSelectionnee === idx ? 'border-coral ring-2 ring-coral/30 shadow-lg' : 'border-cream-dark hover:border-coral/50'} ${modeEdition ? 'cursor-grab active:cursor-grabbing' : 'cursor-pointer'} ${dragIndex === idx ? 'opacity-50' : ''}`}
            >
              <img src={getImageUrl(img)} alt={`Page ${idx + 1}`} className="w-full h-full object-cover" draggable={false} />
            </button>
            <div className="absolute -bottom-2 left-1/2 -translate-x-1/2">
              <span className={`text-xs font-semibold px-2.5 py-1 rounded-full shadow-sm ${imageSelectionnee === idx ? 'bg-coral text-white' : 'bg-ink text-white'}`}>
                {idx + 1}
              </span>
            </div>
            {dragOverIndex === idx && dragIndex !== idx && (
              <div className="absolute -left-1 top-0 bottom-0 w-0.5 bg-coral rounded-full" />
            )}
            {modeEdition && onSupprimer && (
              <button
                data-testid="supprimer-image"
                onClick={(e) => { e.stopPropagation(); onSupprimer(img) }}
                className="absolute -top-2 -right-2 w-6 h-6 bg-red-500 text-white rounded-full text-sm flex items-center justify-center hover:bg-red-600 transition-colors shadow-md"
                title="Supprimer"
              >
                x
              </button>
            )}
          </div>
        ))}
      </div>
      {images.length > 1 && (
        <div className="flex items-center justify-center gap-2">
          <span className="text-sm text-ink-muted">
            Page {imageSelectionnee + 1} sur {images.length}
          </span>
          {modeEdition && (
            <span className="text-xs text-ink-muted ml-2">(glissez pour reorganiser)</span>
          )}
        </div>
      )}
    </div>
  )
}
