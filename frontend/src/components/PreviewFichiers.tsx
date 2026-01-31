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
}: {
  fichier: FichierAvecPreview
  urlPreview: string | null
  onSupprimer: () => void
}) {
  return (
    <div className="bg-white rounded-md overflow-hidden relative aspect-[3/4] shadow-lg group">
      {urlPreview ? (
        <img
          src={urlPreview}
          alt={fichier.fichier.name}
          className="w-full h-full object-cover"
        />
      ) : (
        <div className="w-full h-full flex items-center justify-center bg-cream text-4xl">
          📄
        </div>
      )}

      {/* Bouton supprimer */}
      <button
        type="button"
        className="absolute top-2 right-2 w-7 h-7 rounded-full bg-ink text-white flex items-center justify-center text-lg opacity-0 group-hover:opacity-100 transition-opacity hover:bg-coral"
        onClick={onSupprimer}
        aria-label={`Supprimer ${fichier.fichier.name}`}
      >
        ×
      </button>

      {/* Nom du fichier */}
      <div className="absolute bottom-0 left-0 right-0 px-xs py-xs bg-gradient-to-t from-black/70 to-transparent">
        <p className="text-white text-xs truncate">{fichier.fichier.name}</p>
      </div>
    </div>
  )
}

function BoutonAjouter({ onClick, desactive }: { onClick: () => void; desactive: boolean }) {
  return (
    <button
      type="button"
      className={`
        flex flex-col items-center justify-center gap-xs bg-cream border-2 border-dashed border-ink-muted
        rounded-md aspect-[3/4] text-ink-muted font-medium transition-all
        ${desactive ? 'opacity-50 cursor-not-allowed' : 'hover:border-coral hover:text-coral cursor-pointer'}
      `}
      onClick={onClick}
      disabled={desactive}
      aria-label="Ajouter plus de fichiers"
    >
      <span className="text-2xl">+</span>
      Ajouter
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

  return (
    <div className="mt-lg">
      <div className="grid grid-cols-[repeat(auto-fill,minmax(140px,1fr))] gap-md">
        {fichiers.map((fichier, index) => (
          <CarteFichier
            key={`${fichier.name}-${fichier.lastModified}-${index}`}
            fichier={{ fichier, urlPreview: previews.get(fichier) || null }}
            urlPreview={previews.get(fichier) || null}
            onSupprimer={() => onSupprimer(index)}
          />
        ))}
        <BoutonAjouter onClick={onAjouter} desactive={!peutAjouter} />
      </div>
    </div>
  )
}
