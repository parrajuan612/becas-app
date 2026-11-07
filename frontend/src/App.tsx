import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { AuthProvider, useAuth } from './api/auth'
import Layout from './components/Layout'
import Historial from './pages/Historial'
import Login from './pages/Login'
import NuevaCarga from './pages/NuevaCarga'
import Panel from './pages/Panel'
import Torneos from './pages/Torneos'

function Protegido() {
  const { usuario, cargando } = useAuth()
  if (cargando) return <p className="p-8 text-center text-slate-500">Cargando…</p>
  return usuario ? <Layout /> : <Navigate to="/login" replace />
}

export default function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route element={<Protegido />}>
            <Route index element={<Panel />} />
            <Route path="torneos" element={<Torneos />} />
            <Route path="cargas/nueva" element={<NuevaCarga />} />
            <Route path="historial" element={<Historial />} />
          </Route>
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  )
}
