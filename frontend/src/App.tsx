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
            <Route path="analyser" element={<Analyser />} />
            <Route path="progression" element={<Progression />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </AccessibiliteProvider>
  )
}
