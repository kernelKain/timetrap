import { BrowserRouter, Route, Routes } from 'react-router'
import { AppLayout } from './components/layout/AppLayout'
import { HomePage } from './pages/HomePage'
import { NotFoundPage } from './pages/NotFoundPage'
import { ResultPage } from './pages/ResultPage'
import { ScenarioPage } from './pages/ScenarioPage'

export default function App() {
  const basename = import.meta.env.BASE_URL === '/' ? undefined : import.meta.env.BASE_URL.replace(/\/$/, '')

  return <BrowserRouter basename={basename}><Routes><Route element={<AppLayout />}>
    <Route index element={<HomePage />} />
    <Route path="scenario/new" element={<ScenarioPage />} />
    <Route path="results/:analysisId" element={<ResultPage />} />
    <Route path="*" element={<NotFoundPage />} />
  </Route></Routes></BrowserRouter>
}
