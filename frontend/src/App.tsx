import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { AccessibiliteProvider } from './contexte/AccessibiliteContexte'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import Scanner from './pages/Scanner'
import Cours from './pages/Cours'
import Fiches from './pages/Fiches'
import Quiz from './pages/Quiz'
import Mindmap from './pages/Mindmap'
import Analyser from './pages/Analyser'
import Progression from './pages/Progression'
import Lexique from './pages/Lexique'
import ExamenBlanc from './pages/ExamenBlanc'
import PlanRevisionNouveau from './pages/PlanRevisionNouveau'
import PlanRevisionPage from './pages/PlanRevisionPage'
import Accessibilite from './pages/Accessibilite'

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
            <Route path="examen-blanc" element={<ExamenBlanc />} />
            <Route path="analyser" element={<Analyser />} />
            <Route path="progression" element={<Progression />} />
            <Route path="lexique" element={<Lexique />} />
            <Route path="plans/nouveau" element={<PlanRevisionNouveau />} />
            <Route path="plans/:id" element={<PlanRevisionPage />} />
            <Route path="accessibilite" element={<Accessibilite />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </AccessibiliteProvider>
  )
}
