import { Link } from 'react-router-dom'

export default function Dashboard() {
  return (
    <>
      {/* Header */}
      <header className="mb-xl">
        <h1 className="font-display text-4xl font-bold text-ink mb-xs">
          Bienvenue sur Révise mieux 👋
        </h1>
        <p className="text-ink-light text-lg">
          Transforme tes cours en fiches de révision et quiz interactifs
        </p>
      </header>

      {/* Quick Actions */}
      <div className="grid grid-cols-2 gap-md mb-xl">
        <Link
          to="/scanner"
          className="bg-coral text-white rounded-lg p-lg flex items-center gap-md no-underline transition-all hover:-translate-y-1 hover:shadow-xl hover:bg-coral-dark"
        >
          <div className="w-14 h-14 rounded-md bg-white/20 flex items-center justify-center text-2xl flex-shrink-0">
            📸
          </div>
          <div>
            <h3 className="font-display text-lg font-semibold mb-1">Scanner un cours</h3>
            <p className="text-sm opacity-80">Transforme tes notes en fiches et quiz</p>
          </div>
        </Link>
        <Link
          to="/analyser"
          className="bg-white text-ink rounded-lg p-lg flex items-center gap-md no-underline transition-all border-2 border-transparent hover:-translate-y-1 hover:shadow-xl"
        >
          <div className="w-14 h-14 rounded-md bg-cream flex items-center justify-center text-2xl flex-shrink-0">
            📝
          </div>
          <div>
            <h3 className="font-display text-lg font-semibold mb-1">Analyser une copie</h3>
            <p className="text-sm text-ink-light">Identifie tes erreurs et progresse</p>
          </div>
        </Link>
      </div>

      {/* Stats */}
      <section className="mb-xl">
        <h2 className="font-display text-xl font-semibold text-ink mb-md">Tes statistiques</h2>
        <div className="grid grid-cols-4 gap-md">
          <div className="bg-white rounded-lg p-md text-center">
            <div className="font-display text-4xl font-bold text-coral mb-1">0</div>
            <div className="text-sm text-ink-light">Cours scannés</div>
          </div>
          <div className="bg-white rounded-lg p-md text-center">
            <div className="font-display text-4xl font-bold text-teal mb-1">0</div>
            <div className="text-sm text-ink-light">Quiz complétés</div>
          </div>
          <div className="bg-white rounded-lg p-md text-center">
            <div className="font-display text-4xl font-bold text-gold mb-1">0</div>
            <div className="text-sm text-ink-light">Fiches créées</div>
          </div>
          <div className="bg-white rounded-lg p-md text-center">
            <div className="font-display text-4xl font-bold text-success mb-1">—</div>
            <div className="text-sm text-ink-light">Score moyen</div>
          </div>
        </div>
      </section>

      {/* Empty state */}
      <section className="bg-white rounded-lg p-xl text-center">
        <div className="text-5xl mb-md">📚</div>
        <h2 className="font-display text-xl font-semibold text-ink mb-sm">
          Aucun cours pour le moment
        </h2>
        <p className="text-ink-light mb-lg max-w-md mx-auto">
          Commence par scanner un cours pour générer des fiches de révision et des quiz interactifs.
        </p>
        <Link
          to="/scanner"
          className="inline-flex items-center gap-2 bg-coral text-white px-lg py-3 rounded-full font-semibold no-underline transition-all hover:bg-coral-dark hover:-translate-y-0.5"
        >
          <span>📸</span>
          Scanner mon premier cours
        </Link>
      </section>
    </>
  )
}
