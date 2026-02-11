import { useAccessibilite } from '../contexte/AccessibiliteContexte'
import type { TailleTexte, ModeDaltonien } from '../contexte/AccessibiliteContexte'

const TAILLES: { valeur: TailleTexte; libelle: string; description: string; icone: string }[] = [
  { valeur: 'normal', libelle: 'Normal', description: 'Taille par defaut', icone: 'A' },
  { valeur: 'large', libelle: 'Grand', description: 'Texte agrandi (120%)', icone: 'A+' },
  { valeur: 'xlarge', libelle: 'Tres grand', description: 'Texte tres agrandi (140%)', icone: 'A++' },
]

const MODES_DALTONIEN: { valeur: ModeDaltonien; libelle: string; description: string }[] = [
  { valeur: 'off', libelle: 'Desactive', description: 'Couleurs standard' },
  { valeur: 'protanopia', libelle: 'Protanopie', description: 'Difficulte a percevoir le rouge' },
  { valeur: 'deuteranopia', libelle: 'Deuteranopie', description: 'Difficulte a percevoir le vert' },
  { valeur: 'tritanopia', libelle: 'Tritanopie', description: 'Difficulte a percevoir le bleu' },
]

export default function Accessibilite() {
  const {
    tailleTexte,
    modeDaltonien,
    contrasteEleve,
    setTailleTexte,
    setModeDaltonien,
    setContrasteEleve,
    reinitialiser,
  } = useAccessibilite()

  const aDesModifications =
    tailleTexte !== 'normal' || modeDaltonien !== 'off' || contrasteEleve

  return (
    <>
      <header className="mb-8 md:mb-12">
        <h1 className="font-display text-2xl md:text-4xl font-bold text-ink mb-2">
          Accessibilite
        </h1>
        <p className="text-ink-light text-lg">
          Adapte l'affichage a tes besoins
        </p>
      </header>

      <div className="max-w-2xl space-y-8">
        {/* Taille du texte */}
        <section className="bg-white rounded-lg p-6 md:p-8">
          <h2 className="font-display text-lg font-semibold text-ink mb-1">Taille du texte</h2>
          <p className="text-sm text-ink-light mb-6">Ajuste la taille du texte sur l'ensemble de l'application.</p>
          <div className="grid grid-cols-3 gap-3" role="radiogroup" aria-label="Choisir la taille du texte">
            {TAILLES.map((taille) => {
              const actif = tailleTexte === taille.valeur
              return (
                <button
                  key={taille.valeur}
                  onClick={() => setTailleTexte(taille.valeur)}
                  className={`flex flex-col items-center gap-2 p-4 rounded-lg border-2 transition-all ${
                    actif
                      ? 'border-coral bg-coral/5'
                      : 'border-cream-dark hover:border-ink-muted'
                  }`}
                  role="radio"
                  aria-checked={actif}
                  aria-label={taille.libelle}
                >
                  <span className={`font-display font-bold ${
                    taille.valeur === 'normal' ? 'text-xl' :
                    taille.valeur === 'large' ? 'text-2xl' : 'text-3xl'
                  } ${actif ? 'text-coral' : 'text-ink'}`}>
                    {taille.icone}
                  </span>
                  <span className={`text-sm font-medium ${actif ? 'text-coral' : 'text-ink'}`}>
                    {taille.libelle}
                  </span>
                  <span className="text-xs text-ink-light">{taille.description}</span>
                </button>
              )
            })}
          </div>
        </section>

        {/* Mode daltonien */}
        <section className="bg-white rounded-lg p-6 md:p-8">
          <h2 className="font-display text-lg font-semibold text-ink mb-1">Mode daltonien</h2>
          <p className="text-sm text-ink-light mb-6">Adapte les couleurs pour une meilleure lisibilite.</p>
          <div className="space-y-2">
            {MODES_DALTONIEN.map((mode) => {
              const actif = modeDaltonien === mode.valeur
              return (
                <button
                  key={mode.valeur}
                  onClick={() => setModeDaltonien(mode.valeur)}
                  className={`w-full flex items-center gap-4 p-4 rounded-lg border-2 text-left transition-all ${
                    actif
                      ? 'border-coral bg-coral/5'
                      : 'border-cream-dark hover:border-ink-muted'
                  }`}
                  role="radio"
                  aria-checked={actif}
                >
                  <div className={`w-5 h-5 rounded-full border-2 flex items-center justify-center shrink-0 ${
                    actif ? 'border-coral' : 'border-ink-muted'
                  }`}>
                    {actif && <div className="w-2.5 h-2.5 rounded-full bg-coral" />}
                  </div>
                  <div>
                    <div className={`text-sm font-medium ${actif ? 'text-coral' : 'text-ink'}`}>
                      {mode.libelle}
                    </div>
                    <div className="text-xs text-ink-light">{mode.description}</div>
                  </div>
                </button>
              )
            })}
          </div>
        </section>

        {/* Contraste eleve */}
        <section className="bg-white rounded-lg p-6 md:p-8">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="font-display text-lg font-semibold text-ink mb-1">Contraste eleve</h2>
              <p className="text-sm text-ink-light">Augmente le contraste des couleurs pour une meilleure lisibilite.</p>
            </div>
            <button
              role="switch"
              aria-checked={contrasteEleve}
              onClick={() => setContrasteEleve(!contrasteEleve)}
              className={`relative w-14 h-7 rounded-full transition-colors shrink-0 ml-6 ${
                contrasteEleve ? 'bg-coral' : 'bg-ink-muted/30'
              }`}
              aria-label="Activer ou desactiver le contraste eleve"
            >
              <span
                className={`absolute top-1 w-5 h-5 bg-white rounded-full transition-transform shadow-sm ${
                  contrasteEleve ? 'translate-x-8' : 'translate-x-1'
                }`}
              />
            </button>
          </div>
        </section>

        {/* Apercu */}
        <section className="bg-white rounded-lg p-6 md:p-8">
          <h2 className="font-display text-lg font-semibold text-ink mb-4">Apercu</h2>
          <div className="bg-cream rounded-lg p-6">
            <h3 className="font-display text-xl font-bold text-ink mb-2">Titre exemple</h3>
            <p className="text-ink-light mb-3">
              Texte de contenu normal avec les couleurs du design system.
              Verifie que tout est bien lisible avec tes parametres.
            </p>
            <div className="flex gap-3">
              <span className="px-3 py-1 rounded-full bg-coral text-white text-sm font-medium">Coral</span>
              <span className="px-3 py-1 rounded-full bg-teal text-white text-sm font-medium">Teal</span>
              <span className="px-3 py-1 rounded-full bg-gold text-ink text-sm font-medium">Gold</span>
            </div>
          </div>
        </section>

        {/* Reinitialiser */}
        {aDesModifications && (
          <div className="text-center pb-4">
            <button
              onClick={reinitialiser}
              className="text-sm text-ink-light hover:text-coral transition-colors font-medium"
            >
              Reinitialiser tous les parametres
            </button>
          </div>
        )}
      </div>
    </>
  )
}
