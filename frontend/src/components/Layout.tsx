import { Outlet, NavLink } from 'react-router-dom'
import { useEffect, useState } from 'react'
import { obtenirQuotas } from '../services/api'
import type { StatutQuota } from '../services/api'
import AccessibiliteControles from './AccessibiliteControles'

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
    <div className="mb-8" role="navigation" aria-label={titre}>
      <div className="text-xs font-semibold text-ink-muted uppercase tracking-wider mb-4 px-4">
        {titre}
      </div>
      {liens.map((lien) => (
        <NavLink
          key={lien.vers}
          to={lien.vers}
          end={lien.vers === '/'}
          className={({ isActive }) =>
            `flex items-center gap-4 py-3.5 px-4 rounded-md text-[0.95rem] font-medium transition-all mb-1 ${
              isActive
                ? 'bg-coral text-white'
                : 'text-white/70 hover:bg-white/[0.08] hover:text-white'
            }`
          }
        >
          <span className="text-[1.2rem] w-6 text-center" aria-hidden="true">{lien.icone}</span>
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
    <div className="mb-8 p-4 bg-white/5 rounded-md" role="region" aria-label="Quotas journaliers">
      <div className="text-xs font-semibold text-ink-muted uppercase tracking-wider mb-4">
        Quotas du jour
      </div>
      <div className="space-y-2">
        <div>
          <div className="flex justify-between text-xs text-white/70 mb-1">
            <span id="quota-ocr-label">OCR</span>
            <span aria-live="polite">{quotas.pagesOcrRestantes} restants</span>
          </div>
          <div
            className="h-1.5 bg-white/10 rounded-full overflow-hidden"
            role="progressbar"
            aria-labelledby="quota-ocr-label"
            aria-valuenow={quotas.pagesOcrUtilisees}
            aria-valuemin={0}
            aria-valuemax={quotas.pagesOcrMax}
          >
            <div
              className={`h-full ${couleurBarre(pourcentageOCR)} transition-all`}
              style={{ width: `${pourcentageOCR}%` }}
            />
          </div>
        </div>
        <div>
          <div className="flex justify-between text-xs text-white/70 mb-1">
            <span id="quota-gen-label">Generations</span>
            <span aria-live="polite">{quotas.generationsRestantes} restants</span>
          </div>
          <div
            className="h-1.5 bg-white/10 rounded-full overflow-hidden"
            role="progressbar"
            aria-labelledby="quota-gen-label"
            aria-valuenow={quotas.generationsUtilisees}
            aria-valuemin={0}
            aria-valuemax={quotas.generationsMax}
          >
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
    <div className="flex min-h-screen">
      {/* Skip to content link for keyboard users */}
      <a href="#main-content" className="skip-to-content">
        Aller au contenu principal
      </a>

      {/* Sidebar */}
      <aside className="bg-ink p-8 flex flex-col w-[280px] min-w-[280px] h-screen sticky top-0 overflow-y-auto" role="complementary" aria-label="Navigation principale">
        <a href="/" className="font-display text-2xl font-semibold text-white mb-12" aria-label="Revise mieux - Accueil">
          Revise<span className="text-coral">mieux</span>
        </a>

        <nav className="flex-1" aria-label="Menu principal">
          <SectionNavigation titre="Menu" liens={menuPrincipal} />
          <SectionNavigation titre="Reviser" liens={menuRevision} />
          <SectionNavigation titre="Analyse" liens={menuAnalyse} />
        </nav>

        <AffichageQuotas quotas={quotas} />

        {/* Accessibility controls */}
        <AccessibiliteControles />

        <div className="pt-8 border-t border-white/10 mt-8">
          <div className="flex items-center gap-4 p-4 rounded-md transition-colors cursor-pointer hover:bg-white/[0.08]">
            <div className="w-10 h-10 rounded-full bg-gradient-to-br from-coral to-gold flex items-center justify-center font-semibold text-white text-sm" aria-hidden="true">
              RM
            </div>
            <div className="flex-1">
              <div className="font-semibold text-white text-sm">Revise mieux</div>
              <div className="text-xs text-ink-muted">Mode anonyme</div>
            </div>
          </div>
        </div>
      </aside>

      {/* Main Content */}
      <main id="main-content" className="flex-1 p-12" role="main" tabIndex={-1}>
        <Outlet />
      </main>
    </div>
  )
}
