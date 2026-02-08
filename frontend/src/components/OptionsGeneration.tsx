export interface OptionsGenerationType {
  titre: string
  matiere: string
  genererFiches: boolean
  genererQuiz: boolean
  genererMindmap: boolean
}

interface OptionsGenerationProps {
  options: OptionsGenerationType
  onChange: (options: OptionsGenerationType) => void
}

const MATIERES = [
  { value: '', label: 'Détection automatique' },
  { value: 'mathematiques', label: 'Mathématiques' },
  { value: 'francais', label: 'Français' },
  { value: 'histoire', label: 'Histoire' },
  { value: 'geographie', label: 'Géographie' },
  { value: 'physique', label: 'Physique' },
  { value: 'chimie', label: 'Chimie' },
  { value: 'svt', label: 'SVT' },
  { value: 'sciences', label: 'Sciences' },
  { value: 'anglais', label: 'Anglais' },
  { value: 'espagnol', label: 'Espagnol' },
  { value: 'allemand', label: 'Allemand' },
  { value: 'philosophie', label: 'Philosophie' },
  { value: 'ses', label: 'SES' },
  { value: 'economie', label: 'Économie' },
  { value: 'informatique', label: 'Informatique' },
]

interface CheckboxOptionProps {
  label: string
  icone: string
  cochee: boolean
  onChange: (cochee: boolean) => void
}

function CheckboxOption({ label, icone, cochee, onChange }: CheckboxOptionProps) {
  return (
    <label
      className={`
        flex items-center gap-2 px-4 py-2 rounded-full cursor-pointer transition-all
        ${cochee ? 'bg-coral text-white' : 'bg-cream hover:bg-cream-dark'}
      `}
    >
      <input
        type="checkbox"
        className="hidden"
        checked={cochee}
        onChange={(e) => onChange(e.target.checked)}
      />
      <span>{icone}</span>
      {label}
    </label>
  )
}

export default function OptionsGeneration({ options, onChange }: OptionsGenerationProps) {
  const mettreAJour = <K extends keyof OptionsGenerationType>(
    champ: K,
    valeur: OptionsGenerationType[K]
  ) => {
    onChange({ ...options, [champ]: valeur })
  }

  return (
    <div className="bg-white rounded-lg p-8 mt-8">
      <h3 className="font-display text-xl font-semibold mb-6">Personnalise ta génération</h3>

      {/* Titre */}
      <div className="mb-6">
        <label htmlFor="titre" className="block font-medium mb-2 text-[0.95rem]">
          Titre du cours
        </label>
        <input
          id="titre"
          type="text"
          className="w-full py-3.5 px-4 border-2 border-cream-dark rounded-md text-base transition-colors focus:outline-none focus:border-coral"
          placeholder="Laisser vide pour détection automatique"
          value={options.titre}
          onChange={(e) => mettreAJour('titre', e.target.value)}
        />
        <p className="text-xs text-ink-muted mt-1">
          ✨ Le titre sera détecté automatiquement à partir du contenu
        </p>
      </div>

      {/* Matière */}
      <div className="mb-6">
        <label htmlFor="matiere" className="block font-medium mb-2 text-[0.95rem]">
          Matière
        </label>
        <select
          id="matiere"
          className="w-full py-3.5 px-4 border-2 border-cream-dark rounded-md text-base transition-colors focus:outline-none focus:border-coral bg-white appearance-none cursor-pointer"
          style={{
            backgroundImage: `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='24' height='24' viewBox='0 0 24 24' fill='none' stroke='%238A8A8A' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E")`,
            backgroundRepeat: 'no-repeat',
            backgroundPosition: 'right 12px center',
            backgroundSize: '18px',
            paddingRight: '40px',
          }}
          value={options.matiere}
          onChange={(e) => mettreAJour('matiere', e.target.value)}
        >
          {MATIERES.map((m) => (
            <option key={m.value} value={m.value}>
              {m.label}
            </option>
          ))}
        </select>
        <p className="text-xs text-ink-muted mt-1">
          ✨ La matière sera détectée automatiquement si "Détection automatique" est sélectionné
        </p>
      </div>

      {/* Types de génération */}
      <div>
        <label className="block font-medium mb-2 text-[0.95rem]">Que veux-tu générer ?</label>
        <div className="flex flex-wrap gap-4">
          <CheckboxOption
            label="Fiches de révision"
            icone="📄"
            cochee={options.genererFiches}
            onChange={(v) => mettreAJour('genererFiches', v)}
          />
          <CheckboxOption
            label="Quiz"
            icone="🎯"
            cochee={options.genererQuiz}
            onChange={(v) => mettreAJour('genererQuiz', v)}
          />
          <CheckboxOption
            label="Carte mentale"
            icone="🧠"
            cochee={options.genererMindmap}
            onChange={(v) => mettreAJour('genererMindmap', v)}
          />
        </div>
      </div>
    </div>
  )
}
