import { useCallback, useEffect, useRef, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import IndicateurEtapes from '../components/IndicateurEtapes'
import ZoneUpload from '../components/ZoneUpload'
import PreviewFichiers from '../components/PreviewFichiers'
import OptionsGeneration, { type OptionsGenerationType } from '../components/OptionsGeneration'
import ProcessingSection from '../components/ProcessingSection'
import { envoyerOCRStream, obtenirConfig } from '../services/api'

const ETAPES = [
  { numero: 1, libelle: 'Import' },
  { numero: 2, libelle: 'Options' },
]

const OPTIONS_DEFAUT: OptionsGenerationType = {
  titre: '',
  matiere: '',
  genererFiches: true,
  genererQuiz: true,
  genererMindmap: false,
}

type EtatPage = 'upload' | 'processing' | 'erreur'

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
  const [progression, setProgression] = useState<{ page: number; total: number } | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)
  const abortControllerRef = useRef<AbortController | null>(null)
  const [maxFichiers, setMaxFichiers] = useState(30)

  useEffect(() => {
    obtenirConfig().then((cfg) => setMaxFichiers(cfg.nombreMaxPages))
  }, [])

  const ajouterFichiers = useCallback((nouveauxFichiers: File[]) => {
    setFichiers((prev) => {
      const total = [...prev, ...nouveauxFichiers]
      return total.slice(0, maxFichiers)
    })
    setEtapeActive(2)
  }, [maxFichiers])

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
    abortControllerRef.current?.abort()
    setFichiers([])
    setOptions(OPTIONS_DEFAUT)
    setEtapeActive(1)
    setEtat('upload')
    setErreur(null)
    setProgression(null)
  }, [])

  const auMoinsUneOption =
    options.genererFiches || options.genererQuiz || options.genererMindmap

  const peutSoumettre = fichiers.length > 0 && auMoinsUneOption

  const gererSoumission = async () => {
    if (!peutSoumettre) return

    setEtat('processing')
    setErreur(null)
    setProgression(null)

    const abortController = new AbortController()
    abortControllerRef.current = abortController

    try {
      const resultat = await envoyerOCRStream(
        fichiers,
        (page, total) => setProgression({ page, total }),
        {
          titre: options.titre || undefined,
          matiere: options.matiere || undefined,
          sauvegarder: true,
        },
        abortController.signal
      )

      if (resultat.coursId) {
        navigate(`/cours?id=${resultat.coursId}`)
      } else {
        navigate('/cours')
      }
    } catch (err) {
      if ((err as Error).name === 'AbortError') return

      const message = err instanceof Error ? err.message : 'Une erreur inattendue est survenue'
      const status = (err as Error & { status?: number }).status
      const serverCode = (err as Error & { code?: string }).code

      let code = serverCode || 'ERREUR_INCONNUE'
      if (status === 429 || message.toLowerCase().includes('rate limit')) {
        code = 'QUOTA_DEPASSE'
      } else if (status === 503 || message.toLowerCase().includes('indisponible')) {
        code = 'SERVICE_INDISPONIBLE'
      } else if (status === 400) {
        code = 'FICHIER_INVALIDE'
      }

      setErreur({ code, message })
      setEtat('erreur')
    }
  }

  return (
    <>
      {/* Header avec navigation */}
      <header className="flex items-center gap-4 mb-8 md:gap-8 md:mb-12">
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
            <h2 className="font-display text-xl md:text-3xl font-bold mb-2">Importe tes notes de cours</h2>
            <p className="text-ink-light text-lg mb-8 md:mb-12">
              Prends en photo ou scanne tes notes manuscrites ou imprimées
            </p>

            <ZoneUpload
              onFichiersSelectionnes={ajouterFichiers}
              maxFichiers={maxFichiers}
              fichiersCourants={fichiers.length}
            />

            <PreviewFichiers
              fichiers={fichiers}
              onSupprimer={supprimerFichier}
              onAjouter={ouvrirSelecteur}
              maxFichiers={maxFichiers}
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
                  inline-flex items-center gap-4 py-3 px-8 md:py-4 md:px-12 rounded-full font-semibold text-lg transition-all
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
        <ProcessingSection message="Extraction du texte en cours..." progression={progression} />
      )}

      {/* État: Erreur */}
      {etat === 'erreur' && erreur && (
        <div className="bg-white rounded-lg p-6 md:p-12 text-center">
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

          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <button
              type="button"
              className="px-6 py-3 md:px-8 md:py-4 bg-cream text-ink rounded-full font-medium hover:bg-cream/80 transition-colors"
              onClick={reinitialiser}
            >
              Recommencer
            </button>
            {erreur.code !== 'QUOTA_DEPASSE' && (
              <button
                type="button"
                className="px-6 py-3 md:px-8 md:py-4 bg-teal text-white rounded-full font-medium hover:bg-teal-light transition-colors"
                onClick={gererSoumission}
              >
                Réessayer
              </button>
            )}
          </div>
        </div>
      )}

    </>
  )
}
