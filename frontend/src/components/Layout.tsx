import { NavLink, Outlet } from 'react-router-dom'
import { useAuth } from '../api/auth'

const enlaces = [
  { to: '/', label: 'Panel' },
  { to: '/torneos', label: 'Torneos' },
  { to: '/cargas/nueva', label: 'Nueva carga' },
  { to: '/historial', label: 'Historial' },
]

export default function Layout() {
  const { usuario, logout } = useAuth()
  return (
    <div className="min-h-screen bg-slate-50 text-slate-800">
      <header className="bg-white border-b">
        <div className="max-w-6xl mx-auto px-4 py-3 flex flex-wrap items-center gap-x-4 gap-y-2">
          <span className="font-bold text-indigo-700 mr-auto md:mr-0">BecasApp</span>
          <nav className="order-last w-full md:order-none md:w-auto md:flex-1 flex gap-1 overflow-x-auto">
            {enlaces.map((e) => (
              <NavLink
                key={e.to}
                to={e.to}
                end={e.to === '/'}
                className={({ isActive }) =>
                  `px-3 py-1.5 rounded text-sm whitespace-nowrap ${isActive ? 'bg-indigo-100 text-indigo-800' : 'hover:bg-slate-100'}`
                }
              >
                {e.label}
              </NavLink>
            ))}
          </nav>
          <span className="text-sm text-slate-500 hidden sm:inline">{usuario?.correo}</span>
          <button onClick={logout} className="text-sm text-red-600 hover:underline">
            Salir
          </button>
        </div>
      </header>
      <main className="max-w-6xl mx-auto px-4 py-6">
        <Outlet />
      </main>
    </div>
  )
}
