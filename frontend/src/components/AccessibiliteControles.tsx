import { useState } from 'react'
import { useAccessibilite } from '../contexte/AccessibiliteContexte'
import type { TailleTexte, ModeDaltonien } from '../contexte/AccessibiliteContexte'

const TAILLES: { valeur: TailleTexte; libelle: string; icone: string }[] = [
  { valeur: 'normal', libelle: 'Normal', icone: 'A' },
  { valeur: 'large', libelle: 'Grand', icone: 'A+' },
  { valeur: 'xlarge', libelle: 'Tres grand', icone: 'A++' },
]

const MODES_DALTONIEN: { valeur: ModeDaltonien; libelle: string }[] = [
  { valeur: 'off', libelle: 'Desactive' },
  { valeur: 'protanopia', libelle: 'Protanopie (rouge)' },
  { valeur: 'deuteranopia', libelle: 'Deuteranopie (vert)' },
  { valeur: 'tritanopia', libelle: 'Tritanopie (bleu)' },
]

export default function AccessibiliteControles() {
  const [ouvert, setOuvert] = useState(false)
  const {
    tailleTexte,
    modeDaltonien,
    contrasteEleve,
    setTailleTexte,
    setModeDaltonien,
    setContrasteEleve,
    reinitialiser,
  } = useAccessibilite()

  return (
    <div className="relative">
      {/* Bouton pour ouvrir/fermer le panneau */}
      <button
        onClick={() => setOuvert(!ouvert)}
        className="flex items-center gap-2 px-3 py-2 rounded-md text-sm font-medium text-white/70 hover:bg-white/[0.08] hover:text-white transition-colors"
        aria-expanded={ouvert}
        aria-controls="panneau-accessibilite"
        aria-label="Options d'accessibilite"
      >
        <span className="text-lg" aria-hidden="true">♿</span>
        <span className="hidden sm:inline">Accessibilite</span>
      </button>

      {/* Panneau de contrôles */}
      {ouvert && (
        <div
          id="panneau-accessibilite"
          className="absolute bottom-full left-0 mb-2 w-72 bg-ink-light rounded-md shadow-lg border border-white/10 p-4 z-50"
          role="dialog"
          aria-label="Parametres d'accessibilite"
        >
          <div className="flex justify-between items-center mb-4">
            <h3 className="text-white font-semibold">Accessibilite</h3>
            <button
              onClick={() => setOuvert(false)}
              className="text-white/50 hover:text-white"
              aria-label="Fermer le panneau"
            >
              ✕
            </button>
          </div>

          {/* Taille du texte */}
          <fieldset className="mb-4">
            <legend className="text-xs text-white/70 uppercase tracking-wider mb-2">
              Taille du texte
            </legend>
            <div className="flex gap-1" role="radiogroup" aria-label="Choisir la taille du texte">
              {TAILLES.map((taille) => (
                <button
                  key={taille.valeur}
                  onClick={() => setTailleTexte(taille.valeur)}
                  className={`flex-1 py-2 px-3 rounded text-sm font-medium transition-colors ${
                    tailleTexte === taille.valeur
                      ? 'bg-coral text-white'
                      : 'bg-white/10 text-white/70 hover:bg-white/20'
                  }`}
                  role="radio"
                  aria-checked={tailleTexte === taille.valeur}
                  aria-label={taille.libelle}
                >
                  {taille.icone}
                </button>
              ))}
            </div>
          </fieldset>

          {/* Mode daltonien */}
          <fieldset className="mb-4">
            <legend className="text-xs text-white/70 uppercase tracking-wider mb-2">
              Mode daltonien
            </legend>
            <select
              value={modeDaltonien}
              onChange={(e) => setModeDaltonien(e.target.value as ModeDaltonien)}
              className="w-full bg-white/10 text-white border border-white/20 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-coral"
              aria-label="Selectionner le mode daltonien"
            >
              {MODES_DALTONIEN.map((mode) => (
                <option key={mode.valeur} value={mode.valeur} className="bg-ink text-white">
                  {mode.libelle}
                </option>
              ))}
            </select>
          </fieldset>

          {/* Contraste élevé */}
          <div className="mb-4">
            <label className="flex items-center justify-between cursor-pointer">
              <span className="text-xs text-white/70 uppercase tracking-wider">
                Contraste eleve
              </span>
              <button
                role="switch"
                aria-checked={contrasteEleve}
                onClick={() => setContrasteEleve(!contrasteEleve)}
                className={`relative w-12 h-6 rounded-full transition-colors ${
                  contrasteEleve ? 'bg-coral' : 'bg-white/20'
                }`}
                aria-label="Activer ou desactiver le contraste eleve"
              >
                <span
                  className={`absolute top-1 w-4 h-4 bg-white rounded-full transition-transform ${
                    contrasteEleve ? 'translate-x-7' : 'translate-x-1'
                  }`}
                />
              </button>
            </label>
          </div>

          {/* Bouton réinitialiser */}
          <button
            onClick={reinitialiser}
            className="w-full py-2 text-sm text-white/70 hover:text-white border border-white/20 rounded hover:border-white/40 transition-colors"
            aria-label="Reinitialiser tous les parametres d'accessibilite"
          >
            Reinitialiser
          </button>

          {/* Aperçu */}
          <div className="mt-4 pt-4 border-t border-white/10">
            <p className="text-xs text-white/50 mb-2">Aperçu:</p>
            <div className="bg-cream rounded p-3 text-ink text-sm">
              <p className="font-display font-semibold">Titre exemple</p>
              <p className="font-body">Texte de contenu normal avec des couleurs du design system.</p>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
