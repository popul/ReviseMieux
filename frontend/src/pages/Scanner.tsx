import { useCallback, useRef, useState } from 'react'
import { Link } from 'react-router-dom'
import IndicateurEtapes from '../components/IndicateurEtapes'
import ZoneUpload from '../components/ZoneUpload'
import PreviewFichiers from '../components/PreviewFichiers'
import OptionsGeneration, { type OptionsGenerationType } from '../components/OptionsGeneration'

const MAX_FICHIERS = 10

const ETAPES = [
  { numero: 1, libelle: 'Import' },
  { numero: 2, libelle: 'Options' },
  { numero: 3, libelle: 'Génération' },
]

const OPTIONS_DEFAUT: OptionsGenerationType = {
  titre: '',
  matiere: '',
  genererFiches: true,
  genererQuiz: true,
  genererMindmap: false,
}

export default function Scanner() {
  const [fichiers, setFichiers] = useState<File[]>([])
  const [options, setOptions] = useState<OptionsGenerationType>(OPTIONS_DEFAUT)
  const [etapeActive, setEtapeActive] = useState(1)
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

  const auMoinsUneOption =
    options.genererFiches || options.genererQuiz || options.genererMindmap

  const peutSoumettre = fichiers.length > 0 && auMoinsUneOption

  const gererSoumission = () => {
    // Pour l'instant, juste log - l'intégration API sera dans l'étape 8
    console.log('Soumission avec:', { fichiers, options })
    // TODO: Appeler l'API OCR et passer à l'étape 3
  }

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

      {/* Section Upload */}
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

        {/* Preview des fichiers */}
        <PreviewFichiers
          fichiers={fichiers}
          onSupprimer={supprimerFichier}
          onAjouter={ouvrirSelecteur}
          maxFichiers={MAX_FICHIERS}
        />
      </section>

      {/* Options - visibles si au moins un fichier */}
      {fichiers.length > 0 && (
        <OptionsGeneration options={options} onChange={setOptions} />
      )}

      {/* Bouton de soumission */}
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
  )
}
