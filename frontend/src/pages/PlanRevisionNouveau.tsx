import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { listerCours, creerPlanRevision, type Cours } from '../services/api'

const MATIERES = [
  { value: '', label: 'Choisir une matiere', icone: '📋' },
  { value: 'mathematiques', label: 'Mathematiques', icone: '📐' },
  { value: 'francais', label: 'Francais', icone: '📖' },
  { value: 'histoire', label: 'Histoire', icone: '🏛️' },
  { value: 'geographie', label: 'Geographie', icone: '🗺️' },
  { value: 'sciences', label: 'Sciences', icone: '🔬' },
  { value: 'physique', label: 'Physique', icone: '⚛️' },
  { value: 'chimie', label: 'Chimie', icone: '🧪' },
  { value: 'svt', label: 'SVT', icone: '🌿' },
  { value: 'anglais', label: 'Anglais', icone: '🇬🇧' },
  { value: 'ses', label: 'SES', icone: '📊' },
  { value: 'philosophie', label: 'Philosophie', icone: '🤔' },
]

function getIconeMatiere(matiere: string): string {
  return MATIERES.find(m => m.value === matiere)?.icone || '📋'
}

export default function PlanRevisionNouveau() {
  const navigate = useNavigate()
  const [titre, setTitre] = useState('')
  const [matiere, setMatiere] = useState('')
  const [description, setDescription] = useState('')
  const [dateEcheance, setDateEcheance] = useState('')
  const [coursDisponibles, setCoursDisponibles] = useState<Cours[]>([])
  const [coursSelectionne, setCoursSelectionne] = useState('')
  const [chargement, setChargement] = useState(false)
  const [chargementCours, setChargementCours] = useState(true)
  const [erreur, setErreur] = useState<string | null>(null)

  useEffect(() => {
    listerCours(1, 100)
      .then((res) => setCoursDisponibles(res.cours || []))
      .catch(() => {})
      .finally(() => setChargementCours(false))
  }, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!titre.trim()) return

    setChargement(true)
    setErreur(null)

    try {
      const icone = getIconeMatiere(matiere)
      const res = await creerPlanRevision({
        titre: titre.trim(),
        description: description.trim() || undefined,
        matiere: matiere || undefined,
        iconeMatiere: icone,
        dateEcheance: dateEcheance || undefined,
        coursIds: coursSelectionne ? [coursSelectionne] : [],
      })

      if (res.succes && res.plan) {
        navigate(`/plans/${res.plan.id}`)
      } else {
        setErreur(res.erreur?.message || 'Erreur lors de la creation')
      }
    } catch (err) {
      setErreur(err instanceof Error ? err.message : 'Erreur inconnue')
    } finally {
      setChargement(false)
    }
  }

  return (
    <div className="max-w-3xl mx-auto">
      <header className="mb-8">
        <h1 className="font-display text-xl md:text-3xl font-bold text-ink mb-2">
          Nouveau plan de revision
        </h1>
        <p className="text-ink-light">
          Organise tes cours pour preparer un examen ou un controle
        </p>
      </header>

      {erreur && (
        <div className="bg-red-50 text-red-600 rounded-lg p-4 mb-6 text-sm">
          {erreur}
        </div>
      )}

      <form onSubmit={handleSubmit} className="bg-white rounded-lg p-4 md:p-8 shadow-sm space-y-6">
        {/* Titre */}
        <div>
          <label htmlFor="titre" className="block text-sm font-medium text-ink mb-2">
            Titre du plan *
          </label>
          <input
            id="titre"
            type="text"
            value={titre}
            onChange={(e) => setTitre(e.target.value)}
            placeholder="Ex: Revision bac de maths"
            className="w-full px-4 py-3 rounded-lg border border-cream-dark focus:border-coral focus:outline-none text-ink"
            required
          />
        </div>

        {/* Matiere */}
        <div>
          <label htmlFor="matiere" className="block text-sm font-medium text-ink mb-2">
            Matiere
          </label>
          <div className="flex items-center gap-3">
            <span className="text-2xl">{getIconeMatiere(matiere)}</span>
            <select
              id="matiere"
              value={matiere}
              onChange={(e) => setMatiere(e.target.value)}
              className="flex-1 px-4 py-3 rounded-lg border border-cream-dark focus:border-coral focus:outline-none text-ink bg-white"
            >
              {MATIERES.map(m => (
                <option key={m.value} value={m.value}>{m.label}</option>
              ))}
            </select>
          </div>
        </div>

        {/* Description */}
        <div>
          <label htmlFor="description" className="block text-sm font-medium text-ink mb-2">
            Description
          </label>
          <textarea
            id="description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Ex: Chapitres 5 a 8, derivees et integrales"
            rows={3}
            className="w-full px-4 py-3 rounded-lg border border-cream-dark focus:border-coral focus:outline-none text-ink resize-none"
          />
        </div>

        {/* Date d'echeance */}
        <div>
          <label htmlFor="dateEcheance" className="block text-sm font-medium text-ink mb-2">
            Date de l'examen
          </label>
          <input
            id="dateEcheance"
            type="date"
            value={dateEcheance}
            onChange={(e) => setDateEcheance(e.target.value)}
            className="w-full px-4 py-3 rounded-lg border border-cream-dark focus:border-coral focus:outline-none text-ink"
          />
        </div>

        {/* Selection du cours */}
        <div>
          <label htmlFor="cours" className="block text-sm font-medium text-ink mb-2">
            Cours
          </label>
          {chargementCours ? (
            <div className="text-sm text-ink-muted py-4 text-center">Chargement des cours...</div>
          ) : coursDisponibles.length === 0 ? (
            <div className="bg-cream rounded-lg p-6 text-center">
              <p className="text-sm text-ink-muted mb-3">Aucun cours disponible</p>
              <button
                type="button"
                onClick={() => navigate('/scanner')}
                className="text-sm text-coral font-medium hover:underline"
              >
                Scanner un cours
              </button>
            </div>
          ) : (
            <select
              id="cours"
              value={coursSelectionne}
              onChange={(e) => setCoursSelectionne(e.target.value)}
              className="w-full px-4 py-3 rounded-lg border border-cream-dark focus:border-coral focus:outline-none text-ink bg-white"
            >
              <option value="">Choisir un cours</option>
              {coursDisponibles.map((cours) => (
                <option key={cours.id} value={cours.id}>
                  {cours.titre}{cours.matiere ? ` (${cours.matiere})` : ''}
                </option>
              ))}
            </select>
          )}
        </div>

        {/* Boutons */}
        <div className="flex justify-end gap-4 pt-4">
          <button
            type="button"
            onClick={() => navigate('/')}
            className="px-6 py-3 text-ink-light hover:text-ink transition-colors font-medium"
          >
            Annuler
          </button>
          <button
            type="submit"
            disabled={chargement || !titre.trim()}
            className="px-8 py-3 bg-teal text-white rounded-full font-semibold hover:bg-teal-light transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {chargement ? 'Creation...' : 'Creer le plan'}
          </button>
        </div>
      </form>
    </div>
  )
}
