import { useCallback, useRef, useState } from 'react'
import { Link } from 'react-router-dom'
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
      setZonesIncertaines(resultat.zonesIncertaines)
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
    // TODO: Étape 9 - Lancer la génération des fiches/quiz/mindmap
    console.log('Validation avec texte:', texteEdite)
    console.log('Options:', options)
    // Pour l'instant, afficher un message de succès
    alert('Texte validé ! La génération des supports sera implémentée à l\'étape 9.')
  }, [texteEdite, options])

  return (
    <>
      {/* Header avec navigation */}
      <header className="flex items-center gap-lg mb-xl">
        <Link
          to="/"
          className="flex items-center gap-xs text-ink-light px-sm py-xs rounded-full transition-colors hover:bg-cream hover:text-ink no-underline font-medium"
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
            <h2 className="font-display text-3xl font-bold mb-xs">Importe tes notes de cours</h2>
            <p className="text-ink-light text-lg mb-xl">
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
            <div className="mt-xl text-center">
              <button
                type="button"
                className={`
                  inline-flex items-center gap-sm py-4 px-xl rounded-full font-semibold text-lg transition-all
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
                <p className="mt-sm text-sm text-coral">
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
        <div className="bg-white rounded-lg p-xl text-center">
          <div className="w-16 h-16 mx-auto mb-md bg-coral/10 rounded-full flex items-center justify-center">
            <span className="text-3xl">⚠️</span>
          </div>

          <h3 className="font-display text-xl font-semibold mb-sm text-coral">
            {erreur.code === 'QUOTA_DEPASSE'
              ? 'Quota dépassé'
              : erreur.code === 'SERVICE_INDISPONIBLE'
                ? 'Service indisponible'
                : erreur.code === 'FICHIER_INVALIDE'
                  ? 'Fichier invalide'
                  : 'Une erreur est survenue'}
          </h3>

          <p className="text-ink-light mb-lg">{erreur.message}</p>

          <div className="flex gap-md justify-center">
            <button
              type="button"
              className="px-lg py-sm bg-cream text-ink rounded-full font-medium hover:bg-cream/80 transition-colors"
              onClick={reinitialiser}
            >
              Recommencer
            </button>
            {erreur.code !== 'QUOTA_DEPASSE' && (
              <button
                type="button"
                className="px-lg py-sm bg-teal text-white rounded-full font-medium hover:bg-teal-light transition-colors"
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
        <div className="space-y-lg">
          {/* Résumé */}
          <div className="bg-teal/5 rounded-lg p-md flex items-center gap-md">
            <div className="w-12 h-12 bg-teal/10 rounded-full flex items-center justify-center">
              <span className="text-xl">✓</span>
            </div>
            <div>
              <p className="font-medium text-teal">Extraction réussie</p>
              <p className="text-sm text-ink-light">
                {resultatOCR.nombrePages} page{resultatOCR.nombrePages > 1 ? 's' : ''} traitée
                {resultatOCR.nombrePages > 1 ? 's' : ''}
              </p>
            </div>
            <button
              type="button"
              className="ml-auto px-md py-xs bg-cream text-ink rounded-full text-sm font-medium hover:bg-cream/80 transition-colors"
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
          <div className="bg-white rounded-lg p-md">
            <h4 className="font-medium text-sm text-ink-light mb-sm">Supports à générer</h4>
            <div className="flex gap-sm flex-wrap">
              {options.genererFiches && (
                <span className="px-md py-xs bg-coral/10 text-coral rounded-full text-sm font-medium">
                  Fiches de révision
                </span>
              )}
              {options.genererQuiz && (
                <span className="px-md py-xs bg-teal/10 text-teal rounded-full text-sm font-medium">
                  Quiz interactif
                </span>
              )}
              {options.genererMindmap && (
                <span className="px-md py-xs bg-gold/10 text-gold rounded-full text-sm font-medium">
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
