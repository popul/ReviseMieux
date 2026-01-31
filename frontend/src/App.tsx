import { BrowserRouter, Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import Scanner from './pages/Scanner'
import Fiches from './pages/Fiches'
import Quiz from './pages/Quiz'

// Placeholder pages - à implémenter dans les prochaines étapes
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
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<Dashboard />} />
          <Route path="scanner" element={<Scanner />} />
          <Route path="cours" element={<PageEnConstruction titre="Mes cours" />} />
          <Route path="fiches" element={<Fiches />} />
          <Route path="quiz" element={<Quiz />} />
          <Route path="mindmap" element={<PageEnConstruction titre="Cartes mentales" />} />
          <Route path="analyser" element={<PageEnConstruction titre="Analyser une copie" />} />
          <Route path="progression" element={<PageEnConstruction titre="Ma progression" />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
