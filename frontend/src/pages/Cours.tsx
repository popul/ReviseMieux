import { useEffect, useState, useCallback } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import {
  listerCours,
  obtenirCours,
  supprimerCours,
  mettreAJourCours,
  obtenirFichesCours,
  genererFiches,
  genererQuiz,
  getConceptsByCours,
  retraiterOCRCours,
  listerPlansRevision,
  listerPlansParCours,
  ajouterCoursAuPlan,
  type Cours,
  type ZoneIncertaine,
  type Fiche,
  type BlocTexteOCR,
  type BlocTexteParPage,
  type Concept,
  type ResumeCours,
  type PlanRevisionResume,
} from '../services/api'
import ProcessingSection from '../components/ProcessingSection'
import OverlayTexteOCR from '../components/OverlayTexteOCR'
import ReordonnerPages from '../components/ReordonnerPages'
import ConceptCard from '../components/ConceptCard'

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
        {/* Boutons desktop (colonne) */}
        <div className="hidden md:flex flex-col gap-2" onClick={(e) => e.stopPropagation()}>
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
          <Link
            to={`/examen-blanc?cours=${cours.id}`}
            className="px-6 py-2 bg-ink text-white rounded-full text-sm font-medium hover:bg-ink/80 transition-colors text-center"
          >
            Examen
          </Link>
        </div>
      </div>

      {/* Boutons mobile (ligne) */}
      <div className="flex md:hidden flex-wrap gap-2 mt-4" onClick={(e) => e.stopPropagation()}>
        <Link to={`/fiches?cours=${cours.id}`} className="px-4 py-1.5 bg-coral text-white rounded-full text-xs font-medium hover:bg-coral-dark transition-colors">Fiches</Link>
        <Link to={`/quiz?cours=${cours.id}`} className="px-4 py-1.5 bg-teal text-white rounded-full text-xs font-medium hover:bg-teal-light transition-colors">Quiz</Link>
        <Link to={`/mindmap?cours=${cours.id}`} className="px-4 py-1.5 bg-white text-ink border border-ink rounded-full text-xs font-medium hover:bg-cream transition-colors">Mindmap</Link>
        <Link to={`/examen-blanc?cours=${cours.id}`} className="px-4 py-1.5 bg-ink text-white rounded-full text-xs font-medium hover:bg-ink/80 transition-colors">Examen</Link>
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

