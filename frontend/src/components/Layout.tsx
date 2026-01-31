import { Outlet, NavLink } from 'react-router-dom'
import { useEffect, useState } from 'react'
import { obtenirQuotas } from '../services/api'
import type { StatutQuota } from '../services/api'

interface LienNavigation {
  vers: string
  libelle: string
  icone: string
}

const menuPrincipal: LienNavigation[] = [
  { vers: '/', libelle: 'Tableau de bord', icone: '🏠' },
  { vers: '/scanner', libelle: 'Scanner un cours', icone: '📸' },
  { vers: '/cours', libelle: 'Mes cours', icone: '📚' },
]

const menuRevision: LienNavigation[] = [
  { vers: '/fiches', libelle: 'Fiches de révision', icone: '📄' },
  { vers: '/quiz', libelle: 'Quiz', icone: '🎯' },
  { vers: '/mindmap', libelle: 'Cartes mentales', icone: '🧠' },
]

const menuAnalyse: LienNavigation[] = [
  { vers: '/analyser', libelle: 'Analyser une copie', icone: '📝' },
  { vers: '/progression', libelle: 'Ma progression', icone: '📊' },
]

function SectionNavigation({ titre, liens }: { titre: string; liens: LienNavigation[] }) {
  return (
    <div className="mb-lg">
      <div className="text-xs font-semibold text-ink-muted uppercase tracking-wider mb-sm px-sm">
        {titre}
      </div>
      {liens.map((lien) => (
        <NavLink
          key={lien.vers}
          to={lien.vers}
          className={({ isActive }) =>
            `flex items-center gap-sm py-3.5 px-sm rounded-md text-[0.95rem] font-medium transition-all mb-1 ${
              isActive
                ? 'bg-coral text-white'
                : 'text-white/70 hover:bg-white/[0.08] hover:text-white'
            }`
          }
        >
          <span className="text-[1.2rem] w-6 text-center">{lien.icone}</span>
          {lien.libelle}
        </NavLink>
      ))}
    </div>
  )
}

function AffichageQuotas({ quotas }: { quotas: StatutQuota | null }) {
  if (!quotas) return null

  const pourcentageOCR = (quotas.pagesOcrUtilisees / quotas.pagesOcrMax) * 100
  const pourcentageGen = (quotas.generationsUtilisees / quotas.generationsMax) * 100

  const couleurBarre = (pourcent: number) => {
    if (pourcent >= 90) return 'bg-coral'
    if (pourcent >= 70) return 'bg-gold'
    return 'bg-teal'
  }

  return (
    <div className="mb-lg p-sm bg-white/5 rounded-md">
      <div className="text-xs font-semibold text-ink-muted uppercase tracking-wider mb-sm">
        Quotas du jour
      </div>
      <div className="space-y-2">
        <div>
          <div className="flex justify-between text-xs text-white/70 mb-1">
            <span>OCR</span>
            <span>{quotas.pagesOcrRestantes} restants</span>
          </div>
          <div className="h-1.5 bg-white/10 rounded-full overflow-hidden">
            <div
              className={`h-full ${couleurBarre(pourcentageOCR)} transition-all`}
              style={{ width: `${pourcentageOCR}%` }}
            />
          </div>
        </div>
        <div>
          <div className="flex justify-between text-xs text-white/70 mb-1">
            <span>Générations</span>
            <span>{quotas.generationsRestantes} restants</span>
          </div>
          <div className="h-1.5 bg-white/10 rounded-full overflow-hidden">
            <div
              className={`h-full ${couleurBarre(pourcentageGen)} transition-all`}
              style={{ width: `${pourcentageGen}%` }}
            />
          </div>
        </div>
      </div>
    </div>
  )
}

export default function Layout() {
  const [quotas, setQuotas] = useState<StatutQuota | null>(null)

  useEffect(() => {
    obtenirQuotas()
      .then((response) => {
        if (response.succes && response.quotas) {
          setQuotas(response.quotas)
        }
      })
      .catch(() => {
        // Silencieusement ignorer les erreurs de quota
      })
  }, [])

  return (
    <div className="grid grid-cols-[280px_1fr] min-h-screen">
      {/* Sidebar */}
      <aside className="bg-ink p-lg flex flex-col fixed w-[280px] h-screen overflow-y-auto">
        <a href="/" className="font-display text-2xl font-semibold text-white mb-xl">
          Révise<span className="text-coral">mieux</span>
        </a>

        <nav className="flex-1">
          <SectionNavigation titre="Menu" liens={menuPrincipal} />
          <SectionNavigation titre="Réviser" liens={menuRevision} />
          <SectionNavigation titre="Analyse" liens={menuAnalyse} />
        </nav>

        <AffichageQuotas quotas={quotas} />

        <div className="pt-lg border-t border-white/10">
          <div className="flex items-center gap-sm p-sm rounded-md transition-colors cursor-pointer hover:bg-white/[0.08]">
            <div className="w-10 h-10 rounded-full bg-gradient-to-br from-coral to-gold flex items-center justify-center font-semibold text-white text-sm">
              RM
            </div>
            <div className="flex-1">
              <div className="font-semibold text-white text-sm">Révise mieux</div>
              <div className="text-xs text-ink-muted">Mode anonyme</div>
            </div>
          </div>
        </div>
      </aside>

      {/* Main Content */}
      <main className="ml-[280px] p-xl max-w-[1200px]">
        <Outlet />
      </main>
    </div>
  )
}
