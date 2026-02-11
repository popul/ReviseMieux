import { Outlet, NavLink, useLocation } from 'react-router-dom'
import { useEffect, useState } from 'react'
import { obtenirQuotas, listerPlansRevision } from '../services/api'
import type { StatutQuota, PlanRevisionResume } from '../services/api'

interface LienNavigation {
  vers: string
  libelle: string
  icone: string
}

const menuPrincipal: LienNavigation[] = [
  { vers: '/', libelle: 'Mes cours', icone: '🏠' },
]

const menuBibliotheque: LienNavigation[] = [
  { vers: '/cours', libelle: 'Mes cours', icone: '📚' },
  { vers: '/analyser', libelle: 'Analyser une copie', icone: '🔍' },
  { vers: '/accessibilite', libelle: 'Accessibilite', icone: '♿' },
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

function SectionPlansRevision({ plans }: { plans: PlanRevisionResume[] }) {
  const location = useLocation()

  return (
    <div className="mb-8" role="navigation" aria-label="Plans de revision">
      <div className="flex items-center justify-between mb-4 px-4">
        <div className="text-xs font-semibold text-ink-muted uppercase tracking-wider">
          Mes revisions
        </div>
        <NavLink
          to="/plans/nouveau"
          className="w-5 h-5 flex items-center justify-center rounded bg-white/10 text-white/50 hover:bg-coral hover:text-white transition-all text-xs"
          title="Nouveau plan de revision"
        >
          +
        </NavLink>
      </div>

      {plans.length === 0 ? (
        <NavLink
          to="/plans/nouveau"
          className="flex items-center gap-4 py-3.5 px-4 rounded-md text-[0.95rem] font-medium transition-all mb-1 text-white/40 hover:bg-white/[0.08] hover:text-white/70 border border-dashed border-white/20"
        >
          <span className="text-[1.2rem] w-6 text-center" aria-hidden="true">+</span>
          Creer un plan
        </NavLink>
      ) : (
        plans.map((plan) => {
          const estActif = location.pathname === `/plans/${plan.id}`
          const joursRestants = plan.dateEcheance
            ? Math.ceil((new Date(plan.dateEcheance).getTime() - Date.now()) / (1000 * 60 * 60 * 24))
            : null

          return (
            <NavLink
              key={plan.id}
              to={`/plans/${plan.id}`}
              className={`flex items-start gap-3 py-3 px-4 rounded-md text-sm font-medium transition-all mb-1 ${
                estActif
                  ? 'bg-coral text-white'
                  : 'text-white/70 hover:bg-white/[0.08] hover:text-white'
              }`}
            >
              <span className="text-[1.1rem] w-5 text-center mt-0.5 shrink-0" aria-hidden="true">
                {plan.iconeMatiere || '📋'}
              </span>
              <div className="flex-1 min-w-0">
                <div className="truncate">{plan.titre}</div>
                <div className="flex items-center gap-2 mt-1">
                  <span className={`text-[10px] ${estActif ? 'text-white/70' : 'text-white/40'}`}>
                    {plan.nombreCours} cours
                  </span>
                  {joursRestants !== null && joursRestants >= 0 && (
                    <span className={`text-[10px] px-1.5 py-0.5 rounded-full ${
                      joursRestants <= 3
                        ? 'bg-coral/30 text-coral-light'
                        : joursRestants <= 7
                          ? 'bg-gold/30 text-gold'
                          : estActif ? 'bg-white/20 text-white/70' : 'bg-white/10 text-white/40'
                    }`}>
                      J-{joursRestants}
                    </span>
                  )}
                  {/* Barre de progression */}
                  <div className="flex-1 h-1 bg-white/10 rounded-full overflow-hidden">
                    <div
                      className={`h-full rounded-full transition-all ${
                        estActif ? 'bg-white/50' : 'bg-teal/60'
                      }`}
                      style={{ width: `${plan.progression}%` }}
                    />
                  </div>
                </div>
              </div>
            </NavLink>
          )
        })
      )}
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
  const location = useLocation()
  const [quotas, setQuotas] = useState<StatutQuota | null>(null)
  const [plans, setPlans] = useState<PlanRevisionResume[]>([])
  const [menuOuvert, setMenuOuvert] = useState(false)

  useEffect(() => {
    obtenirQuotas()
      .then((response) => {
        if (response.succes && response.quotas) {
          setQuotas(response.quotas)
        }
      })
      .catch(() => {})
  }, [])

  // Refresh plans quand le pathname change (creation, navigation)
  useEffect(() => {
    listerPlansRevision()
      .then(setPlans)
      .catch(() => {})
  }, [location.pathname])

  // Fermer le menu mobile quand on navigue
  useEffect(() => {
    setMenuOuvert(false)
  }, [location.pathname])

  const contenuSidebar = (
    <>
      <a href="/" className="font-display text-2xl font-semibold text-white mb-12 block" aria-label="Revise mieux - Accueil">
        Revise<span className="text-coral">mieux</span>
      </a>

      <nav className="flex-1" aria-label="Menu principal">
        <SectionNavigation titre="Menu" liens={menuPrincipal} />
        <SectionPlansRevision plans={plans} />
        <SectionNavigation titre="Bibliotheque" liens={menuBibliotheque} />
      </nav>

      <AffichageQuotas quotas={quotas} />

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
    </>
  )

  return (
    <div className="flex min-h-screen">
      {/* Skip to content link for keyboard users */}
      <a href="#main-content" className="skip-to-content">
        Aller au contenu principal
      </a>

      {/* Header mobile */}
      <div className="fixed top-0 left-0 right-0 z-40 flex md:hidden items-center justify-between bg-ink px-4 py-3">
        <a href="/" className="font-display text-xl font-semibold text-white" aria-label="Revise mieux - Accueil">
          Revise<span className="text-coral">mieux</span>
        </a>
        <button
          onClick={() => setMenuOuvert(!menuOuvert)}
          className="w-10 h-10 flex items-center justify-center text-white rounded-md hover:bg-white/10 transition-colors"
          aria-label={menuOuvert ? 'Fermer le menu' : 'Ouvrir le menu'}
        >
          {menuOuvert ? (
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          ) : (
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
              <line x1="3" y1="6" x2="21" y2="6" />
              <line x1="3" y1="12" x2="21" y2="12" />
              <line x1="3" y1="18" x2="21" y2="18" />
            </svg>
          )}
        </button>
      </div>

      {/* Overlay mobile menu */}
      {menuOuvert && (
        <div className="fixed inset-0 z-50 md:hidden">
          <div className="absolute inset-0 bg-black/50" onClick={() => setMenuOuvert(false)} />
          <aside className="absolute inset-y-0 left-0 w-[280px] bg-ink p-8 flex flex-col overflow-y-auto" role="complementary" aria-label="Navigation principale">
            {contenuSidebar}
          </aside>
        </div>
      )}

      {/* Sidebar desktop */}
      <aside className="hidden md:flex bg-ink p-8 flex-col w-[280px] min-w-[280px] h-screen sticky top-0 overflow-y-auto" role="complementary" aria-label="Navigation principale">
        {contenuSidebar}
      </aside>

      {/* Main Content */}
      <main id="main-content" className="flex-1 p-4 md:p-12 pt-16 md:pt-12" role="main" tabIndex={-1}>
        <Outlet />
      </main>
    </div>
  )
}
