import { useEffect, useState } from 'react'

interface FichierAvecPreview {
  fichier: File
  urlPreview: string | null
}

interface PreviewFichiersProps {
  fichiers: File[]
  onSupprimer: (index: number) => void
  onAjouter: () => void
  maxFichiers: number
}

function CarteFichier({
  fichier,
  urlPreview,
  onSupprimer,
  compact,
}: {
  fichier: FichierAvecPreview
  urlPreview: string | null
  onSupprimer: () => void
  compact: boolean
}) {
  return (
    <div className={`bg-white rounded-md overflow-hidden relative shadow group ${
      compact ? 'aspect-square' : 'aspect-[3/4]'
    }`}>
      {urlPreview ? (
        <img
          src={urlPreview}
          alt={fichier.fichier.name}
          className="w-full h-full object-cover"
        />
      ) : (
        <div className="w-full h-full flex items-center justify-center bg-cream text-2xl">
          📄
        </div>
      )}

      {/* Bouton supprimer */}
      <button
        type="button"
        className={`absolute top-1 right-1 rounded-full bg-ink/70 text-white flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity hover:bg-coral ${
          compact ? 'w-5 h-5 text-xs' : 'w-7 h-7 text-lg'
        }`}
        onClick={onSupprimer}
        aria-label={`Supprimer ${fichier.fichier.name}`}
      >
        ×
      </button>

      {/* Nom du fichier */}
      <div className="absolute bottom-0 left-0 right-0 px-1.5 py-1 bg-gradient-to-t from-black/70 to-transparent">
        <p className={`text-white truncate ${compact ? 'text-[10px]' : 'text-xs'}`}>
          {fichier.fichier.name}
        </p>
      </div>
    </div>
  )
}

function BoutonAjouter({ onClick, desactive, compact }: { onClick: () => void; desactive: boolean; compact: boolean }) {
  return (
    <button
      type="button"
      className={`
        flex flex-col items-center justify-center gap-1 bg-cream border-2 border-dashed border-ink-muted
        rounded-md text-ink-muted font-medium transition-all
        ${compact ? 'aspect-square' : 'aspect-[3/4]'}
        ${desactive ? 'opacity-50 cursor-not-allowed' : 'hover:border-coral hover:text-coral cursor-pointer'}
      `}
      onClick={onClick}
      disabled={desactive}
      aria-label="Ajouter plus de fichiers"
    >
      <span className={compact ? 'text-lg' : 'text-2xl'}>+</span>
      <span className={compact ? 'text-[10px]' : 'text-sm'}>Ajouter</span>
    </button>
  )
}

export default function PreviewFichiers({
  fichiers,
  onSupprimer,
  onAjouter,
  maxFichiers,
}: PreviewFichiersProps) {
  const [previews, setPreviews] = useState<Map<File, string>>(new Map())

  // Générer les previews pour les images
  useEffect(() => {
    const nouvellesPreviews = new Map<File, string>()

    fichiers.forEach((fichier) => {
      if (fichier.type.startsWith('image/')) {
        // Réutiliser l'URL existante si possible
        const urlExistante = previews.get(fichier)
        if (urlExistante) {
          nouvellesPreviews.set(fichier, urlExistante)
        } else {
          nouvellesPreviews.set(fichier, URL.createObjectURL(fichier))
        }
      }
    })

    // Libérer les anciennes URLs non utilisées
    previews.forEach((url, fichier) => {
      if (!nouvellesPreviews.has(fichier)) {
        URL.revokeObjectURL(url)
      }
    })

    setPreviews(nouvellesPreviews)

    // Cleanup au démontage
    return () => {
      nouvellesPreviews.forEach((url) => URL.revokeObjectURL(url))
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [fichiers])

  if (fichiers.length === 0) return null

  const peutAjouter = fichiers.length < maxFichiers
  const compact = fichiers.length > 8

  // Grille adaptative : petites vignettes quand beaucoup de fichiers
  const gridClass = compact
    ? 'grid grid-cols-4 sm:grid-cols-6 md:grid-cols-8 lg:grid-cols-10 gap-2'
    : 'grid grid-cols-[repeat(auto-fill,minmax(140px,1fr))] gap-4'

  return (
    <div className="mt-6">
      {/* Compteur */}
      <div className="flex items-center justify-between mb-3">
        <p className="text-sm text-ink-light">
          {fichiers.length} fichier{fichiers.length > 1 ? 's' : ''} selectionne{fichiers.length > 1 ? 's' : ''}
          <span className="text-ink-muted"> / {maxFichiers} max</span>
        </p>
        {fichiers.length > 1 && (
          <button
            type="button"
            className="text-xs text-ink-muted hover:text-coral transition-colors"
            onClick={() => {
              for (let i = fichiers.length - 1; i >= 0; i--) {
                onSupprimer(i)
              }
            }}
          >
            Tout supprimer
          </button>
        )}
      </div>

      <div className={gridClass}>
        {fichiers.map((fichier, index) => (
          <CarteFichier
            key={`${fichier.name}-${fichier.lastModified}-${index}`}
            fichier={{ fichier, urlPreview: previews.get(fichier) || null }}
            urlPreview={previews.get(fichier) || null}
            onSupprimer={() => onSupprimer(index)}
            compact={compact}
          />
        ))}
        <BoutonAjouter onClick={onAjouter} desactive={!peutAjouter} compact={compact} />
      </div>
    </div>
  )
}