// Vue détaillée d'un cours avec mode édition
function CoursDetail({ coursId, onRetour }: { coursId: string; onRetour: () => void }) {
  const [cours, setCours] = useState<Cours | null>(null)
  const [chargement, setChargement] = useState(true)
  const [erreur, setErreur] = useState<string | null>(null)

  // Mode édition
  const [modeEdition, setModeEdition] = useState(false)
  const [texteEdite, setTexteEdite] = useState('')
  const [zonesIncertainesEditees, setZonesIncertainesEditees] = useState<ZoneIncertaine[]>([])
  const [blocsTexteEdites, setBlocsTexteEdites] = useState<BlocTexteParPage[]>([])
  const [sauvegarde, setSauvegarde] = useState(false)
  const [imageSelectionnee, setImageSelectionnee] = useState(0)

  // State pour les fiches
  const [fiches, setFiches] = useState<Fiche[]>([])
  const [chargementFiches, setChargementFiches] = useState(false)
  const [generationFiches, setGenerationFiches] = useState(false)
  const [erreurFiches, setErreurFiches] = useState<string | null>(null)

  // State pour le quiz
  const [generationQuiz, setGenerationQuiz] = useState(false)
  const [erreurQuiz, setErreurQuiz] = useState<string | null>(null)

  // Concepts state
  const [concepts, setConcepts] = useState<Concept[]>([])
  const [chargementConcepts, setChargementConcepts] = useState(false)
  const [conceptSelectionne, setConceptSelectionne] = useState<Concept | null>(null)

  // Resume state
  const [resume, setResume] = useState<ResumeCours | null>(null)

  // Overlay édition directe
  const [overlayModifie, setOverlayModifie] = useState(false)

  // Re-OCR state
  const [reOCREnCours, setReOCREnCours] = useState(false)

  // Plans state
  const [plansMenuOuvert, setPlansMenuOuvert] = useState(false)
  const [tousLesPlans, setTousLesPlans] = useState<PlanRevisionResume[]>([])
  const [plansDuCours, setPlansDuCours] = useState<PlanRevisionResume[]>([])

  const handleOuvrirMenuPlans = async () => {
    if (plansMenuOuvert) {
      setPlansMenuOuvert(false)
      return
    }
    setPlansMenuOuvert(true)
    try {
      const [tous, duCours] = await Promise.all([
        listerPlansRevision(),
        listerPlansParCours(coursId),
      ])
      setTousLesPlans(tous)
      setPlansDuCours(duCours)
    } catch {
      // Ignorer
    }
  }

  const handleAjouterAuPlan = async (planId: string) => {
    try {
      await ajouterCoursAuPlan(planId, coursId)
      setPlansDuCours(prev => [...prev, tousLesPlans.find(p => p.id === planId)!])
    } catch {
      // Ignorer
    }
    setPlansMenuOuvert(false)
  }

  useEffect(() => {
    let cancelled = false

    obtenirCours(coursId)
      .then((c) => {
        if (!cancelled) {
          setCours(c)
          setTexteEdite(c.texteOCR || '')
          setZonesIncertainesEditees(c.zonesIncertaines || [])
          setBlocsTexteEdites(c.blocsTexte || [])
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

  // Activer le mode édition
  const activerEdition = () => {
    if (cours) {
      setTexteEdite(cours.texteOCR || '')
      setZonesIncertainesEditees(cours.zonesIncertaines || [])
      setBlocsTexteEdites(cours.blocsTexte || [])
      setModeEdition(true)
    }
  }

  // Annuler l'édition
  const annulerEdition = () => {
    if (cours) {
      setTexteEdite(cours.texteOCR || '')
      setZonesIncertainesEditees(cours.zonesIncertaines || [])
      setBlocsTexteEdites(cours.blocsTexte || [])
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
        blocsTexte: blocsTexteEdites.length > 0 ? blocsTexteEdites : undefined,
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

  // Charger les concepts du cours
  useEffect(() => {
    if (!coursId) return
    setChargementConcepts(true)
    getConceptsByCours(coursId)
      .then(res => setConcepts(res.concepts || []))
      .catch(() => setConcepts([]))
      .finally(() => setChargementConcepts(false))
  }, [coursId])

  // Parser le résumé depuis les données du cours
  useEffect(() => {
    if (cours?.resume) {
      setResume(cours.resume as unknown as ResumeCours)
    } else {
      setResume(null)
    }
  }, [cours])

  // Polling pour récupérer le résumé et les concepts générés automatiquement
  useEffect(() => {
    if (!cours?.texteOCR) return
    const resumeManquant = !resume
    const conceptsManquants = concepts.length === 0
    if (!resumeManquant && !conceptsManquants) return

    let tentatives = 0
    const maxTentatives = 6
    const timer = setInterval(async () => {
      tentatives++
      if (tentatives > maxTentatives) {
        clearInterval(timer)
        return
      }
      try {
        const coursMAJ = await obtenirCours(coursId)
        setCours(coursMAJ)
        if (coursMAJ.resume) {
          setResume(coursMAJ.resume as unknown as ResumeCours)
        }
        const resConcepts = await getConceptsByCours(coursId)
        if (resConcepts.concepts && resConcepts.concepts.length > 0) {
          setConcepts(resConcepts.concepts)
        }
        // Arrêter si les deux sont arrivés
        const aResume = !!coursMAJ.resume
        const aConcepts = resConcepts.concepts && resConcepts.concepts.length > 0
        if (aResume && aConcepts) {
          clearInterval(timer)
        }
      } catch {
        // Ignorer les erreurs de polling
      }
    }, 5000)

    return () => clearInterval(timer)
  }, [cours?.texteOCR, coursId]) // eslint-disable-line react-hooks/exhaustive-deps

  // Lire le paramètre concept dans l'URL pour deep-linking
  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const conceptId = params.get('concept')
    if (conceptId && concepts.length > 0) {
      const found = concepts.find(c => c.id === conceptId)
      if (found) setConceptSelectionne(found)
    }
  }, [concepts])

  // Scroll vers le concept surligné
  useEffect(() => {
    if (conceptSelectionne) {
      const el = document.getElementById('concept-highlight')
      if (el) el.scrollIntoView({ behavior: 'smooth', block: 'center' })
    }
  }, [conceptSelectionne])

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

  const handleReOCR = async () => {
    if (!coursId) return
    setReOCREnCours(true)
    try {
      const res = await retraiterOCRCours(coursId)
      if (res.succes && cours) {
        // Refresh course data
        const coursMAJ = await obtenirCours(coursId)
        setCours(coursMAJ)
      }
    } catch (err) {
      console.error('Erreur re-OCR:', err)
    } finally {
      setReOCREnCours(false)
    }
  }

  const handleConceptClick = (concept: Concept) => {
    setConceptSelectionne(prev => prev?.id === concept.id ? null : concept)
  }

  // Modifier un bloc de texte OCR
  const modifierBlocTexte = (pageIndex: number, blocIndex: number, nouveauTexte: string) => {
    const nouveauxBlocs = [...blocsTexteEdites]
    if (nouveauxBlocs[pageIndex]) {
      const blocsPage = [...nouveauxBlocs[pageIndex].blocs_texte]
      blocsPage[blocIndex] = { ...blocsPage[blocIndex], texte: nouveauTexte }
      nouveauxBlocs[pageIndex] = { ...nouveauxBlocs[pageIndex], blocs_texte: blocsPage }
      setBlocsTexteEdites(nouveauxBlocs)
      // Mettre a jour le texte OCR global (concatenation de tous les blocs)
      const texteComplet = nouveauxBlocs
        .flatMap((page) => page.blocs_texte.map((b) => b.texte))
        .join('\n')
      setTexteEdite(texteComplet)
      setOverlayModifie(true)
    }
  }

  // Obtenir les blocs de la page courante
  const getBlocsPageCourante = (): BlocTexteOCR[] => {
    const pageData = blocsTexteEdites.find((p) => p.page === imageSelectionnee)
    return pageData?.blocs_texte || []
  }

  // Reordonner les images via drag and drop
  const reordonnerImages = async (nouvelOrdre: string[]) => {
    if (!cours) return
    try {
      const coursModifie = await mettreAJourCours(coursId, { images: nouvelOrdre })
      setCours(coursModifie)
    } catch (err) {
      console.error('Erreur lors du reordonnancement:', err)
    }
  }

  // Générer l'URL de l'image
  const getImageUrl = (nomFichier: string) => {
    return `/api/cours/${coursId}/images/${encodeURIComponent(nomFichier)}`
  }

  // Helper pour surligner le texte du concept sélectionné
  const renderTexteAvecSurlignage = (texte: string, concept: Concept | null): React.ReactNode => {
    if (!concept?.positionDansCours || concept.positionDansCours.debut === undefined) {
      return texte
    }
    const { debut, fin } = concept.positionDansCours
    if (debut >= texte.length || fin > texte.length || debut >= fin) {
      return texte
    }
    return (
      <>
        {texte.slice(0, debut)}
        <mark className="bg-yellow-200 px-0.5 rounded" id="concept-highlight">{texte.slice(debut, fin)}</mark>
        {texte.slice(fin)}
      </>
    )
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
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-8">
        <button
          onClick={onRetour}
          className="flex items-center gap-2 text-ink-light hover:text-ink transition-colors"
        >
          ← Retour aux cours
        </button>
        <div className="flex flex-wrap gap-2">
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
              <Link
                to={`/examen-blanc?cours=${cours.id}`}
                className="px-6 py-2 bg-ink text-white rounded-full text-sm font-medium hover:bg-ink/80 transition-colors"
              >
                Examen blanc
              </Link>
              {/* Ajouter a un plan */}
              <div className="relative">
                <button
                  onClick={handleOuvrirMenuPlans}
                  className="px-6 py-2 bg-white text-teal border border-teal rounded-full text-sm font-medium hover:bg-teal/10 transition-colors"
                >
                  + Plan
                </button>
                {plansMenuOuvert && (
                  <div className="absolute right-0 top-full mt-2 w-64 max-w-[calc(100vw-2rem)] bg-white rounded-lg shadow-xl border border-cream-dark z-50">
                    <div className="p-2">
                      {tousLesPlans.length === 0 ? (
                        <div className="px-3 py-2 text-sm text-ink-muted">Aucun plan</div>
                      ) : (
                        tousLesPlans.map(plan => {
                          const dejaPresent = plansDuCours.some(p => p.id === plan.id)
                          return (
                            <button
                              key={plan.id}
                              onClick={() => !dejaPresent && handleAjouterAuPlan(plan.id)}
                              disabled={dejaPresent}
                              className={`w-full text-left px-3 py-2 rounded text-sm transition-colors ${
                                dejaPresent
                                  ? 'text-ink-muted bg-cream cursor-default'
                                  : 'hover:bg-cream text-ink'
                              }`}
                            >
                              <span className="mr-2">{plan.iconeMatiere}</span>
                              {plan.titre}
                              {dejaPresent && <span className="ml-2 text-xs text-success">✓</span>}
                            </button>
                          )
                        })
                      )}
                      <div className="border-t border-cream mt-1 pt-1">
                        <Link
                          to="/plans/nouveau"
                          className="block px-3 py-2 rounded text-sm text-teal hover:bg-cream transition-colors"
                        >
                          + Creer un nouveau plan
                        </Link>
                      </div>
                    </div>
                  </div>
                )}
              </div>
            </>
          )}
        </div>
      </div>

      {/* Contenu */}
      <div className="bg-white rounded-lg p-4 md:p-8 lg:p-12 shadow-sm">
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

        {/* Section Resume */}
        {!modeEdition && (
          <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6 mb-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-[#1A4D4D]">Résumé du cours</h3>
            </div>
            {!resume && cours.texteOCR && (
              <div className="flex items-center gap-3 text-gray-500">
                <div className="animate-spin h-5 w-5 border-2 border-[#E85D4C] border-t-transparent rounded-full" />
                <span>Résumé en cours de génération...</span>
              </div>
            )}
            {resume && (
              <div className="space-y-4">
                {resume.paragraphe && (
                  <p className="text-gray-700 leading-relaxed bg-[#FBF8F3] p-4 rounded-lg">{resume.paragraphe}</p>
                )}
                {resume.pointsCles && resume.pointsCles.length > 0 && (
                  <div>
                    <h4 className="text-sm font-semibold text-[#1A4D4D] mb-2">Points clés</h4>
                    <ul className="space-y-1">
                      {resume.pointsCles.map((point, i) => (
                        <li key={i} className="flex items-start gap-2 text-sm text-gray-700">
                          <span className="text-[#E85D4C] mt-0.5">&bull;</span>
                          <span>{point}</span>
                        </li>
                      ))}
                    </ul>
                  </div>
                )}
                {resume.structure && resume.structure.length > 0 && (
                  <div>
                    <h4 className="text-sm font-semibold text-[#1A4D4D] mb-2">Structure du cours</h4>
                    <div className="space-y-2">
                      {resume.structure.map((section, i) => (
                        <details key={i} className="border border-gray-200 rounded-lg">
                          <summary className="px-4 py-2 cursor-pointer font-medium text-sm text-[#1A4D4D] hover:bg-gray-50">
                            {section.titre}
                          </summary>
                          <p className="px-4 py-3 text-sm text-gray-600 border-t border-gray-100">{section.contenu}</p>
                        </details>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        )}

        {/* Section Concepts clés */}
        {!modeEdition && (
          <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6 mb-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-[#1A4D4D]">
                Concepts clés
                {concepts.length > 0 && (
                  <span className="ml-2 text-sm font-normal text-gray-400">({concepts.length})</span>
                )}
              </h3>
            </div>
            {concepts.length === 0 && !chargementConcepts && cours.texteOCR && (
              <div className="flex items-center gap-3 text-gray-500">
                <div className="animate-spin h-5 w-5 border-2 border-[#1A4D4D] border-t-transparent rounded-full" />
                <span>Extraction des concepts en cours...</span>
              </div>
            )}
            {chargementConcepts && (
              <div className="flex items-center gap-3 text-gray-500">
                <div className="animate-spin h-5 w-5 border-2 border-[#1A4D4D] border-t-transparent rounded-full" />
                <span>Chargement...</span>
              </div>
            )}
            {concepts.length > 0 && (
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                {concepts.map(concept => (
                  <ConceptCard
                    key={concept.id}
                    concept={concept}
                    onClick={handleConceptClick}
                    isSelected={conceptSelectionne?.id === concept.id}
                  />
                ))}
              </div>
            )}
          </div>
        )}

        {/* Images avec texte superpose */}
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

            {/* Selecteur d'images avec drag and drop */}
            {images.length > 0 && (
              <div className="mb-6">
                <ReordonnerPages
                  images={images}
                  getImageUrl={getImageUrl}
                  imageSelectionnee={imageSelectionnee}
                  onSelectionner={setImageSelectionnee}
                  onReordonner={reordonnerImages}
                  modeEdition={modeEdition}
                  onSupprimer={modeEdition ? supprimerImage : undefined}
                />
              </div>
            )}

            {/* Image principale avec overlay OCR */}
            {images.length > 0 && (
              <div>
                {getBlocsPageCourante().length > 0 ? (
                  <OverlayTexteOCR
                    imageUrl={getImageUrl(images[imageSelectionnee])}
                    blocs={getBlocsPageCourante()}
                    onBlocModifie={(blocIndex, nouveauTexte) => {
                      const pageDataIndex = blocsTexteEdites.findIndex(
                        (p) => p.page === imageSelectionnee
                      )
                      if (pageDataIndex !== -1) {
                        modifierBlocTexte(pageDataIndex, blocIndex, nouveauTexte)
                      }
                    }}
                  />
                ) : (
                  <div className="relative rounded-lg overflow-hidden bg-ink-lighter">
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
                    <div className="absolute bottom-4 left-4 right-4 z-10">
                      <div className="bg-ink/80 backdrop-blur-sm text-white text-xs px-4 py-2 rounded-lg">
                        Le texte ci-dessous correspond a l'ensemble des {images.length} page{images.length > 1 ? 's' : ''} scannee{images.length > 1 ? 's' : ''}
                      </div>
                    </div>
                  </div>
                )}
              </div>
            )}

            {/* Bouton Re-OCR si blocsTexte manquant */}
            {images.length > 0 &&
             (!cours.blocsTexte || cours.blocsTexte.length === 0) && (
              <div className="bg-amber-50 border border-amber-200 rounded-lg p-4 mb-4 mt-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-amber-800">Overlay OCR non disponible</p>
                    <p className="text-xs text-amber-600 mt-1">Ce cours a ete scanne avant la mise a jour. Relancez l'OCR pour activer la superposition du texte.</p>
                  </div>
                  <button
                    onClick={handleReOCR}
                    disabled={reOCREnCours}
                    className="px-4 py-2 bg-[#F5C542] text-[#1A4D4D] rounded-lg text-sm font-medium hover:bg-[#e0b23a] disabled:opacity-50 transition-colors whitespace-nowrap ml-4"
                  >
                    {reOCREnCours ? 'Re-OCR en cours...' : 'Relancer l\'OCR'}
                  </button>
                </div>
                {reOCREnCours && (
                  <div className="flex items-center gap-2 mt-3 text-amber-700">
                    <div className="animate-spin h-4 w-4 border-2 border-amber-600 border-t-transparent rounded-full" />
                    <span className="text-xs">Traitement des images en cours, cela peut prendre quelques secondes...</span>
                  </div>
                )}
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
                {(overlayModifie ? texteEdite : cours.texteOCR)
                  ? renderTexteAvecSurlignage(overlayModifie ? texteEdite : cours.texteOCR, conceptSelectionne)
                  : 'Aucun contenu textuel disponible.'}
              </pre>
            </div>
          )}

          {/* Bouton flottant de sauvegarde overlay */}
          {overlayModifie && !modeEdition && (
            <div className="mt-4 flex justify-end">
              <button
                onClick={async () => {
                  await sauvegarderModifications()
                  setOverlayModifie(false)
                }}
                disabled={sauvegarde}
                className="px-6 py-2 bg-success text-white rounded-full text-sm font-medium hover:bg-success/90 transition-colors disabled:opacity-50 shadow-lg"
              >
                {sauvegarde ? 'Sauvegarde...' : 'Sauvegarder les modifications'}
              </button>
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
      setCours(res.cours || [])
      setTotal(res.total || 0)
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
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-8">
        <div>
          <h1 className="font-display text-xl md:text-3xl font-bold text-ink mb-2">
            Mes cours
          </h1>
          <p className="text-ink-light">
            {total} cours enregistré{total > 1 ? 's' : ''}
          </p>
        </div>
        <Link
          to="/scanner"
          className="inline-flex items-center gap-4 px-6 py-2.5 md:px-8 md:py-3 bg-coral text-white rounded-full font-semibold hover:bg-coral-dark transition-colors self-start sm:self-auto"
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
        <div className="bg-white rounded-lg p-8 md:p-12 text-center">
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
