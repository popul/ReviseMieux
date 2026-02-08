import { useCallback, useEffect, useRef, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import IndicateurEtapes from '../components/IndicateurEtapes'
import ZoneUpload from '../components/ZoneUpload'
import PreviewFichiers from '../components/PreviewFichiers'
import ProcessingSection from '../components/ProcessingSection'
import {
  envoyerOCRCopie,
  analyserCopie,
  obtenirErreursCopie,
  listerCopies,
  supprimerCopie,
  genererRecommandationsCopie,
  type CopieExamen,
  type ErreurAnalyse,
  type ResultatAnalyse,
  type ResultatRecommandations,
} from '../services/api'

const MAX_FICHIERS = 5

const ETAPES = [
  { numero: 1, libelle: 'Import' },
  { numero: 2, libelle: 'Analyse' },
  { numero: 3, libelle: 'Conseils' },
]

type EtatPage = 'liste' | 'upload' | 'ocr_processing' | 'copie_detail' | 'analyse_processing' | 'resultats' | 'erreur'

interface OptionsOCR {
  titre: string
  matiere: string
  noteObtenue: string
  noteTotale: string
  annotationsProfesseur: string
}

const OPTIONS_DEFAUT: OptionsOCR = {
  titre: '',
  matiere: '',
  noteObtenue: '',
  noteTotale: '',
  annotationsProfesseur: '',
}

const MATIERES = [
  'Mathematiques',
  'Francais',
  'Histoire-Geographie',
  'SVT',
  'Physique-Chimie',
  'Anglais',
  'Espagnol',
  'Allemand',
  'Philosophie',
  'SES',
  'NSI',
  'Autre',
]

const TYPES_ERREUR: Record<string, { label: string; icon: string; color: string }> = {
  comprehension: { label: 'Comprehension', icon: '🧠', color: 'coral' },
  methode: { label: 'Methode', icon: '📐', color: 'teal' },
  inattention: { label: 'Inattention', icon: '👀', color: 'gold' },
}

const SEVERITES: Record<string, { label: string; color: string }> = {
  legere: { label: 'Legere', color: 'text-teal' },
  moderate: { label: 'Moderee', color: 'text-gold' },
  grave: { label: 'Grave', color: 'text-coral' },
}

const PRIORITES: Record<number, { label: string; bgColor: string }> = {
  1: { label: 'Urgent', bgColor: 'bg-coral/10' },
  2: { label: 'Important', bgColor: 'bg-gold/10' },
  3: { label: 'Normal', bgColor: 'bg-teal/10' },
  4: { label: 'Secondaire', bgColor: 'bg-cream' },
  5: { label: 'Optionnel', bgColor: 'bg-ink-muted/10' },
}

export default function Analyser() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [fichiers, setFichiers] = useState<File[]>([])
  const [options, setOptions] = useState<OptionsOCR>(OPTIONS_DEFAUT)
  const [etapeActive, setEtapeActive] = useState(1)
  const [etat, setEtat] = useState<EtatPage>('liste')
  const [erreurMessage, setErreurMessage] = useState<string | null>(null)
  const [copies, setCopies] = useState<CopieExamen[]>([])
  const [copieSelectionnee, setCopieSelectionnee] = useState<CopieExamen | null>(null)
  const [resultatAnalyse, setResultatAnalyse] = useState<ResultatAnalyse | null>(null)
  const [erreursAnalyse, setErreursAnalyse] = useState<ErreurAnalyse[]>([])
  const [recommandations, setRecommandations] = useState<ResultatRecommandations | null>(null)
  const [chargementRecommandations, setChargementRecommandations] = useState(false)
  const [chargement, setChargement] = useState(false)
  const [confirmationSuppression, setConfirmationSuppression] = useState<string | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  // Charger les copies existantes
  const chargerCopies = useCallback(async () => {
    try {
      setChargement(true)
      const reponse = await listerCopies()
      if (reponse.succes) {
        setCopies(reponse.copies || [])
      }
    } catch (err) {
      console.error('Erreur chargement copies:', err)
    } finally {
      setChargement(false)
    }
  }, [])

  useEffect(() => {
    const copieId = searchParams.get('copie')
    if (copieId) {
      // Charger la copie specifique et ses erreurs
      chargerCopieEtErreurs(copieId)
    } else {
      chargerCopies()
    }
  }, [searchParams, chargerCopies])

  const chargerCopieEtErreurs = async (copieId: string) => {
    try {
      setChargement(true)
      const reponseCopies = await listerCopies()
      if (reponseCopies.succes) {
        const copie = reponseCopies.copies?.find((c) => c.id === copieId)
        if (copie) {
          setCopieSelectionnee(copie)
          // Charger les erreurs existantes
          const reponseErreurs = await obtenirErreursCopie(copieId)
          if (reponseErreurs.succes && reponseErreurs.erreurs && reponseErreurs.erreurs.length > 0) {
            setErreursAnalyse(reponseErreurs.erreurs)
            setResultatAnalyse({
              erreurs: reponseErreurs.erreurs,
              nombreErreurs: reponseErreurs.nombreErreurs || reponseErreurs.erreurs.length,
              resumeParType: reponseErreurs.comptesParType || {},
              conseilGlobal: '',
            })
            setEtat('resultats')
            setEtapeActive(3)
          } else {
            setEtat('copie_detail')
            setEtapeActive(2)
          }
        } else {
          setEtat('liste')
        }
      }
    } catch (err) {
      console.error('Erreur chargement copie:', err)
      setEtat('liste')
    } finally {
      setChargement(false)
    }
  }

  const ajouterFichiers = useCallback((nouveauxFichiers: File[]) => {
    setFichiers((prev) => {
      const total = [...prev, ...nouveauxFichiers]
      return total.slice(0, MAX_FICHIERS)
    })
  }, [])

  const supprimerFichier = useCallback((index: number) => {
    setFichiers((prev) => prev.filter((_, i) => i !== index))
  }, [])

  const ouvrirSelecteur = useCallback(() => {
    inputRef.current?.click()
  }, [])

  const reinitialiser = useCallback(() => {
    setFichiers([])
    setOptions(OPTIONS_DEFAUT)
    setEtapeActive(1)
    setEtat('liste')
    setErreurMessage(null)
    setCopieSelectionnee(null)
    setResultatAnalyse(null)
    setErreursAnalyse([])
    setRecommandations(null)
    setSearchParams({})
    chargerCopies()
  }, [chargerCopies, setSearchParams])

  const lancerUpload = () => {
    setEtat('upload')
    setEtapeActive(1)
    setFichiers([])
    setOptions(OPTIONS_DEFAUT)
  }

  const peutSoumettre = fichiers.length > 0 && options.titre.trim().length > 0

  const gererSoumissionOCR = async () => {
    if (!peutSoumettre) return

    setEtat('ocr_processing')
    setErreurMessage(null)

    try {
      const resultat = await envoyerOCRCopie(fichiers, {
        titre: options.titre,
        matiere: options.matiere || undefined,
        noteObtenue: options.noteObtenue ? parseFloat(options.noteObtenue) : undefined,
        noteTotale: options.noteTotale ? parseFloat(options.noteTotale) : undefined,
        annotationsProfesseur: options.annotationsProfesseur || undefined,
      })

      if (resultat.succes && resultat.copie) {
        setCopieSelectionnee(resultat.copie)
        setSearchParams({ copie: resultat.copie.id })
        setEtat('copie_detail')
        setEtapeActive(2)
      } else {
        throw new Error(resultat.erreur?.message || 'Erreur OCR')
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Une erreur inattendue est survenue'
      setErreurMessage(message)
      setEtat('erreur')
    }
  }

  const lancerAnalyse = async () => {
    if (!copieSelectionnee) return

    setEtat('analyse_processing')
    setErreurMessage(null)

    try {
      const resultat = await analyserCopie(copieSelectionnee.id)

      if (resultat.succes && resultat.resultat) {
        setResultatAnalyse(resultat.resultat)
        setErreursAnalyse(resultat.resultat.erreurs || [])
        setEtat('resultats')
        setEtapeActive(3)
      } else {
        throw new Error(resultat.erreur?.message || "Erreur lors de l'analyse")
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Une erreur inattendue est survenue'
      setErreurMessage(message)
      setEtat('erreur')
    }
  }

  const gererSuppressionCopie = async (copieId: string) => {
    try {
      await supprimerCopie(copieId)
      setCopies((prev) => prev.filter((c) => c.id !== copieId))
      setConfirmationSuppression(null)
      if (copieSelectionnee?.id === copieId) {
        reinitialiser()
      }
    } catch (err) {
      console.error('Erreur suppression:', err)
    }
  }

  const genererRecommandations = async () => {
    if (!copieSelectionnee) return

    setChargementRecommandations(true)

    try {
      const resultat = await genererRecommandationsCopie(copieSelectionnee.id)

      if (resultat.succes && resultat.recommandations) {
        setRecommandations(resultat.recommandations)
      } else {
        console.error('Erreur generation recommandations:', resultat.erreur?.message)
      }
    } catch (err) {
      console.error('Erreur generation recommandations:', err)
    } finally {
      setChargementRecommandations(false)
    }
  }

  const voirCopie = (copie: CopieExamen) => {
    setCopieSelectionnee(copie)
    setSearchParams({ copie: copie.id })
    setEtat('copie_detail')
    setEtapeActive(2)
  }

  const getScoreColor = (note: number, total: number): string => {
    const pct = (note / total) * 100
    if (pct >= 80) return 'text-teal'
    if (pct >= 60) return 'text-gold'
    if (pct >= 40) return 'text-orange-500'
    return 'text-coral'
  }

  return (
    <>
      {/* Header avec navigation */}
      <header className="flex items-center gap-8 mb-12">
        <Link
          to="/"
          className="flex items-center gap-2 text-ink-light px-4 py-2 rounded-full transition-colors hover:bg-cream hover:text-ink no-underline font-medium"
        >
          ← Retour
        </Link>
        <h1 className="font-display text-xl font-semibold">Analyser une copie</h1>

        <div className="ml-auto hidden md:block">
          <IndicateurEtapes etapes={ETAPES} etapeActive={etapeActive} />
        </div>
      </header>

      {/* Input cache pour le bouton "Ajouter" */}
      <input
        ref={inputRef}
        type="file"
        className="hidden"
        accept=".jpg,.jpeg,.png,.webp,.pdf"
        multiple
        onChange={(e) => {
          if (e.target.files) {
            ajouterFichiers(Array.from(e.target.files))
          }
          e.target.value = ''
        }}
      />

      {/* Etat: Liste des copies */}
      {etat === 'liste' && (
        <section className="space-y-8">
          <div className="text-center">
            <h2 className="font-display text-3xl font-bold mb-2">Mes copies corrigees</h2>
            <p className="text-ink-light text-lg mb-8">
              Scanne tes copies corrigees pour analyser tes erreurs et progresser
            </p>

            <button
              type="button"
              className="inline-flex items-center gap-4 py-3 px-8 rounded-full font-semibold bg-teal text-white hover:bg-teal-light transition-all hover:-translate-y-0.5 hover:shadow-lg"
              onClick={lancerUpload}
            >
              <span>📄</span> Scanner une nouvelle copie
            </button>
          </div>

          {chargement ? (
            <div className="text-center py-12">
              <div className="animate-spin w-8 h-8 border-2 border-teal border-t-transparent rounded-full mx-auto mb-6" />
              <p className="text-ink-light">Chargement...</p>
            </div>
          ) : copies.length === 0 ? (
            <div className="bg-white rounded-lg p-12 text-center">
              <div className="w-16 h-16 mx-auto mb-6 bg-cream rounded-full flex items-center justify-center">
                <span className="text-3xl">📝</span>
              </div>
              <h3 className="font-display text-xl font-semibold mb-4">Aucune copie analysee</h3>
              <p className="text-ink-light">
                Scanne ta premiere copie pour decouvrir tes points forts et axes d'amelioration.
              </p>
            </div>
          ) : (
            <div className="grid gap-6">
              {copies.map((copie) => (
                <div
                  key={copie.id}
                  className="bg-white rounded-lg p-6 hover:shadow-md transition-shadow"
                >
                  <div className="flex items-start gap-6">
                    <div className="w-12 h-12 bg-cream rounded-full flex items-center justify-center flex-shrink-0">
                      <span className="text-xl">📄</span>
                    </div>

                    <div className="flex-1 min-w-0">
                      <h3 className="font-semibold text-lg truncate">{copie.titre}</h3>
                      <div className="flex items-center gap-6 text-sm text-ink-light mt-2">
                        {copie.matiere && (
                          <span className="bg-cream px-4 py-2 rounded-full">{copie.matiere}</span>
                        )}
                        {copie.noteObtenue !== undefined && copie.noteTotale !== undefined && (
                          <span className={`font-semibold ${getScoreColor(copie.noteObtenue, copie.noteTotale)}`}>
                            {copie.noteObtenue}/{copie.noteTotale}
                          </span>
                        )}
                        <span>
                          {new Date(copie.dateCreation).toLocaleDateString('fr-FR', {
                            day: 'numeric',
                            month: 'short',
                          })}
                        </span>
                      </div>
                    </div>

                    <div className="flex items-center gap-4">
                      <button
                        type="button"
                        className="px-6 py-4 bg-teal/10 text-teal rounded-full text-sm font-medium hover:bg-teal/20 transition-colors"
                        onClick={() => voirCopie(copie)}
                      >
                        Analyser
                      </button>
                      <button
                        type="button"
                        className="p-4 text-ink-light hover:text-coral transition-colors"
                        onClick={() => setConfirmationSuppression(copie.id)}
                        title="Supprimer"
                      >
                        🗑️
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>
      )}

      {/* Etat: Upload */}
      {etat === 'upload' && (
        <>
          <section className="text-center">
            <h2 className="font-display text-3xl font-bold mb-2">Scanne ta copie corrigee</h2>
            <p className="text-ink-light text-lg mb-12">
              Prends en photo ou scanne ta copie avec les corrections du professeur
            </p>

            <ZoneUpload
              onFichiersSelectionnes={ajouterFichiers}
              maxFichiers={MAX_FICHIERS}
              fichiersCourants={fichiers.length}
            />

            <PreviewFichiers
              fichiers={fichiers}
              onSupprimer={supprimerFichier}
              onAjouter={ouvrirSelecteur}
              maxFichiers={MAX_FICHIERS}
            />
          </section>

          {fichiers.length > 0 && (
            <section className="bg-white rounded-lg p-8 mt-8">
              <h3 className="font-display text-lg font-semibold mb-6">Informations sur la copie</h3>

              <div className="grid md:grid-cols-2 gap-6">
                <div>
                  <label htmlFor="titre" className="block text-sm font-medium text-ink-light mb-2">
                    Titre de l'examen *
                  </label>
                  <input
                    id="titre"
                    type="text"
                    className="w-full px-6 py-4 border border-ink-muted rounded-lg focus:outline-none focus:border-teal"
                    placeholder="Ex: Controle de Maths - Chapitre 5"
                    value={options.titre}
                    onChange={(e) => setOptions((prev) => ({ ...prev, titre: e.target.value }))}
                  />
                </div>

                <div>
                  <label htmlFor="matiere" className="block text-sm font-medium text-ink-light mb-2">
                    Matiere
                  </label>
                  <select
                    id="matiere"
                    className="w-full px-6 py-4 border border-ink-muted rounded-lg focus:outline-none focus:border-teal bg-white"
                    value={options.matiere}
                    onChange={(e) => setOptions((prev) => ({ ...prev, matiere: e.target.value }))}
                  >
                    <option value="">Choisir une matiere</option>
                    {MATIERES.map((m) => (
                      <option key={m} value={m}>
                        {m}
                      </option>
                    ))}
                  </select>
                </div>

                <div>
                  <label htmlFor="noteObtenue" className="block text-sm font-medium text-ink-light mb-2">
                    Note obtenue
                  </label>
                  <div className="flex items-center gap-4">
                    <input
                      id="noteObtenue"
                      type="number"
                      step="0.5"
                      min="0"
                      className="w-20 px-6 py-4 border border-ink-muted rounded-lg focus:outline-none focus:border-teal"
                      placeholder="12"
                      value={options.noteObtenue}
                      onChange={(e) => setOptions((prev) => ({ ...prev, noteObtenue: e.target.value }))}
                    />
                    <span className="text-ink-light">/</span>
                    <input
                      id="noteTotale"
                      type="number"
                      step="1"
                      min="1"
                      className="w-20 px-6 py-4 border border-ink-muted rounded-lg focus:outline-none focus:border-teal"
                      placeholder="20"
                      value={options.noteTotale}
                      onChange={(e) => setOptions((prev) => ({ ...prev, noteTotale: e.target.value }))}
                    />
                  </div>
                </div>
              </div>

              <div className="mt-6">
                <label htmlFor="annotations" className="block text-sm font-medium text-ink-light mb-2">
                  Annotations du professeur (optionnel)
                </label>
                <textarea
                  id="annotations"
                  className="w-full px-6 py-4 border border-ink-muted rounded-lg focus:outline-none focus:border-teal resize-none"
                  rows={3}
                  placeholder="Recopiez ici les commentaires generaux du professeur si difficiles a lire sur la copie..."
                  value={options.annotationsProfesseur}
                  onChange={(e) => setOptions((prev) => ({ ...prev, annotationsProfesseur: e.target.value }))}
                />
              </div>
            </section>
          )}

          <div className="mt-8 flex gap-6 justify-center">
            <button
              type="button"
              className="px-8 py-4 bg-cream text-ink rounded-full font-medium hover:bg-cream/80 transition-colors"
              onClick={reinitialiser}
            >
              Annuler
            </button>
            <button
              type="button"
              className={`
                inline-flex items-center gap-4 py-3 px-8 rounded-full font-semibold transition-all
                ${
                  peutSoumettre
                    ? 'bg-teal text-white hover:bg-teal-light hover:-translate-y-0.5 hover:shadow-xl'
                    : 'bg-ink-muted text-white cursor-not-allowed'
                }
              `}
              disabled={!peutSoumettre}
              onClick={gererSoumissionOCR}
            >
              <span>🔍</span> Extraire le texte
            </button>
          </div>
        </>
      )}

      {/* Etat: OCR Processing */}
      {etat === 'ocr_processing' && (
        <ProcessingSection message="Extraction du texte en cours..." />
      )}

      {/* Etat: Analyse Processing */}
      {etat === 'analyse_processing' && (
        <ProcessingSection message="Analyse des erreurs en cours..." />
      )}

      {/* Etat: Detail de la copie */}
      {etat === 'copie_detail' && copieSelectionnee && (
        <section className="space-y-8">
          <div className="bg-white rounded-lg p-8">
            <div className="flex items-start justify-between mb-6">
              <div>
                <h2 className="font-display text-2xl font-bold">{copieSelectionnee.titre}</h2>
                <div className="flex items-center gap-6 text-sm text-ink-light mt-2">
                  {copieSelectionnee.matiere && (
                    <span className="bg-cream px-4 py-2 rounded-full">{copieSelectionnee.matiere}</span>
                  )}
                  {copieSelectionnee.noteObtenue !== undefined && copieSelectionnee.noteTotale !== undefined && (
                    <span
                      className={`font-semibold ${getScoreColor(copieSelectionnee.noteObtenue, copieSelectionnee.noteTotale)}`}
                    >
                      {copieSelectionnee.noteObtenue}/{copieSelectionnee.noteTotale}
                    </span>
                  )}
                </div>
              </div>
              <div className="flex items-center gap-2 text-sm">
                <span className="text-ink-light">Confiance OCR:</span>
                <span className={`font-semibold ${copieSelectionnee.confiance >= 0.85 ? 'text-teal' : 'text-gold'}`}>
                  {Math.round(copieSelectionnee.confiance * 100)}%
                </span>
              </div>
            </div>

            <div className="bg-cream/50 rounded-lg p-6 max-h-64 overflow-y-auto">
              <pre className="whitespace-pre-wrap font-sans text-sm">{copieSelectionnee.texteOCR}</pre>
            </div>

            {copieSelectionnee.annotationsProfesseur && (
              <div className="mt-6 p-6 bg-coral/5 rounded-lg border border-coral/20">
                <h4 className="text-sm font-semibold text-coral mb-2">Annotations du professeur</h4>
                <p className="text-sm">{copieSelectionnee.annotationsProfesseur}</p>
              </div>
            )}
          </div>

          <div className="text-center">
            <p className="text-ink-light mb-6">
              L'IA va analyser ta copie pour identifier les types d'erreurs et te donner des conseils personnalises.
            </p>
            <div className="flex gap-6 justify-center">
              <button
                type="button"
                className="px-8 py-4 bg-cream text-ink rounded-full font-medium hover:bg-cream/80 transition-colors"
                onClick={reinitialiser}
              >
                Retour
              </button>
              <button
                type="button"
                className="inline-flex items-center gap-4 py-3 px-8 rounded-full font-semibold bg-coral text-white hover:bg-coral/90 transition-all hover:-translate-y-0.5 hover:shadow-xl"
                onClick={lancerAnalyse}
              >
                <span>🎯</span> Analyser mes erreurs
              </button>
            </div>
          </div>
        </section>
      )}

      {/* Etat: Resultats */}
      {etat === 'resultats' && resultatAnalyse && copieSelectionnee && (
        <section className="space-y-8">
          {/* Resume */}
          <div className="bg-white rounded-lg p-8">
            <div className="flex items-start justify-between mb-6">
              <div>
                <h2 className="font-display text-2xl font-bold">{copieSelectionnee.titre}</h2>
                {copieSelectionnee.noteObtenue !== undefined && copieSelectionnee.noteTotale !== undefined && (
                  <p className="text-ink-light mt-2">
                    Note:{' '}
                    <span
                      className={`font-semibold ${getScoreColor(copieSelectionnee.noteObtenue, copieSelectionnee.noteTotale)}`}
                    >
                      {copieSelectionnee.noteObtenue}/{copieSelectionnee.noteTotale}
                    </span>
                  </p>
                )}
              </div>
              <button
                type="button"
                className="px-6 py-4 bg-cream text-ink rounded-full text-sm font-medium hover:bg-cream/80 transition-colors"
                onClick={reinitialiser}
              >
                Nouvelle analyse
              </button>
            </div>

            {/* Resume par type */}
            <div className="grid grid-cols-3 gap-6">
              {Object.entries(TYPES_ERREUR).map(([type, config]) => {
                const count = resultatAnalyse.resumeParType?.[type] || 0
                return (
                  <div
                    key={type}
                    className={`p-6 rounded-lg bg-${config.color}/10 text-center`}
                  >
                    <div className="text-2xl mb-2">{config.icon}</div>
                    <div className="text-2xl font-bold">{count}</div>
                    <div className="text-sm text-ink-light">{config.label}</div>
                  </div>
                )
              })}
            </div>
          </div>

          {/* Points forts et a ameliorer */}
          {(resultatAnalyse.pointsForts?.length || resultatAnalyse.pointsAAmeliorer?.length) && (
            <div className="grid md:grid-cols-2 gap-6">
              {resultatAnalyse.pointsForts && resultatAnalyse.pointsForts.length > 0 && (
                <div className="bg-teal/5 rounded-lg p-8">
                  <h3 className="font-display text-lg font-semibold text-teal mb-6 flex items-center gap-4">
                    <span>✓</span> Points forts
                  </h3>
                  <ul className="space-y-4">
                    {resultatAnalyse.pointsForts.map((point, i) => (
                      <li key={i} className="flex items-start gap-4 text-sm">
                        <span className="text-teal">•</span>
                        <span>{point}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}

              {resultatAnalyse.pointsAAmeliorer && resultatAnalyse.pointsAAmeliorer.length > 0 && (
                <div className="bg-coral/5 rounded-lg p-8">
                  <h3 className="font-display text-lg font-semibold text-coral mb-6 flex items-center gap-4">
                    <span>↑</span> A ameliorer
                  </h3>
                  <ul className="space-y-4">
                    {resultatAnalyse.pointsAAmeliorer.map((point, i) => (
                      <li key={i} className="flex items-start gap-4 text-sm">
                        <span className="text-coral">•</span>
                        <span>{point}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          )}

          {/* Conseil global */}
          {resultatAnalyse.conseilGlobal && (
            <div className="bg-gold/10 rounded-lg p-8">
              <h3 className="font-display text-lg font-semibold text-gold mb-4 flex items-center gap-4">
                <span>💡</span> Conseil
              </h3>
              <p className="text-sm">{resultatAnalyse.conseilGlobal}</p>
            </div>
          )}

          {/* Liste des erreurs */}
          {erreursAnalyse.length > 0 && (
            <div className="bg-white rounded-lg p-8">
              <h3 className="font-display text-lg font-semibold mb-6">Detail des erreurs</h3>
              <div className="space-y-6">
                {erreursAnalyse.map((erreur) => {
                  const typeConfig = TYPES_ERREUR[erreur.typeErreur] || TYPES_ERREUR.inattention
                  const severiteConfig = SEVERITES[erreur.severite] || SEVERITES.moderate
                  return (
                    <div
                      key={erreur.id}
                      className="border border-ink-muted/30 rounded-lg p-6"
                    >
                      <div className="flex items-center gap-4 mb-4">
                        <span>{typeConfig.icon}</span>
                        <span className={`px-4 py-2 rounded-full text-xs font-medium bg-${typeConfig.color}/10`}>
                          {typeConfig.label}
                        </span>
                        <span className={`text-xs ${severiteConfig.color}`}>{severiteConfig.label}</span>
                      </div>

                      {erreur.texteOriginal && (
                        <div className="bg-coral/5 rounded p-4 mb-4">
                          <p className="text-sm text-coral line-through">{erreur.texteOriginal}</p>
                        </div>
                      )}

                      {erreur.correction && (
                        <div className="bg-teal/5 rounded p-4 mb-4">
                          <p className="text-sm text-teal">{erreur.correction}</p>
                        </div>
                      )}

                      <p className="text-sm text-ink-light mb-4">{erreur.explication}</p>

                      {erreur.conseil && (
                        <div className="bg-gold/5 rounded p-4">
                          <p className="text-xs text-gold">
                            <strong>Conseil:</strong> {erreur.conseil}
                          </p>
                        </div>
                      )}
                    </div>
                  )
                })}
              </div>
            </div>
          )}

          {/* Section Recommandations */}
          <div className="bg-white rounded-lg p-8">
            <div className="flex items-center justify-between mb-6">
              <h3 className="font-display text-lg font-semibold flex items-center gap-4">
                <span>🎯</span> Recommandations personnalisees
              </h3>
              {!recommandations && (
                <button
                  type="button"
                  className={`
                    inline-flex items-center gap-4 px-6 py-4 rounded-full text-sm font-medium transition-all
                    ${
                      chargementRecommandations
                        ? 'bg-ink-muted text-white cursor-wait'
                        : 'bg-teal text-white hover:bg-teal-light hover:-translate-y-0.5'
                    }
                  `}
                  disabled={chargementRecommandations}
                  onClick={genererRecommandations}
                >
                  {chargementRecommandations ? (
                    <>
                      <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                      Generation...
                    </>
                  ) : (
                    <>
                      <span>✨</span> Generer des recommandations
                    </>
                  )}
                </button>
              )}
            </div>

            {!recommandations && !chargementRecommandations && (
              <p className="text-ink-light text-sm">
                Basees sur l'analyse de ta copie, l'IA va te proposer un plan d'action personnalise pour progresser.
              </p>
            )}

            {chargementRecommandations && (
              <div className="text-center py-8">
                <div className="w-8 h-8 border-2 border-teal border-t-transparent rounded-full animate-spin mx-auto mb-6" />
                <p className="text-ink-light">Analyse de tes lacunes en cours...</p>
              </div>
            )}

            {recommandations && (
              <div className="space-y-8">
                {/* Message de motivation */}
                {recommandations.motivation && (
                  <div className="bg-teal/5 rounded-lg p-6 border border-teal/20">
                    <p className="text-sm text-teal font-medium">{recommandations.motivation}</p>
                  </div>
                )}

                {/* Resume */}
                {recommandations.resume && (
                  <div>
                    <h4 className="font-semibold text-sm mb-2">Resume</h4>
                    <p className="text-sm text-ink-light">{recommandations.resume}</p>
                  </div>
                )}

                {/* Liste des recommandations */}
                {recommandations.recommandations.length > 0 && (
                  <div>
                    <h4 className="font-semibold text-sm mb-6">Domaines a reviser ({recommandations.nombreRecommandations})</h4>
                    <div className="space-y-6">
                      {recommandations.recommandations.map((reco, index) => {
                        const prioriteConfig = PRIORITES[reco.priorite] || PRIORITES[3]
                        const severiteConfig = SEVERITES[reco.severiteMax] || SEVERITES.moderate
                        return (
                          <div
                            key={index}
                            className={`rounded-lg p-6 ${prioriteConfig.bgColor}`}
                          >
                            <div className="flex items-center gap-4 mb-4">
                              <span className="w-6 h-6 rounded-full bg-ink text-white text-xs font-bold flex items-center justify-center">
                                {reco.priorite}
                              </span>
                              <span className="font-semibold">{reco.domaine}</span>
                              <span className={`text-xs ${severiteConfig.color}`}>
                                {severiteConfig.label}
                              </span>
                            </div>

                            <p className="text-sm text-ink-light mb-4">{reco.raison}</p>

                            <div className="bg-white/60 rounded p-4">
                              <p className="text-sm">
                                <strong className="text-teal">Action :</strong> {reco.actionSuggerie}
                              </p>
                              {reco.typeQuiz && (
                                <p className="text-xs text-ink-light mt-2">
                                  Quiz suggere : {reco.typeQuiz}
                                </p>
                              )}
                            </div>
                          </div>
                        )
                      })}
                    </div>
                  </div>
                )}

                {/* Plan d'action */}
                {recommandations.planAction && (
                  <div className="bg-gold/10 rounded-lg p-6">
                    <h4 className="font-semibold text-sm mb-2 flex items-center gap-4">
                      <span>📋</span> Plan d'action pour cette semaine
                    </h4>
                    <p className="text-sm">{recommandations.planAction}</p>
                  </div>
                )}

                {/* Prochain quiz */}
                {recommandations.prochainQuiz && (
                  <div className="bg-teal/10 rounded-lg p-6">
                    <h4 className="font-semibold text-sm mb-2 flex items-center gap-4">
                      <span>📝</span> Prochain quiz a faire
                    </h4>
                    <p className="text-sm">{recommandations.prochainQuiz}</p>
                  </div>
                )}
              </div>
            )}
          </div>
        </section>
      )}

      {/* Etat: Erreur */}
      {etat === 'erreur' && erreurMessage && (
        <div className="bg-white rounded-lg p-12 text-center">
          <div className="w-16 h-16 mx-auto mb-6 bg-coral/10 rounded-full flex items-center justify-center">
            <span className="text-3xl">⚠️</span>
          </div>

          <h3 className="font-display text-xl font-semibold mb-4 text-coral">Une erreur est survenue</h3>

          <p className="text-ink-light mb-8">{erreurMessage}</p>

          <div className="flex gap-6 justify-center">
            <button
              type="button"
              className="px-8 py-4 bg-cream text-ink rounded-full font-medium hover:bg-cream/80 transition-colors"
              onClick={reinitialiser}
            >
              Recommencer
            </button>
          </div>
        </div>
      )}

      {/* Modal de confirmation de suppression */}
      {confirmationSuppression && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-8 max-w-sm mx-6">
            <h3 className="font-display text-lg font-semibold mb-4">Supprimer cette copie ?</h3>
            <p className="text-ink-light text-sm mb-8">
              Cette action est irreversible. L'analyse associee sera egalement supprimee.
            </p>
            <div className="flex gap-6 justify-end">
              <button
                type="button"
                className="px-6 py-4 bg-cream text-ink rounded-full text-sm font-medium hover:bg-cream/80 transition-colors"
                onClick={() => setConfirmationSuppression(null)}
              >
                Annuler
              </button>
              <button
                type="button"
                className="px-6 py-4 bg-coral text-white rounded-full text-sm font-medium hover:bg-coral/90 transition-colors"
                onClick={() => gererSuppressionCopie(confirmationSuppression)}
              >
                Supprimer
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  )
}
