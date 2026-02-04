import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { AccessibiliteProvider } from './contexte/AccessibiliteContexte'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import Scanner from './pages/Scanner'
import Cours from './pages/Cours'
import Fiches from './pages/Fiches'
import Quiz from './pages/Quiz'
import Mindmap from './pages/Mindmap'

// Placeholder pages - a implementer dans les prochaines etapes
function PageEnConstruction({ titre }: { titre: string }) {
  return (
    <div className="text-center py-xl">
      <div className="text-5xl mb-md">🚧</div>
      <h1 className="font-display text-2xl font-semibold text-ink mb-sm">{titre}</h1>
      <p className="text-ink-light">Cette page sera disponible prochainement.</p>
    </div>
  )
}

export default function App() {
  return (
    <AccessibiliteProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<Layout />}>
            <Route index element={<Dashboard />} />
            <Route path="scanner" element={<Scanner />} />
            <Route path="cours" element={<Cours />} />
            <Route path="fiches" element={<Fiches />} />
            <Route path="quiz" element={<Quiz />} />
            <Route path="mindmap" element={<Mindmap />} />
            <Route path="analyser" element={<PageEnConstruction titre="Analyser une copie" />} />
            <Route path="progression" element={<PageEnConstruction titre="Ma progression" />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </AccessibiliteProvider>
  )
}
