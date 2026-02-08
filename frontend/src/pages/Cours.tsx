import { useEffect, useState, useCallback } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import {
  listerCours,
  obtenirCours,
  supprimerCours,
  mettreAJourCours,
  obtenirRessourcesCours,
  genererRessources,
  obtenirFichesCours,
  genererFiches,
  genererQuiz,
  type Cours,
  type Ressource,
  type ZoneIncertaine,
  type Fiche,
} from '../services/api'
import ProcessingSection from '../components/ProcessingSection'

// Icônes des matières
const iconesMatiere: Record<string, string> = {
  mathematiques: '📐',
  francais: '📖',
  histoire: '🏛️',
  geographie: '🗺️',
  sciences: '🔬',
  anglais: '🇬🇧',
  physique: '⚛️',
  chimie: '🧪',
  svt: '🌿',
  ses: '📊',
  philosophie: '🤔',
}

function getIconeMatiere(matiere: string | undefined): string {
  if (!matiere) return '📚'
  return iconesMatiere[matiere.toLowerCase()] || '📚'
}

// Format de date
function formatDate(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleDateString('fr-FR', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  })
}

// Skeleton pour les cours
function CoursSkeleton() {
  return (
    <div className="bg-white rounded-lg p-8 animate-pulse">
      <div className="flex items-start gap-6">
        <div className="w-14 h-14 rounded-md bg-cream" />
        <div className="flex-1">
          <div className="h-6 bg-cream rounded w-3/4 mb-2" />
          <div className="h-4 bg-cream rounded w-1/2 mb-3" />
          <div className="flex gap-6">
            <div className="h-4 bg-cream rounded w-20" />
            <div className="h-4 bg-cream rounded w-20" />
          </div>
        </div>
      </div>
    </div>
  )
}

// Carte de cours dans la liste
function CoursCard({
  cours,
  onSupprimer,
  onOuvrir,
}: {
  cours: Cours
  onSupprimer: (id: string) => void
  onOuvrir: (id: string) => void
}) {
  const [confirmationSuppression, setConfirmationSuppression] = useState(false)

  const handleSupprimer = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (confirmationSuppression) {
      onSupprimer(cours.id)
      setConfirmationSuppression(false)
    } else {
      setConfirmationSuppression(true)
    }
  }

  const handleClick = () => {
    if (!confirmationSuppression) {
      onOuvrir(cours.id)
    }
  }

  return (
    <div
      data-testid="cours-card"
      onClick={handleClick}
      className="bg-white rounded-lg p-8 shadow-sm hover:shadow-md transition-shadow border border-cream-dark cursor-pointer"
    >
      <div className="flex items-start gap-6">
        <div className="w-14 h-14 rounded-md bg-cream flex items-center justify-center text-2xl flex-shrink-0">
          {getIconeMatiere(cours.matiere)}
        </div>
        <div className="flex-1 min-w-0">
          <h2 className="font-display text-lg font-semibold text-ink mb-2 truncate">
            {cours.titre || 'Cours sans titre'}
          </h2>
          <p className="text-sm text-ink-muted mb-4 capitalize">
            {cours.matiere || 'Matière non définie'}
          </p>
          <div className="flex items-center gap-8 text-sm text-ink-light">
            <span>Créé le {formatDate(cours.dateCreation)}</span>
            {cours.confiance > 0 && (
              <span className="flex items-center gap-1">
                OCR: {Math.round(cours.confiance * 100)}%
              </span>
            )}
          </div>
        </div>
        <div className="flex flex-col gap-2" onClick={(e) => e.stopPropagation()}>
          <Link
            to={`/fiches?cours=${cours.id}`}
            className="px-6 py-2 bg-coral text-white rounded-full text-sm font-medium hover:bg-coral-dark transition-colors text-center"
          >
            Fiches
          </Link>
          <Link
            to={`/quiz?cours=${cours.id}`}
            className="px-6 py-2 bg-teal text-white rounded-full text-sm font-medium hover:bg-teal-light transition-colors text-center"
          >
            Quiz
          </Link>
          <Link
            to={`/mindmap?cours=${cours.id}`}
            className="px-6 py-2 bg-white text-ink border border-ink rounded-full text-sm font-medium hover:bg-cream transition-colors text-center"
          >
            Mindmap
          </Link>
        </div>
      </div>

      {/* Aperçu du texte OCR */}
      {cours.texteOCR && (
        <div className="mt-6 pt-6 border-t border-cream-dark">
          <p className="text-sm text-ink-light line-clamp-2">
            {cours.texteOCR.slice(0, 200)}
            {cours.texteOCR.length > 200 ? '...' : ''}
          </p>
        </div>
      )}

      {/* Actions */}
      <div className="mt-6 pt-6 border-t border-cream-dark flex justify-end" onClick={(e) => e.stopPropagation()}>
        {confirmationSuppression ? (
          <div className="flex items-center gap-4">
            <span className="text-sm text-ink-muted">Confirmer la suppression ?</span>
            <button
              onClick={handleSupprimer}
              className="px-6 py-1.5 bg-red-500 text-white rounded-full text-sm font-medium hover:bg-red-600 transition-colors"
            >
              Supprimer
            </button>
            <button
              onClick={(e) => { e.stopPropagation(); setConfirmationSuppression(false) }}
              className="px-6 py-1.5 bg-cream text-ink rounded-full text-sm font-medium hover:bg-cream-dark transition-colors"
            >
              Annuler
            </button>
          </div>
        ) : (
          <button
            onClick={handleSupprimer}
            className="text-sm text-ink-muted hover:text-red-500 transition-colors"
          >
            Supprimer
          </button>
        )}
      </div>
    </div>
  )
}

