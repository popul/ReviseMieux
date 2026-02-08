import { useCallback, useRef, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import IndicateurEtapes from '../components/IndicateurEtapes'
import ZoneUpload from '../components/ZoneUpload'
import PreviewFichiers from '../components/PreviewFichiers'
import OptionsGeneration, { type OptionsGenerationType } from '../components/OptionsGeneration'
import ProcessingSection from '../components/ProcessingSection'
import EditeurTexteOCR from '../components/EditeurTexteOCR'
import { envoyerOCR, type ReponseOCR, type ZoneIncertaine } from '../services/api'

const MAX_FICHIERS = 10

const ETAPES = [
  { numero: 1, libelle: 'Import' },
  { numero: 2, libelle: 'Options' },
  { numero: 3, libelle: 'Résultat' },
]

const OPTIONS_DEFAUT: OptionsGenerationType = {
  titre: '',
  matiere: '',
  genererFiches: true,
  genererQuiz: true,
  genererMindmap: false,
}

type EtatPage = 'upload' | 'processing' | 'resultat' | 'erreur'

interface EtatErreur {
  code: string
  message: string
}

export default function Scanner() {
  const navigate = useNavigate()
  const [fichiers, setFichiers] = useState<File[]>([])
  const [options, setOptions] = useState<OptionsGenerationType>(OPTIONS_DEFAUT)
  const [etapeActive, setEtapeActive] = useState(1)
  const [etat, setEtat] = useState<EtatPage>('upload')
  const [erreur, setErreur] = useState<EtatErreur | null>(null)
  const [resultatOCR, setResultatOCR] = useState<ReponseOCR | null>(null)
  const [texteEdite, setTexteEdite] = useState('')
  const [zonesIncertaines, setZonesIncertaines] = useState<ZoneIncertaine[]>([])
  const inputRef = useRef<HTMLInputElement>(null)

  const ajouterFichiers = useCallback((nouveauxFichiers: File[]) => {
    setFichiers((prev) => {
      const total = [...prev, ...nouveauxFichiers]
      return total.slice(0, MAX_FICHIERS)
    })
    setEtapeActive(2)
  }, [])

  const supprimerFichier = useCallback((index: number) => {
    setFichiers((prev) => {
      const nouveaux = prev.filter((_, i) => i !== index)
      if (nouveaux.length === 0) {
        setEtapeActive(1)
      }
      return nouveaux
    })
  }, [])

  const ouvrirSelecteur = useCallback(() => {
    inputRef.current?.click()
  }, [])

  const reinitialiser = useCallback(() => {
    setFichiers([])
    setOptions(OPTIONS_DEFAUT)
    setEtapeActive(1)
    setEtat('upload')
    setErreur(null)
    setResultatOCR(null)
    setTexteEdite('')
    setZonesIncertaines([])
  }, [])

  const auMoinsUneOption =
    options.genererFiches || options.genererQuiz || options.genererMindmap

  const peutSoumettre = fichiers.length > 0 && auMoinsUneOption

  const gererSoumission = async () => {
    if (!peutSoumettre) return

    setEtat('processing')
    setErreur(null)

    try {
      const resultat = await envoyerOCR(fichiers, {
        titre: options.titre || undefined,
        matiere: options.matiere || undefined,
        sauvegarder: true,
      })

      setResultatOCR(resultat)
      setTexteEdite(resultat.texte)
      setZonesIncertaines(resultat.zonesIncertaines || [])
      setEtat('resultat')
      setEtapeActive(3)
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Une erreur inattendue est survenue'

      // Déterminer le code d'erreur
      let code = 'ERREUR_INCONNUE'
      if (message.includes('429') || message.toLowerCase().includes('rate limit')) {
        code = 'QUOTA_DEPASSE'
      } else if (message.includes('503') || message.toLowerCase().includes('indisponible')) {
        code = 'SERVICE_INDISPONIBLE'
      } else if (message.includes('400')) {
        code = 'FICHIER_INVALIDE'
      }

      setErreur({ code, message })
      setEtat('erreur')
    }
  }

  const gererChangementTexte = useCallback((nouveauTexte: string) => {
    setTexteEdite(nouveauTexte)
    // Recalculer les zones incertaines (les invalider si le texte a changé significativement)
    // Pour simplifier, on garde les zones existantes mais on pourrait les recalculer
    setZonesIncertaines([])
  }, [])

  const gererValidation = useCallback(() => {
    // Rediriger vers la page du cours pour générer les supports
    if (resultatOCR?.coursId) {
      navigate(`/cours?id=${resultatOCR.coursId}`)
    } else {
      // Fallback vers la liste des cours si pas d'ID
      navigate('/cours')
    }
  }, [resultatOCR, navigate])

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
        <h1 className="font-display text-xl font-semibold">Scanner un cours</h1>

        <div className="ml-auto hidden md:block">
          <IndicateurEtapes etapes={ETAPES} etapeActive={etapeActive} />
        </div>
      </header>

      {/* Input caché pour le bouton "Ajouter" */}
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

      {/* État: Upload */}
      {etat === 'upload' && (
        <>
          <section className="text-center">
            <h2 className="font-display text-3xl font-bold mb-2">Importe tes notes de cours</h2>
            <p className="text-ink-light text-lg mb-12">
              Prends en photo ou scanne tes notes manuscrites ou imprimées
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
            <OptionsGeneration options={options} onChange={setOptions} />
          )}

          {fichiers.length > 0 && (
            <div className="mt-12 text-center">
              <button
                type="button"
                className={`
                  inline-flex items-center gap-4 py-4 px-12 rounded-full font-semibold text-lg transition-all
                  ${
                    peutSoumettre
                      ? 'bg-teal text-white hover:bg-teal-light hover:-translate-y-0.5 hover:shadow-xl'
                      : 'bg-ink-muted text-white cursor-not-allowed'
                  }
                `}
                disabled={!peutSoumettre}
                onClick={gererSoumission}
              >
                <span>✨</span> Générer mes supports de révision
              </button>

              {!auMoinsUneOption && (
                <p className="mt-4 text-sm text-coral">
                  Sélectionne au moins un type de support à générer
                </p>
              )}
            </div>
          )}
        </>
      )}

      {/* État: Processing */}
      {etat === 'processing' && (
        <ProcessingSection message="Extraction du texte en cours..." />
      )}

      {/* État: Erreur */}
      {etat === 'erreur' && erreur && (
        <div className="bg-white rounded-lg p-12 text-center">
          <div className="w-16 h-16 mx-auto mb-6 bg-coral/10 rounded-full flex items-center justify-center">
            <span className="text-3xl">⚠️</span>
          </div>

          <h3 className="font-display text-xl font-semibold mb-4 text-coral">
            {erreur.code === 'QUOTA_DEPASSE'
              ? 'Quota dépassé'
              : erreur.code === 'SERVICE_INDISPONIBLE'
                ? 'Service indisponible'
                : erreur.code === 'FICHIER_INVALIDE'
                  ? 'Fichier invalide'
                  : 'Une erreur est survenue'}
          </h3>

          <p className="text-ink-light mb-8">{erreur.message}</p>

          <div className="flex gap-6 justify-center">
            <button
              type="button"
              className="px-8 py-4 bg-cream text-ink rounded-full font-medium hover:bg-cream/80 transition-colors"
              onClick={reinitialiser}
            >
              Recommencer
            </button>
            {erreur.code !== 'QUOTA_DEPASSE' && (
              <button
                type="button"
                className="px-8 py-4 bg-teal text-white rounded-full font-medium hover:bg-teal-light transition-colors"
                onClick={gererSoumission}
              >
                Réessayer
              </button>
            )}
          </div>
        </div>
      )}

      {/* État: Résultat */}
      {etat === 'resultat' && resultatOCR && (
        <div className="space-y-8">
          {/* Résumé */}
          <div className="bg-teal/5 rounded-lg p-6 flex items-center gap-6">
            <div className="w-12 h-12 bg-teal/10 rounded-full flex items-center justify-center">
              <span className="text-xl">✓</span>
            </div>
            <div className="flex-1">
              <p className="font-medium text-teal">Extraction réussie</p>
              <p className="text-sm text-ink-light">
                {resultatOCR.nombrePages} page{resultatOCR.nombrePages > 1 ? 's' : ''} traitée
                {resultatOCR.nombrePages > 1 ? 's' : ''}
              </p>
              {/* Afficher le titre et la matière détectés */}
              {(resultatOCR.titreSuggere || resultatOCR.matiereSuggeree) && (
                <div className="mt-2 flex flex-wrap gap-3">
                  {resultatOCR.titreSuggere && (
                    <span className="text-xs bg-teal/10 text-teal px-3 py-1 rounded-full">
                      📝 {resultatOCR.titreSuggere}
                    </span>
                  )}
                  {resultatOCR.matiereSuggeree && (
                    <span className="text-xs bg-coral/10 text-coral px-3 py-1 rounded-full capitalize">
                      📚 {resultatOCR.matiereSuggeree}
                    </span>
                  )}
                </div>
              )}
            </div>
            <button
              type="button"
              className="ml-auto px-6 py-2 bg-cream text-ink rounded-full text-sm font-medium hover:bg-cream/80 transition-colors"
              onClick={reinitialiser}
            >
              Nouveau scan
            </button>
          </div>

          {/* Éditeur de texte OCR */}
          <EditeurTexteOCR
            texte={texteEdite}
            zonesIncertaines={zonesIncertaines}
            confiance={resultatOCR.confiance}
            onTexteChange={gererChangementTexte}
            onValider={gererValidation}
          />

          {/* Récapitulatif des options choisies */}
          <div className="bg-white rounded-lg p-6">
            <h4 className="font-medium text-sm text-ink-light mb-4">Supports à générer</h4>
            <div className="flex gap-4 flex-wrap">
              {options.genererFiches && (
                <span className="px-6 py-2 bg-coral/10 text-coral rounded-full text-sm font-medium">
                  Fiches de révision
                </span>
              )}
              {options.genererQuiz && (
                <span className="px-6 py-2 bg-teal/10 text-teal rounded-full text-sm font-medium">
                  Quiz interactif
                </span>
              )}
              {options.genererMindmap && (
                <span className="px-6 py-2 bg-gold/10 text-gold rounded-full text-sm font-medium">
                  Carte mentale
                </span>
              )}
            </div>
          </div>
        </div>
      )}
    </>
  )
}
