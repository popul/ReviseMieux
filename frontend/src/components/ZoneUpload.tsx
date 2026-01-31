import { useCallback, useRef, useState } from 'react'

interface ZoneUploadProps {
  onFichiersSelectionnes: (fichiers: File[]) => void
  maxFichiers: number
  fichiersCourants: number
  desactivee?: boolean
}

const FORMATS_ACCEPTES = ['image/jpeg', 'image/png', 'image/webp', 'application/pdf']
const EXTENSIONS_ACCEPTEES = '.jpg,.jpeg,.png,.webp,.pdf'

export default function ZoneUpload({
  onFichiersSelectionnes,
  maxFichiers,
  fichiersCourants,
  desactivee = false,
}: ZoneUploadProps) {
  const [estSurvol, setEstSurvol] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)

  const fichiersRestants = maxFichiers - fichiersCourants
  const aDejaDesFichiers = fichiersCourants > 0

  const validerFichiers = useCallback(
    (fichiers: FileList | null) => {
      if (!fichiers) return

      const fichiersValides: File[] = []
      for (let i = 0; i < fichiers.length && fichiersValides.length < fichiersRestants; i++) {
        const fichier = fichiers[i]
        if (FORMATS_ACCEPTES.includes(fichier.type)) {
          fichiersValides.push(fichier)
        }
      }

      if (fichiersValides.length > 0) {
        onFichiersSelectionnes(fichiersValides)
      }
    },
    [fichiersRestants, onFichiersSelectionnes]
  )

  const gererDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault()
      setEstSurvol(false)
      if (!desactivee) {
        validerFichiers(e.dataTransfer.files)
      }
    },
    [desactivee, validerFichiers]
  )

  const gererDragOver = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault()
      if (!desactivee) {
        setEstSurvol(true)
      }
    },
    [desactivee]
  )

  const gererDragLeave = useCallback(() => {
    setEstSurvol(false)
  }, [])

  const gererClic = useCallback(() => {
    if (!desactivee) {
      inputRef.current?.click()
    }
  }, [desactivee])

  const gererChangementInput = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      validerFichiers(e.target.files)
      // Reset l'input pour permettre de resélectionner le même fichier
      e.target.value = ''
    },
    [validerFichiers]
  )

  return (
    <div
      className={`
        bg-white rounded-lg p-xl cursor-pointer transition-all relative overflow-hidden
        ${
          aDejaDesFichiers
            ? 'border-2 border-solid border-success'
            : estSurvol
            ? 'border-[3px] border-dashed border-coral bg-coral/5 scale-[1.01]'
            : 'border-[3px] border-dashed border-ink-muted hover:border-coral hover:bg-coral/[0.02]'
        }
        ${desactivee ? 'opacity-50 cursor-not-allowed' : ''}
      `}
      onDrop={gererDrop}
      onDragOver={gererDragOver}
      onDragLeave={gererDragLeave}
      onClick={gererClic}
      role="button"
      tabIndex={desactivee ? -1 : 0}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault()
          gererClic()
        }
      }}
      aria-label="Zone de téléchargement. Glissez des fichiers ou cliquez pour sélectionner."
    >
      <input
        ref={inputRef}
        type="file"
        className="hidden"
        accept={EXTENSIONS_ACCEPTEES}
        multiple
        onChange={gererChangementInput}
        disabled={desactivee}
      />

      <div className="text-center">
        <span className="text-6xl mb-md block">📸</span>
        <p className="font-display text-xl font-semibold mb-xs">Glisse tes fichiers ici</p>
        <p className="text-ink-muted mb-md">ou clique pour sélectionner</p>

        <button
          type="button"
          className="inline-flex items-center gap-xs bg-coral text-white py-3.5 px-lg rounded-full font-semibold transition-all hover:bg-coral-dark hover:-translate-y-0.5"
          onClick={(e) => {
            e.stopPropagation()
            gererClic()
          }}
          disabled={desactivee}
        >
          <span>📁</span> Choisir des fichiers
        </button>

        <p className="mt-md text-sm text-ink-muted">
          JPG, PNG ou PDF - Max {maxFichiers} pages
        </p>
      </div>
    </div>
  )
}