// Icônes des types de ressources
const iconesTypeRessource: Record<string, string> = {
  video: '🎬',
  article: '📄',
  exercice: '✏️',
  cours: '📖',
  autre: '🔗',
}

function getIconeTypeRessource(type: string): string {
  return iconesTypeRessource[type] || '🔗'
}

// Vue détaillée d'un cours avec mode édition
function CoursDetail({ coursId, onRetour }: { coursId: string; onRetour: () => void }) {
  const [cours, setCours] = useState<Cours | null>(null)
  const [chargement, setChargement] = useState(true)
  const [erreur, setErreur] = useState<string | null>(null)

  // Mode édition
  const [modeEdition, setModeEdition] = useState(false)
  const [texteEdite, setTexteEdite] = useState('')
  const [zonesIncertainesEditees, setZonesIncertainesEditees] = useState<ZoneIncertaine[]>([])
  const [sauvegarde, setSauvegarde] = useState(false)
  const [imageSelectionnee, setImageSelectionnee] = useState(0)

  // State pour les ressources
  const [ressources, setRessources] = useState<Ressource[]>([])
  const [chargementRessources, setChargementRessources] = useState(false)
  const [generationRessources, setGenerationRessources] = useState(false)
  const [erreurRessources, setErreurRessources] = useState<string | null>(null)

  // State pour les fiches
  const [fiches, setFiches] = useState<Fiche[]>([])
  const [chargementFiches, setChargementFiches] = useState(false)
  const [generationFiches, setGenerationFiches] = useState(false)
  const [erreurFiches, setErreurFiches] = useState<string | null>(null)

  // State pour le quiz
  const [generationQuiz, setGenerationQuiz] = useState(false)
  const [erreurQuiz, setErreurQuiz] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    obtenirCours(coursId)
      .then((c) => {
        if (!cancelled) {
          setCours(c)
          setTexteEdite(c.texteOCR || '')
          setZonesIncertainesEditees(c.zonesIncertaines || [])
          setChargement(false)
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setErreur(err.message)
          setChargement(false)
        }
      })

    return () => {
      cancelled = true
    }
  }, [coursId])

  // Charger les ressources existantes
  useEffect(() => {
    let cancelled = false

    if (cours) {
      setChargementRessources(true)
      obtenirRessourcesCours(coursId)
        .then((res) => {
          if (!cancelled && res.succes) {
            setRessources(res.ressources || [])
          }
        })
        .catch(() => {
          // Pas de ressources existantes, ce n'est pas une erreur
        })
        .finally(() => {
          if (!cancelled) {
            setChargementRessources(false)
          }
        })
    }

    return () => {
      cancelled = true
    }
  }, [cours, coursId])

  // Activer le mode édition
  const activerEdition = () => {
    if (cours) {
      setTexteEdite(cours.texteOCR || '')
      setZonesIncertainesEditees(cours.zonesIncertaines || [])
      setModeEdition(true)
    }
  }

  // Annuler l'édition
  const annulerEdition = () => {
    if (cours) {
      setTexteEdite(cours.texteOCR || '')
      setZonesIncertainesEditees(cours.zonesIncertaines || [])
    }
    setModeEdition(false)
  }

  // Sauvegarder les modifications
  const sauvegarderModifications = async () => {
    if (!cours) return

    setSauvegarde(true)
    try {
      const coursModifie = await mettreAJourCours(coursId, {
        texteOCR: texteEdite,
        zonesIncertaines: zonesIncertainesEditees,
      })
      setCours(coursModifie)
      setModeEdition(false)
    } catch (err) {
      setErreur(err instanceof Error ? err.message : 'Erreur lors de la sauvegarde')
    } finally {
      setSauvegarde(false)
    }
  }

  // Corriger une zone incertaine
  const corrigerZone = (index: number, nouveauTexte: string) => {
    const nouvellesZones = [...zonesIncertainesEditees]
    if (nouveauTexte === '') {
      // Supprimer la zone si le texte est vide
      nouvellesZones.splice(index, 1)
    } else {
      nouvellesZones[index] = { ...nouvellesZones[index], texte: nouveauTexte }
    }
    setZonesIncertainesEditees(nouvellesZones)
  }

  // Charger les fiches existantes
  useEffect(() => {
    let cancelled = false

    if (cours) {
      setChargementFiches(true)
      obtenirFichesCours(coursId)
        .then((res) => {
          if (!cancelled && res.succes) {
            setFiches(res.fiches || [])
          }
        })
        .catch(() => {
          // Pas de fiches existantes, ce n'est pas une erreur
        })
        .finally(() => {
          if (!cancelled) {
            setChargementFiches(false)
          }
        })
    }

    return () => {
      cancelled = true
    }
  }, [cours, coursId])

  // Générer des fiches
  const handleGenererFiches = async () => {
    setGenerationFiches(true)
    setErreurFiches(null)
    try {
      const res = await genererFiches(coursId, { nombreFiches: 5 })
      if (res.succes) {
        setFiches(res.fiches)
      } else {
        setErreurFiches(res.erreur?.message || 'Erreur lors de la génération')
      }
    } catch (err) {
      setErreurFiches(err instanceof Error ? err.message : 'Erreur inconnue')
    } finally {
      setGenerationFiches(false)
    }
  }

  // Générer un quiz
  const handleGenererQuiz = async () => {
    setGenerationQuiz(true)
    setErreurQuiz(null)
    try {
      const res = await genererQuiz(coursId, { nombreQuestions: 5 })
      if (res.succes && res.quiz) {
        // Rediriger vers le quiz
        window.location.href = `/quiz?id=${res.quiz.id}`
      } else {
        setErreurQuiz(res.erreur?.message || 'Erreur lors de la génération')
      }
    } catch (err) {
      setErreurQuiz(err instanceof Error ? err.message : 'Erreur inconnue')
    } finally {
      setGenerationQuiz(false)
    }
  }

  // Générer des ressources
  const handleGenererRessources = async () => {
    setGenerationRessources(true)
    setErreurRessources(null)
    try {
      const res = await genererRessources(coursId)
      if (res.succes) {
        setRessources(res.ressources)
      } else {
        setErreurRessources(res.erreur?.message || 'Erreur lors de la génération')
      }
    } catch (err) {
      setErreurRessources(err instanceof Error ? err.message : 'Erreur inconnue')
    } finally {
      setGenerationRessources(false)
    }
  }

  // Générer l'URL de l'image
  const getImageUrl = (nomFichier: string) => {
    return `/api/cours/${coursId}/images/${encodeURIComponent(nomFichier)}`
  }

  if (chargement) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <ProcessingSection message="Chargement du cours..." />
      </div>
    )
  }

  if (erreur || !cours) {
    return (
      <div className="text-center py-12">
        <div className="text-5xl mb-6">😕</div>
        <h1 className="font-display text-2xl font-semibold text-ink mb-4">
          Cours introuvable
        </h1>
        <p className="text-ink-light mb-8">{erreur || 'Ce cours n\'existe pas.'}</p>
        <button
          onClick={onRetour}
          className="inline-flex items-center gap-4 px-8 py-3 bg-coral text-white rounded-full font-medium hover:bg-coral-dark transition-colors"
        >
          ← Retour aux cours
        </button>
      </div>
    )
  }

  // Utiliser les images sauvegardées (pas les noms de fichiers originaux)
  const images = cours.images || []

  // Déplacer une image vers le haut ou le bas
  const deplacerImage = async (index: number, direction: 'haut' | 'bas') => {
    if (!cours) return
    const nouvellesImages = [...images]
    const newIndex = direction === 'haut' ? index - 1 : index + 1
    if (newIndex < 0 || newIndex >= nouvellesImages.length) return

    // Échanger les positions
    ;[nouvellesImages[index], nouvellesImages[newIndex]] = [
      nouvellesImages[newIndex],
      nouvellesImages[index],
    ]

    try {
      const coursModifie = await mettreAJourCours(coursId, { images: nouvellesImages })
      setCours(coursModifie)
      // Mettre à jour l'image sélectionnée si nécessaire
      if (imageSelectionnee === index) {
        setImageSelectionnee(newIndex)
      } else if (imageSelectionnee === newIndex) {
        setImageSelectionnee(index)
      }
    } catch (err) {
      console.error('Erreur lors du déplacement:', err)
    }
  }

  // Supprimer une image
  const supprimerImage = async (nomFichier: string) => {
    if (!cours) return
    try {
      const response = await fetch(`/api/cours/${coursId}/images/${encodeURIComponent(nomFichier)}`, {
        method: 'DELETE',
      })
      if (response.ok) {
        const data = await response.json()
        setCours({ ...cours, images: data.images })
        if (imageSelectionnee >= data.images.length) {
          setImageSelectionnee(Math.max(0, data.images.length - 1))
        }
      }
    } catch (err) {
      console.error('Erreur lors de la suppression:', err)
    }
  }

  // Ajouter une image
  const ajouterImage = async (file: File) => {
    if (!cours) return
    const formData = new FormData()
    formData.append('image', file)

    try {
      const response = await fetch(`/api/cours/${coursId}/images`, {
        method: 'POST',
        body: formData,
      })
      if (response.ok) {
        const data = await response.json()
        setCours({ ...cours, images: data.images })
      }
    } catch (err) {
      console.error('Erreur lors de l\'ajout:', err)
    }
  }

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      ajouterImage(file)
      e.target.value = '' // Reset input
    }
  }

  return (
    <div>
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <button
          onClick={onRetour}
          className="flex items-center gap-2 text-ink-light hover:text-ink transition-colors"
        >
          ← Retour aux cours
        </button>
        <div className="flex gap-4">
          {modeEdition ? (
            <>
              <button
                onClick={annulerEdition}
                disabled={sauvegarde}
                className="px-6 py-2 bg-cream text-ink rounded-full text-sm font-medium hover:bg-cream-dark transition-colors disabled:opacity-50"
              >
                Annuler
              </button>
              <button
                onClick={sauvegarderModifications}
                disabled={sauvegarde}
                className="px-6 py-2 bg-success text-white rounded-full text-sm font-medium hover:bg-success/90 transition-colors disabled:opacity-50"
              >
                {sauvegarde ? 'Sauvegarde...' : 'Enregistrer'}
              </button>
            </>
          ) : (
            <>
              <button
                onClick={activerEdition}
                className="px-6 py-2 bg-gold text-white rounded-full text-sm font-medium hover:bg-gold/90 transition-colors"
              >
                Modifier
              </button>
              <Link
                to={`/fiches?cours=${cours.id}`}
                className="px-6 py-2 bg-coral text-white rounded-full text-sm font-medium hover:bg-coral-dark transition-colors"
              >
                Réviser les fiches
              </Link>
              <Link
                to={`/quiz?cours=${cours.id}`}
                className="px-6 py-2 bg-teal text-white rounded-full text-sm font-medium hover:bg-teal-light transition-colors"
              >
                Lancer un quiz
              </Link>
            </>
          )}
        </div>
      </div>

      {/* Contenu */}
      <div className="bg-white rounded-lg p-12 shadow-sm">
        <div className="flex items-start gap-8 mb-8">
          <div className="w-20 h-20 rounded-lg bg-cream flex items-center justify-center text-4xl flex-shrink-0">
            {getIconeMatiere(cours.matiere)}
          </div>
          <div>
            <h1 className="font-display text-2xl font-bold text-ink mb-2">
              {cours.titre || 'Cours sans titre'}
            </h1>
            <p className="text-ink-muted capitalize">{cours.matiere || 'Matière non définie'}</p>
            <p className="text-sm text-ink-light mt-4">
              Créé le {formatDate(cours.dateCreation)}
            </p>
          </div>
        </div>

        {/* Indicateur OCR */}
        {cours.confiance > 0 && (
          <div className="mb-8 p-6 bg-cream rounded-md">
            <div className="flex items-center justify-between text-sm mb-2">
              <span className="text-ink-muted">Confiance OCR</span>
              <span className="font-semibold text-ink">
                {Math.round(cours.confiance * 100)}%
              </span>
            </div>
            <div className="h-2 bg-cream-dark rounded-full overflow-hidden">
              <div
                className={`h-full rounded-full transition-all ${
                  cours.confiance >= 0.9
                    ? 'bg-success'
                    : cours.confiance >= 0.7
                    ? 'bg-gold'
                    : 'bg-coral'
                }`}
                style={{ width: `${cours.confiance * 100}%` }}
              />
            </div>
          </div>
        )}

        {/* Images avec texte superposé */}
        {(images.length > 0 || modeEdition) && (
          <div className="mb-8" data-testid="images-ocr">
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-semibold text-ink">Images du cours ({images.length})</h3>
              {modeEdition && (
                <label
                  data-testid="ajouter-image"
                  className="px-6 py-2 bg-teal text-white rounded-full text-sm font-medium hover:bg-teal-light transition-colors cursor-pointer"
                >
                  + Ajouter une image
                  <input
                    type="file"
                    accept="image/*"
                    onChange={handleFileSelect}
                    className="hidden"
                  />
                </label>
              )}
            </div>

            {/* Sélecteur d'image avec contrôles de réordonnancement */}
            {images.length > 0 && (
              <div className="flex gap-6 mb-6 overflow-x-auto py-2">
                {images.map((img, idx) => (
                  <div key={idx} className="relative flex-shrink-0">
                    <button
                      data-testid="image-vignette"
                      onClick={() => setImageSelectionnee(idx)}
                      className={`w-24 h-24 rounded-lg overflow-hidden border-3 transition-all ${
                        imageSelectionnee === idx
                          ? 'border-coral ring-2 ring-coral/30 shadow-lg'
                          : 'border-cream-dark hover:border-coral/50'
                      }`}
                    >
                      <img
                        src={getImageUrl(img)}
                        alt={`Page ${idx + 1}`}
                        className="w-full h-full object-cover"
                      />
                    </button>
                    {/* Numéro de page en bas de la vignette */}
                    <div className="absolute -bottom-2 left-1/2 -translate-x-1/2">
                      <span className={`text-xs font-semibold px-2.5 py-1 rounded-full shadow-sm ${
                        imageSelectionnee === idx ? 'bg-coral text-white' : 'bg-ink text-white'
                      }`}>
                        {idx + 1}
                      </span>
                    </div>
                    {/* Bouton supprimer - uniquement en mode édition */}
                    {modeEdition && (
                      <button
                        data-testid="supprimer-image"
                        onClick={() => supprimerImage(img)}
                        className="absolute -top-2 -right-2 w-6 h-6 bg-red-500 text-white rounded-full text-sm flex items-center justify-center hover:bg-red-600 transition-colors shadow-md"
                        title="Supprimer"
                      >
                        ×
                      </button>
                    )}
                  </div>
                ))}
              </div>
            )}

            {/* Boutons de réordonnancement - sous les vignettes */}
            {images.length > 1 && (
              <div className="flex items-center justify-center gap-4 mb-6">
                <button
                  onClick={() => deplacerImage(imageSelectionnee, 'haut')}
                  disabled={imageSelectionnee === 0}
                  className="flex items-center gap-2 px-4 py-2 bg-cream text-ink rounded-full text-sm font-medium hover:bg-cream-dark transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
                >
                  ← Précédent
                </button>
                <span className="text-sm text-ink-muted">
                  Page {imageSelectionnee + 1} sur {images.length}
                </span>
                <button
                  onClick={() => deplacerImage(imageSelectionnee, 'bas')}
                  disabled={imageSelectionnee === images.length - 1}
                  className="flex items-center gap-2 px-4 py-2 bg-cream text-ink rounded-full text-sm font-medium hover:bg-cream-dark transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
                >
                  Suivant →
                </button>
              </div>
            )}

            {/* Image principale avec indicateur de page */}
            {images.length > 0 && (
              <div className="relative rounded-lg overflow-hidden bg-ink-lighter">
                {/* Badge de page en haut à gauche */}
                <div className="absolute top-4 left-4 z-10">
                  <span className="bg-coral text-white text-sm font-semibold px-3 py-1.5 rounded-full shadow-lg">
                    Page {imageSelectionnee + 1} / {images.length}
                  </span>
                </div>
                <img
                  src={getImageUrl(images[imageSelectionnee])}
                  alt={`Page ${imageSelectionnee + 1}`}
                  className="w-full"
                />
                {modeEdition && (
                  <div
                    data-testid="texte-overlay"
                    className="absolute inset-0 bg-white/80 p-6 overflow-y-auto"
                  >
                    <p className="text-xs text-ink-muted mb-2">
                      Texte superposé - Modifiez ci-dessous pour corriger les erreurs OCR
                    </p>
                  </div>
                )}
                {/* Indicateur de correspondance texte-image */}
                <div className="absolute bottom-4 left-4 right-4 z-10">
                  <div className="bg-ink/80 backdrop-blur-sm text-white text-xs px-4 py-2 rounded-lg">
                    💡 Le texte ci-dessous correspond à l'ensemble des {images.length} page{images.length > 1 ? 's' : ''} scannée{images.length > 1 ? 's' : ''}
                  </div>
                </div>
              </div>
            )}

            {/* Message si pas d'images */}
            {images.length === 0 && modeEdition && (
              <div className="text-center py-8 bg-cream rounded-lg">
                <div className="text-3xl mb-4">📷</div>
                <p className="text-sm text-ink-muted">
                  Aucune image pour ce cours.
                  <br />
                  Cliquez sur "Ajouter une image" pour en ajouter.
                </p>
              </div>
            )}
          </div>
        )}

        {/* Zones incertaines */}
        {zonesIncertainesEditees.length > 0 && (
          <div className="mb-8 p-6 bg-gold/10 border border-gold rounded-md">
            <h3 className="font-semibold text-ink mb-4">
              ⚠️ Zones incertaines ({zonesIncertainesEditees.length})
            </h3>
            <ul className="space-y-4">
              {zonesIncertainesEditees.map((zone, i) => (
                <li
                  key={i}
                  data-testid="zone-incertaine"
                  className="flex items-start gap-2"
                >
                  <span className="text-gold mt-1">•</span>
                  {modeEdition ? (
                    <div className="flex-1">
                      <input
                        type="text"
                        data-testid="correction-zone"
                        value={zone.texte}
                        onChange={(e) => corrigerZone(i, e.target.value)}
                        className="w-full px-2 py-1 border border-gold rounded text-sm"
                        placeholder="Corriger ou laisser vide pour supprimer"
                      />
                      <span className="text-xs text-ink-muted">{zone.raison}</span>
                    </div>
                  ) : (
                    <span className="text-sm text-ink-light">
                      "{zone.texte}" - {zone.raison}
                    </span>
                  )}
                </li>
              ))}
            </ul>
          </div>
        )}

        {/* Texte OCR */}
        <div>
          <div className="flex items-center gap-4 mb-4">
            <h3 className="font-semibold text-ink">Contenu du cours</h3>
            {images.length > 0 && (
              <span className="text-xs text-ink-muted bg-cream px-3 py-1 rounded-full">
                📄 Extrait de {images.length} page{images.length > 1 ? 's' : ''}
              </span>
            )}
          </div>
          {modeEdition ? (
            <textarea
              data-testid="texte-ocr-editable"
              value={texteEdite}
              onChange={(e) => setTexteEdite(e.target.value)}
              className="w-full h-[400px] p-6 bg-cream rounded-md text-sm text-ink font-body border border-cream-dark focus:border-coral focus:outline-none resize-none"
              placeholder="Entrez le texte du cours..."
            />
          ) : (
            <div className="p-6 bg-cream rounded-md max-h-[400px] overflow-y-auto">
              <pre className="whitespace-pre-wrap text-sm text-ink-light font-body">
                {cours.texteOCR || 'Aucun contenu textuel disponible.'}
              </pre>
            </div>
          )}
        </div>

        {/* Section Génération IA */}
        {!modeEdition && (
          <div className="mt-8 pt-8 border-t border-cream-dark">
            <h3 className="font-semibold text-ink mb-6">🤖 Générer avec l'IA</h3>

            {/* Boutons de génération */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
              {/* Générer des fiches */}
              <div className="p-6 bg-coral/5 border border-coral/20 rounded-lg">
                <div className="flex items-center gap-4 mb-4">
                  <span className="text-2xl">📝</span>
                  <h4 className="font-medium text-ink">Fiches de révision</h4>
                </div>
                <p className="text-sm text-ink-muted mb-6">
                  Génère automatiquement des fiches question/réponse à partir du contenu du cours.
                </p>
                {erreurFiches && (
                  <div className="mb-4 p-4 bg-red-50 text-red-600 rounded-md text-xs">
                    {erreurFiches}
                  </div>
                )}
                <button
                  onClick={handleGenererFiches}
                  disabled={generationFiches || chargementFiches}
                  className="w-full px-6 py-2 bg-coral text-white rounded-full text-sm font-medium hover:bg-coral-dark transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {generationFiches ? '⏳ Génération en cours...' : fiches.length > 0 ? `🔄 Régénérer (${fiches.length} fiches)` : '✨ Générer des fiches'}
                </button>
              </div>

              {/* Générer un quiz */}
              <div className="p-6 bg-teal/5 border border-teal/20 rounded-lg">
                <div className="flex items-center gap-4 mb-4">
                  <span className="text-2xl">🎯</span>
                  <h4 className="font-medium text-ink">Quiz interactif</h4>
                </div>
                <p className="text-sm text-ink-muted mb-6">
                  Crée un quiz à choix multiples pour tester tes connaissances sur ce cours.
                </p>
                {erreurQuiz && (
                  <div className="mb-4 p-4 bg-red-50 text-red-600 rounded-md text-xs">
                    {erreurQuiz}
                  </div>
                )}
                <button
                  onClick={handleGenererQuiz}
                  disabled={generationQuiz}
                  className="w-full px-6 py-2 bg-teal text-white rounded-full text-sm font-medium hover:bg-teal-light transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {generationQuiz ? '⏳ Création du quiz...' : '✨ Créer un quiz'}
                </button>
              </div>
            </div>

            {/* Aperçu des fiches générées */}
            {(chargementFiches || generationFiches) && (
              <div className="flex items-center justify-center py-8">
                <ProcessingSection
                  message={generationFiches ? 'Génération des fiches en cours...' : 'Chargement...'}
                />
              </div>
            )}

            {!chargementFiches && !generationFiches && fiches.length > 0 && (
              <div className="mb-8">
                <div className="flex items-center justify-between mb-4">
                  <h4 className="font-medium text-ink">📚 Fiches générées ({fiches.length})</h4>
                  <Link
                    to={`/fiches?cours=${coursId}`}
                    className="text-sm text-coral hover:text-coral-dark transition-colors"
                  >
                    Voir toutes les fiches →
                  </Link>
                </div>
                <div className="space-y-4 max-h-[300px] overflow-y-auto">
                  {fiches.slice(0, 3).map((fiche, index) => (
                    <div
                      key={fiche.id}
                      className="p-6 bg-cream rounded-md"
                    >
                      <div className="flex items-start gap-4">
                        <span className="text-coral font-bold text-sm">Q{index + 1}</span>
                        <div className="flex-1">
                          <p className="text-sm font-medium text-ink mb-2">{fiche.question}</p>
                          <p className="text-sm text-ink-light">{fiche.reponse.slice(0, 100)}{fiche.reponse.length > 100 ? '...' : ''}</p>
                        </div>
                        <span className={`text-xs px-2 py-1 rounded-full ${
                          fiche.difficulte === 'facile' ? 'bg-success/20 text-success' :
                          fiche.difficulte === 'moyen' ? 'bg-gold/20 text-gold' :
                          'bg-coral/20 text-coral'
                        }`}>
                          {fiche.difficulte}
                        </span>
                      </div>
                    </div>
                  ))}
                  {fiches.length > 3 && (
                    <p className="text-center text-sm text-ink-muted">
                      + {fiches.length - 3} autres fiches
                    </p>
                  )}
                </div>
              </div>
            )}
          </div>
        )}

        {/* Ressources complémentaires */}
        {!modeEdition && (
          <div className="mt-8 pt-8 border-t border-cream-dark">
            <div className="flex items-center justify-between mb-6">
              <h3 className="font-semibold text-ink">Ressources complémentaires</h3>
              {ressources.length === 0 && !chargementRessources && (
                <button
                  onClick={handleGenererRessources}
                  disabled={generationRessources}
                  className="px-6 py-2 bg-teal text-white rounded-full text-sm font-medium hover:bg-teal-light transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {generationRessources ? 'Génération...' : 'Générer des ressources'}
                </button>
              )}
            </div>

            {/* Avertissement */}
            {ressources.length > 0 && (
              <div className="mb-6 p-4 bg-gold/10 border border-gold rounded-md text-sm text-ink-muted">
                Les liens suggérés sont générés par IA et doivent être vérifiés avant utilisation.
              </div>
            )}

            {/* Erreur */}
            {erreurRessources && (
              <div className="mb-6 p-4 bg-red-50 text-red-600 rounded-md text-sm">
                {erreurRessources}
              </div>
            )}

            {/* Chargement */}
            {(chargementRessources || generationRessources) && (
              <div className="flex items-center justify-center py-8">
                <ProcessingSection
                  message={generationRessources ? 'Recherche de ressources en cours...' : 'Chargement...'}
                />
              </div>
            )}

            {/* Liste des ressources */}
            {!chargementRessources && !generationRessources && ressources.length > 0 && (
              <div className="space-y-4">
                {ressources.map((ressource) => (
                  <div
                    key={ressource.id}
                    className="flex items-start gap-6 p-6 bg-cream rounded-md hover:bg-cream-dark transition-colors"
                  >
                    <div className="w-10 h-10 rounded-md bg-white flex items-center justify-center text-xl flex-shrink-0">
                      {getIconeTypeRessource(ressource.type)}
                    </div>
                    <div className="flex-1 min-w-0">
                      <h4 className="font-medium text-ink mb-2">{ressource.titre}</h4>
                      {ressource.description && (
                        <p className="text-sm text-ink-light mb-2">{ressource.description}</p>
                      )}
                      {ressource.url && (
                        <a
                          href={ressource.url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-sm text-teal hover:text-teal-light transition-colors inline-flex items-center gap-1"
                        >
                          Ouvrir le lien
                          <span aria-hidden="true">↗</span>
                        </a>
                      )}
                    </div>
                    <span className="text-xs text-ink-muted capitalize bg-white px-2 py-1 rounded-full">
                      {ressource.type}
                    </span>
                  </div>
                ))}
              </div>
            )}

            {/* Aucune ressource */}
            {!chargementRessources && !generationRessources && ressources.length === 0 && (
              <div className="text-center py-8 text-ink-muted">
                <div className="text-3xl mb-4">🔍</div>
                <p className="text-sm">
                  Aucune ressource complémentaire disponible.
                  <br />
                  Cliquez sur "Générer des ressources" pour en trouver.
                </p>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}

// Page principale des cours
export default function CoursPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const coursId = searchParams.get('id')

  const [cours, setCours] = useState<Cours[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [chargement, setChargement] = useState(true)
  const [erreur, setErreur] = useState<string | null>(null)

  const limite = 10

  const chargerCours = useCallback(async () => {
    try {
      setChargement(true)
      setErreur(null)
      const res = await listerCours(page, limite)
      setCours(res.cours)
      setTotal(res.total)
    } catch (err) {
      setErreur(err instanceof Error ? err.message : 'Erreur inconnue')
    } finally {
      setChargement(false)
    }
  }, [page])

  useEffect(() => {
    if (!coursId) {
      chargerCours()
    }
  }, [chargerCours, coursId])

  const handleSupprimer = async (id: string) => {
    try {
      await supprimerCours(id)
      setCours((prev) => prev.filter((c) => c.id !== id))
      setTotal((prev) => prev - 1)
    } catch (err) {
      console.error('Erreur lors de la suppression:', err)
    }
  }

  const handleRetour = () => {
    setSearchParams({})
  }

  // Vue détail
  if (coursId) {
    return <CoursDetail coursId={coursId} onRetour={handleRetour} />
  }

  // Vue liste
  const totalPages = Math.ceil(total / limite)

  return (
    <div>
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="font-display text-3xl font-bold text-ink mb-2">
            Mes cours
          </h1>
          <p className="text-ink-light">
            {total} cours enregistré{total > 1 ? 's' : ''}
          </p>
        </div>
        <Link
          to="/scanner"
          className="inline-flex items-center gap-4 px-8 py-3 bg-coral text-white rounded-full font-semibold hover:bg-coral-dark transition-colors"
        >
          <span role="img" aria-label="Scanner">📸</span>
          Scanner un cours
        </Link>
      </div>

      {/* Erreur */}
      {erreur && (
        <div className="bg-red-50 text-red-600 rounded-lg p-6 mb-8">
          <p className="font-medium">Erreur de chargement</p>
          <p className="text-sm">{erreur}</p>
        </div>
      )}

      {/* Liste des cours */}
      {chargement ? (
        <div className="space-y-6">
          <CoursSkeleton />
          <CoursSkeleton />
          <CoursSkeleton />
        </div>
      ) : cours.length === 0 ? (
        <div className="bg-white rounded-lg p-12 text-center">
          <div className="text-5xl mb-6" role="img" aria-label="Livres">📚</div>
          <h2 className="font-display text-xl font-semibold text-ink mb-4">
            Aucun cours enregistré
          </h2>
          <p className="text-ink-light mb-8 max-w-md mx-auto">
            Commencez par scanner un cours pour générer des fiches de révision et des quiz interactifs.
          </p>
          <Link
            to="/scanner"
            className="inline-flex items-center gap-2 bg-coral text-white px-8 py-3 rounded-full font-semibold no-underline transition-all hover:bg-coral-dark hover:-translate-y-0.5"
          >
            <span role="img" aria-label="Scanner">📸</span>
            Scanner mon premier cours
          </Link>
        </div>
      ) : (
        <>
          <div className="space-y-6">
            {cours.map((c) => (
              <CoursCard
                key={c.id}
                cours={c}
                onSupprimer={handleSupprimer}
                onOuvrir={(id) => setSearchParams({ id })}
              />
            ))}
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex items-center justify-center gap-4 mt-8">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                className="px-6 py-2 rounded-full text-sm font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed bg-white border border-cream-dark hover:bg-cream"
              >
                ← Précédent
              </button>
              <span className="px-6 py-2 text-sm text-ink-muted">
                Page {page} sur {totalPages}
              </span>
              <button
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
                className="px-6 py-2 rounded-full text-sm font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed bg-white border border-cream-dark hover:bg-cream"
              >
                Suivant →
              </button>
            </div>
          )}
        </>
      )}
    </div>
  )
}
